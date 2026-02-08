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

// ResolveEntityByName performs a fuzzy name match against entities in a
// campaign using the pg_trgm similarity operator (%). It returns up to
// limit results ordered by descending similarity score.
func (db *DB) ResolveEntityByName(
	ctx context.Context,
	campaignID int64,
	name string,
	limit int,
) ([]models.EntityResolveResult, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 20 {
		limit = 20
	}

	query := `
		SELECT id, name, entity_type, similarity(name, $2) AS similarity
		FROM entities
		WHERE campaign_id = $1 AND name % $2
		ORDER BY similarity DESC
		LIMIT $3`

	rows, err := db.Query(ctx, query, campaignID, name, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve entity by name: %w", err)
	}
	defer rows.Close()

	var results []models.EntityResolveResult
	for rows.Next() {
		var r models.EntityResolveResult
		if err := rows.Scan(&r.ID, &r.Name, &r.EntityType, &r.Similarity); err != nil {
			return nil, fmt.Errorf("failed to scan entity resolve result: %w", err)
		}
		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating entity resolve results: %w", err)
	}

	return results, nil
}
