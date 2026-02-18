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
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/antonypegg/imagineer/internal/agents/canon"
	"github.com/antonypegg/imagineer/internal/agents/graph"
	"github.com/antonypegg/imagineer/internal/agents/ttrpg"
	"github.com/antonypegg/imagineer/internal/analysis"
	"github.com/antonypegg/imagineer/internal/auth"
	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/enrichment"
	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/antonypegg/imagineer/internal/models"
)

// ContentAnalysisHandler handles content analysis API requests.
type ContentAnalysisHandler struct {
	db            *database.DB
	analyzer      *analysis.Analyzer
	enrichCancels sync.Map // map[int64]context.CancelFunc
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

// BatchResolveRequest is the request body for batch-resolving analysis
// items of a given detection type within a single job.
type BatchResolveRequest struct {
	DetectionType string `json:"detectionType"`
	Resolution    string `json:"resolution"`
}

// BatchResolveResponse is the response body for batch resolve, reporting
// the number of items successfully resolved.
type BatchResolveResponse struct {
	Resolved int `json:"resolved"`
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
	phase := r.URL.Query().Get("phase")

	items, err := h.db.ListAnalysisItemsByJob(r.Context(), jobID, resolution, phase)
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

	// Verify the item belongs to a job within the specified campaign.
	var itemCampaignID int64
	err = h.db.QueryRow(r.Context(),
		`SELECT j.campaign_id FROM content_analysis_items i
		 JOIN content_analysis_jobs j ON i.job_id = j.id
		 WHERE i.id = $1`,
		itemID,
	).Scan(&itemCampaignID)
	if errors.Is(err, pgx.ErrNoRows) {
		respondError(w, http.StatusNotFound, "Analysis item not found")
		return
	}
	if err != nil {
		log.Printf("Error verifying item campaign ownership: %v", err)
		respondError(w, http.StatusInternalServerError,
			"Failed to verify analysis item")
		return
	}
	if itemCampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Analysis item not found")
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
		log.Printf("Created entity %d (%q) from analysis item %d in campaign %d",
			entity.ID, *req.EntityName, itemID, campaignID)

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

	// Apply content fix for accepted/new_entity resolutions that have
	// position offsets. This wraps the matched text in [[wiki link]]
	// brackets within the source content.
	var fixJobID int64
	var fixJobIDFound bool
	var detectionType string
	var suggestedContent json.RawMessage
	if req.Resolution == "accepted" || req.Resolution == "new_entity" {
		var posStart, posEnd *int
		var matchedText, srcTable, srcField string
		var srcID int64
		err := h.db.QueryRow(r.Context(),
			`SELECT i.position_start, i.position_end, i.matched_text,
			        i.detection_type, i.job_id, i.suggested_content,
			        j.source_table, j.source_id, j.source_field
			 FROM content_analysis_items i
			 JOIN content_analysis_jobs j ON i.job_id = j.id
			 WHERE i.id = $1`,
			itemID,
		).Scan(&posStart, &posEnd, &matchedText, &detectionType,
			&fixJobID, &suggestedContent, &srcTable, &srcID, &srcField)
		if err != nil {
			log.Printf("Error fetching item details for content fix (item %d): %v",
				itemID, err)
		} else {
			fixJobIDFound = true
			if posStart != nil && posEnd != nil && !strings.HasPrefix(detectionType, "wiki_link_") {
				// Determine the replacement text based on resolution type.
				var replacement string
				if req.Resolution == "new_entity" && req.EntityName != nil && *req.EntityName != "" {
					replacement = "[[" + *req.EntityName + "]]"
				} else if detectionType == "potential_alias" && resolvedEntityID != nil {
					// For potential aliases, use wiki alias syntax
					// to preserve the original display text.
					var entityName string
					err = h.db.QueryRow(r.Context(),
						"SELECT name FROM entities WHERE id = $1",
						*resolvedEntityID,
					).Scan(&entityName)
					if err != nil {
						log.Printf("Error fetching entity name for alias link (item %d): %v",
							itemID, err)
						replacement = "[[" + matchedText + "]]"
					} else {
						replacement = "[[" + entityName + "|" + matchedText + "]]"
					}
				} else {
					replacement = "[[" + matchedText + "]]"
				}

				if fixErr := h.applyContentFix(
					r.Context(), srcTable, srcID, srcField,
					*posStart, *posEnd, matchedText, replacement,
				); fixErr != nil {
					// Content fix failure is non-fatal: log but still
					// return success since the resolution itself succeeded.
					log.Printf("Content fix failed for item %d: %v",
						itemID, fixErr)
				} else {
					// Adjust byte offsets of remaining pending items
					// in the same job so that subsequent fixes apply
					// at the correct positions.
					delta := len(replacement) - (*posEnd - *posStart)
					if delta != 0 {
						adjErr := h.db.Exec(r.Context(),
							`UPDATE content_analysis_items
							 SET position_start = position_start + $1,
							     position_end = position_end + $1
							 WHERE job_id = $2
							   AND id != $3
							   AND resolution = 'pending'
							   AND position_start >= $4`,
							delta, fixJobID, itemID, *posEnd,
						)
						if adjErr != nil {
							log.Printf(
								"Failed to adjust offsets for job %d after item %d: %v",
								fixJobID, itemID, adjErr)
						}
					}
				}
			}
		}
	}

	// Handle relationship_suggestion acceptance: create the actual
	// relationship and auto-resolve any pending inverse suggestion.
	if req.Resolution == "accepted" && detectionType == "relationship_suggestion" && len(suggestedContent) > 0 {
		h.handleRelationshipSuggestion(r.Context(), campaignID, fixJobID, itemID, suggestedContent, req.SuggestedContentOverride)
	}

	// Handle description_update acceptance: apply the suggested
	// description to the resolved entity.
	if req.Resolution == "accepted" && detectionType == "description_update" && len(suggestedContent) > 0 && resolvedEntityID != nil {
		h.handleDescriptionUpdate(r.Context(), *resolvedEntityID, suggestedContent)
	}

	// Handle log_entry acceptance: create a new entity log entry
	// from the suggested content.
	if req.Resolution == "accepted" && detectionType == "log_entry" && len(suggestedContent) > 0 && resolvedEntityID != nil {
		h.handleLogEntry(r.Context(), campaignID, *resolvedEntityID, suggestedContent)
	}

	// Update the job's resolved count. Reuse the job ID fetched above
	// when available; otherwise fall back to a separate lookup.
	jobID := fixJobID
	if !fixJobIDFound {
		jobID, err = h.getItemJobID(r.Context(), itemID)
	}
	if !fixJobIDFound && err != nil {
		log.Printf("Error getting job ID for item %d: %v", itemID, err)
		// The item was resolved successfully; the count update is
		// non-critical so we still return success.
	} else {
		if err := h.db.UpdateJobResolvedCount(r.Context(), jobID); err != nil {
			log.Printf("Error updating job resolved count: %v", err)
		}
		if err := h.db.UpdateJobEnrichmentCount(r.Context(), jobID); err != nil {
			log.Printf("Error updating job enrichment count: %v", err)
		}

		// Auto-trigger enrichment if all Phase 1 items are resolved
		updatedJob, jobErr := h.db.GetAnalysisJob(r.Context(), jobID)
		if jobErr == nil &&
			updatedJob.ResolvedItems == updatedJob.TotalItems &&
			updatedJob.EnrichmentTotal == 0 {
			h.TryAutoEnrich(r.Context(), jobID, userID)
		}
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status": "resolved",
	})
}

