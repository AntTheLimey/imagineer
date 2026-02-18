/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

// Package canon provides the canon expert agent for detecting
// contradictions between new content and established campaign facts
// within the enrichment pipeline.
package canon

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// canonResponse represents the expected JSON output from the LLM.
type canonResponse struct {
	Contradictions []contradiction `json:"contradictions"`
}

// contradiction represents a single detected contradiction between
// new content and established campaign facts.
type contradiction struct {
	ContradictionType string `json:"contradictionType"`
	Severity          string `json:"severity"`
	ConflictingText   string `json:"conflictingText"`
	EstablishedFact   string `json:"establishedFact"`
	Source            string `json:"source"`
	Description       string `json:"description"`
	Suggestion        string `json:"suggestion"`
}

// validContradictionTypes lists the accepted contradiction type values.
var validContradictionTypes = map[string]bool{
	"factual":   true,
	"temporal":  true,
	"character": true,
}

// validSeverities lists the accepted severity levels.
var validSeverities = map[string]bool{
	"info":    true,
	"warning": true,
	"error":   true,
}

// parseCanonResponse parses the LLM response text into structured
// canon analysis results. It strips markdown code fences, unmarshals
// JSON, and validates the structure. On parse failure it returns nil
// with the error, allowing the caller to degrade gracefully.
func parseCanonResponse(raw string) (*canonResponse, error) {
	cleaned := stripCodeFences(raw)
	cleaned = strings.TrimSpace(cleaned)

	if cleaned == "" {
		log.Printf("canon-expert: LLM returned empty response")
		return nil, fmt.Errorf("empty response from LLM")
	}

	var resp canonResponse
	if err := json.Unmarshal([]byte(cleaned), &resp); err != nil {
		log.Printf(
			"canon-expert: failed to parse LLM JSON response: %v (response: %.200s)",
			err, cleaned,
		)
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Ensure contradictions is never nil so downstream code can
	// iterate safely.
	if resp.Contradictions == nil {
		resp.Contradictions = []contradiction{}
	}

	// Validate and normalise each contradiction.
	validated := make([]contradiction, 0, len(resp.Contradictions))
	for _, c := range resp.Contradictions {
		// Skip contradictions with empty description or conflicting text.
		if strings.TrimSpace(c.Description) == "" {
			log.Printf(
				"canon-expert: skipping contradiction with empty description",
			)
			continue
		}
		if strings.TrimSpace(c.ConflictingText) == "" {
			log.Printf(
				"canon-expert: skipping contradiction with empty conflictingText",
			)
			continue
		}

		contradictionType := strings.ToLower(strings.TrimSpace(c.ContradictionType))
		if !validContradictionTypes[contradictionType] {
			log.Printf(
				"canon-expert: normalising unknown contradiction type %q to \"factual\"",
				c.ContradictionType,
			)
			contradictionType = "factual"
		}

		severity := strings.ToLower(strings.TrimSpace(c.Severity))
		if !validSeverities[severity] {
			severity = "warning"
		}

		validated = append(validated, contradiction{
			ContradictionType: contradictionType,
			Severity:          severity,
			ConflictingText:   strings.TrimSpace(c.ConflictingText),
			EstablishedFact:   strings.TrimSpace(c.EstablishedFact),
			Source:            strings.TrimSpace(c.Source),
			Description:       strings.TrimSpace(c.Description),
			Suggestion:        strings.TrimSpace(c.Suggestion),
		})
	}
	resp.Contradictions = validated

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

// contradictionTypeToDetectionType maps a contradiction type to the
// corresponding ContentAnalysisItem detection type.
func contradictionTypeToDetectionType(ct string) string {
	switch ct {
	case "factual", "attribute":
		return "canon_contradiction"
	case "temporal", "timeline":
		return "temporal_inconsistency"
	case "character", "behavior":
		return "character_inconsistency"
	default:
		return "canon_contradiction"
	}
}
