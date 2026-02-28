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
	"github.com/antonypegg/imagineer/internal/models"
)

// RequiredViolation represents an entity that is missing a relationship
// type that the campaign's ontology declares as required.
type RequiredViolation struct {
	EntityID                int64  `json:"entityId"`
	EntityName              string `json:"entityName"`
	EntityType              string `json:"entityType"`
	MissingRelationshipType string `json:"missingRelationshipType"`
}

// requiredRule holds a single row from the required_relationships table.
type requiredRule struct {
	EntityType           string
	RelationshipTypeName string
}

// CheckRequiredRelationships verifies that every entity of a given type
// participates in at least one relationship of each type that the
// campaign's required_relationships table mandates. For example, if a
// rule says entity_type="npc" and relationship_type_name="located_at",
// then every NPC must have at least one located_at relationship (as
// source or target).
//
// Returns nil when db is nil, when no rules exist for the campaign, or
// when all entities satisfy the rules.
func CheckRequiredRelationships(
	ctx context.Context,
	db *database.DB,
	campaignID int64,
	entities []models.Entity,
	relationships []models.Relationship,
) ([]RequiredViolation, error) {
	if db == nil {
		return nil, nil
	}

	// Query required_relationships rules for this campaign.
	rules, err := queryRequiredRules(ctx, db, campaignID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to query required relationships: %w", err,
		)
	}

	if len(rules) == 0 {
		return nil, nil
	}

	// Collect unique relationship type names referenced by the rules.
	ruleTypeNames := make(map[string]bool, len(rules))
	for _, r := range rules {
		ruleTypeNames[r.RelationshipTypeName] = true
	}

	// Query relationship_types to build a name -> id map so we can
	// match against Relationship.RelationshipTypeID.
	nameToID, err := queryRelTypeNameToID(
		ctx, db, campaignID, ruleTypeNames,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to query relationship type IDs: %w", err,
		)
	}

	// Build a set of (entityID, relationshipTypeID) pairs that exist
	// in the current relationship set. An entity satisfies a required
	// relationship if it appears as source or target in a relationship
	// whose type ID matches.
	type entityRelKey struct {
		EntityID int64
		TypeID   int64
	}
	has := make(map[entityRelKey]bool)
	for _, rel := range relationships {
		has[entityRelKey{rel.SourceEntityID, rel.RelationshipTypeID}] = true
		has[entityRelKey{rel.TargetEntityID, rel.RelationshipTypeID}] = true
	}

	// Group rules by entity type for efficient lookup.
	rulesByEntityType := make(map[string][]string)
	for _, r := range rules {
		rulesByEntityType[r.EntityType] = append(
			rulesByEntityType[r.EntityType],
			r.RelationshipTypeName,
		)
	}

	// Check each entity against applicable rules.
	var violations []RequiredViolation
	for _, entity := range entities {
		entityType := string(entity.EntityType)
		requiredTypes, ok := rulesByEntityType[entityType]
		if !ok {
			continue
		}

		for _, rtName := range requiredTypes {
			rtID, idOK := nameToID[rtName]
			if !idOK {
				// Relationship type not found in the campaign;
				// skip rather than producing a false positive.
				continue
			}

			if !has[entityRelKey{entity.ID, rtID}] {
				violations = append(violations, RequiredViolation{
					EntityID:                entity.ID,
					EntityName:              entity.Name,
					EntityType:              entityType,
					MissingRelationshipType: rtName,
				})
			}
		}
	}

	return violations, nil
}

// queryRequiredRules retrieves all required_relationships rows for a
// campaign.
func queryRequiredRules(
	ctx context.Context,
	db *database.DB,
	campaignID int64,
) ([]requiredRule, error) {
	query := `
		SELECT entity_type, relationship_type_name
		FROM required_relationships
		WHERE campaign_id = $1`

	rows, err := db.Query(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to query required_relationships: %w", err,
		)
	}
	defer rows.Close()

	var rules []requiredRule
	for rows.Next() {
		var r requiredRule
		if err := rows.Scan(&r.EntityType, &r.RelationshipTypeName); err != nil {
			return nil, fmt.Errorf(
				"failed to scan required_relationships row: %w", err,
			)
		}
		rules = append(rules, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"error iterating required_relationships rows: %w", err,
		)
	}

	return rules, nil
}

// queryRelTypeNameToID queries relationship_types for the given names
// in a campaign and returns a name -> id mapping.
func queryRelTypeNameToID(
	ctx context.Context,
	db *database.DB,
	campaignID int64,
	names map[string]bool,
) (map[string]int64, error) {
	if len(names) == 0 {
		return nil, nil
	}

	nameList := make([]string, 0, len(names))
	for n := range names {
		nameList = append(nameList, n)
	}

	query := `
		SELECT name, id
		FROM relationship_types
		WHERE name = ANY($1)
		  AND (campaign_id = $2 OR campaign_id IS NULL)`

	rows, err := db.Query(ctx, query, nameList, campaignID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to query relationship_types: %w", err,
		)
	}
	defer rows.Close()

	result := make(map[string]int64, len(nameList))
	for rows.Next() {
		var name string
		var id int64
		if err := rows.Scan(&name, &id); err != nil {
			return nil, fmt.Errorf(
				"failed to scan relationship_types row: %w", err,
			)
		}
		result[name] = id
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"error iterating relationship_types rows: %w", err,
		)
	}

	return result, nil
}
