<!--
  Imagineer - TTRPG Campaign Intelligence Platform

  Copyright (c) 2025 - 2026
  This software is released under The MIT License
-->

# Chapter Upgrades Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use
> superpowers:executing-plans to implement this plan
> task-by-task.

**Goal:** Replace the full-screen ChapterEditorPage with a
unified ChapterViewPage that has inline read/edit modes,
PhaseStrip integration, entity/session/relationship sections,
and a new backend endpoint for chapter-scoped relationships.

**Architecture:** New ChapterViewPage inside AppShell follows
the CampaignOverview pattern (read/edit toggle). A new Go
endpoint `GET /chapters/{chapterId}/relationships` joins
`chapter_entities` with `entity_relationships_view` for
one-query relationship fetching. The existing chapter update
API already supports `?analyze=true&phases=...` query params.

**Tech Stack:** React 18, TypeScript, MUI, React Query,
React Router, Go, chi router, pgx/v5, PostgreSQL.

**Design document:**
`docs/plans/2026-02-26-chapter-upgrades-design.md`

---

## Task 1: Backend — Database Function

Add `ListChapterRelationships` to the database layer.

**Files:**

- Modify: `internal/database/relationships.go` (after
  line 262, `GetEntityRelationships`)
- Test: `internal/database/relationships_test.go`

**Step 1: Write the test**

Add a test for `ListChapterRelationships` that verifies
the function returns relationships where at least one
entity is linked to the given chapter. If the test file
does not exist, create it. The test should:

- Set up test data (entities, chapter_entities links,
  relationships) using test helpers or direct DB calls.
- Call `ListChapterRelationships(ctx, campaignID, chapterID)`.
- Assert the returned relationships include both "internal"
  (both entities in chapter) and "one-hop" (one entity in
  chapter) relationships.
- Assert relationships where neither entity is in the
  chapter are excluded.

**Step 2: Run test to verify it fails**

Run: `go test ./internal/database/ -run TestListChapterRelationships -v`
Expected: FAIL (function does not exist yet)

**Step 3: Write the database function**

Add to `internal/database/relationships.go`:

```go
// ListChapterRelationships returns all relationships
// where at least one entity is linked to the chapter
// via chapter_entities.
func (db *DB) ListChapterRelationships(
    ctx context.Context,
    campaignID, chapterID int64,
) ([]models.Relationship, error) {
    query := `
        SELECT DISTINCT ON (erv.id, erv.direction)
            erv.id, erv.campaign_id,
            erv.from_entity_id, erv.to_entity_id,
            erv.relationship_type_id,
            erv.relationship_type, erv.display_label,
            erv.tone, erv.description, erv.strength,
            erv.created_at, erv.updated_at,
            erv.from_entity_name, erv.from_entity_type,
            erv.to_entity_name, erv.to_entity_type,
            erv.direction
        FROM entity_relationships_view erv
        WHERE erv.campaign_id = $1
          AND (
            erv.from_entity_id IN (
                SELECT entity_id FROM chapter_entities
                WHERE chapter_id = $2
            )
            OR erv.to_entity_id IN (
                SELECT entity_id FROM chapter_entities
                WHERE chapter_id = $2
            )
          )
        ORDER BY erv.id, erv.direction,
                 erv.relationship_type,
                 erv.from_entity_name
    `
    rows, err := db.pool.Query(ctx, query,
        campaignID, chapterID)
    if err != nil {
        return nil, fmt.Errorf(
            "list chapter relationships: %w", err)
    }
    defer rows.Close()

    var results []models.Relationship
    for rows.Next() {
        var r models.Relationship
        err := rows.Scan(
            &r.ID, &r.CampaignID,
            &r.SourceEntityID, &r.TargetEntityID,
            &r.RelationshipTypeID,
            &r.RelationshipTypeName, &r.DisplayLabel,
            &r.Tone, &r.Description, &r.Strength,
            &r.CreatedAt, &r.UpdatedAt,
            &r.SourceEntityName, &r.SourceEntityType,
            &r.TargetEntityName, &r.TargetEntityType,
            &r.Direction,
        )
        if err != nil {
            return nil, fmt.Errorf(
                "scan chapter relationship: %w", err)
        }
        results = append(results, r)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf(
            "iterate chapter relationships: %w", err)
    }
    return results, nil
}
```

