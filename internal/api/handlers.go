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
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/antonypegg/imagineer/internal/analysis"
	"github.com/antonypegg/imagineer/internal/auth"
	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/models"
	"github.com/go-chi/chi/v5"
)

// Handler provides HTTP handlers for the API.
type Handler struct {
	db       *database.DB
	analyzer *analysis.Analyzer
}

// NewHandler creates a new Handler with the given database connection.
func NewHandler(db *database.DB) *Handler {
	return &Handler{
		db:       db,
		analyzer: analysis.NewAnalyzer(db),
	}
}

// EntityWithAnalysis wraps an entity response with optional analysis metadata.
type EntityWithAnalysis struct {
	*models.Entity
	Analysis *models.AnalysisSummary `json:"_analysis,omitempty"`
}

// CampaignWithAnalysis wraps a campaign response with optional analysis metadata.
type CampaignWithAnalysis struct {
	*models.Campaign
	Analysis *models.AnalysisSummary `json:"_analysis,omitempty"`
}

// ChapterWithAnalysis wraps a chapter response with optional analysis metadata.
type ChapterWithAnalysis struct {
	*models.Chapter
	Analysis *models.AnalysisSummary `json:"_analysis,omitempty"`
}

// SessionWithAnalysis wraps a session response with optional analysis metadata.
type SessionWithAnalysis struct {
	*models.Session
	Analysis *models.AnalysisSummary `json:"_analysis,omitempty"`
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

// parseInt64 parses an int64 from the URL parameter named by param.
// It returns the parsed int64, or an error if the parameter is missing or
// not a valid integer.
func parseInt64(r *http.Request, param string) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, param), 10, 64)
}

// verifyCampaignOwnership checks if the authenticated user owns the campaign.
// Returns the user ID if successful, or writes an error response and returns
// 0 if verification fails.
func (h *Handler) verifyCampaignOwnership(w http.ResponseWriter, r *http.Request, campaignID int64) (int64, bool) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return 0, false
	}

	if err := h.db.VerifyCampaignOwnership(r.Context(), campaignID, userID); err != nil {
		respondError(w, http.StatusNotFound, "Campaign not found")
		return 0, false
	}

	return userID, true
}

// isUserCampaignOwner checks if the given user ID matches the campaign's owner.
// This is used to determine if GM-only content (like gm_notes) should be visible.
func (h *Handler) isUserCampaignOwner(ctx context.Context, campaignID int64, userID int64) bool {
	err := h.db.VerifyCampaignOwnership(ctx, campaignID, userID)
	return err == nil
}

// filterEntityGMNotes removes GM notes from an entity if the user is not the
// campaign owner. This protects sensitive GM-only content from being exposed
// to players who may have read access to the campaign.
func filterEntityGMNotes(entity *models.Entity, isGM bool) {
	if !isGM {
		entity.GMNotes = nil
	}
}

// filterEntitiesGMNotes removes GM notes from a slice of entities if the user
// is not the campaign owner. Modifies the entities in place.
func filterEntitiesGMNotes(entities []models.Entity, isGM bool) {
	if isGM {
		return
	}
	for i := range entities {
		entities[i].GMNotes = nil
	}
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
	id, err := parseInt64(r, "id")
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

	id, err := parseInt64(r, "id")
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

	id, err := parseInt64(r, "id")
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

	response := CampaignWithAnalysis{Campaign: campaign}
	if req.Description != nil {
		content := ""
		if req.Description != nil {
			content = *req.Description
		}
		job, items, analyzeErr := h.analyzer.AnalyzeContent(
			r.Context(), campaign.ID,
			"campaigns", "description", campaign.ID,
			content,
		)
		if analyzeErr != nil {
			log.Printf("Content analysis failed for campaign %d: %v", campaign.ID, analyzeErr)
		} else if len(items) > 0 {
			response.Analysis = &models.AnalysisSummary{
				JobID:        job.ID,
				PendingCount: job.TotalItems,
			}
		}
	}
	respondJSON(w, http.StatusOK, response)
}

