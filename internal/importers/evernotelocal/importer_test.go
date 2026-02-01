/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package evernotelocal

import (
	"testing"

	"github.com/antonypegg/imagineer/internal/importers/common"
)

func TestExtractTextFromHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "simple HTML",
			input:    "<p>Hello World</p>",
			expected: "Hello World",
		},
		{
			name:     "HTML with line breaks",
			input:    "<p>Hello</p><p>World</p>",
			expected: "Hello\nWorld",
		},
		{
			name:     "HTML with br tags",
			input:    "Hello<br>World<br/>Test",
			expected: "Hello\nWorld\nTest",
		},
		{
			name:     "HTML with ampersand entity",
			input:    "<p>Hello &amp; World</p>",
			expected: "Hello & World",
		},
		{
			name:     "CDATA wrapper",
			input:    "<![CDATA[<p>Hello World</p>]]>",
			expected: "Hello World",
		},
		{
			name:     "XML declaration",
			input:    "<?xml version=\"1.0\"?><p>Hello World</p>",
			expected: "Hello World",
		},
		{
			name:     "DOCTYPE",
			input:    "<!DOCTYPE html><p>Hello World</p>",
			expected: "Hello World",
		},
		{
			name:     "complex HTML",
			input:    "<div><p>First paragraph</p><p>Second paragraph</p></div>",
			expected: "First paragraph\nSecond paragraph",
		},
		{
			name:     "multiple consecutive newlines reduced",
			input:    "<p>First</p><br><br><br><p>Second</p>",
			expected: "First\n\nSecond",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTextFromHTML(tt.input)
			if result != tt.expected {
				t.Errorf("extractTextFromHTML(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDetectEntityType(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		content  string
		tags     []string
		expected common.EntityType
	}{
		{
			name:     "NPC tag",
			title:    "John Smith",
			content:  "A mysterious stranger",
			tags:     []string{"NPC", "ally"},
			expected: common.EntityTypeNPC,
		},
		{
			name:     "character tag",
			title:    "Jane Doe",
			content:  "Another person",
			tags:     []string{"Character"},
			expected: common.EntityTypeNPC,
		},
		{
			name:     "location tag",
			title:    "The Warehouse",
			content:  "An abandoned building",
			tags:     []string{"Location"},
			expected: common.EntityTypeLocation,
		},
		{
			name:     "place tag",
			title:    "Central Park",
			content:  "A large park",
			tags:     []string{"Place"},
			expected: common.EntityTypeLocation,
		},
		{
			name:     "item tag",
			title:    "Ancient Sword",
			content:  "A magical weapon",
			tags:     []string{"Item"},
			expected: common.EntityTypeItem,
		},
		{
			name:     "artifact tag",
			title:    "The Necronomicon",
			content:  "A dark tome",
			tags:     []string{"Artifact"},
			expected: common.EntityTypeItem,
		},
		{
			name:     "faction tag",
			title:    "The Cult",
			content:  "A secret organization",
			tags:     []string{"Faction"},
			expected: common.EntityTypeFaction,
		},
		{
			name:     "organization tag",
			title:    "MI6",
			content:  "British intelligence",
			tags:     []string{"Organization"},
			expected: common.EntityTypeFaction,
		},
		{
			name:     "clue tag",
			title:    "Bloody Fingerprint",
			content:  "Found at the scene",
			tags:     []string{"Clue"},
			expected: common.EntityTypeClue,
		},
		{
			name:     "evidence tag",
			title:    "Murder Weapon",
			content:  "The knife",
			tags:     []string{"Evidence"},
			expected: common.EntityTypeClue,
		},
		{
			name:     "creature tag",
			title:    "Deep One",
			content:  "A fish-like monster",
			tags:     []string{"Creature"},
			expected: common.EntityTypeCreature,
		},
		{
			name:     "monster tag",
			title:    "Shoggoth",
			content:  "An amorphous horror",
			tags:     []string{"Monster"},
			expected: common.EntityTypeCreature,
		},
		{
			name:     "title with Dr prefix",
			title:    "Dr. Herbert West",
			content:  "A mad scientist",
			tags:     []string{},
			expected: common.EntityTypeNPC,
		},
		{
			name:     "title with Professor prefix",
			title:    "Professor Armitage",
			content:  "A scholar",
			tags:     []string{},
			expected: common.EntityTypeNPC,
		},
		{
			name:     "title with Mr prefix",
			title:    "Mr. Whateley",
			content:  "A farmer",
			tags:     []string{},
			expected: common.EntityTypeNPC,
		},
		{
			name:     "title with hotel",
			title:    "Grand Hotel",
			content:  "A fancy establishment",
			tags:     []string{},
			expected: common.EntityTypeLocation,
		},
		{
			name:     "title with mansion",
			title:    "Arkham Mansion",
			content:  "An old house",
			tags:     []string{},
			expected: common.EntityTypeLocation,
		},
		{
			name:     "title with street",
			title:    "Church Street",
			content:  "A road in town",
			tags:     []string{},
			expected: common.EntityTypeLocation,
		},
		{
			name:     "content with occupation field",
			title:    "Unknown Person",
			content:  "Occupation: Detective\nAge: 35",
			tags:     []string{},
			expected: common.EntityTypeNPC,
		},
		{
			name:     "content with age field",
			title:    "Another Person",
			content:  "Age: 42\nBackground: Unknown",
			tags:     []string{},
			expected: common.EntityTypeNPC,
		},
		{
			name:     "no indicators defaults to other",
			title:    "Random Note",
			content:  "Some general information",
			tags:     []string{},
			expected: common.EntityTypeOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectEntityType(tt.title, tt.content, tt.tags)
			if result != tt.expected {
				t.Errorf("detectEntityType(%q, %q, %v) = %v, want %v",
					tt.title, tt.content, tt.tags, result, tt.expected)
			}
		})
	}
}

func TestNewImporter(t *testing.T) {
	importer := New()
	if importer == nil {
		t.Fatal("New() returned nil")
	}
	if importer.executor == nil {
		t.Error("New() did not initialize executor")
	}
}

func TestImporterName(t *testing.T) {
	importer := New()
	name := importer.Name()
	if name != "Evernote Local" {
		t.Errorf("Name() = %q, want %q", name, "Evernote Local")
	}
}
