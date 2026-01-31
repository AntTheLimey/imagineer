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
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/models"
	"github.com/jackc/pgx/v5"
	"golang.org/x/oauth2"
	googleOAuth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

const (
	// stateCookieName is the name of the cookie used to store OAuth state.
	stateCookieName = "oauth_state"
	// stateCookieMaxAge is the maximum age of the state cookie in seconds (10 minutes).
	stateCookieMaxAge = 600
	// stateLength is the number of random bytes used for the state token.
	stateLength = 32
)

// AuthHandler handles OAuth authentication requests.
type AuthHandler struct {
	db           *database.DB
	oauthConfig  *oauth2.Config
	jwtSecret    []byte
	jwtExpiryHrs int
	frontendURL  string
}

// NewAuthHandler creates a new AuthHandler with the given configuration.
func NewAuthHandler(db *database.DB, oauthCfg *OAuthConfig, jwtSecret []byte, jwtExpiryHrs int) *AuthHandler {
	return &AuthHandler{
		db:           db,
		oauthConfig:  oauthCfg.GoogleOAuth2Config(),
		jwtSecret:    jwtSecret,
		jwtExpiryHrs: jwtExpiryHrs,
		frontendURL:  oauthCfg.FrontendURL,
	}
}

// generateState generates a cryptographically secure, URL-safe base64-encoded state token.
// It returns the token or an error if secure random byte generation fails.
func generateState() (string, error) {
	b := make([]byte, stateLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random state: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// isSecureRequest determines if the request was made over a secure connection.
// It first checks the X-Forwarded-Proto header (used when behind a reverse proxy
// isSecureRequest reports whether the incoming request was made over HTTPS.
// It first checks the X-Forwarded-Proto header (used by reverse proxies) and
// returns true if its value is "https"; if the header is absent, it returns
// true when r.TLS is non-nil.
func isSecureRequest(r *http.Request) bool {
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		return proto == "https"
	}
	return r.TLS != nil
}

// HandleGoogleLogin initiates the Google OAuth flow.
// It generates a secure state token, stores it in a cookie, and redirects
// the user to Google's OAuth consent screen.
func (h *AuthHandler) HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := generateState()
	if err != nil {
		log.Printf("Error generating OAuth state: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Store state in a secure, HTTP-only cookie for CSRF protection
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    state,
		Path:     "/",
		MaxAge:   stateCookieMaxAge,
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to Google's OAuth consent screen
	url := h.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleGoogleCallback handles the OAuth callback from Google.
// It validates the state token, exchanges the authorization code for tokens,
// fetches user info from Google, creates or updates the user in the database,
// generates a JWT, and redirects to the frontend with the auth data.
func (h *AuthHandler) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Validate state parameter to prevent CSRF attacks
	stateCookie, err := r.Cookie(stateCookieName)
	if err != nil {
		log.Printf("OAuth callback missing state cookie: %v", err)
		h.redirectWithError(w, r, "Invalid OAuth state")
		return
	}

	stateParam := r.URL.Query().Get("state")
	if stateParam == "" || stateParam != stateCookie.Value {
		log.Printf("OAuth state mismatch: cookie=%q, param=%q", stateCookie.Value, stateParam)
		h.redirectWithError(w, r, "Invalid OAuth state")
		return
	}

	// Clear the state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteLaxMode,
	})

	// Check for OAuth error response
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		errDesc := r.URL.Query().Get("error_description")
		log.Printf("OAuth error from Google: %s - %s", errParam, errDesc)
		h.redirectWithError(w, r, "Authentication failed")
		return
	}

	// Exchange authorization code for tokens
	code := r.URL.Query().Get("code")
	if code == "" {
		log.Printf("OAuth callback missing authorization code")
		h.redirectWithError(w, r, "Missing authorization code")
		return
	}

	token, err := h.oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("Failed to exchange OAuth code: %v", err)
		h.redirectWithError(w, r, "Failed to exchange authorization code")
		return
	}

	// Fetch user info from Google
	googleUser, err := h.fetchGoogleUserInfo(r.Context(), token)
	if err != nil {
		log.Printf("Failed to fetch Google user info: %v", err)
		h.redirectWithError(w, r, "Failed to fetch user information")
		return
	}

	// Create or update user in database
	user, err := h.findOrCreateUser(r.Context(), googleUser)
	if err != nil {
		log.Printf("Failed to create/update user: %v", err)
		h.redirectWithError(w, r, "Failed to process user account")
		return
	}

	// Generate JWT for the user
	jwt, err := h.generateJWT(user)
	if err != nil {
		log.Printf("Failed to generate JWT: %v", err)
		h.redirectWithError(w, r, "Failed to generate authentication token")
		return
	}

	// Build user JSON for frontend
	userResponse := UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
	}
	userJSON, err := json.Marshal(userResponse)
	if err != nil {
		log.Printf("Failed to marshal user response: %v", err)
		h.redirectWithError(w, r, "Failed to process user data")
		return
	}

	// Redirect to frontend with token and user data
	redirectURL := h.buildFrontendCallbackURL(jwt, string(userJSON))
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// redirectWithError redirects to the frontend auth callback with an error message.
func (h *AuthHandler) redirectWithError(w http.ResponseWriter, r *http.Request, errorMsg string) {
	redirectURL := fmt.Sprintf("%s/auth/callback?error=%s",
		h.frontendURL,
		url.QueryEscape(errorMsg))
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// buildFrontendCallbackURL constructs the frontend callback URL with auth data.
func (h *AuthHandler) buildFrontendCallbackURL(token, userJSON string) string {
	return fmt.Sprintf("%s/auth/callback?token=%s&user=%s",
		h.frontendURL,
		url.QueryEscape(token),
		url.QueryEscape(userJSON))
}

// fetchGoogleUserInfo retrieves user information from Google's OAuth2 API.
func (h *AuthHandler) fetchGoogleUserInfo(ctx context.Context, token *oauth2.Token) (*googleOAuth2.Userinfo, error) {
	oauth2Service, err := googleOAuth2.NewService(ctx,
		option.WithTokenSource(h.oauthConfig.TokenSource(ctx, token)))
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth2 service: %w", err)
	}

	userInfo, err := oauth2Service.Userinfo.Get().Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return userInfo, nil
}

