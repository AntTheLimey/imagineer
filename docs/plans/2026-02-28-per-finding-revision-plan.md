<!--
  Imagineer - TTRPG Campaign Intelligence Platform

  Copyright (c) 2025 - 2026
  This software is released under The MIT License
-->

# Per-Finding Surgical Revision - Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use
> superpowers:executing-plans to implement this plan
> task-by-task.

**Goal:** Replace the monolithic revision flow with
per-finding surgical edits that are generated on demand,
reviewed inline, and applied immediately.

**Architecture:** New per-item revision endpoints replace
the existing per-job revision endpoints. Each finding
generates a targeted edit using only its surrounding
paragraph as context. The frontend shows inline +/- diffs
per finding card with accept/edit/dismiss controls.

**Tech Stack:** Go (chi router, LLM provider interface),
React/TypeScript (MUI, React Query), PostgreSQL

---

## Task 1: Per-Finding Revision Agent

Create a new revision agent that generates surgical edits
for individual findings using only a small context window
around the finding's position.

**Files:**

- Create: `internal/enrichment/finding_revision_agent.go`
- Create:
  `internal/enrichment/finding_revision_agent_test.go`

**Step 1: Write the failing tests**

Create
`internal/enrichment/finding_revision_agent_test.go`
with the following content.

```go
/*-------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------
 */

package enrichment

import (
	"context"
	"strings"
	"testing"

	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProvider returns a fixed response for testing.
type mockProvider struct {
	response string
	err      error
}

func (m *mockProvider) Complete(
	_ context.Context,
	_ llm.CompletionRequest,
) (llm.CompletionResponse, error) {
	if m.err != nil {
		return llm.CompletionResponse{}, m.err
	}
	return llm.CompletionResponse{
		Content: m.response,
	}, nil
}

func TestGenerateFindingRevision_Basic(t *testing.T) {
	provider := &mockProvider{
		response: "The kingdom fell into bitter chaos.",
	}

	input := FindingRevisionInput{
		ContextWindow: "The kingdom fell into chaos.",
		Findings: []FindingDetail{
			{
				DetectionType: "content_suggestion",
				MatchedText:   "fell into chaos",
				Description:   "Too brief",
				Suggestion:    "Add more detail",
			},
		},
	}

	result, err := GenerateFindingRevision(
		context.Background(), provider, input,
	)
	require.NoError(t, err)
	assert.NotEmpty(t, result.RevisedSection)
	assert.Equal(t,
		"The kingdom fell into bitter chaos.",
		result.RevisedSection,
	)
}

func TestGenerateFindingRevision_EmptyContext(
	t *testing.T,
) {
	provider := &mockProvider{}

	input := FindingRevisionInput{
		ContextWindow: "",
		Findings: []FindingDetail{
			{
				DetectionType: "content_suggestion",
				MatchedText:   "test",
			},
		},
	}

	_, err := GenerateFindingRevision(
		context.Background(), provider, input,
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context window")
}

func TestGenerateFindingRevision_NoFindings(
	t *testing.T,
) {
	provider := &mockProvider{}

	input := FindingRevisionInput{
		ContextWindow: "Some content here.",
		Findings:      []FindingDetail{},
	}

	result, err := GenerateFindingRevision(
		context.Background(), provider, input,
	)
	require.NoError(t, err)
	assert.Equal(t, "Some content here.",
		result.RevisedSection,
	)
}

func TestGenerateFindingRevision_WithInstructions(
	t *testing.T,
) {
	var capturedReq llm.CompletionRequest
	provider := &mockProvider{
		response: "Dark and foreboding kingdom.",
	}

	// Wrap to capture the request.
	wrapper := &capturingProvider{
		inner:    provider,
		captured: &capturedReq,
	}

	input := FindingRevisionInput{
		ContextWindow: "The kingdom was peaceful.",
		Findings: []FindingDetail{
			{
				DetectionType: "content_suggestion",
				MatchedText:   "was peaceful",
				Suggestion:    "Change tone",
			},
		},
		Instructions: "make it dark and foreboding",
	}

	result, err := GenerateFindingRevision(
		context.Background(), wrapper, input,
	)
	require.NoError(t, err)
	assert.NotEmpty(t, result.RevisedSection)
	assert.Contains(t,
		capturedReq.UserPrompt,
		"make it dark and foreboding",
	)
}

func TestBuildFindingSystemPrompt(t *testing.T) {
	prompt := buildFindingSystemPrompt()
	assert.Contains(t, prompt, "TTRPG content editor")
	assert.Contains(t, prompt, "surgical")
}

func TestBuildFindingUserPrompt(t *testing.T) {
	input := FindingRevisionInput{
		ContextWindow: "The old manor loomed.",
		Findings: []FindingDetail{
			{
				DetectionType: "canon_contradiction",
				MatchedText:   "old manor",
				Description:   "Manor was destroyed",
				Suggestion:    "Change to ruins",
			},
		},
		Instructions: "keep it spooky",
	}

	prompt := buildFindingUserPrompt(input)
	assert.Contains(t, prompt, "The old manor loomed.")
	assert.Contains(t, prompt, "canon_contradiction")
	assert.Contains(t, prompt, "old manor")
	assert.Contains(t, prompt, "Manor was destroyed")
	assert.Contains(t, prompt, "Change to ruins")
	assert.Contains(t, prompt, "keep it spooky")
}

func TestBuildFindingUserPrompt_NoInstructions(
	t *testing.T,
) {
	input := FindingRevisionInput{
		ContextWindow: "Some text.",
		Findings: []FindingDetail{
			{
				DetectionType: "content_suggestion",
				MatchedText:   "Some",
			},
		},
	}

	prompt := buildFindingUserPrompt(input)
	assert.NotContains(t, prompt, "GM Instructions")
}

func TestBuildFindingUserPrompt_MultipleFindings(
	t *testing.T,
) {
	input := FindingRevisionInput{
		ContextWindow: "The manor was old and dark.",
		Findings: []FindingDetail{
			{
				DetectionType: "content_suggestion",
				MatchedText:   "old",
				Suggestion:    "ancient",
			},
			{
				DetectionType: "mechanics_warning",
				MatchedText:   "dark",
				Description:   "Needs lighting rules",
			},
		},
	}

	prompt := buildFindingUserPrompt(input)
	assert.True(t,
		strings.Contains(prompt, "Finding 1") &&
			strings.Contains(prompt, "Finding 2"),
	)
}

// capturingProvider wraps a provider to capture the
// CompletionRequest for inspection in tests.
type capturingProvider struct {
	inner    llm.Provider
	captured *llm.CompletionRequest
}

func (c *capturingProvider) Complete(
	ctx context.Context,
	req llm.CompletionRequest,
) (llm.CompletionResponse, error) {
	*c.captured = req
	return c.inner.Complete(ctx, req)
}
```

**Step 2: Verify the tests fail**

```bash
cd /Users/antonypegg/PROJECTS/imagineer && \
    go test ./internal/enrichment/ \
    -run TestGenerateFindingRevision -v 2>&1 | head -30
```

The tests should fail because
`finding_revision_agent.go` does not exist yet.

**Step 3: Implement the finding revision agent**

Create
`internal/enrichment/finding_revision_agent.go`
with the following content.