// DeleteCampaign handles DELETE /api/campaigns/:id
// Deletes the campaign only if it belongs to the authenticated user.
func (h *Handler) DeleteCampaign(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	id, err := parseInt64(r, "id")
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
// GM notes are only visible to the campaign owner (GM).
func (h *Handler) ListEntities(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	userID, ok := h.verifyCampaignOwnership(w, r, campaignID)
	if !ok {
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

	// Filter GM notes for non-owners
	isGM := h.isUserCampaignOwner(r.Context(), campaignID, userID)
	filterEntitiesGMNotes(entities, isGM)

	respondJSON(w, http.StatusOK, entities)
}

// CreateEntity handles POST /api/campaigns/:id/entities
// Verifies the user owns the campaign before creating an entity.
func (h *Handler) CreateEntity(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
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
// GM notes are only visible to the campaign owner (GM).
func (h *Handler) GetEntity(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64(r, "id")
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
	userID, ok := h.verifyCampaignOwnership(w, r, entity.CampaignID)
	if !ok {
		return
	}

	// Filter GM notes for non-owners
	isGM := h.isUserCampaignOwner(r.Context(), entity.CampaignID, userID)
	filterEntityGMNotes(entity, isGM)

	respondJSON(w, http.StatusOK, entity)
}

// UpdateEntity handles PUT /api/entities/:id
// Verifies the user owns the entity's campaign before updating the entity.
func (h *Handler) UpdateEntity(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64(r, "id")
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

	// Trigger content analysis if description changed
	response := EntityWithAnalysis{Entity: entity}
	if req.Description != nil {
		content := ""
		if req.Description != nil {
			content = *req.Description
		}
		job, items, analyzeErr := h.analyzer.AnalyzeContent(
			r.Context(), entity.CampaignID,
			"entities", "description", entity.ID,
			content,
		)
		if analyzeErr != nil {
			log.Printf("Content analysis failed for entity %d description: %v", entity.ID, analyzeErr)
		} else if len(items) > 0 {
			response.Analysis = &models.AnalysisSummary{
				JobID:        job.ID,
				PendingCount: job.TotalItems,
			}
		}
	}
	// Also analyze GM notes if changed
	if req.GMNotes != nil {
		content := ""
		if req.GMNotes != nil {
			content = *req.GMNotes
		}
		job, items, analyzeErr := h.analyzer.AnalyzeContent(
			r.Context(), entity.CampaignID,
			"entities", "gm_notes", entity.ID,
			content,
		)
		if analyzeErr != nil {
			log.Printf("Content analysis failed for entity %d gm_notes: %v", entity.ID, analyzeErr)
		} else if response.Analysis == nil && len(items) > 0 {
			response.Analysis = &models.AnalysisSummary{
				JobID:        job.ID,
				PendingCount: job.TotalItems,
			}
		}
	}
	respondJSON(w, http.StatusOK, response)
}

// DeleteEntity handles DELETE /api/entities/:id
// Verifies the user owns the entity's campaign before deleting the entity.
func (h *Handler) DeleteEntity(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64(r, "id")
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
// GM notes are only visible to the campaign owner (GM).
func (h *Handler) SearchEntities(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	userID, ok := h.verifyCampaignOwnership(w, r, campaignID)
	if !ok {
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

	// Filter GM notes for non-owners
	isGM := h.isUserCampaignOwner(r.Context(), campaignID, userID)
	filterEntitiesGMNotes(entities, isGM)

	respondJSON(w, http.StatusOK, entities)
}

// ListRelationships handles GET /api/campaigns/:id/relationships
// Verifies the user owns the campaign before listing relationships.
func (h *Handler) ListRelationships(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
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
	campaignID, err := parseInt64(r, "id")
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

	if req.SourceEntityID == 0 {
		respondError(w, http.StatusBadRequest, "Source entity ID is required")
		return
	}

	if req.TargetEntityID == 0 {
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
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	relationshipID, err := parseInt64(r, "relationshipId")
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
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	relationshipID, err := parseInt64(r, "relationshipId")
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
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	relationshipID, err := parseInt64(r, "relationshipId")
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
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	entityID, err := parseInt64(r, "entityId")
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
	campaignID, err := parseInt64(r, "id")
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
	campaignID, err := parseInt64(r, "id")
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
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	eventID, err := parseInt64(r, "eventId")
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
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	eventID, err := parseInt64(r, "eventId")
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
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	eventID, err := parseInt64(r, "eventId")
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
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	entityID, err := parseInt64(r, "entityId")
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
	campaignID, err := parseInt64(r, "id")
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

// ListRelationshipTypes handles GET /api/campaigns/{id}/relationship-types
// Returns all relationship types available for a campaign (system defaults + custom).
func (h *Handler) ListRelationshipTypes(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	types, err := h.db.ListRelationshipTypes(r.Context(), campaignID)
	if err != nil {
		log.Printf("Error listing relationship types: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list relationship types")
		return
	}

	if types == nil {
		types = []models.RelationshipType{}
	}

	respondJSON(w, http.StatusOK, types)
}

// CreateRelationshipType handles POST /api/campaigns/{id}/relationship-types
// Creates a custom relationship type for a campaign.
func (h *Handler) CreateRelationshipType(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	var req models.CreateRelationshipTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Name is required")
		return
	}

	if req.InverseName == "" {
		respondError(w, http.StatusBadRequest, "Inverse name is required")
		return
	}

	if req.DisplayLabel == "" {
		respondError(w, http.StatusBadRequest, "Display label is required")
		return
	}

	if req.InverseDisplayLabel == "" {
		respondError(w, http.StatusBadRequest, "Inverse display label is required")
		return
	}

	// Validate symmetric constraint: symmetric types must have matching names
	if req.IsSymmetric && req.Name != req.InverseName {
		respondError(w, http.StatusBadRequest, "Symmetric relationship types must have matching name and inverse name")
		return
	}

	relType, err := h.db.CreateRelationshipType(r.Context(), campaignID, req)
	if err != nil {
		log.Printf("Error creating relationship type: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create relationship type")
		return
	}

	respondJSON(w, http.StatusCreated, relType)
}

// DeleteRelationshipType handles DELETE /api/campaigns/{campaignId}/relationship-types/{id}
// Deletes a custom relationship type. System defaults cannot be deleted.
func (h *Handler) DeleteRelationshipType(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	typeID, err := parseInt64(r, "typeId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid relationship type ID")
		return
	}

	// Verify the relationship type exists and belongs to this campaign
	relType, err := h.db.GetRelationshipType(r.Context(), typeID)
	if err != nil {
		log.Printf("Error getting relationship type: %v", err)
		respondError(w, http.StatusNotFound, "Relationship type not found")
		return
	}

	// Check if it's a system default (campaign_id is nil)
	if relType.CampaignID == nil {
		respondError(w, http.StatusForbidden, "Cannot delete system default relationship types")
		return
	}

	// Check if the type belongs to the specified campaign
	if *relType.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Relationship type not found")
		return
	}

	if err := h.db.DeleteRelationshipType(r.Context(), typeID); err != nil {
		log.Printf("Error deleting relationship type: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to delete relationship type")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// maskAPIKey masks an API key, showing only the last 4 characters.
// Returns nil if the input is nil.
func maskAPIKey(key *string) *string {
	if key == nil || *key == "" {
		return nil
	}
	k := *key
	if len(k) <= 4 {
		masked := "****"
		return &masked
	}
	masked := "****" + k[len(k)-4:]
	return &masked
}

// GetUserSettings handles GET /api/user/settings
// Returns the current user's settings with masked API keys.
func (h *Handler) GetUserSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	settings, err := h.db.GetUserSettings(r.Context(), userID)
	if err != nil {
		log.Printf("Error getting user settings: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get user settings")
		return
	}

	// If no settings exist, return an empty settings object
	if settings == nil {
		settings = &models.UserSettings{
			UserID: userID,
		}
	}

	// Build response with masked API keys
	response := models.UserSettingsResponse{
		UserID:            settings.UserID,
		ContentGenService: settings.ContentGenService,
		ContentGenAPIKey:  maskAPIKey(settings.ContentGenAPIKey),
		EmbeddingService:  settings.EmbeddingService,
		EmbeddingAPIKey:   maskAPIKey(settings.EmbeddingAPIKey),
		ImageGenService:   settings.ImageGenService,
		ImageGenAPIKey:    maskAPIKey(settings.ImageGenAPIKey),
		CreatedAt:         settings.CreatedAt,
		UpdatedAt:         settings.UpdatedAt,
	}

	respondJSON(w, http.StatusOK, response)
}

// isMaskedAPIKey checks if a value appears to be a masked API key.
func isMaskedAPIKey(key *string) bool {
	if key == nil {
		return false
	}
	k := *key
	return len(k) >= 4 && k[:4] == "****"
}

// UpdateUserSettings handles PUT /api/user/settings
// Updates the current user's settings.
func (h *Handler) UpdateUserSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req models.UpdateUserSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Clear masked API keys from request to prevent overwriting with masked values
	if isMaskedAPIKey(req.ContentGenAPIKey) {
		req.ContentGenAPIKey = nil
	}
	if isMaskedAPIKey(req.EmbeddingAPIKey) {
		req.EmbeddingAPIKey = nil
	}
	if isMaskedAPIKey(req.ImageGenAPIKey) {
		req.ImageGenAPIKey = nil
	}

	settings, err := h.db.UpdateUserSettings(r.Context(), userID, req)
	if err != nil {
		log.Printf("Error updating user settings: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update user settings")
		return
	}

	// Build response with masked API keys
	response := models.UserSettingsResponse{
		UserID:            settings.UserID,
		ContentGenService: settings.ContentGenService,
		ContentGenAPIKey:  maskAPIKey(settings.ContentGenAPIKey),
		EmbeddingService:  settings.EmbeddingService,
		EmbeddingAPIKey:   maskAPIKey(settings.EmbeddingAPIKey),
		ImageGenService:   settings.ImageGenService,
		ImageGenAPIKey:    maskAPIKey(settings.ImageGenAPIKey),
		CreatedAt:         settings.CreatedAt,
		UpdatedAt:         settings.UpdatedAt,
	}

	respondJSON(w, http.StatusOK, response)
}

// ListPlayerCharacters handles GET /api/campaigns/{id}/player-characters
// Returns all player characters for a campaign.
func (h *Handler) ListPlayerCharacters(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	characters, err := h.db.ListPlayerCharacters(r.Context(), campaignID)
	if err != nil {
		log.Printf("Error listing player characters: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list player characters")
		return
	}

	if characters == nil {
		characters = []models.PlayerCharacter{}
	}

	respondJSON(w, http.StatusOK, characters)
}

// CreatePlayerCharacter handles POST /api/campaigns/{id}/player-characters
// Creates a new player character in a campaign.
func (h *Handler) CreatePlayerCharacter(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	var req models.CreatePlayerCharacterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.CharacterName == "" {
		respondError(w, http.StatusBadRequest, "Character name is required")
		return
	}

	if req.PlayerName == "" {
		respondError(w, http.StatusBadRequest, "Player name is required")
		return
	}

	character, err := h.db.CreatePlayerCharacter(r.Context(), campaignID, req)
	if err != nil {
		log.Printf("Error creating player character: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create player character")
		return
	}

	respondJSON(w, http.StatusCreated, character)
}

// GetPlayerCharacter handles GET /api/campaigns/{id}/player-characters/{pcId}
// Returns a specific player character.
func (h *Handler) GetPlayerCharacter(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	pcID, err := parseInt64(r, "pcId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid player character ID")
		return
	}

	character, err := h.db.GetPlayerCharacter(r.Context(), pcID)
	if err != nil {
		log.Printf("Error getting player character: %v", err)
		respondError(w, http.StatusNotFound, "Player character not found")
		return
	}

	// Verify the character belongs to the specified campaign
	if character.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Player character not found")
		return
	}

	respondJSON(w, http.StatusOK, character)
}

// UpdatePlayerCharacter handles PUT /api/campaigns/{id}/player-characters/{pcId}
// Updates a player character.
func (h *Handler) UpdatePlayerCharacter(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	pcID, err := parseInt64(r, "pcId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid player character ID")
		return
	}

	// Verify the character exists and belongs to the specified campaign
	existing, err := h.db.GetPlayerCharacter(r.Context(), pcID)
	if err != nil {
		log.Printf("Error getting player character: %v", err)
		respondError(w, http.StatusNotFound, "Player character not found")
		return
	}

	if existing.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Player character not found")
		return
	}

	var req models.UpdatePlayerCharacterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	character, err := h.db.UpdatePlayerCharacter(r.Context(), pcID, req)
	if err != nil {
		log.Printf("Error updating player character: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update player character")
		return
	}

	respondJSON(w, http.StatusOK, character)
}

