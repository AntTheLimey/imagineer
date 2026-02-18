/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package graph

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/antonypegg/imagineer/internal/enrichment"
	"github.com/antonypegg/imagineer/internal/models"
)

// buildSystemPrompt returns the system prompt instructing the LLM to
// act as a knowledge graph analyst for TTRPG campaigns. The LLM
// identifies redundant and implied relationships.
func buildSystemPrompt() string {
	return `You are a knowledge graph analyst for TTRPG campaigns. Your role is to
review existing and proposed relationships between campaign entities and
identify structural issues in the graph.

## Finding Types

- **redundant_edge**: Two edges between the same entity pair convey the
  same meaning through different type names. For example, "works_for" and
  "employed_by" between the same NPC and organization are redundant.

- **implied_edge**: A relationship that exists as a traversal through
  intermediate entities does not need a direct edge. For example, if NPC
  Alice leads Faction X and NPC Bob belongs to Faction X, an
  "associated_with" edge between Alice and Bob is implied and unnecessary.

## Rules

1. Only flag genuine graph quality issues. New relationships that add
   independent meaning are NOT redundant.
2. Consider directionality. A -> B and B -> A via an inverse type are
   the same relationship stored once, not a redundancy.
3. Be conservative. When in doubt, do NOT flag a finding.
4. Return an empty findings array if the graph is clean.

## Output Format

Respond with ONLY valid JSON in the following structure:

{
  "findings": [
    {
      "findingType": "redundant_edge|implied_edge",
      "description": "Clear description of the issue",
      "involvedEntities": ["Entity A", "Entity B"],
      "suggestion": "How to resolve the issue"
    }
  ]
}

Do not include any text outside the JSON object.`
}

// buildUserPrompt constructs the user prompt from the pipeline input,
// including existing relationships, proposed new relationships from
// enrichment suggestions, and the entity list with types.
func buildUserPrompt(
	input enrichment.PipelineInput,
	relSuggestions []models.ContentAnalysisItem,
) string {
	var b strings.Builder

	// Include existing relationships.
	if len(input.Relationships) > 0 {
		b.WriteString("## Existing Relationships\n\n")
		b.WriteString("These relationships already exist in the ")
		b.WriteString("campaign graph:\n\n")
		for _, rel := range input.Relationships {
			b.WriteString("- ")
			b.WriteString(rel.SourceEntityName)
			b.WriteString(" (")
			b.WriteString(rel.SourceEntityType)
			b.WriteString(") --[")
			if rel.DisplayLabel != "" {
				b.WriteString(rel.DisplayLabel)
			} else {
				b.WriteString(rel.RelationshipTypeName)
			}
			b.WriteString("]--> ")
			b.WriteString(rel.TargetEntityName)
			b.WriteString(" (")
			b.WriteString(rel.TargetEntityType)
			b.WriteString(")\n")
		}
	}

	// Include proposed new relationships from enrichment.
	if len(relSuggestions) > 0 {
		b.WriteString("\n## Proposed New Relationships\n\n")
		b.WriteString("The enrichment agent has suggested these new ")
		b.WriteString("relationships. Check whether any are redundant ")
		b.WriteString("with existing edges or implied by traversals:\n\n")
		for _, item := range relSuggestions {
			var rs models.RelationshipSuggestion
			if err := json.Unmarshal(
				item.SuggestedContent, &rs,
			); err != nil {
				log.Printf(
					"graph-expert: failed to unmarshal suggestion "+
						"for prompt: %v", err,
				)
				continue
			}
			b.WriteString("- ")
			b.WriteString(rs.SourceEntityName)
			b.WriteString(" --[")
			b.WriteString(rs.RelationshipType)
			b.WriteString("]--> ")
			b.WriteString(rs.TargetEntityName)
			if rs.Description != "" {
				b.WriteString(" (")
				b.WriteString(truncateString(rs.Description, 100))
				b.WriteString(")")
			}
			b.WriteString("\n")
		}
	}

	// Include entity list with types for context.
	if len(input.Entities) > 0 {
		b.WriteString("\n## Campaign Entities\n\n")
		for _, entity := range input.Entities {
			b.WriteString("- **")
			b.WriteString(entity.Name)
			b.WriteString("** (")
			b.WriteString(string(entity.EntityType))
			b.WriteString(", ID: ")
			b.WriteString(fmt.Sprintf("%d", entity.ID))
			b.WriteString(")\n")
		}
	}

	return b.String()
}

// truncateString truncates a string to the specified maximum length,
// appending an ellipsis if truncation occurs.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
