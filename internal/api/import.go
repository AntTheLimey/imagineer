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
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/antonypegg/imagineer/internal/auth"
	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/importers/common"
	"github.com/antonypegg/imagineer/internal/importers/evernote"
	"github.com/antonypegg/imagineer/internal/importers/googledocs"
	"github.com/antonypegg/imagineer/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ImportHandler handles import-related requests.
type ImportHandler struct {
	db                 *database.DB
	evernoteImporter   common.Importer
	googledocsImporter common.Importer
}

// NewImportHandler creates an ImportHandler configured with the provided database
// and default Evernote and Google Docs importers.
func NewImportHandler(db *database.DB) *ImportHandler {
	return &ImportHandler{
		db:                 db,
		evernoteImporter:   evernote.New(),
		googledocsImporter: googledocs.New(),
	}
}

// ImportEvernote handles POST /api/campaigns/{id}/import/evernote
// Verifies the user owns the campaign before importing.
func (h *ImportHandler) ImportEvernote(w http.ResponseWriter, r *http.Request) {
	// Get the campaign ID from URL path
	campaignIDStr := chi.URLParam(r, "id")
	if campaignIDStr == "" {
		respondError(w, http.StatusBadRequest, "Campaign ID is required")
		return
	}

	campaignID, err := uuid.Parse(campaignIDStr)
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
		respondError(w, http.StatusNotFound, "Campaign not found")
		return
	}

	// Parse the multipart form with 32MB max
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	// Get the file
	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "File is required")
		return
	}
	defer file.Close()

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to read file")
		return
	}

	// Parse options from form
	options := common.ImportOptions{
		CampaignID:           campaignID.String(),
		GameSystemCode:       r.FormValue("gameSystemCode"),
		SourceDocument:       header.Filename,
		AutoDetectEntities:   r.FormValue("autoDetectEntities") == "true",
		ExtractRelationships: r.FormValue("extractRelationships") == "true",
		ExtractEvents:        r.FormValue("extractEvents") == "true",
	}

	// Import the content
	result, err := h.evernoteImporter.Import(r.Context(), bytes.NewReader(content), options)
	if err != nil {
		log.Printf("Error importing Evernote file: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to import file: "+err.Error())
		return
	}

	// Create entities in the database
	createdEntities, importErrors := h.createImportedEntities(r.Context(), campaignID, result.Entities)
	result.Errors = append(result.Errors, importErrors...)

	response := map[string]interface{}{
		"imported":      len(createdEntities),
		"entities":      createdEntities,
		"relationships": result.Relationships,
		"events":        result.Events,
		"warnings":      result.Warnings,
		"errors":        result.Errors,
	}

	respondJSON(w, http.StatusOK, response)
}

