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
	"context"
	"errors"
	"testing"

	"github.com/antonypegg/imagineer/internal/agents"
	"github.com/antonypegg/imagineer/internal/enrichment"
	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/antonypegg/imagineer/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProvider implements llm.Provider for testing.
type mockProvider struct {
	response string
	err      error
	called   bool
}

func (m *mockProvider) Complete(
	ctx context.Context,
	req llm.CompletionRequest,
) (llm.CompletionResponse, error) {
	m.called = true
	if m.err != nil {
		return llm.CompletionResponse{}, m.err
	}
	return llm.CompletionResponse{Content: m.response}, nil
}

// ---------------------------------------------------------------------------
// Constructor and metadata tests
// ---------------------------------------------------------------------------

func TestExpert_Name(t *testing.T) {
	expert := NewExpert()
	assert.Equal(t, "canon-expert", expert.Name())
}

func TestExpert_DependsOn(t *testing.T) {
	expert := NewExpert()
	assert.Nil(t, expert.DependsOn())
}

// ---------------------------------------------------------------------------
// Run tests
// ---------------------------------------------------------------------------

func TestExpert_Run_EmptyContent(t *testing.T) {
	provider := &mockProvider{
		response: `{"contradictions":[]}`,
	}
	expert := NewExpert()

	input := enrichment.PipelineInput{
		CampaignID:  1,
		JobID:       10,
		SourceTable: "chapters",
		SourceID:    5,
		Content:     "",
	}

	items, err := expert.Run(context.Background(), provider, input)

	require.NoError(t, err)
	assert.Empty(t, items)
	assert.False(t, provider.called,
		"LLM should not be called when content is empty")
}

func TestExpert_Run_NoRAGContext(t *testing.T) {
	provider := &mockProvider{
		response: `{"contradictions":[]}`,
	}
	expert := NewExpert()

	input := enrichment.PipelineInput{
		CampaignID:  1,
		JobID:       10,
		SourceTable: "chapters",
		SourceID:    5,
		Content:     "Viktor meets Irena at the Silver Fox Inn.",
		Context:     nil,
	}

	items, err := expert.Run(context.Background(), provider, input)

	require.NoError(t, err)
	assert.Empty(t, items)
	assert.False(t, provider.called,
		"LLM should not be called when RAG context is nil")
}

func TestExpert_Run_EmptyRAGResults(t *testing.T) {
	provider := &mockProvider{
		response: `{"contradictions":[]}`,
	}
	expert := NewExpert()

	input := enrichment.PipelineInput{
		CampaignID:  1,
		JobID:       10,
		SourceTable: "chapters",
		SourceID:    5,
		Content:     "Viktor meets Irena at the Silver Fox Inn.",
		Context: &enrichment.RAGContext{
			CampaignResults: []models.SearchResult{},
		},
	}

	items, err := expert.Run(context.Background(), provider, input)

	require.NoError(t, err)
	assert.Empty(t, items)
	assert.False(t, provider.called,
		"LLM should not be called when campaign results are empty")
}

func TestExpert_Run_ValidResponse(t *testing.T) {
	llmResponse := `{
		"contradictions": [
			{
				"contradictionType": "factual",
				"severity": "warning",
				"conflictingText": "Viktor arrived in London in 1923",
				"establishedFact": "Viktor arrived in London in 1921",
				"source": "Chapter 1 overview",
				"description": "The arrival year for Viktor contradicts established canon.",
				"suggestion": "Change the year to 1921 to match Chapter 1."
			},
			{
				"contradictionType": "character",
				"severity": "error",
				"conflictingText": "Irena is described as a bartender",
				"establishedFact": "Irena is a librarian at the Bodleian",
				"source": "Entity: Irena",
				"description": "Irena's occupation contradicts her entity description.",
				"suggestion": "Update to reflect her role as a librarian."
			}
		]
	}`

	provider := &mockProvider{response: llmResponse}
	expert := NewExpert()

	input := enrichment.PipelineInput{
		CampaignID:  1,
		JobID:       42,
		SourceTable: "chapters",
		SourceID:    7,
		Content:     "Viktor arrived in London in 1923. He met Irena, the bartender.",
		Context: &enrichment.RAGContext{
			CampaignResults: []models.SearchResult{
				{
					SourceTable:  "chapters",
					SourceID:     1,
					SourceName:   "Chapter 1",
					ChunkContent: "Viktor arrived in London in 1921.",
				},
			},
		},
	}

	items, err := expert.Run(context.Background(), provider, input)

	require.NoError(t, err)
	require.Len(t, items, 2)

	// Verify the first item (factual -> canon_contradiction).
	assert.Equal(t, int64(42), items[0].JobID)
	assert.Equal(t, "canon_contradiction", items[0].DetectionType)
	assert.Equal(t, "pending", items[0].Resolution)
	assert.Equal(t, "analysis", items[0].Phase)

	// Verify the second item (character -> character_inconsistency).
	assert.Equal(t, int64(42), items[1].JobID)
	assert.Equal(t, "character_inconsistency", items[1].DetectionType)
	assert.Equal(t, "pending", items[1].Resolution)
	assert.Equal(t, "analysis", items[1].Phase)
}

