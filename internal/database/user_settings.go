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
	"github.com/jackc/pgx/v5"
)

// GetUserSettings retrieves user settings by user ID.
// Returns nil if no settings exist for the user.
func (db *DB) GetUserSettings(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error) {
	query := `
        SELECT user_id, content_gen_service, content_gen_api_key,
               embedding_service, embedding_api_key,
               image_gen_service, image_gen_api_key,
               created_at, updated_at
        FROM user_settings
        WHERE user_id = $1`

	var s models.UserSettings
	var contentGenService, embeddingService, imageGenService *string
	err := db.QueryRow(ctx, query, userID).Scan(
		&s.UserID, &contentGenService, &s.ContentGenAPIKey,
		&embeddingService, &s.EmbeddingAPIKey,
		&imageGenService, &s.ImageGenAPIKey,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}

	// Convert string pointers to LLMService pointers
	if contentGenService != nil {
		svc := models.LLMService(*contentGenService)
		s.ContentGenService = &svc
	}
	if embeddingService != nil {
		svc := models.LLMService(*embeddingService)
		s.EmbeddingService = &svc
	}
	if imageGenService != nil {
		svc := models.LLMService(*imageGenService)
		s.ImageGenService = &svc
	}

	return &s, nil
}

// UpdateUserSettings updates or creates user settings using UPSERT.
// Returns the updated settings.
func (db *DB) UpdateUserSettings(ctx context.Context, userID uuid.UUID, req models.UpdateUserSettingsRequest) (*models.UserSettings, error) {
	// First get existing settings to merge with updates
	existing, err := db.GetUserSettings(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Prepare values, merging with existing if they exist
	var contentGenService, embeddingService, imageGenService *string
	var contentGenAPIKey, embeddingAPIKey, imageGenAPIKey *string

	if existing != nil {
		if existing.ContentGenService != nil {
			s := string(*existing.ContentGenService)
			contentGenService = &s
		}
		contentGenAPIKey = existing.ContentGenAPIKey
		if existing.EmbeddingService != nil {
			s := string(*existing.EmbeddingService)
			embeddingService = &s
		}
		embeddingAPIKey = existing.EmbeddingAPIKey
		if existing.ImageGenService != nil {
			s := string(*existing.ImageGenService)
			imageGenService = &s
		}
		imageGenAPIKey = existing.ImageGenAPIKey
	}

	// Apply updates from request
	if req.ContentGenService != nil {
		s := string(*req.ContentGenService)
		contentGenService = &s
	}
	if req.ContentGenAPIKey != nil {
		contentGenAPIKey = req.ContentGenAPIKey
	}
	if req.EmbeddingService != nil {
		s := string(*req.EmbeddingService)
		embeddingService = &s
	}
	if req.EmbeddingAPIKey != nil {
		embeddingAPIKey = req.EmbeddingAPIKey
	}
	if req.ImageGenService != nil {
		s := string(*req.ImageGenService)
		imageGenService = &s
	}
	if req.ImageGenAPIKey != nil {
		imageGenAPIKey = req.ImageGenAPIKey
	}

	query := `
        INSERT INTO user_settings (
            user_id, content_gen_service, content_gen_api_key,
            embedding_service, embedding_api_key,
            image_gen_service, image_gen_api_key
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (user_id) DO UPDATE SET
            content_gen_service = EXCLUDED.content_gen_service,
            content_gen_api_key = EXCLUDED.content_gen_api_key,
            embedding_service = EXCLUDED.embedding_service,
            embedding_api_key = EXCLUDED.embedding_api_key,
            image_gen_service = EXCLUDED.image_gen_service,
            image_gen_api_key = EXCLUDED.image_gen_api_key,
            updated_at = NOW()
        RETURNING user_id, content_gen_service, content_gen_api_key,
                  embedding_service, embedding_api_key,
                  image_gen_service, image_gen_api_key,
                  created_at, updated_at`

	var s models.UserSettings
	var retContentGenService, retEmbeddingService, retImageGenService *string
	err = db.QueryRow(ctx, query,
		userID, contentGenService, contentGenAPIKey,
		embeddingService, embeddingAPIKey,
		imageGenService, imageGenAPIKey,
	).Scan(
		&s.UserID, &retContentGenService, &s.ContentGenAPIKey,
		&retEmbeddingService, &s.EmbeddingAPIKey,
		&retImageGenService, &s.ImageGenAPIKey,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update user settings: %w", err)
	}

	// Convert string pointers to LLMService pointers
	if retContentGenService != nil {
		svc := models.LLMService(*retContentGenService)
		s.ContentGenService = &svc
	}
	if retEmbeddingService != nil {
		svc := models.LLMService(*retEmbeddingService)
		s.EmbeddingService = &svc
	}
	if retImageGenService != nil {
		svc := models.LLMService(*retImageGenService)
		s.ImageGenService = &svc
	}

	return &s, nil
}
