/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package database

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/antonypegg/imagineer/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRelationship_Structure verifies that the Relationship model properly
// handles all fields used by entity_relationships_view queries, including
// the joined fields populated by ListChapterRelationships and
// GetEntityRelationships.
func TestRelationship_Structure(t *testing.T) {
	tone := models.RelationshipTone("positive")
	description := "A trusted ally"
	strength := 8

	r := models.Relationship{
		ID:                   1,
		CampaignID:           42,
		SourceEntityID:       10,
		TargetEntityID:       20,
		RelationshipTypeID:   3,
		Tone:                 &tone,
		Description:          &description,
		Strength:             &strength,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
		RelationshipTypeName: "allied_with",
		DisplayLabel:         "Allied with",
		Direction:            "forward",
		SourceEntityName:     "Arkham Police",
		SourceEntityType:     "faction",
		TargetEntityName:     "Harvey Walters",
		TargetEntityType:     "npc",
	}

	assert.Equal(t, int64(1), r.ID)
	assert.Equal(t, int64(42), r.CampaignID)
	assert.Equal(t, int64(10), r.SourceEntityID)
	assert.Equal(t, int64(20), r.TargetEntityID)
	assert.Equal(t, int64(3), r.RelationshipTypeID)
	assert.Equal(t, "allied_with", r.RelationshipTypeName)
	assert.Equal(t, "Allied with", r.DisplayLabel)
	assert.Equal(t, "forward", r.Direction)
	assert.Equal(t, "Arkham Police", r.SourceEntityName)
	assert.Equal(t, "faction", r.SourceEntityType)
	assert.Equal(t, "Harvey Walters", r.TargetEntityName)
	assert.Equal(t, "npc", r.TargetEntityType)

	require.NotNil(t, r.Tone)
	assert.Equal(t, models.RelationshipTone("positive"), *r.Tone)

	require.NotNil(t, r.Description)
	assert.Equal(t, "A trusted ally", *r.Description)

	require.NotNil(t, r.Strength)
	assert.Equal(t, 8, *r.Strength)
}

// TestRelationship_JSONRoundTrip verifies that a Relationship with all
// view-sourced fields can be marshaled to JSON and unmarshaled without
// data loss. This exercises the same fields that ListChapterRelationships
// and GetEntityRelationships populate.
func TestRelationship_JSONRoundTrip(t *testing.T) {
	tone := models.RelationshipTone("hostile")
	description := "Long-standing rivalry"
	strength := 5
	now := time.Now().Truncate(time.Millisecond)

	original := models.Relationship{
		ID:                   7,
		CampaignID:           1,
		SourceEntityID:       100,
		TargetEntityID:       200,
		RelationshipTypeID:   4,
		Tone:                 &tone,
		Description:          &description,
		Strength:             &strength,
		CreatedAt:            now,
		UpdatedAt:            now,
		RelationshipTypeName: "enemy_of",
		DisplayLabel:         "Enemy of",
		Direction:            "forward",
		SourceEntityName:     "Nyarlathotep",
		SourceEntityType:     "npc",
		TargetEntityName:     "Investigators",
		TargetEntityType:     "faction",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var result models.Relationship
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, original.ID, result.ID)
	assert.Equal(t, original.CampaignID, result.CampaignID)
	assert.Equal(t, original.SourceEntityID, result.SourceEntityID)
	assert.Equal(t, original.TargetEntityID, result.TargetEntityID)
	assert.Equal(t, original.RelationshipTypeID, result.RelationshipTypeID)
	assert.Equal(t, original.RelationshipTypeName, result.RelationshipTypeName)
	assert.Equal(t, original.DisplayLabel, result.DisplayLabel)
	assert.Equal(t, original.Direction, result.Direction)
	assert.Equal(t, original.SourceEntityName, result.SourceEntityName)
	assert.Equal(t, original.SourceEntityType, result.SourceEntityType)
	assert.Equal(t, original.TargetEntityName, result.TargetEntityName)
	assert.Equal(t, original.TargetEntityType, result.TargetEntityType)

	require.NotNil(t, result.Tone)
	assert.Equal(t, *original.Tone, *result.Tone)
	require.NotNil(t, result.Description)
	assert.Equal(t, *original.Description, *result.Description)
	require.NotNil(t, result.Strength)
	assert.Equal(t, *original.Strength, *result.Strength)
}

// TestRelationship_NilOptionalFields verifies that optional pointer fields
// (Tone, Description, Strength) are correctly represented as nil when not
// set. This matches what the database returns for NULL columns.
func TestRelationship_NilOptionalFields(t *testing.T) {
	r := models.Relationship{
		ID:                   2,
		CampaignID:           1,
		SourceEntityID:       10,
		TargetEntityID:       20,
		RelationshipTypeID:   1,
		RelationshipTypeName: "knows",
		DisplayLabel:         "Knows",
		Direction:            "forward",
		SourceEntityName:     "Investigator A",
		SourceEntityType:     "npc",
		TargetEntityName:     "Investigator B",
		TargetEntityType:     "npc",
	}

	assert.Nil(t, r.Tone)
	assert.Nil(t, r.Description)
	assert.Nil(t, r.Strength)

	// Verify JSON output omits nil optional fields
	data, err := json.Marshal(r)
	require.NoError(t, err)

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	_, hasTone := raw["tone"]
	_, hasDesc := raw["description"]
	_, hasStr := raw["strength"]

	assert.False(t, hasTone, "nil tone should be omitted from JSON")
	assert.False(t, hasDesc, "nil description should be omitted from JSON")
	assert.False(t, hasStr, "nil strength should be omitted from JSON")
}

// TestListChapterRelationships_QueryStructure verifies that the
// ListChapterRelationships method exists on the DB type with the correct
// signature. Since this function requires a live database connection for
// full integration testing, this test validates the type signature and
// return types.
func TestListChapterRelationships_QueryStructure(t *testing.T) {
	// Verify the function signature by assigning it to a typed variable.
	// This is a compile-time check that ListChapterRelationships exists
	// on *DB with the expected parameters and return types.
	var _ func(ctx context.Context, campaignID, chapterID int64) ([]models.Relationship, error) = (*DB)(nil).ListChapterRelationships
}
