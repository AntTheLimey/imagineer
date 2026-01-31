/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

// Package api provides HTTP handlers for the Imagineer API.
package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler provides HTTP handlers for the API.
type Handler struct {
	db *database.DB
}

// NewHandler creates a new Handler with the given database connection.
func NewHandler(db *database.DB) *Handler {
	return &Handler{db: db}
}

// respondJSON writes a JSON response.
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	}
}

// respondError writes a JSON error response.
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, models.APIError{
		Code:    status,
		Message: message,
	})
}

// parseUUID parses a UUID from a URL parameter.
func parseUUID(r *http.Request, param string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, param))
}

// ListGameSystems handles GET /api/game-systems
func (h *Handler) ListGameSystems(w http.ResponseWriter, r *http.Request) {
	systems, err := h.db.ListGameSystems(r.Context())
	if err != nil {
		log.Printf("Error listing game systems: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list game systems")
		return
	}

	if systems == nil {
		systems = []models.GameSystem{}
	}

	respondJSON(w, http.StatusOK, systems)
}

// ListCampaigns handles GET /api/campaigns
func (h *Handler) ListCampaigns(w http.ResponseWriter, r *http.Request) {
	campaigns, err := h.db.ListCampaigns(r.Context())
	if err != nil {
		log.Printf("Error listing campaigns: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list campaigns")
		return
	}

	if campaigns == nil {
		campaigns = []models.Campaign{}
	}

	respondJSON(w, http.StatusOK, campaigns)
}

// CreateCampaign handles POST /api/campaigns
func (h *Handler) CreateCampaign(w http.ResponseWriter, r *http.Request) {
	var req models.CreateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Name is required")
		return
	}

	campaign, err := h.db.CreateCampaign(r.Context(), req)
	if err != nil {
		log.Printf("Error creating campaign: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create campaign")
		return
	}

	respondJSON(w, http.StatusCreated, campaign)
}

// GetCampaign handles GET /api/campaigns/:id
func (h *Handler) GetCampaign(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	campaign, err := h.db.GetCampaign(r.Context(), id)
	if err != nil {
		log.Printf("Error getting campaign: %v", err)
		respondError(w, http.StatusNotFound, "Campaign not found")
		return
	}

	respondJSON(w, http.StatusOK, campaign)
}

// UpdateCampaign handles PUT /api/campaigns/:id
func (h *Handler) UpdateCampaign(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	var req models.UpdateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	campaign, err := h.db.UpdateCampaign(r.Context(), id, req)
	if err != nil {
		log.Printf("Error updating campaign: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update campaign")
		return
	}

	respondJSON(w, http.StatusOK, campaign)
}

// DeleteCampaign handles DELETE /api/campaigns/:id
func (h *Handler) DeleteCampaign(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	if err := h.db.DeleteCampaign(r.Context(), id); err != nil {
		log.Printf("Error deleting campaign: %v", err)
		respondError(w, http.StatusNotFound, "Campaign not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListEntities handles GET /api/campaigns/:id/entities
func (h *Handler) ListEntities(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Check for entity_type query parameter
	entityType := r.URL.Query().Get("type")
	var entities []models.Entity

	if entityType != "" {
		entities, err = h.db.ListEntitiesByType(r.Context(), campaignID, models.EntityType(entityType))
	} else {
		entities, err = h.db.ListEntitiesByCampaign(r.Context(), campaignID)
	}

	if err != nil {
		log.Printf("Error listing entities: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list entities")
		return
	}

	if entities == nil {
		entities = []models.Entity{}
	}

	respondJSON(w, http.StatusOK, entities)
}

// CreateEntity handles POST /api/campaigns/:id/entities
func (h *Handler) CreateEntity(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	var req models.CreateEntityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Name is required")
		return
	}

	if req.EntityType == "" {
		respondError(w, http.StatusBadRequest, "Entity type is required")
		return
	}

	entity, err := h.db.CreateEntity(r.Context(), campaignID, req)
	if err != nil {
		log.Printf("Error creating entity: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create entity")
		return
	}

	respondJSON(w, http.StatusCreated, entity)
}

// GetEntity handles GET /api/entities/:id
func (h *Handler) GetEntity(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid entity ID")
		return
	}

	entity, err := h.db.GetEntity(r.Context(), id)
	if err != nil {
		log.Printf("Error getting entity: %v", err)
		respondError(w, http.StatusNotFound, "Entity not found")
		return
	}

	respondJSON(w, http.StatusOK, entity)
}

// UpdateEntity handles PUT /api/entities/:id
func (h *Handler) UpdateEntity(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid entity ID")
		return
	}

	var req models.UpdateEntityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	entity, err := h.db.UpdateEntity(r.Context(), id, req)
	if err != nil {
		log.Printf("Error updating entity: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update entity")
		return
	}

	respondJSON(w, http.StatusOK, entity)
}

// DeleteEntity handles DELETE /api/entities/:id
func (h *Handler) DeleteEntity(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid entity ID")
		return
	}

	if err := h.db.DeleteEntity(r.Context(), id); err != nil {
		log.Printf("Error deleting entity: %v", err)
		respondError(w, http.StatusNotFound, "Entity not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListRelationships handles GET /api/campaigns/:id/relationships
func (h *Handler) ListRelationships(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	relationships, err := h.db.ListRelationshipsByCampaign(r.Context(), campaignID)
	if err != nil {
		log.Printf("Error listing relationships: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list relationships")
		return
	}

	if relationships == nil {
		relationships = []models.Relationship{}
	}

	respondJSON(w, http.StatusOK, relationships)
}

// CreateRelationship handles POST /api/campaigns/:id/relationships
func (h *Handler) CreateRelationship(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	var req models.CreateRelationshipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.SourceEntityID == uuid.Nil {
		respondError(w, http.StatusBadRequest, "Source entity ID is required")
		return
	}

	if req.TargetEntityID == uuid.Nil {
		respondError(w, http.StatusBadRequest, "Target entity ID is required")
		return
	}

	if req.RelationshipType == "" {
		respondError(w, http.StatusBadRequest, "Relationship type is required")
		return
	}

	relationship, err := h.db.CreateRelationship(r.Context(), campaignID, req)
	if err != nil {
		log.Printf("Error creating relationship: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create relationship")
		return
	}

	respondJSON(w, http.StatusCreated, relationship)
}

// ListTimelineEvents handles GET /api/campaigns/:id/timeline
func (h *Handler) ListTimelineEvents(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	events, err := h.db.ListTimelineEventsByCampaign(r.Context(), campaignID)
	if err != nil {
		log.Printf("Error listing timeline events: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list timeline events")
		return
	}

	if events == nil {
		events = []models.TimelineEvent{}
	}

	respondJSON(w, http.StatusOK, events)
}

// CreateTimelineEvent handles POST /api/campaigns/:id/timeline
func (h *Handler) CreateTimelineEvent(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	var req models.CreateTimelineEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Description == "" {
		respondError(w, http.StatusBadRequest, "Description is required")
		return
	}

	event, err := h.db.CreateTimelineEvent(r.Context(), campaignID, req)
	if err != nil {
		log.Printf("Error creating timeline event: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create timeline event")
		return
	}

	respondJSON(w, http.StatusCreated, event)
}

// GetStats handles GET /api/stats
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.db.GetDashboardStats(r.Context())
	if err != nil {
		log.Printf("Error getting stats: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get statistics")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// GetDashboardStats handles GET /api/stats/dashboard
func (h *Handler) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.db.GetFrontendDashboardStats(r.Context())
	if err != nil {
		log.Printf("Error getting dashboard stats: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get dashboard statistics")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}
