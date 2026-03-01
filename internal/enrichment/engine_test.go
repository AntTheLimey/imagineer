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
	"github.com/antonypegg/imagineer/internal/ontology"
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
				ID:                   1,
				CampaignID:           1,
				SourceEntityID:       100,
				TargetEntityID:       200,
				RelationshipTypeName: "knows",
				Description:          &relDesc,
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

func TestEnrichEntity_SkipsDuplicateRelationship(t *testing.T) {
	// The LLM suggests a relationship between entities 1 and 2, but one
	// already exists. The suggestion should be silently skipped.
	llmResponse := `{
		"descriptionUpdates": [],
		"logEntries": [],
		"relationships": [
			{
				"sourceEntityId": 1,
				"sourceEntityName": "Kael",
				"targetEntityId": 2,
				"targetEntityName": "Harbourmaster Grint",
				"relationshipType": "enemy_of",
				"description": "They were seen arguing."
			}
		]
	}`

	provider := &mockProvider{response: llmResponse}
	engine := NewEngine(nil)

	input := EnrichmentInput{
		CampaignID:  1,
		JobID:       42,
		SourceTable: "chapters",
		SourceID:    7,
		Content:     "Kael argued with Harbourmaster Grint.",
		Entity: models.Entity{
			ID:         1,
			EntityType: models.EntityTypeNPC,
			Name:       "Kael",
		},
		Relationships: []models.Relationship{
			{
				ID:                   10,
				CampaignID:           1,
				SourceEntityID:       1,
				TargetEntityID:       2,
				RelationshipTypeName: "knows",
			},
		},
	}

	items, err := engine.EnrichEntity(context.Background(), provider, input)

	require.NoError(t, err)
	// Only description updates and log entries should remain; the
	// relationship suggestion should be filtered out.
	for _, item := range items {
		assert.NotEqual(t, "relationship_suggestion", item.DetectionType,
			"duplicate relationship should have been skipped")
	}
}

func TestEnrichEntity_SkipsDuplicateRelationshipInverse(t *testing.T) {
	// The existing relationship is stored as (2 -> 1) but the LLM
	// suggests (1 -> 2). Because the model uses single-edge storage,
	// these represent the same pair and the suggestion should be skipped.
	llmResponse := `{
		"descriptionUpdates": [],
		"logEntries": [],
		"relationships": [
			{
				"sourceEntityId": 1,
				"sourceEntityName": "Kael",
				"targetEntityId": 2,
				"targetEntityName": "Harbourmaster Grint",
				"relationshipType": "enemy_of",
				"description": "They were seen arguing."
			}
		]
	}`

	provider := &mockProvider{response: llmResponse}
	engine := NewEngine(nil)

	input := EnrichmentInput{
		CampaignID:  1,
		JobID:       42,
		SourceTable: "chapters",
		SourceID:    7,
		Content:     "Kael argued with Harbourmaster Grint.",
		Entity: models.Entity{
			ID:         1,
			EntityType: models.EntityTypeNPC,
			Name:       "Kael",
		},
		Relationships: []models.Relationship{
			{
				ID:                   10,
				CampaignID:           1,
				SourceEntityID:       2, // inverse direction
				TargetEntityID:       1,
				RelationshipTypeName: "knows",
			},
		},
	}

	items, err := engine.EnrichEntity(context.Background(), provider, input)

	require.NoError(t, err)
	for _, item := range items {
		assert.NotEqual(t, "relationship_suggestion", item.DetectionType,
			"inverse duplicate relationship should have been skipped")
	}
}

