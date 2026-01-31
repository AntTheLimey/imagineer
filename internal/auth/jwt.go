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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// JWTClaims represents the claims in a JWT token.
type JWTClaims struct {
	UserID string `json:"sub"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Exp    int64  `json:"exp"`
	Iat    int64  `json:"iat"`
}

// jwtHeader is the standard JWT header for HS256 algorithm.
type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

var defaultHeader = jwtHeader{
	Alg: "HS256",
	Typ: "JWT",
}

// Sign creates a signed JWT token from the claims.
func (c *JWTClaims) Sign(secret []byte) (string, error) {
	// Encode header
	headerJSON, err := json.Marshal(defaultHeader)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JWT header: %w", err)
	}
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	// Encode claims
	claimsJSON, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JWT claims: %w", err)
	}
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Create signature
	message := headerB64 + "." + claimsB64
	signature := signHS256([]byte(message), secret)
	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)

	return message + "." + signatureB64, nil
}

// ParseAndVerify parses a JWT token and verifies its signature.
func ParseAndVerify(tokenString string, secret []byte) (*JWTClaims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid JWT format")
	}

	headerB64, claimsB64, signatureB64 := parts[0], parts[1], parts[2]

	// Verify signature
	message := headerB64 + "." + claimsB64
	expectedSig := signHS256([]byte(message), secret)
	actualSig, err := base64.RawURLEncoding.DecodeString(signatureB64)
	if err != nil {
		return nil, errors.New("invalid JWT signature encoding")
	}

	if !hmac.Equal(expectedSig, actualSig) {
		return nil, errors.New("invalid JWT signature")
	}

	// Decode and verify header
	headerJSON, err := base64.RawURLEncoding.DecodeString(headerB64)
	if err != nil {
		return nil, errors.New("invalid JWT header encoding")
	}

	var header jwtHeader
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, errors.New("invalid JWT header")
	}

	if header.Alg != "HS256" {
		return nil, fmt.Errorf("unsupported JWT algorithm: %s", header.Alg)
	}

	// Decode claims
	claimsJSON, err := base64.RawURLEncoding.DecodeString(claimsB64)
	if err != nil {
		return nil, errors.New("invalid JWT claims encoding")
	}

	var claims JWTClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, errors.New("invalid JWT claims")
	}

	// Verify expiration
	if claims.Exp < time.Now().Unix() {
		return nil, errors.New("JWT token has expired")
	}

	return &claims, nil
}

// signHS256 creates an HMAC-SHA256 signature.
func signHS256(message, secret []byte) []byte {
	h := hmac.New(sha256.New, secret)
	h.Write(message)
	return h.Sum(nil)
}
