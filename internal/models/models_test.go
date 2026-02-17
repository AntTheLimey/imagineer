/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGameSystem_JSONMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		system   GameSystem
		wantJSON map[string]interface{}
	}{
		{
			name: "full game system",
			system: GameSystem{
				ID:              1,
				Name:            "Call of Cthulhu 7e",
				Code:            "coc7e",
				AttributeSchema: json.RawMessage(`{"STR": {"min": 1, "max": 100}}`),
				SkillSchema:     json.RawMessage(`{"Library Use": {"base": 20}}`),
				CreatedAt:       time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
			},
			wantJSON: map[string]interface{}{
				"id":              float64(1),
				"name":            "Call of Cthulhu 7e",
				"code":            "coc7e",
				"attributeSchema": map[string]interface{}{"STR": map[string]interface{}{"min": float64(1), "max": float64(100)}},
				"skillSchema":     map[string]interface{}{"Library Use": map[string]interface{}{"base": float64(20)}},
				"createdAt":       "2025-01-15T10:00:00Z",
			},
		},
		{
			name: "minimal game system",
			system: GameSystem{
				ID:        2,
				Name:      "GURPS 4e",
				Code:      "gurps4e",
				CreatedAt: time.Date(2025, 2, 20, 14, 30, 0, 0, time.UTC),
			},
			wantJSON: map[string]interface{}{
				"id":        float64(2),
				"name":      "GURPS 4e",
				"code":      "gurps4e",
				"createdAt": "2025-02-20T14:30:00Z",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			data, err := json.Marshal(tt.system)
			require.NoError(t, err)

			// Unmarshal to map for comparison
			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			require.NoError(t, err)

			// Verify key fields
			assert.Equal(t, tt.wantJSON["id"], result["id"])
			assert.Equal(t, tt.wantJSON["name"], result["name"])
			assert.Equal(t, tt.wantJSON["code"], result["code"])

			// Round-trip test
			var roundTrip GameSystem
			err = json.Unmarshal(data, &roundTrip)
			require.NoError(t, err)
			assert.Equal(t, tt.system.ID, roundTrip.ID)
			assert.Equal(t, tt.system.Name, roundTrip.Name)
			assert.Equal(t, tt.system.Code, roundTrip.Code)
		})
	}
}

func TestCampaign_JSONMarshalUnmarshal(t *testing.T) {
	description := "A dark investigation in 1920s Arkham"
	systemID := int64(1)

	tests := []struct {
		name     string
		campaign Campaign
	}{
		{
			name: "full campaign",
			campaign: Campaign{
				ID:          3,
				Name:        "Masks of Nyarlathotep",
				SystemID:    &systemID,
				Description: &description,
				Settings:    json.RawMessage(`{"era": "1920s", "location": "Global"}`),
				CreatedAt:   time.Date(2025, 3, 10, 9, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2025, 3, 15, 11, 30, 0, 0, time.UTC),
			},
		},
		{
			name: "minimal campaign",
			campaign: Campaign{
				ID:        4,
				Name:      "Horror on the Orient Express",
				CreatedAt: time.Date(2025, 4, 1, 8, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2025, 4, 1, 8, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.campaign)
			require.NoError(t, err)

			// Unmarshal
			var result Campaign
			err = json.Unmarshal(data, &result)
			require.NoError(t, err)

			// Verify
			assert.Equal(t, tt.campaign.ID, result.ID)
			assert.Equal(t, tt.campaign.Name, result.Name)
			if tt.campaign.SystemID != nil {
				require.NotNil(t, result.SystemID)
				assert.Equal(t, *tt.campaign.SystemID, *result.SystemID)
			}
			if tt.campaign.Description != nil {
				require.NotNil(t, result.Description)
				assert.Equal(t, *tt.campaign.Description, *result.Description)
			}
		})
	}
}

