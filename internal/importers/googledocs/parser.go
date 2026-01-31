// Package googledocs provides an importer for Google Docs content.
package googledocs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/antonypegg/imagineer/internal/importers/common"
)

// Importer implements the common.Importer interface for Google Docs.
type Importer struct {
	httpClient *http.Client
}

// New creates a new Google Docs importer.
func New() *Importer {
	return &Importer{
		httpClient: &http.Client{},
	}
}

// Name returns the name of this importer.
func (i *Importer) Name() string {
	return "Google Docs"
}

// SupportedFormats returns the file formats this importer supports.
func (i *Importer) SupportedFormats() []string {
	return []string{"url", "text/plain"}
}

// Import reads content from a Google Docs URL or plain text and extracts entities.
func (i *Importer) Import(ctx context.Context, source io.Reader, options common.ImportOptions) (*common.ImportResult, error) {
	result := &common.ImportResult{
		Entities:      []common.ExtractedEntity{},
		Relationships: []common.ExtractedRelationship{},
		Events:        []common.ExtractedEvent{},
		Warnings:      []string{},
		Errors:        []string{},
	}

	// Read the content
	content, err := io.ReadAll(source)
	if err != nil {
		return nil, fmt.Errorf("failed to read content: %w", err)
	}

	text := string(content)

	// Check if it's a Google Docs URL
	if strings.HasPrefix(text, "https://docs.google.com/") {
		text, err = i.fetchGoogleDoc(ctx, strings.TrimSpace(text))
		if err != nil {
			return nil, err
		}
	}

	// Parse the content
	entities := i.extractEntities(text, options)
	result.Entities = entities

	if options.ExtractRelationships {
		result.Relationships = i.extractRelationships(text, entities)
	}

	if options.ExtractEvents {
		result.Events = i.extractEvents(text)
	}

	return result, nil
}

// fetchGoogleDoc fetches content from a Google Docs URL.
func (i *Importer) fetchGoogleDoc(ctx context.Context, docURL string) (string, error) {
	// Extract document ID from URL
	docID, err := extractDocID(docURL)
	if err != nil {
		return "", err
	}

	// Construct export URL for plain text
	exportURL := fmt.Sprintf("https://docs.google.com/document/d/%s/export?format=txt", docID)

	req, err := http.NewRequestWithContext(ctx, "GET", exportURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := i.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch document: status %d (document may not be publicly accessible)", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read document: %w", err)
	}

	return string(content), nil
}

// extractDocID extracts the document ID from a Google Docs URL.
func extractDocID(docURL string) (string, error) {
	u, err := url.Parse(docURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// URL format: https://docs.google.com/document/d/{docID}/...
	parts := strings.Split(u.Path, "/")
	for i, part := range parts {
		if part == "d" && i+1 < len(parts) {
			return parts[i+1], nil
		}
	}

	return "", fmt.Errorf("could not extract document ID from URL")
}

// extractEntities parses text content and extracts potential entities.
func (i *Importer) extractEntities(text string, options common.ImportOptions) []common.ExtractedEntity {
	entities := []common.ExtractedEntity{}

	// Split into sections by headers
	sections := splitIntoSections(text)

	for _, section := range sections {
		if section.Title == "" {
			continue
		}

		entityType := common.EntityTypeOther
		if options.AutoDetectEntities {
			entityType = detectEntityTypeFromText(section.Title, section.Content)
		}

		entity := common.ExtractedEntity{
			Name:        section.Title,
			Type:        entityType,
			Description: section.Content,
			Attributes:  extractAttributes(section.Content),
			Tags:        extractTags(section.Content),
			SourceDoc:   options.SourceDocument,
		}

		entities = append(entities, entity)
	}

	return entities
}

// Section represents a section of text with a title.
type Section struct {
	Title   string
	Content string
}

// splitIntoSections splits text into sections based on headers.
func splitIntoSections(text string) []Section {
	sections := []Section{}

	// Match markdown-style headers or all-caps lines
	headerPattern := regexp.MustCompile(`(?m)^(#{1,3}\s+(.+)|([A-Z][A-Z\s]{5,}))$`)
	matches := headerPattern.FindAllStringSubmatchIndex(text, -1)

	if len(matches) == 0 {
		// No headers found, treat entire document as one section
		return []Section{{Title: "Imported Content", Content: strings.TrimSpace(text)}}
	}

	for i, match := range matches {
		// Extract header text
		headerStart := match[0]
		headerEnd := match[1]
		header := strings.TrimSpace(text[headerStart:headerEnd])
		header = strings.TrimLeft(header, "# ")

		// Get content until next header or end of document
		contentStart := headerEnd
		var contentEnd int
		if i+1 < len(matches) {
			contentEnd = matches[i+1][0]
		} else {
			contentEnd = len(text)
		}

		content := strings.TrimSpace(text[contentStart:contentEnd])

		sections = append(sections, Section{
			Title:   header,
			Content: content,
		})
	}

	return sections
}

// detectEntityTypeFromText attempts to determine entity type from text content.
func detectEntityTypeFromText(title, content string) common.EntityType {
	titleLower := strings.ToLower(title)
	contentLower := strings.ToLower(content)
	combined := titleLower + " " + contentLower

	// NPC indicators
	if strings.Contains(combined, "occupation:") ||
		strings.Contains(combined, "age:") ||
		strings.Contains(combined, "appearance:") ||
		strings.Contains(titleLower, "npc") ||
		regexp.MustCompile(`(?i)^(dr\.|mr\.|mrs\.|ms\.|professor|inspector|captain|lord|lady)`).MatchString(title) {
		return common.EntityTypeNPC
	}

	// Location indicators
	if strings.Contains(combined, "located at") ||
		strings.Contains(combined, "address:") ||
		strings.Contains(titleLower, "location") ||
		regexp.MustCompile(`(?i)(hotel|inn|tavern|mansion|house|street|road|city|town|village)`).MatchString(title) {
		return common.EntityTypeLocation
	}

	// Clue indicators
	if strings.Contains(titleLower, "clue") ||
		strings.Contains(combined, "evidence:") ||
		strings.Contains(combined, "found at:") {
		return common.EntityTypeClue
	}

	// Faction indicators
	if strings.Contains(titleLower, "faction") ||
		strings.Contains(titleLower, "cult") ||
		strings.Contains(titleLower, "organization") ||
		strings.Contains(combined, "members:") {
		return common.EntityTypeFaction
	}

	// Item indicators
	if strings.Contains(titleLower, "item") ||
		strings.Contains(titleLower, "artifact") ||
		strings.Contains(combined, "properties:") {
		return common.EntityTypeItem
	}

	// Creature indicators
	if strings.Contains(titleLower, "creature") ||
		strings.Contains(titleLower, "monster") ||
		strings.Contains(combined, "sanity loss:") {
		return common.EntityTypeCreature
	}

	return common.EntityTypeOther
}

// extractAttributes parses content for structured attributes.
func extractAttributes(content string) map[string]interface{} {
	attrs := make(map[string]interface{})

	// Match "Key: Value" patterns
	pattern := regexp.MustCompile(`(?m)^([A-Za-z][A-Za-z\s]{1,20}):\s*(.+)$`)
	matches := pattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		key := strings.TrimSpace(match[1])
		value := strings.TrimSpace(match[2])

		// Convert key to camelCase
		key = toCamelCase(key)

		// Try to parse as JSON for complex values
		var jsonValue interface{}
		if err := json.Unmarshal([]byte(value), &jsonValue); err == nil {
			attrs[key] = jsonValue
		} else {
			attrs[key] = value
		}
	}

	return attrs
}

