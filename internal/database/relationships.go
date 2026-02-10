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

// ListRelationshipsByCampaign retrieves all relationships for a campaign.
func (db *DB) ListRelationshipsByCampaign(ctx context.Context, campaignID int64) ([]models.Relationship, error) {
	query := `
		SELECT r.id, r.campaign_id, r.source_entity_id, r.target_entity_id,
		       r.relationship_type_id, rt.name, r.tone, r.description,
		       r.strength, r.created_at, r.updated_at,
		       se.id, se.name, se.entity_type,
		       te.id, te.name, te.entity_type
		FROM relationships r
		JOIN relationship_types rt ON r.relationship_type_id = rt.id
		JOIN entities se ON r.source_entity_id = se.id
		JOIN entities te ON r.target_entity_id = te.id
		WHERE r.campaign_id = $1
		ORDER BY r.created_at DESC`

	rows, err := db.Query(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to query relationships: %w", err)
	}
	defer rows.Close()

	var relationships []models.Relationship
	for rows.Next() {
		var r models.Relationship
		var seID int64
		var seName string
		var seType models.EntityType
		var teID int64
		var teName string
		var teType models.EntityType

		err := rows.Scan(
			&r.ID, &r.CampaignID, &r.SourceEntityID, &r.TargetEntityID,
			&r.RelationshipTypeID, &r.RelationshipTypeName, &r.Tone, &r.Description,
			&r.Strength, &r.CreatedAt, &r.UpdatedAt,
			&seID, &seName, &seType,
			&teID, &teName, &teType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan relationship: %w", err)
		}

		r.SourceEntity = &models.Entity{
			ID:         seID,
			Name:       seName,
			EntityType: seType,
		}
		r.TargetEntity = &models.Entity{
			ID:         teID,
			Name:       teName,
			EntityType: teType,
		}

		relationships = append(relationships, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating relationships: %w", err)
	}

	return relationships, nil
}

// GetRelationship retrieves a relationship by ID.
func (db *DB) GetRelationship(ctx context.Context, id int64) (*models.Relationship, error) {
	query := `
		SELECT r.id, r.campaign_id, r.source_entity_id, r.target_entity_id,
		       r.relationship_type_id, rt.name, r.tone, r.description,
		       r.strength, r.created_at, r.updated_at
		FROM relationships r
		JOIN relationship_types rt ON r.relationship_type_id = rt.id
		WHERE r.id = $1`

	var r models.Relationship
	err := db.QueryRow(ctx, query, id).Scan(
		&r.ID, &r.CampaignID, &r.SourceEntityID, &r.TargetEntityID,
		&r.RelationshipTypeID, &r.RelationshipTypeName, &r.Tone, &r.Description,
		&r.Strength, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get relationship: %w", err)
	}

	return &r, nil
}

// CreateRelationship creates a new relationship or returns the existing one
// if a duplicate is detected. When a conflict occurs on the unique constraint
// (campaign_id, source_entity_id, target_entity_id, relationship_type_id),
// the existing row is updated with any new description or tone values, making
// this operation idempotent.
func (db *DB) CreateRelationship(ctx context.Context, campaignID int64, req models.CreateRelationshipRequest) (*models.Relationship, error) {
	query := `
		INSERT INTO relationships (campaign_id, source_entity_id, target_entity_id,
		                           relationship_type_id, tone, description, strength)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (campaign_id, source_entity_id, target_entity_id, relationship_type_id)
		DO UPDATE SET
		    description = COALESCE(EXCLUDED.description, relationships.description),
		    tone = COALESCE(EXCLUDED.tone, relationships.tone),
		    strength = COALESCE(EXCLUDED.strength, relationships.strength),
		    updated_at = NOW()
		RETURNING id, campaign_id, source_entity_id, target_entity_id,
		          relationship_type_id, tone, description,
		          strength, created_at, updated_at`

	var r models.Relationship
	err := db.QueryRow(ctx, query,
		campaignID, req.SourceEntityID, req.TargetEntityID,
		req.RelationshipTypeID, req.Tone, req.Description, req.Strength,
	).Scan(
		&r.ID, &r.CampaignID, &r.SourceEntityID, &r.TargetEntityID,
		&r.RelationshipTypeID, &r.Tone, &r.Description,
		&r.Strength, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create relationship: %w", err)
	}

	return &r, nil
}

// UpdateRelationship updates an existing relationship.
func (db *DB) UpdateRelationship(ctx context.Context, id int64, req models.UpdateRelationshipRequest) (*models.Relationship, error) {
	// First get the existing relationship
	existing, err := db.GetRelationship(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	relationshipTypeID := existing.RelationshipTypeID
	if req.RelationshipTypeID != nil {
		relationshipTypeID = *req.RelationshipTypeID
	}

	tone := existing.Tone
	if req.Tone != nil {
		tone = req.Tone
	}

	description := existing.Description
	if req.Description != nil {
		description = req.Description
	}

	strength := existing.Strength
	if req.Strength != nil {
		strength = req.Strength
	}

	query := `
		UPDATE relationships
		SET relationship_type_id = $2, tone = $3, description = $4,
		    strength = $5, updated_at = NOW()
		WHERE id = $1
		RETURNING id, campaign_id, source_entity_id, target_entity_id,
		          relationship_type_id, tone, description,
		          strength, created_at, updated_at`

	var r models.Relationship
	err = db.QueryRow(ctx, query,
		id, relationshipTypeID, tone, description, strength,
	).Scan(
		&r.ID, &r.CampaignID, &r.SourceEntityID, &r.TargetEntityID,
		&r.RelationshipTypeID, &r.Tone, &r.Description,
		&r.Strength, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update relationship: %w", err)
	}

	return &r, nil
}

// DeleteRelationship deletes a relationship by ID.
func (db *DB) DeleteRelationship(ctx context.Context, id int64) error {
	query := `DELETE FROM relationships WHERE id = $1`
	result, err := db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete relationship: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("relationship not found")
	}

	return nil
}

// CountRelationships counts total relationships.
func (db *DB) CountRelationships(ctx context.Context) (int, error) {
	var count int
	err := db.QueryRow(ctx, "SELECT COUNT(*) FROM relationships").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count relationships: %w", err)
	}
	return count, nil
}

