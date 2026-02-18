/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

-- Migration 012: Server-Side Drafts
--
-- This migration:
-- 1. Creates the drafts table for server-side draft persistence
-- 2. Adds indexes for fast lookup by user/target and campaign/table
-- 3. Adds an auto-delete trigger to remove stale drafts when the
--    committed record (entity, chapter, or session) is updated

-- ============================================
-- Drafts Table
-- Server-side persistence of unsaved edits
-- ============================================
CREATE TABLE drafts (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    campaign_id     BIGINT NOT NULL
                    REFERENCES campaigns(id) ON DELETE CASCADE,
    user_id         BIGINT NOT NULL
                    REFERENCES users(id) ON DELETE CASCADE,
    source_table    TEXT NOT NULL,
    source_id       BIGINT NOT NULL DEFAULT 0,
    is_new          BOOLEAN NOT NULL DEFAULT FALSE,
    draft_data      JSONB NOT NULL,
    server_version  INT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE drafts IS 'Server-side persistence of unsaved editor changes, keyed by user, campaign, and source record';
COMMENT ON COLUMN drafts.campaign_id IS 'Campaign this draft belongs to';
COMMENT ON COLUMN drafts.user_id IS 'User who owns this draft';
COMMENT ON COLUMN drafts.source_table IS 'Table name of the record being edited (e.g., entities, chapters, sessions)';
COMMENT ON COLUMN drafts.source_id IS 'Primary key of the record being edited; 0 for new (unsaved) records';
COMMENT ON COLUMN drafts.is_new IS 'TRUE when the draft represents a brand-new record not yet committed';
COMMENT ON COLUMN drafts.draft_data IS 'JSONB snapshot of the in-progress form fields';
COMMENT ON COLUMN drafts.server_version IS 'Version of the committed record at the time the draft was created, used for conflict detection';
COMMENT ON COLUMN drafts.created_at IS 'Timestamp when the draft was first saved';
COMMENT ON COLUMN drafts.updated_at IS 'Timestamp of the most recent draft save';

-- ============================================
-- Indexes
-- ============================================
CREATE UNIQUE INDEX idx_drafts_user_target
    ON drafts (user_id, source_table, source_id, campaign_id);

CREATE INDEX idx_drafts_campaign_table
    ON drafts (campaign_id, source_table);

-- ============================================
-- Auto-update updated_at
-- ============================================
CREATE TRIGGER update_drafts_updated_at
    BEFORE UPDATE ON drafts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Stale Draft Auto-Delete Trigger
-- When a committed record is updated, any
-- matching draft is automatically removed so
-- users do not resume editing from stale data.
-- ============================================
CREATE OR REPLACE FUNCTION delete_stale_draft()
RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM drafts
    WHERE source_table = TG_ARGV[0]
      AND source_id = NEW.id
      AND campaign_id = NEW.campaign_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION delete_stale_draft() IS 'Trigger function that deletes drafts matching the updated source record, preventing stale draft resumption';

CREATE TRIGGER trg_entities_delete_draft
    AFTER UPDATE ON entities
    FOR EACH ROW
    EXECUTE FUNCTION delete_stale_draft('entities');

CREATE TRIGGER trg_chapters_delete_draft
    AFTER UPDATE ON chapters
    FOR EACH ROW
    EXECUTE FUNCTION delete_stale_draft('chapters');

CREATE TRIGGER trg_sessions_delete_draft
    AFTER UPDATE ON sessions
    FOR EACH ROW
    EXECUTE FUNCTION delete_stale_draft('sessions');

-- ============================================
-- Record migration
-- ============================================
INSERT INTO schema_migrations (version) VALUES ('012_server_side_drafts');
