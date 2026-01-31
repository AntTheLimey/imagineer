-- Imagineer Migration: Rename keeper_notes to gm_notes
-- Changes "Keeper" terminology (Call of Cthulhu specific) to generic "GM"

-- ============================================
-- UP Migration
-- ============================================

-- Rename the column from keeper_notes to gm_notes
ALTER TABLE entities RENAME COLUMN keeper_notes TO gm_notes;

-- Update the column comment
COMMENT ON COLUMN entities.gm_notes IS 'GM-only notes, never shown to players';

-- Record migration
INSERT INTO schema_migrations (version) VALUES ('004_rename_keeper_to_gm');

-- ============================================
-- DOWN Migration (for reference, run manually if needed)
-- ============================================
-- ALTER TABLE entities RENAME COLUMN gm_notes TO keeper_notes;
-- COMMENT ON COLUMN entities.keeper_notes IS 'GM-only notes, never shown to players';
-- DELETE FROM schema_migrations WHERE version = '004_rename_keeper_to_gm';
