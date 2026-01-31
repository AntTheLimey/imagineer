-- Imagineer Initial Schema Migration
-- Creates all core tables for the TTRPG Campaign Intelligence Platform

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- Game Systems Table
-- Defines TTRPG systems (CoC 7e, GURPS 4e, FitD, etc.)
-- ============================================
CREATE TABLE game_systems (
    id                       UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name                     TEXT NOT NULL,
    code                     TEXT UNIQUE NOT NULL,
    attribute_schema         JSONB,
    skill_schema             JSONB,
    character_sheet_template JSONB,
    dice_conventions         JSONB,
    created_at               TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE game_systems IS 'TTRPG system definitions (e.g., Call of Cthulhu 7e, GURPS 4e)';
COMMENT ON COLUMN game_systems.code IS 'Short unique identifier (e.g., coc-7e, gurps-4e, fitd)';
COMMENT ON COLUMN game_systems.attribute_schema IS 'JSON schema defining character attributes for this system';

-- ============================================
-- Campaigns Table
-- Individual campaign instances
-- ============================================
CREATE TABLE campaigns (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        TEXT NOT NULL,
    system_id   UUID REFERENCES game_systems(id) ON DELETE SET NULL,
    description TEXT,
    settings    JSONB DEFAULT '{}',
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE campaigns IS 'Individual TTRPG campaigns';
COMMENT ON COLUMN campaigns.settings IS 'Campaign-specific settings and configuration';

CREATE INDEX idx_campaigns_system_id ON campaigns(system_id);

-- ============================================
-- Sessions Table
-- Game sessions within a campaign
-- ============================================
CREATE TABLE sessions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    campaign_id     UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    session_number  INT,
    planned_date    DATE,
    actual_date     DATE,
    status          TEXT DEFAULT 'PLANNED' CHECK (status IN ('PLANNED', 'COMPLETED', 'SKIPPED')),
    prep_notes      TEXT,
    planned_scenes  JSONB,
    actual_notes    TEXT,
    discoveries     JSONB DEFAULT '[]',
    player_decisions JSONB DEFAULT '[]',
    consequences    JSONB DEFAULT '[]',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE sessions IS 'Individual game sessions within a campaign';
COMMENT ON COLUMN sessions.discoveries IS 'Array of {entity_id, how_discovered} objects';
COMMENT ON COLUMN sessions.player_decisions IS 'Notable player choices made during the session';

CREATE INDEX idx_sessions_campaign_id ON sessions(campaign_id);
CREATE INDEX idx_sessions_status ON sessions(status);
CREATE UNIQUE INDEX idx_sessions_campaign_number ON sessions(campaign_id, session_number);

-- ============================================
-- Entities Table (Polymorphic)
-- NPCs, locations, items, factions, clues, etc.
-- ============================================
CREATE TABLE entities (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    campaign_id         UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    entity_type         TEXT NOT NULL CHECK (entity_type IN (
                            'npc', 'location', 'item', 'faction', 'clue',
                            'creature', 'organization', 'event', 'document', 'other'
                        )),
    name                TEXT NOT NULL,
    description         TEXT,
    attributes          JSONB DEFAULT '{}',
    tags                TEXT[] DEFAULT '{}',
    keeper_notes        TEXT,
    discovered_session  UUID REFERENCES sessions(id) ON DELETE SET NULL,
    source_document     TEXT,
    source_confidence   TEXT DEFAULT 'DRAFT' CHECK (source_confidence IN ('DRAFT', 'AUTHORITATIVE', 'SUPERSEDED')),
    version             INT DEFAULT 1,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE entities IS 'Polymorphic table for all campaign entities (NPCs, locations, items, etc.)';
COMMENT ON COLUMN entities.entity_type IS 'Type discriminator for the entity';
COMMENT ON COLUMN entities.attributes IS 'System-specific stats and properties (JSONB)';
COMMENT ON COLUMN entities.keeper_notes IS 'GM-only notes, never shown to players';
COMMENT ON COLUMN entities.source_confidence IS 'Canon status: DRAFT, AUTHORITATIVE, or SUPERSEDED';

CREATE INDEX idx_entities_campaign_id ON entities(campaign_id);
CREATE INDEX idx_entities_type ON entities(entity_type);
CREATE INDEX idx_entities_campaign_type ON entities(campaign_id, entity_type);
CREATE INDEX idx_entities_tags ON entities USING GIN(tags);
CREATE INDEX idx_entities_attributes ON entities USING GIN(attributes);
CREATE INDEX idx_entities_name_trgm ON entities USING GIN(name gin_trgm_ops);

-- ============================================
-- Relationships Table
-- Connections between entities
-- ============================================
CREATE TABLE relationships (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    campaign_id         UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    source_entity_id    UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    target_entity_id    UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    relationship_type   TEXT NOT NULL,
    tone                TEXT CHECK (tone IN ('friendly', 'hostile', 'neutral', 'romantic', 'professional', 'fearful', 'respectful', 'unknown')),
    description         TEXT,
    bidirectional       BOOLEAN DEFAULT false,
    strength            INT CHECK (strength >= 1 AND strength <= 10),
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT no_self_relationship CHECK (source_entity_id != target_entity_id)
);

COMMENT ON TABLE relationships IS 'Connections between campaign entities';
COMMENT ON COLUMN relationships.relationship_type IS 'Type of relationship (knows, owns, located_at, opposes, etc.)';
COMMENT ON COLUMN relationships.strength IS 'Optional weight for graph algorithms (1-10)';

CREATE INDEX idx_relationships_campaign_id ON relationships(campaign_id);
CREATE INDEX idx_relationships_source ON relationships(source_entity_id);
CREATE INDEX idx_relationships_target ON relationships(target_entity_id);
CREATE INDEX idx_relationships_type ON relationships(relationship_type);

-- ============================================
-- Timeline Events Table
-- In-game chronological events
-- ============================================
CREATE TABLE timeline_events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    campaign_id     UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    event_date      DATE,
    event_time      TIME,
    date_precision  TEXT DEFAULT 'exact' CHECK (date_precision IN ('exact', 'approximate', 'month', 'year', 'unknown')),
    description     TEXT NOT NULL,
    entity_ids      UUID[] DEFAULT '{}',
    session_id      UUID REFERENCES sessions(id) ON DELETE SET NULL,
    is_player_known BOOLEAN DEFAULT false,
    source_document TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE timeline_events IS 'In-game chronological events';
COMMENT ON COLUMN timeline_events.event_date IS 'In-game date when the event occurred';
COMMENT ON COLUMN timeline_events.is_player_known IS 'Whether players have learned about this event';
COMMENT ON COLUMN timeline_events.entity_ids IS 'Array of entity IDs involved in this event';

CREATE INDEX idx_timeline_campaign_id ON timeline_events(campaign_id);
CREATE INDEX idx_timeline_date ON timeline_events(event_date);
CREATE INDEX idx_timeline_player_known ON timeline_events(is_player_known);
CREATE INDEX idx_timeline_entity_ids ON timeline_events USING GIN(entity_ids);

-- ============================================
-- Canon Conflicts Table
-- Track contradictions between sources
-- ============================================
CREATE TABLE canon_conflicts (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    campaign_id         UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    entity_id           UUID REFERENCES entities(id) ON DELETE CASCADE,
    field_name          TEXT,
    conflicting_values  JSONB NOT NULL,
    status              TEXT DEFAULT 'DETECTED' CHECK (status IN ('DETECTED', 'ACKNOWLEDGED', 'RESOLVED')),
    resolution          TEXT,
    resolved_at         TIMESTAMPTZ,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE canon_conflicts IS 'Track contradictions between different sources';
COMMENT ON COLUMN canon_conflicts.conflicting_values IS 'Array of {value, source, date} objects';
COMMENT ON COLUMN canon_conflicts.status IS 'DETECTED (new), ACKNOWLEDGED (seen), RESOLVED (fixed)';

CREATE INDEX idx_canon_conflicts_campaign ON canon_conflicts(campaign_id);
CREATE INDEX idx_canon_conflicts_entity ON canon_conflicts(entity_id);
CREATE INDEX idx_canon_conflicts_status ON canon_conflicts(status);

-- ============================================
-- Updated At Trigger Function
-- Automatically updates updated_at timestamp
-- ============================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply trigger to all tables with updated_at
CREATE TRIGGER update_campaigns_updated_at BEFORE UPDATE ON campaigns
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sessions_updated_at BEFORE UPDATE ON sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_entities_updated_at BEFORE UPDATE ON entities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_relationships_updated_at BEFORE UPDATE ON relationships
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_timeline_events_updated_at BEFORE UPDATE ON timeline_events
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Migration Tracking Table
-- ============================================
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO schema_migrations (version) VALUES ('001_initial_schema');
