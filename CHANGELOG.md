# Changelog

All notable changes to the Imagineer project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

- Analysis Wizard (Phase Screens)
  - Replaced the monolithic 4,400-line AnalysisTriagePage
    with a step-by-step wizard where each analysis phase
    (Identify, Revise, Enrich) has its own focused screen
    accessible via nested routes under
    `/campaigns/:id/analysis/:jobId/`.
  - `useAnalysisWizard` hook provides phase filtering via
    detection type groups, phase navigation, and
    auto-advance with a 1500 ms timer when all items in a
    phase are resolved.
  - `AnalysisWizard` shell component renders an MUI Stepper,
    error banner (quota, rate limit, and general errors),
    loading spinner, navigation buttons, and auto-advance
    snackbar.
  - `IdentifyPhasePage` displays a two-column layout with
    items grouped by five detection types, color dots and
    counts, a detail panel with context highlighting, entity
    chip display, a create entity form, and an Accept All
    batch action per detection type group.
  - `EntityAutocomplete` component provides debounced
    (300 ms) MUI Autocomplete for entity search with type
    chips.
  - `RevisePhasePage` groups analysis findings by eight
    detection types with severity indicators, a revision
    workflow (Generate, diff view, edit, Apply with
    iteration counter), and a new mentions section for
    identification items created after revision apply.
  - `DiffView` component renders side-by-side diffs with
    line-by-line comparison, changed-line highlighting
    (red/green), and optional inline editing.
  - `EnrichPhasePage` provides a two-pass enrichment UI
    (Start Enrichment, polling, results) with items
    grouped by eight enrichment detection types,
    type-specific detail views, cancel enrichment support,
    and edit-before-accept for description updates.
  - `autoTagWikiLinks` utility wraps untagged entity names
    in `[[wiki-link]]` syntax with case-insensitive
    matching and skip-already-tagged logic.
  - Graph Health summary section displays a consolidated
    view of graph advisory items with distinctive visual
    treatment.
  - 276 tests across 26 test files cover the wizard
    components.
- Entity Suggestion Quality
  - The LLM prompt for new entity detection now
    requests two to three sentence descriptions
    covering the entity's role, relationships,
    characteristics, and actions from the source
    content.
  - When creating an entity from a
    new_entity_suggestion, the handler now extracts
    and passes the description from suggestedContent
    to the entity record.
  - When creating an entity from an enrichment-phase
    suggestion without position offsets, the handler
    performs a global find-and-replace to insert
    [[wiki link]] tags for all occurrences of the
    entity name in the source content.
- RAG Context for Enrichment and Revision Agents
  - The enrichment agent now performs a
    per-entity hybrid vector search and
    includes campaign context and game system
    schema in its LLM prompt for richer entity
    descriptions and relationship suggestions.
  - The revision agent now receives campaign
    search results and game system schema from
    the handler, enabling context-aware content
    revisions.
- Job Phases Persistence
  - The `job_phases` junction table persists which pipeline
    phases (Identify, Revise, Enrich) the GM selected when
    triggering analysis, so the triage page shows only the
    relevant sections and the GM can resume later.
  - The `current_phase` column on `content_analysis_jobs`
    tracks which phase the job is currently executing.
  - The phase-aware triage page conditionally shows analysis,
    detection, and enrichment sections based on the job's
    selected phases, with backward compatibility for
    pre-existing jobs.
  - The `parsePhases` server-side helper validates phase
    values against an allowlist before storage, preventing
    arbitrary data from reaching the database.
  - Migration 003 adds CHECK constraints on `phase_key`,
    `current_phase`, and `drafts.status` columns.
- Draft Lifecycle Tracking
  - The `revision_count` column on drafts increments
    atomically on each save for engagement tracking.
  - The `status` column on drafts (`active`, `stale`,
    `abandoned`) supports a cleanup index for future
    maintenance jobs.
- Client Phases Integration
  - PhaseStrip selection flows through the API as a
    comma-separated `phases` query parameter on all update
    endpoints (campaigns, entities, chapters, sessions).
  - The `ContentAnalysisJob` type gains `phases` and
    `currentPhase` fields.
  - All update hooks (`useCampaigns`, `useChapters`,
    `useEntities`, `useSessions`) accept an optional `phases`
    parameter.
- Campaign Deletion UI
  - "Delete Campaign" danger zone on Campaign Settings with a
    confirmation dialog requiring the user to type the campaign
    name. On success, the system clears campaign context and
    navigates to `/campaigns`.
