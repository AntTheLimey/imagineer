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

// NewRouter creates and returns an HTTP router configured with common middleware,
// CORS for localhost dev origins, and the application's API routes wired to the
// provided database and handlers.
//
// If authHandler is non-nil, OAuth routes for Google are registered under /api/auth.
// JWT authentication middleware is applied to protected routes (campaigns, entities,
// stats, imports, agents).
//
// NewRouter creates and returns a configured HTTP router for the API, including middleware, CORS,
// public routes, and authentication-protected routes for campaign, entity, user, import, agent and
// statistics endpoints.
// If jwtSecret is empty, NewRouter returns ErrMissingJWTSecret.
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
	contentAnalysisHandler := NewContentAnalysisHandler(db)
	h := NewHandler(db, contentAnalysisHandler)
	importHandler := NewImportHandler(db)
	agentHandler := NewAgentHandler(db)
	entityDetectionHandler := NewEntityDetectionHandler(db)
	entityResolveHandler := NewEntityResolveHandler(db)
	entityLogHandler := NewEntityLogHandler(db)
	sceneHandler := NewSceneHandler(db)
	enrichmentHandler := NewEnrichmentHandler(db)

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

			// User settings
			r.Route("/user/settings", func(r chi.Router) {
				r.Get("/", h.GetUserSettings)
				r.Put("/", h.UpdateUserSettings)
			})

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

					// Campaign content search
					r.Get("/search", h.SearchCampaignContent)

					// Campaign entities
					r.Get("/entities", h.ListEntities)
					r.Post("/entities", h.CreateEntity)
					r.Get("/entities/search", h.SearchEntities)
					r.Get("/entities/resolve", entityResolveHandler.ResolveEntity)

					// Entity-specific routes within campaign context
					r.Route("/entities/{entityId}", func(r chi.Router) {
						r.Get("/relationships", h.GetEntityRelationships)
						r.Get("/timeline", h.GetEntityTimelineEvents)

						// Entity log
						r.Get("/log", entityLogHandler.ListEntityLogs)
						r.Post("/log", entityLogHandler.CreateEntityLog)
						r.Route("/log/{logId}", func(r chi.Router) {
							r.Put("/", entityLogHandler.UpdateEntityLog)
							r.Delete("/", entityLogHandler.DeleteEntityLog)
						})
					})

					// Campaign relationships
					r.Get("/relationships", h.ListRelationships)
					r.Post("/relationships", h.CreateRelationship)
					r.Route("/relationships/{relationshipId}", func(r chi.Router) {
						r.Get("/", h.GetRelationship)
						r.Put("/", h.UpdateRelationship)
						r.Delete("/", h.DeleteRelationship)
					})

					// Campaign relationship types
					r.Get("/relationship-types", h.ListRelationshipTypes)
					r.Post("/relationship-types", h.CreateRelationshipType)
					r.Delete("/relationship-types/{typeId}", h.DeleteRelationshipType)

					// Player characters
					r.Get("/player-characters", h.ListPlayerCharacters)
					r.Post("/player-characters", h.CreatePlayerCharacter)
					r.Route("/player-characters/{pcId}", func(r chi.Router) {
						r.Get("/", h.GetPlayerCharacter)
						r.Put("/", h.UpdatePlayerCharacter)
						r.Delete("/", h.DeletePlayerCharacter)
					})

					// Chapters
					r.Get("/chapters", h.ListChapters)
					r.Post("/chapters", h.CreateChapter)
					r.Post("/chapters/detect-entities", entityDetectionHandler.DetectEntities)
					r.Route("/chapters/{chapterId}", func(r chi.Router) {
						r.Get("/", h.GetChapter)
						r.Put("/", h.UpdateChapter)
						r.Delete("/", h.DeleteChapter)
						r.Get("/sessions", h.ListSessionsByChapter)

						// Chapter entity links
						r.Get("/entities", h.ListChapterEntities)
						r.Post("/entities", h.CreateChapterEntity)
						r.Route("/entities/{linkId}", func(r chi.Router) {
							r.Put("/", h.UpdateChapterEntity)
							r.Delete("/", h.DeleteChapterEntity)
						})
					})

					// Sessions
					r.Get("/sessions", h.ListSessions)
					r.Post("/sessions", h.CreateSession)
					r.Route("/sessions/{sessionId}", func(r chi.Router) {
						r.Get("/", h.GetSession)
						r.Put("/", h.UpdateSession)
						r.Delete("/", h.DeleteSession)

						// Scenes
						r.Get("/scenes", sceneHandler.ListScenes)
						r.Post("/scenes", sceneHandler.CreateScene)
						r.Route("/scenes/{sceneId}", func(r chi.Router) {
							r.Get("/", sceneHandler.GetScene)
							r.Put("/", sceneHandler.UpdateScene)
							r.Delete("/", sceneHandler.DeleteScene)
						})
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

					// Content analysis
					r.Route("/analysis", func(r chi.Router) {
						r.Get("/jobs", contentAnalysisHandler.ListJobs)
						r.Post("/trigger", contentAnalysisHandler.TriggerAnalysis)
						r.Get("/pending-count", contentAnalysisHandler.GetPendingCount)
						r.Route("/jobs/{jobId}", func(r chi.Router) {
							r.Get("/", contentAnalysisHandler.GetJob)
							r.Get("/items", contentAnalysisHandler.ListJobItems)
							r.Put("/resolve-all", contentAnalysisHandler.BatchResolve)
							r.Post("/enrich", enrichmentHandler.TriggerEnrichment)
							r.Post("/cancel-enrichment", contentAnalysisHandler.CancelEnrichment)
							r.Get("/enrichment-stream", enrichmentHandler.EnrichmentStream)
						})
						r.Put("/items/{itemId}", contentAnalysisHandler.ResolveItem)
						r.Put("/items/{itemId}/revert", contentAnalysisHandler.RevertItem)
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

			// Evernote local import endpoints (macOS only)
			r.Route("/import/evernote", func(r chi.Router) {
				r.Get("/status", importHandler.GetEvernoteLocalStatus)
				r.Get("/notebooks", importHandler.ListEvernoteLocalNotebooks)
				r.Get("/notebooks/{name}/notes", importHandler.ListEvernoteLocalNotes)
				r.Post("/import", importHandler.ImportEvernoteLocal)
			})
		})
	})

	// Health check endpoint - always public
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	return r, nil
}
