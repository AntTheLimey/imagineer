/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// gitleaks:allow
const testSecret = "test-secret-key-at-least-32-bytes-long"

// createValidToken creates a valid JWT token for testing.
func createValidToken(t *testing.T, userID string) string {
	t.Helper()

	claims := &JWTClaims{
		UserID: userID,
		Email:  "test@example.com",
		Name:   "Test User",
		Exp:    time.Now().Add(time.Hour).Unix(),
		Iat:    time.Now().Unix(),
	}

	token, err := claims.Sign([]byte(testSecret))
	require.NoError(t, err)
	return token
}

// createExpiredToken creates an expired JWT token for testing.
func createExpiredToken(t *testing.T) string {
	t.Helper()

	claims := &JWTClaims{
		UserID: "user-123",
		Email:  "test@example.com",
		Name:   "Test User",
		Exp:    time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
		Iat:    time.Now().Add(-2 * time.Hour).Unix(),
	}

	token, err := claims.Sign([]byte(testSecret))
	require.NoError(t, err)
	return token
}

// testHandler is a simple handler that records if it was called and
// extracts user info from context.
func testHandler(called *bool, userID *string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		*called = true

		claims, ok := GetUserFromContext(r.Context())
		if ok {
			*userID = claims.UserID
		}

		w.WriteHeader(http.StatusOK)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	expectedUserID := uuid.New().String()
	token := createValidToken(t, expectedUserID)

	var handlerCalled bool
	var extractedUserID string

	middleware := AuthMiddleware(testSecret)
	handler := middleware(testHandler(&handlerCalled, &extractedUserID))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, handlerCalled, "handler should have been called")
	assert.Equal(t, expectedUserID, extractedUserID)
}

func TestAuthMiddleware_MissingAuthorizationHeader(t *testing.T) {
	var handlerCalled bool
	var extractedUserID string

	middleware := AuthMiddleware(testSecret)
	handler := middleware(testHandler(&handlerCalled, &extractedUserID))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, handlerCalled, "handler should not have been called")
	assert.Contains(t, rec.Body.String(), "Authorization header required")
}

func TestAuthMiddleware_InvalidTokenFormat(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		wantErr    string
	}{
		{
			name:       "missing Bearer prefix",
			authHeader: "token-without-bearer",
			wantErr:    "Invalid authorization header format",
		},
		{
			name:       "Basic auth instead of Bearer",
			authHeader: "Basic dXNlcjpwYXNz",
			wantErr:    "Invalid authorization header format",
		},
		{
			name:       "Bearer with empty token",
			authHeader: "Bearer ",
			wantErr:    "Token required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var handlerCalled bool
			var extractedUserID string

			middleware := AuthMiddleware(testSecret)
			handler := middleware(testHandler(&handlerCalled, &extractedUserID))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tt.authHeader)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
			assert.False(t, handlerCalled, "handler should not have been called")
			assert.Contains(t, rec.Body.String(), tt.wantErr)
		})
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	token := createExpiredToken(t)

	var handlerCalled bool
	var extractedUserID string

	middleware := AuthMiddleware(testSecret)
	handler := middleware(testHandler(&handlerCalled, &extractedUserID))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, handlerCalled, "handler should not have been called")
	assert.Contains(t, rec.Body.String(), "Invalid or expired token")
}

func TestAuthMiddleware_InvalidSignature(t *testing.T) {
	// Create token with different secret
	claims := &JWTClaims{
		UserID: "user-123",
		Email:  "test@example.com",
		Name:   "Test User",
		Exp:    time.Now().Add(time.Hour).Unix(),
		Iat:    time.Now().Unix(),
	}
	token, err := claims.Sign([]byte("different-secret-key-32-bytes!"))
	require.NoError(t, err)

	var handlerCalled bool
	var extractedUserID string

	middleware := AuthMiddleware(testSecret)
	handler := middleware(testHandler(&handlerCalled, &extractedUserID))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, handlerCalled, "handler should not have been called")
	assert.Contains(t, rec.Body.String(), "Invalid or expired token")
}

func TestAuthMiddleware_MalformedToken(t *testing.T) {
	var handlerCalled bool
	var extractedUserID string

	middleware := AuthMiddleware(testSecret)
	handler := middleware(testHandler(&handlerCalled, &extractedUserID))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer not-a-valid-jwt-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, handlerCalled, "handler should not have been called")
	assert.Contains(t, rec.Body.String(), "Invalid or expired token")
}

