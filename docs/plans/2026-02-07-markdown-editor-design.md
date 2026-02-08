# Markdown Editor and Entity Analysis System

This design document describes the replacement of the current
HTML-based rich text editor with a Markdown-native editor. The
system stores content as Markdown throughout the entire stack:
editor, API, database, and vectorization pipeline. A wiki-link
syntax enables entity references, and a post-save analysis system
detects entities and suggests enrichments.

This design addresses three problems:

- HTML markup inflates token counts in embedding chunks, which
  degrades vector search quality.
- Entity detection was previously conceived as automatic and
  real-time, but the feature works better as a deliberate
  post-save step.
- Entity references in text need a clear delimiter format because
  entity names can span multiple words.

## Markdown Editor Architecture

The `RichTextEditor` component gives way to a `MarkdownEditor`
that uses TipTap configured for Markdown-native editing. TipTap
remains the editor framework, but the storage format changes from
HTML to Markdown.

### Editor configuration

The editor uses the following TipTap extensions:

- `@tiptap/starter-kit` provides core block and inline nodes.
- `@tiptap/extension-placeholder` provides placeholder text
  (already in use).
- A custom `WikiLink` node extension recognizes
  `[[Entity Name]]` and `[[Entity Name|display text]]` syntax
  and renders wiki links as styled inline chips in the editor.

The toolbar provides these formatting controls:

- Bold, Italic, and Strikethrough for inline formatting.
- H2 and H3 for section headings.
- Bullet List and Ordered List for list structures.
- Blockquote and Horizontal Rule for content organization.
- A Wiki Link insert button for entity references.

### Storage format

The editor serializes content to Markdown on every change via a
TipTap Markdown serializer (such as the `tiptap-markdown` package
or a custom implementation). The `onChange` callback emits a
Markdown string. All API calls send Markdown. All `TEXT` columns
in the database store Markdown.

### Display format

When the application renders stored content outside the editor
(entity previews, campaign overview read mode, search result
snippets), a lightweight Markdown-to-HTML renderer such as
`react-markdown` converts the Markdown for display. The renderer
handles wiki links and presents them as clickable entity links.

### Data migration

Existing HTML content does not need preserving. A fresh database
rebuild handles the format change.

## Wiki Link System

The wiki link system enables users to reference entities directly
within Markdown content using a double-bracket syntax.

### Syntax

Wiki links use `[[Entity Name]]` or
`[[Entity Name|display text]]` syntax. A custom TipTap node
extension called `WikiLink` parses this format.

### Editor behavior

Wiki links render as inline chips that are visually distinct from
surrounding text (for example, with a subtle background color and
an entity-type icon). The chip displays the display text, or the
entity name when no alias is present. Clicking a chip opens the
entity preview popup. Backspace removes the entire chip as a
single unit.

### Inserting wiki links

Users can insert wiki links using two methods:

- The toolbar button opens a small entity search popover. The
  user types to fuzzy-search existing entities, selects one, and
  optionally sets display text. The editor inserts
  `[[Entity Name]]` at the cursor position.
- Typing `[[` triggers inline autocomplete after the user types
  at least 3 characters inside the brackets (debounced at 300ms).
  A dropdown displays matching entities via fuzzy search through
  `pg_trgm`. Selecting an entity closes the brackets. When no
  match exists, the text remains as-is and post-save analysis
  picks up the reference.

For alias links, the user types `|` after the entity name:
`[[Professor Armitage|the old professor]]`. Autocomplete still
operates on the portion before the pipe character.

### Markdown storage

Wiki links are stored literally as `[[Entity Name]]` or
`[[Entity Name|display text]]` in the Markdown string. The system
applies no transformation to these references during storage.

### Read-mode rendering

The Markdown renderer includes a custom plugin that finds
`[[...]]` patterns and renders them as clickable links. Clicking
a link navigates to the entity editor. Unresolved links (where
the entity is not found) render in a warning style, such as red
text or a dashed underline.

### Entity renames