func TestEntity_JSONMarshalUnmarshal(t *testing.T) {
	description := "A shady antiquarian with connections to the cult"
	gmNotes := "Hidden antagonist, reveal in session 5"

	tests := []struct {
		name   string
		entity Entity
	}{
		{
			name: "NPC entity",
			entity: Entity{
				ID:               5,
				CampaignID:       3,
				EntityType:       EntityTypeNPC,
				Name:             "Edward Gavigan",
				Description:      &description,
				Attributes:       json.RawMessage(`{"occupation": "Antiquarian", "age": 45}`),
				Tags:             []string{"antagonist", "cultist", "london"},
				GMNotes:          &gmNotes,
				SourceConfidence: SourceConfidenceAuthoritative,
				Version:          1,
				CreatedAt:        time.Date(2025, 5, 1, 10, 0, 0, 0, time.UTC),
				UpdatedAt:        time.Date(2025, 5, 2, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Location entity",
			entity: Entity{
				ID:               6,
				CampaignID:       3,
				EntityType:       EntityTypeLocation,
				Name:             "Miskatonic University",
				Attributes:       json.RawMessage(`{"city": "Arkham", "founded": 1690}`),
				Tags:             []string{"university", "arkham", "library"},
				SourceConfidence: SourceConfidenceAuthoritative,
				Version:          2,
				CreatedAt:        time.Date(2025, 5, 3, 11, 0, 0, 0, time.UTC),
				UpdatedAt:        time.Date(2025, 5, 10, 16, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.entity)
			require.NoError(t, err)

			// Unmarshal
			var result Entity
			err = json.Unmarshal(data, &result)
			require.NoError(t, err)

			// Verify
			assert.Equal(t, tt.entity.ID, result.ID)
			assert.Equal(t, tt.entity.CampaignID, result.CampaignID)
			assert.Equal(t, tt.entity.EntityType, result.EntityType)
			assert.Equal(t, tt.entity.Name, result.Name)
			assert.Equal(t, tt.entity.SourceConfidence, result.SourceConfidence)
			assert.Equal(t, tt.entity.Version, result.Version)
			assert.Equal(t, tt.entity.Tags, result.Tags)
		})
	}
}

func TestEntityType_Values(t *testing.T) {
	tests := []struct {
		entityType EntityType
		expected   string
	}{
		{EntityTypeNPC, "npc"},
		{EntityTypeLocation, "location"},
		{EntityTypeItem, "item"},
		{EntityTypeFaction, "faction"},
		{EntityTypeClue, "clue"},
		{EntityTypeCreature, "creature"},
		{EntityTypeOrganization, "organization"},
		{EntityTypeEvent, "event"},
		{EntityTypeDocument, "document"},
		{EntityTypeOther, "other"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.entityType))
		})
	}
}

func TestSourceConfidence_Values(t *testing.T) {
	tests := []struct {
		confidence SourceConfidence
		expected   string
	}{
		{SourceConfidenceDraft, "DRAFT"},
		{SourceConfidenceAuthoritative, "AUTHORITATIVE"},
		{SourceConfidenceSuperseded, "SUPERSEDED"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.confidence))
		})
	}
}

func TestRelationshipTone_Values(t *testing.T) {
	tests := []struct {
		tone     RelationshipTone
		expected string
	}{
		{RelationshipToneFriendly, "friendly"},
		{RelationshipToneHostile, "hostile"},
		{RelationshipToneNeutral, "neutral"},
		{RelationshipToneRomantic, "romantic"},
		{RelationshipToneProfessional, "professional"},
		{RelationshipToneFearful, "fearful"},
		{RelationshipToneRespectful, "respectful"},
		{RelationshipToneUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.tone))
		})
	}
}

func TestSessionStatus_Values(t *testing.T) {
	tests := []struct {
		status   SessionStatus
		expected string
	}{
		{SessionStatusPlanned, "PLANNED"},
		{SessionStatusCompleted, "COMPLETED"},
		{SessionStatusSkipped, "SKIPPED"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.status))
		})
	}
}

func TestDatePrecision_Values(t *testing.T) {
	tests := []struct {
		precision DatePrecision
		expected  string
	}{
		{DatePrecisionExact, "exact"},
		{DatePrecisionApproximate, "approximate"},
		{DatePrecisionMonth, "month"},
		{DatePrecisionYear, "year"},
		{DatePrecisionUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.precision))
		})
	}
}

func TestConflictStatus_Values(t *testing.T) {
	tests := []struct {
		status   ConflictStatus
		expected string
	}{
		{ConflictStatusDetected, "DETECTED"},
		{ConflictStatusAcknowledged, "ACKNOWLEDGED"},
		{ConflictStatusResolved, "RESOLVED"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.status))
		})
	}
}

