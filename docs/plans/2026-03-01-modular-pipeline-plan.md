<!--
  Imagineer - TTRPG Campaign Intelligence Platform

  Copyright (c) 2025 - 2026
  This software is released under The MIT License
-->

# Modular Pipeline Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use
> superpowers:executing-plans to implement this plan
> task-by-task.

**Goal:** Replace the monolithic two-stage pipeline with
a Phase Registry that assembles pipelines dynamically
from the GM's phase selections, making each phase an
atomic, independent unit.

**Architecture:** A PhaseRegistry maps phase names to
PipelineAgent sets. A BuildPipeline method creates a
Pipeline from an ordered list of phase names. All agents
(including identification scans) implement PipelineAgent.
Handlers call a single RunPipeline method instead of
branching between AnalyzeContent and
RunContentEnrichment.

**Tech Stack:** Go 1.23, pgx/v5, React/TypeScript

---

### Task 1: Create PhaseRegistry with BuildPipeline

**Files:**

- Create: `internal/enrichment/registry.go`
- Create: `internal/enrichment/registry_test.go`

**Context:** The PhaseRegistry maps short-form phase
names ("identify", "revise", "enrich") to agent lists
and a long-form phase tag ("identification", "analysis",
"enrichment") used on items. BuildPipeline produces a
Pipeline with one Stage per requested phase.

**Step 1: Write the failing tests**

Write these tests in `registry_test.go`:

```go
func TestNewPhaseRegistry(t *testing.T) {
    reg := NewPhaseRegistry(nil)
    assert.NotNil(t, reg)
}

func TestRegister(t *testing.T) {
    reg := NewPhaseRegistry(nil)
    agent := &mockAgent{name: "test"}
    reg.Register("custom", "custom_phase", agent)
    p := reg.BuildPipeline([]string{"custom"})
    assert.Len(t, p.stages, 1)
    assert.Equal(t, "custom", p.stages[0].Name)
    assert.Equal(t, "custom_phase", p.stages[0].Phase)
    assert.Len(t, p.stages[0].Agents, 1)
}

func TestBuildPipeline_MultiplePhases(t *testing.T) {
    reg := NewPhaseRegistry(nil)
    a1 := &mockAgent{name: "a1"}
    a2 := &mockAgent{name: "a2"}
    reg.Register("alpha", "alpha_tag", a1)
    reg.Register("beta", "beta_tag", a2)
    p := reg.BuildPipeline([]string{"alpha", "beta"})
    assert.Len(t, p.stages, 2)
    assert.Equal(t, "alpha", p.stages[0].Name)
    assert.Equal(t, "beta", p.stages[1].Name)
}

func TestBuildPipeline_RepeatedPhase(t *testing.T) {
    reg := NewPhaseRegistry(nil)
    a := &mockAgent{name: "a"}
    reg.Register("x", "x_tag", a)
    p := reg.BuildPipeline([]string{"x", "x"})
    assert.Len(t, p.stages, 2)
}

func TestBuildPipeline_UnknownPhaseSkipped(t *testing.T) {
    reg := NewPhaseRegistry(nil)
    a := &mockAgent{name: "a"}
    reg.Register("known", "k_tag", a)
    p := reg.BuildPipeline(
        []string{"known", "unknown"})
    assert.Len(t, p.stages, 1)
}

func TestBuildPipeline_EmptyPhases(t *testing.T) {
    reg := NewPhaseRegistry(nil)
    p := reg.BuildPipeline(nil)
    assert.Len(t, p.stages, 0)
}
```

Use a minimal mockAgent that implements PipelineAgent
(Name returns the stored name, Run returns nil, DependsOn
returns nil).

Note: the tests access `p.stages` which is currently
unexported. The implementation step will need to export it
or add an accessor. The simplest approach: make the
`stages` field on Pipeline exported (`Stages`). Update
`pipeline.go` accordingly — the only internal reference
to `p.stages` is in `Pipeline.Run()`.

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/enrichment/ -run TestNew -v`
Expected: FAIL (registry.go doesn't exist)

**Step 3: Write the implementation**

In `registry.go`:

```go
package enrichment

import "log"

// phaseEntry holds the agents and item-level phase tag
// for a registered pipeline phase.
type phaseEntry struct {
    phaseTag string          // long-form: "identification"
    agents   []PipelineAgent
}