- TTRPG Expert Scope Awareness
  - `SourceScope` enum on `PipelineInput` derived from
    `SourceTable` (campaigns, chapters, sessions, entities).
    TTRPG Expert and Canon Expert prompts now vary analysis
    and continuity guidance based on content level
    (campaign = world-building, chapter = narrative arc,
    session = tactical detail, entity = character
    consistency).
- Phase Strip UI
  - `PhaseStrip` component replaces `SaveSplitButton` on
    Campaign Overview. Three independent workflow phases
    (Identify, Revise, Enrich) display as labelled checkboxes
    with a Save & Go button. Phase selection persists in
    `localStorage` for downstream use on the triage page.
- Session Workflows Phase 1
  - Full-screen session editor page with two-column prep
    layout (form and scene sidebar), stage tabs (Prep active,
    Play/Wrap-up/Completed as Phase 2 placeholders), autosave,
    draft recovery, and unsaved changes protection.
  - Scene CRUD with full-stack implementation: database table,
    Go model, handlers, routes, React API client, hooks, and
    inline add/edit/delete in the session editor sidebar.
  - Database migration 011 adds the scenes table,
    session_chat_messages table, and play_notes column on
    sessions. The migration also adds vectorization for the
    new tables, updates the `search_campaign_content()`
    function, and drops legacy JSONB columns (planned_scenes,
    discoveries, player_decisions, consequences).
  - Session types updated with a completed stage, removed
    legacy fields, and added playNotes.
  - SessionsManagement page now navigates to the full-screen
    session editor instead of opening a dialog.
  - Chapter editor import button for pasting content into the
    overview field.
  - 28 SessionEditorPage tests covering create and edit modes,
    stage tabs, scene sidebar, breadcrumbs, and user
    interaction.
- Session Workflows Phase 2 — Play Mode
  - Multi-panel Play mode layout replacing the Phase 2
    placeholder: SceneStrip, SceneViewer,
    PlayEntityDrawer, PlayScratchpad,
    PlayEntitySidebar, and ImportNotesDialog.
  - SceneStrip provides horizontal scene navigation with
    status cycling (planned, active, completed, skipped)
    and scene type icons.
  - SceneViewer displays active scene content and prep
    notes in three modes (scene, notes, mixed) with
    wiki-link entity navigation.
  - PlayEntityDrawer shows read-only entity details with
    full description, GM notes, and relationships.
  - PlayScratchpad provides a free-form auto-saving
    textarea for in-game tracking (HP, initiative,
    conditions, loot).
  - PlayEntitySidebar collapses between a 40px icon
    strip with badge counts and a 220px grouped entity
    list.
  - ImportNotesDialog supports importing notes via paste
    or .txt/.md file upload with append/replace modes.
  - SaveSplitButton gains a defaultMode prop so Play
    mode defaults to plain Save (no analysis on
    scratchpad saves).
  - Wrap-up and Completed stage placeholders updated
    from "Coming in Phase 2" to "Coming in Phase 3".
- Save Options (SaveSplitButton)
  - SaveSplitButton component offering three save modes: Save,
    Save & Analyze, and Save, Analyze & Enrich.
  - The backend gates analysis behind `?analyze=true&enrich=true`
    query params on entity, chapter, campaign, and session
    update endpoints.
  - Zero-item analysis results show a green success snackbar
    ("Analysis complete: no issues found") instead of silent
    nothing.
- Entity View Page
  - Read-only entity detail page at
    `/campaigns/{id}/entities/{entityId}` showing entity
    header, description, GM notes, attributes, relationships,
    event log, and metadata.
  - Wiki links and the entity list view icon now navigate to
    the entity view page instead of opening a dialog or
    navigating to edit.
