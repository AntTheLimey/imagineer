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

// ListChapterEntities retrieves all entity links for a chapter.
func (db *DB) ListChapterEntities(ctx context.Context, chapterID int64) ([]models.ChapterEntity, error) {
	query := `
        SELECT ce.id, ce.chapter_id, ce.entity_id, ce.mention_type, ce.created_at,
               e.id, e.campaign_id, e.entity_type, e.name, e.description,
               e.attributes, e.tags, e.gm_notes, e.discovered_session,
               e.source_document, e.source_confidence, e.version,
               e.created_at, e.updated_at
        FROM chapter_entities ce
        JOIN entities e ON ce.entity_id = e.id
        WHERE ce.chapter_id = $1
        ORDER BY ce.mention_type, e.name ASC`

	rows, err := db.Query(ctx, query, chapterID)
	if err != nil {
		return nil, fmt.Errorf("failed to query chapter entities: %w", err)
	}
	defer rows.Close()

	var links []models.ChapterEntity
	for rows.Next() {
		var ce models.ChapterEntity
		var e models.Entity
		err := rows.Scan(
			&ce.ID, &ce.ChapterID, &ce.EntityID, &ce.MentionType, &ce.CreatedAt,
			&e.ID, &e.CampaignID, &e.EntityType, &e.Name, &e.Description,
			&e.Attributes, &e.Tags, &e.GMNotes, &e.DiscoveredSession,
			&e.SourceDocument, &e.SourceConfidence, &e.Version,
			&e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chapter entity: %w", err)
		}
		ce.Entity = &e
		links = append(links, ce)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating chapter entities: %w", err)
	}

	return links, nil
}

// GetChapterEntity retrieves a chapter-entity link by ID.
func (db *DB) GetChapterEntity(ctx context.Context, id int64) (*models.ChapterEntity, error) {
	query := `
        SELECT ce.id, ce.chapter_id, ce.entity_id, ce.mention_type, ce.created_at,
               e.id, e.campaign_id, e.entity_type, e.name, e.description,
               e.attributes, e.tags, e.gm_notes, e.discovered_session,
               e.source_document, e.source_confidence, e.version,
               e.created_at, e.updated_at
        FROM chapter_entities ce
        JOIN entities e ON ce.entity_id = e.id
        WHERE ce.id = $1`

	var ce models.ChapterEntity
	var e models.Entity
	err := db.QueryRow(ctx, query, id).Scan(
		&ce.ID, &ce.ChapterID, &ce.EntityID, &ce.MentionType, &ce.CreatedAt,
		&e.ID, &e.CampaignID, &e.EntityType, &e.Name, &e.Description,
		&e.Attributes, &e.Tags, &e.GMNotes, &e.DiscoveredSession,
		&e.SourceDocument, &e.SourceConfidence, &e.Version,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get chapter entity: %w", err)
	}
	ce.Entity = &e

	return &ce, nil
}

// CreateChapterEntity creates a new link between a chapter and an entity.
func (db *DB) CreateChapterEntity(ctx context.Context, chapterID int64, req models.CreateChapterEntityRequest) (*models.ChapterEntity, error) {
	mentionType := models.ChapterEntityMentionLinked
	if req.MentionType != nil {
		mentionType = *req.MentionType
	}

	query := `
        INSERT INTO chapter_entities (chapter_id, entity_id, mention_type)
        VALUES ($1, $2, $3)
        RETURNING id, chapter_id, entity_id, mention_type, created_at`

	var ce models.ChapterEntity
	err := db.QueryRow(ctx, query, chapterID, req.EntityID, mentionType).Scan(
		&ce.ID, &ce.ChapterID, &ce.EntityID, &ce.MentionType, &ce.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create chapter entity: %w", err)
	}

	return &ce, nil
}

// UpdateChapterEntity updates an existing chapter-entity link.
func (db *DB) UpdateChapterEntity(ctx context.Context, id int64, req models.UpdateChapterEntityRequest) (*models.ChapterEntity, error) {
	// First get the existing link
	existing, err := db.GetChapterEntity(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	mentionType := existing.MentionType
	if req.MentionType != nil {
		mentionType = *req.MentionType
	}

	query := `
        UPDATE chapter_entities
        SET mention_type = $2
        WHERE id = $1
        RETURNING id, chapter_id, entity_id, mention_type, created_at`

	var ce models.ChapterEntity
	err = db.QueryRow(ctx, query, id, mentionType).Scan(
		&ce.ID, &ce.ChapterID, &ce.EntityID, &ce.MentionType, &ce.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update chapter entity: %w", err)
	}

	return &ce, nil
}

// DeleteChapterEntity deletes a chapter-entity link by ID.
func (db *DB) DeleteChapterEntity(ctx context.Context, id int64) error {
	query := `DELETE FROM chapter_entities WHERE id = $1`
	result, err := db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete chapter entity: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("chapter entity not found")
	}

	return nil
}

// DeleteChapterEntityByChapterAndEntity deletes a chapter-entity link by chapter and entity IDs.
func (db *DB) DeleteChapterEntityByChapterAndEntity(ctx context.Context, chapterID, entityID int64) error {
	query := `DELETE FROM chapter_entities WHERE chapter_id = $1 AND entity_id = $2`
	result, err := db.Pool.Exec(ctx, query, chapterID, entityID)
	if err != nil {
		return fmt.Errorf("failed to delete chapter entity: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("chapter entity not found")
	}

	return nil
}

// ListEntitiesByChapter retrieves all entities linked to a chapter (convenience method).
func (db *DB) ListEntitiesByChapter(ctx context.Context, chapterID int64) ([]models.Entity, error) {
	links, err := db.ListChapterEntities(ctx, chapterID)
	if err != nil {
		return nil, err
	}

	entities := make([]models.Entity, 0, len(links))
	for _, link := range links {
		if link.Entity != nil {
			entities = append(entities, *link.Entity)
		}
	}

	return entities, nil
}

// ListChaptersByEntity retrieves all chapters that reference a specific entity.
func (db *DB) ListChaptersByEntity(ctx context.Context, entityID int64) ([]models.Chapter, error) {
	query := `
        SELECT c.id, c.campaign_id, c.title, c.overview, c.sort_order, c.created_at, c.updated_at
        FROM chapters c
        JOIN chapter_entities ce ON c.id = ce.chapter_id
        WHERE ce.entity_id = $1
        ORDER BY c.sort_order ASC, c.created_at ASC`

	rows, err := db.Query(ctx, query, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to query chapters by entity: %w", err)
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
