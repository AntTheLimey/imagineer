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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/antonypegg/imagineer/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ListCampaigns retrieves all campaigns. This is deprecated in favor of
// ListCampaignsByOwner for user-scoped access.
func (db *DB) ListCampaigns(ctx context.Context) ([]models.Campaign, error) {
	query := `
        SELECT c.id, c.name, c.system_id, c.owner_id, c.description, c.settings,
               c.genre, c.image_style_prompt, c.created_at, c.updated_at,
               gs.id, gs.name, gs.code
        FROM campaigns c
        LEFT JOIN game_systems gs ON c.system_id = gs.id
        ORDER BY c.updated_at DESC`

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query campaigns: %w", err)
	}
	defer rows.Close()

	var campaigns []models.Campaign
	for rows.Next() {
		var c models.Campaign
		var gsID *uuid.UUID
		var gsName, gsCode *string
		var genre *string

		err := rows.Scan(
			&c.ID, &c.Name, &c.SystemID, &c.OwnerID, &c.Description, &c.Settings,
			&genre, &c.ImageStylePrompt, &c.CreatedAt, &c.UpdatedAt,
			&gsID, &gsName, &gsCode,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan campaign: %w", err)
		}

		if genre != nil {
			g := models.CampaignGenre(*genre)
			c.Genre = &g
		}

		if gsID != nil {
			c.System = &models.GameSystem{
				ID:   *gsID,
				Name: *gsName,
				Code: *gsCode,
			}
		}

		campaigns = append(campaigns, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating campaigns: %w", err)
	}

	return campaigns, nil
}

// ListCampaignsByOwner retrieves all campaigns owned by a specific user.
// This is the primary method for user-scoped campaign listing.
func (db *DB) ListCampaignsByOwner(ctx context.Context, ownerID uuid.UUID) ([]models.Campaign, error) {
	query := `
        SELECT c.id, c.name, c.system_id, c.owner_id, c.description, c.settings,
               c.genre, c.image_style_prompt, c.created_at, c.updated_at,
               gs.id, gs.name, gs.code
        FROM campaigns c
        LEFT JOIN game_systems gs ON c.system_id = gs.id
        WHERE c.owner_id = $1
        ORDER BY c.updated_at DESC`

	rows, err := db.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query campaigns by owner: %w", err)
	}
	defer rows.Close()

	var campaigns []models.Campaign
	for rows.Next() {
		var c models.Campaign
		var gsID *uuid.UUID
		var gsName, gsCode *string
		var genre *string

		err := rows.Scan(
			&c.ID, &c.Name, &c.SystemID, &c.OwnerID, &c.Description, &c.Settings,
			&genre, &c.ImageStylePrompt, &c.CreatedAt, &c.UpdatedAt,
			&gsID, &gsName, &gsCode,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan campaign: %w", err)
		}

		if genre != nil {
			g := models.CampaignGenre(*genre)
			c.Genre = &g
		}

		if gsID != nil {
			c.System = &models.GameSystem{
				ID:   *gsID,
				Name: *gsName,
				Code: *gsCode,
			}
		}

		campaigns = append(campaigns, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating campaigns: %w", err)
	}

	return campaigns, nil
}

// GetCampaign retrieves a campaign by ID without ownership verification.
// For user-scoped access, use GetCampaignByOwner instead.
func (db *DB) GetCampaign(ctx context.Context, id uuid.UUID) (*models.Campaign, error) {
	query := `
        SELECT c.id, c.name, c.system_id, c.owner_id, c.description, c.settings,
               c.genre, c.image_style_prompt, c.created_at, c.updated_at,
               gs.id, gs.name, gs.code, gs.attribute_schema, gs.skill_schema,
               gs.character_sheet_template, gs.dice_conventions, gs.created_at
        FROM campaigns c
        LEFT JOIN game_systems gs ON c.system_id = gs.id
        WHERE c.id = $1`

	var c models.Campaign
	var gsID *uuid.UUID
	var gsName, gsCode *string
	var gsAttrSchema, gsSkillSchema, gsCharSheet, gsDice []byte
	var gsCreatedAt *interface{}
	var genre *string

	err := db.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.Name, &c.SystemID, &c.OwnerID, &c.Description, &c.Settings,
		&genre, &c.ImageStylePrompt, &c.CreatedAt, &c.UpdatedAt,
		&gsID, &gsName, &gsCode, &gsAttrSchema, &gsSkillSchema,
		&gsCharSheet, &gsDice, &gsCreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign: %w", err)
	}

	if genre != nil {
		g := models.CampaignGenre(*genre)
		c.Genre = &g
	}

	if gsID != nil {
		c.System = &models.GameSystem{
			ID:                     *gsID,
			Name:                   *gsName,
			Code:                   *gsCode,
			AttributeSchema:        gsAttrSchema,
			SkillSchema:            gsSkillSchema,
			CharacterSheetTemplate: gsCharSheet,
			DiceConventions:        gsDice,
		}
	}

	return &c, nil
}

