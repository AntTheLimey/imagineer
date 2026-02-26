/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

-- 005_add_failure_reason.sql
-- Adds a failure_reason column to content_analysis_jobs
-- so the frontend can display why a job failed.

ALTER TABLE content_analysis_jobs
    ADD COLUMN failure_reason TEXT;

COMMENT ON COLUMN content_analysis_jobs.failure_reason
    IS 'Human-readable reason when status is failed '
       '(e.g. API quota exceeded, rate limited).';

INSERT INTO schema_migrations (version) VALUES ('005_add_failure_reason');