// PhaseRegistry maps short-form phase names to their
// agents and phase tags. Use Register to add phases and
// BuildPipeline to assemble a Pipeline from an ordered
// list of phase names.
type PhaseRegistry struct {
    entries map[string]phaseEntry
}

// NewPhaseRegistry creates an empty registry.
func NewPhaseRegistry(db interface{}) *PhaseRegistry {
    return &PhaseRegistry{
        entries: make(map[string]phaseEntry),
    }
}

// Register adds agents for a named phase. The phaseTag
// is the long-form name used on items (e.g.,
// "identification", "analysis", "enrichment").
func (r *PhaseRegistry) Register(
    name, phaseTag string, agents ...PipelineAgent,
) {
    r.entries[name] = phaseEntry{
        phaseTag: phaseTag,
        agents:   agents,
    }
}

// BuildPipeline creates a Pipeline with one Stage per
// requested phase. Unknown phases are logged and skipped.
// Repeated phases produce repeated stages.
func (r *PhaseRegistry) BuildPipeline(
    phases []string,
) *Pipeline {
    var stages []Stage
    for _, name := range phases {
        entry, ok := r.entries[name]
        if !ok {
            log.Printf(
                "pipeline: unknown phase %q, skipping",
                name)
            continue
        }
        stages = append(stages, Stage{
            Name:   name,
            Phase:  entry.phaseTag,
            Agents: entry.agents,
        })
    }
    return &Pipeline{stages: stages}
}
```

Also update `pipeline.go`: export the `stages` field on
Pipeline to `Stages` so tests can inspect it. Update the
one reference in `Pipeline.Run()` from `p.stages` to
`p.Stages`. Update `NewPipeline` similarly. Update
`buildDefaultPipeline` in `source_content.go` if it
references the field directly (it doesn't — it passes
stages to NewPipeline).

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/enrichment/ -run "TestNew|TestRegister|TestBuild" -v`
Expected: PASS

**Step 5: Commit**

```
git add internal/enrichment/registry.go \
    internal/enrichment/registry_test.go \
    internal/enrichment/pipeline.go
git commit -m "feat: add PhaseRegistry with BuildPipeline"
```

---

### Task 2: Create IdentificationAgent

**Files:**

- Create:
  `internal/enrichment/identification_agent.go`
- Create:
  `internal/enrichment/identification_agent_test.go`

**Context:** Wraps the existing scan logic from
`internal/analysis/analyzer.go` into a PipelineAgent.
The Analyzer struct has three scan methods:

- `scanWikiLinks(ctx, campaignID, content) (items, resolvedNames)`
- `scanUntaggedMentions(ctx, campaignID, content, wikiRanges) items`
- `scanMisspellings(ctx, campaignID, content, matchedNames, wikiRanges) items`

Plus helpers `wikiLinkOriginalRanges(content)` and
`buildMatchedNames(resolvedNames, untaggedItems)`.

The IdentificationAgent needs access to the database
(same as the Analyzer) to resolve entity names. It does
NOT need an LLM provider.

**Step 1: Write the failing tests**

Write tests that verify:

- `Name()` returns `"identification"`
- `DependsOn()` returns `nil`
- `Run()` with empty content returns empty items
- `Run()` with content containing wiki links produces
  items (use a test database or mock — follow the
  pattern in `internal/analysis/analyzer_test.go`)

Check `analyzer_test.go` first for test patterns. The
identification agent needs a database connection for
entity lookups. If the analyzer tests use integration
tests with a real DB, follow that pattern. If they use
mocks, follow that pattern.

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/enrichment/ -run TestIdentification -v`
Expected: FAIL

**Step 3: Write the implementation**

```go
package enrichment

import (
    "context"

    "github.com/antonypegg/imagineer/internal/analysis"
    "github.com/antonypegg/imagineer/internal/database"
    "github.com/antonypegg/imagineer/internal/llm"
    "github.com/antonypegg/imagineer/internal/models"
)

// IdentificationAgent wraps the content analysis scanner
// as a PipelineAgent. It detects wiki links, untagged
// entity mentions, and potential misspellings.
type IdentificationAgent struct {
    analyzer *analysis.Analyzer
}

// NewIdentificationAgent creates an
// IdentificationAgent.
func NewIdentificationAgent(
    db *database.DB,
) *IdentificationAgent {
    return &IdentificationAgent{
        analyzer: analysis.NewAnalyzer(db),
    }
}

