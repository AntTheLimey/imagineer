/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

// Package ttrpg provides the TTRPG expert agent for content quality
// analysis within the enrichment pipeline.
package ttrpg

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// expertResponse represents the expected JSON output from the LLM.
type expertResponse struct {
	Report   string    `json:"report"`
	Findings []finding `json:"findings"`
}

// finding represents a single quality finding from the TTRPG expert
// analysis.
type finding struct {
	Category      string `json:"category"`
	Severity      string `json:"severity"`
	Description   string `json:"description"`
	Suggestion    string `json:"suggestion"`
	LineReference string `json:"lineReference,omitempty"`
}

// validCategories lists the accepted finding categories.
var validCategories = map[string]bool{
	"pacing":          true,
	"investigation":   true,
	"spotlight":       true,
	"npc_development": true,
	"mechanics":       true,
	"pc_agency":       true,
	"continuity":      true,
	"setting":         true,
}

// validSeverities lists the accepted severity levels.
var validSeverities = map[string]bool{
	"info":    true,
	"warning": true,
	"error":   true,
}

// parseExpertResponse parses the LLM response text into structured
// expert analysis results. It strips markdown code fences, unmarshals
// JSON, and validates the structure. On parse failure it logs the error
// and returns nil with the error, allowing the caller to degrade
// gracefully.
func parseExpertResponse(raw string) (*expertResponse, error) {
	cleaned := stripCodeFences(raw)
	cleaned = strings.TrimSpace(cleaned)

	if cleaned == "" {
		log.Printf("ttrpg-expert: LLM returned empty response")
		return nil, fmt.Errorf("empty response from LLM")
	}

	var resp expertResponse
	if err := json.Unmarshal([]byte(cleaned), &resp); err != nil {
		log.Printf(
			"ttrpg-expert: failed to parse LLM JSON response: %v (response: %.200s)",
			err, cleaned,
		)
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Ensure findings is never nil so downstream code can iterate
	// safely.
	if resp.Findings == nil {
		resp.Findings = []finding{}
	}

	// Validate and normalise each finding.
	validated := make([]finding, 0, len(resp.Findings))
	for _, f := range resp.Findings {
		category := strings.ToLower(strings.TrimSpace(f.Category))
		if !validCategories[category] {
			log.Printf(
				"ttrpg-expert: skipping finding with invalid category %q",
				f.Category,
			)
			continue
		}

		severity := strings.ToLower(strings.TrimSpace(f.Severity))
		if !validSeverities[severity] {
			severity = "info"
		}

		if strings.TrimSpace(f.Description) == "" {
			log.Printf(
				"ttrpg-expert: skipping finding with empty description (category=%q)",
				category,
			)
			continue
		}

		validated = append(validated, finding{
			Category:      category,
			Severity:      severity,
			Description:   strings.TrimSpace(f.Description),
			Suggestion:    strings.TrimSpace(f.Suggestion),
			LineReference: strings.TrimSpace(f.LineReference),
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

// categoryToDetectionType maps a finding category to the corresponding
// ContentAnalysisItem detection type.
func categoryToDetectionType(category string) string {
	switch category {
	case "pacing":
		return "pacing_note"
	case "investigation":
		return "investigation_gap"
	case "mechanics":
		return "mechanics_warning"
	case "spotlight", "npc_development", "pc_agency", "continuity", "setting":
		return "content_suggestion"
	default:
		return "content_suggestion"
	}
}
