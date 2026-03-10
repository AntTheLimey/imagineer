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
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadSystemPrompt(t *testing.T) {
	// Create a temp template with basic substitutions.
	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "test.tmpl")
	tmplContent := `Welcome to {{.CampaignName}}, a {{.GameSystem}} game.`
	if err := os.WriteFile(tmplPath, []byte(tmplContent), 0644); err != nil {
		t.Fatalf("writing temp template: %v", err)
	}

	ctx := PromptContext{
		CampaignName: "Shadows Over Arkham",
		GameSystem:   "Call of Cthulhu 7e",
	}

	result, err := LoadSystemPrompt(tmplPath, ctx)
	if err != nil {
		t.Fatalf("LoadSystemPrompt returned error: %v", err)
	}

	if !strings.Contains(result, "Shadows Over Arkham") {
		t.Errorf(
			"expected result to contain campaign name, got: %s",
			result,
		)
	}
	if !strings.Contains(result, "Call of Cthulhu 7e") {
		t.Errorf(
			"expected result to contain game system, got: %s",
			result,
		)
	}
}

func TestLoadSystemPromptMissingFile(t *testing.T) {
	_, err := LoadSystemPrompt(
		"/nonexistent/path/template.tmpl",
		PromptContext{},
	)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	if !strings.Contains(err.Error(), "reading prompt template") {
		t.Errorf(
			"expected error to mention reading, got: %v",
			err,
		)
	}
}

func TestLoadSystemPromptWithEntityContext(t *testing.T) {
	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "test.tmpl")
	tmplContent := `Base prompt.
{{if .EntityContext}}
## Context

{{.EntityContext}}
{{end}}`
	if err := os.WriteFile(tmplPath, []byte(tmplContent), 0644); err != nil {
		t.Fatalf("writing temp template: %v", err)
	}

	t.Run("with entity context", func(t *testing.T) {
		ctx := PromptContext{
			EntityContext: "The Keeper is investigating a cult.",
		}

		result, err := LoadSystemPrompt(tmplPath, ctx)
		if err != nil {
			t.Fatalf("LoadSystemPrompt returned error: %v", err)
		}

		if !strings.Contains(result, "## Context") {
			t.Error("expected context heading when EntityContext is set")
		}
		if !strings.Contains(result, "investigating a cult") {
			t.Error("expected entity context content in output")
		}
	})

	t.Run("without entity context", func(t *testing.T) {
		ctx := PromptContext{
			EntityContext: "",
		}

		result, err := LoadSystemPrompt(tmplPath, ctx)
		if err != nil {
			t.Fatalf("LoadSystemPrompt returned error: %v", err)
		}

		if strings.Contains(result, "## Context") {
			t.Error("expected no context heading when EntityContext is empty")
		}
	})
}

func TestLoadSystemPromptAllFields(t *testing.T) {
	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "test.tmpl")
	tmplContent := `Campaign: {{.CampaignName}}
System: {{.GameSystem}}
Date: {{.CurrentDate}}
Scope: {{.ScopeType}} (ID: {{.ScopeID}})
{{if .EntityContext}}Context: {{.EntityContext}}
{{end}}`
	if err := os.WriteFile(tmplPath, []byte(tmplContent), 0644); err != nil {
		t.Fatalf("writing temp template: %v", err)
	}

	ctx := PromptContext{
		CampaignName:  "The Dracula Dossier",
		ScopeType:     "chapter",
		ScopeID:       42,
		GameSystem:    "GURPS 4e",
		CurrentDate:   "2026-03-10",
		EntityContext: "Dracula is active in London.",
	}

	result, err := LoadSystemPrompt(tmplPath, ctx)
	if err != nil {
		t.Fatalf("LoadSystemPrompt returned error: %v", err)
	}

	expectations := map[string]string{
		"CampaignName":  "The Dracula Dossier",
		"GameSystem":    "GURPS 4e",
		"CurrentDate":   "2026-03-10",
		"ScopeType":     "chapter",
		"ScopeID":       "42",
		"EntityContext": "Dracula is active in London.",
	}

	for field, expected := range expectations {
		if !strings.Contains(result, expected) {
			t.Errorf(
				"expected %s value %q in output, got: %s",
				field, expected, result,
			)
		}
	}
}
