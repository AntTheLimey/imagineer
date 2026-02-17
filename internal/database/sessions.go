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
	"errors"
	"fmt"

	"github.com/antonypegg/imagineer/internal/models"
	"github.com/jackc/pgx/v5"
)

// ListSessionsByCampaign retrieves all sessions for a campaign.
func (db *DB) ListSessionsByCampaign(ctx context.Context, campaignID int64) ([]models.Session, error) {
	query := `
        SELECT id, campaign_id, chapter_id, title, session_number, planned_date, actual_date,
               status, stage, prep_notes, actual_notes, play_notes, created_at, updated_at
        FROM sessions
        WHERE campaign_id = $1
        ORDER BY session_number ASC NULLS LAST, created_at ASC`

	rows, err := db.Query(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []models.Session
	for rows.Next() {
		var s models.Session
		var stage *string
		err := rows.Scan(
			&s.ID, &s.CampaignID, &s.ChapterID, &s.Title, &s.SessionNumber, &s.PlannedDate, &s.ActualDate,
			&s.Status, &stage, &s.PrepNotes, &s.ActualNotes, &s.PlayNotes, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		if stage != nil {
			s.Stage = models.SessionStage(*stage)
		}
		sessions = append(sessions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

// ListSessionsByChapter retrieves all sessions for a specific chapter.
func (db *DB) ListSessionsByChapter(ctx context.Context, chapterID int64) ([]models.Session, error) {
	query := `
        SELECT id, campaign_id, chapter_id, title, session_number, planned_date, actual_date,
               status, stage, prep_notes, actual_notes, play_notes, created_at, updated_at
        FROM sessions
        WHERE chapter_id = $1
        ORDER BY session_number ASC NULLS LAST, created_at ASC`

	rows, err := db.Query(ctx, query, chapterID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions by chapter: %w", err)
	}
	defer rows.Close()

	var sessions []models.Session
	for rows.Next() {
		var s models.Session
		var stage *string
		err := rows.Scan(
			&s.ID, &s.CampaignID, &s.ChapterID, &s.Title, &s.SessionNumber, &s.PlannedDate, &s.ActualDate,
			&s.Status, &stage, &s.PrepNotes, &s.ActualNotes, &s.PlayNotes, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		if stage != nil {
			s.Stage = models.SessionStage(*stage)
		}
		sessions = append(sessions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

// GetSession retrieves a session by ID.
// Returns pgx.ErrNoRows (unwrapped) when the session does not exist so
// callers can distinguish 404 from 500 with errors.Is().
func (db *DB) GetSession(ctx context.Context, id int64) (*models.Session, error) {
	query := `
        SELECT id, campaign_id, chapter_id, title, session_number, planned_date, actual_date,
               status, stage, prep_notes, actual_notes, play_notes, created_at, updated_at
        FROM sessions
        WHERE id = $1`

	var s models.Session
	var stage *string
	err := db.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.CampaignID, &s.ChapterID, &s.Title, &s.SessionNumber, &s.PlannedDate, &s.ActualDate,
		&s.Status, &stage, &s.PrepNotes, &s.ActualNotes, &s.PlayNotes, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	if stage != nil {
		s.Stage = models.SessionStage(*stage)
	}

	return &s, nil
}

// CreateSession creates a new session in a campaign.
func (db *DB) CreateSession(ctx context.Context, campaignID int64, req models.CreateSessionRequest) (*models.Session, error) {
	// Default status to PLANNED
	status := models.SessionStatusPlanned

	// Default stage to prep if not provided
	stage := models.SessionStagePrep
	if req.Stage != nil {
		stage = *req.Stage
	}

	query := `
        INSERT INTO sessions (campaign_id, chapter_id, title, session_number, planned_date,
                              status, stage, prep_notes)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, campaign_id, chapter_id, title, session_number, planned_date, actual_date,
                  status, stage, prep_notes, actual_notes, play_notes, created_at, updated_at`

	var s models.Session
	var retStage *string
	err := db.QueryRow(ctx, query, campaignID, req.ChapterID, req.Title, req.SessionNumber,
		req.PlannedDate, status, stage, req.PrepNotes).Scan(
		&s.ID, &s.CampaignID, &s.ChapterID, &s.Title, &s.SessionNumber, &s.PlannedDate, &s.ActualDate,
		&s.Status, &retStage, &s.PrepNotes, &s.ActualNotes, &s.PlayNotes, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	if retStage != nil {
		s.Stage = models.SessionStage(*retStage)
	}

	return &s, nil
}

// UpdateSession updates an existing session.
// The entire read-modify-write is wrapped in a transaction with
// SELECT ... FOR UPDATE to prevent concurrent lost updates.
func (db *DB) UpdateSession(ctx context.Context, id int64, req models.UpdateSessionRequest) (*models.Session, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Fetch the existing session inside the transaction with a row lock.
	selectQuery := `
        SELECT id, campaign_id, chapter_id, title, session_number, planned_date, actual_date,
               status, stage, prep_notes, actual_notes, play_notes, created_at, updated_at
        FROM sessions
        WHERE id = $1
        FOR UPDATE`

	var existing models.Session
	var existingStage *string
	err = tx.QueryRow(ctx, selectQuery, id).Scan(
		&existing.ID, &existing.CampaignID, &existing.ChapterID, &existing.Title,
		&existing.SessionNumber, &existing.PlannedDate, &existing.ActualDate,
		&existing.Status, &existingStage, &existing.PrepNotes, &existing.ActualNotes,
		&existing.PlayNotes, &existing.CreatedAt, &existing.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get session for update: %w", err)
	}
	if existingStage != nil {
		existing.Stage = models.SessionStage(*existingStage)
	}

	// Apply updates - for pointer fields, nil means "no change"
	chapterID := existing.ChapterID
	if req.ChapterID != nil {
		chapterID = req.ChapterID
	}

	title := existing.Title
	if req.Title != nil {
		title = req.Title
	}

	sessionNumber := existing.SessionNumber
	if req.SessionNumber != nil {
		sessionNumber = req.SessionNumber
	}

	plannedDate := existing.PlannedDate
	if req.PlannedDate != nil {
		plannedDate = req.PlannedDate
	}

	actualDate := existing.ActualDate
	if req.ActualDate != nil {
		actualDate = req.ActualDate
	}

	status := existing.Status
	if req.Status != nil {
		status = *req.Status
	}

	stage := existing.Stage
	if req.Stage != nil {
		stage = *req.Stage
	}

	prepNotes := existing.PrepNotes
	if req.PrepNotes != nil {
		prepNotes = req.PrepNotes
	}

	actualNotes := existing.ActualNotes
	if req.ActualNotes != nil {
		actualNotes = req.ActualNotes
	}

	playNotes := existing.PlayNotes
	if req.PlayNotes != nil {
		playNotes = req.PlayNotes
	}

	updateQuery := `
        UPDATE sessions
        SET chapter_id = $2, title = $3, session_number = $4, planned_date = $5,
            actual_date = $6, status = $7, stage = $8, prep_notes = $9,
            actual_notes = $10, play_notes = $11, updated_at = NOW()
        WHERE id = $1
        RETURNING id, campaign_id, chapter_id, title, session_number, planned_date, actual_date,
                  status, stage, prep_notes, actual_notes, play_notes, created_at, updated_at`

	var s models.Session
	var retStage *string
	err = tx.QueryRow(ctx, updateQuery, id, chapterID, title, sessionNumber, plannedDate,
		actualDate, status, stage, prepNotes, actualNotes, playNotes).Scan(
		&s.ID, &s.CampaignID, &s.ChapterID, &s.Title, &s.SessionNumber, &s.PlannedDate, &s.ActualDate,
		&s.Status, &retStage, &s.PrepNotes, &s.ActualNotes, &s.PlayNotes, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}
	if retStage != nil {
		s.Stage = models.SessionStage(*retStage)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &s, nil
}

// DeleteSession deletes a session by ID.
func (db *DB) DeleteSession(ctx context.Context, id int64) error {
	query := `DELETE FROM sessions WHERE id = $1`
	result, err := db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

// CountSessions counts total sessions.
func (db *DB) CountSessions(ctx context.Context) (int, error) {
	var count int
	err := db.QueryRow(ctx, "SELECT COUNT(*) FROM sessions").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count sessions: %w", err)
	}
	return count, nil
}
