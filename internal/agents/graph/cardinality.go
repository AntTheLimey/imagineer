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
// in the cardinality_constraints table. Persisted data is checked
// by the database function check_cardinality_violations(). If
// suggestions are present, Go-side logic adds proposed counts and
// checks whether any new violations would be created.
func CheckCardinality(
	ctx context.Context,
	db *database.DB,
	campaignID int64,
	suggestions []models.ContentAnalysisItem,
) ([]CardinalityViolation, error) {
	if db == nil {
		return nil, nil
	}

	// Query persisted violations from the database function.
	dbViolations, err := queryPersistedViolations(ctx, db, campaignID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to call check_cardinality_violations: %w", err,
		)
	}

	// If there are no suggestions, the DB results are the final answer.
	if len(suggestions) == 0 {
		return dbViolations, nil
	}

	// There are suggestions, so we need to check whether proposed
	// relationships would create additional violations. Query limits
	// from the view.
	limits, err := queryCardinalityLimits(ctx, db, campaignID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to query cardinality constraints: %w", err,
		)
	}

	if len(limits) == 0 {
		// No constraints defined; DB violations should also be empty.
		return dbViolations, nil
	}

	// Build a lookup of limits by relationship type name.
	limitByType := make(map[string]cardinalityLimit, len(limits))
	for _, l := range limits {
		limitByType[l.RelationshipType] = l
	}

	// Build a set of already-reported violations from the DB so we
	// don't duplicate them.
	type violationKey struct {
		EntityID         int64
		RelationshipType string
		Direction        string
	}
	dbViolationSet := make(map[violationKey]bool, len(dbViolations))
	for _, v := range dbViolations {
		dbViolationSet[violationKey{
			EntityID:         v.EntityID,
			RelationshipType: v.RelationshipType,
			Direction:        v.Direction,
		}] = true
	}

	// Count proposed suggestion relationships per
	// (entity, relationship_type, direction).
	type countKey struct {
		EntityID         int64
		RelationshipType string
		Direction        string
	}
	proposalCounts := make(map[countKey]int)

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

		proposalCounts[countKey{rs.SourceEntityID, typeName, "source"}]++
		proposalCounts[countKey{rs.TargetEntityID, typeName, "target"}]++
	}

	if len(proposalCounts) == 0 {
		return dbViolations, nil
	}

	// For each proposed count, query the current persisted count and
	// check if adding proposals would exceed the limit. We need to
	// query entity names/types for new violations. Use a single query
	// to get all current counts for the relevant entities.
	// For simplicity, check each proposal key against limits.
	violations := append([]CardinalityViolation{}, dbViolations...)

	for key, proposalCount := range proposalCounts {
		limit := limitByType[key.RelationshipType]

		var maxAllowed *int
		if key.Direction == "source" {
			maxAllowed = limit.MaxSource
		} else {
			maxAllowed = limit.MaxTarget
		}

		if maxAllowed == nil {
			continue
		}

		// Skip if the DB already reported this as a violation.
		vk := violationKey(key)
		if dbViolationSet[vk] {
			continue
		}

		// Query the current persisted count for this specific
		// (entity, type, direction) combination.
		currentCount, err := queryEntityRelCount(
			ctx, db, campaignID,
			key.EntityID, key.RelationshipType, key.Direction,
		)
		if err != nil {
			log.Printf(
				"graph-expert: failed to query entity rel count "+
					"(entity=%d, type=%s, dir=%s): %v",
				key.EntityID, key.RelationshipType,
				key.Direction, err,
			)
			continue
		}

		totalCount := currentCount + proposalCount
		if totalCount > *maxAllowed {
			// Query entity name/type for the violation report.
			entityName, entityType := queryEntityInfo(
				ctx, db, key.EntityID,
			)

			violations = append(violations, CardinalityViolation{
				EntityID:         key.EntityID,
				EntityName:       entityName,
				EntityType:       entityType,
				RelationshipType: key.RelationshipType,
				Direction:        key.Direction,
				CurrentCount:     totalCount,
				MaxAllowed:       *maxAllowed,
			})
		}
	}

	return violations, nil
}

// queryPersistedViolations calls the database function
// check_cardinality_violations() to get all violations from
// persisted data.
func queryPersistedViolations(
	ctx context.Context,
	db *database.DB,
	campaignID int64,
) ([]CardinalityViolation, error) {
	query := `SELECT * FROM check_cardinality_violations($1)`

	rows, err := db.Query(ctx, query, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var violations []CardinalityViolation
	for rows.Next() {
		var v CardinalityViolation
		if err := rows.Scan(
			&v.EntityID, &v.EntityName, &v.EntityType,
			&v.RelationshipType, &v.Direction,
			&v.CurrentCount, &v.MaxAllowed,
		); err != nil {
			return nil, fmt.Errorf(
				"failed to scan cardinality violation row: %w", err,
			)
		}
		violations = append(violations, v)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"error iterating cardinality violation rows: %w", err,
		)
	}

	return violations, nil
}

// queryCardinalityLimits retrieves cardinality constraints for
// a campaign using the cardinality_constraints_with_names view.
func queryCardinalityLimits(
	ctx context.Context,
	db *database.DB,
	campaignID int64,
) ([]cardinalityLimit, error) {
	query := `
		SELECT relationship_type_name, max_source, max_target
		FROM cardinality_constraints_with_names
		WHERE campaign_id = $1`

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

// queryEntityRelCount queries the current persisted count of
// relationships for a given entity, relationship type, and
// direction within a campaign.
func queryEntityRelCount(
	ctx context.Context,
	db *database.DB,
	campaignID int64,
	entityID int64,
	relTypeName string,
	direction string,
) (int, error) {
	var query string
	if direction == "source" {
		query = `
			SELECT COUNT(*)
			FROM relationships r
			JOIN relationship_types rt ON r.relationship_type_id = rt.id
			WHERE r.source_entity_id = $1
			  AND rt.name = $2
			  AND r.campaign_id = $3`
	} else {
		query = `
			SELECT COUNT(*)
			FROM relationships r
			JOIN relationship_types rt ON r.relationship_type_id = rt.id
			WHERE r.target_entity_id = $1
			  AND rt.name = $2
			  AND r.campaign_id = $3`
	}

	var count int
	err := db.QueryRow(ctx, query, entityID, relTypeName, campaignID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf(
			"failed to count entity relationships: %w", err,
		)
	}

	return count, nil
}

// queryEntityInfo retrieves the name and type for an entity by ID.
// Returns fallback values if the query fails.
func queryEntityInfo(
	ctx context.Context,
	db *database.DB,
	entityID int64,
) (string, string) {
	query := `SELECT name, entity_type FROM entities WHERE id = $1`

	var name, entityType string
	err := db.QueryRow(ctx, query, entityID).Scan(&name, &entityType)
	if err != nil {
		return fmt.Sprintf("entity-%d", entityID), "unknown"
	}

	return name, entityType
}
