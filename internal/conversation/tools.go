/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

// Package conversation implements the conversational
// backend for Imagineer, including tool registration,
// orchestration, and session management.
package conversation

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/antonypegg/imagineer/internal/llm"
)

// ToolExecuteFunc is the function signature for tool
// execution. It receives the raw JSON input from the
// LLM and returns the raw JSON result.
type ToolExecuteFunc func(
	ctx context.Context,
	input json.RawMessage,
) (json.RawMessage, error)

// Tool pairs a tool definition (sent to the LLM) with
// its execution function (called when the LLM invokes
// the tool).
type Tool struct {
	Definition llm.ToolDefinition
	Execute    ToolExecuteFunc
}

// ToolRegistry holds registered tools and dispatches
// execution by name.
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]Tool
	order []string // preserve registration order
}

// NewToolRegistry creates an empty tool registry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry.
func (r *ToolRegistry) Register(tool Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	name := tool.Definition.Name
	if _, exists := r.tools[name]; !exists {
		r.order = append(r.order, name)
	}
	r.tools[name] = tool
}

// Get retrieves a tool by name.
func (r *ToolRegistry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	return t, ok
}

// Definitions returns all tool definitions in
// registration order, suitable for passing to the LLM.
func (r *ToolRegistry) Definitions() []llm.ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	defs := make([]llm.ToolDefinition, 0, len(r.order))
	for _, name := range r.order {
		defs = append(defs, r.tools[name].Definition)
	}
	return defs
}

// Execute runs a tool by name with the given input.
func (r *ToolRegistry) Execute(
	ctx context.Context,
	name string,
	input json.RawMessage,
) (json.RawMessage, error) {
	r.mu.RLock()
	tool, ok := r.tools[name]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
	return tool.Execute(ctx, input)
}
