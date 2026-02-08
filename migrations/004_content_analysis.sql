/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

-- Migration 004: Content Analysis
--
-- This migration:
-- 1. Creates content_analysis_jobs to track analysis runs per content field
-- 2. Creates content_analysis_items for individual detections within a job

-- ============================================
-- Content Analysis Jobs Table
-- Tracks analysis runs for campaign content fields
-- ============================================
CREATE TABLE content_analysis_jobs (
    id              BIGSERIAL PRIMARY KEY,
    campaign_id     BIGINT NOT NULL
                    REFERENCES campaigns(id) ON DELETE CASCADE,
    source_table    TEXT NOT NULL,
    source_id       BIGINT NOT NULL,
    source_field    TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'completed',
    total_items     INT NOT NULL DEFAULT 0,
    resolved_items  INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE content_analysis_jobs IS 'Tracks content analysis runs for campaign content fields';
COMMENT ON COLUMN content_analysis_jobs.id IS 'Auto-incrementing primary key';
COMMENT ON COLUMN content_analysis_jobs.campaign_id IS 'Campaign this analysis job belongs to';
COMMENT ON COLUMN content_analysis_jobs.source_table IS 'Name of the table containing the analysed content (e.g., sessions, chapters)';
COMMENT ON COLUMN content_analysis_jobs.source_id IS 'Row ID within source_table that was analysed';
COMMENT ON COLUMN content_analysis_jobs.source_field IS 'Column name within source_table that was analysed (e.g., prep_notes, overview)';
COMMENT ON COLUMN content_analysis_jobs.status IS 'Job status: completed, pending, failed';
COMMENT ON COLUMN content_analysis_jobs.total_items IS 'Total number of detected items in this analysis run';
COMMENT ON COLUMN content_analysis_jobs.resolved_items IS 'Number of items that have been reviewed or resolved';
COMMENT ON COLUMN content_analysis_jobs.created_at IS 'Timestamp when the analysis job was created';
COMMENT ON COLUMN content_analysis_jobs.updated_at IS 'Timestamp when the analysis job was last updated';

CREATE INDEX idx_content_analysis_jobs_field
    ON content_analysis_jobs(campaign_id, source_table, source_id, source_field);
CREATE INDEX idx_content_analysis_jobs_status
    ON content_analysis_jobs(campaign_id, status);

CREATE TRIGGER update_content_analysis_jobs_updated_at
    BEFORE UPDATE ON content_analysis_jobs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Content Analysis Items Table
-- Individual detections within an analysis job
-- ============================================
CREATE TABLE content_analysis_items (
    id                 BIGSERIAL PRIMARY KEY,
    job_id             BIGINT NOT NULL
                       REFERENCES content_analysis_jobs(id)
                       ON DELETE CASCADE,
    detection_type     TEXT NOT NULL,
    matched_text       TEXT NOT NULL,
    entity_id          BIGINT
                       REFERENCES entities(id) ON DELETE SET NULL,
    similarity         FLOAT,
    context_snippet    TEXT,
    position_start     INT,
    position_end       INT,
    resolution         TEXT NOT NULL DEFAULT 'pending',
    resolved_entity_id BIGINT
                       REFERENCES entities(id) ON DELETE SET NULL,
    resolved_at        TIMESTAMPTZ,
    created_at         TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE content_analysis_items IS 'Individual entity detections within a content analysis job';
COMMENT ON COLUMN content_analysis_items.id IS 'Auto-incrementing primary key';
COMMENT ON COLUMN content_analysis_items.job_id IS 'Analysis job this item belongs to';
COMMENT ON COLUMN content_analysis_items.detection_type IS 'How the entity was detected: wiki_link_resolved, wiki_link_unresolved, untagged_mention, misspelling';
COMMENT ON COLUMN content_analysis_items.matched_text IS 'Text that triggered the detection';
COMMENT ON COLUMN content_analysis_items.entity_id IS 'Existing entity matched by the detection (NULL if no match found)';
COMMENT ON COLUMN content_analysis_items.similarity IS 'Similarity score for vector or fuzzy matches (0.0 to 1.0)';
COMMENT ON COLUMN content_analysis_items.context_snippet IS 'Surrounding text providing context for the detection';
COMMENT ON COLUMN content_analysis_items.position_start IS 'Character offset where the matched text begins in the source field';
COMMENT ON COLUMN content_analysis_items.position_end IS 'Character offset where the matched text ends in the source field';
COMMENT ON COLUMN content_analysis_items.resolution IS 'Review status: pending, accepted, new_entity, dismissed';
COMMENT ON COLUMN content_analysis_items.resolved_entity_id IS 'Entity assigned after resolution (may differ from initial entity_id)';
COMMENT ON COLUMN content_analysis_items.resolved_at IS 'Timestamp when the item was resolved';
COMMENT ON COLUMN content_analysis_items.created_at IS 'Timestamp when the detection was recorded';

CREATE INDEX idx_content_analysis_items_resolution
    ON content_analysis_items(job_id, resolution);

-- ============================================
-- Record migration
-- ============================================
INSERT INTO schema_migrations (version) VALUES ('004_content_analysis');
