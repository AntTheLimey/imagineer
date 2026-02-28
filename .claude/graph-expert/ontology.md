# Relationship Type Ontology

The ontology defines which relationship types are valid between
which entity type pairs. It is enforced at two levels: the
`relationship_type_constraints` table in the database, and the
`ValidateTypePairs` function in `internal/agents/graph/ontology.go`.

## Entity Types

The following entity types participate in relationships:

| Type           | Description                              |
|----------------|------------------------------------------|
| `npc`          | Non-player character                     |
| `location`     | Place or area                            |
| `item`         | Physical object or artifact              |
| `faction`      | Group with shared goals                  |
| `clue`         | Piece of evidence or information         |
| `creature`     | Monster or non-human entity              |
| `organization` | Formal institution                       |
| `event`        | Notable occurrence                       |
| `document`     | Written record or text                   |

The `other` type also exists in the schema but should not be
used for new entities.

## Relationship Type Templates

System-default types seeded from `relationship_type_templates`
(migration 002). Each new campaign copies these templates into
campaign-scoped `relationship_types` rows.

### Asymmetric Types

| Forward Name      | Inverse Name      | Description                   |
|-------------------|-------------------|-------------------------------|
| `owns`            | `owned_by`        | Possesses or controls         |
| `employs`         | `employed_by`     | Employs as worker or servant  |
| `works_for`       | `employs`         | Alias for employed_by         |
| `reports_to`      | `manages`         | Organisational hierarchy      |
| `parent_of`       | `child_of`        | Parental relationship         |
| `located_at`      | `contains`        | Physically located at         |
| `member_of`       | `has_member`      | Membership in group           |
| `created`         | `created_by`      | Made or produced              |
| `rules`           | `ruled_by`        | Political authority over      |
| `headquartered_at`| `headquarters_of` | Primary base of operations    |

### Symmetric Types

| Name          | Description                              |
|---------------|------------------------------------------|
| `knows`       | Acquaintance                             |
| `friend_of`   | Friendly relationship                    |
| `enemy_of`    | Hostile relationship                     |
| `allied_with` | Alliance or partnership                  |

Symmetric types have `is_symmetric = true` and `name` equals
`inverse_name`. The database constraint
`tpl_symmetric_inverse_match` enforces this.

## Type Pair Constraints

The `relationship_type_constraints` table defines which entity
type pairs are valid for each relationship type. The Go function
`ValidateTypePairs` (`ontology.go`) queries this table and flags
suggestions that violate the constraints.

### Constraint Behaviour

- **Constraints exist**: only the listed source-target type
  pairs are valid. All other combinations produce an
  `invalid_type_pair` detection.
- **No constraints**: the relationship type is unconstrained
  and any entity type pair is accepted.
- **Scope**: constraints are resolved per-campaign. The query
  joins through `relationship_types` and filters by
  `campaign_id` or `campaign_id IS NULL` (template types).

### Current Seed Constraints

The `relationship_type_constraints` table has no seed data in
the migrations. Constraints are populated when campaigns add
or customise relationship types. This is a known gap — common
constraints should be seeded to prevent nonsensical suggestions
out of the box.

### Recommended Default Constraints

These constraints reflect the natural ontology for TTRPG
campaign graphs. They should be added as seed data when the
constraint seeding work is done:

| Relationship Type   | Valid Source Types              | Valid Target Types             |
|---------------------|--------------------------------|--------------------------------|
| `owns`              | npc, faction, organization     | item, location, document       |
| `employs`           | faction, organization          | npc                            |
| `works_for`         | npc                            | faction, organization          |
| `reports_to`        | npc                            | npc                            |
| `parent_of`         | npc, creature                  | npc, creature                  |
| `located_at`        | npc, item, creature, event     | location                       |
| `member_of`         | npc                            | faction, organization          |
| `created`           | npc, organization              | item, document, event          |
| `rules`             | npc, faction                   | location, faction, organization|
| `headquartered_at`  | faction, organization          | location                       |
| `knows`             | npc                            | npc                            |
| `friend_of`         | npc                            | npc                            |
| `enemy_of`          | npc, faction, organization     | npc, faction, organization     |
| `allied_with`       | faction, organization          | faction, organization          |

These recommendations follow two principles:

1. **Semantic fit**: only pairs where the relationship makes
   narrative sense.
2. **Conservative scope**: start narrow and widen if campaigns
   need it. It is easier to relax a constraint than to
   retroactively clean up invalid edges.

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
