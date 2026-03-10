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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProceduralToolDefinitions(t *testing.T) {
	// Verify all 8 tools have valid definitions.
	// Use nil DB since we only check definitions.
	tools := BuildProceduralTools(nil, 0, "")
	assert.Len(t, tools, 8)

	names := make(map[string]bool)
	for _, tool := range tools {
		assert.NotEmpty(t, tool.Definition.Name,
			"tool definition must have a name")
		assert.NotEmpty(t, tool.Definition.Description,
			"tool %s must have a description",
			tool.Definition.Name)
		assert.NotNil(t, tool.Definition.InputSchema,
			"tool %s must have an input schema",
			tool.Definition.Name)
		names[tool.Definition.Name] = true
	}

	// Verify expected tool names are present.
	expectedNames := []string{
		"search_entities",
		"get_entity",
		"create_entity",
		"update_entity",
		"create_relationship",
		"get_related_entities",
		"search_content",
		"read_game_schema",
	}
	for _, name := range expectedNames {
		assert.True(t, names[name],
			"expected tool %q to be registered", name)
	}
}

func TestProceduralToolInputSchemas(t *testing.T) {
	tools := BuildProceduralTools(nil, 0, "")

	for _, tool := range tools {
		t.Run(tool.Definition.Name, func(t *testing.T) {
			// Verify each InputSchema is valid JSON.
			var schema map[string]interface{}
			err := json.Unmarshal(
				tool.Definition.InputSchema, &schema)
			require.NoError(t, err,
				"InputSchema must be valid JSON")

			// Verify the schema has a "type" field
			// set to "object".
			assert.Equal(t, "object", schema["type"],
				"InputSchema type must be object")

			// Verify the schema has a "properties"
			// field.
			_, hasProps := schema["properties"]
			assert.True(t, hasProps,
				"InputSchema must have properties")
		})
	}
}

func TestProceduralToolNamesUnique(t *testing.T) {
	tools := BuildProceduralTools(nil, 0, "")
	seen := make(map[string]bool)
	for _, tool := range tools {
		name := tool.Definition.Name
		assert.False(t, seen[name],
			"duplicate tool name: %s", name)
		seen[name] = true
	}
}

func TestReadGameSchemaTool(t *testing.T) {
	// Create a temp schemas dir with a test file.
	dir := t.TempDir()
	content := "name: Test System\ncode: test_system\n"
	err := os.WriteFile(
		filepath.Join(dir, "test_system.yaml"),
		[]byte(content),
		0644)
	require.NoError(t, err)

	tool := buildReadGameSchemaTool(dir)

	result, err := tool.Execute(
		context.Background(),
		json.RawMessage(`{"system_code":"test_system"}`))
	require.NoError(t, err)

	var parsed map[string]string
	err = json.Unmarshal(result, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "test_system",
		parsed["system_code"])
	assert.Contains(t, parsed["content"],
		"Test System")
	assert.Contains(t, parsed["content"],
		"code: test_system")
}

func TestReadGameSchemaToolNotFound(t *testing.T) {
	dir := t.TempDir()
	tool := buildReadGameSchemaTool(dir)

	_, err := tool.Execute(
		context.Background(),
		json.RawMessage(
			`{"system_code":"nonexistent"}`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "schema not found")
}

func TestReadGameSchemaToolPathTraversal(t *testing.T) {
	dir := t.TempDir()
	tool := buildReadGameSchemaTool(dir)

	_, err := tool.Execute(
		context.Background(),
		json.RawMessage(
			`{"system_code":"../../../etc/passwd"}`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid system_code")
}

func TestReadGameSchemaToolEmptyCode(t *testing.T) {
	dir := t.TempDir()
	tool := buildReadGameSchemaTool(dir)

	_, err := tool.Execute(
		context.Background(),
		json.RawMessage(`{"system_code":""}`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(),
		"system_code is required")
}

func TestReadGameSchemaToolInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	tool := buildReadGameSchemaTool(dir)

	_, err := tool.Execute(
		context.Background(),
		json.RawMessage(`{invalid json}`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid input")
}

func TestProceduralToolInputValidation(t *testing.T) {
	// Test that tools with nil DB handle input
	// validation errors before hitting the DB.
	// Each tool should reject invalid JSON and
	// missing required fields.
	tests := []struct {
		name     string
		toolName string
		input    string
		wantErr  string
	}{
		{
			name:     "search_entities invalid json",
			toolName: "search_entities",
			input:    `{bad`,
			wantErr:  "invalid input",
		},
		{
			name:     "search_entities missing query",
			toolName: "search_entities",
			input:    `{"query":""}`,
			wantErr:  "query is required",
		},
		{
			name:     "get_entity invalid json",
			toolName: "get_entity",
			input:    `{bad`,
			wantErr:  "invalid input",
		},
		{
			name:     "get_entity missing id",
			toolName: "get_entity",
			input:    `{"id":0}`,
			wantErr:  "id is required",
		},
		{
			name:     "create_entity missing name",
			toolName: "create_entity",
			input:    `{"entity_type":"npc"}`,
			wantErr:  "name is required",
		},
		{
			name:     "create_entity missing type",
			toolName: "create_entity",
			input:    `{"name":"Bob"}`,
			wantErr:  "entity_type is required",
		},
		{
			name:     "update_entity missing id",
			toolName: "update_entity",
			input:    `{"name":"Bob"}`,
			wantErr:  "id is required",
		},
		{
			name:     "create_relationship missing source",
			toolName: "create_relationship",
			input:    `{"target_entity_id":2,"relationship_type_id":1}`,
			wantErr:  "source_entity_id is required",
		},
		{
			name:     "create_relationship missing target",
			toolName: "create_relationship",
			input:    `{"source_entity_id":1,"relationship_type_id":1}`,
			wantErr:  "target_entity_id is required",
		},
		{
			name:     "create_relationship missing type",
			toolName: "create_relationship",
			input:    `{"source_entity_id":1,"target_entity_id":2}`,
			wantErr:  "relationship_type_id is required",
		},
		{
			name:     "get_related_entities missing id",
			toolName: "get_related_entities",
			input:    `{"entity_id":0}`,
			wantErr:  "entity_id is required",
		},
		{
			name:     "search_content missing query",
			toolName: "search_content",
			input:    `{"query":""}`,
			wantErr:  "query is required",
		},
	}

	tools := BuildProceduralTools(nil, 0, "")
	toolMap := make(map[string]Tool)
	for _, tool := range tools {
		toolMap[tool.Definition.Name] = tool
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool, ok := toolMap[tt.toolName]
			require.True(t, ok,
				"tool %s not found", tt.toolName)

			_, err := tool.Execute(
				context.Background(),
				json.RawMessage(tt.input))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}
