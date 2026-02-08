/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

-- Squashed Schema Migration
-- Replaces migrations 001-010 with the complete final database state.
-- This file represents the canonical schema for the Imagineer platform.

-- ============================================
-- Extensions
-- ============================================
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- ============================================
-- Migration Tracking Table
-- ============================================
CREATE TABLE IF NOT EXISTS schema_migrations (
    version    TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- Updated At Trigger Function
-- Automatically updates updated_at timestamp
-- ============================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- Game Systems Table
-- Defines TTRPG systems (CoC 7e, GURPS 4e, FitD, etc.)
-- ============================================
CREATE TABLE game_systems (
    id                       BIGSERIAL PRIMARY KEY,
    name                     TEXT NOT NULL,
    code                     TEXT UNIQUE NOT NULL,
    attribute_schema         JSONB,
    skill_schema             JSONB,
    character_sheet_template JSONB,
    dice_conventions         JSONB,
    created_at               TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE game_systems IS 'TTRPG system definitions (e.g., Call of Cthulhu 7e, GURPS 4e)';
COMMENT ON COLUMN game_systems.code IS 'Short unique identifier (e.g., coc-7e, gurps-4e, fitd)';
COMMENT ON COLUMN game_systems.attribute_schema IS 'JSON schema defining character attributes for this system';

-- ============================================
-- Users Table
-- User accounts authenticated via Google OAuth
-- ============================================
CREATE TABLE users (
    id         BIGSERIAL PRIMARY KEY,
    google_id  TEXT UNIQUE NOT NULL,
    email      TEXT UNIQUE NOT NULL,
    name       TEXT NOT NULL,
    avatar_url TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

COMMENT ON TABLE users IS 'User accounts authenticated via Google OAuth';
COMMENT ON COLUMN users.google_id IS 'Unique identifier from Google OAuth provider';
COMMENT ON COLUMN users.email IS 'User email address from Google account';
COMMENT ON COLUMN users.name IS 'Display name from Google account';
COMMENT ON COLUMN users.avatar_url IS 'URL to user profile image from Google';
COMMENT ON COLUMN users.deleted_at IS 'Timestamp when user was soft-deleted (NULL means active)';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Campaigns Table
-- Individual campaign instances
-- ============================================
CREATE TABLE campaigns (
    id                BIGSERIAL PRIMARY KEY,
    name              TEXT NOT NULL,
    system_id         BIGINT REFERENCES game_systems(id) ON DELETE SET NULL,
    description       TEXT,
    settings          JSONB DEFAULT '{}',
    owner_id          BIGINT REFERENCES users(id) ON DELETE SET NULL,
    genre             TEXT CHECK (genre IN (
        'anime_manga', 'cyberpunk', 'espionage', 'fantasy', 'gothic',
        'historical', 'horror', 'lovecraftian', 'military',
        'modern_urban_fantasy', 'mystery', 'post_apocalyptic',
        'pulp_adventure', 'science_fiction', 'space_opera', 'steampunk',
        'superhero', 'time_travel', 'western', 'other'
    )),
    image_style_prompt TEXT,
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE campaigns IS 'Individual TTRPG campaigns';
COMMENT ON COLUMN campaigns.settings IS 'Campaign-specific settings and configuration';
COMMENT ON COLUMN campaigns.owner_id IS 'User who owns this campaign (nullable for legacy data)';
COMMENT ON COLUMN campaigns.genre IS 'Campaign genre/setting classification for AI generation context';
COMMENT ON COLUMN campaigns.image_style_prompt IS 'Default style prompt for AI image generation (e.g., "1920s noir photography style")';

CREATE INDEX idx_campaigns_system_id ON campaigns(system_id);
CREATE INDEX idx_campaigns_owner_id ON campaigns(owner_id);

CREATE TRIGGER update_campaigns_updated_at BEFORE UPDATE ON campaigns
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- User Settings Table
-- Stores per-user LLM API keys and service preferences
-- ============================================
CREATE TABLE user_settings (
    user_id             BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    content_gen_service TEXT CHECK (content_gen_service IN ('anthropic', 'openai', 'gemini')),
    content_gen_api_key TEXT,
    embedding_service   TEXT CHECK (embedding_service IN ('voyage', 'openai', 'gemini', 'ollama')),
    embedding_api_key   TEXT,
    image_gen_service   TEXT CHECK (image_gen_service IN ('openai', 'stability')),
    image_gen_api_key   TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE user_settings IS 'User preferences and API keys for LLM services (content, embeddings, images)';
COMMENT ON COLUMN user_settings.user_id IS 'One-to-one relationship with users table';
COMMENT ON COLUMN user_settings.content_gen_service IS 'Selected service for content generation: anthropic (Claude), openai (GPT), or gemini';
COMMENT ON COLUMN user_settings.content_gen_api_key IS 'API key for content generation service (encrypted at application layer)';
COMMENT ON COLUMN user_settings.embedding_service IS 'Selected service for embedding generation: voyage, openai, gemini, or ollama';
COMMENT ON COLUMN user_settings.embedding_api_key IS 'API key for embedding service (encrypted at application layer)';
COMMENT ON COLUMN user_settings.image_gen_service IS 'Selected service for image generation: openai (DALL-E) or stability (Stable Diffusion)';
COMMENT ON COLUMN user_settings.image_gen_api_key IS 'API key for image generation service (encrypted at application layer)';

CREATE TRIGGER update_user_settings_updated_at BEFORE UPDATE ON user_settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Chapters Table
-- Organize sessions into story arcs
-- ============================================
CREATE TABLE chapters (
    id          BIGSERIAL PRIMARY KEY,
    campaign_id BIGINT NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    overview    TEXT,
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE chapters IS 'Story arcs that organize sessions within a campaign';
COMMENT ON COLUMN chapters.campaign_id IS 'Campaign this chapter belongs to';
COMMENT ON COLUMN chapters.title IS 'Chapter title (e.g., "The Arkham Horror")';
COMMENT ON COLUMN chapters.overview IS 'Rich text description of the chapter arc';
COMMENT ON COLUMN chapters.sort_order IS 'Order within campaign for display (0-based)';

CREATE INDEX idx_chapters_campaign_id ON chapters(campaign_id);
CREATE INDEX idx_chapters_sort_order ON chapters(campaign_id, sort_order);

CREATE TRIGGER update_chapters_updated_at BEFORE UPDATE ON chapters
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Sessions Table
-- Game sessions within a campaign
-- ============================================
CREATE TABLE sessions (
    id               BIGSERIAL PRIMARY KEY,
    campaign_id      BIGINT NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    chapter_id       BIGINT REFERENCES chapters(id) ON DELETE SET NULL,
    session_number   INT,
    planned_date     DATE,
    actual_date      DATE,
    status           TEXT DEFAULT 'PLANNED' CHECK (status IN ('PLANNED', 'COMPLETED', 'SKIPPED')),
    stage            TEXT DEFAULT 'prep' CHECK (stage IN ('prep', 'play', 'wrap_up')),
    title            TEXT,
    prep_notes       TEXT,
    planned_scenes   JSONB,
    actual_notes     TEXT,
    discoveries      JSONB DEFAULT '[]',
    player_decisions JSONB DEFAULT '[]',
    consequences     JSONB DEFAULT '[]',
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE sessions IS 'Individual game sessions within a campaign';
COMMENT ON COLUMN sessions.chapter_id IS 'Chapter this session belongs to (NULL for uncategorized)';
COMMENT ON COLUMN sessions.discoveries IS 'Array of {entity_id, how_discovered} objects';
COMMENT ON COLUMN sessions.player_decisions IS 'Notable player choices made during the session';
COMMENT ON COLUMN sessions.stage IS 'Workflow stage: prep (preparation), play (active session), wrap_up (post-session notes)';
COMMENT ON COLUMN sessions.title IS 'Session title (e.g., "The Haunted Mansion")';

CREATE INDEX idx_sessions_campaign_id ON sessions(campaign_id);
CREATE INDEX idx_sessions_status ON sessions(status);
CREATE UNIQUE INDEX idx_sessions_campaign_number ON sessions(campaign_id, session_number);
CREATE INDEX idx_sessions_chapter_id ON sessions(chapter_id);

CREATE TRIGGER update_sessions_updated_at BEFORE UPDATE ON sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Entities Table (Polymorphic)
-- NPCs, locations, items, factions, clues, etc.
-- ============================================
CREATE TABLE entities (
    id                 BIGSERIAL PRIMARY KEY,
    campaign_id        BIGINT NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    entity_type        TEXT NOT NULL CHECK (entity_type IN (
                           'npc', 'location', 'item', 'faction', 'clue',
                           'creature', 'organization', 'event', 'document', 'other'
                       )),
    name               TEXT NOT NULL,
    description        TEXT,
    attributes         JSONB DEFAULT '{}',
    tags               TEXT[] DEFAULT '{}',
    gm_notes           TEXT,
    discovered_session BIGINT REFERENCES sessions(id) ON DELETE SET NULL,
    source_document    TEXT,
    source_confidence  TEXT DEFAULT 'DRAFT' CHECK (source_confidence IN ('DRAFT', 'AUTHORITATIVE', 'SUPERSEDED')),
    version            INT DEFAULT 1,
    created_at         TIMESTAMPTZ DEFAULT NOW(),
    updated_at         TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE entities IS 'Polymorphic table for all campaign entities (NPCs, locations, items, etc.)';
COMMENT ON COLUMN entities.entity_type IS 'Type discriminator for the entity';
COMMENT ON COLUMN entities.attributes IS 'System-specific stats and properties (JSONB)';
COMMENT ON COLUMN entities.gm_notes IS 'GM-only notes, never shown to players';
COMMENT ON COLUMN entities.source_confidence IS 'Canon status: DRAFT, AUTHORITATIVE, or SUPERSEDED';

CREATE INDEX idx_entities_campaign_id ON entities(campaign_id);
CREATE INDEX idx_entities_type ON entities(entity_type);
CREATE INDEX idx_entities_campaign_type ON entities(campaign_id, entity_type);
CREATE INDEX idx_entities_tags ON entities USING GIN(tags);
CREATE INDEX idx_entities_attributes ON entities USING GIN(attributes);
CREATE INDEX idx_entities_name_trgm ON entities USING GIN(name gin_trgm_ops);

CREATE TRIGGER update_entities_updated_at BEFORE UPDATE ON entities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Relationships Table
-- Connections between entities
-- ============================================
CREATE TABLE relationships (
    id                BIGSERIAL PRIMARY KEY,
    campaign_id       BIGINT NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    source_entity_id  BIGINT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    target_entity_id  BIGINT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    relationship_type TEXT NOT NULL,
    tone              TEXT CHECK (tone IN ('friendly', 'hostile', 'neutral', 'romantic', 'professional', 'fearful', 'respectful', 'unknown')),
    description       TEXT,
    bidirectional     BOOLEAN DEFAULT false,
    strength          INT CHECK (strength >= 1 AND strength <= 10),
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT no_self_relationship CHECK (source_entity_id != target_entity_id)
);

COMMENT ON TABLE relationships IS 'Connections between campaign entities';
COMMENT ON COLUMN relationships.relationship_type IS 'Type of relationship (knows, owns, located_at, opposes, etc.)';
COMMENT ON COLUMN relationships.strength IS 'Optional weight for graph algorithms (1-10)';

CREATE INDEX idx_relationships_campaign_id ON relationships(campaign_id);
CREATE INDEX idx_relationships_source ON relationships(source_entity_id);
CREATE INDEX idx_relationships_target ON relationships(target_entity_id);
CREATE INDEX idx_relationships_type ON relationships(relationship_type);

CREATE TRIGGER update_relationships_updated_at BEFORE UPDATE ON relationships
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Relationship Types Table
-- Defines relationship types with inverse mappings
-- ============================================
CREATE TABLE relationship_types (
    id                    BIGSERIAL PRIMARY KEY,
    campaign_id           BIGINT REFERENCES campaigns(id) ON DELETE CASCADE,
    name                  TEXT NOT NULL,
    inverse_name          TEXT NOT NULL,
    is_symmetric          BOOLEAN NOT NULL DEFAULT false,
    display_label         TEXT NOT NULL,
    inverse_display_label TEXT NOT NULL,
    description           TEXT,
    created_at            TIMESTAMPTZ DEFAULT NOW(),
    updated_at            TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT unique_relationship_type_per_campaign
        UNIQUE NULLS NOT DISTINCT (campaign_id, name),
    CONSTRAINT symmetric_inverse_match
        CHECK (NOT is_symmetric OR name = inverse_name)
);

COMMENT ON TABLE relationship_types IS 'Relationship type definitions with inverse mappings, supporting per-campaign customization';
COMMENT ON COLUMN relationship_types.campaign_id IS 'Campaign this type belongs to; NULL for system-wide default types';
COMMENT ON COLUMN relationship_types.name IS 'Relationship type name used in relationships table (e.g., "owns", "knows")';
COMMENT ON COLUMN relationship_types.inverse_name IS 'Inverse relationship type (e.g., "owned_by" for "owns", or same name if symmetric)';
COMMENT ON COLUMN relationship_types.is_symmetric IS 'TRUE if relationship reads the same in both directions (e.g., "knows")';
COMMENT ON COLUMN relationship_types.display_label IS 'Human-friendly label for UI display (e.g., "Owns")';
COMMENT ON COLUMN relationship_types.inverse_display_label IS 'Display label for inverse relationship (e.g., "Is owned by")';

CREATE INDEX idx_relationship_types_name ON relationship_types(name);
CREATE INDEX idx_relationship_types_symmetric ON relationship_types(is_symmetric) WHERE is_symmetric = true;

CREATE TRIGGER update_relationship_types_updated_at BEFORE UPDATE ON relationship_types
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Timeline Events Table
-- In-game chronological events
-- ============================================
CREATE TABLE timeline_events (
    id              BIGSERIAL PRIMARY KEY,
    campaign_id     BIGINT NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    event_date      DATE,
    event_time      TIME,
    date_precision  TEXT DEFAULT 'exact' CHECK (date_precision IN ('exact', 'approximate', 'month', 'year', 'unknown')),
    description     TEXT NOT NULL,
    entity_ids      BIGINT[] DEFAULT '{}',
    session_id      BIGINT REFERENCES sessions(id) ON DELETE SET NULL,
    is_player_known BOOLEAN DEFAULT false,
    source_document TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE timeline_events IS 'In-game chronological events';
COMMENT ON COLUMN timeline_events.event_date IS 'In-game date when the event occurred';
COMMENT ON COLUMN timeline_events.is_player_known IS 'Whether players have learned about this event';
COMMENT ON COLUMN timeline_events.entity_ids IS 'Array of entity IDs involved in this event';

CREATE INDEX idx_timeline_campaign_id ON timeline_events(campaign_id);
CREATE INDEX idx_timeline_date ON timeline_events(event_date);
CREATE INDEX idx_timeline_player_known ON timeline_events(is_player_known);
CREATE INDEX idx_timeline_entity_ids ON timeline_events USING GIN(entity_ids);

CREATE TRIGGER update_timeline_events_updated_at BEFORE UPDATE ON timeline_events
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Canon Conflicts Table
-- Track contradictions between sources
-- ============================================
CREATE TABLE canon_conflicts (
    id                 BIGSERIAL PRIMARY KEY,
    campaign_id        BIGINT NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    entity_id          BIGINT REFERENCES entities(id) ON DELETE CASCADE,
    field_name         TEXT,
    conflicting_values JSONB NOT NULL,
    status             TEXT DEFAULT 'DETECTED' CHECK (status IN ('DETECTED', 'ACKNOWLEDGED', 'RESOLVED')),
    resolution         TEXT,
    resolved_at        TIMESTAMPTZ,
    created_at         TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE canon_conflicts IS 'Track contradictions between different sources';
COMMENT ON COLUMN canon_conflicts.conflicting_values IS 'Array of {value, source, date} objects';
COMMENT ON COLUMN canon_conflicts.status IS 'DETECTED (new), ACKNOWLEDGED (seen), RESOLVED (fixed)';

CREATE INDEX idx_canon_conflicts_campaign ON canon_conflicts(campaign_id);
CREATE INDEX idx_canon_conflicts_entity ON canon_conflicts(entity_id);
CREATE INDEX idx_canon_conflicts_status ON canon_conflicts(status);

-- ============================================
-- Player Characters Table
-- Player characters within campaigns
-- ============================================
CREATE TABLE player_characters (
    id             BIGSERIAL PRIMARY KEY,
    campaign_id    BIGINT NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    entity_id      BIGINT REFERENCES entities(id) ON DELETE SET NULL,
    character_name TEXT NOT NULL,
    player_name    TEXT NOT NULL,
    description    TEXT,
    background     TEXT,
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    updated_at     TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT unique_character_per_campaign UNIQUE (campaign_id, character_name)
);

COMMENT ON TABLE player_characters IS 'Player characters participating in campaigns';
COMMENT ON COLUMN player_characters.campaign_id IS 'Campaign this character belongs to';
COMMENT ON COLUMN player_characters.entity_id IS 'Optional link to entity record for this PC (allows treating PCs as entities in relationships)';
COMMENT ON COLUMN player_characters.character_name IS 'Name of the player character';
COMMENT ON COLUMN player_characters.player_name IS 'Name of the player controlling this character';
COMMENT ON COLUMN player_characters.description IS 'Rich text description of the character (appearance, personality, etc.)';
COMMENT ON COLUMN player_characters.background IS 'Rich text backstory and character history';

CREATE INDEX idx_player_characters_campaign_id ON player_characters(campaign_id);
CREATE INDEX idx_player_characters_entity_id ON player_characters(entity_id);
CREATE INDEX idx_player_characters_player_name ON player_characters(player_name);

CREATE TRIGGER update_player_characters_updated_at BEFORE UPDATE ON player_characters
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Campaign Memories Table
-- Long-term facts and knowledge for a campaign
-- ============================================
CREATE TABLE campaign_memories (
    id                BIGSERIAL PRIMARY KEY,
    campaign_id       BIGINT NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    memory_type       TEXT NOT NULL CHECK (memory_type IN (
                          'premise', 'theme', 'faction_summary', 'plot_thread', 'gm_note'
                      )),
    title             TEXT,
    content           TEXT NOT NULL,
    source_session_id BIGINT REFERENCES sessions(id) ON DELETE SET NULL,
    importance        INT DEFAULT 5 CHECK (importance >= 1 AND importance <= 10),
    is_spoiler        BOOLEAN DEFAULT false,
    gm_created        BOOLEAN DEFAULT false,
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE campaign_memories IS 'Long-term campaign facts and knowledge for AI context';
COMMENT ON COLUMN campaign_memories.campaign_id IS 'Campaign this memory belongs to';
COMMENT ON COLUMN campaign_memories.memory_type IS 'Type of memory: premise, theme, faction_summary, plot_thread, gm_note';
COMMENT ON COLUMN campaign_memories.title IS 'Brief title for the memory';
COMMENT ON COLUMN campaign_memories.content IS 'Full text content of the memory';
COMMENT ON COLUMN campaign_memories.source_session_id IS 'Session where this memory was created (NULL if campaign-level)';
COMMENT ON COLUMN campaign_memories.importance IS 'Priority ranking from 1 (low) to 10 (high)';
COMMENT ON COLUMN campaign_memories.is_spoiler IS 'True if memory contains information hidden from players';
COMMENT ON COLUMN campaign_memories.gm_created IS 'True if GM manually created this memory';

CREATE INDEX idx_campaign_memories_campaign_id ON campaign_memories(campaign_id);
CREATE INDEX idx_campaign_memories_type ON campaign_memories(memory_type);
CREATE INDEX idx_campaign_memories_importance ON campaign_memories(importance DESC);
CREATE INDEX idx_campaign_memories_source_session ON campaign_memories(source_session_id);

CREATE TRIGGER update_campaign_memories_updated_at BEFORE UPDATE ON campaign_memories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Chapter Memories Table
-- Arc-level context for campaign chapters
-- ============================================
CREATE TABLE chapter_memories (
    id             BIGSERIAL PRIMARY KEY,
    campaign_id    BIGINT NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    chapter_id     BIGINT NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
    goals          TEXT[],
    active_threads TEXT[],
    summary        TEXT,
    is_current     BOOLEAN DEFAULT false,
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    updated_at     TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE chapter_memories IS 'Arc-level context and goals for campaign chapters';
COMMENT ON COLUMN chapter_memories.campaign_id IS 'Campaign this memory belongs to';
COMMENT ON COLUMN chapter_memories.chapter_id IS 'Chapter this memory describes';
COMMENT ON COLUMN chapter_memories.goals IS 'Array of goal descriptions for the chapter';
COMMENT ON COLUMN chapter_memories.active_threads IS 'Array of active plot thread descriptions';
COMMENT ON COLUMN chapter_memories.summary IS 'Narrative summary of the chapter arc';
COMMENT ON COLUMN chapter_memories.is_current IS 'True if this is the active chapter in the campaign';

CREATE INDEX idx_chapter_memories_campaign_id ON chapter_memories(campaign_id);
CREATE INDEX idx_chapter_memories_chapter_id ON chapter_memories(chapter_id);
CREATE INDEX idx_chapter_memories_is_current ON chapter_memories(campaign_id, is_current) WHERE is_current = true;

CREATE TRIGGER update_chapter_memories_updated_at BEFORE UPDATE ON chapter_memories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Session Memories Table
-- Session-level details and scene records
-- ============================================
CREATE TABLE session_memories (
    id                 BIGSERIAL PRIMARY KEY,
    session_id         BIGINT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    memory_type        TEXT NOT NULL CHECK (memory_type IN (
                           'summary', 'scene', 'decision', 'discovery', 'moment'
                       )),
    title              TEXT,
    content            TEXT NOT NULL,
    scene_index        INT,
    entities_mentioned BIGINT[],
    importance         INT DEFAULT 5 CHECK (importance >= 1 AND importance <= 10),
    is_player_visible  BOOLEAN DEFAULT false,
    gm_edited          BOOLEAN DEFAULT false,
    created_at         TIMESTAMPTZ DEFAULT NOW(),
    updated_at         TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE session_memories IS 'Session-level memories including summaries, scenes, and key moments';
COMMENT ON COLUMN session_memories.session_id IS 'Session this memory belongs to';
COMMENT ON COLUMN session_memories.memory_type IS 'Type of memory: summary, scene, decision, discovery, moment';
COMMENT ON COLUMN session_memories.title IS 'Brief title for the memory';
COMMENT ON COLUMN session_memories.content IS 'Full text content of the memory';
COMMENT ON COLUMN session_memories.scene_index IS 'Position within session for scene-type memories';
COMMENT ON COLUMN session_memories.entities_mentioned IS 'Array of entity IDs referenced in this memory';
COMMENT ON COLUMN session_memories.importance IS 'Priority ranking from 1 (low) to 10 (high)';
COMMENT ON COLUMN session_memories.is_player_visible IS 'True if players can see this memory';
COMMENT ON COLUMN session_memories.gm_edited IS 'True if GM has edited the AI-generated content';

CREATE INDEX idx_session_memories_session_id ON session_memories(session_id);
CREATE INDEX idx_session_memories_type ON session_memories(memory_type);
CREATE INDEX idx_session_memories_scene_index ON session_memories(session_id, scene_index);
CREATE INDEX idx_session_memories_entities ON session_memories USING GIN(entities_mentioned);
CREATE INDEX idx_session_memories_importance ON session_memories(importance DESC);

CREATE TRIGGER update_session_memories_updated_at BEFORE UPDATE ON session_memories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Session Chat Logs Table
-- Complete conversation histories for AI sessions
-- ============================================
CREATE TABLE session_chat_logs (
    id            BIGSERIAL PRIMARY KEY,
    session_id    BIGINT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    message_index INT NOT NULL,
    role          TEXT NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    content       TEXT NOT NULL,
    tool_calls    JSONB,
    timestamp     TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT unique_session_message UNIQUE (session_id, message_index)
);

COMMENT ON TABLE session_chat_logs IS 'Complete conversation histories for AI-assisted sessions';
COMMENT ON COLUMN session_chat_logs.session_id IS 'Session this chat log belongs to';
COMMENT ON COLUMN session_chat_logs.message_index IS 'Sequential position of message within session (0-based)';
COMMENT ON COLUMN session_chat_logs.role IS 'Message author: user, assistant, or system';
COMMENT ON COLUMN session_chat_logs.content IS 'Full text content of the message';
COMMENT ON COLUMN session_chat_logs.tool_calls IS 'JSON array of tool calls made by the assistant';
COMMENT ON COLUMN session_chat_logs.timestamp IS 'When the message was sent';

CREATE INDEX idx_session_chat_logs_session_id ON session_chat_logs(session_id);
CREATE INDEX idx_session_chat_logs_timestamp ON session_chat_logs(session_id, timestamp);

-- ============================================
-- Memory Entity Extractions Table
-- Queue for detected entities pending GM review
-- ============================================
CREATE TABLE memory_entity_extractions (
    id                BIGSERIAL PRIMARY KEY,
    session_memory_id BIGINT NOT NULL REFERENCES session_memories(id) ON DELETE CASCADE,
    extracted_name    TEXT NOT NULL,
    extracted_type    TEXT CHECK (extracted_type IN (
                          'npc', 'location', 'item', 'faction', 'clue',
                          'creature', 'organization', 'event', 'document', 'other'
                      )),
    evidence_text     TEXT,
    matched_entity_id BIGINT REFERENCES entities(id) ON DELETE SET NULL,
    match_confidence  FLOAT CHECK (match_confidence >= 0.0 AND match_confidence <= 1.0),
    status            TEXT DEFAULT 'pending' CHECK (status IN (
                          'pending', 'approved', 'rejected', 'created'
                      )),
    created_at        TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE memory_entity_extractions IS 'Queue of AI-detected entities pending GM review';
COMMENT ON COLUMN memory_entity_extractions.session_memory_id IS 'Session memory where the entity was detected';
COMMENT ON COLUMN memory_entity_extractions.extracted_name IS 'Name of the detected entity';
COMMENT ON COLUMN memory_entity_extractions.extracted_type IS 'Detected entity type (matches entities.entity_type)';
COMMENT ON COLUMN memory_entity_extractions.evidence_text IS 'Text excerpt where the entity was mentioned';
COMMENT ON COLUMN memory_entity_extractions.matched_entity_id IS 'Existing entity if a match was found';
COMMENT ON COLUMN memory_entity_extractions.match_confidence IS 'Confidence score for entity match (0.0 to 1.0)';
COMMENT ON COLUMN memory_entity_extractions.status IS 'Review status: pending, approved, rejected, or created';

CREATE INDEX idx_memory_entity_extractions_memory_id ON memory_entity_extractions(session_memory_id);
CREATE INDEX idx_memory_entity_extractions_status ON memory_entity_extractions(status);
CREATE INDEX idx_memory_entity_extractions_matched ON memory_entity_extractions(matched_entity_id);

-- ============================================
-- Chapter Entities Table
-- Links entities to chapters with relationship type
-- ============================================
CREATE TABLE chapter_entities (
    id           BIGSERIAL PRIMARY KEY,
    chapter_id   BIGINT NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
    entity_id    BIGINT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    mention_type TEXT DEFAULT 'linked' CHECK (mention_type IN (
                     'linked', 'mentioned', 'featured'
                 )),
    created_at   TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT unique_chapter_entity UNIQUE (chapter_id, entity_id)
);

COMMENT ON TABLE chapter_entities IS 'Links entities to chapters for tracking mentions and associations';
COMMENT ON COLUMN chapter_entities.chapter_id IS 'Chapter containing the entity reference';
COMMENT ON COLUMN chapter_entities.entity_id IS 'Entity being referenced';
COMMENT ON COLUMN chapter_entities.mention_type IS 'How the entity is associated: linked (manual), mentioned (AI-detected), featured (primary)';

CREATE INDEX idx_chapter_entities_chapter ON chapter_entities(chapter_id);
CREATE INDEX idx_chapter_entities_entity ON chapter_entities(entity_id);
CREATE INDEX idx_chapter_entities_mention_type ON chapter_entities(mention_type);

-- ============================================
-- Session Entities Table
-- Links entities to individual game sessions
-- ============================================
CREATE TABLE session_entities (
    id         BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    entity_id  BIGINT NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    role       TEXT DEFAULT 'appeared' CHECK (role IN (
                   'appeared', 'introduced', 'major', 'minor'
               )),
    notes      TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT unique_session_entity UNIQUE (session_id, entity_id)
);

COMMENT ON TABLE session_entities IS 'Links entities to sessions for tracking appearances in individual game sessions';
COMMENT ON COLUMN session_entities.session_id IS 'Session where the entity appeared';
COMMENT ON COLUMN session_entities.entity_id IS 'Entity that appeared in the session';
COMMENT ON COLUMN session_entities.role IS 'How significant was the entity: appeared, introduced, major, minor';
COMMENT ON COLUMN session_entities.notes IS 'Optional context about the entity''s role in this session';

CREATE INDEX idx_session_entities_session ON session_entities(session_id);
CREATE INDEX idx_session_entities_entity ON session_entities(entity_id);
CREATE INDEX idx_session_entities_role ON session_entities(role);

-- ============================================
-- Helper Function: Get Inverse Relationship Type
-- ============================================
CREATE OR REPLACE FUNCTION get_inverse_relationship_type(
    p_campaign_id BIGINT,
    p_relationship_type TEXT
) RETURNS TEXT AS $$
DECLARE
    v_inverse_name TEXT;
BEGIN
    SELECT inverse_name INTO v_inverse_name
    FROM relationship_types
    WHERE campaign_id = p_campaign_id
      AND name = p_relationship_type;

    IF v_inverse_name IS NULL THEN
        SELECT inverse_name INTO v_inverse_name
        FROM relationship_types
        WHERE campaign_id IS NULL
          AND name = p_relationship_type;
    END IF;

    RETURN v_inverse_name;
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION get_inverse_relationship_type IS 'Returns inverse relationship type name, checking campaign-specific first then system defaults';

-- ============================================
-- Helper Function: Validate Relationship Type
-- ============================================
CREATE OR REPLACE FUNCTION validate_relationship_type(
    p_campaign_id BIGINT,
    p_relationship_type TEXT
) RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM relationship_types
        WHERE name = p_relationship_type
          AND (campaign_id = p_campaign_id OR campaign_id IS NULL)
    );
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION validate_relationship_type IS 'Returns TRUE if relationship type exists for campaign (including system defaults)';

-- ============================================
-- View: Available Relationship Types
-- All available types per campaign (defaults + custom)
-- ============================================
CREATE OR REPLACE VIEW available_relationship_types AS
SELECT DISTINCT ON (c.id, rt.name)
    c.id as campaign_id,
    rt.id as relationship_type_id,
    rt.name,
    rt.inverse_name,
    rt.is_symmetric,
    rt.display_label,
    rt.inverse_display_label,
    rt.description,
    rt.campaign_id IS NOT NULL as is_custom
FROM campaigns c
CROSS JOIN relationship_types rt
WHERE rt.campaign_id IS NULL OR rt.campaign_id = c.id
ORDER BY c.id, rt.name, rt.campaign_id NULLS LAST;

COMMENT ON VIEW available_relationship_types IS 'All available relationship types per campaign (system defaults + campaign-specific)';

-- ============================================
-- View: Entity Appearances
-- Unified view across chapters and sessions
-- ============================================
CREATE OR REPLACE VIEW entity_appearances AS
SELECT
    e.id AS entity_id,
    e.name AS entity_name,
    e.entity_type,
    e.campaign_id,
    'chapter' AS appearance_type,
    c.id AS container_id,
    c.title AS container_name,
    ce.mention_type AS role,
    NULL AS notes,
    ce.created_at
FROM entities e
JOIN chapter_entities ce ON ce.entity_id = e.id
JOIN chapters c ON c.id = ce.chapter_id
UNION ALL
SELECT
    e.id AS entity_id,
    e.name AS entity_name,
    e.entity_type,
    e.campaign_id,
    'session' AS appearance_type,
    s.id AS container_id,
    CONCAT('Session ', s.session_number) AS container_name,
    se.role,
    se.notes,
    se.created_at
FROM entities e
JOIN session_entities se ON se.entity_id = e.id
JOIN sessions s ON s.id = se.session_id;

COMMENT ON VIEW entity_appearances IS 'Unified view of all entity appearances across chapters and sessions';

-- ============================================
-- Vectorization Setup (Conditional)
-- Requires pgedge_vectorizer extension
-- ============================================
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_extension WHERE extname = 'pgedge_vectorizer'
    ) THEN
        -- entities.name
        PERFORM pgedge_vectorizer.enable_vectorization(
            source_table := 'entities',
            source_column := 'name',
            chunk_strategy := 'token_based',
            chunk_size := 100,
            chunk_overlap := 10,
            embedding_dimension := 1024
        );

        -- entities.description
        PERFORM pgedge_vectorizer.enable_vectorization(
            source_table := 'entities',
            source_column := 'description',
            chunk_strategy := 'markdown',
            chunk_size := 400,
            chunk_overlap := 50,
            embedding_dimension := 1024
        );

        -- chapters.overview
        PERFORM pgedge_vectorizer.enable_vectorization(
            source_table := 'chapters',
            source_column := 'overview',
            chunk_strategy := 'markdown',
            chunk_size := 300,
            chunk_overlap := 30,
            embedding_dimension := 1024
        );

        -- sessions.prep_notes
        PERFORM pgedge_vectorizer.enable_vectorization(
            source_table := 'sessions',
            source_column := 'prep_notes',
            chunk_strategy := 'markdown',
            chunk_size := 500,
            chunk_overlap := 50,
            embedding_dimension := 1024
        );

        -- sessions.actual_notes
        PERFORM pgedge_vectorizer.enable_vectorization(
            source_table := 'sessions',
            source_column := 'actual_notes',
            chunk_strategy := 'markdown',
            chunk_size := 500,
            chunk_overlap := 50,
            embedding_dimension := 1024
        );

        -- campaign_memories.content
        PERFORM pgedge_vectorizer.enable_vectorization(
            source_table := 'campaign_memories',
            source_column := 'content',
            chunk_strategy := 'markdown',
            chunk_size := 400,
            chunk_overlap := 50,
            embedding_dimension := 1024
        );

        -- Comments on auto-generated chunk tables
        COMMENT ON TABLE entities_name_chunks IS 'Auto-generated by pgedge_vectorizer for entity name embeddings';
        COMMENT ON TABLE entities_description_chunks IS 'Auto-generated by pgedge_vectorizer for entity description embeddings';
        COMMENT ON TABLE chapters_overview_chunks IS 'Auto-generated by pgedge_vectorizer for chapter overview embeddings';
        COMMENT ON TABLE sessions_prep_notes_chunks IS 'Auto-generated by pgedge_vectorizer for session prep notes embeddings';
        COMMENT ON TABLE sessions_actual_notes_chunks IS 'Auto-generated by pgedge_vectorizer for session actual notes embeddings';
        COMMENT ON TABLE campaign_memories_content_chunks IS 'Auto-generated by pgedge_vectorizer for campaign memory embeddings';

        RAISE NOTICE 'pgEdge vectorizer enabled on all content tables';
    ELSE
        RAISE NOTICE 'pgedge_vectorizer extension not found. Skipping vectorization setup.';
        RAISE NOTICE 'To enable semantic search, install the pgedge_vectorizer extension.';
    END IF;
END $$;

-- ============================================
-- Hybrid Search Function (Conditional)
-- Combines vector similarity with text search
-- ============================================
DO $$
BEGIN
    -- Only create function if pgedge_vectorizer extension exists
    IF EXISTS (
        SELECT 1 FROM pg_extension WHERE extname = 'pgedge_vectorizer'
    ) THEN
        -- Drop function if it already exists
        DROP FUNCTION IF EXISTS search_campaign_content(BIGINT, TEXT, INT);

        -- Create the hybrid search function using dynamic SQL
        -- This approach handles missing chunk tables gracefully
        EXECUTE $func$
        CREATE OR REPLACE FUNCTION search_campaign_content(
            p_campaign_id BIGINT,
            p_query TEXT,
            p_limit INT DEFAULT 10
        ) RETURNS TABLE (
            source_table TEXT,
            source_id BIGINT,
            source_name TEXT,
            chunk_content TEXT,
            vector_score FLOAT,
            combined_score FLOAT
        ) AS $body$
        DECLARE
            query_embedding vector;
            sql_query TEXT;
            union_parts TEXT[];
        BEGIN
            -- Generate embedding for the query once
            query_embedding := pgedge_vectorizer.generate_embedding(p_query);

            -- Build array of UNION parts based on which tables exist
            union_parts := ARRAY[]::TEXT[];

            -- Check entities_name_chunks
            IF EXISTS (
                SELECT 1 FROM information_schema.tables
                WHERE table_schema = 'public' AND table_name = 'entities_name_chunks'
            ) THEN
                union_parts := array_append(union_parts, $q$
                    SELECT
                        'entities' AS source_table,
                        e.id AS source_id,
                        e.name AS source_name,
                        c.chunk AS chunk_content,
                        (1 - (c.embedding <=> $1))::FLOAT AS vector_score,
                        (0.7 * (1 - (c.embedding <=> $1)) +
                         0.3 * ts_rank(to_tsvector('english', c.chunk), plainto_tsquery('english', $2)))::FLOAT AS combined_score
                    FROM entities_name_chunks c
                    JOIN entities e ON c.source_id = e.id
                    WHERE e.campaign_id = $3
                $q$);
            END IF;

            -- Check entities_description_chunks
            IF EXISTS (
                SELECT 1 FROM information_schema.tables
                WHERE table_schema = 'public' AND table_name = 'entities_description_chunks'
            ) THEN
                union_parts := array_append(union_parts, $q$
                    SELECT
                        'entities' AS source_table,
                        e.id AS source_id,
                        e.name AS source_name,
                        c.chunk AS chunk_content,
                        (1 - (c.embedding <=> $1))::FLOAT AS vector_score,
                        (0.7 * (1 - (c.embedding <=> $1)) +
                         0.3 * ts_rank(to_tsvector('english', c.chunk), plainto_tsquery('english', $2)))::FLOAT AS combined_score
                    FROM entities_description_chunks c
                    JOIN entities e ON c.source_id = e.id
                    WHERE e.campaign_id = $3
                $q$);
            END IF;

            -- Check chapters_overview_chunks
            IF EXISTS (
                SELECT 1 FROM information_schema.tables
                WHERE table_schema = 'public' AND table_name = 'chapters_overview_chunks'
            ) THEN
                union_parts := array_append(union_parts, $q$
                    SELECT
                        'chapters' AS source_table,
                        ch.id AS source_id,
                        ch.title AS source_name,
                        c.chunk AS chunk_content,
                        (1 - (c.embedding <=> $1))::FLOAT AS vector_score,
                        (0.7 * (1 - (c.embedding <=> $1)) +
                         0.3 * ts_rank(to_tsvector('english', c.chunk), plainto_tsquery('english', $2)))::FLOAT AS combined_score
                    FROM chapters_overview_chunks c
                    JOIN chapters ch ON c.source_id = ch.id
                    WHERE ch.campaign_id = $3
                $q$);
            END IF;

            -- Check sessions_prep_notes_chunks
            IF EXISTS (
                SELECT 1 FROM information_schema.tables
                WHERE table_schema = 'public' AND table_name = 'sessions_prep_notes_chunks'
            ) THEN
                union_parts := array_append(union_parts, $q$
                    SELECT
                        'sessions' AS source_table,
                        s.id AS source_id,
                        COALESCE(s.title, 'Session #' || s.session_number::TEXT) AS source_name,
                        c.chunk AS chunk_content,
                        (1 - (c.embedding <=> $1))::FLOAT AS vector_score,
                        (0.7 * (1 - (c.embedding <=> $1)) +
                         0.3 * ts_rank(to_tsvector('english', c.chunk), plainto_tsquery('english', $2)))::FLOAT AS combined_score
                    FROM sessions_prep_notes_chunks c
                    JOIN sessions s ON c.source_id = s.id
                    WHERE s.campaign_id = $3
                $q$);
            END IF;

            -- Check sessions_actual_notes_chunks
            IF EXISTS (
                SELECT 1 FROM information_schema.tables
                WHERE table_schema = 'public' AND table_name = 'sessions_actual_notes_chunks'
            ) THEN
                union_parts := array_append(union_parts, $q$
                    SELECT
                        'sessions' AS source_table,
                        s.id AS source_id,
                        COALESCE(s.title, 'Session #' || s.session_number::TEXT) AS source_name,
                        c.chunk AS chunk_content,
                        (1 - (c.embedding <=> $1))::FLOAT AS vector_score,
                        (0.7 * (1 - (c.embedding <=> $1)) +
                         0.3 * ts_rank(to_tsvector('english', c.chunk), plainto_tsquery('english', $2)))::FLOAT AS combined_score
                    FROM sessions_actual_notes_chunks c
                    JOIN sessions s ON c.source_id = s.id
                    WHERE s.campaign_id = $3
                $q$);
            END IF;

            -- Check campaign_memories_content_chunks
            IF EXISTS (
                SELECT 1 FROM information_schema.tables
                WHERE table_schema = 'public' AND table_name = 'campaign_memories_content_chunks'
            ) THEN
                union_parts := array_append(union_parts, $q$
                    SELECT
                        'campaign_memories' AS source_table,
                        cm.id AS source_id,
                        COALESCE(cm.title, cm.memory_type) AS source_name,
                        c.chunk AS chunk_content,
                        (1 - (c.embedding <=> $1))::FLOAT AS vector_score,
                        (0.7 * (1 - (c.embedding <=> $1)) +
                         0.3 * ts_rank(to_tsvector('english', c.chunk), plainto_tsquery('english', $2)))::FLOAT AS combined_score
                    FROM campaign_memories_content_chunks c
                    JOIN campaign_memories cm ON c.source_id = cm.id
                    WHERE cm.campaign_id = $3
                $q$);
            END IF;

            -- If no chunk tables exist, return empty result
            IF array_length(union_parts, 1) IS NULL THEN
                RETURN;
            END IF;

            -- Build and execute the final query
            sql_query := array_to_string(union_parts, ' UNION ALL ') ||
                        ' ORDER BY combined_score DESC LIMIT $4';

            RETURN QUERY EXECUTE sql_query
                USING query_embedding, p_query, p_campaign_id, p_limit;
        END;
        $body$ LANGUAGE plpgsql STABLE;
        $func$;

        COMMENT ON FUNCTION search_campaign_content(BIGINT, TEXT, INT) IS
            'Hybrid semantic + text search across all vectorized campaign content. Combines vector similarity (70%) with PostgreSQL text search (30%).';

        RAISE NOTICE 'Hybrid search function search_campaign_content created';
    ELSE
        RAISE NOTICE 'pgedge_vectorizer extension not found. Skipping search function creation.';
    END IF;
END $$;

-- ============================================
-- Record migration
-- ============================================
INSERT INTO schema_migrations (version) VALUES ('001_schema');
