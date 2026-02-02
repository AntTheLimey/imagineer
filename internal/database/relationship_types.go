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

// ListRelationshipTypes returns all relationship types available for a campaign.
// This includes system defaults (campaign_id IS NULL) and campaign-specific types.
// Results are sorted by name.
func (db *DB) ListRelationshipTypes(ctx context.Context, campaignID uuid.UUID) ([]models.RelationshipType, error) {
	query := `
		SELECT id, campaign_id, name, inverse_name, is_symmetric,
		       display_label, inverse_display_label, description,
		       created_at, updated_at
		FROM relationship_types
		WHERE campaign_id IS NULL OR campaign_id = $1
		ORDER BY name`

	rows, err := db.Query(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to query relationship types: %w", err)
	}
	defer rows.Close()

	var types []models.RelationshipType
	for rows.Next() {
		var rt models.RelationshipType
		err := rows.Scan(
			&rt.ID, &rt.CampaignID, &rt.Name, &rt.InverseName, &rt.IsSymmetric,
			&rt.DisplayLabel, &rt.InverseDisplayLabel, &rt.Description,
			&rt.CreatedAt, &rt.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan relationship type: %w", err)
		}
		types = append(types, rt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating relationship types: %w", err)
	}

	return types, nil
}

// GetRelationshipType retrieves a single relationship type by ID.
func (db *DB) GetRelationshipType(ctx context.Context, id uuid.UUID) (*models.RelationshipType, error) {
	query := `
		SELECT id, campaign_id, name, inverse_name, is_symmetric,
		       display_label, inverse_display_label, description,
		       created_at, updated_at
		FROM relationship_types
		WHERE id = $1`

	var rt models.RelationshipType
	err := db.QueryRow(ctx, query, id).Scan(
		&rt.ID, &rt.CampaignID, &rt.Name, &rt.InverseName, &rt.IsSymmetric,
		&rt.DisplayLabel, &rt.InverseDisplayLabel, &rt.Description,
		&rt.CreatedAt, &rt.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get relationship type: %w", err)
	}

	return &rt, nil
}

// CreateRelationshipType creates a new campaign-specific relationship type.
func (db *DB) CreateRelationshipType(ctx context.Context, campaignID uuid.UUID, req models.CreateRelationshipTypeRequest) (*models.RelationshipType, error) {
	id := uuid.New()

	query := `
		INSERT INTO relationship_types (id, campaign_id, name, inverse_name, is_symmetric,
		                                display_label, inverse_display_label, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, campaign_id, name, inverse_name, is_symmetric,
		          display_label, inverse_display_label, description,
		          created_at, updated_at`

	var rt models.RelationshipType
	err := db.QueryRow(ctx, query,
		id, campaignID, req.Name, req.InverseName, req.IsSymmetric,
		req.DisplayLabel, req.InverseDisplayLabel, req.Description,
	).Scan(
		&rt.ID, &rt.CampaignID, &rt.Name, &rt.InverseName, &rt.IsSymmetric,
		&rt.DisplayLabel, &rt.InverseDisplayLabel, &rt.Description,
		&rt.CreatedAt, &rt.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create relationship type: %w", err)
	}

	return &rt, nil
}

// DeleteRelationshipType deletes a campaign-specific relationship type.
// System default types (campaign_id IS NULL) cannot be deleted.
func (db *DB) DeleteRelationshipType(ctx context.Context, id uuid.UUID) error {
	// Only delete if it's a campaign-specific type (not a system default)
	query := `DELETE FROM relationship_types WHERE id = $1 AND campaign_id IS NOT NULL`
	result, err := db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete relationship type: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("relationship type not found or is a system default")
	}

	return nil
}

// GetInverseType returns the inverse type name for a given relationship type.
// It checks campaign-specific types first, then falls back to system defaults.
func (db *DB) GetInverseType(ctx context.Context, campaignID uuid.UUID, typeName string) (string, error) {
	query := `SELECT get_inverse_relationship_type($1, $2)`

	var inverseName *string
	err := db.QueryRow(ctx, query, campaignID, typeName).Scan(&inverseName)
	if err != nil {
		return "", fmt.Errorf("failed to get inverse relationship type: %w", err)
	}

	if inverseName == nil {
		return "", fmt.Errorf("relationship type not found: %s", typeName)
	}

	return *inverseName, nil
}

// IsSystemRelationshipType checks if a relationship type is a system default.
func (db *DB) IsSystemRelationshipType(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT campaign_id IS NULL FROM relationship_types WHERE id = $1`

	var isSystem bool
	err := db.QueryRow(ctx, query, id).Scan(&isSystem)
	if err != nil {
		return false, fmt.Errorf("failed to check relationship type: %w", err)
	}

	return isSystem, nil
}
