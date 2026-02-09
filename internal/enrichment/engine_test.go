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
	"strings"
	"testing"

	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/antonypegg/imagineer/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProvider implements llm.Provider for testing.
type mockProvider struct {
	response string
	err      error
}

func (m *mockProvider) Complete(
	ctx context.Context,
	req llm.CompletionRequest,
) (llm.CompletionResponse, error) {
	if m.err != nil {
		return llm.CompletionResponse{}, m.err
	}
	return llm.CompletionResponse{Content: m.response}, nil
}

// ---------------------------------------------------------------------------
// Parser tests
// ---------------------------------------------------------------------------

func TestParseEnrichmentResponse_Valid(t *testing.T) {
	input := `{
		"descriptionUpdates": [
			{
				"currentDescription": "A mysterious stranger.",
				"suggestedDescription": "A mysterious stranger with a scarred left hand.",
				"rationale": "The session notes mention the scar."
			}
		],
		"logEntries": [
			{
				"content": "Arrived at the tavern and spoke with the party.",
				"occurredAt": "Day 3, evening"
			}
		],
		"relationships": [
			{
				"sourceEntityId": 1,
				"sourceEntityName": "Viktor",
				"targetEntityId": 2,
				"targetEntityName": "The Silver Fox Inn",
				"relationshipType": "located_in",
				"description": "Viktor was seen at the inn."
			}
		]
	}`

	resp, err := parseEnrichmentResponse(input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.DescriptionUpdates, 1)
	assert.Equal(t, "A mysterious stranger with a scarred left hand.",
		resp.DescriptionUpdates[0].SuggestedDescription)
	assert.Len(t, resp.LogEntries, 1)
	assert.Equal(t, "Arrived at the tavern and spoke with the party.",
		resp.LogEntries[0].Content)
	assert.Len(t, resp.Relationships, 1)
	assert.Equal(t, "located_in", resp.Relationships[0].RelationshipType)
}

func TestParseEnrichmentResponse_WithCodeFences(t *testing.T) {
	input := "```json\n" + `{
		"descriptionUpdates": [],
		"logEntries": [
			{
				"content": "Met the informant at the docks."
			}
		],
		"relationships": []
	}` + "\n```"

	resp, err := parseEnrichmentResponse(input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.DescriptionUpdates, 0)
	assert.Len(t, resp.LogEntries, 1)
	assert.Equal(t, "Met the informant at the docks.",
		resp.LogEntries[0].Content)
	assert.Len(t, resp.Relationships, 0)
}

func TestParseEnrichmentResponse_Malformed(t *testing.T) {
	input := "This is not JSON at all, just some text the LLM hallucinated."

	resp, err := parseEnrichmentResponse(input)

	// Should NOT return an error; graceful degradation.
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.DescriptionUpdates, 0)
	assert.Len(t, resp.LogEntries, 0)
	assert.Len(t, resp.Relationships, 0)
}

func TestParseEnrichmentResponse_PartialJSON(t *testing.T) {
	// Only logEntries provided; other arrays should default to empty.
	input := `{
		"logEntries": [
			{
				"content": "Discovered a hidden passage."
			}
		]
	}`

	resp, err := parseEnrichmentResponse(input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.DescriptionUpdates, 0)
	assert.Len(t, resp.LogEntries, 1)
	assert.Equal(t, "Discovered a hidden passage.",
		resp.LogEntries[0].Content)
	assert.Len(t, resp.Relationships, 0)
}

func TestParseEnrichmentResponse_Empty(t *testing.T) {
	resp, err := parseEnrichmentResponse("")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.DescriptionUpdates, 0)
	assert.Len(t, resp.LogEntries, 0)
	assert.Len(t, resp.Relationships, 0)
}

func TestParseEnrichmentResponse_CodeFencesNoLanguage(t *testing.T) {
	input := "```\n" + `{
		"descriptionUpdates": [],
		"logEntries": [],
		"relationships": [
			{
				"sourceEntityId": 10,
				"sourceEntityName": "Alistair",
				"targetEntityId": 20,
				"targetEntityName": "The Crown",
				"relationshipType": "member_of",
				"description": "Alistair is a knight of The Crown."
			}
		]
	}` + "\n```"

	resp, err := parseEnrichmentResponse(input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Relationships, 1)
	assert.Equal(t, "member_of", resp.Relationships[0].RelationshipType)
}

// ---------------------------------------------------------------------------
// Prompt tests
// ---------------------------------------------------------------------------