When a user renames an entity, the backend updates all
`[[Old Name]]` and `[[Old Name|` references across all Markdown
content in that campaign to `[[New Name]]` and `[[New Name|`
respectively. The update runs within the same transaction as the
rename and is scoped to the campaign.

### Index requirement

The implementation must verify or add a composite index and a
trigram GIN index on `entities(campaign_id, name)` for fast fuzzy
autocomplete queries. Autocomplete requires a minimum of 3 typed
characters before the system issues a query.

## Post-Save Entity Analysis

The post-save analysis system examines saved Markdown content and
suggests entity links and enrichments through a two-phase funnel.

### Trigger

When the user saves any Markdown content field (entity
description, chapter overview, session notes, or campaign
description), the backend kicks off analysis as an asynchronous
job.

### Phase 1: Identification (fast, database-only)

The backend scans the saved Markdown and returns a list of
detected items:

- Wiki links `[[...]]` that are resolved, fuzzy-matched, or
  unresolved.
- Untagged text matching existing entity names via fuzzy
  `pg_trgm` search.
- Potential misspellings of existing entities (high similarity
  but not exact matches).

Each item appears in the triage UI with a simple decision:

- Accept link confirms that the text references that entity.
  The system fixes spelling if needed and adds `[[brackets]]`
  when untagged. The item clears from the list and enters the
  Phase 2 queue.
- New entity indicates the text is not a match but represents
  a new entity. The user provides an entity type. The item
  clears and enters the Phase 2 queue as a creation task.
- Not an entity dismisses the item entirely with no further
  processing.

### Phase 2: Enrichment (async, LLM-powered)

Only accepted or confirmed entities enter the enrichment queue.
For each entity, the LLM receives the saved content plus the
entity's current state and suggests:

- Description updates based on new context.
- Entity log entries describing what happened to the entity in
  the content.
- New relationships to other entities mentioned nearby.

These enrichment suggestions appear in the triage UI as they
complete, forming a second wave of items to address. The user
can accept, edit, or decline each suggestion.

No LLM calls are wasted on declined matches. The user's fast
decisions in Phase 1 filter what receives expensive analysis in
Phase 2.

## Triage UI

The triage interface presents detected entities and suggested
enrichments for user review.

### Layout

The triage view uses a full-screen layout via the existing
`FullScreenLayout` component. The header displays the source
context (for example, "Entities found in: Chapter 3 Overview")
along with a "Skip for now" button and a progress indicator
("4 of 7 resolved").

### Left panel: Item list

A vertical list of detected items appears in two sections:

- The Identification section (Phase 1) contains items needing
  a quick decision. Each item shows the matched text, the
  suggested entity (with a confidence percentage for fuzzy
  matches), and the original context snippet. Three action
  buttons appear per item: Accept Link, New Entity, and Not An
  Entity.
- The Enrichment section (Phase 2) contains items that have
  passed Phase 1 and received LLM suggestions. These items
  appear progressively as results arrive. A subtle loading
  indicator shows how many items the system is still analyzing.

Items disappear from the list (or display a checkmark) as the
user resolves them. Unresolved items remain at the top.

### Right panel: Detail view

When the user clicks an item that requires more than a quick
button press, the right panel displays the detail work:

- Accept Link with a spelling correction shows the diff
  ("Professer -> Professor") and the text in context. The user
  can confirm or adjust the correction.
- New Entity displays a creation form with fields for entity
  name, type selector, and an optional initial description.
  The form is lightweight and does not replicate the full entity
  editor.
- Description Update (Phase 2) shows the current entity
  description alongside the LLM's suggested revision as a diff.
  The user can accept, edit, or decline the suggestion.
- Log Entry (Phase 2) shows the suggested log text. The user
  can accept, edit, or decline the entry.
- Relationship (Phase 2) shows the two entities and the
  suggested relationship type. The user can accept the
  suggestion, change the type, or decline.

### Re-entry

When the user clicks "Skip for now," the triage view closes and
the application returns to the previous location. An unobtrusive
badge appears on the source content (for example, a small
indicator on the chapter card or entity row) showing the count of
pending suggestions. Clicking the badge reopens the triage view
with all unresolved items intact.

