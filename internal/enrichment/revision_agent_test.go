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
	"encoding/json"
	"errors"
	"testing"

	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/antonypegg/imagineer/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// RevisionAgent tests
// ---------------------------------------------------------------------------

func TestRevisionAgent_EmptyContent(t *testing.T) {
	ra := NewRevisionAgent()
	provider := &mockProvider{response: ""}

	result, err := ra.GenerateRevision(
		context.Background(),
		provider,
		RevisionInput{
			OriginalContent: "",
			AcceptedItems: []models.ContentAnalysisItem{
				{DetectionType: "pacing_note", MatchedText: "test"},
			},
		},
	)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "original content is required")
}

func TestRevisionAgent_NoAcceptedItems(t *testing.T) {
	// The provider should NOT be called when there are no accepted items.
	ra := NewRevisionAgent()
	provider := &mockProvider{
		err: errors.New("should not be called"),
	}

	originalContent := "The investigators arrived at the manor."
	result, err := ra.GenerateRevision(
		context.Background(),
		provider,
		RevisionInput{
			OriginalContent: originalContent,
			AcceptedItems:   []models.ContentAnalysisItem{},
		},
	)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, originalContent, result.RevisedContent)
	assert.Equal(t, "", result.Summary)
}

func TestRevisionAgent_ValidRevision(t *testing.T) {
	llmResponse := `{
		"revisedContent": "The investigators cautiously approached the manor, their footsteps echoing on the gravel path.",
		"summary": "Added sensory detail to the arrival scene to improve pacing and atmosphere."
	}`

	ra := NewRevisionAgent()
	provider := &mockProvider{response: llmResponse}

	suggestedContent, _ := json.Marshal(map[string]string{
		"description": "The arrival scene lacks sensory detail.",
		"suggestion":  "Add atmospheric description to the approach.",
	})

	result, err := ra.GenerateRevision(
		context.Background(),
		provider,
		RevisionInput{
			OriginalContent: "The investigators arrived at the manor.",
			AcceptedItems: []models.ContentAnalysisItem{
				{
					DetectionType:    "pacing_note",
					MatchedText:      "pacing",
					SuggestedContent: json.RawMessage(suggestedContent),
				},
			},
			SourceTable: "chapters",
			SourceID:    5,
		},
	)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Contains(t, result.RevisedContent, "cautiously approached the manor")
	assert.Contains(t, result.Summary, "sensory detail")
}

func TestRevisionAgent_LLMError(t *testing.T) {
	ra := NewRevisionAgent()
	provider := &mockProvider{
		err: errors.New("API rate limit exceeded"),
	}

	suggestedContent, _ := json.Marshal(map[string]string{
		"description": "A finding",
		"suggestion":  "A suggestion",
	})

	result, err := ra.GenerateRevision(
		context.Background(),
		provider,
		RevisionInput{
			OriginalContent: "Some content to revise.",
			AcceptedItems: []models.ContentAnalysisItem{
				{
					DetectionType:    "content_suggestion",
					MatchedText:      "test",
					SuggestedContent: json.RawMessage(suggestedContent),
				},
			},
		},
	)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "LLM completion failed")
}

func TestRevisionAgent_RawTextFallback(t *testing.T) {
	// LLM returns plain text instead of JSON. The agent should use
	// the raw text as the revised content with an empty summary.
	ra := NewRevisionAgent()
	provider := &mockProvider{
		response: "The investigators cautiously approached the old manor house.",
	}

	suggestedContent, _ := json.Marshal(map[string]string{
		"description": "Needs more detail.",
		"suggestion":  "Add detail.",
	})

	result, err := ra.GenerateRevision(
		context.Background(),
		provider,
		RevisionInput{
			OriginalContent: "The investigators arrived at the manor.",
			AcceptedItems: []models.ContentAnalysisItem{
				{
					DetectionType:    "pacing_note",
					MatchedText:      "pacing",
					SuggestedContent: json.RawMessage(suggestedContent),
				},
			},
		},
	)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "The investigators cautiously approached the old manor house.", result.RevisedContent)
	assert.Equal(t, "", result.Summary)
}

