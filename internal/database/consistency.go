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
	"time"

	"github.com/antonypegg/imagineer/internal/models"
)

// OrphanedEntity represents an entity with no relationships or timeline references.
type OrphanedEntity struct {
	ID         int64             `json:"id"`
	Name       string            `json:"name"`
	EntityType models.EntityType `json:"entityType"`
}

// DuplicateNamePair represents two entities with similar names.
type DuplicateNamePair struct {
	EntityID1  int64   `json:"entityId1"`
	Name1      string  `json:"name1"`
	EntityID2  int64   `json:"entityId2"`
	Name2      string  `json:"name2"`
	Similarity float64 `json:"similarity"`
}

// TimelineConflict represents an entity appearing in multiple events at the same time.
type TimelineConflict struct {
	EntityID   int64     `json:"entityId"`
	EntityName string    `json:"entityName"`
	EventDate  time.Time `json:"eventDate"`
	EventCount int       `json:"eventCount"`
	EventIDs   []int64   `json:"eventIds"`
}

// InvalidReference represents a relationship pointing to a non-existent entity.
type InvalidReference struct {
	RelationshipID  int64  `json:"relationshipId"`
	MissingEntityID int64  `json:"missingEntityId"`
	ReferenceType   string `json:"referenceType"` // "source" or "target"
}

// EmptySession represents a completed session with no entity references.
type EmptySession struct {
	ID            int64 `json:"id"`
	SessionNumber int   `json:"sessionNumber"`
}

