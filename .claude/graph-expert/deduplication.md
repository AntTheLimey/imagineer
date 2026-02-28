# Deduplication Strategies

Strategies for detecting and handling duplicate, redundant,
and implied relationships in Imagineer's campaign knowledge
graphs.

## Layers of Deduplication

The system prevents duplicate relationships at multiple levels:

### 1. Database Layer (Automatic)

These protections operate without the graph expert:

- **Unique constraint**: `(campaign_id, source_entity_id,
  target_entity_id, relationship_type_id)` prevents exact
  duplicate rows.
- **Inverse trigger**: if a relationship from A to B exists,
  the trigger blocks creation of the same relationship from
  B to A.
- **Upsert semantics**: `ON CONFLICT ... DO UPDATE` updates
  metadata (tone, strength, description) rather than failing
  on duplicates.

### 2. Structural Layer (Graph Expert, No LLM)

Type pair validation catches relationships that should not
exist based on entity type constraints. See `ontology.md` for
the constraint definitions.

An `invalid_type_pair` detection often indicates the enrichment
agent confused entity types or used an inappropriate
relationship type. The resolution is typically to either:

- Change the relationship type to one valid for the pair.
- Reject the suggestion entirely.

### 3. Semantic Layer (Graph Expert, With LLM)

The LLM identifies two categories of semantic duplication:

**Redundant edges**: two edges between the same entity pair
that mean the same thing through different type names.

```
Duplicate:
  Viktor --[works_for]--> The Guild
  Viktor --[employed_by]--> The Guild

Resolution: keep the more specific type (works_for) and
remove the other.
```

**Implied edges**: edges that are discoverable by traversing
the existing graph.

```
Implied:
  Alice --[leads]--> Faction X
  Bob --[member_of]--> Faction X
  Alice --[associated_with]--> Bob  ← implied by shared faction

Resolution: remove the direct edge between Alice and Bob.
The connection is discoverable via Faction X.
```

### 4. Constraint Override Layer (Post-Processing)

After all structural and semantic checks produce findings,
`FilterOverriddenFindings` (`overrides.go`) removes findings
that the GM has previously acknowledged via the
`constraint_overrides` table.

This layer operates on three overridable detection types:

| Detection Type          | Override Key Format            |
|-------------------------|--------------------------------|
| `invalid_type_pair`     | `relType:srcType:tgtType`      |
| `cardinality_violation` | `relType:entityId:direction`   |
| `missing_required`      | `entityType:relType`           |

Non-overridable types (`orphan_warning`, `redundant_edge`,
`graph_warning`) always pass through to the GM.

This prevents the system from repeatedly flagging violations
that the GM has reviewed and accepted as intentional
exceptions to the ontology rules.

### 5. Era-Based Archiving

When a relationship changes or ends, it is moved to the
`relationship_archive` table rather than being deleted. The
archive preserves era context, allowing the graph to show
both the current state and its historical evolution.

This is not deduplication per se, but it prevents stale
relationships from accumulating in the active graph while
retaining the historical record for timeline queries.

## Detection Heuristics

### Same-Pair Different-Type Redundancy

When two edges connect the same entity pair (in either
direction) with different relationship type names, check
whether the types are semantically equivalent:

- `works_for` and `employed_by` — same meaning, different
  direction framing.
- `owns` and `possesses` — synonymous.
- `leads` and `commands` — near-synonymous in most TTRPG
  contexts.

The LLM evaluates semantic equivalence because a simple
string comparison would miss these cases.

### Traversal Redundancy

An edge is redundant if the connection it represents is
discoverable by traversing a short path (typically 2 hops)
through intermediate entities:

```
A --[located_at]--> London
B --[located_at]--> London
A --[associated_with]--> B  ← redundant (shared location)
```

Not all shared-node connections are redundant. Two NPCs who
are both members of the same faction may still have an
independent relationship worth recording:

```
Alice --[member_of]--> Faction X
Bob --[member_of]--> Faction X
Alice --[enemy_of]--> Bob  ← NOT redundant (independent meaning)
```

The key test is: **does the direct edge carry meaning that
the traversal does not?** "enemy_of" tells the GM something
that "both are in Faction X" does not. "associated_with" does
not.

### Generic Type as Redundancy Signal

When the enrichment agent suggests a generic type
(`associated_with`, `related_to`, `connected_to`) between
entities that already share a path, the generic edge is
almost certainly redundant. Generic types between entities
with no existing path are suspicious but not necessarily
wrong — they may indicate a relationship the enrichment agent
could not classify more specifically.

## Resolution Strategies

### For Redundant Edges

1. Keep the more specific relationship type.
2. If both types are equally specific, keep the one that
   matches the seed templates (canonical types).
3. Update metadata (tone, strength) if the duplicate carried
   useful metadata not present on the surviving edge.

### For Implied Edges

1. Remove the direct edge.
2. Verify the traversal path still exists (the intermediate
   entities and their relationships are intact).
3. If the direct edge carried metadata (tone, strength) that
   adds narrative value beyond the traversal, consider keeping
   it and noting the exception.

### For Invalid Type Pairs

1. Check whether a valid alternative type exists for the
   entity pair.
2. If yes, suggest changing the relationship type.
3. If no valid type exists, the relationship may indicate a
   gap in the ontology. Flag for human review and suggest
   adding a new constraint if the pair is narratively valid.

## Conservative Approach

The graph expert follows a conservative philosophy:

- **When in doubt, do not flag**. False positives erode trust
  in the system.
- **Advisory only**. All findings require human review before
  action. The expert never auto-removes edges.
- **Structural over semantic**. Structural checks (orphans,
  type pairs) are deterministic and reliable. Semantic checks
  (redundancy, implied edges) are probabilistic and should be
  weighted accordingly.
- **LLM temperature**: semantic checks use temperature 0.2 for
  consistency. Higher temperatures increase false positive
  rates for graph analysis tasks.
