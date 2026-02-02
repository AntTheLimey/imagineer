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
	"github.com/google/uuid"
)

// ListChaptersByCampaign retrieves all chapters for a campaign ordered by sort_order.
func (db *DB) ListChaptersByCampaign(ctx context.Context, campaignID uuid.UUID) ([]models.Chapter, error) {
	query := `
        SELECT id, campaign_id, title, overview, sort_order, created_at, updated_at
        FROM chapters
        WHERE campaign_id = $1
        ORDER BY sort_order ASC, created_at ASC`

	rows, err := db.Query(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to query chapters: %w", err)
	}
	defer rows.Close()

	var chapters []models.Chapter
	for rows.Next() {
		var c models.Chapter
		err := rows.Scan(
			&c.ID, &c.CampaignID, &c.Title, &c.Overview, &c.SortOrder,
			&c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chapter: %w", err)
		}
		chapters = append(chapters, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating chapters: %w", err)
	}

	return chapters, nil
}

// GetChapter retrieves a chapter by ID.
func (db *DB) GetChapter(ctx context.Context, id uuid.UUID) (*models.Chapter, error) {
	query := `
        SELECT id, campaign_id, title, overview, sort_order, created_at, updated_at
        FROM chapters
        WHERE id = $1`

	var c models.Chapter
	err := db.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.CampaignID, &c.Title, &c.Overview, &c.SortOrder,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get chapter: %w", err)
	}

	return &c, nil
}

// CreateChapter creates a new chapter in a campaign.
// If SortOrder is not provided, it will be set to the next available value.
// Uses a transaction with advisory lock to prevent race conditions when
// auto-calculating sort_order for concurrent chapter creation requests.
func (db *DB) CreateChapter(ctx context.Context, campaignID uuid.UUID, req models.CreateChapterRequest) (*models.Chapter, error) {
	id := uuid.New()

	// If sort_order is provided, we can skip the transaction
	if req.SortOrder != nil {
		query := `
            INSERT INTO chapters (id, campaign_id, title, overview, sort_order)
            VALUES ($1, $2, $3, $4, $5)
            RETURNING id, campaign_id, title, overview, sort_order, created_at, updated_at`

		var c models.Chapter
		err := db.QueryRow(ctx, query, id, campaignID, req.Title, req.Overview, *req.SortOrder).Scan(
			&c.ID, &c.CampaignID, &c.Title, &c.Overview, &c.SortOrder,
			&c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create chapter: %w", err)
		}
		return &c, nil
	}

	// Use transaction with advisory lock for auto-calculated sort_order
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // Rollback is a no-op if already committed

	// Acquire advisory lock on campaign_id to serialize sort_order calculation
	// Using the first 8 bytes of the UUID as the lock key
	lockKey := int64(campaignID[0])<<56 | int64(campaignID[1])<<48 |
		int64(campaignID[2])<<40 | int64(campaignID[3])<<32 |
		int64(campaignID[4])<<24 | int64(campaignID[5])<<16 |
		int64(campaignID[6])<<8 | int64(campaignID[7])
	_, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", lockKey)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire advisory lock: %w", err)
	}

	// Get the max sort_order for this campaign and add 1
	var maxSortOrder *int
	err = tx.QueryRow(ctx,
		"SELECT MAX(sort_order) FROM chapters WHERE campaign_id = $1",
		campaignID,
	).Scan(&maxSortOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to determine sort_order: %w", err)
	}

	sortOrder := 0
	if maxSortOrder != nil {
		sortOrder = *maxSortOrder + 1
	}

	query := `
        INSERT INTO chapters (id, campaign_id, title, overview, sort_order)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, campaign_id, title, overview, sort_order, created_at, updated_at`

	var c models.Chapter
	err = tx.QueryRow(ctx, query, id, campaignID, req.Title, req.Overview, sortOrder).Scan(
		&c.ID, &c.CampaignID, &c.Title, &c.Overview, &c.SortOrder,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create chapter: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &c, nil
}

// UpdateChapter updates an existing chapter.
func (db *DB) UpdateChapter(ctx context.Context, id uuid.UUID, req models.UpdateChapterRequest) (*models.Chapter, error) {
	// First get the existing chapter
	existing, err := db.GetChapter(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	title := existing.Title
	if req.Title != nil {
		title = *req.Title
	}

	overview := existing.Overview
	if req.Overview != nil {
		overview = req.Overview
	}

	sortOrder := existing.SortOrder
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	query := `
        UPDATE chapters
        SET title = $2, overview = $3, sort_order = $4, updated_at = NOW()
        WHERE id = $1
        RETURNING id, campaign_id, title, overview, sort_order, created_at, updated_at`

	var c models.Chapter
	err = db.QueryRow(ctx, query, id, title, overview, sortOrder).Scan(
		&c.ID, &c.CampaignID, &c.Title, &c.Overview, &c.SortOrder,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update chapter: %w", err)
	}

	return &c, nil
}

// DeleteChapter deletes a chapter by ID.
func (db *DB) DeleteChapter(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM chapters WHERE id = $1`
	result, err := db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete chapter: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("chapter not found")
	}

	return nil
}