Reference the existing `GetEntityRelationships` at
`internal/database/relationships.go:224-262` for the scan
pattern. The query uses `entity_relationships_view`
(defined in `migrations/001_schema.sql:1084-1134`) and
filters by `chapter_entities.chapter_id`.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/database/ -run TestListChapterRelationships -v`
Expected: PASS

**Step 5: Commit**

```
git add internal/database/relationships.go \
        internal/database/relationships_test.go
git commit -m "feat: add ListChapterRelationships database function"
```

---

## Task 2: Backend — API Handler and Route

Add the `ListChapterRelationships` handler and register
the route.

**Files:**

- Modify: `internal/api/handlers.go` (after line 2321,
  `ListChapterEntities`)
- Modify: `internal/api/router.go:172` (add route inside
  the `/chapters/{chapterId}` group)

**Step 1: Write the handler**

Add to `internal/api/handlers.go` after `ListChapterEntities`:

```go
// ListChapterRelationships returns relationships where
// at least one entity is linked to the chapter.
func (h *Handler) ListChapterRelationships(
    w http.ResponseWriter, r *http.Request,
) {
    campaignID, err := parseInt64(r, "id")
    if err != nil {
        respondError(w, http.StatusBadRequest, err.Error())
        return
    }

    user := getUserFromContext(r.Context())
    _, err = h.verifyOwnership(r.Context(), campaignID,
        user.ID)
    if err != nil {
        respondError(w, http.StatusNotFound,
            "Campaign not found")
        return
    }

    chapterID, err := parseInt64(r, "chapterId")
    if err != nil {
        respondError(w, http.StatusBadRequest, err.Error())
        return
    }

    // Verify chapter belongs to campaign
    chapter, err := h.db.GetChapter(r.Context(), chapterID)
    if err != nil {
        respondError(w, http.StatusNotFound,
            "Chapter not found")
        return
    }
    if chapter.CampaignID != campaignID {
        respondError(w, http.StatusNotFound,
            "Chapter not found")
        return
    }

    rels, err := h.db.ListChapterRelationships(
        r.Context(), campaignID, chapterID)
    if err != nil {
        respondError(w, http.StatusInternalServerError,
            "Failed to list chapter relationships")
        return
    }

    if rels == nil {
        rels = []models.Relationship{}
    }

    respondJSON(w, http.StatusOK, rels)
}
```

Follow the pattern in `ListChapterEntities`
(`handlers.go:2269-2321`) for campaign/chapter ownership
verification.

**Step 2: Register the route**

In `internal/api/router.go`, inside the
`r.Route("/chapters/{chapterId}", ...)` block, add the
route after line 171 (`r.Get("/sessions", ...)`):

```go
r.Get("/relationships", h.ListChapterRelationships)
```

**Step 3: Run tests**

Run: `go test ./internal/api/ -v`
Expected: PASS (existing tests should still pass)

**Step 4: Manual verification**

Run the server and test:

```
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:3001/api/campaigns/11/chapters/1/relationships
```

Expected: JSON array of relationship objects.

**Step 5: Commit**

```
git add internal/api/handlers.go internal/api/router.go
git commit -m "feat: add GET /chapters/:id/relationships endpoint"
```

---

## Task 3: Client — API and Hook for Chapter Relationships

Add the API service function and React Query hook for
chapter relationships.

**Files:**

- Modify: `client/src/api/chapterEntities.ts` (add
  `listRelationships` method)
- Modify: `client/src/hooks/useChapterEntities.ts` (add
  `useChapterRelationships` hook)
- Modify: `client/src/hooks/index.ts` (export new hook)
- Test: `client/src/hooks/useChapterEntities.test.ts`

**Step 1: Add API method**

In `client/src/api/chapterEntities.ts`, add a new method
to the `chapterEntitiesApi` object:

```typescript
/**
 * List relationships involving entities linked to a
 * chapter.
 */
