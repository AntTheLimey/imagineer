<!--
  Imagineer - TTRPG Campaign Intelligence Platform

  Copyright (c) 2025 - 2026
  This software is released under The MIT License
-->

# Phase Screens Design

This design document describes splitting the monolithic
AnalysisTriagePage (4,400 lines handling all three phases
with conditional rendering) into a step-by-step wizard
where each phase gets a dedicated screen. The new structure
isolates each step, enables phase-specific UX, and
addresses several issues discovered during manual testing.

## Goal

The current AnalysisTriagePage handles all three analysis
phases (Identify, Revise, Enrich) in a single component
with complex conditional rendering. This design replaces
that monolith with three separate page components inside a
shared wizard shell. The wizard enforces the phase order
(Identify, Revise, Enrich) and provides navigation,
progress tracking, and auto-advancement.

## Routing

The wizard uses nested routes to direct the user to the
appropriate phase screen.

The following routes are defined:

```
/campaigns/:campaignId/analysis/:jobId
/campaigns/:campaignId/analysis/:jobId/identify
/campaigns/:campaignId/analysis/:jobId/revise
/campaigns/:campaignId/analysis/:jobId/enrich
```

The base URL (without a phase suffix) redirects to the
first active phase. Only phases selected on the PhaseStrip
receive routes.

## AnalysisWizard Shell

A thin parent component replaces AnalysisTriagePage as the
route target. The shell provides the following elements:

- The header displays the job title, source name, and a
  phase stepper showing progress (for example,
  "Identify [12/15] -> Revise [0/3] -> Enrich [pending]").
- The phase stepper bar uses a horizontal MUI Stepper
  showing only the selected phases. The current phase
  appears highlighted, and completed phases display a
  checkmark.
- The content area renders the active phase page via
  nested routes.
- Navigation buttons include "Continue to [next phase]"
  and "Back to Overview."
- Auto-advancement triggers when the GM resolves the last
  pending item in a phase. A brief snackbar
  ("All items resolved!") appears, and the wizard
  auto-navigates to the next phase after 1.5 seconds. A
  manual "Continue" button remains always visible for
  advancing with unresolved items.
- An error banner appears when the job status is "failed"
  and `failure_reason` is set. The banner shows a
  prominent alert explaining the issue (quota exceeded,
  rate limited, or another error).

## Phase 1: Identify

The Identify phase links text to existing entities without
using the LLM or creating entities (except for unresolved
wiki-links). The phase performs pure pattern matching.

### Left Panel

The left panel shows an item list grouped by detection
type with counts:

- `wiki_link_resolved` (green) items are already linked.
  The GM confirms these items, and an "Accept All" batch
  action is available.
- `untagged_mention` (blue) items indicate text that
  matches an existing entity.
- `potential_alias` (purple) items suggest the text might
  be an alias for an existing entity.
- `misspelling` (orange) items indicate a possible
  misspelled entity name.
- `wiki_link_unresolved` (amber) items use wiki-link
  syntax but have no matching entity.

### Right Panel

The right panel shows a detail view with a context
snippet, the matched entity chip, a similarity score, and
action buttons.

### Actions

The following actions are available in the Identify phase:

- Accept (all types with a matched entity) links the text
  to the entity and inserts a wiki-link.
- Reassign (all types) opens an entity autocomplete picker
  to redirect the match to a different entity when the
  original match is wrong.
- Create Entity (unresolved wiki-links only) presents an
  inline form with the name pre-filled from the wiki-link
  text, an entity type picker, and an optional short
  description. A note states: "A full description will be
  generated during the Enrich phase." The system does not
  re-scan the new entity description.
- Dismiss ignores the suggestion.
- Accept All batch-accepts an entire detection type group
  (for example, all resolved wiki-links).

### Completion

The wizard auto-advances when the GM resolves all items.
The GM can also click "Continue" to advance with pending
items. Unresolved items remain pending rather than being
dismissed.

## Phase 2: Revise

The Revise phase handles content quality review and
iterative revision. The GM reviews TTRPG Expert and Canon
Expert findings here. The phase supports multiple revision
cycles.

