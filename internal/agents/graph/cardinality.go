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

	// --- Batch 1: Entity info lookup ---
	// Collect all unique entity IDs from proposal counts.
	entityIDSet := make(map[int64]bool)
	for key := range proposalCounts {
		entityIDSet[key.EntityID] = true
	}

	entityIDs := make([]int64, 0, len(entityIDSet))
	for id := range entityIDSet {
		entityIDs = append(entityIDs, id)
	}

	type entityInfo struct {
		Name       string
		EntityType string
	}
	entityInfoMap := make(map[int64]entityInfo, len(entityIDs))

	entityInfoRows, err := db.Query(ctx,
		`SELECT id, name, entity_type FROM entities WHERE id = ANY($1)`,
		entityIDs,
	)
	if err != nil {
		log.Printf(
			"graph-expert: failed to batch-query entity info: %v", err,
		)
		// Populate fallback values so we can still report violations.
		for _, id := range entityIDs {
			entityInfoMap[id] = entityInfo{
				Name:       fmt.Sprintf("entity-%d", id),
				EntityType: "unknown",
			}
		}
	} else {
		defer entityInfoRows.Close()
		for entityInfoRows.Next() {
			var id int64
			var info entityInfo
			if err := entityInfoRows.Scan(
				&id, &info.Name, &info.EntityType,
			); err != nil {
				log.Printf(
					"graph-expert: failed to scan entity info row: %v",
					err,
				)
				continue
			}
			entityInfoMap[id] = info
		}
		if err := entityInfoRows.Err(); err != nil {
			log.Printf(
				"graph-expert: error iterating entity info rows: %v",
				err,
			)
		}
		entityInfoRows.Close()
	}

	// Fill in fallback values for any entities not found.
	for _, id := range entityIDs {
		if _, ok := entityInfoMap[id]; !ok {
			entityInfoMap[id] = entityInfo{
				Name:       fmt.Sprintf("entity-%d", id),
				EntityType: "unknown",
			}
		}
	}

	// --- Batch 2: Relationship type ID lookup ---
	// Collect all unique relationship type names from proposal counts.
	relTypeNameSet := make(map[string]bool)
	for key := range proposalCounts {
		relTypeNameSet[key.RelationshipType] = true
	}

	relTypeNames := make([]string, 0, len(relTypeNameSet))
	for name := range relTypeNameSet {
		relTypeNames = append(relTypeNames, name)
	}

	// Map relationship type name -> ID for the campaign.
	relTypeIDByName := make(map[string]int64, len(relTypeNames))
	relTypeIDs := make([]int64, 0, len(relTypeNames))

	relTypeRows, err := db.Query(ctx,
		`SELECT id, name FROM relationship_types
		 WHERE name = ANY($1)
		   AND (campaign_id = $2 OR campaign_id IS NULL)`,
		relTypeNames, campaignID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to batch-query relationship type IDs: %w", err,
		)
	}
	defer relTypeRows.Close()

	for relTypeRows.Next() {
		var id int64
		var name string
		if err := relTypeRows.Scan(&id, &name); err != nil {
			return nil, fmt.Errorf(
				"failed to scan relationship type row: %w", err,
			)
		}
		relTypeIDByName[name] = id
		relTypeIDs = append(relTypeIDs, id)
	}
	if err := relTypeRows.Err(); err != nil {
		return nil, fmt.Errorf(
			"error iterating relationship type rows: %w", err,
		)
	}
	relTypeRows.Close()

	// --- Batch 3: Relationship count lookup ---
	// Count existing relationships per (entity, type_id, direction)
	// for all relevant combinations in a single query.
	type relCountKey struct {
		EntityID           int64
		RelationshipTypeID int64
		Direction          string
	}
	existingCounts := make(map[relCountKey]int)

	if len(relTypeIDs) > 0 {
		countRows, err := db.Query(ctx,
			`SELECT entity_id, relationship_type_id, direction, cnt
			 FROM (
			     SELECT source_entity_id AS entity_id,
			            relationship_type_id,
			            'source' AS direction,
			            COUNT(*) AS cnt
			     FROM relationships
			     WHERE campaign_id = $1
			       AND relationship_type_id = ANY($2)
			       AND source_entity_id = ANY($3)
			     GROUP BY source_entity_id, relationship_type_id
			     UNION ALL
			     SELECT target_entity_id AS entity_id,
			            relationship_type_id,
			            'target' AS direction,
			            COUNT(*) AS cnt
			     FROM relationships
			     WHERE campaign_id = $1
			       AND relationship_type_id = ANY($2)
			       AND target_entity_id = ANY($3)
			     GROUP BY target_entity_id, relationship_type_id
			 ) sub`,
			campaignID, relTypeIDs, entityIDs,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to batch-query relationship counts: %w", err,
			)
		}
		defer countRows.Close()

		for countRows.Next() {
			var eid int64
			var rtid int64
			var dir string
			var cnt int
			if err := countRows.Scan(
				&eid, &rtid, &dir, &cnt,
			); err != nil {
				return nil, fmt.Errorf(
					"failed to scan relationship count row: %w", err,
				)
			}
			existingCounts[relCountKey{eid, rtid, dir}] = cnt
		}
		if err := countRows.Err(); err != nil {
			return nil, fmt.Errorf(
				"error iterating relationship count rows: %w", err,
			)
		}
		countRows.Close()
	}

	// Check each proposal against the limits using the batched data.
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

		// Look up the current persisted count from the batched
		// results using the relationship type ID.
		rtID, ok := relTypeIDByName[key.RelationshipType]
		if !ok {
			// Relationship type not found; skip.
			continue
		}

		currentCount := existingCounts[relCountKey{
			key.EntityID, rtID, key.Direction,
		}]

		totalCount := currentCount + proposalCount
		if totalCount > *maxAllowed {
			info := entityInfoMap[key.EntityID]
			violations = append(violations, CardinalityViolation{
				EntityID:         key.EntityID,
				EntityName:       info.Name,
				EntityType:       info.EntityType,
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
