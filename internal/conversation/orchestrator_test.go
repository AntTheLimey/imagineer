/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package conversation

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/antonypegg/imagineer/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// collectEvents drains a stream event channel and
// returns all events as a slice.
func collectEvents(
	ch <-chan llm.StreamEvent,
) []llm.StreamEvent {
	var events []llm.StreamEvent
	for ev := range ch {
		events = append(events, ev)
	}
	return events
}

func TestNewOrchestrator(t *testing.T) {
	t.Run("default cache TTL", func(t *testing.T) {
		provider := &mockStreamingProvider{}
		orch := NewOrchestrator(OrchestratorConfig{
			Provider: provider,
		})
		defer orch.Stop()

		assert.NotNil(t, orch.cache,
			"cache should be created")
		assert.Nil(t, orch.tools,
			"tools should be nil when not provided")
		assert.Nil(t, orch.db,
			"db should be nil when not provided")
		assert.Empty(t, orch.promptPath,
			"promptPath should be empty when not provided")
	})

	t.Run("custom cache TTL", func(t *testing.T) {
		provider := &mockStreamingProvider{}
		orch := NewOrchestrator(OrchestratorConfig{
			Provider: provider,
			CacheTTL: 5 * time.Minute,
		})
		defer orch.Stop()

		assert.NotNil(t, orch.cache,
			"cache should be created with custom TTL")
	})

	t.Run("with tools", func(t *testing.T) {
		provider := &mockStreamingProvider{}
		registry := NewToolRegistry()
		orch := NewOrchestrator(OrchestratorConfig{
			Provider: provider,
			Tools:    registry,
		})
		defer orch.Stop()

		assert.Equal(t, registry, orch.tools,
			"tools should be set from config")
	})
}

func TestOrchestratorSimpleResponse(t *testing.T) {
	provider := &mockStreamingProvider{
		responses: [][]llm.StreamEvent{
			{
				{Type: llm.EventTextDelta,
					Text: "Hello, "},
				{Type: llm.EventTextDelta,
					Text: "how can I help?"},
				{Type: llm.EventUsage,
					Usage: &llm.TokenUsage{
						InputTokens:  50,
						OutputTokens: 10,
					}},
				{Type: llm.EventDone},
			},
		},
	}

	orch := NewOrchestrator(OrchestratorConfig{
		Provider: provider,
	})
	defer orch.Stop()

	ch, err := orch.HandleMessage(
		context.Background(),
		1, 1, 1,
		"Hi there",
	)
	require.NoError(t, err)

	events := collectEvents(ch)

	// Expect: text_delta, text_delta, done.
	var textDeltas []string
	var hasDone bool
	for _, ev := range events {
		switch ev.Type {
		case llm.EventTextDelta:
			textDeltas = append(textDeltas, ev.Text)
		case llm.EventDone:
			hasDone = true
		case llm.EventError:
			t.Fatalf("unexpected error: %v", ev.Error)
		}
	}

	assert.Equal(t,
		[]string{"Hello, ", "how can I help?"},
		textDeltas,
		"should receive text deltas in order")
	assert.True(t, hasDone,
		"should receive done event")
}