func TestEnrichEntity_AllowsNewRelationshipPair(t *testing.T) {
	// The LLM suggests a relationship between entities 1 and 3. An
	// existing relationship exists between 1 and 2 only. The suggestion
	// for the new pair (1, 3) should be kept.
	llmResponse := `{
		"descriptionUpdates": [],
		"logEntries": [],
		"relationships": [
			{
				"sourceEntityId": 1,
				"sourceEntityName": "Kael",
				"targetEntityId": 3,
				"targetEntityName": "The Docks",
				"relationshipType": "located_in",
				"description": "Kael frequents the docks."
			}
		]
	}`

	provider := &mockProvider{response: llmResponse}
	engine := NewEngine(nil)

	input := EnrichmentInput{
		CampaignID:  1,
		JobID:       42,
		SourceTable: "chapters",
		SourceID:    7,
		Content:     "Kael went to the docks.",
		Entity: models.Entity{
			ID:         1,
			EntityType: models.EntityTypeNPC,
			Name:       "Kael",
		},
		Relationships: []models.Relationship{
			{
				ID:                   10,
				CampaignID:           1,
				SourceEntityID:       1,
				TargetEntityID:       2,
				RelationshipTypeName: "knows",
			},
		},
	}

	items, err := engine.EnrichEntity(context.Background(), provider, input)

	require.NoError(t, err)
	// The relationship suggestion for the new pair should be present.
	var found bool
	for _, item := range items {
		if item.DetectionType == "relationship_suggestion" {
			found = true
			var rs models.RelationshipSuggestion
			require.NoError(t, json.Unmarshal(item.SuggestedContent, &rs))
			assert.Equal(t, int64(3), rs.TargetEntityID)
		}
	}
	assert.True(t, found,
		"relationship suggestion for new pair should be present")
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

// ---------------------------------------------------------------------------
// New-entity detection parser tests
// ---------------------------------------------------------------------------

func TestParseNewEntityResponse_Valid(t *testing.T) {
	input := `{
		"new_entities": [
			{
				"name": "Inspector Barrington",
				"entity_type": "npc",
				"description": "A Scotland Yard detective",
				"reasoning": "Named character not in the known list"
			},
			{
				"name": "The Blackwood Estate",
				"entity_type": "location",
				"description": "A manor house on the outskirts of town",
				"reasoning": "Named location mentioned in paragraph 2"
			}
		]
	}`

	resp, err := parseNewEntityResponse(input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.NewEntities, 2)
	assert.Equal(t, "Inspector Barrington", resp.NewEntities[0].Name)
	assert.Equal(t, "npc", resp.NewEntities[0].EntityType)
	assert.Equal(t, "A Scotland Yard detective",
		resp.NewEntities[0].Description)
	assert.Equal(t, "The Blackwood Estate", resp.NewEntities[1].Name)
	assert.Equal(t, "location", resp.NewEntities[1].EntityType)
}

func TestParseNewEntityResponse_Empty(t *testing.T) {
	resp, err := parseNewEntityResponse("")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.NewEntities, 0)
}

func TestParseNewEntityResponse_NoEntities(t *testing.T) {
	input := `{"new_entities": []}`

	resp, err := parseNewEntityResponse(input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.NewEntities, 0)
}

func TestParseNewEntityResponse_WithCodeFences(t *testing.T) {
	input := "```json\n" + `{
		"new_entities": [
			{
				"name": "Captain Holt",
				"entity_type": "npc",
				"description": "A ship captain",
				"reasoning": "Named in dialogue"
			}
		]
	}` + "\n```"

	resp, err := parseNewEntityResponse(input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.NewEntities, 1)
	assert.Equal(t, "Captain Holt", resp.NewEntities[0].Name)
}

func TestParseNewEntityResponse_Malformed(t *testing.T) {
	input := "Here are the entities I found in the text."

	resp, err := parseNewEntityResponse(input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.NewEntities, 0)
}

