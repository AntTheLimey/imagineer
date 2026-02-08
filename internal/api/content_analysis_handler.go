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
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/antonypegg/imagineer/internal/analysis"
	"github.com/antonypegg/imagineer/internal/auth"
	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/models"
)

// ContentAnalysisHandler handles content analysis API requests.
type ContentAnalysisHandler struct {
	db       *database.DB
	analyzer *analysis.Analyzer
}

// NewContentAnalysisHandler creates a new ContentAnalysisHandler.
func NewContentAnalysisHandler(db *database.DB) *ContentAnalysisHandler {
	return &ContentAnalysisHandler{
		db:       db,
		analyzer: analysis.NewAnalyzer(db),
	}
}

// TriggerAnalysisRequest is the request body for triggering content
// analysis on a specific content field.
type TriggerAnalysisRequest struct {
	SourceTable string `json:"sourceTable"`
	SourceID    int64  `json:"sourceId"`
	SourceField string `json:"sourceField"`
}

// TriggerAnalysisResponse is the response body for a triggered analysis.
type TriggerAnalysisResponse struct {
	Job   *models.ContentAnalysisJob   `json:"job"`
	Items []models.ContentAnalysisItem `json:"items"`
}

// PendingCountResponse is the response body for the pending count
// endpoint.
type PendingCountResponse struct {
	Count int `json:"count"`
}

// ListJobs handles GET /api/campaigns/{id}/analysis/jobs
// Returns all content analysis jobs for a campaign.
func (h *ContentAnalysisHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
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

	jobs, err := h.db.ListAnalysisJobsByCampaign(r.Context(), campaignID)
	if err != nil {
		log.Printf("Error listing analysis jobs: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list analysis jobs")
		return
	}

	if jobs == nil {
		jobs = []models.ContentAnalysisJob{}
	}

	respondJSON(w, http.StatusOK, jobs)
}

// GetJob handles GET /api/campaigns/{id}/analysis/jobs/{jobId}
// Returns a specific content analysis job.
func (h *ContentAnalysisHandler) GetJob(w http.ResponseWriter, r *http.Request) {
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

	jobID, err := parseInt64(r, "jobId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid job ID")
		return
	}

	job, err := h.db.GetAnalysisJob(r.Context(), jobID)
	if err != nil {
		log.Printf("Error getting analysis job: %v", err)
		respondError(w, http.StatusNotFound, "Analysis job not found")
		return
	}

	// Verify the job belongs to the specified campaign
	if job.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Analysis job not found")
		return
	}

	respondJSON(w, http.StatusOK, job)
}

// ListJobItems handles GET /api/campaigns/{id}/analysis/jobs/{jobId}/items
// Returns all items for a content analysis job. Supports optional
// ?resolution= query parameter to filter by resolution status.
func (h *ContentAnalysisHandler) ListJobItems(w http.ResponseWriter, r *http.Request) {
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

	jobID, err := parseInt64(r, "jobId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid job ID")
		return
	}

	// Verify the job exists and belongs to this campaign
	job, err := h.db.GetAnalysisJob(r.Context(), jobID)
	if err != nil {
		log.Printf("Error getting analysis job: %v", err)
		respondError(w, http.StatusNotFound, "Analysis job not found")
		return
	}

	if job.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Analysis job not found")
		return
	}

	resolution := r.URL.Query().Get("resolution")

	items, err := h.db.ListAnalysisItemsByJob(r.Context(), jobID, resolution)
	if err != nil {
		log.Printf("Error listing analysis items: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list analysis items")
		return
	}

	if items == nil {
		items = []models.ContentAnalysisItem{}
	}

	respondJSON(w, http.StatusOK, items)
}

// ResolveItem handles PUT /api/campaigns/{id}/analysis/items/{itemId}
// Resolves a content analysis item with the given resolution. Supported
// resolutions are "accepted", "new_entity", and "dismissed".
func (h *ContentAnalysisHandler) ResolveItem(w http.ResponseWriter, r *http.Request) {
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

	itemID, err := parseInt64(r, "itemId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid item ID")
		return
	}

	var req models.ResolveAnalysisItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate the resolution value
	switch req.Resolution {
	case "accepted", "new_entity", "dismissed":
		// valid
	default:
		respondError(w, http.StatusBadRequest,
			"Resolution must be one of: accepted, new_entity, dismissed")
		return
	}

	// Determine the resolved entity ID based on resolution type
	var resolvedEntityID *int64

	switch req.Resolution {
	case "new_entity":
		// Require entity type and name for creating a new entity
		if req.EntityType == nil || *req.EntityType == "" {
			respondError(w, http.StatusBadRequest,
				"Entity type is required for new_entity resolution")
			return
		}
		if req.EntityName == nil || *req.EntityName == "" {
			respondError(w, http.StatusBadRequest,
				"Entity name is required for new_entity resolution")
			return
		}

		createReq := models.CreateEntityRequest{
			EntityType: *req.EntityType,
			Name:       *req.EntityName,
		}

		entity, err := h.db.CreateEntity(r.Context(), campaignID, createReq)
		if err != nil {
			log.Printf("Error creating entity from analysis item: %v", err)
			respondError(w, http.StatusInternalServerError,
				"Failed to create entity")
			return
		}

		resolvedEntityID = &entity.ID

	case "accepted":
		// Use the existing entity_id from the analysis item.
		var entityID *int64
		err := h.db.QueryRow(r.Context(),
			"SELECT entity_id FROM content_analysis_items WHERE id = $1",
			itemID,
		).Scan(&entityID)
		if err != nil {
			log.Printf("Error fetching item entity_id: %v", err)
			respondError(w, http.StatusNotFound, "Analysis item not found")
			return
		}
		resolvedEntityID = entityID

	case "dismissed":
		resolvedEntityID = nil
	}

	// Resolve the analysis item
	if err := h.db.ResolveAnalysisItem(r.Context(), itemID, req.Resolution, resolvedEntityID); err != nil {
		log.Printf("Error resolving analysis item: %v", err)
		respondError(w, http.StatusInternalServerError,
			"Failed to resolve analysis item")
		return
	}

	// Update the job's resolved count
	jobID, err := h.getItemJobID(r.Context(), itemID)
	if err != nil {
		log.Printf("Error getting job ID for item %d: %v", itemID, err)
		// The item was resolved successfully; the count update is
		// non-critical so we still return success.
	} else {
		if err := h.db.UpdateJobResolvedCount(r.Context(), jobID); err != nil {
			log.Printf("Error updating job resolved count: %v", err)
		}
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status": "resolved",
	})
}