// ImportGoogleDocs handles POST /api/campaigns/{id}/import/google-docs
// Verifies the user owns the campaign before importing.
func (h *ImportHandler) ImportGoogleDocs(w http.ResponseWriter, r *http.Request) {
	// Get the campaign ID from URL path
	campaignIDStr := chi.URLParam(r, "id")
	if campaignIDStr == "" {
		respondError(w, http.StatusBadRequest, "Campaign ID is required")
		return
	}

	campaignID, err := uuid.Parse(campaignIDStr)
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
		respondError(w, http.StatusNotFound, "Campaign not found")
		return
	}

	var req struct {
		DocumentURL          string `json:"documentUrl"`
		Content              string `json:"content"`
		GameSystemCode       string `json:"gameSystemCode"`
		SourceDocument       string `json:"sourceDocument"`
		AutoDetectEntities   bool   `json:"autoDetectEntities"`
		ExtractRelationships bool   `json:"extractRelationships"`
		ExtractEvents        bool   `json:"extractEvents"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.DocumentURL == "" && req.Content == "" {
		respondError(w, http.StatusBadRequest, "Either documentUrl or content is required")
		return
	}

	// Use URL or content as the source
	source := req.DocumentURL
	if source == "" {
		source = req.Content
	}

	sourceDoc := req.SourceDocument
	if sourceDoc == "" {
		sourceDoc = req.DocumentURL
	}

	options := common.ImportOptions{
		CampaignID:           campaignID.String(),
		GameSystemCode:       req.GameSystemCode,
		SourceDocument:       sourceDoc,
		AutoDetectEntities:   req.AutoDetectEntities,
		ExtractRelationships: req.ExtractRelationships,
		ExtractEvents:        req.ExtractEvents,
	}

	// Import the content
	result, err := h.googledocsImporter.Import(r.Context(), bytes.NewReader([]byte(source)), options)
	if err != nil {
		log.Printf("Error importing Google Docs content: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to import content: "+err.Error())
		return
	}

	// Create entities in the database
	createdEntities, importErrors := h.createImportedEntities(r.Context(), campaignID, result.Entities)
	result.Errors = append(result.Errors, importErrors...)

	response := map[string]interface{}{
		"imported":      len(createdEntities),
		"entities":      createdEntities,
		"relationships": result.Relationships,
		"events":        result.Events,
		"warnings":      result.Warnings,
		"errors":        result.Errors,
	}

	respondJSON(w, http.StatusOK, response)
}

// ImportFile handles POST /api/campaigns/{id}/import/file
// This is a general file upload endpoint for importing various file types.
// Verifies the user owns the campaign before importing.
func (h *ImportHandler) ImportFile(w http.ResponseWriter, r *http.Request) {
	// Get the campaign ID from URL path
	campaignIDStr := chi.URLParam(r, "id")
	if campaignIDStr == "" {
		respondError(w, http.StatusBadRequest, "Campaign ID is required")
		return
	}

	campaignID, err := uuid.Parse(campaignIDStr)
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
		respondError(w, http.StatusNotFound, "Campaign not found")
		return
	}

	// Parse the multipart form with 32MB max
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	// Get the file
	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "File is required")
		return
	}
	defer file.Close()

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to read file")
		return
	}

	// Parse options from form
	options := common.ImportOptions{
		CampaignID:           campaignID.String(),
		GameSystemCode:       r.FormValue("gameSystemCode"),
		SourceDocument:       header.Filename,
		AutoDetectEntities:   r.FormValue("autoDetectEntities") == "true",
		ExtractRelationships: r.FormValue("extractRelationships") == "true",
		ExtractEvents:        r.FormValue("extractEvents") == "true",
	}

	// For now, use the Evernote importer as a general text importer
	// TODO: Add file type detection and use appropriate importer
	result, err := h.evernoteImporter.Import(r.Context(), bytes.NewReader(content), options)
	if err != nil {
		log.Printf("Error importing file: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to import file: "+err.Error())
		return
	}

	// Create entities in the database
	createdEntities, importErrors := h.createImportedEntities(r.Context(), campaignID, result.Entities)
	result.Errors = append(result.Errors, importErrors...)

	response := map[string]interface{}{
		"imported":      len(createdEntities),
		"entities":      createdEntities,
		"relationships": result.Relationships,
		"events":        result.Events,
		"warnings":      result.Warnings,
		"errors":        result.Errors,
	}

	respondJSON(w, http.StatusOK, response)
}

// createImportedEntities creates entities in the database from imported data.
func (h *ImportHandler) createImportedEntities(ctx context.Context, campaignID uuid.UUID, extracted []common.ExtractedEntity) ([]models.Entity, []string) {
	var created []models.Entity
	var errors []string

	for _, e := range extracted {
		// Convert attributes to JSON
		attributes, err := json.Marshal(e.Attributes)
		if err != nil {
			errors = append(errors, "Failed to marshal attributes for "+e.Name)
			continue
		}

		req := models.CreateEntityRequest{
			EntityType:     models.EntityType(e.Type),
			Name:           e.Name,
			Description:    &e.Description,
			Attributes:     attributes,
			Tags:           e.Tags,
			SourceDocument: &e.SourceDoc,
		}

		entity, err := h.db.CreateEntity(ctx, campaignID, req)
		if err != nil {
			errors = append(errors, "Failed to create entity "+e.Name+": "+err.Error())
			continue
		}

		created = append(created, *entity)
	}

	return created, errors
}