### Left Panel

The left panel contains two sections.

The first section shows analysis findings grouped by
detection type:

- `analysis_report` items are general analysis reports.
- `content_suggestion` items propose content improvements.
- `mechanics_warning` items flag game mechanics issues.
- `investigation_gap` items identify missing investigation
  threads.
- `pacing_note` items comment on session pacing.
- `canon_contradiction` items flag conflicts with
  established canon.
- `temporal_inconsistency` items flag timeline problems.
- `character_inconsistency` items flag character behavior
  contradictions.

Each finding displays severity indicators and inline
Acknowledge/Dismiss buttons.

The second section shows new mentions that appear after
applying a revision. Any new entity mentions found in the
revised content use the same grouping and actions as the
Identify phase (Accept, Reassign, Create Entity for
unresolved wiki-links, and Dismiss).

### Right Panel

The right panel shows a detail view for the selected
analysis finding or new mention.

### Revision Workflow

The revision workflow appears at the top of the page and
includes the following elements:

- A "Generate Revision" button triggers the revision
  agent.
- A side-by-side diff view compares the original content
  with the revised version.
- An edit mode allows the GM to modify the revision before
  applying the revision.
- An "Apply Revision" button writes the revised content
  and triggers automatic re-identification on the new
  content.
- An iteration counter ("Revision 1", "Revision 2", and
  so on) tracks the cycle count via the `revision_count`
  field on drafts.
- After applying a revision, the new mentions section
  populates. The GM resolves the new mentions and then
  either generates another revision cycle or continues to
  the Enrich phase.

### Revision Cycle

The GM follows this cycle: acknowledge findings, generate
a revision, review and edit the diff, apply the revision
(which triggers re-identification), resolve new mentions,
and then repeat or continue to the Enrich phase.

## Phase 3: Enrich

The Enrich phase builds out entity details with LLM
assistance. A two-pass approach gives the GM control over
token spend.

### Pass 1: Entity Scan

Pass 1 is a cheap scan that runs automatically. A single
lightweight LLM call examines the content and returns a
checklist of entities that could benefit from enrichment.

The checklist includes two categories:

- Existing entities list the name, what is improvable
  (description, relationships, or both), and a one-line
  reason.
- New entity suggestions list the name, type, and a
  one-line reason for creation.

The left panel shows this checklist with checkboxes. The
GM selects which entities to enrich. An "Enrich Selected"
button triggers Pass 2 for the checked entities.
Individual "Enrich" buttons per entity are also available.

### Pass 2: Per-Entity Enrichment

Pass 2 runs on demand. The system makes one LLM call per
selected entity and produces the following results:

- `description_update` items suggest description
  improvements.
- `relationship_suggestion` items propose new
  relationships to create.
- `log_entry` items provide event log entries.
- The system analyses existing relationships for that
  entity, suggesting removals for redundant edges,
  suggesting alterations for wrong types or targets, and
  validating type pairs against constraints.

Results stream in and appear in the left panel grouped
under each entity.

### Graph Health

After the selected enrichments complete, the system runs
graph health checks:

- Orphan detection identifies entities with no
  relationships.
- Redundant edge detection scans across the full graph.

The results appear as graph advisory items.

### Right Panel Detail Views

The right panel adapts to the selected item type:

- Description updates show a diff view that the GM can
  edit before accepting, allowing the GM to fix
  inaccuracies.
- Relationship suggestions display a visual diagram with
  type autocomplete. The GM can accept or dismiss the
  suggestion.
- Existing relationship issues show the current
  relationship with the suggested change (remove, alter
  type, or swap direction).
- Log entries display the content and timestamp.
- New entity suggestions show the name, type, and
  description, all of which are editable before creation.
  The system auto-tags descriptions with wiki-links for
  exact entity name matches when the GM accepts the
  suggestion.
- Graph advisory items show a description and
  recommendation. The GM can acknowledge or dismiss the
  advisory.

### Per-Entity Error Handling