- Multi-Agent Enrichment Pipeline (Phases 1-6)
  - Pipeline infrastructure with `PipelineAgent` interface
    (`Name()`, `Run()`, `DependsOn()`), topological sort
    orchestrator for dependency ordering, shared
    `PipelineInput` context, and `EnrichmentAgent` adapter
    wrapping the existing engine as a pipeline stage.
  - Migration 013 adds `agent_name` and `pipeline_run_id`
    columns to `content_analysis_items` for multi-agent
    tracking.
  - TTRPG Expert agent analyses content quality across eight
    dimensions: pacing, investigation (Three Clue Rule),
    spotlight balance, NPC development, mechanics validation,
    PC agency, continuity, and setting.
  - TTRPG Expert produces a holistic report and atomic triage
    items with new detection types (`analysis_report`,
    `content_suggestion`, `mechanics_warning`,
    `investigation_gap`, `pacing_note`) using
    `phase="analysis"` to distinguish analysis output from
    enrichment output.
  - Frontend analysis groups in the triage UI with an
    expandable report view for TTRPG Expert results.
  - Canon Expert agent detects contradictions against
    established campaign facts via RAG, covering factual,
    temporal, and character inconsistencies with new detection
    types (`canon_contradiction`, `temporal_inconsistency`,
    `character_inconsistency`).
  - Frontend canon items in the triage UI with established
    fact and conflicting text rendering.
  - Revision agent generates revised content incorporating
    accepted Stage 1 findings, with two new API endpoints
    (`POST /revision` and `PUT /revision/apply`).
  - Frontend side-by-side diff view for revisions (original
    vs revised) with edit mode and apply workflow using
    green/red line highlighting.
  - Graph Expert agent validates relationships after entity
    enrichment with dual mode: structural checks (orphans,
    type pair constraints) without LLM and semantic checks
    (redundant and implied edges) with LLM.
  - Migration 014 adds the `relationship_type_constraints`
    table with seed data for entity type pair validation.
  - Graph Expert introduces new detection types
    (`graph_warning`, `redundant_edge`, `invalid_type_pair`,
    `orphan_warning`) and depends on the enrichment stage
    via `DependsOn()=["enrichment"]`.
  - ContextBuilder for RAG context assembly using
    content-derived queries (extracts multiple search queries
    from content summary and entity name batches of five),
    deduplication by `(SourceTable, SourceID)` keeping the
    highest combined score, and token budget tracking (4000
    token soft limit with `estimateTokens` heuristic).
  - ContextBuilder wired into all three pipeline handler
    sites with campaign lookup for game system code and
    schema loading via `loadGameSystemSchema`.
  - 36 context tests covering query derivation,
    deduplication, token budgets, and schema loading.
- Content Enrichment Improvements
  - `RunContentEnrichment` method scans content for entity
    mentions independently of Phase 1 analysis, enabling
    discovery of new relationships and descriptions on demand.
  - The enrichment progress indicator on the triage page shows
    real-time status while enrichment runs in the background.
  - Auto-advance to next pending item after accepting or
    dismissing a suggestion on the triage page.
  - Accepted description updates and log entries now apply to
    entities immediately.
- Wiki Link Navigation
  - Wiki links navigate directly to the entity view page
    instead of the entities list.
  - The hover popover shows entity type chip, description
    snippet, and View link with 300 ms open / 200 ms close
    delays.
- Duplicate Relationship Prevention
  - Unique constraint on `(campaign_id, source_entity_id,
    target_entity_id, relationship_type)` prevents duplicate
    relationships.
  - `CreateRelationship` uses `ON CONFLICT DO UPDATE` for
    idempotent upsert.
  - `GetEntityRelationships` query deduplicates inverse
    relationship pairs.
- API Key Encryption at Rest
  - AES-256-GCM encryption for user API keys stored in
    `user_settings` (content generation, embedding, and image
    generation keys).
  - `internal/crypto` package with `Encryptor` struct providing
    `Encrypt`/`Decrypt` methods with random nonces and `enc:`
    prefix for encrypted value identification.
  - Transparent encryption in the database layer — API handlers
    and LLM providers see plaintext, no changes needed.
  - `ENCRYPTION_KEY` environment variable (64 hex chars) enables
    encryption; server warns on startup if not set.
  - Migration 007 NULLs existing plaintext API keys (users
    re-enter via Account Settings).
- LLM Enrichment (Phase 4 — Post-Save Content Analysis)
  - LLM client abstraction supporting Anthropic Claude, OpenAI
    GPT-4o, and Ollama local models with exponential backoff retry.
  - Enrichment engine analyses content for entity description
    updates, log entries, and relationship suggestions.
  - Auto-triggers enrichment when Phase 1 triage items are fully
    resolved (requires LLM configured in Account Settings).
  - SSE streaming and polling fallback for progressive enrichment
    delivery to the triage UI.
  - Triage UI enrichment section with side-by-side description
    diff, log entry preview, and relationship suggestion chips.
  - Relationship type autocomplete with fuzzy-match suggestions
    and inline create-new-type dialog in the triage UI.
  - Auto-resolve inverse relationship suggestions when accepting
    directional relationships.
  - `headquartered_at`/`headquarters_of` system relationship
    type added via migration.
- Entity Log (Chronological Event History)
  - `entity_log` table tracking events per entity with optional
    chapter/session association, occurred_at date, and sort order.
  - Full CRUD API endpoints under
    `/campaigns/{id}/entities/{entityId}/log`.
  - Entity editor log section for viewing, adding, editing, and
    deleting log entries.