listRelationships(
    campaignId: number,
    chapterId: number,
): Promise<Relationship[]> {
    return apiClient.get<Relationship[]>(
        `/campaigns/${campaignId}/chapters/${chapterId}/relationships`,
    );
},
```

Add the `Relationship` import to the existing import
from `'../types'`.

**Step 2: Add query keys**

In `client/src/hooks/useChapterEntities.ts`, extend the
`chapterEntityKeys` object (lines 25-30):

```typescript
export const chapterEntityKeys = {
    all: ['chapterEntities'] as const,
    lists: () => [...chapterEntityKeys.all, 'list'] as const,
    list: (campaignId: number, chapterId: number) =>
        [...chapterEntityKeys.lists(), campaignId, chapterId] as const,
    relationships: (campaignId: number, chapterId: number) =>
        [...chapterEntityKeys.all, 'relationships', campaignId, chapterId] as const,
};
```

**Step 3: Add the hook**

In the same file, add after `useChapterEntities`:

```typescript
/**
 * Fetches relationships involving entities linked to a
 * chapter.
 */
export function useChapterRelationships(
    campaignId: number,
    chapterId: number,
) {
    return useQuery({
        queryKey: chapterEntityKeys.relationships(
            campaignId, chapterId),
        queryFn: () => chapterEntitiesApi.listRelationships(
            campaignId, chapterId),
        enabled: !!campaignId && !!chapterId,
    });
}
```

Add `Relationship` to the type import from `'../types'`.

**Step 4: Export from hooks index**

In `client/src/hooks/index.ts`, find the existing
`useChapterEntities` re-export and add
`useChapterRelationships` alongside it.

**Step 5: Write a basic test**

Create `client/src/hooks/useChapterEntities.test.ts`
(if it doesn't exist) with a test verifying the hook
calls the API and returns data. Use the same test
patterns as `client/src/hooks/useAnalysisWizard.test.ts`.

**Step 6: Run tests**

Run: `cd client && npx vitest run src/hooks/useChapterEntities.test.ts`
Expected: PASS

**Step 7: Commit**

```
git add client/src/api/chapterEntities.ts \
        client/src/hooks/useChapterEntities.ts \
        client/src/hooks/index.ts \
        client/src/hooks/useChapterEntities.test.ts
