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
	"context"
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/antonypegg/imagineer/internal/importers/common"
)

// Importer provides methods to import notes from the local Evernote
// application on macOS.
type Importer struct {
	executor *AppleScriptExecutor
}

// New creates a new Evernote local importer with default settings.
func New() *Importer {
	return &Importer{
		executor: NewAppleScriptExecutor(),
	}
}

// NewWithTimeout creates a new Evernote local importer with a custom timeout.
func NewWithTimeout(timeout time.Duration) *Importer {
	return &Importer{
		executor: NewAppleScriptExecutorWithTimeout(timeout),
	}
}

// Name returns the name of this importer.
func (i *Importer) Name() string {
	return "Evernote Local"
}

// CheckStatus checks if Evernote is available and running.
func (i *Importer) CheckStatus(ctx context.Context) EvernoteStatus {
	return i.executor.CheckEvernoteStatus(ctx)
}

// ListNotebooks retrieves all notebooks from the local Evernote application.
func (i *Importer) ListNotebooks(ctx context.Context) ([]Notebook, error) {
	return i.executor.ListNotebooks(ctx)
}

// ListNotesInNotebook retrieves all notes from a specific notebook.
func (i *Importer) ListNotesInNotebook(ctx context.Context, notebookName string) ([]NoteSummary, error) {
	return i.executor.ListNotesInNotebook(ctx, notebookName)
}

// GetNoteContent retrieves the full content of a note by its note link.
func (i *Importer) GetNoteContent(ctx context.Context, noteLink string) (*NoteContent, error) {
	return i.executor.GetNoteContent(ctx, noteLink)
}

// ImportNotes imports all notes from a notebook and returns an ImportResult.
func (i *Importer) ImportNotes(ctx context.Context, notebookName string, options common.ImportOptions) (*common.ImportResult, error) {
	result := &common.ImportResult{
		Entities:      []common.ExtractedEntity{},
		Relationships: []common.ExtractedRelationship{},
		Events:        []common.ExtractedEvent{},
		Warnings:      []string{},
		Errors:        []string{},
	}

	// Get all notes in the notebook
	notes, err := i.executor.ListNotesInNotebook(ctx, notebookName)
	if err != nil {
		return nil, err
	}

	// Process each note
	for _, noteSummary := range notes {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		// Get full note content
		noteContent, err := i.executor.GetNoteContent(ctx, noteSummary.NoteLink)
		if err != nil {
			result.Warnings = append(result.Warnings,
				"Failed to get content for note '"+noteSummary.Title+"': "+err.Error())
			continue
		}

		entity, err := i.processNote(noteContent, options)
		if err != nil {
			result.Warnings = append(result.Warnings,
				"Error processing note '"+noteSummary.Title+"': "+err.Error())
			continue
		}

		if entity != nil {
			result.Entities = append(result.Entities, *entity)
		}
	}

	return result, nil
}

// processNote converts a note to an extracted entity.
func (i *Importer) processNote(note *NoteContent, options common.ImportOptions) (*common.ExtractedEntity, error) {
	// Extract plain text from the HTML content
	content := extractTextFromHTML(note.HTMLContent)

	// Determine entity type from tags or content
	entityType := common.EntityTypeOther
	if options.AutoDetectEntities {
		entityType = detectEntityType(note.Title, content, note.Tags)
	}

	// Build attributes
	attributes := make(map[string]interface{})
	if !note.Created.IsZero() {
		attributes["importedCreated"] = note.Created.Format(time.RFC3339)
	}
	if !note.Modified.IsZero() {
		attributes["importedUpdated"] = note.Modified.Format(time.RFC3339)
	}

	return &common.ExtractedEntity{
		Name:        note.Title,
		Type:        entityType,
		Description: content,
		Attributes:  attributes,
		Tags:        note.Tags,
		SourceDoc:   options.SourceDocument,
	}, nil
}

// extractTextFromHTML strips HTML tags and extracts plain text.
// This is a simplified version that handles common Evernote HTML patterns.
func extractTextFromHTML(htmlContent string) string {
	// Remove CDATA wrapper if present
	content := strings.TrimPrefix(htmlContent, "<![CDATA[")
	content = strings.TrimSuffix(content, "]]>")

	// Remove XML declaration
	content = regexp.MustCompile(`<\?xml[^?]*\?>`).ReplaceAllString(content, "")

	// Remove DOCTYPE
	content = regexp.MustCompile(`<!DOCTYPE[^>]*>`).ReplaceAllString(content, "")

	// Convert common HTML entities
	content = html.UnescapeString(content)

	// Replace line breaks with newlines
	content = regexp.MustCompile(`<br\s*/?>|</div>|</p>`).ReplaceAllString(content, "\n")

	// Remove remaining HTML tags
	content = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(content, "")

	// Clean up whitespace
	content = regexp.MustCompile(`\n{3,}`).ReplaceAllString(content, "\n\n")
	content = strings.TrimSpace(content)

	return content
}

// detectEntityType attempts to determine the entity type from content and tags.
// This logic mirrors the entity detection in the evernote package parser.
func detectEntityType(title, content string, tags []string) common.EntityType {
	titleLower := strings.ToLower(title)
	contentLower := strings.ToLower(content)

	// Check tags first
	for _, tag := range tags {
		tagLower := strings.ToLower(tag)
		switch {
		case strings.Contains(tagLower, "npc") || strings.Contains(tagLower, "character"):
			return common.EntityTypeNPC
		case strings.Contains(tagLower, "location") || strings.Contains(tagLower, "place"):
			return common.EntityTypeLocation
		case strings.Contains(tagLower, "item") || strings.Contains(tagLower, "artifact"):
			return common.EntityTypeItem
		case strings.Contains(tagLower, "faction") || strings.Contains(tagLower, "organization"):
			return common.EntityTypeFaction
		case strings.Contains(tagLower, "clue") || strings.Contains(tagLower, "evidence"):
			return common.EntityTypeClue
		case strings.Contains(tagLower, "creature") || strings.Contains(tagLower, "monster"):
			return common.EntityTypeCreature
		}
	}

	// Check title patterns
	npcPatterns := []string{"professor", "doctor", "mr.", "mrs.", "ms.", "dr.", "inspector"}
	for _, pattern := range npcPatterns {
		if strings.Contains(titleLower, pattern) {
			return common.EntityTypeNPC
		}
	}

	locationPatterns := []string{"hotel", "house", "mansion", "inn", "tavern", "street", "road", "city", "town"}
	for _, pattern := range locationPatterns {
		if strings.Contains(titleLower, pattern) {
			return common.EntityTypeLocation
		}
	}

	// Check content for entity indicators
	if strings.Contains(contentLower, "occupation:") || strings.Contains(contentLower, "age:") {
		return common.EntityTypeNPC
	}

	return common.EntityTypeOther
}