// GetEntityRelationships retrieves all relationships for a specific entity.
// It queries the entity_relationships_view which provides both forward and
// inverse perspectives. The view's from_entity_id/to_entity_id are mapped
// to SourceEntityID/TargetEntityID for API consistency.
func (db *DB) GetEntityRelationships(ctx context.Context, entityID int64) ([]models.Relationship, error) {
	query := `
		SELECT id, campaign_id, from_entity_id, to_entity_id,
			relationship_type_id, relationship_type, display_label,
			tone, description, strength,
			created_at, updated_at,
			from_entity_name, from_entity_type,
			to_entity_name, to_entity_type,
			direction
		FROM entity_relationships_view
		WHERE from_entity_id = $1
		ORDER BY relationship_type, to_entity_name`

	rows, err := db.Query(ctx, query, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity relationships: %w", err)
	}
	defer rows.Close()

	var relationships []models.Relationship
	for rows.Next() {
		var r models.Relationship
		err := rows.Scan(
			&r.ID, &r.CampaignID, &r.SourceEntityID, &r.TargetEntityID,
			&r.RelationshipTypeID, &r.RelationshipTypeName, &r.DisplayLabel,
			&r.Tone, &r.Description, &r.Strength,
			&r.CreatedAt, &r.UpdatedAt,
			&r.SourceEntityName, &r.SourceEntityType,
			&r.TargetEntityName, &r.TargetEntityType,
			&r.Direction,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan relationship: %w", err)
		}
		relationships = append(relationships, r)
	}

	return relationships, nil
}
