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

	"github.com/antonypegg/imagineer/internal/models"
)

// ListEras handles GET /api/campaigns/{id}/eras
// Returns all eras for the campaign ordered by sequence.
func (h *Handler) ListEras(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	eras, err := h.db.ListEras(r.Context(), campaignID)
	if err != nil {
		log.Printf("Error listing eras: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list eras")
		return
	}

	if eras == nil {
		eras = []models.Era{}
	}

	respondJSON(w, http.StatusOK, eras)
}

// CreateEra handles POST /api/campaigns/{id}/eras
// Creates a new era for the campaign.
func (h *Handler) CreateEra(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	var req models.CreateEraRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Name is required")
		return
	}

	if req.Scale == "" {
		respondError(w, http.StatusBadRequest, "Scale is required")
		return
	}

	switch req.Scale {
	case "mythic", "ancient", "distant", "past", "recent", "now":
		// valid
	default:
		respondError(w, http.StatusBadRequest, "Invalid scale value")
		return
	}

	era, err := h.db.CreateEra(r.Context(), campaignID, req)
	if err != nil {
		log.Printf("Error creating era: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create era")
		return
	}

	respondJSON(w, http.StatusCreated, era)
}

// UpdateEra handles PUT /api/campaigns/{id}/eras/{eraId}
// Updates an existing era.
func (h *Handler) UpdateEra(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	eraID, err := parseInt64(r, "eraId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid era ID")
		return
	}

	var req models.UpdateEraRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	era, err := h.db.UpdateEra(r.Context(), eraID, campaignID, req)
	if err != nil {
		log.Printf("Error updating era: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update era")
		return
	}

	respondJSON(w, http.StatusOK, era)
}

// GetCurrentEra handles GET /api/campaigns/{id}/eras/current
// Returns the era with the highest sequence number for the campaign.
func (h *Handler) GetCurrentEra(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	era, err := h.db.GetCurrentEra(r.Context(), campaignID)
	if err != nil {
		log.Printf("Error getting current era: %v", err)
		respondError(w, http.StatusNotFound, "No eras found for campaign")
		return
	}

	respondJSON(w, http.StatusOK, era)
}

// ListConstraintOverrides handles GET /api/campaigns/{id}/constraint-overrides
// Returns all constraint overrides for the campaign.
func (h *Handler) ListConstraintOverrides(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	overrides, err := h.db.ListConstraintOverrides(r.Context(), campaignID)
	if err != nil {
		log.Printf("Error listing constraint overrides: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list constraint overrides")
		return
	}

	if overrides == nil {
		overrides = []models.ConstraintOverride{}
	}

	respondJSON(w, http.StatusOK, overrides)
}

// CreateConstraintOverride handles POST /api/campaigns/{id}/constraint-overrides
// Records a GM acknowledgement that overrides a specific constraint violation.
func (h *Handler) CreateConstraintOverride(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	var req models.CreateConstraintOverrideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ConstraintType == "" {
		respondError(w, http.StatusBadRequest, "Constraint type is required")
		return
	}

	switch req.ConstraintType {
	case "domain_range", "cardinality", "required":
		// valid
	default:
		respondError(w, http.StatusBadRequest, "Invalid constraint type")
		return
	}

	if req.OverrideKey == "" {
		respondError(w, http.StatusBadRequest, "Override key is required")
		return
	}

	override, err := h.db.CreateConstraintOverride(r.Context(), campaignID, req)
	if err != nil {
		log.Printf("Error creating constraint override: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create constraint override")
		return
	}

	respondJSON(w, http.StatusCreated, override)
}

// DeleteEra handles DELETE /api/campaigns/{id}/eras/{eraId}
// Deletes an era if it is not referenced by any relationships.
func (h *Handler) DeleteEra(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	eraID, err := parseInt64(r, "eraId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid era ID")
		return
	}

	if err := h.db.DeleteEra(r.Context(), eraID, campaignID); err != nil {
		log.Printf("Error deleting era: %v", err)
		respondError(w, http.StatusConflict, "Failed to delete era: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteConstraintOverride handles DELETE /api/campaigns/{id}/constraint-overrides/{overrideId}
// Removes a constraint override so the constraint is enforced again.
func (h *Handler) DeleteConstraintOverride(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	overrideID, err := parseInt64(r, "overrideId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid constraint override ID")
		return
	}

	if err := h.db.DeleteConstraintOverride(r.Context(), overrideID, campaignID); err != nil {
		log.Printf("Error deleting constraint override: %v", err)
		respondError(w, http.StatusNotFound, "Constraint override not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListEntityTypes handles GET /api/campaigns/{id}/entity-types
// Returns all entity types for the campaign from the ontology hierarchy.
func (h *Handler) ListEntityTypes(w http.ResponseWriter, r *http.Request) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid campaign ID")
		return
	}

	// Verify the user owns this campaign
	if _, ok := h.verifyCampaignOwnership(w, r, campaignID); !ok {
		return
	}

	types, err := h.db.ListCampaignEntityTypes(r.Context(), campaignID)
	if err != nil {
		log.Printf("Error listing entity types: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list entity types")
		return
	}

	if types == nil {
		types = []models.CampaignEntityType{}
	}

	respondJSON(w, http.StatusOK, types)
}