git commit -m "feat: add chapter relationships API and hook"
```

---

## Task 4: ChapterViewPage — Read Mode

Create the ChapterViewPage with read-only sections.

**Files:**

- Create: `client/src/pages/ChapterViewPage.tsx`
- Test: `client/src/pages/ChapterViewPage.test.tsx`

**Step 1: Write the test**

Create `client/src/pages/ChapterViewPage.test.tsx` with
tests covering read mode:

- Renders chapter title and overview.
- Shows "No overview yet" when overview is empty.
- Renders linked entities grouped by mention type.
- Renders child sessions with stage indicators.
- Renders relationships section.
- Shows metadata (sort order, dates).
- Edit button is visible.
- Back button navigates to sessions page.

Mock `useChapter`, `useChapterEntities`,
`useSessionsByChapter`, `useChapterRelationships`, and
`useNavigate`. Use `vi.mock` for hook modules. Reference
`client/src/pages/EntityView.tsx:116-621` for the
structural pattern.

**Step 2: Run tests to verify they fail**

Run: `cd client && npx vitest run src/pages/ChapterViewPage.test.tsx`
Expected: FAIL (component does not exist yet)

**Step 3: Write ChapterViewPage — read mode skeleton**

Create `client/src/pages/ChapterViewPage.tsx`. Follow the
EntityView pattern (`client/src/pages/EntityView.tsx`):

- Use `Container maxWidth="md"` (inside AppShell).
- Parse `campaignId` and `chapterId` from URL params.
- Fetch data via `useChapter`, `useChapterEntities`,
  `useSessionsByChapter`, `useChapterRelationships`.
- State: `isEditing` (boolean, default false).

Header section:

- Back button: `navigate(`/campaigns/${campaignId}/sessions`)`.
- Chapter title as Typography `variant="h4"`.
- Edit button (IconButton with EditIcon).
- Delete button with confirmation dialog. Use
  `useDeleteChapter` mutation. On success, navigate to
  sessions page.

Content sections inside a `Paper` component:

1. **Overview** — Use `MarkdownRenderer` component to
   render `chapter.overview`. If empty, show Typography
   `color="text.secondary"` with "No overview yet."

2. **Entities** — Group `chapterEntities` by
   `mentionType` (featured, linked, mentioned). For each
   entity, render a clickable `Link` to
   `/campaigns/${campaignId}/entities/${entity.entityId}`
   with entity name, a `Chip` for entity type (use the
   same color mapping as EntityView), and a small badge
   for mention type.

3. **Sessions** — Section header: "Sessions (N)". Map
   `sessions` to a list. Each shows title (clickable link
   to `/campaigns/${campaignId}/sessions/${session.id}/edit`),
   `plannedDate` or `actualDate`, and a `Chip` for stage
   (prep/play/wrap_up/completed, same colors as
   `SessionStageIndicator`).

4. **Relationships** — Section header: "Relationships".
   Map `relationships` to rows. Each row shows:
   `sourceEntityName` (linked to entity view if in chapter
   entity set), relationship `displayLabel`,
   `targetEntityName` (linked to entity view if in chapter
   entity set). Entities not in the chapter's entity set
   get a subtle styling (e.g., `color="text.secondary"`,
   `fontStyle: 'italic'`).

   Build a `Set<number>` of chapter entity IDs from
   `chapterEntities.map(ce => ce.entityId)` to determine
   which entities are "internal" vs "external".

5. **Metadata** — Sort order, created date, updated date
   in a `Box` with `Typography variant="caption"`.

**Step 4: Run tests**

Run: `cd client && npx vitest run src/pages/ChapterViewPage.test.tsx`
Expected: PASS

**Step 5: Commit**

```
git add client/src/pages/ChapterViewPage.tsx \
        client/src/pages/ChapterViewPage.test.tsx
