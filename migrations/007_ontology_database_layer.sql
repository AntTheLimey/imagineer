/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

-- ============================================
-- Migration 007: Ontology Database Layer
-- Adds indexes, FK corrections, validation
-- triggers, advisory functions, and views
-- for the ontology tables from migration 006.
-- ============================================

-- ============================================
-- Section 1: Indexes
-- B-tree indexes on FK columns and commonly
-- queried columns in ontology tables
-- ============================================

-- campaign_entity_types
CREATE INDEX idx_campaign_entity_types_campaign
    ON campaign_entity_types(campaign_id);

CREATE INDEX idx_campaign_entity_types_parent
    ON campaign_entity_types(campaign_id, parent_name)
    WHERE parent_name IS NOT NULL;

-- eras
CREATE INDEX idx_eras_campaign_id
    ON eras(campaign_id);

-- relationship_archive
CREATE INDEX idx_relationship_archive_campaign
    ON relationship_archive(campaign_id);

CREATE INDEX idx_relationship_archive_era
    ON relationship_archive(era_id)
    WHERE era_id IS NOT NULL;

CREATE INDEX idx_relationship_archive_source
    ON relationship_archive(source_entity_id);

CREATE INDEX idx_relationship_archive_target
    ON relationship_archive(target_entity_id);

-- cardinality_constraints
CREATE INDEX idx_cardinality_constraints_rel_type
    ON cardinality_constraints(relationship_type_id);

-- required_relationships
CREATE INDEX idx_required_relationships_entity_type
    ON required_relationships(campaign_id, entity_type);

-- relationships.era_id and entities.era_id (added by 006)
CREATE INDEX idx_relationships_era
    ON relationships(era_id)
    WHERE era_id IS NOT NULL;

CREATE INDEX idx_entities_era
    ON entities(era_id)
    WHERE era_id IS NOT NULL;

-- ============================================
-- Section 2: Foreign Key Corrections
-- ============================================

-- 2a: Self-referential FK for parent_name
-- Allows entity type inheritance within a
-- campaign (e.g., "deity" inherits "npc").
ALTER TABLE campaign_entity_types
    ADD CONSTRAINT fk_entity_type_parent
    FOREIGN KEY (campaign_id, parent_name)
    REFERENCES campaign_entity_types(campaign_id, name)
    ON DELETE RESTRICT
    DEFERRABLE INITIALLY DEFERRED;

COMMENT ON CONSTRAINT fk_entity_type_parent
    ON campaign_entity_types IS
    'Self-referential FK enforcing that parent_name '
    'references an existing type in the same campaign. '
    'RESTRICT prevents deleting a parent type that has '
    'children. DEFERRABLE so bulk seeding can insert '
    'children before parents within a transaction.';

-- 2b: FK for required_relationships.entity_type
ALTER TABLE required_relationships
    ADD CONSTRAINT fk_required_rel_entity_type
    FOREIGN KEY (campaign_id, entity_type)
    REFERENCES campaign_entity_types(campaign_id, name)
    ON DELETE CASCADE
    DEFERRABLE INITIALLY DEFERRED;

COMMENT ON CONSTRAINT fk_required_rel_entity_type
    ON required_relationships IS
    'Ensures required_relationships.entity_type '
    'references a valid campaign entity type. '
    'CASCADE deletes rules when the type is removed. '
    'DEFERRABLE for bulk seeding within a transaction.';

-- 2c: Change era FKs from SET NULL to RESTRICT
-- Eras are structural timeline markers; deleting
-- one with references would orphan temporal data.

-- relationships.era_id
ALTER TABLE relationships
    DROP CONSTRAINT IF EXISTS relationships_era_id_fkey;

ALTER TABLE relationships
    ADD CONSTRAINT relationships_era_id_fkey
    FOREIGN KEY (era_id) REFERENCES eras(id)
    ON DELETE RESTRICT;

COMMENT ON CONSTRAINT relationships_era_id_fkey
    ON relationships IS
    'RESTRICT prevents deleting an era that still has '
    'relationships attached. Reassign or archive '
    'relationships before removing an era.';

-- relationship_archive.era_id
ALTER TABLE relationship_archive
    DROP CONSTRAINT IF EXISTS relationship_archive_era_id_fkey;

ALTER TABLE relationship_archive
    ADD CONSTRAINT relationship_archive_era_id_fkey
    FOREIGN KEY (era_id) REFERENCES eras(id)
    ON DELETE RESTRICT;

COMMENT ON CONSTRAINT relationship_archive_era_id_fkey
    ON relationship_archive IS
    'RESTRICT prevents deleting an era that has '
    'archived relationships referencing it.';

-- entities.era_id
ALTER TABLE entities
    DROP CONSTRAINT IF EXISTS entities_era_id_fkey;

ALTER TABLE entities
    ADD CONSTRAINT entities_era_id_fkey
    FOREIGN KEY (era_id) REFERENCES eras(id)
    ON DELETE RESTRICT;

