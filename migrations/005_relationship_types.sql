-- Relationship Types Migration
-- Adds support for relationship type definitions with inverse mappings
-- Supports per-campaign customization and system-wide defaults

-- ============================================
-- Relationship Types Table
-- Defines relationship types with inverse mappings
-- ============================================
CREATE TABLE relationship_types (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    campaign_id             UUID REFERENCES campaigns(id) ON DELETE CASCADE,
    name                    TEXT NOT NULL,
    inverse_name            TEXT NOT NULL,
    is_symmetric            BOOLEAN NOT NULL DEFAULT false,
    display_label           TEXT NOT NULL,
    inverse_display_label   TEXT NOT NULL,
    description             TEXT,
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW(),

    -- Ensure unique names per campaign (NULL campaign_id = system-wide)
    CONSTRAINT unique_relationship_type_per_campaign 
        UNIQUE NULLS NOT DISTINCT (campaign_id, name),
    
    -- Symmetric relationships must have matching name and inverse_name
    CONSTRAINT symmetric_inverse_match 
        CHECK (NOT is_symmetric OR name = inverse_name)
);

COMMENT ON TABLE relationship_types IS 'Relationship type definitions with inverse mappings, supporting per-campaign customization';
COMMENT ON COLUMN relationship_types.campaign_id IS 'Campaign this type belongs to; NULL for system-wide default types';
COMMENT ON COLUMN relationship_types.name IS 'Relationship type name used in relationships table (e.g., "owns", "knows")';
COMMENT ON COLUMN relationship_types.inverse_name IS 'Inverse relationship type (e.g., "owned_by" for "owns", or same name if symmetric)';
COMMENT ON COLUMN relationship_types.is_symmetric IS 'TRUE if relationship reads the same in both directions (e.g., "knows")';
COMMENT ON COLUMN relationship_types.display_label IS 'Human-friendly label for UI display (e.g., "Owns")';
COMMENT ON COLUMN relationship_types.inverse_display_label IS 'Display label for inverse relationship (e.g., "Is owned by")';

-- ============================================
-- Indexes
-- ============================================

-- Foreign key index for campaign lookups
CREATE INDEX idx_relationship_types_campaign_id ON relationship_types(campaign_id);

-- Index for fast relationship type lookups when creating relationships
CREATE INDEX idx_relationship_types_name ON relationship_types(name);

-- Composite index for validation queries (checking type exists for campaign)
CREATE INDEX idx_relationship_types_campaign_name ON relationship_types(campaign_id, name);

-- Index for finding symmetric relationships
CREATE INDEX idx_relationship_types_symmetric ON relationship_types(is_symmetric) WHERE is_symmetric = true;

-- ============================================
-- Seed Data: System-Wide Default Types
-- ============================================

-- Ownership relationships (asymmetric)
INSERT INTO relationship_types (campaign_id, name, inverse_name, is_symmetric, display_label, inverse_display_label, description) VALUES
(NULL, 'owns', 'owned_by', false, 'Owns', 'Is owned by', 'Entity possesses or controls another entity'),
(NULL, 'owned_by', 'owns', false, 'Is owned by', 'Owns', 'Inverse of owns relationship');

-- Employment relationships (asymmetric)
INSERT INTO relationship_types (campaign_id, name, inverse_name, is_symmetric, display_label, inverse_display_label, description) VALUES
(NULL, 'employs', 'employed_by', false, 'Employs', 'Is employed by', 'Entity employs another as worker or servant'),
(NULL, 'employed_by', 'employs', false, 'Is employed by', 'Employs', 'Inverse of employs relationship'),
(NULL, 'works_for', 'employs', false, 'Works for', 'Employs', 'Alias for employed_by relationship');

-- Management relationships (asymmetric)
INSERT INTO relationship_types (campaign_id, name, inverse_name, is_symmetric, display_label, inverse_display_label, description) VALUES
(NULL, 'reports_to', 'manages', false, 'Reports to', 'Manages', 'Entity reports to another in organizational hierarchy'),
(NULL, 'manages', 'reports_to', false, 'Manages', 'Reports to', 'Entity manages or supervises another');

-- Family relationships (asymmetric)
INSERT INTO relationship_types (campaign_id, name, inverse_name, is_symmetric, display_label, inverse_display_label, description) VALUES
(NULL, 'parent_of', 'child_of', false, 'Parent of', 'Child of', 'Entity is parent of another'),
(NULL, 'child_of', 'parent_of', false, 'Child of', 'Parent of', 'Entity is child of another');

-- Location relationships (asymmetric)
INSERT INTO relationship_types (campaign_id, name, inverse_name, is_symmetric, display_label, inverse_display_label, description) VALUES
(NULL, 'located_at', 'contains', false, 'Located at', 'Contains', 'Entity is physically located at another location'),
(NULL, 'contains', 'located_at', false, 'Contains', 'Located at', 'Location contains another entity');

