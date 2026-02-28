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
	"testing"

	"github.com/antonypegg/imagineer/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// CheckCardinality tests
// ---------------------------------------------------------------------------

func TestCheckCardinality_NilDB(t *testing.T) {
	// When db is nil, no constraints can be queried so the function
	// should return nil immediately.
	violations, err := CheckCardinality(
		context.Background(),
		nil, // no database
		1,
		[]models.ContentAnalysisItem{},
		[]models.Relationship{},
		[]models.Entity{},
	)

	require.NoError(t, err)
	assert.Nil(t, violations,
		"nil DB should produce no violations")
}

// ---------------------------------------------------------------------------
// parseCountKey tests
// ---------------------------------------------------------------------------

func TestParseCountKey_Valid(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		wantEntityID  int64
		wantTypeName  string
		wantDirection string
	}{
		{
			name:          "simple type name",
			key:           "42:allied_with:source",
			wantEntityID:  42,
			wantTypeName:  "allied_with",
			wantDirection: "source",
		},
		{
			name:          "target direction",
			key:           "7:belongs_to:target",
			wantEntityID:  7,
			wantTypeName:  "belongs_to",
			wantDirection: "target",
		},
		{
			name:          "type name with colons",
			key:           "100:some:complex:type:source",
			wantEntityID:  100,
			wantTypeName:  "some:complex:type",
			wantDirection: "source",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			entityID, typeName, direction := parseCountKey(tc.key)
			assert.Equal(t, tc.wantEntityID, entityID)
			assert.Equal(t, tc.wantTypeName, typeName)
			assert.Equal(t, tc.wantDirection, direction)
		})
	}
}

func TestParseCountKey_Invalid(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{name: "empty string", key: ""},
		{name: "no colons", key: "nocolons"},
		{name: "single colon", key: "42:onlytype"},
		{name: "non-numeric entity ID", key: "abc:type:source"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			entityID, typeName, direction := parseCountKey(tc.key)
			assert.Equal(t, int64(0), entityID)
			assert.Empty(t, typeName)
			assert.Empty(t, direction)
		})
	}
}

// ---------------------------------------------------------------------------
// CardinalityViolation struct tests
// ---------------------------------------------------------------------------

func TestCardinalityViolation_JSONMarshal(t *testing.T) {
	v := CardinalityViolation{
		EntityID:         42,
		EntityName:       "Viktor",
		EntityType:       "npc",
		RelationshipType: "allied_with",
		Direction:        "source",
		CurrentCount:     5,
		MaxAllowed:       3,
	}

	data, err := json.Marshal(v)
	require.NoError(t, err)

	var unmarshalled CardinalityViolation
	err = json.Unmarshal(data, &unmarshalled)
	require.NoError(t, err)

	assert.Equal(t, v, unmarshalled)
}

// ---------------------------------------------------------------------------
// Integration-style test using in-memory data
// ---------------------------------------------------------------------------

