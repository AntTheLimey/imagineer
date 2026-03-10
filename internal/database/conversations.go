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
