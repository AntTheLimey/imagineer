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

// UpdateEra updates an existing era, scoped to the
// given campaign to prevent cross-campaign modification.
func (db *DB) UpdateEra(
	ctx context.Context,
	id int64,
	campaignID int64,
	req models.UpdateEraRequest,
) (*models.Era, error) {
	query := `
        UPDATE eras SET
            sequence    = COALESCE($2, sequence),
            name        = COALESCE($3, name),
            scale       = COALESCE($4, scale),
            description = COALESCE($5, description)
        WHERE id = $1 AND campaign_id = $6
        RETURNING id, campaign_id, sequence, name,
                  scale, description,
                  created_at, updated_at`

	var e models.Era
	err := db.QueryRow(ctx, query,
		id, req.Sequence, req.Name,
		req.Scale, req.Description, campaignID,
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

// DeleteEra deletes an era by ID, scoped to the given
// campaign. Returns an error if the era is still
// referenced by relationships or relationship_archive.
func (db *DB) DeleteEra(
	ctx context.Context,
	eraID int64,
	campaignID int64,
) error {
	// Check for references in relationships
	var refCount int
	checkQuery := `
        SELECT COUNT(*) FROM (
            SELECT 1 FROM relationships
            WHERE era_id = $1
            UNION ALL
            SELECT 1 FROM relationship_archive
            WHERE era_id = $1
        ) refs`
	err := db.QueryRow(ctx, checkQuery, eraID).Scan(
		&refCount)
	if err != nil {
		return fmt.Errorf(
			"failed to check era references: %w", err)
	}
	if refCount > 0 {
		return fmt.Errorf(
			"era is still referenced by %d "+
				"relationship(s)", refCount)
	}

	query := `
        DELETE FROM eras
        WHERE id = $1 AND campaign_id = $2`
	result, err := db.Pool.Exec(ctx, query,
		eraID, campaignID)
	if err != nil {
		return fmt.Errorf(
			"failed to delete era: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("era not found")
	}

	return nil
}