func TestOrchestratorToolUseLoop(t *testing.T) {
	provider := &mockStreamingProvider{
		responses: [][]llm.StreamEvent{
			// First call: tool use.
			{
				{Type: llm.EventTextDelta,
					Text: "Let me look that up. "},
				{Type: llm.EventToolUse,
					ToolName: "echo_tool",
					ToolID:   "call_001",
					ToolInput: json.RawMessage(
						`{"message":"test"}`)},
				{Type: llm.EventDone},
			},
			// Second call: final text.
			{
				{Type: llm.EventTextDelta,
					Text: "Based on the result, "},
				{Type: llm.EventTextDelta,
					Text: "here is the answer."},
				{Type: llm.EventDone},
			},
		},
	}

	registry := newTestSubTools()
	orch := NewOrchestrator(OrchestratorConfig{
		Provider: provider,
	})
	defer orch.Stop()
	orch.tools = registry

	ch, err := orch.HandleMessage(
		context.Background(),
		1, 1, 1,
		"Look something up",
	)
	require.NoError(t, err)

	events := collectEvents(ch)

	var textDeltas []string
	var toolUseEvents []llm.StreamEvent
	var hasDone bool

	for _, ev := range events {
		switch ev.Type {
		case llm.EventTextDelta:
			textDeltas = append(textDeltas, ev.Text)
		case llm.EventToolUse:
			toolUseEvents = append(toolUseEvents, ev)
		case llm.EventDone:
			hasDone = true
		case llm.EventError:
			t.Fatalf("unexpected error: %v", ev.Error)
		}
	}

	// Should see text deltas from both LLM calls.
	assert.Contains(t, textDeltas,
		"Let me look that up. ",
		"should see first text delta")
	assert.Contains(t, textDeltas,
		"Based on the result, ",
		"should see second text delta")
	assert.Contains(t, textDeltas,
		"here is the answer.",
		"should see final text delta")

	// Should see tool use event forwarded.
	require.Len(t, toolUseEvents, 1,
		"should see one tool use event")
	assert.Equal(t, "echo_tool",
		toolUseEvents[0].ToolName)
	assert.Equal(t, "call_001",
		toolUseEvents[0].ToolID)

	assert.True(t, hasDone,
		"should receive done event")

	// Provider should have been called twice.
	assert.Equal(t, 2, provider.callIndex,
		"provider should be called twice")
}

func TestOrchestratorMaxIterations(t *testing.T) {
	// Mock provider always returns tool calls.
	responses := make([][]llm.StreamEvent,
		maxOrchestratorIterations)
	for i := range responses {
		responses[i] = []llm.StreamEvent{
			{Type: llm.EventToolUse,
				ToolName: "echo_tool",
				ToolID: fmt.Sprintf(
					"call_%03d", i),
				ToolInput: json.RawMessage(
					`{"message":"loop"}`)},
			{Type: llm.EventDone},
		}
	}

	provider := &mockStreamingProvider{
		responses: responses,
	}

	registry := newTestSubTools()
	orch := NewOrchestrator(OrchestratorConfig{
		Provider: provider,
	})
	defer orch.Stop()
	orch.tools = registry

	ch, err := orch.HandleMessage(
		context.Background(),
		1, 1, 1,
		"Keep looping",
	)
	require.NoError(t, err)

	events := collectEvents(ch)

	// Should end with an error event.
	var lastError error
	for _, ev := range events {
		if ev.Type == llm.EventError {
			lastError = ev.Error
		}
	}

	require.Error(t, lastError,
		"should receive error event")
	assert.Contains(t, lastError.Error(),
		"exceeded maximum iterations",
		"error should mention iteration limit")
}

func TestOrchestratorStreamError(t *testing.T) {
	provider := &mockStreamingProvider{
		responses: [][]llm.StreamEvent{
			{
				{Type: llm.EventTextDelta,
					Text: "Starting..."},
				{Type: llm.EventError,
					Error: fmt.Errorf(
						"rate limit exceeded")},
			},
		},
	}

	orch := NewOrchestrator(OrchestratorConfig{
		Provider: provider,
	})
	defer orch.Stop()

	ch, err := orch.HandleMessage(
		context.Background(),
		1, 1, 1,
		"Will this fail?",
	)
	require.NoError(t, err)

	events := collectEvents(ch)

	var gotText bool
	var gotError bool
	var errorMsg string

	for _, ev := range events {
		switch ev.Type {
		case llm.EventTextDelta:
			gotText = true
		case llm.EventError:
			gotError = true
			errorMsg = ev.Error.Error()
		}
	}

	assert.True(t, gotText,
		"should receive text before error")
	assert.True(t, gotError,
		"should receive error event")
	assert.Contains(t, errorMsg,
		"rate limit exceeded",
		"error should be forwarded")
}

