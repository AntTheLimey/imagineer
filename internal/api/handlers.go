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

	"github.com/antonypegg/imagineer/internal/auth"
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

// parseUUID parses a UUID from the URL parameter named by param.
// It returns the parsed uuid.UUID, or an error if the parameter is missing or not a valid UUID.
func parseUUID(r *http.Request, param string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, param))
}

// verifyCampaignOwnership checks if the authenticated user owns the campaign.
// Returns the user ID if successful, or writes an error response and returns
// uuid.Nil if verification fails.
func (h *Handler) verifyCampaignOwnership(w http.ResponseWriter, r *http.Request, campaignID uuid.UUID) (uuid.UUID, bool) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return uuid.Nil, false
	}

	if err := h.db.VerifyCampaignOwnership(r.Context(), campaignID, userID); err != nil {
		respondError(w, http.StatusNotFound, "Campaign not found")
		return uuid.Nil, false
	}

	return userID, true
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

// GetGameSystem handles GET /api/game-systems/{id}
func (h *Handler) GetGameSystem(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid game system ID")
		return
	}

	system, err := h.db.GetGameSystem(r.Context(), id)
	if err != nil {
		log.Printf("Error getting game system: %v", err)
		respondError(w, http.StatusNotFound, "Game system not found")
		return
	}

	respondJSON(w, http.StatusOK, system)
}

// GetGameSystemByCode handles GET /api/game-systems/code/{code}
func (h *Handler) GetGameSystemByCode(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		respondError(w, http.StatusBadRequest, "Game system code is required")
		return
	}

	system, err := h.db.GetGameSystemByCode(r.Context(), code)
	if err != nil {
		log.Printf("Error getting game system by code: %v", err)
		respondError(w, http.StatusNotFound, "Game system not found")
		return
	}

	respondJSON(w, http.StatusOK, system)
}

// ListCampaigns handles GET /api/campaigns
// Returns only campaigns owned by the authenticated user.
func (h *Handler) ListCampaigns(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	campaigns, err := h.db.ListCampaignsByOwner(r.Context(), userID)
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
// Creates a new campaign owned by the authenticated user.
func (h *Handler) CreateCampaign(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req models.CreateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Name is required")
		return
	}

	campaign, err := h.db.CreateCampaignWithOwner(r.Context(), req, userID)
	if err != nil {
		log.Printf("Error creating campaign: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create campaign")
		return
	}

	respondJSON(w, http.StatusCreated, campaign)
}

// GetCampaign handles GET /api/campaigns/:id
// Returns the campaign only if it belongs to the authenticated user.
func (h *Handler) GetCampaign(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	id, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	campaign, err := h.db.GetCampaignByOwner(r.Context(), id, userID)
	if err != nil {
		log.Printf("Error getting campaign: %v", err)
		respondError(w, http.StatusNotFound, "Campaign not found")
		return
	}

	respondJSON(w, http.StatusOK, campaign)
}

// UpdateCampaign handles PUT /api/campaigns/:id
// Updates the campaign only if it belongs to the authenticated user.
func (h *Handler) UpdateCampaign(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

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

	campaign, err := h.db.UpdateCampaignByOwner(r.Context(), id, userID, req)
	if err != nil {
		log.Printf("Error updating campaign: %v", err)
		respondError(w, http.StatusNotFound, "Campaign not found")
		return
	}

	respondJSON(w, http.StatusOK, campaign)
}

