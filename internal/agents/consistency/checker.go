/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

// Package consistency provides the consistency checker agent for finding
// plot holes, timeline conflicts, and orphaned entities in campaigns.
package consistency

import (
	"context"
	"fmt"

	"github.com/antonypegg/imagineer/internal/agents"
	"github.com/antonypegg/imagineer/internal/database"
	"github.com/google/uuid"
)

// ConsistencyIssue represents a detected inconsistency in campaign data.
type ConsistencyIssue struct {
	// Type categorizes the inconsistency.
	Type string `json:"type"`

	// Severity indicates the importance of this issue.
	Severity agents.Severity `json:"severity"`

	// EntityID references the primary entity involved.
	EntityID *uuid.UUID `json:"entityId,omitempty"`

	// EntityName is the name of the primary entity involved.
	EntityName string `json:"entityName,omitempty"`

	// Description provides details about the inconsistency.
	Description string `json:"description"`

	// Suggestion recommends how to resolve the issue.
	Suggestion string `json:"suggestion"`

	// RelatedIDs contains IDs of other entities related to this issue.
	RelatedIDs []uuid.UUID `json:"relatedIds,omitempty"`
}

// CheckerResult contains the output of a consistency check.
type CheckerResult struct {
	// CampaignID identifies the campaign that was checked.
	CampaignID uuid.UUID `json:"campaignId"`

	// Issues contains all detected inconsistencies.
	Issues []ConsistencyIssue `json:"issues"`

	// Summary provides aggregate counts by severity.
	Summary IssueSummary `json:"summary"`
}

// IssueSummary provides counts of issues by severity level.
type IssueSummary struct {
	Total    int `json:"total"`
	Critical int `json:"critical"`
	Major    int `json:"major"`
	Minor    int `json:"minor"`
}

// Checker implements the consistency checking agent.
type Checker struct {
	db *database.DB
}

// New creates a new consistency checker agent.
func New(db *database.DB) *Checker {
	return &Checker{db: db}
}

// Name returns the unique identifier for this agent.
func (c *Checker) Name() string {
	return "consistency-checker"
}

// Description returns a human-readable description of what this agent does.
func (c *Checker) Description() string {
	return "Analyzes campaign data to find plot holes, timeline conflicts, orphaned entities, and other inconsistencies"
}

// Run executes the consistency checker with the given parameters.
// Required params: "campaign_id" (uuid.UUID or string)
// Optional params: "entity_type" (string) to filter checks to specific entity types
func (c *Checker) Run(ctx context.Context, params map[string]any) (agents.Result, error) {
	// Parse campaign_id parameter
	campaignID, err := parseCampaignID(params)
	if err != nil {
		return agents.Result{
			Success: false,
			Errors:  []string{err.Error()},
		}, nil
	}

	// Parse optional entity_type filter
	var entityTypeFilter *string
	if et, ok := params["entity_type"].(string); ok && et != "" {
		entityTypeFilter = &et
	}

	// Run all consistency checks
	var issues []ConsistencyIssue

	// Check for orphaned entities
	orphanedIssues, err := c.checkOrphanedEntities(ctx, campaignID, entityTypeFilter)
	if err != nil {
		return agents.Result{
			Success: false,
			Errors:  []string{fmt.Sprintf("failed to check orphaned entities: %v", err)},
		}, nil
	}
	issues = append(issues, orphanedIssues...)

	// Check for duplicate names
	duplicateIssues, err := c.checkDuplicateNames(ctx, campaignID)
	if err != nil {
		return agents.Result{
			Success: false,
			Errors:  []string{fmt.Sprintf("failed to check duplicate names: %v", err)},
		}, nil
	}
	issues = append(issues, duplicateIssues...)

	// Check for timeline conflicts
	timelineIssues, err := c.checkTimelineConflicts(ctx, campaignID)
	if err != nil {
		return agents.Result{
			Success: false,
			Errors:  []string{fmt.Sprintf("failed to check timeline conflicts: %v", err)},
		}, nil
	}
	issues = append(issues, timelineIssues...)

	// Check for invalid references
	referenceIssues, err := c.checkInvalidReferences(ctx, campaignID)
	if err != nil {
		return agents.Result{
			Success: false,
			Errors:  []string{fmt.Sprintf("failed to check invalid references: %v", err)},
		}, nil
	}
	issues = append(issues, referenceIssues...)

	// Check for sessions without discoveries
	sessionIssues, err := c.checkSessionsWithoutDiscoveries(ctx, campaignID)
	if err != nil {
		return agents.Result{
			Success: false,
			Errors:  []string{fmt.Sprintf("failed to check sessions without discoveries: %v", err)},
		}, nil
	}
	issues = append(issues, sessionIssues...)

	// Build result
	result := CheckerResult{
		CampaignID: campaignID,
		Issues:     issues,
		Summary:    summarizeIssues(issues),
	}

	// Convert issues to suggestions for the standard result format
	suggestions := make([]agents.Suggestion, len(issues))
	for i, issue := range issues {
		suggestions[i] = agents.Suggestion{
			Type:        issue.Type,
			Severity:    issue.Severity,
			Description: issue.Description,
			Action:      issue.Suggestion,
			EntityID:    issue.EntityID,
			RelatedIDs:  issue.RelatedIDs,
		}
	}

	return agents.Result{
		Success:     true,
		Data:        result,
		Suggestions: suggestions,
	}, nil
}

