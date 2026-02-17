/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

-- Migration 011: Session Workflows Phase 1
--
-- This migration:
-- 1. Creates scenes table for structured session planning
-- 2. Creates session_chat_messages table for session AI chat
-- 3. Adds play_notes column to sessions
-- 4. Drops JSONB columns from sessions (planned_scenes, discoveries, player_decisions, consequences)
-- 5. Enables vectorization on new content columns
-- 6. Updates search_campaign_content() for new tables

-- ============================================
-- Scenes Table
-- Structured scenes within game sessions
-- ============================================
CREATE TABLE scenes (
    id                BIGSERIAL PRIMARY KEY,
    session_id        BIGINT NOT NULL
                      REFERENCES sessions(id) ON DELETE CASCADE,
    campaign_id       BIGINT NOT NULL
                      REFERENCES campaigns(id) ON DELETE CASCADE,
    title             TEXT NOT NULL,
    description       TEXT,
    scene_type        TEXT NOT NULL DEFAULT 'other',
    status            TEXT NOT NULL DEFAULT 'planned',
    sort_order        INT NOT NULL DEFAULT 0,
    objective         TEXT,
    gm_notes          TEXT,
    entity_ids        BIGINT[] DEFAULT '{}',
    system_data       JSONB DEFAULT '{}',
    source            TEXT NOT NULL DEFAULT 'manual',
    source_confidence TEXT NOT NULL DEFAULT 'DRAFT',
    connections       JSONB DEFAULT '[]',
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE scenes IS 'Structured scenes within game sessions for planning and tracking';
COMMENT ON COLUMN scenes.scene_type IS 'Type of scene: exploration, combat, social, puzzle, chase, travel, downtime, other';
COMMENT ON COLUMN scenes.status IS 'Scene status: planned, active, completed, skipped';
COMMENT ON COLUMN scenes.sort_order IS 'Order within session for display (0-based)';
COMMENT ON COLUMN scenes.entity_ids IS 'Array of entity IDs involved in this scene';
COMMENT ON COLUMN scenes.source_confidence IS 'Canon status: DRAFT, AUTHORITATIVE, or SUPERSEDED';
COMMENT ON COLUMN scenes.connections IS 'JSON array of connections to other scenes or entities';

CREATE INDEX idx_scenes_session ON scenes(session_id, sort_order);
CREATE INDEX idx_scenes_campaign ON scenes(campaign_id);

CREATE TRIGGER update_scenes_updated_at
    BEFORE UPDATE ON scenes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Add play_notes to sessions
-- ============================================
ALTER TABLE sessions ADD COLUMN play_notes TEXT;

COMMENT ON COLUMN sessions.play_notes IS 'Notes taken during active play';

-- ============================================
-- Session Chat Messages Table
-- AI chat messages within session workflow
-- ============================================
CREATE TABLE session_chat_messages (
    id          BIGSERIAL PRIMARY KEY,
    session_id  BIGINT NOT NULL
                REFERENCES sessions(id) ON DELETE CASCADE,
    campaign_id BIGINT NOT NULL
                REFERENCES campaigns(id) ON DELETE CASCADE,
    role        TEXT NOT NULL,
    content     TEXT NOT NULL,
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE session_chat_messages IS 'AI chat messages within session workflow for prep, play, and wrap-up';
COMMENT ON COLUMN session_chat_messages.role IS 'Message author: user or assistant';

CREATE INDEX idx_session_chat_session ON session_chat_messages(session_id, sort_order);
CREATE INDEX idx_session_chat_campaign ON session_chat_messages(campaign_id);

-- ============================================
-- Drop JSONB columns from sessions
-- Replaced by scenes table and other structures
-- ============================================
ALTER TABLE sessions
    DROP COLUMN planned_scenes,
    DROP COLUMN discoveries,
    DROP COLUMN player_decisions,
    DROP COLUMN consequences;

-- ============================================
-- Vectorization Setup (Conditional)
-- Requires pgedge_vectorizer extension
-- ============================================
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_extension WHERE extname = 'pgedge_vectorizer'
    ) THEN
        RAISE NOTICE 'pgedge_vectorizer not found. Skipping.';
        RETURN;
    END IF;

    -- scenes.description vectorization
    PERFORM pgedge_vectorizer.enable_vectorization(
        source_table := 'scenes',
        source_column := 'description',
        chunk_strategy := 'markdown',
        chunk_size := 400,
        chunk_overlap := 50,
        embedding_dimension := 1024
    );

    -- scenes.gm_notes vectorization
    PERFORM pgedge_vectorizer.enable_vectorization(
        source_table := 'scenes',
        source_column := 'gm_notes',
        chunk_strategy := 'markdown',
        chunk_size := 400,
        chunk_overlap := 50,
        embedding_dimension := 1024
    );

    -- session_chat_messages.content vectorization
    PERFORM pgedge_vectorizer.enable_vectorization(
        source_table := 'session_chat_messages',
        source_column := 'content',
        chunk_strategy := 'token_based',
        chunk_size := 200,
        chunk_overlap := 20,
        embedding_dimension := 1024
    );

    RAISE NOTICE 'Vectorization enabled for scenes and session_chat_messages';
END $$;

-- ============================================
-- Replace search_campaign_content()
-- Add: scenes and session_chat_messages branches
-- ============================================
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_extension WHERE extname = 'pgedge_vectorizer'
    ) THEN
        RETURN;
    END IF;

    DROP FUNCTION IF EXISTS search_campaign_content(BIGINT, TEXT, INT);

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
        query_embedding := pgedge_vectorizer.generate_embedding(p_query);
        union_parts := ARRAY[]::TEXT[];

        -- campaigns_description_chunks
        IF EXISTS (
            SELECT 1 FROM information_schema.tables
            WHERE table_schema = 'public'
              AND table_name = 'campaigns_description_chunks'
        ) THEN
            union_parts := array_append(union_parts, $q$
                SELECT
                    'campaigns' AS source_table,
                    camp.id AS source_id,
                    camp.name AS source_name,
                    c.content AS chunk_content,
                    (1 - (c.embedding <=> $1))::FLOAT AS vector_score,
                    (0.7 * (1 - (c.embedding <=> $1)) +
                     0.3 * ts_rank(to_tsvector('english', c.content),
                        plainto_tsquery('english', $2)))::FLOAT
                        AS combined_score
                FROM campaigns_description_chunks c
                JOIN campaigns camp ON c.source_id = camp.id
                WHERE camp.id = $3
            $q$);
        END IF;

        -- entities_name_chunks
        IF EXISTS (
            SELECT 1 FROM information_schema.tables
            WHERE table_schema = 'public'
              AND table_name = 'entities_name_chunks'
        ) THEN
            union_parts := array_append(union_parts, $q$
                SELECT
                    'entities' AS source_table,
                    e.id AS source_id,
                    e.name AS source_name,
                    c.content AS chunk_content,
                    (1 - (c.embedding <=> $1))::FLOAT AS vector_score,
                    (0.7 * (1 - (c.embedding <=> $1)) +
                     0.3 * ts_rank(to_tsvector('english', c.content),
                        plainto_tsquery('english', $2)))::FLOAT
                        AS combined_score
                FROM entities_name_chunks c
                JOIN entities e ON c.source_id = e.id
                WHERE e.campaign_id = $3
            $q$);
        END IF;

        -- entities_description_chunks
        IF EXISTS (
            SELECT 1 FROM information_schema.tables
            WHERE table_schema = 'public'
              AND table_name = 'entities_description_chunks'
        ) THEN
            union_parts := array_append(union_parts, $q$
                SELECT
                    'entities' AS source_table,
                    e.id AS source_id,
                    e.name AS source_name,
                    c.content AS chunk_content,
                    (1 - (c.embedding <=> $1))::FLOAT AS vector_score,
                    (0.7 * (1 - (c.embedding <=> $1)) +
                     0.3 * ts_rank(to_tsvector('english', c.content),
                        plainto_tsquery('english', $2)))::FLOAT
                        AS combined_score
                FROM entities_description_chunks c
                JOIN entities e ON c.source_id = e.id
                WHERE e.campaign_id = $3
            $q$);
        END IF;

        -- chapters_overview_chunks
        IF EXISTS (
            SELECT 1 FROM information_schema.tables
            WHERE table_schema = 'public'
              AND table_name = 'chapters_overview_chunks'
        ) THEN
            union_parts := array_append(union_parts, $q$
                SELECT
                    'chapters' AS source_table,
                    ch.id AS source_id,
                    ch.title AS source_name,
                    c.content AS chunk_content,
                    (1 - (c.embedding <=> $1))::FLOAT AS vector_score,
                    (0.7 * (1 - (c.embedding <=> $1)) +
                     0.3 * ts_rank(to_tsvector('english', c.content),
                        plainto_tsquery('english', $2)))::FLOAT
                        AS combined_score
                FROM chapters_overview_chunks c
                JOIN chapters ch ON c.source_id = ch.id
                WHERE ch.campaign_id = $3
            $q$);
        END IF;

        -- sessions_prep_notes_chunks
        IF EXISTS (
            SELECT 1 FROM information_schema.tables
            WHERE table_schema = 'public'
              AND table_name = 'sessions_prep_notes_chunks'
        ) THEN
            union_parts := array_append(union_parts, $q$
                SELECT
                    'sessions' AS source_table,
                    s.id AS source_id,
                    COALESCE(s.title, 'Session #' ||
                        s.session_number::TEXT) AS source_name,
                    c.content AS chunk_content,
                    (1 - (c.embedding <=> $1))::FLOAT AS vector_score,
                    (0.7 * (1 - (c.embedding <=> $1)) +
                     0.3 * ts_rank(to_tsvector('english', c.content),
                        plainto_tsquery('english', $2)))::FLOAT
                        AS combined_score
                FROM sessions_prep_notes_chunks c
                JOIN sessions s ON c.source_id = s.id
                WHERE s.campaign_id = $3
            $q$);
        END IF;

        -- sessions_actual_notes_chunks
        IF EXISTS (
            SELECT 1 FROM information_schema.tables
            WHERE table_schema = 'public'
              AND table_name = 'sessions_actual_notes_chunks'
        ) THEN
            union_parts := array_append(union_parts, $q$
                SELECT
                    'sessions' AS source_table,
                    s.id AS source_id,
                    COALESCE(s.title, 'Session #' ||
                        s.session_number::TEXT) AS source_name,
                    c.content AS chunk_content,
                    (1 - (c.embedding <=> $1))::FLOAT AS vector_score,
                    (0.7 * (1 - (c.embedding <=> $1)) +
                     0.3 * ts_rank(to_tsvector('english', c.content),
                        plainto_tsquery('english', $2)))::FLOAT
                        AS combined_score
                FROM sessions_actual_notes_chunks c
                JOIN sessions s ON c.source_id = s.id
                WHERE s.campaign_id = $3
            $q$);
        END IF;

        -- campaign_memories_content_chunks
        IF EXISTS (
            SELECT 1 FROM information_schema.tables
            WHERE table_schema = 'public'
              AND table_name = 'campaign_memories_content_chunks'
        ) THEN
            union_parts := array_append(union_parts, $q$
                SELECT
                    'campaign_memories' AS source_table,
                    cm.id AS source_id,
                    COALESCE(cm.title, cm.memory_type) AS source_name,
                    c.content AS chunk_content,
                    (1 - (c.embedding <=> $1))::FLOAT AS vector_score,
                    (0.7 * (1 - (c.embedding <=> $1)) +
                     0.3 * ts_rank(to_tsvector('english', c.content),
                        plainto_tsquery('english', $2)))::FLOAT
                        AS combined_score
                FROM campaign_memories_content_chunks c
                JOIN campaign_memories cm ON c.source_id = cm.id
                WHERE cm.campaign_id = $3
            $q$);
        END IF;

        -- scenes_description_chunks
        IF EXISTS (
            SELECT 1 FROM information_schema.tables
            WHERE table_schema = 'public'
              AND table_name = 'scenes_description_chunks'
        ) THEN
            union_parts := array_append(union_parts, $q$
                SELECT
                    'scenes' AS source_table,
                    s.id AS source_id,
                    s.title AS source_name,
                    c.content AS chunk_content,
                    (1 - (c.embedding <=> $1))::FLOAT AS vector_score,
                    (0.7 * (1 - (c.embedding <=> $1)) +
                     0.3 * ts_rank(to_tsvector('english', c.content),
                        plainto_tsquery('english', $2)))::FLOAT
                        AS combined_score
                FROM scenes_description_chunks c
                JOIN scenes s ON c.source_id = s.id
                WHERE s.campaign_id = $3
            $q$);
        END IF;

        -- scenes_gm_notes_chunks
        IF EXISTS (
            SELECT 1 FROM information_schema.tables
            WHERE table_schema = 'public'
              AND table_name = 'scenes_gm_notes_chunks'
        ) THEN
            union_parts := array_append(union_parts, $q$
                SELECT
                    'scenes' AS source_table,
                    s.id AS source_id,
                    s.title AS source_name,
                    c.content AS chunk_content,
                    (1 - (c.embedding <=> $1))::FLOAT AS vector_score,
                    (0.7 * (1 - (c.embedding <=> $1)) +
                     0.3 * ts_rank(to_tsvector('english', c.content),
                        plainto_tsquery('english', $2)))::FLOAT
                        AS combined_score
                FROM scenes_gm_notes_chunks c
                JOIN scenes s ON c.source_id = s.id
                WHERE s.campaign_id = $3
            $q$);
        END IF;

        -- session_chat_messages_content_chunks
        IF EXISTS (
            SELECT 1 FROM information_schema.tables
            WHERE table_schema = 'public'
              AND table_name = 'session_chat_messages_content_chunks'
        ) THEN
            union_parts := array_append(union_parts, $q$
                SELECT
                    'session_chat_messages' AS source_table,
                    m.id AS source_id,
                    m.role || ': ' || LEFT(m.content, 50) AS source_name,
                    c.content AS chunk_content,
                    (1 - (c.embedding <=> $1))::FLOAT AS vector_score,
                    (0.7 * (1 - (c.embedding <=> $1)) +
                     0.3 * ts_rank(to_tsvector('english', c.content),
                        plainto_tsquery('english', $2)))::FLOAT
                        AS combined_score
                FROM session_chat_messages_content_chunks c
                JOIN session_chat_messages m ON c.source_id = m.id
                WHERE m.campaign_id = $3
            $q$);
        END IF;

        IF array_length(union_parts, 1) IS NULL THEN
            RETURN;
        END IF;

        sql_query := array_to_string(union_parts, ' UNION ALL ') ||
                    ' ORDER BY combined_score DESC LIMIT $4';

        RETURN QUERY EXECUTE sql_query
            USING query_embedding, p_query, p_campaign_id, p_limit;
    END;
    $body$ LANGUAGE plpgsql STABLE;
    $func$;

    COMMENT ON FUNCTION search_campaign_content(BIGINT, TEXT, INT) IS
        'Hybrid semantic + text search across all vectorized campaign content. Combines vector similarity (70%) with PostgreSQL text search (30%).';

    RAISE NOTICE 'search_campaign_content() replaced with scenes and chat message branches';
END $$;

-- ============================================
-- Record migration
-- ============================================
INSERT INTO schema_migrations (version) VALUES ('011_session_workflows');
