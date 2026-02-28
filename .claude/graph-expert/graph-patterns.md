# Graph Patterns and Anti-Patterns

Patterns and anti-patterns for TTRPG campaign knowledge graphs,
aligned with the checks implemented in
`internal/agents/graph/expert.go`.

## Dual-Mode Architecture

The graph expert runs two categories of checks in sequence:

### 1. Structural Checks (No LLM)

These are deterministic and run regardless of LLM availability.

**1a. Orphaned Entity Detection** (`CheckOrphanedEntities`):
- Compares the entity list against the relationship list.
- An entity that appears as neither source nor target in any
  relationship is flagged as `orphan_warning`.
- Input: `PipelineInput.Entities` and
  `PipelineInput.Relationships`.
- Note: `Relationships` is populated from the database before
  pipeline invocation, not just from the current enrichment
  run. This prevents false-positive orphan warnings for
  entities that have existing connections.

**1b. Type Pair Validation** (`ValidateTypePairs`):
- Filters `PipelineInput.PriorResults` for items with
  `DetectionType == "relationship_suggestion"`.
- Queries `relationship_type_constraints` for the campaign.
- Flags suggestions that pair incompatible entity types as
  `invalid_type_pair`.
- If no constraints exist for a relationship type, the
  suggestion is considered valid (constraints are optional).

**1c. Cardinality Check** (`CheckCardinality`):
- Queries `cardinality_constraints` for the campaign.
- Counts existing relationships plus proposed suggestions
  per (entity, relationship type, direction).
- Flags entities that would exceed `max_source` or
  `max_target` as `cardinality_violation`.
- Returns nil early when no constraints exist for the
  campaign (all types default to many-to-many).

**1d. Required Relationships** (`CheckRequiredRelationships`):
- Queries `required_relationships` for the campaign.
- For each entity of a constrained type, verifies it has
  at least one relationship of each required type.
- Flags missing relationships as `missing_required`.
- Skips rules whose relationship type is not found in the
  campaign to avoid false positives.

### 2. Semantic Checks (LLM Required)

These use the LLM to detect issues that require understanding
of meaning, not just structure.

**Redundant Edge Detection**:
- Two edges between the same entity pair with different type
  names that convey the same meaning.
- Example: "works_for" and "employed_by" between the same NPC
  and organisation.
- Detection type: `redundant_edge`.

**Implied Edge Detection**:
- A relationship discoverable by traversing intermediate
  entities.
- Example: if NPC Alice leads Faction X and NPC Bob belongs
  to Faction X, a direct "associated_with" edge between Alice
  and Bob is unnecessary.
- Detection type: `redundant_edge` (mapped from `implied_edge`
  by `findingTypeToDetectionType` in `parser.go`).

### 3. Post-Processing: Override Filtering

After both structural and semantic checks complete,
`FilterOverriddenFindings` (`overrides.go`) removes findings
that match existing GM-acknowledged overrides in the
`constraint_overrides` table.

Overridable detection types:
- `invalid_type_pair` (constraint type: `domain_range`)
- `cardinality_violation` (constraint type: `cardinality`)
- `missing_required` (constraint type: `required`)

Non-overridable types (`orphan_warning`, `redundant_edge`,
`graph_warning`) always pass through.

On error querying overrides, the finding is kept to avoid
silently dropping valid warnings.

### Graceful Degradation

If structural checks succeed but the LLM call fails, the
structural findings are still returned. The LLM error is logged
but does not propagate as an error from `Run()`. This ensures
orphan warnings, type pair violations, cardinality violations,
and missing required relationships are never lost due to
transient LLM failures.

## Good Patterns

### Meaningful Direct Edges

Create direct edges when the relationship carries independent
narrative meaning that is not discoverable through traversal:

```
Viktor (npc) --[employs]--> Maria (npc)
Viktor (npc) --[located_at]--> Silver Fox Inn (location)
The Order (faction) --[headquartered_at]--> Old Chapel (location)
```

Each edge tells the GM something specific. Removing any one
of them would lose information.

### Specific Relationship Types

Use the most specific type available:

```
Good:  Viktor --[employs]--> Maria
Bad:   Viktor --[associated_with]--> Maria
```

