# Imagineer Data Schemas

## Core Entities

### Campaign
```sql
campaigns (
    id              UUID PRIMARY KEY,
    name            TEXT NOT NULL,
    system_id       UUID REFERENCES game_systems(id),
    description     TEXT,
    settings        JSONB,          -- campaign-specific settings
    created_at      TIMESTAMPTZ,
    updated_at      TIMESTAMPTZ
)
```

### GameSystem
```sql
game_systems (
    id                      UUID PRIMARY KEY,
    name                    TEXT NOT NULL,       -- "Call of Cthulhu 7e"
    code                    TEXT UNIQUE,         -- "coc-7e"
    attribute_schema        JSONB,               -- defines stats for this system
    skill_schema            JSONB,
    character_sheet_template JSONB,
    dice_conventions        JSONB,
    created_at              TIMESTAMPTZ
)
```

### Entity (Polymorphic)
```sql
entities (
    id                  UUID PRIMARY KEY,
    campaign_id         UUID REFERENCES campaigns(id),
    entity_type         TEXT NOT NULL,      -- 'npc', 'location', 'item', 'faction', 'clue', etc.
    name                TEXT NOT NULL,
    description         TEXT,
    attributes          JSONB,              -- system-specific stats
    tags                TEXT[],
    keeper_notes        TEXT,               -- never player-visible
    discovered_session  UUID REFERENCES sessions(id),
    source_document     TEXT,               -- canon authority
    source_confidence   TEXT DEFAULT 'DRAFT', -- DRAFT, AUTHORITATIVE, SUPERSEDED
    version             INT DEFAULT 1,
    created_at          TIMESTAMPTZ,
    updated_at          TIMESTAMPTZ
)
```

### Relationship
```sql
relationships (
    id                  UUID PRIMARY KEY,
    campaign_id         UUID REFERENCES campaigns(id),
    source_entity_id    UUID REFERENCES entities(id),
    target_entity_id    UUID REFERENCES entities(id),
    relationship_type   TEXT NOT NULL,      -- 'knows', 'owns', 'located_at', 'opposes', etc.
    tone                TEXT,               -- 'friendly', 'hostile', 'romantic', 'professional'
    description         TEXT,
    bidirectional       BOOLEAN DEFAULT false,
    strength            INT,                -- optional graph weight 1-10
    created_at          TIMESTAMPTZ,
    updated_at          TIMESTAMPTZ
)
```

### Session
```sql
sessions (
    id                  UUID PRIMARY KEY,
    campaign_id         UUID REFERENCES campaigns(id),
    session_number      INT,
    planned_date        DATE,
    actual_date         DATE,
    status              TEXT DEFAULT 'PLANNED', -- PLANNED, COMPLETED, SKIPPED
    prep_notes          TEXT,
    planned_scenes      JSONB,
    actual_notes        TEXT,               -- post-session
    discoveries         JSONB,              -- [{entity_id, how_discovered}]
    player_decisions    JSONB,
    consequences        JSONB,
    created_at          TIMESTAMPTZ,
    updated_at          TIMESTAMPTZ
)
```

### Timeline
```sql
timeline_events (
    id                  UUID PRIMARY KEY,
    campaign_id         UUID REFERENCES campaigns(id),
    event_date          DATE,               -- in-game date
    event_time          TIME,               -- optional
    date_precision      TEXT,               -- 'exact', 'approximate', 'month', 'year'
    description         TEXT NOT NULL,
    entity_ids          UUID[],             -- related entities
    session_id          UUID REFERENCES sessions(id), -- when revealed
    is_player_known     BOOLEAN DEFAULT false,
    source_document     TEXT,
    created_at          TIMESTAMPTZ,
    updated_at          TIMESTAMPTZ
)
```

### CanonConflict
```sql
canon_conflicts (
    id                  UUID PRIMARY KEY,
    campaign_id         UUID REFERENCES campaigns(id),
    entity_id           UUID REFERENCES entities(id),
    field_name          TEXT,
    conflicting_values  JSONB,              -- [{value, source, date}]
    status              TEXT DEFAULT 'DETECTED', -- DETECTED, ACKNOWLEDGED, RESOLVED
    resolution          TEXT,
    resolved_at         TIMESTAMPTZ,
    created_at          TIMESTAMPTZ
)
```

## Game System Schemas

See individual files in `schemas/` directory:
- `coc-7e.yaml` - Call of Cthulhu 7th Edition
- `gurps-4e.yaml` - GURPS 4th Edition
- `fitd.yaml` - Forged in the Dark (Blades, Scum & Villainy, etc.)
