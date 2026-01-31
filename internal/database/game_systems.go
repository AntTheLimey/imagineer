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
)

// ListGameSystems retrieves all game systems.
func (db *DB) ListGameSystems(ctx context.Context) ([]models.GameSystem, error) {
	query := `
        SELECT id, name, code, attribute_schema, skill_schema,
               character_sheet_template, dice_conventions, created_at
        FROM game_systems
        ORDER BY name`

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query game systems: %w", err)
	}
	defer rows.Close()

	var systems []models.GameSystem
	for rows.Next() {
		var gs models.GameSystem
		err := rows.Scan(
			&gs.ID, &gs.Name, &gs.Code, &gs.AttributeSchema,
			&gs.SkillSchema, &gs.CharacterSheetTemplate,
			&gs.DiceConventions, &gs.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game system: %w", err)
		}
		systems = append(systems, gs)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating game systems: %w", err)
	}

	return systems, nil
}

// GetGameSystem retrieves a game system by ID.
func (db *DB) GetGameSystem(ctx context.Context, id uuid.UUID) (*models.GameSystem, error) {
	query := `
        SELECT id, name, code, attribute_schema, skill_schema,
               character_sheet_template, dice_conventions, created_at
        FROM game_systems
        WHERE id = $1`

	var gs models.GameSystem
	err := db.QueryRow(ctx, query, id).Scan(
		&gs.ID, &gs.Name, &gs.Code, &gs.AttributeSchema,
		&gs.SkillSchema, &gs.CharacterSheetTemplate,
		&gs.DiceConventions, &gs.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get game system: %w", err)
	}

	return &gs, nil
}

// GetGameSystemByCode retrieves a game system by its code.
func (db *DB) GetGameSystemByCode(ctx context.Context, code string) (*models.GameSystem, error) {
	query := `
        SELECT id, name, code, attribute_schema, skill_schema,
               character_sheet_template, dice_conventions, created_at
        FROM game_systems
        WHERE code = $1`

	var gs models.GameSystem
	err := db.QueryRow(ctx, query, code).Scan(
		&gs.ID, &gs.Name, &gs.Code, &gs.AttributeSchema,
		&gs.SkillSchema, &gs.CharacterSheetTemplate,
		&gs.DiceConventions, &gs.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get game system by code: %w", err)
	}

	return &gs, nil
}