- Custom Ollama Docker Image
  - Pre-bakes `mxbai-embed-large` embedding model at build time,
    eliminating post-startup model download.
- Wiki Links (`[[Entity Name]]` syntax)
  - `[[Entity Name]]` and `[[Entity Name|display text]]` wiki-link
    syntax in the Markdown editor and renderer.
  - Inline autocomplete triggered by `[[` with fuzzy entity name
    matching via `pg_trgm`.
  - Toolbar insert button for wiki-link insertion via entity
    search.
  - Clickable wiki links in read mode navigate to entity view.
  - Entity rename propagation updates all wiki links across
    campaign content (entities, chapters, sessions, memories,
    timelines, relationships) in a single transaction.
  - Entity resolve API endpoint
    `GET /campaigns/{id}/entities/resolve` for fuzzy name
    matching.
- Ollama Embedding Pipeline (Local-First Semantic Search)
  - Ollama container in Docker Compose for local embedding
    generation using `mxbai-embed-large` model (1024 dimensions).
  - Vectorization extended to chapters, sessions, and campaign
    memories with a hybrid search function.
  - Hybrid search API endpoint `GET /campaigns/{id}/search`
    combining vector similarity (70%) with BM25 text search (30%).
  - Entity detection upgraded to use vector search with
    text-based fallback.
  - Ollama added as embedding service option in Account Settings
    (no API key required for local operation).
- Campaign Description Vectorization
  - Campaign descriptions are now vectorized and searchable
    via the hybrid semantic search endpoint.
  - `search_campaign_content()` SQL function returns campaign
    description matches with `source_table='campaigns'`.
  - Chunk size tuned to 200 tokens for campaign descriptions
    (HTML content requires smaller chunks than plain text).
- Integration Tests for Embedding Pipeline
  - Build-tagged (`integration`) tests verifying vectorization
    availability, chunk creation, search results, and campaign
    description search.
  - `make test-integration` target for running embedding tests
    against live Docker services.
- Chapters and Sessions Management
  - Chapters table, enhanced sessions table with stages, and
    AI memory system tables.
  - Chapter model and CRUD operations in Go backend (`internal/database/chapters.go`).
  - Session model updated with ChapterID, Title, and Stage workflow fields.
  - REST API endpoints for chapters: GET/POST `/campaigns/{id}/chapters`,
    GET/PUT/DELETE `/campaigns/{id}/chapters/{chapterId}`.
  - REST API endpoints for sessions: GET/POST `/campaigns/{id}/sessions`,
    GET/PUT/DELETE `/campaigns/{id}/sessions/{sessionId}`,
    GET `/campaigns/{id}/chapters/{chapterId}/sessions`.
  - Session workflow stages: prep (blue), play (green), wrap_up (orange).
  - React types, API services, and hooks for chapters and sessions.
  - ChapterList component with expand/collapse and delete confirmation.
  - ChapterEditor dialog for creating/editing chapters.
  - SessionList component with stage indicators and quick actions.
  - SessionCard with visual stage indicator and metadata display.
  - SessionEditor dialog with stage navigation and field editing.
  - SessionStageNav for navigating between prep/play/wrap-up stages.
  - SessionStageIndicator showing current stage with appropriate colors.
  - SessionsView integrated into Campaign Dashboard with chapter/session panels.
- App Navigation Restructure
  - HomePage (`HomePage.tsx`) with smart campaign redirect logic - redirects
    to current or latest campaign overview, shows welcome screen if none.
  - CampaignOverview (`CampaignOverview.tsx`) with read-first design and
    inline field editing for name, description, genre, and image style prompt.
  - CreateCampaign (`CreateCampaign.tsx`) full-page form with validation.
  - NoCampaignSelected (`NoCampaignSelected.tsx`) welcome/onboarding screen
    with feature highlights and campaign cards for existing users.
  - Sessions placeholder page (`Sessions.tsx`) showing planned features.
  - App.tsx routing updated to use AppShell consistently with campaign-centric
    URL structure (`/campaigns/{id}/overview`, `/campaigns/{id}/sessions`).
  - Legacy route redirects: `/campaigns` redirects to `/`, `/campaigns/{id}/dashboard`
    redirects to `/campaigns/{id}/overview`.
- Evernote Importer Improvements
  - Evernote 10.x version detection distinguishes between Evernote Legacy (7.x)
    with full AppleScript support and Evernote 10.x with limited support.
  - Helpful error messages guide users to export ENEX files or install
    Evernote Legacy when Evernote 10.x is detected.
