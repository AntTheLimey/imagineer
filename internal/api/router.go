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
	"net/http"
	"time"

	"github.com/antonypegg/imagineer/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// NewRouter creates a new chi router with all routes configured.
func NewRouter(db *database.DB) http.Handler {
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
		// Game Systems
		r.Route("/game-systems", func(r chi.Router) {
			r.Get("/", h.ListGameSystems)
			r.Get("/{id}", h.GetGameSystem)
			r.Get("/code/{code}", h.GetGameSystemByCode)
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

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	return r
}