// DeletePlayerCharacter handles DELETE /api/campaigns/{id}/player-characters/{pcId}
// Deletes a player character.
func (h *Handler) DeletePlayerCharacter(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	pcID, err := parseInt64(r, "pcId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid player character ID")
		return
	}

	// Verify the character exists and belongs to the specified campaign
	existing, err := h.db.GetPlayerCharacter(r.Context(), pcID)
	if err != nil {
		log.Printf("Error getting player character: %v", err)
		respondError(w, http.StatusNotFound, "Player character not found")
		return
	}

	if existing.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Player character not found")
		return
	}

	if err := h.db.DeletePlayerCharacter(r.Context(), pcID); err != nil {
		log.Printf("Error deleting player character: %v", err)
		respondError(w, http.StatusNotFound, "Player character not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListChapters handles GET /api/campaigns/{id}/chapters
// Returns all chapters for a campaign ordered by sort_order.
func (h *Handler) ListChapters(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	chapters, err := h.db.ListChaptersByCampaign(r.Context(), campaignID)
	if err != nil {
		log.Printf("Error listing chapters: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list chapters")
		return
	}

	if chapters == nil {
		chapters = []models.Chapter{}
	}

	respondJSON(w, http.StatusOK, chapters)
}

// GetChapter handles GET /api/campaigns/{id}/chapters/{chapterId}
// Returns a specific chapter.
func (h *Handler) GetChapter(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	chapterID, err := parseInt64(r, "chapterId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid chapter ID")
		return
	}

	chapter, err := h.db.GetChapter(r.Context(), chapterID)
	if err != nil {
		log.Printf("Error getting chapter: %v", err)
		respondError(w, http.StatusNotFound, "Chapter not found")
		return
	}

	// Verify the chapter belongs to the specified campaign
	if chapter.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Chapter not found")
		return
	}

	respondJSON(w, http.StatusOK, chapter)
}