Since enrichment runs per-entity, a failure on one entity
does not block other entities. Failed entities display a
red error badge with the failure reason. The GM can retry
the failed entity later.

## Error Handling and LLM Quota

This section describes how the system handles LLM quota
exhaustion and rate limiting.

### Backend

The backend distinguishes quota errors from rate limits
in LLM retry logic:

- HTTP 402 (Anthropic quota): the system fails
  immediately and tags the error as `quota_exceeded`.
- HTTP 429 with a quota message body (OpenAI): the system
  detects the quota condition via the response body and
  fails immediately instead of retrying.
- HTTP 429 rate limit: the system retries with backoff as
  the current implementation does today.

A new `failure_reason TEXT` column on
`content_analysis_jobs` stores the failure reason on the
job record (for example, "API quota exceeded" or "Rate
limited after retries"). The API returns the failure
reason in the job response.

### Frontend

The AnalysisWizard shell shows an alert banner when the
job status is "failed":

- Quota exceeded: "Your API token limit has been reached.
  Please check your account with [provider] and try again
  later."
- Rate limited: "The AI service is temporarily
  unavailable. Please try again in a few minutes."
- Other errors: "Enrichment encountered an error:
  [reason]."

Per-entity enrichment failures show a red badge on the
specific entity with a retry option. The `useResolveItem`
hook receives an `onError` callback that shows a Snackbar
with the error message.

## Shared Infrastructure

The phase screens share several hooks, contexts, and
components.

### Hooks and Context

The following hooks and contexts support the wizard:

- `useAnalysisWizard` is a new hook providing job data,
  items, phase navigation state, and advancement logic.
- `AnalysisWizardContext` is a React context providing
  filtered items and actions to the phase pages.
- The existing hooks (`useResolveItem`,
  `useBatchResolve`, `useRevertItem`,
  `useGenerateRevision`, and `useApplyRevision`) remain
  unchanged.

### Shared Components

The following components are shared across phases:

- `ItemDetailPanel` is the right-panel detail view,
  parameterized by item type.
- `EntityAutocomplete` is the entity picker for the
  Reassign action.
- `DetectionTypeBadge` already exists and all phases
  share the badge component.

### File Structure

The new file structure organizes the wizard as follows:

```
client/src/pages/
  AnalysisWizard.tsx          (shell with stepper)
  IdentifyPhasePage.tsx       (Phase 1)
  RevisePhasePage.tsx         (Phase 2)
  EnrichPhasePage.tsx         (Phase 3)
client/src/hooks/
  useAnalysisWizard.ts        (shared wizard state)
  useContentAnalysis.ts       (existing, unchanged)
```

### Migration

The existing `AnalysisTriagePage.tsx` (4,400 lines) is
replaced by the four new files listed above. The old file
is deleted after the migration is complete.

### Database

One new column is required:

```sql
ALTER TABLE content_analysis_jobs
    ADD COLUMN failure_reason TEXT;
```

## Issues Addressed

This design addresses the following items from manual
testing:

- Issue (a): LLM quota and limit errors are silent. The
  design adds quota detection, the `failure_reason`
  column, alert banners, per-entity error handling, and
  the `useResolveItem` `onError` callback.
- Issue (b): The graph expert receives incomplete data.
  The Enrich phase now analyses existing relationships
  per-entity, and the graph health step runs with full
  relationship data.
- Issue (c): Entity description auto-tagging and
  pre-acceptance editing are missing. The Enrich phase
  adds edit-before-accept on descriptions and new
  entities. The system auto-tags descriptions with
  wiki-links on accept. The Identify phase adds a
  Create Entity form for unresolved wiki-links.

## Phase Order Rationale

The phase order is Identify, Revise, then Enrich for the
following reasons:

- Identify finds entities without using the LLM.
- Revise reviews and rewrites content before the system
  enriches entities. This approach avoids wasting tokens
  on content that changes during revision.
- Revision can go through multiple cycles. The system
  runs re-identification automatically after each
  revision apply.
- Enrich runs last on stable content. The two-pass
  approach gives the GM control over token spend.
