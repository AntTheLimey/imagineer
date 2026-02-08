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
	"time"

	"github.com/antonypegg/imagineer/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestContentAnalysisJob_JSONMarshal verifies that ContentAnalysisJob can be
// properly marshaled to and unmarshaled from JSON with all fields preserved.
func TestContentAnalysisJob_JSONMarshal(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	original := models.ContentAnalysisJob{
		ID:            42,
		CampaignID:    7,
		SourceTable:   "chapters",
		SourceID:      15,
		SourceField:   "overview",
		Status:        "completed",
		TotalItems:    12,
		ResolvedItems: 8,
		CreatedAt:     now,
		UpdatedAt:     now.Add(5 * time.Minute),
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var result models.ContentAnalysisJob
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, original.ID, result.ID)
	assert.Equal(t, original.CampaignID, result.CampaignID)
	assert.Equal(t, original.SourceTable, result.SourceTable)
	assert.Equal(t, original.SourceID, result.SourceID)
	assert.Equal(t, original.SourceField, result.SourceField)
	assert.Equal(t, original.Status, result.Status)
	assert.Equal(t, original.TotalItems, result.TotalItems)
	assert.Equal(t, original.ResolvedItems, result.ResolvedItems)
	assert.True(t, original.CreatedAt.Equal(result.CreatedAt))
	assert.True(t, original.UpdatedAt.Equal(result.UpdatedAt))
}

// TestContentAnalysisJob_JSONMarshalFromString verifies that ContentAnalysisJob
// can be unmarshaled from raw JSON strings, simulating API input.
func TestContentAnalysisJob_JSONMarshalFromString(t *testing.T) {
	tests := []struct {
		name           string
		jsonData       string
		wantStatus     string
		wantTotal      int
		wantResolved   int
		wantSourceTbl  string
		wantSourceFld  string
	}{
		{
			name:          "pending job with no items",
			jsonData:      `{"id":1,"campaignId":2,"sourceTable":"sessions","sourceId":10,"sourceField":"actualNotes","status":"pending","totalItems":0,"resolvedItems":0,"createdAt":"2025-06-01T12:00:00Z","updatedAt":"2025-06-01T12:00:00Z"}`,
			wantStatus:    "pending",
			wantTotal:     0,
			wantResolved:  0,
			wantSourceTbl: "sessions",
			wantSourceFld: "actualNotes",
		},
		{
			name:          "completed job with all items resolved",
			jsonData:      `{"id":5,"campaignId":3,"sourceTable":"chapters","sourceId":7,"sourceField":"overview","status":"completed","totalItems":6,"resolvedItems":6,"createdAt":"2025-06-01T12:00:00Z","updatedAt":"2025-06-01T12:05:00Z"}`,
			wantStatus:    "completed",
			wantTotal:     6,
			wantResolved:  6,
			wantSourceTbl: "chapters",
			wantSourceFld: "overview",
		},
		{
			name:          "in-progress job with partial resolution",
			jsonData:      `{"id":10,"campaignId":1,"sourceTable":"sessions","sourceId":3,"sourceField":"prepNotes","status":"in_progress","totalItems":10,"resolvedItems":4,"createdAt":"2025-06-01T12:00:00Z","updatedAt":"2025-06-01T12:02:00Z"}`,
			wantStatus:    "in_progress",
			wantTotal:     10,
			wantResolved:  4,
			wantSourceTbl: "sessions",
			wantSourceFld: "prepNotes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var job models.ContentAnalysisJob
			err := json.Unmarshal([]byte(tt.jsonData), &job)
			require.NoError(t, err)

			assert.Equal(t, tt.wantStatus, job.Status)
			assert.Equal(t, tt.wantTotal, job.TotalItems)
			assert.Equal(t, tt.wantResolved, job.ResolvedItems)
			assert.Equal(t, tt.wantSourceTbl, job.SourceTable)
			assert.Equal(t, tt.wantSourceFld, job.SourceField)
		})
	}
}

