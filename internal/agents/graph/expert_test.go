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
	"context"
	"errors"
	"testing"

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
	expert := NewExpert(nil)
	assert.Equal(t, "graph-expert", expert.Name())
}

func TestExpert_DependsOn(t *testing.T) {
	expert := NewExpert(nil)
	deps := expert.DependsOn()
	require.NotNil(t, deps)
	assert.Equal(t, []string{"enrichment"}, deps)
}

// ---------------------------------------------------------------------------
// Run tests
// ---------------------------------------------------------------------------

func TestExpert_Run_NoEntities(t *testing.T) {
	provider := &mockProvider{
		response: `{"findings":[]}`,
	}
	expert := NewExpert(nil)

	input := enrichment.PipelineInput{
		CampaignID:  1,
		JobID:       10,
		SourceTable: "chapters",
		SourceID:    5,
		Content:     "Some chapter content.",
		Entities:    []models.Entity{},
	}

	items, err := expert.Run(context.Background(), provider, input)

	require.NoError(t, err)
	assert.Empty(t, items)
	assert.False(t, provider.called,
		"LLM should not be called when there are no entities")
}

func TestCheckOrphanedEntities_AllConnected(t *testing.T) {
	entities := []models.Entity{
		{ID: 1, Name: "Viktor", EntityType: models.EntityTypeNPC},
		{ID: 2, Name: "Silver Fox Inn", EntityType: models.EntityTypeLocation},
	}

	relationships := []models.Relationship{
		{
			SourceEntityID: 1,
			TargetEntityID: 2,
		},
	}

	orphans := CheckOrphanedEntities(entities, relationships)
	assert.Empty(t, orphans, "all entities are connected, no orphans expected")
}

func TestCheckOrphanedEntities_SomeOrphans(t *testing.T) {
	entities := []models.Entity{
		{ID: 1, Name: "Viktor", EntityType: models.EntityTypeNPC},
		{ID: 2, Name: "Silver Fox Inn", EntityType: models.EntityTypeLocation},
		{ID: 3, Name: "Mysterious Key", EntityType: models.EntityTypeItem},
	}

	// Only entity 1 and 2 are connected; entity 3 is orphaned.
	relationships := []models.Relationship{
		{
			SourceEntityID: 1,
			TargetEntityID: 2,
		},
	}

	orphans := CheckOrphanedEntities(entities, relationships)
	require.Len(t, orphans, 1)
	assert.Equal(t, int64(3), orphans[0].ID)
	assert.Equal(t, "Mysterious Key", orphans[0].Name)
}

func TestCheckOrphanedEntities_Empty(t *testing.T) {
	orphans := CheckOrphanedEntities(
		[]models.Entity{},
		[]models.Relationship{},
	)
	assert.Empty(t, orphans)
}

func TestExpert_Run_OrphanDetection(t *testing.T) {
	// With nil provider, the LLM semantic check is skipped but
	// structural checks (orphan detection) should still run.
	expert := NewExpert(nil)

	entities := []models.Entity{
		{ID: 1, Name: "Viktor", EntityType: models.EntityTypeNPC},
		{ID: 2, Name: "Silver Fox Inn", EntityType: models.EntityTypeLocation},
		{ID: 3, Name: "Mysterious Key", EntityType: models.EntityTypeItem},
	}

	// Only entities 1 and 2 are connected.
	relationships := []models.Relationship{
		{
			SourceEntityID: 1,
			TargetEntityID: 2,
		},
	}

	input := enrichment.PipelineInput{
		CampaignID:    1,
		JobID:         42,
		SourceTable:   "chapters",
		SourceID:      7,
		Content:       "Viktor visits the Silver Fox Inn.",
		Entities:      entities,
		Relationships: relationships,
	}

	items, err := expert.Run(context.Background(), nil, input)

	require.NoError(t, err)
	require.NotEmpty(t, items, "should produce orphan_warning items")

	// Find the orphan warning item for entity 3.
	var foundOrphan bool
	for _, item := range items {
		if item.DetectionType == "orphan_warning" {
			foundOrphan = true
			assert.Equal(t, int64(42), item.JobID)
			assert.Equal(t, "pending", item.Resolution)
			assert.Equal(t, "enrichment", item.Phase)
			assert.Contains(t, item.MatchedText, "Mysterious Key")
		}
	}
	assert.True(t, foundOrphan,
		"expected at least one orphan_warning item for Mysterious Key")
}

func TestExpert_Run_LLMError(t *testing.T) {
	provider := &mockProvider{
		err: errors.New("API rate limit exceeded"),
	}
	expert := NewExpert(nil)

	entities := []models.Entity{
		{ID: 1, Name: "Viktor", EntityType: models.EntityTypeNPC},
		{ID: 2, Name: "Silver Fox Inn", EntityType: models.EntityTypeLocation},
		{ID: 3, Name: "Orphan Item", EntityType: models.EntityTypeItem},
	}

	relationships := []models.Relationship{
		{
			SourceEntityID: 1,
			TargetEntityID: 2,
		},
	}

	input := enrichment.PipelineInput{
		CampaignID:    1,
		JobID:         42,
		SourceTable:   "chapters",
		SourceID:      7,
		Content:       "Viktor visits the Silver Fox Inn.",
		Entities:      entities,
		Relationships: relationships,
	}

	items, err := expert.Run(context.Background(), provider, input)

	// Graceful degradation: LLM error should not prevent structural
	// results from being returned.
	require.NoError(t, err,
		"LLM failure should be logged, not returned as error")
	require.NotEmpty(t, items,
		"structural checks (orphans) should still produce items")

	// Verify structural items are present despite LLM failure.
	var hasOrphan bool
	for _, item := range items {
		if item.DetectionType == "orphan_warning" {
			hasOrphan = true
		}
	}
	assert.True(t, hasOrphan,
		"orphan warnings should be returned even when LLM fails")
}