// extractTags finds tag-like content in text.
func extractTags(content string) []string {
	tags := []string{}

	// Match hashtags
	hashtagPattern := regexp.MustCompile(`#([A-Za-z][A-Za-z0-9_]+)`)
	matches := hashtagPattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		tags = append(tags, match[1])
	}

	return tags
}

// extractRelationships finds potential relationships between entities.
func (i *Importer) extractRelationships(text string, entities []common.ExtractedEntity) []common.ExtractedRelationship {
	relationships := []common.ExtractedRelationship{}

	// Create a map of entity names for quick lookup
	entityNames := make(map[string]bool)
	for _, e := range entities {
		entityNames[strings.ToLower(e.Name)] = true
	}

	// Common relationship patterns
	patterns := []struct {
		pattern string
		relType string
	}{
		{`(\w+)\s+knows\s+(\w+)`, "knows"},
		{`(\w+)\s+works\s+(?:for|with)\s+(\w+)`, "works_with"},
		{`(\w+)\s+(?:is\s+)?located\s+(?:at|in)\s+(\w+)`, "located_at"},
		{`(\w+)\s+owns\s+(\w+)`, "owns"},
		{`(\w+)\s+(?:is\s+)?(?:a\s+)?member\s+of\s+(\w+)`, "member_of"},
	}

	for _, p := range patterns {
		re := regexp.MustCompile(`(?i)` + p.pattern)
		matches := re.FindAllStringSubmatch(text, -1)

		for _, match := range matches {
			if len(match) >= 3 {
				source := match[1]
				target := match[2]

				// Only include if both entities exist
				if entityNames[strings.ToLower(source)] && entityNames[strings.ToLower(target)] {
					relationships = append(relationships, common.ExtractedRelationship{
						SourceName:       source,
						TargetName:       target,
						RelationshipType: p.relType,
					})
				}
			}
		}
	}

	return relationships
}

// extractEvents finds potential timeline events in text.
func (i *Importer) extractEvents(text string) []common.ExtractedEvent {
	events := []common.ExtractedEvent{}

	// Match date patterns
	datePatterns := []struct {
		pattern   string
		precision string
	}{
		{`(\d{4}-\d{2}-\d{2})\s*[:\-]\s*(.+)`, "exact"},
		{`(\w+\s+\d{1,2},?\s+\d{4})\s*[:\-]\s*(.+)`, "exact"},
		{`(\w+\s+\d{4})\s*[:\-]\s*(.+)`, "month"},
		{`(\d{4})\s*[:\-]\s*(.+)`, "year"},
	}

	for _, p := range datePatterns {
		re := regexp.MustCompile(`(?m)` + p.pattern)
		matches := re.FindAllStringSubmatch(text, -1)

		for _, match := range matches {
			if len(match) >= 3 {
				events = append(events, common.ExtractedEvent{
					Date:          match[1],
					DatePrecision: p.precision,
					Description:   strings.TrimSpace(match[2]),
				})
			}
		}
	}

	return events
}

// toCamelCase converts a string to camelCase.
func toCamelCase(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if i == 0 {
			words[i] = strings.ToLower(word)
		} else {
			words[i] = strings.Title(strings.ToLower(word))
		}
	}
	return strings.Join(words, "")
}
