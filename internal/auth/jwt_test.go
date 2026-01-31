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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTSignAndVerify(t *testing.T) {
	secret := []byte("test-secret-key-at-least-32-bytes-long")

	claims := &JWTClaims{
		UserID: "user-123",
		Email:  "test@example.com",
		Name:   "Test User",
		Exp:    time.Now().Add(time.Hour).Unix(),
		Iat:    time.Now().Unix(),
	}

	// Sign the token
	token, err := claims.Sign(secret)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify the token
	parsedClaims, err := ParseAndVerify(token, secret)
	require.NoError(t, err)
	assert.Equal(t, claims.UserID, parsedClaims.UserID)
	assert.Equal(t, claims.Email, parsedClaims.Email)
	assert.Equal(t, claims.Name, parsedClaims.Name)
	assert.Equal(t, claims.Exp, parsedClaims.Exp)
	assert.Equal(t, claims.Iat, parsedClaims.Iat)
}

func TestJWTExpiredToken(t *testing.T) {
	secret := []byte("test-secret-key-at-least-32-bytes-long")

	claims := &JWTClaims{
		UserID: "user-123",
		Email:  "test@example.com",
		Name:   "Test User",
		Exp:    time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
		Iat:    time.Now().Add(-2 * time.Hour).Unix(),
	}

	token, err := claims.Sign(secret)
	require.NoError(t, err)

	_, err = ParseAndVerify(token, secret)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestJWTInvalidSignature(t *testing.T) {
	secret1 := []byte("test-secret-key-1-at-least-32-bytes")
	secret2 := []byte("test-secret-key-2-at-least-32-bytes")

	claims := &JWTClaims{
		UserID: "user-123",
		Email:  "test@example.com",
		Name:   "Test User",
		Exp:    time.Now().Add(time.Hour).Unix(),
		Iat:    time.Now().Unix(),
	}

	// Sign with one secret
	token, err := claims.Sign(secret1)
	require.NoError(t, err)

	// Try to verify with different secret
	_, err = ParseAndVerify(token, secret2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JWT signature")
}

func TestJWTInvalidFormat(t *testing.T) {
	secret := []byte("test-secret-key-at-least-32-bytes-long")

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "single part",
			token: "single",
		},
		{
			name:  "two parts",
			token: "part1.part2",
		},
		{
			name:  "four parts",
			token: "part1.part2.part3.part4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseAndVerify(tt.token, secret)
			require.Error(t, err)
		})
	}
}

func TestJWTTamperedToken(t *testing.T) {
	secret := []byte("test-secret-key-at-least-32-bytes-long")

	claims := &JWTClaims{
		UserID: "user-123",
		Email:  "test@example.com",
		Name:   "Test User",
		Exp:    time.Now().Add(time.Hour).Unix(),
		Iat:    time.Now().Unix(),
	}

	token, err := claims.Sign(secret)
	require.NoError(t, err)

	// Tamper with the token by changing a character
	tamperedToken := token[:len(token)-5] + "XXXXX"

	_, err = ParseAndVerify(tamperedToken, secret)
	require.Error(t, err)
}
