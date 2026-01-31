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
// AuthMiddleware returns HTTP middleware that requires a valid JWT in the
// Authorization header and, on success, attaches the parsed claims to the
// request context under userClaimsKey.
//
// It responds with 401 Unauthorized if the Authorization header is missing,
// does not use the "Bearer " prefix, contains an empty token, or if token
// verification fails. The jwtSecret parameter is used to verify the token.
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
// OptionalAuthMiddleware returns middleware that validates a JWT when provided and attaches its claims to the request context.
// If the Authorization header is missing or contains an empty Bearer token, the request proceeds without authentication; if a token is present but invalid, the middleware responds with HTTP 401 Unauthorized.
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
// GetUserFromContext retrieves JWT claims from the context and reports whether they are present.
// It returns the claims and true if present; otherwise nil and false.
func GetUserFromContext(ctx context.Context) (*JWTClaims, bool) {
	claims, ok := ctx.Value(userClaimsKey).(*JWTClaims)
	return claims, ok
}

// GetUserIDFromContext retrieves the user ID from the request context.
// GetUserIDFromContext retrieves the user's UUID from the request context's JWT claims.
// It returns the parsed UUID and true if claims are present and the UserID is a valid UUID; otherwise it returns uuid.Nil and false.
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