func (a *IdentificationAgent) Name() string {
    return "identification"
}

func (a *IdentificationAgent) DependsOn() []string {
    return nil
}

// Run executes identification scans on the input
// content. The provider argument is unused (no LLM
// calls needed).
func (a *IdentificationAgent) Run(
    ctx context.Context,
    _ llm.Provider,
    input PipelineInput,
) ([]models.ContentAnalysisItem, error) {
    return a.analyzer.ScanContent(
        ctx, input.CampaignID, input.Content)
}
```

This requires exposing a new public method on the
Analyzer. In `internal/analysis/analyzer.go`, add:

```go
// ScanContent runs all identification scans on the
// given content and returns the detected items. Unlike
// AnalyzeContent, this method does not create a job or
// persist items — it only performs the scans.
func (a *Analyzer) ScanContent(
    ctx context.Context,
    campaignID int64,
    content string,
) ([]models.ContentAnalysisItem, error) {
    if content == "" {
        return nil, nil
    }

    var items []models.ContentAnalysisItem

    wikiItems, resolvedNames := a.scanWikiLinks(
        ctx, campaignID, content)
    items = append(items, wikiItems...)

    origRanges := wikiLinkOriginalRanges(content)

    untaggedItems := a.scanUntaggedMentions(
        ctx, campaignID, content, origRanges)
    items = append(items, untaggedItems...)

    matchedNames := buildMatchedNames(
        resolvedNames, untaggedItems)
    misspellingItems := a.scanMisspellings(
        ctx, campaignID, content,
        matchedNames, origRanges)
    items = append(items, misspellingItems...)

    return items, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/enrichment/ -run TestIdentification -v`
Expected: PASS

**Step 5: Commit**

```
git add internal/enrichment/identification_agent.go \
    internal/enrichment/identification_agent_test.go \
    internal/analysis/analyzer.go
git commit -m "feat: add IdentificationAgent wrapping analyzer scans"
```

---

### Task 3: Register default phases

**Files:**

- Modify: `internal/enrichment/registry.go`
- Modify: `internal/enrichment/registry_test.go`
- Modify: `internal/api/source_content.go`

**Context:** Add a `NewDefaultRegistry(db)` function
that creates a PhaseRegistry with all three phases
registered. Replace `buildDefaultPipeline()` with
registry usage.

**Step 1: Write the failing test**

```go
func TestNewDefaultRegistry(t *testing.T) {
    reg := NewDefaultRegistry(nil)
    // Should have 3 phases registered.
    p := reg.BuildPipeline(
        []string{"identify", "revise", "enrich"})
    require.Len(t, p.Stages, 3)

    assert.Equal(t, "identify", p.Stages[0].Name)
    assert.Equal(t, "identification",
        p.Stages[0].Phase)
    assert.Len(t, p.Stages[0].Agents, 1)
    assert.Equal(t, "identification",
        p.Stages[0].Agents[0].Name())

    assert.Equal(t, "revise", p.Stages[1].Name)
    assert.Equal(t, "analysis", p.Stages[1].Phase)
    assert.Len(t, p.Stages[1].Agents, 2)

    assert.Equal(t, "enrich", p.Stages[2].Name)
    assert.Equal(t, "enrichment", p.Stages[2].Phase)
    assert.Len(t, p.Stages[2].Agents, 2)
}

func TestDefaultRegistry_SinglePhase(t *testing.T) {
    reg := NewDefaultRegistry(nil)
    p := reg.BuildPipeline([]string{"revise"})
    require.Len(t, p.Stages, 1)
    assert.Equal(t, "analysis", p.Stages[0].Phase)
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/enrichment/ -run TestNewDefault -v`
Expected: FAIL

**Step 3: Write the implementation**

Add to `registry.go`:

```go
import (
    "github.com/antonypegg/imagineer/internal/agents/canon"
    "github.com/antonypegg/imagineer/internal/agents/graph"
    "github.com/antonypegg/imagineer/internal/agents/ttrpg"
    "github.com/antonypegg/imagineer/internal/database"
)

// NewDefaultRegistry creates a PhaseRegistry with the
// standard three phases registered: identify (wiki link
// and entity scanning), revise (TTRPG and canon
// analysis), and enrich (entity enrichment and graph
// validation).
func NewDefaultRegistry(
    db *database.DB,
) *PhaseRegistry {
    reg := NewPhaseRegistry(db)
    reg.Register("identify", "identification",
        NewIdentificationAgent(db))
    reg.Register("revise", "analysis",
        ttrpg.NewExpert(), canon.NewExpert())
    reg.Register("enrich", "enrichment",
        NewEnrichmentAgent(db), graph.NewExpert(db))
    return reg
}
```

Then update `source_content.go`:

- Delete the `buildDefaultPipeline` function entirely
- It is no longer needed — callers will use the registry

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/enrichment/ -run TestNewDefault -v`
Also: `go build ./...` (verify nothing else breaks)
Expected: PASS

**Step 5: Commit**

```
git add internal/enrichment/registry.go \
    internal/enrichment/registry_test.go \
    internal/api/source_content.go
git commit -m "feat: register default phases and remove buildDefaultPipeline"
```

---

### Task 4: Remove hardcoded Phase from agents

**Files:**

- Modify: `internal/agents/ttrpg/expert.go` (lines
  117, 146)
- Modify: `internal/agents/canon/expert.go` (line 137)
- Modify: `internal/enrichment/engine.go` (lines 183,
  206, 259)
- Modify: `internal/enrichment/parser.go` (line 162)
- Modify: `internal/agents/graph/expert.go` (lines 109,
  165, 218, 268, 380)
- Modify: `internal/enrichment/pipeline.go` (add phase
  tagging in Run)

**Context:** Agents should not set Phase on items. The
pipeline's Run method assigns Phase from the Stage
definition after each agent runs. This decouples agents
from phases so the same agent can participate in
different phases.

**Step 1: Write the failing test**

Add to `pipeline_test.go`:

```go
func TestPipeline_SetsPhaseFromStage(t *testing.T) {
    agent := &mockPipelineAgent{
        name: "test",
        items: []models.ContentAnalysisItem{
            {DetectionType: "finding"},
        },
    }
    p := NewPipeline(nil, []Stage{
        {Name: "s1", Phase: "my_phase",
            Agents: []PipelineAgent{agent}},
    })
    items, err := p.Run(
        context.Background(), nil,
        PipelineInput{})
    require.NoError(t, err)
    require.Len(t, items, 1)
    assert.Equal(t, "my_phase", items[0].Phase)
    assert.Equal(t, "test", items[0].AgentName)
}
```

**Step 2: Run test to verify it fails**

The test may pass if Phase is already set by the agent.
To make it a meaningful test, the agent intentionally
sets `Phase: "wrong"` and we assert the pipeline
overwrites it to `"my_phase"`.

**Step 3: Write the implementation**

In `pipeline.go`, in the `Run` method, after tagging
AgentName, also set Phase:

```go
for i := range items {
    items[i].AgentName = agent.Name()
    items[i].Phase = stage.Phase
}
```

Then remove all hardcoded `Phase:` assignments from:

- `internal/agents/ttrpg/expert.go`: remove
  `Phase: "analysis"` at lines 117 and 146
- `internal/agents/canon/expert.go`: remove
  `Phase: "analysis"` at line 137
- `internal/enrichment/engine.go`: remove
  `Phase: "enrichment"` at lines 183, 206, 259
- `internal/enrichment/parser.go`: remove
  `Phase: "enrichment"` at line 162
- `internal/agents/graph/expert.go`: remove
  `Phase: "enrichment"` at lines 109, 165, 218,
  268, 380

**Step 4: Run full test suite**

Run: `make test-all`
Expected: Some existing tests may assert specific Phase
values. Update those tests to still pass — the phase
tagging now happens in the pipeline, so tests that run
agents directly (outside a pipeline) will get empty
Phase. For agent-level tests, either remove the Phase
assertion or accept empty string. For pipeline-level
tests, Phase should match the Stage.Phase.

**Step 5: Commit**

```
git add internal/agents/ttrpg/expert.go \
    internal/agents/canon/expert.go \
    internal/enrichment/engine.go \
    internal/enrichment/parser.go \
    internal/agents/graph/expert.go \
    internal/enrichment/pipeline.go \
    internal/enrichment/pipeline_test.go
git commit -m "refactor: move phase tagging from agents to pipeline"
```

---

### Task 5: Remove Graph expert's hard dependency

**Files:**

- Modify: `internal/agents/graph/expert.go`
  (line 52-54)
- Modify: `internal/agents/graph/expert_test.go`
  (if any tests assert DependsOn)

**Context:** The graph expert currently declares
`DependsOn: ["enrichment"]`. This is removed so it can
run independently. When registered in the Enrich phase
alongside the enrichment agent, registration order
ensures the enrichment agent runs first.

**Step 1: Change DependsOn to return nil**

In `expert.go`, change:

```go
func (e *Expert) DependsOn() []string {
    return []string{"enrichment"}
}
```

to:

```go
func (e *Expert) DependsOn() []string {
    return nil
}
```

**Step 2: Run tests**

Run: `go test ./internal/agents/graph/ -v`
Expected: PASS (graph expert already handles empty
PriorResults gracefully)

**Step 3: Commit**

```
git add internal/agents/graph/expert.go
git commit -m \
    "refactor: remove graph expert hard dependency on enrichment agent"
```

---

### Task 6: Create RunPipeline method

**Files:**

- Modify:
  `internal/api/content_analysis_handler.go`
- Create:
  `internal/api/content_analysis_handler_pipeline_test.go`
  (or add to existing test file)

**Context:** A single RunPipeline method replaces both
`RunContentEnrichment` and the `AnalyzeContent` call
pattern. It creates a job, builds a pipeline from the
registry, runs it in a background goroutine, saves
items, and updates job status.

The ContentAnalysisHandler struct needs a new field:
`registry *enrichment.PhaseRegistry`. This replaces the
`analyzer` field (the IdentificationAgent now holds the
Analyzer internally).

**Step 1: Update ContentAnalysisHandler struct**

Change the struct from:

```go
type ContentAnalysisHandler struct {
    db            *database.DB
    analyzer      *analysis.Analyzer
    enrichCancels sync.Map
}
```

to:

```go
type ContentAnalysisHandler struct {
    db            *database.DB
    registry      *enrichment.PhaseRegistry
    enrichCancels sync.Map
}
```

Update the constructor:

```go
func NewContentAnalysisHandler(
    db *database.DB,
) *ContentAnalysisHandler {
    return &ContentAnalysisHandler{
        db:       db,
        registry: enrichment.NewDefaultRegistry(db),
    }
}
```

**Step 2: Write RunPipeline method**

```go
// RunPipeline creates an analysis job and runs the
// pipeline for the given phases in a background
// goroutine. It returns the created job immediately.
func (h *ContentAnalysisHandler) RunPipeline(
    ctx context.Context,
    campaignID int64,
    sourceTable, sourceField string,
    sourceID int64,
    content string,
    userID int64,
    phases []string,
) (*models.ContentAnalysisJob, error) {
    // Delete previous jobs for this source.
    if err := h.db.DeleteAnalysisJobsForSource(
        ctx, campaignID, sourceTable,
        sourceID, sourceField,
    ); err != nil {
        return nil, fmt.Errorf(
            "failed to delete old jobs: %w", err)
    }

    // Create job.
    job := &models.ContentAnalysisJob{
        CampaignID:  campaignID,
        SourceTable: sourceTable,
        SourceID:    sourceID,
        SourceField: sourceField,
        Status:      "running",
        Phases:      phases,
    }
    createdJob, err := h.db.CreateAnalysisJob(
        ctx, job)
    if err != nil {
        return nil, fmt.Errorf(
            "failed to create job: %w", err)
    }

    // Build pipeline from registry.
    pipeline := h.registry.BuildPipeline(phases)

    // Fetch user settings for LLM provider.
    settings, err := h.db.GetUserSettings(
        ctx, userID)
    if err != nil || settings == nil {
        // No LLM — can only run identify phase.
        // Still proceed; agents that need LLM will
        // get nil provider and should handle it.
    }
    var provider llm.Provider
    if settings != nil &&
        settings.ContentGenService != nil &&
        settings.ContentGenAPIKey != nil {
        provider, _ = llm.NewProvider(
            *settings.ContentGenService,
            *settings.ContentGenAPIKey)
    }

    // Look up campaign for game system.
    campaign, _ := h.db.GetCampaign(
        ctx, campaignID)
    var gameSystemCode string
    var gameSystemID *int64
    if campaign != nil {
        if campaign.System != nil {
            gameSystemCode = campaign.System.Code
        }
        gameSystemID = campaign.SystemID
    }

    jobID := createdJob.ID

    // Spawn background goroutine.
    bgCtx, cancel := context.WithTimeout(
        context.Background(), 10*time.Minute)
    h.enrichCancels.Store(jobID, cancel)
    go func() {
        defer h.enrichCancels.Delete(jobID)
        defer cancel()

        // Build RAG context.
        ctxBuilder := enrichment.NewContextBuilder(
            h.db, "")
        ragCtx, _ := ctxBuilder.BuildContext(
            bgCtx, campaignID, content,
            gameSystemCode, nil)

        // Load relationships.
        relationships, _ :=
            h.db.ListRelationshipsByCampaign(
                bgCtx, campaignID)

        input := enrichment.PipelineInput{
            CampaignID:    campaignID,
            JobID:         jobID,
            SourceTable:   sourceTable,
            SourceID:      sourceID,
            SourceField:   sourceField,
            SourceScope:   enrichment.ScopeFromSourceTable(
                sourceTable),
            Content:       content,
            Relationships: relationships,
            GameSystemID:  gameSystemID,
            Context:       ragCtx,
            Ontology:      h.db.Ontology,
        }

        items, err := pipeline.Run(
            bgCtx, provider, input)
        if err != nil {
            log.Printf(
                "Pipeline: run failed for job %d: %v",
                jobID, err)
            _ = h.db.SetJobFailureReason(
                context.Background(), jobID,
                "Pipeline encountered an error")
            return
        }

        // Assign job ID and persist items.
        for i := range items {
            items[i].JobID = jobID
        }
        if len(items) > 0 {
            if err := h.db.CreateAnalysisItems(
                bgCtx, items); err != nil {
                log.Printf(
                    "Pipeline: failed to save items "+
                        "for job %d: %v",
                    jobID, err)
                return
            }
        }

        // Update job counts and status.
        // Count items by phase for the appropriate
        // total fields.
        identifyCount := 0
        enrichCount := 0
        for _, item := range items {
            switch item.Phase {
            case "identification":
                identifyCount++
            case "enrichment":
                enrichCount++
            }
        }

        _ = h.db.Exec(bgCtx,
            `UPDATE content_analysis_jobs
             SET status = 'completed',
                 total_items = $2,
                 enrichment_total = $3
             WHERE id = $1`,
            jobID, identifyCount, enrichCount)

        log.Printf(
            "Pipeline: completed job %d — %d items",
            jobID, len(items))
    }()

    return createdJob, nil
}
```

Note: this is a starting point. The implementer should
check how `RunContentEnrichment` handles SSE streaming
notifications (search for `h.sseHub` or `h.notifier`
or similar patterns) and replicate that in RunPipeline.

**Step 3: Verify compilation**

Run: `go build ./...`
Expected: PASS

**Step 4: Commit**

```
git add internal/api/content_analysis_handler.go
git commit -m "feat: add RunPipeline method to ContentAnalysisHandler"
```

---

### Task 7: Update handlers to use RunPipeline

**Files:**

- Modify: `internal/api/handlers.go`

**Context:** Replace all `shouldAnalyze` /
`shouldEnrich` branching with a single call to
`RunPipeline`. Every handler that currently has this
pattern:

```go
shouldAnalyze := r.URL.Query().Get(
    "analyze") == "true"
shouldEnrich := r.URL.Query().Get(
    "enrich") == "true"
phases := parsePhases(r)

if shouldAnalyze {
    job, _, _ := h.analyzer.AnalyzeContent(...)
    if h.caHandler != nil {
        h.caHandler.RunContentEnrichment(...)
    }
} else if shouldEnrich && h.caHandler != nil {
    job := h.createEnrichmentJob(...)
    h.caHandler.RunContentEnrichment(...)
}
```

becomes:

```go
phases := parsePhases(r)
if len(phases) > 0 && content != "" &&
    h.caHandler != nil {
    job, err := h.caHandler.RunPipeline(
        r.Context(), campaignID,
        sourceTable, sourceField, sourceID,
        content, userID, phases)
    if err != nil {
        log.Printf("Pipeline failed: %v", err)
    } else {
        response.Analysis = &models.AnalysisSummary{
            JobID:        job.ID,
            PendingCount: 0,
        }
    }
}
```

Apply this pattern to ALL handler locations. There are
7 locations in handlers.go:

1. CreateCampaign — campaigns/description
2. UpdateCampaign — campaigns/description
3. UpdateEntity — entities/description
4. UpdateEntity — entities/gm_notes
5. UpdateChapter — chapters/overview
6. UpdateSession — sessions/prep_notes
7. UpdateSession — sessions/actual_notes

Also:

- Remove `shouldAnalyze` and `shouldEnrich` variable
  declarations from each handler
- Remove the `createEnrichmentJob` method
- Remove the `analyzer` field from the Handler struct
  (it is no longer needed — identification runs via
  the pipeline)
- Keep `parsePhases()` — it is still used

**Step 1: Make all changes**

Apply the pattern to all 7 locations. Check that
`campaignID`, `sourceTable`, `sourceField`,
`sourceID`, `content`, and `userID` are all available
at each call site.

**Step 2: Verify compilation**

Run: `go build ./...`
Expected: PASS

**Step 3: Run full test suite**

Run: `make test-all`
Expected: PASS (some handler tests may need updating
if they mock AnalyzeContent or RunContentEnrichment)

**Step 4: Commit**

```
git add internal/api/handlers.go
git commit -m "refactor: replace shouldAnalyze/shouldEnrich with RunPipeline"
```

---

### Task 8: Remove old code paths

**Files:**

- Modify:
  `internal/api/content_analysis_handler.go`
- Modify: `internal/analysis/analyzer.go`

**Context:** Remove the code that is no longer called.

**Step 1: Remove from content_analysis_handler.go**

- Delete `RunContentEnrichment()` method
- Delete `TryAutoEnrich()` method
- Remove callers of `TryAutoEnrich` (search for
  `h.TryAutoEnrich` in the same file)

**Step 2: Remove from analyzer.go**

- Delete `AnalyzeContent()` method (the public entry
  point that created jobs). Keep `ScanContent()` (used
  by IdentificationAgent) and all private scan methods.

**Step 3: Verify compilation**

Run: `go build ./...`
Expected: PASS

**Step 4: Run full test suite**

Run: `make test-all`
Expected: PASS (update any tests that reference
removed methods)

**Step 5: Commit**

```
git add internal/api/content_analysis_handler.go \
    internal/analysis/analyzer.go
git commit -m \
    "chore: remove RunContentEnrichment, TryAutoEnrich, and AnalyzeContent"
```

---

### Task 9: Update frontend API client

**Files:**

- Modify: `client/src/api/campaigns.ts`
- Modify: `client/src/api/types.ts`

**Context:** Remove `analyze` and `enrich` boolean
params. The `phases` array is the only mechanism now.

**Step 1: Update types.ts**

Change `AnalysisOptions` from:

```typescript
export interface AnalysisOptions {
    analyze?: boolean;
    enrich?: boolean;
    phases?: string[];
}
```

to:

```typescript
export interface AnalysisOptions {
    phases?: string[];
}
```

**Step 2: Update campaigns.ts**

Remove lines that append `analyze` and `enrich` params.
Keep the `phases` param handling. The create and update
methods should only append `?phases=...` when
`options.phases` is non-empty.

**Step 3: Verify TypeScript compilation**

Run: `cd client && npx tsc --noEmit`
Expected: PASS (fix any references to removed fields)

**Step 4: Run frontend tests**

Run: `cd client && npx vitest run`
Expected: PASS

**Step 5: Commit**

```
git add client/src/api/campaigns.ts \
    client/src/api/types.ts
git commit -m "chore: remove analyze/enrich params from frontend API client"
```

---

### Task 10: Final verification and cleanup

**Step 1: Run full test suite**

Run: `make test-all`
Expected: `=== All tests passed ===`

**Step 2: Search for stale references**

Search for any remaining references to:

- `shouldAnalyze`
- `shouldEnrich`
- `RunContentEnrichment`
- `TryAutoEnrich`
- `buildDefaultPipeline`
- `AnalyzeContent` (as a called method, not test
  references)
- `analyze.*true` in frontend code
- `enrich.*true` in frontend code (as query params,
  not phase names)

Remove any found.

**Step 3: Rebuild and restart server**

```
go build -o bin/server ./cmd/server
# kill old server, start new one
```

**Step 4: Commit any cleanup**

```
git add -A
git commit -m "chore: remove stale references to old pipeline code"
```
