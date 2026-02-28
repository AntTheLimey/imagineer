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

// ArchiveRelationship copies a relationship to the
// archive table and deletes it from the active table.
// The eraID parameter is optional; if nil, the
// relationship is archived without an era reference.
func (db *DB) ArchiveRelationship(
	ctx context.Context,
	relationshipID int64,
	eraID *int64,
) error {
	// Copy to archive in a single statement.
	archiveQuery := `
        INSERT INTO relationship_archive
            (campaign_id, source_entity_id,
             target_entity_id, relationship_type_id,
             era_id, tone, description, strength,
             original_created_at)
        SELECT campaign_id, source_entity_id,
               target_entity_id, relationship_type_id,
               COALESCE($2, era_id), tone, description,
               strength, created_at
        FROM relationships
        WHERE id = $1`

	tag, err := db.Pool.Exec(ctx, archiveQuery,
		relationshipID, eraID)
	if err != nil {
		return fmt.Errorf(
			"failed to archive relationship: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf(
			"relationship %d not found", relationshipID)
	}

	// Delete from active table.
	deleteQuery := `
        DELETE FROM relationships WHERE id = $1`
	_, err = db.Pool.Exec(ctx, deleteQuery,
		relationshipID)
	if err != nil {
		return fmt.Errorf(
			"failed to delete archived relationship: %w",
			err)
	}

	return nil
}
