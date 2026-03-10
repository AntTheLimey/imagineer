/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package database

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/antonypegg/imagineer/internal/models"
)

// GetOrCreateConversation calls the get_or_create_conversation()
// stored function. Returns the conversation and whether it was
// newly created.
func (db *DB) GetOrCreateConversation(
	ctx context.Context,
	campaignID int64,
	scopeType models.ScopeType,
	scopeID int64,
) (*models.Conversation, bool, error) {
	var conv models.Conversation
	var wasCreated bool
	err := db.QueryRow(ctx,
		`SELECT id, campaign_id, scope_type,
			scope_id, title, summary,
			summary_tokens, created_at,
			updated_at, was_created
		 FROM get_or_create_conversation($1, $2, $3)`,
		campaignID,
		string(scopeType),
		scopeID,
	).Scan(
		&conv.ID, &conv.CampaignID,
		&conv.ScopeType, &conv.ScopeID,
		&conv.Title, &conv.Summary,
		&conv.SummaryTokens,
		&conv.CreatedAt, &conv.UpdatedAt,
		&wasCreated,
	)
	if err != nil {
		return nil, false, fmt.Errorf(
			"get_or_create_conversation: %w", err)
	}
	return &conv, wasCreated, nil
}

// GetConversation retrieves a conversation by ID.
func (db *DB) GetConversation(
	ctx context.Context,
	id int64,
) (*models.Conversation, error) {
	var conv models.Conversation
	err := db.QueryRow(ctx,
		`SELECT id, campaign_id, scope_type,
			scope_id, title, summary,
			summary_tokens,
			created_at, updated_at
		 FROM conversations
		 WHERE id = $1`,
		id,
	).Scan(
		&conv.ID, &conv.CampaignID,
		&conv.ScopeType, &conv.ScopeID,
		&conv.Title, &conv.Summary,
		&conv.SummaryTokens,
		&conv.CreatedAt, &conv.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get conversation %d: %w",
			id, err)
	}
	return &conv, nil
}

// ListConversations queries the conversation_list view,
// returning conversations with message preview and counts.
func (db *DB) ListConversations(
	ctx context.Context,
	campaignID int64,
	scopeType models.ScopeType,
	scopeID int64,
	limit int,
) ([]models.ConversationListItem, error) {
	if limit <= 0 {
		limit = 20
	}

	query := `SELECT id, campaign_id,
			scope_type, scope_id, title,
			summary_tokens, created_at,
			updated_at, last_message_role,
			last_message_preview,
			last_message_at, total_messages,
			uncompacted_messages
		 FROM conversation_list
		 WHERE campaign_id = $1`
	args := []any{campaignID}
	argN := 2

	if scopeType != "" {
		query += fmt.Sprintf(
			" AND scope_type = $%d", argN)
		args = append(args, scopeType)
		argN++
	}
	if scopeID > 0 {
		query += fmt.Sprintf(
			" AND scope_id = $%d", argN)
		args = append(args, scopeID)
		argN++
	}

	query += fmt.Sprintf(
		" ORDER BY updated_at DESC LIMIT $%d",
		argN)
	args = append(args, limit)

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to list conversations: %w", err)
	}
	defer rows.Close()

	var items []models.ConversationListItem
	for rows.Next() {
		var c models.ConversationListItem
		err := rows.Scan(
			&c.ID, &c.CampaignID,
			&c.ScopeType, &c.ScopeID,
			&c.Title, &c.SummaryTokens,
			&c.CreatedAt, &c.UpdatedAt,
			&c.LastMessageRole,
			&c.LastMessagePreview,
			&c.LastMessageAt,
			&c.TotalMessages,
			&c.UncompactedMessages,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to scan conversation: %w", err)
		}
		items = append(items, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"conversation rows error: %w", err)
	}
	return items, nil
}

// CreateMessage inserts a message into a conversation.
func (db *DB) CreateMessage(
	ctx context.Context,
	conversationID int64,
	role models.MessageRole,
	content string,
	toolName *string,
	toolUseID *string,
	toolInput json.RawMessage,
	toolResult json.RawMessage,
	tokens *int,
) (*models.Message, error) {
	var msg models.Message
	err := db.QueryRow(ctx,
		`INSERT INTO messages
			(conversation_id, role, content,
			 tool_name, tool_use_id,
			 tool_input, tool_result, tokens)
		 VALUES ($1, $2, $3, $4, $5, $6,
				 $7, $8)
		 RETURNING id, conversation_id, role,
			content, tool_name, tool_use_id,
			tool_input, tool_result, tokens,
			compacted, created_at`,
		conversationID, string(role), content,
		toolName, toolUseID,
		toolInput, toolResult, tokens,
	).Scan(
		&msg.ID, &msg.ConversationID,
		&msg.Role, &msg.Content,
		&msg.ToolName, &msg.ToolUseID,
		&msg.ToolInput, &msg.ToolResult,
		&msg.Tokens,
		&msg.Compacted, &msg.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create message: %w", err)
	}
	return &msg, nil
}