func TestOptionalAuthMiddleware_ValidToken(t *testing.T) {
	expectedUserID := uuid.New().String()
	token := createValidToken(t, expectedUserID)

	var handlerCalled bool
	var extractedUserID string

	middleware := OptionalAuthMiddleware(testSecret)
	handler := middleware(testHandler(&handlerCalled, &extractedUserID))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, handlerCalled, "handler should have been called")
	assert.Equal(t, expectedUserID, extractedUserID)
}

func TestOptionalAuthMiddleware_MissingAuthorizationHeader(t *testing.T) {
	var handlerCalled bool
	var extractedUserID string

	middleware := OptionalAuthMiddleware(testSecret)
	handler := middleware(testHandler(&handlerCalled, &extractedUserID))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Should allow unauthenticated access
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, handlerCalled, "handler should have been called")
	assert.Empty(t, extractedUserID, "no user should be in context")
}

func TestOptionalAuthMiddleware_EmptyBearerToken(t *testing.T) {
	var handlerCalled bool
	var extractedUserID string

	middleware := OptionalAuthMiddleware(testSecret)
	handler := middleware(testHandler(&handlerCalled, &extractedUserID))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer ")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Should allow unauthenticated access for empty token
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, handlerCalled, "handler should have been called")
	assert.Empty(t, extractedUserID, "no user should be in context")
}

func TestOptionalAuthMiddleware_InvalidToken(t *testing.T) {
	var handlerCalled bool
	var extractedUserID string

	middleware := OptionalAuthMiddleware(testSecret)
	handler := middleware(testHandler(&handlerCalled, &extractedUserID))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Should reject invalid token even for optional auth
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, handlerCalled, "handler should not have been called")
	assert.Contains(t, rec.Body.String(), "Invalid or expired token")
}

func TestOptionalAuthMiddleware_ExpiredToken(t *testing.T) {
	token := createExpiredToken(t)

	var handlerCalled bool
	var extractedUserID string

	middleware := OptionalAuthMiddleware(testSecret)
	handler := middleware(testHandler(&handlerCalled, &extractedUserID))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Should reject expired token even for optional auth
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, handlerCalled, "handler should not have been called")
}

func TestOptionalAuthMiddleware_InvalidFormat(t *testing.T) {
	var handlerCalled bool
	var extractedUserID string

	middleware := OptionalAuthMiddleware(testSecret)
	handler := middleware(testHandler(&handlerCalled, &extractedUserID))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Should reject invalid format
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, handlerCalled, "handler should not have been called")
	assert.Contains(t, rec.Body.String(), "Invalid authorization header format")
}

func TestGetUserFromContext_WithClaims(t *testing.T) {
	expectedUserID := uuid.New().String()
	token := createValidToken(t, expectedUserID)

	middleware := AuthMiddleware(testSecret)

	var gotClaims *JWTClaims
	var gotOK bool

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotClaims, gotOK = GetUserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.True(t, gotOK)
	require.NotNil(t, gotClaims)
	assert.Equal(t, expectedUserID, gotClaims.UserID)
	assert.Equal(t, "test@example.com", gotClaims.Email)
	assert.Equal(t, "Test User", gotClaims.Name)
}

func TestGetUserFromContext_WithoutClaims(t *testing.T) {
	var gotClaims *JWTClaims
	var gotOK bool

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotClaims, gotOK = GetUserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.False(t, gotOK)
	assert.Nil(t, gotClaims)
}

func TestGetUserIDFromContext_ValidUUID(t *testing.T) {
	expectedUserID := uuid.New()
	token := createValidToken(t, expectedUserID.String())

	middleware := AuthMiddleware(testSecret)

	var gotUserID uuid.UUID
	var gotOK bool

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID, gotOK = GetUserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.True(t, gotOK)
	assert.Equal(t, expectedUserID, gotUserID)
}

func TestGetUserIDFromContext_InvalidUUID(t *testing.T) {
	// Create token with non-UUID user ID
	token := createValidToken(t, "not-a-uuid")

	middleware := AuthMiddleware(testSecret)

	var gotUserID uuid.UUID
	var gotOK bool

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID, gotOK = GetUserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.False(t, gotOK)
	assert.Equal(t, uuid.Nil, gotUserID)
}

func TestGetUserIDFromContext_NoClaims(t *testing.T) {
	var gotUserID uuid.UUID
	var gotOK bool

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID, gotOK = GetUserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.False(t, gotOK)
	assert.Equal(t, uuid.Nil, gotUserID)
}