// CreateChapter handles POST /api/campaigns/{id}/chapters
// Creates a new chapter in a campaign.
func (h *Handler) CreateChapter(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	var req models.CreateChapterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Title == "" {
		respondError(w, http.StatusBadRequest, "Title is required")
		return
	}

	chapter, err := h.db.CreateChapter(r.Context(), campaignID, req)
	if err != nil {
		log.Printf("Error creating chapter: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create chapter")
		return
	}

	respondJSON(w, http.StatusCreated, chapter)
}

// UpdateChapter handles PUT /api/campaigns/{id}/chapters/{chapterId}
// Updates a chapter.
func (h *Handler) UpdateChapter(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	chapterID, err := parseInt64(r, "chapterId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid chapter ID")
		return
	}

	// Verify the chapter exists and belongs to the specified campaign
	existing, err := h.db.GetChapter(r.Context(), chapterID)
	if err != nil {
		log.Printf("Error getting chapter: %v", err)
		respondError(w, http.StatusNotFound, "Chapter not found")
		return
	}

	if existing.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Chapter not found")
		return
	}

	var req models.UpdateChapterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	chapter, err := h.db.UpdateChapter(r.Context(), chapterID, req)
	if err != nil {
		log.Printf("Error updating chapter: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update chapter")
		return
	}

	response := ChapterWithAnalysis{Chapter: chapter}
	if req.Overview != nil {
		content := ""
		if req.Overview != nil {
			content = *req.Overview
		}
		job, items, analyzeErr := h.analyzer.AnalyzeContent(
			r.Context(), chapter.CampaignID,
			"chapters", "overview", chapter.ID,
			content,
		)
		if analyzeErr != nil {
			log.Printf("Content analysis failed for chapter %d: %v", chapter.ID, analyzeErr)
		} else if len(items) > 0 {
			response.Analysis = &models.AnalysisSummary{
				JobID:        job.ID,
				PendingCount: job.TotalItems,
			}
		}
	}
	respondJSON(w, http.StatusOK, response)
}

