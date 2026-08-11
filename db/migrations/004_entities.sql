/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

-- 004_entities.sql — world entities (spec §4.4; decision 22: no campaign_id).
CREATE TABLE world.entities (
    id           UUID PRIMARY KEY DEFAULT uuidv7(),
    world_id     UUID NOT NULL REFERENCES world.worlds(id) ON DELETE CASCADE,
    type_id      UUID NOT NULL REFERENCES world.types(id),
    name         TEXT NOT NULL,
    description  TEXT,
    attributes   JSONB NOT NULL DEFAULT '{}',
    system_data  JSONB NOT NULL DEFAULT '{}',
    canon_status TEXT NOT NULL DEFAULT 'draft'
                 CHECK (canon_status IN ('draft','authoritative','superseded')),
    gm_notes     TEXT,
    deleted_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE world.entities IS 'Every entity is world canon from birth (decision 22). Relations are entities too, via world.entity_relations.';
COMMENT ON COLUMN world.entities.system_data IS 'Opaque to the ontology (decision 3). Sheets live here.';
COMMENT ON COLUMN world.entities.gm_notes IS 'Never crosses the player boundary — enforced at the query layer (§8c).';

CREATE INDEX ON world.entities (world_id);
CREATE INDEX ON world.entities (world_id, type_id);
CREATE INDEX entities_name_trgm ON world.entities USING GIN (name gin_trgm_ops);
CREATE INDEX ON world.entities USING GIN (attributes);
CREATE TRIGGER touch BEFORE UPDATE ON world.entities
    FOR EACH ROW EXECUTE FUNCTION core.touch_updated_at();
