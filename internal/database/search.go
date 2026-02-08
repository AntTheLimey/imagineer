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

	"github.com/antonypegg/imagineer/internal/models"
)

// SearchCampaignContent performs hybrid vector+BM25 search across all
// vectorized campaign content by calling the search_campaign_content
// SQL function.
func (db *DB) SearchCampaignContent(
	ctx context.Context,
	campaignID int64,
	query string,
	limit int,
) ([]models.SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	rows, err := db.Query(ctx,
		`SELECT source_table, source_id, source_name,
                chunk_content, vector_score, combined_score
           FROM search_campaign_content($1, $2, $3)`,
		campaignID, query, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.SearchResult
	for rows.Next() {
		var r models.SearchResult
		if err := rows.Scan(
			&r.SourceTable, &r.SourceID, &r.SourceName,
			&r.ChunkContent, &r.VectorScore, &r.CombinedScore,
		); err != nil {
			return nil, err
		}
		results = append(results, r)
	}

	return results, rows.Err()
}

// IsVectorizationAvailable checks whether the pgedge_vectorizer
// extension is installed in the database.
func (db *DB) IsVectorizationAvailable(ctx context.Context) bool {
	var exists bool
	err := db.QueryRow(ctx,
		`SELECT EXISTS(
            SELECT 1 FROM pg_extension WHERE extname = 'pgedge_vectorizer'
        )`,
	).Scan(&exists)
	return err == nil && exists
}