```go
/*-------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------
 */

package enrichment

import (
	"context"
	"fmt"
	"strings"

	"github.com/antonypegg/imagineer/internal/llm"
)

// FindingRevisionInput contains the context and findings
// for a surgical per-finding revision.
type FindingRevisionInput struct {
	// ContextWindow is the paragraph(s) surrounding the
	// finding's position -- typically the target paragraph
	// plus one before and one after.
	ContextWindow string

	// Findings lists the individual detections to address
	// within the context window.
	Findings []FindingDetail

	// Instructions contains optional GM-provided guidance
	// for the revision (e.g., "make it darker").
	Instructions string
}

// FindingDetail describes a single analysis finding to
// be addressed by the revision agent.
type FindingDetail struct {
	DetectionType string // e.g., "content_suggestion"
	MatchedText   string // The text span that was flagged
	Description   string // What the finding describes
	Suggestion    string // Suggested replacement or fix
}

// FindingRevisionResult holds the output of a
// per-finding revision.
type FindingRevisionResult struct {
	RevisedSection string
}

// GenerateFindingRevision produces a surgical edit for
// one or more findings within a context window. It
// builds a focused prompt and calls the LLM with a small
// token budget (1024 max). If no findings are provided,
// the context window is returned unchanged without an
// LLM call.
func GenerateFindingRevision(
	ctx context.Context,
	provider llm.Provider,
	input FindingRevisionInput,
) (*FindingRevisionResult, error) {
	if input.ContextWindow == "" {
		return nil, fmt.Errorf(
			"context window is required for " +
				"finding revision",
		)
	}

	if len(input.Findings) == 0 {
		return &FindingRevisionResult{
			RevisedSection: input.ContextWindow,
		}, nil
	}

	systemPrompt := buildFindingSystemPrompt()
	userPrompt := buildFindingUserPrompt(input)

	resp, err := provider.Complete(ctx,
		llm.CompletionRequest{
			SystemPrompt: systemPrompt,
			UserPrompt:   userPrompt,
			MaxTokens:    1024,
			Temperature:  0.4,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"LLM completion failed: %w", err,
		)
	}

	revised := strings.TrimSpace(resp.Content)
	return &FindingRevisionResult{
		RevisedSection: revised,
	}, nil
}

// buildFindingSystemPrompt returns the system prompt for
// surgical per-finding revisions. Unlike the full-document
// revision prompt, this instructs the LLM to edit only
// the provided section and return plain text (not JSON).
func buildFindingSystemPrompt() string {
	return `You are a surgical TTRPG content editor. ` +
		`You will receive a short section of text ` +
		`and one or more findings that need to be ` +
		`addressed within that section.

Rules:
- Make ONLY the changes needed to address the findings.
- Preserve the author's voice, tone, and writing style.
- Preserve all markdown formatting.
- Do NOT add content beyond what the findings require.
- Do NOT wrap your response in code fences or JSON.
- Return ONLY the revised section text, nothing else.`
}

// buildFindingUserPrompt constructs the user prompt with
// the context window, findings, and optional GM
// instructions.
func buildFindingUserPrompt(
	input FindingRevisionInput,
) string {
	var b strings.Builder

	b.WriteString("## Section to Edit\n\n")
	b.WriteString(input.ContextWindow)
	b.WriteString("\n\n")

	b.WriteString("## Findings\n\n")
	for i, f := range input.Findings {
		fmt.Fprintf(&b, "### Finding %d\n\n", i+1)
		fmt.Fprintf(&b,
			"**Detection Type**: %s\n",
			f.DetectionType,
		)
		fmt.Fprintf(&b,
			"**Matched Text**: %s\n",
			f.MatchedText,
		)
		if f.Description != "" {
			fmt.Fprintf(&b,
				"**Description**: %s\n",
				f.Description,
			)
		}
		if f.Suggestion != "" {
			fmt.Fprintf(&b,
				"**Suggestion**: %s\n",
				f.Suggestion,
			)
		}
		b.WriteString("\n")
	}

	if input.Instructions != "" {
		b.WriteString("## GM Instructions\n\n")
		b.WriteString(input.Instructions)
		b.WriteString("\n")
	}

	return b.String()
}
```

**Step 4: Verify the tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer && \
    go test ./internal/enrichment/ \
    -run TestGenerateFindingRevision -v
```

```bash
cd /Users/antonypegg/PROJECTS/imagineer && \
    go test ./internal/enrichment/ \
    -run TestBuildFinding -v
```

All tests should pass.

**Commit message:**
`feat: add per-finding surgical revision agent`

---

## Task 2: Context Window Extraction

Create utility functions for extracting the paragraph(s)
around a finding's position and grouping items that share
the same paragraph.

**Files:**

- Create: `internal/enrichment/context_window.go`
- Create: `internal/enrichment/context_window_test.go`

**Step 1: Write the failing tests**

Create `internal/enrichment/context_window_test.go`
with the following content.

```go
/*-------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------
 */

package enrichment

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractContextWindow_MiddleParagraph(
	t *testing.T,
) {
	content := "First paragraph.\n\n" +
		"Second paragraph with the target.\n\n" +
		"Third paragraph."

	// Position inside "Second paragraph" (offset 18).
	window, ws, we := ExtractContextWindow(
		content, 18, 35,
	)
	assert.Contains(t, window, "First paragraph.")
	assert.Contains(t, window,
		"Second paragraph with the target.",
	)
	assert.Contains(t, window, "Third paragraph.")
	assert.Equal(t, 0, ws)
	assert.Equal(t, len(content), we)
}

func TestExtractContextWindow_FirstParagraph(
	t *testing.T,
) {
	content := "First paragraph.\n\n" +
		"Second paragraph.\n\n" +
		"Third paragraph."

	// Position inside "First paragraph" (offset 5).
	window, ws, _ := ExtractContextWindow(
		content, 5, 10,
	)
	assert.Contains(t, window, "First paragraph.")
	assert.Contains(t, window, "Second paragraph.")
	assert.Equal(t, 0, ws)
	// Should not include third paragraph -- only target
	// plus one after.
	assert.NotContains(t, window, "Third paragraph.")
}

func TestExtractContextWindow_LastParagraph(
	t *testing.T,
) {
	content := "First paragraph.\n\n" +
		"Second paragraph.\n\n" +
		"Third paragraph."

	// Position inside "Third paragraph" (offset 38).
	window, _, we := ExtractContextWindow(
		content, 38, 45,
	)
	assert.Contains(t, window, "Second paragraph.")
	assert.Contains(t, window, "Third paragraph.")
	assert.Equal(t, len(content), we)
	// Should not include first paragraph -- only target
	// plus one before.
	assert.NotContains(t, window, "First paragraph.")
}

func TestExtractContextWindow_SingleParagraph(
	t *testing.T,
) {
	content := "Only one paragraph here."

	window, ws, we := ExtractContextWindow(
		content, 5, 10,
	)
	assert.Equal(t, content, window)
	assert.Equal(t, 0, ws)
	assert.Equal(t, len(content), we)
}

func TestExtractContextWindow_OutOfBounds(
	t *testing.T,
) {
	content := "Short."

	// Position beyond content length.
	window, ws, we := ExtractContextWindow(
		content, 100, 200,
	)
	assert.Equal(t, content, window)
	assert.Equal(t, 0, ws)
	assert.Equal(t, len(content), we)
}

func TestGroupItemsByParagraph_NoOverlap(
	t *testing.T,
) {
	content := "First paragraph.\n\n" +
		"Second paragraph.\n\n" +
		"Third paragraph."

	items := []ItemPosition{
		{ID: 1, Start: 5, End: 10},  // First para
		{ID: 2, Start: 22, End: 28}, // Second para
	}

	groups := GroupItemsByParagraph(items, content)
	require.Len(t, groups, 2)
	assert.Len(t, groups[0], 1)
	assert.Len(t, groups[1], 1)
}

func TestGroupItemsByParagraph_SameParagraph(
	t *testing.T,
) {
	content := "The kingdom fell into chaos " +
		"and darkness.\n\n" +
		"Other paragraph."

	items := []ItemPosition{
		{ID: 1, Start: 4, End: 11},  // "kingdom"
		{ID: 2, Start: 22, End: 27}, // "chaos"
	}

	groups := GroupItemsByParagraph(items, content)
	require.Len(t, groups, 1)
	assert.Len(t, groups[0], 2)
}

func TestGroupItemsByParagraph_Empty(t *testing.T) {
	groups := GroupItemsByParagraph(
		[]ItemPosition{}, "content",
	)
	assert.Empty(t, groups)
}

func TestGroupItemsByParagraph_OverlappingPositions(
	t *testing.T,
) {
	content := "A long paragraph with many words."

	items := []ItemPosition{
		{ID: 1, Start: 2, End: 15},
		{ID: 2, Start: 10, End: 25},
	}

	groups := GroupItemsByParagraph(items, content)
	require.Len(t, groups, 1)
	assert.Len(t, groups[0], 2)
}
```

**Step 2: Verify the tests fail**

```bash
cd /Users/antonypegg/PROJECTS/imagineer && \
    go test ./internal/enrichment/ \
    -run "TestExtractContextWindow|TestGroupItems" \
    -v 2>&1 | head -20
