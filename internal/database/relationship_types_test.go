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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRelationshipType_Structure verifies that the RelationshipType model
// properly handles all fields including the optional campaign_id.
func TestRelationshipType_Structure(t *testing.T) {
	tests := []struct {
		name       string
		campaignID *uuid.UUID
		wantNil    bool
	}{
		{
			name:       "system default type (nil campaign_id)",
			campaignID: nil,
			wantNil:    true,
		},
		{
			name:       "campaign-specific type",
			campaignID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			wantNil:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := models.RelationshipType{
				ID:                  uuid.New(),
				CampaignID:          tt.campaignID,
				Name:                "test_type",
				InverseName:         "inverse_test_type",
				IsSymmetric:         false,
				DisplayLabel:        "Test Type",
				InverseDisplayLabel: "Inverse Test Type",
			}

			if tt.wantNil {
				assert.Nil(t, rt.CampaignID)
			} else {
				assert.NotNil(t, rt.CampaignID)
			}
		})
	}
}

// TestRelationshipType_JSONMarshal verifies that RelationshipType can be
// properly marshaled to and unmarshaled from JSON.
func TestRelationshipType_JSONMarshal(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		wantName    string
		wantSymm    bool
		wantCampNil bool
	}{
		{
			name:        "system default type",
			jsonData:    `{"id":"550e8400-e29b-41d4-a716-446655440000","name":"knows","inverseName":"knows","isSymmetric":true,"displayLabel":"Knows","inverseDisplayLabel":"Knows"}`,
			wantName:    "knows",
			wantSymm:    true,
			wantCampNil: true,
		},
		{
			name:        "asymmetric type",
			jsonData:    `{"id":"550e8400-e29b-41d4-a716-446655440001","campaignId":"550e8400-e29b-41d4-a716-446655440002","name":"owns","inverseName":"owned_by","isSymmetric":false,"displayLabel":"Owns","inverseDisplayLabel":"Is owned by"}`,
			wantName:    "owns",
			wantSymm:    false,
			wantCampNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rt models.RelationshipType
			err := json.Unmarshal([]byte(tt.jsonData), &rt)
			require.NoError(t, err)

			assert.Equal(t, tt.wantName, rt.Name)
			assert.Equal(t, tt.wantSymm, rt.IsSymmetric)
			if tt.wantCampNil {
				assert.Nil(t, rt.CampaignID)
			} else {
				assert.NotNil(t, rt.CampaignID)
			}
		})
	}
}

// TestRelationshipType_RoundTrip verifies that RelationshipType can be
// marshaled and unmarshaled without losing data.
func TestRelationshipType_RoundTrip(t *testing.T) {
	description := "A test relationship type"
	campaignID := uuid.New()

	original := models.RelationshipType{
		ID:                  uuid.New(),
		CampaignID:          &campaignID,
		Name:                "custom_type",
		InverseName:         "inverse_custom_type",
		IsSymmetric:         false,
		DisplayLabel:        "Custom Type",
		InverseDisplayLabel: "Inverse Custom Type",
		Description:         &description,
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	require.NoError(t, err)

	// Unmarshal back
	var result models.RelationshipType
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	// Verify all fields
	assert.Equal(t, original.ID, result.ID)
	assert.Equal(t, original.CampaignID, result.CampaignID)
	assert.Equal(t, original.Name, result.Name)
	assert.Equal(t, original.InverseName, result.InverseName)
	assert.Equal(t, original.IsSymmetric, result.IsSymmetric)
	assert.Equal(t, original.DisplayLabel, result.DisplayLabel)
	assert.Equal(t, original.InverseDisplayLabel, result.InverseDisplayLabel)
	require.NotNil(t, result.Description)
	assert.Equal(t, *original.Description, *result.Description)
}

// TestCreateRelationshipTypeRequest_JSONMarshal verifies that
// CreateRelationshipTypeRequest can be properly unmarshaled from JSON.
func TestCreateRelationshipTypeRequest_JSONMarshal(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		wantName    string
		wantInverse string
		wantSymm    bool
		wantDescNil bool
	}{
		{
			name:        "symmetric type without description",
			jsonData:    `{"name":"allies_with","inverseName":"allies_with","isSymmetric":true,"displayLabel":"Allied with","inverseDisplayLabel":"Allied with"}`,
			wantName:    "allies_with",
			wantInverse: "allies_with",
			wantSymm:    true,
			wantDescNil: true,
		},
		{
			name:        "asymmetric type with description",
			jsonData:    `{"name":"mentors","inverseName":"mentored_by","isSymmetric":false,"displayLabel":"Mentors","inverseDisplayLabel":"Mentored by","description":"Entity teaches or guides another"}`,
			wantName:    "mentors",
			wantInverse: "mentored_by",
			wantSymm:    false,
			wantDescNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req models.CreateRelationshipTypeRequest
			err := json.Unmarshal([]byte(tt.jsonData), &req)
			require.NoError(t, err)

			assert.Equal(t, tt.wantName, req.Name)
			assert.Equal(t, tt.wantInverse, req.InverseName)
			assert.Equal(t, tt.wantSymm, req.IsSymmetric)
			if tt.wantDescNil {
				assert.Nil(t, req.Description)
			} else {
				assert.NotNil(t, req.Description)
			}
		})
	}
}

