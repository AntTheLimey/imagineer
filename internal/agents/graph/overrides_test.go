/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package graph

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/antonypegg/imagineer/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// FilterOverriddenFindings tests
// ---------------------------------------------------------------------------

func TestFilterOverriddenFindings_NilDB(t *testing.T) {
	items := []models.ContentAnalysisItem{
		{DetectionType: "invalid_type_pair", MatchedText: "employs"},
		{DetectionType: "orphan_warning", MatchedText: "Viktor"},
	}

	result := FilterOverriddenFindings(
		context.TODO(), nil, 1, items,
	)

	assert.Equal(t, items, result,
		"nil DB should return items unchanged")
}

func TestFilterOverriddenFindings_EmptyItems(t *testing.T) {
	result := FilterOverriddenFindings(
		context.TODO(), nil, 1, []models.ContentAnalysisItem{},
	)

	assert.Empty(t, result,
		"empty items should return empty slice")
}

func TestFilterOverriddenFindings_NilItems(t *testing.T) {
	result := FilterOverriddenFindings(
		context.TODO(), nil, 1, nil,
	)

	assert.Nil(t, result,
		"nil items should return nil")
}

// ---------------------------------------------------------------------------
// extractOverrideKey tests
// ---------------------------------------------------------------------------

func TestExtractOverrideKey_NonOverridableTypes(t *testing.T) {
	nonOverridable := []string{
		"orphan_warning",
		"redundant_edge",
		"graph_warning",
		"relationship_suggestion",
		"unknown_type",
		"",
	}

	for _, dt := range nonOverridable {
		t.Run(dt, func(t *testing.T) {
			item := models.ContentAnalysisItem{
				DetectionType: dt,
			}
			_, ok := extractOverrideKey(item)
			assert.False(t, ok,
				"detection type %q should not be overridable", dt)
		})
	}
}

func TestExtractOverrideKey_InvalidTypePair(t *testing.T) {
	content, err := json.Marshal(map[string]interface{}{
		"relationshipType": "employs",
		"sourceEntityType": "location",
		"targetEntityType": "event",
		"description":      "Not valid between these types.",
	})
	require.NoError(t, err)

	item := models.ContentAnalysisItem{
		DetectionType:    "invalid_type_pair",
		SuggestedContent: json.RawMessage(content),
	}

	info, ok := extractOverrideKey(item)

	require.True(t, ok)
	assert.Equal(t, "domain_range", info.constraintType)
	assert.Equal(t, "employs:location:event", info.overrideKey)
}

func TestExtractOverrideKey_CardinalityViolation(t *testing.T) {
	content, err := json.Marshal(map[string]interface{}{
		"relationshipType": "allied_with",
		"entityId":         42,
		"direction":        "source",
		"description":      "Too many.",
	})
	require.NoError(t, err)

	item := models.ContentAnalysisItem{
		DetectionType:    "cardinality_violation",
		SuggestedContent: json.RawMessage(content),
	}

	info, ok := extractOverrideKey(item)

	require.True(t, ok)
	assert.Equal(t, "cardinality", info.constraintType)
	assert.Equal(t, "allied_with:42:source", info.overrideKey)
}

func TestExtractOverrideKey_MissingRequired(t *testing.T) {
	content, err := json.Marshal(map[string]interface{}{
		"entityType":              "npc",
		"missingRelationshipType": "located_at",
		"description":             "Missing required relationship.",
	})
	require.NoError(t, err)

	item := models.ContentAnalysisItem{
		DetectionType:    "missing_required",
		SuggestedContent: json.RawMessage(content),
	}

	info, ok := extractOverrideKey(item)

	require.True(t, ok)
	assert.Equal(t, "required", info.constraintType)
	assert.Equal(t, "npc:located_at", info.overrideKey)
}

// ---------------------------------------------------------------------------
// Empty / missing SuggestedContent tests
// ---------------------------------------------------------------------------

func TestExtractOverrideKey_EmptySuggestedContent(t *testing.T) {
	overridableTypes := []string{
		"invalid_type_pair",
		"cardinality_violation",
		"missing_required",
	}

	for _, dt := range overridableTypes {
		t.Run(dt+"_empty", func(t *testing.T) {
			item := models.ContentAnalysisItem{
				DetectionType:    dt,
				SuggestedContent: json.RawMessage{},
			}
			_, ok := extractOverrideKey(item)
			assert.False(t, ok,
				"empty SuggestedContent should not extract a key")
		})
	}
}

func TestExtractOverrideKey_NilSuggestedContent(t *testing.T) {
	overridableTypes := []string{
		"invalid_type_pair",
		"cardinality_violation",
		"missing_required",
	}

	for _, dt := range overridableTypes {
		t.Run(dt+"_nil", func(t *testing.T) {
			item := models.ContentAnalysisItem{
				DetectionType:    dt,
				SuggestedContent: nil,
			}
			_, ok := extractOverrideKey(item)
			assert.False(t, ok,
				"nil SuggestedContent should not extract a key")
		})
	}
}

func TestExtractOverrideKey_MalformedJSON(t *testing.T) {
	overridableTypes := []string{
		"invalid_type_pair",
		"cardinality_violation",
		"missing_required",
	}

	for _, dt := range overridableTypes {
		t.Run(dt+"_malformed", func(t *testing.T) {
			item := models.ContentAnalysisItem{
				DetectionType:    dt,
				SuggestedContent: json.RawMessage(`{not valid json`),
			}
			_, ok := extractOverrideKey(item)
			assert.False(t, ok,
				"malformed JSON should not extract a key")
		})
	}
}

