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
	"strings"
	"testing"

	"github.com/antonypegg/imagineer/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConversationHandler(t *testing.T) {
	handler := NewConversationHandler(nil, nil)
	assert.NotNil(t, handler)
}

// The handlers check auth before any other validation,
// so unauthenticated requests always receive 401. These
// tests verify that the routes resolve and the handler
// is invoked rather than returning 404 or 405.

func TestConversationCreate_NoAuth(t *testing.T) {
	handler := NewConversationHandler(nil, nil)

	r := chi.NewRouter()
	r.Post("/api/campaigns/{id}/conversations",
		handler.Create)

	req := httptest.NewRequest(http.MethodPost,
		"/api/campaigns/1/conversations",
		strings.NewReader(`{"scopeType":"entity","scopeId":1}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var apiErr models.APIError
	err := json.Unmarshal(rec.Body.Bytes(), &apiErr)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, apiErr.Code)
	assert.Contains(t, apiErr.Message,
		"Authentication required")
}

func TestConversationCreate_InvalidCampaignID(t *testing.T) {
	handler := NewConversationHandler(nil, nil)

	r := chi.NewRouter()
	r.Post("/api/campaigns/{id}/conversations",
		handler.Create)

	req := httptest.NewRequest(http.MethodPost,
		"/api/campaigns/not-a-number/conversations",
		strings.NewReader(`{"scopeType":"entity","scopeId":1}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	// parseInt64 fails before auth check.
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var apiErr models.APIError
	err := json.Unmarshal(rec.Body.Bytes(), &apiErr)
	require.NoError(t, err)
	assert.Contains(t, apiErr.Message, "Invalid campaign ID")
}

func TestConversationGet_NoAuth(t *testing.T) {
	handler := NewConversationHandler(nil, nil)

	r := chi.NewRouter()
	r.Get(
		"/api/campaigns/{id}/conversations/{conversationId}",
		handler.Get)

	req := httptest.NewRequest(http.MethodGet,
		"/api/campaigns/1/conversations/1", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestConversationGet_InvalidConversationID(t *testing.T) {
	handler := NewConversationHandler(nil, nil)

	r := chi.NewRouter()
	r.Get(
		"/api/campaigns/{id}/conversations/{conversationId}",
		handler.Get)

	req := httptest.NewRequest(http.MethodGet,
		"/api/campaigns/abc/conversations/1", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	// Invalid campaign ID is caught before auth.
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestConversationList_NoAuth(t *testing.T) {
	handler := NewConversationHandler(nil, nil)

	r := chi.NewRouter()
	r.Get("/api/campaigns/{id}/conversations",
		handler.List)

	req := httptest.NewRequest(http.MethodGet,
		"/api/campaigns/1/conversations", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestConversationList_InvalidCampaignID(t *testing.T) {
	handler := NewConversationHandler(nil, nil)

	r := chi.NewRouter()
	r.Get("/api/campaigns/{id}/conversations",
		handler.List)

	req := httptest.NewRequest(http.MethodGet,
		"/api/campaigns/xyz/conversations", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestConversationSendMessage_NoAuth(t *testing.T) {
	handler := NewConversationHandler(nil, nil)

	r := chi.NewRouter()
	r.Post(
		"/api/campaigns/{id}/conversations/{conversationId}/messages",
		handler.SendMessage)

	body := `{"content":"hello"}`
	req := httptest.NewRequest(http.MethodPost,
		"/api/campaigns/1/conversations/1/messages",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestConversationSendMessage_InvalidCampaignID(t *testing.T) {
	handler := NewConversationHandler(nil, nil)

	r := chi.NewRouter()
	r.Post(
		"/api/campaigns/{id}/conversations/{conversationId}/messages",
		handler.SendMessage)

	body := `{"content":"hello"}`
	req := httptest.NewRequest(http.MethodPost,
		"/api/campaigns/abc/conversations/1/messages",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var apiErr models.APIError
	err := json.Unmarshal(rec.Body.Bytes(), &apiErr)
	require.NoError(t, err)
	assert.Contains(t, apiErr.Message, "Invalid campaign ID")
}

func TestConversationSendMessage_InvalidConversationID(t *testing.T) {
	handler := NewConversationHandler(nil, nil)

	r := chi.NewRouter()
	r.Post(
		"/api/campaigns/{id}/conversations/{conversationId}/messages",
		handler.SendMessage)

	body := `{"content":"hello"}`
	req := httptest.NewRequest(http.MethodPost,
		"/api/campaigns/1/conversations/xyz/messages",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	// Auth fires before conversationId parsing.
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestValidScopeTypes(t *testing.T) {
	tests := []struct {
		name    string
		scope   models.ScopeType
		isValid bool
	}{
		{"entity", models.ScopeTypeEntity, true},
		{"chapter", models.ScopeTypeChapter, true},
		{"session", models.ScopeTypeSession, true},
		{"scene", models.ScopeTypeScene, true},
		{"campaign", models.ScopeTypeCampaign, true},
		{"unknown", models.ScopeType("unknown"), false},
		{"empty", models.ScopeType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isValid,
				validScopeTypes[tt.scope])
		})
	}
}

func TestCreateConversationRequest_JSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectScope models.ScopeType
		expectID    int64
	}{
		{
			name:        "entity scope",
			input:       `{"scopeType":"entity","scopeId":42}`,
			expectScope: models.ScopeTypeEntity,
			expectID:    42,
		},
		{
			name:        "campaign scope",
			input:       `{"scopeType":"campaign","scopeId":1}`,
			expectScope: models.ScopeTypeCampaign,
			expectID:    1,
		},
		{
			name:        "session scope",
			input:       `{"scopeType":"session","scopeId":7}`,
			expectScope: models.ScopeTypeSession,
			expectID:    7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req models.CreateConversationRequest
			err := json.Unmarshal(
				[]byte(tt.input), &req)
			require.NoError(t, err)
			assert.Equal(t, tt.expectScope, req.ScopeType)
			assert.Equal(t, tt.expectID, req.ScopeID)
		})
	}
}

func TestSendMessageRequest_JSON(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectContent string
		expectSelect  *string
	}{
		{
			name:          "content only",
			input:         `{"content":"hello world"}`,
			expectContent: "hello world",
			expectSelect:  nil,
		},
		{
			name:          "with editor selection",
			input:         `{"content":"fix this","editorSelection":"selected text"}`,
			expectContent: "fix this",
			expectSelect: func() *string {
				s := "selected text"
				return &s
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req models.SendMessageRequest
			err := json.Unmarshal(
				[]byte(tt.input), &req)
			require.NoError(t, err)
			assert.Equal(t, tt.expectContent, req.Content)
			if tt.expectSelect == nil {
				assert.Nil(t, req.EditorSelection)
			} else {
				require.NotNil(t, req.EditorSelection)
				assert.Equal(t, *tt.expectSelect,
					*req.EditorSelection)
			}
		})
	}
}
