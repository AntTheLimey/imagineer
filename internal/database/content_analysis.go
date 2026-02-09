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
	"strings"

	"github.com/antonypegg/imagineer/internal/models"
	"github.com/jackc/pgx/v5"
)

// CreateAnalysisJob inserts a new content analysis job and returns the
// populated record including server-generated fields.
func (db *DB) CreateAnalysisJob(ctx context.Context, job *models.ContentAnalysisJob) (*models.ContentAnalysisJob, error) {
	query := `
		INSERT INTO content_analysis_jobs
			(campaign_id, source_table, source_id, source_field,
			 status, total_items, resolved_items)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, campaign_id, source_table, source_id, source_field,
		          status, total_items, resolved_items, created_at, updated_at`

	var j models.ContentAnalysisJob
	err := db.QueryRow(ctx, query,
		job.CampaignID, job.SourceTable, job.SourceID, job.SourceField,
		job.Status, job.TotalItems, job.ResolvedItems,
	).Scan(
		&j.ID, &j.CampaignID, &j.SourceTable, &j.SourceID, &j.SourceField,
		&j.Status, &j.TotalItems, &j.ResolvedItems, &j.CreatedAt, &j.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create analysis job: %w", err)
	}

	return &j, nil
}

// GetAnalysisJob retrieves a content analysis job by ID.
func (db *DB) GetAnalysisJob(ctx context.Context, id int64) (*models.ContentAnalysisJob, error) {
	query := `
		SELECT id, campaign_id, source_table, source_id, source_field,
		       status, total_items, resolved_items, created_at, updated_at
		FROM content_analysis_jobs
		WHERE id = $1`

	var j models.ContentAnalysisJob
	err := db.QueryRow(ctx, query, id).Scan(
		&j.ID, &j.CampaignID, &j.SourceTable, &j.SourceID, &j.SourceField,
		&j.Status, &j.TotalItems, &j.ResolvedItems, &j.CreatedAt, &j.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get analysis job: %w", err)
	}

	return &j, nil
}

// ListAnalysisJobsByCampaign retrieves all content analysis jobs for a
// campaign, ordered by most recent first.
func (db *DB) ListAnalysisJobsByCampaign(ctx context.Context, campaignID int64) ([]models.ContentAnalysisJob, error) {
	query := `
		SELECT id, campaign_id, source_table, source_id, source_field,
		       status, total_items, resolved_items, created_at, updated_at
		FROM content_analysis_jobs
		WHERE campaign_id = $1
		ORDER BY created_at DESC`

	rows, err := db.Query(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to list analysis jobs: %w", err)
	}
	defer rows.Close()

	return scanAnalysisJobs(rows)
}

// GetLatestAnalysisJob retrieves the most recent analysis job for a
// specific content source field.
func (db *DB) GetLatestAnalysisJob(ctx context.Context, campaignID int64, sourceTable string, sourceID int64, sourceField string) (*models.ContentAnalysisJob, error) {
	query := `
		SELECT id, campaign_id, source_table, source_id, source_field,
		       status, total_items, resolved_items, created_at, updated_at
		FROM content_analysis_jobs
		WHERE campaign_id = $1
		  AND source_table = $2
		  AND source_id = $3
		  AND source_field = $4
		ORDER BY created_at DESC
		LIMIT 1`

	var j models.ContentAnalysisJob
	err := db.QueryRow(ctx, query, campaignID, sourceTable, sourceID, sourceField).Scan(
		&j.ID, &j.CampaignID, &j.SourceTable, &j.SourceID, &j.SourceField,
		&j.Status, &j.TotalItems, &j.ResolvedItems, &j.CreatedAt, &j.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest analysis job: %w", err)
	}

	return &j, nil
}

