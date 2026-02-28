<!--
  Imagineer - TTRPG Campaign Intelligence Platform

  Copyright (c) 2025 - 2026
  This software is released under The MIT License
-->

# Campaign Ontology Schema Design

This design document describes a formal ontology schema for
Imagineer's campaign knowledge graphs. The ontology defines
entity types, relationship types, and constraints that govern
graph structure. It replaces the current ad-hoc constraint
mechanism with a comprehensive, evolvable schema that seeds
each campaign and grows with the GM's world.

## Context and Motivation

The current system has ontology enforcement scattered across
three loosely coupled places.

A hardcoded Go and SQL entity type enum defines ten fixed
entity types. The `relationship_type_constraints` table exists
in the database but contains no seed data, so the graph expert
validates against an empty constraint set. The enrichment
pipeline's LLM prompt lists suggested relationship types by
example but has no formal schema to constrain suggestions.

This produces several problems. The enrichment agent suggests
nonsensical relationships because no domain or range
constraints exist. The graph expert cannot catch invalid type
pairs because no constraints are seeded. Campaigns cannot add
new entity types or relationship types without code changes.
There is no way to express cardinality rules or required
relationships. As the graph grows, structural integrity
degrades without formal validation.

The ontology schema addresses all of these problems while
preserving GM creative freedom through advisory enforcement.

## Design Goals

The ontology schema serves three purposes.

First, LLM guardrails. The enrichment agent receives the
campaign's ontology in its prompt so suggestions align with
valid types and constraints from the start, reducing the
volume of invalid suggestions the graph expert must catch.

Second, evolvability. GMs evolve their worlds through
storytelling. The ontology must support new entity types,
new relationship types, and new constraints without code
changes or database migrations. Campaigns start from a rich
global default and diverge independently.

Third, graph quality at scale. As campaigns grow to hundreds
or thousands of entities, the ontology enables automated
validation, cleanup, and reporting against a well-defined
structural standard.

## Architectural Decisions

### Native Schema with Future JSON-LD Export

The ontology is defined in YAML files and enforced through
PostgreSQL tables. This keeps the system simple and fast.
JSON-LD export is deferred to the publishing pipeline, where
campaign worlds need to leave the system in an interoperable
format. The native schema borrows concepts from OWL (domain
and range constraints, class hierarchy, cardinality) without
requiring OWL tooling or an RDF store.

### YAML Files as Global Source, Database for Enforcement

The global ontology lives in version-controlled YAML files
under `schemas/ontology/`. On campaign creation, the system
loads the YAML into campaign-scoped database tables. Campaigns
then evolve independently. Changes to the global YAML
propagate to new campaigns; existing campaigns keep their
version unless explicitly upgraded.

This follows the established pattern used by
`relationship_type_templates`, where template rows are copied
to campaign-scoped `relationship_types` on campaign creation.

### Advisory Enforcement (Warn, Not Block)

All constraints are advisory. Violations produce findings
for GM review, not hard rejections. When a GM acknowledges a
constraint violation (for example, accepting that a sentient
tavern can employ NPCs), the campaign's ontology evolves to
permit that pattern going forward.

This protects GM creative freedom while catching genuine
errors from LLM enrichment. The ontology sets expectations;
the GM has final authority.

### Separation from Game System Schemas

The ontology owns entity relationships and graph structure.
Game system schemas (`schemas/*.yaml`) own entity properties
and game mechanics. These two systems are complementary with
no overlap. A Call of Cthulhu tome is a `document` (sub-type
`relic`) in the ontology; its Cthulhu Mythos skill impact is
defined in `schemas/coc-7e.yaml`. Adding more game systems
does not affect the ontology, and extending the ontology does
not affect game system schemas.

## Entity Type Hierarchy

The ontology defines a type hierarchy where abstract parent
types enable constraints to be defined once and inherited by
all children. Campaigns can add sub-types under any concrete
or abstract parent.

### Abstract Parent Types

Abstract types are not instantiated directly. They exist to
group concrete types and enable constraint inheritance.