// DeleteChapter handles DELETE /api/campaigns/{id}/chapters/{chapterId}
// Deletes a chapter.
func (h *Handler) DeleteChapter(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	chapterID, err := parseInt64(r, "chapterId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid chapter ID")
		return
	}

	// Verify the chapter exists and belongs to the specified campaign
	existing, err := h.db.GetChapter(r.Context(), chapterID)
	if err != nil {
		log.Printf("Error getting chapter: %v", err)
		respondError(w, http.StatusNotFound, "Chapter not found")
		return
	}

	if existing.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Chapter not found")
		return
	}

	if err := h.db.DeleteChapter(r.Context(), chapterID); err != nil {
		log.Printf("Error deleting chapter: %v", err)
		respondError(w, http.StatusNotFound, "Chapter not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListSessions handles GET /api/campaigns/{id}/sessions
// Returns all sessions for a campaign.
func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	sessions, err := h.db.ListSessionsByCampaign(r.Context(), campaignID)
	if err != nil {
		log.Printf("Error listing sessions: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list sessions")
		return
	}

	if sessions == nil {
		sessions = []models.Session{}
	}

	respondJSON(w, http.StatusOK, sessions)
}

// ListSessionsByChapter handles GET /api/campaigns/{id}/chapters/{chapterId}/sessions
// Returns all sessions for a specific chapter.
func (h *Handler) ListSessionsByChapter(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	chapterID, err := parseInt64(r, "chapterId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid chapter ID")
		return
	}

	// Verify the chapter exists and belongs to the specified campaign
	chapter, err := h.db.GetChapter(r.Context(), chapterID)
	if err != nil {
		log.Printf("Error getting chapter: %v", err)
		respondError(w, http.StatusNotFound, "Chapter not found")
		return
	}

	if chapter.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Chapter not found")
		return
	}

	sessions, err := h.db.ListSessionsByChapter(r.Context(), chapterID)
	if err != nil {
		log.Printf("Error listing sessions by chapter: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list sessions")
		return
	}

	if sessions == nil {
		sessions = []models.Session{}
	}

	respondJSON(w, http.StatusOK, sessions)
}

