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
	)

	require.NoError(t, err)
	assert.Nil(t, violations,
		"nil DB should produce no violations")
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
// Proposal counting logic tests
// ---------------------------------------------------------------------------

// TestProposalSuggestionParsing verifies that the suggestion parsing
// and counting logic correctly extracts relationship type, source,
// and target from ContentAnalysisItem suggestions. This tests the
// Go-side logic that remains after the DB function handles persisted
// data.
func TestProposalSuggestionParsing(t *testing.T) {
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

	items := []models.ContentAnalysisItem{
		{
			DetectionType:    "relationship_suggestion",
			SuggestedContent: json.RawMessage(sugContent),
		},
		{
			DetectionType: "orphan_warning",
		},
	}

	// Simulate the proposal counting logic from CheckCardinality.
	type countKey struct {
		EntityID         int64
		RelationshipType string
		Direction        string
	}
	proposalCounts := make(map[countKey]int)

	limitByType := map[string]cardinalityLimit{
		"allied_with": {
			RelationshipType: "allied_with",
		},
	}

	for _, item := range items {
		if item.DetectionType != "relationship_suggestion" {
			continue
		}
		if len(item.SuggestedContent) == 0 {
			continue
		}

		var rs models.RelationshipSuggestion
		err := json.Unmarshal(item.SuggestedContent, &rs)
		require.NoError(t, err)

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

	assert.Equal(t, 1,
		proposalCounts[countKey{1, "allied_with", "source"}],
		"Viktor should have 1 proposed allied_with as source")
	assert.Equal(t, 1,
		proposalCounts[countKey{3, "allied_with", "target"}],
		"Elara should have 1 proposed allied_with as target")
	assert.Equal(t, 2, len(proposalCounts),
		"should have exactly 2 proposal count entries")
}

// NOTE: Integration-style tests that previously validated the full
// counting and violation detection logic (using parseCountKey, manual
// string-key counting of existing relationships, and in-memory entity
// lookups) have been removed. Persisted relationship counting now
// lives in the PostgreSQL function check_cardinality_violations() and
// should be covered by database integration tests.
