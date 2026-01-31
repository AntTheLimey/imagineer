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

// ListRelationshipsByCampaign retrieves all relationships for a campaign.
func (db *DB) ListRelationshipsByCampaign(ctx context.Context, campaignID uuid.UUID) ([]models.Relationship, error) {
	query := `
        SELECT r.id, r.campaign_id, r.source_entity_id, r.target_entity_id,
               r.relationship_type, r.tone, r.description, r.bidirectional,
               r.strength, r.created_at, r.updated_at,
               se.id, se.name, se.entity_type,
               te.id, te.name, te.entity_type
        FROM relationships r
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
		var seID uuid.UUID
		var seName string
		var seType models.EntityType
		var teID uuid.UUID
		var teName string
		var teType models.EntityType

		err := rows.Scan(
			&r.ID, &r.CampaignID, &r.SourceEntityID, &r.TargetEntityID,
			&r.RelationshipType, &r.Tone, &r.Description, &r.Bidirectional,
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
func (db *DB) GetRelationship(ctx context.Context, id uuid.UUID) (*models.Relationship, error) {
	query := `
        SELECT id, campaign_id, source_entity_id, target_entity_id,
               relationship_type, tone, description, bidirectional,
               strength, created_at, updated_at
        FROM relationships
        WHERE id = $1`

	var r models.Relationship
	err := db.QueryRow(ctx, query, id).Scan(
		&r.ID, &r.CampaignID, &r.SourceEntityID, &r.TargetEntityID,
		&r.RelationshipType, &r.Tone, &r.Description, &r.Bidirectional,
		&r.Strength, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get relationship: %w", err)
	}

	return &r, nil
}

// CreateRelationship creates a new relationship.
func (db *DB) CreateRelationship(ctx context.Context, campaignID uuid.UUID, req models.CreateRelationshipRequest) (*models.Relationship, error) {
	id := uuid.New()

	query := `
        INSERT INTO relationships (id, campaign_id, source_entity_id, target_entity_id,
                                   relationship_type, tone, description, bidirectional, strength)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING id, campaign_id, source_entity_id, target_entity_id,
                  relationship_type, tone, description, bidirectional,
                  strength, created_at, updated_at`

	var r models.Relationship
	err := db.QueryRow(ctx, query,
		id, campaignID, req.SourceEntityID, req.TargetEntityID,
		req.RelationshipType, req.Tone, req.Description, req.Bidirectional, req.Strength,
	).Scan(
		&r.ID, &r.CampaignID, &r.SourceEntityID, &r.TargetEntityID,
		&r.RelationshipType, &r.Tone, &r.Description, &r.Bidirectional,
		&r.Strength, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create relationship: %w", err)
	}

	return &r, nil
}

// UpdateRelationship updates an existing relationship.
func (db *DB) UpdateRelationship(ctx context.Context, id uuid.UUID, req models.UpdateRelationshipRequest) (*models.Relationship, error) {
	// First get the existing relationship
	existing, err := db.GetRelationship(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	relationshipType := existing.RelationshipType
	if req.RelationshipType != nil {
		relationshipType = *req.RelationshipType
	}

	tone := existing.Tone
	if req.Tone != nil {
		tone = req.Tone
	}

	description := existing.Description
	if req.Description != nil {
		description = req.Description
	}

	bidirectional := existing.Bidirectional
	if req.Bidirectional != nil {
		bidirectional = *req.Bidirectional
	}

	strength := existing.Strength
	if req.Strength != nil {
		strength = req.Strength
	}

	query := `
        UPDATE relationships
        SET relationship_type = $2, tone = $3, description = $4,
            bidirectional = $5, strength = $6
        WHERE id = $1
        RETURNING id, campaign_id, source_entity_id, target_entity_id,
                  relationship_type, tone, description, bidirectional,
                  strength, created_at, updated_at`

	var r models.Relationship
	err = db.QueryRow(ctx, query,
		id, relationshipType, tone, description, bidirectional, strength,
	).Scan(
		&r.ID, &r.CampaignID, &r.SourceEntityID, &r.TargetEntityID,
		&r.RelationshipType, &r.Tone, &r.Description, &r.Bidirectional,
		&r.Strength, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update relationship: %w", err)
	}

	return &r, nil
}

// DeleteRelationship deletes a relationship by ID.
func (db *DB) DeleteRelationship(ctx context.Context, id uuid.UUID) error {
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
func (db *DB) GetEntityRelationships(ctx context.Context, entityID uuid.UUID) ([]models.Relationship, error) {
	query := `
        SELECT r.id, r.campaign_id, r.source_entity_id, r.target_entity_id,
               r.relationship_type, r.tone, r.description, r.bidirectional,
               r.strength, r.created_at, r.updated_at,
               se.id, se.name, se.entity_type,
               te.id, te.name, te.entity_type
        FROM relationships r
        JOIN entities se ON r.source_entity_id = se.id
        JOIN entities te ON r.target_entity_id = te.id
        WHERE r.source_entity_id = $1 OR r.target_entity_id = $1
        ORDER BY r.created_at DESC`

	rows, err := db.Query(ctx, query, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to query entity relationships: %w", err)
	}
	defer rows.Close()

	var relationships []models.Relationship
	for rows.Next() {
		var r models.Relationship
		var seID uuid.UUID
		var seName string
		var seType models.EntityType
		var teID uuid.UUID
		var teName string
		var teType models.EntityType

		err := rows.Scan(
			&r.ID, &r.CampaignID, &r.SourceEntityID, &r.TargetEntityID,
			&r.RelationshipType, &r.Tone, &r.Description, &r.Bidirectional,
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

	return relationships, nil
}