// TestContentAnalysisItem_JSONMarshal verifies that ContentAnalysisItem can be
// properly marshaled to and unmarshaled from JSON, including nullable fields
// such as EntityID, Similarity, ContextSnippet, and position fields.
func TestContentAnalysisItem_JSONMarshal(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	resolvedAt := now.Add(3 * time.Minute)
	entityID := int64(99)
	similarity := 0.85
	context := "...the cultist approached the [[Arkham Library]]..."
	posStart := 25
	posEnd := 40
	resolvedEntityID := int64(99)

	original := models.ContentAnalysisItem{
		ID:               1,
		JobID:            42,
		DetectionType:    "wiki_link_resolved",
		MatchedText:      "Arkham Library",
		EntityID:         &entityID,
		Similarity:       &similarity,
		ContextSnippet:   &context,
		PositionStart:    &posStart,
		PositionEnd:      &posEnd,
		Resolution:       "accepted",
		ResolvedEntityID: &resolvedEntityID,
		ResolvedAt:       &resolvedAt,
		CreatedAt:        now,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var result models.ContentAnalysisItem
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, original.ID, result.ID)
	assert.Equal(t, original.JobID, result.JobID)
	assert.Equal(t, original.DetectionType, result.DetectionType)
	assert.Equal(t, original.MatchedText, result.MatchedText)
	require.NotNil(t, result.EntityID)
	assert.Equal(t, *original.EntityID, *result.EntityID)
	require.NotNil(t, result.Similarity)
	assert.Equal(t, *original.Similarity, *result.Similarity)
	require.NotNil(t, result.ContextSnippet)
	assert.Equal(t, *original.ContextSnippet, *result.ContextSnippet)
	require.NotNil(t, result.PositionStart)
	assert.Equal(t, *original.PositionStart, *result.PositionStart)
	require.NotNil(t, result.PositionEnd)
	assert.Equal(t, *original.PositionEnd, *result.PositionEnd)
	assert.Equal(t, original.Resolution, result.Resolution)
	require.NotNil(t, result.ResolvedEntityID)
	assert.Equal(t, *original.ResolvedEntityID, *result.ResolvedEntityID)
	require.NotNil(t, result.ResolvedAt)
	assert.True(t, original.ResolvedAt.Equal(*result.ResolvedAt))
	assert.True(t, original.CreatedAt.Equal(result.CreatedAt))
}

// TestContentAnalysisItem_JoinedFields verifies that the EntityName and
// EntityType joined fields serialize properly when populated. These fields
// are not stored in the database but are joined from the entities table.
func TestContentAnalysisItem_JoinedFields(t *testing.T) {
	entityName := "Professor Armitage"
	entityType := models.EntityTypeNPC

	item := models.ContentAnalysisItem{
		ID:            5,
		JobID:         10,
		DetectionType: "untagged_mention",
		MatchedText:   "Professor Armitage",
		Resolution:    "pending",
		CreatedAt:     time.Now().Truncate(time.Second),
		EntityName:    &entityName,
		EntityType:    &entityType,
	}

	data, err := json.Marshal(item)
	require.NoError(t, err)

	// Verify the joined fields are present in the JSON output.
	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	assert.Equal(t, "Professor Armitage", raw["entityName"])
	assert.Equal(t, "npc", raw["entityType"])

	// Round-trip the struct and verify joined fields survive.
	var result models.ContentAnalysisItem
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	require.NotNil(t, result.EntityName)
	assert.Equal(t, "Professor Armitage", *result.EntityName)
	require.NotNil(t, result.EntityType)
	assert.Equal(t, models.EntityTypeNPC, *result.EntityType)
}

