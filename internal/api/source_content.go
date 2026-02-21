/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package api

import (
	"context"
	"fmt"

	"github.com/antonypegg/imagineer/internal/agents/canon"
	"github.com/antonypegg/imagineer/internal/agents/graph"
	"github.com/antonypegg/imagineer/internal/agents/ttrpg"
	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/enrichment"
)

// fetchSourceContent retrieves the text content from the appropriate
// source table and field using parameterized queries. All queries are
// scoped by campaignID to prevent cross-campaign data access (IDOR).
func fetchSourceContent(
	ctx context.Context,
	db *database.DB,
	campaignID int64,
	sourceTable string,
	sourceID int64,
	sourceField string,
) (string, error) {
	var query string

	switch sourceTable + "." + sourceField {
	case "entities.description":
		query = "SELECT COALESCE(description, '') FROM entities WHERE id = $1 AND campaign_id = $2"
	case "entities.gm_notes":
		query = "SELECT COALESCE(gm_notes, '') FROM entities WHERE id = $1 AND campaign_id = $2"
	case "chapters.overview":
		query = "SELECT COALESCE(overview, '') FROM chapters WHERE id = $1 AND campaign_id = $2"
	case "sessions.prep_notes":
		query = "SELECT COALESCE(s.prep_notes, '') FROM sessions s JOIN chapters c ON s.chapter_id = c.id WHERE s.id = $1 AND c.campaign_id = $2"
	case "sessions.actual_notes":
		query = "SELECT COALESCE(s.actual_notes, '') FROM sessions s JOIN chapters c ON s.chapter_id = c.id WHERE s.id = $1 AND c.campaign_id = $2"
	case "campaigns.description":
		query = "SELECT COALESCE(description, '') FROM campaigns WHERE id = $1 AND id = $2"
	default:
		return "", fmt.Errorf("unsupported source: %s.%s", sourceTable, sourceField)
	}

	var content string
	err := db.QueryRow(ctx, query, sourceID, campaignID).Scan(&content)
	if err != nil {
		return "", fmt.Errorf("failed to fetch content from %s.%s: %w",
			sourceTable, sourceField, err)
	}

	return content, nil
}

// buildDefaultPipeline creates a Pipeline with the standard two-stage
// layout used for content enrichment: an analysis stage (TTRPG expert
// + canon expert) followed by an enrichment stage (enrichment agent +
// graph expert).
func buildDefaultPipeline(db *database.DB) *enrichment.Pipeline {
	ttrpgAgent := ttrpg.NewExpert()
	canonAgent := canon.NewExpert()
	enrichAgent := enrichment.NewEnrichmentAgent(db)
	graphAgent := graph.NewExpert(db)

	return enrichment.NewPipeline(db, []enrichment.Stage{
		{
			Name:   "analysis",
			Phase:  "analysis",
			Agents: []enrichment.PipelineAgent{ttrpgAgent, canonAgent},
		},
		{
			Name:   "enrichment",
			Phase:  "enrichment",
			Agents: []enrichment.PipelineAgent{enrichAgent, graphAgent},
		},
	})
}
