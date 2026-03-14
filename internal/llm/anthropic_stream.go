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
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// AnthropicStreamProvider implements the StreamingProvider
// interface for Anthropic's Messages API with SSE streaming.
type AnthropicStreamProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewAnthropicStreamProvider creates a new streaming
// Anthropic provider. The apiKey must be non-empty.
func NewAnthropicStreamProvider(
	apiKey string,
) (*AnthropicStreamProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("anthropic API key is required")
	}
	return &AnthropicStreamProvider{
		apiKey:  apiKey,
		baseURL: anthropicAPIURL,
		client: &http.Client{
			Transport: &http.Transport{
				ResponseHeaderTimeout: 30 * time.Second,
			},
		},
	}, nil
}

// anthropicStreamRequest is the request body sent to the
// Anthropic Messages API with streaming enabled.
type anthropicStreamRequest struct {
	Model       string                    `json:"model"`
	MaxTokens   int                       `json:"max_tokens"`
	System      string                    `json:"system,omitempty"`
	Messages    []anthropicStreamMessage  `json:"messages"`
	Tools       []anthropicToolDefinition `json:"tools,omitempty"`
	Stream      bool                      `json:"stream"`
	Temperature float64                   `json:"temperature"`
}

// anthropicStreamMessage represents a message in the
// Anthropic API format, supporting both simple string
// content and structured content blocks.
type anthropicStreamMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

// anthropicContentBlock represents a typed content block
// in the Anthropic API message format.
type anthropicContentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   string          `json:"content,omitempty"`
}

// anthropicToolDefinition is the Anthropic API format
// for tool definitions.
type anthropicToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema,omitempty"`
}

// sseContentBlockStart is the parsed data payload for a
// content_block_start SSE event.
type sseContentBlockStart struct {
	Index        int `json:"index"`
	ContentBlock struct {
		Type  string          `json:"type"`
		ID    string          `json:"id,omitempty"`
		Name  string          `json:"name,omitempty"`
		Text  string          `json:"text,omitempty"`
		Input json.RawMessage `json:"input,omitempty"`
	} `json:"content_block"`
}

// sseContentBlockDelta is the parsed data payload for a
// content_block_delta SSE event.
type sseContentBlockDelta struct {
	Index int `json:"index"`
	Delta struct {
		Type        string `json:"type"`
		Text        string `json:"text,omitempty"`
		PartialJSON string `json:"partial_json,omitempty"`
	} `json:"delta"`
}

// sseMessageStart is the parsed data payload for a
// message_start SSE event.
type sseMessageStart struct {
	Message struct {
		Usage struct {
			InputTokens int `json:"input_tokens"`
		} `json:"usage"`
	} `json:"message"`
}

// sseMessageDelta is the parsed data payload for a
// message_delta SSE event.
type sseMessageDelta struct {
	Delta struct {
		StopReason string `json:"stop_reason"`
	} `json:"delta"`
	Usage struct {
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// CompleteStream sends a streaming request to the Anthropic
// Messages API and returns a channel of StreamEvent values.
// The channel is closed when the response ends or an error
// occurs.
func (p *AnthropicStreamProvider) CompleteStream(
	ctx context.Context,
	req StreamingRequest,
) (<-chan StreamEvent, error) {
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}

	messages := formatStreamMessages(req.Messages)

	body := anthropicStreamRequest{
		Model:       anthropicModel,
		MaxTokens:   maxTokens,
		System:      req.SystemPrompt,
		Messages:    messages,
		Stream:      true,
		Temperature: req.Temperature,
	}

	if len(req.Tools) > 0 {
		tools := make(
			[]anthropicToolDefinition, len(req.Tools),
		)
		for i, t := range req.Tools {
			tools[i] = anthropicToolDefinition{
				Name:        t.Name,
				Description: t.Description,
				InputSchema: t.InputSchema,
			}
		}
		body.Tools = tools
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to marshal request: %w", err,
		)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx, http.MethodPost,
		p.baseURL, bytes.NewReader(payload),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create request: %w", err,
		)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", anthropicAPIVersion)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		var bodyText strings.Builder
		for scanner.Scan() {
			bodyText.WriteString(scanner.Text())
		}
		return nil, fmt.Errorf(
			"anthropic API error (status %d): %s",
			resp.StatusCode, bodyText.String(),
		)
	}

	ch := make(chan StreamEvent, 16)
	go p.readSSEStream(ctx, resp, ch)

	return ch, nil
}

