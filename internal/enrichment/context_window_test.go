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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// ExtractContextWindow tests
// ---------------------------------------------------------------------------

func TestExtractContextWindow_MiddleParagraph(t *testing.T) {
	// Three paragraphs separated by "\n\n". Position falls in the
	// middle paragraph; the result should include the target plus
	// one paragraph before and one after.
	content := "First paragraph here.\n\nSecond paragraph here.\n\nThird paragraph here."

	// "Second paragraph here." starts at byte 23 (after "First paragraph here.\n\n").
	start := 23
	end := 45 // end of "Second paragraph here."

	window, wStart, wEnd := ExtractContextWindow(content, start, end)

	assert.Equal(t, content, window,
		"middle paragraph should return preceding + target + following")
	assert.Equal(t, 0, wStart,
		"window should start at the beginning (includes first paragraph)")
	assert.Equal(t, len(content), wEnd,
		"window should extend to the end (includes last paragraph)")
}

func TestExtractContextWindow_FirstParagraph(t *testing.T) {
	// Position falls in the first paragraph. There is no preceding
	// paragraph, so the result should include target + one after.
	content := "First paragraph.\n\nSecond paragraph.\n\nThird paragraph."

	start := 0
	end := 16 // "First paragraph."

	window, wStart, wEnd := ExtractContextWindow(content, start, end)

	// Should include first and second paragraphs but not the third.
	expected := "First paragraph.\n\nSecond paragraph."
	assert.Equal(t, expected, window,
		"first paragraph should return target + one after")
	assert.Equal(t, 0, wStart,
		"window should start at the beginning")
	assert.Equal(t, len(expected), wEnd,
		"window should end after the second paragraph")
}

func TestExtractContextWindow_LastParagraph(t *testing.T) {
	// Position falls in the last paragraph. There is no following
	// paragraph, so the result should include one before + target.
	content := "First paragraph.\n\nSecond paragraph.\n\nThird paragraph."

	// "Third paragraph." starts at byte 37.
	start := 37
	end := 53 // "Third paragraph."

	window, wStart, wEnd := ExtractContextWindow(content, start, end)

	// Should include second and third paragraphs but not the first.
	expected := "Second paragraph.\n\nThird paragraph."
	assert.Equal(t, expected, window,
		"last paragraph should return one before + target")
	assert.Equal(t, 18, wStart,
		"window should start at the second paragraph")
	assert.Equal(t, len(content), wEnd,
		"window should extend to the end of content")
}

func TestExtractContextWindow_SingleParagraph(t *testing.T) {
	// Content with no paragraph breaks should return the entire
	// content.
	content := "This is a single paragraph with no breaks."

	start := 5
	end := 10

	window, wStart, wEnd := ExtractContextWindow(content, start, end)

	assert.Equal(t, content, window,
		"single paragraph should return entire content")
	assert.Equal(t, 0, wStart)
	assert.Equal(t, len(content), wEnd)
}

func TestExtractContextWindow_OutOfBounds(t *testing.T) {
	// When the position is beyond the content length, the function
	// should return the entire content.
	content := "Short content.\n\nAnother paragraph."

	start := 500
	end := 600

	window, wStart, wEnd := ExtractContextWindow(content, start, end)

	assert.Equal(t, content, window,
		"out-of-bounds position should return entire content")
	assert.Equal(t, 0, wStart)
	assert.Equal(t, len(content), wEnd)
}

func TestExtractContextWindow_EmptyContent(t *testing.T) {
	window, wStart, wEnd := ExtractContextWindow("", 0, 0)

	assert.Equal(t, "", window)
	assert.Equal(t, 0, wStart)
	assert.Equal(t, 0, wEnd)
}

func TestExtractContextWindow_PositionInDelimiter(t *testing.T) {
	// When the position falls in the "\n\n" delimiter between
	// paragraphs, the item should be assigned to the next paragraph.
	content := "First paragraph.\n\nSecond paragraph.\n\nThird paragraph."

	// Position at byte 16 is the first '\n' of the delimiter.
	start := 16
	end := 17

	window, wStart, wEnd := ExtractContextWindow(content, start, end)

	// The position should be assigned to the second paragraph.
	// Window should include: first (before) + second (target) + third (after).
	assert.Equal(t, content, window,
		"position in delimiter should be assigned to next paragraph")
	assert.Equal(t, 0, wStart)
	assert.Equal(t, len(content), wEnd)
}

func TestExtractContextWindow_NegativeStart(t *testing.T) {
	content := "First paragraph.\n\nSecond paragraph."

	// Negative start should be clamped to 0.
	window, wStart, wEnd := ExtractContextWindow(content, -5, 5)

	assert.Equal(t, "First paragraph.\n\nSecond paragraph.", window,
		"negative start should be clamped, returning first paragraph + next")
	assert.Equal(t, 0, wStart)
	assert.Equal(t, len(content), wEnd)
}

func TestExtractContextWindow_FourParagraphs(t *testing.T) {
	// Four paragraphs: position in second. Window should include
	// paragraphs 1, 2, and 3 but not 4.
	content := "Para one.\n\nPara two.\n\nPara three.\n\nPara four."

	// "Para two." starts at byte 11.
	start := 11
	end := 20 // "Para two."

	window, wStart, wEnd := ExtractContextWindow(content, start, end)

	expected := "Para one.\n\nPara two.\n\nPara three."
	assert.Equal(t, expected, window,
		"should include one before, target, one after")
	assert.Equal(t, 0, wStart)
	assert.Equal(t, len(expected), wEnd)
}