func TestOrchestratorProviderError(t *testing.T) {
	// Mock provider has no responses, so
	// CompleteStream returns an error.
	provider := &mockStreamingProvider{
		responses: nil,
	}

	orch := NewOrchestrator(OrchestratorConfig{
		Provider: provider,
	})
	defer orch.Stop()

	ch, err := orch.HandleMessage(
		context.Background(),
		1, 1, 1,
		"Will this fail?",
	)
	require.NoError(t, err)

	events := collectEvents(ch)

	require.NotEmpty(t, events,
		"should receive at least one event")

	var foundError bool
	for _, ev := range events {
		if ev.Type == llm.EventError {
			foundError = true
			assert.Contains(t, ev.Error.Error(),
				"LLM stream failed",
				"error should mention stream failure")
		}
	}

	assert.True(t, foundError,
		"should receive error event")
}

func TestOrchestratorContextCancellation(t *testing.T) {
	// Create a provider that blocks until context
	// is cancelled.
	blockingProvider := &blockingStreamingProvider{
		ready: make(chan struct{}),
	}

	orch := NewOrchestrator(OrchestratorConfig{
		Provider: blockingProvider,
	})
	defer orch.Stop()

	ctx, cancel := context.WithCancel(
		context.Background())

	ch, err := orch.HandleMessage(
		ctx, 1, 1, 1, "Cancel me",
	)
	require.NoError(t, err)

	// Wait for the provider to be called.
	<-blockingProvider.ready

	// Cancel the context.
	cancel()

	events := collectEvents(ch)

	// Should get an error about cancellation.
	var gotError bool
	for _, ev := range events {
		if ev.Type == llm.EventError {
			gotError = true
		}
	}
	assert.True(t, gotError,
		"should receive error on cancellation")
}

// blockingStreamingProvider blocks CompleteStream
// until the context is cancelled. It signals readiness
// via the ready channel.
type blockingStreamingProvider struct {
	ready chan struct{}
}

func (b *blockingStreamingProvider) CompleteStream(
	ctx context.Context,
	req llm.StreamingRequest,
) (<-chan llm.StreamEvent, error) {
	close(b.ready)
	<-ctx.Done()
	return nil, ctx.Err()
}

func TestMessagesToStreaming(t *testing.T) {
	toolName := "search_entities"
	toolID := "call_42"
	toolInput := json.RawMessage(`{"query":"test"}`)
	toolResult := json.RawMessage(`{"results":[]}`)

	msgs := []models.Message{
		{
			ID:             1,
			ConversationID: 10,
			Role:           models.MessageRoleUser,
			Content:        "Hello",
		},
		{
			ID:             2,
			ConversationID: 10,
			Role:           models.MessageRoleAssistant,
			Content:        "Hi there!",
		},
		{
			ID:             3,
			ConversationID: 10,
			Role:           models.MessageRoleToolCall,
			Content:        "",
			ToolName:       &toolName,
			ToolUseID:      &toolID,
			ToolInput:      toolInput,
		},
		{
			ID:             4,
			ConversationID: 10,
			Role:           models.MessageRoleToolResult,
			Content:        "",
			ToolName:       &toolName,
			ToolUseID:      &toolID,
			ToolResult:     toolResult,
		},
	}

	result := messagesToStreaming(msgs)

	require.Len(t, result, 4,
		"should convert all messages")

	// User message.
	assert.Equal(t, "user", result[0].Role)
	assert.Equal(t, "Hello", result[0].Content)
	assert.Empty(t, result[0].ToolName)

	// Assistant message.
	assert.Equal(t, "assistant", result[1].Role)
	assert.Equal(t, "Hi there!", result[1].Content)

	// Tool call message.
	assert.Equal(t, "tool_call", result[2].Role)
	assert.Equal(t, "search_entities",
		result[2].ToolName)
	assert.Equal(t, "call_42", result[2].ToolUseID)
	assert.JSONEq(t, `{"query":"test"}`,
		string(result[2].ToolInput))
	assert.Nil(t, result[2].ToolResult)

	// Tool result message.
	assert.Equal(t, "tool_result", result[3].Role)
	assert.Equal(t, "search_entities",
		result[3].ToolName)
	assert.Equal(t, "call_42", result[3].ToolUseID)
	assert.JSONEq(t, `{"results":[]}`,
		string(result[3].ToolResult))
}

