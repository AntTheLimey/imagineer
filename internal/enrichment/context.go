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
	"log"
	"os"
	"path/filepath"

	"github.com/antonypegg/imagineer/internal/database"
)

// maxSearchQueryLen is the maximum number of characters from the source
// content used as the vector search query. Longer content is truncated
// to this length before being sent to SearchCampaignContent.
const maxSearchQueryLen = 200

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

// BuildContext assembles a RAGContext by querying campaign content via
// vector search and loading the game system schema YAML. Both sources
// are optional: if vectorization is unavailable or the schema file
// cannot be read, the corresponding field is left empty and no error
// is returned.
func (cb *ContextBuilder) BuildContext(
	ctx context.Context,
	campaignID int64,
	content string,
	gameSystemCode string,
) (*RAGContext, error) {
	ragCtx := &RAGContext{}

	// Retrieve relevant campaign content via hybrid vector search.
	if cb.db.IsVectorizationAvailable(ctx) {
		query := truncateQuery(content)
		if query != "" {
			results, err := cb.db.SearchCampaignContent(
				ctx, campaignID, query, 20,
			)
			if err != nil {
				log.Printf(
					"enrichment: vector search failed for "+
						"campaign %d: %v",
					campaignID, err,
				)
			} else {
				ragCtx.CampaignResults = results
			}
		}
	}

	// Load the game system schema YAML if a system code was provided.
	if gameSystemCode != "" {
		ragCtx.GameSystemYAML = cb.loadGameSystemSchema(
			gameSystemCode,
		)
	}

	return ragCtx, nil
}

// loadGameSystemSchema reads the YAML schema file for the given game
// system code. It returns the file contents as a string, or an empty
// string if the file does not exist or cannot be read.
func (cb *ContextBuilder) loadGameSystemSchema(code string) string {
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
