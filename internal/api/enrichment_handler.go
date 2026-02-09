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
	"time"

	"github.com/antonypegg/imagineer/internal/auth"
	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/enrichment"
	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/antonypegg/imagineer/internal/models"
)

// EnrichmentHandler handles enrichment trigger and streaming API requests.
type EnrichmentHandler struct {
	db *database.DB
}

// NewEnrichmentHandler creates a new EnrichmentHandler.
func NewEnrichmentHandler(db *database.DB) *EnrichmentHandler {
	return &EnrichmentHandler{db: db}
}

// triggerEnrichmentResponse is the response body for the trigger
// enrichment endpoint.
type triggerEnrichmentResponse struct {
	Status      string `json:"status"`
	EntityCount int    `json:"entityCount,omitempty"`
	Message     string `json:"message,omitempty"`
}

// TriggerEnrichment handles POST /api/campaigns/{id}/analysis/jobs/{jobId}/enrich
// Triggers LLM-based enrichment for all accepted Phase 1 entities in a job.
func (h *EnrichmentHandler) TriggerEnrichment(w http.ResponseWriter, r *http.Request) {
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

	// Get user settings for LLM configuration
	settings, err := h.db.GetUserSettings(r.Context(), userID)
	if err != nil {
		log.Printf("Error getting user settings: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get user settings")
		return
	}

	if settings == nil || settings.ContentGenService == nil || settings.ContentGenAPIKey == nil {
		respondError(w, http.StatusBadRequest,
			"LLM service not configured. Configure an LLM in Account Settings.")
		return
	}

	// Create LLM provider
	provider, err := llm.NewProvider(*settings.ContentGenService, *settings.ContentGenAPIKey)
	if err != nil {
		log.Printf("Error creating LLM provider: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create LLM provider")
		return
	}

	// Collect accepted Phase 1 items with resolved entity IDs
	items, err := h.db.ListAnalysisItemsByJob(r.Context(), jobID, "", "identification")
	if err != nil {
		log.Printf("Error listing analysis items: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list analysis items")
		return
	}

	// Gather unique entity IDs from accepted/new_entity items
	entityIDSet := make(map[int64]bool)
	for _, item := range items {
		if (item.Resolution == "accepted" || item.Resolution == "new_entity") &&
			item.ResolvedEntityID != nil {
			entityIDSet[*item.ResolvedEntityID] = true
		}
	}

	if len(entityIDSet) == 0 {
		respondJSON(w, http.StatusOK, triggerEnrichmentResponse{
			Status:  "no_entities",
			Message: "No accepted entities to enrich",
		})
		return
	}

	entityIDs := make([]int64, 0, len(entityIDSet))
	for id := range entityIDSet {
		entityIDs = append(entityIDs, id)
	}

	// Update job status to enriching
	if err := h.db.Exec(r.Context(),
		"UPDATE content_analysis_jobs SET status = 'enriching' WHERE id = $1",
		jobID,
	); err != nil {
		log.Printf("Error updating job status: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update job status")
		return
	}

	// Fetch source content for the enrichment engine
	content, err := h.fetchSourceContent(r.Context(), job.SourceTable, job.SourceID, job.SourceField)
	if err != nil {
		log.Printf("Error fetching source content: %v", err)
		// Reset job status and return error
		_ = h.db.Exec(r.Context(),
			"UPDATE content_analysis_jobs SET status = 'completed' WHERE id = $1",
			jobID,
		)
		respondError(w, http.StatusInternalServerError, "Failed to fetch source content")
		return
	}

	// Fetch all campaign entities for context
	allEntities, err := h.db.ListEntitiesByCampaign(r.Context(), campaignID)
	if err != nil {
		log.Printf("Error listing campaign entities: %v", err)
		allEntities = []models.Entity{}
	}

	// Create enrichment engine
	engine := enrichment.NewEngine(h.db)

	// Spawn background goroutine for enrichment processing
	go func() {
		bgCtx := context.Background()

		defer func() {
			if r := recover(); r != nil {
				log.Printf("Enrichment: panic in background goroutine for job %d: %v",
					jobID, r)
				if err := h.db.Exec(bgCtx,
					"UPDATE content_analysis_jobs SET status = 'failed' WHERE id = $1",
					jobID,
				); err != nil {
					log.Printf("Enrichment: error setting job %d to failed after panic: %v",
						jobID, err)
				}
			}
		}()

		for _, entityID := range entityIDs {
			entity, err := h.db.GetEntity(bgCtx, entityID)
			if err != nil {
				log.Printf("Enrichment: error fetching entity %d: %v", entityID, err)
				continue
			}

			relationships, err := h.db.GetEntityRelationships(bgCtx, entityID)
			if err != nil {
				log.Printf("Enrichment: error fetching relationships for entity %d: %v",
					entityID, err)
				relationships = []models.Relationship{}
			}

			// Build list of other entities (exclude the current entity)
			otherEntities := make([]models.Entity, 0, len(allEntities)-1)
			for _, e := range allEntities {
				if e.ID != entityID {
					otherEntities = append(otherEntities, e)
				}
			}

			input := enrichment.EnrichmentInput{
				CampaignID:    campaignID,
				JobID:         jobID,
				SourceTable:   job.SourceTable,
				SourceID:      job.SourceID,
				Content:       content,
				Entity:        *entity,
				OtherEntities: otherEntities,
				Relationships: relationships,
			}

			enrichItems, err := engine.EnrichEntity(bgCtx, provider, input)
			if err != nil {
				log.Printf("Enrichment: error enriching entity %d: %v", entityID, err)
				continue
			}

			if len(enrichItems) > 0 {
				if err := h.db.CreateAnalysisItems(bgCtx, enrichItems); err != nil {
					log.Printf("Enrichment: error inserting items for entity %d: %v",
						entityID, err)
					continue
				}

				if err := h.db.Exec(bgCtx,
					"UPDATE content_analysis_jobs SET enrichment_total = enrichment_total + $1 WHERE id = $2",
					len(enrichItems), jobID,
				); err != nil {
					log.Printf("Enrichment: error updating job counts for entity %d: %v",
						entityID, err)
				}
			}
		}

		// Mark enrichment as completed
		if err := h.db.Exec(bgCtx,
			"UPDATE content_analysis_jobs SET status = 'completed' WHERE id = $1",
			jobID,
		); err != nil {
			log.Printf("Enrichment: error updating job status to completed: %v", err)
		}
	}()

	respondJSON(w, http.StatusOK, triggerEnrichmentResponse{
		Status:      "enriching",
		EntityCount: len(entityIDs),
	})
}

