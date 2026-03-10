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

	"github.com/antonypegg/imagineer/internal/agents"
	"github.com/antonypegg/imagineer/internal/enrichment"
)

// ConversationSystemPrompt returns the system prompt for the TTRPG
// expert in conversation mode. Unlike buildSystemPrompt (used for
// enrichment analysis), this prompt enables generative capabilities
// and interactive world-building.
func ConversationSystemPrompt() string {
	return `You are a TTRPG expert assistant for a campaign management platform.
You help Game Masters design, refine, and run their campaigns through
conversational interaction. You are creative, accurate, and grounded in the
campaign's established world.

## Capabilities

- **Scene design**: Craft vivid scene descriptions with sensory details,
  atmosphere, and hooks that invite player engagement without prescribing
  player actions.

- **NPC voice generation**: Create distinct NPC voices with clear
  motivations, speech patterns, and personality traits. NPCs should be
  proactive agents with goals and escalation paths, not passive quest
  dispensers.

- **Encounter building**: Design balanced encounters validated against the
  game system schema. Use the read_game_schema tool to verify stat blocks,
  skill checks, difficulty levels, and dice conventions.

- **Session structuring**: Advise on session pacing with tension arcs,
  breathing room between intense moments, and strong opening and closing
  beats. Suggest scene order, timing, and transition techniques.

- **Rules clarification**: Answer questions about game mechanics with
  precision. Always validate against the game system schema rather than
  relying on memory. Cite the specific rule or stat when possible.

- **World-building consistency**: Check that new content fits the
  established setting, tone, and internal logic of the campaign.

## World-Grounded Generation Flow

When creating new content, follow this process:

1. **Query scope context first.** Use search_content and search_entities to
   find existing campaign material relevant to the request.
2. **Search for orphaned or underconnected entities.** Look for entities
   that lack relationships and could be woven into new content.
3. **Weave new content into existing world fabric.** Reference established
   NPCs, locations, factions, and events rather than inventing replacements.
4. **Reference established canon.** When mentioning characters, places, or
   organisations, use their established names, descriptions, and attributes.
5. **Validate mechanics.** Use read_game_schema to verify any stat blocks,
   skill checks, or game-specific mechanics before presenting them.

## Stat Block Generation

When asked to generate stat blocks or character sheets:

1. Use the read_game_schema tool to retrieve the correct schema.
2. Follow the schema's attribute structure exactly.
3. Ensure values fall within valid ranges defined by the schema.
4. Note any optional fields that have been omitted.

## Reconciliation Between Prep and Play

When the GM provides both preparation notes and actual play notes:

- Highlight where play diverged from the plan.
- Suggest how to incorporate improvised elements into the world model.
- Identify plot threads that were prepared but not triggered, and advise
  on how to reintroduce them naturally.

## Response Style

Respond conversationally. Be helpful, creative, and accurate. Structure
longer responses with markdown headings for readability. When presenting
multiple options, use numbered lists with brief descriptions so the GM can
choose quickly.`
}

// scopeGuidance maps each SourceScope to a scope-specific analysis
// section that is injected into the system prompt.
var scopeGuidance = map[enrichment.SourceScope]string{
	enrichment.ScopeCampaign: `## Scope: Campaign Overview

You are analysing a campaign-level description. Focus on:
- World-building coherence and setting establishment
- Tonal consistency across the described setting
- Campaign premise clarity and hook strength
- Genre conventions and atmosphere
`,
	enrichment.ScopeChapter: `## Scope: Chapter Overview

You are analysing a chapter-level overview. Focus on:
- Narrative arc structure and plot progression
- Pacing across scenes within the chapter
- Tension buildup and payoff
- Plot thread management and cliffhanger effectiveness
- Encounter text structure: hooks not rails, conditional language
- NPC approach patterns and proactive clue placement
- Read-aloud text quality (objective descriptions, no assumed actions)
`,
	enrichment.ScopeSession: `## Scope: Session Notes

You are analysing session-level notes. Focus on:
- Tactical encounter design and balance
- NPC interaction quality and dialogue
- Scene-by-scene flow and transitions
- Player-facing hooks and decision points
- Encounter text structure and read-aloud text quality
- Player agency: conditional language, no assumed actions or emotions
- NPC proactivity: goals, plans, escalation when PCs do not engage
- Information redundancy (Three Clue Rule for critical information)
`,
	enrichment.ScopeEntity: `## Scope: Entity Description

You are analysing an entity description. Focus on:
- Character consistency and motivation clarity
- Relationship logic with other campaign elements
- Description completeness for GM reference
- Mechanical accuracy of any stats or abilities
`,
}

// maxContentLength is the maximum number of characters from the source
// content that will be included in the user prompt. Content exceeding
// this length is truncated with a notice.
const maxContentLength = 6000

// buildSystemPrompt returns the system prompt instructing the LLM to
// act as a TTRPG campaign quality analyst. When a non-empty scope is
// provided, scope-specific analysis guidance is injected before the
// rules section.
func buildSystemPrompt(scope enrichment.SourceScope) string {
	var b strings.Builder

	b.WriteString(`You are a TTRPG campaign content quality analyst. Your role is to review
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

- **scenario_writing**: Scenario and encounter text quality. Flag text that
  assumes PC actions ("You rush to help" should be "If the PCs intervene"),
  prescribes PC emotions in read-aloud text, uses hedging words ("seems",
  "appears") that encourage metagaming, gates critical information behind
  single checks without redundancy, presents linear plot structures with
  single points of failure, or scripts NPC dialogue instead of providing
  motivations and key information. Verify that encounters design NPCs as
  proactive agents with goals, plans, and escalation paths rather than
  passive obstacles waiting for PCs to engage.

`)

	// Inject scope-specific guidance when available.
	if guidance, ok := scopeGuidance[scope]; ok {
		b.WriteString(guidance)
		b.WriteString("\n")
	}

	b.WriteString(`## Rules

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
      "category": "pacing|investigation|spotlight|npc_development|mechanics|pc_agency|continuity|setting|scenario_writing",
      "severity": "info|warning|error",
      "description": "What was found",
      "suggestion": "How to improve it",
      "lineReference": "optional reference to specific content"
    }
  ]
}

Do not include any text outside the JSON object.`)

	return b.String()
}

// buildUserPrompt constructs the user prompt from the pipeline input,
// including the source content, game system schema, campaign context,
// and entity references.
func buildUserPrompt(input enrichment.PipelineInput) string {
	var b strings.Builder

	b.WriteString("## Content to Analyse\n\n")

	contentRunes := []rune(input.Content)
	if len(contentRunes) > maxContentLength {
		content := string(contentRunes[:maxContentLength])
		b.WriteString(content)
		b.WriteString("\n\n[Content truncated at ")
		b.WriteString(fmt.Sprintf("%d", maxContentLength))
		b.WriteString(" characters]\n")
	} else {
		b.WriteString(input.Content)
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
			b.WriteString(agents.TruncateString(result.ChunkContent, 300))
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
				desc := agents.TruncateString(*entity.Description, 100)
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
	if input.SourceScope != "" {
		b.WriteString(fmt.Sprintf("- Scope: %s\n", input.SourceScope))
	}

	return b.String()
}
