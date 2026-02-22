<!--
  Imagineer - TTRPG Campaign Intelligence Platform

  Copyright (c) 2025 - 2026
  This software is released under The MIT License
-->

# RAG Context for Enrichment and Revision Agents
# Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use
> superpowers:executing-plans to implement this plan
> task-by-task.

**Goal:** Add RAG context (hybrid vector search results
and game system schema) to the enrichment agent and
revision agent prompts.

**Architecture:** Two independent changes. The enrichment
agent performs a per-entity vector search inside its
`Run()` method and passes results plus game system YAML
through to the prompt builder via new `EnrichmentInput`
fields. The revision handler populates the existing
`GameSystemYAML` field and a new `CampaignResults` field
on `RevisionInput`, and the prompt builder formats them
into the LLM prompt.

**Tech Stack:** Go, PostgreSQL (pgedge_vectorizer hybrid
search).

---

### Task 1: Add RAG Fields to EnrichmentInput

**Files:**

- Modify: `internal/enrichment/engine.go:44-53`
- Test: `internal/enrichment/engine_test.go`

**Step 1: Write the failing test**

In `engine_test.go`, add a test that asserts the
enrichment prompt includes campaign context and game
system YAML when the new fields are populated.

```go
func TestBuildUserPrompt_WithRAGContext(t *testing.T) {
	desc := "A grizzled veteran."
	input := EnrichmentInput{
		CampaignID:  1,
		JobID:       10,
		SourceTable: "chapters",
		SourceID:    5,
		Content:     "Viktor entered the tavern.",
		Entity: models.Entity{
			ID:         100,
			CampaignID: 1,
			EntityType: models.EntityTypeNPC,
			Name:       "Viktor",
			Description: &desc,
		},
		CampaignResults: []models.SearchResult{
			{
				SourceTable: "chapters",
				SourceID:    2,
				SourceName:  "Chapter 2",
				ChunkContent: "Viktor was first seen " +
					"arriving at the docks.",
			},
		},
		GameSystemYAML: "name: Call of Cthulhu 7e\n" +
			"skills:\n  - Spot Hidden",
	}

	prompt := buildUserPrompt(input)

	assert.Contains(t, prompt,
		"## Campaign Context")
	assert.Contains(t, prompt,
		"Chapter 2")
	assert.Contains(t, prompt,
		"arriving at the docks")
	assert.Contains(t, prompt,
		"## Game System Schema")
	assert.Contains(t, prompt,
		"Call of Cthulhu 7e")
}

func TestBuildUserPrompt_NoRAGContext(t *testing.T) {
	input := EnrichmentInput{
		Content: "Some content.",
		Entity: models.Entity{
			ID:         1,
			EntityType: models.EntityTypeNPC,
			Name:       "Test",
		},
	}

	prompt := buildUserPrompt(input)

	assert.NotContains(t, prompt,
		"## Campaign Context")
	assert.NotContains(t, prompt,
		"## Game System Schema")
}
```

**Step 2: Run the test to verify that it fails**

Run the following command.

```bash
go test ./internal/enrichment/ \
    -run "TestBuildUserPrompt_WithRAGContext|TestBuildUserPrompt_NoRAGContext" -v
```

Expected: FAIL (the fields do not exist on
`EnrichmentInput`).

**Step 3: Add the fields to EnrichmentInput**

In `engine.go`, add two fields to the `EnrichmentInput`
struct after the `Relationships` field.

```go
// EnrichmentInput contains everything needed to enrich
// a single entity from a content source.
type EnrichmentInput struct {
	CampaignID      int64
	JobID           int64
	SourceTable     string
	SourceID        int64
	Content         string // Source content (Markdown)
	Entity          models.Entity
	OtherEntities   []models.Entity
	Relationships   []models.Relationship
	CampaignResults []models.SearchResult // RAG: campaign vector search results
	GameSystemYAML  string               // RAG: game system schema
}
```

**Step 4: Update the prompt builder**

In `prompts.go`, in the `buildUserPrompt` function, add
two new sections after the "Other Entities in This Content"
block and before the final instruction line. Insert the
following code before the line that starts with
`b.WriteString("Analyse the source content")`.

```go
	// Campaign context from RAG (vector search results).
	if len(input.CampaignResults) > 0 {
		fmt.Fprintf(&b,
			"## Campaign Context\n\n")
		fmt.Fprintf(&b,
			"Related content from the campaign "+
				"(for continuity and context):\n\n")
		for _, r := range input.CampaignResults {
			chunk := r.ChunkContent
			if len([]rune(chunk)) > 300 {
				chunk = string(
					[]rune(chunk)[:300]) + "..."
			}
			fmt.Fprintf(&b,
				"- **%s** (%s): %s\n",
				r.SourceName, r.SourceTable, chunk)
		}
		b.WriteString("\n")
	}

	// Game system schema from RAG.
	if input.GameSystemYAML != "" {
		fmt.Fprintf(&b,
			"## Game System Schema\n\n")
		fmt.Fprintf(&b, "```yaml\n%s\n```\n\n",
			input.GameSystemYAML)
	}
```

