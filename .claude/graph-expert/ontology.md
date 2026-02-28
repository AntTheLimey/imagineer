# Relationship Type Ontology

The ontology defines entity types, relationship types, and the
constraints that govern which combinations are valid. It is the
source of truth for graph validation at two levels:

1. **YAML definitions** in `schemas/ontology/` — canonical
   specifications loaded per-campaign.
2. **Go validation** in `internal/agents/graph/` — runtime
   checks against the seeded database tables.

## Ontology YAML Files

| File                             | Purpose                     |
|----------------------------------|-----------------------------|
| `schemas/ontology/entity-types.yaml` | Entity type hierarchy   |
| `schemas/ontology/relationship-types.yaml` | ~80 rel types    |
| `schemas/ontology/constraints.yaml` | Domain/range, cardinality,|
|                                  | required relationships      |

When a campaign is created, the loader reads these YAML files
and seeds the campaign-scoped database tables. Campaigns can
then customise their own constraints.

## Entity Types

Entity types form a hierarchy with abstract parent types that
group concrete children for constraint inheritance. Abstract
types cannot be assigned to entities directly but can appear
in domain/range constraints — the loader resolves them to
their concrete descendants.

### Top-Level Abstract Types

| Abstract Type | Concrete Children                      |
|---------------|----------------------------------------|
| `person`      | player, game_master, character         |
| `character`   | pc, npc                                |
| `agent`       | character, creature, faction,          |
|               | organization                           |
| `place`       | location                               |
| `artifact`    | item, document                         |
| `narrative`   | event, clue                            |

### Concrete Types with Sub-Types

Several concrete types have their own sub-type children:

| Parent        | Sub-Types                              |
|---------------|----------------------------------------|
| `creature`    | beast, undead, construct, spirit,      |
|               | deity, aberration                      |
| `organization`| government, corporation, cult, guild,  |
|               | military, criminal                     |
| `item`        | weapon, armor, vehicle, treasure,      |
|               | relic, tool, consumable                |
| `document`    | spell, map, letter, prophecy,          |
|               | contract, journal                      |
| `event`       | battle, ritual, disaster, discovery,   |
|               | betrayal_event, celebration            |

The `other` type also exists in the schema but should not be
used for new entities.

### Constraint Inheritance

When a constraint references an abstract type like `agent`,
it applies to all concrete descendants: npc, pc, creature,
faction, organization, and their sub-types. This reduces
duplication in constraint definitions.

## Relationship Types

Approximately 80 relationship types are defined in
`schemas/ontology/relationship-types.yaml`, organised by
narrative category. Each campaign copies these templates
into campaign-scoped `relationship_types` rows.

### Categories

| Category                   | Example Types                   |
|----------------------------|---------------------------------|
| Meta-Game                  | plays, formerly_played, narrates|
| Kinship and Family         | parent_of, sibling_of, spouse_of|
| Social and Interpersonal   | knows, friend_of, rival_of,    |
|                            | mentors, trusts, betrayed       |
| Power and Authority        | rules, employs, commands,       |
|                            | serves, vassal_of, imprisons    |
| Spatial and Territorial    | located_at, headquartered_at,   |
|                            | part_of, borders, haunts        |
| Possession and Creation    | owns, created, wields, seeks    |
| Knowledge and Information  | discovered, conceals,           |
|                            | recorded_in, studies            |
| Conflict and Alliance      | enemy_of, allied_with,          |
|                            | at_war_with, conspires_against  |
| Affiliation and Membership | member_of, founded, leads,      |
|                            | defected_from, infiltrates      |
| Supernatural and Mystical  | bound_to, cursed_by, summoned,  |
|                            | worships, corrupted_by          |
| Temporal and Causal        | caused, triggered,              |
|                            | participated_in, witnessed      |
| Economic                   | trades_with, supplies,          |
|                            | finances, indebted_to           |
| Event Relationships        | murdered, poisoned, wounded,    |
|                            | rescued, captured, deceived     |
| Sci-Fi and Cyberpunk       | uploaded_to, augmented_by,      |
|                            | cloned_from, hacked             |
| Horror                     | possessed_by, infected_by,      |
|                            | fears, feeds_on                 |
| Superhero                  | alter_ego_of, empowered_by,     |
|                            | nemesis_of                      |
| Historical and Political   | conquered, exiled_from,         |
|                            | succeeded, negotiated_with      |
| Romance and Intrigue       | courts, rejected, disguised_as, |
|                            | blackmails                      |

### Genre Tags

Each relationship type carries genre tags (e.g. `universal`,
`fantasy`, `horror`, `scifi`, `superhero`, `historical`,
`romance`). The enrichment prompt uses these to filter
suggestions appropriate to the campaign's genre.

