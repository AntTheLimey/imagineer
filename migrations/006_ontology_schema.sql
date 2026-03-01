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
-- Campaign Entity Types Table
-- Campaign-scoped entity type hierarchy
-- ============================================
CREATE TABLE campaign_entity_types (
    id          BIGSERIAL PRIMARY KEY,
    campaign_id BIGINT NOT NULL REFERENCES campaigns(id)
                ON DELETE CASCADE,
    name        TEXT NOT NULL,
    parent_name TEXT,
    abstract    BOOLEAN NOT NULL DEFAULT false,
    description TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(campaign_id, name)
);

COMMENT ON TABLE campaign_entity_types IS
    'Campaign-scoped entity type hierarchy seeded '
    'from schemas/ontology/entity-types.yaml';

CREATE TRIGGER update_campaign_entity_types_updated_at
    BEFORE UPDATE ON campaign_entity_types
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Eras Table
-- Named periods in the fictional timeline
-- ============================================
CREATE TABLE eras (
    id          BIGSERIAL PRIMARY KEY,
    campaign_id BIGINT NOT NULL REFERENCES campaigns(id)
                ON DELETE CASCADE,
    sequence    INT NOT NULL,
    name        TEXT NOT NULL,
    scale       TEXT NOT NULL DEFAULT 'now'
                CHECK (scale IN (
                    'mythic', 'ancient', 'distant',
                    'past', 'recent', 'now'
                )),
    description TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(campaign_id, sequence),
    UNIQUE(campaign_id, name)
);

COMMENT ON TABLE eras IS
    'Named periods in a campaign fictional timeline';

CREATE TRIGGER update_eras_updated_at
    BEFORE UPDATE ON eras
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Relationship Archive Table
-- Archived relationships with era reference
-- ============================================
CREATE TABLE relationship_archive (
    id                   BIGSERIAL PRIMARY KEY,
    campaign_id          BIGINT NOT NULL
                         REFERENCES campaigns(id)
                         ON DELETE CASCADE,
    source_entity_id     BIGINT NOT NULL
                         REFERENCES entities(id)
                         ON DELETE CASCADE,
    target_entity_id     BIGINT NOT NULL
                         REFERENCES entities(id)
                         ON DELETE CASCADE,
    relationship_type_id BIGINT NOT NULL
                         REFERENCES relationship_types(id)
                         ON DELETE CASCADE,
    era_id               BIGINT REFERENCES eras(id)
                         ON DELETE SET NULL,
    tone                 TEXT CHECK (tone IN (
                             'friendly', 'hostile',
                             'neutral', 'romantic',
                             'professional', 'fearful',
                             'respectful', 'unknown'
                         )),
    description          TEXT,
    strength             INT CHECK (
                             strength >= 1
                             AND strength <= 10
                         ),
    archived_at          TIMESTAMPTZ DEFAULT NOW(),
    original_created_at  TIMESTAMPTZ
);

COMMENT ON TABLE relationship_archive IS
    'Archived relationships preserving historical '
    'graph state per era';

-- ============================================
-- Cardinality Constraints Table
-- ============================================
CREATE TABLE cardinality_constraints (
    id                   BIGSERIAL PRIMARY KEY,
    campaign_id          BIGINT NOT NULL
                         REFERENCES campaigns(id)
                         ON DELETE CASCADE,
    relationship_type_id BIGINT NOT NULL
                         REFERENCES relationship_types(id)
                         ON DELETE CASCADE,
    max_source           INT,
    max_target           INT,
    created_at           TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(campaign_id, relationship_type_id)
);

COMMENT ON TABLE cardinality_constraints IS
    'Campaign-scoped cardinality limits per '
    'relationship type';

-- ============================================
-- Required Relationships Table
-- ============================================
CREATE TABLE required_relationships (
    id                     BIGSERIAL PRIMARY KEY,
    campaign_id            BIGINT NOT NULL
                           REFERENCES campaigns(id)
                           ON DELETE CASCADE,
    entity_type            TEXT NOT NULL,
    relationship_type_name TEXT NOT NULL,
    created_at             TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(campaign_id, entity_type,
           relationship_type_name)
);

COMMENT ON TABLE required_relationships IS
    'Advisory rules for relationships every entity '
    'of a given type should have';

-- ============================================
-- Constraint Overrides Table
-- GM acknowledgements that evolve the campaign
-- ontology
-- ============================================
CREATE TABLE constraint_overrides (
    id              BIGSERIAL PRIMARY KEY,
    campaign_id     BIGINT NOT NULL
                    REFERENCES campaigns(id)
                    ON DELETE CASCADE,
    constraint_type TEXT NOT NULL
                    CHECK (constraint_type IN (
                        'domain_range', 'cardinality',
                        'required'
                    )),
    override_key    TEXT NOT NULL,
    acknowledged_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(campaign_id, constraint_type, override_key)
);

COMMENT ON TABLE constraint_overrides IS
    'GM acknowledgements that override specific '
    'constraint violations';

-- ============================================
-- Add era_id to existing tables
-- ============================================

-- Add era_id column to existing relationships table
ALTER TABLE relationships
    ADD COLUMN era_id BIGINT
    REFERENCES eras(id) ON DELETE SET NULL;

-- Add era_id column to entities table
ALTER TABLE entities
    ADD COLUMN era_id BIGINT
    REFERENCES eras(id) ON DELETE SET NULL;

-- Drop the hardcoded CHECK constraint on entity_type.
-- Validation is now against campaign_entity_types.
ALTER TABLE entities
    DROP CONSTRAINT IF EXISTS entities_entity_type_check;

-- ============================================
-- Record migration
-- ============================================
INSERT INTO schema_migrations (version)
VALUES ('006_ontology_schema');
