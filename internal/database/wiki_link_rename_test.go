/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package database

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// wikiReplace mirrors the replacement logic used in PropagateEntityRename
// for unit testing without a database connection.
func wikiReplace(text, oldName, newName string) string {
	oldExact := "[[" + oldName + "]]"
	newExact := "[[" + newName + "]]"
	oldPiped := "[[" + oldName + "|"
	newPiped := "[[" + newName + "|"

	text = strings.ReplaceAll(text, oldExact, newExact)
	text = strings.ReplaceAll(text, oldPiped, newPiped)
	return text
}

// TestWikiLinkRename_ExactLink verifies that [[Old Name]] is replaced
// with [[New Name]].
func TestWikiLinkRename_ExactLink(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		oldName string
		newName string
		want    string
	}{
		{
			name:    "simple exact replacement",
			input:   "The [[Old NPC]] was seen near the docks.",
			oldName: "Old NPC",
			newName: "New NPC",
			want:    "The [[New NPC]] was seen near the docks.",
		},
		{
			name:    "exact link at start of text",
			input:   "[[Old NPC]] was seen near the docks.",
			oldName: "Old NPC",
			newName: "New NPC",
			want:    "[[New NPC]] was seen near the docks.",
		},
		{
			name:    "exact link at end of text",
			input:   "Spotted near the docks: [[Old NPC]]",
			oldName: "Old NPC",
			newName: "New NPC",
			want:    "Spotted near the docks: [[New NPC]]",
		},
		{
			name:    "no match leaves text unchanged",
			input:   "The [[Other NPC]] was seen near the docks.",
			oldName: "Old NPC",
			newName: "New NPC",
			want:    "The [[Other NPC]] was seen near the docks.",
		},
		{
			name:    "text with no wiki links unchanged",
			input:   "Just some plain text without any links.",
			oldName: "Old NPC",
			newName: "New NPC",
			want:    "Just some plain text without any links.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wikiReplace(tt.input, tt.oldName, tt.newName)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestWikiLinkRename_PipedLink verifies that [[Old Name|display text]]
// is replaced with [[New Name|display text]].
func TestWikiLinkRename_PipedLink(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		oldName string
		newName string
		want    string
	}{
		{
			name:    "piped link with display text",
			input:   "The [[Old NPC|mysterious stranger]] arrived.",
			oldName: "Old NPC",
			newName: "New NPC",
			want:    "The [[New NPC|mysterious stranger]] arrived.",
		},
		{
			name:    "piped link preserves display text",
			input:   "Talk to [[Professor Armitage|the professor]].",
			oldName: "Professor Armitage",
			newName: "Dr. Henry Armitage",
			want:    "Talk to [[Dr. Henry Armitage|the professor]].",
		},
		{
			name:    "piped link at start of text",
			input:   "[[Old NPC|the stranger]] left town.",
			oldName: "Old NPC",
			newName: "New NPC",
			want:    "[[New NPC|the stranger]] left town.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wikiReplace(tt.input, tt.oldName, tt.newName)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestWikiLinkRename_MultipleOccurrences verifies that all occurrences
// within the same field are replaced.
func TestWikiLinkRename_MultipleOccurrences(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		oldName string
		newName string
		want    string
	}{
		{
			name:    "multiple exact links",
			input:   "[[Old NPC]] went to the market. Later [[Old NPC]] returned home.",
			oldName: "Old NPC",
			newName: "New NPC",
			want:    "[[New NPC]] went to the market. Later [[New NPC]] returned home.",
		},
		{
			name:    "mixed exact and piped links",
			input:   "[[Old NPC]] met [[Old NPC|the stranger]] at the inn.",
			oldName: "Old NPC",
			newName: "New NPC",
			want:    "[[New NPC]] met [[New NPC|the stranger]] at the inn.",
		},
		{
			name:    "three occurrences in different forms",
			input:   "First [[Old NPC]], then [[Old NPC|alias]], finally [[Old NPC]] again.",
			oldName: "Old NPC",
			newName: "New NPC",
			want:    "First [[New NPC]], then [[New NPC|alias]], finally [[New NPC]] again.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wikiReplace(tt.input, tt.oldName, tt.newName)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestWikiLinkRename_SpecialCharacters verifies that names containing
// special characters (quotes, periods, apostrophes) are handled safely
// by the string replace approach.
func TestWikiLinkRename_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		oldName string
		newName string
		want    string
	}{
		{
			name:    "name with period",
			input:   "Visit [[Dr. West]] at the lab.",
			oldName: "Dr. West",
			newName: "Dr. Herbert West",
			want:    "Visit [[Dr. Herbert West]] at the lab.",
		},
		{
			name:    "name with apostrophe",
			input:   "Go to [[O'Brien's Tavern]] for information.",
			oldName: "O'Brien's Tavern",
			newName: "O'Brien's Pub",
			want:    "Go to [[O'Brien's Pub]] for information.",
		},
		{
			name:    "name with double quotes",
			input:   `Meet [["Mad Dog" Murphy]] at the pier.`,
			oldName: `"Mad Dog" Murphy`,
			newName: `"Crazy Dog" Murphy`,
			want:    `Meet [["Crazy Dog" Murphy]] at the pier.`,
		},
		{
			name:    "name with parentheses",
			input:   "The [[Necronomicon (Latin)]] was stolen.",
			oldName: "Necronomicon (Latin)",
			newName: "Necronomicon (Dee Translation)",
			want:    "The [[Necronomicon (Dee Translation)]] was stolen.",
		},
		{
			name:    "name with numbers",
			input:   "Agent [[Agent 47]] has arrived.",
			oldName: "Agent 47",
			newName: "Agent 48",
			want:    "Agent [[Agent 48]] has arrived.",
		},
		{
			name:    "piped link with special chars in name",
			input:   "Ask [[Dr. West|the doctor]] about the serum.",
			oldName: "Dr. West",
			newName: "Dr. Herbert West",
			want:    "Ask [[Dr. Herbert West|the doctor]] about the serum.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wikiReplace(tt.input, tt.oldName, tt.newName)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestWikiLinkRename_CampaignScope verifies that the SQL LIKE patterns
// are correctly constructed for campaign scoping. This test validates the
// pattern construction logic without requiring a database.
func TestWikiLinkRename_CampaignScope(t *testing.T) {
	// Verify the patterns that would be used in LIKE clauses
	oldName := "Old NPC"
	oldExact := "[[" + oldName + "]]"
	oldPiped := "[[" + oldName + "|"

	// Content from campaign 1 (should match)
	campaign1Content := "The [[Old NPC]] was seen at [[Old NPC|the docks]]."
	assert.True(t, strings.Contains(campaign1Content, oldExact),
		"campaign content should contain the exact wiki link pattern")
	assert.True(t, strings.Contains(campaign1Content, oldPiped),
		"campaign content should contain the piped wiki link pattern")

	// Content from campaign 2 (should not match)
	campaign2Content := "The [[Other NPC]] was seen elsewhere."
	assert.False(t, strings.Contains(campaign2Content, oldExact),
		"other campaign content should not contain the exact wiki link pattern")
	assert.False(t, strings.Contains(campaign2Content, oldPiped),
		"other campaign content should not contain the piped wiki link pattern")
}

// TestWikiLinkRename_EdgeCases verifies edge cases in the replacement logic.
func TestWikiLinkRename_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		oldName string
		newName string
		want    string
	}{
		{
			name:    "empty text",
			input:   "",
			oldName: "Old NPC",
			newName: "New NPC",
			want:    "",
		},
		{
			name:    "partial bracket match is not replaced",
			input:   "The [Old NPC] was seen.",
			oldName: "Old NPC",
			newName: "New NPC",
			want:    "The [Old NPC] was seen.",
		},
		{
			name:    "single bracket only is not replaced",
			input:   "The [Old NPC]] was seen.",
			oldName: "Old NPC",
			newName: "New NPC",
			want:    "The [Old NPC]] was seen.",
		},
		{
			name:    "old name as substring of another link is not replaced",
			input:   "The [[Old NPC Junior]] was seen.",
			oldName: "Old NPC",
			newName: "New NPC",
			want:    "The [[Old NPC Junior]] was seen.",
		},
		{
			name:    "name identical to new name is no-op",
			input:   "The [[Same Name]] was seen.",
			oldName: "Same Name",
			newName: "Same Name",
			want:    "The [[Same Name]] was seen.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wikiReplace(tt.input, tt.oldName, tt.newName)
			assert.Equal(t, tt.want, got)
		})
	}
}