// parseCampaignID extracts and validates the campaign_id parameter.
func parseCampaignID(params map[string]any) (uuid.UUID, error) {
	raw, ok := params["campaign_id"]
	if !ok {
		return uuid.Nil, fmt.Errorf("campaign_id parameter is required")
	}

	switch v := raw.(type) {
	case uuid.UUID:
		return v, nil
	case string:
		id, err := uuid.Parse(v)
		if err != nil {
			return uuid.Nil, fmt.Errorf("invalid campaign_id: %w", err)
		}
		return id, nil
	default:
		return uuid.Nil, fmt.Errorf("campaign_id must be a UUID or string")
	}
}

// checkOrphanedEntities finds entities with no relationships or timeline references.
func (c *Checker) checkOrphanedEntities(ctx context.Context, campaignID uuid.UUID, entityTypeFilter *string) ([]ConsistencyIssue, error) {
	orphans, err := c.db.FindOrphanedEntities(ctx, campaignID, entityTypeFilter)
	if err != nil {
		return nil, err
	}

	issues := make([]ConsistencyIssue, len(orphans))
	for i, orphan := range orphans {
		issues[i] = ConsistencyIssue{
			Type:        "orphaned_entity",
			Severity:    agents.SeverityMinor,
			EntityID:    &orphan.ID,
			EntityName:  orphan.Name,
			Description: fmt.Sprintf("Entity '%s' (%s) has no relationships or timeline references", orphan.Name, orphan.EntityType),
			Suggestion:  "Consider linking this entity to others or removing if unused",
		}
	}

	return issues, nil
}

// checkDuplicateNames finds entities with very similar names.
func (c *Checker) checkDuplicateNames(ctx context.Context, campaignID uuid.UUID) ([]ConsistencyIssue, error) {
	duplicates, err := c.db.FindDuplicateNames(ctx, campaignID, 0.7)
	if err != nil {
		return nil, err
	}

	var issues []ConsistencyIssue
	for _, dup := range duplicates {
		issue := ConsistencyIssue{
			Type:        "duplicate_name",
			Severity:    agents.SeverityMajor,
			EntityID:    &dup.EntityID1,
			EntityName:  dup.Name1,
			Description: fmt.Sprintf("Entities '%s' and '%s' have very similar names (%.0f%% similarity)", dup.Name1, dup.Name2, dup.Similarity*100),
			Suggestion:  "These entities may be duplicates. Consider merging.",
			RelatedIDs:  []uuid.UUID{dup.EntityID2},
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

// checkTimelineConflicts finds entities appearing in multiple events at the same time.
func (c *Checker) checkTimelineConflicts(ctx context.Context, campaignID uuid.UUID) ([]ConsistencyIssue, error) {
	conflicts, err := c.db.FindTimelineConflicts(ctx, campaignID)
	if err != nil {
		return nil, err
	}

	var issues []ConsistencyIssue
	for _, conflict := range conflicts {
		issue := ConsistencyIssue{
			Type:        "timeline_conflict",
			Severity:    agents.SeverityCritical,
			EntityID:    &conflict.EntityID,
			EntityName:  conflict.EntityName,
			Description: fmt.Sprintf("Entity '%s' appears in %d events on %s", conflict.EntityName, conflict.EventCount, conflict.EventDate.Format("2006-01-02")),
			Suggestion:  "Entity cannot be in two places at once. Verify timeline.",
			RelatedIDs:  conflict.EventIDs,
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

// checkInvalidReferences finds relationships or events referencing non-existent entities.
func (c *Checker) checkInvalidReferences(ctx context.Context, campaignID uuid.UUID) ([]ConsistencyIssue, error) {
	invalidRefs, err := c.db.FindMissingRelationshipTargets(ctx, campaignID)
	if err != nil {
		return nil, err
	}

	var issues []ConsistencyIssue
	for _, ref := range invalidRefs {
		issue := ConsistencyIssue{
			Type:        "invalid_reference",
			Severity:    agents.SeverityCritical,
			EntityID:    &ref.RelationshipID,
			Description: fmt.Sprintf("Relationship references missing entity (ID: %s)", ref.MissingEntityID),
			Suggestion:  "Reference points to missing entity. Update or remove.",
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

// checkSessionsWithoutDiscoveries finds completed sessions with no discoveries.
func (c *Checker) checkSessionsWithoutDiscoveries(ctx context.Context, campaignID uuid.UUID) ([]ConsistencyIssue, error) {
	sessions, err := c.db.FindSessionsWithoutEntities(ctx, campaignID)
	if err != nil {
		return nil, err
	}

	var issues []ConsistencyIssue
	for _, session := range sessions {
		sessionID := session.ID
		issue := ConsistencyIssue{
			Type:        "session_without_discoveries",
			Severity:    agents.SeverityMinor,
			EntityID:    &sessionID,
			EntityName:  fmt.Sprintf("Session %d", session.SessionNumber),
			Description: fmt.Sprintf("Completed session %d has no entity discoveries or references", session.SessionNumber),
			Suggestion:  "Consider adding discovered entities to this session.",
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

// summarizeIssues calculates aggregate counts by severity.
func summarizeIssues(issues []ConsistencyIssue) IssueSummary {
	summary := IssueSummary{Total: len(issues)}
	for _, issue := range issues {
		switch issue.Severity {
		case agents.SeverityCritical:
			summary.Critical++
		case agents.SeverityMajor:
			summary.Major++
		case agents.SeverityMinor:
			summary.Minor++
		}
	}
	return summary
}
