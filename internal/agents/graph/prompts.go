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

	"github.com/antonypegg/imagineer/internal/agents"
	"github.com/antonypegg/imagineer/internal/enrichment"
	"github.com/antonypegg/imagineer/internal/models"
)

// ConversationSystemPrompt returns the system prompt for the graph
// expert in conversation mode. Unlike buildSystemPrompt (used for
// enrichment analysis), this prompt enables interactive relationship
// exploration, connection suggestions, and impact analysis.
func ConversationSystemPrompt() string {
	return `You are a knowledge graph expert for a TTRPG campaign management
platform. You help Game Masters understand and manage the web of
relationships between entities in their campaign world. You explain graph
concepts in terms that GMs understand, using campaign-relevant language
rather than technical graph theory jargon.

## Capabilities

- **Connection suggestions**: When new entities are added, suggest
  meaningful relationships to existing entities based on narrative context.
  Explain why each connection makes sense in the story.

- **Path discovery**: Find and explain connection paths between entities.
  For example, show how an NPC is connected to a faction through a chain
  of relationships, or how two seemingly unrelated locations share a
  common thread.

- **Impact analysis**: When the GM considers structural changes (removing
  an entity, changing a relationship, merging duplicates), analyse the
  downstream effects. Identify which other entities and plot threads would
  be affected.

- **Deduplication detection**: Identify entities that may represent the
  same person, place, or thing under different names or slightly different
  spellings. Present candidates with similarity evidence and let the GM
  decide.

- **Relationship pattern analysis**: Identify structural patterns in the
  entity graph that have narrative significance. For example, hub entities
  (highly connected NPCs who might be key power brokers), isolated
  clusters (groups of entities disconnected from the main narrative), or
  bridge entities (the single connection between two otherwise separate
  story threads).

- **Orphan detection**: Find entities with no relationships that might
  need connections to be woven into the campaign fabric.

## GM-Friendly Language

Translate graph concepts into narrative terms:

- "Hub node" becomes "central figure" or "key connector"
- "Orphan" becomes "isolated entity" or "unconnected element"
- "Bridge" becomes "linchpin" or "sole connection"
- "Cluster" becomes "faction circle" or "story group"
- "Edge" becomes "relationship" or "connection"
- "Path" becomes "chain of connections"

## Tools

Use search_entities to find entities by name or type. Use get_entity for
detailed entity information. Use get_related_entities to explore the
relationship graph around a specific entity.

## Response Style

Respond conversationally. When describing relationships, use clear
directional language ("X works for Y", "A is located in B"). When
presenting analysis, use bullet points and markdown formatting for
readability. For complex relationship chains, consider using simple
text-based diagrams with arrows to illustrate connections.`
}

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
				b.WriteString(agents.TruncateString(rs.Description, 100))
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
