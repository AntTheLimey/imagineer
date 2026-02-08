/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

-- Migration 003: Campaign Description Vectorization
--
-- This migration:
-- 1. Adds vectorization on campaigns.description (not in 001_schema.sql)
-- 2. Replaces search_campaign_content() to add campaigns branch
--    and fix c.chunk -> c.content (actual column name)

-- ============================================
-- Campaign Description Vectorization
-- ============================================
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_extension WHERE extname = 'pgedge_vectorizer'
    ) THEN
        RAISE NOTICE 'pgedge_vectorizer not found. Skipping.';
        RETURN;
    END IF;

    -- Add campaigns.description vectorization (not in 001_schema.sql)
    PERFORM pgedge_vectorizer.enable_vectorization(
        source_table := 'campaigns',
        source_column := 'description',
        chunk_strategy := 'markdown',
        chunk_size := 200,
        chunk_overlap := 25,
        embedding_dimension := 1024
    );

    COMMENT ON TABLE campaigns_description_chunks IS
        'pgedge_vectorizer: campaign description embeddings';

    RAISE NOTICE 'campaigns.description vectorization enabled';
END $$;

-- ============================================
-- Replace search_campaign_content()
-- Fix: c.chunk -> c.content (actual column name)
-- Add: campaigns_description_chunks branch
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

    RAISE NOTICE 'search_campaign_content() replaced with fixes';
END $$;

-- ============================================
-- Record migration
-- ============================================
INSERT INTO schema_migrations (version) VALUES ('003_campaign_vectorization');
