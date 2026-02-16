# Imagineer Task Tracker

## MVP Definition

The Minimum Viable Product enables a GM to:

1. Log in (Google Auth) and manage their account.
2. Create and fully own their campaigns.
3. Create entities with rich text descriptions and AI assistance.
4. Plan sessions with scene structure and AI-powered entity
   detection.
5. Never lose work due to accidental navigation or browser issues.
6. Search and find content with semantic search across their
   campaign.

**Core Principle:** AI assistance is embedded in every feature from
day one. The differentiator for Imagineer is intelligent campaign
management, not just storage. Every MVP feature includes its AI
component.

**MVP Tag Legend:**

- `[MVP-1]` - Must have: Blocks basic usage if missing.
- `[MVP-2]` - Should have: Core experience features.
- `[MVP-3]` - Nice to have: Polish and convenience.

---

## In Progress

### Session Stage Workflows

The basic session CRUD (create, list, edit, delete) and stage
navigation components are complete. The remaining work focuses on
AI-powered workflows within each stage that make sessions useful
beyond simple text editing.

#### Session Prep

- [ ] `[MVP-1]` Session Prep screen layout with AI-generated
  previous session summary, rich text entry, entity sidebar, and
  AI assistant panel.
- [ ] `[MVP-1]` AI-generated previous session summary display
  (collapsible, from previous wrap-up).
- [ ] `[MVP-1]` Entity detection in prep notes (AI scans for
  existing entities and suggests new entities to create).
- [ ] `[MVP-2]` Prep breakdown into reorderable events and
  encounters for play traversal.

#### Session Play

- [ ] `[MVP-1]` Session Play screen layout with prep note
  traversal, AI chat assistant panel, and dynamic entity lists.
- [ ] `[MVP-1]` Play mode navigation to traverse prep events
  and encounters linearly.
- [ ] `[MVP-1]` AI Play Assistant chat for on-the-fly help,
  reactions, side content, and rules questions.
- [ ] `[MVP-1]` Real-time entity extraction from chat
  conversations with accept or defer options.
- [ ] `[MVP-2]` Draft canon mode for entities deferred during
  play and flagged for wrap-up review.

#### Session Wrap-up

- [ ] `[MVP-1]` Session Wrap-up screen layout with AI-generated
  summary (editable), additional notes entry, and entity
  sidebars.
- [ ] `[MVP-1]` AI session summary generation that synthesizes
  prep notes, play chat, and wrap-up notes.
- [ ] `[MVP-1]` Knowledge extraction to review entities created
  during play and approve draft canon items.
- [ ] `[MVP-1]` AI-suggested relationship creation between
  session entities with GM approval workflow.
- [ ] `[MVP-2]` Campaign story integration that adds session
  summaries to the campaign timeline.

### AI Features

These features depend on the context assembly service and RAG
infrastructure.

- [ ] `[MVP-1]` AI Memory System with campaign, chapter, and
  session memory hierarchy (persistent context, inheritance,
  management UI).
- [ ] `[MVP-1]` Context assembly service with campaign-scoped
  content retrieval and token budget management.
- [ ] `[MVP-1]` Entity AI assistance buttons: "Flesh out
  details", "Check canon", and "Suggest relationships".
- [ ] `[MVP-1]` AI summary panel component (reusable, displays
  AI-generated summaries, collapsible, context-aware).
- [ ] `[MVP-1]` Token compression for chat context (preserve
  full logs, compress for AI context window).

### Infrastructure and UI

- [ ] `[MVP-1]` API key validation on save in account settings.
- [ ] `[MVP-1]` Player Characters management UI (list view,
  create/edit, link PCs to entity records).
- [ ] `[MVP-1]` Three-panel layout for session views (basic
  campaign nav exists, but the full three-panel layout with AI
  panel does not).

### Evernote Import

- [ ] `[MVP-1]` Fix and verify Evernote ENEX file import.
- [ ] `[MVP-1]` Import preview before committing.
- [ ] `[MVP-1]` Progress indicator during import.
- [ ] `[MVP-2]` AI entity extraction from imported notes
  (post-import processing).

### Search Enhancements

- [ ] `[MVP-2]` Search filters by entity type, date range, and
  relationships.
- [ ] `[MVP-3]` Keyword and boolean search operators.

### Security

- [ ] `[MVP-3]` Import validation with duplicate detection
  (Levenshtein name similarity check, post-import consistency
  checks).

### Deferred Design

The enrichment pipeline improvements have a design document at
`docs/plans/2026-02-10-enrichment-pipeline-improvements-design.md`
and cover content-diff, budgets, parallel sub-agents, and
relationship type governance.

---

## MVP Backlog

### Scenes / Vignettes / Encounters