```

**Step 3: Implement context window extraction**

Create `internal/enrichment/context_window.go` with
the following content.

```go
/*-------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------
 */

package enrichment

import (
	"strings"
)

// ItemPosition is a lightweight struct for grouping
// items by their paragraph position without importing
// the full models package.
type ItemPosition struct {
	ID    int64
	Start int
	End   int
}

// ExtractContextWindow returns the paragraph containing
// the position range [start, end), plus one paragraph
// before and one after. Paragraphs are delimited by
// double newlines ("\n\n"). The returned windowStart and
// windowEnd are byte offsets into the original content.
func ExtractContextWindow(
	content string, start, end int,
) (window string, windowStart, windowEnd int) {
	if content == "" {
		return "", 0, 0
	}

	// Clamp positions to content bounds.
	if start < 0 {
		start = 0
	}
	if end > len(content) {
		end = len(content)
	}
	if start >= len(content) {
		return content, 0, len(content)
	}

	// Split into paragraphs, tracking byte offsets.
	type paraRange struct {
		start int
		end   int
	}
	var paragraphs []paraRange

	parts := strings.Split(content, "\n\n")
	offset := 0
	for i, part := range parts {
		pStart := offset
		pEnd := offset + len(part)
		paragraphs = append(paragraphs, paraRange{
			start: pStart,
			end:   pEnd,
		})
		// Account for the "\n\n" delimiter.
		offset = pEnd
		if i < len(parts)-1 {
			offset += 2
		}
	}

	// Find which paragraph contains the start position.
	targetIdx := len(paragraphs) - 1
	for i, p := range paragraphs {
		if start >= p.start && start < p.end {
			targetIdx = i
			break
		}
		// If start falls in a delimiter between
		// paragraphs, assign to the next paragraph.
		if i < len(paragraphs)-1 &&
			start >= p.end &&
			start < paragraphs[i+1].start {
			targetIdx = i + 1
			break
		}
	}

	// Expand to include one paragraph before and after.
	fromIdx := targetIdx - 1
	if fromIdx < 0 {
		fromIdx = 0
	}
	toIdx := targetIdx + 1
	if toIdx >= len(paragraphs) {
		toIdx = len(paragraphs) - 1
	}

	windowStart = paragraphs[fromIdx].start
	windowEnd = paragraphs[toIdx].end

	return content[windowStart:windowEnd],
		windowStart, windowEnd
}

// GroupItemsByParagraph groups items whose positions
// fall within the same paragraph (delimited by "\n\n")
// or whose position ranges overlap. Items are returned
// in the same relative order they appear in the input.
func GroupItemsByParagraph(
	items []ItemPosition, content string,
) [][]ItemPosition {
	if len(items) == 0 {
		return nil
	}

	// Build paragraph boundaries.
	type paraRange struct {
		start int
		end   int
	}
	var paragraphs []paraRange

	parts := strings.Split(content, "\n\n")
	offset := 0
	for i, part := range parts {
		pStart := offset
		pEnd := offset + len(part)
		paragraphs = append(paragraphs, paraRange{
			start: pStart,
			end:   pEnd,
		})
		offset = pEnd
		if i < len(parts)-1 {
			offset += 2
		}
	}

	// Assign each item to a paragraph index.
	paraIdx := func(pos int) int {
		for i, p := range paragraphs {
			if pos >= p.start && pos < p.end {
				return i
			}
			if i < len(paragraphs)-1 &&
				pos >= p.end &&
				pos < paragraphs[i+1].start {
				return i + 1
			}
		}
		return len(paragraphs) - 1
	}

	// Group by paragraph index. Items sharing a
	// paragraph are always in the same group.
	groupMap := make(map[int][]ItemPosition)
	var groupOrder []int

	for _, item := range items {
		idx := paraIdx(item.Start)
		if _, exists := groupMap[idx]; !exists {
			groupOrder = append(groupOrder, idx)
		}
		groupMap[idx] = append(groupMap[idx], item)
	}

	var result [][]ItemPosition
	for _, idx := range groupOrder {
		result = append(result, groupMap[idx])
	}
	return result
}
```

**Step 4: Verify the tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer && \
    go test ./internal/enrichment/ \
    -run "TestExtractContextWindow|TestGroupItems" -v
```

All tests should pass.

**Commit message:**
`feat: add context window extraction and paragraph grouping`

---

## Task 3: Database Operations for Position Shifting

Add database functions to shift item positions after a
revision is applied and to fetch multiple items by ID.

**Files:**

- Modify: `internal/database/content_analysis.go`
  (lines 283+)
- Create: `internal/database/content_analysis_shift_test.go`

**Step 1: Write the failing tests**

Create
`internal/database/content_analysis_shift_test.go`
with the following content. These are integration tests
that require a running database.

```go
/*-------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------
 */

package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShiftItemPositions(t *testing.T) {
	if testDB == nil {
		t.Skip("database not available")
	}

	ctx := context.Background()

	// Create a campaign, source, and analysis job for
	// test isolation.
	campaign := createTestCampaign(t, ctx)
	chapter := createTestChapter(t, ctx, campaign.ID)
	job := createTestAnalysisJob(
		t, ctx, campaign.ID, chapter.ID,
	)

	// Create items at known positions.
	start1, end1 := 10, 20
	start2, end2 := 30, 40
	start3, end3 := 50, 60

	item1 := createTestAnalysisItem(
		t, ctx, job.ID, &start1, &end1,
	)
	item2 := createTestAnalysisItem(
		t, ctx, job.ID, &start2, &end2,
	)
	item3 := createTestAnalysisItem(
		t, ctx, job.ID, &start3, &end3,
	)

	// Shift positions after position 25 by +5.
	err := testDB.ShiftItemPositions(
		ctx, job.ID, 25, 5,
	)
	require.NoError(t, err)

	// Item 1 (position 10-20) should be unchanged.
	items, err := testDB.GetAnalysisItemsByIDs(
		ctx, []int64{item1.ID},
	)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, 10, *items[0].PositionStart)
	assert.Equal(t, 20, *items[0].PositionEnd)

	// Item 2 (position 30-40) should be shifted to 35-45.
	items, err = testDB.GetAnalysisItemsByIDs(
		ctx, []int64{item2.ID},
	)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, 35, *items[0].PositionStart)
	assert.Equal(t, 45, *items[0].PositionEnd)

	// Item 3 (position 50-60) should be shifted to 55-65.
	items, err = testDB.GetAnalysisItemsByIDs(
		ctx, []int64{item3.ID},
	)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, 55, *items[0].PositionStart)
	assert.Equal(t, 65, *items[0].PositionEnd)
}

func TestShiftItemPositions_NegativeDelta(
	t *testing.T,
) {
	if testDB == nil {
		t.Skip("database not available")
	}

	ctx := context.Background()
	campaign := createTestCampaign(t, ctx)
	chapter := createTestChapter(t, ctx, campaign.ID)
	job := createTestAnalysisJob(
		t, ctx, campaign.ID, chapter.ID,
	)

	start, end := 50, 60
	item := createTestAnalysisItem(
		t, ctx, job.ID, &start, &end,
	)

	// Shift by -10 after position 20.
	err := testDB.ShiftItemPositions(
		ctx, job.ID, 20, -10,
	)
	require.NoError(t, err)

	items, err := testDB.GetAnalysisItemsByIDs(
		ctx, []int64{item.ID},
	)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, 40, *items[0].PositionStart)
	assert.Equal(t, 50, *items[0].PositionEnd)
}

func TestGetAnalysisItemsByIDs(t *testing.T) {
	if testDB == nil {
		t.Skip("database not available")
	}

	ctx := context.Background()
	campaign := createTestCampaign(t, ctx)
	chapter := createTestChapter(t, ctx, campaign.ID)
	job := createTestAnalysisJob(
		t, ctx, campaign.ID, chapter.ID,
	)

	start1, end1 := 10, 20
	start2, end2 := 30, 40

	item1 := createTestAnalysisItem(
		t, ctx, job.ID, &start1, &end1,
	)
	item2 := createTestAnalysisItem(
		t, ctx, job.ID, &start2, &end2,
	)

	items, err := testDB.GetAnalysisItemsByIDs(
		ctx, []int64{item1.ID, item2.ID},
	)
	require.NoError(t, err)
	require.Len(t, items, 2)
}

func TestGetAnalysisItemsByIDs_Empty(t *testing.T) {
	if testDB == nil {
		t.Skip("database not available")
	}

	ctx := context.Background()
	items, err := testDB.GetAnalysisItemsByIDs(
		ctx, []int64{},
	)
	require.NoError(t, err)
	assert.Empty(t, items)
}
```

