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
	"errors"
	"fmt"

	"github.com/antonypegg/imagineer/internal/models"
	"github.com/jackc/pgx/v5"
)

// sceneColumns is the standard column list for scene queries.
const sceneColumns = `id, session_id, campaign_id, title, description,
	scene_type, status, sort_order, objective, gm_notes,
	entity_ids, system_data, source, source_confidence,
	connections, created_at, updated_at`

// scanScene scans a single row into a models.Scene.
func scanScene(row pgx.Row) (*models.Scene, error) {
	var s models.Scene
	var entityIDs []int64
	err := row.Scan(
		&s.ID, &s.SessionID, &s.CampaignID, &s.Title, &s.Description,
		&s.SceneType, &s.Status, &s.SortOrder, &s.Objective, &s.GMNotes,
		&entityIDs, &s.SystemData, &s.Source, &s.SourceConfidence,
		&s.Connections, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if entityIDs == nil {
		entityIDs = []int64{}
	}
	s.EntityIDs = entityIDs
	return &s, nil
}

// scanScenes scans multiple scene rows.
func scanScenes(rows pgx.Rows) ([]models.Scene, error) {
	var scenes []models.Scene
	for rows.Next() {
		var s models.Scene
		var entityIDs []int64
		err := rows.Scan(
			&s.ID, &s.SessionID, &s.CampaignID, &s.Title, &s.Description,
			&s.SceneType, &s.Status, &s.SortOrder, &s.Objective, &s.GMNotes,
			&entityIDs, &s.SystemData, &s.Source, &s.SourceConfidence,
			&s.Connections, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan scene: %w", err)
		}
		if entityIDs == nil {
			entityIDs = []int64{}
		}
		s.EntityIDs = entityIDs
		scenes = append(scenes, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating scenes: %w", err)
	}

	return scenes, nil
}

// ListScenesBySession retrieves all scenes for a session ordered by
// sort_order ascending, then created_at ascending.
func (db *DB) ListScenesBySession(ctx context.Context, sessionID int64) ([]models.Scene, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM scenes
		WHERE session_id = $1
		ORDER BY sort_order ASC, created_at ASC`, sceneColumns)

	rows, err := db.Query(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query scenes: %w", err)
	}
	defer rows.Close()

	scenes, err := scanScenes(rows)
	if err != nil {
		return nil, err
	}

	if scenes == nil {
		scenes = []models.Scene{}
	}

	return scenes, nil
}

// GetScene retrieves a scene by ID.
// Returns pgx.ErrNoRows (unwrapped) when the scene does not exist so
// callers can distinguish 404 from 500 with errors.Is().
func (db *DB) GetScene(ctx context.Context, id int64) (*models.Scene, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM scenes
		WHERE id = $1`, sceneColumns)

	s, err := scanScene(db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get scene: %w", err)
	}

	return s, nil
}

// CreateScene creates a new scene in a session.
// If SortOrder is not provided, it will be set to the next available value
// using a transaction with an advisory lock to prevent race conditions.
func (db *DB) CreateScene(ctx context.Context, sessionID, campaignID int64, req models.CreateSceneRequest) (*models.Scene, error) {
	// Resolve defaults
	sceneType := "other"
	if req.SceneType != nil {
		sceneType = *req.SceneType
	}

	source := "manual"
	if req.Source != nil {
		source = *req.Source
	}

	sourceConfidence := models.SourceConfidenceDraft
	if req.SourceConfidence != nil {
		sourceConfidence = *req.SourceConfidence
	}

	entityIDs := []int64{}
	if req.EntityIDs != nil {
		entityIDs = req.EntityIDs
	}

	systemData := json.RawMessage("{}")
	if req.SystemData != nil {
		systemData = req.SystemData
	}

	connections := json.RawMessage("[]")
	if req.Connections != nil {
		connections = req.Connections
	}

	insertQuery := fmt.Sprintf(`
		INSERT INTO scenes (session_id, campaign_id, title, description,
			scene_type, sort_order, objective, gm_notes,
			entity_ids, system_data, source, source_confidence, connections)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING %s`, sceneColumns)

	// If sort_order is provided, skip the transaction
	if req.SortOrder != nil {
		s, err := scanScene(db.QueryRow(ctx, insertQuery,
			sessionID, campaignID, req.Title, req.Description,
			sceneType, *req.SortOrder, req.Objective, req.GMNotes,
			entityIDs, systemData, source, sourceConfidence, connections,
		))
		if err != nil {
			return nil, fmt.Errorf("failed to create scene: %w", err)
		}
		return s, nil
	}

	// Use transaction with advisory lock for auto-calculated sort_order
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // Rollback is a no-op if already committed

	// Acquire advisory lock on session_id to serialize sort_order calculation
	_, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire advisory lock: %w", err)
	}

	// Get the max sort_order for this session and add 1
	var maxSortOrder *int
	err = tx.QueryRow(ctx,
		"SELECT MAX(sort_order) FROM scenes WHERE session_id = $1",
		sessionID,
	).Scan(&maxSortOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to determine sort_order: %w", err)
	}

	sortOrder := 0
	if maxSortOrder != nil {
		sortOrder = *maxSortOrder + 1
	}

	s, err := scanScene(tx.QueryRow(ctx, insertQuery,
		sessionID, campaignID, req.Title, req.Description,
		sceneType, sortOrder, req.Objective, req.GMNotes,
		entityIDs, systemData, source, sourceConfidence, connections,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create scene: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s, nil
}

// UpdateScene updates an existing scene.
// The entire read-modify-write is wrapped in a transaction with
// SELECT ... FOR UPDATE to prevent concurrent lost updates.
func (db *DB) UpdateScene(ctx context.Context, id int64, req models.UpdateSceneRequest) (*models.Scene, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Fetch the existing scene inside the transaction with a row lock.
	selectQuery := fmt.Sprintf(`
		SELECT %s
		FROM scenes
		WHERE id = $1
		FOR UPDATE`, sceneColumns)

	existing, err := scanScene(tx.QueryRow(ctx, selectQuery, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get scene for update: %w", err)
	}

	// Apply updates - for pointer fields, nil means "no change"
	title := existing.Title
	if req.Title != nil {
		title = *req.Title
	}

	description := existing.Description
	if req.Description != nil {
		description = req.Description
	}

	sceneType := existing.SceneType
	if req.SceneType != nil {
		sceneType = *req.SceneType
	}

	status := existing.Status
	if req.Status != nil {
		status = *req.Status
	}

	sortOrder := existing.SortOrder
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	objective := existing.Objective
	if req.Objective != nil {
		objective = req.Objective
	}

	gmNotes := existing.GMNotes
	if req.GMNotes != nil {
		gmNotes = req.GMNotes
	}

	entityIDs := existing.EntityIDs
	if req.EntityIDs != nil {
		entityIDs = req.EntityIDs
	}
	if entityIDs == nil {
		entityIDs = []int64{}
	}

	systemData := existing.SystemData
	if req.SystemData != nil {
		systemData = req.SystemData
	}

	source := existing.Source
	if req.Source != nil {
		source = *req.Source
	}

	sourceConfidence := existing.SourceConfidence
	if req.SourceConfidence != nil {
		sourceConfidence = *req.SourceConfidence
	}

	connections := existing.Connections
	if req.Connections != nil {
		connections = req.Connections
	}

	updateQuery := fmt.Sprintf(`
		UPDATE scenes
		SET title = $2, description = $3, scene_type = $4, status = $5,
			sort_order = $6, objective = $7, gm_notes = $8,
			entity_ids = $9, system_data = $10, source = $11,
			source_confidence = $12, connections = $13, updated_at = NOW()
		WHERE id = $1
		RETURNING %s`, sceneColumns)

	s, err := scanScene(tx.QueryRow(ctx, updateQuery,
		id, title, description, sceneType, status,
		sortOrder, objective, gmNotes,
		entityIDs, systemData, source,
		sourceConfidence, connections,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to update scene: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s, nil
}

// DeleteScene deletes a scene by ID.
// Returns pgx.ErrNoRows when no scene matches the given ID.
func (db *DB) DeleteScene(ctx context.Context, id int64) error {
	result, err := db.Pool.Exec(ctx, "DELETE FROM scenes WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete scene: %w", err)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