func TestMessagesToStreamingEmpty(t *testing.T) {
	result := messagesToStreaming(nil)
	assert.Empty(t, result,
		"nil input should return empty slice")

	result = messagesToStreaming([]models.Message{})
	assert.Empty(t, result,
		"empty input should return empty slice")
}

func TestMessagesToStreamingNilOptionalFields(
	t *testing.T,
) {
	msgs := []models.Message{
		{
			Role:    models.MessageRoleUser,
			Content: "Test",
		},
	}

	result := messagesToStreaming(msgs)
	require.Len(t, result, 1)
	assert.Empty(t, result[0].ToolName,
		"nil ToolName should become empty string")
	assert.Empty(t, result[0].ToolUseID,
		"nil ToolUseID should become empty string")
	assert.Nil(t, result[0].ToolInput,
		"nil ToolInput should remain nil")
	assert.Nil(t, result[0].ToolResult,
		"nil ToolResult should remain nil")
}

func TestOrchestratorCacheUpdated(t *testing.T) {
	provider := &mockStreamingProvider{
		responses: [][]llm.StreamEvent{
			{
				{Type: llm.EventTextDelta,
					Text: "Cached response"},
				{Type: llm.EventDone},
			},
		},
	}

	orch := NewOrchestrator(OrchestratorConfig{
		Provider: provider,
	})
	defer orch.Stop()

	ch, err := orch.HandleMessage(
		context.Background(),
		42, 1, 1,
		"Cache this",
	)
	require.NoError(t, err)

	// Drain events.
	_ = collectEvents(ch)

	// Verify cache was updated.
	entry, ok := orch.cache.Get(42)
	require.True(t, ok,
		"cache should have entry for conversation 42")
	require.NotEmpty(t, entry.Messages,
		"cached messages should not be empty")

	// Last message should be the assistant response.
	lastMsg := entry.Messages[len(entry.Messages)-1]
	assert.Equal(t, "assistant", lastMsg.Role)
	assert.Equal(t, "Cached response", lastMsg.Content)
}

func TestOrchestratorUsesCache(t *testing.T) {
	callCount := 0
	provider := &trackingStreamingProvider{
		inner: &mockStreamingProvider{
			responses: [][]llm.StreamEvent{
				{
					{Type: llm.EventTextDelta,
						Text: "First response"},
					{Type: llm.EventDone},
				},
				{
					{Type: llm.EventTextDelta,
						Text: "Second response"},
					{Type: llm.EventDone},
				},
			},
		},
		onCall: func(req llm.StreamingRequest) {
			callCount++
		},
	}

	orch := NewOrchestrator(OrchestratorConfig{
		Provider: provider,
	})
	defer orch.Stop()

	// First message.
	ch, err := orch.HandleMessage(
		context.Background(),
		42, 1, 1,
		"First message",
	)
	require.NoError(t, err)
	_ = collectEvents(ch)

	assert.Equal(t, 1, callCount,
		"first call should invoke provider once")

	// Second message -- should use cached context.
	ch, err = orch.HandleMessage(
		context.Background(),
		42, 1, 1,
		"Second message",
	)
	require.NoError(t, err)
	_ = collectEvents(ch)

	assert.Equal(t, 2, callCount,
		"second call should invoke provider again")

	// Verify the cache has both messages.
	entry, ok := orch.cache.Get(42)
	require.True(t, ok)
	// Should have: user1, assistant1, user2, assistant2.
	assert.Len(t, entry.Messages, 4,
		"cache should have 4 messages after two turns")
}

// trackingStreamingProvider wraps a provider and
// calls onCall for each CompleteStream invocation.
type trackingStreamingProvider struct {
	inner  llm.StreamingProvider
	onCall func(req llm.StreamingRequest)
}

