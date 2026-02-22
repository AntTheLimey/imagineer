<!--
  Imagineer - TTRPG Campaign Intelligence Platform

  Copyright (c) 2025 - 2026
  This software is released under The MIT License
-->

# Entity Suggestion Quality Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use
> superpowers:executing-plans to implement this plan
> task-by-task.

**Goal:** Improve new entity suggestion descriptions,
pass descriptions through to created entities, and
insert wiki link tags when creating entities from
enrichment-phase suggestions.

**Architecture:** Three changes to the Go backend. The
LLM prompt gets better description guidance. The
ResolveItem handler extracts the description from
suggestedContent and passes the description to
CreateEntityRequest. A new helper performs global
wiki-link insertion for entities created without
position offsets.

**Tech Stack:** Go, PostgreSQL, LLM prompt engineering.

---

### Task 1: Update LLM Prompt for Better Descriptions

**Files:**

- Modify: `internal/enrichment/prompts.go:213-246`
- Test: `internal/enrichment/engine_test.go:963-972`

**Step 1: Update the existing prompt test**

In `engine_test.go`, update
`TestBuildNewEntityDetectionSystemPrompt` to assert the
new description guidance.

```go
func TestBuildNewEntityDetectionSystemPrompt(t *testing.T) {
	prompt := buildNewEntityDetectionSystemPrompt()

	assert.Contains(t, prompt, "TTRPG campaign analyst")
	assert.Contains(t, prompt, "new_entities")
	assert.Contains(t, prompt, "entity_type")
	assert.Contains(t, prompt, "npc")
	assert.Contains(t, prompt, "location")
	assert.Contains(t, prompt, "valid JSON only")
	assert.Contains(t, prompt,
		"two to three sentences")
}
```

**Step 2: Run the test to verify that the test fails**

Run the following command.

```bash
go test ./internal/enrichment/ \
    -run TestBuildNewEntityDetectionSystemPrompt -v
```