COMMENT ON CONSTRAINT entities_era_id_fkey
    ON entities IS
    'RESTRICT prevents deleting an era that still has '
    'entities attached. Reassign entities before '
    'removing an era.';

-- 2d: Explicit RESTRICT on
-- relationships.relationship_type_id
-- The unnamed FK defaults to NO ACTION, which
-- behaves like RESTRICT but is implicit. Make
-- intent explicit for documentation clarity.
ALTER TABLE relationships
    DROP CONSTRAINT IF EXISTS
        relationships_relationship_type_id_fkey;

ALTER TABLE relationships
    ADD CONSTRAINT relationships_relationship_type_id_fkey
    FOREIGN KEY (relationship_type_id)
    REFERENCES relationship_types(id)
    ON DELETE RESTRICT;

COMMENT ON CONSTRAINT relationships_relationship_type_id_fkey
    ON relationships IS
    'RESTRICT is used (not CASCADE) because deleting '
    'a relationship type that is still in use would '
    'silently destroy graph data. The application must '
    'reassign or remove relationships before deleting '
    'a type.';

-- ============================================
-- Section 3: Triggers
-- ============================================

-- 3a: Entity type validation trigger
-- Validates that entities reference a concrete
-- (non-abstract) campaign entity type. Skips
-- validation for legacy campaigns that have no
-- entity types seeded.
CREATE OR REPLACE FUNCTION validate_entity_type()
RETURNS TRIGGER AS $$
BEGIN
    -- Guard: skip validation for legacy campaigns
    -- that have no entity types seeded yet.
    IF NOT EXISTS (
        SELECT 1 FROM campaign_entity_types
        WHERE campaign_id = NEW.campaign_id
        LIMIT 1
    ) THEN
        RETURN NEW;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM campaign_entity_types
        WHERE campaign_id = NEW.campaign_id
          AND name = NEW.entity_type
          AND abstract = false
    ) THEN
        RAISE EXCEPTION
            'entity_type "%" is not a valid concrete '
            'type for campaign %',
            NEW.entity_type, NEW.campaign_id
            USING ERRCODE = 'check_violation';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION validate_entity_type() IS
    'Validates that an entity''s entity_type references '
    'a concrete (non-abstract) type in '
    'campaign_entity_types. Skips validation for legacy '
    'campaigns with no seeded entity types.';

CREATE TRIGGER check_entity_type
    BEFORE INSERT OR UPDATE OF entity_type ON entities
    FOR EACH ROW
    EXECUTE FUNCTION validate_entity_type();

COMMENT ON TRIGGER check_entity_type ON entities IS
    'Fires on INSERT or UPDATE of entity_type to '
    'ensure the type is a valid concrete campaign '
    'entity type.';

-- 3b: Relationship type pair validation trigger
-- ADVISORY enforcement: warns but never blocks.
-- The GM must always be able to create any
-- relationship regardless of ontology constraints.
CREATE OR REPLACE FUNCTION validate_relationship_type_pair()
RETURNS TRIGGER AS $$
DECLARE
    v_source_type TEXT;
    v_target_type TEXT;
BEGIN
    -- Skip if no constraints exist for this
    -- relationship type (open-world default)
    IF NOT EXISTS (
        SELECT 1 FROM relationship_type_constraints
        WHERE relationship_type_id =
              NEW.relationship_type_id
    ) THEN
        RETURN NEW;
    END IF;

    -- Look up entity types
    SELECT entity_type INTO v_source_type
    FROM entities WHERE id = NEW.source_entity_id;

    SELECT entity_type INTO v_target_type
    FROM entities WHERE id = NEW.target_entity_id;

    -- Validate the source/target type pair
    IF NOT EXISTS (
        SELECT 1 FROM relationship_type_constraints
        WHERE relationship_type_id =
              NEW.relationship_type_id
          AND source_entity_type = v_source_type
          AND target_entity_type = v_target_type
    ) THEN
        RAISE WARNING
            'Relationship type % does not allow % -> %',
            NEW.relationship_type_id,
            v_source_type,
            v_target_type;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION validate_relationship_type_pair() IS
    'Advisory validation of relationship source/target '
    'entity type pairs against '
    'relationship_type_constraints. Issues a WARNING '
    'but never blocks — the GM can always override.';

CREATE TRIGGER check_relationship_type_pair
    BEFORE INSERT OR UPDATE ON relationships
    FOR EACH ROW
    EXECUTE FUNCTION validate_relationship_type_pair();

COMMENT ON TRIGGER check_relationship_type_pair
    ON relationships IS
    'Advisory trigger that warns when a relationship '
    'violates domain/range constraints. Never blocks '
    'the operation.';

-- ============================================
-- Section 4: Database Functions
-- Advisory reporting functions for ontology
-- constraint violations
-- ============================================

