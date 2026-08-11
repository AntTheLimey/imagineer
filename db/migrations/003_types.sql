/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

-- 003_types.sql — per-world type tree (spec §4.3).
CREATE TABLE world.types (
    id             UUID PRIMARY KEY DEFAULT uuidv7(),
    world_id       UUID NOT NULL REFERENCES world.worlds(id) ON DELETE CASCADE,
    parent_id      UUID REFERENCES world.types(id) ON DELETE CASCADE,
    kind           TEXT NOT NULL CHECK (kind IN ('entity','relation','vocabulary')),
    name           TEXT NOT NULL,
    display_label  TEXT NOT NULL,
    description    TEXT,
    abstract       BOOLEAN NOT NULL DEFAULT false,
    relation_props JSONB NOT NULL DEFAULT '{}',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE world.types IS 'One tree for entity kinds, predicates (kind=relation), and vocabularies like tone (kind=vocabulary). parent_id nullable: roots are insertable (v3 correction).';
COMMENT ON COLUMN world.types.relation_props IS 'kind=relation only: inverse_name, inverse_display_label, is_symmetric, source/target type constraints.';

CREATE UNIQUE INDEX types_sibling_name
    ON world.types (world_id, parent_id, lower(name)) NULLS NOT DISTINCT;
CREATE INDEX ON world.types (world_id, kind);
CREATE TRIGGER touch BEFORE UPDATE ON world.types
    FOR EACH ROW EXECUTE FUNCTION core.touch_updated_at();

CREATE FUNCTION world.types_no_cycle() RETURNS trigger AS $$
DECLARE
    cur UUID := NEW.parent_id;
BEGIN
    WHILE cur IS NOT NULL LOOP
        IF cur = NEW.id THEN
            RAISE EXCEPTION 'type tree cycle involving %', NEW.id;
        END IF;
        SELECT parent_id INTO cur FROM world.types WHERE id = cur;
    END LOOP;
    RETURN NEW;
END $$ LANGUAGE plpgsql;

CREATE TRIGGER no_cycle BEFORE INSERT OR UPDATE OF parent_id ON world.types
    FOR EACH ROW EXECUTE FUNCTION world.types_no_cycle();