**Step 2: Verify the tests fail**

```bash
cd /Users/antonypegg/PROJECTS/imagineer && \
    go test ./internal/database/ \
    -run "TestShiftItemPositions|TestGetAnalysisItemsByIDs" \
    -v 2>&1 | head -20
```

**Step 3: Add the database functions**

Add the following functions to
`internal/database/content_analysis.go` after the
`ResolveAnalysisItem` function (line 283).

```go
// ShiftItemPositions adjusts the position_start and
// position_end of all analysis items in a job whose
// position_start is greater than afterPosition. This is
// used after applying a revision to keep subsequent
// item positions aligned with the modified content.
func (db *DB) ShiftItemPositions(
	ctx context.Context,
	jobID int64,
	afterPosition int,
	delta int,
) error {
	query := `
		UPDATE content_analysis_items
		SET position_start = position_start + $3,
		    position_end = position_end + $3
		WHERE job_id = $1
		  AND position_start > $2
		  AND position_start IS NOT NULL`

	return db.Exec(ctx, query,
		jobID, afterPosition, delta,
	)
}

// GetAnalysisItemsByIDs retrieves multiple analysis items
// by their IDs, joining on entities to populate
// entity_name and entity_type. Items are returned in the
// order of their position_start.
func (db *DB) GetAnalysisItemsByIDs(
	ctx context.Context,
	itemIDs []int64,
) ([]models.ContentAnalysisItem, error) {
	if len(itemIDs) == 0 {
		return nil, nil
	}

	// Build a parameterised IN clause.
	placeholders := make([]string, len(itemIDs))
	args := make([]interface{}, len(itemIDs))
	for i, id := range itemIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT i.id, i.job_id, i.detection_type,
		       i.matched_text, i.entity_id,
		       i.similarity, i.context_snippet,
		       i.position_start, i.position_end,
		       i.resolution, i.resolved_entity_id,
		       i.resolved_at, i.created_at,
		       i.suggested_content, i.phase,
		       i.agent_name, i.pipeline_run_id,
		       e.name AS entity_name, e.entity_type
		FROM content_analysis_items i
		LEFT JOIN entities e ON i.entity_id = e.id
		WHERE i.id IN (%s)
		ORDER BY i.position_start ASC`,
		strings.Join(placeholders, ", "),
	)

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get analysis items by IDs: %w",
			err,
		)
	}
	defer rows.Close()

	return scanAnalysisItems(rows)
}
```

**Step 4: Verify the tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer && \
    go test ./internal/database/ \
    -run "TestShiftItemPositions|TestGetAnalysisItemsByIDs" \
    -v
```

All tests should pass.

**Commit message:**
`feat: add position shifting and batch item lookup`

---

## Task 4: Backend Handlers

Add the `GenerateItemRevision` and `ApplyItemRevision`
handlers and register them in the router. These handlers
implement the new per-finding revision endpoints.

**Files:**

- Modify: `internal/api/content_analysis_handler.go`
  (after line 2209)
- Modify: `internal/api/router.go` (after line 247)

**Step 1: Add request/response types and handlers**

Add the following code to
`internal/api/content_analysis_handler.go` after the
`ApplyRevision` handler (after line 2209).