// GetSession handles GET /api/campaigns/{id}/sessions/{sessionId}
// Returns a specific session.
func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	sessionID, err := parseInt64(r, "sessionId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	session, err := h.db.GetSession(r.Context(), sessionID)
	if err != nil {
		log.Printf("Error getting session: %v", err)
		respondError(w, http.StatusNotFound, "Session not found")
		return
	}

	// Verify the session belongs to the specified campaign
	if session.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Session not found")
		return
	}

	respondJSON(w, http.StatusOK, session)
}

// CreateSession handles POST /api/campaigns/{id}/sessions
// Creates a new session in a campaign.
func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	var req models.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// If a chapter ID is provided, verify it belongs to this campaign
	if req.ChapterID != nil {
		chapter, err := h.db.GetChapter(r.Context(), *req.ChapterID)
		if err != nil {
			log.Printf("Error getting chapter: %v", err)
			respondError(w, http.StatusBadRequest, "Invalid chapter ID")
			return
		}
		if chapter.CampaignID != campaignID {
			respondError(w, http.StatusBadRequest, "Chapter does not belong to this campaign")
			return
		}
	}

	session, err := h.db.CreateSession(r.Context(), campaignID, req)
	if err != nil {
		log.Printf("Error creating session: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	respondJSON(w, http.StatusCreated, session)
}

// UpdateSession handles PUT /api/campaigns/{id}/sessions/{sessionId}
// Updates a session.
func (h *Handler) UpdateSession(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	sessionID, err := parseInt64(r, "sessionId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	// Verify the session exists and belongs to the specified campaign
	existing, err := h.db.GetSession(r.Context(), sessionID)
	if err != nil {
		log.Printf("Error getting session: %v", err)
		respondError(w, http.StatusNotFound, "Session not found")
		return
	}

	if existing.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Session not found")
		return
	}

	var req models.UpdateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// If a chapter ID is provided, verify it belongs to this campaign
	if req.ChapterID != nil {
		chapter, err := h.db.GetChapter(r.Context(), *req.ChapterID)
		if err != nil {
			log.Printf("Error getting chapter: %v", err)
			respondError(w, http.StatusBadRequest, "Invalid chapter ID")
			return
		}
		if chapter.CampaignID != campaignID {
			respondError(w, http.StatusBadRequest, "Chapter does not belong to this campaign")
			return
		}
	}

	session, err := h.db.UpdateSession(r.Context(), sessionID, req)
	if err != nil {
		log.Printf("Error updating session: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update session")
		return
	}

	response := SessionWithAnalysis{Session: session}
	if req.PrepNotes != nil {
		content := ""
		if req.PrepNotes != nil {
			content = *req.PrepNotes
		}
		job, items, analyzeErr := h.analyzer.AnalyzeContent(
			r.Context(), session.CampaignID,
			"sessions", "prep_notes", session.ID,
			content,
		)
		if analyzeErr != nil {
			log.Printf("Content analysis failed for session %d prep_notes: %v", session.ID, analyzeErr)
		} else if len(items) > 0 {
			response.Analysis = &models.AnalysisSummary{
				JobID:        job.ID,
				PendingCount: job.TotalItems,
			}
		}
	}
	if req.ActualNotes != nil {
		content := ""
		if req.ActualNotes != nil {
			content = *req.ActualNotes
		}
		job, items, analyzeErr := h.analyzer.AnalyzeContent(
			r.Context(), session.CampaignID,
			"sessions", "actual_notes", session.ID,
			content,
		)
		if analyzeErr != nil {
			log.Printf("Content analysis failed for session %d actual_notes: %v", session.ID, analyzeErr)
		} else if response.Analysis == nil && len(items) > 0 {
			response.Analysis = &models.AnalysisSummary{
				JobID:        job.ID,
				PendingCount: job.TotalItems,
			}
		}
	}
	respondJSON(w, http.StatusOK, response)
}

// DeleteSession handles DELETE /api/campaigns/{id}/sessions/{sessionId}
// Deletes a session.
func (h *Handler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	sessionID, err := parseInt64(r, "sessionId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	// Verify the session exists and belongs to the specified campaign
	existing, err := h.db.GetSession(r.Context(), sessionID)
	if err != nil {
		log.Printf("Error getting session: %v", err)
		respondError(w, http.StatusNotFound, "Session not found")
		return
	}

	if existing.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Session not found")
		return
	}

	if err := h.db.DeleteSession(r.Context(), sessionID); err != nil {
		log.Printf("Error deleting session: %v", err)
		respondError(w, http.StatusNotFound, "Session not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListChapterEntities handles GET /api/campaigns/{id}/chapters/{chapterId}/entities
// Returns all entities linked to a chapter.
func (h *Handler) ListChapterEntities(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	userID, ok := h.verifyCampaignOwnership(w, r, campaignID)
	if !ok {
		return
	}

	chapterID, err := parseInt64(r, "chapterId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid chapter ID")
		return
	}

	// Verify the chapter belongs to the specified campaign
	chapter, err := h.db.GetChapter(r.Context(), chapterID)
	if err != nil {
		log.Printf("Error getting chapter: %v", err)
		respondError(w, http.StatusNotFound, "Chapter not found")
		return
	}

	if chapter.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Chapter not found")
		return
	}

	links, err := h.db.ListChapterEntities(r.Context(), chapterID)
	if err != nil {
		log.Printf("Error listing chapter entities: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list chapter entities")
		return
	}

	if links == nil {
		links = []models.ChapterEntity{}
	}

	// Filter GM notes from linked entities for non-owners
	isGM := h.isUserCampaignOwner(r.Context(), campaignID, userID)
	for i := range links {
		if links[i].Entity != nil {
			filterEntityGMNotes(links[i].Entity, isGM)
		}
	}

	respondJSON(w, http.StatusOK, links)
}

// CreateChapterEntity handles POST /api/campaigns/{id}/chapters/{chapterId}/entities
// Links an entity to a chapter.
func (h *Handler) CreateChapterEntity(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	chapterID, err := parseInt64(r, "chapterId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid chapter ID")
		return
	}

	// Verify the chapter belongs to the specified campaign
	chapter, err := h.db.GetChapter(r.Context(), chapterID)
	if err != nil {
		log.Printf("Error getting chapter: %v", err)
		respondError(w, http.StatusNotFound, "Chapter not found")
		return
	}

	if chapter.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Chapter not found")
		return
	}

	var req models.CreateChapterEntityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.EntityID == 0 {
		respondError(w, http.StatusBadRequest, "Entity ID is required")
		return
	}

	// Verify the entity exists and belongs to the same campaign
	entity, err := h.db.GetEntity(r.Context(), req.EntityID)
	if err != nil {
		log.Printf("Error getting entity: %v", err)
		respondError(w, http.StatusBadRequest, "Entity not found")
		return
	}

	if entity.CampaignID != campaignID {
		respondError(w, http.StatusBadRequest, "Entity does not belong to this campaign")
		return
	}

	link, err := h.db.CreateChapterEntity(r.Context(), chapterID, req)
	if err != nil {
		log.Printf("Error creating chapter entity: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to link entity to chapter")
		return
	}

	respondJSON(w, http.StatusCreated, link)
}

