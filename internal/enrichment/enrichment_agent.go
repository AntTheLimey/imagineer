/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package enrichment

import (
	"context"
	"log"
	"strings"

	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/antonypegg/imagineer/internal/models"
)

// EnrichmentAgent adapts the enrichment Engine as a PipelineAgent,
// allowing it to participate in the multi-stage analysis pipeline.
// It enriches entities mentioned in the source content and detects
// new entity suggestions.
type EnrichmentAgent struct {
	engine *Engine
	db     *database.DB
}

// NewEnrichmentAgent creates a new EnrichmentAgent wrapping a fresh
// Engine instance backed by the given database handle.
func NewEnrichmentAgent(db *database.DB) *EnrichmentAgent {
	return &EnrichmentAgent{
		engine: NewEngine(db),
		db:     db,
	}
}

// Name returns the unique identifier for this pipeline agent.
func (a *EnrichmentAgent) Name() string {
	return "enrichment"
}

// DependsOn returns the names of agents that must run before this one.
// The enrichment agent has no dependencies and can run independently.
func (a *EnrichmentAgent) DependsOn() []string {
	return nil
}

// Run executes entity enrichment and new-entity detection against the
// provided content. It identifies entities mentioned in the content,
// enriches each one via the LLM engine, and runs new-entity detection.
// Individual entity failures are logged and skipped so that a single
// failure does not abort the entire pipeline stage.
func (a *EnrichmentAgent) Run(
	ctx context.Context,
	provider llm.Provider,
	input PipelineInput,
) ([]models.ContentAnalysisItem, error) {
	if input.Content == "" {
		return []models.ContentAnalysisItem{}, nil
	}

	// Determine entities to enrich. Prefer entities provided in the
	// pipeline input; fall back to scanning content for entity name
	// mentions.
	var allCampaignEntities []models.Entity
	entities := input.Entities
	if len(entities) == 0 {
		var err error
		allCampaignEntities, err = a.db.ListEntitiesByCampaign(ctx, input.CampaignID)
		if err != nil {
			log.Printf(
				"enrichment-agent: failed to list entities for campaign %d: %v",
				input.CampaignID, err,
			)
			// Cannot determine entities; run only new-entity detection.
			allCampaignEntities = nil
		}

		lowerContent := strings.ToLower(input.Content)
		for _, entity := range allCampaignEntities {
			if entity.Name == "" {
				continue
			}
			if strings.Contains(lowerContent, strings.ToLower(entity.Name)) {
				entities = append(entities, entity)
			}
		}
	}

	var allItems []models.ContentAnalysisItem

	// Build the full entity list for the OtherEntities field and for
	// new-entity detection. DetectNewEntities needs ALL campaign
	// entities (not just the mentioned subset) so it can check for
	// duplicates against every existing entity.
	allKnownEntities := allCampaignEntities
	if len(input.Entities) > 0 {
		// Input entities were explicit; load campaign entities for the
		// full known-entity list used by DetectNewEntities.
		campaignEntities, err := a.db.ListEntitiesByCampaign(ctx, input.CampaignID)
		if err != nil {
			log.Printf(
				"enrichment-agent: failed to list campaign entities for new-entity detection: %v",
				err,
			)
			allKnownEntities = entities
		} else {
			allKnownEntities = campaignEntities
		}
	} else if allKnownEntities == nil {
		// Fallback: if ListEntitiesByCampaign failed above, use
		// whatever mentioned entities we found.
		allKnownEntities = entities
	}

	// Enrich each entity individually.
	for i, entity := range entities {
		if ctx.Err() != nil {
			log.Printf("enrichment-agent: context cancelled, stopping enrichment")
			break
		}

		// Build the OtherEntities list (all entities minus the current one).
		otherEntities := make([]models.Entity, 0, len(entities)-1)
		for j, e := range entities {
			if j != i {
				otherEntities = append(otherEntities, e)
			}
		}

		// Fetch relationships for the current entity.
		relationships, err := a.db.GetEntityRelationships(ctx, entity.ID)
		if err != nil {
			log.Printf(
				"enrichment-agent: failed to get relationships for entity %d: %v",
				entity.ID, err,
			)
			relationships = nil
		}

		enrichInput := EnrichmentInput{
			CampaignID:    input.CampaignID,
			JobID:         input.JobID,
			SourceTable:   input.SourceTable,
			SourceID:      input.SourceID,
			Content:       input.Content,
			Entity:        entity,
			OtherEntities: otherEntities,
			Relationships: relationships,
		}

		items, err := a.engine.EnrichEntity(ctx, provider, enrichInput)
		if err != nil {
			log.Printf(
				"enrichment-agent: failed to enrich entity %d (%q): %v",
				entity.ID, entity.Name, err,
			)
			continue
		}

		allItems = append(allItems, items...)
	}

	// Detect new entities not yet in the campaign database.
	if ctx.Err() == nil {
		newEntityItems, err := a.engine.DetectNewEntities(
			ctx, provider, input.CampaignID, input.JobID,
			input.Content, allKnownEntities,
		)
		if err != nil {
			log.Printf(
				"enrichment-agent: new-entity detection failed for campaign %d: %v",
				input.CampaignID, err,
			)
		} else {
			allItems = append(allItems, newEntityItems...)
		}
	}

	if allItems == nil {
		allItems = []models.ContentAnalysisItem{}
	}

	return allItems, nil
}
