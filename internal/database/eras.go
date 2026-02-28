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

// CreateEra creates a new era for a campaign.
func (db *DB) CreateEra(
	ctx context.Context,
	campaignID int64,
	req models.CreateEraRequest,
) (*models.Era, error) {
	query := `
        INSERT INTO eras
            (campaign_id, sequence, name, scale,
             description)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, campaign_id, sequence, name,
                  scale, description,
                  created_at, updated_at`

	var e models.Era
	err := db.QueryRow(ctx, query,
		campaignID, req.Sequence, req.Name,
		req.Scale, req.Description,
	).Scan(
		&e.ID, &e.CampaignID, &e.Sequence,
		&e.Name, &e.Scale, &e.Description,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create era: %w", err)
	}

	return &e, nil
}

// ListEras retrieves all eras for a campaign ordered
// by sequence.
func (db *DB) ListEras(
	ctx context.Context,
	campaignID int64,
) ([]models.Era, error) {
	query := `
        SELECT id, campaign_id, sequence, name,
               scale, description,
               created_at, updated_at
        FROM eras
        WHERE campaign_id = $1
        ORDER BY sequence`

	rows, err := db.Query(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to list eras: %w", err)
	}
	defer rows.Close()

	var eras []models.Era
	for rows.Next() {
		var e models.Era
		err := rows.Scan(
			&e.ID, &e.CampaignID, &e.Sequence,
			&e.Name, &e.Scale, &e.Description,
			&e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to scan era: %w", err)
		}
		eras = append(eras, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"error iterating eras: %w", err)
	}

	return eras, nil
}

// GetCurrentEra returns the era with the highest
// sequence number for the campaign.
func (db *DB) GetCurrentEra(
	ctx context.Context,
	campaignID int64,
) (*models.Era, error) {
	query := `
        SELECT id, campaign_id, sequence, name,
               scale, description,
               created_at, updated_at
        FROM eras
        WHERE campaign_id = $1
        ORDER BY sequence DESC
        LIMIT 1`

	var e models.Era
	err := db.QueryRow(ctx, query, campaignID).Scan(
		&e.ID, &e.CampaignID, &e.Sequence,
		&e.Name, &e.Scale, &e.Description,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get current era: %w", err)
	}

	return &e, nil
}

// UpdateEra updates an existing era.
func (db *DB) UpdateEra(
	ctx context.Context,
	id int64,
	req models.UpdateEraRequest,
) (*models.Era, error) {
	query := `
        UPDATE eras SET
            sequence    = COALESCE($2, sequence),
            name        = COALESCE($3, name),
            scale       = COALESCE($4, scale),
            description = COALESCE($5, description)
        WHERE id = $1
        RETURNING id, campaign_id, sequence, name,
                  scale, description,
                  created_at, updated_at`

	var e models.Era
	err := db.QueryRow(ctx, query,
		id, req.Sequence, req.Name,
		req.Scale, req.Description,
	).Scan(
		&e.ID, &e.CampaignID, &e.Sequence,
		&e.Name, &e.Scale, &e.Description,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to update era: %w", err)
	}

	return &e, nil
}
