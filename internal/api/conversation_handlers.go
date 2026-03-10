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
	"github.com/antonypegg/imagineer/internal/conversation"
	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/antonypegg/imagineer/internal/models"
	"github.com/jackc/pgx/v5"
)

// validScopeTypes defines the permitted scope_type values
// for conversation creation.
var validScopeTypes = map[models.ScopeType]bool{
	models.ScopeTypeEntity:   true,
	models.ScopeTypeChapter:  true,
	models.ScopeTypeSession:  true,
	models.ScopeTypeScene:    true,
	models.ScopeTypeCampaign: true,
}

// ConversationHandler handles conversation API requests.
type ConversationHandler struct {
	db           *database.DB
	orchestrator *conversation.Orchestrator
}

// NewConversationHandler creates a new ConversationHandler.
func NewConversationHandler(
	db *database.DB,
	orch *conversation.Orchestrator,
) *ConversationHandler {
	return &ConversationHandler{
		db:           db,
		orchestrator: orch,
	}
}

// Create handles POST /api/campaigns/{id}/conversations.
// It creates a new conversation or returns an existing one
// for the given scope.
func (h *ConversationHandler) Create(
	w http.ResponseWriter, r *http.Request,
) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest,
			"Invalid campaign ID")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized,
			"Authentication required")
		return
	}

	if err := h.db.VerifyCampaignOwnership(
		r.Context(), campaignID, userID,
	); err != nil {
		respondError(w, http.StatusNotFound,
			"Campaign not found")
		return
	}

	var req models.CreateConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest,
			"Invalid request body")
		return
	}

	if req.ScopeType == "" {
		respondError(w, http.StatusBadRequest,
			"scope_type is required")
		return
	}

	if !validScopeTypes[req.ScopeType] {
		respondError(w, http.StatusBadRequest,
			"Invalid scope_type")
		return
	}

	conv, wasCreated, err := h.db.GetOrCreateConversation(
		r.Context(), campaignID, req.ScopeType, req.ScopeID,
	)
	if err != nil {
		log.Printf("Error creating conversation: %v", err)
		respondError(w, http.StatusInternalServerError,
			"Failed to create conversation")
		return
	}

	status := http.StatusOK
	if wasCreated {
		status = http.StatusCreated
	}

	respondJSON(w, status, conv)
}

// Get handles GET /api/campaigns/{id}/conversations/{conversationId}.
// It returns a single conversation.
func (h *ConversationHandler) Get(
	w http.ResponseWriter, r *http.Request,
) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest,
			"Invalid campaign ID")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized,
			"Authentication required")
		return
	}

	if err := h.db.VerifyCampaignOwnership(
		r.Context(), campaignID, userID,
	); err != nil {
		respondError(w, http.StatusNotFound,
			"Campaign not found")
		return
	}

	conversationID, err := parseInt64(r, "conversationId")
	if err != nil {
		respondError(w, http.StatusBadRequest,
			"Invalid conversation ID")
		return
	}

	conv, err := h.db.GetConversation(
		r.Context(), conversationID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(w, http.StatusNotFound,
				"Conversation not found")
			return
		}
		log.Printf("Error getting conversation: %v", err)
		respondError(w, http.StatusInternalServerError,
			"Failed to get conversation")
		return
	}

	if conv.CampaignID != campaignID {
		respondError(w, http.StatusNotFound,
			"Conversation not found")
		return
	}

	respondJSON(w, http.StatusOK, conv)
}

// List handles GET /api/campaigns/{id}/conversations.
// It returns conversations for a campaign, optionally
// filtered by scope_type and scope_id.
func (h *ConversationHandler) List(
	w http.ResponseWriter, r *http.Request,
) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest,
			"Invalid campaign ID")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized,
			"Authentication required")
		return
	}

	if err := h.db.VerifyCampaignOwnership(
		r.Context(), campaignID, userID,
	); err != nil {
		respondError(w, http.StatusNotFound,
			"Campaign not found")
		return
	}

	scopeType := models.ScopeType(
		r.URL.Query().Get("scope_type"),
	)

	var scopeID int64
	if raw := r.URL.Query().Get("scope_id"); raw != "" {
		scopeID, err = strconv.ParseInt(raw, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest,
				"Invalid scope_id")
			return
		}
	}

	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err == nil && parsed > 0 {
			limit = parsed
		}
	}

	items, err := h.db.ListConversations(
		r.Context(), campaignID,
		scopeType, scopeID, limit,
	)
	if err != nil {
		log.Printf("Error listing conversations: %v", err)
		respondError(w, http.StatusInternalServerError,
			"Failed to list conversations")
		return
	}

	if items == nil {
		items = []models.ConversationListItem{}
	}

	respondJSON(w, http.StatusOK, items)
}

// SendMessage handles POST /api/campaigns/{id}/conversations/{conversationId}/messages.
// It streams the assistant response as Server-Sent Events.
func (h *ConversationHandler) SendMessage(
	w http.ResponseWriter, r *http.Request,
) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest,
			"Invalid campaign ID")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized,
			"Authentication required")
		return
	}

	if err := h.db.VerifyCampaignOwnership(
		r.Context(), campaignID, userID,
	); err != nil {
		respondError(w, http.StatusNotFound,
			"Campaign not found")
		return
	}

	conversationID, err := parseInt64(r, "conversationId")
	if err != nil {
		respondError(w, http.StatusBadRequest,
			"Invalid conversation ID")
		return
	}

	var req models.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest,
			"Invalid request body")
		return
	}

	if req.Content == "" {
		respondError(w, http.StatusBadRequest,
			"Content is required")
		return
	}

	// Set SSE headers before streaming begins.
	sse := conversation.NewSSEWriter(w)
	sse.SetHeaders()

	eventCh, err := h.orchestrator.HandleMessage(
		r.Context(), conversationID, campaignID,
		userID, req.Content,
	)
	if err != nil {
		// Headers are already set for SSE, so send
		// the error as an SSE event rather than a
		// JSON error response.
		_ = sse.WriteError(err.Error())
		return
	}

	var inputTokens, outputTokens int

	for ev := range eventCh {
		switch ev.Type {
		case llm.EventTextDelta:
			if err := sse.WriteTextDelta(ev.Text); err != nil {
				log.Printf(
					"SSE write error: %v", err)
				return
			}

		case llm.EventToolUse:
			if err := sse.WriteToolUse(
				ev.ToolName, ev.ToolInput,
			); err != nil {
				log.Printf(
					"SSE write error: %v", err)
				return
			}

		case llm.EventUsage:
			if ev.Usage != nil {
				inputTokens += ev.Usage.InputTokens
				outputTokens += ev.Usage.OutputTokens
			}

		case llm.EventError:
			_ = sse.WriteError(ev.Error.Error())
			return

		case llm.EventDone:
			_ = sse.WriteDone(
				0, inputTokens, outputTokens)
			return
		}
	}

	// Channel closed without a Done event; send done
	// so the client knows the stream ended.
	_ = sse.WriteDone(0, inputTokens, outputTokens)
}
