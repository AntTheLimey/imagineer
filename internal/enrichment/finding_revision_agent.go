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
	"fmt"
	"strings"

	"github.com/antonypegg/imagineer/internal/llm"
)

// FindingRevisionInput contains everything needed to generate a surgical
// revision for a small context window containing one or more findings.
type FindingRevisionInput struct {
	ContextWindow string          // The text surrounding the finding(s)
	Findings      []FindingDetail // Individual findings within the window
	Instructions  string          // Optional GM instructions for the revision
}

// FindingDetail describes a single finding to be addressed during revision.
type FindingDetail struct {
	DetectionType string // e.g., "spelling", "mechanics_warning", "pacing_note"
	MatchedText   string // The text span the finding applies to
	Description   string // What the finding detected
	Suggestion    string // How to fix it
}

// FindingRevisionResult contains the revised text for the context window.
type FindingRevisionResult struct {
	RevisedSection string // The revised context window text
}

// GenerateFindingRevision produces a revised version of a context window
// that incorporates the specified findings. It validates input, returns
// the context unchanged when there are no findings, and otherwise calls
// the LLM provider to generate the surgical edit.
func GenerateFindingRevision(
	ctx context.Context,
	provider llm.Provider,
	input FindingRevisionInput,
) (*FindingRevisionResult, error) {
	if input.ContextWindow == "" {
		return nil, fmt.Errorf("context window is required for finding revision")
	}

	if len(input.Findings) == 0 {
		return &FindingRevisionResult{
			RevisedSection: input.ContextWindow,
		}, nil
	}

	systemPrompt := buildFindingSystemPrompt()
	userPrompt := buildFindingUserPrompt(input)

	resp, err := provider.Complete(ctx, llm.CompletionRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    1024,
		Temperature:  0.4,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM completion failed: %w", err)
	}

	return &FindingRevisionResult{
		RevisedSection: strings.TrimSpace(resp.Content),
	}, nil
}

// buildFindingSystemPrompt constructs the system prompt that instructs
// the LLM to perform surgical editing on a small text section.
func buildFindingSystemPrompt() string {
	return `You are a TTRPG content editor performing surgical edits on a small section of text.

Rules:
- Make ONLY the changes described in the findings below.
- Preserve the author's voice, style, and all surrounding text that is not affected by a finding.
- Output the revised section as plain text. Do not wrap it in JSON, code fences, or any other structure.
- Do not add new content beyond what is needed to address the findings.
- Do not remove content unless a finding explicitly calls for removal.
- Maintain paragraph breaks, formatting, and whitespace from the original.`
}

// buildFindingUserPrompt constructs the user prompt containing the context
// window, individual findings, and optional GM instructions.
func buildFindingUserPrompt(input FindingRevisionInput) string {
	var b strings.Builder

	b.WriteString("## Context Window\n\n")
	b.WriteString(input.ContextWindow)
	b.WriteString("\n\n")

	b.WriteString("## Findings\n\n")
	for i, f := range input.Findings {
		fmt.Fprintf(&b, "### Finding %d\n\n", i+1)
		fmt.Fprintf(&b, "**Detection Type**: %s\n", f.DetectionType)
		fmt.Fprintf(&b, "**Matched Text**: %s\n", f.MatchedText)
		if f.Description != "" {
			fmt.Fprintf(&b, "**Description**: %s\n", f.Description)
		}
		if f.Suggestion != "" {
			fmt.Fprintf(&b, "**Suggestion**: %s\n", f.Suggestion)
		}
		b.WriteString("\n")
	}

	if input.Instructions != "" {
		b.WriteString("## GM Instructions\n\n")
		b.WriteString(input.Instructions)
		b.WriteString("\n\n")
	}

	return b.String()
}
