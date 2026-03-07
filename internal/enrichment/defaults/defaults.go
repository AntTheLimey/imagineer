/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package defaults

import (
	"github.com/antonypegg/imagineer/internal/agents/canon"
	"github.com/antonypegg/imagineer/internal/agents/graph"
	"github.com/antonypegg/imagineer/internal/agents/ttrpg"
	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/enrichment"
)

// NewDefaultRegistry creates a PhaseRegistry with the standard three
// phases registered: identify (wiki link and entity scanning), revise
// (TTRPG and canon analysis), and enrich (entity enrichment and graph
// validation).
func NewDefaultRegistry(db *database.DB) *enrichment.PhaseRegistry {
	reg := enrichment.NewPhaseRegistry(db)
	reg.Register("identify", "identification",
		enrichment.NewIdentificationAgent(db))
	reg.Register("revise", "analysis",
		ttrpg.NewExpert(), canon.NewExpert())
	reg.Register("enrich", "enrichment",
		enrichment.NewEnrichmentAgent(db), graph.NewExpert(db))
	return reg
}