// TestContentAnalysisItem_JoinedFieldsAllEntityTypes verifies that the
// EntityType joined field correctly handles all known entity types.
func TestContentAnalysisItem_JoinedFieldsAllEntityTypes(t *testing.T) {
	entityTypes := []models.EntityType{
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

	for _, et := range entityTypes {
		t.Run(string(et), func(t *testing.T) {
			entityName := "Test Entity"
			entityType := et

			item := models.ContentAnalysisItem{
				ID:            1,
				JobID:         1,
				DetectionType: "wiki_link_resolved",
				MatchedText:   entityName,
				Resolution:    "accepted",
				CreatedAt:     time.Now().Truncate(time.Second),
				EntityName:    &entityName,
				EntityType:    &entityType,
			}

			data, err := json.Marshal(item)
			require.NoError(t, err)

			var result models.ContentAnalysisItem
			err = json.Unmarshal(data, &result)
			require.NoError(t, err)

			require.NotNil(t, result.EntityType)
			assert.Equal(t, et, *result.EntityType)
		})
	}
}

// TestResolveAnalysisItemRequest_JSONMarshal tests all resolution types:
// "pending", "accepted", "new_entity", and "dismissed".
func TestResolveAnalysisItemRequest_JSONMarshal(t *testing.T) {
	tests := []struct {
		name           string
		jsonData       string
		wantResolution string
		wantEntityType *models.EntityType
		wantEntityName *string
	}{
		{
			name:           "pending resolution",
			jsonData:       `{"resolution":"pending"}`,
			wantResolution: "pending",
			wantEntityType: nil,
			wantEntityName: nil,
		},
		{
			name:           "accepted resolution",
			jsonData:       `{"resolution":"accepted"}`,
			wantResolution: "accepted",
			wantEntityType: nil,
			wantEntityName: nil,
		},
		{
			name:           "new_entity resolution with entity details",
			jsonData:       `{"resolution":"new_entity","entityType":"npc","entityName":"Inspector Legrasse"}`,
			wantResolution: "new_entity",
			wantEntityType: entityTypePtr(models.EntityTypeNPC),
			wantEntityName: stringPtr("Inspector Legrasse"),
		},
		{
			name:           "dismissed resolution",
			jsonData:       `{"resolution":"dismissed"}`,
			wantResolution: "dismissed",
			wantEntityType: nil,
			wantEntityName: nil,
		},
		{
			name:           "new_entity resolution with location type",
			jsonData:       `{"resolution":"new_entity","entityType":"location","entityName":"R'lyeh"}`,
			wantResolution: "new_entity",
			wantEntityType: entityTypePtr(models.EntityTypeLocation),
			wantEntityName: stringPtr("R'lyeh"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req models.ResolveAnalysisItemRequest
			err := json.Unmarshal([]byte(tt.jsonData), &req)
			require.NoError(t, err)

			assert.Equal(t, tt.wantResolution, req.Resolution)

			if tt.wantEntityType != nil {
				require.NotNil(t, req.EntityType)
				assert.Equal(t, *tt.wantEntityType, *req.EntityType)
			} else {
				assert.Nil(t, req.EntityType)
			}

			if tt.wantEntityName != nil {
				require.NotNil(t, req.EntityName)
				assert.Equal(t, *tt.wantEntityName, *req.EntityName)
			} else {
				assert.Nil(t, req.EntityName)
			}
		})
	}
}

// TestResolveAnalysisItemRequest_RoundTrip verifies that a fully populated
// ResolveAnalysisItemRequest survives marshal and unmarshal without data loss.
func TestResolveAnalysisItemRequest_RoundTrip(t *testing.T) {
	entityType := models.EntityTypeFaction
	entityName := "The Order of the Silver Twilight"

	original := models.ResolveAnalysisItemRequest{
		Resolution: "new_entity",
		EntityType: &entityType,
		EntityName: &entityName,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var result models.ResolveAnalysisItemRequest
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, original.Resolution, result.Resolution)
	require.NotNil(t, result.EntityType)
	assert.Equal(t, *original.EntityType, *result.EntityType)
	require.NotNil(t, result.EntityName)
	assert.Equal(t, *original.EntityName, *result.EntityName)
}

// TestAnalysisSummary_JSONMarshal verifies that AnalysisSummary marshals and
// unmarshals correctly, preserving the job ID and pending count.
func TestAnalysisSummary_JSONMarshal(t *testing.T) {
	tests := []struct {
		name             string
		summary          models.AnalysisSummary
		wantJobID        int64
		wantPendingCount int
	}{
		{
			name: "no pending items",
			summary: models.AnalysisSummary{
				JobID:        1,
				PendingCount: 0,
			},
			wantJobID:        1,
			wantPendingCount: 0,
		},
		{
			name: "several pending items",
			summary: models.AnalysisSummary{
				JobID:        55,
				PendingCount: 12,
			},
			wantJobID:        55,
			wantPendingCount: 12,
		},
		{
			name: "large pending count",
			summary: models.AnalysisSummary{
				JobID:        100,
				PendingCount: 500,
			},
			wantJobID:        100,
			wantPendingCount: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.summary)
			require.NoError(t, err)

			var result models.AnalysisSummary
			err = json.Unmarshal(data, &result)
			require.NoError(t, err)

			assert.Equal(t, tt.wantJobID, result.JobID)
			assert.Equal(t, tt.wantPendingCount, result.PendingCount)
		})
	}
}

