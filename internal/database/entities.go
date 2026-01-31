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
	"encoding/json"
	"fmt"

	"github.com/antonypegg/imagineer/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ListEntitiesByCampaign retrieves all entities for a campaign.
func (db *DB) ListEntitiesByCampaign(ctx context.Context, campaignID uuid.UUID) ([]models.Entity, error) {
	query := `
        SELECT id, campaign_id, entity_type, name, description, attributes,
               tags, keeper_notes, discovered_session, source_document,
               source_confidence, version, created_at, updated_at
        FROM entities
        WHERE campaign_id = $1
        ORDER BY name`

	rows, err := db.Query(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to query entities: %w", err)
	}
	defer rows.Close()

	return scanEntities(rows)
}

// ListEntitiesByType retrieves all entities of a specific type for a campaign.
func (db *DB) ListEntitiesByType(ctx context.Context, campaignID uuid.UUID, entityType models.EntityType) ([]models.Entity, error) {
	query := `
        SELECT id, campaign_id, entity_type, name, description, attributes,
               tags, keeper_notes, discovered_session, source_document,
               source_confidence, version, created_at, updated_at
        FROM entities
        WHERE campaign_id = $1 AND entity_type = $2
        ORDER BY name`

	rows, err := db.Query(ctx, query, campaignID, entityType)
	if err != nil {
		return nil, fmt.Errorf("failed to query entities by type: %w", err)
	}
	defer rows.Close()

	return scanEntities(rows)
}

// GetEntity retrieves an entity by ID.
func (db *DB) GetEntity(ctx context.Context, id uuid.UUID) (*models.Entity, error) {
	query := `
        SELECT id, campaign_id, entity_type, name, description, attributes,
               tags, keeper_notes, discovered_session, source_document,
               source_confidence, version, created_at, updated_at
        FROM entities
        WHERE id = $1`

	var e models.Entity
	err := db.QueryRow(ctx, query, id).Scan(
		&e.ID, &e.CampaignID, &e.EntityType, &e.Name, &e.Description,
		&e.Attributes, &e.Tags, &e.KeeperNotes, &e.DiscoveredSession,
		&e.SourceDocument, &e.SourceConfidence, &e.Version,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	return &e, nil
}

// CreateEntity creates a new entity.
func (db *DB) CreateEntity(ctx context.Context, campaignID uuid.UUID, req models.CreateEntityRequest) (*models.Entity, error) {
	id := uuid.New()

	attributes := req.Attributes
	if attributes == nil {
		attributes = json.RawMessage("{}")
	}

	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}

	sourceConfidence := models.SourceConfidenceDraft
	if req.SourceConfidence != nil {
		sourceConfidence = *req.SourceConfidence
	}

	query := `
        INSERT INTO entities (id, campaign_id, entity_type, name, description,
                              attributes, tags, keeper_notes, discovered_session,
                              source_document, source_confidence)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        RETURNING id, campaign_id, entity_type, name, description, attributes,
                  tags, keeper_notes, discovered_session, source_document,
                  source_confidence, version, created_at, updated_at`

	var e models.Entity
	err := db.QueryRow(ctx, query,
		id, campaignID, req.EntityType, req.Name, req.Description,
		attributes, tags, req.KeeperNotes, req.DiscoveredSession,
		req.SourceDocument, sourceConfidence,
	).Scan(
		&e.ID, &e.CampaignID, &e.EntityType, &e.Name, &e.Description,
		&e.Attributes, &e.Tags, &e.KeeperNotes, &e.DiscoveredSession,
		&e.SourceDocument, &e.SourceConfidence, &e.Version,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}

	return &e, nil
}

