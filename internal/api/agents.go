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
	"strconv"

	"github.com/antonypegg/imagineer/internal/agents/consistency"
	"github.com/antonypegg/imagineer/internal/auth"
	"github.com/antonypegg/imagineer/internal/database"
	"github.com/go-chi/chi/v5"
)

// AgentHandler provides HTTP handlers for agent operations.
type AgentHandler struct {
	db                 *database.DB
	consistencyChecker *consistency.Checker
}

// NewAgentHandler creates a new AgentHandler with the given database connection.
func NewAgentHandler(db *database.DB) *AgentHandler {
	return &AgentHandler{
		db:                 db,
		consistencyChecker: consistency.New(db),
	}
}

// ConsistencyCheckRequest represents the request body for a consistency check.
type ConsistencyCheckRequest struct {
	// EntityType optionally filters checks to a specific entity type.
	EntityType string `json:"entityType,omitempty"`
}

// RunConsistencyCheck handles POST /api/campaigns/{id}/agents/consistency-check
// Verifies the user owns the campaign before running the consistency check.
func (h *AgentHandler) RunConsistencyCheck(w http.ResponseWriter, r *http.Request) {
	campaignIDStr := chi.URLParam(r, "id")
	campaignID, err := strconv.ParseInt(campaignIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	if err := h.db.VerifyCampaignOwnership(r.Context(), campaignID, userID); err != nil {
		log.Printf("Error verifying campaign ownership: %v", err)
		respondError(w, http.StatusNotFound, "Campaign not found")
		return
	}

	// Parse optional request body
	var req ConsistencyCheckRequest
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
	}

	// Build agent parameters
	params := map[string]any{
		"campaign_id": campaignID,
	}
	if req.EntityType != "" {
		params["entity_type"] = req.EntityType
	}

	// Run the consistency checker
	result, err := h.consistencyChecker.Run(r.Context(), params)
	if err != nil {
		log.Printf("Error running consistency check: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to run consistency check")
		return
	}

	respondJSON(w, http.StatusOK, result)
}
