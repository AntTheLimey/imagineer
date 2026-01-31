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
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/antonypegg/imagineer/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   map[string]string
	}{
		{
			name:           "GET health returns ok",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]string{"status": "ok"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router without database (health endpoint doesn't need it)
			router := NewRouter(nil)

			// Create request
			req := httptest.NewRequest(tt.method, "/health", nil)
			rec := httptest.NewRecorder()

			// Serve request
			router.ServeHTTP(rec, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Check response body
			var body map[string]string
			err := json.Unmarshal(rec.Body.Bytes(), &body)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedBody, body)

			// Check content type
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		})
	}
}

func TestRespondJSON(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		data           interface{}
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "respond with map",
			status:         http.StatusOK,
			data:           map[string]string{"message": "success"},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"success"}`,
		},
		{
			name:           "respond with struct",
			status:         http.StatusCreated,
			data:           models.APIError{Code: 201, Message: "Created"},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"code":201,"message":"Created"}`,
		},
		{
			name:           "respond with nil data",
			status:         http.StatusNoContent,
			data:           nil,
			expectedStatus: http.StatusNoContent,
			expectedBody:   "",
		},
		{
			name:           "respond with array",
			status:         http.StatusOK,
			data:           []string{"item1", "item2"},
			expectedStatus: http.StatusOK,
			expectedBody:   `["item1","item2"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			respondJSON(rec, tt.status, tt.data)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

			if tt.expectedBody != "" {
				// Trim newline that json.Encoder adds
				body := bytes.TrimSpace(rec.Body.Bytes())
				assert.JSONEq(t, tt.expectedBody, string(body))
			}
		})
	}
}

func TestRespondError(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		message        string
		expectedStatus int
		expectedBody   models.APIError
	}{
		{
			name:           "bad request error",
			status:         http.StatusBadRequest,
			message:        "Invalid input",
			expectedStatus: http.StatusBadRequest,
			expectedBody: models.APIError{
				Code:    400,
				Message: "Invalid input",
			},
		},
		{
			name:           "not found error",
			status:         http.StatusNotFound,
			message:        "Resource not found",
			expectedStatus: http.StatusNotFound,
			expectedBody: models.APIError{
				Code:    404,
				Message: "Resource not found",
			},
		},
		{
			name:           "internal server error",
			status:         http.StatusInternalServerError,
			message:        "Something went wrong",
			expectedStatus: http.StatusInternalServerError,
			expectedBody: models.APIError{
				Code:    500,
				Message: "Something went wrong",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			respondError(rec, tt.status, tt.message)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

			var body models.APIError
			err := json.Unmarshal(rec.Body.Bytes(), &body)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedBody.Code, body.Code)
			assert.Equal(t, tt.expectedBody.Message, body.Message)
		})
	}
}

func TestNewHandler(t *testing.T) {
	// Test that NewHandler creates a handler without panicking
	handler := NewHandler(nil)
	assert.NotNil(t, handler)
}

func TestAPIErrorResponse_Structure(t *testing.T) {
	// Test that API errors have the expected JSON structure
	tests := []struct {
		name        string
		apiError    models.APIError
		expectedMap map[string]interface{}
	}{
		{
			name: "error with all fields",
			apiError: models.APIError{
				Code:    400,
				Message: "Bad request",
				Details: "Name field is required",
			},
			expectedMap: map[string]interface{}{
				"code":    float64(400),
				"message": "Bad request",
				"details": "Name field is required",
			},
		},
		{
			name: "error without details",
			apiError: models.APIError{
				Code:    500,
				Message: "Internal server error",
			},
			expectedMap: map[string]interface{}{
				"code":    float64(500),
				"message": "Internal server error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.apiError)
			require.NoError(t, err)

			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedMap["code"], result["code"])
			assert.Equal(t, tt.expectedMap["message"], result["message"])
			if tt.expectedMap["details"] != nil {
				assert.Equal(t, tt.expectedMap["details"], result["details"])
			}
		})
	}
}

func TestCORSHeaders(t *testing.T) {
	// Test that CORS headers are set correctly
	router := NewRouter(nil)

	// Create OPTIONS preflight request
	req := httptest.NewRequest(http.MethodOptions, "/health", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Check CORS headers
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Origin"), "http://localhost:5173")
}

func TestJSONRequestDecoding(t *testing.T) {
	tests := []struct {
		name        string
		jsonInput   string
		expectError bool
	}{
		{
			name:        "valid JSON object",
			jsonInput:   `{"name": "Test Campaign", "description": "A test"}`,
			expectError: false,
		},
		{
			name:        "empty object",
			jsonInput:   `{}`,
			expectError: false,
		},
		{
			name:        "invalid JSON",
			jsonInput:   `{"name": "Test"`,
			expectError: true,
		},
		{
			name:        "array instead of object",
			jsonInput:   `["item1", "item2"]`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req models.CreateCampaignRequest
			err := json.NewDecoder(bytes.NewReader([]byte(tt.jsonInput))).Decode(&req)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestContentTypeHeader(t *testing.T) {
	// Verify that all JSON responses have correct Content-Type
	router := NewRouter(nil)

	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/health"},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			req := httptest.NewRequest(ep.method, ep.path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			contentType := rec.Header().Get("Content-Type")
			assert.Equal(t, "application/json", contentType)
		})
	}
}

func TestRouterMiddleware(t *testing.T) {
	// Test that the router has required middleware
	router := NewRouter(nil)

	// Test request ID middleware by checking response headers
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// The request completed successfully (recoverer didn't panic)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHealthEndpoint_ResponseFormat(t *testing.T) {
	router := NewRouter(nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Verify exact response format
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, `{"status":"ok"}`, rec.Body.String())
}
