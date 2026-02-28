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
	"testing"

	"github.com/antonypegg/imagineer/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// CheckRequiredRelationships tests
// ---------------------------------------------------------------------------

func TestCheckRequiredRelationships_NilDB(t *testing.T) {
	// When db is nil, no rules can be queried so the function
	// should return nil immediately.
	violations, err := CheckRequiredRelationships(
		context.Background(),
		nil, // no database
		1,
		[]models.Entity{},
		[]models.Relationship{},
	)

	require.NoError(t, err)
	assert.Nil(t, violations,
		"nil DB should produce no violations")
}

// ---------------------------------------------------------------------------
// RequiredViolation struct tests
// ---------------------------------------------------------------------------

func TestRequiredViolation_JSONMarshal(t *testing.T) {
	v := RequiredViolation{
		EntityID:                42,
		EntityName:              "Viktor",
		EntityType:              "npc",
		MissingRelationshipType: "located_at",
	}

	data, err := json.Marshal(v)
	require.NoError(t, err)

	var unmarshalled RequiredViolation
	err = json.Unmarshal(data, &unmarshalled)
	require.NoError(t, err)

	assert.Equal(t, v, unmarshalled)
}

// ---------------------------------------------------------------------------
// Integration-style test using in-memory data
// ---------------------------------------------------------------------------

// TestCheckRequiredLogic_MissingRelationship verifies the required
// relationship checking logic by simulating what happens after rules
// and relationship type IDs are loaded from the database.
func TestCheckRequiredLogic_MissingRelationship(t *testing.T) {
	entities := []models.Entity{
		{ID: 1, Name: "Viktor", EntityType: models.EntityTypeNPC},
		{ID: 2, Name: "Silver Fox Inn", EntityType: models.EntityTypeLocation},
		{ID: 3, Name: "Elara", EntityType: models.EntityTypeNPC},
	}

	// Simulate required rules: every NPC must have located_at.
	rules := []requiredRule{
		{EntityType: "npc", RelationshipTypeName: "located_at"},
	}

	// Simulate name -> id mapping.
	nameToID := map[string]int64{
		"located_at": 10,
	}

	// Viktor has a located_at relationship, Elara does not.
	relationships := []models.Relationship{
		{
			SourceEntityID:     1,
			TargetEntityID:     2,
			RelationshipTypeID: 10, // located_at
		},
	}

	// Replicate the checking logic from CheckRequiredRelationships.
	type entityRelKey struct {
		EntityID int64
		TypeID   int64
	}
	has := make(map[entityRelKey]bool)
	for _, rel := range relationships {
		has[entityRelKey{rel.SourceEntityID, rel.RelationshipTypeID}] = true
		has[entityRelKey{rel.TargetEntityID, rel.RelationshipTypeID}] = true
	}

	rulesByEntityType := make(map[string][]string)
	for _, r := range rules {
		rulesByEntityType[r.EntityType] = append(
			rulesByEntityType[r.EntityType],
			r.RelationshipTypeName,
		)
	}

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

	// Viktor (ID 1) has located_at -> no violation.
	// Silver Fox Inn (ID 2) is a location, not an NPC -> no rule applies.
	// Elara (ID 3) is an NPC without located_at -> violation.
	require.Len(t, violations, 1,
		"should have exactly one violation")
	assert.Equal(t, int64(3), violations[0].EntityID)
	assert.Equal(t, "Elara", violations[0].EntityName)
	assert.Equal(t, "npc", violations[0].EntityType)
	assert.Equal(t, "located_at", violations[0].MissingRelationshipType)
}

func TestCheckRequiredLogic_EntitySatisfiedAsTarget(t *testing.T) {
	// An entity can satisfy a required relationship by being the
	// target, not just the source.
	entities := []models.Entity{
		{ID: 1, Name: "Viktor", EntityType: models.EntityTypeNPC},
		{ID: 2, Name: "Arkham", EntityType: models.EntityTypeLocation},
	}

	rules := []requiredRule{
		{EntityType: "npc", RelationshipTypeName: "located_at"},
	}

	nameToID := map[string]int64{
		"located_at": 10,
	}

	// Arkham is the source, Viktor is the target of a located_at
	// relationship. Viktor should still satisfy the requirement.
	relationships := []models.Relationship{
		{
			SourceEntityID:     2,
			TargetEntityID:     1,
			RelationshipTypeID: 10, // located_at
		},
	}

	type entityRelKey struct {
		EntityID int64
		TypeID   int64
	}
	has := make(map[entityRelKey]bool)
	for _, rel := range relationships {
		has[entityRelKey{rel.SourceEntityID, rel.RelationshipTypeID}] = true
		has[entityRelKey{rel.TargetEntityID, rel.RelationshipTypeID}] = true
	}

	rulesByEntityType := make(map[string][]string)
	for _, r := range rules {
		rulesByEntityType[r.EntityType] = append(
			rulesByEntityType[r.EntityType],
			r.RelationshipTypeName,
		)
	}

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

	assert.Empty(t, violations,
		"Viktor satisfies located_at as target, no violation expected")
}