- GM Session Patterns Knowledge Base
  - `.claude/ttrpg-expert/gm-session-patterns.md` documents GM preparation
    techniques including Lazy DM's Eight Steps, Three Clue Rule, Five Room
    Dungeon, and Fronts/Clocks systems.
- AI Memory System Design
  - `docs/memory-system-design.md` documents three-tier memory architecture
    (Campaign, Chapter, Session) with token compression and entity extraction.
- Favicon (`client/public/favicon.svg`) added for browser tab branding.

- UX Foundation (Phase 1)
  - Three-panel layout component (`ThreePanelLayout.tsx`) with responsive
    behavior: desktop (all 3 panels), tablet (drawer for left), mobile (bottom
    navigation tabs).
  - Full-screen editor layout (`FullScreenLayout.tsx`) with back navigation,
    breadcrumbs, dirty state indicator, and action buttons.
  - App shell (`AppShell.tsx`) with header, campaign dropdown, and user menu.
  - Campaign context (`CampaignContext.tsx`) tracks current campaign with
    localStorage persistence.
  - Draft context (`DraftContext.tsx`) manages localStorage drafts for data
    loss prevention.
  - `useDraft` hook for localStorage draft storage with server version tracking.
  - `useAutosave` hook saves drafts every 30 seconds automatically.
  - `useUnsavedChanges` hook tracks dirty state and blocks navigation with
    confirmation dialog.
  - Full-screen entity editor (`EntityEditor.tsx`) with draft recovery,
    autosave, and duplicate detection.
  - Route structure updated to support full-screen editor views outside the
    main app shell.
- RPG terminology knowledge base (`.claude/ttrpg-expert/rpg-terminology.md`)
  for TTRPG expert agent with comprehensive terminology reference covering GM
  titles across systems.
- Multi-User Foundation & Authentication
  - Users table and owner_id foreign key on campaigns table
    (included in the consolidated schema migration).
  - User model and database operations in Go backend handle user creation,
    retrieval, and campaign ownership.
  - Google OAuth authentication flow enables sign-in and sign-up via Google.
  - JWT token generation and validation secure API requests.
  - Authentication middleware extracts user from JWT and attaches the user
    to the request context.
  - Campaign ownership and data isolation ensure users only see their own
    campaigns.
  - React AuthContext manages client-side authentication state.
  - Login page provides Google sign-in button and OAuth flow initiation.
  - OAuth callback handling completes the authentication flow and stores
    tokens.
  - Protected routes in React app redirect unauthenticated users to login.
  - User display in navigation shows avatar, name, and logout button.
- First analysis agent: consistency-checker
  - Agent interface (`internal/agents/agent.go`) defines Result, Suggestion,
    and Source structs for all analysis agents.
  - Consistency checker (`internal/agents/consistency/checker.go`) implements
    five data integrity checks:
    - Orphaned entities detection finds entities with no relationships or
      timeline references.
    - Duplicate name detection uses pg_trgm similarity threshold of 0.7.
    - Timeline conflict detection identifies entities in multiple events at
      the same time.
    - Invalid reference detection finds broken relationship pointers.
    - Sessions without discoveries identifies completed sessions that added
      no entities.
  - Database consistency queries (`internal/database/consistency.go`) provide
    efficient PostgreSQL queries for all checks.
  - API endpoint POST `/api/campaigns/{id}/agents/consistency-check` exposes
    the checker via the REST API.
  - Comprehensive unit tests cover both the checker and database functions.
- 15 new backend API endpoints for complete frontend-backend integration:
  - Game system lookup by ID and code
  - Entity search by name (fuzzy matching)
  - Relationship CRUD and entity relationship queries
  - Timeline event CRUD and entity timeline queries
  - Campaign-specific statistics
- Fully functional Entities page with:
  - Entity list with type filtering and pagination
  - Create dialog with duplicate detection
  - View/Edit dialogs with validation
  - Delete with confirmation
  - Tags, GM notes, and source confidence tracking
- Fully functional Timeline page with:
  - Vertical timeline layout with colored date indicators
  - Filter panel (date precision, visibility, date range)
  - Sort toggle (newest/oldest first)
  - Create/Edit/Delete dialogs with entity linking
- Import page connected to API with:
  - Campaign selector
  - Loading states and error handling
  - Import result summary display
  - Multi-file upload support
- Server configuration with testable constants (`cmd/server/config.go`)
- Database entity tests for array handling (`entities_test.go`)
- Client environment example file (`client/.env.example`)