// readSSEStream reads SSE events from the HTTP response
// body and sends parsed StreamEvent values on the channel.
// It closes both the response body and the channel when
// finished.
func (p *AnthropicStreamProvider) readSSEStream(
	ctx context.Context,
	resp *http.Response,
	ch chan<- StreamEvent,
) {
	defer close(ch)
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	// Increase buffer size for large SSE payloads.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var currentEvent string

	// Track current content block state.
	var blockType string
	var toolName string
	var toolID string
	var toolInputJSON strings.Builder

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			ch <- StreamEvent{
				Type:  EventError,
				Error: ctx.Err(),
			}
			return
		default:
		}

		line := scanner.Text()

		if strings.HasPrefix(line, "event: ") {
			currentEvent = strings.TrimPrefix(
				line, "event: ",
			)
			continue
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		switch currentEvent {
		case "message_start":
			var msg sseMessageStart
			if err := json.Unmarshal(
				[]byte(data), &msg,
			); err != nil {
				ch <- StreamEvent{
					Type:  EventError,
					Error: fmt.Errorf("parse message_start: %w", err),
				}
				continue
			}
			if msg.Message.Usage.InputTokens > 0 {
				ch <- StreamEvent{
					Type: EventUsage,
					Usage: &TokenUsage{
						InputTokens: msg.Message.Usage.InputTokens,
					},
				}
			}

		case "content_block_start":
			var block sseContentBlockStart
			if err := json.Unmarshal(
				[]byte(data), &block,
			); err != nil {
				ch <- StreamEvent{
					Type:  EventError,
					Error: fmt.Errorf("parse content_block_start: %w", err),
				}
				continue
			}
			blockType = block.ContentBlock.Type
			if blockType == "tool_use" {
				toolName = block.ContentBlock.Name
				toolID = block.ContentBlock.ID
				toolInputJSON.Reset()
			}

		case "content_block_delta":
			var delta sseContentBlockDelta
			if err := json.Unmarshal(
				[]byte(data), &delta,
			); err != nil {
				continue
			}
			switch delta.Delta.Type {
			case "text_delta":
				ch <- StreamEvent{
					Type: EventTextDelta,
					Text: delta.Delta.Text,
				}
			case "input_json_delta":
				toolInputJSON.WriteString(
					delta.Delta.PartialJSON,
				)
			}

		case "content_block_stop":
			if blockType == "tool_use" {
				ch <- StreamEvent{
					Type:      EventToolUse,
					ToolName:  toolName,
					ToolID:    toolID,
					ToolInput: json.RawMessage(toolInputJSON.String()),
				}
				blockType = ""
				toolName = ""
				toolID = ""
				toolInputJSON.Reset()
			}

		case "message_delta":
			var md sseMessageDelta
			if err := json.Unmarshal(
				[]byte(data), &md,
			); err != nil {
				ch <- StreamEvent{
					Type:  EventError,
					Error: fmt.Errorf("parse message_delta: %w", err),
				}
				continue
			}
			if md.Usage.OutputTokens > 0 {
				ch <- StreamEvent{
					Type: EventUsage,
					Usage: &TokenUsage{
						OutputTokens: md.Usage.OutputTokens,
					},
				}
			}

		case "message_stop":
			ch <- StreamEvent{Type: EventDone}
			return
		}
	}

	if err := scanner.Err(); err != nil {
		ch <- StreamEvent{
			Type:  EventError,
			Error: fmt.Errorf("SSE read error: %w", err),
		}
	}
}

// formatStreamMessages converts StreamingMessage values
// into the Anthropic API message format, handling both
// simple text messages and tool use/result content blocks.
func formatStreamMessages(
	msgs []StreamingMessage,
) []anthropicStreamMessage {
	result := make([]anthropicStreamMessage, 0, len(msgs))

	for _, m := range msgs {
		switch {
		case m.ToolResult != nil && m.ToolUseID != "":
			// Tool result message: user role with
			// tool_result content block.
			block := anthropicContentBlock{
				Type:      "tool_result",
				ToolUseID: m.ToolUseID,
				Content:   string(m.ToolResult),
			}
			result = append(result, anthropicStreamMessage{
				Role:    "user",
				Content: []anthropicContentBlock{block},
			})

		case m.ToolName != "" && m.ToolUseID != "":
			// Tool use message: assistant role with
			// tool_use content block.
			block := anthropicContentBlock{
				Type:  "tool_use",
				ID:    m.ToolUseID,
				Name:  m.ToolName,
				Input: m.ToolInput,
			}
			result = append(result, anthropicStreamMessage{
				Role:    "assistant",
				Content: []anthropicContentBlock{block},
			})

		default:
			// Simple text message.
			result = append(result, anthropicStreamMessage{
				Role:    m.Role,
				Content: m.Content,
			})
		}
	}

	return result
}