- [ ] `[MVP-2]` Scene data model (title, summary, involved
  entities, GM notes).
- [ ] `[MVP-2]` Scene editor UI with entity sidebar.
- [ ] `[MVP-2]` Scenes reorderable within session.
- [ ] `[MVP-3]` Memorable moments per scene (Funny, Dramatic,
  Epic, Intriguing).
- [ ] Scene outline with collapsible beats.
- [ ] Link scenes to timeline events.

### Campaign Permissions & Sharing

- [ ] `[MVP-3]` Invite other users to be a player in your
  campaign as a GM.
- [ ] `[MVP-3]` Email-based invite system with notifications.
- [ ] Invite resend functionality.
- [ ] Read-only access to campaigns you have accepted an
  invitation to as a player.
- [ ] Accept or decline invitations to other GMs' campaigns
  from your dashboard.
- [ ] Permission levels: Read, ReadWrite, ReadWriteDelete.

### Undo / Redo

- [ ] `[MVP-3]` Undo/redo stack for text content.

---

## Post-MVP Backlog

Features planned for after the initial release.

### Multi-User Enhancements

- [ ] Role-based access control (User, Admin roles).
- [ ] Admin user management.
- [ ] Additional OAuth providers (Apple, GitHub) if needed.
- [ ] Full token refresh endpoint (basic implementation
  complete).

### World Elements (New Entity Types)

- [ ] Custom entity type creation UI (allow users to define
  new types).
- [ ] Professions and careers as a separate entity type.
- [ ] Races and sexes as configurable entities.
- [ ] Languages with mastery levels (None, Semi, Proficient,
  Native).
- [ ] Equipment tracking linked to characters.
- [ ] Life date ranges (birth/death tracking) for NPCs.
- [ ] Player character attachment (link player user to
  controlled character).
- [ ] Sibling and family relationships (family trees).
- [ ] Organizations with organization types.
- [ ] Political units (kingdoms, city-states, provinces) with
  types.
- [ ] Cultures with descriptions.
- [ ] Currencies with exchange rates.
- [ ] Land features (mountains, forests) with subtypes.
- [ ] Writings (books, stories) with chapter support.
- [ ] Glossary and terms with definitions and term groups.

### Evernote Import Enhancements

- [ ] Evernote import via API call (not just file upload) to
  list spaces, notebooks, and notes and import selectively.
- [ ] AI processing of imported Evernotes to extract entities
  and sessions.
- [ ] Management UI to verify AI-extracted entities before
  ingestion with deduplication and merge approval.

### Visual & Interactive Features

#### Maps

- [ ] Map builder interface for entity locations.
- [ ] Multiple maps per campaign.
- [ ] Layer-based map system (organize drawing layers).
- [ ] Shape drawing (points, lines, polygons, rectangles,
  ellipses).
- [ ] Color-coded features with custom color system.
- [ ] Icon placement with icon categories and groups.
- [ ] GeoJSON feature support.

#### Graph Visualization

- [ ] Relationship graph visualization on entity detail page.
- [ ] Interactive graph exploration (Cytoscape).
- [ ] Graph queries filtered by element type.
- [ ] Full campaign relationship mapping visualization.

#### Calendar & Timeline

- [ ] Custom calendar support for non-Earth worlds (fantasy
  calendars).
- [ ] Configurable months, weekdays, and holidays.
- [ ] Event scheduling on custom calendars.
- [ ] Timeline visualizations with date range queries.

### Dashboards

- [ ] Custom dashboard creation with widgets.
- [ ] Dashboard widgets showing element lists.
- [ ] GM dashboard vs player dashboard variants.
- [ ] Real-time updates via WebSocket.

### RAG Enhancements

- [ ] Search result ranking tuning (A/B test vector vs BM25
  weights, boost recent content).
- [ ] Caching layer for frequent queries (cache embeddings,
  invalidate on content updates).
- [ ] Scenes vectorization (`scenes.summary` and
  `scenes.gm_notes`).

### Rulebook Knowledgebase

- [ ] Database schema for rulebook storage
  (`rulebook_sources`, `rulebook_sections` with hierarchical
  headings and vectorization).
- [ ] PDF import pipeline (text extraction, heading detection,
  page tracking, table preservation, image extraction).
- [ ] Admin UI for rulebook management (upload, review,
  edit, delete, re-import).
- [ ] Rulebook content for supported systems: Call of Cthulhu
  7e, GURPS 4e, Forged in the Dark SRD, and Blades in the
  Dark.
- [ ] Rulebook retrieval in context builder (scoped to game
  system).

### AI / Content Generation

- [ ] AI text generation for world building.
- [ ] AI image generation support.
- [ ] Prompt-based content creation.
- [ ] NPC stat block generation (using rulebook RAG).
- [ ] Character creation assistance (using rulebook RAG).
- [ ] Encounter balancing suggestions.

