/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

-- Migration 007: Encrypt API keys at rest
--
-- API keys were previously stored as plaintext. This migration NULLs
-- all existing plaintext keys so they can be re-entered and stored
-- encrypted at the application layer (AES-256-GCM).

-- NULL all existing plaintext API keys
UPDATE user_settings
SET content_gen_api_key = NULL,
    embedding_api_key = NULL,
    image_gen_api_key = NULL;

-- Record migration
INSERT INTO schema_migrations (version) VALUES ('007_encrypt_api_keys');

-- Update column comments to note encryption
COMMENT ON COLUMN user_settings.content_gen_api_key IS
    'Content generation API key — encrypted at application layer (AES-256-GCM)';
COMMENT ON COLUMN user_settings.embedding_api_key IS
    'Embedding API key — encrypted at application layer (AES-256-GCM)';
COMMENT ON COLUMN user_settings.image_gen_api_key IS
    'Image generation API key — encrypted at application layer (AES-256-GCM)';
