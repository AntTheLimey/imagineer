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

// ListSessionsByCampaign retrieves all sessions for a campaign.
func (db *DB) ListSessionsByCampaign(ctx context.Context, campaignID uuid.UUID) ([]models.Session, error) {
	query := `
        SELECT id, campaign_id, session_number, planned_date, actual_date,
               status, prep_notes, planned_scenes, actual_notes, discoveries,
               player_decisions, consequences, created_at, updated_at
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
		err := rows.Scan(
			&s.ID, &s.CampaignID, &s.SessionNumber, &s.PlannedDate, &s.ActualDate,
			&s.Status, &s.PrepNotes, &s.PlannedScenes, &s.ActualNotes, &s.Discoveries,
			&s.PlayerDecisions, &s.Consequences, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

// GetSession retrieves a session by ID.
func (db *DB) GetSession(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	query := `
        SELECT id, campaign_id, session_number, planned_date, actual_date,
               status, prep_notes, planned_scenes, actual_notes, discoveries,
               player_decisions, consequences, created_at, updated_at
        FROM sessions
        WHERE id = $1`

	var s models.Session
	err := db.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.CampaignID, &s.SessionNumber, &s.PlannedDate, &s.ActualDate,
		&s.Status, &s.PrepNotes, &s.PlannedScenes, &s.ActualNotes, &s.Discoveries,
		&s.PlayerDecisions, &s.Consequences, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &s, nil
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