```
person (abstract)
  Description: A real-world participant in the game.
  Children: player, game_master, character

character (abstract)
  Parent: person
  Description: An in-world persona.
  Children: pc, npc

agent (abstract)
  Description: An entity that can act in the fiction.
  Children: character, creature, faction, organization

place (abstract)
  Description: A physical or conceptual location.
  Children: location

artifact (abstract)
  Description: A tangible object or record.
  Children: item, document

narrative (abstract)
  Description: A story element that connects other entities.
  Children: event, clue
```

### Concrete Types

#### Meta-Game Types (Outside the Fiction)

These represent real-world participants:

- `player` — a real person at the table. Has meta-game
  relationships to their characters. Players may come and go
  across sessions.

- `game_master` — the narrator or GM. Controls NPCs, manages
  the world, drives the story.

#### Character Types (Bridge Between Real and Fictional)

- `pc` (player character) — the protagonist, played by a
  player. PCs are at the heart of the story. When a PC dies,
  the player may create a new one or take over an existing
  NPC (which then becomes a PC).

- `npc` (non-player character) — controlled by the GM. Can
  transition to PC when a player adopts them.

The `pc` and `npc` boundary is fluid. The system tracks
transitions between these types to preserve provenance.

#### Agent Types

- `creature` — a monster or non-human entity.
  Sub-types: beast, undead, construct, spirit, deity,
  aberration.

- `faction` — a group with shared goals.

- `organization` — a formal institution.
  Sub-types: government, corporation, cult, guild, military,
  criminal.

#### Place Types

- `location` — a physical or conceptual place. Locations
  nest via the `part_of` relationship rather than through
  rigid sub-types, so the same system handles dungeon rooms,
  city districts, kingdoms, planets, and galactic sectors.

#### Artifact Types

- `item` — a physical object or tool.
  Sub-types: weapon, armor, vehicle, treasure, relic, tool,
  consumable.

- `document` — a written record or text.
  Sub-types: spell, map, letter, prophecy, contract, journal.

#### Narrative Types

- `event` — a notable occurrence.
  Sub-types: battle, ritual, disaster, discovery, betrayal,
  celebration.

- `clue` — a piece of evidence or information.

### Campaign Extensibility

Campaigns can add sub-types under any concrete or abstract
parent type. A nautical campaign might add `vessel` under
`item`. A superhero campaign might add `power_source` under
`artifact`. Custom sub-types inherit all constraints defined
on their parent type.

The hardcoded entity type CHECK constraint in the `entities`
table and the Go enum in `models.go` must be replaced with
a dynamic lookup against the campaign's entity type hierarchy.

## Relationship Types

The ontology defines approximately 80 relationship types
across 11 narrative categories plus one meta-game category.
All types are defined in
`schemas/ontology/relationship-types.yaml` and seeded into
campaign-scoped `relationship_types` rows on creation.

### Meta-Game Relationships

These connect real-world participants to the fiction:

| Forward           | Inverse             | Sym | Domain      | Range     |
|-------------------|---------------------|-----|-------------|-----------|
| plays             | played_by           | no  | player      | pc        |
| formerly_played   | formerly_played_by  | no  | player      | pc, npc   |
| narrates          | narrated_by         | no  | game_master | any       |

The `plays` / `played_by` relationship is the most important
edge in the graph. It bridges the real world and the fiction.
When a PC dies, `plays` becomes `formerly_played`. When a
player takes over an NPC, the NPC transitions to `pc` and a
new `plays` edge is created.

### Kinship and Family (Universal)

| Forward       | Inverse         | Sym | Domain        | Range         |
|---------------|-----------------|-----|---------------|---------------|
| parent_of     | child_of        | no  | npc, creature | npc, creature |
| sibling_of    | sibling_of      | yes | npc           | npc           |
| spouse_of     | spouse_of       | yes | npc           | npc           |
| betrothed_to  | betrothed_to    | yes | npc           | npc           |
| ancestor_of   | descendant_of   | no  | npc           | npc           |

### Social and Interpersonal (Universal)

| Forward       | Inverse         | Sym | Domain | Range        |
|---------------|-----------------|-----|--------|--------------|
| knows         | knows           | yes | npc    | npc          |
| friend_of     | friend_of       | yes | npc    | npc          |
| rival_of      | rival_of        | yes | agent  | agent        |
| lover_of      | lover_of        | yes | npc    | npc          |
| mentors       | mentored_by     | no  | npc    | npc          |
| protects      | protected_by    | no  | agent  | agent, place |
| trusts        | trusted_by      | no  | npc    | npc          |
| betrayed      | betrayed_by     | no  | npc    | npc          |

