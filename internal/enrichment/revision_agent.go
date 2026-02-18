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
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/antonypegg/imagineer/internal/models"
)

// RevisionInput contains everything needed to generate a revision.
type RevisionInput struct {
	OriginalContent string                       // The source content to revise
	AcceptedItems   []models.ContentAnalysisItem // Accepted Stage 1 findings
	SourceTable     string                       // e.g., "chapters", "sessions"
	SourceID        int64
	GameSystemYAML  string // Optional game system context
}

// RevisionResult contains the generated revision.
type RevisionResult struct {
	RevisedContent string `json:"revisedContent"`
	Summary        string `json:"summary"` // Brief summary of changes made
}

// RevisionAgent generates revised content incorporating accepted findings.
type RevisionAgent struct{}

// NewRevisionAgent creates a new RevisionAgent.
func NewRevisionAgent() *RevisionAgent {
	return &RevisionAgent{}
}

// GenerateRevision produces revised content based on accepted analysis
// findings. It builds a prompt from the original content and accepted
// items, calls the LLM provider, and parses the response into a
// RevisionResult. If no accepted items are provided, the original
// content is returned unchanged without an LLM call.
func (ra *RevisionAgent) GenerateRevision(
	ctx context.Context,
	provider llm.Provider,
	input RevisionInput,
) (*RevisionResult, error) {
	if input.OriginalContent == "" {
		return nil, fmt.Errorf("original content is required for revision")
	}

	if len(input.AcceptedItems) == 0 {
		return &RevisionResult{
			RevisedContent: input.OriginalContent,
			Summary:        "",
		}, nil
	}

	systemPrompt := buildRevisionSystemPrompt()
	userPrompt := buildRevisionUserPrompt(input)

	resp, err := provider.Complete(ctx, llm.CompletionRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    8192,
		Temperature:  0.4,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM completion failed: %w", err)
	}

	return parseRevisionResponse(resp.Content), nil
}

// buildRevisionSystemPrompt constructs the system prompt that instructs
// the LLM to act as a TTRPG content editor.
func buildRevisionSystemPrompt() string {
	return `You are a TTRPG content editor. Revise the following content to address the accepted suggestions while preserving the author's voice and style.

Rules:
- Make ONLY the changes suggested by the accepted findings.
- Do not add new content beyond what is needed to address the findings.
- Preserve formatting, markdown structure, and the author's writing style.
- Return valid JSON with two fields:
  - "revisedContent": the full revised text
  - "summary": a 2-3 sentence description of the changes made

Respond with valid JSON only.`
}

// buildRevisionUserPrompt constructs the user prompt containing the
// original content, accepted findings, and optional game system context.
func buildRevisionUserPrompt(input RevisionInput) string {
	var b strings.Builder

	b.WriteString("## Original Content\n\n")
	b.WriteString(input.OriginalContent)
	b.WriteString("\n\n")

	b.WriteString("## Accepted Findings\n\n")
	for i, item := range input.AcceptedItems {
		fmt.Fprintf(&b, "### Finding %d\n\n", i+1)
		fmt.Fprintf(&b, "**Detection Type**: %s\n", item.DetectionType)
		fmt.Fprintf(&b, "**Matched Text**: %s\n", item.MatchedText)

		if len(item.SuggestedContent) > 0 {
			var content map[string]string
			if err := json.Unmarshal(item.SuggestedContent, &content); err == nil {
				if desc, ok := content["description"]; ok && desc != "" {
					fmt.Fprintf(&b, "**Description**: %s\n", desc)
				}
				if sug, ok := content["suggestion"]; ok && sug != "" {
					fmt.Fprintf(&b, "**Suggestion**: %s\n", sug)
				}
			}
		}

		b.WriteString("\n")
	}

	if input.GameSystemYAML != "" {
		b.WriteString("## Game System Context\n\n")
		b.WriteString("```yaml\n")
		b.WriteString(input.GameSystemYAML)
		b.WriteString("\n```\n\n")
	}

	return b.String()
}

// parseRevisionResponse parses the LLM response into a RevisionResult.
// It first attempts to parse as JSON. If that fails, the raw response
// text is used as the revised content with an empty summary (fallback
// for cases where the LLM returns plain text instead of JSON).
func parseRevisionResponse(raw string) *RevisionResult {
	cleaned := stripCodeFences(raw)
	cleaned = strings.TrimSpace(cleaned)

	if cleaned == "" {
		return &RevisionResult{
			RevisedContent: "",
			Summary:        "",
		}
	}

	var result RevisionResult
	if err := json.Unmarshal([]byte(cleaned), &result); err == nil {
		return &result
	}

	// Fallback: treat raw response as the revised content directly.
	return &RevisionResult{
		RevisedContent: strings.TrimSpace(raw),
		Summary:        "",
	}
}