// BatchResolve handles PUT /api/campaigns/{id}/analysis/jobs/{jobId}/resolve-all
// Batch-resolves all pending analysis items for a job that match the given
// detectionType. Items are processed in reverse position order (highest
// position_start first) so that content fixes do not invalidate the byte
// offsets of items processed later.
func (h *ContentAnalysisHandler) BatchResolve(w http.ResponseWriter, r *http.Request) {
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

	// Verify the job exists and belongs to this campaign.
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

	var req BatchResolveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.DetectionType == "" {
		respondError(w, http.StatusBadRequest, "detectionType is required")
		return
	}

	switch req.Resolution {
	case "accepted", "dismissed":
		// valid
	default:
		respondError(w, http.StatusBadRequest,
			"Resolution must be one of: accepted, dismissed")
		return
	}

	// Fetch all pending items for this job.
	items, err := h.db.ListAnalysisItemsByJob(r.Context(), jobID, "pending", "")
	if err != nil {
		log.Printf("Error listing pending analysis items for job %d: %v",
			jobID, err)
		respondError(w, http.StatusInternalServerError,
			"Failed to list pending analysis items")
		return
	}

	// Filter to only items matching the requested detection type.
	var filtered []models.ContentAnalysisItem
	for _, item := range items {
		if item.DetectionType == req.DetectionType {
			filtered = append(filtered, item)
		}
	}

	// Sort by position_start DESCENDING so that content fixes are
	// applied back-to-front, preserving earlier byte offsets.
	sort.Slice(filtered, func(i, j int) bool {
		si := 0
		if filtered[i].PositionStart != nil {
			si = *filtered[i].PositionStart
		}
		sj := 0
		if filtered[j].PositionStart != nil {
			sj = *filtered[j].PositionStart
		}
		return si > sj
	})

	resolved := 0
	for _, item := range filtered {
		// Determine the resolved entity ID (same logic as ResolveItem
		// for the "accepted" case).
		var resolvedEntityID *int64
		if req.Resolution == "accepted" {
			resolvedEntityID = item.EntityID
		}

		// Resolve the item in the database.
		if err := h.db.ResolveAnalysisItem(
			r.Context(), item.ID, req.Resolution, resolvedEntityID,
		); err != nil {
			log.Printf("Error resolving analysis item %d in batch: %v",
				item.ID, err)
			continue
		}

		// Apply content fix for accepted resolutions that have
		// position offsets.
		if req.Resolution == "accepted" &&
			item.PositionStart != nil && item.PositionEnd != nil {
			var replacement string
			switch item.DetectionType {
			case "potential_alias":
				if resolvedEntityID != nil {
					var entityName string
					err := h.db.QueryRow(r.Context(),
						"SELECT name FROM entities WHERE id = $1",
						*resolvedEntityID,
					).Scan(&entityName)
					if err != nil {
						log.Printf(
							"Error fetching entity name for alias link (item %d): %v",
							item.ID, err)
						replacement = "[[" + item.MatchedText + "]]"
					} else {
						replacement = "[[" + entityName + "|" + item.MatchedText + "]]"
					}
				} else {
					replacement = "[[" + item.MatchedText + "]]"
				}
			case "misspelling":
				replacement = "[[" + item.MatchedText + "]]"
			default:
				// untagged_mention and any other type
				replacement = "[[" + item.MatchedText + "]]"
			}

			if fixErr := h.applyContentFix(
				r.Context(), job.SourceTable, job.SourceID,
				job.SourceField,
				*item.PositionStart, *item.PositionEnd,
				item.MatchedText, replacement,
			); fixErr != nil {
				log.Printf("Content fix failed for item %d in batch: %v",
					item.ID, fixErr)
			}
		}

		resolved++
	}

	// Update the job's resolved count to reflect all changes.
	if err := h.db.UpdateJobResolvedCount(r.Context(), jobID); err != nil {
		log.Printf("Error updating job resolved count for job %d: %v",
			jobID, err)
	}

	// Auto-trigger enrichment if all Phase 1 items are resolved
	updatedJob, jobErr := h.db.GetAnalysisJob(r.Context(), jobID)
	if jobErr == nil &&
		updatedJob.ResolvedItems == updatedJob.TotalItems &&
		updatedJob.EnrichmentTotal == 0 {
		h.TryAutoEnrich(r.Context(), jobID, userID)
	}

	respondJSON(w, http.StatusOK, BatchResolveResponse{Resolved: resolved})
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
// Returns the number of pending analysis items. When sourceTable and
// sourceId query parameters are provided, the count is scoped to that
// specific source; otherwise it returns the campaign-wide total.
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
	var sourceID int64
	if sourceIDStr := r.URL.Query().Get("sourceId"); sourceIDStr != "" {
		sourceID, err = strconv.ParseInt(sourceIDStr, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid sourceId")
			return
		}
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

// applyContentFix modifies the source content at the specified position,
// replacing the matched text with the given replacement string. It verifies
// that the content has not changed since analysis before applying the fix.
func (h *ContentAnalysisHandler) applyContentFix(
	ctx context.Context,
	sourceTable string,
	sourceID int64,
	sourceField string,
	posStart int,
	posEnd int,
	matchedText string,
	replacement string,
) error {
	// Build safe SQL queries based on table and field combination.
	// Each combination uses a hardcoded query to prevent SQL injection.
	var selectSQL, updateSQL string

	switch sourceTable {
	case "entities":
		switch sourceField {
		case "description":
			selectSQL = "SELECT COALESCE(description, '') FROM entities WHERE id = $1"
			updateSQL = "UPDATE entities SET description = $2, updated_at = NOW() WHERE id = $1"
		case "gm_notes":
			selectSQL = "SELECT COALESCE(gm_notes, '') FROM entities WHERE id = $1"
			updateSQL = "UPDATE entities SET gm_notes = $2, updated_at = NOW() WHERE id = $1"
		default:
			return fmt.Errorf("unsupported field %q for table %q",
				sourceField, sourceTable)
		}
	case "chapters":
		switch sourceField {
		case "overview":
			selectSQL = "SELECT COALESCE(overview, '') FROM chapters WHERE id = $1"
			updateSQL = "UPDATE chapters SET overview = $2, updated_at = NOW() WHERE id = $1"
		default:
			return fmt.Errorf("unsupported field %q for table %q",
				sourceField, sourceTable)
		}
	case "sessions":
		switch sourceField {
		case "prep_notes":
			selectSQL = "SELECT COALESCE(prep_notes, '') FROM sessions WHERE id = $1"
			updateSQL = "UPDATE sessions SET prep_notes = $2, updated_at = NOW() WHERE id = $1"
		case "actual_notes":
			selectSQL = "SELECT COALESCE(actual_notes, '') FROM sessions WHERE id = $1"
			updateSQL = "UPDATE sessions SET actual_notes = $2, updated_at = NOW() WHERE id = $1"
		default:
			return fmt.Errorf("unsupported field %q for table %q",
				sourceField, sourceTable)
		}
	case "campaigns":
		switch sourceField {
		case "description":
			selectSQL = "SELECT COALESCE(description, '') FROM campaigns WHERE id = $1"
			updateSQL = "UPDATE campaigns SET description = $2, updated_at = NOW() WHERE id = $1"
		default:
			return fmt.Errorf("unsupported field %q for table %q",
				sourceField, sourceTable)
		}
	default:
		return fmt.Errorf("unsupported source table: %s", sourceTable)
	}

	// Fetch the current content.
	var content string
	if err := h.db.QueryRow(ctx, selectSQL, sourceID).Scan(&content); err != nil {
		return fmt.Errorf("failed to fetch content from %s.%s: %w",
			sourceTable, sourceField, err)
	}

	// Stale offset protection: verify the text at the recorded position
	// still matches what the analysis detected.
	if posStart < 0 || posEnd > len(content) || posStart > posEnd {
		return fmt.Errorf(
			"position offsets [%d:%d] are out of range for content of length %d",
			posStart, posEnd, len(content))
	}

	currentText := content[posStart:posEnd]
	if currentText != matchedText {
		return fmt.Errorf(
			"content has changed since analysis: expected %q at [%d:%d], found %q",
			matchedText, posStart, posEnd, currentText)
	}

	// Build the new content by splicing in the replacement.
	newContent := content[:posStart] + replacement + content[posEnd:]

	// Update the source record with the new content.
	err := h.db.Exec(ctx, updateSQL, sourceID, newContent)
	if err != nil {
		return fmt.Errorf("failed to update content in %s.%s: %w",
			sourceTable, sourceField, err)
	}

	return nil
}

// wikiLinkPattern matches wiki links of the form [[text]] or
// [[entity|display]].
var wikiLinkPattern = regexp.MustCompile(`\[\[([^\]|]+?)(?:\|([^\]]+?))?\]\]`)

// RevertItem handles PUT /api/campaigns/{id}/analysis/items/{itemId}/revert
// Reverts a previously resolved content analysis item back to pending
// status and removes the wiki link that was inserted into the source
// content.
func (h *ContentAnalysisHandler) RevertItem(w http.ResponseWriter, r *http.Request) {
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

	// Verify the item belongs to a job within the specified campaign.
	var itemCampaignID int64
	err = h.db.QueryRow(r.Context(),
		`SELECT j.campaign_id FROM content_analysis_items i
		 JOIN content_analysis_jobs j ON i.job_id = j.id
		 WHERE i.id = $1`,
		itemID,
	).Scan(&itemCampaignID)
	if errors.Is(err, pgx.ErrNoRows) {
		respondError(w, http.StatusNotFound, "Analysis item not found")
		return
	}
	if err != nil {
		log.Printf("Error verifying item campaign ownership: %v", err)
		respondError(w, http.StatusInternalServerError,
			"Failed to verify analysis item")
		return
	}
	if itemCampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Analysis item not found")
		return
	}

	// Fetch the item details including its current resolution and
	// position information needed for reverting the content fix.
	var (
		posStart, posEnd        *int
		matchedText, resolution string
		srcTable, srcField      string
		srcID, jobID            int64
	)
	err = h.db.QueryRow(r.Context(),
		`SELECT i.position_start, i.position_end, i.matched_text,
		        i.resolution, i.job_id,
		        j.source_table, j.source_id, j.source_field
		 FROM content_analysis_items i
		 JOIN content_analysis_jobs j ON i.job_id = j.id
		 WHERE i.id = $1`,
		itemID,
	).Scan(&posStart, &posEnd, &matchedText, &resolution,
		&jobID, &srcTable, &srcID, &srcField)
	if err != nil {
		log.Printf("Error fetching item details for revert (item %d): %v",
			itemID, err)
		respondError(w, http.StatusInternalServerError,
			"Failed to fetch analysis item details")
		return
	}

	// The item must already be resolved (not pending) to be reverted.
	if resolution == "pending" {
		respondError(w, http.StatusBadRequest,
			"Item is already pending and cannot be reverted")
		return
	}

	// If the original resolution inserted a wiki link (accepted or
	// new_entity) and position data is available, attempt to remove
	// the wiki link from the source content.
	if (resolution == "accepted" || resolution == "new_entity") &&
		posStart != nil && posEnd != nil {

		delta, fixErr := h.revertContentFix(
			r.Context(), srcTable, srcID, srcField,
			*posStart, matchedText,
		)
		if fixErr != nil {
			// Content revert failure is non-fatal: log but still
			// proceed with unresolving the item.
			log.Printf("Content revert failed for item %d: %v",
				itemID, fixErr)
		} else if delta != 0 {
			// Adjust byte offsets of remaining pending items in
			// the same job. The wiki link brackets added extra
			// characters that are now removed, so delta is
			// negative.
			adjErr := h.db.Exec(r.Context(),
				`UPDATE content_analysis_items
				 SET position_start = position_start + $1,
				     position_end = position_end + $1
				 WHERE job_id = $2
				   AND id != $3
				   AND resolution = 'pending'
				   AND position_start >= $4`,
				delta, jobID, itemID, *posStart,
			)
			if adjErr != nil {
				log.Printf(
					"Failed to adjust offsets for job %d after revert of item %d: %v",
					jobID, itemID, adjErr)
			}
		}
	}

	// Set the item back to pending.
	unresolveErr := h.db.Exec(r.Context(),
		`UPDATE content_analysis_items
		 SET resolution = 'pending',
		     resolved_entity_id = NULL,
		     resolved_at = NULL
		 WHERE id = $1`,
		itemID,
	)
	if unresolveErr != nil {
		log.Printf("Error unresolving analysis item %d: %v",
			itemID, unresolveErr)
		respondError(w, http.StatusInternalServerError,
			"Failed to revert analysis item")
		return
	}

	// Update the job's resolved count.
	if err := h.db.UpdateJobResolvedCount(r.Context(), jobID); err != nil {
		log.Printf("Error updating job resolved count: %v", err)
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status": "reverted",
	})
}