func TestRevisionAgent_MalformedJSON(t *testing.T) {
	// LLM returns invalid JSON. Same behaviour as raw text fallback.
	ra := NewRevisionAgent()
	provider := &mockProvider{
		response: `{"revisedContent": "some text", broken json here}`,
	}

	suggestedContent, _ := json.Marshal(map[string]string{
		"description": "A finding",
		"suggestion":  "Fix it",
	})

	result, err := ra.GenerateRevision(
		context.Background(),
		provider,
		RevisionInput{
			OriginalContent: "Original content.",
			AcceptedItems: []models.ContentAnalysisItem{
				{
					DetectionType:    "content_suggestion",
					MatchedText:      "test",
					SuggestedContent: json.RawMessage(suggestedContent),
				},
			},
		},
	)

	require.NoError(t, err)
	require.NotNil(t, result)
	// Falls back to raw text.
	assert.Contains(t, result.RevisedContent, "revisedContent")
	assert.Equal(t, "", result.Summary)
}

func TestRevisionAgent_IncludesGameSystem(t *testing.T) {
	// Verify that when GameSystemYAML is provided, it appears in the
	// prompt sent to the LLM.
	var captured llm.CompletionRequest
	innerProvider := &mockProvider{
		response: `{"revisedContent": "revised", "summary": "updated"}`,
	}
	captureProvider := &capturingProvider{
		inner:    innerProvider,
		captured: &captured,
	}

	ra := NewRevisionAgent()

	gameYAML := "name: Call of Cthulhu 7e\nskills:\n  - Spot Hidden\n  - Library Use"
	suggestedContent, _ := json.Marshal(map[string]string{
		"description": "A finding",
		"suggestion":  "A suggestion",
	})

	_, err := ra.GenerateRevision(
		context.Background(),
		captureProvider,
		RevisionInput{
			OriginalContent: "The investigators searched the room.",
			AcceptedItems: []models.ContentAnalysisItem{
				{
					DetectionType:    "mechanics_warning",
					MatchedText:      "mechanics",
					SuggestedContent: json.RawMessage(suggestedContent),
				},
			},
			GameSystemYAML: gameYAML,
		},
	)

	require.NoError(t, err)
	assert.Contains(t, captured.UserPrompt, "## Game System Context")
	assert.Contains(t, captured.UserPrompt, "Call of Cthulhu 7e")
	assert.Contains(t, captured.UserPrompt, "Spot Hidden")
}

func TestRevisionAgent_FindingsInPrompt(t *testing.T) {
	// Verify that accepted items' descriptions and suggestions appear
	// in the user prompt sent to the LLM.
	var captured llm.CompletionRequest
	innerProvider := &mockProvider{
		response: `{"revisedContent": "revised", "summary": "updated"}`,
	}
	captureProvider := &capturingProvider{
		inner:    innerProvider,
		captured: &captured,
	}

	ra := NewRevisionAgent()

	finding1Content, _ := json.Marshal(map[string]string{
		"description": "The pacing drags in the middle section.",
		"suggestion":  "Tighten the dialogue and remove redundant exposition.",
	})
	finding2Content, _ := json.Marshal(map[string]string{
		"description": "NPC motivation is unclear.",
		"suggestion":  "Add a line hinting at the NPC's true goal.",
	})

	_, err := ra.GenerateRevision(
		context.Background(),
		captureProvider,
		RevisionInput{
			OriginalContent: "A long chapter about the investigators.",
			AcceptedItems: []models.ContentAnalysisItem{
				{
					DetectionType:    "pacing_note",
					MatchedText:      "pacing",
					SuggestedContent: json.RawMessage(finding1Content),
				},
				{
					DetectionType:    "content_suggestion",
					MatchedText:      "npc_development",
					SuggestedContent: json.RawMessage(finding2Content),
				},
			},
		},
	)

	require.NoError(t, err)

	// Verify both findings appear in the prompt.
	assert.Contains(t, captured.UserPrompt, "The pacing drags in the middle section.")
	assert.Contains(t, captured.UserPrompt, "Tighten the dialogue and remove redundant exposition.")
	assert.Contains(t, captured.UserPrompt, "NPC motivation is unclear.")
	assert.Contains(t, captured.UserPrompt, "Add a line hinting at the NPC's true goal.")

	// Verify detection types appear.
	assert.Contains(t, captured.UserPrompt, "pacing_note")
	assert.Contains(t, captured.UserPrompt, "content_suggestion")

	// Verify original content appears.
	assert.Contains(t, captured.UserPrompt, "A long chapter about the investigators.")

	// Verify finding numbering.
	assert.Contains(t, captured.UserPrompt, "### Finding 1")
	assert.Contains(t, captured.UserPrompt, "### Finding 2")
}