// EnrichmentStream handles GET /api/campaigns/{id}/analysis/jobs/{jobId}/enrichment-stream
// Streams enrichment progress and new items via Server-Sent Events.
func (h *EnrichmentHandler) EnrichmentStream(w http.ResponseWriter, r *http.Request) {
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

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		respondError(w, http.StatusInternalServerError, "Streaming not supported")
		return
	}

	// Track the last item ID to only send new items
	var lastItemID int64
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-timeout:
			fmt.Fprintf(w, "event: timeout\ndata: {\"message\":\"Stream timeout\"}\n\n")
			flusher.Flush()
			return
		case <-ticker.C:
			// Fetch current job status
			currentJob, err := h.db.GetAnalysisJob(r.Context(), jobID)
			if err != nil {
				log.Printf("SSE: error fetching job %d: %v", jobID, err)
				continue
			}

			// Fetch new enrichment items
			enrichItems, err := h.db.ListAnalysisItemsByJob(r.Context(), jobID, "", "enrichment")
			if err != nil {
				log.Printf("SSE: error listing enrichment items for job %d: %v", jobID, err)
				continue
			}

			// Send new items that we have not sent yet
			for _, item := range enrichItems {
				if item.ID > lastItemID {
					itemJSON, err := json.Marshal(item)
					if err != nil {
						log.Printf("SSE: error marshalling item %d: %v", item.ID, err)
						continue
					}
					fmt.Fprintf(w, "event: enrichment_item\ndata: %s\n\n", itemJSON)
					lastItemID = item.ID
				}
			}

			// Send progress heartbeat
			progressJSON, _ := json.Marshal(map[string]interface{}{
				"total":    currentJob.EnrichmentTotal,
				"resolved": currentJob.EnrichmentResolved,
				"status":   currentJob.Status,
			})
			fmt.Fprintf(w, "event: enrichment_progress\ndata: %s\n\n", progressJSON)
			flusher.Flush()

			// Close stream when job is completed (all items were already
			// sent above, so no additional check is needed)
			if currentJob.Status == "completed" {
				fmt.Fprintf(w, "event: enrichment_complete\ndata: {\"status\":\"completed\"}\n\n")
				flusher.Flush()
				return
			}
		}
	}
}

// fetchSourceContent retrieves the text content from the appropriate
// source table and field using parameterized queries.
func (h *EnrichmentHandler) fetchSourceContent(
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