### Fixed

- Migration 005 adds `failure_reason` TEXT column to
  `content_analysis_jobs` for storing the reason a job
  failed. The `SetJobFailureReason` helper stores the
  failure reason atomically with `status='failed'`.
- LLM quota detection: a new `QuotaExceededError` type
  distinguishes HTTP 402 quota exceeded from 429 rate
  limiting, and `doWithRetry` fails immediately on quota
  errors instead of retrying.
- `PipelineInput.Relationships` now populates from the
  database before pipeline invocation, eliminating
  false-positive orphan warnings from the Graph Expert.
- `CreateAnalysisJob` now wraps job and phase inserts in a
  database transaction for atomicity.
- Missing `rows.Err()` check after iterating phase rows in
  `GetAnalysisJob` and `GetLatestAnalysisJob`; extracted a
  shared `loadJobPhases` helper to eliminate duplication.
- Migration backfill adds `enrich` and `revise` phases for
  existing jobs that already have enrichment or analysis
  items, preventing those items from being hidden on the
  triage page.
- Markdown paste in TipTap editor rendering raw Markdown
  syntax (headings, lists, blockquotes, horizontal rules) as
  plain text instead of formatted content. A new
  `MarkdownPasteHandler` extension intercepts plain-text
  paste events containing block-level Markdown patterns and
  converts them to rich content.
- Cross-campaign IDOR in `fetchSourceContent` queries; all
  content retrieval now scopes by `campaign_id` to prevent
  unauthorized access to other campaigns' data (VULN-004).
- Request body size limits added to all API endpoints that
  accept JSON bodies, preventing oversized payloads via
  `http.MaxBytesReader` with a 1 MB limit (VULN-001).
- GM notes stripped from entity objects before passing to
  the enrichment pipeline, preventing sensitive GM-only
  content from leaking into LLM prompts (VULN-003).
- Background enrichment goroutines now use a 10-minute
  timeout instead of unbounded `context.Background()`,
  preventing resource leaks from stalled LLM calls
  (VULN-005).
- Entity type validated against known constants when
  resolving `new_entity` analysis items, rejecting invalid
  types with HTTP 400 (VULN-006).
- Byte-based string truncation in Canon Expert and Graph
  Expert prompts replaced with rune-based truncation,
  preventing split multi-byte UTF-8 characters
  (BUG-002/003).
- Failed enrichment pipeline runs now set job status to
  `'failed'` instead of `'completed'`, correctly reflecting
  the error state (MAJOR-005).
- Path traversal vulnerability in `loadGameSystemSchema`; the
  function now validates file paths before loading schema files.
- Rune-safe truncation in TTRPG Expert prompts preventing
  invalid UTF-8 sequences when truncating multi-byte content.
- `allKnownEntities` bug where `DetectNewEntities` did not see
  all campaign entities; the function now receives the complete
  entity list.
- Missing `acknowledged` resolution in backend validation; the
  resolver now accepts the `acknowledged` value.
- Unused `campaignId` prop removed from the CompletedStage
  component.
- Redundant `@types/diff` devDependency removed from the
  client package.
- Enrichment skipping when triggered via "Save, Analyze &
  Enrich" button; the system now runs `RunContentEnrichment`
  independently of Phase 1 analysis results.
- Triage page showing "no items to review" while enrichment
  was still running; the page now shows a progress indicator.
- Accepted description updates and log entries silently
  discarded instead of applied to entities.
- Auto-advance after accepting a triage item not respecting
  grouped entity display order.
- View icon on entities list opening a popup dialog instead of
  navigating to the entity view page.
- Current description invisible in dark theme on triage UI;
  fixed by using theme-aware background color.
- Suggested description blank in MarkdownEditor due to TipTap
  value initialization timing; fixed by passing value directly
  as a prop.
- Analysis badge missing enrichment-phase pending items on
  campaign overview; fixed by removing the phase filter from
  `CountPendingAnalysisItems`.
- MCP server authentication configured via `INIT_TOKENS` in
  Docker Compose.
- Entity API URL routes for get/update/delete now correctly use
  `/entities/{id}` instead of `/campaigns/{campaignId}/entities/{id}`.
- Tags input in entity forms now captures typed text on blur (added
  `autoSelect` and `onBlur` handler to MUI Autocomplete).
- Import routes now use RESTful pattern `/api/campaigns/{id}/import/*`
  instead of `/api/import/*` with campaignId in body
- Server default port changed from 8080 to 3001 to avoid conflict with
  MCP server