// TriggerAnalysis handles POST /api/campaigns/{id}/analysis/trigger
// Triggers content analysis on a specific content field.
func (h *ContentAnalysisHandler) TriggerAnalysis(w http.ResponseWriter, r *http.Request) {
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

	var req TriggerAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.SourceTable == "" || req.SourceField == "" {
		respondError(w, http.StatusBadRequest,
			"sourceTable and sourceField are required")
		return
	}

	if req.SourceID <= 0 {
		respondError(w, http.StatusBadRequest, "sourceId must be a positive integer")
		return
	}

	// Fetch the content from the appropriate table and field
	content, err := h.fetchSourceContent(r.Context(), req.SourceTable, req.SourceID, req.SourceField)
	if err != nil {
		log.Printf("Error fetching source content: %v", err)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	job, items, err := h.analyzer.AnalyzeContent(
		r.Context(), campaignID,
		req.SourceTable, req.SourceField, req.SourceID,
		content,
	)
	if err != nil {
		log.Printf("Error analyzing content: %v", err)
		respondError(w, http.StatusInternalServerError,
			"Failed to analyze content")
		return
	}

	if items == nil {
		items = []models.ContentAnalysisItem{}
	}

	respondJSON(w, http.StatusOK, TriggerAnalysisResponse{
		Job:   job,
		Items: items,
	})
}

// GetPendingCount handles GET /api/campaigns/{id}/analysis/pending-count
// Returns the number of pending analysis items for a specific source.
// Query parameters: sourceTable, sourceId.
func (h *ContentAnalysisHandler) GetPendingCount(w http.ResponseWriter, r *http.Request) {
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

	sourceTable := r.URL.Query().Get("sourceTable")
	if sourceTable == "" {
		respondError(w, http.StatusBadRequest,
			"Query parameter 'sourceTable' is required")
		return
	}

	sourceIDStr := r.URL.Query().Get("sourceId")
	if sourceIDStr == "" {
		respondError(w, http.StatusBadRequest,
			"Query parameter 'sourceId' is required")
		return
	}

	sourceID, err := strconv.ParseInt(sourceIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid sourceId")
		return
	}

	count, err := h.db.CountPendingAnalysisItems(r.Context(), campaignID, sourceTable, sourceID)
	if err != nil {
		log.Printf("Error counting pending analysis items: %v", err)
		respondError(w, http.StatusInternalServerError,
			"Failed to count pending analysis items")
		return
	}

	respondJSON(w, http.StatusOK, PendingCountResponse{Count: count})
}

// getItemJobID retrieves the job_id for a content analysis item.
func (h *ContentAnalysisHandler) getItemJobID(ctx context.Context, itemID int64) (int64, error) {
	var jobID int64
	err := h.db.QueryRow(ctx,
		"SELECT job_id FROM content_analysis_items WHERE id = $1",
		itemID,
	).Scan(&jobID)
	return jobID, err
}

// fetchSourceContent retrieves the text content from the appropriate
// source table and field using parameterized queries.
func (h *ContentAnalysisHandler) fetchSourceContent(
	ctx context.Context,
	sourceTable string,
	sourceID int64,
	sourceField string,
) (string, error) {
	var query string

	switch sourceTable + "." + sourceField {
	case "entities.description":
		query = "SELECT COALESCE(description, '') FROM entities WHERE id = $1"
	case "entities.gm_notes":
		query = "SELECT COALESCE(gm_notes, '') FROM entities WHERE id = $1"
	case "chapters.overview":
		query = "SELECT COALESCE(overview, '') FROM chapters WHERE id = $1"
	case "sessions.prep_notes":
		query = "SELECT COALESCE(prep_notes, '') FROM sessions WHERE id = $1"
	case "sessions.actual_notes":
		query = "SELECT COALESCE(actual_notes, '') FROM sessions WHERE id = $1"
	case "campaigns.description":
		query = "SELECT COALESCE(description, '') FROM campaigns WHERE id = $1"
	default:
		return "", fmt.Errorf("unsupported source: %s.%s", sourceTable, sourceField)
	}

	var content string
	err := h.db.QueryRow(ctx, query, sourceID).Scan(&content)
	if err != nil {
		return "", fmt.Errorf("failed to fetch content from %s.%s: %w",
			sourceTable, sourceField, err)
	}

	return content, nil
}