git commit -m "feat: add ChapterViewPage read mode"
```

---

## Task 5: ChapterViewPage — Edit Mode with PhaseStrip

Add inline editing and the Save & Analyze flow.

**Files:**

- Modify: `client/src/pages/ChapterViewPage.tsx`
- Modify: `client/src/pages/ChapterViewPage.test.tsx`

**Step 1: Write edit mode tests**

Add tests to ChapterViewPage.test.tsx:

- Clicking Edit switches to edit mode (title becomes
  input, MarkdownEditor appears, PhaseStrip appears).
- Save without phases selected exits edit mode.
- Save with phases navigates to analysis wizard.
- Entity add/remove controls appear in edit mode.
- Cancel/discard exits edit mode without saving.

**Step 2: Run tests to verify they fail**

Run: `cd client && npx vitest run src/pages/ChapterViewPage.test.tsx`
Expected: FAIL (edit mode not implemented yet)

**Step 3: Implement edit mode**

Add to ChapterViewPage:

State additions:

```typescript
const [isEditing, setIsEditing] = useState(false);
const [editTitle, setEditTitle] = useState('');
const [editOverview, setEditOverview] = useState('');
const [editSortOrder, setEditSortOrder] = useState(0);
```

When entering edit mode (`handleToggleEdit`), populate
edit state from current chapter data.

Header changes in edit mode:

- Replace Edit button with PhaseStrip. Import
  `PhaseStrip` from
  `'../components/PhaseStrip/PhaseStrip'` and
  `PhaseSelection` type.
- Check LLM availability via `useUserSettings()`.
  Build `disabledPhases` object when no LLM configured
  (same pattern as `CampaignOverview.tsx:158-165`).
- Wire `PhaseStrip.onSave` to `handleSave`.

Content changes in edit mode:

- Title: `TextField` instead of Typography.
- Overview: `MarkdownEditor` instead of MarkdownRenderer.
  Import from `'../components/MarkdownEditor'`.
- Sort order: `TextField type="number"`.
- Entities: Add an entity autocomplete
  (`EntityAutocomplete` from
  `'../components/EntityAutocomplete'`) and remove
  buttons (IconButton with DeleteIcon) for each entity.
  Wire to `useCreateChapterEntity` and
  `useDeleteChapterEntity`.

`handleSave` function (follow `CampaignOverview.tsx:292-353`):

```typescript
const handleSave = async (phases: PhaseSelection) => {
    const phaseKeys = Object.entries(phases)
        .filter(([, v]) => v)
        .map(([k]) => k);

    const analyze = phases.identify || phases.revise;
    const enrich = phases.enrich;

    const result = await updateChapter.mutateAsync({
        campaignId,
        chapterId,
        input: {
            title: editTitle,
            overview: editOverview,
            sortOrder: editSortOrder,
        },
        options: {
            analyze,
            enrich,
            phases: phaseKeys.length > 0
                ? phaseKeys : undefined,
        },
    });

    // Navigate to wizard if analysis started
    const analysisResult =
        (result as any)?._analysis;
    if (phaseKeys.length > 0 && analysisResult?.jobId) {
        navigate(
            `/campaigns/${campaignId}/analysis/${analysisResult.jobId}`
        );
        return;
    }

    setIsEditing(false);
};
```

Draft management: integrate `useServerDraft` with
`sourceTable: 'chapters'`, same pattern as
`ChapterEditorPage.tsx:137-161`.

**Step 4: Run tests**

Run: `cd client && npx vitest run src/pages/ChapterViewPage.test.tsx`
Expected: PASS

**Step 5: Commit**

```
git add client/src/pages/ChapterViewPage.tsx \
        client/src/pages/ChapterViewPage.test.tsx
git commit -m "feat: add ChapterViewPage edit mode with PhaseStrip"
```

---

## Task 6: Routes and Navigation Updates

Wire the new ChapterViewPage into the router and update
all navigation references.

**Files:**

- Modify: `client/src/App.tsx:177,208-216`
- Modify: `client/src/pages/SessionsManagement.tsx:50-51`
- Modify: `client/src/pages/CampaignDashboard.tsx:320`
- Modify: `client/src/pages/ChapterEditorPage.tsx:325,335`
- Modify: `client/src/components/Chapters/ChapterList.tsx`

**Step 1: Add the view route in App.tsx**

In `client/src/App.tsx`, inside the AppShellWrapper
routes (near line 177 where EntityView is), add:

```tsx
<Route
    path="/campaigns/:campaignId/chapters/:chapterId"
    element={<ChapterViewPage />}
/>
```

Import `ChapterViewPage` at the top.

**Step 2: Remove the edit route**

In `client/src/App.tsx`, delete the edit route at
lines 213-216:

```tsx
// DELETE THIS:
<Route path="/campaigns/:campaignId/chapters/:chapterId/edit"
    element={<ChapterEditorPage />} />
