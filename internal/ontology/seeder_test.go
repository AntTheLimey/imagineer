/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package ontology

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDB records SQL calls for verification without
// requiring a real database connection.
type mockDB struct {
	execCalls  []mockExecCall
	queryRowFn func(ctx context.Context, sql string,
		args ...interface{}) pgx.Row
}

type mockExecCall struct {
	SQL  string
	Args []interface{}
}

func (m *mockDB) Exec(ctx context.Context, sql string,
	arguments ...interface{},
) (pgconn.CommandTag, error) {
	m.execCalls = append(m.execCalls, mockExecCall{
		SQL:  sql,
		Args: arguments,
	})
	return pgconn.NewCommandTag("INSERT 0 1"), nil
}

func (m *mockDB) QueryRow(ctx context.Context,
	sql string, args ...interface{},
) pgx.Row {
	if m.queryRowFn != nil {
		return m.queryRowFn(ctx, sql, args...)
	}
	return nil
}

func TestSeedCampaignEntityTypes(t *testing.T) {
	et := &EntityTypeFile{
		Types: map[string]EntityTypeDef{
			"agent": {
				Abstract:    true,
				Description: "An entity that can act",
				Children:    []string{"npc"},
			},
			"npc": {
				Parent:      "agent",
				Abstract:    false,
				Description: "Non-player character",
			},
		},
	}

	db := &mockDB{}
	err := SeedCampaignEntityTypes(
		context.Background(), db, 42, et)
	require.NoError(t, err)

	// Should have made exactly one Exec call.
	require.Len(t, db.execCalls, 1)

	call := db.execCalls[0]
	assert.Contains(t, call.SQL,
		"INSERT INTO campaign_entity_types")

	// Two types x 5 params = 10 args.
	assert.Len(t, call.Args, 10)

	// Verify campaign_id appears in args.
	assert.Equal(t, int64(42), call.Args[0])
}

func TestSeedCampaignRelationshipTypes(t *testing.T) {
	rt := &RelationshipTypeFile{
		Types: map[string]RelationshipTypeDef{
			"owns": {
				Inverse:             "owned_by",
				Symmetric:           false,
				DisplayLabel:        "Owns",
				InverseDisplayLabel: "Owned by",
				Description:         "Ownership",
			},
			"knows": {
				Inverse:             "knows",
				Symmetric:           true,
				DisplayLabel:        "Knows",
				InverseDisplayLabel: "Knows",
				Description:         "Acquaintance",
			},
		},
	}

	db := &mockDB{}
	err := SeedCampaignRelationshipTypes(
		context.Background(), db, 42, rt)
	require.NoError(t, err)

	require.Len(t, db.execCalls, 1)
	call := db.execCalls[0]
	assert.Contains(t, call.SQL,
		"INSERT INTO relationship_types")

	// Two types x 7 params = 14 args.
	assert.Len(t, call.Args, 14)
}

func TestSeedCampaignRequiredRelationships(t *testing.T) {
	c := &ConstraintsFile{
		Required: map[string][]string{
			"npc": {"located_at"},
			"pc":  {"played_by", "located_at"},
		},
	}

	db := &mockDB{}
	err := SeedCampaignRequiredRelationships(
		context.Background(), db, 42, c)
	require.NoError(t, err)

	require.Len(t, db.execCalls, 1)
	call := db.execCalls[0]
	assert.Contains(t, call.SQL,
		"INSERT INTO required_relationships")

	// 3 required relationships x 3 params = 9 args.
	assert.Len(t, call.Args, 9)
}

func TestSeedDefaultEra(t *testing.T) {
	db := &mockDB{}
	err := SeedDefaultEra(
		context.Background(), db, 42)
	require.NoError(t, err)

	require.Len(t, db.execCalls, 1)
	call := db.execCalls[0]
	assert.Contains(t, call.SQL, "INSERT INTO eras")
	assert.Contains(t, call.Args, int64(42))
}

func TestSeedEmptyInputs(t *testing.T) {
	db := &mockDB{}

	// Empty entity types should not call Exec.
	err := SeedCampaignEntityTypes(
		context.Background(), db, 42,
		&EntityTypeFile{Types: map[string]EntityTypeDef{}})
	require.NoError(t, err)
	assert.Empty(t, db.execCalls)

	// Empty relationship types should not call Exec.
	err = SeedCampaignRelationshipTypes(
		context.Background(), db, 42,
		&RelationshipTypeFile{
			Types: map[string]RelationshipTypeDef{}})
	require.NoError(t, err)
	assert.Empty(t, db.execCalls)

	// Empty required relationships should not call Exec.
	err = SeedCampaignRequiredRelationships(
		context.Background(), db, 42,
		&ConstraintsFile{Required: map[string][]string{}})
	require.NoError(t, err)
	assert.Empty(t, db.execCalls)
}

func TestResolveTypes(t *testing.T) {
	et := &EntityTypeFile{
		Types: map[string]EntityTypeDef{
			"agent": {
				Abstract: true,
				Children: []string{"npc", "creature"},
			},
			"npc": {
				Parent:   "agent",
				Abstract: false,
			},
			"creature": {
				Parent:   "agent",
				Abstract: false,
				Children: []string{"undead"},
			},
			"undead": {
				Parent:   "creature",
				Abstract: false,
			},
			"location": {
				Abstract: false,
			},
		},
	}

	// Test resolving an abstract type.
	result := resolveTypes([]string{"agent"}, et)
	assert.Contains(t, result, "npc")
	assert.Contains(t, result, "creature")
	assert.Contains(t, result, "undead")
	assert.NotContains(t, result, "agent")

	// Test resolving a concrete type.
	result = resolveTypes([]string{"location"}, et)
	assert.Equal(t, []string{"location"}, result)

	// Test resolving "any".
	result = resolveTypes([]string{"any"}, et)
	assert.Contains(t, result, "npc")
	assert.Contains(t, result, "creature")
	assert.Contains(t, result, "undead")
	assert.Contains(t, result, "location")
	assert.NotContains(t, result, "agent")
}
