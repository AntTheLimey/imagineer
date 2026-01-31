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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2/google"
)

func TestNewOAuthConfigFromEnv(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		wantErr     bool
		errContains string
	}{
		{
			name: "all variables set",
			envVars: map[string]string{
				"GOOGLE_CLIENT_ID":     "test-client-id",
				"GOOGLE_CLIENT_SECRET": "test-client-secret",
				"GOOGLE_REDIRECT_URL":  "http://localhost:8080/callback",
			},
			wantErr: false,
		},
		{
			name: "missing client ID",
			envVars: map[string]string{
				"GOOGLE_CLIENT_SECRET": "test-client-secret",
				"GOOGLE_REDIRECT_URL":  "http://localhost:8080/callback",
			},
			wantErr:     true,
			errContains: "GOOGLE_CLIENT_ID",
		},
		{
			name: "missing client secret",
			envVars: map[string]string{
				"GOOGLE_CLIENT_ID":    "test-client-id",
				"GOOGLE_REDIRECT_URL": "http://localhost:8080/callback",
			},
			wantErr:     true,
			errContains: "GOOGLE_CLIENT_SECRET",
		},
		{
			name: "missing redirect URL",
			envVars: map[string]string{
				"GOOGLE_CLIENT_ID":     "test-client-id",
				"GOOGLE_CLIENT_SECRET": "test-client-secret",
			},
			wantErr:     true,
			errContains: "GOOGLE_REDIRECT_URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all OAuth env vars first
			os.Unsetenv("GOOGLE_CLIENT_ID")
			os.Unsetenv("GOOGLE_CLIENT_SECRET")
			os.Unsetenv("GOOGLE_REDIRECT_URL")

			// Set the env vars for this test
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg, err := NewOAuthConfigFromEnv()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, cfg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, cfg)
				assert.Equal(t, tt.envVars["GOOGLE_CLIENT_ID"], cfg.ClientID)
				assert.Equal(t, tt.envVars["GOOGLE_CLIENT_SECRET"], cfg.ClientSecret)
				assert.Equal(t, tt.envVars["GOOGLE_REDIRECT_URL"], cfg.RedirectURL)
			}
		})
	}
}

func TestGoogleOAuth2Config(t *testing.T) {
	cfg := &OAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
	}

	oauth2Cfg := cfg.GoogleOAuth2Config()

	assert.Equal(t, cfg.ClientID, oauth2Cfg.ClientID)
	assert.Equal(t, cfg.ClientSecret, oauth2Cfg.ClientSecret)
	assert.Equal(t, cfg.RedirectURL, oauth2Cfg.RedirectURL)
	assert.Equal(t, google.Endpoint, oauth2Cfg.Endpoint)
	assert.Contains(t, oauth2Cfg.Scopes, "https://www.googleapis.com/auth/userinfo.email")
	assert.Contains(t, oauth2Cfg.Scopes, "https://www.googleapis.com/auth/userinfo.profile")
}