// GetCampaignByOwner retrieves a campaign by ID and verifies ownership.
// Returns an error if the campaign doesn't exist or doesn't belong to the owner.
// This is the primary method for user-scoped campaign access.
func (db *DB) GetCampaignByOwner(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) (*models.Campaign, error) {
	query := `
        SELECT c.id, c.name, c.system_id, c.owner_id, c.description, c.settings,
               c.genre, c.image_style_prompt, c.created_at, c.updated_at,
               gs.id, gs.name, gs.code, gs.attribute_schema, gs.skill_schema,
               gs.character_sheet_template, gs.dice_conventions, gs.created_at
        FROM campaigns c
        LEFT JOIN game_systems gs ON c.system_id = gs.id
        WHERE c.id = $1 AND c.owner_id = $2`

	var c models.Campaign
	var gsID *uuid.UUID
	var gsName, gsCode *string
	var gsAttrSchema, gsSkillSchema, gsCharSheet, gsDice []byte
	var gsCreatedAt *interface{}
	var genre *string

	err := db.QueryRow(ctx, query, id, ownerID).Scan(
		&c.ID, &c.Name, &c.SystemID, &c.OwnerID, &c.Description, &c.Settings,
		&genre, &c.ImageStylePrompt, &c.CreatedAt, &c.UpdatedAt,
		&gsID, &gsName, &gsCode, &gsAttrSchema, &gsSkillSchema,
		&gsCharSheet, &gsDice, &gsCreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign: %w", err)
	}

	if genre != nil {
		g := models.CampaignGenre(*genre)
		c.Genre = &g
	}

	if gsID != nil {
		c.System = &models.GameSystem{
			ID:                     *gsID,
			Name:                   *gsName,
			Code:                   *gsCode,
			AttributeSchema:        gsAttrSchema,
			SkillSchema:            gsSkillSchema,
			CharacterSheetTemplate: gsCharSheet,
			DiceConventions:        gsDice,
		}
	}

	return &c, nil
}

// CreateCampaign creates a new campaign without an owner.
// For user-scoped creation, use CreateCampaignWithOwner instead.
func (db *DB) CreateCampaign(ctx context.Context, req models.CreateCampaignRequest) (*models.Campaign, error) {
	id := uuid.New()

	settings := req.Settings
	if settings == nil {
		settings = json.RawMessage("{}")
	}

	// Convert genre to string pointer for database
	var genreStr *string
	if req.Genre != nil {
		s := string(*req.Genre)
		genreStr = &s
	}

	query := `
        INSERT INTO campaigns (id, name, system_id, description, settings, genre, image_style_prompt)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, name, system_id, owner_id, description, settings, genre, image_style_prompt, created_at, updated_at`

	var c models.Campaign
	var retGenre *string
	err := db.QueryRow(ctx, query, id, req.Name, req.SystemID, req.Description, settings, genreStr, req.ImageStylePrompt).Scan(
		&c.ID, &c.Name, &c.SystemID, &c.OwnerID, &c.Description, &c.Settings,
		&retGenre, &c.ImageStylePrompt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create campaign: %w", err)
	}

	if retGenre != nil {
		g := models.CampaignGenre(*retGenre)
		c.Genre = &g
	}

	return &c, nil
}

