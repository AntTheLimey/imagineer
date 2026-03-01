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
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/models"
)

// overrideKeyInfo holds the constraint type and override key
// extracted from a ContentAnalysisItem's DetectionType and
// SuggestedContent JSON.
type overrideKeyInfo struct {
	constraintType string
	overrideKey    string
}

// FilterOverriddenFindings removes findings from the list that
// match existing constraint overrides for the campaign. Items
// with detection types that are not overridable (orphan_warning,
// redundant_edge, graph_warning) pass through unchanged.
//
// All overrides for the campaign are loaded in a single query
// and checked in-memory, replacing per-item database round-trips.
func FilterOverriddenFindings(
	ctx context.Context,
	db *database.DB,
	campaignID int64,
	items []models.ContentAnalysisItem,
) []models.ContentAnalysisItem {
	if db == nil || len(items) == 0 {
		return items
	}

	// Load all constraint overrides for the campaign in one query.
	overrideSet, err := loadOverrideSet(ctx, db, campaignID)
	if err != nil {
		log.Printf(
			"graph-expert: failed to load constraint overrides "+
				"for campaign %d, returning items unfiltered: %v",
			campaignID, err,
		)
		return items
	}

	filtered := make([]models.ContentAnalysisItem, 0, len(items))

	for _, item := range items {
		info, ok := extractOverrideKey(item)
		if !ok {
			// Not an overridable detection type, or could not
			// extract the key. Keep the item.
			filtered = append(filtered, item)
			continue
		}

		lookupKey := info.constraintType + ":" + info.overrideKey
		if overrideSet[lookupKey] {
			log.Printf(
				"graph-expert: filtering overridden finding "+
					"(type=%s, key=%s)",
				info.constraintType, info.overrideKey,
			)
			continue
		}

		filtered = append(filtered, item)
	}

	return filtered
}

// loadOverrideSet queries all constraint overrides for a campaign
// and returns a set of "constraintType:overrideKey" strings for
// fast in-memory lookup.
func loadOverrideSet(
	ctx context.Context,
	db *database.DB,
	campaignID int64,
) (map[string]bool, error) {
	query := `
		SELECT constraint_type, override_key
		FROM constraint_overrides
		WHERE campaign_id = $1`

	rows, err := db.Query(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to query constraint overrides: %w", err,
		)
	}
	defer rows.Close()

	overrides := make(map[string]bool)
	for rows.Next() {
		var constraintType, overrideKey string
		if err := rows.Scan(&constraintType, &overrideKey); err != nil {
			return nil, fmt.Errorf(
				"failed to scan constraint override row: %w", err,
			)
		}
		overrides[constraintType+":"+overrideKey] = true
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"error iterating constraint override rows: %w", err,
		)
	}

	return overrides, nil
}

// extractOverrideKey determines the constraint_type and
// override_key for an item based on its DetectionType and
// SuggestedContent. Returns false if the item is not
// overridable or the key cannot be extracted.
func extractOverrideKey(
	item models.ContentAnalysisItem,
) (overrideKeyInfo, bool) {
	switch item.DetectionType {
	case "invalid_type_pair":
		return extractTypePairKey(item.SuggestedContent)
	case "cardinality_violation":
		return extractCardinalityKey(item.SuggestedContent)
	case "missing_required":
		return extractRequiredKey(item.SuggestedContent)
	default:
		// orphan_warning, redundant_edge, graph_warning, etc.
		// are not overridable.
		return overrideKeyInfo{}, false
	}
}

// extractTypePairKey parses the SuggestedContent JSON for an
// invalid_type_pair finding and builds the override key in the
// format "{relType}:{sourceType}:{targetType}".
func extractTypePairKey(
	content json.RawMessage,
) (overrideKeyInfo, bool) {
	if len(content) == 0 {
		return overrideKeyInfo{}, false
	}

	var data struct {
		RelationshipType string `json:"relationshipType"`
		SourceEntityType string `json:"sourceEntityType"`
		TargetEntityType string `json:"targetEntityType"`
	}
	if err := json.Unmarshal(content, &data); err != nil {
		log.Printf(
			"graph-expert: failed to parse type pair "+
				"SuggestedContent for override check: %v", err,
		)
		return overrideKeyInfo{}, false
	}

	if data.RelationshipType == "" ||
		data.SourceEntityType == "" ||
		data.TargetEntityType == "" {
		return overrideKeyInfo{}, false
	}

	return overrideKeyInfo{
		constraintType: "domain_range",
		overrideKey: fmt.Sprintf(
			"%s:%s:%s",
			data.RelationshipType,
			data.SourceEntityType,
			data.TargetEntityType,
		),
	}, true
}

// extractCardinalityKey parses the SuggestedContent JSON for a
// cardinality_violation finding and builds the override key in
// the format "{relType}:{entityId}:{direction}".
func extractCardinalityKey(
	content json.RawMessage,
) (overrideKeyInfo, bool) {
	if len(content) == 0 {
		return overrideKeyInfo{}, false
	}

	var data struct {
		RelationshipType string `json:"relationshipType"`
		EntityID         int64  `json:"entityId"`
		Direction        string `json:"direction"`
	}
	if err := json.Unmarshal(content, &data); err != nil {
		log.Printf(
			"graph-expert: failed to parse cardinality "+
				"SuggestedContent for override check: %v", err,
		)
		return overrideKeyInfo{}, false
	}

	if data.RelationshipType == "" ||
		data.EntityID == 0 ||
		data.Direction == "" {
		return overrideKeyInfo{}, false
	}

	return overrideKeyInfo{
		constraintType: "cardinality",
		overrideKey: fmt.Sprintf(
			"%s:%d:%s",
			data.RelationshipType,
			data.EntityID,
			data.Direction,
		),
	}, true
}

// extractRequiredKey parses the SuggestedContent JSON for a
// missing_required finding and builds the override key in the
// format "{entityType}:{relType}".
func extractRequiredKey(
	content json.RawMessage,
) (overrideKeyInfo, bool) {
	if len(content) == 0 {
		return overrideKeyInfo{}, false
	}

	var data struct {
		EntityType              string `json:"entityType"`
		MissingRelationshipType string `json:"missingRelationshipType"`
	}
	if err := json.Unmarshal(content, &data); err != nil {
		log.Printf(
			"graph-expert: failed to parse required "+
				"SuggestedContent for override check: %v", err,
		)
		return overrideKeyInfo{}, false
	}

	if data.EntityType == "" ||
		data.MissingRelationshipType == "" {
		return overrideKeyInfo{}, false
	}

	return overrideKeyInfo{
		constraintType: "required",
		overrideKey: fmt.Sprintf(
			"%s:%s",
			data.EntityType,
			data.MissingRelationshipType,
		),
	}, true
}