```go
// -------------------------------------------------------
// Per-finding revision (surgical edits)
// -------------------------------------------------------

// GenerateItemRevisionRequest is the request body for
// generating a per-finding revision.
type GenerateItemRevisionRequest struct {
	ItemIDs      []int64 `json:"itemIds"`
	Instructions string  `json:"instructions,omitempty"`
}

// GenerateItemRevisionResponse is the response body for
// a per-finding revision generation.
type GenerateItemRevisionResponse struct {
	OriginalSection string `json:"originalSection"`
	RevisedSection  string `json:"revisedSection"`
}

// ApplyItemRevisionRequest is the request body for
// applying a per-finding revision to the source content.
type ApplyItemRevisionRequest struct {
	ItemIDs        []int64 `json:"itemIds"`
	RevisedSection string  `json:"revisedSection"`
}

// GenerateItemRevision handles
// POST /api/campaigns/{id}/analysis/items/revision
//
// Generates a surgical revision for one or more analysis
// items. The handler extracts a context window around the
// items' positions, builds a focused prompt, and returns
// the original and revised sections for inline diffing.
func (h *ContentAnalysisHandler) GenerateItemRevision(
	w http.ResponseWriter, r *http.Request,
) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest,
			"Invalid campaign ID")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized,
			"Authentication required")
		return
	}

	if err := h.db.VerifyCampaignOwnership(
		r.Context(), campaignID, userID,
	); err != nil {
		respondError(w, http.StatusNotFound,
			"Campaign not found")
		return
	}

	r.Body = http.MaxBytesReader(
		w, r.Body, maxRequestBodyBytes,
	)
	var req GenerateItemRevisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest,
			"Invalid request body")
		return
	}

	if len(req.ItemIDs) == 0 {
		respondError(w, http.StatusBadRequest,
			"itemIds is required")
		return
	}

	// Load the requested items.
	items, err := h.db.GetAnalysisItemsByIDs(
		r.Context(), req.ItemIDs,
	)
	if err != nil {
		log.Printf(
			"GenerateItemRevision: error loading "+
				"items: %v", err,
		)
		respondError(w, http.StatusInternalServerError,
			"Failed to load analysis items")
		return
	}

	if len(items) == 0 {
		respondError(w, http.StatusNotFound,
			"No analysis items found")
		return
	}

	// Verify all items belong to the same job and that
	// the job belongs to this campaign.
	jobID := items[0].JobID
	for _, item := range items {
		if item.JobID != jobID {
			respondError(w, http.StatusBadRequest,
				"All items must belong to the same job")
			return
		}
	}

	job, err := h.db.GetAnalysisJob(
		r.Context(), jobID,
	)
	if err != nil {
		respondError(w, http.StatusNotFound,
			"Analysis job not found")
		return
	}
	if job.CampaignID != campaignID {
		respondError(w, http.StatusNotFound,
			"Analysis job not found in this campaign")
		return
	}

	// Validate that all items have position data.
	for _, item := range items {
		if item.PositionStart == nil ||
			item.PositionEnd == nil {
			respondError(w, http.StatusBadRequest,
				fmt.Sprintf(
					"Item %d has no position data; "+
						"it cannot be revised",
					item.ID,
				),
			)
			return
		}
	}

	// Load the source content.
	sourceContent, err := h.loadSourceContent(
		r.Context(), job,
	)
	if err != nil {
		log.Printf(
			"GenerateItemRevision: error loading "+
				"source: %v", err,
		)
		respondError(w, http.StatusInternalServerError,
			"Failed to load source content")
		return
	}

	// Validate that matched text still aligns with
	// source content (stale position check).
	for _, item := range items {
		start := *item.PositionStart
		end := *item.PositionEnd
		if end > len(sourceContent) {
			respondError(w, http.StatusConflict,
				fmt.Sprintf(
					"Item %d position is beyond "+
						"content length; the "+
						"finding may be stale",
					item.ID,
				),
			)
			return
		}
		actual := sourceContent[start:end]
		if actual != item.MatchedText {
			respondError(w, http.StatusConflict,
				fmt.Sprintf(
					"Item %d position does not "+
						"match source content; "+
						"the finding may be stale "+
						"(expected %q, got %q)",
					item.ID,
					item.MatchedText,
					actual,
				),
			)
			return
		}
	}

	// Determine the combined position range.
	minStart := *items[0].PositionStart
	maxEnd := *items[0].PositionEnd
	for _, item := range items[1:] {
		if *item.PositionStart < minStart {
			minStart = *item.PositionStart
		}
		if *item.PositionEnd > maxEnd {
			maxEnd = *item.PositionEnd
		}
	}

	// Extract context window.
	contextWindow, windowStart, windowEnd :=
		enrichment.ExtractContextWindow(
			sourceContent, minStart, maxEnd,
		)

	// Build finding details from items.
	findings := make(
		[]enrichment.FindingDetail, 0, len(items),
	)
	for _, item := range items {
		detail := enrichment.FindingDetail{
			DetectionType: item.DetectionType,
			MatchedText:   item.MatchedText,
		}
		if len(item.SuggestedContent) > 0 {
			var content map[string]string
			if err := json.Unmarshal(
				item.SuggestedContent, &content,
			); err == nil {
				detail.Description = content["description"]
				detail.Suggestion = content["suggestion"]
			}
		}
		findings = append(findings, detail)
	}

	// Get the LLM provider.
	settings, err := h.db.GetUserSettings(
		r.Context(), userID,
	)
	if err != nil {
		log.Printf(
			"GenerateItemRevision: error getting "+
				"user settings: %v", err,
		)
		respondError(w, http.StatusInternalServerError,
			"Failed to get user settings")
		return
	}
	if settings == nil ||
		settings.ContentGenService == nil ||
		settings.ContentGenAPIKey == nil {
		respondError(w, http.StatusBadRequest,
			"LLM service not configured. "+
				"Configure an LLM in Account Settings.")
		return
	}

	provider, err := llm.NewProvider(
		*settings.ContentGenService,
		*settings.ContentGenAPIKey,
	)
	if err != nil {
		log.Printf(
			"GenerateItemRevision: error creating "+
				"LLM provider: %v", err,
		)
		respondError(w, http.StatusInternalServerError,
			"Failed to create LLM provider")
		return
	}

	// Call the finding revision agent.
	llmCtx, llmCancel := context.WithTimeout(
		context.Background(), 30*time.Second,
	)
	defer llmCancel()

	result, err := enrichment.GenerateFindingRevision(
		llmCtx, provider,
		enrichment.FindingRevisionInput{
			ContextWindow: contextWindow,
			Findings:      findings,
			Instructions:  req.Instructions,
		},
	)
	if err != nil {
		log.Printf(
			"GenerateItemRevision: revision agent "+
				"failed: %v", err,
		)
		respondError(w, http.StatusInternalServerError,
			"Failed to generate revision")
		return
	}

	_ = windowStart
	_ = windowEnd

	respondJSON(w, http.StatusOK,
		GenerateItemRevisionResponse{
			OriginalSection: contextWindow,
			RevisedSection:  result.RevisedSection,
		},
	)
}

// ApplyItemRevision handles
// PUT /api/campaigns/{id}/analysis/items/revision/apply
//
// Applies a per-finding revision to the source content.
// The handler splices the revised section into the source
// at the items' position range, shifts positions of
// subsequent items, and marks the items as accepted.
func (h *ContentAnalysisHandler) ApplyItemRevision(
	w http.ResponseWriter, r *http.Request,
) {
	campaignID, err := parseInt64(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest,
			"Invalid campaign ID")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized,
			"Authentication required")
		return
	}

	if err := h.db.VerifyCampaignOwnership(
		r.Context(), campaignID, userID,
	); err != nil {
		respondError(w, http.StatusNotFound,
			"Campaign not found")
		return
	}

	r.Body = http.MaxBytesReader(
		w, r.Body, maxRequestBodyBytes,
	)
	var req ApplyItemRevisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest,
			"Invalid request body")
		return
	}

	if len(req.ItemIDs) == 0 {
		respondError(w, http.StatusBadRequest,
			"itemIds is required")
		return
	}
	if req.RevisedSection == "" {
		respondError(w, http.StatusBadRequest,
			"revisedSection is required")
		return
	}

	// Load items.
	items, err := h.db.GetAnalysisItemsByIDs(
		r.Context(), req.ItemIDs,
	)
	if err != nil {
		log.Printf(
			"ApplyItemRevision: error loading "+
				"items: %v", err,
		)
		respondError(w, http.StatusInternalServerError,
			"Failed to load analysis items")
		return
	}
	if len(items) == 0 {
		respondError(w, http.StatusNotFound,
			"No analysis items found")
		return
	}

	// Verify all items belong to the same job and
	// campaign.
	jobID := items[0].JobID
	for _, item := range items {
		if item.JobID != jobID {
			respondError(w, http.StatusBadRequest,
				"All items must belong to the same job")
			return
		}
	}

	job, err := h.db.GetAnalysisJob(
		r.Context(), jobID,
	)
	if err != nil {
		respondError(w, http.StatusNotFound,
			"Analysis job not found")
		return
	}
	if job.CampaignID != campaignID {
		respondError(w, http.StatusNotFound,
			"Analysis job not found in this campaign")
		return
	}

	// Validate position data.
	for _, item := range items {
		if item.PositionStart == nil ||
			item.PositionEnd == nil {
			respondError(w, http.StatusBadRequest,
				fmt.Sprintf(
					"Item %d has no position data",
					item.ID,
				),
			)
			return
		}
	}

	// Load source content.
	sourceContent, err := h.loadSourceContent(
		r.Context(), job,
	)
	if err != nil {
		log.Printf(
			"ApplyItemRevision: error loading "+
				"source: %v", err,
		)
		respondError(w, http.StatusInternalServerError,
			"Failed to load source content")
		return
	}

	// Determine the context window that was revised.
	minStart := *items[0].PositionStart
	maxEnd := *items[0].PositionEnd
	for _, item := range items[1:] {
		if *item.PositionStart < minStart {
			minStart = *item.PositionStart
		}
		if *item.PositionEnd > maxEnd {
			maxEnd = *item.PositionEnd
		}
	}

	// Extract the same context window that was used
	// during generation.
	_, windowStart, windowEnd :=
		enrichment.ExtractContextWindow(
			sourceContent, minStart, maxEnd,
		)

	// Splice the revised section into the source.
	newContent := sourceContent[:windowStart] +
		req.RevisedSection +
		sourceContent[windowEnd:]

	// Calculate the length delta for position shifting.
	oldLen := windowEnd - windowStart
	newLen := len(req.RevisedSection)
	delta := newLen - oldLen

	// Update the source content in the database.
	if err := h.updateSourceContent(
		r.Context(),
		job.SourceTable,
		job.SourceID,
		job.SourceField,
		newContent,
	); err != nil {
		log.Printf(
			"ApplyItemRevision: error updating "+
				"source: %v", err,
		)
		respondError(w, http.StatusInternalServerError,
			"Failed to update source content")
		return
	}

	// Shift positions of subsequent items.
	if delta != 0 {
		if err := h.db.ShiftItemPositions(
			r.Context(), jobID, windowEnd, delta,
		); err != nil {
			log.Printf(
				"ApplyItemRevision: error shifting "+
					"positions: %v", err,
			)
			// Non-fatal: the content was already saved.
		}
	}

	// Mark items as accepted.
	for _, item := range items {
		if err := h.db.ResolveAnalysisItem(
			r.Context(), item.ID, "accepted", nil,
		); err != nil {
			log.Printf(
				"ApplyItemRevision: error resolving "+
					"item %d: %v", item.ID, err,
			)
		}
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status": "applied",
	})
}

// loadSourceContent reads the source text for an
// analysis job based on its source table and field.
func (h *ContentAnalysisHandler) loadSourceContent(
	ctx context.Context,
	job *models.ContentAnalysisJob,
) (string, error) {
	switch job.SourceTable {
	case "chapters":
		chapter, err := h.db.GetChapter(
			ctx, job.SourceID,
		)
		if err != nil {
			return "", fmt.Errorf(
				"failed to fetch chapter %d: %w",
				job.SourceID, err,
			)
		}
		if chapter.Overview != nil {
			return *chapter.Overview, nil
		}
		return "", nil

	case "sessions":
		session, err := h.db.GetSession(
			ctx, job.SourceID,
		)
		if err != nil {
			return "", fmt.Errorf(
				"failed to fetch session %d: %w",
				job.SourceID, err,
			)
		}
		switch job.SourceField {
		case "prep_notes":
			if session.PrepNotes != nil {
				return *session.PrepNotes, nil
			}
		case "actual_notes":
			if session.ActualNotes != nil {
				return *session.ActualNotes, nil
			}
		default:
			return "", fmt.Errorf(
				"unsupported session field: %s",
				job.SourceField,
			)
		}
		return "", nil

	case "campaigns":
		campaign, err := h.db.GetCampaign(
			ctx, job.SourceID,
		)
		if err != nil {
			return "", fmt.Errorf(
				"failed to fetch campaign %d: %w",
				job.SourceID, err,
			)
		}
		if campaign.Description != nil {
			return *campaign.Description, nil
		}
		return "", nil

	default:
		return "", fmt.Errorf(
			"unsupported source table: %s",
			job.SourceTable,
		)
	}
}
```