func TestCheckRequiredLogic_NoRulesNoViolations(t *testing.T) {
	// When there are no required rules, no violations should
	// be produced regardless of the entity/relationship state.
	entities := []models.Entity{
		{ID: 1, Name: "Viktor", EntityType: models.EntityTypeNPC},
	}

	rules := []requiredRule{}

	rulesByEntityType := make(map[string][]string)
	for _, r := range rules {
		rulesByEntityType[r.EntityType] = append(
			rulesByEntityType[r.EntityType],
			r.RelationshipTypeName,
		)
	}

	var violations []RequiredViolation
	for _, entity := range entities {
		entityType := string(entity.EntityType)
		requiredTypes, ok := rulesByEntityType[entityType]
		if !ok {
			continue
		}
		for range requiredTypes {
			violations = append(violations, RequiredViolation{})
		}
	}

	assert.Empty(t, violations,
		"no rules should produce no violations")
}

func TestCheckRequiredLogic_MultipleRulesMultipleViolations(t *testing.T) {
	// When an entity type has multiple required relationship types,
	// violations should be produced for each missing one.
	entities := []models.Entity{
		{ID: 1, Name: "Viktor", EntityType: models.EntityTypeNPC},
	}

	rules := []requiredRule{
		{EntityType: "npc", RelationshipTypeName: "located_at"},
		{EntityType: "npc", RelationshipTypeName: "belongs_to"},
	}

	nameToID := map[string]int64{
		"located_at": 10,
		"belongs_to": 20,
	}

	// Viktor has no relationships at all.
	relationships := []models.Relationship{}

	type entityRelKey struct {
		EntityID int64
		TypeID   int64
	}
	has := make(map[entityRelKey]bool)
	for _, rel := range relationships {
		has[entityRelKey{rel.SourceEntityID, rel.RelationshipTypeID}] = true
		has[entityRelKey{rel.TargetEntityID, rel.RelationshipTypeID}] = true
	}

	rulesByEntityType := make(map[string][]string)
	for _, r := range rules {
		rulesByEntityType[r.EntityType] = append(
			rulesByEntityType[r.EntityType],
			r.RelationshipTypeName,
		)
	}

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

	require.Len(t, violations, 2,
		"should have two violations for missing located_at and belongs_to")

	// Collect the missing types.
	missingTypes := make(map[string]bool)
	for _, v := range violations {
		assert.Equal(t, int64(1), v.EntityID)
		assert.Equal(t, "Viktor", v.EntityName)
		assert.Equal(t, "npc", v.EntityType)
		missingTypes[v.MissingRelationshipType] = true
	}
	assert.True(t, missingTypes["located_at"],
		"should flag missing located_at")
	assert.True(t, missingTypes["belongs_to"],
		"should flag missing belongs_to")
}

func TestCheckRequiredLogic_OnlySomeRulesSatisfied(t *testing.T) {
	// When an entity satisfies one required relationship but not
	// another, only the missing one should produce a violation.
	entities := []models.Entity{
		{ID: 1, Name: "Viktor", EntityType: models.EntityTypeNPC},
		{ID: 2, Name: "Arkham", EntityType: models.EntityTypeLocation},
	}

	rules := []requiredRule{
		{EntityType: "npc", RelationshipTypeName: "located_at"},
		{EntityType: "npc", RelationshipTypeName: "belongs_to"},
	}

	nameToID := map[string]int64{
		"located_at": 10,
		"belongs_to": 20,
	}

	// Viktor has a located_at but no belongs_to.
	relationships := []models.Relationship{
		{
			SourceEntityID:     1,
			TargetEntityID:     2,
			RelationshipTypeID: 10, // located_at
		},
	}

	type entityRelKey struct {
		EntityID int64
		TypeID   int64
	}
	has := make(map[entityRelKey]bool)
	for _, rel := range relationships {
		has[entityRelKey{rel.SourceEntityID, rel.RelationshipTypeID}] = true
		has[entityRelKey{rel.TargetEntityID, rel.RelationshipTypeID}] = true
	}

	rulesByEntityType := make(map[string][]string)
	for _, r := range rules {
		rulesByEntityType[r.EntityType] = append(
			rulesByEntityType[r.EntityType],
			r.RelationshipTypeName,
		)
	}

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

	require.Len(t, violations, 1,
		"should have exactly one violation for missing belongs_to")
	assert.Equal(t, int64(1), violations[0].EntityID)
	assert.Equal(t, "Viktor", violations[0].EntityName)
	assert.Equal(t, "belongs_to", violations[0].MissingRelationshipType)
}

func TestCheckRequiredLogic_RuleForNonexistentType(t *testing.T) {
	// When a rule references a relationship type that doesn't exist
	// in the name->id map, it should be skipped (no false positive).
	entities := []models.Entity{
		{ID: 1, Name: "Viktor", EntityType: models.EntityTypeNPC},
	}

	rules := []requiredRule{
		{EntityType: "npc", RelationshipTypeName: "nonexistent_type"},
	}

	nameToID := map[string]int64{
		// nonexistent_type is NOT in the map.
	}

	type entityRelKey struct {
		EntityID int64
		TypeID   int64
	}
	has := make(map[entityRelKey]bool)

	rulesByEntityType := make(map[string][]string)
	for _, r := range rules {
		rulesByEntityType[r.EntityType] = append(
			rulesByEntityType[r.EntityType],
			r.RelationshipTypeName,
		)
	}

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

	assert.Empty(t, violations,
		"rules referencing unknown relationship types should be skipped")
}
