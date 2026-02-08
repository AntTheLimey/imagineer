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
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/antonypegg/imagineer/internal/auth"
	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/models"
	"github.com/go-chi/chi/v5"
)

// EntityDetectionHandler handles entity detection requests.
type EntityDetectionHandler struct {
	db *database.DB
}

// NewEntityDetectionHandler creates a new EntityDetectionHandler.
func NewEntityDetectionHandler(db *database.DB) *EntityDetectionHandler {
	return &EntityDetectionHandler{db: db}
}

// DetectEntitiesRequest is the request body for entity detection.
type DetectEntitiesRequest struct {
	TextSegments     []string `json:"textSegments"`
	ExcludeEntityIDs []string `json:"excludeEntityIds"`
}

// EntitySuggestion represents a detected entity match.
type EntitySuggestion struct {
	Text       string         `json:"text"`
	Entity     *models.Entity `json:"entity"`
	Similarity float64        `json:"similarity"`
}

// DetectEntitiesResponse is the response for entity detection.
type DetectEntitiesResponse struct {
	Suggestions []EntitySuggestion `json:"suggestions"`
	Configured  bool               `json:"configured"`
}

// DetectEntities handles POST /api/campaigns/:id/chapters/detect-entities
// Detects entities mentioned in chapter content using text-based search.
func (h *EntityDetectionHandler) DetectEntities(w http.ResponseWriter, r *http.Request) {
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

	// Parse request body
	var req DetectEntitiesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Fetch user's embedding settings to determine if semantic search is configured
	settings, err := h.db.GetUserSettings(r.Context(), userID)
	if err != nil {
		log.Printf("Error fetching user settings: %v", err)
		// Continue with fallback - settings may not exist yet
		settings = nil
	}

	// Check if embedding is configured.
	// Ollama does not require an API key -- it runs locally.
	embeddingConfigured := settings != nil &&
		settings.EmbeddingService != nil

	if embeddingConfigured {
		svc := *settings.EmbeddingService
		if svc == models.LLMServiceOllama {
			// Ollama needs no API key; consider it configured if
			// the vectorizer extension is available.
			embeddingConfigured = h.db.IsVectorizationAvailable(r.Context())
		} else {
			// Cloud services require an API key.
			embeddingConfigured = settings.EmbeddingAPIKey != nil &&
				*settings.EmbeddingAPIKey != ""
		}
	}

	// Build exclusion set for filtering
	excludeSet := make(map[int64]bool)
	for _, idStr := range req.ExcludeEntityIDs {
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
			excludeSet[id] = true
		}
	}

	var suggestions []EntitySuggestion

	// Try vector-based detection first
	if h.db.IsVectorizationAvailable(r.Context()) {
		suggestions, err = h.detectEntitiesFromVectors(
			r.Context(), campaignID, req.TextSegments, excludeSet,
		)
		if err != nil {
			log.Printf("Vector search failed, falling back to text: %v", err)
			suggestions = nil
		}
	}

	// Fall back to text-based detection
	if suggestions == nil {
		suggestions, err = h.detectEntitiesFromText(
			r.Context(), campaignID, req.TextSegments, excludeSet,
		)
		if err != nil {
			log.Printf("Error detecting entities: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to detect entities")
			return
		}
	}

	response := DetectEntitiesResponse{
		Suggestions: suggestions,
		Configured:  embeddingConfigured,
	}

	respondJSON(w, http.StatusOK, response)
}