**Step 2: Register routes in the router**

In `internal/api/router.go`, add the new per-item
revision routes inside the `/analysis` route group,
after the existing `r.Put("/items/{itemId}/revert", ...)`
line (line 247). Insert the following lines.

```go
					// Per-finding revision (surgical)
					r.Post("/items/revision",
						contentAnalysisHandler.GenerateItemRevision)
					r.Put("/items/revision/apply",
						contentAnalysisHandler.ApplyItemRevision)
```

The routes section should now look like this (context for
placement).

```go
					r.Put("/items/{itemId}",
						contentAnalysisHandler.ResolveItem)
					r.Put("/items/{itemId}/revert",
						contentAnalysisHandler.RevertItem)

					// Per-finding revision (surgical)
					r.Post("/items/revision",
						contentAnalysisHandler.GenerateItemRevision)
					r.Put("/items/revision/apply",
						contentAnalysisHandler.ApplyItemRevision)
```

**Step 3: Verify the build compiles**

```bash
cd /Users/antonypegg/PROJECTS/imagineer && \
    go build ./...
```

**Step 4: Verify existing tests still pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer && \
    go test ./internal/api/ -v -count=1 2>&1 | tail -20
```

**Commit message:**
`feat: add per-finding revision handlers and routes`

---

## Task 5: Frontend API Client and Hooks

Add the TypeScript API functions and React Query hooks
for the new per-finding revision endpoints.

**Files:**

- Modify: `client/src/api/contentAnalysis.ts`
  (after line 310)
- Modify: `client/src/hooks/useContentAnalysis.ts`
  (after line 369)

**Step 1: Add TypeScript types and API functions**

In `client/src/api/contentAnalysis.ts`, add the
following types and functions after the existing
`revisionApi` object (after line 310, before the
`export default` line).

```typescript
/**
 * Request payload for generating a per-finding revision.
 */
export interface GenerateItemRevisionRequest {
    itemIds: number[];
    instructions?: string;
}

/**
 * Response from generating a per-finding revision.
 */
export interface GenerateItemRevisionResponse {
    originalSection: string;
    revisedSection: string;
}

/**
 * Request payload for applying a per-finding revision.
 */
export interface ApplyItemRevisionRequest {
    itemIds: number[];
    revisedSection: string;
}

/**
 * Per-finding revision API endpoints.
 */
export const itemRevisionApi = {
    /**
     * Generate a surgical revision for one or more
     * analysis items. The backend extracts a context
     * window around the items and uses a focused LLM
     * call.
     */
    generateItemRevision(
        campaignId: number,
        req: GenerateItemRevisionRequest,
    ): Promise<GenerateItemRevisionResponse> {
        return apiClient.post<
            GenerateItemRevisionResponse
        >(
            `/campaigns/${campaignId}` +
                `/analysis/items/revision`,
            req,
        );
    },

    /**
     * Apply a per-finding revision (possibly edited by
     * the GM) to the source content.
     */
    applyItemRevision(
        campaignId: number,
        req: ApplyItemRevisionRequest,
    ): Promise<ApplyRevisionResponse> {
        return apiClient.put<ApplyRevisionResponse>(
            `/campaigns/${campaignId}` +
                `/analysis/items/revision/apply`,
            req,
        );
    },
};
```

**Step 2: Add React Query mutation hooks**

In `client/src/hooks/useContentAnalysis.ts`, add the
following hooks after the existing `useApplyRevision`
function (after line 369).

First, add the new imports at the top of the file.
Update the import from `../api/contentAnalysis` to
include the new types.

```typescript
import {
    // ... existing imports ...
    itemRevisionApi,
    type GenerateItemRevisionRequest,
    type GenerateItemRevisionResponse,
    type ApplyItemRevisionRequest,
} from '../api/contentAnalysis';
```

Then add the hooks.

```typescript
/**
 * Mutation to generate a per-finding surgical revision
 * for one or more analysis items. Returns the original
 * and revised sections for inline diffing.
 *
 * @param campaignId - The campaign the items belong to.
 */
export function useGenerateItemRevision(
    campaignId: number,
) {
    return useMutation<
        GenerateItemRevisionResponse,
        Error,
        GenerateItemRevisionRequest
    >({
        mutationFn: (req) =>
            itemRevisionApi.generateItemRevision(
                campaignId, req,
            ),
    });
}

/**
 * Mutation to apply a per-finding revision to the
 * source content. Invalidates content analysis,
 * campaign, entity, chapter, and session caches on
 * success.
 *
 * @param campaignId - The campaign the items belong to.
 */
export function useApplyItemRevision(
    campaignId: number,
) {
    const queryClient = useQueryClient();
    return useMutation<
        ApplyRevisionResponse,
        Error,
        ApplyItemRevisionRequest
    >({
        mutationFn: (req) =>
            itemRevisionApi.applyItemRevision(
                campaignId, req,
            ),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: contentAnalysisKeys.all,
            });
            queryClient.invalidateQueries({
                queryKey: campaignKeys.all,
            });
            queryClient.invalidateQueries({
                queryKey: entityKeys.all,
            });
            queryClient.invalidateQueries({
                queryKey: chapterKeys.all,
            });
            queryClient.invalidateQueries({
                queryKey: sessionKeys.all,
            });
        },
    });
}
```

**Step 3: Verify TypeScript compiles**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client && \
    npx tsc --noEmit 2>&1 | tail -20
```

**Commit message:**
`feat: add per-finding revision API client and hooks`

---

## Task 6: RevisePhasePage Redesign

Replace the monolithic revision workflow at the top of
RevisePhasePage with per-finding inline revision cards.
Each finding card gets its own generate/accept/edit/
dismiss controls.

**Files:**

- Modify: `client/src/pages/RevisePhasePage.tsx`

**Step 1: Update imports**

At the top of `client/src/pages/RevisePhasePage.tsx`,
replace the import of `useGenerateRevision` and
`useApplyRevision` with the new per-finding hooks.

Replace the hooks import (lines 49-53).

```typescript
import {
    useResolveItem,
    useGenerateItemRevision,
    useApplyItemRevision,
} from '../hooks/useContentAnalysis';
```

Add `Collapse` and `AutoFixHigh` to the MUI imports.

```typescript
import {
    Box,
    Grid,
    List,
    ListItemButton,
    ListItemText,
    Typography,
    Chip,
    IconButton,
    Button,
    Stack,
    Divider,
    Alert,
    Paper,
    Tooltip,
    TextField,
    CircularProgress,
    Collapse,
} from '@mui/material';
import {
    Check,
    CheckCircle,
    Close,
    Edit,
    AutoFixHigh,
} from '@mui/icons-material';
```

**Step 2: Add per-finding revision state**

Replace the old local state block (lines 206-218) with
per-finding state. Remove `revisionContent`, `isEditing`,
`revisionCount`, and `revisionError`. Add per-item
revision tracking.

