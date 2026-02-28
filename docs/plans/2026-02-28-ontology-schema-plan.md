<!--
  Imagineer - TTRPG Campaign Intelligence Platform

  Copyright (c) 2025 - 2026
  This software is released under The MIT License
-->

# Ontology Schema Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use
> superpowers:executing-plans to implement this plan
> task-by-task.

**Goal:** Replace the ad-hoc constraint mechanism with a
formal, evolvable ontology schema that seeds each campaign
with ~80 relationship types, a hierarchical entity type
system, domain/range constraints, and an era-based temporal
model.

**Architecture:** YAML files in `schemas/ontology/` define
the global ontology. A Go loader package (`internal/ontology/`)
parses YAML and seeds campaign-scoped database tables on
campaign creation. The graph expert validates against the
populated constraint tables. Advisory enforcement warns GMs
but never blocks creative decisions.

**Tech Stack:** Go 1.22, PostgreSQL 16, `gopkg.in/yaml.v3`,
`testify`, React/TypeScript (UI tasks).

**Design document:**
`docs/plans/2026-02-28-ontology-schema-design.md`

---

## Phase 1: YAML Ontology Files

### Task 1: Create Entity Types YAML

Define the entity type hierarchy with abstract parents,
concrete types, and sub-types.

**Files:**

- Create: `schemas/ontology/entity-types.yaml`

**Step 1: Write the YAML file**

```yaml
# Entity Type Hierarchy
#
# Abstract types group concrete types for constraint
# inheritance. Campaigns can add sub-types under any
# type.

types:
  # --- Meta-game (outside the fiction) ---
  person:
    abstract: true
    description: >-
      A real-world participant in the game.
    children:
      - player
      - game_master
      - character

  player:
    parent: person
    abstract: false
    description: >-
      A real person at the table.

  game_master:
    parent: person
    abstract: false
    description: >-
      The narrator or GM.

  # --- Characters (bridge real and fictional) ---
  character:
    parent: person
    abstract: true
    description: >-
      An in-world persona.
    children:
      - pc
      - npc

  pc:
    parent: character
    abstract: false
    description: >-
      Player character — a protagonist played by a player.

  npc:
    parent: character
    abstract: false
    description: >-
      Non-player character — controlled by the GM.

  # --- Agents (can act in the fiction) ---
  agent:
    abstract: true
    description: >-
      An entity that can act, decide, or exert influence.
    children:
      - character
      - creature
      - faction
      - organization

  creature:
    parent: agent
    abstract: false
    description: >-
      A monster or non-human entity.
    children:
      - beast
      - undead
      - construct
      - spirit
      - deity
      - aberration

  beast:
    parent: creature
    abstract: false
    description: Natural animal or monster.

  undead:
    parent: creature
    abstract: false
    description: Animated dead — vampire, zombie, lich.

  construct:
    parent: creature
    abstract: false
    description: Artificial being — golem, android.

  spirit:
    parent: creature
    abstract: false
    description: Incorporeal entity — ghost, elemental, fey.

  deity:
    parent: creature
    abstract: false
    description: God, demigod, or cosmic entity.

  aberration:
    parent: creature
    abstract: false
    description: Alien or eldritch being.

  faction:
    parent: agent
    abstract: false
    description: >-
      A group with shared goals.

  organization:
    parent: agent
    abstract: false
    description: >-
      A formal institution.
    children:
      - government
      - corporation
      - cult
      - guild
      - military
      - criminal

  government:
    parent: organization
    abstract: false
    description: State or sovereign body.

  corporation:
    parent: organization
    abstract: false
    description: Commercial enterprise.

  cult:
    parent: organization
    abstract: false
    description: Secretive religious or ideological group.

  guild:
    parent: organization
    abstract: false
    description: Professional or trade association.

  military:
    parent: organization
    abstract: false
    description: Armed force or regiment.

  criminal:
    parent: organization
    abstract: false
    description: Organised crime syndicate or gang.

  # --- Places ---
  place:
    abstract: true
    description: >-
      A physical or conceptual location.
    children:
      - location

  location:
    parent: place
    abstract: false
    description: >-
      A physical or conceptual place. Locations nest
      via the part_of relationship.

  # --- Artifacts ---
  artifact:
    abstract: true
    description: >-
      A tangible object or record.
    children:
      - item
      - document

  item:
    parent: artifact
    abstract: false
    description: >-
      A physical object or tool.
    children:
      - weapon
      - armor
      - vehicle
      - treasure
      - relic
      - tool
      - consumable

  weapon:
    parent: item
    abstract: false
    description: Instrument of harm.

  armor:
    parent: item
    abstract: false
    description: Protective equipment.

  vehicle:
    parent: item
    abstract: false
    description: Conveyance — ship, cart, starship.

  treasure:
    parent: item
    abstract: false
    description: Valuable hoard or cache.

  relic:
    parent: item
    abstract: false
    description: Significant artifact or magical item.

  tool:
    parent: item
    abstract: false
    description: Practical device or instrument.

  consumable:
    parent: item
    abstract: false
    description: Single-use item — potion, scroll, ration.

  document:
    parent: artifact
    abstract: false
    description: >-
      A written record or text.
    children:
      - spell
      - map
      - letter
      - prophecy
      - contract
      - journal

  spell:
    parent: document
    abstract: false
    description: Magical text or ritual instructions.

  map:
    parent: document
    abstract: false
    description: Cartographic record.

  letter:
    parent: document
    abstract: false
    description: Correspondence or message.

  prophecy:
    parent: document
    abstract: false
    description: Prediction or vision record.

  contract:
    parent: document
    abstract: false
    description: Binding agreement or deed.

  journal:
    parent: document
    abstract: false
    description: Personal diary or log.

  # --- Narrative ---
  narrative:
    abstract: true
    description: >-
      A story element that connects other entities.
    children:
      - event
      - clue

  event:
    parent: narrative
    abstract: false
    description: >-
      A notable occurrence.
    children:
      - battle
      - ritual
      - disaster
      - discovery
      - betrayal_event
      - celebration

  battle:
    parent: event
    abstract: false
    description: Armed conflict or skirmish.

  ritual:
    parent: event
    abstract: false
    description: Ceremony, summoning, or formal rite.

  disaster:
    parent: event
    abstract: false
    description: Catastrophe — natural or caused.

  discovery:
    parent: event
    abstract: false
    description: Finding or revelation.

  betrayal_event:
    parent: event
    abstract: false
    description: Act of treachery or broken faith.

  celebration:
    parent: event
    abstract: false
    description: Festival, feast, or victory.

  clue:
    parent: narrative
    abstract: false
    description: >-
      A piece of evidence or information.
```

