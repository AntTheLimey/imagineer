/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

// Package graph provides the graph expert agent for validating
// relationship suggestions and checking graph hygiene after entity
// enrichment within the enrichment pipeline.
package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/enrichment"
	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/antonypegg/imagineer/internal/models"
)

// Expert implements the enrichment.PipelineAgent interface for graph
// hygiene validation. It performs structural checks (orphaned entities,
// type pair validation) without an LLM and semantic checks (redundant
// or implied relationships) using an LLM. The agent runs after the
// enrichment agent and inspects its relationship suggestions.
type Expert struct {
	db *database.DB
}

// NewExpert creates a new graph expert agent with the given database
// handle, which is needed for querying relationship type constraints.
func NewExpert(db *database.DB) *Expert {
	return &Expert{db: db}
}

// Name returns the unique identifier for this pipeline agent.
func (e *Expert) Name() string {
	return "graph-expert"
}

// DependsOn returns the names of agents that must run before this one.
// The graph expert depends on the enrichment agent so it can inspect
// relationship suggestions produced during enrichment.
func (e *Expert) DependsOn() []string {
	return []string{"enrichment"}
}

// Run executes graph hygiene analysis. It performs two categories of
// checks:
//
//  1. Structural checks (no LLM needed): orphaned entity detection
//     and type pair constraint validation.
//  2. Semantic checks (LLM needed): redundant and implied relationship
//     detection among proposed and existing edges.
//
// If structural checks succeed but the LLM call fails, structural
// findings are still returned (graceful degradation).
func (e *Expert) Run(
	ctx context.Context,
	provider llm.Provider,
	input enrichment.PipelineInput,
) ([]models.ContentAnalysisItem, error) {
	if len(input.Entities) == 0 {
		return []models.ContentAnalysisItem{}, nil
	}

	now := time.Now()
	var allItems []models.ContentAnalysisItem

	// -----------------------------------------------------------------
	// 1. Structural checks (no LLM required)
	// -----------------------------------------------------------------

	// 1a. Check for orphaned entities (zero relationships).
	orphans := CheckOrphanedEntities(input.Entities, input.Relationships)
	for _, orphan := range orphans {
		detail, err := json.Marshal(map[string]interface{}{
			"entityId":   orphan.ID,
			"entityName": orphan.Name,
			"entityType": string(orphan.EntityType),
			"description": "Entity has no relationships. " +
				"Consider adding connections to other entities " +
				"or verifying that this entity is relevant.",
		})
		if err != nil {
			log.Printf(
				"graph-expert: failed to marshal orphan finding for entity %d: %v",
				orphan.ID, err,
			)
			continue
		}

		entityID := orphan.ID
		allItems = append(allItems, models.ContentAnalysisItem{
			JobID:            input.JobID,
			DetectionType:    "orphan_warning",
			MatchedText:      orphan.Name,
			EntityID:         &entityID,
			Resolution:       "pending",
			SuggestedContent: json.RawMessage(detail),
			Phase:            "enrichment",
			CreatedAt:        now,
		})
	}

	// 1b. Validate type pairs for relationship suggestions from the
	// enrichment agent. Filter PriorResults for relationship_suggestion
	// items produced by the enrichment agent.
	var relSuggestions []models.ContentAnalysisItem
	for _, item := range input.PriorResults {
		if item.DetectionType == "relationship_suggestion" {
			relSuggestions = append(relSuggestions, item)
		}
	}

	if len(relSuggestions) > 0 {
		violations, err := ValidateTypePairs(
			ctx, e.db, input.CampaignID,
			relSuggestions, input.Entities,
		)
		if err != nil {
			log.Printf(
				"graph-expert: type pair validation failed: %v", err,
			)
			// Continue with other checks; do not abort.
		} else {
			for _, v := range violations {
				detail, err := json.Marshal(map[string]interface{}{
					"sourceEntityId":   v.SourceEntityID,
					"sourceEntityName": v.SourceEntityName,
					"sourceEntityType": v.SourceEntityType,
					"targetEntityId":   v.TargetEntityID,
					"targetEntityName": v.TargetEntityName,
					"targetEntityType": v.TargetEntityType,
					"relationshipType": v.RelationshipType,
					"validPairs":       v.ValidPairs,
					"description": "Relationship type " +
						v.RelationshipType +
						" is not valid between " +
						v.SourceEntityType + " and " +
						v.TargetEntityType + ".",
				})
				if err != nil {
					log.Printf(
						"graph-expert: failed to marshal type pair violation: %v",
						err,
					)
					continue
				}

				allItems = append(allItems, models.ContentAnalysisItem{
					JobID:            input.JobID,
					DetectionType:    "invalid_type_pair",
					MatchedText:      v.RelationshipType,
					Resolution:       "pending",
					SuggestedContent: json.RawMessage(detail),
					Phase:            "enrichment",
					CreatedAt:        now,
				})
			}
		}
	}

	// 1c. Cardinality check: verify that existing relationships plus
	// proposed suggestions do not exceed per-type cardinality limits.
	if e.db != nil {
		cardViolations, err := CheckCardinality(
			ctx, e.db, input.CampaignID,
			relSuggestions,
		)
		if err != nil {
			log.Printf(
				"graph-expert: cardinality check failed: %v", err,
			)
			// Continue with other checks; do not abort.
		} else {
			for _, v := range cardViolations {
				detail, err := json.Marshal(map[string]interface{}{
					"entityId":         v.EntityID,
					"entityName":       v.EntityName,
					"entityType":       v.EntityType,
					"relationshipType": v.RelationshipType,
					"direction":        v.Direction,
					"currentCount":     v.CurrentCount,
					"maxAllowed":       v.MaxAllowed,
					"description": fmt.Sprintf(
						"Entity %s would have %d %s relationships as %s, "+
							"exceeding the limit of %d.",
						v.EntityName, v.CurrentCount,
						v.RelationshipType, v.Direction,
						v.MaxAllowed,
					),
				})
				if err != nil {
					log.Printf(
						"graph-expert: failed to marshal cardinality violation: %v",
						err,
					)
					continue
				}

				entityID := v.EntityID
				allItems = append(allItems, models.ContentAnalysisItem{
					JobID:            input.JobID,
					DetectionType:    "cardinality_violation",
					MatchedText:      v.RelationshipType,
					EntityID:         &entityID,
					Resolution:       "pending",
					SuggestedContent: json.RawMessage(detail),
					Phase:            "enrichment",
					CreatedAt:        now,
				})
			}
		}
	}

	// 1d. Required relationships: verify that every entity of a type
	// with required relationship rules participates in at least one
	// relationship of each required type.
	if e.db != nil {
		reqViolations, err := CheckRequiredRelationships(
			ctx, e.db, input.CampaignID,
		)
		if err != nil {
			log.Printf(
				"graph-expert: required relationship check failed: %v",
				err,
			)
			// Continue with other checks; do not abort.
		} else {
			for _, v := range reqViolations {
				detail, err := json.Marshal(map[string]interface{}{
					"entityId":                v.EntityID,
					"entityName":              v.EntityName,
					"entityType":              v.EntityType,
					"missingRelationshipType": v.MissingRelationshipType,
					"description": fmt.Sprintf(
						"Entity %s (%s) is missing a required %s "+
							"relationship.",
						v.EntityName, v.EntityType,
						v.MissingRelationshipType,
					),
				})
				if err != nil {
					log.Printf(
						"graph-expert: failed to marshal required relationship violation: %v",
						err,
					)
					continue
				}

				entityID := v.EntityID
				allItems = append(allItems, models.ContentAnalysisItem{
					JobID:            input.JobID,
					DetectionType:    "missing_required",
					MatchedText:      v.MissingRelationshipType,
					EntityID:         &entityID,
					Resolution:       "pending",
					SuggestedContent: json.RawMessage(detail),
					Phase:            "enrichment",
					CreatedAt:        now,
				})
			}
		}
	}

	// -----------------------------------------------------------------
	// 2. Semantic checks (LLM required)
	// -----------------------------------------------------------------

	// Only attempt semantic checks when there are relationships or
	// relationship suggestions to analyse and an LLM provider is
	// available.
	if provider != nil &&
		(len(input.Relationships) > 0 || len(relSuggestions) > 0) {
		semanticItems := e.runSemanticChecks(
			ctx, provider, input, relSuggestions, now,
		)
		allItems = append(allItems, semanticItems...)
	}

	if allItems == nil {
		allItems = []models.ContentAnalysisItem{}
	}

	// Filter out findings that match existing constraint overrides
	// so that acknowledged violations are not re-reported.
	if e.db != nil {
		allItems = FilterOverriddenFindings(
			ctx, e.db, input.CampaignID, allItems,
		)
	}

	return allItems, nil
}