func TestBuildUserPrompt(t *testing.T) {
	desc := "A grizzled veteran."
	relDesc := "Viktor knows Elara"
	input := EnrichmentInput{
		CampaignID:  1,
		JobID:       10,
		SourceTable: "chapters",
		SourceID:    5,
		Content:     "Viktor entered the room and greeted Elara warmly.",
		Entity: models.Entity{
			ID:          100,
			CampaignID:  1,
			EntityType:  models.EntityTypeNPC,
			Name:        "Viktor",
			Description: &desc,
		},
		OtherEntities: []models.Entity{
			{
				ID:         200,
				CampaignID: 1,
				EntityType: models.EntityTypeNPC,
				Name:       "Elara",
			},
		},
		Relationships: []models.Relationship{
			{
				ID:               1,
				CampaignID:       1,
				SourceEntityID:   100,
				TargetEntityID:   200,
				RelationshipType: "knows",
				Description:      &relDesc,
				SourceEntity: &models.Entity{
					ID:   100,
					Name: "Viktor",
				},
				TargetEntity: &models.Entity{
					ID:   200,
					Name: "Elara",
				},
			},
		},
	}

	prompt := buildUserPrompt(input)

	assert.Contains(t, prompt, "Viktor entered the room")
	assert.Contains(t, prompt, "**Name**: Viktor")
	assert.Contains(t, prompt, "**Type**: npc")
	assert.Contains(t, prompt, "A grizzled veteran.")
	assert.Contains(t, prompt, "Viktor -[knows]-> Elara")
	assert.Contains(t, prompt, "Viktor knows Elara")
	assert.Contains(t, prompt, "**Elara** (ID: 200, Type: npc)")
	assert.Contains(t, prompt, "Respond with JSON only.")
}

func TestBuildUserPrompt_NoDescription(t *testing.T) {
	input := EnrichmentInput{
		Content: "Some content about the market.",
		Entity: models.Entity{
			ID:         1,
			EntityType: models.EntityTypeLocation,
			Name:       "The Market",
		},
	}

	prompt := buildUserPrompt(input)

	assert.Contains(t, prompt, "**Current Description**: (none)")
	// Should not include relationship or other entities sections.
	assert.NotContains(t, prompt, "## Existing Relationships")
	assert.NotContains(t, prompt, "## Other Entities")
}

// ---------------------------------------------------------------------------
// Truncation tests
// ---------------------------------------------------------------------------

func TestTruncateContent(t *testing.T) {
	t.Run("short content unchanged", func(t *testing.T) {
		content := "Viktor walked into the inn."
		result := truncateContent(content, "Viktor", 100)
		assert.Equal(t, content, result)
	})

	t.Run("long content truncated around mention", func(t *testing.T) {
		// Build a long string with entity name in the middle.
		before := strings.Repeat("a", 3000)
		after := strings.Repeat("b", 3000)
		content := before + "Viktor" + after

		result := truncateContent(content, "Viktor", 200)

		assert.Contains(t, result, "Viktor")
		assert.Contains(t, result, "[...]")
		// Should be approximately 200 chars of actual content plus markers.
		contentOnly := strings.ReplaceAll(result, "[...]", "")
		contentOnly = strings.TrimSpace(contentOnly)
		assert.LessOrEqual(t, len(contentOnly), 210)
	})

	t.Run("entity not found truncates from start", func(t *testing.T) {
		content := strings.Repeat("x", 5000)
		result := truncateContent(content, "NotPresent", 200)

		assert.True(t, strings.HasSuffix(result, "[...]"))
		assert.Len(t, strings.ReplaceAll(result, "\n\n[...]", ""), 200)
	})

	t.Run("entity near start", func(t *testing.T) {
		content := "Viktor" + strings.Repeat("z", 5000)
		result := truncateContent(content, "Viktor", 200)

		assert.Contains(t, result, "Viktor")
		// Should not have leading [...] since entity is at the start.
		assert.False(t, strings.HasPrefix(result, "[...]"))
	})

	t.Run("entity near end", func(t *testing.T) {
		content := strings.Repeat("w", 5000) + "Viktor"
		result := truncateContent(content, "Viktor", 200)

		assert.Contains(t, result, "Viktor")
		// Should not have trailing [...] since entity is at the end.
		assert.False(t, strings.HasSuffix(result, "[...]"))
	})

	t.Run("case insensitive search", func(t *testing.T) {
		content := strings.Repeat("q", 3000) + "VIKTOR" + strings.Repeat("r", 3000)
		result := truncateContent(content, "Viktor", 200)

		assert.Contains(t, result, "VIKTOR")
	})

	t.Run("unicode content not corrupted", func(t *testing.T) {
		// Build content with multi-byte characters (Japanese + accented).
		// Each of these characters is 3 bytes in UTF-8, so byte-based
		// slicing would split them and produce invalid UTF-8.
		prefix := strings.Repeat("\u6771", 100) // 100 x "east" (CJK)
		suffix := strings.Repeat("\u00e9", 100) // 100 x "e-acute"
		entityName := "Viktor"
		content := prefix + entityName + suffix

		result := truncateContent(content, entityName, 50)

		// The entity name must survive truncation intact.
		assert.Contains(t, result, entityName)

		// The result must be valid UTF-8 with no replacement characters.
		assert.NotContains(t, result, "\uFFFD",
			"result contains Unicode replacement character, indicating corruption")

		// Every rune must be decodable (no partial multi-byte sequences).
		for i, r := range result {
			assert.NotEqual(t, rune(0xFFFD), r,
				"corrupt rune at index %d", i)
		}
	})
}