### Power and Authority (Universal)

| Forward         | Inverse           | Sym | Domain         | Range              |
|-----------------|-------------------|-----|----------------|--------------------|
| rules           | ruled_by          | no  | npc, faction   | place, faction,    |
|                 |                   |     |                | organization       |
| commands        | commanded_by      | no  | npc            | npc, agent         |
| employs         | employed_by       | no  | agent          | npc                |
| reports_to      | manages           | no  | npc            | npc                |
| serves          | served_by         | no  | npc            | agent              |
| patron_of       | client_of         | no  | npc, org       | npc, org           |
| vassal_of       | liege_of          | no  | npc, faction   | npc, faction       |
| imprisons       | imprisoned_by     | no  | agent, place   | npc, creature      |

### Spatial and Territorial (Universal)

| Forward          | Inverse          | Sym | Domain                 | Range    |
|------------------|------------------|-----|------------------------|----------|
| located_at       | contains         | no  | agent, artifact, event | place    |
| headquartered_at | headquarters_of  | no  | faction, org           | place    |
| part_of          | has_part         | no  | any                    | any      |
| borders          | borders          | yes | place                  | place    |
| connected_to     | connected_to     | yes | place                  | place    |
| originated_from  | origin_of        | no  | npc, creature, artifact| place    |
| guards           | guarded_by       | no  | agent                  | place,   |
|                  |                  |     |                        | artifact |
| haunts           | haunted_by       | no  | npc, creature          | place    |
| controls         | controlled_by    | no  | agent                  | place    |

The `part_of` / `has_part` relationship is critical for
location nesting. A cellar is `part_of` a house, which is
`part_of` a town, which is `part_of` a kingdom. This also
works for organisational structure (a division is `part_of`
an army) and documents (a chapter is `part_of` a book). The
domain and range are `any` because structural containment
applies broadly.

### Possession and Creation (Universal)

| Forward         | Inverse         | Sym | Domain          | Range           |
|-----------------|-----------------|-----|-----------------|-----------------|
| owns            | owned_by        | no  | agent           | artifact, place |
| created         | created_by      | no  | agent           | artifact, event |
| wields          | wielded_by      | no  | npc, creature   | item            |
| seeks           | sought_by       | no  | agent           | agent, artifact,|
|                 |                 |     |                 | place           |
| inherited_from  | bequeathed_to   | no  | npc             | npc, artifact   |

### Knowledge and Information (Universal)

| Forward      | Inverse        | Sym | Domain | Range    |
|--------------|----------------|-----|--------|----------|
| discovered   | discovered_by  | no  | npc    | any      |
| conceals     | concealed_by   | no  | agent  | any      |
| recorded_in  | records        | no  | any    | document |
| studies      | studied_by     | no  | npc    | any      |

### Conflict and Alliance (Universal)

| Forward            | Inverse            | Sym | Domain | Range |
|--------------------|--------------------|-----|--------|-------|
| enemy_of           | enemy_of           | yes | agent  | agent |
| allied_with        | allied_with        | yes | agent  | agent |
| at_war_with        | at_war_with        | yes | agent  | agent |
| conspires_against  | target_of          | no  | agent  | agent |
| hunts              | hunted_by          | no  | agent  | agent |
| opposes            | opposed_by         | no  | agent  | agent |

### Affiliation and Membership (Universal)

| Forward        | Inverse         | Sym | Domain | Range        |
|----------------|-----------------|-----|--------|--------------|
| member_of      | has_member      | no  | npc    | faction, org |
| founded        | founded_by      | no  | npc    | faction, org,|
|                |                 |     |        | place        |
| leads          | led_by          | no  | npc    | agent        |
| defected_from  | lost_member     | no  | npc    | faction, org |
| infiltrates    | infiltrated_by  | no  | npc    | faction, org |

### Supernatural and Mystical (Fantasy, Horror, Sci-Fi)