-- Membership relationships (asymmetric)
INSERT INTO relationship_types (campaign_id, name, inverse_name, is_symmetric, display_label, inverse_display_label, description) VALUES
(NULL, 'member_of', 'has_member', false, 'Member of', 'Has member', 'Entity is member of organization or faction'),
(NULL, 'has_member', 'member_of', false, 'Has member', 'Member of', 'Organization has entity as member');

-- Creation relationships (asymmetric)
INSERT INTO relationship_types (campaign_id, name, inverse_name, is_symmetric, display_label, inverse_display_label, description) VALUES
(NULL, 'created', 'created_by', false, 'Created', 'Created by', 'Entity created or made another entity'),
(NULL, 'created_by', 'created', false, 'Created by', 'Created', 'Entity was created by another');

-- Political relationships (asymmetric)
INSERT INTO relationship_types (campaign_id, name, inverse_name, is_symmetric, display_label, inverse_display_label, description) VALUES
(NULL, 'rules', 'ruled_by', false, 'Rules', 'Ruled by', 'Entity has political authority over another'),
(NULL, 'ruled_by', 'rules', false, 'Ruled by', 'Rules', 'Entity is under political authority of another');

-- Social relationships (symmetric)
INSERT INTO relationship_types (campaign_id, name, inverse_name, is_symmetric, display_label, inverse_display_label, description) VALUES
(NULL, 'knows', 'knows', true, 'Knows', 'Knows', 'Entity is acquainted with another'),
(NULL, 'friend_of', 'friend_of', true, 'Friend of', 'Friend of', 'Entity has friendly relationship with another'),
(NULL, 'enemy_of', 'enemy_of', true, 'Enemy of', 'Enemy of', 'Entity has hostile relationship with another'),
(NULL, 'allied_with', 'allied_with', true, 'Allied with', 'Allied with', 'Entity has alliance or partnership with another');

-- ============================================
-- Helper Function: Get Inverse Relationship Type
-- ============================================

-- Function to look up the inverse of a relationship type
-- Checks campaign-specific types first, then falls back to system defaults
CREATE OR REPLACE FUNCTION get_inverse_relationship_type(
    p_campaign_id UUID,
    p_relationship_type TEXT
) RETURNS TEXT AS $$
DECLARE
    v_inverse_name TEXT;
BEGIN
    -- First try campaign-specific type
    SELECT inverse_name INTO v_inverse_name
    FROM relationship_types
    WHERE campaign_id = p_campaign_id
      AND name = p_relationship_type;
    
    -- If not found, try system-wide default
    IF v_inverse_name IS NULL THEN
        SELECT inverse_name INTO v_inverse_name
        FROM relationship_types
        WHERE campaign_id IS NULL
          AND name = p_relationship_type;
    END IF;
    
    RETURN v_inverse_name;
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION get_inverse_relationship_type IS 'Returns inverse relationship type name, checking campaign-specific first then system defaults';

-- ============================================
-- Helper Function: Validate Relationship Type
-- ============================================

-- Function to check if a relationship type exists for a campaign
CREATE OR REPLACE FUNCTION validate_relationship_type(
    p_campaign_id UUID,
    p_relationship_type TEXT
) RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM relationship_types
        WHERE name = p_relationship_type
          AND (campaign_id = p_campaign_id OR campaign_id IS NULL)
    );
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION validate_relationship_type IS 'Returns TRUE if relationship type exists for campaign (including system defaults)';

-- ============================================
-- Helper View: All Available Relationship Types
-- ============================================

-- View that shows all available relationship types for each campaign
-- Combines system-wide defaults with campaign-specific types
CREATE OR REPLACE VIEW available_relationship_types AS
SELECT DISTINCT ON (c.id, rt.name)
    c.id as campaign_id,
    rt.id as relationship_type_id,
    rt.name,
    rt.inverse_name,
    rt.is_symmetric,
    rt.display_label,
    rt.inverse_display_label,
    rt.description,
    rt.campaign_id IS NOT NULL as is_custom
FROM campaigns c
CROSS JOIN relationship_types rt
WHERE rt.campaign_id IS NULL OR rt.campaign_id = c.id
ORDER BY c.id, rt.name, rt.campaign_id NULLS LAST;

COMMENT ON VIEW available_relationship_types IS 'All available relationship types per campaign (system defaults + campaign-specific)';

-- ============================================
-- Updated Timestamp Trigger
-- ============================================

CREATE OR REPLACE FUNCTION update_relationship_types_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_relationship_types_updated_at
    BEFORE UPDATE ON relationship_types
    FOR EACH ROW
    EXECUTE FUNCTION update_relationship_types_updated_at();