## Data Model Changes

This section describes the database schema changes required for
the entity analysis system.

### Markdown content columns

The existing `TEXT` columns require no schema changes. The columns
`entities.description`, `campaigns.description`,
`chapters.overview`, `sessions.prep_notes`,
`sessions.actual_notes`, and `campaign_memories.content` store
Markdown instead of HTML. The column type remains the same; only
the content format changes.

### New table: content_analysis_jobs

The `content_analysis_jobs` table tracks each post-save analysis
run.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | Primary key. |
| source_table | TEXT | The table name, e.g., 'chapters'. |
| source_id | UUID | The row that triggered the analysis. |
| source_column | TEXT | The column name, e.g., 'overview'. |
| campaign_id | UUID | Foreign key to campaigns. |
| status | TEXT | One of: pending, analyzing, complete. |
| created_at | TIMESTAMPTZ | When the save triggered the job. |

### New table: content_analysis_items

The `content_analysis_items` table stores individual suggestions
within a job.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | Primary key. |
| job_id | UUID | FK to content_analysis_jobs. |
| phase | TEXT | identification or enrichment. |
| item_type | TEXT | See item types below. |
| matched_text | TEXT | The text found in content. |
| entity_id | UUID | FK to entities (nullable). |
| suggested_entity_name | TEXT | For new entity suggestions. |
| suggested_entity_type | TEXT | For new entity suggestions. |
| suggested_content | JSONB | Diff, description, or details. |
| confidence | FLOAT | Fuzzy match confidence (0-1). |
| context_snippet | TEXT | Surrounding text for display. |
| status | TEXT | pending, accepted, or declined. |
| resolved_at | TIMESTAMPTZ | When the user acted on it. |

Valid `item_type` values are: `wiki_link`, `untagged_match`,
`spelling`, `new_entity`, `description_update`, `log_entry`,
and `relationship`.

### New table: entity_log

The `entity_log` table records events that happen to entities
across the campaign.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | Primary key. |
| entity_id | UUID | Foreign key to entities. |
| campaign_id | UUID | Foreign key to campaigns. |
| chapter_id | UUID | Foreign key to chapters (nullable). |
| session_id | UUID | Foreign key to sessions (nullable). |
| source_table | TEXT | Where the event was discovered. |
| source_id | UUID | Which row contains the source. |
| content | TEXT | What happened to the entity. |
| occurred_at | TEXT | Free-form in-world date (nullable). |
| sort_order | INT | Chronological ordering (nullable). |
| created_at | TIMESTAMPTZ | When the log entry was created. |

The `occurred_at` field uses free text to support fantasy
calendars without requiring a world calendar system. The
`sort_order` integer handles chronological ordering independently
of date format.

## API Changes

This section describes the new and modified API endpoints for the
Markdown editor and entity analysis system.

### Modified endpoints (Markdown storage)

The existing CRUD endpoints require no structural changes. These
endpoints already accept and return `TEXT` content. The only
difference is that the content is now Markdown instead of HTML.
The API contract (JSON field names and types) remains the same.

### New endpoints for entity analysis

The following endpoints support the analysis workflow.

`GET /campaigns/{id}/analysis/jobs`
:   List analysis jobs, filterable by status.

`GET /campaigns/{id}/analysis/jobs/{jobId}`
:   Get a job with all of its items.

`PATCH /campaigns/{id}/analysis/items/{itemId}`
:   Resolve an item: accept, decline, or edit.

`POST /campaigns/{id}/analysis/items/{itemId}/enqueue`
:   Push an accepted Phase 1 item into the Phase 2
    LLM queue.

### New endpoints for entity log

The following endpoints manage entity log entries.

`GET /campaigns/{id}/entities/{entityId}/log`
:   Get log entries sorted by sort_order.

`POST /campaigns/{id}/entities/{entityId}/log`
:   Create a log entry.

`PUT /campaigns/{id}/entities/{entityId}/log/{logId}`
:   Edit a log entry.