func TestParseNewEntityResponse_NullArray(t *testing.T) {
	// JSON with null instead of array.
	input := `{"new_entities": null}`

	resp, err := parseNewEntityResponse(input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotNil(t, resp.NewEntities, "should normalise nil to empty slice")
	assert.Len(t, resp.NewEntities, 0)
}

// ---------------------------------------------------------------------------
// New-entity detection converter tests
// ---------------------------------------------------------------------------

func TestConvertNewEntitiesToItems(t *testing.T) {
	resp := &newEntityResponse{
		NewEntities: []newEntitySuggestion{
			{
				Name:        "Lady Ashworth",
				EntityType:  "npc",
				Description: "A wealthy patron of the arts",
				Reasoning:   "Named in the chapter introduction",
			},
			{
				Name:        "The Iron Key",
				EntityType:  "item",
				Description: "An ornate iron key found in the study",
				Reasoning:   "Described as a named artefact",
			},
		},
	}

	items := convertNewEntitiesToItems(1, 42, resp)

	require.Len(t, items, 2)

	// First item.
	assert.Equal(t, int64(42), items[0].JobID)
	assert.Equal(t, "new_entity_suggestion", items[0].DetectionType)
	assert.Equal(t, "Lady Ashworth", items[0].MatchedText)
	assert.Equal(t, "pending", items[0].Resolution)
	assert.Equal(t, "enrichment", items[0].Phase)
	assert.Nil(t, items[0].EntityID)

	var content0 map[string]string
	require.NoError(t, json.Unmarshal(items[0].SuggestedContent, &content0))
	assert.Equal(t, "npc", content0["entity_type"])
	assert.Equal(t, "A wealthy patron of the arts", content0["description"])
	assert.Equal(t, "Named in the chapter introduction",
		content0["reasoning"])

	// Second item.
	assert.Equal(t, "The Iron Key", items[1].MatchedText)
	var content1 map[string]string
	require.NoError(t, json.Unmarshal(items[1].SuggestedContent, &content1))
	assert.Equal(t, "item", content1["entity_type"])
}

func TestConvertNewEntitiesToItems_NilResponse(t *testing.T) {
	items := convertNewEntitiesToItems(1, 42, nil)

	require.NotNil(t, items)
	assert.Empty(t, items)
}

func TestConvertNewEntitiesToItems_EmptyResponse(t *testing.T) {
	resp := &newEntityResponse{
		NewEntities: []newEntitySuggestion{},
	}

	items := convertNewEntitiesToItems(1, 42, resp)

	require.NotNil(t, items)
	assert.Empty(t, items)
}

// ---------------------------------------------------------------------------
// New-entity detection prompt tests
// ---------------------------------------------------------------------------

func TestBuildNewEntityDetectionSystemPrompt(t *testing.T) {
	prompt := buildNewEntityDetectionSystemPrompt()

	assert.Contains(t, prompt, "TTRPG campaign analyst")
	assert.Contains(t, prompt, "new_entities")
	assert.Contains(t, prompt, "entity_type")
	assert.Contains(t, prompt, "npc")
	assert.Contains(t, prompt, "location")
	assert.Contains(t, prompt, "valid JSON only")
	assert.Contains(t, prompt, "two to three sentences")
}

func TestBuildNewEntityDetectionUserPrompt(t *testing.T) {
	knownEntities := []models.Entity{
		{
			ID:         1,
			Name:       "Viktor",
			EntityType: models.EntityTypeNPC,
		},
		{
			ID:         2,
			Name:       "The Silver Fox Inn",
			EntityType: models.EntityTypeLocation,
		},
	}

	prompt := buildNewEntityDetectionUserPrompt(
		"Viktor and Inspector Barrington met at The Silver Fox Inn.",
		knownEntities,
	)

	assert.Contains(t, prompt, "## Source Content")
	assert.Contains(t, prompt, "Viktor and Inspector Barrington")
	assert.Contains(t, prompt, "## Known Entities")
	assert.Contains(t, prompt, "Viktor (npc)")
	assert.Contains(t, prompt, "The Silver Fox Inn (location)")
	assert.Contains(t, prompt, "NOT in the known entities list")
}

func TestBuildNewEntityDetectionUserPrompt_NoKnownEntities(t *testing.T) {
	prompt := buildNewEntityDetectionUserPrompt(
		"A tale of adventure.",
		nil,
	)

	assert.Contains(t, prompt, "## Source Content")
	assert.NotContains(t, prompt, "## Known Entities")
	assert.Contains(t, prompt, "NOT in the known entities list")
}

func TestBuildNewEntityDetectionUserPrompt_TruncatesLongContent(t *testing.T) {
	longContent := strings.Repeat("x", maxContentChars+500)

	prompt := buildNewEntityDetectionUserPrompt(longContent, nil)

	// The content section should be truncated to maxContentChars runes.
	assert.NotContains(t, prompt, strings.Repeat("x", maxContentChars+1))
}

// ---------------------------------------------------------------------------
// DetectNewEntities engine tests
// ---------------------------------------------------------------------------

func TestDetectNewEntities_Success(t *testing.T) {
	llmResponse := `{
		"new_entities": [
			{
				"name": "Inspector Barrington",
				"entity_type": "npc",
				"description": "A Scotland Yard detective",
				"reasoning": "Named character not in the known list"
			}
		]
	}`

	provider := &mockProvider{response: llmResponse}
	engine := NewEngine(nil)

	knownEntities := []models.Entity{
		{ID: 1, Name: "Viktor", EntityType: models.EntityTypeNPC},
	}

	items, err := engine.DetectNewEntities(
		context.Background(),
		provider,
		1,  // campaignID
		42, // jobID
		"Viktor met Inspector Barrington at the docks.",
		knownEntities,
	)

	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, int64(42), items[0].JobID)
	assert.Equal(t, "new_entity_suggestion", items[0].DetectionType)
	assert.Equal(t, "Inspector Barrington", items[0].MatchedText)
	assert.Equal(t, "pending", items[0].Resolution)
	assert.Equal(t, "enrichment", items[0].Phase)
	assert.Nil(t, items[0].EntityID)

	var content map[string]string
	require.NoError(t, json.Unmarshal(items[0].SuggestedContent, &content))
	assert.Equal(t, "npc", content["entity_type"])
	assert.Equal(t, "A Scotland Yard detective", content["description"])
}

