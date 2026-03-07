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

	"github.com/antonypegg/imagineer/internal/analysis"
	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/antonypegg/imagineer/internal/models"
)

// IdentificationAgent wraps the content analysis scanner as a
// PipelineAgent. It detects wiki links, untagged entity mentions,
// and potential misspellings.
type IdentificationAgent struct {
	analyzer *analysis.Analyzer
}

// NewIdentificationAgent creates an IdentificationAgent backed by
// the given database handle.
func NewIdentificationAgent(db *database.DB) *IdentificationAgent {
	return &IdentificationAgent{
		analyzer: analysis.NewAnalyzer(db),
	}
}

// Name returns the unique identifier for this pipeline agent.
func (a *IdentificationAgent) Name() string {
	return "identification"
}

// DependsOn returns the names of agents that must run before this one.
// The identification agent has no dependencies and can run independently.
func (a *IdentificationAgent) DependsOn() []string {
	return nil
}

// Run executes identification scans on the input content. The
// provider argument is unused (no LLM calls needed).
func (a *IdentificationAgent) Run(
	ctx context.Context,
	_ llm.Provider,
	input PipelineInput,
) ([]models.ContentAnalysisItem, error) {
	return a.analyzer.ScanContent(
		ctx, input.CampaignID, input.Content)
}
