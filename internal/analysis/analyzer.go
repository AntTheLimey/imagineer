/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

// Package analysis provides content analysis for detecting entity
// references, untagged mentions, and potential misspellings in campaign
// text content.
package analysis

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/models"
)

// wikiLinkRe matches wiki-link syntax: [[Name]] or [[Name|Display]].
var wikiLinkRe = regexp.MustCompile(`\[\[([^\]|]+?)(?:\|([^\]]*?))?\]\]`)

// capitalizedPhraseRe matches capitalized word sequences (1-5 words)
// where the first word starts with an uppercase Unicode letter.
var capitalizedPhraseRe = regexp.MustCompile(`\p{Lu}[\p{L}'-]*(?:\s+[\p{L}][\p{L}'-]*){0,4}`)

// contextRadius is the number of characters to include on each side of
// a match when building context snippets.
const contextRadius = 50

// maxMisspellingCandidates caps the number of misspelling items to
// prevent noise in the results.
const maxMisspellingCandidates = 20

// similarityThresholdResolved is the minimum similarity score for a
// wiki link to be considered fully resolved.
const similarityThresholdResolved = 0.9

// similarityThresholdMinimum is the minimum similarity score for a
// fuzzy match to be considered a possible suggestion.
const similarityThresholdMinimum = 0.6

// wikiLinkRange represents the plain-text position range of a
// stripped wiki link.
type wikiLinkRange struct {
	start, end int
}

// Analyzer performs content analysis against campaign entities stored
// in the database.
type Analyzer struct {
	db *database.DB
}

// NewAnalyzer creates an Analyzer backed by the given database handle.
func NewAnalyzer(db *database.DB) *Analyzer {
	return &Analyzer{db: db}
}

// AnalyzeContent scans content for wiki links, untagged entity
// mentions, and potential misspellings. It persists the results as a
// ContentAnalysisJob with associated ContentAnalysisItems and returns
// both.
func (a *Analyzer) AnalyzeContent(
	ctx context.Context,
	campaignID int64,
	sourceTable, sourceField string,
	sourceID int64,
	content string,
) (*models.ContentAnalysisJob, []models.ContentAnalysisItem, error) {
	var items []models.ContentAnalysisItem

	if content != "" {
		// Scan 1: extract wiki links and resolve them against
		// known entities.
		wikiItems, resolvedNames := a.scanWikiLinks(ctx, campaignID, content)
		items = append(items, wikiItems...)

		// Build a plain-text version of the content with wiki
		// link markup removed, used for scans 2 and 3. Also
		// collect position ranges of stripped wiki links so
		// scan 2 can skip occurrences that fall inside them.
		plainText, wikiRanges := stripWikiLinksWithRanges(content)

		// Scan 2: detect untagged entity name mentions in plain
		// text.
		untaggedItems := a.scanUntaggedMentions(ctx, campaignID, plainText, wikiRanges)
		items = append(items, untaggedItems...)

		// Scan 3: detect possible misspellings of entity names.
		matchedNames := buildMatchedNames(resolvedNames, untaggedItems)
		misspellingItems := a.scanMisspellings(ctx, campaignID, plainText, matchedNames, wikiRanges)
		items = append(items, misspellingItems...)
	}

	// Delete any previous analysis jobs for this exact source field
	// before creating a new one.
	err := a.db.DeleteAnalysisJobsForSource(ctx, campaignID, sourceTable, sourceID, sourceField)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to delete old analysis jobs: %w", err)
	}

	job := &models.ContentAnalysisJob{
		CampaignID:    campaignID,
		SourceTable:   sourceTable,
		SourceID:      sourceID,
		SourceField:   sourceField,
		Status:        "completed",
		TotalItems:    len(items),
		ResolvedItems: 0,
	}

	createdJob, err := a.db.CreateAnalysisJob(ctx, job)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create analysis job: %w", err)
	}

	// Assign the new job ID to every item before batch insert.
	for i := range items {
		items[i].JobID = createdJob.ID
	}

	if len(items) > 0 {
		if err := a.db.CreateAnalysisItems(ctx, items); err != nil {
			return nil, nil, fmt.Errorf("failed to create analysis items: %w", err)
		}
	}

	return createdJob, items, nil
}