**Step 2: Verify YAML parses correctly**

Run: `python3 -c "import yaml; yaml.safe_load(open('schemas/ontology/entity-types.yaml'))"`
Expected: no error.

**Step 3: Commit**

```
git add schemas/ontology/entity-types.yaml
git commit -m "feat: add entity type hierarchy YAML"
```

---

### Task 2: Create Relationship Types YAML

Define all ~80 relationship types organised by narrative
category.

**Files:**

- Create: `schemas/ontology/relationship-types.yaml`

**Step 1: Write the YAML file**

The file defines each relationship type with forward name,
inverse name, symmetry flag, display labels, domain, range,
description, and genre tags.

Structure per type:

```yaml
types:
  # --- Meta-Game ---
  plays:
    inverse: played_by
    symmetric: false
    display_label: Plays
    inverse_display_label: Played by
    domain: [player]
    range: [pc]
    genre: [universal]
    description: >-
      Active player-character bond.

  # ... all ~80 types following the same structure,
  # organised by category as defined in the design
  # document sections: Meta-Game, Kinship, Social,
  # Power, Spatial, Possession, Knowledge, Conflict,
  # Affiliation, Supernatural, Temporal, Economic,
  # Event, Sci-Fi, Horror, Superhero, Historical,
  # Romance.
```

Every relationship type from the design document's tables
goes into this file. The 14 existing template types
(`owns`, `employs`, `works_for`, `reports_to`, `parent_of`,
`located_at`, `member_of`, `created`, `rules`,
`headquartered_at`, `knows`, `friend_of`, `enemy_of`,
`allied_with`) must be included with their existing names
to maintain backward compatibility.

**Step 2: Verify YAML parses**

Run: `python3 -c "import yaml; d=yaml.safe_load(open('schemas/ontology/relationship-types.yaml')); print(len(d['types']), 'types')"`
Expected: approximately 80 types.

**Step 3: Commit**

```
git add schemas/ontology/relationship-types.yaml
git commit -m "feat: add relationship types YAML (~80 types)"
```

---

### Task 3: Create Constraints YAML

Define domain/range constraints, cardinality defaults, and
required relationships.

**Files:**

- Create: `schemas/ontology/constraints.yaml`

**Step 1: Write the YAML file**

