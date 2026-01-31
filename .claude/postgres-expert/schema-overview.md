# Schema Overview

Database schema documentation for Imagineer.

## Entity Relationship Diagram

```
game_systems
    │
    ▼
campaigns ──────────┬──────────┬──────────┬──────────┐
    │               │          │          │          │
    ▼               ▼          ▼          ▼          ▼
entities      sessions   relationships  timeline   conflicts
                                         events
```

## Table Definitions

### game_systems

```sql
CREATE TABLE game_systems (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    code TEXT NOT NULL UNIQUE,
    attribute_schema JSONB NOT NULL DEFAULT '{}',
    skill_schema JSONB NOT NULL DEFAULT '{}',
    dice_conventions JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT now()
);
```

### campaigns

```sql
CREATE TABLE campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    system_id UUID NOT NULL REFERENCES game_systems(id),
    description TEXT,
    settings JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);
```

### entities

```sql
CREATE TABLE entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id),
    entity_type TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    attributes JSONB NOT NULL DEFAULT '{}',
    tags TEXT[] DEFAULT '{}',
    keeper_notes TEXT,
    discovered_session UUID REFERENCES sessions(id),
    source_document TEXT,
    source_confidence TEXT DEFAULT 'DRAFT',
    version INTEGER DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);
```

### relationships

```sql
CREATE TABLE relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id),
    source_entity_id UUID NOT NULL REFERENCES entities(id),
    target_entity_id UUID NOT NULL REFERENCES entities(id),
    relationship_type TEXT NOT NULL,
    tone TEXT,
    description TEXT,
    bidirectional BOOLEAN DEFAULT false,
    strength INTEGER,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);
```

### sessions

```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id),
    session_number INTEGER,
    planned_date DATE,
    actual_date DATE,
    status TEXT DEFAULT 'PLANNED',
    prep_notes TEXT,
    planned_scenes JSONB DEFAULT '[]',
    actual_notes TEXT,
    discoveries JSONB DEFAULT '[]',
    player_decisions JSONB DEFAULT '[]',
    consequences JSONB DEFAULT '[]',
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);
```

### timeline_events

```sql
CREATE TABLE timeline_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id),
    event_date DATE,
    event_time TIME,
    date_precision TEXT DEFAULT 'unknown',
    description TEXT NOT NULL,
    entity_ids UUID[] DEFAULT '{}',
    session_id UUID REFERENCES sessions(id),
    is_player_known BOOLEAN DEFAULT false,
    source_document TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);
```

### canon_conflicts

```sql
CREATE TABLE canon_conflicts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id),
    entity_id UUID REFERENCES entities(id),
    field_name TEXT,
    conflicting_values JSONB NOT NULL,
    status TEXT DEFAULT 'DETECTED',
    resolution TEXT,
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now()
);
```

## Indexes

### Foreign Key Indexes

```sql
CREATE INDEX idx_campaigns_system_id ON campaigns(system_id);
CREATE INDEX idx_entities_campaign_id ON entities(campaign_id);
CREATE INDEX idx_relationships_campaign_id ON relationships(campaign_id);
CREATE INDEX idx_relationships_source ON relationships(source_entity_id);
CREATE INDEX idx_relationships_target ON relationships(target_entity_id);
```

### Query Indexes

```sql
-- Entity search by type
CREATE INDEX idx_entities_type ON entities(entity_type);

-- JSONB containment queries
CREATE INDEX idx_entities_attributes ON entities USING GIN(attributes);

-- Array searches
CREATE INDEX idx_entities_tags ON entities USING GIN(tags);

-- Fuzzy name matching
CREATE INDEX idx_entities_name_trgm ON entities USING GIN(name gin_trgm_ops);
```

## Common Queries

### Find entities by name

```sql
SELECT * FROM entities
WHERE campaign_id = $1
  AND name % $2  -- Trigram similarity
ORDER BY similarity(name, $2) DESC
LIMIT 10;
```

### Find relationships for entity

```sql
SELECT r.*, e.name as target_name
FROM relationships r
JOIN entities e ON r.target_entity_id = e.id
WHERE r.source_entity_id = $1;
```

### Timeline for campaign

```sql
SELECT * FROM timeline_events
WHERE campaign_id = $1
ORDER BY event_date, event_time;
```