// runSemanticChecks calls the LLM to identify redundant or implied
// relationships. If the LLM call fails, findings are logged and an
// empty slice is returned so that structural findings are preserved.
func (e *Expert) runSemanticChecks(
	ctx context.Context,
	provider llm.Provider,
	input enrichment.PipelineInput,
	relSuggestions []models.ContentAnalysisItem,
	now time.Time,
) []models.ContentAnalysisItem {
	systemPrompt := buildSystemPrompt()
	userPrompt := buildUserPrompt(input, relSuggestions)

	resp, err := provider.Complete(ctx, llm.CompletionRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    2048,
		Temperature:  0.2,
	})
	if err != nil {
		log.Printf(
			"graph-expert: LLM call failed for job %d, returning structural findings only: %v",
			input.JobID, err,
		)
		return nil
	}

	parsed, err := parseGraphResponse(resp.Content)
	if err != nil {
		log.Printf(
			"graph-expert: parse error for job %d, returning structural findings only: %v",
			input.JobID, err,
		)
		return nil
	}

	return convertLLMFindings(input.JobID, parsed, now)
}

// convertLLMFindings transforms parsed LLM findings into
// ContentAnalysisItems.
func convertLLMFindings(
	jobID int64,
	resp *graphResponse,
	now time.Time,
) []models.ContentAnalysisItem {
	if resp == nil || len(resp.Findings) == 0 {
		return nil
	}

	items := make(
		[]models.ContentAnalysisItem, 0, len(resp.Findings),
	)

	for _, f := range resp.Findings {
		detail, err := json.Marshal(map[string]interface{}{
			"findingType":      f.FindingType,
			"description":      f.Description,
			"involvedEntities": f.InvolvedEntities,
			"suggestion":       f.Suggestion,
		})
		if err != nil {
			log.Printf(
				"graph-expert: failed to marshal LLM finding for job %d: %v",
				jobID, err,
			)
			continue
		}

		items = append(items, models.ContentAnalysisItem{
			JobID:            jobID,
			DetectionType:    findingTypeToDetectionType(f.FindingType),
			MatchedText:      f.Description,
			Resolution:       "pending",
			SuggestedContent: json.RawMessage(detail),
			Phase:            "enrichment",
			CreatedAt:        now,
		})
	}

	return items
}