```yaml
# Ontology Constraints
#
# All constraints are advisory (warn, not block).
# Domain/range reference abstract parent types where
# possible; the loader resolves to concrete children.

domain_range:
  # Each key is a relationship type name.
  # domain: valid source entity types.
  # range: valid target entity types.
  plays:
    domain: [player]
    range: [pc]
  formerly_played:
    domain: [player]
    range: [pc, npc]
  narrates:
    domain: [game_master]
    range: [any]
  parent_of:
    domain: [npc, creature]
    range: [npc, creature]
  sibling_of:
    domain: [npc]
    range: [npc]
  spouse_of:
    domain: [npc]
    range: [npc]
  # ... all relationship types with domain/range
  # constraints as defined in the design document
  # tables. Types not listed here are unconstrained.

# Cardinality defaults (all many-to-many).
# Campaigns can tighten these.
cardinality: {}

# Required relationships (advisory).
# Every entity of the given type should have at least
# one relationship of the named type.
required:
  npc:
    - located_at
  pc:
    - played_by
    - located_at
  creature:
    - located_at
  faction:
    - headquartered_at
  organization:
    - headquartered_at
```

**Step 2: Verify YAML parses**

Run: `python3 -c "import yaml; yaml.safe_load(open('schemas/ontology/constraints.yaml'))"`
Expected: no error.

**Step 3: Commit**

```
git add schemas/ontology/constraints.yaml
git commit -m "feat: add ontology constraints YAML"
```

---

## Phase 2: Database Migration

### Task 4: Write the Migration

Add new tables for the ontology schema. This migration adds
tables only; it does not modify existing tables yet.

**Files:**

- Create: `migrations/003_ontology_schema.sql`
- Test: manual verification via `make migrate`

**Step 1: Write the migration SQL**

```sql
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
    UNIQUE(campaign_id, name)
);

COMMENT ON TABLE campaign_entity_types IS
    'Campaign-scoped entity type hierarchy seeded '
    'from schemas/ontology/entity-types.yaml';

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

-- Add era_id column to existing relationships table
ALTER TABLE relationships
    ADD COLUMN era_id BIGINT
    REFERENCES eras(id) ON DELETE SET NULL;

-- Add era_id column to entities table
ALTER TABLE entities
    ADD COLUMN era_id BIGINT
    REFERENCES eras(id) ON DELETE SET NULL;
```

**Step 2: Apply the migration locally**

Run: `make migrate`
Expected: migration applies without error.

**Step 3: Verify tables exist**

Run:
```
docker exec imagineer-postgres psql -U imagineer \
  -d imagineer -c "\dt campaign_entity_types"
```
Expected: table listed.

**Step 4: Commit**

```
git add migrations/003_ontology_schema.sql
git commit -m "feat: add ontology schema migration"
```

---

## Phase 3: YAML Loader Package

### Task 5: Create Ontology Types and Parser

A new `internal/ontology/` package that parses the YAML
files into Go structs.

**Files:**

- Create: `internal/ontology/types.go`
- Create: `internal/ontology/loader.go`
- Create: `internal/ontology/loader_test.go`

**Step 1: Write the failing test**

`internal/ontology/loader_test.go`:

```go
package ontology

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestLoadEntityTypes(t *testing.T) {
    et, err := LoadEntityTypes(
        "../../schemas/ontology/entity-types.yaml",
    )
    require.NoError(t, err)
    require.NotNil(t, et)

    // Check abstract parents exist.
    assert.True(t, et.Types["agent"].Abstract)
    assert.True(t, et.Types["person"].Abstract)

    // Check concrete types exist.
    assert.False(t, et.Types["npc"].Abstract)
    assert.Equal(t, "character", et.Types["npc"].Parent)

    // Check sub-types.
    assert.Equal(t, "creature",
        et.Types["undead"].Parent)
}

func TestLoadRelationshipTypes(t *testing.T) {
    rt, err := LoadRelationshipTypes(
        "../../schemas/ontology/relationship-types.yaml",
    )
    require.NoError(t, err)
    require.NotNil(t, rt)

    // Check an existing type.
    owns := rt.Types["owns"]
    assert.Equal(t, "owned_by", owns.Inverse)
    assert.False(t, owns.Symmetric)

    // Check a symmetric type.
    knows := rt.Types["knows"]
    assert.True(t, knows.Symmetric)

    // Check total count is approximately 80.
    assert.Greater(t, len(rt.Types), 70)
}

func TestLoadConstraints(t *testing.T) {
    c, err := LoadConstraints(
        "../../schemas/ontology/constraints.yaml",
    )
    require.NoError(t, err)
    require.NotNil(t, c)

    // Check domain/range for a known type.
    plays := c.DomainRange["plays"]
    assert.Contains(t, plays.Domain, "player")
    assert.Contains(t, plays.Range, "pc")

    // Check required relationships.
    assert.Contains(t, c.Required["npc"], "located_at")
}

func TestResolveConcreteTypes(t *testing.T) {
    et, err := LoadEntityTypes(
        "../../schemas/ontology/entity-types.yaml",
    )
    require.NoError(t, err)

    // "agent" should resolve to npc, pc, creature,
    // faction, organization and all their sub-types.
    concrete := et.ResolveToConcreteTypes("agent")
    assert.Contains(t, concrete, "npc")
    assert.Contains(t, concrete, "pc")
    assert.Contains(t, concrete, "creature")
    assert.Contains(t, concrete, "undead")
    assert.NotContains(t, concrete, "agent")
}
```

