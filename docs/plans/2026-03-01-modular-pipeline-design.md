<!--
  Imagineer - TTRPG Campaign Intelligence Platform

  Copyright (c) 2025 - 2026
  This software is released under The MIT License
-->

# Modular Pipeline Design

## Problem

The current pipeline is a monolithic two-stage block
(analysis then enrichment) that always runs both stages.
There is no way to assemble it based on which phases the
GM selected. The three phases (Identify, Revise, Enrich)
are treated as dependent on each other, but they are
independent. Any phase can run alone, and phases can
repeat (e.g., Identify after Enrich to catch newly
created entities).

Additionally, the Identify phase lives in a completely
separate code path (`analyzer.AnalyzeContent()`) from the
Revise/Enrich phases (`RunContentEnrichment()`), causing
duplicated orchestration logic across 7+ handler locations
with `shouldAnalyze` / `shouldEnrich` branching.

## Design

Replace the hardcoded pipeline with a Phase Registry that
maps phase names to agents, and a Pipeline Builder that
assembles pipelines dynamically from the GM's phase
selections.

### Core Principles

1. **Atomic phases**: each phase is a self-contained,
   independent unit of work. It produces a complete set
   of items for its purpose, requires no output from
   other phases, and can run alone or repeat without
   side effects. PriorResults gives a phase optional
   visibility into earlier output but is never required.
2. **Composable pipelines**: the GM's phase choices
   define the pipeline stages, in order. The pipeline
   is simply an ordered sequence of atomic phases.
3. **Repeatable phases**: a phase can appear multiple
   times (e.g., [identify, enrich, identify])
4. **Uniform interface**: all agents (including
   identification scans) implement PipelineAgent
5. **Phase tagging from stage**: agents do not hardcode
   their phase -- the stage assigns it

### Phase Registry

A `PhaseRegistry` struct in
`internal/enrichment/registry.go` maps phase names to
their agents. The server initialises the registry once at
startup with the database handle.

```go
type PhaseRegistry struct {
    phases map[string][]PipelineAgent
}

func NewPhaseRegistry(db *database.DB) *PhaseRegistry
func (r *PhaseRegistry) Register(
    phase string, agents ...PipelineAgent,
)
func (r *PhaseRegistry) BuildPipeline(
    phases []string,
) *Pipeline
```

Default registration:

| Phase      | Agents                       | Phase Tag        |
|------------|------------------------------|------------------|
| `identify` | IdentificationAgent          | `identification` |
| `revise`   | TTRPGExpert, CanonExpert     | `analysis`       |
| `enrich`   | EnrichmentAgent, GraphExpert | `enrichment`     |

`BuildPipeline()` iterates the phases slice and creates
one Stage per entry, looking up agents from the registry.
Unknown phases are skipped with a log warning. Repeated
phases produce repeated stages.

### IdentificationAgent

A new PipelineAgent wrapper in
`internal/enrichment/identification_agent.go` wraps the
existing scan logic from `internal/analysis/analyzer.go`.
It calls the same three scan functions (wiki links,
untagged mentions, misspellings) through the PipelineAgent
interface.

The Analyzer struct's scan methods become internal helpers
used by the IdentificationAgent. The public
`AnalyzeContent()` method is removed.

### Handler Simplification

All handlers use a single unified method instead of the
`shouldAnalyze` / `shouldEnrich` branching:

```go
func (h *ContentAnalysisHandler) RunPipeline(
    ctx context.Context,
    campaignID int64,
    sourceTable, sourceField string,
    sourceID int64,
    content string,
    userID int64,
    phases []string,
) (*models.ContentAnalysisJob, error)
```

This method:

1. Deletes any previous analysis jobs for the source
2. Creates a new job with the given phases
3. Builds a pipeline from the registry using
   `BuildPipeline(phases)`
4. Runs the pipeline synchronously within a background
   goroutine
5. Saves all produced items to the job
6. Updates job status and counts

Each handler shrinks from ~30 lines of analysis branching
to ~5 lines: parse phases, call RunPipeline, set
response.

### Pipeline Execution Changes

**Phase tagging from stage.** After each stage completes,
the pipeline tags all returned items with the stage's
Phase field. Agents remove their hardcoded Phase
assignments.

```go
for i := range items {
    items[i].AgentName = agent.Name()
    items[i].Phase = stage.Phase
}
```

**PriorResults.** The existing mechanism continues
unchanged. Before each stage, `input.PriorResults` is
populated with items from all prior stages. This enables
patterns like [identify, enrich, identify] where the
second identify stage can see entities created by the
enrich stage.

**Synchronous pipeline, async caller.** The pipeline
itself is always synchronous. The caller (RunPipeline)
spawns a background goroutine with a context timeout.

### Graph Expert Independence

The graph expert's hard dependency on the enrichment
agent (`DependsOn: ["enrichment"]`) is removed. The graph
expert already handles the case where no
relationship_suggestion items exist in PriorResults -- it
skips semantic validation and still runs structural checks
(orphan detection, type pair validation, cardinality,
missing required relationships).

When registered in the Enrich phase alongside the
enrichment agent, the enrichment agent runs first
(registration order) and PriorResults naturally provides
its output to the graph expert.

### API Changes

- `?analyze=true` and `?enrich=true` query params are
  removed
- `?phases=identify,revise,enrich` becomes the single
  mechanism
- The frontend already sends `?phases=...` via
  PhaseStrip -- no frontend API changes needed

### Backend Removal List

The following functions and code paths are removed:

- `buildDefaultPipeline()` in `source_content.go`
- `RunContentEnrichment()` in
  `content_analysis_handler.go`
- `createEnrichmentJob()` in
  `content_analysis_handler.go`
- `TryAutoEnrich()` in `content_analysis_handler.go`
- `analyzer.AnalyzeContent()` as a public method
- `shouldAnalyze` / `shouldEnrich` branching in all 7+
  handlers
- References to `analyze=true` / `enrich=true` in
  frontend API client

### Database

No schema changes are required. The existing
`job_phases` table, `content_analysis_items.phase`
column, and `content_analysis_jobs` structure all support
this design.

### Testing

- Existing pipeline tests continue to work
- New tests cover: PhaseRegistry, BuildPipeline,
  IdentificationAgent
- Handler tests that mock AnalyzeContent need updating
  to mock RunPipeline

### Migration Path

1. Create PhaseRegistry and IdentificationAgent
2. Create RunPipeline method
3. Update all handlers to use RunPipeline
4. Remove old code paths
5. Remove frontend references to analyze/enrich query
   params