func TestRevisionAgent_NilAcceptedItems(t *testing.T) {
	// nil accepted items should behave the same as empty: return
	// original content unchanged without calling the provider.
	ra := NewRevisionAgent()
	provider := &mockProvider{
		err: errors.New("should not be called"),
	}

	originalContent := "The manor loomed ahead."
	result, err := ra.GenerateRevision(
		context.Background(),
		provider,
		RevisionInput{
			OriginalContent: originalContent,
			AcceptedItems:   nil,
		},
	)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, originalContent, result.RevisedContent)
	assert.Equal(t, "", result.Summary)
}

func TestRevisionAgent_VerifiesProviderRequest(t *testing.T) {
	// Verify that the provider receives the expected request parameters.
	var captured llm.CompletionRequest
	innerProvider := &mockProvider{
		response: `{"revisedContent": "revised text", "summary": "changes"}`,
	}
	captureProvider := &capturingProvider{
		inner:    innerProvider,
		captured: &captured,
	}

	ra := NewRevisionAgent()

	suggestedContent, _ := json.Marshal(map[string]string{
		"description": "A finding",
		"suggestion":  "A suggestion",
	})

	_, err := ra.GenerateRevision(
		context.Background(),
		captureProvider,
		RevisionInput{
			OriginalContent: "Content to revise.",
			AcceptedItems: []models.ContentAnalysisItem{
				{
					DetectionType:    "pacing_note",
					MatchedText:      "test",
					SuggestedContent: json.RawMessage(suggestedContent),
				},
			},
		},
	)

	require.NoError(t, err)
	assert.Equal(t, 8192, captured.MaxTokens)
	assert.InDelta(t, 0.4, captured.Temperature, 0.001)
	assert.Contains(t, captured.SystemPrompt, "TTRPG content editor")
	assert.Contains(t, captured.SystemPrompt, "Respond with valid JSON only")
	assert.Contains(t, captured.UserPrompt, "Content to revise.")
}

func TestRevisionAgent_CodeFencedResponse(t *testing.T) {
	// LLM wraps its JSON response in markdown code fences.
	llmResponse := "```json\n" + `{
		"revisedContent": "The investigators arrived at the dark manor.",
		"summary": "Added atmosphere descriptor."
	}` + "\n```"

	ra := NewRevisionAgent()
	provider := &mockProvider{response: llmResponse}

	suggestedContent, _ := json.Marshal(map[string]string{
		"description": "Needs atmosphere.",
		"suggestion":  "Add descriptors.",
	})

	result, err := ra.GenerateRevision(
		context.Background(),
		provider,
		RevisionInput{
			OriginalContent: "The investigators arrived at the manor.",
			AcceptedItems: []models.ContentAnalysisItem{
				{
					DetectionType:    "setting",
					MatchedText:      "setting",
					SuggestedContent: json.RawMessage(suggestedContent),
				},
			},
		},
	)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "The investigators arrived at the dark manor.", result.RevisedContent)
	assert.Equal(t, "Added atmosphere descriptor.", result.Summary)
}
