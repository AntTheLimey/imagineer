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

	"github.com/antonypegg/imagineer/internal/agents"
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
	"pacing":           true,
	"investigation":    true,
	"spotlight":        true,
	"npc_development":  true,
	"mechanics":        true,
	"pc_agency":        true,
	"continuity":       true,
	"setting":          true,
	"scenario_writing": true,
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
	cleaned := agents.StripCodeFences(raw)
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
	case "spotlight", "npc_development", "pc_agency", "continuity", "setting", "scenario_writing":
		return "content_suggestion"
	default:
		return "content_suggestion"
	}
}