// CreateCampaignWithOwner creates a new campaign with an owner.
// This is the primary method for user-scoped campaign creation.
func (db *DB) CreateCampaignWithOwner(ctx context.Context, req models.CreateCampaignRequest, ownerID uuid.UUID) (*models.Campaign, error) {
	id := uuid.New()

	settings := req.Settings
	if settings == nil {
		settings = json.RawMessage("{}")
	}

	// Convert genre to string pointer for database
	var genreStr *string
	if req.Genre != nil {
		s := string(*req.Genre)
		genreStr = &s
	}

	query := `
        INSERT INTO campaigns (id, name, system_id, owner_id, description, settings, genre, image_style_prompt)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, name, system_id, owner_id, description, settings, genre, image_style_prompt, created_at, updated_at`

	var c models.Campaign
	var retGenre *string
	err := db.QueryRow(ctx, query, id, req.Name, req.SystemID, ownerID, req.Description, settings, genreStr, req.ImageStylePrompt).Scan(
		&c.ID, &c.Name, &c.SystemID, &c.OwnerID, &c.Description, &c.Settings,
		&retGenre, &c.ImageStylePrompt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create campaign: %w", err)
	}

	if retGenre != nil {
		g := models.CampaignGenre(*retGenre)
		c.Genre = &g
	}

	return &c, nil
}

