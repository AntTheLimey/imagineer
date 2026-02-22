/*-------------------------------------------------------------------------
 *
 * 003_job_phases_and_draft_lifecycle.sql
 *
 * Adds job phases junction table, current_phase tracking on
 * content_analysis_jobs, and revision_count/status lifecycle
 * columns on drafts.
 *
 *-------------------------------------------------------------------------
 */

-- ============================================
-- Part A: Job Phases
-- ============================================

-- Junction table linking analysis jobs to their phases
CREATE TABLE job_phases (
    job_id     BIGINT NOT NULL
               REFERENCES content_analysis_jobs(id)
               ON DELETE CASCADE,
    phase_key  TEXT NOT NULL
               CHECK (phase_key IN ('identify', 'revise', 'enrich')),
    sort_order INT NOT NULL DEFAULT 0,
    PRIMARY KEY (job_id, phase_key)
);

COMMENT ON TABLE job_phases IS 'Junction table linking content analysis jobs to their processing phases';
COMMENT ON COLUMN job_phases.job_id IS 'Analysis job this phase belongs to';
COMMENT ON COLUMN job_phases.phase_key IS 'Phase identifier: identify, revise, or enrich';
COMMENT ON COLUMN job_phases.sort_order IS 'Execution order of this phase within the job (0-based)';

CREATE INDEX idx_job_phases_phase
    ON job_phases (phase_key);

COMMENT ON INDEX idx_job_phases_phase IS 'Supports lookups by phase key across all jobs';

-- Track which phase a job is currently executing
ALTER TABLE content_analysis_jobs
    ADD COLUMN current_phase TEXT
        CHECK (current_phase IN ('identify', 'revise', 'enrich'));

COMMENT ON COLUMN content_analysis_jobs.current_phase IS 'Phase the job is currently executing (NULL when not running or completed)';

-- Backfill existing jobs with the default identify phase
INSERT INTO job_phases (job_id, phase_key, sort_order)
SELECT id, 'identify', 0
FROM content_analysis_jobs;

-- Backfill enrich phase for jobs that already have enrichment items
INSERT INTO job_phases (job_id, phase_key, sort_order)
SELECT DISTINCT i.job_id, 'enrich', 1
FROM content_analysis_items i
WHERE i.phase = 'enrichment'
  AND NOT EXISTS (
      SELECT 1 FROM job_phases jp
      WHERE jp.job_id = i.job_id AND jp.phase_key = 'enrich'
  );

-- Backfill revise phase for jobs that already have analysis items
INSERT INTO job_phases (job_id, phase_key, sort_order)
SELECT DISTINCT i.job_id, 'revise', 2
FROM content_analysis_items i
WHERE i.phase = 'analysis'
  AND NOT EXISTS (
      SELECT 1 FROM job_phases jp
      WHERE jp.job_id = i.job_id AND jp.phase_key = 'revise'
  );

-- ============================================
-- Part B: Draft Lifecycle
-- ============================================

-- Add revision tracking and lifecycle status to drafts
ALTER TABLE drafts
    ADD COLUMN revision_count INT NOT NULL DEFAULT 0,
    ADD COLUMN status TEXT NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'stale', 'abandoned'));

COMMENT ON COLUMN drafts.revision_count IS 'Number of times the draft has been saved (incremented on each update)';
COMMENT ON COLUMN drafts.status IS 'Lifecycle status of the draft: active (in progress) or other future states (e.g., stale, abandoned)';

CREATE INDEX idx_drafts_cleanup
    ON drafts (status, updated_at);

COMMENT ON INDEX idx_drafts_cleanup IS 'Supports periodic cleanup queries that find drafts by status and age';

-- ============================================
-- Record migration
-- ============================================
INSERT INTO schema_migrations (version) VALUES ('003_job_phases_and_draft_lifecycle');