func (t *trackingStreamingProvider) CompleteStream(
	ctx context.Context,
	req llm.StreamingRequest,
) (<-chan llm.StreamEvent, error) {
	if t.onCall != nil {
		t.onCall(req)
	}
	return t.inner.CompleteStream(ctx, req)
}

func TestOrchestratorMultipleToolCalls(t *testing.T) {
	provider := &mockStreamingProvider{
		responses: [][]llm.StreamEvent{
			// First call: two tool calls.
			{
				{Type: llm.EventToolUse,
					ToolName: "echo_tool",
					ToolID:   "call_001",
					ToolInput: json.RawMessage(
						`{"message":"first"}`)},
				{Type: llm.EventToolUse,
					ToolName: "echo_tool",
					ToolID:   "call_002",
					ToolInput: json.RawMessage(
						`{"message":"second"}`)},
				{Type: llm.EventDone},
			},
			// Second call: final text.
			{
				{Type: llm.EventTextDelta,
					Text: "Both tools succeeded."},
				{Type: llm.EventDone},
			},
		},
	}

	registry := newTestSubTools()
	orch := NewOrchestrator(OrchestratorConfig{
		Provider: provider,
	})
	defer orch.Stop()
	orch.tools = registry

	ch, err := orch.HandleMessage(
		context.Background(),
		1, 1, 1,
		"Use two tools",
	)
	require.NoError(t, err)

	events := collectEvents(ch)

	var toolUseCount int
	var hasFinalText bool

	for _, ev := range events {
		switch ev.Type {
		case llm.EventToolUse:
			toolUseCount++
		case llm.EventTextDelta:
			if ev.Text == "Both tools succeeded." {
				hasFinalText = true
			}
		}
	}

	assert.Equal(t, 2, toolUseCount,
		"should see two tool use events")
	assert.True(t, hasFinalText,
		"should see final text response")
}

func TestOrchestratorToolExecutionError(t *testing.T) {
	provider := &mockStreamingProvider{
		responses: [][]llm.StreamEvent{
			// Tool call to failing tool.
			{
				{Type: llm.EventToolUse,
					ToolName:  "failing_tool",
					ToolID:    "call_fail",
					ToolInput: json.RawMessage(`{}`)},
				{Type: llm.EventDone},
			},
			// LLM recovers after error result.
			{
				{Type: llm.EventTextDelta,
					Text: "The tool failed."},
				{Type: llm.EventDone},
			},
		},
	}

	registry := NewToolRegistry()
	registry.Register(Tool{
		Definition: llm.ToolDefinition{
			Name:        "failing_tool",
			Description: "Always fails",
			InputSchema: json.RawMessage(
				`{"type":"object","properties":{}}`),
		},
		Execute: func(
			ctx context.Context,
			input json.RawMessage,
		) (json.RawMessage, error) {
			return nil, fmt.Errorf("database unavailable")
		},
	})

	orch := NewOrchestrator(OrchestratorConfig{
		Provider: provider,
	})
	defer orch.Stop()
	orch.tools = registry

	ch, err := orch.HandleMessage(
		context.Background(),
		1, 1, 1,
		"Try the failing tool",
	)
	require.NoError(t, err)

	events := collectEvents(ch)

	var hasFinalText bool
	var hasError bool

	for _, ev := range events {
		switch ev.Type {
		case llm.EventTextDelta:
			if ev.Text == "The tool failed." {
				hasFinalText = true
			}
		case llm.EventError:
			hasError = true
		}
	}

	assert.True(t, hasFinalText,
		"LLM should recover from tool error")
	assert.False(t, hasError,
		"tool execution error should not be a "+
			"stream error")
}

func TestOrchestratorStop(t *testing.T) {
	orch := NewOrchestrator(OrchestratorConfig{
		Provider: &mockStreamingProvider{},
	})

	// Should not panic when called multiple times.
	orch.Stop()
	orch.Stop()
}
