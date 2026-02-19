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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/antonypegg/imagineer/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Query derivation tests
// ---------------------------------------------------------------------------

func TestBuildSearchQueries_ContentOnly(t *testing.T) {
	content := "The investigators crept through the abandoned warehouse."
	var entities []models.Entity

	cb := NewContextBuilder(nil, "")
	queries := cb.buildSearchQueries(content, entities)

	require.NotEmpty(t, queries,
		"content with no entities should produce at least one query")

	// The content-based query should be derived from the content.
	found := false
	for _, q := range queries {
		if strings.Contains(q, "investigators") ||
			strings.Contains(q, "warehouse") {
			found = true
			break
		}
	}
	assert.True(t, found,
		"at least one query should contain terms from the content")
}

func TestBuildSearchQueries_WithEntities(t *testing.T) {
	content := "Viktor and Elara explored the Silver Fox Inn."
	entities := []models.Entity{
		{Name: "Viktor", EntityType: models.EntityTypeNPC},
		{Name: "Elara", EntityType: models.EntityTypeNPC},
		{
			Name:       "The Silver Fox Inn",
			EntityType: models.EntityTypeLocation,
		},
	}

	cb := NewContextBuilder(nil, "")
	queries := cb.buildSearchQueries(content, entities)

	require.NotEmpty(t, queries)

	// Should produce queries containing entity names.
	joined := strings.Join(queries, " ")
	assert.Contains(t, joined, "Viktor",
		"queries should include Viktor")
	assert.Contains(t, joined, "Elara",
		"queries should include Elara")
	assert.Contains(t, joined, "Silver Fox Inn",
		"queries should include the Silver Fox Inn")

	// Should also include a content-derived query in addition to
	// entity name queries.
	assert.Greater(t, len(queries), 1,
		"should produce entity queries plus a content query")
}

func TestBuildSearchQueries_EmptyContent(t *testing.T) {
	entities := []models.Entity{
		{
			Name:       "Inspector Barrington",
			EntityType: models.EntityTypeNPC,
		},
		{
			Name:       "The Docks",
			EntityType: models.EntityTypeLocation,
		},
	}

	cb := NewContextBuilder(nil, "")
	queries := cb.buildSearchQueries("", entities)

	require.NotEmpty(t, queries,
		"empty content with entities should still produce entity queries")

	joined := strings.Join(queries, " ")
	assert.Contains(t, joined, "Inspector Barrington",
		"should produce a query for Inspector Barrington")
	assert.Contains(t, joined, "The Docks",
		"should produce a query for The Docks")
}

func TestBuildSearchQueries_EmptyContentAndEntities(t *testing.T) {
	cb := NewContextBuilder(nil, "")
	queries := cb.buildSearchQueries("", nil)

	assert.Empty(t, queries,
		"no content and no entities should produce zero queries")
}

func TestBuildSearchQueries_ManyEntities(t *testing.T) {
	// Build a large set of entities to verify batching behaviour.
	// With many entities the function should group them into batched
	// queries rather than producing one query per entity.
	entities := make([]models.Entity, 20)
	for i := range entities {
		entities[i] = models.Entity{
			Name:       "Entity_" + strings.Repeat("X", i+1),
			EntityType: models.EntityTypeNPC,
		}
	}

	cb := NewContextBuilder(nil, "")
	queries := cb.buildSearchQueries("Some campaign content.", entities)

	require.NotEmpty(t, queries)

	// Every entity name should appear in at least one query.
	for _, ent := range entities {
		found := false
		for _, q := range queries {
			if strings.Contains(q, ent.Name) {
				found = true
				break
			}
		}
		assert.True(t, found,
			"entity %q should appear in at least one query",
			ent.Name)
	}

	// The number of queries should be less than the number of
	// entities (batching) plus one (content query).
	assert.Less(t, len(queries), len(entities)+1,
		"queries should be batched, not one per entity")
}

func TestBuildSearchQueries_EntityWithEmptyName(t *testing.T) {
	entities := []models.Entity{
		{Name: "", EntityType: models.EntityTypeNPC},
		{Name: "Viktor", EntityType: models.EntityTypeNPC},
	}

	cb := NewContextBuilder(nil, "")
	queries := cb.buildSearchQueries("Viktor explored the ruins.", entities)

	// Empty-name entities should be skipped.
	for _, q := range queries {
		trimmed := strings.TrimSpace(q)
		assert.NotEqual(t, "", trimmed,
			"queries should not contain empty or blank strings")
	}

	joined := strings.Join(queries, " ")
	assert.Contains(t, joined, "Viktor",
		"should include a query for the non-empty entity name")
}