func TestExtractContextWindow_ByteOffsets(t *testing.T) {
	// Verify that windowStart and windowEnd can be used to slice
	// the original content to reproduce the window.
	content := "Alpha.\n\nBeta.\n\nGamma.\n\nDelta."

	// "Gamma." starts at byte 15.
	start := 15
	end := 21 // "Gamma."

	window, wStart, wEnd := ExtractContextWindow(content, start, end)

	assert.Equal(t, window, content[wStart:wEnd],
		"content[windowStart:windowEnd] should equal the returned window")
}

// ---------------------------------------------------------------------------
// GroupItemsByParagraph tests
// ---------------------------------------------------------------------------

func TestGroupItemsByParagraph_NoOverlap(t *testing.T) {
	// Two items in different paragraphs should produce separate groups.
	content := "First paragraph.\n\nSecond paragraph.\n\nThird paragraph."

	items := []ItemPosition{
		{ID: 1, Start: 0, End: 16},  // in first paragraph
		{ID: 2, Start: 37, End: 53}, // in third paragraph
	}

	groups := GroupItemsByParagraph(items, content)

	require.Len(t, groups, 2,
		"items in different paragraphs should produce separate groups")

	// Verify each group has exactly one item.
	for _, g := range groups {
		assert.Len(t, g, 1)
	}

	// Verify both IDs are present.
	allIDs := collectIDs(groups)
	assert.Contains(t, allIDs, int64(1))
	assert.Contains(t, allIDs, int64(2))
}

func TestGroupItemsByParagraph_SameParagraph(t *testing.T) {
	// Two items in the same paragraph should be grouped together.
	content := "First paragraph with multiple findings.\n\nSecond paragraph."

	items := []ItemPosition{
		{ID: 1, Start: 0, End: 5},
		{ID: 2, Start: 10, End: 20},
	}

	groups := GroupItemsByParagraph(items, content)

	require.Len(t, groups, 1,
		"items in same paragraph should produce one group")
	assert.Len(t, groups[0], 2)
}

func TestGroupItemsByParagraph_Empty(t *testing.T) {
	groups := GroupItemsByParagraph(nil, "Some content.")

	assert.Empty(t, groups,
		"empty input should produce empty output")
}

func TestGroupItemsByParagraph_EmptySlice(t *testing.T) {
	groups := GroupItemsByParagraph([]ItemPosition{}, "Some content.")

	assert.Empty(t, groups,
		"empty slice should produce empty output")
}

func TestGroupItemsByParagraph_OverlappingPositions(t *testing.T) {
	// Two items with overlapping byte ranges in the same paragraph
	// should be grouped together.
	content := "A long paragraph with overlapping ranges."

	items := []ItemPosition{
		{ID: 1, Start: 0, End: 20},
		{ID: 2, Start: 15, End: 35},
	}

	groups := GroupItemsByParagraph(items, content)

	require.Len(t, groups, 1,
		"overlapping items in same paragraph should produce one group")
	assert.Len(t, groups[0], 2)
}

func TestGroupItemsByParagraph_MultipleGroups(t *testing.T) {
	// Three items: two in paragraph 1, one in paragraph 2.
	content := "First paragraph.\n\nSecond paragraph."

	items := []ItemPosition{
		{ID: 1, Start: 0, End: 5},
		{ID: 2, Start: 10, End: 16},
		{ID: 3, Start: 18, End: 30},
	}

	groups := GroupItemsByParagraph(items, content)

	require.Len(t, groups, 2,
		"should produce two groups")

	// Find which group has 2 items and which has 1.
	var twoGroup, oneGroup []ItemPosition
	for _, g := range groups {
		if len(g) == 2 {
			twoGroup = g
		} else {
			oneGroup = g
		}
	}

	require.NotNil(t, twoGroup, "should have a group with 2 items")
	require.NotNil(t, oneGroup, "should have a group with 1 item")

	// The two-item group should contain IDs 1 and 2 (first paragraph).
	twoIDs := make(map[int64]bool)
	for _, item := range twoGroup {
		twoIDs[item.ID] = true
	}
	assert.True(t, twoIDs[1] && twoIDs[2],
		"group with 2 items should contain IDs 1 and 2")

	// The one-item group should contain ID 3 (second paragraph).
	assert.Equal(t, int64(3), oneGroup[0].ID)
}

func TestGroupItemsByParagraph_SingleItem(t *testing.T) {
	content := "First paragraph.\n\nSecond paragraph."

	items := []ItemPosition{
		{ID: 42, Start: 0, End: 10},
	}

	groups := GroupItemsByParagraph(items, content)

	require.Len(t, groups, 1)
	assert.Len(t, groups[0], 1)
	assert.Equal(t, int64(42), groups[0][0].ID)
}

func TestGroupItemsByParagraph_PreservesItemData(t *testing.T) {
	// Verify that grouping preserves all fields of each ItemPosition.
	content := "Only one paragraph."

	items := []ItemPosition{
		{ID: 100, Start: 0, End: 5},
		{ID: 200, Start: 10, End: 19},
	}

	groups := GroupItemsByParagraph(items, content)

	require.Len(t, groups, 1)
	require.Len(t, groups[0], 2)

	idMap := make(map[int64]ItemPosition)
	for _, item := range groups[0] {
		idMap[item.ID] = item
	}

	assert.Equal(t, 0, idMap[100].Start)
	assert.Equal(t, 5, idMap[100].End)
	assert.Equal(t, 10, idMap[200].Start)
	assert.Equal(t, 19, idMap[200].End)
}

// collectIDs extracts all IDs from grouped items for assertion helpers.
func collectIDs(groups [][]ItemPosition) []int64 {
	var ids []int64
	for _, g := range groups {
		for _, item := range g {
			ids = append(ids, item.ID)
		}
	}
	return ids
}
