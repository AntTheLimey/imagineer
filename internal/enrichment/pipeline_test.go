/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package enrichment

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/antonypegg/imagineer/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mock pipeline agent
// ---------------------------------------------------------------------------

// mockPipelineAgent implements PipelineAgent for testing without a real
// LLM provider or database.
type mockPipelineAgent struct {
	name      string
	deps      []string
	items     []models.ContentAnalysisItem
	err       error
	called    bool
	callCount int
}

func (m *mockPipelineAgent) Name() string        { return m.name }
func (m *mockPipelineAgent) DependsOn() []string { return m.deps }

func (m *mockPipelineAgent) Run(
	ctx context.Context,
	provider llm.Provider,
	input PipelineInput,
) ([]models.ContentAnalysisItem, error) {
	m.called = true
	m.callCount++
	if m.err != nil {
		return nil, m.err
	}
	return m.items, nil
}

// ---------------------------------------------------------------------------
// Pipeline.Run tests
// ---------------------------------------------------------------------------

func TestPipelineRun_SingleAgent(t *testing.T) {
	agent := &mockPipelineAgent{
		name: "entity-detector",
		items: []models.ContentAnalysisItem{
			{MatchedText: "Nyarlathotep", DetectionType: "entity_mention"},
			{MatchedText: "Arkham", DetectionType: "entity_mention"},
		},
	}

	pipeline := NewPipeline(nil, []Stage{
		{Name: "analysis", Phase: "analysis", Agents: []PipelineAgent{agent}},
	})

	items, err := pipeline.Run(context.Background(), nil, PipelineInput{
		CampaignID: 1,
		Content:    "Nyarlathotep was spotted in Arkham.",
	})

	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.True(t, agent.called)
	assert.Equal(t, 1, agent.callCount)
	assert.Equal(t, "Nyarlathotep", items[0].MatchedText)
	assert.Equal(t, "Arkham", items[1].MatchedText)
}

func TestPipelineRun_MultipleStages(t *testing.T) {
	agentA := &mockPipelineAgent{
		name: "stage-one-agent",
		items: []models.ContentAnalysisItem{
			{MatchedText: "item-from-stage-one"},
		},
	}
	agentB := &mockPipelineAgent{
		name: "stage-two-agent",
		items: []models.ContentAnalysisItem{
			{MatchedText: "item-from-stage-two"},
		},
	}

	pipeline := NewPipeline(nil, []Stage{
		{Name: "analysis", Phase: "analysis", Agents: []PipelineAgent{agentA}},
		{Name: "enrichment", Phase: "enrichment", Agents: []PipelineAgent{agentB}},
	})

	items, err := pipeline.Run(context.Background(), nil, PipelineInput{
		CampaignID: 1,
		Content:    "test content",
	})

	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.True(t, agentA.called)
	assert.True(t, agentB.called)

	// Items from stage one should appear before items from stage two.
	assert.Equal(t, "item-from-stage-one", items[0].MatchedText)
	assert.Equal(t, "item-from-stage-two", items[1].MatchedText)
}

func TestPipelineRun_AgentFailure(t *testing.T) {
	failingAgent := &mockPipelineAgent{
		name: "failing-agent",
		err:  errors.New("LLM unavailable"),
	}
	healthyAgent := &mockPipelineAgent{
		name: "healthy-agent",
		items: []models.ContentAnalysisItem{
			{MatchedText: "survived"},
		},
	}

	pipeline := NewPipeline(nil, []Stage{
		{
			Name:  "analysis",
			Phase: "analysis",
			Agents: []PipelineAgent{
				failingAgent,
				healthyAgent,
			},
		},
	})

	items, err := pipeline.Run(context.Background(), nil, PipelineInput{
		CampaignID: 1,
		Content:    "test content",
	})

	require.NoError(t, err, "pipeline should not return an error on agent failure")
	assert.Len(t, items, 1)
	assert.Equal(t, "survived", items[0].MatchedText)
	assert.True(t, failingAgent.called, "failing agent should still have been called")
	assert.True(t, healthyAgent.called, "healthy agent should have been called despite prior failure")
}

