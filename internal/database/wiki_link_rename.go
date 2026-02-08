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

	"github.com/jackc/pgx/v5"
)

// escapeLikePattern escapes SQL LIKE wildcard characters (% and _)
// so they are treated as literals in LIKE expressions.
func escapeLikePattern(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// PropagateEntityRename updates all wiki links throughout campaign content
// when an entity is renamed. It replaces both [[Old Name]] and [[Old Name|
// patterns with the new name. This function must be called within an
// existing transaction.
func PropagateEntityRename(ctx context.Context, tx pgx.Tx, campaignID int64, oldName, newName string) (int64, error) {
	oldExact := "[[" + oldName + "]]"
	newExact := "[[" + newName + "]]"
	oldPiped := "[[" + oldName + "|"
	newPiped := "[[" + newName + "|"

	// Escape wildcards for LIKE patterns so that % and _ in entity
	// names are matched literally rather than as SQL wildcards.
	oldExactLike := "%" + escapeLikePattern("[["+oldName+"]]") + "%"
	oldPipedLike := "%" + escapeLikePattern("[["+oldName+"|") + "%"

	var totalUpdated int64

	// 1. entities: description, gm_notes
	result, err := tx.Exec(ctx, `
		UPDATE entities SET
			description = replace(replace(description, $2, $3), $4, $5),
			gm_notes = replace(replace(gm_notes, $2, $3), $4, $5)
		WHERE campaign_id = $1
			AND (description LIKE $6 ESCAPE '\' OR description LIKE $7 ESCAPE '\'
				OR gm_notes LIKE $6 ESCAPE '\' OR gm_notes LIKE $7 ESCAPE '\')`,
		campaignID, oldExact, newExact, oldPiped, newPiped, oldExactLike, oldPipedLike,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update entities: %w", err)
	}
	totalUpdated += result.RowsAffected()

	// 2. chapters: overview
	result, err = tx.Exec(ctx, `
		UPDATE chapters SET
			overview = replace(replace(overview, $2, $3), $4, $5)
		WHERE campaign_id = $1
			AND (overview LIKE $6 ESCAPE '\' OR overview LIKE $7 ESCAPE '\')`,
		campaignID, oldExact, newExact, oldPiped, newPiped, oldExactLike, oldPipedLike,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update chapters: %w", err)
	}
	totalUpdated += result.RowsAffected()

	// 3. sessions: prep_notes, actual_notes
	result, err = tx.Exec(ctx, `
		UPDATE sessions SET
			prep_notes = replace(replace(prep_notes, $2, $3), $4, $5),
			actual_notes = replace(replace(actual_notes, $2, $3), $4, $5)
		WHERE campaign_id = $1
			AND (prep_notes LIKE $6 ESCAPE '\' OR prep_notes LIKE $7 ESCAPE '\'
				OR actual_notes LIKE $6 ESCAPE '\' OR actual_notes LIKE $7 ESCAPE '\')`,
		campaignID, oldExact, newExact, oldPiped, newPiped, oldExactLike, oldPipedLike,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update sessions: %w", err)
	}
	totalUpdated += result.RowsAffected()

	// 4. campaigns: description
	result, err = tx.Exec(ctx, `
		UPDATE campaigns SET
			description = replace(replace(description, $2, $3), $4, $5)
		WHERE id = $1
			AND (description LIKE $6 ESCAPE '\' OR description LIKE $7 ESCAPE '\')`,
		campaignID, oldExact, newExact, oldPiped, newPiped, oldExactLike, oldPipedLike,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update campaigns: %w", err)
	}
	totalUpdated += result.RowsAffected()

	// 5. campaign_memories: content
	result, err = tx.Exec(ctx, `
		UPDATE campaign_memories SET
			content = replace(replace(content, $2, $3), $4, $5)
		WHERE campaign_id = $1
			AND (content LIKE $6 ESCAPE '\' OR content LIKE $7 ESCAPE '\')`,
		campaignID, oldExact, newExact, oldPiped, newPiped, oldExactLike, oldPipedLike,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update campaign_memories: %w", err)
	}
	totalUpdated += result.RowsAffected()

	// 6. timeline_events: description
	result, err = tx.Exec(ctx, `
		UPDATE timeline_events SET
			description = replace(replace(description, $2, $3), $4, $5)
		WHERE campaign_id = $1
			AND (description LIKE $6 ESCAPE '\' OR description LIKE $7 ESCAPE '\')`,
		campaignID, oldExact, newExact, oldPiped, newPiped, oldExactLike, oldPipedLike,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update timeline_events: %w", err)
	}
	totalUpdated += result.RowsAffected()

	// 7. player_characters: description, background
	result, err = tx.Exec(ctx, `
		UPDATE player_characters SET
			description = replace(replace(description, $2, $3), $4, $5),
			background = replace(replace(background, $2, $3), $4, $5)
		WHERE campaign_id = $1
			AND (description LIKE $6 ESCAPE '\' OR description LIKE $7 ESCAPE '\'
				OR background LIKE $6 ESCAPE '\' OR background LIKE $7 ESCAPE '\')`,
		campaignID, oldExact, newExact, oldPiped, newPiped, oldExactLike, oldPipedLike,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update player_characters: %w", err)
	}
	totalUpdated += result.RowsAffected()

	// 8. relationships: description
	result, err = tx.Exec(ctx, `
		UPDATE relationships SET
			description = replace(replace(description, $2, $3), $4, $5)
		WHERE campaign_id = $1
			AND (description LIKE $6 ESCAPE '\' OR description LIKE $7 ESCAPE '\')`,
		campaignID, oldExact, newExact, oldPiped, newPiped, oldExactLike, oldPipedLike,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update relationships: %w", err)
	}
	totalUpdated += result.RowsAffected()

	return totalUpdated, nil
}
