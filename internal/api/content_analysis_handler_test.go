/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/antonypegg/imagineer/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContentAnalysisHandler(t *testing.T) {
	handler := NewContentAnalysisHandler(nil)
	assert.NotNil(t, handler)
}

func TestContentAnalysisHandler_AuthEnforcement(t *testing.T) {
	handler := NewContentAnalysisHandler(nil)

	tests := []struct {
		name   string
		method string
		path   string
		route  string
	}{
		{
			name:   "ListJobs requires auth",
			method: http.MethodGet,
			path:   "/api/campaigns/1/analysis/jobs",
			route:  "/api/campaigns/{id}/analysis/jobs",
		},
		{
			name:   "GetJob requires auth",
			method: http.MethodGet,
			path:   "/api/campaigns/1/analysis/jobs/1",
			route:  "/api/campaigns/{id}/analysis/jobs/{jobId}",
		},
		{
			name:   "ListJobItems requires auth",
			method: http.MethodGet,
			path:   "/api/campaigns/1/analysis/jobs/1/items",
			route:  "/api/campaigns/{id}/analysis/jobs/{jobId}/items",
		},
		{
			name:   "ResolveItem requires auth",
			method: http.MethodPut,
			path:   "/api/campaigns/1/analysis/items/1",
			route:  "/api/campaigns/{id}/analysis/items/{itemId}",
		},
		{
			name:   "TriggerAnalysis requires auth",
			method: http.MethodPost,
			path:   "/api/campaigns/1/analysis/trigger",
			route:  "/api/campaigns/{id}/analysis/trigger",
		},
		{
			name:   "GetPendingCount requires auth",
			method: http.MethodGet,
			path:   "/api/campaigns/1/analysis/pending-count?sourceTable=entities&sourceId=1",
			route:  "/api/campaigns/{id}/analysis/pending-count",
		},
		{
			name:   "BatchResolve requires auth",
			method: http.MethodPut,
			path:   "/api/campaigns/1/analysis/jobs/1/resolve-all",
			route:  "/api/campaigns/{id}/analysis/jobs/{jobId}/resolve-all",
		},
		{
			name:   "RevertItem requires auth",
			method: http.MethodPut,
			path:   "/api/campaigns/1/analysis/items/1/revert",
			route:  "/api/campaigns/{id}/analysis/items/{itemId}/revert",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()

			switch tt.method {
			case http.MethodGet:
				switch {
				case tt.route == "/api/campaigns/{id}/analysis/jobs":
					r.Get(tt.route, handler.ListJobs)
				case tt.route == "/api/campaigns/{id}/analysis/jobs/{jobId}":
					r.Get(tt.route, handler.GetJob)
				case tt.route == "/api/campaigns/{id}/analysis/jobs/{jobId}/items":
					r.Get(tt.route, handler.ListJobItems)
				case tt.route == "/api/campaigns/{id}/analysis/pending-count":
					r.Get(tt.route, handler.GetPendingCount)
				}
			case http.MethodPut:
				switch {
				case tt.route == "/api/campaigns/{id}/analysis/jobs/{jobId}/resolve-all":
					r.Put(tt.route, handler.BatchResolve)
				case tt.route == "/api/campaigns/{id}/analysis/items/{itemId}/revert":
					r.Put(tt.route, handler.RevertItem)
				default:
					r.Put(tt.route, handler.ResolveItem)
				}
			case http.MethodPost:
				r.Post(tt.route, handler.TriggerAnalysis)
			}

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)

			var apiErr models.APIError
			err := json.Unmarshal(rec.Body.Bytes(), &apiErr)
			require.NoError(t, err)
			assert.Equal(t, http.StatusUnauthorized, apiErr.Code)
			assert.Contains(t, apiErr.Message, "Authentication required")
		})
	}
}

