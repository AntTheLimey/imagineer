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
	"github.com/jackc/pgx/v5"
)

// CreateEntityLog inserts a new entity log entry and returns the
// populated record.
func (db *DB) CreateEntityLog(ctx context.Context, entityID, campaignID int64, req models.CreateEntityLogRequest) (*models.EntityLog, error) {
	query := `
		INSERT INTO entity_log
			(entity_id, campaign_id, chapter_id, session_id,
			 content, occurred_at, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, entity_id, campaign_id, chapter_id, session_id,
		          source_table, source_id, content, occurred_at,
		          sort_order, created_at`

	var l models.EntityLog
	err := db.QueryRow(ctx, query,
		entityID, campaignID, req.ChapterID, req.SessionID,
		req.Content, req.OccurredAt, req.SortOrder,
	).Scan(
		&l.ID, &l.EntityID, &l.CampaignID, &l.ChapterID, &l.SessionID,
		&l.SourceTable, &l.SourceID, &l.Content, &l.OccurredAt,
		&l.SortOrder, &l.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create entity log: %w", err)
	}

	return &l, nil
}

// GetEntityLog retrieves a single entity log entry by ID.
func (db *DB) GetEntityLog(ctx context.Context, id int64) (*models.EntityLog, error) {
	query := `
		SELECT id, entity_id, campaign_id, chapter_id, session_id,
		       source_table, source_id, content, occurred_at,
		       sort_order, created_at
		FROM entity_log
		WHERE id = $1`

	var l models.EntityLog
	err := db.QueryRow(ctx, query, id).Scan(
		&l.ID, &l.EntityID, &l.CampaignID, &l.ChapterID, &l.SessionID,
		&l.SourceTable, &l.SourceID, &l.Content, &l.OccurredAt,
		&l.SortOrder, &l.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity log: %w", err)
	}

	return &l, nil
}

// ListEntityLogs retrieves all log entries for an entity, sorted by
// sort_order ascending (NULLs last), then by created_at ascending.
func (db *DB) ListEntityLogs(ctx context.Context, entityID int64) ([]models.EntityLog, error) {
	query := `
		SELECT id, entity_id, campaign_id, chapter_id, session_id,
		       source_table, source_id, content, occurred_at,
		       sort_order, created_at
		FROM entity_log
		WHERE entity_id = $1
		ORDER BY sort_order ASC NULLS LAST, created_at ASC`

	rows, err := db.Query(ctx, query, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to list entity logs: %w", err)
	}
	defer rows.Close()

	return scanEntityLogs(rows)
}

// UpdateEntityLog updates an entity log entry and returns the updated
// record. Uses COALESCE to preserve existing values when fields are nil.
func (db *DB) UpdateEntityLog(ctx context.Context, id int64, req models.UpdateEntityLogRequest) (*models.EntityLog, error) {
	query := `
		UPDATE entity_log
		SET content    = COALESCE($2, content),
		    chapter_id = COALESCE($3, chapter_id),
		    session_id = COALESCE($4, session_id),
		    occurred_at = COALESCE($5, occurred_at),
		    sort_order = COALESCE($6, sort_order)
		WHERE id = $1
		RETURNING id, entity_id, campaign_id, chapter_id, session_id,
		          source_table, source_id, content, occurred_at,
		          sort_order, created_at`

	var l models.EntityLog
	err := db.QueryRow(ctx, query,
		id, req.Content, req.ChapterID, req.SessionID,
		req.OccurredAt, req.SortOrder,
	).Scan(
		&l.ID, &l.EntityID, &l.CampaignID, &l.ChapterID, &l.SessionID,
		&l.SourceTable, &l.SourceID, &l.Content, &l.OccurredAt,
		&l.SortOrder, &l.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update entity log: %w", err)
	}

	return &l, nil
}

// DeleteEntityLog deletes an entity log entry by ID.
func (db *DB) DeleteEntityLog(ctx context.Context, id int64) error {
	return db.Exec(ctx, "DELETE FROM entity_log WHERE id = $1", id)
}

// scanEntityLogs scans multiple entity log rows.
func scanEntityLogs(rows pgx.Rows) ([]models.EntityLog, error) {
	var logs []models.EntityLog
	for rows.Next() {
		var l models.EntityLog
		err := rows.Scan(
			&l.ID, &l.EntityID, &l.CampaignID, &l.ChapterID, &l.SessionID,
			&l.SourceTable, &l.SourceID, &l.Content, &l.OccurredAt,
			&l.SortOrder, &l.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entity log: %w", err)
		}
		logs = append(logs, l)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating entity logs: %w", err)
	}

	return logs, nil
}
