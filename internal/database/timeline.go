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
	"github.com/lib/pq"
)

// ListTimelineEventsByCampaign retrieves all timeline events for a campaign.
func (db *DB) ListTimelineEventsByCampaign(ctx context.Context, campaignID int64) ([]models.TimelineEvent, error) {
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
		var entityIDs []int64
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
func (db *DB) GetTimelineEvent(ctx context.Context, id int64) (*models.TimelineEvent, error) {
	query := `
        SELECT id, campaign_id, event_date, event_time, date_precision,
               description, entity_ids, session_id, is_player_known,
               source_document, created_at, updated_at
        FROM timeline_events
        WHERE id = $1`

	var e models.TimelineEvent
	var entityIDs []int64
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
func (db *DB) CreateTimelineEvent(ctx context.Context, campaignID int64, req models.CreateTimelineEventRequest) (*models.TimelineEvent, error) {
	entityIDs := req.EntityIDs
	if entityIDs == nil {
		entityIDs = []int64{}
	}

	query := `
        INSERT INTO timeline_events (campaign_id, event_date, event_time,
                                     date_precision, description, entity_ids,
                                     session_id, is_player_known, source_document)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING id, campaign_id, event_date, event_time, date_precision,
                  description, entity_ids, session_id, is_player_known,
                  source_document, created_at, updated_at`

	var e models.TimelineEvent
	var returnedEntityIDs []int64
	err := db.QueryRow(ctx, query,
		campaignID, req.EventDate, req.EventTime, req.DatePrecision,
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

// UpdateTimelineEvent updates an existing timeline event.
func (db *DB) UpdateTimelineEvent(ctx context.Context, id int64, req models.UpdateTimelineEventRequest) (*models.TimelineEvent, error) {
	// First get the existing event
	existing, err := db.GetTimelineEvent(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	eventDate := existing.EventDate
	if req.EventDate != nil {
		eventDate = req.EventDate
	}

	eventTime := existing.EventTime
	if req.EventTime != nil {
		eventTime = req.EventTime
	}

	datePrecision := existing.DatePrecision
	if req.DatePrecision != nil {
		datePrecision = *req.DatePrecision
	}

	description := existing.Description
	if req.Description != nil {
		description = *req.Description
	}

	entityIDs := existing.EntityIDs
	if req.EntityIDs != nil {
		entityIDs = req.EntityIDs
	}

	sessionID := existing.SessionID
	if req.SessionID != nil {
		sessionID = req.SessionID
	}

	isPlayerKnown := existing.IsPlayerKnown
	if req.IsPlayerKnown != nil {
		isPlayerKnown = *req.IsPlayerKnown
	}

	sourceDocument := existing.SourceDocument
	if req.SourceDocument != nil {
		sourceDocument = req.SourceDocument
	}

	query := `
        UPDATE timeline_events
        SET event_date = $2, event_time = $3, date_precision = $4,
            description = $5, entity_ids = $6, session_id = $7,
            is_player_known = $8, source_document = $9
        WHERE id = $1
        RETURNING id, campaign_id, event_date, event_time, date_precision,
                  description, entity_ids, session_id, is_player_known,
                  source_document, created_at, updated_at`

	var e models.TimelineEvent
	var returnedEntityIDs []int64
	err = db.QueryRow(ctx, query,
		id, eventDate, eventTime, datePrecision, description,
		pq.Array(entityIDs), sessionID, isPlayerKnown, sourceDocument,
	).Scan(
		&e.ID, &e.CampaignID, &e.EventDate, &e.EventTime, &e.DatePrecision,
		&e.Description, pq.Array(&returnedEntityIDs), &e.SessionID, &e.IsPlayerKnown,
		&e.SourceDocument, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update timeline event: %w", err)
	}
	e.EntityIDs = returnedEntityIDs

	return &e, nil
}

// DeleteTimelineEvent deletes a timeline event by ID.
func (db *DB) DeleteTimelineEvent(ctx context.Context, id int64) error {
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

// GetTimelineEventsForEntity retrieves all timeline events involving a specific entity.
func (db *DB) GetTimelineEventsForEntity(ctx context.Context, entityID int64) ([]models.TimelineEvent, error) {
	query := `
        SELECT id, campaign_id, event_date, event_time, date_precision,
               description, entity_ids, session_id, is_player_known,
               source_document, created_at, updated_at
        FROM timeline_events
        WHERE $1 = ANY(entity_ids)
        ORDER BY event_date ASC NULLS LAST, created_at ASC`

	rows, err := db.Query(ctx, query, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to query timeline events for entity: %w", err)
	}
	defer rows.Close()

	var events []models.TimelineEvent
	for rows.Next() {
		var e models.TimelineEvent
		var entityIDs []int64
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
