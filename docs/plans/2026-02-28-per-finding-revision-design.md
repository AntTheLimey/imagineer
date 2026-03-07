<!--
  Imagineer - TTRPG Campaign Intelligence Platform

  Copyright (c) 2025 - 2026
  This software is released under The MIT License
-->

# Per-Finding Surgical Revision Flow

## Problem

The current revision flow sends the entire source document, all
acknowledged findings, the full game system YAML schema, and RAG
campaign search results in a single massive LLM call. This is:

- Slow: 2+ minutes per revision, often timing out
- Expensive: ~6000+ input tokens and ~4000 output tokens per
  call
- Wasteful: rewrites unchanged portions, generates for
  dismissed findings

## Design

Replace the single-call full-document revision with per-finding
surgical edits generated on demand.

### Core Principles

1. **Per-finding granularity**: each finding generates its own
   targeted edit, not a full document rewrite
2. **On-demand generation**: edits are only generated when the
   GM clicks "Generate Suggestion" on a finding -- no wasted
   tokens on findings the GM will dismiss
3. **Immediate apply**: accepting an edit updates the source
   content in the DB immediately
4. **GM guidance**: optional text field lets the GM add
   instructions before generating (e.g., "make it darker,
   mention the assassinations")
5. **Overlapping groups**: findings that share the same
   paragraph or have overlapping position ranges are grouped
   and handled in one LLM call for coherence

### Data Flow

```
GM clicks "Generate Suggestion" on a finding (or group)
  -> POST /api/campaigns/{id}/analysis/items/revision
     Body: { itemIds: [1, 2], instructions?: "..." }
  -> Backend:
     1. Load items (matchedText, positionStart, positionEnd,
        suggestedContent)
     2. Load source content, extract context window
        (~2 paragraphs around position range)
     3. Small LLM call (~300-800 input tokens):
        System: "You are a TTRPG content editor..."
        User: "Here is a section of text. Apply these
               changes: [findings]. GM instructions:
               [optional]. Return the revised section."
     4. Return { originalSection, revisedSection }
  -> UI shows +/- diff inline under the finding
  -> GM clicks [Accept]:
     -> PUT /api/campaigns/{id}/analysis/items/revision/apply
        Body: { itemIds: [1, 2], revisedSection: "..." }
     -> Splices revised section into source content at
        position
     -> Shifts positions of subsequent items in the same job
     -> Marks items as accepted
  -> GM clicks [Edit]:
     -> Opens inline text editor on revisedSection
     -> GM modifies, then clicks "Save & Apply"
     -> Same apply flow with edited text
  -> GM clicks [Dismiss]:
     -> Marks items as dismissed, no content change
```

### Token Efficiency

| Metric              | Current          | New              |
|---------------------|------------------|------------------|
| Input per revision  | ~6000 tokens     | ~300-800 tokens  |
|                     | (full doc + all  | (paragraph +     |
|                     | items + YAML     | finding)         |
|                     | + RAG)           |                  |
| Output per revision | ~4000 tokens     | ~200-500 tokens  |
|                     | (full rewrite)   | (paragraph)      |
| Calls per session   | 1 massive call   | N small calls    |
|                     |                  | (only for items  |
|                     |                  | GM wants)        |
| Wasted tokens       | High (full doc   | Near zero        |
|                     | rewrite,         |                  |
|                     | dismissed items) |                  |
| Latency             | 2+ minutes       | 5-10 seconds per |
|                     |                  | finding          |

### API Changes

#### New Endpoints

**POST /api/campaigns/{id}/analysis/items/revision**

Generate a revision suggestion for one or more grouped items.

Request body:

```json
{
  "itemIds": [425, 426],
  "instructions": "make it dark and foreboding"
}
```

Response:

```json
{
  "originalSection": "The kingdom fell into chaos.",
  "revisedSection": "The kingdom fell into chaos. A bitter regency..."
}
```

The handler:

1. Loads all items by ID, validates they belong to the
   campaign
2. Validates items share overlapping positions or same
   paragraph
3. Loads the source content from the parent job's source
   table/field/ID
4. Extracts context window: the paragraph(s) containing the
   items' position range, plus one paragraph before and after
   (bounded by double newlines)
5. Builds a focused prompt with: context window, all
   findings' details, optional GM instructions
6. Calls LLM with max_tokens=1024, temperature=0.4
7. Returns original and revised sections

**PUT /api/campaigns/{id}/analysis/items/revision/apply**

Apply a revision (possibly edited by the GM) to the source
content.

Request body:

```json
{
  "itemIds": [425, 426],
  "revisedSection": "The kingdom fell into chaos. A bitter regency..."
}
```

The handler:

1. Loads items, identifies the position range to replace
2. Loads source content, splices revisedSection in at the
   position range
3. Updates source content in the DB (chapter overview,
   session notes, campaign description, entity description,
   etc.)
4. Calculates length delta (new length - old length)
5. Shifts positions of all items in the same job where
   positionStart > the edit point:
   `position_start += delta, position_end += delta`
6. Marks all items in itemIds as resolution=accepted
7. Clears any previously generated suggestions for items
   whose positions were shifted (forces regeneration)
8. Returns the updated source content

#### Deprecated Endpoints

These existing endpoints become unused:

- `POST /api/campaigns/{id}/analysis/jobs/{jobId}/revision`
- `PUT /api/campaigns/{id}/analysis/jobs/{jobId}/revision/apply`

They can be kept temporarily for backward compatibility but
should be removed in a future release.

### Frontend Changes

#### RevisePhasePage

**Finding grouping**: On load, group items by overlapping
position ranges or shared paragraph boundaries. Most groups
will contain a single item.

**Finding card** (per group):

- Lists all findings in the group with their detection type,
  description, and matched text
- Collapsible "Additional instructions" text field
- "Generate Suggestion" button
- After generation: +/- diff view showing originalSection vs
  revisedSection
- Three action buttons: Accept, Edit, Dismiss
- Accept calls the apply endpoint, shows success indicator,
  and removes the card (or marks it as done)
- Edit opens an inline text editor on revisedSection with a
  "Save & Apply" button
- Dismiss marks all items in the group as dismissed

**Loading state**: Scoped to the individual finding card while
generating. Other findings remain interactive.

**Remove**: The current "Generate Revision" button and
full-document diff modal.

#### Existing Components

The MarkdownEditor component can be reused for the edit
functionality. The diff display pattern from the current
revision view can be adapted for inline per-finding diffs.

### Edge Cases

**Findings without position data**: Some analysis items (like
analysis_report) are general commentary without specific text
positions. These remain as review-only items -- the GM reads
them and acknowledges or dismisses, but there is no text to
revise.

**Overlapping findings**: Grouped and handled in one LLM call.
The GM sees a combined diff and accepts/dismisses the group as
a whole.

**Position shifts after apply**: After accepting a revision,
all subsequent items' positions are shifted by the length
delta. Any previously generated suggestion for shifted items
is cleared, requiring the GM to regenerate if they want to see
the updated diff.

**Empty matched text**: For "add new content" suggestions where
there is no existing text to replace, the diff shows pure
additions at the insertion point. The positionStart and
positionEnd would be equal, indicating an insertion.

**Source content already changed**: If the source content was
modified outside the analysis flow (e.g., direct edit),
position data may be stale. The generation endpoint should
validate that the text at positionStart-positionEnd still
matches matchedText. If not, return an error indicating the
finding is stale.

### Revision Agent Changes

The existing RevisionAgent
(internal/enrichment/revision_agent.go) is replaced with a
simpler per-finding revision function:

- Input: context window (1-3 paragraphs), list of findings
  with their suggestions, optional GM instructions
- System prompt: focused on surgical editing (not full
  document rewriting)
- Max tokens: 1024 (down from 8192)
- No game system YAML or RAG context needed (the finding's
  suggestedContent already contains the relevant analysis)
- Output: the revised section text only

### Migration Path

1. Implement new endpoints and UI alongside existing flow
2. Remove the old "Generate Revision" button and endpoints
3. No database migration needed -- uses existing item position
   fields