### File Management

- [ ] File attachments to entities.
- [ ] Cloud storage integration.
- [ ] File upload and download support.
- [ ] File metadata tracking.

### Document Import & Processing

- [ ] PDF import service (Go or Python microservice) with text
  extraction, heading detection, table extraction, page
  tracking, and multi-column layout handling.
- [ ] EPUB and MOBI import for digital rulebooks.
- [ ] Markdown import with frontmatter parsing.
- [ ] Word document (.docx) import.
- [ ] Import job queue with progress tracking.
- [ ] Import preview before committing to database.

### Canon Management

- [ ] Canon conflict detection and resolution UI.
- [ ] Source confidence tracking workflow.
- [ ] Campaign notes hierarchy (GM-only content).
- [ ] Noteable attachments (attach notes to any element).

### Testing

- [ ] Add comprehensive UI tests (React components).
- [ ] Add API integration tests (Go handlers).
- [ ] Add database integration tests (entity CRUD with real
  DB).

### Infrastructure

- [ ] MCP custom tool definitions for Imagineer operations.

---

## Completed

### Account Settings

- [x] Account settings page accessible from user menu with
  service selectors for content generation, embedding, and
  image generation.
- [x] LLM API key configuration with service selector
  (Anthropic, OpenAI, Gemini for content; Ollama, Voyage,
  OpenAI, Gemini for embedding; OpenAI, Stability AI for
  image).
- [x] Secure storage of API keys with AES-256-GCM encryption
  (migration 007, `internal/crypto` package, `ENCRYPTION_KEY`
  env var).

### Campaign Dashboard & Settings

- [x] Campaign dashboard with left navigation menu (Entities,
  Sessions, Import via CampaignNav component).
- [x] Campaign settings page with name, description
  (MarkdownEditor), RPG system selector, genre dropdown, and
  image style prompt.

### Chapters (Basic CRUD)

- [x] Chapter data model (migration 001, `models.go`).
- [x] Chapter CRUD API endpoints (handlers, router with full
  REST routes).
- [x] Chapter list view (ChapterList component).
- [x] Chapter editor UI (ChapterEditorPage with MarkdownEditor,
  entity panel, SaveSplitButton, analysis and enrichment
  integration).
- [x] Entity sidebar for chapters (ChapterEntityPanel
  component).

### Sessions (Basic CRUD)

- [x] Session data model (migration 001, `models.go` with
  stage: prep/play/wrap_up).
- [x] Session CRUD API endpoints (handlers, router with full
  REST routes including listByChapter).
- [x] Session list view (SessionList, SessionsManagement
  two-column layout).
- [x] Session creation UI (SessionEditor with title, date,
  chapter assignment, stage).
- [x] Stage navigation component (SessionStageNav,
  SessionStageIndicator).

### Entity Editor

- [x] Rich text description editor (MarkdownEditor replaces
  JSON attributes box).
- [x] Relationship association in editor (RelationshipEditor
  component with AddRelationshipDialog).

### Data Loss Prevention

- [x] Auto-draft to local storage (useAutosave hook).
- [x] Dirty state tracking with unsaved indicator
  (useUnsavedChanges hook).
- [x] Navigation confirmation (used in ChapterEditorPage,
  EntityEditor).
- [x] Draft recovery on return (ChapterEditorPage has draft
  recovery).
- [x] Save and Save & Close actions (SaveSplitButton
  component).

### UI Layout

- [x] Full-screen editing views (entity editor and chapter
  editor use FullScreenLayout).
- [x] Entity mention system (wiki links `[[Entity Name]]`
  with autocomplete, fuzzy matching, hover popover).
- [x] Entity preview popup (wiki link hover popover shows
  entity type, description snippet, and View link).

### Single-Edge Relationship Model

- [x] Migration 009: single-edge LPG model with
  `relationship_type_id` FK.
- [x] Template pattern: `relationship_type_templates` seeded
  per campaign.
- [x] DB trigger prevents inverse relationships.
- [x] `entity_relationships_view` for bidirectional display.

### Text Storage & Vectorization

- [x] Text stored in separate TEXT columns (not in JSONB) for
  pgedge_vectorizer chunking.
- [x] All narrative text columns are vectorized by
  pgedge_vectorizer.

### Multi-User Foundation & Authentication

- [x] Users table, Google OAuth integration, JWT token
  generation, auth middleware, and campaign scoping.
- [x] React AuthContext, login page, OAuth callback, protected
  routes, and user display in navigation.

### Search & RAG Infrastructure

- [x] Semantic search across all campaign elements
  (RAG-powered).