// TestAnalysisSummary_JSONKeys verifies the exact JSON field names used by
// AnalysisSummary, ensuring API contract compatibility.
func TestAnalysisSummary_JSONKeys(t *testing.T) {
	summary := models.AnalysisSummary{
		JobID:        42,
		PendingCount: 7,
	}

	data, err := json.Marshal(summary)
	require.NoError(t, err)

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	assert.Contains(t, raw, "jobId")
	assert.Contains(t, raw, "pendingCount")
	assert.Equal(t, float64(42), raw["jobId"])
	assert.Equal(t, float64(7), raw["pendingCount"])
}

// TestContentAnalysisItem_NullableFields verifies that items with nil optional
// fields marshal without those fields present in the JSON output, confirming
// correct omitempty behavior.
func TestContentAnalysisItem_NullableFields(t *testing.T) {
	tests := []struct {
		name           string
		item           models.ContentAnalysisItem
		absentFields   []string
		presentFields  []string
	}{
		{
			name: "all optional fields nil",
			item: models.ContentAnalysisItem{
				ID:            1,
				JobID:         10,
				DetectionType: "wiki_link_unresolved",
				MatchedText:   "Unknown Entity",
				Resolution:    "pending",
				CreatedAt:     time.Now().Truncate(time.Second),
			},
			absentFields: []string{
				"entityId", "similarity", "contextSnippet",
				"positionStart", "positionEnd",
				"resolvedEntityId", "resolvedAt",
				"entityName", "entityType",
			},
			presentFields: []string{
				"id", "jobId", "detectionType", "matchedText",
				"resolution", "createdAt",
			},
		},
		{
			name: "only entityId populated",
			item: models.ContentAnalysisItem{
				ID:            2,
				JobID:         10,
				DetectionType: "wiki_link_resolved",
				MatchedText:   "Arkham",
				EntityID:      int64Ptr(50),
				Resolution:    "accepted",
				CreatedAt:     time.Now().Truncate(time.Second),
			},
			absentFields: []string{
				"similarity", "contextSnippet",
				"positionStart", "positionEnd",
				"resolvedEntityId", "resolvedAt",
				"entityName", "entityType",
			},
			presentFields: []string{
				"id", "jobId", "detectionType", "matchedText",
				"entityId", "resolution", "createdAt",
			},
		},
		{
			name: "entityId and similarity populated",
			item: models.ContentAnalysisItem{
				ID:            3,
				JobID:         10,
				DetectionType: "untagged_mention",
				MatchedText:   "Dr. West",
				EntityID:      int64Ptr(77),
				Similarity:    float64Ptr(0.92),
				Resolution:    "pending",
				CreatedAt:     time.Now().Truncate(time.Second),
			},
			absentFields: []string{
				"contextSnippet",
				"positionStart", "positionEnd",
				"resolvedEntityId", "resolvedAt",
				"entityName", "entityType",
			},
			presentFields: []string{
				"id", "jobId", "detectionType", "matchedText",
				"entityId", "similarity", "resolution", "createdAt",
			},
		},
		{
			name: "joined fields populated without database fields",
			item: models.ContentAnalysisItem{
				ID:            4,
				JobID:         10,
				DetectionType: "wiki_link_resolved",
				MatchedText:   "Miskatonic University",
				EntityID:      int64Ptr(30),
				Resolution:    "accepted",
				CreatedAt:     time.Now().Truncate(time.Second),
				EntityName:    stringPtr("Miskatonic University"),
				EntityType:    entityTypePtr(models.EntityTypeLocation),
			},
			absentFields: []string{
				"similarity", "contextSnippet",
				"positionStart", "positionEnd",
				"resolvedEntityId", "resolvedAt",
			},
			presentFields: []string{
				"id", "jobId", "detectionType", "matchedText",
				"entityId", "resolution", "createdAt",
				"entityName", "entityType",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.item)
			require.NoError(t, err)

			var raw map[string]interface{}
			err = json.Unmarshal(data, &raw)
			require.NoError(t, err)

			for _, field := range tt.absentFields {
				assert.NotContains(t, raw, field,
					"field %q should be absent when nil", field)
			}

			for _, field := range tt.presentFields {
				assert.Contains(t, raw, field,
					"field %q should be present", field)
			}
		})
	}
}

