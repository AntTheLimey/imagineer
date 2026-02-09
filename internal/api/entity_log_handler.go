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
	"encoding/json"
	"log"
	"net/http"

	"github.com/antonypegg/imagineer/internal/auth"
	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/models"
)

// EntityLogHandler handles entity log CRUD API requests.
type EntityLogHandler struct {
	db *database.DB
}

// NewEntityLogHandler creates a new EntityLogHandler.
func NewEntityLogHandler(db *database.DB) *EntityLogHandler {
	return &EntityLogHandler{db: db}
}

// verifyEntityBelongsToCampaign checks that the entity exists and belongs
// to the specified campaign. Returns false and writes an error response
// if the check fails.
func (h *EntityLogHandler) verifyEntityBelongsToCampaign(
	w http.ResponseWriter,
	r *http.Request,
	entityID, campaignID int64,
) bool {
	var entityCampaignID int64
	err := h.db.QueryRow(r.Context(),
		"SELECT campaign_id FROM entities WHERE id = $1",
		entityID,
	).Scan(&entityCampaignID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Entity not found")
		return false
	}

	if entityCampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Entity not found")
		return false
	}

	return true
}

// ListEntityLogs handles GET /api/campaigns/{id}/entities/{entityId}/log
// Returns all log entries for an entity.
func (h *EntityLogHandler) ListEntityLogs(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	if err := h.db.VerifyCampaignOwnership(r.Context(), campaignID, userID); err != nil {
		respondError(w, http.StatusNotFound, "Campaign not found")
		return
	}

	entityID, err := parseInt64(r, "entityId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid entity ID")
		return
	}

	if !h.verifyEntityBelongsToCampaign(w, r, entityID, campaignID) {
		return
	}

	logs, err := h.db.ListEntityLogs(r.Context(), entityID)
	if err != nil {
		log.Printf("Error listing entity logs: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list entity logs")
		return
	}

	if logs == nil {
		logs = []models.EntityLog{}
	}

	respondJSON(w, http.StatusOK, logs)
}

// CreateEntityLog handles POST /api/campaigns/{id}/entities/{entityId}/log
// Creates a new log entry for an entity.
func (h *EntityLogHandler) CreateEntityLog(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	if err := h.db.VerifyCampaignOwnership(r.Context(), campaignID, userID); err != nil {
		respondError(w, http.StatusNotFound, "Campaign not found")
		return
	}

	entityID, err := parseInt64(r, "entityId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid entity ID")
		return
	}

	if !h.verifyEntityBelongsToCampaign(w, r, entityID, campaignID) {
		return
	}

	var req models.CreateEntityLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Content == "" {
		respondError(w, http.StatusBadRequest, "Content is required")
		return
	}

	logEntry, err := h.db.CreateEntityLog(r.Context(), entityID, campaignID, req)
	if err != nil {
		log.Printf("Error creating entity log: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create entity log")
		return
	}

	respondJSON(w, http.StatusCreated, logEntry)
}

// UpdateEntityLog handles PUT /api/campaigns/{id}/entities/{entityId}/log/{logId}
// Updates an existing entity log entry.
func (h *EntityLogHandler) UpdateEntityLog(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	if err := h.db.VerifyCampaignOwnership(r.Context(), campaignID, userID); err != nil {
		respondError(w, http.StatusNotFound, "Campaign not found")
		return
	}

	entityID, err := parseInt64(r, "entityId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid entity ID")
		return
	}

	if !h.verifyEntityBelongsToCampaign(w, r, entityID, campaignID) {
		return
	}

	logID, err := parseInt64(r, "logId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid log ID")
		return
	}

	// Verify the log entry exists and belongs to this entity
	existingLog, err := h.db.GetEntityLog(r.Context(), logID)
	if err != nil {
		log.Printf("Error getting entity log: %v", err)
		respondError(w, http.StatusNotFound, "Entity log not found")
		return
	}

	if existingLog.EntityID != entityID {
		respondError(w, http.StatusNotFound, "Entity log not found")
		return
	}

	var req models.UpdateEntityLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	logEntry, err := h.db.UpdateEntityLog(r.Context(), logID, req)
	if err != nil {
		log.Printf("Error updating entity log: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update entity log")
		return
	}

	respondJSON(w, http.StatusOK, logEntry)
}

// DeleteEntityLog handles DELETE /api/campaigns/{id}/entities/{entityId}/log/{logId}
// Deletes an entity log entry.
func (h *EntityLogHandler) DeleteEntityLog(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	if err := h.db.VerifyCampaignOwnership(r.Context(), campaignID, userID); err != nil {
		respondError(w, http.StatusNotFound, "Campaign not found")
		return
	}

	entityID, err := parseInt64(r, "entityId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid entity ID")
		return
	}

	if !h.verifyEntityBelongsToCampaign(w, r, entityID, campaignID) {
		return
	}

	logID, err := parseInt64(r, "logId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid log ID")
		return
	}

	// Verify the log entry exists and belongs to this entity
	existingLog, err := h.db.GetEntityLog(r.Context(), logID)
	if err != nil {
		log.Printf("Error getting entity log: %v", err)
		respondError(w, http.StatusNotFound, "Entity log not found")
		return
	}

	if existingLog.EntityID != entityID {
		respondError(w, http.StatusNotFound, "Entity log not found")
		return
	}

	if err := h.db.DeleteEntityLog(r.Context(), logID); err != nil {
		log.Printf("Error deleting entity log: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to delete entity log")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
