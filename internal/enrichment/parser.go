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
	"encoding/json"
	"log"
	"strings"

	"github.com/antonypegg/imagineer/internal/models"
)

// enrichmentResponse represents the expected JSON output from the LLM.
type enrichmentResponse struct {
	DescriptionUpdates []models.DescriptionUpdateSuggestion `json:"descriptionUpdates"`
	LogEntries         []models.LogEntrySuggestion          `json:"logEntries"`
	Relationships      []models.RelationshipSuggestion      `json:"relationships"`
}

// parseEnrichmentResponse parses the LLM response text into structured
// enrichment suggestions. It strips markdown code fences, tolerates
// malformed output (logs and skips rather than failing), and returns
// whatever it can parse. On complete parse failure it returns an empty
// response rather than an error, providing graceful degradation.
func parseEnrichmentResponse(responseText string) (*enrichmentResponse, error) {
	cleaned := stripCodeFences(responseText)
	cleaned = strings.TrimSpace(cleaned)

	if cleaned == "" {
		log.Printf("enrichment: LLM returned empty response")
		return &enrichmentResponse{}, nil
	}

	var resp enrichmentResponse
	if err := json.Unmarshal([]byte(cleaned), &resp); err != nil {
		log.Printf(
			"enrichment: failed to parse LLM JSON response: %v (response: %.200s)",
			err, cleaned,
		)
		return &enrichmentResponse{}, nil
	}

	// Ensure arrays are never nil so downstream code can iterate safely.
	if resp.DescriptionUpdates == nil {
		resp.DescriptionUpdates = []models.DescriptionUpdateSuggestion{}
	}
	if resp.LogEntries == nil {
		resp.LogEntries = []models.LogEntrySuggestion{}
	}
	if resp.Relationships == nil {
		resp.Relationships = []models.RelationshipSuggestion{}
	}

	return &resp, nil
}

// stripCodeFences removes markdown code fences from the LLM response.
// It handles both ```json ... ``` and ``` ... ``` patterns, including
// fences with trailing language identifiers.
func stripCodeFences(text string) string {
	text = strings.TrimSpace(text)

	// Check for opening code fence with optional language tag.
	if strings.HasPrefix(text, "```") {
		// Remove the opening fence line.
		idx := strings.Index(text, "\n")
		if idx >= 0 {
			text = text[idx+1:]
		} else {
			// The entire text is just the opening fence.
			return ""
		}

		// Remove the closing fence.
		if strings.HasSuffix(strings.TrimSpace(text), "```") {
			text = strings.TrimSpace(text)
			text = text[:len(text)-3]
		}
	}

	return strings.TrimSpace(text)
}
