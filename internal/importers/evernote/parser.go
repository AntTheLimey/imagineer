// Package evernote provides an importer for Evernote .enex export files.
package evernote

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/antonypegg/imagineer/internal/importers/common"
)

// Note represents a single note in an Evernote export.
type Note struct {
	Title   string `xml:"title"`
	Content string `xml:"content"`
	Created string `xml:"created"`
	Updated string `xml:"updated"`
	Tags    []Tag  `xml:"tag"`
}

// Tag represents a tag on an Evernote note.
type Tag struct {
	Name string `xml:",chardata"`
}

// Export represents the root element of an Evernote .enex file.
type Export struct {
	XMLName xml.Name `xml:"en-export"`
	Notes   []Note   `xml:"note"`
}

// Importer implements the common.Importer interface for Evernote files.
type Importer struct{}

// New creates a new Evernote importer.
func New() *Importer {
	return &Importer{}
}

// Name returns the name of this importer.
func (i *Importer) Name() string {
	return "Evernote"
}

// SupportedFormats returns the file formats this importer supports.
func (i *Importer) SupportedFormats() []string {
	return []string{".enex"}
}

// Import reads an Evernote .enex file and extracts entities.
func (i *Importer) Import(ctx context.Context, source io.Reader, options common.ImportOptions) (*common.ImportResult, error) {
	result := &common.ImportResult{
		Entities:      []common.ExtractedEntity{},
		Relationships: []common.ExtractedRelationship{},
		Events:        []common.ExtractedEvent{},
		Warnings:      []string{},
		Errors:        []string{},
	}

	// Parse the XML
	var export Export
	decoder := xml.NewDecoder(source)
	if err := decoder.Decode(&export); err != nil {
		return nil, fmt.Errorf("failed to parse Evernote file: %w", err)
	}

	// Process each note
	for _, note := range export.Notes {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		entity, err := i.processNote(note, options)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Error processing note '%s': %v", note.Title, err))
			continue
		}

		if entity != nil {
			result.Entities = append(result.Entities, *entity)
		}
	}

	return result, nil
}

// processNote converts an Evernote note to an extracted entity.
func (i *Importer) processNote(note Note, options common.ImportOptions) (*common.ExtractedEntity, error) {
	// Extract plain text from the HTML content
	content := extractTextFromHTML(note.Content)

	// Determine entity type from tags or content
	entityType := common.EntityTypeOther
	if options.AutoDetectEntities {
		entityType = detectEntityType(note.Title, content, note.Tags)
	}

	// Collect tags
	tags := make([]string, len(note.Tags))
	for i, tag := range note.Tags {
		tags[i] = tag.Name
	}

	// Build attributes
	attributes := make(map[string]interface{})
	if note.Created != "" {
		if t, err := parseEvernoteDate(note.Created); err == nil {
			attributes["importedCreated"] = t.Format(time.RFC3339)
		}
	}
	if note.Updated != "" {
		if t, err := parseEvernoteDate(note.Updated); err == nil {
			attributes["importedUpdated"] = t.Format(time.RFC3339)
		}
	}

	return &common.ExtractedEntity{
		Name:        note.Title,
		Type:        entityType,
		Description: content,
		Attributes:  attributes,
		Tags:        tags,
		SourceDoc:   options.SourceDocument,
	}, nil
}

// extractTextFromHTML strips HTML tags and extracts plain text.
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

// detectEntityType attempts to determine the entity type from content.
func detectEntityType(title, content string, tags []Tag) common.EntityType {
	titleLower := strings.ToLower(title)
	contentLower := strings.ToLower(content)

	// Check tags first
	for _, tag := range tags {
		tagLower := strings.ToLower(tag.Name)
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

// parseEvernoteDate parses an Evernote date string (format: 20231215T143022Z).
func parseEvernoteDate(dateStr string) (time.Time, error) {
	return time.Parse("20060102T150405Z", dateStr)
}
