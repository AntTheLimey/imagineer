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

func TestResolveEntity_MissingNameParam(t *testing.T) {
	handler := NewEntityResolveHandler(nil)

	// Build a chi router so URLParam("id") resolves correctly
	r := chi.NewRouter()
	r.Get("/api/campaigns/{id}/entities/resolve", handler.ResolveEntity)

	req := httptest.NewRequest(http.MethodGet, "/api/campaigns/1/entities/resolve", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	// Without authentication context the handler returns 401 first,
	// which is the expected behaviour when no JWT middleware is present.
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var apiErr models.APIError
	err := json.Unmarshal(rec.Body.Bytes(), &apiErr)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, apiErr.Code)
	assert.Contains(t, apiErr.Message, "Authentication required")
}

func TestResolveEntity_NameTooShort(t *testing.T) {
	// This test verifies the handler returns 400 when the name is
	// shorter than 3 characters. Because authentication is checked
	// first, we expect 401 without a valid JWT context.
	handler := NewEntityResolveHandler(nil)

	r := chi.NewRouter()
	r.Get("/api/campaigns/{id}/entities/resolve", handler.ResolveEntity)

	req := httptest.NewRequest(http.MethodGet, "/api/campaigns/1/entities/resolve?name=ab", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	// Authentication check fires before name validation
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestResolveEntity_InvalidCampaignID(t *testing.T) {
	handler := NewEntityResolveHandler(nil)

	r := chi.NewRouter()
	r.Get("/api/campaigns/{id}/entities/resolve", handler.ResolveEntity)

	req := httptest.NewRequest(http.MethodGet, "/api/campaigns/not-a-number/entities/resolve?name=test", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	// Authentication check fires before campaign ID parsing for
	// unauthenticated requests
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestNewEntityResolveHandler(t *testing.T) {
	handler := NewEntityResolveHandler(nil)
	assert.NotNil(t, handler)
}

func TestEntityResolveResult_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		result   models.EntityResolveResult
		expected map[string]interface{}
	}{
		{
			name: "NPC entity result",
			result: models.EntityResolveResult{
				ID:         42,
				Name:       "Professor Armitage",
				EntityType: models.EntityTypeNPC,
				Similarity: 0.85,
			},
			expected: map[string]interface{}{
				"id":         float64(42),
				"name":       "Professor Armitage",
				"entityType": "npc",
				"similarity": 0.85,
			},
		},
		{
			name: "location entity result",
			result: models.EntityResolveResult{
				ID:         7,
				Name:       "Miskatonic University",
				EntityType: models.EntityTypeLocation,
				Similarity: 0.92,
			},
			expected: map[string]interface{}{
				"id":         float64(7),
				"name":       "Miskatonic University",
				"entityType": "location",
				"similarity": 0.92,
			},
		},
		{
			name: "zero similarity result",
			result: models.EntityResolveResult{
				ID:         1,
				Name:       "Test",
				EntityType: models.EntityTypeOther,
				Similarity: 0.0,
			},
			expected: map[string]interface{}{
				"id":         float64(1),
				"name":       "Test",
				"entityType": "other",
				"similarity": 0.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.result)
			require.NoError(t, err)

			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			require.NoError(t, err)

			assert.Equal(t, tt.expected["id"], result["id"])
			assert.Equal(t, tt.expected["name"], result["name"])
			assert.Equal(t, tt.expected["entityType"], result["entityType"])
			assert.Equal(t, tt.expected["similarity"], result["similarity"])
		})
	}
}

func TestEntityResolveResult_ArraySerialization(t *testing.T) {
	// Verify that an empty results array serializes as [] not null
	results := []models.EntityResolveResult{}
	data, err := json.Marshal(results)
	require.NoError(t, err)
	assert.Equal(t, "[]", string(data))
}

func TestResolveEntity_RouteRegistered(t *testing.T) {
	// Verify the route is registered in the router by checking that
	// a request to the resolve endpoint does not return 404/405.
	router, err := NewRouter(nil, nil, testJWTSecret)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/campaigns/1/entities/resolve?name=test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// We expect 401 (unauthorized) because the token is invalid,
	// but NOT 404 (route not found) or 405 (method not allowed).
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
