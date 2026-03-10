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

	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockStreamingProvider returns predetermined event
// sequences for each call to CompleteStream.
type mockStreamingProvider struct {
	responses [][]llm.StreamEvent
	callIndex int
}

func (m *mockStreamingProvider) CompleteStream(
	ctx context.Context,
	req llm.StreamingRequest,
) (<-chan llm.StreamEvent, error) {
	if m.callIndex >= len(m.responses) {
		return nil, fmt.Errorf(
			"unexpected call %d", m.callIndex)
	}
	events := m.responses[m.callIndex]
	m.callIndex++
	ch := make(chan llm.StreamEvent, len(events))
	for _, e := range events {
		ch <- e
	}
	close(ch)
	return ch, nil
}

// newTestSubTools creates a ToolRegistry with a simple
// echo tool for testing agent tool-use loops.
func newTestSubTools() *ToolRegistry {
	registry := NewToolRegistry()
	registry.Register(Tool{
		Definition: llm.ToolDefinition{
			Name:        "echo_tool",
			Description: "Echoes back the input",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"message": {
						"type": "string"
					}
				},
				"required": ["message"]
			}`),
		},
		Execute: func(
			ctx context.Context,
			input json.RawMessage,
		) (json.RawMessage, error) {
			return json.Marshal(map[string]string{
				"echoed": string(input),
			})
		},
	})
	return registry
}

// newFullTestRegistry creates a ToolRegistry that has
// all the tool names the agent tools expect to filter.
// Each tool simply returns a fixed JSON response.
func newFullTestRegistry() *ToolRegistry {
	registry := NewToolRegistry()
	names := []string{
		"search_entities", "get_entity",
		"create_entity", "update_entity",
		"create_relationship", "get_related_entities",
		"search_content", "read_game_schema",
	}
	for _, name := range names {
		n := name // capture
		registry.Register(Tool{
			Definition: llm.ToolDefinition{
				Name:        n,
				Description: "Test tool: " + n,
				InputSchema: json.RawMessage(
					`{"type":"object","properties":{}}`),
			},
			Execute: func(
				ctx context.Context,
				input json.RawMessage,
			) (json.RawMessage, error) {
				return json.Marshal(
					map[string]string{
						"tool": n,
					})
			},
		})
	}
	return registry
}

func TestAgentToolDefinitions(t *testing.T) {
	provider := &mockStreamingProvider{}
	registry := newFullTestRegistry()
	tools := BuildAgentTools(registry, provider)

	assert.Len(t, tools, 3,
		"expected 3 agent tools")

	expectedNames := map[string]bool{
		"ask_ttrpg_expert": true,
		"ask_canon_expert": true,
		"ask_graph_expert": true,
	}

	for _, tool := range tools {
		assert.NotEmpty(t, tool.Definition.Name,
			"tool must have a name")
		assert.NotEmpty(t, tool.Definition.Description,
			"tool %s must have a description",
			tool.Definition.Name)
		assert.NotNil(t, tool.Definition.InputSchema,
			"tool %s must have an input schema",
			tool.Definition.Name)
		assert.True(t,
			expectedNames[tool.Definition.Name],
			"unexpected tool name: %s",
			tool.Definition.Name)
	}
}

func TestAgentToolInputSchemas(t *testing.T) {
	provider := &mockStreamingProvider{}
	registry := newFullTestRegistry()
	tools := BuildAgentTools(registry, provider)

	for _, tool := range tools {
		t.Run(tool.Definition.Name, func(t *testing.T) {
			var schema map[string]interface{}
			err := json.Unmarshal(
				tool.Definition.InputSchema, &schema)
			require.NoError(t, err,
				"InputSchema must be valid JSON")

			assert.Equal(t, "object", schema["type"],
				"InputSchema type must be object")

			props, hasProps := schema["properties"]
			assert.True(t, hasProps,
				"InputSchema must have properties")

			propsMap, ok := props.(map[string]interface{})
			require.True(t, ok,
				"properties must be an object")
			_, hasQuestion := propsMap["question"]
			assert.True(t, hasQuestion,
				"InputSchema must have question property")

			required, hasRequired := schema["required"]
			assert.True(t, hasRequired,
				"InputSchema must have required fields")

			reqSlice, ok := required.([]interface{})
			require.True(t, ok,
				"required must be an array")
			assert.Contains(t, reqSlice, "question",
				"question must be required")
		})
	}
}

func TestAgentToolNamesUnique(t *testing.T) {
	provider := &mockStreamingProvider{}
	registry := newFullTestRegistry()
	tools := BuildAgentTools(registry, provider)

	seen := make(map[string]bool)
	for _, tool := range tools {
		name := tool.Definition.Name
		assert.False(t, seen[name],
			"duplicate tool name: %s", name)
		seen[name] = true
	}
}

func TestAgentToolInputValidation(t *testing.T) {
	provider := &mockStreamingProvider{}
	registry := newFullTestRegistry()
	tools := BuildAgentTools(registry, provider)

	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "invalid json",
			input:   `{bad`,
			wantErr: "invalid input",
		},
		{
			name:    "missing question",
			input:   `{}`,
			wantErr: "question is required",
		},
		{
			name:    "empty question",
			input:   `{"question":""}`,
			wantErr: "question is required",
		},
	}

	for _, tool := range tools {
		for _, tt := range tests {
			testName := fmt.Sprintf("%s/%s",
				tool.Definition.Name, tt.name)
			t.Run(testName, func(t *testing.T) {
				_, err := tool.Execute(
					context.Background(),
					json.RawMessage(tt.input))
				require.Error(t, err)
				assert.Contains(t, err.Error(),
					tt.wantErr)
			})
		}
	}
}

func TestBuildAgentToolTextResponse(t *testing.T) {
	// Mock provider returns a simple text response
	// with no tool calls.
	provider := &mockStreamingProvider{
		responses: [][]llm.StreamEvent{
			{
				{Type: llm.EventTextDelta,
					Text: "Here is my expert "},
				{Type: llm.EventTextDelta,
					Text: "analysis of the situation."},
				{Type: llm.EventDone},
			},
		},
	}

	subTools := newTestSubTools()
	tool := buildAgentTool(
		"test_expert",
		"A test expert",
		"You are a test expert.",
		subTools,
		provider,
	)

	result, err := tool.Execute(
		context.Background(),
		json.RawMessage(
			`{"question":"What do you think?"}`))
	require.NoError(t, err)

	var output agentToolOutput
	err = json.Unmarshal(result, &output)
	require.NoError(t, err)
	assert.Equal(t,
		"Here is my expert analysis of the situation.",
		output.Response)
}

func TestBuildAgentToolWithToolUse(t *testing.T) {
	// Mock provider: first call returns a tool call,
	// second call returns text after seeing the result.
	provider := &mockStreamingProvider{
		responses: [][]llm.StreamEvent{
			// First response: tool call.
			{
				{Type: llm.EventTextDelta,
					Text: "Let me check that. "},
				{Type: llm.EventToolUse,
					ToolName: "echo_tool",
					ToolID:   "call_001",
					ToolInput: json.RawMessage(
						`{"message":"hello"}`)},
				{Type: llm.EventDone},
			},
			// Second response: text after tool result.
			{
				{Type: llm.EventTextDelta,
					Text: "Based on the result, "},
				{Type: llm.EventTextDelta,
					Text: "everything looks good."},
				{Type: llm.EventDone},
			},
		},
	}

	subTools := newTestSubTools()
	tool := buildAgentTool(
		"test_expert",
		"A test expert",
		"You are a test expert.",
		subTools,
		provider,
	)

	result, err := tool.Execute(
		context.Background(),
		json.RawMessage(
			`{"question":"Check something"}`))
	require.NoError(t, err)

	var output agentToolOutput
	err = json.Unmarshal(result, &output)
	require.NoError(t, err)
	assert.Equal(t,
		"Based on the result, everything looks good.",
		output.Response)

	// Verify the provider was called twice.
	assert.Equal(t, 2, provider.callIndex,
		"provider should have been called twice")
}

func TestBuildAgentToolMultipleToolCalls(t *testing.T) {
	// Mock provider: first call returns two tool calls,
	// second call returns final text.
	provider := &mockStreamingProvider{
		responses: [][]llm.StreamEvent{
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
			{
				{Type: llm.EventTextDelta,
					Text: "Both calls succeeded."},
				{Type: llm.EventDone},
			},
		},
	}

	subTools := newTestSubTools()
	tool := buildAgentTool(
		"test_expert",
		"A test expert",
		"You are a test expert.",
		subTools,
		provider,
	)

	result, err := tool.Execute(
		context.Background(),
		json.RawMessage(
			`{"question":"Run two tools"}`))
	require.NoError(t, err)

	var output agentToolOutput
	err = json.Unmarshal(result, &output)
	require.NoError(t, err)
	assert.Equal(t, "Both calls succeeded.",
		output.Response)
}

func TestBuildAgentToolMaxIterations(t *testing.T) {
	// Mock provider always returns a tool call, never
	// text-only. Build enough responses for all
	// iterations.
	responses := make([][]llm.StreamEvent,
		maxAgentIterations)
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

	subTools := newTestSubTools()
	tool := buildAgentTool(
		"test_expert",
		"A test expert",
		"You are a test expert.",
		subTools,
		provider,
	)

	_, err := tool.Execute(
		context.Background(),
		json.RawMessage(
			`{"question":"Loop forever"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(),
		"exceeded maximum iterations")
}

func TestBuildAgentToolStreamError(t *testing.T) {
	// Mock provider returns an error event.
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

	subTools := newTestSubTools()
	tool := buildAgentTool(
		"test_expert",
		"A test expert",
		"You are a test expert.",
		subTools,
		provider,
	)

	_, err := tool.Execute(
		context.Background(),
		json.RawMessage(
			`{"question":"Will this fail?"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(),
		"rate limit exceeded")
}

func TestBuildAgentToolProviderError(t *testing.T) {
	// Mock provider returns an error from
	// CompleteStream itself (not via events).
	provider := &mockStreamingProvider{
		responses: nil, // no responses available
	}

	subTools := newTestSubTools()
	tool := buildAgentTool(
		"test_expert",
		"A test expert",
		"You are a test expert.",
		subTools,
		provider,
	)

	_, err := tool.Execute(
		context.Background(),
		json.RawMessage(
			`{"question":"Will this fail?"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(),
		"agent stream failed")
}

func TestBuildAgentToolSubToolError(t *testing.T) {
	// When a sub-tool execution fails, the error
	// should be passed back as a tool result so the
	// expert can recover.
	provider := &mockStreamingProvider{
		responses: [][]llm.StreamEvent{
			{
				{Type: llm.EventToolUse,
					ToolName:  "failing_tool",
					ToolID:    "call_fail",
					ToolInput: json.RawMessage(`{}`)},
				{Type: llm.EventDone},
			},
			{
				{Type: llm.EventTextDelta,
					Text: "The tool failed, but I recovered."},
				{Type: llm.EventDone},
			},
		},
	}

	subTools := NewToolRegistry()
	subTools.Register(Tool{
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

	tool := buildAgentTool(
		"test_expert",
		"A test expert",
		"You are a test expert.",
		subTools,
		provider,
	)

	result, err := tool.Execute(
		context.Background(),
		json.RawMessage(
			`{"question":"Try the failing tool"}`))
	require.NoError(t, err)

	var output agentToolOutput
	err = json.Unmarshal(result, &output)
	require.NoError(t, err)
	assert.Equal(t,
		"The tool failed, but I recovered.",
		output.Response)
}

func TestBuildAgentToolContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	provider := &mockStreamingProvider{}
	subTools := NewToolRegistry()

	tool := buildAgentTool(
		"test_agent", "Test agent",
		"You are a test agent.",
		subTools, provider)

	_, err := tool.Execute(ctx,
		json.RawMessage(`{"question":"test"}`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cancel")
}

func TestFilterTools(t *testing.T) {
	source := newFullTestRegistry()

	t.Run("filters to requested names", func(t *testing.T) {
		filtered := filterTools(source,
			"search_entities", "get_entity")
		defs := filtered.Definitions()
		assert.Len(t, defs, 2)

		names := make(map[string]bool)
		for _, d := range defs {
			names[d.Name] = true
		}
		assert.True(t, names["search_entities"])
		assert.True(t, names["get_entity"])
	})

	t.Run("skips missing tools", func(t *testing.T) {
		filtered := filterTools(source,
			"search_entities", "nonexistent_tool")
		defs := filtered.Definitions()
		assert.Len(t, defs, 1)
		assert.Equal(t, "search_entities",
			defs[0].Name)
	})

	t.Run("empty filter returns empty registry",
		func(t *testing.T) {
			filtered := filterTools(source)
			defs := filtered.Definitions()
			assert.Len(t, defs, 0)
		})
}

func TestBuildAgentToolsSubToolAssignments(t *testing.T) {
	registry := newFullTestRegistry()
	provider := &mockStreamingProvider{}

	// We cannot directly inspect the sub-tool
	// registries, but we can verify the tools are
	// built with the correct names and that
	// BuildAgentTools does not panic.
	tools := BuildAgentTools(registry, provider)

	names := make([]string, len(tools))
	for i, tool := range tools {
		names[i] = tool.Definition.Name
	}

	assert.Equal(t, []string{
		"ask_ttrpg_expert",
		"ask_canon_expert",
		"ask_graph_expert",
	}, names)
}

func TestCollectStreamEventsEmpty(t *testing.T) {
	ch := make(chan llm.StreamEvent)
	close(ch)

	text, calls, err := collectStreamEvents(ch)
	require.NoError(t, err)
	assert.Empty(t, text)
	assert.Empty(t, calls)
}

func TestCollectStreamEventsTextOnly(t *testing.T) {
	ch := make(chan llm.StreamEvent, 3)
	ch <- llm.StreamEvent{
		Type: llm.EventTextDelta, Text: "Hello "}
	ch <- llm.StreamEvent{
		Type: llm.EventTextDelta, Text: "world"}
	ch <- llm.StreamEvent{Type: llm.EventDone}
	close(ch)

	text, calls, err := collectStreamEvents(ch)
	require.NoError(t, err)
	assert.Equal(t, "Hello world", text)
	assert.Empty(t, calls)
}

func TestCollectStreamEventsWithToolUse(t *testing.T) {
	ch := make(chan llm.StreamEvent, 4)
	ch <- llm.StreamEvent{
		Type: llm.EventTextDelta, Text: "Checking..."}
	ch <- llm.StreamEvent{
		Type:      llm.EventToolUse,
		ToolName:  "my_tool",
		ToolID:    "id_1",
		ToolInput: json.RawMessage(`{"key":"val"}`),
	}
	ch <- llm.StreamEvent{
		Type:  llm.EventUsage,
		Usage: &llm.TokenUsage{InputTokens: 100},
	}
	ch <- llm.StreamEvent{Type: llm.EventDone}
	close(ch)

	text, calls, err := collectStreamEvents(ch)
	require.NoError(t, err)
	assert.Equal(t, "Checking...", text)
	require.Len(t, calls, 1)
	assert.Equal(t, "my_tool", calls[0].Name)
	assert.Equal(t, "id_1", calls[0].ID)
	assert.JSONEq(t, `{"key":"val"}`,
		string(calls[0].Input))
}