func TestExpert_Run_LLMError(t *testing.T) {
	provider := &mockProvider{
		err: errors.New("API rate limit exceeded"),
	}
	expert := NewExpert()

	input := enrichment.PipelineInput{
		CampaignID:  1,
		JobID:       42,
		SourceTable: "chapters",
		SourceID:    7,
		Content:     "Some session content.",
		Context: &enrichment.RAGContext{
			CampaignResults: []models.SearchResult{
				{
					SourceTable:  "chapters",
					SourceID:     1,
					SourceName:   "Chapter 1",
					ChunkContent: "Established canon content.",
				},
			},
		},
	}

	items, err := expert.Run(context.Background(), provider, input)

	assert.Error(t, err)
	assert.Nil(t, items)
}

func TestExpert_Run_MalformedJSON(t *testing.T) {
	provider := &mockProvider{
		response: "I cannot comply with your request. Here is my analysis in prose form.",
	}
	expert := NewExpert()

	input := enrichment.PipelineInput{
		CampaignID:  1,
		JobID:       42,
		SourceTable: "chapters",
		SourceID:    7,
		Content:     "Some session content.",
		Context: &enrichment.RAGContext{
			CampaignResults: []models.SearchResult{
				{
					SourceTable:  "chapters",
					SourceID:     1,
					SourceName:   "Chapter 1",
					ChunkContent: "Established canon content.",
				},
			},
		},
	}

	items, err := expert.Run(context.Background(), provider, input)

	// Graceful degradation: no error, empty items.
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestExpert_Run_EmptyContradictions(t *testing.T) {
	llmResponse := `{
		"contradictions": []
	}`

	provider := &mockProvider{response: llmResponse}
	expert := NewExpert()

	input := enrichment.PipelineInput{
		CampaignID:  1,
		JobID:       42,
		SourceTable: "chapters",
		SourceID:    7,
		Content:     "A well-written session with no contradictions.",
		Context: &enrichment.RAGContext{
			CampaignResults: []models.SearchResult{
				{
					SourceTable:  "chapters",
					SourceID:     1,
					SourceName:   "Chapter 1",
					ChunkContent: "Previously established facts.",
				},
			},
		},
	}

	items, err := expert.Run(context.Background(), provider, input)

	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestExpert_Run_AllItemsHaveAnalysisPhase(t *testing.T) {
	llmResponse := `{
		"contradictions": [
			{
				"contradictionType": "factual",
				"severity": "warning",
				"conflictingText": "The inn is on Baker Street",
				"establishedFact": "The inn is on Fleet Street",
				"source": "Entity: Silver Fox Inn",
				"description": "Street name contradicts entity record.",
				"suggestion": "Change to Fleet Street."
			},
			{
				"contradictionType": "temporal",
				"severity": "info",
				"conflictingText": "Events happen in March 1925",
				"establishedFact": "Previous session set in January 1925",
				"source": "Session 3 notes",
				"description": "Two-month time gap is not explained.",
				"suggestion": "Add a brief bridging passage."
			},
			{
				"contradictionType": "character",
				"severity": "error",
				"conflictingText": "Dr. Armitage is young",
				"establishedFact": "Dr. Armitage is described as elderly",
				"source": "Entity: Dr. Armitage",
				"description": "Age description contradicts entity.",
				"suggestion": "Use 'elderly' consistently."
			}
		]
	}`

	provider := &mockProvider{response: llmResponse}
	expert := NewExpert()

	input := enrichment.PipelineInput{
		CampaignID:  1,
		JobID:       42,
		SourceTable: "chapters",
		SourceID:    7,
		Content:     "The investigators visit the inn on Baker Street in March 1925.",
		Context: &enrichment.RAGContext{
			CampaignResults: []models.SearchResult{
				{
					SourceTable:  "chapters",
					SourceID:     1,
					SourceName:   "Chapter 1",
					ChunkContent: "The Silver Fox Inn is on Fleet Street.",
				},
			},
		},
	}

	items, err := expert.Run(context.Background(), provider, input)

	require.NoError(t, err)
	require.Len(t, items, 3, "expected 3 contradiction items")

	for i, item := range items {
		assert.Equal(t, "analysis", item.Phase,
			"item %d should have Phase='analysis'", i)
	}
}

// ---------------------------------------------------------------------------
// Parser tests
// ---------------------------------------------------------------------------

func TestParseCanonResponse_Valid(t *testing.T) {
	input := `{
		"contradictions": [
			{
				"contradictionType": "factual",
				"severity": "warning",
				"conflictingText": "Viktor arrived in 1923",
				"establishedFact": "Viktor arrived in 1921",
				"source": "Chapter 1",
				"description": "Year of arrival contradicts canon.",
				"suggestion": "Use 1921."
			}
		]
	}`

	resp, err := parseCanonResponse(input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Contradictions, 1)
	assert.Equal(t, "factual", resp.Contradictions[0].ContradictionType)
	assert.Equal(t, "warning", resp.Contradictions[0].Severity)
	assert.Equal(t, "Viktor arrived in 1923", resp.Contradictions[0].ConflictingText)
	assert.Equal(t, "Viktor arrived in 1921", resp.Contradictions[0].EstablishedFact)
	assert.Equal(t, "Chapter 1", resp.Contradictions[0].Source)
	assert.Equal(t, "Year of arrival contradicts canon.", resp.Contradictions[0].Description)
	assert.Equal(t, "Use 1921.", resp.Contradictions[0].Suggestion)
}

func TestParseCanonResponse_WithCodeFences(t *testing.T) {
	input := "```json\n" + `{
		"contradictions": [
			{
				"contradictionType": "temporal",
				"severity": "info",
				"conflictingText": "Events in March",
				"establishedFact": "Events in January",
				"source": "Session 3",
				"description": "Time gap not explained.",
				"suggestion": "Add bridging text."
			}
		]
	}` + "\n```"

	resp, err := parseCanonResponse(input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Contradictions, 1)
	assert.Equal(t, "temporal", resp.Contradictions[0].ContradictionType)
}

func TestParseCanonResponse_Invalid(t *testing.T) {
	input := "This is not JSON at all, just prose from the LLM."

	resp, err := parseCanonResponse(input)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// ---------------------------------------------------------------------------
// stripCodeFences tests
// ---------------------------------------------------------------------------

func TestStripCodeFences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "json code fence",
			input:    "```json\n{\"key\": \"value\"}\n```",
			expected: "{\"key\": \"value\"}",
		},
		{
			name:     "plain code fence",
			input:    "```\n{\"key\": \"value\"}\n```",
			expected: "{\"key\": \"value\"}",
		},
		{
			name:     "no code fence",
			input:    "{\"key\": \"value\"}",
			expected: "{\"key\": \"value\"}",
		},
		{
			name:     "code fence with extra whitespace",
			input:    "  ```json\n{\"key\": \"value\"}\n```  ",
			expected: "{\"key\": \"value\"}",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only fences no content",
			input:    "```json\n\n```",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := agents.StripCodeFences(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// ---------------------------------------------------------------------------
// contradictionTypeToDetectionType tests
// ---------------------------------------------------------------------------

func TestContradictionTypeToDetectionType(t *testing.T) {
	tests := []struct {
		contradictionType string
		expected          string
	}{
		{"factual", "canon_contradiction"},
		{"temporal", "temporal_inconsistency"},
		{"character", "character_inconsistency"},
		{"unknown", "canon_contradiction"},
		{"", "canon_contradiction"},
	}

	for _, tc := range tests {
		t.Run(tc.contradictionType, func(t *testing.T) {
			result := contradictionTypeToDetectionType(tc.contradictionType)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// ---------------------------------------------------------------------------
// Prompt tests
// ---------------------------------------------------------------------------

func TestBuildUserPrompt_WithRAGContext(t *testing.T) {
	input := enrichment.PipelineInput{
		CampaignID:  1,
		JobID:       10,
		SourceTable: "chapters",
		SourceID:    5,
		Content:     "Viktor enters the Silver Fox Inn.",
		Context: &enrichment.RAGContext{
			CampaignResults: []models.SearchResult{
				{
					SourceTable:  "entities",
					SourceID:     100,
					SourceName:   "Viktor",
					ChunkContent: "Viktor is a Russian emigre who arrived in London in 1921.",
				},
				{
					SourceTable:  "chapters",
					SourceID:     1,
					SourceName:   "Chapter 1",
					ChunkContent: "The investigators first meet Viktor at the docks.",
				},
			},
		},
	}

	prompt := buildUserPrompt(input)

	assert.Contains(t, prompt, "Viktor enters the Silver Fox Inn.")
	assert.Contains(t, prompt, "Russian emigre")
	assert.Contains(t, prompt, "investigators first meet Viktor")
}

func TestBuildUserPrompt_WithEntities(t *testing.T) {
	viktorDesc := "A Russian emigre and former soldier."
	irenaDesc := "A librarian at the Bodleian Library."

	input := enrichment.PipelineInput{
		CampaignID:  1,
		JobID:       10,
		SourceTable: "chapters",
		SourceID:    5,
		Content:     "Viktor and Irena discuss the missing manuscript.",
		Entities: []models.Entity{
			{
				ID:          100,
				CampaignID:  1,
				Name:        "Viktor",
				EntityType:  models.EntityTypeNPC,
				Description: &viktorDesc,
			},
			{
				ID:          200,
				CampaignID:  1,
				Name:        "Irena",
				EntityType:  models.EntityTypeNPC,
				Description: &irenaDesc,
			},
		},
		Context: &enrichment.RAGContext{
			CampaignResults: []models.SearchResult{
				{
					SourceTable:  "chapters",
					SourceID:     1,
					SourceName:   "Chapter 1",
					ChunkContent: "Background context.",
				},
			},
		},
	}

	prompt := buildUserPrompt(input)

	assert.Contains(t, prompt, "Viktor and Irena discuss the missing manuscript.")
	assert.Contains(t, prompt, "Viktor")
	assert.Contains(t, prompt, "Russian emigre")
	assert.Contains(t, prompt, "Irena")
	assert.Contains(t, prompt, "librarian")
}

func TestBuildUserPrompt_WithScope(t *testing.T) {
	input := enrichment.PipelineInput{
		CampaignID:  1,
		JobID:       10,
		SourceTable: "sessions",
		SourceID:    5,
		SourceScope: enrichment.ScopeSession,
		Content:     "The investigators question the barkeep.",
		Context: &enrichment.RAGContext{
			CampaignResults: []models.SearchResult{
				{
					SourceTable:  "chapters",
					SourceID:     1,
					SourceName:   "Chapter 1",
					ChunkContent: "Background context.",
				},
			},
		},
	}

	prompt := buildUserPrompt(input)

	assert.Contains(t, prompt, "- Scope: session")
}

func TestBuildUserPrompt_WithoutScope(t *testing.T) {
	input := enrichment.PipelineInput{
		CampaignID:  1,
		JobID:       10,
		SourceTable: "chapters",
		SourceID:    5,
		Content:     "A brief chapter overview.",
		Context: &enrichment.RAGContext{
			CampaignResults: []models.SearchResult{
				{
					SourceTable:  "chapters",
					SourceID:     1,
					SourceName:   "Chapter 1",
					ChunkContent: "Background context.",
				},
			},
		},
	}

	prompt := buildUserPrompt(input)

	assert.NotContains(t, prompt, "Scope:")
}

// ---------------------------------------------------------------------------
// System prompt tests
// ---------------------------------------------------------------------------

func TestBuildSystemPrompt(t *testing.T) {
	prompt := buildSystemPrompt("")

	assert.NotEmpty(t, prompt)
	// The system prompt should instruct the LLM to return JSON.
	assert.Contains(t, prompt, "JSON")
	// It should mention the expected response structure.
	assert.Contains(t, prompt, "contradictions")
}

func TestBuildSystemPrompt_Scoped(t *testing.T) {
	tests := []struct {
		scope    enrichment.SourceScope
		contains string
	}{
		{enrichment.ScopeCampaign, "Campaign Overview"},
		{enrichment.ScopeChapter, "Chapter Overview"},
		{enrichment.ScopeSession, "Session Notes"},
		{enrichment.ScopeEntity, "Entity Description"},
	}
	for _, tc := range tests {
		t.Run(string(tc.scope), func(t *testing.T) {
			prompt := buildSystemPrompt(tc.scope)
			assert.Contains(t, prompt, tc.contains)
			// Should still contain the standard sections.
			assert.Contains(t, prompt, "JSON")
			assert.Contains(t, prompt, "contradictions")
		})
	}
}

func TestBuildSystemPrompt_UnknownScope(t *testing.T) {
	prompt := buildSystemPrompt("unknown")

	assert.NotEmpty(t, prompt)
	// Should still contain the standard sections but no scope heading.
	assert.Contains(t, prompt, "JSON")
	assert.NotContains(t, prompt, "## Scope:")
}