- Vite proxy now correctly targets port 3001
- Entity creation bug caused by pgx/pq library incompatibility - removed
  pq.Array() calls and use pgx native array handling
- Timeline page crash when event.entityIds is undefined (CodeRabbit)
- Timeline page forEach lint error from implicit return (CodeRabbit)
- Campaign scoping security issue in relationship and timeline handlers -
  now verify resources belong to the campaign specified in URL (CodeRabbit)
- Worked around pgedge_vectorizer UUID primary key bug: chunk
  tables now correctly use UUID `source_id` columns instead of
  BIGINT, matching Imagineer's UUID primary keys. Applied via
  SQL hotfix pending upstream fix in pgedge_vectorizer.
- Fixed `search_campaign_content()` SQL function referencing
  non-existent column `c.chunk` (correct column is `c.content`
  in pgedge_vectorizer chunk tables).

### Changed

- Replaced the monolithic AnalysisTriagePage (4,755 lines)
  with the AnalysisWizard component and three focused phase
  pages (IdentifyPhasePage, RevisePhasePage,
  EnrichPhasePage). The old triage page and its test file
  are deleted.
- CampaignOverview builds a phases array from PhaseStrip
  selection and passes the array through the API; removed
  localStorage-based phase persistence.
- Extracted shared utilities to reduce code duplication
  (-198 net lines): `agents.StripCodeFences` (was 4 copies),
  `agents.TruncateString` (was 3 copies), shared
  `fetchSourceContent` (was 2 copies), and
  `buildDefaultPipeline` factory (was 3 copies).
- Squashed 10 incremental database migrations into two files:
  `001_schema.sql` (complete schema with 20 tables, indexes,
  triggers, functions, views, and vectorization) and
  `002_seed_data.sql` (game systems and default relationship
  types). The database is still in active development, so no
  data preservation was needed.
- Renamed `keeper_notes` column to `gm_notes` throughout the
  codebase, moving from Call of Cthulhu-specific to generic RPG
  terminology.
- GM notes filtered at API level; non-campaign-owners receive empty `gmNotes`
  field in entity responses.
- Client-side GM notes UI hidden for non-owners using `useCampaignOwnership`
  hook.
- CLAUDE.md Task Workflow now includes step 6 "Document" for automatic
  CHANGELOG.md and Todo.md updates after completing user-facing changes
- Added "Tracking Files" subsection to Documentation guidelines
- Primary agent role now enforces mandatory delegation (never writes code
  directly)
- Added mcp-server-expert and design-compliance-validator sub-agents
- Added documentation synchronization requirements
- Added database test requirement ("Do not skip database tests")
- Added copyright notice requirement for source files
- Upgraded PostgreSQL from 17 to 18 (pgedge-postgres:18-spock5-standard)
- CI workflow now tests against PostgreSQL 17 and 18 (was 16 and 17)
- Added PostgreSQL extensions auto-configuration on container startup:
  - pg_trgm, vector, pgedge_vectorizer, pg_tokenizer, vchord_bm25,
    vectorize, pgmq, pg_cron
- Added `make test-db` to verify database extensions are installed

### Added

#### API Implementation
- Go backend API server with chi router and middleware (CORS, logging, timeout)
- Database package (`internal/database/`) with PostgreSQL connection pooling via pgx/v5
- Model structs (`internal/models/`) matching the database schema
- API handlers (`internal/api/`) for campaigns, entities, relationships, timeline, import
- REST API endpoints: game-systems, campaigns CRUD, entities CRUD, relationships, timeline, stats, import
- React API client layer (`client/src/api/`) with typed fetch wrapper
- React Query hooks (`client/src/hooks/`) for all API domains
- Loading and error states in Dashboard and Campaigns components

#### Project Infrastructure
- Initial project structure following Go standard layout (cmd/, internal/, pkg/)
- Docker Compose configuration with pgEdge PostgreSQL 17 and MCP server
- Makefile with comprehensive build, test, and development commands
- Environment configuration (.env.example) for secrets management
- GitHub Actions CI workflows for server (Go) and client (React)

#### Database
- PostgreSQL schema with UUID primary keys and JSONB columns
- Core tables: game_systems, campaigns, sessions, entities, relationships,
  timeline_events, canon_conflicts, schema_migrations
- Schema migration (001_schema.sql) with all tables, indexes,
  triggers, functions, views, and vectorization
- Seed data migration (002_seed_data.sql) for CoC 7e, GURPS 4e,
  FitD game systems and default relationship types
- GIN indexes for JSONB and array columns
- Trigram index for fuzzy name matching
- Automatic updated_at triggers
- Migration runner script (scripts/migrate.sh) with colored output