```typescript
    // -- Local state -----------------------------------

    const [selectedItemId, setSelectedItemId] = useState<
        number | null
    >(null);

    // Per-finding revision state keyed by item ID (or
    // the first item ID in a group).
    const [revisionState, setRevisionState] = useState<
        Record<
            number,
            {
                originalSection?: string;
                revisedSection?: string;
                editedSection?: string;
                isEditing?: boolean;
                instructions?: string;
                showInstructions?: boolean;
                error?: string;
            }
        >
    >({});
```

**Step 3: Replace hooks**

Replace the hook calls (lines 203-204) with the new
per-finding hooks.

```typescript
    const generateItemRevision =
        useGenerateItemRevision(Number(campaignId));
    const applyItemRevision =
        useApplyItemRevision(Number(campaignId));
```

**Step 4: Add per-finding handlers**

Replace `handleGenerateRevision` and
`handleApplyRevision` (lines 346-381) with the following
per-finding handlers.

```typescript
    const handleGenerateItemRevision = (
        itemIds: number[],
    ) => {
        const key = itemIds[0];
        const instructions =
            revisionState[key]?.instructions;

        generateItemRevision.mutate(
            {
                itemIds,
                instructions: instructions || undefined,
            },
            {
                onSuccess: (data) => {
                    setRevisionState((prev) => ({
                        ...prev,
                        [key]: {
                            ...prev[key],
                            originalSection:
                                data.originalSection,
                            revisedSection:
                                data.revisedSection,
                            editedSection:
                                data.revisedSection,
                            isEditing: false,
                            error: undefined,
                        },
                    }));
                },
                onError: (err: Error) => {
                    setRevisionState((prev) => ({
                        ...prev,
                        [key]: {
                            ...prev[key],
                            error: err.message,
                        },
                    }));
                },
            },
        );
    };

    const handleApplyItemRevision = (
        itemIds: number[],
    ) => {
        const key = itemIds[0];
        const state = revisionState[key];
        if (!state?.editedSection) return;

        applyItemRevision.mutate(
            {
                itemIds,
                revisedSection: state.editedSection,
            },
            {
                onSuccess: () => {
                    setRevisionState((prev) => {
                        const next = { ...prev };
                        delete next[key];
                        return next;
                    });
                },
                onError: (err: Error) => {
                    setRevisionState((prev) => ({
                        ...prev,
                        [key]: {
                            ...prev[key],
                            error: err.message,
                        },
                    }));
                },
            },
        );
    };

    const handleDismissGroup = (itemIds: number[]) => {
        for (const id of itemIds) {
            handleResolve(id, 'dismissed');
        }
        const key = itemIds[0];
        setRevisionState((prev) => {
            const next = { ...prev };
            delete next[key];
            return next;
        });
    };
```

**Step 5: Remove old revision workflow UI**

Remove the entire `<Paper variant="outlined">` revision
workflow section at the top of the render (the block
from lines 388-583 that contains the "Generate Revision"
button, the diff view, and the "Apply Revision" button).

Replace it with a simple header.

```typescript
            {/* Per-finding revision header */}
            <Paper variant="outlined" sx={{ p: 2 }}>
                <Typography variant="h6">
                    Revise Findings
                </Typography>
                <Typography
                    variant="body2"
                    color="text.secondary"
                    sx={{ mt: 0.5 }}
                >
                    Generate targeted revisions for
                    individual findings. Click
                    &quot;Generate Suggestion&quot; on any
                    finding to create a surgical edit.
                </Typography>
            </Paper>
```

**Step 6: Add inline revision controls to finding cards**

Within the finding card rendering (inside the
`group.items.map` block, after the `ListItemButton` for
each item), add an inline revision panel that appears
when an item is acknowledged. Insert this block after
the closing `</ListItemButton>` tag for each item
(after line 779).