**Step 2: Run the test to verify it fails**

Run: `go test -v ./internal/ontology/...`
Expected: FAIL — package does not exist.

**Step 3: Write the types**

`internal/ontology/types.go`:

```go
package ontology

// EntityTypeFile represents the parsed
// entity-types.yaml file.
type EntityTypeFile struct {
    Types map[string]EntityTypeDef `yaml:"types"`
}

// EntityTypeDef defines a single entity type.
type EntityTypeDef struct {
    Parent      string   `yaml:"parent"`
    Abstract    bool     `yaml:"abstract"`
    Description string   `yaml:"description"`
    Children    []string `yaml:"children"`
}

// ResolveToConcreteTypes returns all non-abstract
// descendants of the given type name, including the
// type itself if it is concrete.
func (f *EntityTypeFile) ResolveToConcreteTypes(
    name string,
) []string {
    var result []string
    f.collectConcrete(name, &result)
    return result
}

func (f *EntityTypeFile) collectConcrete(
    name string, result *[]string,
) {
    def, ok := f.Types[name]
    if !ok {
        return
    }
    if !def.Abstract {
        *result = append(*result, name)
    }
    for _, child := range def.Children {
        f.collectConcrete(child, result)
    }
}

// RelationshipTypeFile represents the parsed
// relationship-types.yaml file.
type RelationshipTypeFile struct {
    Types map[string]RelationshipTypeDef `yaml:"types"`
}

// RelationshipTypeDef defines a single relationship
// type.
type RelationshipTypeDef struct {
    Inverse             string   `yaml:"inverse"`
    Symmetric           bool     `yaml:"symmetric"`
    DisplayLabel        string   `yaml:"display_label"`
    InverseDisplayLabel string   `yaml:"inverse_display_label"`
    Domain              []string `yaml:"domain"`
    Range               []string `yaml:"range"`
    Genre               []string `yaml:"genre"`
    Description         string   `yaml:"description"`
}

// ConstraintsFile represents the parsed
// constraints.yaml file.
type ConstraintsFile struct {
    DomainRange map[string]DomainRangeDef `yaml:"domain_range"`
    Cardinality map[string]CardinalityDef `yaml:"cardinality"`
    Required    map[string][]string       `yaml:"required"`
}

// DomainRangeDef defines valid source and target
// entity types for a relationship type.
type DomainRangeDef struct {
    Domain []string `yaml:"domain"`
    Range  []string `yaml:"range"`
}

// CardinalityDef defines limits on how many
// relationships of a type an entity can have.
type CardinalityDef struct {
    MaxSource *int `yaml:"max_source"`
    MaxTarget *int `yaml:"max_target"`
}
```

**Step 4: Write the loader**

`internal/ontology/loader.go`:

```go
package ontology

import (
    "fmt"
    "os"

    "gopkg.in/yaml.v3"
)

// LoadEntityTypes parses the entity types YAML file.
func LoadEntityTypes(path string) (
    *EntityTypeFile, error,
) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf(
            "read entity types: %w", err)
    }
    var f EntityTypeFile
    if err := yaml.Unmarshal(data, &f); err != nil {
        return nil, fmt.Errorf(
            "parse entity types: %w", err)
    }
    return &f, nil
}

// LoadRelationshipTypes parses the relationship
// types YAML file.
func LoadRelationshipTypes(path string) (
    *RelationshipTypeFile, error,
) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf(
            "read relationship types: %w", err)
    }
    var f RelationshipTypeFile
    if err := yaml.Unmarshal(data, &f); err != nil {
        return nil, fmt.Errorf(
            "parse relationship types: %w", err)
    }
    return &f, nil
}

// LoadConstraints parses the constraints YAML file.
func LoadConstraints(path string) (
    *ConstraintsFile, error,
) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf(
            "read constraints: %w", err)
    }
    var f ConstraintsFile
    if err := yaml.Unmarshal(data, &f); err != nil {
        return nil, fmt.Errorf(
            "parse constraints: %w", err)
    }
    return &f, nil
}
```