`DELETE /campaigns/{id}/entities/{entityId}/log/{logId}`
:   Delete a log entry.

### New endpoint for wiki link resolution

The following endpoint supports editor autocomplete.

`GET /campaigns/{id}/entities/resolve?name={text}`
:   Fuzzy-match an entity name and return candidates with
    confidence scores. The endpoint requires a minimum of
    3 characters.

### Analysis trigger

The existing save endpoints (`PUT` for entities, chapters,
sessions, and campaigns) gain a side effect: after a successful
save of a Markdown content field, the backend creates a
`content_analysis_jobs` row and runs Phase 1 identification. The
save response includes the job ID so the client can navigate to
the triage view.

## Client Component Changes

This section describes the React component additions, removals,
and modifications required for the Markdown editor system.

### Components to remove

The following components and utilities are no longer needed:

- `RichTextEditor/RichTextEditor.tsx` is replaced entirely by
  the new Markdown editor.
- `RichTextEditor/EditorToolbar.tsx` is replaced by the new
  toolbar component.
- `utils/sanitizeHtml.ts` is no longer needed because the system
  no longer handles HTML content.
- The `dompurify` dependency is removed.

### New components

The implementation introduces these new components:

- `MarkdownEditor/MarkdownEditor.tsx` wraps TipTap configured
  with Markdown serialization. The component uses the same prop
  interface as the old editor (`value: string`,
  `onChange: (markdown: string) => void`) so that pages require
  minimal changes. The component includes the WikiLink custom
  node.
- `MarkdownEditor/EditorToolbar.tsx` provides toolbar buttons
  for bold, italic, strikethrough, H2, H3, bullet list, ordered
  list, blockquote, horizontal rule, and wiki link insert.
- `MarkdownEditor/WikiLinkNode.tsx` implements the custom TipTap
  node extension. The extension parses `[[...]]` syntax, renders
  styled chips in the editor, and provides inline autocomplete
  when the user types `[[` followed by at least 3 characters.
- `MarkdownRenderer/MarkdownRenderer.tsx` renders stored Markdown
  as HTML for read-mode display. The component wraps
  `react-markdown` with a custom plugin for wiki link rendering.
  All content display areas outside the editor use this
  component.
- `EntityAnalysis/TriageView.tsx` implements the full-screen
  triage view using `FullScreenLayout`. The component contains
  the item list and detail panel.
- `EntityAnalysis/TriageItemList.tsx` renders the left-panel
  checklist of Phase 1 and Phase 2 items with action buttons.
- `EntityAnalysis/TriageDetailPanel.tsx` renders the right-panel
  detail view for diffs, entity creation forms, and relationship
  management.

### Modified pages

The following pages require updates:

- `EntityEditor.tsx`, `ChapterEditorPage.tsx`, and
  `CampaignOverview.tsx` swap the `RichTextEditor` import to
  `MarkdownEditor`. The prop interface is identical, so the
  changes are minimal.
- Any component that uses `sanitizeHtml()` or
  `dangerouslySetInnerHTML` replaces that usage with a
  `<MarkdownRenderer content={...} />` element.

### New dependencies

The implementation adds these packages:

- `tiptap-markdown` (or an equivalent) provides Markdown
  serialization for TipTap.
- `react-markdown` provides Markdown-to-HTML rendering for read
  mode.

### Removed dependencies

The implementation removes `dompurify` because the system no
longer processes HTML content.

## Implementation Phases

The project is divided into four sequential phases.

### Phase 1: Markdown Editor

This phase establishes the core Markdown editing capability.

- Replace `RichTextEditor` with `MarkdownEditor`.
- Configure TipTap for Markdown serialization.
- Implement the toolbar with bold, italic, strikethrough,
  headers, lists, blockquote, and horizontal rule controls.
- Remove `sanitizeHtml`, DOMPurify, and
  `dangerouslySetInnerHTML` usage.
