/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

-- 005_relations.sql — relations are entities: subclass table (spec §4.5, §4.6).
CREATE TABLE world.entity_relations (
    id             UUID PRIMARY KEY REFERENCES world.entities(id) ON DELETE CASCADE,
    source_id      UUID NOT NULL REFERENCES world.entities(id) ON DELETE CASCADE,
    target_id      UUID NOT NULL REFERENCES world.entities(id) ON DELETE CASCADE,
    predicate_id   UUID NOT NULL REFERENCES world.types(id),
    validity       NUMMULTIRANGE NOT NULL
                   DEFAULT nummultirange(numrange(NULL, NULL)),
    tone_id        UUID REFERENCES world.types(id),
    strength       INT CHECK (strength BETWEEN 1 AND 10),
    established_by UUID REFERENCES world.entities(id),
    terminated_by  UUID REFERENCES world.entities(id),
    CONSTRAINT no_self_relation CHECK (source_id <> target_id),
    -- An empty multirange overlaps nothing, so it would slip past the
    -- non-overlap trigger and allow unlimited duplicate rows per edge.
    -- numrange(5,5) collapses to empty, which a caller with lo = hi
    -- produces silently, so the degenerate case is reachable by accident.
    CONSTRAINT validity_not_empty CHECK (validity <> '{}'::nummultirange)
);
COMMENT ON TABLE world.entity_relations IS 'Subclass of world.entities — the id IS an entity id (never INHERITS: the v3 correction). Stored direction only; inverses are presentation.';
COMMENT ON COLUMN world.entity_relations.validity IS 'Set of hour intervals over the world line. Recurrence allowed across disjoint intervals; overlap of the same (source,target,predicate) forbidden.';
COMMENT ON COLUMN world.entity_relations.established_by IS 'Transition (spec §4.6): the occurrence explaining the interval opening. NULL = asserted, not explained.';

CREATE INDEX rel_edge ON world.entity_relations (source_id, target_id, predicate_id);
CREATE INDEX ON world.entity_relations (target_id);
CREATE INDEX ON world.entity_relations (predicate_id);

CREATE FUNCTION world.relations_no_overlap() RETURNS trigger AS $$
BEGIN
    -- All writes to a given (source, target, predicate) edge serialize
    -- through this advisory lock, so the EXISTS probe below always sees
    -- committed truth for that edge. Without it, two concurrent inserts
    -- of overlapping intervals on the same edge each take a snapshot in
    -- which the other row is invisible under READ COMMITTED, both find
    -- no conflict, and both commit. Taking the lock first makes the
    -- second transaction block until the first commits; the probe then
    -- runs under a fresh snapshot and sees the conflicting row. An
    -- advisory xact lock is used rather than row locks because the row
    -- being defended against may not exist yet. Rows written earlier in
    -- the SAME transaction need no lock: they are already visible.
    PERFORM pg_advisory_xact_lock(hashtextextended('world.entity_relations:'
        || NEW.source_id::text || NEW.target_id::text || NEW.predicate_id::text, 0));
    IF EXISTS (
        SELECT 1 FROM world.entity_relations r
        WHERE r.source_id = NEW.source_id
          AND r.target_id = NEW.target_id
          AND r.predicate_id = NEW.predicate_id
          AND r.id <> NEW.id
          AND r.validity && NEW.validity
    ) THEN
        RAISE EXCEPTION 'overlapping validity for identical (source, target, predicate)';
    END IF;
    RETURN NEW;
END $$ LANGUAGE plpgsql;
COMMENT ON FUNCTION world.relations_no_overlap() IS 'Recurrence-safe non-overlap: a constraint trigger states the rule exactly, where multirange GiST operator classes are not dependable across PG versions.';

CREATE CONSTRAINT TRIGGER no_overlap
    AFTER INSERT OR UPDATE OF source_id, target_id, predicate_id, validity
    ON world.entity_relations
    FOR EACH ROW EXECUTE FUNCTION world.relations_no_overlap();
