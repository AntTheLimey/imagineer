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

import "strings"

// ItemPosition identifies a finding's byte range within a content string.
// ID is the finding or item identifier; Start and End are byte offsets
// into the original content (Start is inclusive, End is exclusive).
type ItemPosition struct {
	ID    int64
	Start int
	End   int
}

// paragraph records the byte-offset boundaries of a single paragraph
// within a content string.
type paragraph struct {
	start int // byte offset of first char in content
	end   int // byte offset past last char in content
}

// splitParagraphs splits content on "\n\n" delimiters and returns a
// slice of paragraph structs with byte-offset boundaries.
func splitParagraphs(content string) []paragraph {
	delimiter := "\n\n"
	var paragraphs []paragraph
	offset := 0
	remaining := content

	for {
		idx := strings.Index(remaining, delimiter)
		if idx == -1 {
			paragraphs = append(paragraphs, paragraph{
				start: offset,
				end:   offset + len(remaining),
			})
			break
		}
		paragraphs = append(paragraphs, paragraph{
			start: offset,
			end:   offset + idx,
		})
		offset += idx + len(delimiter)
		remaining = remaining[idx+len(delimiter):]
	}

	return paragraphs
}

// findParagraphIndex returns the index of the paragraph that contains
// the given byte position. If the position falls in a "\n\n" delimiter
// between paragraphs, it is assigned to the next paragraph.
func findParagraphIndex(paragraphs []paragraph, pos int) int {
	for i, p := range paragraphs {
		if pos >= p.start && pos < p.end {
			return i
		}
		// Check if pos falls in the delimiter after this paragraph
		// and before the next one.
		if i+1 < len(paragraphs) {
			nextStart := paragraphs[i+1].start
			if pos >= p.end && pos < nextStart {
				return i + 1
			}
		}
	}
	// Default to the last paragraph.
	return len(paragraphs) - 1
}

// ExtractContextWindow returns the paragraph containing the byte range
// [start, end) plus one paragraph before and one paragraph after.
// Paragraphs are delimited by "\n\n". The returned windowStart and
// windowEnd are byte offsets into the original content so that
// content[windowStart:windowEnd] == window.
//
// If start is beyond the content length, or if start and end are
// invalid, the entire content is returned.
func ExtractContextWindow(
	content string,
	start, end int,
) (window string, windowStart, windowEnd int) {
	if content == "" {
		return "", 0, 0
	}

	// Clamp positions to content bounds.
	if start < 0 {
		start = 0
	}
	if start >= len(content) {
		return content, 0, len(content)
	}
	paragraphs := splitParagraphs(content)
	if len(paragraphs) == 0 {
		return content, 0, len(content)
	}

	targetIdx := findParagraphIndex(paragraphs, start)

	// Compute the range: target +/- 1 paragraph.
	first := targetIdx - 1
	if first < 0 {
		first = 0
	}
	last := targetIdx + 1
	if last >= len(paragraphs) {
		last = len(paragraphs) - 1
	}

	windowStart = paragraphs[first].start
	windowEnd = paragraphs[last].end

	window = content[windowStart:windowEnd]
	return window, windowStart, windowEnd
}

// GroupItemsByParagraph groups items that fall within the same paragraph.
// Paragraphs are delimited by "\n\n". Items whose byte ranges overlap
// the same paragraph are placed in the same group. The order of groups
// follows the paragraph order in the content; within each group the
// original item order is preserved.
func GroupItemsByParagraph(
	items []ItemPosition,
	content string,
) [][]ItemPosition {
	if len(items) == 0 {
		return nil
	}

	paragraphs := splitParagraphs(content)

	// Assign each item to a paragraph index.
	paraForItem := make([]int, len(items))
	for i, item := range items {
		pos := item.Start
		if pos < 0 {
			pos = 0
		}
		paraForItem[i] = findParagraphIndex(paragraphs, pos)
	}

	// Group items by paragraph index, preserving order.
	groupMap := make(map[int][]ItemPosition)
	var order []int
	seen := make(map[int]bool)

	for i, pIdx := range paraForItem {
		if !seen[pIdx] {
			seen[pIdx] = true
			order = append(order, pIdx)
		}
		groupMap[pIdx] = append(groupMap[pIdx], items[i])
	}

	groups := make([][]ItemPosition, 0, len(order))
	for _, pIdx := range order {
		groups = append(groups, groupMap[pIdx])
	}

	return groups
}
