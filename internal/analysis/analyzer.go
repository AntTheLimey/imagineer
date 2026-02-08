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

// capitalizedPhraseRe matches capitalized word sequences (2-4 words)
// where the first word starts with an uppercase letter.
var capitalizedPhraseRe = regexp.MustCompile(`[A-Z][a-zA-Z'-]*(?:\s+[A-Za-z][a-zA-Z'-]*){0,3}`)

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
const similarityThresholdMinimum = 0.4

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
		// link markup removed, used for scans 2 and 3.
		plainText := stripWikiLinks(content)

		// Scan 2: detect untagged entity name mentions in plain
		// text.
		untaggedItems := a.scanUntaggedMentions(ctx, campaignID, plainText, resolvedNames)
		items = append(items, untaggedItems...)

		// Scan 3: detect possible misspellings of entity names.
		matchedNames := buildMatchedNames(resolvedNames, untaggedItems)
		misspellingItems := a.scanMisspellings(ctx, campaignID, plainText, matchedNames)
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

		items = append(items, item)
	}

	return items, resolvedNames
}

// scanUntaggedMentions searches the plain text for entity names that
// appear without wiki link markup. Entities already resolved as wiki
// links are excluded.
func (a *Analyzer) scanUntaggedMentions(
	ctx context.Context,
	campaignID int64,
	plainText string,
	resolvedNames map[string]bool,
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
		if resolvedNames[strings.ToLower(entity.Name)] {
			continue
		}

		lowerName := strings.ToLower(entity.Name)
		idx := strings.Index(lowerPlain, lowerName)
		if idx < 0 {
			continue
		}

		end := idx + len(entity.Name)
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

	return items
}

// scanMisspellings extracts capitalized phrases from plain text and
// checks each against the entity list using fuzzy matching. Only
// results with similarity between the minimum threshold and the
// resolved threshold are kept.
func (a *Analyzer) scanMisspellings(
	ctx context.Context,
	campaignID int64,
	plainText string,
	matchedNames map[string]bool,
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

		results, err := a.db.ResolveEntityByName(ctx, campaignID, phrase, 1)
		if err != nil {
			log.Printf("analysis: failed to resolve phrase %q: %v", phrase, err)
			continue
		}

		if len(results) == 0 {
			continue
		}

		r := results[0]
		if r.Similarity < similarityThresholdMinimum || r.Similarity >= similarityThresholdResolved {
			continue
		}

		snippet := extractContextSnippet(plainText, loc[0], loc[1])
		entityID := r.ID

		items = append(items, models.ContentAnalysisItem{
			DetectionType:  "misspelling",
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

// stripWikiLinks replaces wiki link markup with plain display text.
// [[Name]] becomes Name; [[Name|Display]] becomes Display.
func stripWikiLinks(content string) string {
	return wikiLinkRe.ReplaceAllStringFunc(content, func(match string) string {
		sub := wikiLinkRe.FindStringSubmatch(match)
		if len(sub) >= 3 && sub[2] != "" {
			return sub[2]
		}
		return sub[1]
	})
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

