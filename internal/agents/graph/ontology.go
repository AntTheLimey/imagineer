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
	"strings"

	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/models"
)

// TypePairViolation represents a relationship suggestion that uses an
// invalid entity type combination according to the
// relationship_type_constraints table.
type TypePairViolation struct {
	SourceEntityID   int64
	SourceEntityName string
	SourceEntityType string
	TargetEntityID   int64
	TargetEntityName string
	TargetEntityType string
	RelationshipType string
	ValidPairs       string // human-readable description of allowed pairs
}

// CheckOrphanedEntities finds entities with no relationships at all.
// It compares the provided entity list against the relationship list
// and returns entities that appear as neither source nor target in any
// relationship.
func CheckOrphanedEntities(
	entities []models.Entity,
	relationships []models.Relationship,
) []models.Entity {
	if len(entities) == 0 {
		return nil
	}

	// Build a set of entity IDs that participate in at least one
	// relationship.
	connected := make(map[int64]bool, len(relationships)*2)
	for _, r := range relationships {
		connected[r.SourceEntityID] = true
		connected[r.TargetEntityID] = true
	}

	var orphans []models.Entity
	for _, entity := range entities {
		if !connected[entity.ID] {
			orphans = append(orphans, entity)
		}
	}

	return orphans
}

// ValidateTypePairs checks if relationship suggestions use valid
// entity type combinations according to the relationship_type_constraints
// table. It returns violations for suggestions that pair incompatible
// entity types. If no constraints exist for a relationship type, the
// suggestion is considered valid (constraints are optional).
func ValidateTypePairs(
	ctx context.Context,
	db *database.DB,
	campaignID int64,
	suggestions []models.ContentAnalysisItem,
	entities []models.Entity,
) ([]TypePairViolation, error) {
	if len(suggestions) == 0 {
		return nil, nil
	}

	// Build an entity lookup by ID for quick type resolution.
	entityByID := make(map[int64]models.Entity, len(entities))
	for _, e := range entities {
		entityByID[e.ID] = e
	}

	// Parse each suggestion to extract relationship details.
	type parsedSuggestion struct {
		suggestion models.RelationshipSuggestion
		item       models.ContentAnalysisItem
	}

	var parsed []parsedSuggestion
	for _, item := range suggestions {
		if len(item.SuggestedContent) == 0 {
			continue
		}

		var rs models.RelationshipSuggestion
		if err := json.Unmarshal(item.SuggestedContent, &rs); err != nil {
			log.Printf(
				"graph-expert: failed to unmarshal relationship suggestion: %v",
				err,
			)
			continue
		}

		parsed = append(parsed, parsedSuggestion{
			suggestion: rs,
			item:       item,
		})
	}

	if len(parsed) == 0 {
		return nil, nil
	}

	// Collect unique relationship type names for a single DB query.
	typeNames := make(map[string]bool, len(parsed))
	for _, p := range parsed {
		if p.suggestion.RelationshipType != "" {
			typeNames[p.suggestion.RelationshipType] = true
		}
	}

	if len(typeNames) == 0 {
		return nil, nil
	}

	// Query constraints for all referenced relationship types in
	// this campaign. Campaign-scoped types use the campaign_id
	// column; template types have campaign_id IS NULL.
	constraints, err := queryConstraints(ctx, db, campaignID, typeNames)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to query type pair constraints: %w", err,
		)
	}

	// Validate each suggestion against the constraints.
	var violations []TypePairViolation
	for _, p := range parsed {
		rtName := p.suggestion.RelationshipType
		pairs, hasConstraints := constraints[rtName]
		if !hasConstraints {
			// No constraints defined for this type; it is valid.
			continue
		}

		// Resolve entity types.
		sourceEntity, sourceOK := entityByID[p.suggestion.SourceEntityID]
		targetEntity, targetOK := entityByID[p.suggestion.TargetEntityID]
		if !sourceOK || !targetOK {
			// Cannot validate without knowing both entity types.
			continue
		}

		sourceType := string(sourceEntity.EntityType)
		targetType := string(targetEntity.EntityType)

		// Check whether this source-target pair is allowed.
		pairKey := sourceType + ":" + targetType
		if !pairs[pairKey] {
			violations = append(violations, TypePairViolation{
				SourceEntityID:   p.suggestion.SourceEntityID,
				SourceEntityName: p.suggestion.SourceEntityName,
				SourceEntityType: sourceType,
				TargetEntityID:   p.suggestion.TargetEntityID,
				TargetEntityName: p.suggestion.TargetEntityName,
				TargetEntityType: targetType,
				RelationshipType: rtName,
				ValidPairs:       formatValidPairs(pairs),
			})
		}
	}

	return violations, nil
}

// constraintMap maps relationship type name -> set of valid
// "sourceType:targetType" pairs.
type constraintMap map[string]map[string]bool

// queryConstraints retrieves type pair constraints for the given
// relationship type names from the database. It queries both
// campaign-scoped types and template types (campaign_id IS NULL).
func queryConstraints(
	ctx context.Context,
	db *database.DB,
	campaignID int64,
	typeNames map[string]bool,
) (constraintMap, error) {
	if len(typeNames) == 0 {
		return nil, nil
	}

	// Build the IN clause for type names.
	names := make([]string, 0, len(typeNames))
	for name := range typeNames {
		names = append(names, name)
	}

	// Query constraints joining through relationship_types to
	// resolve type names.
	query := `
		SELECT rt.name, rtc.source_entity_type, rtc.target_entity_type
		FROM relationship_type_constraints rtc
		JOIN relationship_types rt ON rtc.relationship_type_id = rt.id
		WHERE rt.name = ANY($1)
		  AND (rt.campaign_id = $2 OR rt.campaign_id IS NULL)`

	rows, err := db.Query(ctx, query, names, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to query constraints: %w", err)
	}
	defer rows.Close()

	result := make(constraintMap)
	for rows.Next() {
		var rtName, sourceType, targetType string
		if err := rows.Scan(&rtName, &sourceType, &targetType); err != nil {
			return nil, fmt.Errorf("failed to scan constraint row: %w", err)
		}

		if result[rtName] == nil {
			result[rtName] = make(map[string]bool)
		}
		result[rtName][sourceType+":"+targetType] = true
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating constraint rows: %w", err)
	}

	return result, nil
}

// formatValidPairs produces a human-readable description of the valid
// source:target type pairs for error messages.
func formatValidPairs(pairs map[string]bool) string {
	if len(pairs) == 0 {
		return "none"
	}

	descriptions := make([]string, 0, len(pairs))
	for pair := range pairs {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			descriptions = append(descriptions,
				parts[0]+" -> "+parts[1],
			)
		}
	}

	return strings.Join(descriptions, ", ")
}