| Forward          | Inverse          | Sym | Domain        | Range              |
|------------------|------------------|-----|---------------|--------------------|
| bound_to         | bound_to         | yes | any           | any                |
| cursed_by        | cursed           | no  | agent         | any                |
| blessed_by       | blessed          | no  | agent         | any                |
| summoned         | summoned_by      | no  | npc, creature | npc, creature      |
| worships         | worshipped_by    | no  | npc           | npc, creature      |
| prophesied       | prophesied_by    | no  | npc, event    | npc, document      |
| transformed_by   | transformed      | no  | npc, creature | npc, creature,     |
|                  |                  |     |               | artifact           |
| corrupted_by     | corrupted        | no  | agent         | any                |

### Temporal and Causal (Universal)

| Forward          | Inverse          | Sym | Domain | Range |
|------------------|------------------|-----|--------|-------|
| caused           | caused_by        | no  | any    | event |
| triggered        | triggered_by     | no  | event  | event |
| participated_in  | involved         | no  | agent  | event |
| witnessed        | witnessed_by     | no  | npc    | event |

### Economic (Universal)

| Forward       | Inverse        | Sym | Domain | Range |
|---------------|----------------|-----|--------|-------|
| trades_with   | trades_with    | yes | agent  | agent |
| supplies      | supplied_by    | no  | agent  | agent |
| finances      | financed_by    | no  | agent  | agent |
| indebted_to   | creditor_of    | no  | agent  | agent |

### Event Relationships (Universal)

These capture significant narrative events between entities:

| Forward       | Inverse         | Sym | Domain | Range         |
|---------------|-----------------|-----|--------|---------------|
| murdered      | murdered_by     | no  | agent  | npc, creature |
| poisoned      | poisoned_by     | no  | agent  | npc, creature |
| wounded       | wounded_by      | no  | agent  | npc, creature |
| rescued       | rescued_by      | no  | agent  | agent         |
| captured      | captured_by     | no  | agent  | agent         |
| freed         | freed_by        | no  | agent  | agent         |
| deceived      | deceived_by     | no  | agent  | agent         |
| stole_from    | robbed_by       | no  | agent  | agent         |

### Genre-Specific Relationship Types

#### Sci-Fi and Cyberpunk

| Forward        | Inverse        | Sym | Domain        | Range              |
|----------------|----------------|-----|---------------|--------------------|
| uploaded_to    | hosts          | no  | agent         | artifact, place    |
| augmented_by   | augments       | no  | artifact      | npc, creature      |
| cloned_from    | clone_of       | no  | npc, creature | npc, creature      |
| programmed_by  | programmed     | no  | npc           | creature, artifact |
| hacked         | hacked_by      | no  | agent         | agent, artifact    |

#### Horror

| Forward       | Inverse           | Sym | Domain        | Range         |
|---------------|-------------------|-----|---------------|---------------|
| possessed_by  | possesses_spirit  | no  | npc, creature | npc, creature |
| infected_by   | infected          | no  | agent         | npc, creature |
| fears         | feared_by         | no  | npc           | any           |
| feeds_on      | preyed_upon_by    | no  | creature      | npc, creature |

#### Superhero

| Forward       | Inverse       | Sym | Domain | Range          |
|---------------|---------------|-----|--------|----------------|
| alter_ego_of  | alter_ego_of  | yes | npc    | npc            |
| empowered_by  | empowers      | no  | any    | npc, creature  |
| nemesis_of    | nemesis_of    | yes | agent  | agent          |

#### Historical and Political

| Forward          | Inverse          | Sym | Domain | Range              |
|------------------|------------------|-----|--------|--------------------|
| conquered        | conquered_by     | no  | agent  | agent, place       |
| exiled_from      | exiled           | no  | npc    | place, faction,    |
|                  |                  |     |        | organization       |
| succeeded        | predecessor_of   | no  | npc    | npc                |
| negotiated_with  | negotiated_with  | yes | agent  | agent              |

#### Romance and Intrigue

| Forward      | Inverse           | Sym | Domain | Range |
|--------------|-------------------|-----|--------|-------|
| courts       | courted_by        | no  | npc    | npc   |
| rejected     | rejected_by       | no  | npc    | npc   |
| disguised_as | impersonated_by   | no  | npc    | npc   |
| blackmails   | blackmailed_by    | no  | agent  | agent |