func TestContentAnalysisJob_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		job      models.ContentAnalysisJob
		expected map[string]interface{}
	}{
		{
			name: "completed job",
			job: models.ContentAnalysisJob{
				ID:            1,
				CampaignID:    10,
				SourceTable:   "entities",
				SourceID:      42,
				SourceField:   "description",
				Status:        "completed",
				TotalItems:    5,
				ResolvedItems: 3,
			},
			expected: map[string]interface{}{
				"id":            float64(1),
				"campaignId":    float64(10),
				"sourceTable":   "entities",
				"sourceId":      float64(42),
				"sourceField":   "description",
				"status":        "completed",
				"totalItems":    float64(5),
				"resolvedItems": float64(3),
			},
		},
		{
			name: "pending job with zero items",
			job: models.ContentAnalysisJob{
				ID:            2,
				CampaignID:    20,
				SourceTable:   "chapters",
				SourceID:      7,
				SourceField:   "overview",
				Status:        "pending",
				TotalItems:    0,
				ResolvedItems: 0,
			},
			expected: map[string]interface{}{
				"id":            float64(2),
				"campaignId":    float64(20),
				"sourceTable":   "chapters",
				"sourceId":      float64(7),
				"sourceField":   "overview",
				"status":        "pending",
				"totalItems":    float64(0),
				"resolvedItems": float64(0),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.job)
			require.NoError(t, err)

			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			require.NoError(t, err)

			assert.Equal(t, tt.expected["id"], result["id"])
			assert.Equal(t, tt.expected["campaignId"], result["campaignId"])
			assert.Equal(t, tt.expected["sourceTable"], result["sourceTable"])
			assert.Equal(t, tt.expected["sourceId"], result["sourceId"])
			assert.Equal(t, tt.expected["sourceField"], result["sourceField"])
			assert.Equal(t, tt.expected["status"], result["status"])
			assert.Equal(t, tt.expected["totalItems"], result["totalItems"])
			assert.Equal(t, tt.expected["resolvedItems"], result["resolvedItems"])
		})
	}
}