// revertContentFix locates and removes a wiki link in the source content
// that was inserted when the analysis item was resolved. It searches for
// a wiki link pattern near the recorded position whose display text
// matches the item's matched text. On success it returns the byte offset
// delta (negative, since brackets are removed).
func (h *ContentAnalysisHandler) revertContentFix(
	ctx context.Context,
	sourceTable string,
	sourceID int64,
	sourceField string,
	posStart int,
	matchedText string,
) (int, error) {
	// Fetch the current source text.
	content, err := h.fetchSourceContent(
		ctx, sourceTable, sourceID, sourceField,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch content for revert: %w", err)
	}

	// Define a search window around the expected position (+-50 chars)
	// to account for drift from other fixes.
	windowStart := posStart - 50
	if windowStart < 0 {
		windowStart = 0
	}
	windowEnd := posStart + len(matchedText) + 50
	if windowEnd > len(content) {
		windowEnd = len(content)
	}

	window := content[windowStart:windowEnd]

	// Find all wiki links in the window and pick the one whose
	// display text matches the item's matched text.
	matches := wikiLinkPattern.FindAllStringSubmatchIndex(window, -1)
	foundOffset := -1
	foundLen := 0
	for _, m := range matches {
		// m[0]:m[1] is the full match [[...]]
		// m[2]:m[3] is capture group 1 (entity name or simple text)
		// m[4]:m[5] is capture group 2 (display text in alias form),
		//           or -1 if not present.
		var displayText string
		if m[4] >= 0 && m[5] >= 0 {
			// Alias form: [[entity|display]]
			displayText = window[m[4]:m[5]]
		} else {
			// Simple form: [[display]]
			displayText = window[m[2]:m[3]]
		}

		if displayText == matchedText {
			foundOffset = windowStart + m[0]
			foundLen = m[1] - m[0]
			break
		}
	}

	if foundOffset < 0 {
		return 0, fmt.Errorf(
			"wiki link containing %q not found near position %d",
			matchedText, posStart)
	}

	// Compute the delta before modifying content. The wiki link is
	// replaced by just the plain matched text, so the content shrinks
	// by (linkLength - matchedTextLength).
	delta := len(matchedText) - foundLen

	// Replace the wiki link with just the plain matched text.
	newContent := content[:foundOffset] + matchedText +
		content[foundOffset+foundLen:]

	// Build the update SQL using the same safe switch pattern as
	// applyContentFix.
	var updateSQL string
	switch sourceTable {
	case "entities":
		switch sourceField {
		case "description":
			updateSQL = "UPDATE entities SET description = $2, updated_at = NOW() WHERE id = $1"
		case "gm_notes":
			updateSQL = "UPDATE entities SET gm_notes = $2, updated_at = NOW() WHERE id = $1"
		default:
			return 0, fmt.Errorf("unsupported field %q for table %q",
				sourceField, sourceTable)
		}
	case "chapters":
		switch sourceField {
		case "overview":
			updateSQL = "UPDATE chapters SET overview = $2, updated_at = NOW() WHERE id = $1"
		default:
			return 0, fmt.Errorf("unsupported field %q for table %q",
				sourceField, sourceTable)
		}
	case "sessions":
		switch sourceField {
		case "prep_notes":
			updateSQL = "UPDATE sessions SET prep_notes = $2, updated_at = NOW() WHERE id = $1"
		case "actual_notes":
			updateSQL = "UPDATE sessions SET actual_notes = $2, updated_at = NOW() WHERE id = $1"
		default:
			return 0, fmt.Errorf("unsupported field %q for table %q",
				sourceField, sourceTable)
		}
	case "campaigns":
		switch sourceField {
		case "description":
			updateSQL = "UPDATE campaigns SET description = $2, updated_at = NOW() WHERE id = $1"
		default:
			return 0, fmt.Errorf("unsupported field %q for table %q",
				sourceField, sourceTable)
		}
	default:
		return 0, fmt.Errorf("unsupported source table: %s", sourceTable)
	}

	if err := h.db.Exec(ctx, updateSQL, sourceID, newContent); err != nil {
		return 0, fmt.Errorf("failed to update content in %s.%s: %w",
			sourceTable, sourceField, err)
	}

	return delta, nil
}