```

Keep the `/chapters/new` route (lines 208-211).

**Step 3: Update SessionsManagement navigation**

In `client/src/pages/SessionsManagement.tsx`, change
`handleEditChapter` (line 50-52) to navigate to the
view page:

```typescript
const handleEditChapter = (chapter: Chapter) => {
    navigate(
        `/campaigns/${campaignId}/chapters/${chapter.id}`
    );
};
```

Rename to `handleViewChapter` for clarity. Update the
prop passed to ChapterList at line 95.

**Step 4: Update CampaignDashboard navigation**

In `client/src/pages/CampaignDashboard.tsx`, update the
`onEditChapter` handler at line 320 to navigate to the
view page instead of the editor.

**Step 5: Update ChapterEditorPage post-create navigation**

In `client/src/pages/ChapterEditorPage.tsx`, change the
post-create navigation (lines 325, 335) from
`/chapters/${newChapter.id}/edit` to
`/chapters/${newChapter.id}` (the view page):

```typescript
navigate(
    `/campaigns/${campaignId}/chapters/${newChapter.id}`,
    { replace: true }
);
```

**Step 6: Update ChapterList**

In `client/src/components/Chapters/ChapterList.tsx`:

- Change the chapter title/row click to call
  `onEditChapter` (or rename to `onViewChapter`).
  Currently `onSelectChapter` just sets selection state
  for the sessions panel. The title click should
  navigate to the view page.
- Remove the expand/collapse inline preview (the
  collapsible overview snippet).
- Remove the edit icon button (editing happens on the
  view page).
- Keep the delete button.

**Step 7: Run all tests**

Run: `cd client && npx vitest run`
Expected: PASS (all 26+ test files pass)

Fix any broken tests that reference the old edit route
or removed ChapterList behaviors.

**Step 8: Commit**

```
git add client/src/App.tsx \
        client/src/pages/SessionsManagement.tsx \
        client/src/pages/CampaignDashboard.tsx \
        client/src/pages/ChapterEditorPage.tsx \
        client/src/components/Chapters/ChapterList.tsx
git commit -m "feat: wire ChapterViewPage routes and update navigation"
```

---

## Task 7: Final Integration and Cleanup

Run all tests, verify manually, and update docs.

**Files:**

- Modify: `CHANGELOG.md`
- Modify: `Todo.md`

**Step 1: Run full test suite**

Run: `cd client && npx vitest run`
Expected: all tests pass

Run: `go test ./...`
Expected: all tests pass

**Step 2: Manual smoke test**

Start the server and client. Verify:

1. Sessions page: clicking a chapter title navigates to
   the Chapter View page.
2. Chapter View: overview renders with wiki links,
   entities grouped by mention type, sessions listed
   with stage chips, relationships section shows
   connected entities.
3. Edit mode: clicking Edit shows TextFields and
   MarkdownEditor, PhaseStrip appears.
4. Save & Analyze: selecting phases and saving navigates
   to the analysis wizard.
5. New chapter: `/chapters/new` still works, creates and
   navigates to the view page.
6. Delete: confirmation dialog works, navigates back to
   sessions.

**Step 3: Update CHANGELOG.md**

Add entries under `[Unreleased]`:

### Added

- Chapter View page at
  `/campaigns/:id/chapters/:chapterId` with read mode
  showing overview, linked entities grouped by mention
  type, child sessions with stage indicators,
  relationships between chapter entities plus one hop,
  and metadata.
- Inline editing on the Chapter View page with
  PhaseStrip integration matching the CampaignOverview
  pattern.
- `GET /chapters/:chapterId/relationships` backend
  endpoint returning relationships scoped to a
  chapter's entity set.

### Changed

- Chapter navigation updated: clicking a chapter in the
  list navigates to the View page instead of the editor.
- ChapterEditorPage retained only for creating new
  chapters.
- ChapterList simplified: expand/collapse inline preview
  and edit icon removed.

**Step 4: Update Todo.md**

Check off the "Read-only chapter view page" item in
the MVP Backlog section and add any new items discovered
during implementation.

**Step 5: Commit**

```
git add CHANGELOG.md Todo.md
git commit -m "docs: update CHANGELOG and Todo for chapter upgrades"
```

---

## Task Summary

| # | Task | Type | Estimated Scope |
|---|------|------|-----------------|
| 1 | DB function | Backend | ~60 lines Go |
| 2 | API handler + route | Backend | ~50 lines Go |
| 3 | API + hook | Frontend | ~50 lines TS |
| 4 | ChapterViewPage read | Frontend | ~400 lines TSX |
| 5 | ChapterViewPage edit | Frontend | ~200 lines TSX |
| 6 | Routes + navigation | Frontend | ~30 lines changes |
| 7 | Integration + docs | Both | Testing + docs |

Dependencies: Task 1 before Task 2. Task 2 before
Task 3. Task 3 before Task 4. Task 4 before Task 5.
Task 5 before Task 6. Task 6 before Task 7.
