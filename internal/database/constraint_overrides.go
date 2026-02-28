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

// CreateConstraintOverride records a GM acknowledgement
// that overrides a specific constraint violation.
func (db *DB) CreateConstraintOverride(
	ctx context.Context,
	campaignID int64,
	req models.CreateConstraintOverrideRequest,
) (*models.ConstraintOverride, error) {
	query := `
        INSERT INTO constraint_overrides
            (campaign_id, constraint_type, override_key)
        VALUES ($1, $2, $3)
        ON CONFLICT (campaign_id, constraint_type,
                     override_key)
        DO UPDATE SET acknowledged_at = NOW()
        RETURNING id, campaign_id, constraint_type,
                  override_key, acknowledged_at`

	var o models.ConstraintOverride
	err := db.QueryRow(ctx, query,
		campaignID, req.ConstraintType,
		req.OverrideKey,
	).Scan(
		&o.ID, &o.CampaignID, &o.ConstraintType,
		&o.OverrideKey, &o.AcknowledgedAt,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create constraint override: %w",
			err)
	}

	return &o, nil
}

// ListConstraintOverrides returns all constraint
// overrides for a campaign.
func (db *DB) ListConstraintOverrides(
	ctx context.Context,
	campaignID int64,
) ([]models.ConstraintOverride, error) {
	query := `
        SELECT id, campaign_id, constraint_type,
               override_key, acknowledged_at
        FROM constraint_overrides
        WHERE campaign_id = $1
        ORDER BY acknowledged_at DESC`

	rows, err := db.Query(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to list constraint overrides: %w",
			err)
	}
	defer rows.Close()

	var overrides []models.ConstraintOverride
	for rows.Next() {
		var o models.ConstraintOverride
		err := rows.Scan(
			&o.ID, &o.CampaignID, &o.ConstraintType,
			&o.OverrideKey, &o.AcknowledgedAt,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to scan constraint override: %w",
				err)
		}
		overrides = append(overrides, o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"error iterating constraint overrides: %w",
			err)
	}

	return overrides, nil
}

// HasConstraintOverride checks if a specific constraint
// override exists for a campaign.
func (db *DB) HasConstraintOverride(
	ctx context.Context,
	campaignID int64,
	constraintType string,
	overrideKey string,
) (bool, error) {
	query := `
        SELECT EXISTS (
            SELECT 1 FROM constraint_overrides
            WHERE campaign_id = $1
              AND constraint_type = $2
              AND override_key = $3
        )`
	var exists bool
	err := db.QueryRow(ctx, query,
		campaignID, constraintType, overrideKey,
	).Scan(&exists)
	return exists, err
}