// ListMessages returns messages for a conversation
// with pagination. Ordered by created_at ascending.
func (db *DB) ListMessages(
	ctx context.Context,
	conversationID int64,
	limit int,
	beforeID int64,
) ([]models.Message, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `SELECT id, conversation_id, role,
			content, tool_name, tool_use_id,
			tool_input, tool_result, tokens,
			compacted, created_at
		 FROM messages
		 WHERE conversation_id = $1`
	args := []any{conversationID}
	argN := 2

	if beforeID > 0 {
		query += fmt.Sprintf(" AND id < $%d", argN)
		args = append(args, beforeID)
		argN++
	}

	query += fmt.Sprintf(
		" ORDER BY created_at ASC LIMIT $%d",
		argN)
	args = append(args, limit)

	return db.scanMessages(ctx, query, args...)
}

// AssembleConversationContext calls the
// assemble_conversation_context() stored function.
func (db *DB) AssembleConversationContext(
	ctx context.Context,
	conversationID int64,
) (*models.ConversationContext, error) {
	rows, err := db.Query(ctx,
		`SELECT conversation_id, campaign_id,
			scope_type, scope_id, summary,
			summary_tokens,
			uncompacted_token_estimate,
			message_id, role, content,
			tool_name, tool_use_id,
			tool_input, tool_result,
			message_tokens, message_created_at
		 FROM assemble_conversation_context($1)`,
		conversationID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"assemble_conversation_context: %w",
			err)
	}
	defer rows.Close()

	var cc models.ConversationContext
	first := true
	for rows.Next() {
		var m models.Message
		var convID, campaignID, scopeID int64
		var scopeType string
		var summary *string
		var summaryTokens, tokenEstimate int

		err := rows.Scan(
			&convID, &campaignID,
			&scopeType, &scopeID,
			&summary, &summaryTokens,
			&tokenEstimate,
			&m.ID, &m.Role, &m.Content,
			&m.ToolName, &m.ToolUseID,
			&m.ToolInput, &m.ToolResult,
			&m.Tokens, &m.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"scan context row: %w", err)
		}

		if first {
			cc.Conversation = models.Conversation{
				ID:            convID,
				CampaignID:    campaignID,
				ScopeType:     models.ScopeType(scopeType),
				ScopeID:       scopeID,
				Summary:       summary,
				SummaryTokens: summaryTokens,
			}
			cc.TokenEstimate = tokenEstimate
			first = false
		}

		m.ConversationID = convID
		cc.Messages = append(cc.Messages, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"context rows error: %w", err)
	}
	if first {
		return nil, fmt.Errorf(
			"conversation %d not found",
			conversationID)
	}
	return &cc, nil
}

// CompactMessages calls the compact_messages()
// stored function.
func (db *DB) CompactMessages(
	ctx context.Context,
	conversationID int64,
	messageIDs []int64,
	newSummary string,
	summaryTokens int,
) (int, error) {
	var compactedCount int
	err := db.QueryRow(ctx,
		`SELECT compact_messages($1, $2, $3, $4)`,
		conversationID,
		messageIDs,
		newSummary,
		summaryTokens,
	).Scan(&compactedCount)
	if err != nil {
		return 0, fmt.Errorf(
			"compact_messages: %w", err)
	}
	return compactedCount, nil
}

// scanMessages is a shared helper for scanning message rows.
func (db *DB) scanMessages(
	ctx context.Context,
	query string,
	args ...any,
) ([]models.Message, error) {
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to query messages: %w", err)
	}
	defer rows.Close()

	var msgs []models.Message
	for rows.Next() {
		var m models.Message
		err := rows.Scan(
			&m.ID, &m.ConversationID,
			&m.Role, &m.Content,
			&m.ToolName, &m.ToolUseID,
			&m.ToolInput, &m.ToolResult,
			&m.Tokens,
			&m.Compacted, &m.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to scan message: %w", err)
		}
		msgs = append(msgs, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"message rows error: %w", err)
	}
	return msgs, nil
}
