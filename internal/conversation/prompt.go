/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package conversation

import (
	"bytes"
	"fmt"
	"os"
	"text/template"
)

// PromptContext holds the data needed to render
// the system prompt template.
type PromptContext struct {
	CampaignName  string
	ScopeType     string
	ScopeID       int64
	GameSystem    string
	CurrentDate   string
	EntityContext string
}

// LoadSystemPrompt parses the template at templatePath
// and renders it with the given context.
func LoadSystemPrompt(
	templatePath string,
	ctx PromptContext,
) (string, error) {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf(
			"reading prompt template %s: %w",
			templatePath, err,
		)
	}

	tmpl, err := template.New("conversation").Parse(
		string(content),
	)
	if err != nil {
		return "", fmt.Errorf(
			"parsing prompt template %s: %w",
			templatePath, err,
		)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", fmt.Errorf(
			"executing prompt template %s: %w",
			templatePath, err,
		)
	}

	return buf.String(), nil
}