// handleRelationshipSuggestion creates a relationship from an accepted
// relationship_suggestion enrichment item. The single-edge model stores
// only the canonical forward direction; the database view provides both
// perspectives, so no inverse row is needed.
func (h *ContentAnalysisHandler) handleRelationshipSuggestion(
	ctx context.Context,
	campaignID int64,
	jobID int64,
	itemID int64,
	suggestedContent json.RawMessage,
	override map[string]interface{},
) {
	var suggestion models.RelationshipSuggestion
	if err := json.Unmarshal(suggestedContent, &suggestion); err != nil {
		log.Printf("Error parsing relationship suggestion for item %d: %v",
			itemID, err)
		return
	}

	// Determine the final relationship type name, allowing the user to
	// override the LLM-suggested type from the triage UI.
	relationshipType := suggestion.RelationshipType
	if override != nil {
		if overriddenType, ok := override["relationshipType"].(string); ok && overriddenType != "" {
			relationshipType = overriddenType
		}
	}

	// Look up the relationship type ID by name.
	relTypes, err := h.db.ListRelationshipTypes(ctx, campaignID)
	if err != nil {
		log.Printf("Error listing relationship types for campaign %d: %v",
			campaignID, err)
		return
	}

	var relTypeID int64
	for _, rt := range relTypes {
		if rt.Name == relationshipType {
			relTypeID = rt.ID
			break
		}
	}
	if relTypeID == 0 {
		log.Printf("Relationship type %q not found for campaign %d",
			relationshipType, campaignID)
		return
	}

	// Create the forward relationship only. The entity_relationships_view
	// automatically provides the inverse perspective.
	desc := suggestion.Description
	_, err = h.db.CreateRelationship(ctx, campaignID, models.CreateRelationshipRequest{
		SourceEntityID:     suggestion.SourceEntityID,
		TargetEntityID:     suggestion.TargetEntityID,
		RelationshipTypeID: relTypeID,
		Description:        &desc,
	})
	if err != nil {
		// Gracefully handle unique_violation from the inverse-pair
		// trigger. This occurs when the inverse edge already exists.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			log.Printf("Relationship already exists (unique_violation) "+
				"for item %d, skipping: %v", itemID, err)
			return
		}
		log.Printf("Error creating relationship from suggestion (item %d): %v",
			itemID, err)
		return
	}

	log.Printf("Created relationship %s -> %s (%s) from enrichment item %d",
		suggestion.SourceEntityName, suggestion.TargetEntityName,
		relationshipType, itemID)
}