// scanWikiLinks extracts [[Entity]] and [[Entity|Display]] references,
// resolves each against the campaign entity list using fuzzy matching,
// and returns analysis items plus the set of entity names that were
// resolved (similarity >= threshold).
func (a *Analyzer) scanWikiLinks(
	ctx context.Context,
	campaignID int64,
	content string,
) ([]models.ContentAnalysisItem, map[string]bool) {
	resolvedNames := make(map[string]bool)
	var items []models.ContentAnalysisItem

	matches := wikiLinkRe.FindAllStringSubmatchIndex(content, -1)
	for _, loc := range matches {
		// loc[0]:loc[1] is the full match including brackets.
		// loc[2]:loc[3] is the entity name (group 1).
		entityName := content[loc[2]:loc[3]]

		item := models.ContentAnalysisItem{
			MatchedText: entityName,
			Resolution:  "pending",
		}

		start := loc[0]
		end := loc[1]
		item.PositionStart = &start
		item.PositionEnd = &end

		snippet := extractContextSnippet(content, start, end)
		item.ContextSnippet = &snippet

		results, err := a.db.ResolveEntityByName(ctx, campaignID, entityName, 1)
		if err != nil {
			log.Printf("analysis: failed to resolve wiki link %q: %v", entityName, err)
			item.DetectionType = "wiki_link_unresolved"
			items = append(items, item)
			continue
		}

		if len(results) > 0 {
			r := results[0]
			item.EntityID = &r.ID
			item.Similarity = &r.Similarity

			if r.Similarity >= similarityThresholdResolved {
				item.DetectionType = "wiki_link_resolved"
				resolvedNames[strings.ToLower(entityName)] = true
			} else if r.Similarity >= similarityThresholdMinimum {
				item.DetectionType = "wiki_link_unresolved"
			} else {
				// Below minimum threshold: treat as fully
				// unresolved with no entity suggestion.
				item.DetectionType = "wiki_link_unresolved"
				item.EntityID = nil
				item.Similarity = nil
			}
		} else {
			item.DetectionType = "wiki_link_unresolved"
		}

		// Only create items for unresolved wiki links. Resolved
		// links are correctly tagged and need no user action;
		// they still contribute to resolvedNames above so later
		// scans skip those entity names.
		if item.DetectionType != "wiki_link_resolved" {
			items = append(items, item)
		}
	}

	return items, resolvedNames
}

// scanUntaggedMentions searches the plain text for entity names that
// appear without wiki link markup. Each occurrence is checked against
// the wiki link ranges; occurrences that overlap a stripped wiki link
// are skipped so only genuinely untagged mentions are flagged.
func (a *Analyzer) scanUntaggedMentions(
	ctx context.Context,
	campaignID int64,
	plainText string,
	wikiRanges []wikiLinkRange,
) []models.ContentAnalysisItem {
	entities, err := a.db.ListEntitiesByCampaign(ctx, campaignID)
	if err != nil {
		log.Printf("analysis: failed to list entities for campaign %d: %v", campaignID, err)
		return nil
	}

	var items []models.ContentAnalysisItem
	lowerPlain := strings.ToLower(plainText)

	for _, entity := range entities {
		if len(entity.Name) < 3 {
			continue
		}

		lowerName := strings.ToLower(entity.Name)
		nameLen := len(lowerName)

		// Find ALL occurrences of this entity name in the
		// plain text, not just the first one.
		searchFrom := 0
		for searchFrom <= len(lowerPlain)-nameLen {
			idx := strings.Index(lowerPlain[searchFrom:], lowerName)
			if idx < 0 {
				break
			}
			idx += searchFrom
			end := idx + len(entity.Name)

			// Advance past this match for the next iteration.
			searchFrom = idx + 1

			// Skip if this occurrence overlaps with any wiki
			// link range in plain-text coordinates.
			if overlapsWikiRange(idx, end, wikiRanges) {
				continue
			}

			sim := 1.0
			entityID := entity.ID
			snippet := extractContextSnippet(plainText, idx, end)

			items = append(items, models.ContentAnalysisItem{
				DetectionType:  "untagged_mention",
				MatchedText:    plainText[idx:end],
				EntityID:       &entityID,
				Similarity:     &sim,
				ContextSnippet: &snippet,
				PositionStart:  &idx,
				PositionEnd:    &end,
				Resolution:     "pending",
			})
		}
	}

	return items
}

// overlapsWikiRange reports whether the text range [start, end)
// overlaps with any of the provided wiki link ranges.
func overlapsWikiRange(start, end int, ranges []wikiLinkRange) bool {
	for _, r := range ranges {
		if start < r.end && end > r.start {
			return true
		}
	}
	return false
}

