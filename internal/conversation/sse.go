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
	"fmt"
	"net/http"
)

// textDeltaData is the JSON payload for text_delta events.
type textDeltaData struct {
	Text string `json:"text"`
}

// toolUseData is the JSON payload for tool_use events.
type toolUseData struct {
	Tool  string          `json:"tool"`
	Input json.RawMessage `json:"input"`
}

// toolResultData is the JSON payload for tool_result events.
type toolResultData struct {
	Tool   string          `json:"tool"`
	Result json.RawMessage `json:"result"`
}

// doneData is the JSON payload for done events.
type doneData struct {
	MessageID    int64 `json:"message_id"`
	InputTokens  int   `json:"input_tokens"`
	OutputTokens int   `json:"output_tokens"`
}

// errorData is the JSON payload for error events.
type errorData struct {
	Error string `json:"error"`
}

// SSEWriter formats and sends Server-Sent Events to an
// HTTP response for conversation streaming.
type SSEWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

// NewSSEWriter creates an SSEWriter for the given
// ResponseWriter. If the writer implements http.Flusher,
// events are flushed immediately after each write.
func NewSSEWriter(w http.ResponseWriter) *SSEWriter {
	flusher, _ := w.(http.Flusher)
	return &SSEWriter{
		w:       w,
		flusher: flusher,
	}
}

// SetHeaders configures the response headers for an SSE
// stream. Call this before writing any events.
func (s *SSEWriter) SetHeaders() {
	h := s.w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")
}

// WriteTextDelta sends a text_delta event containing a
// fragment of the assistant's response text.
func (s *SSEWriter) WriteTextDelta(text string) error {
	return s.writeEvent("text_delta", textDeltaData{
		Text: text,
	})
}

// WriteToolUse sends a tool_use event indicating the LLM
// has invoked a tool with the given input.
func (s *SSEWriter) WriteToolUse(
	name string,
	input json.RawMessage,
) error {
	return s.writeEvent("tool_use", toolUseData{
		Tool:  name,
		Input: input,
	})
}

// WriteToolResult sends a tool_result event containing the
// output from a tool execution.
func (s *SSEWriter) WriteToolResult(
	name string,
	result json.RawMessage,
) error {
	return s.writeEvent("tool_result", toolResultData{
		Tool:   name,
		Result: result,
	})
}

// WriteDone sends a done event signalling the end of the
// response stream, including the message ID and token
// usage counts.
func (s *SSEWriter) WriteDone(
	messageID int64,
	inputTokens, outputTokens int,
) error {
	return s.writeEvent("done", doneData{
		MessageID:    messageID,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
	})
}

// WriteError sends an error event with a human-readable
// error message.
func (s *SSEWriter) WriteError(msg string) error {
	return s.writeEvent("error", errorData{
		Error: msg,
	})
}

// writeEvent marshals data to JSON and writes a single
// SSE event in the format:
//
//	event: <type>\ndata: <json>\n\n
func (s *SSEWriter) writeEvent(
	eventType string,
	data interface{},
) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf(
			"failed to marshal SSE data: %w", err)
	}
	_, err = fmt.Fprintf(s.w,
		"event: %s\ndata: %s\n\n", eventType, jsonData)
	if err != nil {
		return fmt.Errorf(
			"failed to write SSE event: %w", err)
	}
	if s.flusher != nil {
		s.flusher.Flush()
	}
	return nil
}