**Step 5: Run tests to verify they pass**

Run: `go test -v ./internal/ontology/...`
Expected: PASS.

**Step 6: Commit**

```
git add internal/ontology/
git commit -m "feat: add ontology YAML loader package"
```

---

### Task 6: Campaign Seeder

Functions that populate campaign-scoped tables from parsed
YAML structs.

**Files:**

- Create: `internal/ontology/seeder.go`
- Create: `internal/ontology/seeder_test.go`

**Step 1: Write the failing test**

`internal/ontology/seeder_test.go` — test that
`SeedCampaignEntityTypes` inserts rows into
`campaign_entity_types` for a given campaign ID. Use a
test database or mock the DB interface.

The seeder needs these functions:

- `SeedCampaignEntityTypes(ctx, db, campaignID,
  *EntityTypeFile) error`
- `SeedCampaignRelationshipTypes(ctx, db, campaignID,
  *RelationshipTypeFile) error`
- `SeedCampaignConstraints(ctx, db, campaignID,
  *ConstraintsFile, *EntityTypeFile) error`
- `SeedCampaignRequiredRelationships(ctx, db,
  campaignID, *ConstraintsFile) error`
- `SeedDefaultEra(ctx, db, campaignID) error`

For constraint seeding, the `EntityTypeFile` is needed to
resolve abstract parent types to concrete types in the
`relationship_type_constraints` table.

`SeedDefaultEra` creates a single era named "Present Day"
with sequence 1 and scale "now" for every new campaign.

**Step 2: Write the seeder implementation**

Each function uses batch INSERT statements for efficiency.
`SeedCampaignRelationshipTypes` replaces the current
template-copying SQL in `campaigns.go:308-319`.

**Step 3: Run tests**

Run: `go test -v ./internal/ontology/...`
Expected: PASS.

**Step 4: Commit**

```
git add internal/ontology/seeder.go \
        internal/ontology/seeder_test.go
git commit -m "feat: add ontology campaign seeder"
```

---

## Phase 4: Campaign Creation Integration

### Task 7: Replace Template Copying with Ontology Seeding

Modify `CreateCampaignWithOwner` to use the ontology
seeder instead of the INSERT...SELECT from
`relationship_type_templates`.

**Files:**

- Modify: `internal/database/campaigns.go:308-319`
- Modify: `cmd/server/main.go` (load YAML at startup)
- Test: existing campaign creation tests must still pass

**Step 1: Load ontology YAML at server startup**

In `cmd/server/main.go`, load the three YAML files into
a global `*ontology.Ontology` struct and pass it to the
handler or DB layer. The struct holds all three parsed
files.

**Step 2: Modify CreateCampaignWithOwner**

Replace the INSERT...SELECT from
`relationship_type_templates` (lines 308-319) with calls
to:

```go
ontology.SeedCampaignEntityTypes(ctx, db, c.ID, ont.EntityTypes)
ontology.SeedCampaignRelationshipTypes(ctx, db, c.ID, ont.RelationshipTypes)
ontology.SeedCampaignConstraints(ctx, db, c.ID, ont.Constraints, ont.EntityTypes)
ontology.SeedCampaignRequiredRelationships(ctx, db, c.ID, ont.Constraints)
ontology.SeedDefaultEra(ctx, db, c.ID)
```

**Step 3: Run existing tests**

Run: `go test -v ./internal/database/... ./internal/api/...`
Expected: PASS. The campaign creation tests must still
work. The seeded relationship types must include all 14
original templates (backward compatibility).

**Step 4: Commit**

```
git add internal/database/campaigns.go cmd/server/main.go
git commit -m "feat: seed campaigns from ontology YAML"
```

---

## Phase 5: Entity Type Dynamic Lookup

### Task 8: Replace Hardcoded Entity Type Enum

Replace the CHECK constraint on `entities.entity_type`
and the Go `EntityType` constants with a dynamic lookup
against `campaign_entity_types`.

**Files:**

- Modify: `migrations/003_ontology_schema.sql` (add
  ALTER TABLE to drop CHECK, add FK-like validation)
