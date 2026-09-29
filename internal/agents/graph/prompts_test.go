/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package graph

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
	assert.Contains(t, prompt, "Connection suggestions",
		"should mention connection suggestions")
	assert.Contains(t, prompt, "Path discovery",
		"should mention path discovery")
	assert.Contains(t, prompt, "Impact analysis",
		"should mention impact analysis")
	assert.Contains(t, prompt, "Deduplication",
		"should mention deduplication detection")
}

func TestConversationSystemPromptContainsRelationshipPatterns(t *testing.T) {
	prompt := ConversationSystemPrompt()
	assert.Contains(t, prompt, "Relationship pattern",
		"should mention relationship pattern analysis")
	assert.Contains(t, prompt, "Orphan",
		"should mention orphan detection")
	assert.Contains(t, prompt, "hub",
		"should mention hub entities")
}

func TestConversationSystemPromptUsesGMFriendlyLanguage(t *testing.T) {
	prompt := ConversationSystemPrompt()
	assert.Contains(t, prompt, "GM-Friendly",
		"should have GM-friendly language section")
	assert.Contains(t, prompt, "central figure",
		"should translate hub node to central figure")
	assert.Contains(t, prompt, "linchpin",
		"should translate bridge to linchpin")
}

func TestConversationSystemPromptIsConversational(t *testing.T) {
	prompt := ConversationSystemPrompt()
	assert.Contains(t, prompt, "conversationally",
		"should instruct conversational responses")
	assert.NotContains(t, prompt, "Respond with ONLY valid JSON",
		"should not instruct JSON output")
}

func TestConversationSystemPromptReferencesTools(t *testing.T) {
	prompt := ConversationSystemPrompt()
	assert.Contains(t, prompt, "search_entities",
		"should reference search_entities tool")
	assert.Contains(t, prompt, "get_entity",
		"should reference get_entity tool")
	assert.Contains(t, prompt, "get_related_entities",
		"should reference get_related_entities tool")
}