// detectEntitiesFromVectors uses the pgedge_vectorizer to find entities
// semantically similar to the provided text segments.
func (h *EntityDetectionHandler) detectEntitiesFromVectors(
	ctx context.Context,
	campaignID int64,
	textSegments []string,
	excludeSet map[int64]bool,
) ([]EntitySuggestion, error) {
	// Concatenate segments into a single search query (capped at 2000 chars)
	combined := strings.Join(textSegments, " ")
	if len(combined) > 2000 {
		combined = combined[:2000]
	}
	if combined == "" {
		return []EntitySuggestion{}, nil
	}

	// Search entity chunks via the SQL search function.
	// We request more results than needed to allow for filtering.
	results, err := h.db.SearchCampaignContent(ctx, campaignID, combined, 50)
	if err != nil {
		return nil, err
	}

	// Build entity lookup for deduplication
	seen := make(map[int64]bool)
	var suggestions []EntitySuggestion

	for _, r := range results {
		// Only include entity results (skip chapters, sessions, etc.)
		if r.SourceTable != "entities" {
			continue
		}

		if seen[r.SourceID] || excludeSet[r.SourceID] {
			continue
		}

		// Fetch the full entity
		entity, err := h.db.GetEntity(ctx, r.SourceID)
		if err != nil {
			log.Printf("Error fetching entity %d: %v", r.SourceID, err)
			continue
		}

		suggestions = append(suggestions, EntitySuggestion{
			Text:       r.SourceName,
			Entity:     entity,
			Similarity: r.CombinedScore,
		})
		seen[r.SourceID] = true

		// Cap at 20 suggestions
		if len(suggestions) >= 20 {
			break
		}
	}

	return suggestions, nil
}

// detectEntitiesFromText performs text-based entity detection using ILIKE pattern matching.
// It extracts potential entity names from text segments and searches for matches.
func (h *EntityDetectionHandler) detectEntitiesFromText(
	ctx context.Context,
	campaignID int64,
	textSegments []string,
	excludeSet map[int64]bool,
) ([]EntitySuggestion, error) {
	// First, get all entities for this campaign to build a lookup
	entities, err := h.db.ListEntitiesByCampaign(ctx, campaignID)
	if err != nil {
		return nil, err
	}

	// Build entity name lookup map (lowercase for case-insensitive matching)
	entityByName := make(map[string]*models.Entity)
	for i := range entities {
		normalizedName := strings.ToLower(strings.TrimSpace(entities[i].Name))
		entityByName[normalizedName] = &entities[i]
	}

	// Track unique suggestions to avoid duplicates
	seen := make(map[int64]bool)
	var suggestions []EntitySuggestion

	// Search for entity names in each text segment
	for _, text := range textSegments {
		if text == "" {
			continue
		}

		normalizedText := strings.ToLower(text)

		// Check each entity name against the text
		for name, entity := range entityByName {
			// Skip if already suggested or excluded
			if seen[entity.ID] || excludeSet[entity.ID] {
				continue
			}

			// Check if entity name appears in the text
			if strings.Contains(normalizedText, name) {
				// Calculate a basic similarity score based on how well the name matches
				similarity := calculateSimilarity(name, text)

				// Make a copy of the entity to avoid pointer issues
				entityCopy := *entity
				suggestions = append(suggestions, EntitySuggestion{
					Text:       entity.Name,
					Entity:     &entityCopy,
					Similarity: similarity,
				})
				seen[entity.ID] = true
			}
		}
	}

	// Sort suggestions by similarity (highest first)
	sortSuggestionsBySimilarity(suggestions)

	return suggestions, nil
}

// calculateSimilarity computes a basic similarity score between an entity name and text.
// Returns a value between 0 and 1, where 1 is the best match.
func calculateSimilarity(entityName, text string) float64 {
	normalizedName := strings.ToLower(strings.TrimSpace(entityName))
	normalizedText := strings.ToLower(text)

	// Base score for finding the name in text
	baseSimilarity := 0.5

	// Check for exact word boundary match (higher score)
	words := strings.Fields(normalizedText)
	for _, word := range words {
		if word == normalizedName {
			return 0.95 // Near-exact match
		}
	}

	// Check if it's part of a multi-word match
	if strings.Contains(normalizedText, normalizedName) {
		// Calculate how much of the text is the entity name
		ratio := float64(len(normalizedName)) / float64(len(normalizedText))
		// Higher ratio = better match, scale to 0.5-0.9 range
		return baseSimilarity + (ratio * 0.4)
	}

	return baseSimilarity
}

// sortSuggestionsBySimilarity sorts suggestions by similarity in descending order.
func sortSuggestionsBySimilarity(suggestions []EntitySuggestion) {
	// Simple bubble sort for small arrays (typically < 100 items)
	n := len(suggestions)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if suggestions[j].Similarity < suggestions[j+1].Similarity {
				suggestions[j], suggestions[j+1] = suggestions[j+1], suggestions[j]
			}
		}
	}
}