// FindOrphanedEntities retrieves entities with no relationships or timeline references.
func (db *DB) FindOrphanedEntities(ctx context.Context, campaignID int64, entityTypeFilter *string) ([]OrphanedEntity, error) {
	query := `
		SELECT e.id, e.name, e.entity_type
		FROM entities e
		WHERE e.campaign_id = $1
		  AND NOT EXISTS (
			SELECT 1 FROM relationships r
			WHERE r.source_entity_id = e.id OR r.target_entity_id = e.id
		  )
		  AND NOT EXISTS (
			SELECT 1 FROM timeline_events t
			WHERE e.id = ANY(t.entity_ids)
		  )
		  AND ($2::text IS NULL OR e.entity_type = $2)
		ORDER BY e.name`

	rows, err := db.Query(ctx, query, campaignID, entityTypeFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to find orphaned entities: %w", err)
	}
	defer rows.Close()

	var orphans []OrphanedEntity
	for rows.Next() {
		var o OrphanedEntity
		if err := rows.Scan(&o.ID, &o.Name, &o.EntityType); err != nil {
			return nil, fmt.Errorf("failed to scan orphaned entity: %w", err)
		}
		orphans = append(orphans, o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating orphaned entities: %w", err)
	}

	return orphans, nil
}

// FindDuplicateNames retrieves pairs of entities with similar names using pg_trgm.
// The threshold parameter specifies the minimum similarity score (0.0 to 1.0).
func (db *DB) FindDuplicateNames(ctx context.Context, campaignID int64, threshold float64) ([]DuplicateNamePair, error) {
	query := `
		SELECT
			e1.id AS entity_id1,
			e1.name AS name1,
			e2.id AS entity_id2,
			e2.name AS name2,
			similarity(e1.name, e2.name) AS sim
		FROM entities e1
		JOIN entities e2 ON e1.campaign_id = e2.campaign_id
			AND e1.id < e2.id
			AND similarity(e1.name, e2.name) > $2
		WHERE e1.campaign_id = $1
		ORDER BY sim DESC`

	rows, err := db.Query(ctx, query, campaignID, threshold)
	if err != nil {
		return nil, fmt.Errorf("failed to find duplicate names: %w", err)
	}
	defer rows.Close()

	var duplicates []DuplicateNamePair
	for rows.Next() {
		var d DuplicateNamePair
		if err := rows.Scan(&d.EntityID1, &d.Name1, &d.EntityID2, &d.Name2, &d.Similarity); err != nil {
			return nil, fmt.Errorf("failed to scan duplicate name pair: %w", err)
		}
		duplicates = append(duplicates, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating duplicate names: %w", err)
	}

	return duplicates, nil
}

// FindTimelineConflicts retrieves entities appearing in multiple events at the same date.
func (db *DB) FindTimelineConflicts(ctx context.Context, campaignID int64) ([]TimelineConflict, error) {
	// First, find entities with potential conflicts
	query := `
		WITH entity_event_dates AS (
			SELECT
				unnest(t.entity_ids) AS entity_id,
				t.id AS event_id,
				t.event_date
			FROM timeline_events t
			WHERE t.campaign_id = $1
			  AND t.event_date IS NOT NULL
			  AND t.date_precision = 'exact'
		),
		conflicts AS (
			SELECT
				eed.entity_id,
				eed.event_date,
				COUNT(*) AS event_count,
				array_agg(eed.event_id) AS event_ids
			FROM entity_event_dates eed
			GROUP BY eed.entity_id, eed.event_date
			HAVING COUNT(*) > 1
		)
		SELECT
			c.entity_id,
			e.name AS entity_name,
			c.event_date,
			c.event_count,
			c.event_ids
		FROM conflicts c
		JOIN entities e ON c.entity_id = e.id
		ORDER BY c.event_date, e.name`

	rows, err := db.Query(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to find timeline conflicts: %w", err)
	}
	defer rows.Close()

	var conflicts []TimelineConflict
	for rows.Next() {
		var c TimelineConflict
		if err := rows.Scan(&c.EntityID, &c.EntityName, &c.EventDate, &c.EventCount, &c.EventIDs); err != nil {
			return nil, fmt.Errorf("failed to scan timeline conflict: %w", err)
		}
		conflicts = append(conflicts, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating timeline conflicts: %w", err)
	}

	return conflicts, nil
}

// FindMissingRelationshipTargets retrieves relationships pointing to non-existent entities.
func (db *DB) FindMissingRelationshipTargets(ctx context.Context, campaignID int64) ([]InvalidReference, error) {
	query := `
		SELECT r.id, r.source_entity_id, 'source'
		FROM relationships r
		WHERE r.campaign_id = $1
		  AND NOT EXISTS (SELECT 1 FROM entities e WHERE e.id = r.source_entity_id)
		UNION ALL
		SELECT r.id, r.target_entity_id, 'target'
		FROM relationships r
		WHERE r.campaign_id = $1
		  AND NOT EXISTS (SELECT 1 FROM entities e WHERE e.id = r.target_entity_id)`

	rows, err := db.Query(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to find missing relationship targets: %w", err)
	}
	defer rows.Close()

	var invalidRefs []InvalidReference
	for rows.Next() {
		var ref InvalidReference
		if err := rows.Scan(&ref.RelationshipID, &ref.MissingEntityID, &ref.ReferenceType); err != nil {
			return nil, fmt.Errorf("failed to scan invalid reference: %w", err)
		}
		invalidRefs = append(invalidRefs, ref)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating invalid references: %w", err)
	}

	return invalidRefs, nil
}

// FindSessionsWithoutEntities retrieves completed sessions with no entity discoveries.
func (db *DB) FindSessionsWithoutEntities(ctx context.Context, campaignID int64) ([]EmptySession, error) {
	query := `
		SELECT s.id, COALESCE(s.session_number, 0) as session_number
		FROM sessions s
		WHERE s.campaign_id = $1
		  AND s.status = 'COMPLETED'
		  AND NOT EXISTS (
			SELECT 1 FROM entities e
			WHERE e.discovered_session = s.id
		  )
		  AND (s.discoveries IS NULL OR s.discoveries = '[]'::jsonb OR s.discoveries = 'null'::jsonb)
		ORDER BY s.session_number`

	rows, err := db.Query(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to find sessions without entities: %w", err)
	}
	defer rows.Close()

	var sessions []EmptySession
	for rows.Next() {
		var s EmptySession
		if err := rows.Scan(&s.ID, &s.SessionNumber); err != nil {
			return nil, fmt.Errorf("failed to scan empty session: %w", err)
		}
		sessions = append(sessions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating empty sessions: %w", err)
	}

	return sessions, nil
}
