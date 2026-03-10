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

func TestAnthropicStreamUsageTracking(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type",
					"text/event-stream")
				flusher := w.(http.Flusher)

				fmt.Fprintf(w,
					"event: message_start\n"+
						"data: {\"type\":\"message_start\","+
						"\"message\":{\"usage\":{"+
						"\"input_tokens\":42}}}\n\n")
				flusher.Flush()

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
						"\"text\":\"Hi\"}}\n\n")
				flusher.Flush()

				fmt.Fprintf(w,
					"event: message_delta\n"+
						"data: {\"delta\":{\"stop_reason"+
						"\":\"end_turn\"},\"usage\":"+
						"{\"output_tokens\":10}}\n\n")
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

	var inputTokens, outputTokens int
	for ev := range ch {
		if ev.Type == EventUsage && ev.Usage != nil {
			inputTokens += ev.Usage.InputTokens
			outputTokens += ev.Usage.OutputTokens
		}
	}
	assert.Equal(t, 42, inputTokens)
	assert.Equal(t, 10, outputTokens)
}

func TestAnthropicStreamHTTPError(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTooManyRequests)
				fmt.Fprintf(w,
					`{"error":{"message":"rate limited"}}`)
			},
		),
	)
	defer server.Close()

	provider := &AnthropicStreamProvider{
		apiKey:  "test-key",
		baseURL: server.URL,
		client:  server.Client(),
	}

	_, err := provider.CompleteStream(
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
	require.Error(t, err)
	assert.Contains(t, err.Error(), "429")
}

func TestAnthropicStreamContextCancel(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type",
					"text/event-stream")
				flusher := w.(http.Flusher)

				fmt.Fprintf(w,
					"event: content_block_delta\n"+
						"data: {\"index\":0,\"delta\":"+
						"{\"type\":\"text_delta\","+
						"\"text\":\"Hi\"}}\n\n")
				flusher.Flush()

				// Block until client disconnects.
				<-r.Context().Done()
			},
		),
	)
	defer server.Close()

	ctx, cancel := context.WithCancel(
		context.Background())

	provider := &AnthropicStreamProvider{
		apiKey:  "test-key",
		baseURL: server.URL,
		client:  server.Client(),
	}

	ch, err := provider.CompleteStream(ctx,
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

	// Read the text delta event.
	ev := <-ch
	assert.Equal(t, EventTextDelta, ev.Type)

	// Cancel the context.
	cancel()

	// Channel should close eventually.
	for range ch {
		// drain
	}
	// If we get here, the channel closed --
	// context cancellation worked.
}