Expected: FAIL (the prompt does not contain "two to
three sentences").

**Step 3: Update the system prompt**

In `prompts.go`, replace
`buildNewEntityDetectionSystemPrompt` (lines 213-246)
with the following function.

```go
func buildNewEntityDetectionSystemPrompt() string {
	return `You are a TTRPG campaign analyst. Analyse
content to identify named entities (NPCs, locations,
items, factions, creatures, organizations, events,
documents, clues) mentioned but NOT in the campaign
database.

Rules:
- Only identify proper nouns and clear named entities.
- Do NOT identify generic references like "the tavern",
  "a guard", "the stranger", or "some soldiers".
- Only identify entities that are clearly distinct from
  any entity in the known entities list.
- For each entity, write a description of two to three
  sentences summarising everything known from the source
  content: their role, relationships to other characters
  or places, notable characteristics, and any actions
  described.

Supported entity types:
  npc, location, item, faction, clue, creature,
  organization, event, document, other

You MUST respond with valid JSON only. No markdown, no
commentary outside the JSON object.

Response format:
{
  "new_entities": [
    {
      "name": "Inspector Barrington",
      "entity_type": "npc",
      "description": "Inspector Barrington is a Scotland
Yard detective assigned to investigate the
disappearances in Whitechapel. He is described as
methodical and sceptical of the occult, preferring
forensic evidence over superstition. He first appears at
the crime scene where he clashes with the investigators
over jurisdiction.",
      "reasoning": "Named character appearing in
paragraph 3 who is not in the known entities list"
    }
  ]
}

If no new entities are found, return:
{"new_entities": []}`
}
```

**Step 4: Run the test to verify that the test passes**

Run the following command.

```bash
go test ./internal/enrichment/ \
    -run TestBuildNewEntityDetectionSystemPrompt -v
```

Expected: PASS.

**Step 5: Commit the changes**

```bash
git add internal/enrichment/prompts.go \
       internal/enrichment/engine_test.go
git commit -m "feat: improve LLM prompt for entity \
  suggestion descriptions"
```

---

### Task 2: Pass Description Through on Entity Creation

**Files:**

- Modify:
  `internal/api/content_analysis_handler.go:290-336`

**Step 1: Verify that existing tests still pass**

Skip a dedicated unit test for this step. Verify via
the existing handler tests that existing behaviour
remains intact.

Run the following command.

```bash
go test ./internal/api/ -v
```

Expected: PASS.

**Step 2: Fetch suggestedContent before the resolution
switch**

In `content_analysis_handler.go`, after the resolution
validation (line 288) and before the resolution switch
(line 293), add a query to fetch the suggestedContent
for the item. The handler uses this data in the
`new_entity` case to extract the description and in the
content-fix block for position data.

```go
	// Fetch suggestedContent for the item (used by
	// new_entity to extract description, and by the
	// content-fix block for position data).
	var itemSuggestedContent json.RawMessage
	var itemDetectionType string
	var itemSrcTable, itemSrcField string
	var itemSrcID int64
	var itemPosStart, itemPosEnd *int
	var itemMatchedText string
	err = h.db.QueryRow(r.Context(),
		`SELECT COALESCE(suggested_content, '{}'),
		        detection_type,
		        position_start, position_end,
		        matched_text,
		        j.source_table, j.source_id,
		        j.source_field
		 FROM content_analysis_items i
		 JOIN content_analysis_jobs j
		   ON i.job_id = j.id
		 WHERE i.id = $1`,
		itemID,
	).Scan(&itemSuggestedContent, &itemDetectionType,
		&itemPosStart, &itemPosEnd,
		&itemMatchedText,
		&itemSrcTable, &itemSrcID, &itemSrcField)
	if err != nil {
		log.Printf(
			"Error fetching item details: %v", err)
		respondError(w, http.StatusNotFound,
			"Analysis item not found")
		return
	}
```

**Step 3: Extract the description in the new_entity
case**

In the `new_entity` case (around line 321), before
creating the entity, extract the description from
suggestedContent.

```go
		// Extract description from suggestedContent
		// if available.
		var desc *string
		var sc map[string]interface{}
		if err := json.Unmarshal(
			itemSuggestedContent, &sc,
		); err == nil {
			if d, ok := sc["description"].(string); ok && d != "" {
				desc = &d
			}
		}

		createReq := models.CreateEntityRequest{
			EntityType:  *req.EntityType,
			Name:        *req.EntityName,
			Description: desc,
		}
```

**Step 4: Remove the duplicate item-details query**

The existing query at lines 377-387 that fetches
posStart, posEnd, matchedText, detectionType,
suggestedContent, srcTable, srcID, and srcField now
duplicates the query added in Step 2. Replace that
block to reuse the variables already fetched.

Replace lines 370-392 (the content-fix query block)
with the following code that reuses the already-fetched
variables.

```go
	// Apply content fix for accepted/new_entity
	// resolutions that have position offsets.
	var fixJobIDFound bool
	fixJobID := jobID
	posStart := itemPosStart
	posEnd := itemPosEnd
	matchedText := itemMatchedText
	detectionType := itemDetectionType
	suggestedContent := itemSuggestedContent
	srcTable := itemSrcTable
	srcID := itemSrcID
	srcField := itemSrcField
	fixJobIDFound = true
```

Keep the existing if-block starting at the
`posStart != nil && posEnd != nil` check (line 393
onwards) unchanged.

**Step 5: Run the tests**

Run the following command.

```bash
go test ./internal/api/ -v
```

Expected: PASS (existing tests still work).

**Step 6: Commit the changes**

```bash
git add internal/api/content_analysis_handler.go
git commit -m "feat: pass entity description from \
  suggestedContent on creation"
```

---

### Task 3: Global Wiki-Link Insertion for New Entities

**Files:**

- Modify:
  `internal/api/content_analysis_handler.go`

**Step 1: Add a helper function for global wiki-link
insertion**

Add a new method to `ContentAnalysisHandler` after the
existing `applyContentFix` method (around line 902).
The `applyGlobalWikiLinks` method replaces all
occurrences of the entity name in the source content
with `[[entityName]]`, skipping text already inside
`[[...]]` brackets. The method uses word-boundary
matching to avoid partial replacements within longer
words.

```go
// applyGlobalWikiLinks replaces all occurrences of
// entityName in the source content with
// [[entityName]], skipping text already inside
// [[...]] brackets. Uses word-boundary matching to
// avoid partial replacements within longer words.
func (h *ContentAnalysisHandler) applyGlobalWikiLinks(
	ctx context.Context,
	campaignID int64,
	sourceTable string,
	sourceID int64,
	sourceField string,
	entityName string,
) error {
	content, err := fetchSourceContent(
		ctx, h.db, campaignID,
		sourceTable, sourceID, sourceField)
	if err != nil {
		return fmt.Errorf(
			"failed to fetch source content: %w", err)
	}

	// Build a regex that matches the entity name on
	// word boundaries, but not inside [[...]].
	escaped := regexp.QuoteMeta(entityName)
	namePattern := regexp.MustCompile(
		`\b` + escaped + `\b`)

	// Find all wiki-link spans so the method can skip
	// matches inside them.
	linkSpans := wikiLinkPattern.FindAllStringIndex(
		content, -1)

	// Process matches in reverse order so earlier
	// byte offsets remain valid after each
	// replacement.
	matches := namePattern.FindAllStringIndex(
		content, -1)
	replacement := "[[" + entityName + "]]"
	changed := false
	for i := len(matches) - 1; i >= 0; i-- {
		m := matches[i]
		if insideWikiLink(m[0], m[1], linkSpans) {
			continue
		}
		content = content[:m[0]] + replacement +
			content[m[1]:]
		changed = true
	}

	if !changed {
		return nil
	}

	// Write back using the same table/field routing
	// as applyContentFix.
	return h.updateSourceContent(
		ctx, sourceTable, sourceID, sourceField,
		content)
}

// insideWikiLink checks whether the range
// [start, end) falls within any of the given
// wiki-link spans.
func insideWikiLink(
	start, end int, spans [][]int,
) bool {
	for _, span := range spans {
		if start >= span[0] && end <= span[1] {
			return true
		}
	}
	return false
}

// updateSourceContent writes content back to the
// specified source table and field.
func (h *ContentAnalysisHandler) updateSourceContent(
	ctx context.Context,
	sourceTable string,
	sourceID int64,
	sourceField string,
	content string,
) error {
	var updateSQL string
	switch sourceTable {
	case "entities":
		switch sourceField {
		case "description":
			updateSQL = `UPDATE entities
				SET description = $2,
				    updated_at = NOW()
				WHERE id = $1`
		case "gm_notes":
			updateSQL = `UPDATE entities
				SET gm_notes = $2,
				    updated_at = NOW()
				WHERE id = $1`
		default:
			return fmt.Errorf(
				"unsupported field %q for table %q",
				sourceField, sourceTable)
		}
	case "chapters":
		switch sourceField {
		case "overview":
			updateSQL = `UPDATE chapters
				SET overview = $2,
				    updated_at = NOW()
				WHERE id = $1`
		default:
			return fmt.Errorf(
				"unsupported field %q for table %q",
				sourceField, sourceTable)
		}
	case "sessions":
		switch sourceField {
		case "prep_notes":
			updateSQL = `UPDATE sessions
				SET prep_notes = $2,
				    updated_at = NOW()
				WHERE id = $1`
		case "actual_notes":
			updateSQL = `UPDATE sessions
				SET actual_notes = $2,
				    updated_at = NOW()
				WHERE id = $1`
		default:
			return fmt.Errorf(
				"unsupported field %q for table %q",
				sourceField, sourceTable)
		}
	case "campaigns":
		switch sourceField {
		case "description":
			updateSQL = `UPDATE campaigns
				SET description = $2,
				    updated_at = NOW()
				WHERE id = $1`
		default:
			return fmt.Errorf(
				"unsupported field %q for table %q",
				sourceField, sourceTable)
		}
	default:
		return fmt.Errorf(
			"unsupported source table: %s",
			sourceTable)
	}

	return h.db.Exec(
		ctx, updateSQL, sourceID, content)
}
```

Note: The `fetchSourceContent` function in
`source_content.go` takes a `campaignID` parameter. The
`applyGlobalWikiLinks` signature includes `campaignID
int64` and passes the value to `fetchSourceContent`.

**Step 2: Call the helper from the ResolveItem handler**

In the ResolveItem handler, after the existing
content-fix block (after line 448 where the
position-based fix ends), add a fallback for items
without positions.

```go
			// For new_entity_suggestion items without
			// position offsets, do a global
			// find-and-replace across the source
			// content.
			if req.Resolution == "new_entity" &&
				posStart == nil && posEnd == nil &&
				req.EntityName != nil {
				if linkErr :=
					h.applyGlobalWikiLinks(
						r.Context(),
						campaignID,
						srcTable, srcID,
						srcField,
						*req.EntityName,
					); linkErr != nil {
					log.Printf(
						"Global wiki-link insertion "+
							"failed for item %d: %v",
						itemID, linkErr)
				}
			}
```

**Step 3: Run the tests**

Run both test suites to confirm that all tests pass.

```bash
go test ./internal/api/ -v
```

```bash
go test ./internal/enrichment/ -v
```

Expected: PASS for both suites.

**Step 4: Commit the changes**

```bash
git add internal/api/content_analysis_handler.go
git commit -m "feat: insert wiki links globally when \
  creating entities from enrichment suggestions"
```

---

### Task 4: Final Verification

**Step 1: Run the full test suite**

Run the following command.

```bash
make test-all
```

Expected: All tests pass with no lint errors.

**Step 2: Update Todo.md**

Mark the three Entity Suggestion Quality items as
completed in `Todo.md`.

**Step 3: Update CHANGELOG.md**

Add entries for the three changes under a new section.

**Step 4: Commit the changes**

```bash
git add Todo.md CHANGELOG.md
git commit -m "docs: update CHANGELOG and Todo for \
  entity suggestion quality"
```
