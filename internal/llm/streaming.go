/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package llm

import (
	"context"
	"encoding/json"
)

// StreamEventType identifies the kind of streaming event.
type StreamEventType int

const (
	EventTextDelta StreamEventType = iota
	EventToolUse
	EventUsage
	EventDone
	EventError
)

// StreamEvent represents a single event from the
// streaming LLM response.
type StreamEvent struct {
	Type      StreamEventType
	Text      string
	ToolName  string
	ToolInput json.RawMessage
	ToolID    string
	Usage     *TokenUsage
	Error     error
}

// TokenUsage records token counts from an LLM call.
type TokenUsage struct {
	InputTokens  int
	OutputTokens int
}

// ToolDefinition describes a tool the LLM can call.
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// StreamingMessage represents a message in a
// multi-turn conversation for the streaming API.
type StreamingMessage struct {
	Role       string          `json:"role"`
	Content    string          `json:"content,omitempty"`
	ToolUseID  string          `json:"tool_use_id,omitempty"`
	ToolName   string          `json:"tool_name,omitempty"`
	ToolInput  json.RawMessage `json:"tool_input,omitempty"`
	ToolResult json.RawMessage `json:"tool_result,omitempty"`
}

// StreamingRequest contains all parameters for a
// streaming LLM call with tool use.
type StreamingRequest struct {
	SystemPrompt string
	Messages     []StreamingMessage
	Tools        []ToolDefinition
	MaxTokens    int
	Temperature  float64
}

// StreamingProvider streams LLM responses with tool
// use support. Implementors return a channel that
// emits StreamEvent values until the response
// completes.
type StreamingProvider interface {
	CompleteStream(
		ctx context.Context,
		req StreamingRequest,
	) (<-chan StreamEvent, error)
}