#### Game System Schemas
- Call of Cthulhu 7th Edition schema (schemas/coc-7e.yaml)
  - Percentile-based characteristics (STR, CON, SIZ, DEX, APP, INT, POW, EDU)
  - Derived attributes (HP, MP, SAN, Luck, MOV, Build, Damage Bonus)
  - Skill categories and sample skills
  - Roll mechanics with success levels
  - Sanity mechanics
- GURPS 4th Edition schema (schemas/gurps-4e.yaml)
  - Point-buy primary attributes (ST, DX, IQ, HT)
  - Secondary characteristics with formulas
  - Skill difficulty levels and cost progression
  - Advantages and disadvantages
- Forged in the Dark schema (schemas/fitd.yaml)
  - D6 dice pool mechanics
  - Action ratings grouped by attribute
  - Position and effect system
  - Stress, trauma, and harm levels
  - Faction system with tier and status

#### React Web Client
- Vite-based React 18 application with TypeScript
- Material-UI (MUI) component library with dark theme
- React Router for navigation
- React Query for server state management
- Vitest for testing with coverage
- ESLint for code quality
- Pages: Dashboard, Campaigns, Entities, Timeline, Import
- Responsive layout with sidebar navigation

#### Content Importers
- Common importer interface (internal/importers/common)
  - ExtractedEntity, ExtractedRelationship, ExtractedEvent types
  - ImportResult with entities, relationships, events, warnings, errors
  - ImportOptions for configuration
- Evernote importer (internal/importers/evernote)
  - Parses .enex XML export files
  - Extracts notes as entities with tags
  - Auto-detects entity types from content
  - Preserves creation/update timestamps
- Google Docs importer (internal/importers/googledocs)
  - Fetches public documents via export URL
  - Parses sections from headers
  - Extracts structured attributes from "Key: Value" patterns
  - Extracts relationships from text patterns
  - Extracts timeline events with dates

#### Claude Code Integration
- CLAUDE.md with comprehensive standing instructions
  - Sub-agent coordination pattern
  - Task workflow (Understand, Plan, Delegate, Verify, Review, Document, Report)
  - Automatic CHANGELOG.md and Todo.md update instructions
  - Canon management guidelines
  - Code standards and database conventions
  - Testing requirements with coverage goals
- Sub-agent definitions (.claude/agents/)
  - golang-expert: Go development and architecture
  - react-expert: React/TypeScript development
  - ttrpg-expert: Game mechanics and canon management
  - postgres-expert: Database design and optimization
  - testing-expert: Test strategies and patterns

#### Documentation
- design.md: Architecture Decision Records (ADRs)
- SCHEMAS.md: Database schema documentation
- AGENTS.md: Analysis agent registry and interfaces
- README files for project overview

#### Scripts and Automation
- scripts/migrate.sh: Database migration runner
- scripts/backup.sh: PostgreSQL backup with rotation
- scripts/restore.sh: Interactive backup restoration

### Security

- Security review identified areas for hardening before production: authentication, GM note filtering, SSL for database, environment-based credentials

### Infrastructure Details

#### Docker Services
- **imagineer-postgres**: pgEdge PostgreSQL 17 with Spock replication support
- **imagineer-mcp**: pgEdge MCP server for AI integration

#### Makefile Commands
| Command | Description |
|---------|-------------|
| `make up` | Start all Docker services |
| `make down` | Stop all services |
| `make status` | Show service and database status |
| `make psql` | Open PostgreSQL session |
| `make migrate` | Run pending migrations |
| `make test-all` | Run all tests with coverage and linting |
| `make client-dev` | Start React development server |
| `make client-build` | Build React client for production |

#### CI/CD Workflows
- **ci-server.yml**: Go tests with matrix (Go 1.23/1.24, PostgreSQL 16/17)
- **ci-client.yml**: React tests with Node.js 20/22

---

## Project Statistics

- **Database Tables**: 8 (including schema_migrations)
- **API Endpoints**: 35+ (game-systems, campaigns, entities, relationships,
  timeline, stats, import)
- **Game Systems**: 3 (Call of Cthulhu 7e, GURPS 4e, Forged in the Dark)
- **Sub-Agents**: 7 (golang, react, ttrpg, postgres, testing, mcp-server,
  design-compliance experts)
- **React Pages**: 5 (Dashboard, Campaigns, Entities, Timeline, Import)
- **React Query Hooks**: 31 (for all API domains)
- **Importers**: 2 (Evernote, Google Docs)
