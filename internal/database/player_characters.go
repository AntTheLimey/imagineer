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
	"errors"
	"fmt"

	"github.com/antonypegg/imagineer/internal/models"
	"github.com/jackc/pgx/v5"
)

// ListPlayerCharacters retrieves all player characters for a campaign.
func (db *DB) ListPlayerCharacters(ctx context.Context, campaignID int64) ([]models.PlayerCharacter, error) {
	query := `
        SELECT id, campaign_id, entity_id, character_name, player_name,
               description, background, created_at, updated_at
        FROM player_characters
        WHERE campaign_id = $1
        ORDER BY character_name`

	rows, err := db.Query(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to query player characters: %w", err)
	}
	defer rows.Close()

	return scanPlayerCharacters(rows)
}

// GetPlayerCharacter retrieves a player character by ID.
func (db *DB) GetPlayerCharacter(ctx context.Context, id int64) (*models.PlayerCharacter, error) {
	query := `
        SELECT id, campaign_id, entity_id, character_name, player_name,
               description, background, created_at, updated_at
        FROM player_characters
        WHERE id = $1`

	var pc models.PlayerCharacter
	err := db.QueryRow(ctx, query, id).Scan(
		&pc.ID, &pc.CampaignID, &pc.EntityID, &pc.CharacterName, &pc.PlayerName,
		&pc.Description, &pc.Background, &pc.CreatedAt, &pc.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("player character not found")
		}
		return nil, fmt.Errorf("failed to get player character: %w", err)
	}

	return &pc, nil
}

// CreatePlayerCharacter creates a new player character in a campaign.
func (db *DB) CreatePlayerCharacter(ctx context.Context, campaignID int64, req models.CreatePlayerCharacterRequest) (*models.PlayerCharacter, error) {
	query := `
        INSERT INTO player_characters (campaign_id, entity_id, character_name,
                                        player_name, description, background)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, campaign_id, entity_id, character_name, player_name,
                  description, background, created_at, updated_at`

	var pc models.PlayerCharacter
	err := db.QueryRow(ctx, query,
		campaignID, req.EntityID, req.CharacterName,
		req.PlayerName, req.Description, req.Background,
	).Scan(
		&pc.ID, &pc.CampaignID, &pc.EntityID, &pc.CharacterName, &pc.PlayerName,
		&pc.Description, &pc.Background, &pc.CreatedAt, &pc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create player character: %w", err)
	}

	return &pc, nil
}

// UpdatePlayerCharacter updates an existing player character using COALESCE.
// Uses database-side COALESCE to atomically merge updates, preventing race conditions.
// NULL values in the request preserve existing values in the database.
func (db *DB) UpdatePlayerCharacter(ctx context.Context, id int64, req models.UpdatePlayerCharacterRequest) (*models.PlayerCharacter, error) {
	query := `
        UPDATE player_characters
        SET entity_id = COALESCE($2, entity_id),
            character_name = COALESCE($3, character_name),
            player_name = COALESCE($4, player_name),
            description = COALESCE($5, description),
            background = COALESCE($6, background)
        WHERE id = $1
        RETURNING id, campaign_id, entity_id, character_name, player_name,
                  description, background, created_at, updated_at`

	var pc models.PlayerCharacter
	err := db.QueryRow(ctx, query,
		id, req.EntityID, req.CharacterName, req.PlayerName, req.Description, req.Background,
	).Scan(
		&pc.ID, &pc.CampaignID, &pc.EntityID, &pc.CharacterName, &pc.PlayerName,
		&pc.Description, &pc.Background, &pc.CreatedAt, &pc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("player character not found")
		}
		return nil, fmt.Errorf("failed to update player character: %w", err)
	}

	return &pc, nil
}

// DeletePlayerCharacter deletes a player character by ID.
func (db *DB) DeletePlayerCharacter(ctx context.Context, id int64) error {
	query := `DELETE FROM player_characters WHERE id = $1`
	result, err := db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete player character: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("player character not found")
	}

	return nil
}

// scanPlayerCharacters scans multiple player character rows.
func scanPlayerCharacters(rows pgx.Rows) ([]models.PlayerCharacter, error) {
	var characters []models.PlayerCharacter
	for rows.Next() {
		var pc models.PlayerCharacter
		err := rows.Scan(
			&pc.ID, &pc.CampaignID, &pc.EntityID, &pc.CharacterName, &pc.PlayerName,
			&pc.Description, &pc.Background, &pc.CreatedAt, &pc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan player character: %w", err)
		}
		characters = append(characters, pc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating player characters: %w", err)
	}

	return characters, nil
}
