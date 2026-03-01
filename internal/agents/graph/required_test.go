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

// NOTE: Integration-style tests that previously validated the in-memory
// checking logic (using requiredRule, nameToID maps, and entity/relationship
// slices) have been removed. That logic now lives in the PostgreSQL function
// check_required_relationships() and should be covered by database
// integration tests.