**Step 5: Run the test to verify that it passes**

Run the following command.

```bash
go test ./internal/enrichment/ \
    -run "TestBuildUserPrompt" -v
```

Expected: PASS (all `TestBuildUserPrompt*` tests pass,
including the existing ones).

**Step 6: Commit the changes**

```bash
git add internal/enrichment/engine.go \
       internal/enrichment/prompts.go \
       internal/enrichment/engine_test.go
git commit -m "feat: add RAG context fields to \
  EnrichmentInput and enrichment prompt"
```

---

### Task 2: Wire RAG Context in EnrichmentAgent.Run()

**Files:**

- Modify:
  `internal/enrichment/enrichment_agent.go:57-189`

**Step 1: Verify that existing tests still pass**

Run the following command.

```bash
go test ./internal/enrichment/ -v
```

Expected: PASS.

**Step 2: Add per-entity vector search and populate RAG
fields**

In `enrichment_agent.go`, in the `Run()` method, add RAG
context preparation before the per-entity loop (after the
`allKnownEntities` block, before the comment
`// Enrich each entity individually.`).

```go
	// Prepare RAG context. Extract game system YAML
	// from pipeline context and check vectorization
	// availability for per-entity search.
	var gameSystemYAML string
	var vectorAvailable bool
	if input.Context != nil {
		gameSystemYAML = input.Context.GameSystemYAML
		vectorAvailable = a.db.IsVectorizationAvailable(ctx)
	}
```

Then inside the per-entity loop, after building
`relationships` and before constructing `enrichInput`,
add a per-entity vector search.

```go
		// Per-entity RAG: search campaign content
		// for context about this specific entity.
		var campaignResults []models.SearchResult
		if vectorAvailable {
			results, searchErr :=
				a.db.SearchCampaignContent(
					ctx, input.CampaignID,
					entity.Name, 5)
			if searchErr != nil {
				log.Printf(
					"enrichment-agent: RAG search "+
						"failed for entity %d: %v",
					entity.ID, searchErr)
			} else {
				campaignResults = results
			}
		}
```

Then update the `enrichInput` construction to include the
new fields.

```go
		enrichInput := EnrichmentInput{
			CampaignID:      input.CampaignID,
			JobID:           input.JobID,
			SourceTable:     input.SourceTable,
			SourceID:        input.SourceID,
			Content:         input.Content,
			Entity:          entity,
			OtherEntities:   otherEntities,
			Relationships:   relationships,
			CampaignResults: campaignResults,
			GameSystemYAML:  gameSystemYAML,
		}
```

**Step 3: Run the tests**

Run the following command.

```bash
go test ./internal/enrichment/ -v
```

Expected: PASS.

**Step 4: Commit the changes**

```bash
git add internal/enrichment/enrichment_agent.go
git commit -m "feat: wire per-entity RAG context in \
  EnrichmentAgent.Run()"
```

---

### Task 3: Add Campaign Context to RevisionInput

**Files:**

- Modify:
  `internal/enrichment/revision_agent.go:25-138`
- Test:
  `internal/enrichment/revision_agent_test.go`

**Step 1: Write the failing test**

In `revision_agent_test.go`, add a test that asserts
campaign context appears in the revision prompt when
provided.

```go
func TestRevisionAgent_CampaignContext(t *testing.T) {
	var captured llm.CompletionRequest
	innerProvider := &mockProvider{
		response: `{"revisedContent": "revised", ` +
			`"summary": "updated"}`,
	}
	captureProvider := &capturingProvider{
		inner:    innerProvider,
		captured: &captured,
	}

	ra := NewRevisionAgent()

	suggestedContent, _ := json.Marshal(
		map[string]string{
			"description": "A finding",
		})

	_, err := ra.GenerateRevision(
		context.Background(),
		captureProvider,
		RevisionInput{
			OriginalContent: "The investigators " +
				"entered the manor.",
			AcceptedItems: []models.ContentAnalysisItem{
				{
					DetectionType:    "spelling",
					MatchedText:      "manor",
					SuggestedContent: suggestedContent,
				},
			},
			CampaignResults: []models.SearchResult{
				{
					SourceTable:  "chapters",
					SourceID:     2,
					SourceName:   "Chapter 2",
					ChunkContent: "The manor was built " +
						"in 1823 by Lord Ashton.",
				},
			},
		},
	)

	require.NoError(t, err)
	assert.Contains(t, captured.UserPrompt,
		"## Campaign Context")
	assert.Contains(t, captured.UserPrompt,
		"Chapter 2")
	assert.Contains(t, captured.UserPrompt,
		"Lord Ashton")
}
```

