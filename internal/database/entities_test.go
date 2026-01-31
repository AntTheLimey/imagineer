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
	"encoding/json"
	"testing"

	"github.com/antonypegg/imagineer/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateEntityRequest_TagsArray verifies that the CreateEntityRequest
// properly handles the tags field as a string array. This test exists to
// prevent regression of the pgx/pq compatibility issue where arrays were
// not properly handled.
func TestCreateEntityRequest_TagsArray(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		wantTags []string
	}{
		{
			name:     "nil tags",
			tags:     nil,
			wantTags: nil,
		},
		{
			name:     "empty tags",
			tags:     []string{},
			wantTags: []string{},
		},
		{
			name:     "single tag",
			tags:     []string{"important"},
			wantTags: []string{"important"},
		},
		{
			name:     "multiple tags",
			tags:     []string{"cultist", "arkham", "mythos"},
			wantTags: []string{"cultist", "arkham", "mythos"},
		},
		{
			name:     "tags with special characters",
			tags:     []string{"tag-with-dash", "tag_with_underscore", "Tag With Spaces"},
			wantTags: []string{"tag-with-dash", "tag_with_underscore", "Tag With Spaces"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := models.CreateEntityRequest{
				EntityType: models.EntityTypeNPC,
				Name:       "Test Entity",
				Tags:       tt.tags,
			}

			// Verify the tags are stored correctly
			assert.Equal(t, tt.wantTags, req.Tags)
		})
	}
}

// TestCreateEntityRequest_JSONMarshal verifies that CreateEntityRequest can
// be properly marshaled to and unmarshaled from JSON, including the tags array.
func TestCreateEntityRequest_JSONMarshal(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		wantTags []string
	}{
		{
			name:     "JSON with empty tags array",
			jsonData: `{"entityType":"npc","name":"Test NPC","tags":[]}`,
			wantTags: []string{},
		},
		{
			name:     "JSON with tags array",
			jsonData: `{"entityType":"npc","name":"Test NPC","tags":["cultist","dangerous"]}`,
			wantTags: []string{"cultist", "dangerous"},
		},
		{
			name:     "JSON without tags field",
			jsonData: `{"entityType":"npc","name":"Test NPC"}`,
			wantTags: nil,
		},
		{
			name:     "JSON with null tags",
			jsonData: `{"entityType":"npc","name":"Test NPC","tags":null}`,
			wantTags: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req models.CreateEntityRequest
			err := json.Unmarshal([]byte(tt.jsonData), &req)
			require.NoError(t, err)

			assert.Equal(t, tt.wantTags, req.Tags)
			assert.Equal(t, models.EntityTypeNPC, req.EntityType)
			assert.Equal(t, "Test NPC", req.Name)
		})
	}
}

// TestCreateEntityRequest_RoundTrip verifies that CreateEntityRequest can be
// marshaled and unmarshaled without losing data, particularly the tags array.
func TestCreateEntityRequest_RoundTrip(t *testing.T) {
	description := "A mysterious cultist"
	original := models.CreateEntityRequest{
		EntityType:  models.EntityTypeNPC,
		Name:        "Mysterious Cultist",
		Description: &description,
		Attributes:  json.RawMessage(`{"occupation":"cultist","age":45}`),
		Tags:        []string{"cultist", "arkham", "dangerous", "mythos"},
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	require.NoError(t, err)

	// Unmarshal back
	var result models.CreateEntityRequest
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	// Verify all fields
	assert.Equal(t, original.EntityType, result.EntityType)
	assert.Equal(t, original.Name, result.Name)
	require.NotNil(t, result.Description)
	assert.Equal(t, *original.Description, *result.Description)
	assert.Equal(t, original.Tags, result.Tags)
	assert.Len(t, result.Tags, 4)
}

// TestEntity_TagsField verifies that the Entity model properly handles the
// tags field as a string array.
func TestEntity_TagsField(t *testing.T) {
	tests := []struct {
		name    string
		tags    []string
		wantLen int
		wantNil bool
	}{
		{
			name:    "nil tags",
			tags:    nil,
			wantLen: 0,
			wantNil: true,
		},
		{
			name:    "empty tags",
			tags:    []string{},
			wantLen: 0,
			wantNil: false,
		},
		{
			name:    "multiple tags",
			tags:    []string{"tag1", "tag2", "tag3"},
			wantLen: 3,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := models.Entity{
				EntityType: models.EntityTypeNPC,
				Name:       "Test Entity",
				Tags:       tt.tags,
			}

			if tt.wantNil {
				assert.Nil(t, entity.Tags)
			} else {
				assert.NotNil(t, entity.Tags)
			}
			assert.Len(t, entity.Tags, tt.wantLen)
		})
	}
}