- Modify: `internal/models/models.go:69-83`
- Modify: `internal/database/entities.go`
- Modify: `internal/api/handlers.go:478-516`
- Test: `internal/database/entities_test.go`
- Test: `internal/api/handlers_test.go`

**Step 1: Migration change**

Add to the migration:

```sql
-- Drop the hardcoded CHECK constraint on entity_type.
-- Validation is now against campaign_entity_types.
ALTER TABLE entities
    DROP CONSTRAINT IF EXISTS entities_entity_type_check;
```

The Go model retains `EntityType` as a `string` type but
removes the const block of fixed values. The existing
constants remain as convenience references but are not
used for validation.

**Step 2: Add validation function**

`internal/database/entities.go` — add a function:

```go
// ValidateEntityType checks if the given entity type
// is valid for the campaign (exists in
// campaign_entity_types and is not abstract).
func (db *DB) ValidateEntityType(
    ctx context.Context,
    campaignID int64,
    entityType string,
) (bool, error) {
    query := `
        SELECT EXISTS (
            SELECT 1 FROM campaign_entity_types
            WHERE campaign_id = $1
              AND name = $2
              AND abstract = false
        )`
    var valid bool
    err := db.QueryRow(ctx, query,
        campaignID, entityType).Scan(&valid)
    return valid, err
}
```

**Step 3: Update CreateEntity handler**

In `handlers.go:478-516`, after decoding the request and
before creating the entity, call `ValidateEntityType`.
If the type is not valid, return a warning in the
response (advisory, not blocking — per design).

**Step 4: Run tests**

Run: `make test-all`
Expected: PASS.

**Step 5: Commit**

```
git add migrations/003_ontology_schema.sql \
        internal/models/models.go \
        internal/database/entities.go \
        internal/api/handlers.go \
        internal/database/entities_test.go
git commit -m "feat: replace hardcoded entity type enum \
with dynamic campaign-scoped lookup"
```

---

## Phase 6: Eras System

### Task 9: Era CRUD Operations

Add database functions for managing eras.

**Files:**

- Create: `internal/database/eras.go`
- Create: `internal/database/eras_test.go`

**Step 1: Write failing tests**

Test `CreateEra`, `ListEras`, `GetEra`,
`GetCurrentEra`, `UpdateEra`.

**Step 2: Implement CRUD**

```go
func (db *DB) CreateEra(ctx context.Context,
    campaignID int64,
    req models.CreateEraRequest,
) (*models.Era, error)

func (db *DB) ListEras(ctx context.Context,
    campaignID int64,
) ([]models.Era, error)

func (db *DB) GetCurrentEra(ctx context.Context,
    campaignID int64,
) (*models.Era, error)
// Returns the era with the highest sequence number.

func (db *DB) UpdateEra(ctx context.Context,
    id int64, req models.UpdateEraRequest,
) (*models.Era, error)
```

**Step 3: Add Era model to models.go**

```go
type Era struct {
    ID          int64     `json:"id"`
    CampaignID  int64     `json:"campaignId"`
    Sequence    int       `json:"sequence"`
    Name        string    `json:"name"`
    Scale       string    `json:"scale"`
    Description *string   `json:"description,omitempty"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}
```

**Step 4: Run tests**

Run: `go test -v ./internal/database/...`
Expected: PASS.

**Step 5: Commit**

```
git add internal/database/eras.go \
        internal/database/eras_test.go \
        internal/models/models.go
git commit -m "feat: add era CRUD operations"
```

---

### Task 10: Relationship Archiving

When a relationship changes, archive the old state with
its era.

**Files:**

- Create: `internal/database/relationship_archive.go`
- Create: `internal/database/relationship_archive_test.go`
- Modify: `internal/database/relationships.go`

**Step 1: Write failing test**

Test that `ArchiveRelationship` moves a relationship to
`relationship_archive` with the correct era ID.

**Step 2: Implement archiving**

```go
// ArchiveRelationship copies a relationship to the
// archive table and deletes it from the active table.
func (db *DB) ArchiveRelationship(
    ctx context.Context,
    relationshipID int64,
    eraID *int64,
) error
```

**Step 3: Run tests**

Run: `go test -v ./internal/database/...`
Expected: PASS.

**Step 4: Commit**

```
git add internal/database/relationship_archive.go \
        internal/database/relationship_archive_test.go \
        internal/database/relationships.go
