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
	"errors"
	"log"
	"net/http"

	"github.com/antonypegg/imagineer/internal/auth"
	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/models"
	"github.com/jackc/pgx/v5"
)

// SceneHandler handles scene CRUD API requests.
type SceneHandler struct {
	db *database.DB
}

// NewSceneHandler creates a new SceneHandler.
func NewSceneHandler(db *database.DB) *SceneHandler {
	return &SceneHandler{db: db}
}

// verifySessionBelongsToCampaign checks that the session exists and
// belongs to the specified campaign. Returns false and writes an error
// response if the check fails.
func (h *SceneHandler) verifySessionBelongsToCampaign(
	w http.ResponseWriter,
	r *http.Request,
	sessionID, campaignID int64,
) bool {
	session, err := h.db.GetSession(r.Context(), sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(w, http.StatusNotFound, "Session not found")
			return false
		}
		log.Printf("Error verifying session ownership: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to verify session")
		return false
	}
	if session.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Session not found")
		return false
	}
	return true
}

// ListScenes handles GET /api/campaigns/{id}/sessions/{sessionId}/scenes
// Returns all scenes for a session.
func (h *SceneHandler) ListScenes(w http.ResponseWriter, r *http.Request) {
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

	sessionID, err := parseInt64(r, "sessionId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	if !h.verifySessionBelongsToCampaign(w, r, sessionID, campaignID) {
		return
	}

	scenes, err := h.db.ListScenesBySession(r.Context(), sessionID)
	if err != nil {
		log.Printf("Error listing scenes: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list scenes")
		return
	}

	if scenes == nil {
		scenes = []models.Scene{}
	}

	respondJSON(w, http.StatusOK, scenes)
}

// GetScene handles GET /api/campaigns/{id}/sessions/{sessionId}/scenes/{sceneId}
// Returns a single scene.
func (h *SceneHandler) GetScene(w http.ResponseWriter, r *http.Request) {
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

	sessionID, err := parseInt64(r, "sessionId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	if !h.verifySessionBelongsToCampaign(w, r, sessionID, campaignID) {
		return
	}

	sceneID, err := parseInt64(r, "sceneId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid scene ID")
		return
	}

	scene, err := h.db.GetScene(r.Context(), sceneID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(w, http.StatusNotFound, "Scene not found")
			return
		}
		log.Printf("Error getting scene: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get scene")
		return
	}

	if scene.SessionID != sessionID {
		respondError(w, http.StatusNotFound, "Scene not found")
		return
	}

	respondJSON(w, http.StatusOK, scene)
}

// CreateScene handles POST /api/campaigns/{id}/sessions/{sessionId}/scenes
// Creates a new scene in a session.
func (h *SceneHandler) CreateScene(w http.ResponseWriter, r *http.Request) {
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

	sessionID, err := parseInt64(r, "sessionId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	if !h.verifySessionBelongsToCampaign(w, r, sessionID, campaignID) {
		return
	}

	var req models.CreateSceneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Title == "" {
		respondError(w, http.StatusBadRequest, "Title is required")
		return
	}

	scene, err := h.db.CreateScene(r.Context(), sessionID, campaignID, req)
	if err != nil {
		log.Printf("Error creating scene: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create scene")
		return
	}

	respondJSON(w, http.StatusCreated, scene)
}

// UpdateScene handles PUT /api/campaigns/{id}/sessions/{sessionId}/scenes/{sceneId}
// Updates an existing scene.
func (h *SceneHandler) UpdateScene(w http.ResponseWriter, r *http.Request) {
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

	sessionID, err := parseInt64(r, "sessionId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	if !h.verifySessionBelongsToCampaign(w, r, sessionID, campaignID) {
		return
	}

	sceneID, err := parseInt64(r, "sceneId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid scene ID")
		return
	}

	// Verify the scene exists and belongs to this session
	existingScene, err := h.db.GetScene(r.Context(), sceneID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(w, http.StatusNotFound, "Scene not found")
			return
		}
		log.Printf("Error getting scene for update: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get scene")
		return
	}

	if existingScene.SessionID != sessionID {
		respondError(w, http.StatusNotFound, "Scene not found")
		return
	}

	var req models.UpdateSceneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	scene, err := h.db.UpdateScene(r.Context(), sceneID, req)
	if err != nil {
		log.Printf("Error updating scene: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update scene")
		return
	}

	respondJSON(w, http.StatusOK, scene)
}

// DeleteScene handles DELETE /api/campaigns/{id}/sessions/{sessionId}/scenes/{sceneId}
// Deletes a scene.
func (h *SceneHandler) DeleteScene(w http.ResponseWriter, r *http.Request) {
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

	sessionID, err := parseInt64(r, "sessionId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	if !h.verifySessionBelongsToCampaign(w, r, sessionID, campaignID) {
		return
	}

	sceneID, err := parseInt64(r, "sceneId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid scene ID")
		return
	}

	// Verify the scene exists and belongs to this session
	existingScene, err := h.db.GetScene(r.Context(), sceneID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(w, http.StatusNotFound, "Scene not found")
			return
		}
		log.Printf("Error getting scene for delete: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get scene")
		return
	}

	if existingScene.SessionID != sessionID {
		respondError(w, http.StatusNotFound, "Scene not found")
		return
	}

	if err := h.db.DeleteScene(r.Context(), sceneID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(w, http.StatusNotFound, "Scene not found")
			return
		}
		log.Printf("Error deleting scene: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to delete scene")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
