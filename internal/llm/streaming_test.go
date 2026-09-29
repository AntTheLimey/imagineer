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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStreamEventTypes(t *testing.T) {
	// Verify distinct values
	types := []StreamEventType{
		EventTextDelta,
		EventToolUse,
		EventUsage,
		EventDone,
		EventError,
	}
	seen := make(map[StreamEventType]bool)
	for _, et := range types {
		assert.False(t, seen[et],
			"duplicate event type: %d", et)
		seen[et] = true
	}
}

func TestStreamingRequestDefaults(t *testing.T) {
	req := StreamingRequest{
		SystemPrompt: "You are a helpful assistant.",
		MaxTokens:    4096,
		Temperature:  0.7,
	}
	assert.Equal(t, 4096, req.MaxTokens)
	assert.Equal(t, 0.7, req.Temperature)
	assert.Empty(t, req.Messages)
	assert.Empty(t, req.Tools)
}
