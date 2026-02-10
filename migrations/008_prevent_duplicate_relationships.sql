-- Migration: Prevent duplicate relationships
--
-- Removes existing duplicate relationships (keeping the oldest) and
-- adds a unique constraint to prevent future duplicates.

-- Step 1: Remove duplicates, keeping the row with the lowest id.
DELETE FROM relationships
WHERE id NOT IN (
    SELECT MIN(id)
    FROM relationships
    GROUP BY campaign_id, source_entity_id, target_entity_id, relationship_type
);

-- Step 2: Add unique constraint.
ALTER TABLE relationships
ADD CONSTRAINT uq_relationships_source_target_type
UNIQUE (campaign_id, source_entity_id, target_entity_id, relationship_type);

COMMENT ON CONSTRAINT uq_relationships_source_target_type ON relationships IS
    'Prevents duplicate relationships between the same entities with the same type within a campaign';
