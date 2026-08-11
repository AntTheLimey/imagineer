/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

-- 007_property_graph.sql — SQL/PGQ over the core (spec §10).
-- SOURCE/DESTINATION KEY references name the graph ELEMENT, not the table:
-- the element's default name is the unqualified relation name, so the target
-- is `entities`, never `world.entities` (PG19 rejects a qualified name here).
CREATE PROPERTY GRAPH world.canon_graph
    VERTEX TABLES (
        world.entities
            KEY (id)
            LABEL entity
            PROPERTIES (id, world_id, name, canon_status)
    )
    EDGE TABLES (
        world.entity_relations
            KEY (id)
            SOURCE KEY (source_id) REFERENCES entities (id)
            DESTINATION KEY (target_id) REFERENCES entities (id)
            LABEL relates
            PROPERTIES (id, predicate_id)
    );
COMMENT ON PROPERTY GRAPH world.canon_graph IS 'PGQ view of the world plane. PG19: fixed-depth patterns only; deep traversal stays recursive CTE.';
