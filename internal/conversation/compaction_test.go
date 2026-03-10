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
	"testing"

	"github.com/antonypegg/imagineer/internal/models"
	"github.com/stretchr/testify/assert"
)

func intPtr(v int) *int { return &v }

func TestShouldCompact(t *testing.T) {
	orch := NewOrchestrator(OrchestratorConfig{
		Provider: &mockStreamingProvider{},
	})
	defer orch.Stop()

	messages := []models.Message{
		{Tokens: intPtr(3000)},
		{Tokens: intPtr(3000)},
		{Tokens: intPtr(3000)},
	}

	assert.True(t, orch.ShouldCompact(messages),
		"9000 tokens should exceed threshold of 8000")
}

func TestShouldCompactBelowThreshold(t *testing.T) {
	orch := NewOrchestrator(OrchestratorConfig{
		Provider: &mockStreamingProvider{},
	})
	defer orch.Stop()

	messages := []models.Message{
		{Tokens: intPtr(2000)},
		{Tokens: intPtr(2000)},
	}

	assert.False(t, orch.ShouldCompact(messages),
		"4000 tokens should not exceed threshold")
}

func TestShouldCompactNilTokens(t *testing.T) {
	orch := NewOrchestrator(OrchestratorConfig{
		Provider: &mockStreamingProvider{},
	})
	defer orch.Stop()

	messages := []models.Message{
		{Tokens: nil},
		{Tokens: intPtr(5000)},
		{Tokens: nil},
		{Tokens: intPtr(2000)},
	}

	assert.False(t, orch.ShouldCompact(messages),
		"nil tokens treated as 0; total 7000 < 8000")
}

func TestShouldCompactEmpty(t *testing.T) {
	orch := NewOrchestrator(OrchestratorConfig{
		Provider: &mockStreamingProvider{},
	})
	defer orch.Stop()

	assert.False(t, orch.ShouldCompact(nil),
		"nil message list should return false")
	assert.False(t,
		orch.ShouldCompact([]models.Message{}),
		"empty message list should return false")
}

func TestShouldCompactExactThreshold(t *testing.T) {
	orch := NewOrchestrator(OrchestratorConfig{
		Provider: &mockStreamingProvider{},
	})
	defer orch.Stop()

	messages := []models.Message{
		{Tokens: intPtr(4000)},
		{Tokens: intPtr(4000)},
	}

	assert.False(t, orch.ShouldCompact(messages),
		"exactly 8000 tokens should return false; "+
			"must EXCEED threshold")
}

func TestCompactTooFewMessages(t *testing.T) {
	// Compact requires a DB to load context, so when
	// DB is nil it returns an error. This test verifies
	// the nil-DB guard fires before we attempt the
	// "too few messages" check.
	orch := NewOrchestrator(OrchestratorConfig{
		Provider: &mockStreamingProvider{},
	})
	defer orch.Stop()

	err := orch.Compact(
		t.Context(), 1, 1, 1)
	assert.Error(t, err,
		"should return error when DB is nil")
	assert.Contains(t, err.Error(),
		"database is required",
		"error should mention missing database")
}

func TestBuildCompactionPrompt(t *testing.T) {
	toolName := "search_entities"
	messages := []models.Message{
		{
			Role:    models.MessageRoleUser,
			Content: "Tell me about the NPCs.",
		},
		{
			Role:    models.MessageRoleAssistant,
			Content: "There are several NPCs in your campaign.",
		},
		{
			Role:     models.MessageRoleToolCall,
			ToolName: &toolName,
		},
		{
			Role:    models.MessageRoleUser,
			Content: "Create a new NPC named Aldric.",
		},
	}

	prompt := BuildCompactionPrompt(
		"Previously, the GM set up the campaign.",
		messages,
	)

	assert.Contains(t, prompt,
		"## Existing Summary",
		"should include existing summary heading")
	assert.Contains(t, prompt,
		"Previously, the GM set up the campaign.",
		"should include existing summary text")
	assert.Contains(t, prompt,
		"Merge the following conversation",
		"should instruct merging with existing summary")
	assert.Contains(t, prompt,
		"## Conversation to Summarise",
		"should include conversation heading")
	assert.Contains(t, prompt,
		"**GM**: Tell me about the NPCs.",
		"should format user messages as GM")
	assert.Contains(t, prompt,
		"**Assistant**: There are several NPCs",
		"should format assistant messages")
	assert.Contains(t, prompt,
		"**Tool Call** (search_entities)",
		"should format tool call messages")
	assert.Contains(t, prompt,
		"**GM**: Create a new NPC named Aldric.",
		"should include all user messages")
}

func TestBuildCompactionPromptNoExistingSummary(
	t *testing.T,
) {
	messages := []models.Message{
		{
			Role:    models.MessageRoleUser,
			Content: "Hello, let's plan a session.",
		},
		{
			Role:    models.MessageRoleAssistant,
			Content: "Sure! What would you like to plan?",
		},
	}

	prompt := BuildCompactionPrompt("", messages)

	assert.NotContains(t, prompt,
		"## Existing Summary",
		"should not include existing summary heading")
	assert.NotContains(t, prompt,
		"Merge the following",
		"should not include merge instruction")
	assert.Contains(t, prompt,
		"## Conversation to Summarise",
		"should include conversation heading")
	assert.Contains(t, prompt,
		"**GM**: Hello, let's plan a session.",
		"should include user message")
	assert.Contains(t, prompt,
		"**Assistant**: Sure! What would you like to plan?",
		"should include assistant message")
}

func TestBuildCompactionPromptToolResult(t *testing.T) {
	toolName := "get_entity"
	messages := []models.Message{
		{
			Role:     models.MessageRoleToolResult,
			ToolName: &toolName,
		},
	}

	prompt := BuildCompactionPrompt("", messages)

	assert.Contains(t, prompt,
		"**Tool Result** (get_entity)",
		"should format tool result with tool name")
}

func TestCompactionConstants(t *testing.T) {
	assert.Equal(t, 8000, CompactionThreshold,
		"threshold should be 8000 tokens")
	assert.Equal(t, 4, minCompactableMessages,
		"minimum compactable messages should be 4")
}
