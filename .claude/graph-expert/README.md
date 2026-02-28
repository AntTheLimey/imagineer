# Graph Expert Knowledge Base

Domain knowledge for the graph expert sub-agent. This knowledge
base is the advisory counterpart to the Go implementation in
`internal/agents/graph/` — they represent the same expert and
must stay in sync.

## Documents

- [ontology.md](ontology.md) — Relationship type ontology, valid
  entity type pairs, and constraint definitions.
- [graph-patterns.md](graph-patterns.md) — Graph patterns,
  anti-patterns, and the dual-mode check architecture.
- [deduplication.md](deduplication.md) — Deduplication and
  redundancy detection strategies.

## Quick Reference

### Detection Types

| Detection Type          | Source     | Description                   |
|-------------------------|------------|-------------------------------|
| `orphan_warning`        | Structural | Entity has zero relationships |
| `invalid_type_pair`     | Structural | Type applied to wrong pair    |
| `cardinality_violation` | Structural | Exceeds max relationship count|
| `missing_required`      | Structural | Required relationship absent  |
| `redundant_edge`        | Semantic   | Duplicate or implied edge     |
| `graph_warning`         | Semantic   | Unclassified LLM finding      |

### Pipeline Position

| Property   | Value                                         |
|------------|-----------------------------------------------|
| Agent name | `graph-expert`                                |
| Depends on | `enrichment` (runs after enrichment agent)    |
| Input      | `PipelineInput` with entities, relationships, |
|            | prior results, and source content             |
| Output     | `[]ContentAnalysisItem` (advisory findings)   |

### Key Implementation Files

| File                                       | Purpose                |
|--------------------------------------------|------------------------|
| `internal/agents/graph/expert.go`          | Pipeline agent entry   |
| `internal/agents/graph/ontology.go`        | Type pair validation   |
| `internal/agents/graph/cardinality.go`     | Cardinality checks     |
| `internal/agents/graph/required.go`        | Required rel checks    |
| `internal/agents/graph/overrides.go`       | Override filtering     |
| `internal/agents/graph/prompts.go`         | LLM prompt construction|
| `internal/agents/graph/parser.go`          | LLM response parsing   |
| `internal/agents/graph/expert_test.go`     | Test suite             |
| `internal/agents/graph/cardinality_test.go`| Cardinality tests      |
| `internal/agents/graph/required_test.go`   | Required rel tests     |
| `internal/agents/graph/overrides_test.go`  | Override tests         |

### Ontology Source of Truth

The ontology is defined in YAML files under `schemas/ontology/`:

| File                    | Contents                          |
|-------------------------|-----------------------------------|
| `entity-types.yaml`    | Entity type hierarchy with        |
|                         | abstract parents                  |
| `relationship-types.yaml`| ~80 relationship types across   |
|                         | narrative categories              |
| `constraints.yaml`     | Domain/range constraints,         |
|                         | cardinality, required rels        |

These YAML files are the canonical definitions. Migrations
seed the database from these files per-campaign.

## Synchronisation

This knowledge base and the Go implementation in
`internal/agents/graph/` are two facets of the same expert.
When either changes, the other must be updated:

- **New detection type added in Go** → update detection types
  table above and relevant document.
- **New ontology rule added in YAML or migrations** → update
  `ontology.md` with the constraint.
- **New pattern discovered during advisory work** → update
  `graph-patterns.md` and ensure the Go implementation can
  detect it (or note it as a gap).
- **LLM prompt changed** → update `deduplication.md` if the
  finding types or rules changed.

Last Updated: 2026-02-28