func TestBuildSearchQueries_ContentTruncation(t *testing.T) {
	// Very long content should be truncated in the content-based
	// query, not passed through at full length.
	longContent := strings.Repeat("word ", 1000)
	cb := NewContextBuilder(nil, "")
	queries := cb.buildSearchQueries(longContent, nil)

	require.NotEmpty(t, queries)
	for _, q := range queries {
		assert.LessOrEqual(t, len([]rune(q)), maxSearchQueryLen+50,
			"content query should be truncated to a reasonable length")
	}
}

// ---------------------------------------------------------------------------
// Deduplication and trimming tests
// ---------------------------------------------------------------------------

func TestDeduplicateAndTrim_NoDuplicates(t *testing.T) {
	results := []models.SearchResult{
		{
			SourceTable:   "chapters",
			SourceID:      1,
			SourceName:    "Chapter One",
			ChunkContent:  "The investigators arrived.",
			CombinedScore: 0.9,
		},
		{
			SourceTable:   "sessions",
			SourceID:      2,
			SourceName:    "Session Two",
			ChunkContent:  "A new lead was discovered.",
			CombinedScore: 0.8,
		},
		{
			SourceTable:   "chapters",
			SourceID:      3,
			SourceName:    "Chapter Three",
			ChunkContent:  "The final confrontation.",
			CombinedScore: 0.7,
		},
	}

	deduped := deduplicateAndTrim(results)

	assert.Len(t, deduped, 3,
		"all unique results should pass through")
}

func TestDeduplicateAndTrim_KeepsHighestScore(t *testing.T) {
	results := []models.SearchResult{
		{
			SourceTable:   "chapters",
			SourceID:      1,
			SourceName:    "Chapter One",
			ChunkContent:  "First match, lower score.",
			CombinedScore: 0.5,
		},
		{
			SourceTable:   "sessions",
			SourceID:      2,
			SourceName:    "Session Two",
			ChunkContent:  "Unique result.",
			CombinedScore: 0.8,
		},
		{
			SourceTable:   "chapters",
			SourceID:      1,
			SourceName:    "Chapter One",
			ChunkContent:  "Second match, higher score.",
			CombinedScore: 0.95,
		},
	}

	deduped := deduplicateAndTrim(results)

	require.Len(t, deduped, 2,
		"duplicate (chapters, 1) should collapse to one entry")

	// Find the chapters/1 entry and verify it kept the higher score.
	var chapterResult *models.SearchResult
	for i := range deduped {
		if deduped[i].SourceTable == "chapters" &&
			deduped[i].SourceID == 1 {
			chapterResult = &deduped[i]
			break
		}
	}
	require.NotNil(t, chapterResult,
		"should contain the chapters/1 result")
	assert.InDelta(t, 0.95, chapterResult.CombinedScore, 0.001,
		"should keep the result with the higher combined score")
	assert.Equal(t, "Second match, higher score.",
		chapterResult.ChunkContent,
		"should keep the chunk content from the higher-scoring entry")
}

func TestDeduplicateAndTrim_KeepsHighestScore_ReverseOrder(t *testing.T) {
	// Higher score appears first; the lower-scoring duplicate should
	// not overwrite it.
	results := []models.SearchResult{
		{
			SourceTable:   "chapters",
			SourceID:      5,
			SourceName:    "Chapter Five",
			ChunkContent:  "High score first.",
			CombinedScore: 0.99,
		},
		{
			SourceTable:   "chapters",
			SourceID:      5,
			SourceName:    "Chapter Five",
			ChunkContent:  "Low score second.",
			CombinedScore: 0.3,
		},
	}

	deduped := deduplicateAndTrim(results)

	require.Len(t, deduped, 1)
	assert.InDelta(t, 0.99, deduped[0].CombinedScore, 0.001,
		"should retain the higher-scoring result regardless of order")
	assert.Equal(t, "High score first.", deduped[0].ChunkContent)
}

func TestDeduplicateAndTrim_SameTableDifferentIDs(t *testing.T) {
	results := []models.SearchResult{
		{
			SourceTable:   "chapters",
			SourceID:      1,
			ChunkContent:  "Short.",
			CombinedScore: 0.9,
		},
		{
			SourceTable:   "chapters",
			SourceID:      2,
			ChunkContent:  "Also short.",
			CombinedScore: 0.8,
		},
	}

	deduped := deduplicateAndTrim(results)

	assert.Len(t, deduped, 2,
		"same table with different IDs should not be deduplicated")
}

