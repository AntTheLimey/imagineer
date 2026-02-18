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
	"strconv"

	"github.com/antonypegg/imagineer/internal/auth"
	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

// DraftHandler handles draft auto-save API requests.
type DraftHandler struct {
	db *database.DB
}

// NewDraftHandler creates a new DraftHandler.
func NewDraftHandler(db *database.DB) *DraftHandler {
	return &DraftHandler{db: db}
}

// SaveDraft handles PUT /api/campaigns/{id}/drafts
// Saves or updates a draft for the authenticated user.
func (h *DraftHandler) SaveDraft(w http.ResponseWriter, r *http.Request) {
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

	var req models.SaveDraftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.SourceTable == "" {
		respondError(w, http.StatusBadRequest, "Source table is required")
		return
	}

	if len(req.DraftData) == 0 || string(req.DraftData) == "null" {
		respondError(w, http.StatusBadRequest, "Draft data is required")
		return
	}

	draft, err := h.db.SaveDraft(r.Context(), campaignID, userID, req)
	if err != nil {
		log.Printf("Error saving draft: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to save draft")
		return
	}

	respondJSON(w, http.StatusOK, draft)
}

// ListDraftIndicators handles GET /api/campaigns/{id}/drafts
// Returns lightweight draft indicators for a campaign and user.
func (h *DraftHandler) ListDraftIndicators(w http.ResponseWriter, r *http.Request) {
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

	sourceTable := r.URL.Query().Get("source_table")

	indicators, err := h.db.ListDraftIndicators(r.Context(), campaignID, userID, sourceTable)
	if err != nil {
		log.Printf("Error listing draft indicators: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list draft indicators")
		return
	}

	if indicators == nil {
		indicators = []models.DraftIndicator{}
	}

	respondJSON(w, http.StatusOK, indicators)
}

// GetDraft handles GET /api/campaigns/{id}/drafts/{st}/{sid}
// Returns a specific draft by source table and source ID.
func (h *DraftHandler) GetDraft(w http.ResponseWriter, r *http.Request) {
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

	sourceTable := chi.URLParam(r, "st")
	sourceID, err := strconv.ParseInt(chi.URLParam(r, "sid"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid source ID")
		return
	}

	draft, err := h.db.GetDraft(r.Context(), userID, sourceTable, sourceID, campaignID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(w, http.StatusNotFound, "Draft not found")
			return
		}
		log.Printf("Error getting draft: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get draft")
		return
	}

	respondJSON(w, http.StatusOK, draft)
}

// DeleteDraft handles DELETE /api/campaigns/{id}/drafts/{st}/{sid}
// Deletes a specific draft by source table and source ID.
func (h *DraftHandler) DeleteDraft(w http.ResponseWriter, r *http.Request) {
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

	sourceTable := chi.URLParam(r, "st")
	sourceID, err := strconv.ParseInt(chi.URLParam(r, "sid"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid source ID")
		return
	}

	if err := h.db.DeleteDraft(r.Context(), userID, sourceTable, sourceID, campaignID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(w, http.StatusNotFound, "Draft not found")
			return
		}
		log.Printf("Error deleting draft: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to delete draft")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// BeaconSave handles POST /api/campaigns/{id}/drafts/beacon
// Saves a draft using navigator.sendBeacon() semantics. Accepts
// JSON body (sendBeacon sends as a Blob) and responds with an
// empty JSON object.
func (h *DraftHandler) BeaconSave(w http.ResponseWriter, r *http.Request) {
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

	var req models.SaveDraftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.SourceTable == "" {
		respondError(w, http.StatusBadRequest, "Source table is required")
		return
	}

	if len(req.DraftData) == 0 || string(req.DraftData) == "null" {
		respondError(w, http.StatusBadRequest, "Draft data is required")
		return
	}

	if _, err := h.db.SaveDraft(r.Context(), campaignID, userID, req); err != nil {
		log.Printf("Error saving draft via beacon: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to save draft")
		return
	}

	respondJSON(w, http.StatusOK, struct{}{})
}