// ---------------------------------------------------------------------------
// Engine integration tests (with mock provider)
// ---------------------------------------------------------------------------

func TestEnrichEntity_Success(t *testing.T) {
	llmResponse := `{
		"descriptionUpdates": [
			{
				"currentDescription": "A mysterious figure.",
				"suggestedDescription": "A mysterious figure who frequents the docks at night.",
				"rationale": "Session notes describe nightly visits to the harbour."
			}
		],
		"logEntries": [
			{
				"content": "Was seen arguing with the harbourmaster.",
				"occurredAt": "Session 3"
			}
		],
		"relationships": [
			{
				"sourceEntityId": 1,
				"sourceEntityName": "Kael",
				"targetEntityId": 2,
				"targetEntityName": "Harbourmaster Grint",
				"relationshipType": "enemy_of",
				"description": "They were seen arguing at the docks."
			}
		]
	}`

	provider := &mockProvider{response: llmResponse}
	engine := NewEngine(nil)

	desc := "A mysterious figure."
	input := EnrichmentInput{
		CampaignID:  1,
		JobID:       42,
		SourceTable: "chapters",
		SourceID:    7,
		Content:     "Kael was seen at the docks arguing with Harbourmaster Grint.",
		Entity: models.Entity{
			ID:          1,
			CampaignID:  1,
			EntityType:  models.EntityTypeNPC,
			Name:        "Kael",
			Description: &desc,
		},
		OtherEntities: []models.Entity{
			{ID: 2, Name: "Harbourmaster Grint", EntityType: models.EntityTypeNPC},
		},
	}

	items, err := engine.EnrichEntity(context.Background(), provider, input)

	require.NoError(t, err)
	require.Len(t, items, 3)

	// Verify description update item.
	assert.Equal(t, int64(42), items[0].JobID)
	assert.Equal(t, "description_update", items[0].DetectionType)
	assert.Equal(t, "Kael", items[0].MatchedText)
	assert.Equal(t, "pending", items[0].Resolution)
	assert.Equal(t, "enrichment", items[0].Phase)
	assert.NotNil(t, items[0].EntityID)
	assert.Equal(t, int64(1), *items[0].EntityID)

	var du models.DescriptionUpdateSuggestion
	require.NoError(t, json.Unmarshal(items[0].SuggestedContent, &du))
	assert.Equal(t, "A mysterious figure who frequents the docks at night.",
		du.SuggestedDescription)

	// Verify log entry item.
	assert.Equal(t, "log_entry", items[1].DetectionType)

	var le models.LogEntrySuggestion
	require.NoError(t, json.Unmarshal(items[1].SuggestedContent, &le))
	assert.Equal(t, "Was seen arguing with the harbourmaster.", le.Content)

	// Verify relationship suggestion item.
	assert.Equal(t, "relationship_suggestion", items[2].DetectionType)

	var rs models.RelationshipSuggestion
	require.NoError(t, json.Unmarshal(items[2].SuggestedContent, &rs))
	assert.Equal(t, "enemy_of", rs.RelationshipType)
	assert.Equal(t, int64(2), rs.TargetEntityID)
}

