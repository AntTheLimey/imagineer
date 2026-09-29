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
	"encoding/json"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildToolRegistry(t *testing.T) {
	// Build with nil DB and a mock provider.
	// Definitions are created without executing tools,
	// so nil DB is safe.
	provider := &mockStreamingProvider{}
	registry := BuildToolRegistry(
		nil, 1, provider, "")

	defs := registry.Definitions()
	assert.Len(t, defs, 13,
		"expected 13 tools total: "+
			"8 procedural + 2 document + 3 agent")
}

func TestBuildToolRegistryToolNames(t *testing.T) {
	provider := &mockStreamingProvider{}
	registry := BuildToolRegistry(
		nil, 1, provider, "")

	expectedNames := []string{
		// 8 procedural tools
		"search_entities",
		"get_entity",
		"create_entity",
		"update_entity",
		"create_relationship",
		"get_related_entities",
		"search_content",
		"read_game_schema",
		// 2 document tools
		"read_document",
		"edit_document",
		// 3 agent tools
		"ask_ttrpg_expert",
		"ask_canon_expert",
		"ask_graph_expert",
	}

	defs := registry.Definitions()
	actualNames := make([]string, len(defs))
	for i, d := range defs {
		actualNames[i] = d.Name
	}

	// Sort both slices for comparison.
	sort.Strings(expectedNames)
	sort.Strings(actualNames)
	assert.Equal(t, expectedNames, actualNames,
		"registry must contain exactly the "+
			"expected tool names")
}

func TestBuildToolRegistryNoDuplicates(t *testing.T) {
	provider := &mockStreamingProvider{}
	registry := BuildToolRegistry(
		nil, 1, provider, "")

	defs := registry.Definitions()
	seen := make(map[string]bool, len(defs))
	for _, d := range defs {
		assert.False(t, seen[d.Name],
			"duplicate tool name: %s", d.Name)
		seen[d.Name] = true
	}
}

func TestBuildToolRegistryDefinitionsValid(t *testing.T) {
	provider := &mockStreamingProvider{}
	registry := BuildToolRegistry(
		nil, 1, provider, "")

	defs := registry.Definitions()
	for _, d := range defs {
		t.Run(d.Name, func(t *testing.T) {
			assert.NotEmpty(t, d.Name,
				"tool must have a name")
			assert.NotEmpty(t, d.Description,
				"tool %s must have a description",
				d.Name)
			require.NotNil(t, d.InputSchema,
				"tool %s must have an input schema",
				d.Name)

			// Verify the schema is valid JSON.
			var schema map[string]interface{}
			err := json.Unmarshal(
				d.InputSchema, &schema)
			require.NoError(t, err,
				"tool %s InputSchema must be "+
					"valid JSON", d.Name)

			// Every tool schema should be an object
			// type.
			assert.Equal(t, "object",
				schema["type"],
				"tool %s InputSchema type must "+
					"be object", d.Name)
		})
	}
}
