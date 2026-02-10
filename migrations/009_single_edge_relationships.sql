/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

-- Migration 009: Single-Edge Relationship Model
--
-- Refactors the relationship model from dual-edge to single-edge LPG
-- (Labelled Property Graph). Instead of storing both forward and inverse
-- directions as separate rows, each relationship is stored once with the
-- canonical forward direction. The inverse is derived from the
-- relationship type's inverse_name at query time via a view.

BEGIN;

-- ============================================
-- Step 1: Create relationship_type_templates
-- System-default types copied to each new campaign
-- ============================================
CREATE TABLE relationship_type_templates (
    id                    BIGSERIAL PRIMARY KEY,
    name                  TEXT NOT NULL UNIQUE,
    inverse_name          TEXT NOT NULL,
    is_symmetric          BOOLEAN NOT NULL DEFAULT false,
    display_label         TEXT NOT NULL,
    inverse_display_label TEXT NOT NULL,
    description           TEXT,
    created_at            TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT tpl_symmetric_inverse_match
        CHECK (NOT is_symmetric OR name = inverse_name)
);

COMMENT ON TABLE relationship_type_templates IS
    'System-default relationship types copied to each new campaign';

-- ============================================
-- Step 2: Populate templates from existing
-- NULL-campaign forward types
-- ============================================
INSERT INTO relationship_type_templates
    (name, inverse_name, is_symmetric, display_label,
     inverse_display_label, description)
SELECT name, inverse_name, is_symmetric, display_label,
       inverse_display_label, description
FROM relationship_types
WHERE campaign_id IS NULL
AND name IN (
    'owns', 'employs', 'works_for', 'reports_to', 'parent_of',
    'located_at', 'member_of', 'created', 'rules',
    'headquartered_at',
    'knows', 'friend_of', 'enemy_of', 'allied_with'
);

-- ============================================
-- Step 3: Seed campaign-scoped types for each
-- existing campaign
-- ============================================
INSERT INTO relationship_types
    (campaign_id, name, inverse_name, is_symmetric, display_label,
     inverse_display_label, description)
SELECT c.id, t.name, t.inverse_name, t.is_symmetric, t.display_label,
       t.inverse_display_label, t.description
FROM campaigns c
CROSS JOIN relationship_type_templates t
ON CONFLICT (campaign_id, name) DO NOTHING;

-- ============================================
-- Step 4: Add relationship_type_id column
-- (nullable initially)
-- ============================================
ALTER TABLE relationships ADD COLUMN relationship_type_id BIGINT;

-- ============================================
-- Step 5: Populate relationship_type_id
-- ============================================

-- Forward types: direct name match to campaign-scoped types
UPDATE relationships r
SET relationship_type_id = rt.id
FROM relationship_types rt
WHERE rt.campaign_id = r.campaign_id
AND rt.name = r.relationship_type;

-- Inverse types: match against inverse_name and swap source/target
UPDATE relationships r
SET relationship_type_id = rt.id,
    source_entity_id = r.target_entity_id,
    target_entity_id = r.source_entity_id
FROM relationship_types rt
WHERE rt.campaign_id = r.campaign_id
AND rt.inverse_name = r.relationship_type
AND rt.name != rt.inverse_name
AND r.relationship_type_id IS NULL;

-- ============================================
-- Step 6: Handle any remaining unmapped
-- relationships by creating campaign-scoped
-- types for custom type names
-- ============================================
INSERT INTO relationship_types
    (campaign_id, name, inverse_name, is_symmetric, display_label,
     inverse_display_label, description)
SELECT DISTINCT r.campaign_id, r.relationship_type,
       r.relationship_type, true,
       INITCAP(REPLACE(r.relationship_type, '_', ' ')),
       INITCAP(REPLACE(r.relationship_type, '_', ' ')),
       'Auto-created during migration'
FROM relationships r
WHERE r.relationship_type_id IS NULL
ON CONFLICT (campaign_id, name) DO NOTHING;

-- Now map any remaining unmapped relationships
UPDATE relationships r
SET relationship_type_id = rt.id
FROM relationship_types rt
WHERE rt.campaign_id = r.campaign_id
AND rt.name = r.relationship_type
AND r.relationship_type_id IS NULL;

-- ============================================
-- Step 7: Deduplicate rows
-- After entity swapping, some rows may now be
-- duplicates
-- ============================================
DELETE FROM relationships
WHERE id NOT IN (
    SELECT MIN(id)
    FROM relationships
    GROUP BY campaign_id, source_entity_id,
             target_entity_id, relationship_type_id
);

-- ============================================
-- Step 8: Make relationship_type_id NOT NULL
-- and add foreign key
-- ============================================
ALTER TABLE relationships
    ALTER COLUMN relationship_type_id SET NOT NULL;

ALTER TABLE relationships
    ADD CONSTRAINT fk_relationships_type
    FOREIGN KEY (relationship_type_id)
    REFERENCES relationship_types(id);

-- ============================================
-- Step 9: Drop old columns and constraint,
-- add new unique constraint
-- ============================================
ALTER TABLE relationships
    DROP CONSTRAINT IF EXISTS uq_relationships_source_target_type;

ALTER TABLE relationships
    DROP COLUMN relationship_type,
    DROP COLUMN bidirectional;

