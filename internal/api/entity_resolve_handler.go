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
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/antonypegg/imagineer/internal/auth"
	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/models"
	"github.com/go-chi/chi/v5"
)

// EntityResolveHandler handles entity name resolution requests for
// wiki-link autocomplete.
type EntityResolveHandler struct {
	db *database.DB
}

// NewEntityResolveHandler creates a new EntityResolveHandler.
func NewEntityResolveHandler(db *database.DB) *EntityResolveHandler {
	return &EntityResolveHandler{db: db}
}

// ResolveEntity handles GET /api/campaigns/{id}/entities/resolve?name={text}&limit={n}
// It fuzzy-matches entity names using pg_trgm similarity for wiki-link
// autocomplete suggestions.
func (h *EntityResolveHandler) ResolveEntity(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Get campaign ID from URL
	campaignIDStr := chi.URLParam(r, "id")
	campaignID, err := strconv.ParseInt(campaignIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify user owns the campaign
	if err := h.db.VerifyCampaignOwnership(r.Context(), campaignID, userID); err != nil {
		respondError(w, http.StatusNotFound, "Campaign not found")
		return
	}

	// Validate name query parameter
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		respondError(w, http.StatusBadRequest, "Query parameter 'name' is required")
		return
	}
	if len(name) < 3 {
		respondError(w, http.StatusBadRequest, "Query parameter 'name' must be at least 3 characters")
		return
	}

	// Parse and clamp limit
	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, parseErr := strconv.Atoi(limitStr); parseErr == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > 20 {
		limit = 20
	}

	results, err := h.db.ResolveEntityByName(r.Context(), campaignID, name, limit)
	if err != nil {
		log.Printf("Error resolving entity name: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to resolve entity name")
		return
	}

	if results == nil {
		results = []models.EntityResolveResult{}
	}

	respondJSON(w, http.StatusOK, results)
}