func TestAPIError_JSONMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		apiError APIError
	}{
		{
			name: "error with details",
			apiError: APIError{
				Code:    400,
				Message: "Invalid request",
				Details: "Name field is required",
			},
		},
		{
			name: "error without details",
			apiError: APIError{
				Code:    404,
				Message: "Not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.apiError)
			require.NoError(t, err)

			// Unmarshal
			var result APIError
			err = json.Unmarshal(data, &result)
			require.NoError(t, err)

			// Verify
			assert.Equal(t, tt.apiError.Code, result.Code)
			assert.Equal(t, tt.apiError.Message, result.Message)
			assert.Equal(t, tt.apiError.Details, result.Details)
		})
	}
}

func TestCreateCampaignRequest_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected CreateCampaignRequest
		wantErr  bool
	}{
		{
			name:     "valid request with name only",
			jsonData: `{"name": "Test Campaign"}`,
			expected: CreateCampaignRequest{
				Name: "Test Campaign",
			},
		},
		{
			name:     "valid request with all fields",
			jsonData: `{"name": "Full Campaign", "systemId": 1, "description": "A test description", "settings": {"era": "1920s"}}`,
			expected: CreateCampaignRequest{
				Name: "Full Campaign",
			},
		},
		{
			name:     "invalid JSON",
			jsonData: `{"name": }`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result CreateCampaignRequest
			err := json.Unmarshal([]byte(tt.jsonData), &result)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected.Name, result.Name)
		})
	}
}

func TestCreateEntityRequest_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected CreateEntityRequest
		wantErr  bool
	}{
		{
			name:     "valid NPC request",
			jsonData: `{"entityType": "npc", "name": "Test NPC"}`,
			expected: CreateEntityRequest{
				EntityType: EntityTypeNPC,
				Name:       "Test NPC",
			},
		},
		{
			name:     "valid location request with tags",
			jsonData: `{"entityType": "location", "name": "Test Location", "tags": ["city", "arkham"]}`,
			expected: CreateEntityRequest{
				EntityType: EntityTypeLocation,
				Name:       "Test Location",
				Tags:       []string{"city", "arkham"},
			},
		},
		{
			name:     "invalid JSON",
			jsonData: `{"entityType": "npc", "name": `,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result CreateEntityRequest
			err := json.Unmarshal([]byte(tt.jsonData), &result)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected.EntityType, result.EntityType)
			assert.Equal(t, tt.expected.Name, result.Name)
			if tt.expected.Tags != nil {
				assert.Equal(t, tt.expected.Tags, result.Tags)
			}
		})
	}
}

func TestDashboardStats_JSONMarshalUnmarshal(t *testing.T) {
	stats := DashboardStats{
		TotalCampaigns:     5,
		TotalEntities:      150,
		TotalRelationships: 75,
		TotalSessions:      30,
		EntitiesByType: map[string]int{
			"npc":      50,
			"location": 40,
			"item":     30,
			"clue":     30,
		},
		RecentCampaigns: []Campaign{
			{
				ID:        7,
				Name:      "Recent Campaign",
				CreatedAt: time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC),
			},
		},
	}

	// Marshal
	data, err := json.Marshal(stats)
	require.NoError(t, err)

	// Unmarshal
	var result DashboardStats
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, stats.TotalCampaigns, result.TotalCampaigns)
	assert.Equal(t, stats.TotalEntities, result.TotalEntities)
	assert.Equal(t, stats.TotalRelationships, result.TotalRelationships)
	assert.Equal(t, stats.TotalSessions, result.TotalSessions)
	assert.Equal(t, stats.EntitiesByType, result.EntitiesByType)
	assert.Len(t, result.RecentCampaigns, 1)
	assert.Equal(t, stats.RecentCampaigns[0].Name, result.RecentCampaigns[0].Name)
}

func TestFrontendDashboardStats_JSONMarshalUnmarshal(t *testing.T) {
	stats := FrontendDashboardStats{
		CampaignCount:      5,
		NPCCount:           50,
		LocationCount:      40,
		TimelineEventCount: 20,
		ItemCount:          30,
		FactionCount:       10,
		TotalEntityCount:   150,
	}

	// Marshal
	data, err := json.Marshal(stats)
	require.NoError(t, err)

	// Unmarshal
	var result FrontendDashboardStats
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, stats, result)
}

func TestSessionStage_Values(t *testing.T) {
	tests := []struct {
		stage    SessionStage
		expected string
	}{
		{SessionStagePrep, "prep"},
		{SessionStagePlay, "play"},
		{SessionStageWrapUp, "wrap_up"},
		{SessionStageCompleted, "completed"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.stage))
		})
	}
}