- Add the `MarkdownRenderer` component for read-mode display.
- Add the `react-markdown` dependency and remove `dompurify`.
- Update all pages to use the new components.
- Revisit vectorization chunk sizes because Markdown is leaner
  than HTML.

### Phase 2: Wiki Links

This phase adds entity reference support within Markdown content.

- Implement the `WikiLinkNode` TipTap extension.
- Add `[[` autocomplete with a 3-character minimum, debounced at
  300ms.
- Add the toolbar wiki link insert button.
- Implement the `GET /campaigns/{id}/entities/resolve` endpoint.
- Verify or add the entity name index for fast fuzzy search.
- Add the `MarkdownRenderer` wiki link plugin for clickable
  entity links in read mode.
- Implement entity rename propagation across campaign content.

### Phase 3: Post-Save Analysis (Identification)

This phase introduces the Phase 1 analysis pipeline and triage
interface.

- Create the `content_analysis_jobs` and
  `content_analysis_items` tables.
- Add save-endpoint triggers for Phase 1 analysis (wiki link
  resolution, untagged entity detection, and spelling fuzzy
  match).
- Build the triage full-screen view with the item list and
  detail panel.
- Implement accept, decline, and new entity actions.
- Add "Skip for now" functionality with a badge indicator for
  pending items.
- Queue accepted items for Phase 2 enrichment processing.

### Phase 4: LLM Enrichment

This phase adds LLM-powered enrichment suggestions.

- Integrate LLM processing for entity discovery from content.
- Implement description update suggestions with a diff view.
- Add relationship suggestions between co-occurring entities.
- Create the `entity_log` table with free-text dates and sort
  order.
- Add log entry suggestions from the LLM.
- Implement progressive loading in the triage UI as results
  arrive.

## Testing Strategy

This section defines the test coverage required for each major
component of the system.

### MarkdownEditor component tests

The following tests validate the Markdown editor component:

- The component renders with an initial Markdown value.
- Toolbar buttons toggle formatting (bold wraps the selection
  in `**`).
- The `onChange` callback emits valid Markdown on edits.
- The wiki link node renders `[[Entity Name]]` as a chip.
- Wiki link autocomplete triggers after `[[` plus 3 characters.
- Wiki link autocomplete does not trigger with fewer than 3
  characters.
- Alias syntax `[[Name|display]]` renders correctly.
- The component handles empty and null content gracefully.

### MarkdownRenderer component tests

The following tests validate the Markdown renderer component:

- The renderer displays headings, bold, italic, lists,
  blockquotes, and horizontal rules correctly.
- Wiki links render as clickable entity links.
- Unresolved wiki links render in a warning style.
- The component handles empty and null content without crashing.
- The renderer blocks XSS via crafted Markdown (no raw HTML
  pass-through).

### Wiki link resolution endpoint tests

The following tests validate the fuzzy search endpoint:

- An exact match returns a confidence of 1.0.
- A fuzzy match returns candidates ranked by similarity.
- Misspellings return the correct entity with lower confidence.
- Campaign scoping excludes entities from other campaigns.
- Empty or short queries return empty results.

### Post-save analysis tests

The following tests validate the analysis pipeline:

- Saving content triggers job creation.
- The system detects and matches wiki links in the content.
- The system detects untagged entity names with fuzzy matching.
- Accepting an item changes the item status and enqueues the
  item for Phase 2.
- Declining an item removes the item from the pending list.
- Selecting "New entity" creates the entity and enqueues the
  entity for enrichment.
- "Skip for now" preserves all pending items.

### Entity log tests

The following tests validate entity log operations:

- CRUD operations on log entries work correctly.
- Query results respect the sort order.
- Chapter and session associations are optional.
- Log entries are scoped to the campaign.

### Integration tests

The following tests validate end-to-end workflows:

- The full flow works: save Markdown, the system creates an
  analysis job, items appear, the user accepts an item,
  enrichment is queued, a suggestion appears, the user accepts
  the suggestion, and the entity is updated.
- An entity rename propagates through wiki links across all
  content in the campaign.
- Vectorization produces cleaner chunks from Markdown than from
  HTML.