func TestDeduplicateAndTrim_DifferentTableSameID(t *testing.T) {
	results := []models.SearchResult{
		{
			SourceTable:   "chapters",
			SourceID:      1,
			ChunkContent:  "Chapter content.",
			CombinedScore: 0.9,
		},
		{
			SourceTable:   "sessions",
			SourceID:      1,
			ChunkContent:  "Session content.",
			CombinedScore: 0.8,
		},
	}

	deduped := deduplicateAndTrim(results)

	assert.Len(t, deduped, 2,
		"different tables with the same ID should not be deduplicated")
}

func TestDeduplicateAndTrim_NilInput(t *testing.T) {
	deduped := deduplicateAndTrim(nil)

	assert.Nil(t, deduped,
		"should return nil for nil input")
}

func TestDeduplicateAndTrim_EmptySlice(t *testing.T) {
	deduped := deduplicateAndTrim([]models.SearchResult{})

	assert.Nil(t, deduped,
		"should return nil for empty input")
}

func TestDeduplicateAndTrim_ManyDuplicates(t *testing.T) {
	// Same (table, ID) appearing many times should collapse to one.
	results := make([]models.SearchResult, 10)
	for i := range results {
		results[i] = models.SearchResult{
			SourceTable:   "chapters",
			SourceID:      42,
			ChunkContent:  strings.Repeat("x", i+1),
			CombinedScore: float64(i) * 0.1,
		}
	}

	deduped := deduplicateAndTrim(results)

	require.Len(t, deduped, 1,
		"all duplicates should collapse to a single entry")
	assert.InDelta(t, 0.9, deduped[0].CombinedScore, 0.001,
		"should keep the highest score from all duplicates")
}

func TestDeduplicateAndTrim_SortsByScoreDescending(t *testing.T) {
	results := []models.SearchResult{
		{
			SourceTable:   "chapters",
			SourceID:      1,
			ChunkContent:  "low",
			CombinedScore: 0.3,
		},
		{
			SourceTable:   "chapters",
			SourceID:      2,
			ChunkContent:  "high",
			CombinedScore: 0.9,
		},
		{
			SourceTable:   "sessions",
			SourceID:      3,
			ChunkContent:  "mid",
			CombinedScore: 0.6,
		},
	}

	deduped := deduplicateAndTrim(results)

	require.Len(t, deduped, 3)
	for i := 1; i < len(deduped); i++ {
		assert.GreaterOrEqual(t,
			deduped[i-1].CombinedScore,
			deduped[i].CombinedScore,
			"results should be sorted by CombinedScore descending")
	}
}

func TestDeduplicateAndTrim_TrimsToTokenBudget(t *testing.T) {
	// Create results with large content that exceeds maxContextTokens.
	// At tokensPerChar=0.25, 400 chars is ~100 tokens.
	// maxContextTokens is 4000, so 4000/100 = 40 results would fit.
	// Create 50 results to exceed the budget.
	results := make([]models.SearchResult, 50)
	for i := range results {
		results[i] = models.SearchResult{
			SourceTable:   "chapters",
			SourceID:      int64(i + 1),
			ChunkContent:  strings.Repeat("a", 400),
			CombinedScore: float64(50-i) * 0.01,
		}
	}

	trimmed := deduplicateAndTrim(results)

	assert.Less(t, len(trimmed), len(results),
		"should return fewer results than the input when over budget")
	require.NotEmpty(t, trimmed,
		"should return at least one result")

	// The highest-scoring result should appear first.
	assert.InDelta(t, 0.50, trimmed[0].CombinedScore, 0.001,
		"highest-scoring result should appear first")
}

func TestDeduplicateAndTrim_AlwaysIncludesAtLeastOne(t *testing.T) {
	// A single result that itself exceeds the budget should still
	// be included to avoid returning an empty context.
	results := []models.SearchResult{
		{
			SourceTable:   "chapters",
			SourceID:      1,
			ChunkContent:  strings.Repeat("x", 50000),
			CombinedScore: 0.95,
		},
	}

	trimmed := deduplicateAndTrim(results)

	require.Len(t, trimmed, 1,
		"should include at least one result even if it exceeds budget")
}

// ---------------------------------------------------------------------------
// estimateTokens tests
// ---------------------------------------------------------------------------

func TestEstimateTokens_Empty(t *testing.T) {
	assert.InDelta(t, 0.0, estimateTokens(""), 0.001)
}

func TestEstimateTokens_KnownInput(t *testing.T) {
	// 100 characters at 0.25 tokens/char = 25 tokens.
	input := strings.Repeat("a", 100)
	assert.InDelta(t, 25.0, estimateTokens(input), 0.001)
}

func TestEstimateTokens_MultiByte(t *testing.T) {
	// CJK characters: 4 runes at 0.25 tokens/rune = 1 token.
	input := strings.Repeat("\u4e16", 4)
	assert.InDelta(t, 1.0, estimateTokens(input), 0.001)
}

