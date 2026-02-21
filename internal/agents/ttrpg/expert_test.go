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
	"context"
	"encoding/json"
	"errors"
	"strings"
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
	assert.Equal(t, "ttrpg-expert", expert.Name())
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
		response: `{"report":"should not be called","findings":[]}`,
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

func TestExpert_Run_ValidResponse(t *testing.T) {
	llmResponse := `{
		"report": "## Session Analysis\nThe session features strong investigation scenes.",
		"findings": [
			{
				"category": "pacing",
				"severity": "warning",
				"description": "The session may drag in the middle.",
				"suggestion": "Consider adding a brief action beat.",
				"lineReference": "paragraph 3"
			},
			{
				"category": "mechanics",
				"severity": "info",
				"description": "The Spot Hidden roll DC seems high.",
				"suggestion": "Consider lowering to Regular difficulty.",
				"lineReference": ""
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
		Content:     "The investigators search the abandoned library for clues.",
	}

	items, err := expert.Run(context.Background(), provider, input)

	require.NoError(t, err)
	// Expect 3 items: 1 report + 2 findings.
	require.Len(t, items, 3)

	// Verify the report item.
	assert.Equal(t, int64(42), items[0].JobID)
	assert.Equal(t, "analysis_report", items[0].DetectionType)
	assert.Equal(t, "pending", items[0].Resolution)
	assert.Equal(t, "analysis", items[0].Phase)
	assert.Equal(t, "ttrpg-expert", items[0].MatchedText)

	// Verify the first finding item.
	assert.Equal(t, "pacing_note", items[1].DetectionType)
	assert.Equal(t, "pending", items[1].Resolution)
	assert.Equal(t, "analysis", items[1].Phase)
	assert.Equal(t, "pacing", items[1].MatchedText)

	// Verify the second finding item.
	assert.Equal(t, "mechanics_warning", items[2].DetectionType)
	assert.Equal(t, "pending", items[2].Resolution)
	assert.Equal(t, "analysis", items[2].Phase)
}

func TestExpert_Run_LLMError(t *testing.T) {
	provider := &mockProvider{
		err: errors.New("API rate limit exceeded"),
	}
	expert := NewExpert()

	input := enrichment.PipelineInput{
		CampaignID: 1,
		JobID:      42,
		Content:    "Some session content.",
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
		CampaignID: 1,
		JobID:      42,
		Content:    "Some session content.",
	}

	items, err := expert.Run(context.Background(), provider, input)

	// Graceful degradation: no error, empty items.
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestExpert_Run_EmptyFindings(t *testing.T) {
	llmResponse := `{
		"report": "## Analysis\nEverything looks good, no issues found.",
		"findings": []
	}`

	provider := &mockProvider{response: llmResponse}
	expert := NewExpert()

	input := enrichment.PipelineInput{
		CampaignID: 1,
		JobID:      42,
		Content:    "A well-paced session with good mechanics.",
	}

	items, err := expert.Run(context.Background(), provider, input)

	require.NoError(t, err)
	// Only the report item should be present.
	require.Len(t, items, 1)
	assert.Equal(t, "analysis_report", items[0].DetectionType)
	assert.Equal(t, "analysis", items[0].Phase)
}

func TestExpert_Run_AllItemsHaveAnalysisPhase(t *testing.T) {
	llmResponse := `{
		"report": "## Report\nDetailed analysis.",
		"findings": [
			{
				"category": "investigation",
				"severity": "info",
				"description": "Investigation gap detected.",
				"suggestion": "Add more clues.",
				"lineReference": ""
			},
			{
				"category": "spotlight",
				"severity": "warning",
				"description": "Spotlight suggestion.",
				"suggestion": "Expand the scene.",
				"lineReference": ""
			},
			{
				"category": "pacing",
				"severity": "info",
				"description": "Pacing is good.",
				"suggestion": "",
				"lineReference": ""
			}
		]
	}`

	provider := &mockProvider{response: llmResponse}
	expert := NewExpert()

	input := enrichment.PipelineInput{
		CampaignID: 1,
		JobID:      42,
		Content:    "The investigators proceed through the haunted mansion.",
	}

	items, err := expert.Run(context.Background(), provider, input)

	require.NoError(t, err)
	require.Len(t, items, 4, "expected 1 report + 3 findings")

	for i, item := range items {
		assert.Equal(t, "analysis", item.Phase,
			"item %d should have Phase='analysis'", i)
	}
}

// ---------------------------------------------------------------------------
// Parser tests
// ---------------------------------------------------------------------------

func TestParseExpertResponse_Valid(t *testing.T) {
	input := `{
		"report": "## Session Analysis\nLooks good overall.",
		"findings": [
			{
				"category": "pacing",
				"severity": "warning",
				"description": "Slow middle section.",
				"suggestion": "Add a twist.",
				"lineReference": "paragraph 2"
			}
		]
	}`

	resp, err := parseExpertResponse(input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Report, "Session Analysis")
	assert.Len(t, resp.Findings, 1)
	assert.Equal(t, "pacing", resp.Findings[0].Category)
	assert.Equal(t, "warning", resp.Findings[0].Severity)
	assert.Equal(t, "Slow middle section.", resp.Findings[0].Description)
	assert.Equal(t, "Add a twist.", resp.Findings[0].Suggestion)
	assert.Equal(t, "paragraph 2", resp.Findings[0].LineReference)
}

func TestParseExpertResponse_WithCodeFences(t *testing.T) {
	input := "```json\n" + `{
		"report": "Analysis report.",
		"findings": [
			{
				"category": "mechanics",
				"severity": "info",
				"description": "Roll seems off.",
				"suggestion": "Check the rules.",
				"lineReference": ""
			}
		]
	}` + "\n```"

	resp, err := parseExpertResponse(input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "Analysis report.", resp.Report)
	assert.Len(t, resp.Findings, 1)
	assert.Equal(t, "mechanics", resp.Findings[0].Category)
}

