<!--
  Imagineer - TTRPG Campaign Intelligence Platform

  Copyright (c) 2025 - 2026
  This software is released under The MIT License
-->

# RAG Context for Enrichment and Revision Agents

This design document describes adding RAG context (hybrid
vector search results and game system schema) to the
enrichment agent and the revision agent. The analysis-stage
agents (ttrpg-expert, canon-expert) already consume RAG
context; the enrichment and revision agents currently
ignore it.

## Problem

### Enrichment Agent Ignores RAG Context

The `EnrichmentAgent.Run()` receives the full
`PipelineInput` (which includes `Context *RAGContext`), but
when building per-entity `EnrichmentInput` structs it only
passes the source content, entity data, and relationships.
The `EnrichmentInput` struct has no fields for campaign
search results or game system schema. Entity description
updates and relationship suggestions are generated without
broader campaign context.

### Revision Agent Has Dead Code

The `RevisionInput` struct has a `GameSystemYAML` field and
the prompt builder checks for it, but the handler that
invokes the revision agent never populates it. The struct
also has no field for campaign search results.

## Design

### 1. Per-Entity RAG for the Enrichment Agent

The enrichment agent enriches entities one at a time. For
each entity, the agent performs a targeted vector search
using the entity name as the query. This produces results
more relevant to the specific entity than the pipeline-level
content-summary query.

Changes to `EnrichmentInput` (in `engine.go`):

- Add `CampaignResults []models.SearchResult`.
- Add `GameSystemYAML string`.

Changes to `EnrichmentAgent.Run()` (in
`enrichment_agent.go`):

- Before the per-entity loop, check
  `db.IsVectorizationAvailable()` and extract
  `GameSystemYAML` from `input.Context`.
- Inside the per-entity loop, call
  `db.SearchCampaignContent()` with the entity name as
  query (limit 5).
- Populate the new fields on `EnrichmentInput`.

Changes to the enrichment prompt builder (in
`prompts.go:buildUserPrompt`):

- Append a "Campaign context" section listing campaign
  search results as bullet points with source name and
  truncated chunk content.
- Append a "Game system schema" section with the YAML in a
  code block.

The vector search is guarded by
`IsVectorizationAvailable()`. When vectorization is
unavailable, the new fields are left empty and the prompt
sections are omitted.

### 2. Populate RevisionInput Fields

Changes to `RevisionInput` (in `revision_agent.go`):

- Add `CampaignResults []models.SearchResult`.

Changes to the revision handler (in
`content_analysis_handler.go`):

- Load the game system schema YAML from disk and set
  `RevisionInput.GameSystemYAML`.
- Run a content-summary vector search and set
  `RevisionInput.CampaignResults`.

Changes to `buildRevisionUserPrompt()`:

- Add a "Campaign context" section (same format as the
  enrichment agent).
- The existing `GameSystemYAML` section requires no changes
  since the code is already written but was never
  triggered.

## Files to Modify

| File | Change |
|------|--------|
| `internal/enrichment/engine.go` | Add RAG fields to `EnrichmentInput`. |
| `internal/enrichment/enrichment_agent.go` | Per-entity vector search, populate RAG fields. |
| `internal/enrichment/prompts.go` | Campaign context and game system sections in enrichment prompt. |
| `internal/enrichment/revision_agent.go` | Add `CampaignResults` field, update prompt builder. |
| `internal/api/content_analysis_handler.go` | Populate `GameSystemYAML` and `CampaignResults` for revision. |

## What This Does Not Change

- No changes to the graph expert agent.
- No changes to analysis-stage agents (already have RAG).
- No changes to `ContextBuilder` or `RAGContext`.
- No new database queries or tables (reuses
  `SearchCampaignContent`).
- No frontend changes.
- No new API endpoints.
- No migrations.