// scanMisspellings extracts capitalized phrases from plain text and
// checks each against the entity list using fuzzy matching. Only
// results with similarity between the minimum threshold and the
// resolved threshold are kept. Matches are classified as
// "potential_alias" when the matched phrase is a substring of the
// entity name or vice versa, and as "misspelling" otherwise.
func (a *Analyzer) scanMisspellings(
	ctx context.Context,
	campaignID int64,
	plainText string,
	matchedNames map[string]bool,
	wikiRanges []wikiLinkRange,
) []models.ContentAnalysisItem {
	var items []models.ContentAnalysisItem

	candidates := capitalizedPhraseRe.FindAllStringIndex(plainText, -1)
	for _, loc := range candidates {
		if len(items) >= maxMisspellingCandidates {
			break
		}

		phrase := plainText[loc[0]:loc[1]]

		// Skip single-character results or phrases already
		// matched in earlier scans.
		if len(phrase) < 2 {
			continue
		}
		if matchedNames[strings.ToLower(phrase)] {
			continue
		}

		// Skip phrases that overlap with a stripped wiki link
		// range to avoid false matches across link boundaries.
		if overlapsWikiRange(loc[0], loc[1], wikiRanges) {
			continue
		}

		results, err := a.db.ResolveEntityByName(ctx, campaignID, phrase, 1)
		if err != nil {
			log.Printf("analysis: failed to resolve phrase %q: %v", phrase, err)
			continue
		}

		if len(results) == 0 {
			continue
		}

		r := results[0]

		// Skip fragments that cover less than half the entity name.
		// These are typically just the opening words of a longer name
		// (e.g., "Canticle of" matching "Canticle of Æternity").
		if float64(len(phrase)) < float64(len(r.Name))*0.5 {
			continue
		}

		if r.Similarity < similarityThresholdMinimum || r.Similarity >= similarityThresholdResolved {
			continue
		}

		snippet := extractContextSnippet(plainText, loc[0], loc[1])
		entityID := r.ID

		// Classify as alias or misspelling based on substring
		// relationship.
		detectionType := "misspelling"
		lowerPhrase := strings.ToLower(phrase)
		lowerEntityName := strings.ToLower(r.Name)
		if strings.Contains(lowerEntityName, lowerPhrase) || strings.Contains(lowerPhrase, lowerEntityName) {
			detectionType = "potential_alias"
		}

		items = append(items, models.ContentAnalysisItem{
			DetectionType:  detectionType,
			MatchedText:    phrase,
			EntityID:       &entityID,
			Similarity:     &r.Similarity,
			ContextSnippet: &snippet,
			PositionStart:  &loc[0],
			PositionEnd:    &loc[1],
			Resolution:     "pending",
		})
	}

	return items
}

// extractContextSnippet returns a substring of content surrounding the
// range [start, end), padded by up to contextRadius characters on each
// side. The returned string is clamped to content boundaries.
func extractContextSnippet(content string, start, end int) string {
	snippetStart := start - contextRadius
	if snippetStart < 0 {
		snippetStart = 0
	}

	snippetEnd := end + contextRadius
	if snippetEnd > len(content) {
		snippetEnd = len(content)
	}

	return content[snippetStart:snippetEnd]
}

// stripWikiLinksWithRanges replaces wiki link markup with plain
// display text and returns the stripped text plus position ranges
// of where each link's text sits in the output. For [[Name]] the
// display text is Name; for [[Name|Display]] it is Display.
func stripWikiLinksWithRanges(content string) (string, []wikiLinkRange) {
	var result strings.Builder
	var ranges []wikiLinkRange

	matches := wikiLinkRe.FindAllStringSubmatchIndex(content, -1)
	prev := 0

	for _, loc := range matches {
		// Append everything before this match.
		result.WriteString(content[prev:loc[0]])

		// Determine the display text for this wiki link.
		var displayText string
		if loc[4] >= 0 && loc[5] >= 0 && content[loc[4]:loc[5]] != "" {
			// [[Name|Display]] - use the display alias.
			displayText = content[loc[4]:loc[5]]
		} else {
			// [[Name]] - use the entity name.
			displayText = content[loc[2]:loc[3]]
		}

		// Record the range of this link's display text in the
		// output coordinate space.
		outStart := result.Len()
		result.WriteString(displayText)
		outEnd := result.Len()

		ranges = append(ranges, wikiLinkRange{
			start: outStart,
			end:   outEnd,
		})

		prev = loc[1]
	}

	// Append any trailing content after the last match.
	result.WriteString(content[prev:])

	return result.String(), ranges
}

// stripWikiLinks replaces wiki link markup with plain display text.
// [[Name]] becomes Name; [[Name|Display]] becomes Display.
func stripWikiLinks(content string) string {
	text, _ := stripWikiLinksWithRanges(content)
	return text
}

// buildMatchedNames collects the lowercased names of all entities that
// were matched in scans 1 and 2, so scan 3 can skip them.
func buildMatchedNames(
	resolvedNames map[string]bool,
	untaggedItems []models.ContentAnalysisItem,
) map[string]bool {
	matched := make(map[string]bool, len(resolvedNames)+len(untaggedItems))
	for name := range resolvedNames {
		matched[name] = true
	}
	for _, item := range untaggedItems {
		matched[strings.ToLower(item.MatchedText)] = true
	}
	return matched
}
