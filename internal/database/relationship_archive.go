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
)

// ArchiveRelationship atomically moves a relationship
// from the active table to the archive table using a
// single CTE statement. The eraID parameter is optional;
// if nil, the existing era_id on the relationship is
// preserved.
func (db *DB) ArchiveRelationship(
	ctx context.Context,
	relationshipID int64,
	eraID *int64,
) error {
	// Use a CTE to atomically delete and archive
	// in a single statement. The DELETE runs first
	// and RETURNING feeds the INSERT, so no row can
	// exist in both tables simultaneously.
	query := `
        WITH archived AS (
            DELETE FROM relationships
            WHERE id = $1
            RETURNING campaign_id, source_entity_id,
                      target_entity_id,
                      relationship_type_id,
                      era_id, tone, description,
                      strength, created_at
        )
        INSERT INTO relationship_archive
            (campaign_id, source_entity_id,
             target_entity_id, relationship_type_id,
             era_id, tone, description, strength,
             original_created_at)
        SELECT campaign_id, source_entity_id,
               target_entity_id, relationship_type_id,
               COALESCE($2, era_id), tone, description,
               strength, created_at
        FROM archived`

	tag, err := db.Pool.Exec(ctx, query,
		relationshipID, eraID)
	if err != nil {
		return fmt.Errorf(
			"failed to archive relationship: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf(
			"relationship %d not found", relationshipID)
	}

	return nil
}
