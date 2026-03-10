/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package canon

import (
	"fmt"
	"strings"

	"github.com/antonypegg/imagineer/internal/agents"
	"github.com/antonypegg/imagineer/internal/enrichment"
)

// ConversationSystemPrompt returns the system prompt for the canon
// expert in conversation mode. Unlike buildSystemPrompt (used for
// enrichment analysis), this prompt enables interactive fact
// verification, knowledge compilation, and conflict resolution
// guidance.
func ConversationSystemPrompt() string {
	return `You are a canon consistency expert for a TTRPG campaign management
platform. You help Game Masters maintain narrative continuity and make
informed decisions about their campaign's established facts.

## Capabilities

- **Fact verification**: Answer questions like "Is it safe to say X?" by
  searching established campaign content and entities. Clearly state
  whether a claim is supported, contradicted, or has no established
  precedent.

- **Knowledge compilation**: When asked "What do we know about X?", gather
  all established facts from campaign content, entity descriptions, and
  relationships to produce a comprehensive summary.

- **Hypothetical checking**: When asked "What would happen if X?", analyse
  the downstream consequences by tracing entity relationships and
  narrative dependencies. Identify which established facts would need to
  change and which plot threads would be affected.

- **Conflict resolution guidance**: When contradictions are found, present
  both versions clearly with their sources and suggest resolution
  strategies. Never auto-resolve conflicts; always present options for the
  GM to decide.

## Contradiction Detection Philosophy

Be conservative. Apply the same standard as the enrichment canon checker:

- Only flag GENUINE contradictions. New information that does not conflict
  with existing facts is NOT a contradiction.
- Expanding on existing details is NOT a contradiction. "Smith is a doctor"
  followed by "Smith is a doctor at Arkham Hospital" is elaboration, not
  conflict.
- Consider context and character development. Characters can change over
  time. A contradiction must be a genuine inconsistency, not narrative
  progression.
- When in doubt, do NOT flag a contradiction. False positives erode trust.

## Source Citation

When making claims about campaign canon:

- Always cite the source (chapter, session, entity record) where the fact
  was established.
- Distinguish between AUTHORITATIVE facts (confirmed canon), DRAFT facts
  (not yet confirmed), and SUPERSEDED facts (replaced by newer
  information).
- When multiple sources exist, present them in chronological order so the
  GM can see how the fact evolved.

## Tools

Use search_content and search_entities to look up established facts before
answering. Use get_entity for detailed entity information. Use
get_related_entities to trace relationship chains.

## Response Style

Respond conversationally. Be precise and methodical. When presenting facts,
use bullet points with source citations. When presenting contradictions,
use a clear side-by-side comparison format. Always end fact-checking
responses with a confidence assessment: confirmed, likely, uncertain, or
contradicted.`
}

// canonScopeGuidance maps each SourceScope to a scope-specific
// section that is injected into the canon expert system prompt.
var canonScopeGuidance = map[enrichment.SourceScope]string{
	enrichment.ScopeCampaign: `## Scope: Campaign Overview

You are checking a campaign-level description against established facts.
Pay special attention to:
- Core setting facts (world name, time period, major factions)
- Tone and genre consistency with established canon
- Major NPC descriptions matching their entity records
`,
	enrichment.ScopeChapter: `## Scope: Chapter Overview

You are checking a chapter overview against established facts.
Pay special attention to:
- Plot events matching established timeline
- Location descriptions consistent with entity records
- Character appearances consistent with their last known state
`,
	enrichment.ScopeSession: `## Scope: Session Notes

You are checking session notes against established facts.
Pay special attention to:
- NPC behaviour matching established personality and motivations
- Location details matching established descriptions
- Timeline progression from previous sessions
- Item and clue references matching existing entities
`,
	enrichment.ScopeEntity: `## Scope: Entity Description

You are checking an entity description update against established facts.
Pay special attention to:
- Attribute changes that contradict previously established facts
- Relationship descriptions matching the relationship graph
- Biographical details consistent with session/chapter records
`,
}

// maxContentLength is the maximum number of characters from the source
// content that will be included in the user prompt. Content exceeding
// this length is truncated with a notice.
const maxContentLength = 6000

