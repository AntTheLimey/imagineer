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
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the claims in a JWT token.
// It embeds jwt.RegisteredClaims for standard JWT validation while
// maintaining backward-compatible Exp and Iat fields.
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	// Exp and Iat are kept for backward compatibility but are derived from RegisteredClaims
	Exp int64 `json:"-"`
	Iat int64 `json:"-"`
	jwt.RegisteredClaims
}

// Sign creates a signed JWT token from the claims.
func (c *JWTClaims) Sign(secret []byte) (string, error) {
	// Set RegisteredClaims from our fields
	if c.Exp != 0 {
		c.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(time.Unix(c.Exp, 0))
	}
	if c.Iat != 0 {
		c.RegisteredClaims.IssuedAt = jwt.NewNumericDate(time.Unix(c.Iat, 0))
	}
	if c.UserID != "" {
		c.RegisteredClaims.Subject = c.UserID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %w", err)
	}

	return signedToken, nil
}

// ParseAndVerify parses a JWT string, verifies its HS256 HMAC-SHA256 signature,
// validates the header and expiration, and returns the decoded claims.
//
// It returns an error for any of the following conditions: token does not have
// three dot-separated parts; signature is not valid base64 URL encoding or does
// not match; header or claims fail base64 decoding or JSON unmarshaling; the
// header algorithm is not "HS256"; or the token has expired.
func ParseAndVerify(tokenString string, secret []byte) (*JWTClaims, error) {
	// Create parser with leeway for clock skew
	parser := jwt.NewParser(
		jwt.WithLeeway(5*time.Second),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)

	claims := &JWTClaims{}
	token, err := parser.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("JWT token has expired")
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, errors.New("invalid JWT format")
		}
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return nil, errors.New("invalid JWT signature")
		}
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid JWT token")
	}

	// Populate backward-compatible fields from RegisteredClaims
	if claims.RegisteredClaims.Subject != "" {
		claims.UserID = claims.RegisteredClaims.Subject
	}
	if claims.RegisteredClaims.ExpiresAt != nil {
		claims.Exp = claims.RegisteredClaims.ExpiresAt.Unix()
	}
	if claims.RegisteredClaims.IssuedAt != nil {
		claims.Iat = claims.RegisteredClaims.IssuedAt.Unix()
	}

	return claims, nil
}
