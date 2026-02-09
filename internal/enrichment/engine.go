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
	CampaignID    int64
	JobID         int64
	SourceTable   string
	SourceID      int64
	Content       string // Source content (Markdown)
	Entity        models.Entity
	OtherEntities []models.Entity       // Other entities mentioned in the same content
	Relationships []models.Relationship // Existing relationships for this entity
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

	// Relationship suggestions.
	for _, rs := range resp.Relationships {
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
