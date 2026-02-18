/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

-- Migration 013: Pipeline Tracking
--
-- This migration:
-- 1. Adds agent_name to content_analysis_items to record which analysis
--    agent produced the detection
-- 2. Adds pipeline_run_id to content_analysis_items to correlate items
--    produced within the same pipeline execution

-- ============================================
-- Add pipeline tracking columns
-- ============================================
ALTER TABLE content_analysis_items
    ADD COLUMN IF NOT EXISTS agent_name TEXT;

ALTER TABLE content_analysis_items
    ADD COLUMN IF NOT EXISTS pipeline_run_id BIGINT;

COMMENT ON COLUMN content_analysis_items.agent_name IS
    'Name of the analysis agent that produced this detection (e.g., entity_detector, enrichment_agent)';

COMMENT ON COLUMN content_analysis_items.pipeline_run_id IS
    'Identifier correlating all items produced within the same pipeline execution';

-- ============================================
-- Record migration
-- ============================================
INSERT INTO schema_migrations (version) VALUES ('013_pipeline_tracking');