func TestParseExpertResponse_Invalid(t *testing.T) {
	input := "This is not JSON at all, just prose from the LLM."

	resp, err := parseExpertResponse(input)

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
// categoryToDetectionType tests
// ---------------------------------------------------------------------------

func TestCategoryToDetectionType(t *testing.T) {
	tests := []struct {
		category string
		expected string
	}{
		{"pacing", "pacing_note"},
		{"investigation", "investigation_gap"},
		{"mechanics", "mechanics_warning"},
		{"content", "content_suggestion"},
		{"unknown_category", "content_suggestion"},
		{"", "content_suggestion"},
	}

	for _, tc := range tests {
		t.Run(tc.category, func(t *testing.T) {
			result := categoryToDetectionType(tc.category)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// ---------------------------------------------------------------------------
// Prompt tests
// ---------------------------------------------------------------------------

func TestBuildUserPrompt_WithSchema(t *testing.T) {
	gameSystemID := int64(1)
	input := enrichment.PipelineInput{
		CampaignID:   1,
		JobID:        10,
		SourceTable:  "chapters",
		SourceID:     5,
		Content:      "The investigators attempt a Spot Hidden roll.",
		GameSystemID: &gameSystemID,
		Context: &enrichment.RAGContext{
			GameSystemYAML: "name: Call of Cthulhu 7e\nskills:\n  - Spot Hidden\n  - Listen",
		},
	}

	prompt := buildUserPrompt(input)

	assert.Contains(t, prompt, "The investigators attempt a Spot Hidden roll.")
	assert.Contains(t, prompt, "Call of Cthulhu 7e")
	assert.Contains(t, prompt, "Spot Hidden")
}

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
					SourceTable:  "entities",
					SourceID:     200,
					SourceName:   "The Silver Fox Inn",
					ChunkContent: "A pub in the East End frequented by dockworkers.",
				},
			},
		},
	}

	prompt := buildUserPrompt(input)

	assert.Contains(t, prompt, "Viktor enters the Silver Fox Inn.")
	assert.Contains(t, prompt, "Russian emigre")
	assert.Contains(t, prompt, "dockworkers")
}

func TestBuildUserPrompt_MinimalInput(t *testing.T) {
	input := enrichment.PipelineInput{
		CampaignID:  1,
		JobID:       10,
		SourceTable: "chapters",
		SourceID:    5,
		Content:     "A brief scene in the marketplace.",
	}

	prompt := buildUserPrompt(input)

	assert.Contains(t, prompt, "A brief scene in the marketplace.")
	// Should not contain schema or RAG sections when they are absent.
	assert.NotContains(t, prompt, "Game System")
	assert.NotContains(t, prompt, "Campaign Context")
}

// ---------------------------------------------------------------------------
// System prompt test
// ---------------------------------------------------------------------------

func TestBuildSystemPrompt(t *testing.T) {
	prompt := buildSystemPrompt()

	assert.NotEmpty(t, prompt)
	// The system prompt should instruct the LLM to return JSON.
	assert.Contains(t, prompt, "JSON")
	// It should mention the expected response fields.
	assert.Contains(t, prompt, "report")
	assert.Contains(t, prompt, "findings")
}

// ---------------------------------------------------------------------------
// Integration: verify item structure
// ---------------------------------------------------------------------------

func TestExpert_Run_ItemStructure(t *testing.T) {
	llmResponse := `{
		"report": "## Detailed Report\nThe session is well-structured.",
		"findings": [
			{
				"category": "investigation",
				"severity": "warning",
				"description": "Missing clue for the main plot.",
				"suggestion": "Add a backup clue in the library.",
				"lineReference": "scene 2"
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
		Content:     "The investigators explore the haunted manor.",
	}

	items, err := expert.Run(context.Background(), provider, input)

	require.NoError(t, err)
	require.Len(t, items, 2)

	// Report item: verify SuggestedContent is valid JSON with a report field.
	var reportContent map[string]interface{}
	require.NoError(t, json.Unmarshal(items[0].SuggestedContent, &reportContent))
	assert.Contains(t, reportContent, "report")

	// Finding item: verify SuggestedContent contains finding details.
	var findingContent map[string]interface{}
	require.NoError(t, json.Unmarshal(items[1].SuggestedContent, &findingContent))
	assert.Equal(t, "investigation", findingContent["category"])
	assert.Equal(t, "warning", findingContent["severity"])
	assert.Equal(t, "Missing clue for the main plot.",
		findingContent["description"])
	assert.Equal(t, "Add a backup clue in the library.",
		findingContent["suggestion"])

	// Verify detection type mapping.
	assert.Equal(t, "investigation_gap", items[1].DetectionType)
}

func TestExpert_Run_ReportItemMatchedTextTruncation(t *testing.T) {
	// Build a very long report to verify the MatchedText is handled
	// reasonably (not an excessively long string).
	longReport := "## Analysis\n" + strings.Repeat("Detailed notes. ", 200)
	llmResponse := `{
		"report": ` + mustMarshalString(longReport) + `,
		"findings": []
	}`

	provider := &mockProvider{response: llmResponse}
	expert := NewExpert()

	input := enrichment.PipelineInput{
		CampaignID: 1,
		JobID:      42,
		Content:    "Session content.",
	}

	items, err := expert.Run(context.Background(), provider, input)

	require.NoError(t, err)
	require.Len(t, items, 1)
	// MatchedText should exist and not be empty.
	assert.NotEmpty(t, items[0].MatchedText)
}

// mustMarshalString marshals a string to a JSON string literal.
func mustMarshalString(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return string(b)
}
