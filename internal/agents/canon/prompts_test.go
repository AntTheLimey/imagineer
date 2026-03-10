/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package canon

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
	assert.Contains(t, prompt, "Fact verification",
		"should mention fact verification")
	assert.Contains(t, prompt, "Knowledge compilation",
		"should mention knowledge compilation")
	assert.Contains(t, prompt, "Hypothetical checking",
		"should mention hypothetical checking")
	assert.Contains(t, prompt, "Conflict resolution",
		"should mention conflict resolution")
}

func TestConversationSystemPromptContainsConservativePhilosophy(t *testing.T) {
	prompt := ConversationSystemPrompt()
	assert.Contains(t, prompt, "conservative",
		"should mention conservative contradiction detection")
	assert.Contains(t, prompt, "False positives",
		"should warn against false positives")
	assert.Contains(t, prompt, "elaboration",
		"should distinguish elaboration from contradiction")
}

func TestConversationSystemPromptContainsCitationGuidance(t *testing.T) {
	prompt := ConversationSystemPrompt()
	assert.Contains(t, prompt, "cite",
		"should mention citing sources")
	assert.Contains(t, prompt, "AUTHORITATIVE",
		"should mention source confidence levels")
	assert.Contains(t, prompt, "DRAFT",
		"should mention DRAFT confidence level")
	assert.Contains(t, prompt, "SUPERSEDED",
		"should mention SUPERSEDED confidence level")
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
	assert.Contains(t, prompt, "search_content",
		"should reference search_content tool")
	assert.Contains(t, prompt, "search_entities",
		"should reference search_entities tool")
	assert.Contains(t, prompt, "get_entity",
		"should reference get_entity tool")
	assert.Contains(t, prompt, "get_related_entities",
		"should reference get_related_entities tool")
}