// DeleteCampaign handles DELETE /api/campaigns/:id
// Deletes the campaign only if it belongs to the authenticated user.
func (h *Handler) DeleteCampaign(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	id, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	if err := h.db.DeleteCampaignByOwner(r.Context(), id, userID); err != nil {
		log.Printf("Error deleting campaign: %v", err)
		respondError(w, http.StatusNotFound, "Campaign not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListEntities handles GET /api/campaigns/:id/entities
// Verifies the user owns the campaign before listing entities.
func (h *Handler) ListEntities(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
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
// Verifies the user owns the campaign before creating an entity.
func (h *Handler) CreateEntity(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
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
// Verifies the user owns the entity's campaign before returning the entity.
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

	// Verify the user owns the entity's campaign
	if _, ok := h.verifyCampaignOwnership(w, r, entity.CampaignID); !ok {
		return
	}

	respondJSON(w, http.StatusOK, entity)
}

// UpdateEntity handles PUT /api/entities/:id
// Verifies the user owns the entity's campaign before updating the entity.
func (h *Handler) UpdateEntity(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid entity ID")
		return
	}

	// Fetch the entity first to get its campaign ID
	existingEntity, err := h.db.GetEntity(r.Context(), id)
	if err != nil {
		log.Printf("Error getting entity: %v", err)
		respondError(w, http.StatusNotFound, "Entity not found")
		return
	}

	// Verify the user owns the entity's campaign
	if _, ok := h.verifyCampaignOwnership(w, r, existingEntity.CampaignID); !ok {
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
// Verifies the user owns the entity's campaign before deleting the entity.
func (h *Handler) DeleteEntity(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid entity ID")
		return
	}

	// Fetch the entity first to get its campaign ID
	existingEntity, err := h.db.GetEntity(r.Context(), id)
	if err != nil {
		log.Printf("Error getting entity: %v", err)
		respondError(w, http.StatusNotFound, "Entity not found")
		return
	}

	// Verify the user owns the entity's campaign
	if _, ok := h.verifyCampaignOwnership(w, r, existingEntity.CampaignID); !ok {
		return
	}

	if err := h.db.DeleteEntity(r.Context(), id); err != nil {
		log.Printf("Error deleting entity: %v", err)
		respondError(w, http.StatusNotFound, "Entity not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SearchEntities handles GET /api/campaigns/{campaignId}/entities/search
// Verifies the user owns the campaign before searching entities.
func (h *Handler) SearchEntities(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		respondError(w, http.StatusBadRequest, "Name query parameter is required")
		return
	}

	// Default limit of 10 results
	limit := 10

	entities, err := h.db.SearchEntitiesByName(r.Context(), campaignID, name, limit)
	if err != nil {
		log.Printf("Error searching entities: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to search entities")
		return
	}

	if entities == nil {
		entities = []models.Entity{}
	}

	respondJSON(w, http.StatusOK, entities)
}

// ListRelationships handles GET /api/campaigns/:id/relationships
// Verifies the user owns the campaign before listing relationships.
func (h *Handler) ListRelationships(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
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
// Verifies the user owns the campaign before creating a relationship.
func (h *Handler) CreateRelationship(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
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

// GetRelationship handles GET /api/campaigns/{campaignId}/relationships/{id}
// Verifies the user owns the campaign before getting the relationship.
func (h *Handler) GetRelationship(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	relationshipID, err := parseUUID(r, "relationshipId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid relationship ID")
		return
	}

	relationship, err := h.db.GetRelationship(r.Context(), relationshipID)
	if err != nil {
		log.Printf("Error getting relationship: %v", err)
		respondError(w, http.StatusNotFound, "Relationship not found")
		return
	}

	// Verify the relationship belongs to the specified campaign
	if relationship.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Relationship not found")
		return
	}

	respondJSON(w, http.StatusOK, relationship)
}

// UpdateRelationship handles PUT /api/campaigns/{campaignId}/relationships/{id}
// Verifies the user owns the campaign before updating the relationship.
func (h *Handler) UpdateRelationship(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	relationshipID, err := parseUUID(r, "relationshipId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid relationship ID")
		return
	}

	// Verify the relationship belongs to the specified campaign before updating
	existing, err := h.db.GetRelationship(r.Context(), relationshipID)
	if err != nil {
		log.Printf("Error getting relationship: %v", err)
		respondError(w, http.StatusNotFound, "Relationship not found")
		return
	}

	if existing.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Relationship not found")
		return
	}

	var req models.UpdateRelationshipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	relationship, err := h.db.UpdateRelationship(r.Context(), relationshipID, req)
	if err != nil {
		log.Printf("Error updating relationship: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update relationship")
		return
	}

	respondJSON(w, http.StatusOK, relationship)
}

// DeleteRelationship handles DELETE /api/campaigns/{campaignId}/relationships/{id}
// Verifies the user owns the campaign before deleting the relationship.
func (h *Handler) DeleteRelationship(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	relationshipID, err := parseUUID(r, "relationshipId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid relationship ID")
		return
	}

	// Verify the relationship belongs to the specified campaign before deleting
	existing, err := h.db.GetRelationship(r.Context(), relationshipID)
	if err != nil {
		log.Printf("Error getting relationship: %v", err)
		respondError(w, http.StatusNotFound, "Relationship not found")
		return
	}

	if existing.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Relationship not found")
		return
	}

	if err := h.db.DeleteRelationship(r.Context(), relationshipID); err != nil {
		log.Printf("Error deleting relationship: %v", err)
		respondError(w, http.StatusNotFound, "Relationship not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetEntityRelationships handles GET /api/campaigns/{campaignId}/entities/{entityId}/relationships
// Verifies the user owns the campaign before getting entity relationships.
func (h *Handler) GetEntityRelationships(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	entityID, err := parseUUID(r, "entityId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid entity ID")
		return
	}

	// Verify the entity belongs to the specified campaign
	entity, err := h.db.GetEntity(r.Context(), entityID)
	if err != nil {
		log.Printf("Error getting entity: %v", err)
		respondError(w, http.StatusNotFound, "Entity not found")
		return
	}

	if entity.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Entity not found")
		return
	}

	relationships, err := h.db.GetEntityRelationships(r.Context(), entityID)
	if err != nil {
		log.Printf("Error getting entity relationships: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get entity relationships")
		return
	}

	if relationships == nil {
		relationships = []models.Relationship{}
	}

	respondJSON(w, http.StatusOK, relationships)
}

// ListTimelineEvents handles GET /api/campaigns/:id/timeline
// Verifies the user owns the campaign before listing timeline events.
func (h *Handler) ListTimelineEvents(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
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
// Verifies the user owns the campaign before creating a timeline event.
func (h *Handler) CreateTimelineEvent(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
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

// GetTimelineEvent handles GET /api/campaigns/{campaignId}/timeline/{id}
// Verifies the user owns the campaign before getting the timeline event.
func (h *Handler) GetTimelineEvent(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	eventID, err := parseUUID(r, "eventId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid timeline event ID")
		return
	}

	event, err := h.db.GetTimelineEvent(r.Context(), eventID)
	if err != nil {
		log.Printf("Error getting timeline event: %v", err)
		respondError(w, http.StatusNotFound, "Timeline event not found")
		return
	}

	// Verify the event belongs to the specified campaign
	if event.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Timeline event not found")
		return
	}

	respondJSON(w, http.StatusOK, event)
}

// UpdateTimelineEvent handles PUT /api/campaigns/{campaignId}/timeline/{id}
// Verifies the user owns the campaign before updating the timeline event.
func (h *Handler) UpdateTimelineEvent(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	eventID, err := parseUUID(r, "eventId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid timeline event ID")
		return
	}

	// Verify the event belongs to the specified campaign before updating
	existing, err := h.db.GetTimelineEvent(r.Context(), eventID)
	if err != nil {
		log.Printf("Error getting timeline event: %v", err)
		respondError(w, http.StatusNotFound, "Timeline event not found")
		return
	}

	if existing.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Timeline event not found")
		return
	}

	var req models.UpdateTimelineEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	event, err := h.db.UpdateTimelineEvent(r.Context(), eventID, req)
	if err != nil {
		log.Printf("Error updating timeline event: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update timeline event")
		return
	}

	respondJSON(w, http.StatusOK, event)
}

// DeleteTimelineEvent handles DELETE /api/campaigns/{campaignId}/timeline/{id}
// Verifies the user owns the campaign before deleting the timeline event.
func (h *Handler) DeleteTimelineEvent(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	eventID, err := parseUUID(r, "eventId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid timeline event ID")
		return
	}

	// Verify the event belongs to the specified campaign before deleting
	existing, err := h.db.GetTimelineEvent(r.Context(), eventID)
	if err != nil {
		log.Printf("Error getting timeline event: %v", err)
		respondError(w, http.StatusNotFound, "Timeline event not found")
		return
	}

	if existing.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Timeline event not found")
		return
	}

	if err := h.db.DeleteTimelineEvent(r.Context(), eventID); err != nil {
		log.Printf("Error deleting timeline event: %v", err)
		respondError(w, http.StatusNotFound, "Timeline event not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetEntityTimelineEvents handles GET /api/campaigns/{campaignId}/entities/{entityId}/timeline
// Verifies the user owns the campaign before getting entity timeline events.
func (h *Handler) GetEntityTimelineEvents(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	entityID, err := parseUUID(r, "entityId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid entity ID")
		return
	}

	// Verify the entity belongs to the specified campaign
	entity, err := h.db.GetEntity(r.Context(), entityID)
	if err != nil {
		log.Printf("Error getting entity: %v", err)
		respondError(w, http.StatusNotFound, "Entity not found")
		return
	}

	if entity.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Entity not found")
		return
	}

	events, err := h.db.GetTimelineEventsForEntity(r.Context(), entityID)
	if err != nil {
		log.Printf("Error getting entity timeline events: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get entity timeline events")
		return
	}

	if events == nil {
		events = []models.TimelineEvent{}
	}

	respondJSON(w, http.StatusOK, events)
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

// GetCampaignStats handles GET /api/campaigns/{campaignId}/stats
// Verifies the user owns the campaign before getting stats.
func (h *Handler) GetCampaignStats(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	stats, err := h.db.GetCampaignStats(r.Context(), campaignID)
	if err != nil {
		log.Printf("Error getting campaign stats: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get campaign statistics")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}