```typescript
                            {/* Inline revision panel */}
                            {item.resolution ===
                                'acknowledged' && (
                                <Box
                                    sx={{
                                        pl: 4,
                                        pr: 2,
                                        py: 1,
                                        bgcolor:
                                            'background.default',
                                        borderLeft: 3,
                                        borderColor:
                                            'success.main',
                                    }}
                                >
                                    {/* Instructions */}
                                    <Stack
                                        spacing={1}
                                        sx={{ mb: 1 }}
                                    >
                                        <Button
                                            size="small"
                                            variant="text"
                                            onClick={() =>
                                                setRevisionState(
                                                    (
                                                        prev,
                                                    ) => ({
                                                        ...prev,
                                                        [item.id]:
                                                            {
                                                                ...prev[
                                                                    item
                                                                        .id
                                                                ],
                                                                showInstructions:
                                                                    !prev[
                                                                        item
                                                                            .id
                                                                    ]
                                                                        ?.showInstructions,
                                                            },
                                                    }),
                                                )
                                            }
                                        >
                                            {revisionState[
                                                item.id
                                            ]
                                                ?.showInstructions
                                                ? 'Hide instructions'
                                                : 'Add instructions'}
                                        </Button>
                                        <Collapse
                                            in={
                                                revisionState[
                                                    item
                                                        .id
                                                ]
                                                    ?.showInstructions ??
                                                false
                                            }
                                        >
                                            <TextField
                                                size="small"
                                                fullWidth
                                                placeholder="e.g., make it darker, mention the assassinations"
                                                value={
                                                    revisionState[
                                                        item
                                                            .id
                                                    ]
                                                        ?.instructions ??
                                                    ''
                                                }
                                                onChange={(
                                                    e,
                                                ) =>
                                                    setRevisionState(
                                                        (
                                                            prev,
                                                        ) => ({
                                                            ...prev,
                                                            [item.id]:
                                                                {
                                                                    ...prev[
                                                                        item
                                                                            .id
                                                                    ],
                                                                    instructions:
                                                                        e
                                                                            .target
                                                                            .value,
                                                                },
                                                        }),
                                                    )
                                                }
                                            />
                                        </Collapse>
                                    </Stack>

                                    {/* Generate button */}
                                    {!revisionState[
                                        item.id
                                    ]
                                        ?.revisedSection && (
                                        <Button
                                            size="small"
                                            variant="contained"
                                            startIcon={
                                                generateItemRevision.isPending ? (
                                                    <CircularProgress
                                                        size={
                                                            14
                                                        }
                                                        color="inherit"
                                                    />
                                                ) : (
                                                    <AutoFixHigh />
                                                )
                                            }
                                            disabled={
                                                generateItemRevision.isPending
                                            }
                                            onClick={() =>
                                                handleGenerateItemRevision(
                                                    [
                                                        item.id,
                                                    ],
                                                )
                                            }
                                        >
                                            Generate
                                            Suggestion
                                        </Button>
                                    )}

                                    {/* Error */}
                                    {revisionState[
                                        item.id
                                    ]?.error && (
                                        <Alert
                                            severity="error"
                                            sx={{
                                                mt: 1,
                                            }}
                                            onClose={() =>
                                                setRevisionState(
                                                    (
                                                        prev,
                                                    ) => ({
                                                        ...prev,
                                                        [item.id]:
                                                            {
                                                                ...prev[
                                                                    item
                                                                        .id
                                                                ],
                                                                error: undefined,
                                                            },
                                                    }),
                                                )
                                            }
                                        >
                                            {
                                                revisionState[
                                                    item
                                                        .id
                                                ]?.error
                                            }
                                        </Alert>
                                    )}

                                    {/* Diff view */}
                                    {revisionState[
                                        item.id
                                    ]
                                        ?.revisedSection && (
                                        <Box
                                            sx={{
                                                mt: 1,
                                            }}
                                        >
                                            <Grid
                                                container
                                                spacing={
                                                    1
                                                }
                                            >
                                                <Grid
                                                    item
                                                    xs={
                                                        6
                                                    }
                                                >
                                                    <Paper
                                                        variant="outlined"
                                                        sx={{
                                                            p: 1,
                                                            bgcolor:
                                                                'error.50',
                                                            maxHeight: 200,
                                                            overflow:
                                                                'auto',
                                                        }}
                                                    >
                                                        <Typography
                                                            variant="caption"
                                                            color="text.secondary"
                                                        >
                                                            Original
                                                        </Typography>
                                                        <Typography
                                                            variant="body2"
                                                            sx={{
                                                                whiteSpace:
                                                                    'pre-wrap',
                                                                fontSize:
                                                                    '0.8rem',
                                                            }}
                                                        >
                                                            {
                                                                revisionState[
                                                                    item
                                                                        .id
                                                                ]
                                                                    ?.originalSection
                                                            }
                                                        </Typography>
                                                    </Paper>
                                                </Grid>
                                                <Grid
                                                    item
                                                    xs={
                                                        6
                                                    }
                                                >
                                                    <Paper
                                                        variant="outlined"
                                                        sx={{
                                                            p: 1,
                                                            bgcolor:
                                                                'success.50',
                                                            maxHeight: 200,
                                                            overflow:
                                                                'auto',
                                                        }}
                                                    >
                                                        <Typography
                                                            variant="caption"
                                                            color="text.secondary"
                                                        >
                                                            Revised
                                                        </Typography>
                                                        {revisionState[
                                                            item
                                                                .id
                                                        ]
                                                            ?.isEditing ? (
                                                            <TextField
                                                                multiline
                                                                fullWidth
                                                                size="small"
                                                                value={
                                                                    revisionState[
                                                                        item
                                                                            .id
                                                                    ]
                                                                        ?.editedSection ??
                                                                    ''
                                                                }
                                                                onChange={(
                                                                    e,
                                                                ) =>
                                                                    setRevisionState(
                                                                        (
                                                                            prev,
                                                                        ) => ({
                                                                            ...prev,
                                                                            [item.id]:
                                                                                {
                                                                                    ...prev[
                                                                                        item
                                                                                            .id
                                                                                    ],
                                                                                    editedSection:
                                                                                        e
                                                                                            .target
                                                                                            .value,
                                                                                },
                                                                        }),
                                                                    )
                                                                }
                                                                minRows={
                                                                    3
                                                                }
                                                                inputProps={{
                                                                    'aria-label':
                                                                        'Edit revised section',
                                                                    style: {
                                                                        fontSize:
                                                                            '0.8rem',
                                                                    },
                                                                }}
                                                            />
                                                        ) : (
                                                            <Typography
                                                                variant="body2"
                                                                sx={{
                                                                    whiteSpace:
                                                                        'pre-wrap',
                                                                    fontSize:
                                                                        '0.8rem',
                                                                }}
                                                            >
                                                                {
                                                                    revisionState[
                                                                        item
                                                                            .id
                                                                    ]
                                                                        ?.editedSection
                                                                }
                                                            </Typography>
                                                        )}
                                                    </Paper>
                                                </Grid>
                                            </Grid>

                                            {/* Action buttons */}
                                            <Stack
                                                direction="row"
                                                spacing={
                                                    1
                                                }
                                                sx={{
                                                    mt: 1,
                                                }}
                                            >
                                                <Button
                                                    size="small"
                                                    variant="contained"
                                                    color="success"
                                                    startIcon={
                                                        <Check />
                                                    }
                                                    disabled={
                                                        applyItemRevision.isPending
                                                    }
                                                    onClick={() =>
                                                        handleApplyItemRevision(
                                                            [
                                                                item.id,
                                                            ],
                                                        )
                                                    }
                                                >
                                                    Accept
                                                </Button>
                                                <Button
                                                    size="small"
                                                    variant="outlined"
                                                    startIcon={
                                                        <Edit />
                                                    }
                                                    onClick={() =>
                                                        setRevisionState(
                                                            (
                                                                prev,
                                                            ) => ({
                                                                ...prev,
                                                                [item.id]:
                                                                    {
                                                                        ...prev[
                                                                            item
                                                                                .id
                                                                        ],
                                                                        isEditing:
                                                                            !prev[
                                                                                item
                                                                                    .id
                                                                            ]
                                                                                ?.isEditing,
                                                                    },
                                                            }),
                                                        )
                                                    }
                                                >
                                                    {revisionState[
                                                        item
                                                            .id
                                                    ]
                                                        ?.isEditing
                                                        ? 'Preview'
                                                        : 'Edit'}
                                                </Button>
                                                <Button
                                                    size="small"
                                                    variant="outlined"
                                                    color="error"
                                                    startIcon={
                                                        <Close />
                                                    }
                                                    onClick={() =>
                                                        handleDismissGroup(
                                                            [
                                                                item.id,
                                                            ],
                                                        )
                                                    }
                                                >
                                                    Dismiss
                                                </Button>
                                            </Stack>
                                        </Box>
                                    )}
                                </Box>
                            )}
```

**Step 7: Verify TypeScript compiles**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client && \
    npx tsc --noEmit 2>&1 | tail -20
```

**Commit message:**
`feat: redesign RevisePhasePage with per-finding inline revision`

---

## Task 7: Deprecate Old Revision Endpoints

Mark the old full-document revision handlers as
deprecated and add log warnings. Remove the old revision
UI state that is no longer used.

**Files:**

- Modify: `internal/api/content_analysis_handler.go`
  (lines 1862-2209)
- Modify: `internal/api/router.go` (lines 243-244)

**Step 1: Add deprecation comments and log warnings**

In `internal/api/content_analysis_handler.go`, update
the `GenerateRevision` handler's doc comment (line 1876)
to include a deprecation notice.

```go
// GenerateRevision handles POST
// /api/campaigns/{id}/analysis/jobs/{jobId}/revision
//
// Deprecated: Use GenerateItemRevision instead. This
// endpoint generates a full-document revision using all
// acknowledged items and will be removed in a future
// release.
func (h *ContentAnalysisHandler) GenerateRevision(
	w http.ResponseWriter, r *http.Request,
) {
	log.Printf(
		"DEPRECATED: GenerateRevision called for "+
			"campaign %s, job %s. "+
			"Use POST /analysis/items/revision instead.",
		chi.URLParam(r, "id"),
		chi.URLParam(r, "jobId"),
	)
```

The rest of the handler body remains unchanged. Only the
doc comment and the new `log.Printf` at the start of the
function body are added.

Update the `ApplyRevision` handler similarly
(line 2069).

```go
// ApplyRevision handles PUT
// /api/campaigns/{id}/analysis/jobs/{jobId}/revision/apply
//
// Deprecated: Use ApplyItemRevision instead. This
// endpoint applies a full-document revision and will be
// removed in a future release.
func (h *ContentAnalysisHandler) ApplyRevision(
	w http.ResponseWriter, r *http.Request,
) {
	log.Printf(
		"DEPRECATED: ApplyRevision called for "+
			"campaign %s, job %s. "+
			"Use PUT /analysis/items/revision/apply instead.",
		chi.URLParam(r, "id"),
		chi.URLParam(r, "jobId"),
	)
```

**Step 2: Add deprecation comments to routes**

In `internal/api/router.go`, add comments to the old
routes (lines 243-244).

```go
						// Deprecated: use /analysis/items/revision instead.
						r.Post("/revision",
							contentAnalysisHandler.GenerateRevision)
						// Deprecated: use /analysis/items/revision/apply instead.
						r.Put("/revision/apply",
							contentAnalysisHandler.ApplyRevision)
```

**Step 3: Remove old revision UI state from
RevisePhasePage**

Verify that the old state variables (`revisionContent`,
`isEditing`, `revisionCount`, `revisionError`) were
already removed in Task 6. If any references to the old
`useGenerateRevision` or `useApplyRevision` hooks remain,
remove them now.

Also remove the import of `PlayArrow` from
`@mui/icons-material` since the old "Generate Revision"
button that used it has been removed.

**Step 4: Verify everything builds**

```bash
cd /Users/antonypegg/PROJECTS/imagineer && \
    go build ./...
```

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client && \
    npx tsc --noEmit
```

**Step 5: Run all tests**

```bash
cd /Users/antonypegg/PROJECTS/imagineer && \
    make test-all
```

**Commit message:**
`chore: deprecate old full-document revision endpoints`