func TestExpert_Run_AllItemsHaveEnrichmentPhase(t *testing.T) {
	llmResponse := `{
		"findings": [
			{
				"findingType": "redundant_edge",
				"description": "Duplicate relationship detected.",
				"involvedEntities": ["Viktor", "Silver Fox Inn"],
				"suggestion": "Remove the redundant edge."
			}
		]
	}`

	provider := &mockProvider{response: llmResponse}
	expert := NewExpert(nil)

	entities := []models.Entity{
		{ID: 1, Name: "Viktor", EntityType: models.EntityTypeNPC},
		{ID: 2, Name: "Silver Fox Inn", EntityType: models.EntityTypeLocation},
	}

	relationships := []models.Relationship{
		{
			SourceEntityID: 1,
			TargetEntityID: 2,
		},
	}

	input := enrichment.PipelineInput{
		CampaignID:    1,
		JobID:         42,
		SourceTable:   "chapters",
		SourceID:      7,
		Content:       "Viktor visits the Silver Fox Inn.",
		Entities:      entities,
		Relationships: relationships,
	}

	items, err := expert.Run(context.Background(), provider, input)

	require.NoError(t, err)
	require.NotEmpty(t, items)

	for i, item := range items {
		assert.Equal(t, "enrichment", item.Phase,
			"item %d should have Phase='enrichment'", i)
	}
}

// ---------------------------------------------------------------------------
// Parser tests
// ---------------------------------------------------------------------------

func TestParseGraphResponse_Valid(t *testing.T) {
	input := `{
		"findings": [
			{
				"findingType": "redundant_edge",
				"description": "Duplicate relationship detected.",
				"involvedEntities": ["Viktor", "Silver Fox Inn"],
				"suggestion": "Remove the redundant edge."
			},
			{
				"findingType": "implied_edge",
				"description": "Relationship is implied through faction.",
				"involvedEntities": ["Alice", "Bob"],
				"suggestion": "Remove direct edge; traverse via faction."
			}
		]
	}`

	resp, err := parseGraphResponse(input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Findings, 2)
	assert.Equal(t, "redundant_edge", resp.Findings[0].FindingType)
	assert.Equal(t, "Duplicate relationship detected.",
		resp.Findings[0].Description)
	assert.Equal(t, "Remove the redundant edge.",
		resp.Findings[0].Suggestion)
	assert.Equal(t, []string{"Viktor", "Silver Fox Inn"},
		resp.Findings[0].InvolvedEntities)
	assert.Equal(t, "implied_edge", resp.Findings[1].FindingType)
}

func TestParseGraphResponse_WithCodeFences(t *testing.T) {
	input := "```json\n" + `{
		"findings": [
			{
				"findingType": "redundant_edge",
				"description": "Entity has many connections.",
				"involvedEntities": ["Viktor"],
				"suggestion": "Review for hub overload."
			}
		]
	}` + "\n```"

	resp, err := parseGraphResponse(input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Findings, 1)
	assert.Equal(t, "redundant_edge", resp.Findings[0].FindingType)
}

func TestParseGraphResponse_Invalid(t *testing.T) {
	input := "This is not JSON at all, just prose from the LLM."

	resp, err := parseGraphResponse(input)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestParseGraphResponse_UnknownFindingType(t *testing.T) {
	input := `{
		"findings": [
			{
				"findingType": "unknown_type",
				"description": "Something unusual detected.",
				"involvedEntities": ["Entity A"],
				"suggestion": "Investigate further."
			}
		]
	}`

	resp, err := parseGraphResponse(input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Findings, 1)
	// Unknown types are normalised to "redundant_edge".
	assert.Equal(t, "redundant_edge", resp.Findings[0].FindingType)
}

func TestParseGraphResponse_EmptyDescription(t *testing.T) {
	input := `{
		"findings": [
			{
				"findingType": "redundant_edge",
				"description": "",
				"involvedEntities": ["Viktor"],
				"suggestion": "Remove the edge."
			}
		]
	}`

	resp, err := parseGraphResponse(input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Findings,
		"findings with empty description should be skipped")
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
			result := stripCodeFences(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// ---------------------------------------------------------------------------
// findingTypeToDetectionType tests
// ---------------------------------------------------------------------------

func TestFindingTypeToDetectionType(t *testing.T) {
	tests := []struct {
		findingType string
		expected    string
	}{
		{"redundant_edge", "redundant_edge"},
		{"implied_edge", "redundant_edge"},
		{"unknown_type", "graph_warning"},
		{"", "graph_warning"},
	}

	for _, tc := range tests {
		t.Run(tc.findingType, func(t *testing.T) {
			result := findingTypeToDetectionType(tc.findingType)
			assert.Equal(t, tc.expected, result)
		})
	}
}
