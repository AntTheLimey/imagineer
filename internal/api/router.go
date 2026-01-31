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
	"errors"
	"net/http"
	"time"

	"github.com/antonypegg/imagineer/internal/auth"
	"github.com/antonypegg/imagineer/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// ErrMissingJWTSecret is returned when the JWT secret is not configured.
var ErrMissingJWTSecret = errors.New("JWT secret is required for authentication")

// NewRouter creates a new chi router with all routes configured.
// If authHandler is nil, auth routes will not be registered.
// Returns an error if jwtSecret is empty, as authentication is required.
func NewRouter(db *database.DB, authHandler *auth.AuthHandler, jwtSecret string) (http.Handler, error) {
	if jwtSecret == "" {
		return nil, ErrMissingJWTSecret
	}
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS configuration for localhost:5173 (Vite dev server)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Create handlers
	h := NewHandler(db)
	importHandler := NewImportHandler(db)
	agentHandler := NewAgentHandler(db)

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Authentication routes (only if auth handler is configured)
		// These routes are always public - they handle the OAuth flow
		if authHandler != nil {
			r.Route("/auth", func(r chi.Router) {
				r.Get("/google", authHandler.HandleGoogleLogin)
				r.Get("/google/callback", authHandler.HandleGoogleCallback)
			})
		}

		// Game Systems - public reference data, no authentication required
		r.Route("/game-systems", func(r chi.Router) {
			r.Get("/", h.ListGameSystems)
			r.Get("/{id}", h.GetGameSystem)
			r.Get("/code/{code}", h.GetGameSystemByCode)
		})

		// Protected routes - require authentication
		r.Group(func(r chi.Router) {
			// Apply authentication middleware (jwtSecret is validated at router creation)
			r.Use(auth.AuthMiddleware(jwtSecret))

			// Campaigns
			r.Route("/campaigns", func(r chi.Router) {
				r.Get("/", h.ListCampaigns)
				r.Post("/", h.CreateCampaign)

				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", h.GetCampaign)
					r.Put("/", h.UpdateCampaign)
					r.Delete("/", h.DeleteCampaign)

					// Campaign stats
					r.Get("/stats", h.GetCampaignStats)

					// Campaign entities
					r.Get("/entities", h.ListEntities)
					r.Post("/entities", h.CreateEntity)
					r.Get("/entities/search", h.SearchEntities)

					// Entity-specific routes within campaign context
					r.Route("/entities/{entityId}", func(r chi.Router) {
						r.Get("/relationships", h.GetEntityRelationships)
						r.Get("/timeline", h.GetEntityTimelineEvents)
					})

					// Campaign relationships
					r.Get("/relationships", h.ListRelationships)
					r.Post("/relationships", h.CreateRelationship)
					r.Route("/relationships/{relationshipId}", func(r chi.Router) {
						r.Get("/", h.GetRelationship)
						r.Put("/", h.UpdateRelationship)
						r.Delete("/", h.DeleteRelationship)
					})

					// Campaign timeline
					r.Get("/timeline", h.ListTimelineEvents)
					r.Post("/timeline", h.CreateTimelineEvent)
					r.Route("/timeline/{eventId}", func(r chi.Router) {
						r.Get("/", h.GetTimelineEvent)
						r.Put("/", h.UpdateTimelineEvent)
						r.Delete("/", h.DeleteTimelineEvent)
					})

					// Campaign import endpoints
					r.Route("/import", func(r chi.Router) {
						r.Post("/evernote", importHandler.ImportEvernote)
						r.Post("/google-docs", importHandler.ImportGoogleDocs)
						r.Post("/file", importHandler.ImportFile)
					})

					// Campaign agent endpoints
					r.Route("/agents", func(r chi.Router) {
						r.Post("/consistency-check", agentHandler.RunConsistencyCheck)
					})
				})
			})

			// Entities (direct access by ID)
			r.Route("/entities", func(r chi.Router) {
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", h.GetEntity)
					r.Put("/", h.UpdateEntity)
					r.Delete("/", h.DeleteEntity)
				})
			})

			// Statistics
			r.Get("/stats", h.GetStats)
			r.Get("/stats/dashboard", h.GetDashboardStats)
		})
	})

	// Health check endpoint - always public
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	return r, nil
}