git commit -m "feat: add relationship archiving with era"
```

---

## Phase 7: Graph Expert Enhancements

### Task 11: Cardinality Check

Add a cardinality validation check to the graph expert.

**Files:**

- Modify: `internal/agents/graph/expert.go`
- Create: `internal/agents/graph/cardinality.go`
- Create: `internal/agents/graph/cardinality_test.go`

**Step 1: Write failing test**

Test that `CheckCardinality` returns violations when an
entity exceeds the `max_source` or `max_target` limit
for a relationship type.

**Step 2: Implement CheckCardinality**

Query `cardinality_constraints` for the campaign.
Compare against the count of existing relationships
plus proposed new relationships from enrichment.

**Step 3: Wire into Expert.Run**

Add cardinality check after type pair validation in
`expert.go:122`, producing `cardinality_violation`
detection type items.

**Step 4: Run tests**

Run: `go test -v ./internal/agents/graph/...`
Expected: PASS.

**Step 5: Commit**

```
git add internal/agents/graph/cardinality.go \
        internal/agents/graph/cardinality_test.go \
        internal/agents/graph/expert.go
git commit -m "feat: add cardinality check to graph expert"
```

---

### Task 12: Required Relationship Check

Add required relationship validation to the graph expert.

**Files:**

- Modify: `internal/agents/graph/expert.go`
- Create: `internal/agents/graph/required.go`
- Create: `internal/agents/graph/required_test.go`

**Step 1: Write failing test**

Test that `CheckRequiredRelationships` flags an NPC
without a `located_at` relationship when the required
relationships table mandates it.

**Step 2: Implement CheckRequiredRelationships**

Query `required_relationships` for the campaign. For
each entity in the pipeline input, check if it has the
required relationship types.

**Step 3: Wire into Expert.Run**

Add after cardinality check, producing
`missing_required` detection type items. This upgrades
the existing orphan detection with more specific
findings.

**Step 4: Run tests**

Run: `go test -v ./internal/agents/graph/...`
Expected: PASS.

**Step 5: Commit**

```
git add internal/agents/graph/required.go \
        internal/agents/graph/required_test.go \
        internal/agents/graph/expert.go
git commit -m "feat: add required relationship check"
```

---

### Task 13: Constraint Override Awareness

Teach the graph expert to consult `constraint_overrides`
before flagging violations.

**Files:**

- Modify: `internal/agents/graph/expert.go`
- Modify: `internal/agents/graph/ontology.go`
- Test: `internal/agents/graph/expert_test.go`

**Step 1: Write failing test**

Test that a type pair violation is NOT produced when a
matching `constraint_overrides` row exists for the
campaign.

**Step 2: Implement override lookup**

Add a function to query `constraint_overrides` and
filter out findings that match an existing override.
The `override_key` format encodes the constraint type
and specific violation (for example,
`domain_range:employs:location:event`).

**Step 3: Run tests**

Run: `go test -v ./internal/agents/graph/...`
Expected: PASS.

**Step 4: Commit**

```
git add internal/agents/graph/expert.go \
        internal/agents/graph/ontology.go \
        internal/agents/graph/expert_test.go
git commit -m "feat: constraint override awareness"
```

---

## Phase 8: Enrichment Pipeline Ontology Integration

### Task 14: Include Ontology in Enrichment Prompt

Update the enrichment agent's system prompt to include
the campaign's valid entity types, relationship types,
and domain/range constraints.

**Files:**

- Modify: `internal/enrichment/prompts.go:24-84`
- Modify: `internal/enrichment/engine.go` (pass ontology
  data to prompt builder)
- Modify: `internal/enrichment/types.go` (add ontology
  fields to `EnrichmentInput`)
- Test: `internal/enrichment/prompts_test.go`

**Step 1: Write failing test**

Test that `buildSystemPrompt` includes ontology data
when provided.

**Step 2: Extend EnrichmentInput**

Add fields for valid entity types, valid relationship
types, and constraints. These are loaded from the
campaign's `campaign_entity_types` and
`relationship_types` tables.

**Step 3: Update system prompt**

Add a section to the system prompt:

```
## Valid Entity Types
Only suggest entities with these types: npc, pc,
location, item, ...

## Valid Relationship Types
Use only these relationship types. Each type lists
which entity type pairs it accepts:
- located_at: agent -> place
- employs: agent -> npc
- ...

Do not invent new relationship types unless none of
the above fit.
```

**Step 4: Run tests**

Run: `go test -v ./internal/enrichment/...`
Expected: PASS.

**Step 5: Commit**

```
git add internal/enrichment/prompts.go \
        internal/enrichment/engine.go \
        internal/enrichment/types.go \
        internal/enrichment/prompts_test.go
