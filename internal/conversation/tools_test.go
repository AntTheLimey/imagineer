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
	"testing"

	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolRegistryRegisterAndGet(t *testing.T) {
	registry := NewToolRegistry()

	registry.Register(Tool{
		Definition: llm.ToolDefinition{
			Name:        "search_entities",
			Description: "Search campaign entities",
		},
		Execute: func(ctx context.Context,
			input json.RawMessage,
		) (json.RawMessage, error) {
			return json.Marshal(map[string]string{
				"result": "found",
			})
		},
	})

	tool, ok := registry.Get("search_entities")
	assert.True(t, ok)
	assert.Equal(t, "search_entities",
		tool.Definition.Name)
}

func TestToolRegistryDefinitions(t *testing.T) {
	registry := NewToolRegistry()
	registry.Register(Tool{
		Definition: llm.ToolDefinition{
			Name:        "tool_a",
			Description: "Tool A",
		},
	})
	registry.Register(Tool{
		Definition: llm.ToolDefinition{
			Name:        "tool_b",
			Description: "Tool B",
		},
	})

	defs := registry.Definitions()
	assert.Len(t, defs, 2)
}

func TestToolRegistryExecute(t *testing.T) {
	registry := NewToolRegistry()
	registry.Register(Tool{
		Definition: llm.ToolDefinition{
			Name: "echo",
		},
		Execute: func(ctx context.Context,
			input json.RawMessage,
		) (json.RawMessage, error) {
			return input, nil
		},
	})

	input := json.RawMessage(`{"msg":"hello"}`)
	result, err := registry.Execute(
		context.Background(), "echo", input)
	require.NoError(t, err)
	assert.JSONEq(t, `{"msg":"hello"}`,
		string(result))
}

func TestToolRegistryExecuteNotFound(t *testing.T) {
	registry := NewToolRegistry()
	_, err := registry.Execute(
		context.Background(), "nonexistent", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown tool")
}
