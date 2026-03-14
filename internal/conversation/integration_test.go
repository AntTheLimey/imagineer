//go:build integration

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
	"sync/atomic"
	"testing"

	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConversationEndToEnd simulates a full conversation
// where the LLM first calls a tool, receives the result,
// and then produces a final text response. This exercises
// the complete orchestrator tool-use loop end-to-end
// with a mock provider but no database.
func TestConversationEndToEnd(t *testing.T) {
	provider := &mockStreamingProvider{
		responses: [][]llm.StreamEvent{
			// First call: tool use.
			{
				{Type: llm.EventToolUse,
					ToolName: "search_entities",
					ToolID:   "call_001",
					ToolInput: json.RawMessage(
						`{"query":"NPCs near the inn"}`)},
				{Type: llm.EventUsage,
					Usage: &llm.TokenUsage{
						InputTokens:  500,
						OutputTokens: 50,
					}},
				{Type: llm.EventDone},
			},
			// Second call: text response after
			// receiving tool result.
			{
				{Type: llm.EventTextDelta,
					Text: "I found several NPCs near "},
				{Type: llm.EventTextDelta,
					Text: "the inn. "},
				{Type: llm.EventTextDelta,
					Text: "The innkeeper is Greta."},
				{Type: llm.EventUsage,
					Usage: &llm.TokenUsage{
						InputTokens:  800,
						OutputTokens: 100,
					}},
				{Type: llm.EventDone},
			},
		},
	}

	registry := NewToolRegistry()
	registry.Register(Tool{
		Definition: llm.ToolDefinition{
			Name:        "search_entities",
			Description: "Search entities",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"query": {"type": "string"}
				},
				"required": ["query"]
			}`),
		},
		Execute: func(
			ctx context.Context,
			input json.RawMessage,
		) (json.RawMessage, error) {
			return json.RawMessage(`[
				{"id": 1, "name": "Greta",
				 "entity_type": "npc",
				 "description": "The innkeeper"}
			]`), nil
		},
	})

	orch := NewOrchestrator(OrchestratorConfig{
		Provider: provider,
		Tools:    registry,
	})
	defer orch.Stop()

	ctx := context.Background()
	ch, err := orch.HandleMessage(
		ctx, 1, 1, 1,
		"Find NPCs near the inn")
	require.NoError(t, err)

	events := collectEvents(ch)

	var hasToolUse, hasTextDelta, hasDone bool
	var accumulatedText string

	for _, ev := range events {
		switch ev.Type {
		case llm.EventToolUse:
			hasToolUse = true
			assert.Equal(t, "search_entities",
				ev.ToolName)
		case llm.EventTextDelta:
			hasTextDelta = true
			accumulatedText += ev.Text
		case llm.EventDone:
			hasDone = true
		case llm.EventError:
			t.Fatalf("unexpected error event: %v",
				ev.Error)
		}
	}

	assert.True(t, hasToolUse,
		"should have tool_use events")
	assert.True(t, hasTextDelta,
		"should have text_delta events")
	assert.True(t, hasDone,
		"should have done event")
	assert.Contains(t, accumulatedText, "Greta",
		"response should mention the entity found")

	// Verify provider was called exactly 2 times:
	// once for the tool call, once for the final
	// response.
	assert.Equal(t, 2, provider.callIndex,
		"provider should be called twice: "+
			"tool use + final")
}

// TestConversationEndToEndNoToolUse verifies that a
// simple text-only response (no tool calls) flows
// through the orchestrator correctly.
func TestConversationEndToEndNoToolUse(t *testing.T) {
	provider := &mockStreamingProvider{
		responses: [][]llm.StreamEvent{
			{
				{Type: llm.EventTextDelta,
					Text: "Welcome to the "},
				{Type: llm.EventTextDelta,
					Text: "campaign! "},
				{Type: llm.EventTextDelta,
					Text: "How can I help you today?"},
				{Type: llm.EventUsage,
					Usage: &llm.TokenUsage{
						InputTokens:  100,
						OutputTokens: 20,
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
		"Hello")
	require.NoError(t, err)

	events := collectEvents(ch)

	var accumulatedText string
	var hasDone bool
	var hasToolUse bool

	for _, ev := range events {
		switch ev.Type {
		case llm.EventTextDelta:
			accumulatedText += ev.Text
		case llm.EventDone:
			hasDone = true
		case llm.EventToolUse:
			hasToolUse = true
		case llm.EventError:
			t.Fatalf("unexpected error event: %v",
				ev.Error)
		}
	}

	assert.False(t, hasToolUse,
		"should not have any tool_use events")
	assert.True(t, hasDone,
		"should have done event")
	assert.Equal(t,
		"Welcome to the campaign! "+
			"How can I help you today?",
		accumulatedText,
		"accumulated text should match all deltas")

	// Provider should have been called exactly once.
	assert.Equal(t, 1, provider.callIndex,
		"provider should be called once")
}

// TestConversationEndToEndMultipleToolCalls verifies
// the orchestrator handles an LLM response that
// includes multiple tool calls in a single turn,
// followed by a final text response.
func TestConversationEndToEndMultipleToolCalls(
	t *testing.T,
) {
	var searchCalled, getCalled atomic.Int32

	provider := &mockStreamingProvider{
		responses: [][]llm.StreamEvent{
			// First call: two tool calls.
			{
				{Type: llm.EventToolUse,
					ToolName: "search_entities",
					ToolID:   "call_001",
					ToolInput: json.RawMessage(
						`{"query":"tavern"}`)},
				{Type: llm.EventToolUse,
					ToolName: "get_entity",
					ToolID:   "call_002",
					ToolInput: json.RawMessage(
						`{"id":42}`)},
				{Type: llm.EventUsage,
					Usage: &llm.TokenUsage{
						InputTokens:  600,
						OutputTokens: 80,
					}},
				{Type: llm.EventDone},
			},
			// Second call: final text.
			{
				{Type: llm.EventTextDelta,
					Text: "The Silver Stag tavern is "},
				{Type: llm.EventTextDelta,
					Text: "run by Greta the innkeeper."},
				{Type: llm.EventUsage,
					Usage: &llm.TokenUsage{
						InputTokens:  900,
						OutputTokens: 40,
					}},
				{Type: llm.EventDone},
			},
		},
	}

	registry := NewToolRegistry()
	registry.Register(Tool{
		Definition: llm.ToolDefinition{
			Name:        "search_entities",
			Description: "Search entities",
			InputSchema: json.RawMessage(
				`{"type":"object","properties":{` +
					`"query":{"type":"string"}}}`),
		},
		Execute: func(
			ctx context.Context,
			input json.RawMessage,
		) (json.RawMessage, error) {
			searchCalled.Add(1)
			return json.RawMessage(
				`[{"id":42,"name":"Silver Stag"}]`,
			), nil
		},
	})
	registry.Register(Tool{
		Definition: llm.ToolDefinition{
			Name:        "get_entity",
			Description: "Get entity by ID",
			InputSchema: json.RawMessage(
				`{"type":"object","properties":{` +
					`"id":{"type":"integer"}}}`),
		},
		Execute: func(
			ctx context.Context,
			input json.RawMessage,
		) (json.RawMessage, error) {
			getCalled.Add(1)
			return json.RawMessage(
				`{"id":42,"name":"Silver Stag",` +
					`"type":"location",` +
					`"description":"A cozy tavern"}`,
			), nil
		},
	})

	orch := NewOrchestrator(OrchestratorConfig{
		Provider: provider,
		Tools:    registry,
	})
	defer orch.Stop()

	ch, err := orch.HandleMessage(
		context.Background(),
		1, 1, 1,
		"Tell me about the tavern")
	require.NoError(t, err)

	events := collectEvents(ch)

	var toolUseCount int
	var accumulatedText string
	var hasDone bool
	toolNames := make(map[string]bool)

	for _, ev := range events {
		switch ev.Type {
		case llm.EventToolUse:
			toolUseCount++
			toolNames[ev.ToolName] = true
		case llm.EventTextDelta:
			accumulatedText += ev.Text
		case llm.EventDone:
			hasDone = true
		case llm.EventError:
			t.Fatalf("unexpected error event: %v",
				ev.Error)
		}
	}

	assert.Equal(t, 2, toolUseCount,
		"should see two tool use events")
	assert.True(t, toolNames["search_entities"],
		"should call search_entities")
	assert.True(t, toolNames["get_entity"],
		"should call get_entity")
	assert.True(t, hasDone,
		"should have done event")
	assert.Contains(t, accumulatedText,
		"Silver Stag",
		"response should mention the tavern")
	assert.Contains(t, accumulatedText,
		"Greta",
		"response should mention the innkeeper")

	// Both tools should have been executed exactly
	// once.
	assert.Equal(t, int32(1), searchCalled.Load(),
		"search_entities should be called once")
	assert.Equal(t, int32(1), getCalled.Load(),
		"get_entity should be called once")

	// Provider should have been called twice.
	assert.Equal(t, 2, provider.callIndex,
		"provider should be called twice")
}

// TestConversationEndToEndToolError verifies that when
// a tool execution fails, the orchestrator passes the
// error back to the LLM as a tool result, and the LLM
// recovers gracefully with a text response.
func TestConversationEndToEndToolError(t *testing.T) {
	provider := &mockStreamingProvider{
		responses: [][]llm.StreamEvent{
			// First call: tool call.
			{
				{Type: llm.EventToolUse,
					ToolName: "search_entities",
					ToolID:   "call_001",
					ToolInput: json.RawMessage(
						`{"query":"dragons"}`)},
				{Type: llm.EventDone},
			},
			// Second call: LLM recovers from the
			// tool error and produces text.
			{
				{Type: llm.EventTextDelta,
					Text: "I was unable to search "},
				{Type: llm.EventTextDelta,
					Text: "for entities at this time. "},
				{Type: llm.EventTextDelta,
					Text: "Please try again later."},
				{Type: llm.EventUsage,
					Usage: &llm.TokenUsage{
						InputTokens:  400,
						OutputTokens: 30,
					}},
				{Type: llm.EventDone},
			},
		},
	}

	registry := NewToolRegistry()
	registry.Register(Tool{
		Definition: llm.ToolDefinition{
			Name:        "search_entities",
			Description: "Search entities",
			InputSchema: json.RawMessage(
				`{"type":"object","properties":{` +
					`"query":{"type":"string"}}}`),
		},
		Execute: func(
			ctx context.Context,
			input json.RawMessage,
		) (json.RawMessage, error) {
			return nil, fmt.Errorf(
				"database connection lost")
		},
	})

	orch := NewOrchestrator(OrchestratorConfig{
		Provider: provider,
		Tools:    registry,
	})
	defer orch.Stop()

	ch, err := orch.HandleMessage(
		context.Background(),
		1, 1, 1,
		"Find dragons nearby")
	require.NoError(t, err)

	events := collectEvents(ch)

	var accumulatedText string
	var hasToolUse, hasDone bool
	var hasStreamError bool

	for _, ev := range events {
		switch ev.Type {
		case llm.EventToolUse:
			hasToolUse = true
		case llm.EventTextDelta:
			accumulatedText += ev.Text
		case llm.EventDone:
			hasDone = true
		case llm.EventError:
			hasStreamError = true
		}
	}

	assert.True(t, hasToolUse,
		"should have tool_use event")
	assert.True(t, hasDone,
		"should have done event")
	assert.False(t, hasStreamError,
		"tool execution error should not produce "+
			"a stream error event")
	assert.Contains(t, accumulatedText,
		"unable to search",
		"LLM should acknowledge the tool failure")

	// Provider should have been called twice: once
	// for the tool call, once after the error result.
	assert.Equal(t, 2, provider.callIndex,
		"provider should be called twice")
}

// TestConversationEndToEndCachePreservation verifies
// that the session cache is correctly populated after
// a full tool-use conversation, preserving all message
// history including tool interactions.
func TestConversationEndToEndCachePreservation(
	t *testing.T,
) {
	provider := &mockStreamingProvider{
		responses: [][]llm.StreamEvent{
			// First call: tool use.
			{
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
					Text: "Echo complete."},
				{Type: llm.EventDone},
			},
		},
	}

	registry := newTestSubTools()

	orch := NewOrchestrator(OrchestratorConfig{
		Provider: provider,
		Tools:    registry,
	})
	defer orch.Stop()

	const convID int64 = 99

	ch, err := orch.HandleMessage(
		context.Background(),
		convID, 1, 1,
		"Echo something")
	require.NoError(t, err)
	_ = collectEvents(ch)

	// Verify the cache has the full conversation
	// history.
	entry, ok := orch.cache.Get(convID)
	require.True(t, ok,
		"cache should have entry for conversation")
	require.NotEmpty(t, entry.Messages,
		"cached messages should not be empty")

	// Expected message sequence:
	// 1. user message
	// 2. assistant tool call
	// 3. user tool result
	// 4. assistant final response
	require.Len(t, entry.Messages, 4,
		"cache should have 4 messages: "+
			"user, tool call, tool result, "+
			"assistant response")

	assert.Equal(t, "user",
		entry.Messages[0].Role)
	assert.Equal(t, "Echo something",
		entry.Messages[0].Content)
	assert.Equal(t, "assistant",
		entry.Messages[1].Role)
	assert.Equal(t, "echo_tool",
		entry.Messages[1].ToolName)
	assert.Equal(t, "user",
		entry.Messages[2].Role)
	assert.Equal(t, "call_001",
		entry.Messages[2].ToolUseID)
	assert.Equal(t, "assistant",
		entry.Messages[3].Role)
	assert.Equal(t, "Echo complete.",
		entry.Messages[3].Content)
}