### Symmetric Types

Symmetric types have `name == inverse_name` and
`is_symmetric = true`. The database constraint
`tpl_symmetric_inverse_match` enforces this. Examples:
`knows`, `sibling_of`, `enemy_of`, `allied_with`.

## Domain/Range Constraints

The `constraints.yaml` file defines which entity type pairs
are valid for each relationship type using domain (source)
and range (target) lists. These are loaded into the
`relationship_type_constraints` table per-campaign.

### Constraint Behaviour

- **Constraints exist**: only the listed source-target type
  pairs are valid. All other combinations produce an
  `invalid_type_pair` detection.
- **No constraints**: the relationship type is unconstrained
  and any entity type pair is accepted.
- **Abstract type resolution**: constraints referencing
  `agent`, `artifact`, or other abstract types are expanded
  to their concrete descendants during loading.
- **Scope**: constraints are resolved per-campaign. The
  query joins through `relationship_types` and filters by
  `campaign_id` or `campaign_id IS NULL` (template types).

### Validation

The Go function `ValidateTypePairs` (`ontology.go`) queries
the `relationship_type_constraints` table and flags
suggestions that violate the domain/range constraints.

## Cardinality Constraints

The `cardinality_constraints` table limits how many
relationships of a given type an entity can have in each
direction (source or target). By default all types are
many-to-many; campaigns can tighten these limits.

### Constraint Fields

| Column       | Meaning                                 |
|--------------|-----------------------------------------|
| `max_source` | Max times an entity can be the source   |
|              | of this relationship type (NULL = no    |
|              | limit)                                  |
| `max_target` | Max times an entity can be the target   |
|              | of this relationship type (NULL = no    |
|              | limit)                                  |

### Validation

The Go function `CheckCardinality` (`cardinality.go`)
counts existing relationships plus proposed suggestions
and flags any entity that would exceed the limit as a
`cardinality_violation`.

## Required Relationships

The `required_relationships` table declares that every
entity of a given type must participate in at least one
relationship of a specified type. Default rules from the
YAML:

| Entity Type    | Required Relationship Type       |
|----------------|----------------------------------|
| `npc`          | `located_at`                     |
| `pc`           | `played_by`, `located_at`        |
| `creature`     | `located_at`                     |
| `faction`      | `headquartered_at`               |
| `organization` | `headquartered_at`               |

### Validation

The Go function `CheckRequiredRelationships`
(`required.go`) checks every entity against applicable
rules and flags missing relationships as
`missing_required`.

## Constraint Overrides

GMs can acknowledge constraint violations they consider
acceptable. The `constraint_overrides` table stores these
acknowledgements keyed by constraint type and override key.

### Overridable Detection Types

| Detection Type          | Constraint Type | Key Format             |
|-------------------------|-----------------|------------------------|
| `invalid_type_pair`     | `domain_range`  | `relType:srcType:tgtType`|
| `cardinality_violation` | `cardinality`   | `relType:entityId:dir` |
| `missing_required`      | `required`      | `entityType:relType`   |

Non-overridable types (`orphan_warning`, `redundant_edge`,
`graph_warning`) always pass through to the GM.

The Go function `FilterOverriddenFindings` (`overrides.go`)
runs as a post-processing step after all checks, removing
findings that match existing overrides.

## Ontology Design Principles

### Direct Edges vs Traversals

Not every connection needs a direct edge. If NPC Alice is in
London and NPC Bob is in London, the system should not create
an "associated_with" edge between them. The connection is
discoverable by traversing their shared location. Reserve
direct edges for relationships that carry independent meaning.

### Specificity Over Generality

Prefer specific relationship types over generic ones. "employs"
conveys more meaning than "associated_with". Generic types like
"related_to" add edges without adding information.

### Single-Edge Storage

The system stores one row per logical relationship. The
`entity_relationships_view` presents both forward and inverse
perspectives. A database trigger prevents creation of inverse
edges. `ValidateTypePairs` works with forward-direction
constraints only — the inverse is implicit.

## Era-Based Temporal Model

Relationships can reference eras via the `eras` table. When
a relationship changes or ends, it is moved to the
`relationship_archive` table, preserving era context. This
allows the graph to represent both current state and
historical progression.

### Database Tables

| Table                  | Purpose                          |
|------------------------|----------------------------------|
| `eras`                 | Named time periods per-campaign  |
| `relationship_archive` | Historical relationships with    |
|                        | era references                   |

## Enrichment Integration

The enrichment agent's system prompt now includes valid
entity types and relationship types from the ontology YAML
files. This guides the LLM to suggest types that conform to
the schema, reducing invalid_type_pair findings downstream.
