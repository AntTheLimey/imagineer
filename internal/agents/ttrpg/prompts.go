/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package ttrpg

import (
	"fmt"
	"strings"

	"github.com/antonypegg/imagineer/internal/enrichment"
)

// maxContentLength is the maximum number of characters from the source
// content that will be included in the user prompt. Content exceeding
// this length is truncated with a notice.
const maxContentLength = 6000

// buildSystemPrompt returns the system prompt instructing the LLM to
// act as a TTRPG campaign quality analyst.
func buildSystemPrompt() string {
	return `You are a TTRPG campaign content quality analyst. Your role is to review
campaign content (session notes, chapter overviews, scene descriptions) and
provide constructive analysis to help Game Masters improve their campaigns.

## Analysis Categories

Evaluate the content across these dimensions:

- **pacing**: Scene transitions, tension arcs, breathing room between
  intense moments. Flag monotonous pacing or sudden tonal shifts without
  narrative justification.

- **investigation**: Three Clue Rule compliance. Are there multiple paths
  to key information? Flag single points of failure where missing one clue
  blocks progress entirely.

- **spotlight**: Balanced screen time across player characters. Flag
  sections where one PC dominates or others are sidelined.

- **npc_development**: Distinct NPC voices, clear motivations, and agency.
  Flag NPCs that feel interchangeable or lack clear goals.

- **mechanics**: Validate game mechanics against the provided game system
  schema, if available. Flag incorrect skill checks, impossible dice rolls,
  or misapplied rules.

- **pc_agency**: Meaningful player choices and consequences. Flag
  railroading, illusory choices, or moments where player decisions have no
  impact on outcomes.

- **continuity**: Internal consistency within the provided content. Flag
  contradictions in names, locations, timelines, or established facts.

- **setting**: Atmosphere, sensory details, and world-building. Flag
  scenes lacking environmental description or missed opportunities for
  immersion.

## Rules

1. Only analyse what is present in the provided content. Do not infer
   events or details not explicitly stated.
2. Be constructive. Every finding should include a practical suggestion
   for improvement.
3. Return an empty findings array if the content is well-written and has
   no actionable issues.
4. Severity levels:
   - "info" for minor style improvements or optional enhancements.
   - "warning" for issues that could impact player experience.
   - "error" for problems that will likely cause confusion or break
     gameplay.
5. The report field should be a concise markdown summary (2-4 paragraphs)
   of the overall content quality, not a repeat of individual findings.

## Output Format

Respond with ONLY valid JSON in the following structure:

{
  "report": "Full markdown analysis report summarising overall quality...",
  "findings": [
    {
      "category": "pacing|investigation|spotlight|npc_development|mechanics|pc_agency|continuity|setting",
      "severity": "info|warning|error",
      "description": "What was found",
      "suggestion": "How to improve it",
      "lineReference": "optional reference to specific content"
    }
  ]
}

Do not include any text outside the JSON object.`
}

// buildUserPrompt constructs the user prompt from the pipeline input,
// including the source content, game system schema, campaign context,
// and entity references.
func buildUserPrompt(input enrichment.PipelineInput) string {
	var b strings.Builder

	b.WriteString("## Content to Analyse\n\n")

	content := input.Content
	if len(content) > maxContentLength {
		content = content[:maxContentLength]
		b.WriteString(content)
		b.WriteString("\n\n[Content truncated at ")
		b.WriteString(fmt.Sprintf("%d", maxContentLength))
		b.WriteString(" characters]\n")
	} else {
		b.WriteString(content)
		b.WriteString("\n")
	}

	// Include game system schema if available.
	if input.Context != nil && input.Context.GameSystemYAML != "" {
		b.WriteString("\n## Game System Schema\n\n")
		b.WriteString("Use this schema to validate any game mechanics ")
		b.WriteString("referenced in the content:\n\n```yaml\n")
		b.WriteString(input.Context.GameSystemYAML)
		b.WriteString("\n```\n")
	}

	// Include campaign context from RAG results if available.
	if input.Context != nil && len(input.Context.CampaignResults) > 0 {
		b.WriteString("\n## Campaign Context\n\n")
		b.WriteString("Related campaign content for reference ")
		b.WriteString("(use for continuity checks):\n\n")
		for _, result := range input.Context.CampaignResults {
			b.WriteString("- **")
			b.WriteString(result.SourceName)
			b.WriteString("** (")
			b.WriteString(result.SourceTable)
			b.WriteString("): ")
			b.WriteString(truncateString(result.ChunkContent, 300))
			b.WriteString("\n")
		}
	}

	// Include entity list for reference.
	if len(input.Entities) > 0 {
		b.WriteString("\n## Known Entities\n\n")
		b.WriteString("These entities exist in the campaign:\n\n")
		for _, entity := range input.Entities {
			b.WriteString("- **")
			b.WriteString(entity.Name)
			b.WriteString("** (")
			b.WriteString(string(entity.EntityType))
			b.WriteString(")")
			if entity.Description != nil && *entity.Description != "" {
				desc := truncateString(*entity.Description, 100)
				b.WriteString(": ")
				b.WriteString(desc)
			}
			b.WriteString("\n")
		}
	}

	// Include relationship context if available.
	if len(input.Relationships) > 0 {
		b.WriteString("\n## Known Relationships\n\n")
		for _, rel := range input.Relationships {
			b.WriteString("- ")
			b.WriteString(rel.SourceEntityName)
			b.WriteString(" -> ")
			b.WriteString(rel.TargetEntityName)
			if rel.DisplayLabel != "" {
				b.WriteString(" (")
				b.WriteString(rel.DisplayLabel)
				b.WriteString(")")
			}
			b.WriteString("\n")
		}
	}

	// Add metadata.
	b.WriteString("\n## Metadata\n\n")
	b.WriteString(fmt.Sprintf("- Source: %s (ID: %d)\n", input.SourceTable, input.SourceID))
	if input.SourceField != "" {
		b.WriteString(fmt.Sprintf("- Field: %s\n", input.SourceField))
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