func TestPipelineRun_EmptyPipeline(t *testing.T) {
	pipeline := NewPipeline(nil, []Stage{})

	items, err := pipeline.Run(context.Background(), nil, PipelineInput{
		CampaignID: 1,
		Content:    "test content",
	})

	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestPipelineRun_EmptyStage(t *testing.T) {
	pipeline := NewPipeline(nil, []Stage{
		{Name: "empty-stage", Phase: "analysis", Agents: []PipelineAgent{}},
	})

	items, err := pipeline.Run(context.Background(), nil, PipelineInput{
		CampaignID: 1,
		Content:    "test content",
	})

	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestPipelineRun_AgentNameTagging(t *testing.T) {
	agent := &mockPipelineAgent{
		name: "relationship-detector",
		items: []models.ContentAnalysisItem{
			{MatchedText: "item-one"},
			{MatchedText: "item-two"},
			{MatchedText: "item-three"},
		},
	}

	pipeline := NewPipeline(nil, []Stage{
		{Name: "analysis", Phase: "analysis", Agents: []PipelineAgent{agent}},
	})

	items, err := pipeline.Run(context.Background(), nil, PipelineInput{
		CampaignID: 1,
		Content:    "test content",
	})

	require.NoError(t, err)
	require.Len(t, items, 3)

	for _, item := range items {
		assert.Equal(t, "relationship-detector", item.AgentName,
			"every item should be tagged with the producing agent's name")
	}
}

func TestPipelineRun_AgentNameTaggingMultipleAgents(t *testing.T) {
	agentA := &mockPipelineAgent{
		name: "agent-alpha",
		items: []models.ContentAnalysisItem{
			{MatchedText: "from-alpha"},
		},
	}
	agentB := &mockPipelineAgent{
		name: "agent-beta",
		items: []models.ContentAnalysisItem{
			{MatchedText: "from-beta"},
		},
	}

	pipeline := NewPipeline(nil, []Stage{
		{
			Name:   "analysis",
			Phase:  "analysis",
			Agents: []PipelineAgent{agentA, agentB},
		},
	})

	items, err := pipeline.Run(context.Background(), nil, PipelineInput{
		CampaignID: 1,
		Content:    "test content",
	})

	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, "agent-alpha", items[0].AgentName)
	assert.Equal(t, "agent-beta", items[1].AgentName)
}

func TestPipelineRun_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	secondAgent := &mockPipelineAgent{
		name: "after-cancel",
		items: []models.ContentAnalysisItem{
			{MatchedText: "should-not-appear"},
		},
	}

	// Use a functional mock to cancel context mid-pipeline.
	canceller := &contextCancellerAgent{
		name:     "canceller",
		cancelFn: cancel,
		items: []models.ContentAnalysisItem{
			{MatchedText: "before-cancel"},
		},
	}

	// The pipeline runs agents sequentially. After the canceller runs
	// and cancels the context, the second agent's Run call will receive
	// a cancelled context. However, the pipeline itself does not check
	// ctx.Err() between agents -- it is up to each agent to respect
	// context cancellation. Our mock agents do not check context, so
	// both agents will run. This test verifies that the pipeline does
	// not panic and returns results from all agents that completed.
	pipeline := NewPipeline(nil, []Stage{
		{
			Name:  "analysis",
			Phase: "analysis",
			Agents: []PipelineAgent{
				canceller,
				secondAgent,
			},
		},
	})

	items, err := pipeline.Run(ctx, nil, PipelineInput{
		CampaignID: 1,
		Content:    "test content",
	})

	require.NoError(t, err)
	// The canceller's items are returned. The second agent also runs
	// because Pipeline.Run does not check ctx between agent calls.
	assert.True(t, canceller.called)
	assert.True(t, secondAgent.called)
	assert.NotEmpty(t, items)
	assert.Equal(t, "before-cancel", items[0].MatchedText)
}

// contextCancellerAgent is a PipelineAgent that cancels a context during
// its Run method, used to test pipeline behaviour under cancellation.
type contextCancellerAgent struct {
	name     string
	cancelFn context.CancelFunc
	items    []models.ContentAnalysisItem
	called   bool
}

func (a *contextCancellerAgent) Name() string        { return a.name }
func (a *contextCancellerAgent) DependsOn() []string { return nil }