func TestDetectNewEntities_EmptyContent(t *testing.T) {
	provider := &mockProvider{response: `{"new_entities": []}`}
	engine := NewEngine(nil)

	items, err := engine.DetectNewEntities(
		context.Background(), provider, 1, 42, "", nil,
	)

	require.NoError(t, err)
	require.NotNil(t, items)
	assert.Empty(t, items)
}

func TestDetectNewEntities_LLMError(t *testing.T) {
	provider := &mockProvider{
		err: errors.New("service unavailable"),
	}
	engine := NewEngine(nil)

	items, err := engine.DetectNewEntities(
		context.Background(), provider, 1, 42, "Some content.", nil,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LLM completion failed")
	assert.Nil(t, items)
}

func TestDetectNewEntities_MalformedResponse(t *testing.T) {
	provider := &mockProvider{
		response: "I found some entities for you!",
	}
	engine := NewEngine(nil)

	items, err := engine.DetectNewEntities(
		context.Background(), provider, 1, 42, "Some content.", nil,
	)

	require.NoError(t, err)
	require.NotNil(t, items, "should return non-nil empty slice")
	assert.Empty(t, items)
}

func TestDetectNewEntities_VerifiesProviderRequest(t *testing.T) {
	var captured llm.CompletionRequest
	provider := &mockProvider{
		response: `{"new_entities": []}`,
	}
	captureProvider := &capturingProvider{
		inner:    provider,
		captured: &captured,
	}
	engine := NewEngine(nil)

	knownEntities := []models.Entity{
		{ID: 1, Name: "Raven", EntityType: models.EntityTypeNPC},
	}

	_, err := engine.DetectNewEntities(
		context.Background(),
		captureProvider,
		1,
		42,
		"Content about Raven and the mysterious Dr. Crane.",
		knownEntities,
	)
	require.NoError(t, err)

	assert.Equal(t, 2048, captured.MaxTokens)
	assert.InDelta(t, 0.3, captured.Temperature, 0.001)
	assert.Contains(t, captured.SystemPrompt, "TTRPG campaign analyst")
	assert.Contains(t, captured.UserPrompt, "Raven (npc)")
	assert.Contains(t, captured.UserPrompt, "Dr. Crane")
}

func TestDetectNewEntities_NoNewEntitiesFound(t *testing.T) {
	llmResponse := `{"new_entities": []}`

	provider := &mockProvider{response: llmResponse}
	engine := NewEngine(nil)

	items, err := engine.DetectNewEntities(
		context.Background(), provider, 1, 42,
		"Nothing new here, just Viktor again.",
		[]models.Entity{
			{ID: 1, Name: "Viktor", EntityType: models.EntityTypeNPC},
		},
	)

	require.NoError(t, err)
	require.NotNil(t, items)
	assert.Empty(t, items)
}

func TestBuildUserPrompt_WithRAGContext(t *testing.T) {
	desc := "A grizzled veteran."
	input := EnrichmentInput{
		CampaignID:  1,
		JobID:       10,
		SourceTable: "chapters",
		SourceID:    5,
		Content:     "Viktor entered the tavern.",
		Entity: models.Entity{
			ID:          100,
			CampaignID:  1,
			EntityType:  models.EntityTypeNPC,
			Name:        "Viktor",
			Description: &desc,
		},
		CampaignResults: []models.SearchResult{
			{
				SourceTable:  "chapters",
				SourceID:     2,
				SourceName:   "Chapter 2",
				ChunkContent: "Viktor was first seen arriving at the docks.",
			},
		},
		GameSystemYAML: "name: Call of Cthulhu 7e\nskills:\n  - Spot Hidden",
	}

	prompt := buildUserPrompt(input)

	assert.Contains(t, prompt, "## Campaign Context")
	assert.Contains(t, prompt, "Chapter 2")
	assert.Contains(t, prompt, "arriving at the docks")
	assert.Contains(t, prompt, "## Game System Schema")
	assert.Contains(t, prompt, "Call of Cthulhu 7e")
}

func TestBuildUserPrompt_NoRAGContext(t *testing.T) {
	input := EnrichmentInput{
		Content: "Some content.",
		Entity: models.Entity{
			ID:         1,
			EntityType: models.EntityTypeNPC,
			Name:       "Test",
		},
	}

	prompt := buildUserPrompt(input)

	assert.NotContains(t, prompt, "## Campaign Context")
	assert.NotContains(t, prompt, "## Game System Schema")
}

// ---------------------------------------------------------------------------
// Ontology system prompt tests
// ---------------------------------------------------------------------------

func TestBuildSystemPrompt_NilOntology(t *testing.T) {
	prompt := buildSystemPrompt(nil)

	assert.Contains(t, prompt, "TTRPG campaign analyst")
	assert.NotContains(t, prompt, "## Valid Entity Types")
	assert.NotContains(t, prompt, "## Valid Relationship Types")
}

func TestBuildSystemPrompt_EmptyOntology(t *testing.T) {
	ont := &ontology.Ontology{}

	prompt := buildSystemPrompt(ont)

	assert.Contains(t, prompt, "TTRPG campaign analyst")
	assert.NotContains(t, prompt, "## Valid Entity Types")
	assert.NotContains(t, prompt, "## Valid Relationship Types")
}

func TestBuildSystemPrompt_WithEntityTypes(t *testing.T) {
	ont := &ontology.Ontology{
		EntityTypes: &ontology.EntityTypeFile{
			Types: map[string]ontology.EntityTypeDef{
				"entity": {Abstract: true, Description: "Root type"},
				"agent":  {Abstract: true, Parent: "entity"},
				"npc":    {Abstract: false, Parent: "agent"},
				"pc":     {Abstract: false, Parent: "agent"},
				"place":  {Abstract: true, Parent: "entity"},
				"location": {
					Abstract: false,
					Parent:   "place",
				},
				"item": {Abstract: false, Parent: "entity"},
			},
		},
	}

	prompt := buildSystemPrompt(ont)

	assert.Contains(t, prompt, "## Valid Entity Types")
	assert.Contains(t, prompt, "item")
	assert.Contains(t, prompt, "location")
	assert.Contains(t, prompt, "npc")
	assert.Contains(t, prompt, "pc")
	// Abstract types should not appear in the list.
	// Check that they don't appear after "Only suggest entities with these types:"
	idx := strings.Index(prompt, "Only suggest entities with these types:")
	require.Greater(t, idx, 0)
	entityTypeLine := prompt[idx:]
	assert.NotContains(t, entityTypeLine, "entity,")
	assert.NotContains(t, entityTypeLine, "agent,")
	assert.NotContains(t, entityTypeLine, "place,")
}

func TestBuildSystemPrompt_WithRelationshipTypes(t *testing.T) {
	ont := &ontology.Ontology{
		RelationshipTypes: &ontology.RelationshipTypeFile{
			Types: map[string]ontology.RelationshipTypeDef{
				"located_at": {
					Domain: []string{"agent", "artifact"},
					Range:  []string{"place"},
				},
				"member_of": {
					Domain: []string{"npc"},
					Range:  []string{"faction", "organization"},
				},
				"knows": {
					Domain: []string{},
					Range:  []string{},
				},
			},
		},
	}

	prompt := buildSystemPrompt(ont)

	assert.Contains(t, prompt, "## Valid Relationship Types")
	assert.Contains(t, prompt, "located_at: agent, artifact -> place")
	assert.Contains(t, prompt, "member_of: npc -> faction, organization")
	assert.Contains(t, prompt, "knows")
	assert.Contains(t, prompt, "Do not invent new relationship types")
}

func TestBuildSystemPrompt_WithFullOntology(t *testing.T) {
	ont := &ontology.Ontology{
		EntityTypes: &ontology.EntityTypeFile{
			Types: map[string]ontology.EntityTypeDef{
				"npc":      {Abstract: false},
				"location": {Abstract: false},
				"item":     {Abstract: false},
			},
		},
		RelationshipTypes: &ontology.RelationshipTypeFile{
			Types: map[string]ontology.RelationshipTypeDef{
				"owns": {
					Domain: []string{"npc"},
					Range:  []string{"item"},
				},
			},
		},
	}

	prompt := buildSystemPrompt(ont)

	// Both sections should be present.
	assert.Contains(t, prompt, "## Valid Entity Types")
	assert.Contains(t, prompt, "## Valid Relationship Types")
	// Base prompt should still be present.
	assert.Contains(t, prompt, "TTRPG campaign analyst")
	assert.Contains(t, prompt, "descriptionUpdates")
}

func TestBuildSystemPrompt_RelTypesNoDomainRange(t *testing.T) {
	ont := &ontology.Ontology{
		RelationshipTypes: &ontology.RelationshipTypeFile{
			Types: map[string]ontology.RelationshipTypeDef{
				"ally_of":  {},
				"enemy_of": {},
			},
		},
	}

	prompt := buildSystemPrompt(ont)

	assert.Contains(t, prompt, "## Valid Relationship Types")
	assert.Contains(t, prompt, "- ally_of")
	assert.Contains(t, prompt, "- enemy_of")
	// Without domain/range, lines should not have ": ... -> ..."
	assert.NotContains(t, prompt, "ally_of:")
	assert.NotContains(t, prompt, "enemy_of:")
}

func TestBuildSystemPrompt_OnlyAbstractEntityTypes(t *testing.T) {
	ont := &ontology.Ontology{
		EntityTypes: &ontology.EntityTypeFile{
			Types: map[string]ontology.EntityTypeDef{
				"entity": {Abstract: true},
				"agent":  {Abstract: true},
			},
		},
	}

	prompt := buildSystemPrompt(ont)

	// No concrete types means no entity types section.
	assert.NotContains(t, prompt, "## Valid Entity Types")
}

func TestEnrichEntity_OntologyThreadedToPrompt(t *testing.T) {
	// Verify that when ontology is set on EnrichmentInput, the
	// system prompt includes ontology information.
	var captured llm.CompletionRequest
	provider := &mockProvider{
		response: `{"descriptionUpdates":[],"logEntries":[],"relationships":[]}`,
	}
	captureProvider := &capturingProvider{
		inner:    provider,
		captured: &captured,
	}
	engine := NewEngine(nil)

	ont := &ontology.Ontology{
		EntityTypes: &ontology.EntityTypeFile{
			Types: map[string]ontology.EntityTypeDef{
				"npc":      {Abstract: false},
				"location": {Abstract: false},
			},
		},
		RelationshipTypes: &ontology.RelationshipTypeFile{
			Types: map[string]ontology.RelationshipTypeDef{
				"located_at": {
					Domain: []string{"npc"},
					Range:  []string{"location"},
				},
			},
		},
	}

	input := EnrichmentInput{
		CampaignID: 1,
		JobID:      42,
		Content:    "Kael is at the docks.",
		Entity: models.Entity{
			ID:         1,
			EntityType: models.EntityTypeNPC,
			Name:       "Kael",
		},
		Ontology: ont,
	}

	_, err := engine.EnrichEntity(
		context.Background(), captureProvider, input,
	)
	require.NoError(t, err)

	assert.Contains(t, captured.SystemPrompt, "## Valid Entity Types")
	assert.Contains(t, captured.SystemPrompt, "npc")
	assert.Contains(t, captured.SystemPrompt, "location")
	assert.Contains(t, captured.SystemPrompt, "## Valid Relationship Types")
	assert.Contains(t, captured.SystemPrompt, "located_at: npc -> location")
}

func TestEnrichEntity_NilOntologyBackwardCompatible(t *testing.T) {
	// Verify that with nil ontology, the prompt is unchanged from
	// the base prompt (backward compatibility).
	var captured llm.CompletionRequest
	provider := &mockProvider{
		response: `{"descriptionUpdates":[],"logEntries":[],"relationships":[]}`,
	}
	captureProvider := &capturingProvider{
		inner:    provider,
		captured: &captured,
	}
	engine := NewEngine(nil)

	input := EnrichmentInput{
		CampaignID: 1,
		JobID:      42,
		Content:    "Kael is at the docks.",
		Entity: models.Entity{
			ID:         1,
			EntityType: models.EntityTypeNPC,
			Name:       "Kael",
		},
		Ontology: nil,
	}

	_, err := engine.EnrichEntity(
		context.Background(), captureProvider, input,
	)
	require.NoError(t, err)

	assert.Contains(t, captured.SystemPrompt, "TTRPG campaign analyst")
	assert.NotContains(t, captured.SystemPrompt, "## Valid Entity Types")
	assert.NotContains(t, captured.SystemPrompt, "## Valid Relationship Types")
}