-- 4a: check_required_relationships(campaign_id)
-- Returns entities that are missing a required
-- relationship type defined in
-- required_relationships.
CREATE OR REPLACE FUNCTION check_required_relationships(
    p_campaign_id BIGINT
)
RETURNS TABLE (
    entity_id                 BIGINT,
    entity_name               TEXT,
    entity_type               TEXT,
    missing_relationship_type TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        e.id AS entity_id,
        e.name AS entity_name,
        e.entity_type::TEXT AS entity_type,
        rr.relationship_type_name
            AS missing_relationship_type
    FROM required_relationships rr
    JOIN entities e
        ON e.campaign_id = rr.campaign_id
       AND e.entity_type = rr.entity_type
    JOIN relationship_types rt
        ON rt.name = rr.relationship_type_name
       AND rt.campaign_id = rr.campaign_id
    WHERE rr.campaign_id = p_campaign_id
      AND NOT EXISTS (
          SELECT 1 FROM relationships r
          WHERE r.campaign_id = p_campaign_id
            AND r.relationship_type_id = rt.id
            AND (r.source_entity_id = e.id
                 OR r.target_entity_id = e.id)
      );
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION check_required_relationships(BIGINT)
    IS 'Returns entities missing a required '
       'relationship type. Used by the Graph Expert '
       'for advisory ontology hygiene checks.';

-- 4b: check_cardinality_violations(campaign_id)
-- Returns entities that exceed the maximum
-- cardinality defined in cardinality_constraints.
CREATE OR REPLACE FUNCTION check_cardinality_violations(
    p_campaign_id BIGINT
)
RETURNS TABLE (
    entity_id          BIGINT,
    entity_name        TEXT,
    entity_type        TEXT,
    relationship_type  TEXT,
    direction          TEXT,
    current_count      INT,
    max_allowed        INT
) AS $$
BEGIN
    RETURN QUERY
    -- Source direction violations
    SELECT
        e.id,
        e.name,
        e.entity_type::TEXT,
        rt.name,
        'source'::TEXT,
        COUNT(*)::INT,
        cc.max_source
    FROM relationships r
    JOIN relationship_types rt
        ON r.relationship_type_id = rt.id
    JOIN cardinality_constraints cc
        ON cc.campaign_id = r.campaign_id
       AND cc.relationship_type_id = rt.id
    JOIN entities e
        ON e.id = r.source_entity_id
    WHERE r.campaign_id = p_campaign_id
      AND cc.max_source IS NOT NULL
    GROUP BY e.id, e.name, e.entity_type,
             rt.name, cc.max_source
    HAVING COUNT(*) > cc.max_source

    UNION ALL

    -- Target direction violations
    SELECT
        e.id,
        e.name,
        e.entity_type::TEXT,
        rt.name,
        'target'::TEXT,
        COUNT(*)::INT,
        cc.max_target
    FROM relationships r
    JOIN relationship_types rt
        ON r.relationship_type_id = rt.id
    JOIN cardinality_constraints cc
        ON cc.campaign_id = r.campaign_id
       AND cc.relationship_type_id = rt.id
    JOIN entities e
        ON e.id = r.target_entity_id
    WHERE r.campaign_id = p_campaign_id
      AND cc.max_target IS NOT NULL
    GROUP BY e.id, e.name, e.entity_type,
             rt.name, cc.max_target
    HAVING COUNT(*) > cc.max_target;
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION check_cardinality_violations(BIGINT)
    IS 'Returns entities that exceed max cardinality '
       'limits. Used by the Graph Expert for advisory '
       'ontology hygiene checks.';

-- ============================================
-- Section 5: Views
-- ============================================

-- 5a: cardinality_constraints_with_names
-- Joins cardinality_constraints with
-- relationship_types for human-readable output.
CREATE VIEW cardinality_constraints_with_names AS
SELECT
    cc.id,
    cc.campaign_id,
    cc.relationship_type_id,
    rt.name AS relationship_type_name,
    cc.max_source,
    cc.max_target,
    cc.created_at
FROM cardinality_constraints cc
JOIN relationship_types rt
    ON cc.relationship_type_id = rt.id;

COMMENT ON VIEW cardinality_constraints_with_names IS
    'Cardinality constraints joined with relationship '
    'type names for human-readable display';

-- 5b: orphaned_entities
-- Entities with no relationships at all.
CREATE VIEW orphaned_entities AS
SELECT
    e.id,
    e.campaign_id,
    e.entity_type,
    e.name
FROM entities e
WHERE NOT EXISTS (
    SELECT 1 FROM relationships r
    WHERE r.source_entity_id = e.id
       OR r.target_entity_id = e.id
);

COMMENT ON VIEW orphaned_entities IS
    'Entities with no relationships. Used by the '
    'Graph Expert to identify disconnected nodes '
    'in the campaign knowledge graph.';

-- ============================================
-- Section 6: Record Migration
-- ============================================
INSERT INTO schema_migrations (version)
VALUES ('007_ontology_database_layer');
