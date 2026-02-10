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
	"time"

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

// newEntityResponse represents the expected JSON output from the LLM
// for new-entity detection.
type newEntityResponse struct {
	NewEntities []newEntitySuggestion `json:"new_entities"`
}

// newEntitySuggestion represents a single entity suggestion from the
// new-entity detection LLM call.
type newEntitySuggestion struct {
	Name        string `json:"name"`
	EntityType  string `json:"entity_type"`
	Description string `json:"description"`
	Reasoning   string `json:"reasoning"`
}

// parseNewEntityResponse parses the LLM response text into structured
// new-entity suggestions. It follows the same graceful-degradation pattern
// as parseEnrichmentResponse: strip code fences, unmarshal JSON, and
// return an empty response on failure rather than propagating errors.
func parseNewEntityResponse(responseText string) (*newEntityResponse, error) {
	cleaned := stripCodeFences(responseText)
	cleaned = strings.TrimSpace(cleaned)

	if cleaned == "" {
		log.Printf("enrichment: LLM returned empty response for new-entity detection")
		return &newEntityResponse{}, nil
	}

	var resp newEntityResponse
	if err := json.Unmarshal([]byte(cleaned), &resp); err != nil {
		log.Printf(
			"enrichment: failed to parse new-entity JSON response: %v (response: %.200s)",
			err, cleaned,
		)
		return &newEntityResponse{}, nil
	}

	if resp.NewEntities == nil {
		resp.NewEntities = []newEntitySuggestion{}
	}

	return &resp, nil
}

// convertNewEntitiesToItems transforms a parsed new-entity response into
// a slice of ContentAnalysisItems with detection_type="new_entity_suggestion".
func convertNewEntitiesToItems(
	campaignID int64,
	jobID int64,
	resp *newEntityResponse,
) []models.ContentAnalysisItem {
	if resp == nil {
		return []models.ContentAnalysisItem{}
	}

	now := time.Now()
	items := []models.ContentAnalysisItem{}

	allowedTypes := map[string]bool{
		"npc": true, "location": true, "item": true,
		"faction": true, "clue": true, "creature": true,
		"organization": true, "event": true, "document": true,
		"other": true,
	}

	for _, suggestion := range resp.NewEntities {
		name := strings.TrimSpace(suggestion.Name)
		if name == "" {
			continue
		}

		entityType := strings.ToLower(strings.TrimSpace(suggestion.EntityType))
		if !allowedTypes[entityType] {
			entityType = "other"
		}

		content, err := json.Marshal(map[string]string{
			"entity_type": entityType,
			"description": suggestion.Description,
			"reasoning":   suggestion.Reasoning,
		})
		if err != nil {
			log.Printf(
				"enrichment: failed to marshal new-entity suggestion %q: %v",
				suggestion.Name, err,
			)
			continue
		}

		items = append(items, models.ContentAnalysisItem{
			JobID:            jobID,
			DetectionType:    "new_entity_suggestion",
			MatchedText:      name,
			Resolution:       "pending",
			SuggestedContent: json.RawMessage(content),
			Phase:            "enrichment",
			CreatedAt:        now,
		})
	}

	return items
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
