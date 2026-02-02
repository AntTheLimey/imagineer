/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

-- Settings and Player Characters Migration
-- Adds user settings for API keys and player character management

-- ============================================
-- User Settings Table
-- Stores per-user LLM API keys and service preferences
-- ============================================
CREATE TABLE user_settings (
    user_id                 UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    -- Content generation service configuration
    content_gen_service     TEXT CHECK (content_gen_service IN ('anthropic', 'openai', 'gemini')),
    content_gen_api_key     TEXT,  -- Encrypted at application layer
    -- Embedding generation service configuration
    embedding_service       TEXT CHECK (embedding_service IN ('voyage', 'openai', 'gemini')),
    embedding_api_key       TEXT,  -- Encrypted at application layer
    -- Image generation service configuration
    image_gen_service       TEXT CHECK (image_gen_service IN ('openai', 'stability')),
    image_gen_api_key       TEXT,  -- Encrypted at application layer
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE user_settings IS 'User preferences and API keys for LLM services (content, embeddings, images)';
COMMENT ON COLUMN user_settings.user_id IS 'One-to-one relationship with users table';
COMMENT ON COLUMN user_settings.content_gen_service IS 'Selected service for content generation: anthropic (Claude), openai (GPT), or gemini';
COMMENT ON COLUMN user_settings.content_gen_api_key IS 'API key for content generation service (encrypted at application layer)';
COMMENT ON COLUMN user_settings.embedding_service IS 'Selected service for embedding generation: voyage, openai, or gemini';
COMMENT ON COLUMN user_settings.embedding_api_key IS 'API key for embedding service (encrypted at application layer)';
COMMENT ON COLUMN user_settings.image_gen_service IS 'Selected service for image generation: openai (DALL-E) or stability (Stable Diffusion)';
COMMENT ON COLUMN user_settings.image_gen_api_key IS 'API key for image generation service (encrypted at application layer)';

-- ============================================
-- Campaign Settings Extensions
-- Add genre and image generation settings
-- ============================================
ALTER TABLE campaigns
    ADD COLUMN genre TEXT CHECK (genre IN (
        'anime_manga',
        'cyberpunk',
        'espionage',
        'fantasy',
        'gothic',
        'historical',
        'horror',
        'lovecraftian',
        'military',
        'modern_urban_fantasy',
        'mystery',
        'post_apocalyptic',
        'pulp_adventure',
        'science_fiction',
        'space_opera',
        'steampunk',
        'superhero',
        'time_travel',
        'western',
        'other'
    )),
    ADD COLUMN image_style_prompt TEXT;

COMMENT ON COLUMN campaigns.genre IS 'Campaign genre/setting classification for AI generation context';
COMMENT ON COLUMN campaigns.image_style_prompt IS 'Default style prompt for AI image generation (e.g., "1920s noir photography style")';

-- ============================================
-- Player Characters Table
-- Stores player characters within campaigns
-- ============================================
CREATE TABLE player_characters (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    campaign_id         UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    entity_id           UUID REFERENCES entities(id) ON DELETE SET NULL,
    character_name      TEXT NOT NULL,
    player_name         TEXT NOT NULL,
    description         TEXT,  -- Rich text character description
    background          TEXT,  -- Rich text character backstory
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW(),
    
    -- Ensure character names are unique within a campaign
    CONSTRAINT unique_character_per_campaign UNIQUE (campaign_id, character_name)
);

COMMENT ON TABLE player_characters IS 'Player characters participating in campaigns';
COMMENT ON COLUMN player_characters.campaign_id IS 'Campaign this character belongs to';
COMMENT ON COLUMN player_characters.entity_id IS 'Optional link to entity record for this PC (allows treating PCs as entities in relationships)';
COMMENT ON COLUMN player_characters.character_name IS 'Name of the player character';
COMMENT ON COLUMN player_characters.player_name IS 'Name of the player controlling this character';
COMMENT ON COLUMN player_characters.description IS 'Rich text description of the character (appearance, personality, etc.)';
COMMENT ON COLUMN player_characters.background IS 'Rich text backstory and character history';

-- ============================================
-- Indexes
-- ============================================

-- Foreign key index for campaign lookups
CREATE INDEX idx_player_characters_campaign_id ON player_characters(campaign_id);

-- Foreign key index for entity linkage
CREATE INDEX idx_player_characters_entity_id ON player_characters(entity_id);

-- Index for player lookups (find all characters for a player)
CREATE INDEX idx_player_characters_player_name ON player_characters(player_name);

-- ============================================
-- Updated Timestamp Triggers
-- ============================================

-- Trigger for user_settings
CREATE TRIGGER update_user_settings_updated_at 
    BEFORE UPDATE ON user_settings
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger for player_characters
CREATE TRIGGER update_player_characters_updated_at 
    BEFORE UPDATE ON player_characters
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Record migration
-- ============================================
INSERT INTO schema_migrations (version) VALUES ('006_settings_and_pcs');