// handleDescriptionUpdate applies an accepted description update to
// the entity.
func (h *ContentAnalysisHandler) handleDescriptionUpdate(
	ctx context.Context,
	entityID int64,
	suggestedContent json.RawMessage,
) {
	var suggestion models.DescriptionUpdateSuggestion
	if err := json.Unmarshal(suggestedContent, &suggestion); err != nil {
		log.Printf("handleDescriptionUpdate: failed to unmarshal suggestion for entity %d: %v",
			entityID, err)
		return
	}

	_, err := h.db.UpdateEntity(ctx, entityID, models.UpdateEntityRequest{
		Description: &suggestion.SuggestedDescription,
	})
	if err != nil {
		log.Printf("handleDescriptionUpdate: failed to update entity %d description: %v",
			entityID, err)
	}
}

// handleLogEntry creates a new entity log entry from an accepted
// log_entry enrichment suggestion.
func (h *ContentAnalysisHandler) handleLogEntry(
	ctx context.Context,
	campaignID int64,
	entityID int64,
	suggestedContent json.RawMessage,
) {
	var suggestion models.LogEntrySuggestion
	if err := json.Unmarshal(suggestedContent, &suggestion); err != nil {
		log.Printf("handleLogEntry: failed to unmarshal suggestion for entity %d: %v",
			entityID, err)
		return
	}

	_, err := h.db.CreateEntityLog(ctx, entityID, campaignID, models.CreateEntityLogRequest{
		Content:    suggestion.Content,
		OccurredAt: suggestion.OccurredAt,
	})
	if err != nil {
		log.Printf("handleLogEntry: failed to create log entry for entity %d: %v",
			entityID, err)
	}
}

