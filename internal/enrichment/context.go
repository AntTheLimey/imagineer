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
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/models"
)

// maxSearchQueryLen is the maximum number of characters from the source
// content used as the vector search query. Longer content is truncated
// to this length before being sent to SearchCampaignContent.
const maxSearchQueryLen = 200

// maxContextTokens is the soft limit for total RAG context tokens
// across all search results included in a single pipeline run.
const maxContextTokens = 4000

// tokensPerChar is a rough estimate of tokens per character, used for
// token budget tracking. English text averages around 0.25 tokens per
// character; this intentionally over-estimates slightly to stay within
// the budget.
const tokensPerChar = 0.25

// maxContentSummaryLen is the maximum number of characters taken from
// the beginning of the source content for the content-summary query.
const maxContentSummaryLen = 150

// entityBatchSize is the number of entity names grouped into a single
// vector search query.
const entityBatchSize = 5

// searchLimitPerQuery is the maximum number of results requested from
// each individual search query.
const searchLimitPerQuery = 10

// ContextBuilder assembles shared RAG context for the enrichment
// pipeline. It performs vector search against campaign content and
// loads game system schema YAML, degrading gracefully when either
// source is unavailable.
type ContextBuilder struct {
	db         *database.DB
	schemasDir string
}

// NewContextBuilder creates a ContextBuilder. If schemasDir is empty
// it defaults to "schemas".
func NewContextBuilder(db *database.DB, schemasDir string) *ContextBuilder {
	if schemasDir == "" {
		schemasDir = "schemas"
	}
	return &ContextBuilder{
		db:         db,
		schemasDir: schemasDir,
	}
}

// BuildContext assembles a RAGContext by deriving multiple search
// queries from the source content and entity names, executing them
// via hybrid vector search, deduplicating and trimming results to
// fit within a token budget, and loading the game system schema YAML.
// Both search and schema sources are optional: if vectorization is
// unavailable or the schema file cannot be read, the corresponding
// field is left empty and no error is returned.
func (cb *ContextBuilder) BuildContext(
	ctx context.Context,
	campaignID int64,
	content string,
	gameSystemCode string,
	entities []models.Entity,
) (*RAGContext, error) {
	ragCtx := &RAGContext{}

	// Retrieve relevant campaign content via hybrid vector search
	// using multiple content-derived queries.
	if cb.db.IsVectorizationAvailable(ctx) {
		queries := cb.buildSearchQueries(content, entities)
		var allResults []models.SearchResult

		for _, query := range queries {
			if query == "" {
				continue
			}
			results, err := cb.db.SearchCampaignContent(
				ctx, campaignID, query, searchLimitPerQuery,
			)
			if err != nil {
				log.Printf(
					"enrichment: vector search failed for "+
						"campaign %d (query %q): %v",
					campaignID,
					truncateQuery(query),
					err,
				)
				continue
			}
			allResults = append(allResults, results...)
		}

		ragCtx.CampaignResults = deduplicateAndTrim(allResults)
	}

	// Load the game system schema YAML if a system code was provided.
	if gameSystemCode != "" {
		ragCtx.GameSystemYAML = cb.loadGameSystemSchema(
			gameSystemCode,
		)
	}

	return ragCtx, nil
}

// buildSearchQueries derives multiple search queries from the source
// content and entity list. It returns a content-summary query and
// zero or more entity-name batch queries.
func (cb *ContextBuilder) buildSearchQueries(
	content string,
	entities []models.Entity,
) []string {
	var queries []string

	// 1. Content-summary query: first ~150 chars of content plus any
	//    entity names that appear in the content.
	summaryQuery := buildContentSummaryQuery(content, entities)
	if summaryQuery != "" {
		queries = append(queries, summaryQuery)
	}

	// 2. Entity-name queries: batch entity names into groups for
	//    efficient vector search.
	entityQueries := buildEntityNameQueries(entities)
	queries = append(queries, entityQueries...)

	return queries
}

