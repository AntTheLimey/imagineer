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
	"errors"
	"fmt"

	"github.com/antonypegg/imagineer/internal/models"
	"github.com/jackc/pgx/v5"
)

// validSourceTables defines the allowed source_table values for drafts.
var validSourceTables = map[string]bool{
	"entities": true,
	"chapters": true,
	"sessions": true,
}

// draftColumns is the standard column list for draft queries.
const draftColumns = `id, campaign_id, user_id, source_table, source_id,
	is_new, draft_data, server_version, created_at, updated_at`

// validateSourceTable checks that the given source table name is one of
// the permitted values.
func validateSourceTable(sourceTable string) error {
	if !validSourceTables[sourceTable] {
		return fmt.Errorf("invalid source_table %q: must be one of entities, chapters, sessions", sourceTable)
	}
	return nil
}

// scanDraft scans a single row into a models.Draft.
func scanDraft(row pgx.Row) (*models.Draft, error) {
	var d models.Draft
	err := row.Scan(
		&d.ID, &d.CampaignID, &d.UserID, &d.SourceTable, &d.SourceID,
		&d.IsNew, &d.DraftData, &d.ServerVersion,
		&d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// SaveDraft upserts a draft using the unique index on
// (user_id, source_table, source_id, campaign_id). On conflict it
// updates draft_data, server_version, is_new, and updated_at.
func (db *DB) SaveDraft(ctx context.Context, campaignID, userID int64, req models.SaveDraftRequest) (*models.Draft, error) {
	if err := validateSourceTable(req.SourceTable); err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		INSERT INTO drafts
			(campaign_id, user_id, source_table, source_id,
			 is_new, draft_data, server_version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, source_table, source_id, campaign_id)
		DO UPDATE SET
			draft_data     = EXCLUDED.draft_data,
			server_version = EXCLUDED.server_version,
			is_new         = EXCLUDED.is_new,
			updated_at     = NOW()
		RETURNING %s`, draftColumns)

	d, err := scanDraft(db.QueryRow(ctx, query,
		campaignID, userID, req.SourceTable, req.SourceID,
		req.IsNew, req.DraftData, req.ServerVersion,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to save draft: %w", err)
	}

	return d, nil
}

// GetDraft retrieves a draft by its composite key
// (user_id, source_table, source_id, campaign_id).
// Returns pgx.ErrNoRows (unwrapped) when no draft is found so
// callers can distinguish 404 from 500 with errors.Is().
func (db *DB) GetDraft(ctx context.Context, userID int64, sourceTable string, sourceID, campaignID int64) (*models.Draft, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM drafts
		WHERE user_id = $1
		  AND source_table = $2
		  AND source_id = $3
		  AND campaign_id = $4`, draftColumns)

	d, err := scanDraft(db.QueryRow(ctx, query,
		userID, sourceTable, sourceID, campaignID,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get draft: %w", err)
	}

	return d, nil
}

// DeleteDraft deletes a draft by its composite key
// (user_id, source_table, source_id, campaign_id).
// Returns pgx.ErrNoRows when no draft matches the given key.
func (db *DB) DeleteDraft(ctx context.Context, userID int64, sourceTable string, sourceID, campaignID int64) error {
	result, err := db.Pool.Exec(ctx, `
		DELETE FROM drafts
		WHERE user_id = $1
		  AND source_table = $2
		  AND source_id = $3
		  AND campaign_id = $4`,
		userID, sourceTable, sourceID, campaignID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete draft: %w", err)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// ListDraftIndicators returns lightweight draft indicators for a
// campaign and user. When sourceTable is non-empty, results are
// filtered to that source table only. Returns an empty slice (not
// nil) when no drafts match.
func (db *DB) ListDraftIndicators(ctx context.Context, campaignID, userID int64, sourceTable string) ([]models.DraftIndicator, error) {
	var rows pgx.Rows
	var err error

	if sourceTable != "" {
		query := `
			SELECT source_table, source_id, is_new, updated_at
			FROM drafts
			WHERE campaign_id = $1
			  AND user_id = $2
			  AND source_table = $3`
		rows, err = db.Query(ctx, query, campaignID, userID, sourceTable)
	} else {
		query := `
			SELECT source_table, source_id, is_new, updated_at
			FROM drafts
			WHERE campaign_id = $1
			  AND user_id = $2`
		rows, err = db.Query(ctx, query, campaignID, userID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list draft indicators: %w", err)
	}
	defer rows.Close()

	indicators, err := scanDraftIndicators(rows)
	if err != nil {
		return nil, err
	}

	if indicators == nil {
		indicators = []models.DraftIndicator{}
	}

	return indicators, nil
}

// scanDraftIndicators scans multiple draft indicator rows.
func scanDraftIndicators(rows pgx.Rows) ([]models.DraftIndicator, error) {
	var indicators []models.DraftIndicator
	for rows.Next() {
		var d models.DraftIndicator
		err := rows.Scan(
			&d.SourceTable, &d.SourceID, &d.IsNew, &d.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan draft indicator: %w", err)
		}
		indicators = append(indicators, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating draft indicators: %w", err)
	}

	return indicators, nil
}
