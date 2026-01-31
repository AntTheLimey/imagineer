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
	"github.com/lib/pq"
)

// ListTimelineEventsByCampaign retrieves all timeline events for a campaign.
func (db *DB) ListTimelineEventsByCampaign(ctx context.Context, campaignID uuid.UUID) ([]models.TimelineEvent, error) {
	query := `
        SELECT id, campaign_id, event_date, event_time, date_precision,
               description, entity_ids, session_id, is_player_known,
               source_document, created_at, updated_at
        FROM timeline_events
        WHERE campaign_id = $1
        ORDER BY event_date ASC NULLS LAST, created_at ASC`

	rows, err := db.Query(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to query timeline events: %w", err)
	}
	defer rows.Close()

	var events []models.TimelineEvent
	for rows.Next() {
		var e models.TimelineEvent
		var entityIDs []uuid.UUID
		err := rows.Scan(
			&e.ID, &e.CampaignID, &e.EventDate, &e.EventTime, &e.DatePrecision,
			&e.Description, pq.Array(&entityIDs), &e.SessionID, &e.IsPlayerKnown,
			&e.SourceDocument, &e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan timeline event: %w", err)
		}
		e.EntityIDs = entityIDs
		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating timeline events: %w", err)
	}

	return events, nil
}

// GetTimelineEvent retrieves a timeline event by ID.
func (db *DB) GetTimelineEvent(ctx context.Context, id uuid.UUID) (*models.TimelineEvent, error) {
	query := `
        SELECT id, campaign_id, event_date, event_time, date_precision,
               description, entity_ids, session_id, is_player_known,
               source_document, created_at, updated_at
        FROM timeline_events
        WHERE id = $1`

	var e models.TimelineEvent
	var entityIDs []uuid.UUID
	err := db.QueryRow(ctx, query, id).Scan(
		&e.ID, &e.CampaignID, &e.EventDate, &e.EventTime, &e.DatePrecision,
		&e.Description, pq.Array(&entityIDs), &e.SessionID, &e.IsPlayerKnown,
		&e.SourceDocument, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get timeline event: %w", err)
	}
	e.EntityIDs = entityIDs

	return &e, nil
}

// CreateTimelineEvent creates a new timeline event.
func (db *DB) CreateTimelineEvent(ctx context.Context, campaignID uuid.UUID, req models.CreateTimelineEventRequest) (*models.TimelineEvent, error) {
	id := uuid.New()

	entityIDs := req.EntityIDs
	if entityIDs == nil {
		entityIDs = []uuid.UUID{}
	}

	query := `
        INSERT INTO timeline_events (id, campaign_id, event_date, event_time,
                                     date_precision, description, entity_ids,
                                     session_id, is_player_known, source_document)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING id, campaign_id, event_date, event_time, date_precision,
                  description, entity_ids, session_id, is_player_known,
                  source_document, created_at, updated_at`

	var e models.TimelineEvent
	var returnedEntityIDs []uuid.UUID
	err := db.QueryRow(ctx, query,
		id, campaignID, req.EventDate, req.EventTime, req.DatePrecision,
		req.Description, pq.Array(entityIDs), req.SessionID, req.IsPlayerKnown,
		req.SourceDocument,
	).Scan(
		&e.ID, &e.CampaignID, &e.EventDate, &e.EventTime, &e.DatePrecision,
		&e.Description, pq.Array(&returnedEntityIDs), &e.SessionID, &e.IsPlayerKnown,
		&e.SourceDocument, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create timeline event: %w", err)
	}
	e.EntityIDs = returnedEntityIDs

	return &e, nil
}

// DeleteTimelineEvent deletes a timeline event by ID.
func (db *DB) DeleteTimelineEvent(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM timeline_events WHERE id = $1`
	result, err := db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete timeline event: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("timeline event not found")
	}

	return nil
}