// buildContentSummaryQuery constructs a search query from the first
// maxContentSummaryLen characters of content, appending entity names
// that appear in the content for better relevance.
func buildContentSummaryQuery(
	content string,
	entities []models.Entity,
) string {
	if content == "" {
		return ""
	}

	// Take the first ~150 characters of content.
	runes := []rune(content)
	limit := maxContentSummaryLen
	if len(runes) < limit {
		limit = len(runes)
	}
	summary := strings.TrimSpace(string(runes[:limit]))

	// Append entity names found in the content for additional
	// relevance signal.
	contentLower := strings.ToLower(content)
	var mentioned []string
	for _, e := range entities {
		if e.Name == "" {
			continue
		}
		if strings.Contains(contentLower, strings.ToLower(e.Name)) {
			mentioned = append(mentioned, e.Name)
		}
	}

	if len(mentioned) > 0 {
		summary = summary + " " + strings.Join(mentioned, ", ")
	}

	return summary
}

// buildEntityNameQueries groups entity names into batches and returns
// one search query per batch. This produces targeted queries that
// find content related to specific entities.
func buildEntityNameQueries(entities []models.Entity) []string {
	if len(entities) == 0 {
		return nil
	}

	var names []string
	for _, e := range entities {
		if e.Name != "" {
			names = append(names, e.Name)
		}
	}

	if len(names) == 0 {
		return nil
	}

	var queries []string
	for i := 0; i < len(names); i += entityBatchSize {
		end := i + entityBatchSize
		if end > len(names) {
			end = len(names)
		}
		batch := strings.Join(names[i:end], ", ")
		queries = append(queries, batch)
	}

	return queries
}

// deduplicateAndTrim removes duplicate search results (by SourceTable
// and SourceID, keeping the highest CombinedScore), sorts by
// CombinedScore descending, and trims to fit within maxContextTokens.
func deduplicateAndTrim(results []models.SearchResult) []models.SearchResult {
	if len(results) == 0 {
		return nil
	}

	// Deduplicate by (SourceTable, SourceID), keeping the entry
	// with the highest CombinedScore.
	type dedupKey struct {
		table string
		id    int64
	}
	best := make(map[dedupKey]models.SearchResult, len(results))

	for _, r := range results {
		key := dedupKey{table: r.SourceTable, id: r.SourceID}
		existing, found := best[key]
		if !found || r.CombinedScore > existing.CombinedScore {
			best[key] = r
		}
	}

	// Collect deduplicated results into a slice.
	deduped := make([]models.SearchResult, 0, len(best))
	for _, r := range best {
		deduped = append(deduped, r)
	}

	// Sort by CombinedScore descending (highest relevance first).
	sort.Slice(deduped, func(i, j int) bool {
		return deduped[i].CombinedScore > deduped[j].CombinedScore
	})

	// Trim to fit within the token budget.
	var trimmed []models.SearchResult
	var totalTokens float64

	for _, r := range deduped {
		tokens := estimateTokens(r.ChunkContent)
		if totalTokens+tokens > maxContextTokens && len(trimmed) > 0 {
			break
		}
		trimmed = append(trimmed, r)
		totalTokens += tokens
	}

	return trimmed
}

// estimateTokens returns a rough token count for a string based on
// its character length. This uses a simple heuristic suitable for
// budget tracking where precision is not critical.
func estimateTokens(s string) float64 {
	return float64(len([]rune(s))) * tokensPerChar
}

// loadGameSystemSchema reads the YAML schema file for the given game
// system code. It returns the file contents as a string, or an empty
// string if the file does not exist or cannot be read. The code is
// validated to contain only alphanumeric characters and hyphens to
// prevent path traversal attacks.
func (cb *ContextBuilder) loadGameSystemSchema(code string) string {
	for _, r := range code {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-') {
			log.Printf("enrichment: invalid game system code %q", code)
			return ""
		}
	}
	path := filepath.Join(cb.schemasDir, code+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf(
			"enrichment: failed to load game system schema %q: %v",
			path, err,
		)
		return ""
	}
	return string(data)
}

// truncateQuery returns the first maxSearchQueryLen characters of s.
// If s is shorter than the limit it is returned unchanged.
func truncateQuery(s string) string {
	runes := []rune(s)
	if len(runes) <= maxSearchQueryLen {
		return s
	}
	return string(runes[:maxSearchQueryLen])
}

// String returns a human-readable summary of the ContextBuilder
// configuration for logging and debugging.
func (cb *ContextBuilder) String() string {
	return fmt.Sprintf(
		"ContextBuilder(schemasDir=%q, maxTokens=%d)",
		cb.schemasDir, maxContextTokens,
	)
}
