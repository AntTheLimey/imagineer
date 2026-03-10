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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocumentToolDefinitions(t *testing.T) {
	// Verify both tools have valid definitions.
	// Use nil DB since we only check definitions.
	tools := BuildDocumentTools(nil, 0)
	assert.Len(t, tools, 2)

	for _, tool := range tools {
		assert.NotEmpty(t, tool.Definition.Name,
			"tool definition must have a name")
		assert.NotEmpty(t, tool.Definition.Description,
			"tool %s must have a description",
			tool.Definition.Name)
		assert.NotNil(t, tool.Definition.InputSchema,
			"tool %s must have an input schema",
			tool.Definition.Name)
	}

	// Verify expected tool names are present.
	names := make(map[string]bool)
	for _, tool := range tools {
		names[tool.Definition.Name] = true
	}
	assert.True(t, names["read_document"],
		"expected read_document tool")
	assert.True(t, names["edit_document"],
		"expected edit_document tool")
}

func TestDocumentToolInputSchemas(t *testing.T) {
	tools := BuildDocumentTools(nil, 0)

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

			// Verify the schema has "required" fields.
			_, hasRequired := schema["required"]
			assert.True(t, hasRequired,
				"InputSchema must have required fields")
		})
	}
}

func TestDocumentToolNamesUnique(t *testing.T) {
	tools := BuildDocumentTools(nil, 0)
	seen := make(map[string]bool)
	for _, tool := range tools {
		name := tool.Definition.Name
		assert.False(t, seen[name],
			"duplicate tool name: %s", name)
		seen[name] = true
	}
}

func TestReadDocumentInputValidation(t *testing.T) {
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
			name:    "missing scope_type",
			input:   `{"scope_id": 1}`,
			wantErr: "scope_type is required",
		},
		{
			name:    "empty scope_type",
			input:   `{"scope_type": "", "scope_id": 1}`,
			wantErr: "scope_type is required",
		},
		{
			name:    "missing scope_id",
			input:   `{"scope_type": "chapter"}`,
			wantErr: "scope_id is required",
		},
		{
			name:    "zero scope_id",
			input:   `{"scope_type": "chapter", "scope_id": 0}`,
			wantErr: "scope_id is required",
		},
		{
			name:    "invalid scope_type",
			input:   `{"scope_type": "unknown", "scope_id": 1}`,
			wantErr: "invalid scope_type",
		},
		{
			name:    "campaign scope rejection",
			input:   `{"scope_type": "campaign", "scope_id": 1}`,
			wantErr: "campaign scope does not have a document",
		},
	}

	// Build tools with nil DB — validation happens
	// before any DB calls.
	tools := BuildDocumentTools(nil, 0)
	var readTool Tool
	for _, tool := range tools {
		if tool.Definition.Name == "read_document" {
			readTool = tool
			break
		}
	}
	require.NotEmpty(t, readTool.Definition.Name,
		"read_document tool not found")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := readTool.Execute(
				context.Background(),
				json.RawMessage(tt.input))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestEditDocumentInputValidation(t *testing.T) {
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
			name:    "missing scope_type",
			input:   `{"scope_id": 1, "operation": "append", "new_text": "hello"}`,
			wantErr: "scope_type is required",
		},
		{
			name:    "missing scope_id",
			input:   `{"scope_type": "chapter", "operation": "append", "new_text": "hello"}`,
			wantErr: "scope_id is required",
		},
		{
			name:    "missing operation",
			input:   `{"scope_type": "chapter", "scope_id": 1, "new_text": "hello"}`,
			wantErr: "operation is required",
		},
		{
			name:    "missing new_text",
			input:   `{"scope_type": "chapter", "scope_id": 1, "operation": "append"}`,
			wantErr: "new_text is required",
		},
		{
			name:    "empty new_text",
			input:   `{"scope_type": "chapter", "scope_id": 1, "operation": "append", "new_text": ""}`,
			wantErr: "new_text is required",
		},
		{
			name:    "invalid operation",
			input:   `{"scope_type": "chapter", "scope_id": 1, "operation": "delete", "new_text": "hello"}`,
			wantErr: "invalid operation",
		},
		{
			name:    "replace without old_text",
			input:   `{"scope_type": "chapter", "scope_id": 1, "operation": "replace", "new_text": "hello"}`,
			wantErr: "old_text is required for replace operation",
		},
		{
			name:    "replace with empty old_text",
			input:   `{"scope_type": "chapter", "scope_id": 1, "operation": "replace", "old_text": "", "new_text": "hello"}`,
			wantErr: "old_text is required for replace operation",
		},
		{
			name:    "invalid scope_type",
			input:   `{"scope_type": "unknown", "scope_id": 1, "operation": "append", "new_text": "hello"}`,
			wantErr: "invalid scope_type",
		},
		{
			name:    "campaign scope rejection",
			input:   `{"scope_type": "campaign", "scope_id": 1, "operation": "append", "new_text": "hello"}`,
			wantErr: "campaign scope does not have a document",
		},
	}

	// Build tools with nil DB — validation happens
	// before any DB calls.
	tools := BuildDocumentTools(nil, 0)
	var editTool Tool
	for _, tool := range tools {
		if tool.Definition.Name == "edit_document" {
			editTool = tool
			break
		}
	}
	require.NotEmpty(t, editTool.Definition.Name,
		"edit_document tool not found")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := editTool.Execute(
				context.Background(),
				json.RawMessage(tt.input))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}