func (a *contextCancellerAgent) Run(
	ctx context.Context,
	provider llm.Provider,
	input PipelineInput,
) ([]models.ContentAnalysisItem, error) {
	a.called = true
	a.cancelFn()
	return a.items, nil
}

// ---------------------------------------------------------------------------
// topologicalSort tests
// ---------------------------------------------------------------------------

func TestTopologicalSort_NoDependencies(t *testing.T) {
	agents := []PipelineAgent{
		&mockPipelineAgent{name: "alpha"},
		&mockPipelineAgent{name: "beta"},
		&mockPipelineAgent{name: "gamma"},
	}

	sorted, skipped := topologicalSort(agents)

	assert.Empty(t, skipped)
	require.Len(t, sorted, 3)

	// With no dependencies, agents should appear in their original order.
	names := make([]string, len(sorted))
	for i, a := range sorted {
		names[i] = a.Name()
	}
	assert.Equal(t, []string{"alpha", "beta", "gamma"}, names)
}

func TestTopologicalSort_LinearDeps(t *testing.T) {
	// A depends on B, B depends on C. Expected order: C, B, A.
	agentC := &mockPipelineAgent{name: "C"}
	agentB := &mockPipelineAgent{name: "B", deps: []string{"C"}}
	agentA := &mockPipelineAgent{name: "A", deps: []string{"B"}}

	sorted, skipped := topologicalSort([]PipelineAgent{agentA, agentB, agentC})

	assert.Empty(t, skipped)
	require.Len(t, sorted, 3)

	names := make([]string, len(sorted))
	for i, a := range sorted {
		names[i] = a.Name()
	}
	assert.Equal(t, []string{"C", "B", "A"}, names)
}

func TestTopologicalSort_DiamondDeps(t *testing.T) {
	// D depends on B and C; B depends on A; C depends on A.
	// Expected: A first, then B and C (in input order), then D.
	agentA := &mockPipelineAgent{name: "A"}
	agentB := &mockPipelineAgent{name: "B", deps: []string{"A"}}
	agentC := &mockPipelineAgent{name: "C", deps: []string{"A"}}
	agentD := &mockPipelineAgent{name: "D", deps: []string{"B", "C"}}

	sorted, skipped := topologicalSort([]PipelineAgent{
		agentD, agentB, agentC, agentA,
	})

	assert.Empty(t, skipped)
	require.Len(t, sorted, 4)

	names := make([]string, len(sorted))
	for i, a := range sorted {
		names[i] = a.Name()
	}

	// A must come before B and C; B and C must come before D.
	posA := indexOf(names, "A")
	posB := indexOf(names, "B")
	posC := indexOf(names, "C")
	posD := indexOf(names, "D")

	assert.Less(t, posA, posB, "A should come before B")
	assert.Less(t, posA, posC, "A should come before C")
	assert.Less(t, posB, posD, "B should come before D")
	assert.Less(t, posC, posD, "C should come before D")
}

func TestTopologicalSort_CircularDeps(t *testing.T) {
	// A depends on B, B depends on A -- both should be skipped.
	agentA := &mockPipelineAgent{name: "A", deps: []string{"B"}}
	agentB := &mockPipelineAgent{name: "B", deps: []string{"A"}}
	agentC := &mockPipelineAgent{name: "C"} // no cycle

	sorted, skipped := topologicalSort([]PipelineAgent{agentA, agentB, agentC})

	assert.Len(t, skipped, 2)
	assert.Contains(t, skipped, "A")
	assert.Contains(t, skipped, "B")

	require.Len(t, sorted, 1)
	assert.Equal(t, "C", sorted[0].Name())
}

func TestTopologicalSort_SelfDependency(t *testing.T) {
	// An agent that depends on itself is a trivial cycle.
	agent := &mockPipelineAgent{name: "narcissist", deps: []string{"narcissist"}}

	sorted, skipped := topologicalSort([]PipelineAgent{agent})

	assert.Len(t, skipped, 1)
	assert.Contains(t, skipped, "narcissist")
	assert.Empty(t, sorted)
}

func TestTopologicalSort_ExternalDependency(t *testing.T) {
	// Agent depends on "prior-stage-agent" which is not in this stage.
	// The dependency should be treated as satisfied.
	agent := &mockPipelineAgent{
		name: "enricher",
		deps: []string{"prior-stage-agent"},
	}

	sorted, skipped := topologicalSort([]PipelineAgent{agent})

	assert.Empty(t, skipped)
	require.Len(t, sorted, 1)
	assert.Equal(t, "enricher", sorted[0].Name())
}

