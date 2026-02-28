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

// CardinalityViolation represents a case where an entity exceeds
// (or would exceed) the maximum number of relationships of a given
// type in a particular direction.
type CardinalityViolation struct {
	EntityID         int64  `json:"entityId"`
	EntityName       string `json:"entityName"`
	EntityType       string `json:"entityType"`
	RelationshipType string `json:"relationshipType"`
	Direction        string `json:"direction"` // "source" or "target"
	CurrentCount     int    `json:"currentCount"`
	MaxAllowed       int    `json:"maxAllowed"`
}

// cardinalityLimit holds the parsed limits from a single
// cardinality_constraints row joined with the relationship type name.
type cardinalityLimit struct {
	RelationshipType string
	MaxSource        *int
	MaxTarget        *int
}

// CheckCardinality validates that existing relationships plus
// proposed new suggestions don't exceed cardinality limits defined
// in the cardinality_constraints table. If no constraints exist for
// the campaign, it returns nil early.
func CheckCardinality(
	ctx context.Context,
	db *database.DB,
	campaignID int64,
	suggestions []models.ContentAnalysisItem,
	relationships []models.Relationship,
	entities []models.Entity,
) ([]CardinalityViolation, error) {
	if db == nil {
		return nil, nil
	}

	// Query cardinality constraints for this campaign.
	limits, err := queryCardinalityLimits(ctx, db, campaignID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to query cardinality constraints: %w", err,
		)
	}

	if len(limits) == 0 {
		return nil, nil
	}

	// Build a lookup of limits by relationship type name.
	limitByType := make(map[string]cardinalityLimit, len(limits))
	for _, l := range limits {
		limitByType[l.RelationshipType] = l
	}

	// Build entity lookup for name/type resolution.
	entityByID := make(map[int64]models.Entity, len(entities))
	for _, e := range entities {
		entityByID[e.ID] = e
	}

	// Count existing relationships per (entity, relationship_type, direction).
	// Key format: "entityID:relationshipType:direction"
	counts := make(map[string]int)
	for _, r := range relationships {
		typeName := r.RelationshipTypeName
		if typeName == "" {
			continue
		}
		if _, hasLimit := limitByType[typeName]; !hasLimit {
			continue
		}
		sourceKey := fmt.Sprintf("%d:%s:source", r.SourceEntityID, typeName)
		targetKey := fmt.Sprintf("%d:%s:target", r.TargetEntityID, typeName)
		counts[sourceKey]++
		counts[targetKey]++
	}

	// Parse suggestions and add proposed counts.
	for _, item := range suggestions {
		if item.DetectionType != "relationship_suggestion" {
			continue
		}
		if len(item.SuggestedContent) == 0 {
			continue
		}

		var rs models.RelationshipSuggestion
		if err := json.Unmarshal(item.SuggestedContent, &rs); err != nil {
			log.Printf(
				"graph-expert: failed to unmarshal suggestion for cardinality check: %v",
				err,
			)
			continue
		}

		typeName := rs.RelationshipType
		if typeName == "" {
			continue
		}
		if _, hasLimit := limitByType[typeName]; !hasLimit {
			continue
		}

		sourceKey := fmt.Sprintf("%d:%s:source", rs.SourceEntityID, typeName)
		targetKey := fmt.Sprintf("%d:%s:target", rs.TargetEntityID, typeName)
		counts[sourceKey]++
		counts[targetKey]++
	}

	// Compare totals against limits and collect violations.
	var violations []CardinalityViolation
	// Use a set to avoid duplicate violations for the same
	// (entity, type, direction) combination.
	seen := make(map[string]bool)

	for key, count := range counts {
		var entityID int64
		var typeName, direction string
		_, err := fmt.Sscanf(key, "%d:%s", &entityID, &typeName)
		if err != nil {
			// Parse the key manually since Sscanf doesn't handle
			// colons within the type name well.
			continue
		}

		// Parse key manually for robustness.
		entityID, typeName, direction = parseCountKey(key)
		if typeName == "" {
			continue
		}

		limit, ok := limitByType[typeName]
		if !ok {
			continue
		}

		var maxAllowed *int
		if direction == "source" {
			maxAllowed = limit.MaxSource
		} else {
			maxAllowed = limit.MaxTarget
		}

		if maxAllowed == nil {
			continue
		}

		if count > *maxAllowed {
			violationKey := fmt.Sprintf(
				"%d:%s:%s", entityID, typeName, direction,
			)
			if seen[violationKey] {
				continue
			}
			seen[violationKey] = true

			entity, ok := entityByID[entityID]
			entityName := fmt.Sprintf("entity-%d", entityID)
			entityType := "unknown"
			if ok {
				entityName = entity.Name
				entityType = string(entity.EntityType)
			}

			violations = append(violations, CardinalityViolation{
				EntityID:         entityID,
				EntityName:       entityName,
				EntityType:       entityType,
				RelationshipType: typeName,
				Direction:        direction,
				CurrentCount:     count,
				MaxAllowed:       *maxAllowed,
			})
		}
	}

	return violations, nil
}

// parseCountKey extracts (entityID, typeName, direction) from a
// count map key of the form "entityID:typeName:direction".
func parseCountKey(key string) (int64, string, string) {
	// Find the first colon to extract entityID.
	firstColon := -1
	for i, ch := range key {
		if ch == ':' {
			firstColon = i
			break
		}
	}
	if firstColon < 0 || firstColon >= len(key)-1 {
		return 0, "", ""
	}

	var entityID int64
	_, err := fmt.Sscanf(key[:firstColon], "%d", &entityID)
	if err != nil {
		return 0, "", ""
	}

	// Find the last colon to extract direction.
	lastColon := -1
	for i := len(key) - 1; i > firstColon; i-- {
		if key[i] == ':' {
			lastColon = i
			break
		}
	}
	if lastColon < 0 || lastColon >= len(key)-1 {
		return 0, "", ""
	}

	typeName := key[firstColon+1 : lastColon]
	direction := key[lastColon+1:]

	return entityID, typeName, direction
}

// queryCardinalityLimits retrieves cardinality constraints for
// a campaign by joining with relationship_types to resolve
// type names.
func queryCardinalityLimits(
	ctx context.Context,
	db *database.DB,
	campaignID int64,
) ([]cardinalityLimit, error) {
	query := `
		SELECT rt.name, cc.max_source, cc.max_target
		FROM cardinality_constraints cc
		JOIN relationship_types rt ON cc.relationship_type_id = rt.id
		WHERE cc.campaign_id = $1`

	rows, err := db.Query(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to query cardinality constraints: %w", err,
		)
	}
	defer rows.Close()

	var limits []cardinalityLimit
	for rows.Next() {
		var l cardinalityLimit
		if err := rows.Scan(
			&l.RelationshipType, &l.MaxSource, &l.MaxTarget,
		); err != nil {
			return nil, fmt.Errorf(
				"failed to scan cardinality constraint row: %w", err,
			)
		}
		limits = append(limits, l)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"error iterating cardinality constraint rows: %w", err,
		)
	}

	return limits, nil
}