// ---------------------------------------------------------------------------
// Partial / missing fields tests
// ---------------------------------------------------------------------------

func TestExtractTypePairKey_MissingFields(t *testing.T) {
	tests := []struct {
		name    string
		content map[string]interface{}
	}{
		{
			name: "missing relationshipType",
			content: map[string]interface{}{
				"sourceEntityType": "location",
				"targetEntityType": "event",
			},
		},
		{
			name: "missing sourceEntityType",
			content: map[string]interface{}{
				"relationshipType": "employs",
				"targetEntityType": "event",
			},
		},
		{
			name: "missing targetEntityType",
			content: map[string]interface{}{
				"relationshipType": "employs",
				"sourceEntityType": "location",
			},
		},
		{
			name: "all empty strings",
			content: map[string]interface{}{
				"relationshipType": "",
				"sourceEntityType": "",
				"targetEntityType": "",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.content)
			require.NoError(t, err)

			_, ok := extractTypePairKey(json.RawMessage(data))
			assert.False(t, ok,
				"should not extract key with missing fields")
		})
	}
}

func TestExtractCardinalityKey_MissingFields(t *testing.T) {
	tests := []struct {
		name    string
		content map[string]interface{}
	}{
		{
			name: "missing relationshipType",
			content: map[string]interface{}{
				"entityId":  42,
				"direction": "source",
			},
		},
		{
			name: "missing entityId",
			content: map[string]interface{}{
				"relationshipType": "allied_with",
				"direction":        "source",
			},
		},
		{
			name: "missing direction",
			content: map[string]interface{}{
				"relationshipType": "allied_with",
				"entityId":         42,
			},
		},
		{
			name: "zero entityId",
			content: map[string]interface{}{
				"relationshipType": "allied_with",
				"entityId":         0,
				"direction":        "source",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.content)
			require.NoError(t, err)

			_, ok := extractCardinalityKey(json.RawMessage(data))
			assert.False(t, ok,
				"should not extract key with missing fields")
		})
	}
}

func TestExtractRequiredKey_MissingFields(t *testing.T) {
	tests := []struct {
		name    string
		content map[string]interface{}
	}{
		{
			name: "missing entityType",
			content: map[string]interface{}{
				"missingRelationshipType": "located_at",
			},
		},
		{
			name: "missing missingRelationshipType",
			content: map[string]interface{}{
				"entityType": "npc",
			},
		},
		{
			name: "both empty strings",
			content: map[string]interface{}{
				"entityType":              "",
				"missingRelationshipType": "",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.content)
			require.NoError(t, err)

			_, ok := extractRequiredKey(json.RawMessage(data))
			assert.False(t, ok,
				"should not extract key with missing fields")
		})
	}
}

// ---------------------------------------------------------------------------
// Mixed items passthrough test
// ---------------------------------------------------------------------------

func TestFilterOverriddenFindings_NonOverridablePassThrough(t *testing.T) {
	// When db is nil, all items should pass through regardless
	// of detection type. This verifies that non-overridable types
	// are never filtered even in principle.
	items := []models.ContentAnalysisItem{
		{DetectionType: "orphan_warning", MatchedText: "Viktor"},
		{DetectionType: "redundant_edge", MatchedText: "duplicate"},
		{DetectionType: "graph_warning", MatchedText: "warning"},
	}

	result := FilterOverriddenFindings(
		context.TODO(), nil, 1, items,
	)

	require.Len(t, result, 3)
	assert.Equal(t, "orphan_warning", result[0].DetectionType)
	assert.Equal(t, "redundant_edge", result[1].DetectionType)
	assert.Equal(t, "graph_warning", result[2].DetectionType)
}

// ---------------------------------------------------------------------------
// overrideKeyInfo construction tests
// ---------------------------------------------------------------------------

func TestOverrideKeyInfo_TypePairFormat(t *testing.T) {
	// Verify the exact key format for domain_range overrides.
	content, err := json.Marshal(map[string]interface{}{
		"relationshipType": "rules",
		"sourceEntityType": "faction",
		"targetEntityType": "location",
	})
	require.NoError(t, err)

	info, ok := extractTypePairKey(json.RawMessage(content))

	require.True(t, ok)
	assert.Equal(t, "domain_range", info.constraintType)
	assert.Equal(t, "rules:faction:location", info.overrideKey)
}

func TestOverrideKeyInfo_CardinalityFormat(t *testing.T) {
	// Verify the exact key format for cardinality overrides.
	content, err := json.Marshal(map[string]interface{}{
		"relationshipType": "belongs_to",
		"entityId":         100,
		"direction":        "target",
	})
	require.NoError(t, err)

	info, ok := extractCardinalityKey(json.RawMessage(content))

	require.True(t, ok)
	assert.Equal(t, "cardinality", info.constraintType)
	assert.Equal(t, "belongs_to:100:target", info.overrideKey)
}

func TestOverrideKeyInfo_RequiredFormat(t *testing.T) {
	// Verify the exact key format for required overrides.
	content, err := json.Marshal(map[string]interface{}{
		"entityType":              "location",
		"missingRelationshipType": "contains",
	})
	require.NoError(t, err)

	info, ok := extractRequiredKey(json.RawMessage(content))

	require.True(t, ok)
	assert.Equal(t, "required", info.constraintType)
	assert.Equal(t, "location:contains", info.overrideKey)
}
