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
	"log"
	"strings"

	"github.com/antonypegg/imagineer/internal/models"
	"github.com/jackc/pgx/v5"
)

// GetUserSettings retrieves user settings by user ID.
// Returns nil if no settings exist for the user.
func (db *DB) GetUserSettings(ctx context.Context, userID int64) (*models.UserSettings, error) {
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

	// Decrypt API keys
	s.ContentGenAPIKey = db.decryptKey(s.ContentGenAPIKey)
	s.EmbeddingAPIKey = db.decryptKey(s.EmbeddingAPIKey)
	s.ImageGenAPIKey = db.decryptKey(s.ImageGenAPIKey)

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

// UpdateUserSettings updates or creates user settings using UPSERT with COALESCE.
// Uses database-side COALESCE to atomically merge updates, preventing race conditions.
// NULL values in the request preserve existing values in the database.
func (db *DB) UpdateUserSettings(ctx context.Context, userID int64, req models.UpdateUserSettingsRequest) (*models.UserSettings, error) {
	// Convert request LLMService pointers to strings for database
	var contentGenService, embeddingService, imageGenService *string
	var contentGenAPIKey, embeddingAPIKey, imageGenAPIKey *string

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

	// Encrypt API keys before writing
	contentGenAPIKey = db.encryptKey(contentGenAPIKey)
	embeddingAPIKey = db.encryptKey(embeddingAPIKey)
	imageGenAPIKey = db.encryptKey(imageGenAPIKey)

	// Use COALESCE to atomically merge: NULL in request preserves existing value
	query := `
        INSERT INTO user_settings (
            user_id, content_gen_service, content_gen_api_key,
            embedding_service, embedding_api_key,
            image_gen_service, image_gen_api_key
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (user_id) DO UPDATE SET
            content_gen_service = COALESCE(EXCLUDED.content_gen_service, user_settings.content_gen_service),
            content_gen_api_key = COALESCE(EXCLUDED.content_gen_api_key, user_settings.content_gen_api_key),
            embedding_service = COALESCE(EXCLUDED.embedding_service, user_settings.embedding_service),
            embedding_api_key = COALESCE(EXCLUDED.embedding_api_key, user_settings.embedding_api_key),
            image_gen_service = COALESCE(EXCLUDED.image_gen_service, user_settings.image_gen_service),
            image_gen_api_key = COALESCE(EXCLUDED.image_gen_api_key, user_settings.image_gen_api_key),
            updated_at = NOW()
        RETURNING user_id, content_gen_service, content_gen_api_key,
                  embedding_service, embedding_api_key,
                  image_gen_service, image_gen_api_key,
                  created_at, updated_at`

	var s models.UserSettings
	var retContentGenService, retEmbeddingService, retImageGenService *string
	err := db.QueryRow(ctx, query,
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

	// Decrypt API keys from RETURNING values
	s.ContentGenAPIKey = db.decryptKey(s.ContentGenAPIKey)
	s.EmbeddingAPIKey = db.decryptKey(s.EmbeddingAPIKey)
	s.ImageGenAPIKey = db.decryptKey(s.ImageGenAPIKey)

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

// encryptKey encrypts an API key if encryption is configured.
// Returns the original value if Encryptor is nil or the key is nil/empty.
func (db *DB) encryptKey(key *string) *string {
	if db.Encryptor == nil || key == nil || *key == "" {
		return key
	}
	encrypted, err := db.Encryptor.Encrypt(*key)
	if err != nil {
		log.Printf("WARNING: failed to encrypt API key: %v", err)
		return key
	}
	return &encrypted
}

// decryptKey decrypts an API key if encryption is configured.
// Returns nil on decryption failure (user must re-enter key).
// When no Encryptor is available but the stored value carries the "enc:"
// prefix, the key is treated as unreadable and nil is returned so the
// caller sees it as absent rather than leaking ciphertext.
func (db *DB) decryptKey(key *string) *string {
	if key == nil || *key == "" {
		return key
	}
	if db.Encryptor == nil {
		if strings.HasPrefix(*key, "enc:") {
			log.Printf("WARNING: encrypted API key found but ENCRYPTION_KEY not set (re-enter required)")
			return nil
		}
		return key
	}
	decrypted, err := db.Encryptor.Decrypt(*key)
	if err != nil {
		log.Printf("WARNING: failed to decrypt API key (re-enter required): %v", err)
		return nil
	}
	return &decrypted
}
