/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

-- 006_eras.sql — named ranges on the world line (spec §4.7).
CREATE TABLE world.eras (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    world_id    UUID NOT NULL REFERENCES world.worlds(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    span        NUMRANGE NOT NULL,
    sequence    INT NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (world_id, sequence)
);
COMMENT ON TABLE world.eras IS 'Named hour-ranges on the world line. Fuzziness is range width — no scale enum (spec §4.7).';
COMMENT ON COLUMN world.eras.sequence IS 'Display and sort order of eras within a world. Unique per world; independent of span, which may overlap or leave gaps.';
CREATE UNIQUE INDEX eras_name ON world.eras (world_id, lower(name));
CREATE TRIGGER touch BEFORE UPDATE ON world.eras
    FOR EACH ROW EXECUTE FUNCTION core.touch_updated_at();