// UpdateChapterEntity handles PUT /api/campaigns/{id}/chapters/{chapterId}/entities/{linkId}
// Updates a chapter-entity link (e.g., change mention_type).
func (h *Handler) UpdateChapterEntity(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	chapterID, err := parseInt64(r, "chapterId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid chapter ID")
		return
	}

	// Verify the chapter belongs to the specified campaign
	chapter, err := h.db.GetChapter(r.Context(), chapterID)
	if err != nil {
		log.Printf("Error getting chapter: %v", err)
		respondError(w, http.StatusNotFound, "Chapter not found")
		return
	}

	if chapter.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Chapter not found")
		return
	}

	linkID, err := parseInt64(r, "linkId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid link ID")
		return
	}

	// Verify the link exists and belongs to the specified chapter
	existingLink, err := h.db.GetChapterEntity(r.Context(), linkID)
	if err != nil {
		log.Printf("Error getting chapter entity: %v", err)
		respondError(w, http.StatusNotFound, "Chapter entity link not found")
		return
	}

	if existingLink.ChapterID != chapterID {
		respondError(w, http.StatusNotFound, "Chapter entity link not found")
		return
	}

	var req models.UpdateChapterEntityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	link, err := h.db.UpdateChapterEntity(r.Context(), linkID, req)
	if err != nil {
		log.Printf("Error updating chapter entity: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update chapter entity link")
		return
	}

	respondJSON(w, http.StatusOK, link)
}

