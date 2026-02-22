/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package canon

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/antonypegg/imagineer/internal/enrichment"
	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/antonypegg/imagineer/internal/models"
)

// Expert implements the enrichment.PipelineAgent interface for canon
// contradiction detection. It compares new campaign content against
// established facts retrieved via RAG context and flags genuine
// contradictions for GM review.
type Expert struct{}

// NewExpert creates a new canon expert agent.
func NewExpert() *Expert {
	return &Expert{}
}

// Name returns the unique identifier for this pipeline agent.
func (e *Expert) Name() string {
	return "canon-expert"
}

// DependsOn returns the names of agents that must run before this one.
// The canon expert has no dependencies and can run independently.
func (e *Expert) DependsOn() []string {
	return nil
}

// Run executes canon contradiction detection against the provided
// content. It compares the new content against established campaign
// facts from the RAG context, calls the LLM, parses the structured
// response, and converts detected contradictions into
// ContentAnalysisItems. On LLM error the error is returned for the
// pipeline to handle. On parse error, the agent logs the failure and
// returns an empty slice (graceful degradation).
func (e *Expert) Run(
	ctx context.Context,
	provider llm.Provider,
	input enrichment.PipelineInput,
) ([]models.ContentAnalysisItem, error) {
	if input.Content == "" {
		return []models.ContentAnalysisItem{}, nil
	}

	// Canon checking requires established facts to compare against.
	// Without RAG context, there is nothing to check for
	// contradictions.
	if input.Context == nil || len(input.Context.CampaignResults) == 0 {
		return []models.ContentAnalysisItem{}, nil
	}

	systemPrompt := buildSystemPrompt(input.SourceScope)
	userPrompt := buildUserPrompt(input)

	resp, err := provider.Complete(ctx, llm.CompletionRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    4096,
		Temperature:  0.2,
	})
	if err != nil {
		return nil, err
	}

	parsed, err := parseCanonResponse(resp.Content)
	if err != nil {
		log.Printf(
			"canon-expert: parse error for job %d, returning empty results: %v",
			input.JobID, err,
		)
		return []models.ContentAnalysisItem{}, nil
	}

	items := convertToItems(input.JobID, parsed)
	return items, nil
}

// convertToItems transforms a parsed canon response into a slice of
// ContentAnalysisItems. Each detected contradiction becomes an item
// with a detection type derived from the contradiction category,
// and the full contradiction detail stored as JSON in
// SuggestedContent.
func convertToItems(
	jobID int64,
	resp *canonResponse,
) []models.ContentAnalysisItem {
	if resp == nil || len(resp.Contradictions) == 0 {
		return []models.ContentAnalysisItem{}
	}

	now := time.Now()
	items := make(
		[]models.ContentAnalysisItem, 0, len(resp.Contradictions),
	)

	for _, c := range resp.Contradictions {
		detail, err := json.Marshal(map[string]string{
			"contradiction_type": c.ContradictionType,
			"severity":           c.Severity,
			"established_fact":   c.EstablishedFact,
			"source":             c.Source,
			"conflicting_text":   c.ConflictingText,
			"description":        c.Description,
			"suggestion":         c.Suggestion,
		})
		if err != nil {
			log.Printf(
				"canon-expert: failed to marshal contradiction for job %d: %v",
				jobID, err,
			)
			continue
		}

		items = append(items, models.ContentAnalysisItem{
			JobID:            jobID,
			DetectionType:    contradictionTypeToDetectionType(c.ContradictionType),
			MatchedText:      c.ConflictingText,
			Resolution:       "pending",
			SuggestedContent: json.RawMessage(detail),
			Phase:            "analysis",
			CreatedAt:        now,
		})
	}

	return items
}
