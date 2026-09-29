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
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSEWriterTextDelta(t *testing.T) {
	w := httptest.NewRecorder()
	sse := NewSSEWriter(w)
	err := sse.WriteTextDelta("Hello world")
	require.NoError(t, err)
	assert.Contains(t, w.Body.String(),
		"event: text_delta")
	assert.Contains(t, w.Body.String(),
		`"text":"Hello world"`)
}

func TestSSEWriterToolUse(t *testing.T) {
	w := httptest.NewRecorder()
	sse := NewSSEWriter(w)
	err := sse.WriteToolUse("search_entities",
		json.RawMessage(`{"query":"inn"}`))
	require.NoError(t, err)
	assert.Contains(t, w.Body.String(),
		"event: tool_use")
	assert.Contains(t, w.Body.String(),
		"search_entities")
}

func TestSSEWriterToolResult(t *testing.T) {
	w := httptest.NewRecorder()
	sse := NewSSEWriter(w)
	err := sse.WriteToolResult("search_entities",
		json.RawMessage(`{"count":3}`))
	require.NoError(t, err)
	body := w.Body.String()
	assert.Contains(t, body, "event: tool_result")
	assert.Contains(t, body,
		`"tool":"search_entities"`)
	assert.Contains(t, body, `"count":3`)
}

func TestSSEWriterDone(t *testing.T) {
	w := httptest.NewRecorder()
	sse := NewSSEWriter(w)
	err := sse.WriteDone(42, 1200, 380)
	require.NoError(t, err)
	body := w.Body.String()
	assert.Contains(t, body, "event: done")
	assert.Contains(t, body, `"message_id":42`)
	assert.Contains(t, body, `"input_tokens":1200`)
	assert.Contains(t, body, `"output_tokens":380`)
}

func TestSSEWriterError(t *testing.T) {
	w := httptest.NewRecorder()
	sse := NewSSEWriter(w)
	err := sse.WriteError("something went wrong")
	require.NoError(t, err)
	body := w.Body.String()
	assert.Contains(t, body, "event: error")
	assert.Contains(t, body,
		`"error":"something went wrong"`)
}

func TestSSEWriterSetHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	sse := NewSSEWriter(w)
	sse.SetHeaders()
	h := w.Header()
	assert.Equal(t, "text/event-stream",
		h.Get("Content-Type"))
	assert.Equal(t, "no-cache",
		h.Get("Cache-Control"))
	assert.Equal(t, "keep-alive",
		h.Get("Connection"))
	assert.Equal(t, "no",
		h.Get("X-Accel-Buffering"))
}

func TestSSEWriterMultipleEvents(t *testing.T) {
	w := httptest.NewRecorder()
	sse := NewSSEWriter(w)

	require.NoError(t,
		sse.WriteTextDelta("chunk one"))
	require.NoError(t,
		sse.WriteTextDelta("chunk two"))
	require.NoError(t,
		sse.WriteToolUse("lookup",
			json.RawMessage(`{"id":1}`)))
	require.NoError(t,
		sse.WriteDone(99, 500, 200))

	body := w.Body.String()

	// Each event must be separated by a blank line
	// (double newline after each event).
	events := strings.Split(
		strings.TrimSpace(body), "\n\n")
	assert.Len(t, events, 4,
		"expected four separate events")

	// Verify event order.
	assert.Contains(t, events[0], "event: text_delta")
	assert.Contains(t, events[0], "chunk one")
	assert.Contains(t, events[1], "event: text_delta")
	assert.Contains(t, events[1], "chunk two")
	assert.Contains(t, events[2], "event: tool_use")
	assert.Contains(t, events[3], "event: done")
}

func TestSSEWriterSpecialCharacters(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "quotes",
			text: `He said "hello"`,
			want: `He said \"hello\"`,
		},
		{
			name: "newlines",
			text: "line1\nline2",
			want: `line1\nline2`,
		},
		{
			name: "unicode",
			text: "cafe\u0301 \u2603 \U0001F600",
			want: "cafe\u0301 \u2603 \U0001F600",
		},
		{
			name: "backslash",
			text: `path\to\file`,
			want: `path\\to\\file`,
		},
		{
			name: "tabs",
			text: "col1\tcol2",
			want: `col1\tcol2`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			sse := NewSSEWriter(w)
			err := sse.WriteTextDelta(tc.text)
			require.NoError(t, err)

			body := w.Body.String()
			assert.Contains(t, body,
				"event: text_delta")
			assert.Contains(t, body, tc.want)

			// Verify the data line is valid JSON.
			lines := strings.Split(body, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "data: ") {
					raw := strings.TrimPrefix(
						line, "data: ")
					var parsed textDeltaData
					err := json.Unmarshal(
						[]byte(raw), &parsed)
					require.NoError(t, err,
						"data line must be valid JSON")
					assert.Equal(t, tc.text,
						parsed.Text)
				}
			}
		})
	}
}
