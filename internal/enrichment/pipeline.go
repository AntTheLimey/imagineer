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

	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/antonypegg/imagineer/internal/models"
)

// RAGContext holds retrieved context shared across all pipeline agents.
type RAGContext struct {
	CampaignResults []models.SearchResult
	GameSystemYAML  string
}

// PipelineInput contains everything needed for a pipeline run.
type PipelineInput struct {
	CampaignID    int64
	JobID         int64
	SourceTable   string
	SourceID      int64
	SourceField   string
	Content       string
	Entities      []models.Entity
	Relationships []models.Relationship
	GameSystemID  *int64
	Context       *RAGContext

	// PriorResults holds items produced by agents that ran in
	// earlier pipeline stages. The pipeline populates this field
	// before each stage so that downstream agents can inspect
	// upstream output without a database round-trip.
	PriorResults []models.ContentAnalysisItem
}

// PipelineAgent extends the base agents.Agent interface with pipeline-
// specific methods.
type PipelineAgent interface {
	// Name returns the unique identifier for this agent.
	Name() string
	// Run executes the agent and returns items for triage.
	Run(ctx context.Context, provider llm.Provider, input PipelineInput) ([]models.ContentAnalysisItem, error)
	// DependsOn returns names of agents that must run before this one.
	DependsOn() []string
}

// Stage groups agents that run at the same pipeline phase.
type Stage struct {
	Name   string
	Phase  string // "analysis" or "enrichment"
	Agents []PipelineAgent
}

// Pipeline orchestrates multi-stage content analysis.
type Pipeline struct {
	db     *database.DB
	stages []Stage
}

// NewPipeline creates a new Pipeline with the given database handle and
// ordered stages. Stages are executed sequentially; agents within each
// stage are sorted by their declared dependencies.
func NewPipeline(db *database.DB, stages []Stage) *Pipeline {
	return &Pipeline{
		db:     db,
		stages: stages,
	}
}

// Run executes every stage in order, running each stage's agents in
// dependency order. Items produced by all agents across all stages are
// collected and returned. If an individual agent fails, the error is
// logged and the pipeline continues (graceful degradation).
func (p *Pipeline) Run(
	ctx context.Context,
	provider llm.Provider,
	input PipelineInput,
) ([]models.ContentAnalysisItem, error) {
	var allItems []models.ContentAnalysisItem

	for _, stage := range p.stages {
		// Make items from prior stages available to agents in
		// this stage via the PriorResults field.
		input.PriorResults = allItems

		sorted, skipped := topologicalSort(stage.Agents)

		for _, name := range skipped {
			log.Printf(
				"pipeline: skipping agent %q in stage %q due to circular dependency",
				name, stage.Name,
			)
		}

		for _, agent := range sorted {
			items, err := agent.Run(ctx, provider, input)
			if err != nil {
				log.Printf(
					"pipeline: agent %q in stage %q failed: %v",
					agent.Name(), stage.Name, err,
				)
				continue
			}

			// Tag every item with the agent that produced it.
			for i := range items {
				items[i].AgentName = agent.Name()
			}

			allItems = append(allItems, items...)
		}
	}

	return allItems, nil
}

// topologicalSort orders agents so that each agent runs after all of
// its declared dependencies. Agents with no dependencies come first.
// If a circular dependency is detected, the offending agents are
// returned in the skipped slice and excluded from the sorted output.
func topologicalSort(agents []PipelineAgent) (sorted []PipelineAgent, skipped []string) {
	// Build lookup maps.
	byName := make(map[string]PipelineAgent, len(agents))
	for _, a := range agents {
		byName[a.Name()] = a
	}

	// Track visit state for cycle detection.
	const (
		unvisited = 0
		visiting  = 1
		visited   = 2
	)
	state := make(map[string]int, len(agents))
	for _, a := range agents {
		state[a.Name()] = unvisited
	}

	// Collect names that are part of a cycle.
	cyclic := make(map[string]bool)

	// Recursive DFS visit. Returns false if a cycle is detected.
	var order []PipelineAgent
	var visit func(name string) bool
	visit = func(name string) bool {
		switch state[name] {
		case visited:
			return true
		case visiting:
			// Cycle detected.
			return false
		}

		state[name] = visiting

		agent, exists := byName[name]
		if !exists {
			// Dependency references an agent not in this stage;
			// treat as satisfied (it may have run in a prior stage).
			state[name] = visited
			return true
		}

		for _, dep := range agent.DependsOn() {
			if !visit(dep) {
				cyclic[name] = true
				return false
			}
		}

		state[name] = visited
		order = append(order, agent)
		return true
	}

	for _, a := range agents {
		if state[a.Name()] == unvisited {
			if !visit(a.Name()) {
				cyclic[a.Name()] = true
				// Reset visiting nodes so we can continue with
				// remaining agents.
				for n, s := range state {
					if s == visiting {
						cyclic[n] = true
						state[n] = unvisited
					}
				}
			}
		}
	}

	// Build the skipped list from cyclic agents.
	for _, a := range agents {
		if cyclic[a.Name()] {
			skipped = append(skipped, a.Name())
		}
	}

	// Filter out any cyclic agents that may have leaked into order.
	sorted = make([]PipelineAgent, 0, len(order))
	for _, a := range order {
		if !cyclic[a.Name()] {
			sorted = append(sorted, a)
		}
	}

	return sorted, skipped
}

// String returns a human-readable description of the pipeline for
// debugging and logging.
func (p *Pipeline) String() string {
	total := 0
	for _, s := range p.stages {
		total += len(s.Agents)
	}
	return fmt.Sprintf(
		"Pipeline(%d stages, %d agents)",
		len(p.stages), total,
	)
}
