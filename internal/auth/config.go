/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

// Package auth provides Google OAuth authentication for the Imagineer API.
package auth

import (
	"errors"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OAuthConfig holds the configuration for Google OAuth.
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	FrontendURL  string
}

// NewOAuthConfigFromEnv creates an OAuthConfig from environment variables.
// Required environment variables:
//   - GOOGLE_CLIENT_ID: Google OAuth client ID
//   - GOOGLE_CLIENT_SECRET: Google OAuth client secret
//   - GOOGLE_REDIRECT_URL: OAuth callback URL (e.g., http://localhost:3001/api/auth/google/callback)
//   - FRONTEND_URL: Frontend application URL (e.g., http://localhost:5173)
func NewOAuthConfigFromEnv() (*OAuthConfig, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID == "" {
		return nil, errors.New("GOOGLE_CLIENT_ID environment variable is required")
	}

	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if clientSecret == "" {
		return nil, errors.New("GOOGLE_CLIENT_SECRET environment variable is required")
	}

	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")
	if redirectURL == "" {
		return nil, errors.New("GOOGLE_REDIRECT_URL environment variable is required")
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		// Default to localhost:5173 for development
		frontendURL = "http://localhost:5173"
	}

	return &OAuthConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		FrontendURL:  frontendURL,
	}, nil
}

// GoogleOAuth2Config creates an oauth2.Config for Google authentication.
// The config requests access to the user's basic profile and email address.
func (c *OAuthConfig) GoogleOAuth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		RedirectURL:  c.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}
