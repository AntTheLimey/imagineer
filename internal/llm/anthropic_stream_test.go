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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnthropicStreamTextResponse(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type",
					"text/event-stream")
				flusher := w.(http.Flusher)

				fmt.Fprintf(w,
					"event: content_block_start\n"+
						"data: {\"index\":0,"+
						"\"content_block\":{\"type\":"+
						"\"text\",\"text\":\"\"}}\n\n")
				flusher.Flush()

				fmt.Fprintf(w,
					"event: content_block_delta\n"+
						"data: {\"index\":0,\"delta\":"+
						"{\"type\":\"text_delta\","+
						"\"text\":\"Hello\"}}\n\n")
				flusher.Flush()

				fmt.Fprintf(w,
					"event: message_delta\n"+
						"data: {\"delta\":{\"stop_reason"+
						"\":\"end_turn\"},\"usage\":"+
						"{\"output_tokens\":5}}\n\n")
				flusher.Flush()

				fmt.Fprintf(w,
					"event: message_stop\n"+
						"data: {}\n\n")
				flusher.Flush()
			},
		),
	)
	defer server.Close()

	provider := &AnthropicStreamProvider{
		apiKey:  "test-key",
		baseURL: server.URL,
		client:  server.Client(),
	}

	ch, err := provider.CompleteStream(
		context.Background(),
		StreamingRequest{
			SystemPrompt: "Test",
			Messages: []StreamingMessage{
				{Role: "user", Content: "Hi"},
			},
			MaxTokens:   100,
			Temperature: 0.7,
		},
	)
	require.NoError(t, err)

	var events []StreamEvent
	for ev := range ch {
		events = append(events, ev)
	}

	hasText := false
	hasDone := false
	for _, ev := range events {
		if ev.Type == EventTextDelta {
			hasText = true
			assert.Equal(t, "Hello", ev.Text)
		}
		if ev.Type == EventDone {
			hasDone = true
		}
	}
	assert.True(t, hasText, "expected text delta event")
	assert.True(t, hasDone, "expected done event")
}

func TestAnthropicStreamToolUseResponse(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type",
					"text/event-stream")
				flusher := w.(http.Flusher)

				fmt.Fprintf(w,
					"event: content_block_start\n"+
						"data: {\"index\":0,"+
						"\"content_block\":{\"type\":"+
						"\"tool_use\",\"id\":"+
						"\"toolu_123\",\"name\":"+
						"\"search_entities\","+
						"\"input\":{}}}\n\n")
				flusher.Flush()

				fmt.Fprintf(w,
					"event: content_block_delta\n"+
						"data: {\"index\":0,\"delta\":"+
						"{\"type\":\"input_json_delta\","+
						"\"partial_json\":"+
						"\"{\\\"query\\\":\\\"inn\\\"}\""+
						"}}\n\n")
				flusher.Flush()

				fmt.Fprintf(w,
					"event: content_block_stop\n"+
						"data: {\"index\":0}\n\n")
				flusher.Flush()

				fmt.Fprintf(w,
					"event: message_delta\n"+
						"data: {\"delta\":{\"stop_reason"+
						"\":\"tool_use\"},\"usage\":"+
						"{\"output_tokens\":20}}\n\n")
				flusher.Flush()

				fmt.Fprintf(w,
					"event: message_stop\n"+
						"data: {}\n\n")
				flusher.Flush()
			},
		),
	)
	defer server.Close()

	provider := &AnthropicStreamProvider{
		apiKey:  "test-key",
		baseURL: server.URL,
		client:  server.Client(),
	}

	ch, err := provider.CompleteStream(
		context.Background(),
		StreamingRequest{
			SystemPrompt: "Test",
			Messages: []StreamingMessage{
				{Role: "user", Content: "Find inns"},
			},
			Tools: []ToolDefinition{
				{
					Name:        "search_entities",
					Description: "Search entities",
				},
			},
			MaxTokens:   100,
			Temperature: 0.7,
		},
	)
	require.NoError(t, err)

	var events []StreamEvent
	for ev := range ch {
		events = append(events, ev)
	}

	hasToolUse := false
	for _, ev := range events {
		if ev.Type == EventToolUse {
			hasToolUse = true
			assert.Equal(t, "search_entities",
				ev.ToolName)
			assert.Equal(t, "toolu_123", ev.ToolID)
		}
	}
	assert.True(t, hasToolUse,
		"expected tool use event")
}