func TestContentAnalysisItem_JSONSerialization(t *testing.T) {
	entityID := int64(99)
	similarity := 0.87
	contextSnippet := "The cultists gathered at Innsmouth..."
	positionStart := 42
	positionEnd := 56
	resolvedEntityID := int64(100)
	entityName := "Deep Ones"
	entityType := models.EntityTypeNPC

	tests := []struct {
		name     string
		item     models.ContentAnalysisItem
		expected map[string]interface{}
	}{
		{
			name: "fully populated item",
			item: models.ContentAnalysisItem{
				ID:               1,
				JobID:            10,
				DetectionType:    "entity_mention",
				MatchedText:      "Deep Ones",
				EntityID:         &entityID,
				Similarity:       &similarity,
				ContextSnippet:   &contextSnippet,
				PositionStart:    &positionStart,
				PositionEnd:      &positionEnd,
				Resolution:       "accepted",
				ResolvedEntityID: &resolvedEntityID,
				EntityName:       &entityName,
				EntityType:       &entityType,
			},
			expected: map[string]interface{}{
				"id":               float64(1),
				"jobId":            float64(10),
				"detectionType":    "entity_mention",
				"matchedText":      "Deep Ones",
				"entityId":         float64(99),
				"similarity":       0.87,
				"contextSnippet":   "The cultists gathered at Innsmouth...",
				"positionStart":    float64(42),
				"positionEnd":      float64(56),
				"resolution":       "accepted",
				"resolvedEntityId": float64(100),
				"entityName":       "Deep Ones",
				"entityType":       "npc",
			},
		},
		{
			name: "item with nil optional fields",
			item: models.ContentAnalysisItem{
				ID:            2,
				JobID:         20,
				DetectionType: "new_entity",
				MatchedText:   "Unknown Cultist",
				Resolution:    "pending",
			},
			expected: map[string]interface{}{
				"id":            float64(2),
				"jobId":         float64(20),
				"detectionType": "new_entity",
				"matchedText":   "Unknown Cultist",
				"resolution":    "pending",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.item)
			require.NoError(t, err)

			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			require.NoError(t, err)

			assert.Equal(t, tt.expected["id"], result["id"])
			assert.Equal(t, tt.expected["jobId"], result["jobId"])
			assert.Equal(t, tt.expected["detectionType"], result["detectionType"])
			assert.Equal(t, tt.expected["matchedText"], result["matchedText"])
			assert.Equal(t, tt.expected["resolution"], result["resolution"])

			// Check optional fields: present when populated, absent when nil
			if tt.expected["entityId"] != nil {
				assert.Equal(t, tt.expected["entityId"], result["entityId"])
			} else {
				assert.Nil(t, result["entityId"])
			}

			if tt.expected["similarity"] != nil {
				assert.Equal(t, tt.expected["similarity"], result["similarity"])
			} else {
				assert.Nil(t, result["similarity"])
			}

			if tt.expected["contextSnippet"] != nil {
				assert.Equal(t, tt.expected["contextSnippet"], result["contextSnippet"])
			} else {
				assert.Nil(t, result["contextSnippet"])
			}

			if tt.expected["positionStart"] != nil {
				assert.Equal(t, tt.expected["positionStart"], result["positionStart"])
			} else {
				assert.Nil(t, result["positionStart"])
			}

			if tt.expected["positionEnd"] != nil {
				assert.Equal(t, tt.expected["positionEnd"], result["positionEnd"])
			} else {
				assert.Nil(t, result["positionEnd"])
			}

			if tt.expected["resolvedEntityId"] != nil {
				assert.Equal(t, tt.expected["resolvedEntityId"], result["resolvedEntityId"])
			} else {
				assert.Nil(t, result["resolvedEntityId"])
			}

			if tt.expected["entityName"] != nil {
				assert.Equal(t, tt.expected["entityName"], result["entityName"])
			} else {
				assert.Nil(t, result["entityName"])
			}

			if tt.expected["entityType"] != nil {
				assert.Equal(t, tt.expected["entityType"], result["entityType"])
			} else {
				assert.Nil(t, result["entityType"])
			}
		})
	}
}

func TestAnalysisSummary_JSONSerialization(t *testing.T) {
	summary := models.AnalysisSummary{
		JobID:        42,
		PendingCount: 7,
	}

	data, err := json.Marshal(summary)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, float64(42), result["jobId"])
	assert.Equal(t, float64(7), result["pendingCount"])
}

func TestContentAnalysis_RoutesRegistered(t *testing.T) {
	router, err := NewRouter(nil, nil, testJWTSecret)
	require.NoError(t, err)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "ListJobs route registered",
			method: http.MethodGet,
			path:   "/api/campaigns/1/analysis/jobs",
		},
		{
			name:   "GetJob route registered",
			method: http.MethodGet,
			path:   "/api/campaigns/1/analysis/jobs/1",
		},
		{
			name:   "ListJobItems route registered",
			method: http.MethodGet,
			path:   "/api/campaigns/1/analysis/jobs/1/items",
		},
		{
			name:   "ResolveItem route registered",
			method: http.MethodPut,
			path:   "/api/campaigns/1/analysis/items/1",
		},
		{
			name:   "TriggerAnalysis route registered",
			method: http.MethodPost,
			path:   "/api/campaigns/1/analysis/trigger",
		},
		{
			name:   "GetPendingCount route registered",
			method: http.MethodGet,
			path:   "/api/campaigns/1/analysis/pending-count?sourceTable=entities&sourceId=1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Authorization", "Bearer invalid-token")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			// We expect 401 (unauthorized) because the token is invalid,
			// but NOT 404 (route not found) or 405 (method not allowed).
			assert.Equal(t, http.StatusUnauthorized, rec.Code,
				"Expected 401 for %s %s, got %d", tt.method, tt.path, rec.Code)
		})
	}
}
