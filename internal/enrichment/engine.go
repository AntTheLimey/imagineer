/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

// Package enrichment provides the LLM-based enrichment engine that
// analyses content alongside entity state and produces structured
// suggestions for description updates, log entries, and relationships.
package enrichment

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/antonypegg/imagineer/internal/models"
)

// Engine orchestrates LLM-based entity enrichment. It builds prompts
// from content and entity state, calls an LLM provider, and converts
// the structured response into ContentAnalysisItems for triage.
type Engine struct {
	db *database.DB
}

// NewEngine creates a new enrichment engine. The database reference is
// stored for future use (e.g., persisting items directly) but is not
// required by EnrichEntity, which receives all data via EnrichmentInput.
func NewEngine(db *database.DB) *Engine {
	return &Engine{db: db}
}

// EnrichmentInput contains everything needed to enrich a single entity
// from a content source.
type EnrichmentInput struct {
	CampaignID      int64
	JobID           int64
	SourceTable     string
	SourceID        int64
	Content         string // Source content (Markdown)
	Entity          models.Entity
	OtherEntities   []models.Entity       // Other entities mentioned in the same content
	Relationships   []models.Relationship // Existing relationships for this entity
	CampaignResults []models.SearchResult // RAG: campaign vector search results
	GameSystemYAML  string                // RAG: game system schema
}

// EnrichEntity sends content and entity state to the LLM and returns
// enrichment items (description updates, log entries, relationship
// suggestions). Items are created with phase="enrichment" and
// resolution="pending".
func (e *Engine) EnrichEntity(
	ctx context.Context,
	provider llm.Provider,
	input EnrichmentInput,
) ([]models.ContentAnalysisItem, error) {
	// Validate required input fields.
	if input.Entity.ID == 0 {
		return nil, fmt.Errorf("enrichment: entity ID is required")
	}
	if input.Entity.Name == "" {
		return nil, fmt.Errorf("enrichment: entity name is required")
	}
	if input.Content == "" {
		// Nothing to enrich; return empty slice without error.
		return []models.ContentAnalysisItem{}, nil
	}

	systemPrompt := buildSystemPrompt()
	userPrompt := buildUserPrompt(input)

	resp, err := provider.Complete(ctx, llm.CompletionRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    2048,
		Temperature:  0.3,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM completion failed: %w", err)
	}

	parsed, err := parseEnrichmentResponse(resp.Content)
	if err != nil {
		// Log but do not propagate parse errors; return empty items
		// for graceful degradation.
		log.Printf(
			"enrichment: failed to parse LLM response for entity %d: %v",
			input.Entity.ID, err,
		)
		return []models.ContentAnalysisItem{}, nil
	}

	items := convertToItems(input, parsed)
	return items, nil
}

// DetectNewEntities analyses source content to identify named entities
// that are mentioned but do not yet exist in the campaign database.
// It returns ContentAnalysisItems with detection_type="new_entity_suggestion"
// for each candidate. On LLM or parse errors it degrades gracefully,
// returning an empty slice rather than propagating the error.
func (e *Engine) DetectNewEntities(
	ctx context.Context,
	provider llm.Provider,
	campaignID int64,
	jobID int64,
	sourceContent string,
	knownEntities []models.Entity,
) ([]models.ContentAnalysisItem, error) {
	if sourceContent == "" {
		return []models.ContentAnalysisItem{}, nil
	}

	systemPrompt := buildNewEntityDetectionSystemPrompt()
	userPrompt := buildNewEntityDetectionUserPrompt(
		sourceContent, knownEntities,
	)

	resp, err := provider.Complete(ctx, llm.CompletionRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    2048,
		Temperature:  0.3,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM completion failed: %w", err)
	}

	parsed, err := parseNewEntityResponse(resp.Content)
	if err != nil {
		log.Printf(
			"enrichment: failed to parse new-entity LLM response "+
				"for campaign %d: %v",
			campaignID, err,
		)
		return []models.ContentAnalysisItem{}, nil
	}

	items := convertNewEntitiesToItems(campaignID, jobID, parsed)
	return items, nil
}

// convertToItems transforms a parsed enrichment response into a slice
// of ContentAnalysisItems ready for triage.
func convertToItems(input EnrichmentInput, resp *enrichmentResponse) []models.ContentAnalysisItem {
	if resp == nil {
		return []models.ContentAnalysisItem{}
	}

	now := time.Now()
	entityID := input.Entity.ID
	items := []models.ContentAnalysisItem{}

	// Description update suggestions.
	for _, du := range resp.DescriptionUpdates {
		content, err := json.Marshal(du)
		if err != nil {
			log.Printf(
				"enrichment: failed to marshal description update for entity %d: %v",
				entityID, err,
			)
			continue
		}

		items = append(items, models.ContentAnalysisItem{
			JobID:            input.JobID,
			DetectionType:    "description_update",
			MatchedText:      input.Entity.Name,
			EntityID:         &entityID,
			Resolution:       "pending",
			SuggestedContent: json.RawMessage(content),
			Phase:            "enrichment",
			CreatedAt:        now,
		})
	}

	// Log entry suggestions.
	for _, le := range resp.LogEntries {
		content, err := json.Marshal(le)
		if err != nil {
			log.Printf(
				"enrichment: failed to marshal log entry for entity %d: %v",
				entityID, err,
			)
			continue
		}

		items = append(items, models.ContentAnalysisItem{
			JobID:            input.JobID,
			DetectionType:    "log_entry",
			MatchedText:      input.Entity.Name,
			EntityID:         &entityID,
			Resolution:       "pending",
			SuggestedContent: json.RawMessage(content),
			Phase:            "enrichment",
			CreatedAt:        now,
		})
	}

	// Build a set of existing entity pairs (both directions) for
	// deduplication. Any existing relationship between two entities
	// means we skip new suggestions for that pair.
	type entityPair struct {
		a, b int64
	}
	existingPairs := make(map[entityPair]bool)
	for _, rel := range input.Relationships {
		// Normalise the pair so (a,b) and (b,a) match.
		a, b := rel.SourceEntityID, rel.TargetEntityID
		if a > b {
			a, b = b, a
		}
		existingPairs[entityPair{a, b}] = true
	}

	// Relationship suggestions.
	for _, rs := range resp.Relationships {
		// Skip if a relationship already exists between these entities.
		a, b := rs.SourceEntityID, rs.TargetEntityID
		if a > b {
			a, b = b, a
		}
		if existingPairs[entityPair{a, b}] {
			log.Printf(
				"enrichment: skipping duplicate relationship suggestion "+
					"between entities %d and %d for entity %d",
				rs.SourceEntityID, rs.TargetEntityID, entityID,
			)
			continue
		}

		content, err := json.Marshal(rs)
		if err != nil {
			log.Printf(
				"enrichment: failed to marshal relationship suggestion for entity %d: %v",
				entityID, err,
			)
			continue
		}

		items = append(items, models.ContentAnalysisItem{
			JobID:            input.JobID,
			DetectionType:    "relationship_suggestion",
			MatchedText:      input.Entity.Name,
			EntityID:         &entityID,
			Resolution:       "pending",
			SuggestedContent: json.RawMessage(content),
			Phase:            "enrichment",
			CreatedAt:        now,
		})
	}

	return items
}