**Step 2: Run the test to verify that it fails**

Run the following command.

```bash
go test ./internal/enrichment/ \
    -run TestRevisionAgent_CampaignContext -v
```

Expected: FAIL (the `CampaignResults` field does not exist
on `RevisionInput`).

**Step 3: Add CampaignResults field to RevisionInput**

In `revision_agent.go`, add the new field to
`RevisionInput` after `GameSystemYAML`.

```go
// RevisionInput contains everything needed to generate
// a revision.
type RevisionInput struct {
	OriginalContent string
	AcceptedItems   []models.ContentAnalysisItem
	SourceTable     string
	SourceID        int64
	GameSystemYAML  string               // Optional game system context
	CampaignResults []models.SearchResult // RAG: campaign search results
}
```

**Step 4: Update buildRevisionUserPrompt**

In `revision_agent.go`, in `buildRevisionUserPrompt`,
add a campaign context section after the game system YAML
section (before the closing `return b.String()`).

```go
	if len(input.CampaignResults) > 0 {
		b.WriteString("## Campaign Context\n\n")
		b.WriteString("Related content from the " +
			"campaign (for continuity):\n\n")
		for _, r := range input.CampaignResults {
			chunk := r.ChunkContent
			if len([]rune(chunk)) > 300 {
				chunk = string(
					[]rune(chunk)[:300]) + "..."
			}
			fmt.Fprintf(&b,
				"- **%s** (%s): %s\n",
				r.SourceName, r.SourceTable, chunk)
		}
		b.WriteString("\n")
	}
```

**Step 5: Run the test to verify that it passes**

Run the following command.

```bash
go test ./internal/enrichment/ \
    -run TestRevisionAgent -v
```

Expected: PASS (all `TestRevisionAgent*` tests pass).

**Step 6: Commit the changes**

```bash
git add internal/enrichment/revision_agent.go \
       internal/enrichment/revision_agent_test.go
git commit -m "feat: add campaign context to \
  RevisionInput and revision prompt"
```

---

### Task 4: Populate Revision RAG Fields in the Handler

**Files:**

- Modify:
  `internal/api/content_analysis_handler.go:1958-1964`

**Step 1: Verify that existing tests still pass**

Run the following command.

```bash
go test ./internal/api/ -v
```

Expected: PASS.

**Step 2: Add RAG context to the revision handler**

In `content_analysis_handler.go`, before the revision
input construction at line 1958, add code to look up the
campaign's game system code and build RAG context. Insert
the following block after the LLM provider creation (after
line 1956) and before the comment
`// Call the RevisionAgent to generate revised content.`

```go
	// Build RAG context for the revision agent.
	campaign, campaignErr := h.db.GetCampaign(
		r.Context(), campaignID)
	var gameSystemCode string
	if campaignErr != nil {
		log.Printf(
			"GenerateRevision: failed to get "+
				"campaign %d: %v",
			campaignID, campaignErr)
	} else if campaign.System != nil {
		gameSystemCode = campaign.System.Code
	}

	ctxBuilder := enrichment.NewContextBuilder(
		h.db, "")
	ragCtx, ragErr := ctxBuilder.BuildContext(
		r.Context(), campaignID, originalContent,
		gameSystemCode, nil)
	if ragErr != nil {
		log.Printf(
			"GenerateRevision: failed to build "+
				"RAG context for job %d: %v",
			jobID, ragErr)
	}
```

Then update the `revisionInput` construction to include
the RAG fields.

```go
	// Call the RevisionAgent to generate revised content.
	revisionInput := enrichment.RevisionInput{
		OriginalContent: originalContent,
		AcceptedItems:   acceptedItems,
		SourceTable:     job.SourceTable,
		SourceID:        job.SourceID,
	}
	if ragCtx != nil {
		revisionInput.GameSystemYAML =
			ragCtx.GameSystemYAML
		revisionInput.CampaignResults =
			ragCtx.CampaignResults
	}
```

**Step 3: Run the tests**

Run the following command.

```bash
go test ./internal/api/ -v
```

Expected: PASS.

**Step 4: Commit the changes**

```bash
git add internal/api/content_analysis_handler.go
git commit -m "feat: populate RAG context for revision \
  agent in handler"
```

---

### Task 5: Final Verification

**Step 1: Run the full test suite**

Run the following command.

```bash
make test-all
```

Expected: All tests pass with no lint errors.

**Step 2: Update Todo.md**

Add a completed RAG integration entry to the completed
section of `Todo.md`, or check off any existing items
related to RAG integration for enrichment/revision.

**Step 3: Update CHANGELOG.md**

Add entries for the RAG context changes under the
Unreleased section.

**Step 4: Commit the changes**

```bash
git add Todo.md CHANGELOG.md
git commit -m "docs: update CHANGELOG and Todo for RAG \
  enrichment and revision integration"
```
