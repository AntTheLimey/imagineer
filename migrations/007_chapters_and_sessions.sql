/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

-- Chapters and Sessions Migration
-- Adds chapters for session organization, enhances sessions table,
-- and creates the AI memory system tables

-- ============================================
-- Chapters Table
-- Organize sessions into story arcs
-- ============================================
CREATE TABLE chapters (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
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

-- Indexes for chapters
CREATE INDEX idx_chapters_campaign_id ON chapters(campaign_id);
CREATE INDEX idx_chapters_sort_order ON chapters(campaign_id, sort_order);

-- ============================================
-- Sessions Table Modifications
-- Add chapter reference, workflow stage, and title
-- ============================================

-- Add chapter reference (nullable for existing sessions)
ALTER TABLE sessions
    ADD COLUMN chapter_id UUID REFERENCES chapters(id) ON DELETE SET NULL;

-- Add workflow stage for session lifecycle
ALTER TABLE sessions
    ADD COLUMN stage TEXT DEFAULT 'prep' CHECK (stage IN ('prep', 'play', 'wrap_up'));

-- Add session title for descriptive naming
ALTER TABLE sessions
    ADD COLUMN title TEXT;

COMMENT ON COLUMN sessions.chapter_id IS 'Chapter this session belongs to (NULL for uncategorized)';
COMMENT ON COLUMN sessions.stage IS 'Workflow stage: prep (preparation), play (active session), wrap_up (post-session notes)';
COMMENT ON COLUMN sessions.title IS 'Session title (e.g., "The Haunted Mansion")';

-- Index for chapter lookups
CREATE INDEX idx_sessions_chapter_id ON sessions(chapter_id);

-- ============================================
-- Campaign Memories Table
-- Long-term facts and knowledge for a campaign
-- ============================================
CREATE TABLE campaign_memories (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    campaign_id       UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    memory_type       TEXT NOT NULL CHECK (memory_type IN (
                          'premise', 'theme', 'faction_summary', 'plot_thread', 'gm_note'
                      )),
    title             TEXT,
    content           TEXT NOT NULL,
    source_session_id UUID REFERENCES sessions(id) ON DELETE SET NULL,
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

-- Indexes for campaign_memories
CREATE INDEX idx_campaign_memories_campaign_id ON campaign_memories(campaign_id);
CREATE INDEX idx_campaign_memories_type ON campaign_memories(memory_type);
CREATE INDEX idx_campaign_memories_importance ON campaign_memories(importance DESC);
CREATE INDEX idx_campaign_memories_source_session ON campaign_memories(source_session_id);

-- ============================================
-- Chapter Memories Table
-- Arc-level context for campaign chapters
-- ============================================
CREATE TABLE chapter_memories (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    campaign_id    UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    chapter_id     UUID NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
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

-- Indexes for chapter_memories
CREATE INDEX idx_chapter_memories_campaign_id ON chapter_memories(campaign_id);
CREATE INDEX idx_chapter_memories_chapter_id ON chapter_memories(chapter_id);
CREATE INDEX idx_chapter_memories_is_current ON chapter_memories(campaign_id, is_current) WHERE is_current = true;

-- ============================================
-- Session Memories Table
-- Session-level details and scene records
-- ============================================
CREATE TABLE session_memories (
    id                 UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id         UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    memory_type        TEXT NOT NULL CHECK (memory_type IN (
                           'summary', 'scene', 'decision', 'discovery', 'moment'
                       )),
    title              TEXT,
    content            TEXT NOT NULL,
    scene_index        INT,
    entities_mentioned UUID[],
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
COMMENT ON COLUMN session_memories.entities_mentioned IS 'Array of entity UUIDs referenced in this memory';
COMMENT ON COLUMN session_memories.importance IS 'Priority ranking from 1 (low) to 10 (high)';
COMMENT ON COLUMN session_memories.is_player_visible IS 'True if players can see this memory';
COMMENT ON COLUMN session_memories.gm_edited IS 'True if GM has edited the AI-generated content';

-- Indexes for session_memories
CREATE INDEX idx_session_memories_session_id ON session_memories(session_id);
CREATE INDEX idx_session_memories_type ON session_memories(memory_type);
CREATE INDEX idx_session_memories_scene_index ON session_memories(session_id, scene_index);
CREATE INDEX idx_session_memories_entities ON session_memories USING GIN(entities_mentioned);
CREATE INDEX idx_session_memories_importance ON session_memories(importance DESC);

-- ============================================
-- Session Chat Logs Table
-- Complete conversation histories for AI sessions
-- ============================================
CREATE TABLE session_chat_logs (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id    UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
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

-- Indexes for session_chat_logs
CREATE INDEX idx_session_chat_logs_session_id ON session_chat_logs(session_id);
CREATE INDEX idx_session_chat_logs_timestamp ON session_chat_logs(session_id, timestamp);

-- ============================================
-- Memory Entity Extractions Table
-- Queue for detected entities pending GM review
-- ============================================
CREATE TABLE memory_entity_extractions (
    id                 UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_memory_id  UUID NOT NULL REFERENCES session_memories(id) ON DELETE CASCADE,
    extracted_name     TEXT NOT NULL,
    extracted_type     TEXT CHECK (extracted_type IN (
                           'npc', 'location', 'item', 'faction', 'clue',
                           'creature', 'organization', 'event', 'document', 'other'
                       )),
    evidence_text      TEXT,
    matched_entity_id  UUID REFERENCES entities(id) ON DELETE SET NULL,
    match_confidence   FLOAT CHECK (match_confidence >= 0.0 AND match_confidence <= 1.0),
    status             TEXT DEFAULT 'pending' CHECK (status IN (
                           'pending', 'approved', 'rejected', 'created'
                       )),
    created_at         TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE memory_entity_extractions IS 'Queue of AI-detected entities pending GM review';
COMMENT ON COLUMN memory_entity_extractions.session_memory_id IS 'Session memory where the entity was detected';
COMMENT ON COLUMN memory_entity_extractions.extracted_name IS 'Name of the detected entity';
COMMENT ON COLUMN memory_entity_extractions.extracted_type IS 'Detected entity type (matches entities.entity_type)';
COMMENT ON COLUMN memory_entity_extractions.evidence_text IS 'Text excerpt where the entity was mentioned';
COMMENT ON COLUMN memory_entity_extractions.matched_entity_id IS 'Existing entity if a match was found';
COMMENT ON COLUMN memory_entity_extractions.match_confidence IS 'Confidence score for entity match (0.0 to 1.0)';
COMMENT ON COLUMN memory_entity_extractions.status IS 'Review status: pending, approved, rejected, or created';

-- Indexes for memory_entity_extractions
CREATE INDEX idx_memory_entity_extractions_memory_id ON memory_entity_extractions(session_memory_id);
CREATE INDEX idx_memory_entity_extractions_status ON memory_entity_extractions(status);
CREATE INDEX idx_memory_entity_extractions_matched ON memory_entity_extractions(matched_entity_id);

-- ============================================
-- Updated Timestamp Triggers
-- ============================================

-- Trigger for chapters
CREATE TRIGGER update_chapters_updated_at
    BEFORE UPDATE ON chapters
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger for campaign_memories
CREATE TRIGGER update_campaign_memories_updated_at
    BEFORE UPDATE ON campaign_memories
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger for chapter_memories
CREATE TRIGGER update_chapter_memories_updated_at
    BEFORE UPDATE ON chapter_memories
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger for session_memories
CREATE TRIGGER update_session_memories_updated_at
    BEFORE UPDATE ON session_memories
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Record migration
-- ============================================
INSERT INTO schema_migrations (version) VALUES ('007_chapters_and_sessions');