// CancelEnrichment handles POST /api/campaigns/{id}/analysis/jobs/{jobId}/cancel-enrichment
// Cancels a running LLM enrichment for the specified job.
func (h *ContentAnalysisHandler) CancelEnrichment(w http.ResponseWriter, r *http.Request) {
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

	// Verify the job belongs to this campaign.
	job, err := h.db.GetAnalysisJob(r.Context(), jobID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Analysis job not found")
		return
	}
	if job.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Analysis job not found")
		return
	}

	if job.Status != "enriching" {
		respondError(w, http.StatusConflict, "Enrichment is not running")
		return
	}

	// Cancel the enrichment goroutine if running.
	if cancelVal, ok := h.enrichCancels.LoadAndDelete(jobID); ok {
		cancelVal.(context.CancelFunc)()
	}

	// Set job status to completed.
	if err := h.db.Exec(r.Context(),
		"UPDATE content_analysis_jobs SET status = 'completed' WHERE id = $1",
		jobID,
	); err != nil {
		log.Printf("CancelEnrichment: failed to update job %d status: %v", jobID, err)
		respondError(w, http.StatusInternalServerError, "Failed to update job status")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

// RunContentEnrichment runs LLM enrichment directly on content without
// depending on Phase 1 analysis results. It delegates entity discovery
// and enrichment to the Pipeline and EnrichmentAgent. Results are saved
// as enrichment-phase analysis items on the given job. The method runs
// enrichment in a background goroutine and returns immediately.
func (h *ContentAnalysisHandler) RunContentEnrichment(
	ctx context.Context,
	jobID int64,
	campaignID int64,
	content string,
	userID int64,
) {
	// 1. Fetch user settings to check LLM configuration.
	settings, err := h.db.GetUserSettings(ctx, userID)
	if err != nil || settings == nil {
		log.Printf("Content-enrich: skipping job %d — no user settings found for user %d",
			jobID, userID)
		return
	}
	if settings.ContentGenService == nil || settings.ContentGenAPIKey == nil {
		log.Printf("Content-enrich: skipping job %d — no LLM configured (service=%v, key=%v)",
			jobID, settings.ContentGenService != nil, settings.ContentGenAPIKey != nil)
		return
	}

	// 2. Create LLM provider from user settings.
	provider, err := llm.NewProvider(
		*settings.ContentGenService, *settings.ContentGenAPIKey,
	)
	if err != nil {
		log.Printf("Content-enrich: failed to create LLM provider for job %d: %v",
			jobID, err)
		return
	}

	// 3. Get the job for source info.
	job, err := h.db.GetAnalysisJob(ctx, jobID)
	if err != nil {
		log.Printf("Content-enrich: failed to get job %d: %v", jobID, err)
		return
	}

	// 4. Set job status to "enriching".
	if err := h.db.Exec(ctx,
		"UPDATE content_analysis_jobs SET status = 'enriching' WHERE id = $1",
		jobID,
	); err != nil {
		log.Printf("Content-enrich: failed to set job %d status: %v",
			jobID, err)
		return
	}

	// 5. Build pipeline and spawn background goroutine for enrichment.
	ttrpgAgent := ttrpg.NewExpert()
	canonAgent := canon.NewExpert()
	enrichAgent := enrichment.NewEnrichmentAgent(h.db)
	graphAgent := graph.NewExpert(h.db)
	pipeline := enrichment.NewPipeline(h.db, []enrichment.Stage{
		{
			Name:   "analysis",
			Phase:  "analysis",
			Agents: []enrichment.PipelineAgent{ttrpgAgent, canonAgent},
		},
		{
			Name:   "enrichment",
			Phase:  "enrichment",
			Agents: []enrichment.PipelineAgent{enrichAgent, graphAgent},
		},
	})

	bgCtx, cancel := context.WithCancel(context.Background())
	h.enrichCancels.Store(jobID, cancel)
	go func() {
		defer h.enrichCancels.Delete(jobID)
		defer cancel()
		log.Printf("Content-enrich: starting enrichment for job %d", jobID)

		input := enrichment.PipelineInput{
			CampaignID:  campaignID,
			JobID:       jobID,
			SourceTable: job.SourceTable,
			SourceID:    job.SourceID,
			Content:     content,
		}

		enrichItems, err := pipeline.Run(bgCtx, provider, input)
		if err != nil {
			log.Printf("Content-enrich: pipeline run failed for job %d: %v",
				jobID, err)
		} else if len(enrichItems) > 0 {
			if err := h.db.CreateAnalysisItems(bgCtx, enrichItems); err != nil {
				log.Printf(
					"Content-enrich: failed to save items for job %d: %v",
					jobID, err)
			} else {
				if err := h.db.Exec(bgCtx,
					"UPDATE content_analysis_jobs SET enrichment_total = enrichment_total + $1 WHERE id = $2",
					len(enrichItems), jobID,
				); err != nil {
					log.Printf(
						"Content-enrich: failed to update enrichment count for job %d: %v",
						jobID, err)
				}
			}
		}

		// Mark enrichment as completed.
		log.Printf("Content-enrich: completed enrichment for job %d", jobID)
		if err := h.db.Exec(context.Background(),
			"UPDATE content_analysis_jobs SET status = 'completed' WHERE id = $1",
			jobID,
		); err != nil {
			log.Printf(
				"Content-enrich: failed to set job %d status to completed: %v",
				jobID, err)
		}
	}()
}

// TryAutoEnrich automatically triggers LLM enrichment when all Phase 1
// identification items have been resolved. It silently returns if the
// user has not configured an LLM provider or if there are no accepted
// entities to enrich.
func (h *ContentAnalysisHandler) TryAutoEnrich(
	ctx context.Context, jobID int64, userID int64,
) {
	// 1. Fetch user settings to check LLM configuration.
	settings, err := h.db.GetUserSettings(ctx, userID)
	if err != nil || settings == nil {
		log.Printf("Auto-enrich: skipping job %d — no user settings found for user %d", jobID, userID)
		return
	}
	if settings.ContentGenService == nil || settings.ContentGenAPIKey == nil {
		log.Printf("Auto-enrich: skipping job %d — no LLM configured (service=%v, key=%v)",
			jobID, settings.ContentGenService != nil, settings.ContentGenAPIKey != nil)
		return
	}

	// 2. Create LLM provider from user settings.
	provider, err := llm.NewProvider(
		*settings.ContentGenService, *settings.ContentGenAPIKey,
	)
	if err != nil {
		log.Printf("Auto-enrich: failed to create LLM provider for job %d: %v",
			jobID, err)
		return
	}

	// 3. Get the job for source info.
	job, err := h.db.GetAnalysisJob(ctx, jobID)
	if err != nil {
		log.Printf("Auto-enrich: failed to get job %d: %v", jobID, err)
		return
	}

	// 4. Collect accepted items with resolved entity IDs.
	items, err := h.db.ListAnalysisItemsByJob(ctx, jobID, "", "identification")
	if err != nil {
		log.Printf("Auto-enrich: failed to list items for job %d: %v",
			jobID, err)
		return
	}

	entityIDSet := make(map[int64]bool)
	for _, item := range items {
		if (item.Resolution == "accepted" || item.Resolution == "new_entity") &&
			item.ResolvedEntityID != nil {
			entityIDSet[*item.ResolvedEntityID] = true
		}
	}

	if len(entityIDSet) == 0 {
		log.Printf("Auto-enrich: skipping job %d — no accepted entities with resolved_entity_id", jobID)
		return
	}

	entityIDs := make([]int64, 0, len(entityIDSet))
	for id := range entityIDSet {
		entityIDs = append(entityIDs, id)
	}

	// 5. Set job status to "enriching".
	if err := h.db.Exec(ctx,
		"UPDATE content_analysis_jobs SET status = 'enriching' WHERE id = $1",
		jobID,
	); err != nil {
		log.Printf("Auto-enrich: failed to set job %d status: %v",
			jobID, err)
		return
	}

	// 6. Fetch source content.
	content, err := h.fetchSourceContent(
		ctx, job.SourceTable, job.SourceID, job.SourceField,
	)
	if err != nil {
		log.Printf("Auto-enrich: failed to fetch content for job %d: %v",
			jobID, err)
		_ = h.db.Exec(ctx,
			"UPDATE content_analysis_jobs SET status = 'completed' WHERE id = $1",
			jobID,
		)
		return
	}

	// 7. Pre-load entity objects from accepted entity IDs.
	entities := make([]models.Entity, 0, len(entityIDs))
	for _, eid := range entityIDs {
		entity, err := h.db.GetEntity(ctx, eid)
		if err != nil {
			log.Printf("Auto-enrich: failed to get entity %d: %v", eid, err)
			continue
		}
		entities = append(entities, *entity)
	}

	// 8. Build pipeline and spawn background goroutine for enrichment.
	ttrpgAgent := ttrpg.NewExpert()
	canonAgent := canon.NewExpert()
	enrichAgent := enrichment.NewEnrichmentAgent(h.db)
	graphAgent := graph.NewExpert(h.db)
	pipeline := enrichment.NewPipeline(h.db, []enrichment.Stage{
		{
			Name:   "analysis",
			Phase:  "analysis",
			Agents: []enrichment.PipelineAgent{ttrpgAgent, canonAgent},
		},
		{
			Name:   "enrichment",
			Phase:  "enrichment",
			Agents: []enrichment.PipelineAgent{enrichAgent, graphAgent},
		},
	})

	bgCtx, cancel := context.WithCancel(context.Background())
	h.enrichCancels.Store(jobID, cancel)
	go func() {
		defer h.enrichCancels.Delete(jobID)
		defer cancel()
		log.Printf("Auto-enrich: starting enrichment for job %d with %d entities",
			jobID, len(entities))

		input := enrichment.PipelineInput{
			CampaignID:  job.CampaignID,
			JobID:       jobID,
			SourceTable: job.SourceTable,
			SourceID:    job.SourceID,
			Content:     content,
			Entities:    entities,
		}

		enrichItems, err := pipeline.Run(bgCtx, provider, input)
		if err != nil {
			log.Printf("Auto-enrich: pipeline run failed for job %d: %v",
				jobID, err)
		} else if len(enrichItems) > 0 {
			if err := h.db.CreateAnalysisItems(bgCtx, enrichItems); err != nil {
				log.Printf(
					"Auto-enrich: failed to save items for job %d: %v",
					jobID, err)
			} else {
				if err := h.db.Exec(bgCtx,
					"UPDATE content_analysis_jobs SET enrichment_total = enrichment_total + $1 WHERE id = $2",
					len(enrichItems), jobID,
				); err != nil {
					log.Printf(
						"Auto-enrich: failed to update enrichment count for job %d: %v",
						jobID, err)
				}
			}
		}

		// Mark enrichment as completed.
		log.Printf("Auto-enrich: completed enrichment for job %d", jobID)
		if err := h.db.Exec(context.Background(),
			"UPDATE content_analysis_jobs SET status = 'completed' WHERE id = $1",
			jobID,
		); err != nil {
			log.Printf(
				"Auto-enrich: failed to set job %d status to completed: %v",
				jobID, err)
		}
	}()
}

// GenerateRevisionResponse is the response body for the generate
// revision endpoint.
type GenerateRevisionResponse struct {
	RevisedContent  string `json:"revisedContent"`
	Summary         string `json:"summary"`
	OriginalContent string `json:"originalContent"`
}

// ApplyRevisionRequest is the request body for applying a revision to
// the source content.
type ApplyRevisionRequest struct {
	RevisedContent string `json:"revisedContent"`
}

// GenerateRevision handles POST /api/campaigns/{id}/analysis/jobs/{jobId}/revision
// Generates a revised version of the source content by incorporating
// accepted analysis findings via the RevisionAgent LLM pipeline.
func (h *ContentAnalysisHandler) GenerateRevision(w http.ResponseWriter, r *http.Request) {
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

	// Verify the job exists and belongs to this campaign.
	job, err := h.db.GetAnalysisJob(r.Context(), jobID)
	if err != nil {
		log.Printf("GenerateRevision: error getting analysis job %d: %v", jobID, err)
		respondError(w, http.StatusNotFound, "Analysis job not found")
		return
	}
	if job.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Analysis job not found")
		return
	}

	// Fetch the source content based on the job's source table.
	var originalContent string
	switch job.SourceTable {
	case "chapters":
		chapter, chErr := h.db.GetChapter(r.Context(), job.SourceID)
		if chErr != nil {
			log.Printf("GenerateRevision: error fetching chapter %d: %v",
				job.SourceID, chErr)
			respondError(w, http.StatusInternalServerError,
				"Failed to fetch source content")
			return
		}
		if chapter.Overview != nil {
			originalContent = *chapter.Overview
		}
	case "sessions":
		session, sErr := h.db.GetSession(r.Context(), job.SourceID)
		if sErr != nil {
			log.Printf("GenerateRevision: error fetching session %d: %v",
				job.SourceID, sErr)
			respondError(w, http.StatusInternalServerError,
				"Failed to fetch source content")
			return
		}
		// Use the field indicated by the job's source_field.
		switch job.SourceField {
		case "prep_notes":
			if session.PrepNotes != nil {
				originalContent = *session.PrepNotes
			}
		case "actual_notes":
			if session.ActualNotes != nil {
				originalContent = *session.ActualNotes
			}
		default:
			respondError(w, http.StatusBadRequest,
				fmt.Sprintf("Unsupported session field: %s", job.SourceField))
			return
		}
	default:
		respondError(w, http.StatusBadRequest,
			fmt.Sprintf("Unsupported source table for revision: %s", job.SourceTable))
		return
	}

	// Get accepted analysis items for this job.
	acceptedItems, err := h.db.ListAnalysisItemsByJob(
		r.Context(), jobID, "acknowledged", "analysis",
	)
	if err != nil {
		log.Printf("GenerateRevision: error listing accepted items for job %d: %v",
			jobID, err)
		respondError(w, http.StatusInternalServerError,
			"Failed to list accepted analysis items")
		return
	}

	// If no accepted items, return the original content unchanged.
	if len(acceptedItems) == 0 {
		respondJSON(w, http.StatusOK, GenerateRevisionResponse{
			RevisedContent:  originalContent,
			Summary:         "",
			OriginalContent: originalContent,
		})
		return
	}

	// Get the LLM provider from user settings.
	settings, err := h.db.GetUserSettings(r.Context(), userID)
	if err != nil {
		log.Printf("GenerateRevision: error getting user settings: %v", err)
		respondError(w, http.StatusInternalServerError,
			"Failed to get user settings")
		return
	}
	if settings == nil || settings.ContentGenService == nil || settings.ContentGenAPIKey == nil {
		respondError(w, http.StatusBadRequest,
			"LLM service not configured. Configure an LLM in Account Settings.")
		return
	}

	provider, err := llm.NewProvider(
		*settings.ContentGenService, *settings.ContentGenAPIKey,
	)
	if err != nil {
		log.Printf("GenerateRevision: error creating LLM provider: %v", err)
		respondError(w, http.StatusInternalServerError,
			"Failed to create LLM provider")
		return
	}

	// Call the RevisionAgent to generate revised content.
	revisionInput := enrichment.RevisionInput{
		OriginalContent: originalContent,
		AcceptedItems:   acceptedItems,
		SourceTable:     job.SourceTable,
		SourceID:        job.SourceID,
	}

	result, err := enrichment.NewRevisionAgent().GenerateRevision(
		r.Context(), provider, revisionInput,
	)
	if err != nil {
		log.Printf("GenerateRevision: revision agent failed for job %d: %v",
			jobID, err)
		respondError(w, http.StatusInternalServerError,
			"Failed to generate revision")
		return
	}

	respondJSON(w, http.StatusOK, GenerateRevisionResponse{
		RevisedContent:  result.RevisedContent,
		Summary:         result.Summary,
		OriginalContent: originalContent,
	})
}

