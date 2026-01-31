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
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

// userClaimsKey is the context key for storing JWT claims.
const userClaimsKey contextKey = "userClaims"

// AuthMiddleware returns chi middleware that validates JWT tokens.
// It extracts the JWT from the Authorization header, validates it,
// and attaches the user claims to the request context.
// Returns 401 Unauthorized if the token is missing or invalid.
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	secret := []byte(jwtSecret)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Validate Bearer prefix
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == "" {
				http.Error(w, "Token required", http.StatusUnauthorized)
				return
			}

			// Parse and verify the token
			claims, err := ParseAndVerify(tokenString, secret)
			if err != nil {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Attach claims to context and continue
			ctx := context.WithValue(r.Context(), userClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuthMiddleware returns chi middleware that validates JWT tokens
// but does not fail if the token is missing. This is useful for routes that
// work with or without authentication.
// If a token is present and valid, the claims are attached to the context.
// If no token is present, the request proceeds without claims.
// Returns 401 Unauthorized only if a token is present but invalid.
func OptionalAuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	secret := []byte(jwtSecret)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// No auth header - proceed without authentication
				next.ServeHTTP(w, r)
				return
			}

			// Validate Bearer prefix
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == "" {
				// Empty token after Bearer - proceed without authentication
				next.ServeHTTP(w, r)
				return
			}

			// Parse and verify the token
			claims, err := ParseAndVerify(tokenString, secret)
			if err != nil {
				// Token present but invalid
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Attach claims to context and continue
			ctx := context.WithValue(r.Context(), userClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext retrieves the JWT claims from the request context.
// Returns the claims and true if present, nil and false otherwise.
func GetUserFromContext(ctx context.Context) (*JWTClaims, bool) {
	claims, ok := ctx.Value(userClaimsKey).(*JWTClaims)
	return claims, ok
}

// GetUserIDFromContext retrieves the user ID from the request context.
// Returns the UUID and true if present, uuid.Nil and false otherwise.
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	claims, ok := GetUserFromContext(ctx)
	if !ok {
		return uuid.Nil, false
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, false
	}

	return userID, true
}