- [x] pgedge_vectorizer configured in docker-compose with
  Ollama for local embedding.
- [x] Vectorization enabled on entities, sessions, timeline
  events, and campaign descriptions.
- [x] Hybrid search API endpoint (vector similarity + BM25
  lexical, combined scoring, campaign scoping).
- [x] pgedge_vectorizer UUID primary key workaround (hotfix
  SQL patches, rebuilt chunk tables).
- [x] Integration tests for embedding pipeline
  (`search_integration_test.go`, Makefile target).

### Content Analysis & Enrichment

- [x] LLM client abstraction (Anthropic, OpenAI, Ollama).
- [x] Enrichment engine with prompt construction and JSON
  parsing.
- [x] Auto-trigger enrichment on Phase 1 triage completion.
- [x] Entity log table and CRUD API/UI.
- [x] Triage UI enrichment section (diff view, log entries,
  relationships).
- [x] Custom Ollama Docker image with pre-baked embedding
  model.
- [x] SSE streaming and polling for progressive delivery.
- [x] Relationship type autocomplete with inverse
  auto-resolve.
- [x] `RunContentEnrichment` runs independently of Phase 1
  analysis for on-demand entity mention scanning.
- [x] Enrichment progress indicator on the triage page.
- [x] Auto-advance to next pending item after accepting or
  dismissing a suggestion.
- [x] Accepted description updates and log entries apply to
  entities immediately.

### Entity & Campaign Features

- [x] Entity view page at
  `/campaigns/{id}/entities/{entityId}` with header,
  description, GM notes, attributes, relationships, event
  log, and metadata.
- [x] SaveSplitButton component with Save, Save & Analyze,
  and Save, Analyze & Enrich modes.
- [x] Duplicate relationship prevention (unique constraint,
  `ON CONFLICT DO UPDATE` upsert, inverse pair
  deduplication).
- [x] Campaigns auto-assigned to creator as GM/owner.
- [x] Only the campaign owner can modify their campaigns.
- [x] Campaign description vectorization (migration 003).
- [x] GM notes filtering (API filters `gm_notes` for
  non-owners, client hides UI).

### Wiki Links & Markdown

- [x] Wiki links (`[[Entity Name]]` and
  `[[Entity Name|display text]]`) with inline autocomplete,
  toolbar insert, clickable navigation, and entity rename
  propagation.
- [x] Markdown editor using tiptap-markdown for native
  Markdown round-trip serialization.

### Core Infrastructure

- [x] Initial project setup and core data model (SCHEMAS.md).
- [x] Database migrations squashed to 001_schema +
  002_seed_data.
- [x] Game system schemas (CoC 7e, GURPS 4e, FitD).
- [x] Docker Compose configuration (PostgreSQL + MCP Server).
- [x] Migration runner, backup, and restore scripts.
- [x] Claude sub-agent setup (CLAUDE.md, `.claude/agents/`).
- [x] GitHub Actions CI workflows.
- [x] React/Vite client scaffolding.
- [x] Makefile with test-all, lint, and coverage commands.

### Consistency Checker Agent

- [x] Agent interface with Result, Suggestion, Source structs.
- [x] Five checks: orphaned entities, duplicate names,
  timeline conflicts, invalid references, sessions without
  discoveries.
- [x] Database queries in `internal/database/consistency.go`.
- [x] API endpoint
  `POST /api/campaigns/{id}/agents/consistency-check`.

### Client Pages & Importers

- [x] Client pages (Dashboard, Campaigns, Entities, Timeline,
  Import).
- [x] Evernote importer (`internal/importers/evernote`).
- [x] Google Docs importer (`internal/importers/googledocs`).
- [x] All 15 backend API endpoints implemented.

### Bug Fixes

- [x] Fix import route mismatch (client vs backend URL
  pattern).
- [x] Fix pgx/pq array compatibility issue (entity creation).
- [x] Fix Timeline.tsx entityIds undefined crash.
- [x] Fix Timeline.tsx forEach lint error.
- [x] Add campaign scoping to relationship and timeline
  handlers.
- [x] Fix entity API URL routes (get/update/delete use
  `/entities/{id}`).
- [x] Fix Tags input UX (autoSelect and onBlur for better
  capture).
- [x] Fix `c.chunk` -> `c.content` bug in search function.
- [x] Dark theme, suggested description, and analysis badge
  fixes.

### Infrastructure Updates

- [x] Update PostgreSQL from 17 to 18
  (pgedge-postgres:18-spock5-standard).
- [x] Configure PostgreSQL extensions (vector, vectorizer,
  pg_cron).
- [x] Rename `keeper_notes` to `gm_notes` (CoC-specific to
  generic terminology).
- [x] Add RPG terminology knowledge base for TTRPG expert
  agent.
