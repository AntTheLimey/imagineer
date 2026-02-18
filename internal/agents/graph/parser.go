/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package graph

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// graphResponse represents the expected JSON output from the LLM.
type graphResponse struct {
	Findings []graphFinding `json:"findings"`
}

// graphFinding represents a single graph hygiene finding from the LLM.
type graphFinding struct {
	FindingType      string   `json:"findingType"`
	Description      string   `json:"description"`
	InvolvedEntities []string `json:"involvedEntities"`
	Suggestion       string   `json:"suggestion"`
}

// validFindingTypes lists the accepted finding type values.
var validFindingTypes = map[string]bool{
	"redundant_edge": true,
	"implied_edge":   true,
}

// parseGraphResponse parses the LLM response text into structured
// graph analysis results. It strips markdown code fences, unmarshals
// JSON, and validates the structure. On parse failure it returns nil
// with the error, allowing the caller to degrade gracefully.
func parseGraphResponse(raw string) (*graphResponse, error) {
	cleaned := stripCodeFences(raw)
	cleaned = strings.TrimSpace(cleaned)

	if cleaned == "" {
		log.Printf("graph-expert: LLM returned empty response")
		return nil, fmt.Errorf("empty response from LLM")
	}

	var resp graphResponse
	if err := json.Unmarshal([]byte(cleaned), &resp); err != nil {
		log.Printf(
			"graph-expert: failed to parse LLM JSON response: %v (response: %.200s)",
			err, cleaned,
		)
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Ensure findings is never nil so downstream code can iterate
	// safely.
	if resp.Findings == nil {
		resp.Findings = []graphFinding{}
	}

	// Validate and normalise each finding.
	validated := make([]graphFinding, 0, len(resp.Findings))
	for _, f := range resp.Findings {
		findingType := strings.ToLower(strings.TrimSpace(f.FindingType))
		if !validFindingTypes[findingType] {
			log.Printf(
				"graph-expert: normalising unknown finding type %q to \"redundant_edge\"",
				f.FindingType,
			)
			findingType = "redundant_edge"
		}

		if strings.TrimSpace(f.Description) == "" {
			log.Printf(
				"graph-expert: skipping finding with empty description (type=%q)",
				findingType,
			)
			continue
		}

		// Ensure InvolvedEntities is never nil.
		involvedEntities := f.InvolvedEntities
		if involvedEntities == nil {
			involvedEntities = []string{}
		}

		validated = append(validated, graphFinding{
			FindingType:      findingType,
			Description:      strings.TrimSpace(f.Description),
			InvolvedEntities: involvedEntities,
			Suggestion:       strings.TrimSpace(f.Suggestion),
		})
	}
	resp.Findings = validated

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

// findingTypeToDetectionType maps an LLM finding type to the
// corresponding ContentAnalysisItem detection type.
func findingTypeToDetectionType(ft string) string {
	switch ft {
	case "redundant_edge":
		return "redundant_edge"
	case "implied_edge":
		return "redundant_edge"
	default:
		return "graph_warning"
	}
}
