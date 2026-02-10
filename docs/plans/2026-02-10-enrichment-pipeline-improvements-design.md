<!--
  Imagineer - TTRPG Campaign Intelligence Platform

  Copyright (c) 2025 - 2026
  This software is released under The MIT License
-->

# Enrichment Pipeline Improvements Design

This design document describes improvements to the LLM enrichment
pipeline that address relationship type proliferation, repeated
suggestions, token waste, low-value output, and monolithic prompt
structure. The improvements ship as six independent, incremental
changes.

## Context and Motivation

The current enrichment pipeline has several problems identified
during testing.

The LLM creates custom relationship types freely during enrichment
instead of using the canonical template types. This behaviour
results in semantically duplicative types (for example,
`located_at`, `located_in`, `operates_in`, and
`base_of_operations_for` all mean roughly "entity is at a
location"). In one campaign with 15 entities, the LLM created 23
custom types beyond the 14 templates.

Every enrichment run treats each entity as if the system has never
analysed the entity before. The same description expansion gets
suggested repeatedly, even after the user has already accepted the
suggestion.

The system sends full content to the LLM on every run, even when
nothing has changed since the last enrichment. This wastes tokens
and increases cost without producing new insights.

After two or three runs, the LLM produces diminishing-returns
tweaks rather than substantive improvements. Log entry suggestions
become duplicative.

A single large LLM prompt asks for description updates, log
entries, relationships, and new entities simultaneously. This
monolithic approach produces lower-quality results than focused
prompts would produce.

## Design: Parallel Sub-Agent Architecture

The system replaces the current single-prompt enrichment with
parallel focused sub-agents. Each sub-agent has a narrower scope
and a simpler prompt, producing higher-quality results.

### Sub-Agents

The system runs these sub-agents in parallel per entity:

- The Description Agent analyses content for information that
  expands or corrects the entity description. The agent only
  suggests updates when substantive new information is found.
- The Log Entry Agent identifies chronological events involving the
  entity. The agent focuses on actionable, in-game events rather
  than restatements of existing information.
- The Relationship Agent identifies relationships between the
  entity and other entities mentioned in the content. The agent
  suggests relationship types as strings and constrains suggestions
  to canonical forward types from the campaign's
  `relationship_types` table when possible. The agent proposes new
  types only when truly needed.
- The New Entity Agent scans content for entity mentions not yet in
  the campaign database. The agent suggests creation with type,
  name, and initial description.

### Benefits

The parallel sub-agent architecture provides the following
advantages:

- Each agent's prompt is smaller and more focused, producing
  better results.
- Independent agents run in parallel, reducing wall-clock time.
- If one agent fails, the others still succeed.
- Each agent's prompt can be tuned independently without affecting
  the other agents.
- Total token cost increases slightly (the system sends content
  four times instead of once per entity) but result quality
  improves and each response is more reliable to parse.

## Design: Content-Diff Enrichment

The system only sends changed content to the LLM, not the full
content on every run.

### Mechanism

The system stores the content hash per
`(source_table, source_id)` at enrichment time. On subsequent
runs, the system follows this logic:

- The system computes the diff between the current content and the
  last-enriched content.
- If the diff is empty, the system skips enrichment entirely (zero
  tokens, zero cost).
- If the diff is small, the system sends only the changed portions
  to the LLM.
- If the content is entirely new (no prior enrichment), the system
  sends the full content.

### Storage

A single row per content source is sufficient. The following
statement creates the `enrichment_history` table:

```sql
CREATE TABLE enrichment_history (
    id               BIGSERIAL PRIMARY KEY,
    campaign_id      BIGINT NOT NULL
                     REFERENCES campaigns(id)
                         ON DELETE CASCADE,
    source_table     TEXT NOT NULL,
    source_id        BIGINT NOT NULL,
    content_hash     TEXT NOT NULL,
    enriched_at      TIMESTAMPTZ DEFAULT NOW(),
    enrichment_count INT NOT NULL DEFAULT 1,
    UNIQUE (campaign_id, source_table, source_id)
);
```

### Benefits

Content-diff enrichment provides the following advantages:

- The system naturally prevents "suggesting the same description
  update again" since the LLM only sees new information.
- The GM can freely edit content, and the system diffs against the
  last-enriched version.
- No complex tracking of which suggestions were made or accepted
  is required.
- Cost is proportional to the size of the change: small edits cost
  small tokens.

## Design: Enrichment Budget

The system caps the total enrichment runs per content source.
Each content type has a configurable budget.

### Budget Levels

Each content type has a budget representing the number of
enrichment passes the system allows:

- Campaign overview has a configurable default of 3.
- Chapter has a configurable default of 5.
- Session setup has a configurable default of 3.
- Session wrap-up has a configurable default of 3.
- Session play is unconstrained because the chat-based format
  means anything could come up.

### Storage

The `enrichment_history.enrichment_count` column tracks how many
times the system has enriched a content source. When the count
reaches the budget, the system skips enrichment unless the user
explicitly requests the enrichment.

### Configuration

The system stores budgets in a system settings table or user
settings. Budgets support the following modes:

- The user can disable enrichment entirely so that enrichment
  never auto-runs.
- The user can set a content type budget to 0 to skip that type.
- The user can set a high number for unlimited effective budget.

## Design: Relationship Type Governance

The LLM must not freely create relationship types during
enrichment. Type creation requires GM approval.

### Enrichment Behaviour

The enrichment relationship agent suggests relationship types as
strings. When the suggested type does not exist in the campaign's
`relationship_types` table, the system follows this process:

1. The system flags the suggestion as "new type needed" in the
   analysis item.
2. On the triage page, new type suggestions appear before
   relationship suggestions (grouped at the top).
3. For each new type suggestion, the UI shows the suggested type
   name and description, semantically similar existing types (via
   fuzzy match against existing type names and display labels),
   and three options: Accept (creates the type), Map to existing
   type (select from a dropdown), or Dismiss.
4. If the user maps to an existing type, the system updates all
   relationship suggestions using the dismissed type to use the
   selected alternative.

### Triage Page Ordering

The triage page presents enrichment results in this order:

1. New relationship type suggestions appear first (the user must
   resolve these before proceeding).
2. Relationship suggestions appear second (these depend on types
   being resolved).
3. Description update suggestions appear third.
4. Log entry suggestions appear fourth.
5. New entity suggestions appear fifth.

The user can review items in groups 2 through 5 in any order.
The user must resolve group 1 before group 2 becomes actionable.

### Relationship Type Management

The system adds a campaign settings area for managing relationship
types. The management interface provides the following
capabilities:

- The interface lists all types for the campaign with name,
  inverse name, display labels, and symmetric flag.
- The user can create new types manually.
- The user can edit existing types (rename, change inverse, toggle
  symmetric).
- The user can delete unused types (with confirmation if
  relationships exist using the type).
- The interface displays the count of relationships using each
  type.

The relationship type management interface is accessible from the
campaign settings or a dedicated "Relationship Types" section in
the campaign navigation.

## One-Time Cleanup

The existing campaign data has 23 LLM-created relationship types
that require review. A migration or admin tool should perform the
following tasks:

1. The tool lists all non-template relationship types per
   campaign.
2. The tool suggests mappings to canonical types based on semantic
   similarity.
3. The GM confirms or overrides each mapping.
4. The tool merges the selected types (updates relationship rows
   and deletes the old type).

This cleanup works best as an interactive tool rather than an
automated migration, since the GM should decide which mappings are
correct.

## Implementation Order

The system implements these improvements in the following order:

1. Content-diff enrichment provides the biggest token savings and
   is the simplest to implement.
2. The enrichment budget layers on top of the diff mechanism.
3. Parallel sub-agents provide a quality improvement.
4. Relationship type governance and triage ordering add approval
   controls.
5. The relationship type management UI gives the GM direct control
   over types.
6. The one-time cleanup tool addresses existing data quality.

Each item is independently valuable and the team can ship each
item incrementally.

## References

The following files contain the current implementation:

- The enrichment engine lives in
  `internal/enrichment/engine.go` and
  `internal/enrichment/prompts.go`.
- The triage UI lives in
  `client/src/pages/AnalysisTriagePage.tsx`.
- The relationship types are defined in
  `internal/database/relationship_types.go`.
- The enrichment handler lives in
  `internal/api/content_analysis_handler.go`.