// CreateAnalysisItems batch-inserts analysis items using a single query
// with multiple VALUES rows.
func (db *DB) CreateAnalysisItems(ctx context.Context, items []models.ContentAnalysisItem) error {
	if len(items) == 0 {
		return nil
	}

	const cols = 10 // number of columns per row
	valueStrings := make([]string, 0, len(items))
	args := make([]interface{}, 0, len(items)*cols)

	for i, item := range items {
		base := i * cols
		valueStrings = append(valueStrings, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5,
			base+6, base+7, base+8, base+9, base+10,
		))
		args = append(args,
			item.JobID, item.DetectionType, item.MatchedText,
			item.EntityID, item.Similarity, item.ContextSnippet,
			item.PositionStart, item.PositionEnd, item.Resolution,
			item.ResolvedEntityID,
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO content_analysis_items
			(job_id, detection_type, matched_text, entity_id, similarity,
			 context_snippet, position_start, position_end, resolution,
			 resolved_entity_id)
		VALUES %s`, strings.Join(valueStrings, ", "))

	return db.Exec(ctx, query, args...)
}

// ListAnalysisItemsByJob retrieves analysis items for a job, joining on
// entities to populate entity_name and entity_type. When resolution is
// non-empty, only items matching that resolution are returned. Results
// are ordered by position_start ascending.
func (db *DB) ListAnalysisItemsByJob(ctx context.Context, jobID int64, resolution string) ([]models.ContentAnalysisItem, error) {
	query := `
		SELECT i.id, i.job_id, i.detection_type, i.matched_text,
		       i.entity_id, i.similarity, i.context_snippet,
		       i.position_start, i.position_end, i.resolution,
		       i.resolved_entity_id, i.resolved_at, i.created_at,
		       e.name AS entity_name, e.entity_type
		FROM content_analysis_items i
		LEFT JOIN entities e ON i.entity_id = e.id
		WHERE i.job_id = $1`

	args := []interface{}{jobID}

	if resolution != "" {
		query += ` AND i.resolution = $2`
		args = append(args, resolution)
	}

	query += ` ORDER BY i.position_start ASC`

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list analysis items: %w", err)
	}
	defer rows.Close()

	return scanAnalysisItems(rows)
}

// ResolveAnalysisItem updates an analysis item's resolution status and
// optionally links it to an entity.
func (db *DB) ResolveAnalysisItem(ctx context.Context, itemID int64, resolution string, resolvedEntityID *int64) error {
	query := `
		UPDATE content_analysis_items
		SET resolution = $2, resolved_entity_id = $3, resolved_at = NOW()
		WHERE id = $1`

	return db.Exec(ctx, query, itemID, resolution, resolvedEntityID)
}

// UpdateJobResolvedCount recalculates and updates the resolved_items
// count for a content analysis job based on items that are no longer
// pending.
func (db *DB) UpdateJobResolvedCount(ctx context.Context, jobID int64) error {
	query := `
		UPDATE content_analysis_jobs
		SET resolved_items = (
			SELECT COUNT(*)
			FROM content_analysis_items
			WHERE job_id = $1 AND resolution != 'pending'
		)
		WHERE id = $1`

	return db.Exec(ctx, query, jobID)
}

// CountPendingAnalysisItems counts analysis items with resolution =
// 'pending' for a given campaign. When sourceTable and sourceID are
// provided (non-empty / non-zero), the count is scoped to that
// specific source; otherwise it returns the campaign-wide total.
func (db *DB) CountPendingAnalysisItems(ctx context.Context, campaignID int64, sourceTable string, sourceID int64) (int, error) {
	var count int
	var err error

	if sourceTable != "" && sourceID != 0 {
		query := `
			SELECT COUNT(*)
			FROM content_analysis_items i
			JOIN content_analysis_jobs j ON i.job_id = j.id
			WHERE j.campaign_id = $1
			  AND j.source_table = $2
			  AND j.source_id = $3
			  AND i.resolution = 'pending'`
		err = db.QueryRow(ctx, query, campaignID, sourceTable, sourceID).Scan(&count)
	} else {
		query := `
			SELECT COUNT(*)
			FROM content_analysis_items i
			JOIN content_analysis_jobs j ON i.job_id = j.id
			WHERE j.campaign_id = $1
			  AND i.resolution = 'pending'`
		err = db.QueryRow(ctx, query, campaignID).Scan(&count)
	}

	if err != nil {
		return 0, fmt.Errorf("failed to count pending analysis items: %w", err)
	}

	return count, nil
}

// DeleteAnalysisJobsForSource deletes all analysis jobs (and their items
// via CASCADE) for a specific content source field.
func (db *DB) DeleteAnalysisJobsForSource(ctx context.Context, campaignID int64, sourceTable string, sourceID int64, sourceField string) error {
	query := `
		DELETE FROM content_analysis_jobs
		WHERE campaign_id = $1
		  AND source_table = $2
		  AND source_id = $3
		  AND source_field = $4`

	return db.Exec(ctx, query, campaignID, sourceTable, sourceID, sourceField)
}

// scanAnalysisJobs scans multiple content analysis job rows.
func scanAnalysisJobs(rows pgx.Rows) ([]models.ContentAnalysisJob, error) {
	var jobs []models.ContentAnalysisJob
	for rows.Next() {
		var j models.ContentAnalysisJob
		err := rows.Scan(
			&j.ID, &j.CampaignID, &j.SourceTable, &j.SourceID, &j.SourceField,
			&j.Status, &j.TotalItems, &j.ResolvedItems, &j.CreatedAt, &j.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan analysis job: %w", err)
		}
		jobs = append(jobs, j)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating analysis jobs: %w", err)
	}

	return jobs, nil
}

// scanAnalysisItems scans multiple content analysis item rows including
// joined entity fields.
func scanAnalysisItems(rows pgx.Rows) ([]models.ContentAnalysisItem, error) {
	var items []models.ContentAnalysisItem
	for rows.Next() {
		var item models.ContentAnalysisItem
		err := rows.Scan(
			&item.ID, &item.JobID, &item.DetectionType, &item.MatchedText,
			&item.EntityID, &item.Similarity, &item.ContextSnippet,
			&item.PositionStart, &item.PositionEnd, &item.Resolution,
			&item.ResolvedEntityID, &item.ResolvedAt, &item.CreatedAt,
			&item.EntityName, &item.EntityType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan analysis item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating analysis items: %w", err)
	}

	return items, nil
}
