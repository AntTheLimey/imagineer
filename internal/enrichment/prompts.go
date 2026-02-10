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
	"fmt"
	"strings"

	"github.com/antonypegg/imagineer/internal/models"
)

// maxContentChars is the default maximum number of characters to include
// from the source content in the user prompt.
const maxContentChars = 4000

// buildSystemPrompt returns the system prompt for the enrichment LLM
// call. It instructs the model to act as a TTRPG campaign analyst and
// return structured JSON.
func buildSystemPrompt() string {
	return `You are a TTRPG campaign analyst assistant. Your job is to analyse
session notes, chapter content, and other campaign writing to suggest
enrichments for campaign entities (NPCs, locations, items, factions, etc.).

Given a piece of campaign content and the current state of an entity that
appears in it, you must produce structured JSON suggesting:

1. **Description updates** - improvements or additions to the entity's
   description based on new information revealed in the content.
2. **Log entries** - chronological event entries that should be added to
   the entity's history log based on what happened in the content.
3. **Relationships** - connections between this entity and other entities
   mentioned in the same content.

Rules:
- Only suggest changes supported by the provided content.
- Do not invent information not present in the source material.
- Keep descriptions concise and in a style consistent with TTRPG notes.
- For relationships, use descriptive types like "ally_of", "enemy_of",
  "located_in", "member_of", "owns", "works_for", "knows", etc.
- If the content does not reveal new information about the entity, return
  empty arrays.
- Do not duplicate existing relationships listed in the input.

You MUST respond with valid JSON only. No markdown, no commentary outside
the JSON object.

Response format:
{
  "descriptionUpdates": [
    {
      "currentDescription": "the entity's current description",
      "suggestedDescription": "an improved description incorporating new info",
      "rationale": "brief explanation of what changed and why"
    }
  ],
  "logEntries": [
    {
      "content": "what happened to or involving this entity",
      "occurredAt": "optional in-game date or time reference"
    }
  ],
  "relationships": [
    {
      "sourceEntityId": 123,
      "sourceEntityName": "Source Entity",
      "targetEntityId": 456,
      "targetEntityName": "Target Entity",
      "relationshipType": "type_of_relationship",
      "description": "brief description of the relationship"
    }
  ]
}`
}