// UpdateEntity updates an existing entity.
func (db *DB) UpdateEntity(ctx context.Context, id uuid.UUID, req models.UpdateEntityRequest) (*models.Entity, error) {
	// First get the existing entity
	existing, err := db.GetEntity(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	entityType := existing.EntityType
	if req.EntityType != nil {
		entityType = *req.EntityType
	}

	name := existing.Name
	if req.Name != nil {
		name = *req.Name
	}

	description := existing.Description
	if req.Description != nil {
		description = req.Description
	}

	attributes := existing.Attributes
	if req.Attributes != nil {
		attributes = req.Attributes
	}

	tags := existing.Tags
	if req.Tags != nil {
		tags = req.Tags
	}

	keeperNotes := existing.KeeperNotes
	if req.KeeperNotes != nil {
		keeperNotes = req.KeeperNotes
	}

	discoveredSession := existing.DiscoveredSession
	if req.DiscoveredSession != nil {
		discoveredSession = req.DiscoveredSession
	}

	sourceDocument := existing.SourceDocument
	if req.SourceDocument != nil {
		sourceDocument = req.SourceDocument
	}

	sourceConfidence := existing.SourceConfidence
	if req.SourceConfidence != nil {
		sourceConfidence = *req.SourceConfidence
	}

	query := `
        UPDATE entities
        SET entity_type = $2, name = $3, description = $4, attributes = $5,
            tags = $6, keeper_notes = $7, discovered_session = $8,
            source_document = $9, source_confidence = $10, version = version + 1
        WHERE id = $1
        RETURNING id, campaign_id, entity_type, name, description, attributes,
                  tags, keeper_notes, discovered_session, source_document,
                  source_confidence, version, created_at, updated_at`

	var e models.Entity
	err = db.QueryRow(ctx, query,
		id, entityType, name, description, attributes,
		tags, keeperNotes, discoveredSession,
		sourceDocument, sourceConfidence,
	).Scan(
		&e.ID, &e.CampaignID, &e.EntityType, &e.Name, &e.Description,
		&e.Attributes, &e.Tags, &e.KeeperNotes, &e.DiscoveredSession,
		&e.SourceDocument, &e.SourceConfidence, &e.Version,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update entity: %w", err)
	}

	return &e, nil
}

// DeleteEntity deletes an entity by ID.
func (db *DB) DeleteEntity(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM entities WHERE id = $1`
	result, err := db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete entity: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("entity not found")
	}

	return nil
}

// CountEntities counts total entities.
func (db *DB) CountEntities(ctx context.Context) (int, error) {
	var count int
	err := db.QueryRow(ctx, "SELECT COUNT(*) FROM entities").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count entities: %w", err)
	}
	return count, nil
}

// CountEntitiesByType counts entities grouped by type.
func (db *DB) CountEntitiesByType(ctx context.Context) (map[string]int, error) {
	query := `
        SELECT entity_type, COUNT(*)
        FROM entities
        GROUP BY entity_type`

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to count entities by type: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var entityType string
		var count int
		if err := rows.Scan(&entityType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan entity count: %w", err)
		}
		counts[entityType] = count
	}

	return counts, nil
}

// SearchEntitiesByName searches for entities with similar names.
func (db *DB) SearchEntitiesByName(ctx context.Context, campaignID uuid.UUID, name string, limit int) ([]models.Entity, error) {
	query := `
        SELECT id, campaign_id, entity_type, name, description, attributes,
               tags, keeper_notes, discovered_session, source_document,
               source_confidence, version, created_at, updated_at
        FROM entities
        WHERE campaign_id = $1 AND name % $2
        ORDER BY similarity(name, $2) DESC
        LIMIT $3`

	rows, err := db.Query(ctx, query, campaignID, name, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search entities: %w", err)
	}
	defer rows.Close()

	return scanEntities(rows)
}

// scanEntities scans multiple entity rows.
func scanEntities(rows pgx.Rows) ([]models.Entity, error) {
	var entities []models.Entity
	for rows.Next() {
		var e models.Entity
		err := rows.Scan(
			&e.ID, &e.CampaignID, &e.EntityType, &e.Name, &e.Description,
			&e.Attributes, &e.Tags, &e.KeeperNotes, &e.DiscoveredSession,
			&e.SourceDocument, &e.SourceConfidence, &e.Version,
			&e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entity: %w", err)
		}
		entities = append(entities, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating entities: %w", err)
	}

	return entities, nil
}