func TestTopologicalSort_MixedExternalAndInternalDeps(t *testing.T) {
	// Agent A depends on B (in stage) and X (external).
	agentB := &mockPipelineAgent{name: "B"}
	agentA := &mockPipelineAgent{name: "A", deps: []string{"B", "X"}}

	sorted, skipped := topologicalSort([]PipelineAgent{agentA, agentB})

	assert.Empty(t, skipped)
	require.Len(t, sorted, 2)

	names := make([]string, len(sorted))
	for i, a := range sorted {
		names[i] = a.Name()
	}
	assert.Equal(t, []string{"B", "A"}, names)
}

func TestTopologicalSort_EmptyInput(t *testing.T) {
	sorted, skipped := topologicalSort([]PipelineAgent{})

	assert.Empty(t, sorted)
	assert.Empty(t, skipped)
}

// ---------------------------------------------------------------------------
// truncateQuery tests
// ---------------------------------------------------------------------------

func TestContextBuilder_TruncateQuery_Short(t *testing.T) {
	input := "A short query."
	result := truncateQuery(input)
	assert.Equal(t, input, result)
}

func TestContextBuilder_TruncateQuery_ExactLimit(t *testing.T) {
	// Build a string of exactly maxSearchQueryLen runes.
	input := strings.Repeat("a", maxSearchQueryLen)
	result := truncateQuery(input)
	assert.Equal(t, input, result)
	assert.Len(t, []rune(result), maxSearchQueryLen)
}

func TestContextBuilder_TruncateQuery_OverLimit(t *testing.T) {
	input := strings.Repeat("b", maxSearchQueryLen+50)
	result := truncateQuery(input)
	assert.Len(t, []rune(result), maxSearchQueryLen)
	assert.Equal(t, strings.Repeat("b", maxSearchQueryLen), result)
}

func TestContextBuilder_TruncateQuery_Empty(t *testing.T) {
	result := truncateQuery("")
	assert.Equal(t, "", result)
}

func TestContextBuilder_TruncateQuery_MultiByte(t *testing.T) {
	// Use multi-byte characters to verify rune-based (not byte-based)
	// truncation. Each character here is 3 bytes in UTF-8.
	input := strings.Repeat("\u4e16", maxSearchQueryLen+10) // Chinese character
	result := truncateQuery(input)
	assert.Len(t, []rune(result), maxSearchQueryLen)
	assert.Equal(t, strings.Repeat("\u4e16", maxSearchQueryLen), result)
}

// ---------------------------------------------------------------------------
// Pipeline.String test
// ---------------------------------------------------------------------------

func TestPipelineString(t *testing.T) {
	pipeline := NewPipeline(nil, []Stage{
		{
			Name: "stage-1",
			Agents: []PipelineAgent{
				&mockPipelineAgent{name: "a"},
				&mockPipelineAgent{name: "b"},
			},
		},
		{
			Name: "stage-2",
			Agents: []PipelineAgent{
				&mockPipelineAgent{name: "c"},
			},
		},
	})

	s := pipeline.String()
	assert.Equal(t, "Pipeline(2 stages, 3 agents)", s)
}

// ---------------------------------------------------------------------------
// ScopeFromSourceTable tests
// ---------------------------------------------------------------------------

func TestScopeFromSourceTable(t *testing.T) {
	tests := []struct {
		table    string
		expected SourceScope
	}{
		{"campaigns", ScopeCampaign},
		{"chapters", ScopeChapter},
		{"sessions", ScopeSession},
		{"entities", ScopeEntity},
		{"unknown", ScopeCampaign},
		{"", ScopeCampaign},
	}
	for _, tc := range tests {
		t.Run(tc.table, func(t *testing.T) {
			assert.Equal(t, tc.expected, ScopeFromSourceTable(tc.table))
		})
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// indexOf returns the position of target in slice, or -1 if not found.
func indexOf(slice []string, target string) int {
	for i, v := range slice {
		if v == target {
			return i
		}
	}
	return -1
}