// DeleteChapterEntity handles DELETE /api/campaigns/{id}/chapters/{chapterId}/entities/{linkId}
// Removes an entity link from a chapter.
func (h *Handler) DeleteChapterEntity(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	chapterID, err := parseInt64(r, "chapterId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid chapter ID")
		return
	}

	// Verify the chapter belongs to the specified campaign
	chapter, err := h.db.GetChapter(r.Context(), chapterID)
	if err != nil {
		log.Printf("Error getting chapter: %v", err)
		respondError(w, http.StatusNotFound, "Chapter not found")
		return
	}

	if chapter.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Chapter not found")
		return
	}

	linkID, err := parseInt64(r, "linkId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid link ID")
		return
	}

	// Verify the link exists and belongs to the specified chapter
	existingLink, err := h.db.GetChapterEntity(r.Context(), linkID)
	if err != nil {
		log.Printf("Error getting chapter entity: %v", err)
		respondError(w, http.StatusNotFound, "Chapter entity link not found")
		return
	}

	if existingLink.ChapterID != chapterID {
		respondError(w, http.StatusNotFound, "Chapter entity link not found")
		return
	}

	if err := h.db.DeleteChapterEntity(r.Context(), linkID); err != nil {
		log.Printf("Error deleting chapter entity: %v", err)
		respondError(w, http.StatusNotFound, "Chapter entity link not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SearchCampaignContent handles GET /api/campaigns/{id}/search
// Performs hybrid vector+BM25 search across all campaign content.
func (h *Handler) SearchCampaignContent(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		respondError(w, http.StatusBadRequest, "Query parameter 'q' is required")
		return
	}

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if !h.db.IsVectorizationAvailable(r.Context()) {
		respondError(w, http.StatusServiceUnavailable,
			"Semantic search is not available. The vectorizer extension is not installed.")
		return
	}

	results, err := h.db.SearchCampaignContent(r.Context(), campaignID, query, limit)
	if err != nil {
		log.Printf("Error searching campaign content: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to search campaign content")
		return
	}

	if results == nil {
		results = []models.SearchResult{}
	}

	respondJSON(w, http.StatusOK, results)
}