func TestSession_JSONMarshalUnmarshal(t *testing.T) {
	playNotes := "The players chose to investigate the warehouse"
	prepNotes := "Set up the docks encounter"
	title := "Session 5: The Docks"
	chapterID := int64(2)

	session := Session{
		ID:         10,
		CampaignID: 3,
		ChapterID:  &chapterID,
		Title:      &title,
		Status:     SessionStatusPlanned,
		Stage:      SessionStagePrep,
		PrepNotes:  &prepNotes,
		PlayNotes:  &playNotes,
		CreatedAt:  time.Date(2025, 7, 1, 10, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2025, 7, 1, 10, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(session)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	// Verify new field is present
	assert.Equal(t, playNotes, result["playNotes"])

	// Verify removed fields are NOT present
	assert.NotContains(t, result, "plannedScenes")
	assert.NotContains(t, result, "discoveries")
	assert.NotContains(t, result, "playerDecisions")
	assert.NotContains(t, result, "consequences")

	// Round-trip test
	var roundTrip Session
	err = json.Unmarshal(data, &roundTrip)
	require.NoError(t, err)
	assert.Equal(t, session.ID, roundTrip.ID)
	assert.Equal(t, session.Stage, roundTrip.Stage)
	assert.NotNil(t, roundTrip.PlayNotes)
	assert.Equal(t, playNotes, *roundTrip.PlayNotes)
}

func TestScene_JSONMarshalUnmarshal(t *testing.T) {
	description := "The investigators arrive at the docks"
	objective := "Find the hidden cargo"

	scene := Scene{
		ID:               1,
		SessionID:        10,
		CampaignID:       3,
		Title:            "Arrival at the Docks",
		Description:      &description,
		SceneType:        "exploration",
		Status:           "planned",
		SortOrder:        0,
		Objective:        &objective,
		EntityIDs:        []int64{5, 6, 7},
		SystemData:       json.RawMessage(`{"encounter_type": "investigation"}`),
		Source:           "manual",
		SourceConfidence: SourceConfidenceDraft,
		Connections:      json.RawMessage(`[]`),
		CreatedAt:        time.Date(2025, 7, 1, 10, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2025, 7, 1, 10, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(scene)
	require.NoError(t, err)

	var result Scene
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, scene.ID, result.ID)
	assert.Equal(t, scene.SessionID, result.SessionID)
	assert.Equal(t, scene.CampaignID, result.CampaignID)
	assert.Equal(t, scene.Title, result.Title)
	require.NotNil(t, result.Description)
	assert.Equal(t, *scene.Description, *result.Description)
	assert.Equal(t, scene.SceneType, result.SceneType)
	assert.Equal(t, scene.Status, result.Status)
	assert.Equal(t, scene.SortOrder, result.SortOrder)
	assert.Equal(t, scene.EntityIDs, result.EntityIDs)
	assert.Equal(t, scene.Source, result.Source)
	assert.Equal(t, scene.SourceConfidence, result.SourceConfidence)
}

func TestCreateSceneRequest_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected CreateSceneRequest
		wantErr  bool
	}{
		{
			name:     "minimal scene",
			jsonData: `{"title": "Test Scene"}`,
			expected: CreateSceneRequest{Title: "Test Scene"},
		},
		{
			name:     "full scene",
			jsonData: `{"title": "Full Scene", "sceneType": "combat", "description": "A fight!", "entityIds": [1, 2]}`,
			expected: CreateSceneRequest{
				Title:     "Full Scene",
				EntityIDs: []int64{1, 2},
			},
		},
		{
			name:     "invalid JSON",
			jsonData: `{"title": }`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result CreateSceneRequest
			err := json.Unmarshal([]byte(tt.jsonData), &result)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected.Title, result.Title)
			if tt.expected.EntityIDs != nil {
				assert.Equal(t, tt.expected.EntityIDs, result.EntityIDs)
			}
		})
	}
}

func TestSessionChatMessage_JSONMarshalUnmarshal(t *testing.T) {
	msg := SessionChatMessage{
		ID:         1,
		SessionID:  10,
		CampaignID: 3,
		Role:       "user",
		Content:    "What NPCs are at the docks?",
		SortOrder:  0,
		CreatedAt:  time.Date(2025, 7, 1, 10, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var result SessionChatMessage
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, msg.ID, result.ID)
	assert.Equal(t, msg.SessionID, result.SessionID)
	assert.Equal(t, msg.Role, result.Role)
	assert.Equal(t, msg.Content, result.Content)
	assert.Equal(t, msg.SortOrder, result.SortOrder)
}