// TestContentAnalysisItem_DetectionTypes verifies the four detection types are
// valid strings: "wiki_link_resolved", "wiki_link_unresolved",
// "untagged_mention", and "misspelling".
func TestContentAnalysisItem_DetectionTypes(t *testing.T) {
	detectionTypes := []struct {
		name  string
		value string
	}{
		{
			name:  "wiki_link_resolved",
			value: "wiki_link_resolved",
		},
		{
			name:  "wiki_link_unresolved",
			value: "wiki_link_unresolved",
		},
		{
			name:  "untagged_mention",
			value: "untagged_mention",
		},
		{
			name:  "misspelling",
			value: "misspelling",
		},
	}

	for _, dt := range detectionTypes {
		t.Run(dt.name, func(t *testing.T) {
			// Verify the type is a non-empty string.
			assert.NotEmpty(t, dt.value)

			// Verify the type can be used in a ContentAnalysisItem and
			// survives a JSON round-trip.
			item := models.ContentAnalysisItem{
				ID:            1,
				JobID:         1,
				DetectionType: dt.value,
				MatchedText:   "Test Detection",
				Resolution:    "pending",
				CreatedAt:     time.Now().Truncate(time.Second),
			}

			data, err := json.Marshal(item)
			require.NoError(t, err)

			var result models.ContentAnalysisItem
			err = json.Unmarshal(data, &result)
			require.NoError(t, err)

			assert.Equal(t, dt.value, result.DetectionType)
		})
	}
}

// TestContentAnalysisItem_DetectionTypesInJSON verifies that detection types
// are correctly represented in the JSON output.
func TestContentAnalysisItem_DetectionTypesInJSON(t *testing.T) {
	detectionTypes := []string{
		"wiki_link_resolved",
		"wiki_link_unresolved",
		"untagged_mention",
		"misspelling",
	}

	for _, dt := range detectionTypes {
		t.Run(dt, func(t *testing.T) {
			jsonStr := `{"id":1,"jobId":1,"detectionType":"` + dt +
				`","matchedText":"Test","resolution":"pending",` +
				`"createdAt":"2025-06-01T12:00:00Z"}`

			var item models.ContentAnalysisItem
			err := json.Unmarshal([]byte(jsonStr), &item)
			require.NoError(t, err)

			assert.Equal(t, dt, item.DetectionType)
		})
	}
}

// ---------- helper functions ----------

// stringPtr returns a pointer to the given string value.
func stringPtr(s string) *string {
	return &s
}

// int64Ptr returns a pointer to the given int64 value.
func int64Ptr(i int64) *int64 {
	return &i
}

// float64Ptr returns a pointer to the given float64 value.
func float64Ptr(f float64) *float64 {
	return &f
}

// entityTypePtr returns a pointer to the given EntityType value.
func entityTypePtr(et models.EntityType) *models.EntityType {
	return &et
}