// buildUserPrompt constructs the user prompt from the enrichment input.
// It includes the source content (truncated around entity mentions),
// the entity's current state, existing relationships, and a list of
// other entities mentioned in the same content.
func buildUserPrompt(input EnrichmentInput) string {
	var b strings.Builder

	// Source content, truncated around entity mentions.
	truncated := truncateContent(input.Content, input.Entity.Name, maxContentChars)
	fmt.Fprintf(&b, "## Source Content\n\n%s\n\n", truncated)

	// Entity current state.
	fmt.Fprintf(&b, "## Entity to Enrich\n\n")
	fmt.Fprintf(&b, "- **ID**: %d\n", input.Entity.ID)
	fmt.Fprintf(&b, "- **Name**: %s\n", input.Entity.Name)
	fmt.Fprintf(&b, "- **Type**: %s\n", string(input.Entity.EntityType))
	if input.Entity.Description != nil && *input.Entity.Description != "" {
		fmt.Fprintf(&b, "- **Current Description**: %s\n", *input.Entity.Description)
	} else {
		fmt.Fprintf(&b, "- **Current Description**: (none)\n")
	}
	b.WriteString("\n")

	// Existing relationships.
	if len(input.Relationships) > 0 {
		fmt.Fprintf(&b, "## Existing Relationships\n\n")
		for _, rel := range input.Relationships {
			desc := ""
			if rel.Description != nil {
				desc = *rel.Description
			}

			sourceName := fmt.Sprintf("Entity %d", rel.SourceEntityID)
			targetName := fmt.Sprintf("Entity %d", rel.TargetEntityID)
			if rel.SourceEntity != nil {
				sourceName = rel.SourceEntity.Name
			}
			if rel.TargetEntity != nil {
				targetName = rel.TargetEntity.Name
			}

			fmt.Fprintf(&b, "- %s -[%s]-> %s",
				sourceName, rel.RelationshipType, targetName)
			if desc != "" {
				fmt.Fprintf(&b, " (%s)", desc)
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Other entities mentioned in the same content.
	if len(input.OtherEntities) > 0 {
		fmt.Fprintf(&b, "## Other Entities in This Content\n\n")
		for _, ent := range input.OtherEntities {
			fmt.Fprintf(&b, "- **%s** (ID: %d, Type: %s)\n",
				ent.Name, ent.ID, string(ent.EntityType))
		}
		b.WriteString("\n")
	}

	b.WriteString("Analyse the source content and produce enrichment ")
	b.WriteString("suggestions for the entity above. Respond with JSON only.")

	return b.String()
}

// truncateContent truncates content to approximately maxChars runes
// centred around entity mentions. If the content is shorter than
// maxChars, it is returned unchanged. When truncation occurs, a window
// is placed around the first mention of the entity name and the result
// is prefixed/suffixed with "[...]" markers.
//
// All length comparisons and offsets use rune counts so that multi-byte
// Unicode characters (e.g., CJK, accented Latin) are never split.
func truncateContent(content string, entityName string, maxChars int) string {
	runes := []rune(content)
	if len(runes) <= maxChars {
		return content
	}

	// Find the first mention of the entity name (case-insensitive).
	// strings.Index returns a byte offset, so convert to rune offset.
	lowerContent := strings.ToLower(content)
	lowerName := strings.ToLower(entityName)
	byteIdx := strings.Index(lowerContent, lowerName)

	if byteIdx < 0 {
		// Entity name not found; take from the start.
		return string(runes[:maxChars]) + "\n\n[...]"
	}

	// Convert byte offset to rune offset.
	idx := len([]rune(content[:byteIdx]))

	// Centre a window of maxChars runes around the mention.
	half := maxChars / 2
	start := idx - half
	end := idx + half

	if start < 0 {
		end += -start
		start = 0
	}
	if end > len(runes) {
		start -= end - len(runes)
		end = len(runes)
		if start < 0 {
			start = 0
		}
	}

	var result strings.Builder
	if start > 0 {
		result.WriteString("[...]\n\n")
	}
	result.WriteString(string(runes[start:end]))
	if end < len(runes) {
		result.WriteString("\n\n[...]")
	}

	return result.String()
}

// buildNewEntityDetectionSystemPrompt returns the system prompt for the
// new-entity detection LLM call. It instructs the model to identify named
// entities in content that are not already in the campaign database.
func buildNewEntityDetectionSystemPrompt() string {
	return `You are a TTRPG campaign analyst. Analyse content to identify
named entities (NPCs, locations, items, factions, creatures, organizations,
events, documents, clues) mentioned but NOT in the campaign database.

Rules:
- Only identify proper nouns and clear named entities.
- Do NOT identify generic references like "the tavern", "a guard",
  "the stranger", or "some soldiers".
- Only identify entities that are clearly distinct from any entity in the
  known entities list.

Supported entity types:
  npc, location, item, faction, clue, creature, organization, event,
  document, other

You MUST respond with valid JSON only. No markdown, no commentary outside
the JSON object.

Response format:
{
  "new_entities": [
    {
      "name": "Inspector Barrington",
      "entity_type": "npc",
      "description": "A Scotland Yard detective mentioned in the chapter",
      "reasoning": "Named character appearing in paragraph 3 who is not in the known entities list"
    }
  ]
}

If no new entities are found, return:
{"new_entities": []}`
}

// buildNewEntityDetectionUserPrompt constructs the user prompt for new-
// entity detection. It includes the source content (truncated to
// maxContentChars from the start) and a list of known entity names and
// types so the LLM can avoid suggesting entities already in the database.
func buildNewEntityDetectionUserPrompt(
	sourceContent string,
	knownEntities []models.Entity,
) string {
	var b strings.Builder

	// Source content, truncated from the start if too long.
	runes := []rune(sourceContent)
	if len(runes) > maxContentChars {
		runes = runes[:maxContentChars]
	}
	fmt.Fprintf(&b, "## Source Content\n\n%s\n\n", string(runes))

	// Known entities already in the campaign database.
	if len(knownEntities) > 0 {
		fmt.Fprintf(&b, "## Known Entities (already in database)\n\n")
		for _, ent := range knownEntities {
			fmt.Fprintf(&b, "- %s (%s)\n",
				ent.Name, string(ent.EntityType))
		}
		b.WriteString("\n")
	}

	b.WriteString("Identify named entities in the content above that ")
	b.WriteString("are NOT in the known entities list. ")
	b.WriteString("Respond with JSON only.")

	return b.String()
}
