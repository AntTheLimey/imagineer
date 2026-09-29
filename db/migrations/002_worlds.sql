/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

-- 002_worlds.sql — worlds and calendars (spec §4.1, §4.2, §7).
CREATE TABLE world.worlds (
    id            UUID PRIMARY KEY DEFAULT uuidv7(),
    owner_id      UUID NOT NULL REFERENCES app.users(id),
    name          TEXT NOT NULL,
    description   TEXT,
    system_id     UUID,
    genre         TEXT,
    is_template   BOOLEAN NOT NULL DEFAULT false,
    forked_from   UUID REFERENCES world.worlds(id),
    fork_metadata JSONB NOT NULL DEFAULT '{}',
    deleted_at    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE world.worlds IS 'One row per world, template or live. Creation is fork (spec §8).';
COMMENT ON COLUMN world.worlds.system_id IS 'FK to world.system_packs added in Plan 02.';
COMMENT ON COLUMN world.worlds.forked_from IS 'Lineage; NULL only for the base template.';
CREATE INDEX ON world.worlds (owner_id);
CREATE TRIGGER touch BEFORE UPDATE ON world.worlds
    FOR EACH ROW EXECUTE FUNCTION core.touch_updated_at();

CREATE TABLE world.calendars (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    world_id        UUID NOT NULL REFERENCES world.worlds(id) ON DELETE CASCADE,
    is_primary      BOOLEAN NOT NULL DEFAULT false,
    epoch_label     TEXT NOT NULL,
    display_mapping JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE world.calendars IS 'Meaning of the numeric hours line (spec §7). Unit is always hours.';
COMMENT ON COLUMN world.calendars.display_mapping IS 'Named scheme ("gregorian-rata-die") or table-driven months/weekdays with hours-per-day.';
CREATE INDEX ON world.calendars (world_id);
CREATE UNIQUE INDEX one_primary_calendar_per_world
    ON world.calendars (world_id) WHERE is_primary;
CREATE TRIGGER touch BEFORE UPDATE ON world.calendars
    FOR EACH ROW EXECUTE FUNCTION core.touch_updated_at();