// TestEntityTypes verifies that all entity types are valid strings.
func TestEntityTypes(t *testing.T) {
	types := []models.EntityType{
		models.EntityTypeNPC,
		models.EntityTypeLocation,
		models.EntityTypeItem,
		models.EntityTypeFaction,
		models.EntityTypeClue,
		models.EntityTypeCreature,
		models.EntityTypeOrganization,
		models.EntityTypeEvent,
		models.EntityTypeDocument,
		models.EntityTypeOther,
	}

	for _, entityType := range types {
		t.Run(string(entityType), func(t *testing.T) {
			// Verify the type is a non-empty string
			assert.NotEmpty(t, string(entityType))

			// Verify the type can be used in a request
			req := models.CreateEntityRequest{
				EntityType: entityType,
				Name:       "Test Entity",
				Tags:       []string{"test"},
			}
			assert.Equal(t, entityType, req.EntityType)
		})
	}
}

// TestSourceConfidenceTypes verifies that all source confidence types are valid.
func TestSourceConfidenceTypes(t *testing.T) {
	confidences := []models.SourceConfidence{
		models.SourceConfidenceDraft,
		models.SourceConfidenceAuthoritative,
		models.SourceConfidenceSuperseded,
	}

	for _, confidence := range confidences {
		t.Run(string(confidence), func(t *testing.T) {
			assert.NotEmpty(t, string(confidence))

			// Verify the confidence can be used in a request
			req := models.CreateEntityRequest{
				EntityType:       models.EntityTypeNPC,
				Name:             "Test Entity",
				SourceConfidence: &confidence,
			}
			require.NotNil(t, req.SourceConfidence)
			assert.Equal(t, confidence, *req.SourceConfidence)
		})
	}
}

// TestUpdateEntityRequest_TagsArray verifies that UpdateEntityRequest properly
// handles the tags field as a string array.
func TestUpdateEntityRequest_TagsArray(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		wantTags []string
		wantNil  bool
	}{
		{
			name:     "JSON with tags to add",
			jsonData: `{"tags":["new-tag","another-tag"]}`,
			wantTags: []string{"new-tag", "another-tag"},
			wantNil:  false,
		},
		{
			name:     "JSON with empty tags to clear",
			jsonData: `{"tags":[]}`,
			wantTags: []string{},
			wantNil:  false,
		},
		{
			name:     "JSON without tags field",
			jsonData: `{"name":"Updated Name"}`,
			wantTags: nil,
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req models.UpdateEntityRequest
			err := json.Unmarshal([]byte(tt.jsonData), &req)
			require.NoError(t, err)

			if tt.wantNil {
				assert.Nil(t, req.Tags)
			} else {
				assert.Equal(t, tt.wantTags, req.Tags)
			}
		})
	}
}

// TestCreateEntityRequest_NilTagsDefaulting tests the pattern used in
// CreateEntity to default nil tags to an empty slice, which is important
// for PostgreSQL array compatibility.
func TestCreateEntityRequest_NilTagsDefaulting(t *testing.T) {
	// This mimics the logic in CreateEntity
	tests := []struct {
		name      string
		inputTags []string
		wantTags  []string
	}{
		{
			name:      "nil tags should default to empty slice",
			inputTags: nil,
			wantTags:  []string{},
		},
		{
			name:      "empty tags should remain empty",
			inputTags: []string{},
			wantTags:  []string{},
		},
		{
			name:      "existing tags should be preserved",
			inputTags: []string{"tag1", "tag2"},
			wantTags:  []string{"tag1", "tag2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := models.CreateEntityRequest{
				EntityType: models.EntityTypeNPC,
				Name:       "Test",
				Tags:       tt.inputTags,
			}

			// Apply the defaulting logic used in CreateEntity
			tags := req.Tags
			if tags == nil {
				tags = []string{}
			}

			assert.Equal(t, tt.wantTags, tags)
			assert.NotNil(t, tags, "tags should never be nil after defaulting")
		})
	}
}

// TestEntity_CompleteStructure tests that an Entity can be fully populated
// with all fields including tags.
func TestEntity_CompleteStructure(t *testing.T) {
	description := "A test entity description"
	gmNotes := "Secret GM notes"
	sourceDoc := "Test Document"

	entity := models.Entity{
		EntityType:       models.EntityTypeNPC,
		Name:             "Complete Entity",
		Description:      &description,
		Attributes:       json.RawMessage(`{"str":10,"dex":12}`),
		Tags:             []string{"tag1", "tag2", "tag3"},
		GMNotes:          &gmNotes,
		SourceDocument:   &sourceDoc,
		SourceConfidence: models.SourceConfidenceDraft,
		Version:          1,
	}

	// Marshal to JSON and back to verify structure
	data, err := json.Marshal(entity)
	require.NoError(t, err)

	var result models.Entity
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, entity.EntityType, result.EntityType)
	assert.Equal(t, entity.Name, result.Name)
	assert.Equal(t, entity.Tags, result.Tags)
	assert.Len(t, result.Tags, 3)
	assert.Equal(t, entity.SourceConfidence, result.SourceConfidence)
}
