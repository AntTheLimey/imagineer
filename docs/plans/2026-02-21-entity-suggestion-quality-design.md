<!--
  Imagineer - TTRPG Campaign Intelligence Platform

  Copyright (c) 2025 - 2026
  This software is released under The MIT License
-->

# Entity Suggestion Quality

This design document describes improvements to the quality
and completeness of new entity suggestions from the enrichment
pipeline. Three independent changes address description depth,
data propagation, and wiki link insertion.

## Problem

Three issues degrade the entity creation experience when a user
accepts new entity suggestions from content analysis.

### Terse Descriptions

The LLM prompt in `buildNewEntityDetectionSystemPrompt()`
produces terse one-line descriptions because the example is
just "A Scotland Yard detective mentioned in the chapter" and
the prompt provides no guidance on desired detail. The resulting
descriptions lack context about the entity's role,
relationships, characteristics, and actions.

### Discarded Description on Creation

When the user clicks "Create Entity" on a
`new_entity_suggestion`, the handler at
`content_analysis_handler.go:321` only passes `Name` and
`EntityType` to `CreateEntityRequest`. The handler discards the
`Description` from `suggestedContent`, creating entities with
no description.

### Missing Wiki Links

Enrichment-phase `new_entity_suggestion` items have no
`position_start`/`position_end` offsets. The wiki link
insertion logic at line 393 silently skips these items because
the offsets are nil. The system creates the entity but never
tags the source text.

## Design

### 1. Better LLM Descriptions

The first change updates
`buildNewEntityDetectionSystemPrompt()` in
`internal/enrichment/prompts.go` to produce richer
descriptions.

The updated prompt makes the following changes:

- The prompt instructs the LLM to provide two to three
  sentences covering the entity's role, relationships,
  characteristics, and actions mentioned in the source
  content.
- The prompt replaces the existing single-sentence example
  with a multi-sentence example that demonstrates the
  expected level of detail.

This change affects a single function and requires no
structural modifications.

### 2. Pass Description Through on Entity Creation

The second change ensures the handler propagates the
description when creating an entity from a suggestion.

In `content_analysis_handler.go`, the item-details query
(currently at lines 377-387) loads `suggestedContent` but
executes after the resolution switch statement. The fix moves
this query before the switch so that `suggestedContent` is
available in the `new_entity` case. The handler then extracts
the `description` field from `suggestedContent` and sets the
value on `CreateEntityRequest.Description`.

This restructuring also simplifies the later content-fix block,
which currently re-fetches the same data.

### 3. Wiki Link Insertion for New Entities

The third change adds wiki link insertion for new entity
suggestions that lack position offsets.

After creating the entity and resolving the analysis item, when
`detection_type` equals `"new_entity_suggestion"` and positions
are nil, the handler performs a global find-and-replace on the
source content. The process follows these steps:

1. The handler fetches the source content using the existing
   `fetchSourceContent` function.
2. The handler finds all occurrences of the entity name using
   a case-sensitive exact match on word boundaries.
3. The handler skips occurrences already inside `[[...]]`
   brackets.
4. The handler replaces each remaining match with
   `[[entity name]]`.
5. The handler writes the updated content back using the
   existing `applySourceContentUpdate` function.

This approach reuses the existing source content fetch and
update infrastructure without requiring position offsets.

## Files to Modify

The following table lists the files that require changes:

| File | Change |
|------|--------|
| `internal/enrichment/prompts.go` | Description guidance in the system prompt. |
| `internal/api/content_analysis_handler.go` | Move the item query earlier, extract the description, and add global wiki link insertion. |

## What This Does Not Change

This design intentionally excludes the following areas:

- The frontend requires no changes because descriptions
  already render from `suggestedContent`.
- No migration changes are necessary.
- No new API endpoints are required.
