/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package ttrpg

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/antonypegg/imagineer/internal/enrichment"
	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/antonypegg/imagineer/internal/models"
)

// Expert implements the enrichment.PipelineAgent interface for TTRPG
// content quality analysis. It evaluates campaign content across
// multiple dimensions including pacing, investigation design, spotlight
// balance, NPC development, mechanics accuracy, player agency,
// continuity, and setting atmosphere.
type Expert struct{}

// NewExpert creates a new TTRPG expert agent.
func NewExpert() *Expert {
	return &Expert{}
}

// Name returns the unique identifier for this pipeline agent.
func (e *Expert) Name() string {
	return "ttrpg-expert"
}

// DependsOn returns the names of agents that must run before this one.
// The TTRPG expert has no dependencies and can run independently.
func (e *Expert) DependsOn() []string {
	return nil
}

// Run executes TTRPG content quality analysis against the provided
// content. It builds a system and user prompt, calls the LLM, parses
// the structured response, and converts findings into
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

	systemPrompt := buildSystemPrompt()
	userPrompt := buildUserPrompt(input)

	resp, err := provider.Complete(ctx, llm.CompletionRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    4096,
		Temperature:  0.3,
	})
	if err != nil {
		return nil, err
	}

	parsed, err := parseExpertResponse(resp.Content)
	if err != nil {
		log.Printf(
			"ttrpg-expert: parse error for job %d, returning empty results: %v",
			input.JobID, err,
		)
		return []models.ContentAnalysisItem{}, nil
	}

	items := convertToItems(input.JobID, parsed)
	return items, nil
}

// convertToItems transforms a parsed expert response into a slice of
// ContentAnalysisItems. It creates one analysis_report item containing
// the full markdown report, plus individual items for each finding
// with detection types mapped from the finding category.
func convertToItems(jobID int64, resp *expertResponse) []models.ContentAnalysisItem {
	if resp == nil {
		return []models.ContentAnalysisItem{}
	}

	now := time.Now()
	items := make([]models.ContentAnalysisItem, 0, 1+len(resp.Findings))

	// Create the overall analysis report item.
	if resp.Report != "" {
		reportContent, err := json.Marshal(map[string]string{
			"report": resp.Report,
		})
		if err != nil {
			log.Printf(
				"ttrpg-expert: failed to marshal report for job %d: %v",
				jobID, err,
			)
		} else {
			items = append(items, models.ContentAnalysisItem{
				JobID:            jobID,
				DetectionType:    "analysis_report",
				MatchedText:      "ttrpg-expert",
				Resolution:       "pending",
				SuggestedContent: json.RawMessage(reportContent),
				Phase:            "analysis",
				CreatedAt:        now,
			})
		}
	}

	// Create individual items for each finding.
	for _, f := range resp.Findings {
		findingContent, err := json.Marshal(map[string]string{
			"category":      f.Category,
			"severity":      f.Severity,
			"description":   f.Description,
			"suggestion":    f.Suggestion,
			"lineReference": f.LineReference,
		})
		if err != nil {
			log.Printf(
				"ttrpg-expert: failed to marshal finding %q for job %d: %v",
				f.Category, jobID, err,
			)
			continue
		}

		items = append(items, models.ContentAnalysisItem{
			JobID:            jobID,
			DetectionType:    categoryToDetectionType(f.Category),
			MatchedText:      f.Category,
			Resolution:       "pending",
			SuggestedContent: json.RawMessage(findingContent),
			Phase:            "analysis",
			CreatedAt:        now,
		})
	}

	return items
}