## Constraints

All constraints are advisory. Violations produce findings
for GM review and acknowledgement, never hard rejections.

### Domain and Range Constraints

Each relationship type specifies which entity type pairs are
valid. Constraints reference abstract parent types where
possible. When checking `employs` with domain `[agent]` and
range `[npc]`, the system resolves `agent` to its concrete
children (npc, creature, faction, organization) and validates
the source entity's type against that expanded set.

If no constraints are defined for a relationship type, any
entity type pair is valid. This preserves backward
compatibility with existing campaigns that predate the
ontology.

### Cardinality Constraints

Cardinality rules express how many relationships of a given
type an entity can have. The default is `many-to-many` (no
constraint). Campaigns can tighten cardinality where their
setting requires it.

Cardinality is campaign-scoped, not global. A campaign set
in a culture with monogamous marriage might set `spouse_of`
to `one-to-one`. A campaign featuring dynastic politics
leaves it as `many-to-many`. The global default imposes no
cardinality limits.

### Required Relationships

Required relationship rules express expectations about what
connections an entity should have. For example, every NPC
should have at least one `located_at` relationship. These
rules upgrade orphan detection from "this entity has zero
connections" to "this NPC is missing a location," which is
more actionable.

Required relationships produce advisory findings, not
blocking errors. An NPC without a location is flagged for
review but not prevented from being created.

### Constraint Acknowledgement

When a GM acknowledges a constraint violation, the
acknowledgement is recorded as a campaign-level override.
The system does not flag the same pattern again for that
campaign. This is how the campaign's ontology evolves: the
GM's creative decisions extend the effective ontology beyond
the global defaults.

## Temporal Model: Eras

Imagineer's knowledge graphs contain relationships that span
vastly different time scales, from events that happened
moments ago to mythic history from millennia past. The
system does not use earth-centric timestamps or calendars
because fictional worlds have their own time systems, and
multiverse settings may have branching or parallel timelines.

### Eras

An era is a named period in the fictional timeline. Eras are
ordered by a sequence number within their campaign and carry
no absolute duration. The system derives relative time
distance from era position.

```
Sequence  Name                       Scale
1         The Old Empire             mythic
2         The Interregnum            ancient
3         The Twelve Kingdoms        distant
4         Charles's Ascent           recent
5         Present Day                now
```

The scale values are: `mythic`, `ancient`, `distant`, `past`,
`recent`, `now`. These are metadata on the era, not on
individual relationships. When the campaign advances (a time
skip, a generational change, new PCs who are children of
previous PCs), a new era begins and everything before it
shifts contextually.

### Relationships and Events Tag Their Era

Each relationship and event entity references the era in
which it exists or occurred. Current-state relationships
belong to the latest era. When a relationship changes (an
alliance breaks, a ruler is deposed), the old relationship
is archived with its era and a new relationship is created
in the current era.

### Session Log is Separate

The era system tracks fictional time. What happened at the
table (which session, what the players discovered) is tracked
separately in the session log and entity logs. "Session 12:
the PCs discover the truth about Charles's lineage" is a
play event. The historical relationships it reveals belong
to their respective fictional eras.

### Era-Scoped Queries

The system can query relationships by era to answer questions
about historical state. "What was the political map during
the Interregnum?" queries relationships tagged with era 2.
"What changed during Charles's Ascent?" queries the
difference between eras 3 and 4.

## Storage Architecture

### YAML Source Files

The global ontology lives in `schemas/ontology/`:

```
schemas/ontology/
  entity-types.yaml     Type hierarchy with abstract parents
  relationship-types.yaml   All relationship type definitions
  constraints.yaml      Domain/range, cardinality, required
```

These files are the authoritative global definition. They
are version-controlled, diffable, and human-readable.

### Database Tables

Campaign-scoped enforcement uses these tables:

Existing tables (require modifications):

- `relationship_types` — already campaign-scoped. Seed from
  YAML instead of `relationship_type_templates` table.

- `relationship_type_constraints` — already exists with the
  correct schema. Seed with domain/range constraints from
  YAML.

New tables:

- `campaign_entity_types` — campaign-scoped entity type
  hierarchy. Columns: `campaign_id`, `name`, `parent_name`,
  `abstract`, `description`. Seeded from
  `entity-types.yaml`.

- `eras` — named periods in the fictional timeline. Columns:
  `campaign_id`, `sequence`, `name`, `scale`, `description`.

- `relationship_archive` — archived relationships with era
  reference. Same schema as `entity_relationships` plus
  `era_id` and `archived_at`.

- `cardinality_constraints` — campaign-scoped cardinality
  rules. Columns: `campaign_id`, `relationship_type_id`,
  `max_source`, `max_target`.

- `required_relationships` — campaign-scoped required
  relationship rules. Columns: `campaign_id`,
  `entity_type`, `relationship_type_name`.

- `constraint_overrides` — GM acknowledgements that override
  specific constraint violations. Columns: `campaign_id`,
  `constraint_type`, `override_key`, `acknowledged_at`.

### Entity Type Migration

The hardcoded CHECK constraint on the `entities` table and
the Go `EntityType` enum in `models.go` must be replaced
with a foreign key to `campaign_entity_types`. This allows
campaigns to define custom entity types without code changes.

## Enforcement Points

### Enrichment Pipeline

The enrichment agent's system prompt includes the campaign's
ontology: valid entity types, valid relationship types, and
domain/range constraints. This reduces invalid suggestions
at the source. The prompt references canonical forward type
names and lists which entity type pairs each type accepts.

### Graph Expert

The graph expert already implements structural checks
(orphan detection, type pair validation) and semantic checks
(redundancy, implied edges). With the ontology schema:

- Type pair validation (`ValidateTypePairs`) queries the now
  populated `relationship_type_constraints` table.
- A new cardinality check queries
  `cardinality_constraints`.
- A new required relationship check queries
  `required_relationships`.
- Constraint overrides are consulted before flagging
  previously acknowledged patterns.

### API Handlers

Entity creation and relationship creation handlers perform
advisory validation against the campaign's ontology. Invalid
entity types or relationship type pairs produce warnings in
the response but do not block the operation.

### User Interface

The graph health section in the Enrich phase displays
constraint violations for GM review. The GM can acknowledge
violations, which creates a `constraint_overrides` entry and
evolves the campaign's effective ontology.

## Synchronisation

The `.claude/graph-expert/` knowledge base documents the
ontology patterns and the Go implementation enforces them.
When the ontology schema changes, both must be updated:

- New relationship types in YAML require corresponding
  documentation in the knowledge base.
- New constraint types in the Go implementation require
  documentation of the enforcement logic.
- The agent definition at `.claude/agents/graph-expert.md`
  references both and includes a synchronisation protocol.

## Game System Compatibility

The ontology has been validated against all four supported
game systems (Call of Cthulhu 7e, GURPS 4e, Forged in the
Dark, D&D 5e 2024). Game-system-specific concepts map onto
generic ontology types:

- CoC tomes are `document` entities (sub-type `relic`).
  Reading creates a `studied` relationship. Sanity impact
  is a game mechanic tracked by the system schema.
- GURPS Patron, Enemy, Dependent, and Ally advantages map
  to `patron_of`, `enemy_of`, `protects`, and
  `allied_with`.
- FitD faction standings map to relationship tone and
  strength metadata. Turf maps to `controls`.
- D&D warlock pacts map to `bound_to` and `patron_of`.
  Attunement is a game mechanic, not a relationship.

The ontology models narrative relationships. Game system
schemas model mechanical effects. They are complementary.

## Summary of Changes

This design requires the following changes:

- Three new YAML files under `schemas/ontology/`.
- Six new database tables (campaign_entity_types, eras,
  relationship_archive, cardinality_constraints,
  required_relationships, constraint_overrides).
- Modifications to two existing tables
  (relationship_types seeding, relationship_type_constraints
  seeding).
- Migration of the hardcoded entity type enum to a dynamic
  campaign-scoped lookup.
- Updates to the enrichment pipeline prompt to include the
  campaign's ontology.
- Updates to the graph expert to check cardinality and
  required relationships.
- Updates to the graph expert knowledge base to reflect the
  new ontology structure.
- API and UI changes for constraint acknowledgement.
