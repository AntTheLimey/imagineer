/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package consistency

import (
	"context"
	"testing"

	"github.com/antonypegg/imagineer/internal/agents"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestChecker_Name verifies the agent name is correct.
func TestChecker_Name(t *testing.T) {
	checker := &Checker{}
	assert.Equal(t, "consistency-checker", checker.Name())
}

// TestChecker_Description verifies the agent description is non-empty.
func TestChecker_Description(t *testing.T) {
	checker := &Checker{}
	desc := checker.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, desc, "inconsistencies")
}

// TestParseCampaignID_ValidInt64 tests parsing a valid int64 parameter.
func TestParseCampaignID_ValidInt64(t *testing.T) {
	expected := int64(42)
	params := map[string]any{
		"campaign_id": expected,
	}

	result, err := parseCampaignID(params)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

// TestParseCampaignID_ValidString tests parsing a valid int64 string parameter.
func TestParseCampaignID_ValidString(t *testing.T) {
	expected := int64(42)
	params := map[string]any{
		"campaign_id": "42",
	}

	result, err := parseCampaignID(params)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

// TestParseCampaignID_Missing tests error when campaign_id is missing.
func TestParseCampaignID_Missing(t *testing.T) {
	params := map[string]any{}

	_, err := parseCampaignID(params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "campaign_id parameter is required")
}

// TestParseCampaignID_InvalidString tests error when campaign_id is a non-numeric string.
func TestParseCampaignID_InvalidString(t *testing.T) {
	params := map[string]any{
		"campaign_id": "not-a-number",
	}

	_, err := parseCampaignID(params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid campaign_id")
}

// TestParseCampaignID_InvalidType tests error when campaign_id is wrong type.
func TestParseCampaignID_InvalidType(t *testing.T) {
	params := map[string]any{
		"campaign_id": true,
	}

	_, err := parseCampaignID(params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be a number or string")
}

// TestSummarizeIssues_Empty tests summarizing an empty issue list.
func TestSummarizeIssues_Empty(t *testing.T) {
	issues := []ConsistencyIssue{}
	summary := summarizeIssues(issues)

	assert.Equal(t, 0, summary.Total)
	assert.Equal(t, 0, summary.Critical)
	assert.Equal(t, 0, summary.Major)
	assert.Equal(t, 0, summary.Minor)
}

// TestSummarizeIssues_Mixed tests summarizing issues with mixed severities.
func TestSummarizeIssues_Mixed(t *testing.T) {
	issues := []ConsistencyIssue{
		{Type: "test1", Severity: agents.SeverityCritical},
		{Type: "test2", Severity: agents.SeverityCritical},
		{Type: "test3", Severity: agents.SeverityMajor},
		{Type: "test4", Severity: agents.SeverityMajor},
		{Type: "test5", Severity: agents.SeverityMajor},
		{Type: "test6", Severity: agents.SeverityMinor},
	}
	summary := summarizeIssues(issues)

	assert.Equal(t, 6, summary.Total)
	assert.Equal(t, 2, summary.Critical)
	assert.Equal(t, 3, summary.Major)
	assert.Equal(t, 1, summary.Minor)
}

// TestConsistencyIssue_Structure tests the ConsistencyIssue struct fields.
func TestConsistencyIssue_Structure(t *testing.T) {
	entityID := int64(100)
	relatedID := int64(200)

	issue := ConsistencyIssue{
		Type:        "orphaned_entity",
		Severity:    agents.SeverityMinor,
		EntityID:    &entityID,
		EntityName:  "Test Entity",
		Description: "Entity has no relationships",
		Suggestion:  "Consider linking or removing",
		RelatedIDs:  []int64{relatedID},
	}

	assert.Equal(t, "orphaned_entity", issue.Type)
	assert.Equal(t, agents.SeverityMinor, issue.Severity)
	assert.Equal(t, entityID, *issue.EntityID)
	assert.Equal(t, "Test Entity", issue.EntityName)
	assert.Equal(t, "Entity has no relationships", issue.Description)
	assert.Equal(t, "Consider linking or removing", issue.Suggestion)
	assert.Len(t, issue.RelatedIDs, 1)
	assert.Equal(t, relatedID, issue.RelatedIDs[0])
}

// TestCheckerResult_Structure tests the CheckerResult struct fields.
func TestCheckerResult_Structure(t *testing.T) {
	campaignID := int64(1)
	entityID := int64(2)

	result := CheckerResult{
		CampaignID: campaignID,
		Issues: []ConsistencyIssue{
			{
				Type:       "test",
				Severity:   agents.SeverityMajor,
				EntityID:   &entityID,
				EntityName: "Test",
			},
		},
		Summary: IssueSummary{
			Total: 1,
			Major: 1,
		},
	}

	assert.Equal(t, campaignID, result.CampaignID)
	assert.Len(t, result.Issues, 1)
	assert.Equal(t, 1, result.Summary.Total)
	assert.Equal(t, 1, result.Summary.Major)
}

// TestChecker_Run_MissingCampaignID tests Run with missing campaign_id.
func TestChecker_Run_MissingCampaignID(t *testing.T) {
	checker := &Checker{db: nil}
	params := map[string]any{}

	result, err := checker.Run(context.Background(), params)
	require.NoError(t, err) // Errors are returned in Result, not as error
	assert.False(t, result.Success)
	assert.NotEmpty(t, result.Errors)
	assert.Contains(t, result.Errors[0], "campaign_id parameter is required")
}

// TestChecker_Run_InvalidCampaignID tests Run with invalid campaign_id.
func TestChecker_Run_InvalidCampaignID(t *testing.T) {
	checker := &Checker{db: nil}
	params := map[string]any{
		"campaign_id": "invalid",
	}

	result, err := checker.Run(context.Background(), params)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.NotEmpty(t, result.Errors)
	assert.Contains(t, result.Errors[0], "invalid campaign_id")
}

// TestIssueSummary_Structure tests the IssueSummary struct.
func TestIssueSummary_Structure(t *testing.T) {
	summary := IssueSummary{
		Total:    10,
		Critical: 2,
		Major:    3,
		Minor:    5,
	}

	assert.Equal(t, 10, summary.Total)
	assert.Equal(t, 2, summary.Critical)
	assert.Equal(t, 3, summary.Major)
	assert.Equal(t, 5, summary.Minor)
}

// TestSeverityValues tests that severity constants have expected values.
func TestSeverityValues(t *testing.T) {
	assert.Equal(t, agents.Severity("minor"), agents.SeverityMinor)
	assert.Equal(t, agents.Severity("major"), agents.SeverityMajor)
	assert.Equal(t, agents.Severity("critical"), agents.SeverityCritical)
}