### Appropriate Metadata

Use tone and strength to add narrative depth without creating
extra edges:

```
Viktor --[enemy_of {tone: hostile, strength: 8}]--> Karl
```

One edge with metadata, not separate edges for the rivalry and
the hostility.

## Anti-Patterns

### Hub Entity Overload

When a single entity connects to a large percentage of all
other entities, it functions as a noise hub. Common culprits:

- A capital city with "visited_by" edges to every NPC.
- A major faction with "associated_with" edges to every entity.

Hub connections should carry meaningful narrative weight. If
most NPCs simply exist in a city, their location is context,
not a relationship worth storing per-entity.

### Generic Type Proliferation

Overusing "associated_with", "related_to", or "connected_to"
adds edges without adding information. These generic types
make the graph denser without making it more useful.

If the LLM enrichment agent suggests a generic type, the graph
expert should flag it for review or suggest a more specific
alternative.

### Redundant Symmetric Pairs

Two edges between the same pair using related types:

```
Anti-pattern:
  Viktor --[works_for]--> The Guild
  The Guild --[employs]--> Viktor

Correct:
  Viktor --[works_for]--> The Guild
  (inverse "employs" derived automatically by the view)
```

The single-edge model and database trigger prevent this at the
storage level, but the LLM may suggest edges that would create
logical redundancy even if the trigger blocks the physical
duplicate.

### Implied Traversal Edges

Edges that duplicate information already reachable by graph
traversal:

```
Anti-pattern:
  Alice --[member_of]--> Faction X
  Bob --[member_of]--> Faction X
  Alice --[associated_with]--> Bob  (implied by shared faction)

Correct:
  Alice --[member_of]--> Faction X
  Bob --[member_of]--> Faction X
  (connection between Alice and Bob discovered via Faction X)
```

The LLM semantic check specifically targets this anti-pattern.

### Orphaned Entities

Entities with zero relationships indicate either:

1. **Incomplete data**: the entity was extracted but its
   relationships were not yet identified. Common during
   initial content processing.
2. **Irrelevant entity**: the entity was extracted in error
   or is not meaningful to the campaign.

The graph expert flags these for human review — it does not
auto-delete or auto-connect them.

## LLM Prompt Design

The semantic check prompt (`prompts.go`) follows these rules:

1. **Conservative**: only flag genuine issues. New relationships
   that add independent meaning are not redundant.
2. **Direction-aware**: A→B and B→A via an inverse type are the
   same relationship stored once, not a redundancy.
3. **Empty when clean**: return an empty findings array if no
   issues are found.

The prompt includes three sections built from `PipelineInput`:

- **Existing Relationships**: current edges in the campaign
  graph, formatted as `Source (type) --[label]--> Target (type)`.
- **Proposed New Relationships**: suggestions from the
  enrichment agent, with descriptions truncated to 100
  characters (rune-based via `agents.TruncateString`).
- **Campaign Entities**: full entity list with types and IDs
  for context.

## LLM Response Parsing

The parser (`parser.go`) handles:

- Markdown code fence stripping (via `agents.StripCodeFences`).
- JSON unmarshalling into `graphResponse`.
- Finding type validation: only `redundant_edge` and
  `implied_edge` are accepted. Unknown types are normalised
  to `redundant_edge` with a log warning.
- Empty descriptions are skipped.
- Nil `involvedEntities` arrays are normalised to empty slices.

### Detection Type Mapping

**Structural checks** (no mapping needed — detection type is
set directly by the Go function):

| Check Function               | Detection Type          |
|------------------------------|-------------------------|
| `CheckOrphanedEntities`      | `orphan_warning`        |
| `ValidateTypePairs`          | `invalid_type_pair`     |
| `CheckCardinality`           | `cardinality_violation` |
| `CheckRequiredRelationships` | `missing_required`      |

**LLM semantic checks** (mapped in `parser.go`):

| LLM Finding Type | Detection Type   |
|-------------------|-----------------|
| `redundant_edge`  | `redundant_edge` |
| `implied_edge`    | `redundant_edge` |
| (unknown)         | `graph_warning`  |