// ApplyRevision handles PUT /api/campaigns/{id}/analysis/jobs/{jobId}/revision/apply
// Applies a previously generated revision to the source content, updating
// the underlying chapter overview or session notes field.
func (h *ContentAnalysisHandler) ApplyRevision(w http.ResponseWriter, r *http.Request) {
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

	var req ApplyRevisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.RevisedContent == "" {
		respondError(w, http.StatusBadRequest, "revisedContent is required")
		return
	}

	// Verify the job exists and belongs to this campaign.
	job, err := h.db.GetAnalysisJob(r.Context(), jobID)
	if err != nil {
		log.Printf("ApplyRevision: error getting analysis job %d: %v", jobID, err)
		respondError(w, http.StatusNotFound, "Analysis job not found")
		return
	}
	if job.CampaignID != campaignID {
		respondError(w, http.StatusNotFound, "Analysis job not found")
		return
	}

	// Apply the revised content to the appropriate source.
	switch job.SourceTable {
	case "chapters":
		chapter, chErr := h.db.GetChapter(r.Context(), job.SourceID)
		if chErr != nil {
			log.Printf("ApplyRevision: error fetching chapter %d: %v",
				job.SourceID, chErr)
			respondError(w, http.StatusInternalServerError,
				"Failed to fetch source content")
			return
		}
		_ = chapter // Verify the chapter exists before updating.
		_, updateErr := h.db.UpdateChapter(r.Context(), job.SourceID,
			models.UpdateChapterRequest{
				Overview: &req.RevisedContent,
			},
		)
		if updateErr != nil {
			log.Printf("ApplyRevision: error updating chapter %d: %v",
				job.SourceID, updateErr)
			respondError(w, http.StatusInternalServerError,
				"Failed to apply revision to chapter")
			return
		}

	case "sessions":
		session, sErr := h.db.GetSession(r.Context(), job.SourceID)
		if sErr != nil {
			log.Printf("ApplyRevision: error fetching session %d: %v",
				job.SourceID, sErr)
			respondError(w, http.StatusInternalServerError,
				"Failed to fetch source content")
			return
		}
		_ = session // Verify the session exists before updating.

		updateReq := models.UpdateSessionRequest{}
		switch job.SourceField {
		case "prep_notes":
			updateReq.PrepNotes = &req.RevisedContent
		case "actual_notes":
			updateReq.ActualNotes = &req.RevisedContent
		default:
			respondError(w, http.StatusBadRequest,
				fmt.Sprintf("Unsupported session field: %s", job.SourceField))
			return
		}

		_, updateErr := h.db.UpdateSession(r.Context(), job.SourceID, updateReq)
		if updateErr != nil {
			log.Printf("ApplyRevision: error updating session %d: %v",
				job.SourceID, updateErr)
			respondError(w, http.StatusInternalServerError,
				"Failed to apply revision to session")
			return
		}

	default:
		respondError(w, http.StatusBadRequest,
			fmt.Sprintf("Unsupported source table for revision: %s", job.SourceTable))
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status": "applied",
	})
}
