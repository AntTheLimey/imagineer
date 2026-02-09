/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package analysis

import (
	"testing"

	"github.com/antonypegg/imagineer/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStripWikiLinks verifies that wiki link markup is replaced with
// plain display text for both simple and aliased link forms.
func TestStripWikiLinks(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple wiki link",
			input: "[[Arkham]]",
			want:  "Arkham",
		},
		{
			name:  "aliased wiki link",
			input: "[[Nyarlathotep|The Crawling Chaos]]",
			want:  "The Crawling Chaos",
		},
		{
			name:  "mixed links in sentence",
			input: "Visit [[Arkham]] and talk to [[Dr. Armitage|the Professor]]",
			want:  "Visit Arkham and talk to the Professor",
		},
		{
			name:  "no wiki links",
			input: "No wiki links here",
			want:  "No wiki links here",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "multiple simple and aliased links",
			input: "[[A]] and [[B]] and [[C|D]]",
			want:  "A and B and D",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripWikiLinks(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestExtractContextSnippet verifies that context snippets are correctly
// extracted from content with appropriate padding on each side.
func TestExtractContextSnippet(t *testing.T) {
	tests := []struct {
		name    string
		content string
		start   int
		end     int
		want    string
	}{
		{
			name:    "short content covers entire string",
			content: "Arkham is spooky",
			start:   0,
			end:     6,
			want:    "Arkham is spooky",
		},
		{
			name: "long content clipped on both sides",
			// 59 a's + " " + "MATCH" + " " + 59 b's = 129 chars.
			// MATCH is at indices [62:67].
			// snippet start = 62-50 = 12, end = 67+50 = 117.
			content: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa MATCH bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			start:   62,
			end:     67,
			want:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa MATCH bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		},
		{
			name: "match at very start",
			// start=0, end=6 => snippet [0:56].
			content: "Arkham is a city in Lovecraft's fiction with many dark secrets and elder things lurking beneath the surface of everyday life there",
			start:   0,
			end:     6,
			want:    "Arkham is a city in Lovecraft's fiction with many dark s",
		},
		{
			name: "match at very end",
			// 109-char string. start=103, end=109 => snippet [53:109].
			content: "The investigators traveled through the dark and misty countryside to the ancient and foreboding city of Arkham",
			start:   103,
			end:     109,
			want:    " countryside to the ancient and foreboding city of Arkham",
		},
		{
			name:    "zero-length match",
			content: "Some text here for testing purposes with extra padding on each side of the content",
			start:   10,
			end:     10,
			want:    "Some text here for testing purposes with extra padding on ea",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractContextSnippet(tt.content, tt.start, tt.end)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestBuildMatchedNames verifies that resolved wiki link names and
// untagged mention names are merged into a single lowercased set.
func TestBuildMatchedNames(t *testing.T) {
	tests := []struct {
		name          string
		resolvedNames map[string]bool
		untaggedItems []models.ContentAnalysisItem
		want          map[string]bool
	}{
		{
			name:          "empty inputs",
			resolvedNames: map[string]bool{},
			untaggedItems: nil,
			want:          map[string]bool{},
		},
		{
			name: "only resolved names",
			resolvedNames: map[string]bool{
				"arkham":       true,
				"nyarlathotep": true,
			},
			untaggedItems: nil,
			want: map[string]bool{
				"arkham":       true,
				"nyarlathotep": true,
			},
		},
		{
			name:          "only untagged items",
			resolvedNames: map[string]bool{},
			untaggedItems: []models.ContentAnalysisItem{
				{MatchedText: "Dr. Armitage"},
				{MatchedText: "Miskatonic University"},
			},
			want: map[string]bool{
				"dr. armitage":          true,
				"miskatonic university": true,
			},
		},
		{
			name: "both inputs merged",
			resolvedNames: map[string]bool{
				"arkham": true,
			},
			untaggedItems: []models.ContentAnalysisItem{
				{MatchedText: "Dr. Armitage"},
			},
			want: map[string]bool{
				"arkham":       true,
				"dr. armitage": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildMatchedNames(tt.resolvedNames, tt.untaggedItems)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestWikiLinkRegex verifies the compiled wikiLinkRe pattern against
// various wiki link formats and edge cases.
func TestWikiLinkRegex(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantCount  int
		wantGroup1 string
		wantGroup2 string
	}{
		{
			name:       "simple link",
			input:      "[[Arkham]]",
			wantCount:  1,
			wantGroup1: "Arkham",
			wantGroup2: "",
		},
		{
			name:       "aliased link",
			input:      "[[Dr. Armitage|the Professor]]",
			wantCount:  1,
			wantGroup1: "Dr. Armitage",
			wantGroup2: "the Professor",
		},
		{
			name:      "multiple links",
			input:     "[[A]] some text [[B|C]]",
			wantCount: 2,
		},
		{
			name:      "no links",
			input:     "No links",
			wantCount: 0,
		},
		{
			name:      "empty brackets not valid",
			input:     "[[]]",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := wikiLinkRe.FindAllStringSubmatch(tt.input, -1)
			assert.Len(t, matches, tt.wantCount)

			if tt.wantCount == 1 && len(matches) == 1 {
				require.GreaterOrEqual(t, len(matches[0]), 2)
				assert.Equal(t, tt.wantGroup1, matches[0][1])
				if tt.wantGroup2 != "" {
					require.GreaterOrEqual(t, len(matches[0]), 3)
					assert.Equal(t, tt.wantGroup2, matches[0][2])
				} else {
					if len(matches[0]) >= 3 {
						assert.Empty(t, matches[0][2])
					}
				}
			}
		})
	}
}

// TestCapitalizedPhraseRegex verifies the compiled capitalizedPhraseRe
// pattern detects capitalized word sequences and ignores lowercase text.
func TestCapitalizedPhraseRegex(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantMatch bool
		wantFirst string
	}{
		{
			name: "title with period breaks at dot",
			// The period in "Dr." is not in the regex character
			// class, so the match stops at "Dr" and "Armitage
			// walked" starts a separate match.
			input:     "Dr. Armitage walked",
			wantMatch: true,
			wantFirst: "Dr",
		},
		{
			name:      "all lowercase no match",
			input:     "the old man",
			wantMatch: false,
		},
		{
			name:      "multi-word capitalized phrase",
			input:     "Miskatonic University Library",
			wantMatch: true,
			wantFirst: "Miskatonic University Library",
		},
		{
			name:      "all caps word matches",
			input:     "SHOUTING",
			wantMatch: true,
			wantFirst: "SHOUTING",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := capitalizedPhraseRe.FindAllString(tt.input, -1)
			if tt.wantMatch {
				require.NotEmpty(t, matches, "expected at least one match")
				assert.Equal(t, tt.wantFirst, matches[0])
			} else {
				assert.Empty(t, matches, "expected no matches")
			}
		})
	}
}

// TestExtractContextSnippet_Boundaries exercises edge cases around
// content length and position boundaries.
func TestExtractContextSnippet_Boundaries(t *testing.T) {
	tests := []struct {
		name    string
		content string
		start   int
		end     int
	}{
		{
			name:    "content shorter than context radius",
			content: "short",
			start:   1,
			end:     4,
		},
		{
			name:    "start is zero",
			content: "beginning of the long text that stretches well beyond the context radius boundary on at least one side definitely",
			start:   0,
			end:     9,
		},
		{
			name:    "end equals content length",
			content: "text that stretches well beyond the context radius boundary on at least one side to reach the very end",
			start:   96,
			end:     101,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractContextSnippet(tt.content, tt.start, tt.end)

			// The snippet must contain the matched portion.
			matched := tt.content[tt.start:tt.end]
			assert.Contains(t, got, matched)

			// The snippet must not exceed the content boundaries.
			assert.True(t, len(got) <= len(tt.content),
				"snippet length %d exceeds content length %d", len(got), len(tt.content))

			// The snippet must be a valid substring of content.
			assert.Contains(t, tt.content, got,
				"snippet must be a contiguous substring of content")
		})
	}
}