func TestEnrichEntity_LLMError(t *testing.T) {
	provider := &mockProvider{
		err: errors.New("API rate limit exceeded"),
	}
	engine := NewEngine(nil)

	input := EnrichmentInput{
		CampaignID: 1,
		JobID:      42,
		Content:    "Some content.",
		Entity: models.Entity{
			ID:         1,
			EntityType: models.EntityTypeNPC,
			Name:       "Test",
		},
	}

	items, err := engine.EnrichEntity(context.Background(), provider, input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LLM completion failed")
	assert.Nil(t, items)
}

func TestEnrichEntity_EmptyResponse(t *testing.T) {
	// LLM returns valid JSON with empty arrays.
	llmResponse := `{
		"descriptionUpdates": [],
		"logEntries": [],
		"relationships": []
	}`

	provider := &mockProvider{response: llmResponse}
	engine := NewEngine(nil)

	input := EnrichmentInput{
		CampaignID: 1,
		JobID:      42,
		Content:    "Nothing interesting happened.",
		Entity: models.Entity{
			ID:         1,
			EntityType: models.EntityTypeNPC,
			Name:       "Boring NPC",
		},
	}

	items, err := engine.EnrichEntity(context.Background(), provider, input)

	require.NoError(t, err)
	require.NotNil(t, items, "should return non-nil empty slice")
	assert.Empty(t, items)
}

func TestEnrichEntity_MalformedLLMResponse(t *testing.T) {
	// LLM returns garbage; engine should degrade gracefully.
	provider := &mockProvider{response: "I cannot comply with your request."}
	engine := NewEngine(nil)

	input := EnrichmentInput{
		CampaignID: 1,
		JobID:      42,
		Content:    "Some session notes.",
		Entity: models.Entity{
			ID:         1,
			EntityType: models.EntityTypeNPC,
			Name:       "Test",
		},
	}

	items, err := engine.EnrichEntity(context.Background(), provider, input)

	// Should not error; returns non-nil empty slice.
	require.NoError(t, err)
	require.NotNil(t, items, "should return non-nil empty slice on parse failure")
	assert.Empty(t, items)
}

func TestEnrichEntity_EmptyContent(t *testing.T) {
	provider := &mockProvider{
		response: `{"descriptionUpdates":[],"logEntries":[],"relationships":[]}`,
	}
	engine := NewEngine(nil)

	input := EnrichmentInput{
		CampaignID: 1,
		JobID:      42,
		Content:    "",
		Entity: models.Entity{
			ID:         1,
			EntityType: models.EntityTypeNPC,
			Name:       "Test",
		},
	}

	items, err := engine.EnrichEntity(context.Background(), provider, input)

	require.NoError(t, err)
	require.NotNil(t, items)
	assert.Empty(t, items)
}

func TestEnrichEntity_ZeroEntityID(t *testing.T) {
	provider := &mockProvider{
		response: `{"descriptionUpdates":[],"logEntries":[],"relationships":[]}`,
	}
	engine := NewEngine(nil)

	input := EnrichmentInput{
		CampaignID: 1,
		JobID:      42,
		Content:    "Some content.",
		Entity: models.Entity{
			ID:         0,
			EntityType: models.EntityTypeNPC,
			Name:       "Test",
		},
	}

	items, err := engine.EnrichEntity(context.Background(), provider, input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "entity ID is required")
	assert.Nil(t, items)
}

func TestEnrichEntity_EmptyEntityName(t *testing.T) {
	provider := &mockProvider{
		response: `{"descriptionUpdates":[],"logEntries":[],"relationships":[]}`,
	}
	engine := NewEngine(nil)

	input := EnrichmentInput{
		CampaignID: 1,
		JobID:      42,
		Content:    "Some content.",
		Entity: models.Entity{
			ID:         1,
			EntityType: models.EntityTypeNPC,
			Name:       "",
		},
	}

	items, err := engine.EnrichEntity(context.Background(), provider, input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "entity name is required")
	assert.Nil(t, items)
}

func TestEnrichEntity_VerifiesProviderRequest(t *testing.T) {
	// Capture the request sent to the provider.
	var captured llm.CompletionRequest
	provider := &mockProvider{
		response: `{"descriptionUpdates":[],"logEntries":[],"relationships":[]}`,
	}

	// Wrap mock to capture the request.
	captureProvider := &capturingProvider{
		inner:    provider,
		captured: &captured,
	}
	engine := NewEngine(nil)

	desc := "A known thief."
	input := EnrichmentInput{
		CampaignID: 1,
		JobID:      42,
		Content:    "Content about Raven.",
		Entity: models.Entity{
			ID:          5,
			EntityType:  models.EntityTypeNPC,
			Name:        "Raven",
			Description: &desc,
		},
	}

	_, err := engine.EnrichEntity(
		context.Background(), captureProvider, input,
	)
	require.NoError(t, err)

	assert.Equal(t, 2048, captured.MaxTokens)
	assert.InDelta(t, 0.3, captured.Temperature, 0.001)
	assert.Contains(t, captured.SystemPrompt, "TTRPG campaign analyst")
	assert.Contains(t, captured.UserPrompt, "Raven")
}

// capturingProvider wraps a Provider to capture the CompletionRequest.
type capturingProvider struct {
	inner    llm.Provider
	captured *llm.CompletionRequest
}

func (c *capturingProvider) Complete(
	ctx context.Context,
	req llm.CompletionRequest,
) (llm.CompletionResponse, error) {
	*c.captured = req
	return c.inner.Complete(ctx, req)
}
