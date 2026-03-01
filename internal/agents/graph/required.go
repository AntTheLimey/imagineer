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
	"fmt"

	"github.com/antonypegg/imagineer/internal/database"
)

// RequiredViolation represents an entity that is missing a relationship
// type that the campaign's ontology declares as required.
type RequiredViolation struct {
	EntityID                int64  `json:"entityId"`
	EntityName              string `json:"entityName"`
	EntityType              string `json:"entityType"`
	MissingRelationshipType string `json:"missingRelationshipType"`
}

// CheckRequiredRelationships verifies that every entity of a given type
// participates in at least one relationship of each type that the
// campaign's required_relationships table mandates. It delegates all
// checking to the database function check_required_relationships(),
// which performs the entity/relationship cross-check in a single
// round-trip.
//
// Returns nil when db is nil or when all entities satisfy the rules.
func CheckRequiredRelationships(
	ctx context.Context,
	db *database.DB,
	campaignID int64,
) ([]RequiredViolation, error) {
	if db == nil {
		return nil, nil
	}

	query := `SELECT * FROM check_required_relationships($1)`

	rows, err := db.Query(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to call check_required_relationships: %w", err,
		)
	}
	defer rows.Close()

	var violations []RequiredViolation
	for rows.Next() {
		var v RequiredViolation
		if err := rows.Scan(
			&v.EntityID, &v.EntityName,
			&v.EntityType, &v.MissingRelationshipType,
		); err != nil {
			return nil, fmt.Errorf(
				"failed to scan required violation row: %w", err,
			)
		}
		violations = append(violations, v)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"error iterating required violation rows: %w", err,
		)
	}

	return violations, nil
}