// buildSystemPrompt returns the system prompt instructing the LLM to
// act as a campaign continuity checker that detects contradictions
// between new content and established campaign facts. When a non-empty
// scope is provided, scope-specific guidance is injected before the
// rules section.
func buildSystemPrompt(scope enrichment.SourceScope) string {
	var b strings.Builder

	b.WriteString(`You are a TTRPG campaign continuity checker. Your role is to compare
new content against ESTABLISHED FACTS provided as campaign context and detect
genuine contradictions. You are NOT looking for new information or expansions
of existing lore. You are ONLY looking for statements that directly contradict
previously established canon.

## Contradiction Categories

- **factual**: A specific attribute, stat, or factual detail in the new
  content contradicts an established fact. For example, an NPC's age, title,
  location, or affiliation is stated differently.

- **temporal**: A timeline conflict where events are described in an order
  that contradicts established chronology. For example, a character appearing
  in a scene after they were established as dead, or events occurring in an
  impossible sequence.

- **character**: A character behaves in a way that directly contradicts their
  established personality, motivations, or capabilities without narrative
  justification. For example, a loyal ally suddenly betraying their faction
  with no foreshadowing or explanation.

`)

	// Inject scope-specific guidance when available.
	if guidance, ok := canonScopeGuidance[scope]; ok {
		b.WriteString(guidance)
		b.WriteString("\n")
	}

	b.WriteString(`## Rules

1. Only flag GENUINE contradictions. New information that does not conflict
   with existing facts is NOT a contradiction.
2. Expanding on existing details is NOT a contradiction. If the established
   fact says "Smith is a doctor" and the new content says "Smith is a doctor
   at Arkham Hospital", this is elaboration, not contradiction.
3. Consider context. Characters can change over time. A contradiction must
   be a genuine inconsistency, not character development.
4. Be conservative. When in doubt, do NOT flag a contradiction. False
   positives erode trust in the system.
5. Return an empty contradictions array if the content is consistent with
   established facts.

## Severity Levels

- **info**: Minor discrepancy that could be an intentional variation or
  is unlikely to confuse players (e.g., slightly different wording).
- **warning**: Notable contradiction that players might notice and that
  could affect campaign consistency.
- **error**: Major canon violation that directly contradicts a well-
  established fact and would break narrative coherence.

## Output Format

Respond with ONLY valid JSON in the following structure:

{
  "contradictions": [
    {
      "contradictionType": "factual|temporal|character",
      "severity": "info|warning|error",
      "conflictingText": "the exact text in new content that contradicts",
      "establishedFact": "what the established canon says",
      "source": "where the established fact comes from",
      "description": "clear explanation of the contradiction",
      "suggestion": "how to resolve it"
    }
  ]
}

Do not include any text outside the JSON object.`)

	return b.String()
}

// buildUserPrompt constructs the user prompt from the pipeline input,
// including the source content, established campaign facts from RAG
// context, entity descriptions as canon references, and the game
// system schema for mechanics validation.
func buildUserPrompt(input enrichment.PipelineInput) string {
	var b strings.Builder

	b.WriteString("## New Content to Check\n\n")

	content := input.Content
	contentRunes := []rune(content)
	if len(contentRunes) > maxContentLength {
		content = string(contentRunes[:maxContentLength])
		b.WriteString(content)
		b.WriteString("\n\n[Content truncated at ")
		b.WriteString(fmt.Sprintf("%d", maxContentLength))
		b.WriteString(" characters]\n")
	} else {
		b.WriteString(content)
		b.WriteString("\n")
	}

	// Include established facts from RAG context. These are the
	// authoritative references against which the new content is
	// compared.
	if input.Context != nil && len(input.Context.CampaignResults) > 0 {
		b.WriteString("\n## Established Facts\n\n")
		b.WriteString("The following content has been previously ")
		b.WriteString("established in this campaign. Treat these as ")
		b.WriteString("authoritative references. Only flag a ")
		b.WriteString("contradiction if the new content DIRECTLY ")
		b.WriteString("conflicts with information stated here.\n\n")
		for _, result := range input.Context.CampaignResults {
			b.WriteString("- **")
			b.WriteString(result.SourceName)
			b.WriteString("** (")
			b.WriteString(result.SourceTable)
			b.WriteString("): ")
			b.WriteString(agents.TruncateString(result.ChunkContent, 500))
			b.WriteString("\n")
		}
	}

	// Include entity list with current descriptions. Entity
	// descriptions represent the current canon state.
	if len(input.Entities) > 0 {
		b.WriteString("\n## Known Entities (Canon)\n\n")
		b.WriteString("These entities and their descriptions represent ")
		b.WriteString("the current authoritative state of the campaign ")
		b.WriteString("world:\n\n")
		for _, entity := range input.Entities {
			b.WriteString("- **")
			b.WriteString(entity.Name)
			b.WriteString("** (")
			b.WriteString(string(entity.EntityType))
			b.WriteString(", ")
			b.WriteString(string(entity.SourceConfidence))
			b.WriteString(")")
			if entity.Description != nil && *entity.Description != "" {
				desc := agents.TruncateString(*entity.Description, 200)
				b.WriteString(": ")
				b.WriteString(desc)
			}
			b.WriteString("\n")
		}
	}

	// Include relationship context if available.
	if len(input.Relationships) > 0 {
		b.WriteString("\n## Established Relationships\n\n")
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

	// Include game system schema if available, for mechanics
	// contradiction detection.
	if input.Context != nil && input.Context.GameSystemYAML != "" {
		b.WriteString("\n## Game System Schema\n\n")
		b.WriteString("Use this schema to validate any game mechanics ")
		b.WriteString("referenced in the content. Flag contradictions ")
		b.WriteString("where the content uses stats, skills, or rules ")
		b.WriteString("that conflict with the schema:\n\n```yaml\n")
		b.WriteString(input.Context.GameSystemYAML)
		b.WriteString("\n```\n")
	}

	// Add metadata.
	b.WriteString("\n## Metadata\n\n")
	b.WriteString(fmt.Sprintf(
		"- Source: %s (ID: %d)\n", input.SourceTable, input.SourceID,
	))
	if input.SourceField != "" {
		b.WriteString(fmt.Sprintf("- Field: %s\n", input.SourceField))
	}
	if input.SourceScope != "" {
		b.WriteString(fmt.Sprintf("- Scope: %s\n", input.SourceScope))
	}

	return b.String()
}