// UpdateCampaign updates an existing campaign without ownership verification.
// For user-scoped updates, use UpdateCampaignByOwner instead.
func (db *DB) UpdateCampaign(ctx context.Context, id uuid.UUID, req models.UpdateCampaignRequest) (*models.Campaign, error) {
	// First get the existing campaign
	existing, err := db.GetCampaign(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	name := existing.Name
	if req.Name != nil {
		name = *req.Name
	}

	systemID := existing.SystemID
	if req.SystemID != nil {
		systemID = req.SystemID
	}

	description := existing.Description
	if req.Description != nil {
		description = req.Description
	}

	settings := existing.Settings
	if req.Settings != nil {
		settings = req.Settings
	}

	genre := existing.Genre
	if req.Genre != nil {
		genre = req.Genre
	}

	imageStylePrompt := existing.ImageStylePrompt
	if req.ImageStylePrompt != nil {
		imageStylePrompt = req.ImageStylePrompt
	}

	// Convert genre to string pointer for database
	var genreStr *string
	if genre != nil {
		s := string(*genre)
		genreStr = &s
	}

	query := `
        UPDATE campaigns
        SET name = $2, system_id = $3, description = $4, settings = $5, genre = $6, image_style_prompt = $7
        WHERE id = $1
        RETURNING id, name, system_id, owner_id, description, settings, genre, image_style_prompt, created_at, updated_at`

	var c models.Campaign
	var retGenre *string
	err = db.QueryRow(ctx, query, id, name, systemID, description, settings, genreStr, imageStylePrompt).Scan(
		&c.ID, &c.Name, &c.SystemID, &c.OwnerID, &c.Description, &c.Settings,
		&retGenre, &c.ImageStylePrompt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update campaign: %w", err)
	}

	if retGenre != nil {
		g := models.CampaignGenre(*retGenre)
		c.Genre = &g
	}

	return &c, nil
}

// UpdateCampaignByOwner updates an existing campaign with ownership verification.
// Returns an error if the campaign doesn't exist or doesn't belong to the owner.
// This is the primary method for user-scoped campaign updates.
func (db *DB) UpdateCampaignByOwner(ctx context.Context, id uuid.UUID, ownerID uuid.UUID, req models.UpdateCampaignRequest) (*models.Campaign, error) {
	// First verify the campaign exists and belongs to the owner
	existing, err := db.GetCampaignByOwner(ctx, id, ownerID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	name := existing.Name
	if req.Name != nil {
		name = *req.Name
	}

	systemID := existing.SystemID
	if req.SystemID != nil {
		systemID = req.SystemID
	}

	description := existing.Description
	if req.Description != nil {
		description = req.Description
	}

	settings := existing.Settings
	if req.Settings != nil {
		settings = req.Settings
	}

	genre := existing.Genre
	if req.Genre != nil {
		genre = req.Genre
	}

	imageStylePrompt := existing.ImageStylePrompt
	if req.ImageStylePrompt != nil {
		imageStylePrompt = req.ImageStylePrompt
	}

	// Convert genre to string pointer for database
	var genreStr *string
	if genre != nil {
		s := string(*genre)
		genreStr = &s
	}

	query := `
        UPDATE campaigns
        SET name = $2, system_id = $3, description = $4, settings = $5, genre = $6, image_style_prompt = $7
        WHERE id = $1 AND owner_id = $8
        RETURNING id, name, system_id, owner_id, description, settings, genre, image_style_prompt, created_at, updated_at`

	var c models.Campaign
	var retGenre *string
	err = db.QueryRow(ctx, query, id, name, systemID, description, settings, genreStr, imageStylePrompt, ownerID).Scan(
		&c.ID, &c.Name, &c.SystemID, &c.OwnerID, &c.Description, &c.Settings,
		&retGenre, &c.ImageStylePrompt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update campaign: %w", err)
	}

	if retGenre != nil {
		g := models.CampaignGenre(*retGenre)
		c.Genre = &g
	}

	return &c, nil
}

// DeleteCampaign deletes a campaign by ID without ownership verification.
// For user-scoped deletes, use DeleteCampaignByOwner instead.
func (db *DB) DeleteCampaign(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM campaigns WHERE id = $1`
	result, err := db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete campaign: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("campaign not found")
	}

	return nil
}

// DeleteCampaignByOwner deletes a campaign by ID with ownership verification.
// Returns an error if the campaign doesn't exist or doesn't belong to the owner.
// This is the primary method for user-scoped campaign deletion.
func (db *DB) DeleteCampaignByOwner(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) error {
	query := `DELETE FROM campaigns WHERE id = $1 AND owner_id = $2`
	result, err := db.Pool.Exec(ctx, query, id, ownerID)
	if err != nil {
		return fmt.Errorf("failed to delete campaign: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("campaign not found")
	}

	return nil
}

// VerifyCampaignOwnership checks if a campaign belongs to a specific user.
// Returns nil if the campaign exists and belongs to the owner.
// Returns an error if the campaign doesn't exist or doesn't belong to the owner.
func (db *DB) VerifyCampaignOwnership(ctx context.Context, campaignID uuid.UUID, ownerID uuid.UUID) error {
	query := `SELECT 1 FROM campaigns WHERE id = $1 AND owner_id = $2`
	var exists int
	err := db.QueryRow(ctx, query, campaignID, ownerID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("campaign not found")
		}
		return fmt.Errorf("failed to verify campaign ownership: %w", err)
	}
	return nil
}

// GetRecentCampaigns retrieves the most recently updated campaigns.
func (db *DB) GetRecentCampaigns(ctx context.Context, limit int) ([]models.Campaign, error) {
	query := `
        SELECT id, name, system_id, owner_id, description, settings, genre, image_style_prompt, created_at, updated_at
        FROM campaigns
        ORDER BY updated_at DESC
        LIMIT $1`

	rows, err := db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent campaigns: %w", err)
	}
	defer rows.Close()

	var campaigns []models.Campaign
	for rows.Next() {
		var c models.Campaign
		var genre *string
		err := rows.Scan(
			&c.ID, &c.Name, &c.SystemID, &c.OwnerID, &c.Description, &c.Settings,
			&genre, &c.ImageStylePrompt, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan campaign: %w", err)
		}
		if genre != nil {
			g := models.CampaignGenre(*genre)
			c.Genre = &g
		}
		campaigns = append(campaigns, c)
	}

	return campaigns, nil
}

// GetCampaignsByOwnerID retrieves all campaigns owned by a specific user.
func (db *DB) GetCampaignsByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]models.Campaign, error) {
	query := `
        SELECT c.id, c.name, c.system_id, c.owner_id, c.description, c.settings,
               c.genre, c.image_style_prompt, c.created_at, c.updated_at,
               gs.id, gs.name, gs.code
        FROM campaigns c
        LEFT JOIN game_systems gs ON c.system_id = gs.id
        WHERE c.owner_id = $1
        ORDER BY c.updated_at DESC`

	rows, err := db.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query campaigns by owner: %w", err)
	}
	defer rows.Close()

	var campaigns []models.Campaign
	for rows.Next() {
		var c models.Campaign
		var gsID *uuid.UUID
		var gsName, gsCode *string
		var genre *string

		err := rows.Scan(
			&c.ID, &c.Name, &c.SystemID, &c.OwnerID, &c.Description, &c.Settings,
			&genre, &c.ImageStylePrompt, &c.CreatedAt, &c.UpdatedAt,
			&gsID, &gsName, &gsCode,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan campaign: %w", err)
		}

		if genre != nil {
			g := models.CampaignGenre(*genre)
			c.Genre = &g
		}

		if gsID != nil {
			c.System = &models.GameSystem{
				ID:   *gsID,
				Name: *gsName,
				Code: *gsCode,
			}
		}

		campaigns = append(campaigns, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating campaigns: %w", err)
	}

	return campaigns, nil
}