git commit -m "feat: include ontology in enrichment prompt"
```

---

## Phase 9: API Endpoints

### Task 15: Era API Endpoints

Add REST endpoints for era management.

**Files:**

- Modify: `internal/api/handlers.go`
- Modify: `internal/api/routes.go`
- Test: `internal/api/handlers_test.go`

Endpoints:

- `GET /api/campaigns/:id/eras` — list eras
- `POST /api/campaigns/:id/eras` — create era
- `PUT /api/eras/:id` — update era
- `GET /api/campaigns/:id/eras/current` — get current

**Step 1: Write handler tests**

**Step 2: Implement handlers**

**Step 3: Register routes**

**Step 4: Run tests**

Run: `go test -v ./internal/api/...`
Expected: PASS.

**Step 5: Commit**

```
git add internal/api/handlers.go \
        internal/api/routes.go \
        internal/api/handlers_test.go
git commit -m "feat: add era API endpoints"
```

---

### Task 16: Constraint Override API Endpoint

Add endpoint for GM to acknowledge constraint violations.

**Files:**

- Modify: `internal/api/handlers.go`
- Create: `internal/database/constraint_overrides.go`
- Test: `internal/api/handlers_test.go`

Endpoint:

- `POST /api/campaigns/:id/constraint-overrides` —
  acknowledge a constraint violation.

Request body:

```json
{
    "constraintType": "domain_range",
    "overrideKey": "employs:location:event"
}
```

**Step 1-5: Test, implement, commit**

```
git commit -m "feat: add constraint override endpoint"
```

---

## Phase 10: Knowledge Base and Documentation

### Task 17: Update Graph Expert Knowledge Base

Update the `.claude/graph-expert/` knowledge base to
reflect the new ontology schema.

**Files:**

- Modify: `.claude/graph-expert/README.md`
- Modify: `.claude/graph-expert/ontology.md`
- Modify: `.claude/graph-expert/graph-patterns.md`
- Modify: `.claude/graph-expert/deduplication.md`

Update detection types table to include
`cardinality_violation` and `missing_required`. Update
ontology documentation to reference YAML files as the
source of truth. Update graph patterns to document the
new checks.

**Commit:**

```
git commit -m "docs: update graph-expert knowledge base \
for ontology schema"
```

---

### Task 18: Update CHANGELOG and Todo

**Files:**

- Modify: `CHANGELOG.md`
- Modify: `Todo.md`

Document all changes. Mark completed todo items. Add
any new tasks discovered during implementation.

**Commit:**

```
git commit -m "docs: update CHANGELOG and Todo for \
ontology schema"
```

---

## Task Summary

| # | Task | Type | Phase |
|---|------|------|-------|
| 1 | Entity types YAML | Config | 1 |
| 2 | Relationship types YAML | Config | 1 |
| 3 | Constraints YAML | Config | 1 |
| 4 | Database migration | SQL | 2 |
| 5 | Ontology loader package | Backend | 3 |
| 6 | Campaign seeder | Backend | 3 |
| 7 | Campaign creation integration | Backend | 4 |
| 8 | Entity type dynamic lookup | Backend | 5 |
| 9 | Era CRUD | Backend | 6 |
| 10 | Relationship archiving | Backend | 6 |
| 11 | Cardinality check | Backend | 7 |
| 12 | Required relationship check | Backend | 7 |
| 13 | Constraint override awareness | Backend | 7 |
| 14 | Enrichment prompt ontology | Backend | 8 |
| 15 | Era API endpoints | Backend | 9 |
| 16 | Constraint override endpoint | Backend | 9 |
| 17 | Knowledge base update | Docs | 10 |
| 18 | CHANGELOG and Todo | Docs | 10 |

## Dependencies

Tasks must be completed in phase order:

- Phase 1 (Tasks 1-3) has no dependencies.
- Phase 2 (Task 4) depends on Phase 1.
- Phase 3 (Tasks 5-6) depends on Phases 1 and 2.
- Phase 4 (Task 7) depends on Phase 3.
- Phase 5 (Task 8) depends on Phase 4.
- Phase 6 (Tasks 9-10) depends on Phase 2.
- Phase 7 (Tasks 11-13) depends on Phases 4 and 6.
- Phase 8 (Task 14) depends on Phase 4.
- Phase 9 (Tasks 15-16) depends on Phases 6 and 7.
- Phase 10 (Tasks 17-18) depends on all prior phases.

Within a phase, tasks can run in parallel unless they
modify the same files. Tasks 1, 2, and 3 are fully
independent. Tasks 9 and 10 are independent. Tasks 11
and 12 are independent. Tasks 15 and 16 are independent.