ALTER TABLE relationships
    ADD CONSTRAINT uq_relationships_source_target_type
    UNIQUE (campaign_id, source_entity_id, target_entity_id,
            relationship_type_id);

-- ============================================
-- Step 10: Remove NULL-campaign relationship
-- types and enforce NOT NULL on campaign_id
-- ============================================
DELETE FROM relationship_types WHERE campaign_id IS NULL;

ALTER TABLE relationship_types
    ALTER COLUMN campaign_id SET NOT NULL;

-- ============================================
-- Step 11: Remove inverse-only type rows from
-- campaign-scoped types
-- ============================================
DELETE FROM relationship_types rt
WHERE NOT rt.is_symmetric
AND EXISTS (
    SELECT 1 FROM relationship_types rt2
    WHERE rt2.campaign_id = rt.campaign_id
    AND rt2.inverse_name = rt.name
    AND rt2.name != rt.name
    AND rt2.id < rt.id
);

-- ============================================
-- Step 12: Create trigger to prevent inverse
-- relationship duplicates
-- ============================================
CREATE OR REPLACE FUNCTION prevent_inverse_relationship()
RETURNS TRIGGER AS $$
BEGIN
    -- For symmetric types: block same type with swapped entities
    IF EXISTS (
        SELECT 1 FROM relationships r
        JOIN relationship_types rt ON rt.id = NEW.relationship_type_id
        WHERE rt.is_symmetric = true
        AND r.campaign_id = NEW.campaign_id
        AND r.source_entity_id = NEW.target_entity_id
        AND r.target_entity_id = NEW.source_entity_id
        AND r.relationship_type_id = NEW.relationship_type_id
        AND (TG_OP = 'INSERT' OR r.id != NEW.id)
    ) THEN
        RAISE EXCEPTION 'Symmetric inverse relationship already exists'
            USING ERRCODE = 'unique_violation';
    END IF;

    -- For asymmetric types: block if inverse type exists with
    -- swapped entities
    IF EXISTS (
        SELECT 1 FROM relationships r
        JOIN relationship_types rt_new
            ON rt_new.id = NEW.relationship_type_id
        JOIN relationship_types rt_inv
            ON rt_inv.campaign_id = rt_new.campaign_id
            AND rt_inv.name = rt_new.inverse_name
        WHERE rt_new.is_symmetric = false
        AND r.campaign_id = NEW.campaign_id
        AND r.source_entity_id = NEW.target_entity_id
        AND r.target_entity_id = NEW.source_entity_id
        AND r.relationship_type_id = rt_inv.id
        AND (TG_OP = 'INSERT' OR r.id != NEW.id)
    ) THEN
        RAISE EXCEPTION 'Inverse relationship already exists'
            USING ERRCODE = 'unique_violation';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION prevent_inverse_relationship() IS
    'Prevents inserting a relationship that is the logical inverse '
    'of an existing one, for both symmetric and asymmetric types';

CREATE TRIGGER trg_prevent_inverse_relationship
    BEFORE INSERT OR UPDATE ON relationships
    FOR EACH ROW
    EXECUTE FUNCTION prevent_inverse_relationship();

COMMENT ON TRIGGER trg_prevent_inverse_relationship
    ON relationships IS
    'Fires before insert or update to prevent duplicate inverse '
    'relationships';

-- ============================================
-- Step 13: Create display view
-- Provides both forward and inverse directions
-- for querying entity relationships
-- ============================================
CREATE VIEW entity_relationships_view AS
-- Forward: entity is source, use display_label
SELECT
    r.id,
    r.campaign_id,
    r.source_entity_id AS from_entity_id,
    r.target_entity_id AS to_entity_id,
    r.relationship_type_id,
    rt.name AS relationship_type,
    rt.display_label,
    r.tone,
    r.description,
    r.strength,
    r.created_at,
    r.updated_at,
    se.name AS from_entity_name,
    se.entity_type AS from_entity_type,
    te.name AS to_entity_name,
    te.entity_type AS to_entity_type,
    'forward' AS direction
FROM relationships r
JOIN relationship_types rt ON rt.id = r.relationship_type_id
JOIN entities se ON se.id = r.source_entity_id
JOIN entities te ON te.id = r.target_entity_id

UNION ALL

-- Inverse: entity is target, flip source/target, use inverse labels
SELECT
    r.id,
    r.campaign_id,
    r.target_entity_id AS from_entity_id,
    r.source_entity_id AS to_entity_id,
    r.relationship_type_id,
    rt.inverse_name AS relationship_type,
    rt.inverse_display_label AS display_label,
    r.tone,
    r.description,
    r.strength,
    r.created_at,
    r.updated_at,
    te.name AS from_entity_name,
    te.entity_type AS from_entity_type,
    se.name AS to_entity_name,
    se.entity_type AS to_entity_type,
    'inverse' AS direction
FROM relationships r
JOIN relationship_types rt ON rt.id = r.relationship_type_id
JOIN entities se ON se.id = r.source_entity_id
JOIN entities te ON te.id = r.target_entity_id
WHERE rt.is_symmetric = false;

COMMENT ON VIEW entity_relationships_view IS
    'Provides both forward and inverse views of all relationships '
    'for display purposes, joining entity names and type labels';

-- ============================================
-- Step 14: Record migration
-- ============================================
INSERT INTO schema_migrations (version)
    VALUES ('009_single_edge_relationships');

COMMIT;
