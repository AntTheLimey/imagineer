/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

-- Migration 014: Relationship Type Constraints
--
-- Adds a table defining which entity type pairs are valid for each
-- relationship type. Used by the Graph Expert agent to flag
-- invalid_type_pair warnings.

-- ============================================
-- Create relationship_type_constraints table
-- ============================================
CREATE TABLE IF NOT EXISTS relationship_type_constraints (
	id                   bigserial PRIMARY KEY,
	relationship_type_id bigint NOT NULL
		REFERENCES relationship_types(id) ON DELETE CASCADE,
	source_entity_type   text NOT NULL,
	target_entity_type   text NOT NULL,
	created_at           timestamptz DEFAULT NOW(),
	UNIQUE(relationship_type_id, source_entity_type, target_entity_type)
);

COMMENT ON TABLE relationship_type_constraints IS
	'Defines valid entity type pairs for each relationship type, used by Graph Expert for ontology validation.';

COMMENT ON COLUMN relationship_type_constraints.relationship_type_id IS
	'The relationship type this constraint applies to.';

COMMENT ON COLUMN relationship_type_constraints.source_entity_type IS
	'The valid source entity type (e.g., npc, location, item).';

COMMENT ON COLUMN relationship_type_constraints.target_entity_type IS
	'The valid target entity type.';

-- ============================================
-- Seed default constraints for system types
-- ============================================
-- "located_at" should be entity → location
INSERT INTO relationship_type_constraints (relationship_type_id, source_entity_type, target_entity_type)
SELECT rt.id, unnest(ARRAY['npc', 'item', 'creature', 'organization', 'faction', 'event']), 'location'
FROM relationship_types rt
WHERE rt.name = 'located_at' AND rt.campaign_id IS NULL
ON CONFLICT DO NOTHING;

-- "member_of" should be npc/creature → faction/organization
INSERT INTO relationship_type_constraints (relationship_type_id, source_entity_type, target_entity_type)
SELECT rt.id, unnest(ARRAY['npc', 'creature']), unnest(ARRAY['faction', 'organization'])
FROM relationship_types rt
WHERE rt.name = 'member_of' AND rt.campaign_id IS NULL
ON CONFLICT DO NOTHING;

-- "owns" should be npc/faction/organization → item
INSERT INTO relationship_type_constraints (relationship_type_id, source_entity_type, target_entity_type)
SELECT rt.id, unnest(ARRAY['npc', 'faction', 'organization']), 'item'
FROM relationship_types rt
WHERE rt.name = 'owns' AND rt.campaign_id IS NULL
ON CONFLICT DO NOTHING;

-- ============================================
-- Record migration
-- ============================================
INSERT INTO schema_migrations (version) VALUES ('014_relationship_type_constraints')
ON CONFLICT DO NOTHING;
