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
	"log"

	"github.com/antonypegg/imagineer/internal/database"
)

// phaseEntry holds the metadata for a registered pipeline phase.
type phaseEntry struct {
	phaseTag string
	agents   []PipelineAgent
}

// PhaseRegistry maintains a named collection of pipeline phases that
// can be assembled into a Pipeline via BuildPipeline.
type PhaseRegistry struct {
	db      *database.DB
	entries map[string]phaseEntry
}

// NewPhaseRegistry creates an empty PhaseRegistry.
func NewPhaseRegistry(db *database.DB) *PhaseRegistry {
	return &PhaseRegistry{
		db:      db,
		entries: make(map[string]phaseEntry),
	}
}

// Register adds a named phase with the given phaseTag and agents to
// the registry. If a phase with the same name already exists, it is
// overwritten.
func (r *PhaseRegistry) Register(name, phaseTag string, agents ...PipelineAgent) {
	r.entries[name] = phaseEntry{
		phaseTag: phaseTag,
		agents:   agents,
	}
}

// BuildPipeline creates a Pipeline containing one Stage per requested
// phase name. Unknown phases are logged and skipped. Repeated phase
// names produce repeated stages.
func (r *PhaseRegistry) BuildPipeline(phases []string) *Pipeline {
	var stages []Stage
	for _, name := range phases {
		entry, ok := r.entries[name]
		if !ok {
			log.Printf("PhaseRegistry: unknown phase %q, skipping", name)
			continue
		}
		stages = append(stages, Stage{
			Name:   name,
			Phase:  entry.phaseTag,
			Agents: entry.agents,
		})
	}
	return NewPipeline(r.db, stages)
}
