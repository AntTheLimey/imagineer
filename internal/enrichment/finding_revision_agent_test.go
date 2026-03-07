/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package enrichment

import (
	"context"
	"errors"
	"testing"

	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// findingMockProvider implements llm.Provider for finding-revision tests.
// Separate from the package-level mockProvider to keep tests self-contained.
// ---------------------------------------------------------------------------

type findingMockProvider struct {
	response string
	err      error
	captured *llm.CompletionRequest // set on each call if non-nil
}

func (m *findingMockProvider) Complete(
	ctx context.Context,
	req llm.CompletionRequest,
) (llm.CompletionResponse, error) {
	if m.captured != nil {
		*m.captured = req
	}
	if m.err != nil {
		return llm.CompletionResponse{}, m.err
	}
	return llm.CompletionResponse{Content: m.response}, nil
}

// ---------------------------------------------------------------------------
// GenerateFindingRevision tests
// ---------------------------------------------------------------------------

func TestGenerateFindingRevision_Basic(t *testing.T) {
	provider := &findingMockProvider{
		response: "The dragon swooped down, its crimson scales glinting in the fading light.",
	}

	input := FindingRevisionInput{
		ContextWindow: "The dragon swooped down, its scales glinting in the light.",
		Findings: []FindingDetail{
			{
				DetectionType: "style_suggestion",
				MatchedText:   "its scales glinting in the light",
				Description:   "The description lacks vivid detail.",
				Suggestion:    "Add colour and atmosphere to the dragon description.",
			},
		},
	}

	result, err := GenerateFindingRevision(context.Background(), provider, input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Contains(t, result.RevisedSection, "crimson scales")
}

func TestGenerateFindingRevision_EmptyContext(t *testing.T) {
	provider := &findingMockProvider{
		response: "should not be called",
	}

	input := FindingRevisionInput{
		ContextWindow: "",
		Findings: []FindingDetail{
			{
				DetectionType: "spelling",
				MatchedText:   "teh",
				Description:   "Typo detected.",
				Suggestion:    "Replace with 'the'.",
			},
		},
	}

	result, err := GenerateFindingRevision(context.Background(), provider, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "context window is required")
}

func TestGenerateFindingRevision_NoFindings(t *testing.T) {
	// When there are no findings, the context should be returned unchanged
	// without calling the LLM provider.
	provider := &findingMockProvider{
		err: errors.New("should not be called"),
	}

	originalContext := "The investigators crept through the darkened hallway."
	input := FindingRevisionInput{
		ContextWindow: originalContext,
		Findings:      []FindingDetail{},
	}

	result, err := GenerateFindingRevision(context.Background(), provider, input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, originalContext, result.RevisedSection)
}

func TestGenerateFindingRevision_NilFindings(t *testing.T) {
	// nil findings should behave the same as empty.
	provider := &findingMockProvider{
		err: errors.New("should not be called"),
	}

	originalContext := "The manor loomed ahead."
	input := FindingRevisionInput{
		ContextWindow: originalContext,
		Findings:      nil,
	}

	result, err := GenerateFindingRevision(context.Background(), provider, input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, originalContext, result.RevisedSection)
}

func TestGenerateFindingRevision_WithInstructions(t *testing.T) {
	// Verify that GM instructions reach the prompt sent to the LLM.
	var captured llm.CompletionRequest
	provider := &findingMockProvider{
		response: "revised text",
		captured: &captured,
	}

	input := FindingRevisionInput{
		ContextWindow: "The NPC said hello.",
		Findings: []FindingDetail{
			{
				DetectionType: "tone_mismatch",
				MatchedText:   "said hello",
				Description:   "Tone is too casual for the setting.",
				Suggestion:    "Use more formal language.",
			},
		},
		Instructions: "Keep the Victorian English tone throughout. Do not modernise dialogue.",
	}

	result, err := GenerateFindingRevision(context.Background(), provider, input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Contains(t, captured.UserPrompt, "GM Instructions")
	assert.Contains(t, captured.UserPrompt, "Victorian English tone")
	assert.Contains(t, captured.UserPrompt, "Do not modernise dialogue")
}

func TestGenerateFindingRevision_LLMError(t *testing.T) {
	provider := &findingMockProvider{
		err: errors.New("API rate limit exceeded"),
	}

	input := FindingRevisionInput{
		ContextWindow: "Some text to revise.",
		Findings: []FindingDetail{
			{
				DetectionType: "grammar",
				MatchedText:   "Some text",
				Description:   "Grammar issue.",
				Suggestion:    "Fix grammar.",
			},
		},
	}

	result, err := GenerateFindingRevision(context.Background(), provider, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "LLM completion failed")
}

func TestGenerateFindingRevision_VerifiesProviderRequest(t *testing.T) {
	// Verify MaxTokens and Temperature are set correctly.
	var captured llm.CompletionRequest
	provider := &findingMockProvider{
		response: "revised",
		captured: &captured,
	}

	input := FindingRevisionInput{
		ContextWindow: "Some context window text.",
		Findings: []FindingDetail{
			{
				DetectionType: "spelling",
				MatchedText:   "teh",
				Description:   "Typo.",
				Suggestion:    "the",
			},
		},
	}

	_, err := GenerateFindingRevision(context.Background(), provider, input)

	require.NoError(t, err)
	assert.Equal(t, 1024, captured.MaxTokens)
	assert.InDelta(t, 0.4, captured.Temperature, 0.001)
}

// ---------------------------------------------------------------------------
// buildFindingSystemPrompt tests
// ---------------------------------------------------------------------------

func TestBuildFindingSystemPrompt(t *testing.T) {
	prompt := buildFindingSystemPrompt()

	// The system prompt should instruct the LLM to perform surgical editing.
	assert.Contains(t, prompt, "surgical")
	assert.Contains(t, prompt, "plain text")
	assert.Contains(t, prompt, "TTRPG")
	// It should warn the LLM to preserve surrounding text.
	assert.Contains(t, prompt, "Preserve")
}

// ---------------------------------------------------------------------------
// buildFindingUserPrompt tests
// ---------------------------------------------------------------------------

func TestBuildFindingUserPrompt(t *testing.T) {
	input := FindingRevisionInput{
		ContextWindow: "The rogue picked the lock with ease.",
		Findings: []FindingDetail{
			{
				DetectionType: "mechanics_warning",
				MatchedText:   "picked the lock with ease",
				Description:   "Lock-picking should require a skill check.",
				Suggestion:    "Add a Dexterity check or mention the rogue's proficiency.",
			},
		},
		Instructions: "Use D&D 5e mechanics.",
	}

	prompt := buildFindingUserPrompt(input)

	// Context window section.
	assert.Contains(t, prompt, "Context Window")
	assert.Contains(t, prompt, "The rogue picked the lock with ease.")

	// Finding details.
	assert.Contains(t, prompt, "mechanics_warning")
	assert.Contains(t, prompt, "picked the lock with ease")
	assert.Contains(t, prompt, "Lock-picking should require a skill check.")
	assert.Contains(t, prompt, "Add a Dexterity check")

	// GM Instructions section.
	assert.Contains(t, prompt, "GM Instructions")
	assert.Contains(t, prompt, "Use D&D 5e mechanics.")
}

func TestBuildFindingUserPrompt_NoInstructions(t *testing.T) {
	input := FindingRevisionInput{
		ContextWindow: "A quiet village square.",
		Findings: []FindingDetail{
			{
				DetectionType: "atmosphere",
				MatchedText:   "quiet village square",
				Description:   "Lacks sensory detail.",
				Suggestion:    "Add sounds and smells.",
			},
		},
		Instructions: "",
	}

	prompt := buildFindingUserPrompt(input)

	// GM Instructions section should not appear when instructions are empty.
	assert.NotContains(t, prompt, "GM Instructions")
	// But context and findings should still be present.
	assert.Contains(t, prompt, "Context Window")
	assert.Contains(t, prompt, "quiet village square")
	assert.Contains(t, prompt, "atmosphere")
}

func TestBuildFindingUserPrompt_MultipleFindings(t *testing.T) {
	input := FindingRevisionInput{
		ContextWindow: "The wizard cast fireball. The rogue dodged the trap.",
		Findings: []FindingDetail{
			{
				DetectionType: "mechanics_warning",
				MatchedText:   "cast fireball",
				Description:   "Fireball requires a spell slot.",
				Suggestion:    "Mention the spell slot expenditure.",
			},
			{
				DetectionType: "pacing_note",
				MatchedText:   "dodged the trap",
				Description:   "The trap evasion is too abrupt.",
				Suggestion:    "Add tension with a Dexterity saving throw description.",
			},
			{
				DetectionType: "continuity_error",
				MatchedText:   "The wizard",
				Description:   "The wizard was previously described as unconscious.",
				Suggestion:    "Note the wizard's recovery or adjust the scene.",
			},
		},
	}

	prompt := buildFindingUserPrompt(input)

	// Numbered findings should appear.
	assert.Contains(t, prompt, "Finding 1")
	assert.Contains(t, prompt, "Finding 2")
	assert.Contains(t, prompt, "Finding 3")

	// Each finding's details should be present.
	assert.Contains(t, prompt, "mechanics_warning")
	assert.Contains(t, prompt, "pacing_note")
	assert.Contains(t, prompt, "continuity_error")
	assert.Contains(t, prompt, "Fireball requires a spell slot.")
	assert.Contains(t, prompt, "The trap evasion is too abrupt.")
	assert.Contains(t, prompt, "previously described as unconscious")
}
