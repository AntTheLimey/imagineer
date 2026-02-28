<!--
  Imagineer - TTRPG Campaign Intelligence Platform

  Copyright (c) 2025 - 2026
  This software is released under The MIT License
-->

# Chapter Upgrades Design

This design document describes replacing the separate
full-screen ChapterEditorPage with a unified
ChapterViewPage that supports both read and edit modes.
The new page matches the CampaignOverview pattern and adds
a backend endpoint for chapter-scoped relationships.

## Goals

This upgrade achieves the following objectives:

- Chapters gain a proper View page showing an overview,
  entities, sessions, and relationships.
- The edit pattern matches CampaignOverview with inline
  editing and a PhaseStrip for Save and Analyze.
- The separate ChapterEditorPage is retired for existing
  chapters.
- A simplified creation route remains for new chapters.

## ChapterViewPage

The ChapterViewPage replaces both the read-only chapter
display and the full-screen editor with a single unified
component.

### Route

The page uses the following route, rendered inside AppShell
(not full-screen) to match the EntityView pattern:

```
/campaigns/:campaignId/chapters/:chapterId
```

### Read Mode

Read mode is the default state when a user navigates to
a chapter.

The header bar contains the following elements:

- A back button navigates to the Sessions page.
- The chapter title displays as an h4 heading.
- An Edit button and a Delete button (with a confirmation
  dialog) provide management actions.

The content area displays five sections.

The Overview section renders chapter content with the
MarkdownRenderer component, supporting wiki links. An
empty state shows a "No overview yet" placeholder when
no content exists.

The Entities section groups entities by mention type:
Featured first, then Linked, then Mentioned. Each entity
displays a clickable name linking to the entity view, a
type chip, and a mention type badge.

The Sessions section lists child sessions with a title,
date, and stage indicator chip. Each session links to the
session editor when clicked. The section header displays
a count of child sessions.

The Relationships section shows relationships where both
entities link to the chapter, plus one hop outward
(relationships where one entity belongs to the chapter
and the other does not). Each row displays the source
entity, the relationship type, and the target entity.
Non-chapter entities display with a subtle "external"
visual indicator to distinguish them from chapter
entities.

The Metadata section shows the sort order, the created
date, and the updated date.

### Edit Mode

The Edit button in the header activates edit mode. The
following elements change when entering edit mode:

- The title becomes a TextField.
- The overview switches from MarkdownRenderer to
  MarkdownEditor.
- The sort order becomes editable.
- The entity section gains add and remove controls
  through manual entity management via autocomplete.
  AI-powered entity detection is deferred to a later
  enhancement.
- The PhaseStrip appears in the header, replacing the
  Edit button.
- The Delete button remains visible.

### Save and Analyze Flow

The Save and Analyze flow matches CampaignOverview
exactly. The user follows these steps:

1. Click the Edit button and modify the title, overview,
   or entities.
2. Review the PhaseStrip, which shows Identify, Revise,
   and Enrich checkboxes. The Revise and Enrich
   checkboxes are disabled when no LLM is configured
   (checked via `useUserSettings`).
3. Click "Save & Go" on the PhaseStrip.

The `handleSave` function calls `updateChapter` with
`options: { analyze, enrich, phases }` as parameters.
The backend returns `_analysis.jobId` in the response.

When the user selected phases and the backend returns a
job ID, the system navigates to
`/campaigns/:id/analysis/:jobId`. When no phases are
selected, the system performs a plain save and exits
edit mode.

Draft management uses `useServerDraft`, matching the
current ChapterEditorPage behavior.

No backend changes are needed for this flow. The chapter
update handler already supports `?analyze=true`,
`?enrich=true`, and `?phases=...` query parameters.

## New Backend Endpoint

The system requires one new endpoint to support the
chapter-scoped relationship view.

### GET /campaigns/:id/chapters/:chapterId/relationships

This endpoint returns relationships where at least one
entity links to the chapter via `chapter_entities`. The
endpoint enables the "chapter entities plus one hop"
relationship view without N+1 client-side API calls.

The response format matches the existing entity
relationships endpoint: an array of relationship objects
containing the source entity, the target entity, and the
relationship type.

The implementation joins `chapter_entities` with
`entity_relationships_view` where either
`source_entity_id` or `target_entity_id` belongs to
the chapter's entity set.

## Routes and Navigation

This section describes the routing changes for the
chapter upgrade.

### New Routes

The following route is added:

- `/campaigns/:campaignId/chapters/:chapterId` renders
  the ChapterViewPage inside AppShell.

### Kept Routes

The following route remains unchanged:

- `/campaigns/:campaignId/chapters/new` renders a
  simplified creation form. After creation, the system
  navigates to the new chapter's View page.

### Retired Routes

The following route is removed:

- `/campaigns/:campaignId/chapters/:chapterId/edit` is
  removed because editing is now inline on the View page.

### Navigation Updates

The following navigation targets change:

- ChapterList item click navigates to the chapter view
  instead of the editor. Clicking a chapter title or row
  in the list navigates to the View page. The current
  expand/collapse inline preview and the edit icon button
  are removed; all chapter interaction starts from the
  View page.
- ChapterCard click navigates to the chapter view.
- The analysis wizard "back to source" link navigates to
  the chapter view.

## Retired Components

This section describes components that the upgrade
retires or preserves.

- ChapterEditorPage is retired for existing chapters.
  The creation flow for `/chapters/new` can reuse a
  stripped-down version or become a new
  ChapterCreatePage.
- The ChapterEditor dialog remains for quick-create from
  the ChapterList header button.
- ChapterEntityPanel (AI entity detection) is not
  currently used by ChapterEditorPage. AI entity
  detection in chapters is deferred to a later
  enhancement.

## Data Requirements

This section describes the hooks and backend functions
that the ChapterViewPage consumes.

### Existing Hooks

The following existing hooks require no changes:

- `useChapter(campaignId, chapterId)` fetches chapter
  data.
- `useChapterEntities(campaignId, chapterId)` fetches
  entity links.
- `useSessionsByChapter(campaignId, chapterId)` fetches
  child sessions.
- `useUpdateChapter()` saves with analysis options.
- `useServerDraft()` manages draft state.
- `useUserSettings()` checks LLM availability.

### New Hooks and Functions

The upgrade requires the following additions:

- `useChapterRelationships(campaignId, chapterId)` is a
  new React hook for the relationships endpoint.
- `ListChapterRelationships(ctx, campaignID, chapterID)`
  is a new Go database function.
- A `ListChapterRelationships` handler registers in the
  router to serve the new endpoint.

## Deferred Work

The following features are explicitly deferred to later
enhancements:

- AI entity detection panel (ChapterEntityPanel)
  integration into edit mode is deferred.
- Relationship editing from the chapter view is deferred.
  The chapter view displays relationships as read-only.
