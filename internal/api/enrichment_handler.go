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
	content, err := fetchSourceContent(r.Context(), h.db, campaignID, job.SourceTable, job.SourceID, job.SourceField)
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

	// Build the pipeline with analysis and enrichment stages.
	pipeline := buildDefaultPipeline(h.db)

	// Pre-load entity objects from the accepted entity IDs.
	entities := make([]models.Entity, 0, len(entityIDs))
	for _, eid := range entityIDs {
		entity, err := h.db.GetEntity(r.Context(), eid)
		if err != nil {
			log.Printf("Enrichment: error fetching entity %d: %v", eid, err)
			continue
		}
		entities = append(entities, *entity)
	}

	// Look up the campaign to get its game system code for RAG context.
	campaign, campaignErr := h.db.GetCampaign(r.Context(), campaignID)
	var gameSystemCode string
	if campaignErr != nil {
		log.Printf("Enrichment: failed to get campaign %d: %v", campaignID, campaignErr)
	} else if campaign.System != nil {
		gameSystemCode = campaign.System.Code
	}

	// Spawn background goroutine for enrichment processing
	go func() {
		bgCtx, bgCancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer bgCancel()

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

		// Build RAG context for the enrichment pipeline.
		ctxBuilder := enrichment.NewContextBuilder(h.db, "")
		ragCtx, ragErr := ctxBuilder.BuildContext(bgCtx, campaignID, content, gameSystemCode, entities)
		if ragErr != nil {
			log.Printf("Enrichment: failed to build RAG context for job %d: %v",
				jobID, ragErr)
		}

		var gameSystemID *int64
		if campaign != nil {
			gameSystemID = campaign.SystemID
		}

		// Strip GM-only content from entities before passing to pipeline.
		for i := range entities {
			entities[i].GMNotes = nil
		}

		input := enrichment.PipelineInput{
			CampaignID:   campaignID,
			JobID:        jobID,
			SourceTable:  job.SourceTable,
			SourceID:     job.SourceID,
			SourceScope:  enrichment.ScopeFromSourceTable(job.SourceTable),
			Content:      content,
			Entities:     entities,
			GameSystemID: gameSystemID,
			Context:      ragCtx,
		}

		enrichItems, err := pipeline.Run(bgCtx, provider, input)
		if err != nil {
			log.Printf("Enrichment: pipeline run failed for job %d: %v", jobID, err)
			if dbErr := h.db.Exec(bgCtx,
				"UPDATE content_analysis_jobs SET status = 'failed' WHERE id = $1",
				jobID,
			); dbErr != nil {
				log.Printf("Enrichment: error setting job %d to failed: %v", jobID, dbErr)
			}
			return
		}

		if len(enrichItems) > 0 {
			if err := h.db.CreateAnalysisItems(bgCtx, enrichItems); err != nil {
				log.Printf("Enrichment: error inserting items for job %d: %v",
					jobID, err)
			} else {
				_ = h.db.Exec(bgCtx,
					"UPDATE content_analysis_jobs SET enrichment_total = enrichment_total + $1 WHERE id = $2",
					len(enrichItems), jobID)
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
