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

// ListRelationshipTypes returns all relationship types for a campaign.
// All types are campaign-scoped, seeded from templates on campaign creation.
// Results are sorted by name.
func (db *DB) ListRelationshipTypes(ctx context.Context, campaignID int64) ([]models.RelationshipType, error) {
	query := `
		SELECT id, campaign_id, name, inverse_name, is_symmetric,
		       display_label, inverse_display_label, description,
		       created_at, updated_at
		FROM relationship_types
		WHERE campaign_id = $1
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
func (db *DB) GetRelationshipType(ctx context.Context, id int64) (*models.RelationshipType, error) {
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
func (db *DB) CreateRelationshipType(ctx context.Context, campaignID int64, req models.CreateRelationshipTypeRequest) (*models.RelationshipType, error) {
	query := `
		INSERT INTO relationship_types (campaign_id, name, inverse_name, is_symmetric,
		                                display_label, inverse_display_label, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, campaign_id, name, inverse_name, is_symmetric,
		          display_label, inverse_display_label, description,
		          created_at, updated_at`

	var rt models.RelationshipType
	err := db.QueryRow(ctx, query,
		campaignID, req.Name, req.InverseName, req.IsSymmetric,
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

// DeleteRelationshipType deletes a relationship type by ID.
// All types are campaign-scoped and deletable.
func (db *DB) DeleteRelationshipType(ctx context.Context, id int64) error {
	query := `DELETE FROM relationship_types WHERE id = $1`
	result, err := db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete relationship type: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("relationship type not found")
	}

	return nil
}

// GetRelationshipTypeByName retrieves a relationship type by its name for a
// given campaign. All types are campaign-scoped.
func (db *DB) GetRelationshipTypeByName(ctx context.Context, campaignID int64, name string) (*models.RelationshipType, error) {
	query := `
		SELECT id, campaign_id, name, inverse_name, is_symmetric,
		       display_label, inverse_display_label, description,
		       created_at, updated_at
		FROM relationship_types
		WHERE name = $1
		  AND campaign_id = $2`

	var rt models.RelationshipType
	err := db.QueryRow(ctx, query, name, campaignID).Scan(
		&rt.ID, &rt.CampaignID, &rt.Name, &rt.InverseName, &rt.IsSymmetric,
		&rt.DisplayLabel, &rt.InverseDisplayLabel, &rt.Description,
		&rt.CreatedAt, &rt.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get relationship type by name: %w", err)
	}

	return &rt, nil
}

// GetInverseType returns the inverse type name for a given relationship type.
// It checks campaign-specific types first, then falls back to system defaults.
func (db *DB) GetInverseType(ctx context.Context, campaignID int64, typeName string) (string, error) {
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
func (db *DB) IsSystemRelationshipType(ctx context.Context, id int64) (bool, error) {
	query := `SELECT campaign_id IS NULL FROM relationship_types WHERE id = $1`

	var isSystem bool
	err := db.QueryRow(ctx, query, id).Scan(&isSystem)
	if err != nil {
		return false, fmt.Errorf("failed to check relationship type: %w", err)
	}

	return isSystem, nil
}