// TestCheckCardinalityLogic verifies the counting and violation
// detection logic by testing the internal helpers directly.
// The full CheckCardinality function requires a database connection
// to query cardinality_constraints; this test validates the logic
// that runs after constraints are loaded.
func TestCheckCardinalityLogic_CountsAndViolations(t *testing.T) {
	// Simulate the counting logic from CheckCardinality.
	// This mirrors what happens inside the function after
	// limits are loaded from the database.

	entities := []models.Entity{
		{ID: 1, Name: "Viktor", EntityType: models.EntityTypeNPC},
		{ID: 2, Name: "Silver Fox Inn", EntityType: models.EntityTypeLocation},
		{ID: 3, Name: "Elara", EntityType: models.EntityTypeNPC},
	}

	entityByID := make(map[int64]models.Entity, len(entities))
	for _, e := range entities {
		entityByID[e.ID] = e
	}

	// Simulate limits: allied_with has max_source=2, max_target=nil (unlimited)
	maxSource := 2
	limitByType := map[string]cardinalityLimit{
		"allied_with": {
			RelationshipType: "allied_with",
			MaxSource:        &maxSource,
			MaxTarget:        nil,
		},
	}

	// Existing relationships: Viktor is source in 2 allied_with relationships.
	relationships := []models.Relationship{
		{
			SourceEntityID:       1,
			TargetEntityID:       2,
			RelationshipTypeName: "allied_with",
		},
		{
			SourceEntityID:       1,
			TargetEntityID:       3,
			RelationshipTypeName: "allied_with",
		},
	}

	// A suggestion that would add a third allied_with from Viktor.
	suggestion := models.RelationshipSuggestion{
		SourceEntityID:   1,
		SourceEntityName: "Viktor",
		TargetEntityID:   3,
		TargetEntityName: "Elara",
		RelationshipType: "allied_with",
		Description:      "Viktor is allied with Elara.",
	}

	sugContent, err := json.Marshal(suggestion)
	require.NoError(t, err)

	suggestions := []models.ContentAnalysisItem{
		{
			DetectionType:    "relationship_suggestion",
			SuggestedContent: json.RawMessage(sugContent),
		},
	}

	// Replicate the counting logic.
	counts := make(map[string]int)
	for _, r := range relationships {
		typeName := r.RelationshipTypeName
		if _, hasLimit := limitByType[typeName]; !hasLimit {
			continue
		}
		sourceKey := keyFor(r.SourceEntityID, typeName, "source")
		targetKey := keyFor(r.TargetEntityID, typeName, "target")
		counts[sourceKey]++
		counts[targetKey]++
	}

	for _, item := range suggestions {
		if item.DetectionType != "relationship_suggestion" {
			continue
		}
		var rs models.RelationshipSuggestion
		err := json.Unmarshal(item.SuggestedContent, &rs)
		require.NoError(t, err)

		typeName := rs.RelationshipType
		if _, hasLimit := limitByType[typeName]; !hasLimit {
			continue
		}
		sourceKey := keyFor(rs.SourceEntityID, typeName, "source")
		targetKey := keyFor(rs.TargetEntityID, typeName, "target")
		counts[sourceKey]++
		counts[targetKey]++
	}

	// Viktor as source of allied_with: 2 existing + 1 suggestion = 3
	viktorSourceKey := keyFor(1, "allied_with", "source")
	assert.Equal(t, 3, counts[viktorSourceKey],
		"Viktor should have 3 total allied_with as source")

	// Check for violations.
	var violations []CardinalityViolation
	for key, count := range counts {
		entityID, typeName, direction := parseCountKey(key)
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
			entity := entityByID[entityID]
			violations = append(violations, CardinalityViolation{
				EntityID:         entityID,
				EntityName:       entity.Name,
				EntityType:       string(entity.EntityType),
				RelationshipType: typeName,
				Direction:        direction,
				CurrentCount:     count,
				MaxAllowed:       *maxAllowed,
			})
		}
	}

	require.Len(t, violations, 1,
		"should have exactly one violation")
	assert.Equal(t, int64(1), violations[0].EntityID)
	assert.Equal(t, "Viktor", violations[0].EntityName)
	assert.Equal(t, "npc", violations[0].EntityType)
	assert.Equal(t, "allied_with", violations[0].RelationshipType)
	assert.Equal(t, "source", violations[0].Direction)
	assert.Equal(t, 3, violations[0].CurrentCount)
	assert.Equal(t, 2, violations[0].MaxAllowed)
}

func TestCheckCardinalityLogic_NoViolation(t *testing.T) {
	// When counts are within limits, no violations should be produced.

	maxSource := 5
	limitByType := map[string]cardinalityLimit{
		"allied_with": {
			RelationshipType: "allied_with",
			MaxSource:        &maxSource,
			MaxTarget:        nil,
		},
	}

	relationships := []models.Relationship{
		{
			SourceEntityID:       1,
			TargetEntityID:       2,
			RelationshipTypeName: "allied_with",
		},
	}

	counts := make(map[string]int)
	for _, r := range relationships {
		typeName := r.RelationshipTypeName
		if _, hasLimit := limitByType[typeName]; !hasLimit {
			continue
		}
		sourceKey := keyFor(r.SourceEntityID, typeName, "source")
		counts[sourceKey]++
	}

	// Count is 1, limit is 5 => no violation.
	var violations []CardinalityViolation
	for key, count := range counts {
		_, typeName, direction := parseCountKey(key)
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
			violations = append(violations, CardinalityViolation{})
		}
	}

	assert.Empty(t, violations,
		"should have no violations when within limits")
}

// keyFor is a test helper that builds a count map key.
func keyFor(entityID int64, typeName, direction string) string {
	return fmt.Sprintf("%d:%s:%s", entityID, typeName, direction)
}