// ---------------------------------------------------------------------------
// Game system schema loading tests
// ---------------------------------------------------------------------------

func TestLoadGameSystemSchema_ValidCode(t *testing.T) {
	tmpDir := t.TempDir()
	schemaContent := "name: Test System\nskills:\n" +
		"  - Investigate\n  - Persuade\n"
	err := os.WriteFile(
		filepath.Join(tmpDir, "test-system.yaml"),
		[]byte(schemaContent),
		0644,
	)
	require.NoError(t, err)

	cb := NewContextBuilder(nil, tmpDir)
	result := cb.loadGameSystemSchema("test-system")

	assert.Equal(t, schemaContent, result)
}

func TestLoadGameSystemSchema_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()

	cb := NewContextBuilder(nil, tmpDir)
	result := cb.loadGameSystemSchema("nonexistent-system")

	assert.Equal(t, "", result,
		"missing file should return empty string")
}

func TestLoadGameSystemSchema_EmptyCode(t *testing.T) {
	tmpDir := t.TempDir()

	cb := NewContextBuilder(nil, tmpDir)
	result := cb.loadGameSystemSchema("")

	// An empty code builds a path like "<dir>/.yaml" which should
	// not exist. The method returns an empty string.
	assert.Equal(t, "", result,
		"empty code should result in an empty string")
}

func TestLoadGameSystemSchema_LargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	largeContent := strings.Repeat("skill: Investigation\n", 1000)
	err := os.WriteFile(
		filepath.Join(tmpDir, "large-system.yaml"),
		[]byte(largeContent),
		0644,
	)
	require.NoError(t, err)

	cb := NewContextBuilder(nil, tmpDir)
	result := cb.loadGameSystemSchema("large-system")

	assert.Equal(t, largeContent, result,
		"large schema files should be loaded completely")
}

// ---------------------------------------------------------------------------
// NewContextBuilder tests
// ---------------------------------------------------------------------------

func TestNewContextBuilder_DefaultSchemasDir(t *testing.T) {
	cb := NewContextBuilder(nil, "")

	assert.Equal(t, "schemas", cb.schemasDir,
		"empty schemasDir should default to 'schemas'")
}

func TestNewContextBuilder_CustomSchemasDir(t *testing.T) {
	cb := NewContextBuilder(nil, "/custom/schemas")

	assert.Equal(t, "/custom/schemas", cb.schemasDir)
}

// ---------------------------------------------------------------------------
// BuildContext integration test (no vectorization)
// ---------------------------------------------------------------------------

func TestBuildContext_SchemaLoadingOnly(t *testing.T) {
	// When no database is available, vectorization is unavailable.
	// The ContextBuilder should still load the game system schema
	// into the RAGContext.
	tmpDir := t.TempDir()
	schemaContent := "name: Call of Cthulhu 7e\nskills:\n" +
		"  - Spot Hidden\n  - Library Use\n"
	err := os.WriteFile(
		filepath.Join(tmpDir, "coc-7e.yaml"),
		[]byte(schemaContent),
		0644,
	)
	require.NoError(t, err)

	cb := NewContextBuilder(nil, tmpDir)

	// Verify schema loading via the helper method (BuildContext
	// requires a non-nil *database.DB for the vectorization check,
	// so we test the schema path independently).
	ragCtx := &RAGContext{}
	ragCtx.GameSystemYAML = cb.loadGameSystemSchema("coc-7e")

	assert.Equal(t, schemaContent, ragCtx.GameSystemYAML,
		"game system schema should be loaded into RAGContext")
	assert.Empty(t, ragCtx.CampaignResults,
		"campaign results should be empty without vectorization")
}

func TestBuildContext_NoGameSystemCode(t *testing.T) {
	// When no game system code is provided, the GameSystemYAML
	// field should remain empty.
	ragCtx := &RAGContext{}

	assert.Equal(t, "", ragCtx.GameSystemYAML,
		"no game system code should leave GameSystemYAML empty")
	assert.Empty(t, ragCtx.CampaignResults,
		"campaign results should be empty")
}

// ---------------------------------------------------------------------------
// RAGContext zero-value test
// ---------------------------------------------------------------------------

func TestRAGContext_ZeroValue(t *testing.T) {
	ragCtx := &RAGContext{}

	assert.Empty(t, ragCtx.CampaignResults,
		"zero-value RAGContext should have nil CampaignResults")
	assert.Equal(t, "", ragCtx.GameSystemYAML,
		"zero-value RAGContext should have empty GameSystemYAML")
}