// findOrCreateUser looks up a user by Google ID and creates them if not found.
// If the user exists, their profile is updated with the latest information.
func (h *AuthHandler) findOrCreateUser(ctx context.Context, googleUser *googleOAuth2.Userinfo) (*models.User, error) {
	// Try to find existing user by Google ID
	user, err := h.db.GetUserByGoogleID(ctx, googleUser.Id)
	if err == nil {
		// User exists - update their profile with latest info
		user.Email = googleUser.Email
		user.Name = googleUser.Name
		if googleUser.Picture != "" {
			user.AvatarURL = &googleUser.Picture
		}
		if err := h.db.UpdateUser(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
		return user, nil
	}

	// Check if the error is "not found" - if so, create new user
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to lookup user: %w", err)
	}

	// Create new user
	user = &models.User{
		GoogleID: googleUser.Id,
		Email:    googleUser.Email,
		Name:     googleUser.Name,
	}
	if googleUser.Picture != "" {
		user.AvatarURL = &googleUser.Picture
	}

	if err := h.db.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// generateJWT creates a JWT token for the authenticated user.
func (h *AuthHandler) generateJWT(user *models.User) (string, error) {
	expiresAt := time.Now().Add(time.Duration(h.jwtExpiryHrs) * time.Hour)

	claims := JWTClaims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Name:   user.Name,
		Exp:    expiresAt.Unix(),
		Iat:    time.Now().Unix(),
	}

	return claims.Sign(h.jwtSecret)
}

// AuthResponse is the JSON response returned after successful authentication.
type AuthResponse struct {
	Token     string       `json:"token"`
	ExpiresIn int          `json:"expiresIn"`
	User      UserResponse `json:"user"`
}

// UserResponse is the user information included in the auth response.
type UserResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	AvatarURL *string `json:"avatarUrl,omitempty"`
}