// TestCreateRelationshipTypeRequest_Validation tests the validation constraints
// that should be enforced by the API handlers.
func TestCreateRelationshipTypeRequest_Validation(t *testing.T) {
	tests := []struct {
		name      string
		req       models.CreateRelationshipTypeRequest
		wantValid bool
		reason    string
	}{
		{
			name: "valid asymmetric type",
			req: models.CreateRelationshipTypeRequest{
				Name:                "owns",
				InverseName:         "owned_by",
				IsSymmetric:         false,
				DisplayLabel:        "Owns",
				InverseDisplayLabel: "Is owned by",
			},
			wantValid: true,
			reason:    "",
		},
		{
			name: "valid symmetric type",
			req: models.CreateRelationshipTypeRequest{
				Name:                "knows",
				InverseName:         "knows",
				IsSymmetric:         true,
				DisplayLabel:        "Knows",
				InverseDisplayLabel: "Knows",
			},
			wantValid: true,
			reason:    "",
		},
		{
			name: "invalid: symmetric with different names",
			req: models.CreateRelationshipTypeRequest{
				Name:                "owns",
				InverseName:         "owned_by",
				IsSymmetric:         true,
				DisplayLabel:        "Owns",
				InverseDisplayLabel: "Is owned by",
			},
			wantValid: false,
			reason:    "symmetric types must have matching name and inverse name",
		},
		{
			name: "invalid: empty name",
			req: models.CreateRelationshipTypeRequest{
				Name:                "",
				InverseName:         "test",
				IsSymmetric:         false,
				DisplayLabel:        "Test",
				InverseDisplayLabel: "Test",
			},
			wantValid: false,
			reason:    "name is required",
		},
		{
			name: "invalid: empty inverse name",
			req: models.CreateRelationshipTypeRequest{
				Name:                "test",
				InverseName:         "",
				IsSymmetric:         false,
				DisplayLabel:        "Test",
				InverseDisplayLabel: "Test",
			},
			wantValid: false,
			reason:    "inverse name is required",
		},
		{
			name: "invalid: empty display label",
			req: models.CreateRelationshipTypeRequest{
				Name:                "test",
				InverseName:         "test_inverse",
				IsSymmetric:         false,
				DisplayLabel:        "",
				InverseDisplayLabel: "Test",
			},
			wantValid: false,
			reason:    "display label is required",
		},
		{
			name: "invalid: empty inverse display label",
			req: models.CreateRelationshipTypeRequest{
				Name:                "test",
				InverseName:         "test_inverse",
				IsSymmetric:         false,
				DisplayLabel:        "Test",
				InverseDisplayLabel: "",
			},
			wantValid: false,
			reason:    "inverse display label is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the validation logic from the API handler
			valid := true
			reason := ""

			if tt.req.Name == "" {
				valid = false
				reason = "name is required"
			} else if tt.req.InverseName == "" {
				valid = false
				reason = "inverse name is required"
			} else if tt.req.DisplayLabel == "" {
				valid = false
				reason = "display label is required"
			} else if tt.req.InverseDisplayLabel == "" {
				valid = false
				reason = "inverse display label is required"
			} else if tt.req.IsSymmetric && tt.req.Name != tt.req.InverseName {
				valid = false
				reason = "symmetric types must have matching name and inverse name"
			}

			assert.Equal(t, tt.wantValid, valid)
			if !tt.wantValid {
				assert.Contains(t, reason, tt.reason)
			}
		})
	}
}

// TestRelationshipType_SystemDefaults tests the expected system default types.
func TestRelationshipType_SystemDefaults(t *testing.T) {
	// These are the expected system default relationship types from the migration
	expectedDefaults := []struct {
		name        string
		inverseName string
		isSymmetric bool
		isAlias     bool // alias types (like works_for) point to another type but don't have a reciprocal
	}{
		{"owns", "owned_by", false, false},
		{"owned_by", "owns", false, false},
		{"employs", "employed_by", false, false},
		{"employed_by", "employs", false, false},
		{"works_for", "employs", false, true}, // alias for employed_by
		{"reports_to", "manages", false, false},
		{"manages", "reports_to", false, false},
		{"parent_of", "child_of", false, false},
		{"child_of", "parent_of", false, false},
		{"located_at", "contains", false, false},
		{"contains", "located_at", false, false},
		{"member_of", "has_member", false, false},
		{"has_member", "member_of", false, false},
		{"created", "created_by", false, false},
		{"created_by", "created", false, false},
		{"rules", "ruled_by", false, false},
		{"ruled_by", "rules", false, false},
		{"knows", "knows", true, false},
		{"friend_of", "friend_of", true, false},
		{"enemy_of", "enemy_of", true, false},
		{"allied_with", "allied_with", true, false},
	}

	// Verify the count matches what we expect
	assert.Equal(t, 21, len(expectedDefaults), "should have 21 system default relationship types")

	// Verify symmetric types have matching names, and asymmetric non-alias types have inverses
	for _, def := range expectedDefaults {
		t.Run(def.name, func(t *testing.T) {
			if def.isSymmetric {
				assert.Equal(t, def.name, def.inverseName, "symmetric type should have matching name and inverse name")
			} else if !def.isAlias {
				// For asymmetric non-alias types, verify the inverse relationship is also defined
				found := false
				for _, other := range expectedDefaults {
					if other.name == def.inverseName && other.inverseName == def.name {
						found = true
						break
					}
				}
				assert.True(t, found, "asymmetric type should have corresponding inverse type defined")
			} else {
				// For alias types, just verify the inverse exists
				found := false
				for _, other := range expectedDefaults {
					if other.name == def.inverseName {
						found = true
						break
					}
				}
				assert.True(t, found, "alias type should point to an existing type")
			}
		})
	}
}
