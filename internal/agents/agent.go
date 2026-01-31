/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

// Package agents provides the agent infrastructure for campaign analysis.
package agents

import (
	"context"

	"github.com/google/uuid"
)

// Severity represents the importance level of an issue or suggestion.
type Severity string

const (
	// SeverityMinor indicates a low-priority issue that may warrant attention.
	SeverityMinor Severity = "minor"

	// SeverityMajor indicates a significant issue that should be addressed.
	SeverityMajor Severity = "major"

	// SeverityCritical indicates a high-priority issue requiring immediate attention.
	SeverityCritical Severity = "critical"
)

// Agent defines the interface that all analysis agents must implement.
type Agent interface {
	// Name returns the unique identifier for this agent.
	Name() string

	// Description returns a human-readable description of what this agent does.
	Description() string

	// Run executes the agent with the given parameters and returns the result.
	Run(ctx context.Context, params map[string]any) (Result, error)
}

// Result represents the output of an agent execution.
type Result struct {
	// Success indicates whether the agent completed successfully.
	Success bool `json:"success"`

	// Data contains the agent-specific output data.
	Data any `json:"data,omitempty"`

	// Suggestions contains actionable recommendations from the agent.
	Suggestions []Suggestion `json:"suggestions,omitempty"`

	// Errors contains any error messages encountered during execution.
	Errors []string `json:"errors,omitempty"`

	// Sources provides attribution for RAG-retrieved content.
	Sources []Source `json:"sources,omitempty"`
}

// Suggestion represents an actionable recommendation from an agent.
type Suggestion struct {
	// Type categorizes the suggestion.
	Type string `json:"type"`

	// Severity indicates the importance of this suggestion.
	Severity Severity `json:"severity"`

	// Description provides details about the suggestion.
	Description string `json:"description"`

	// Action describes what should be done to address this suggestion.
	Action string `json:"action"`

	// EntityID references the entity this suggestion relates to, if any.
	EntityID *uuid.UUID `json:"entityId,omitempty"`

	// RelatedIDs contains IDs of other entities related to this suggestion.
	RelatedIDs []uuid.UUID `json:"relatedIds,omitempty"`
}

// Source represents the origin of retrieved content for attribution.
type Source struct {
	// Type indicates the source type: "campaign" or "rulebook".
	Type string `json:"type"`

	// EntityID references the entity this content came from, if applicable.
	EntityID *uuid.UUID `json:"entityId,omitempty"`

	// EntityName is the human-readable name of the source entity.
	EntityName string `json:"entityName,omitempty"`

	// ChunkText contains the relevant text excerpt from the source.
	ChunkText string `json:"chunkText,omitempty"`

	// Score indicates the relevance score for RAG-retrieved content.
	Score float64 `json:"score,omitempty"`
}
