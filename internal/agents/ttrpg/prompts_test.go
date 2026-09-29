/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package ttrpg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// ConversationSystemPrompt tests
// ---------------------------------------------------------------------------

func TestConversationSystemPromptNotEmpty(t *testing.T) {
	prompt := ConversationSystemPrompt()
	assert.NotEmpty(t, prompt)
}

func TestConversationSystemPromptContainsKeyCapabilities(t *testing.T) {
	prompt := ConversationSystemPrompt()
	assert.Contains(t, prompt, "scene",
		"should mention scene design")
	assert.Contains(t, prompt, "NPC",
		"should mention NPC voice generation")
	assert.Contains(t, prompt, "encounter",
		"should mention encounter building")
	assert.Contains(t, prompt, "schema",
		"should mention game system schema validation")
}

func TestConversationSystemPromptContainsWorldGroundedFlow(t *testing.T) {
	prompt := ConversationSystemPrompt()
	assert.Contains(t, prompt, "search_content",
		"should reference search_content tool")
	assert.Contains(t, prompt, "search_entities",
		"should reference search_entities tool")
	assert.Contains(t, prompt, "read_game_schema",
		"should reference read_game_schema tool")
	assert.Contains(t, prompt, "orphaned",
		"should mention orphaned entities")
}

func TestConversationSystemPromptIsConversational(t *testing.T) {
	prompt := ConversationSystemPrompt()
	assert.Contains(t, prompt, "conversationally",
		"should instruct conversational responses")
	assert.NotContains(t, prompt, "JSON",
		"should not instruct JSON output")
}

func TestConversationSystemPromptContainsStatBlockGuidance(t *testing.T) {
	prompt := ConversationSystemPrompt()
	assert.Contains(t, prompt, "stat block",
		"should mention stat block generation")
	assert.Contains(t, prompt, "Reconciliation",
		"should mention prep vs play reconciliation")
}
