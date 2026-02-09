# Imagineer Task Tracker

## MVP Definition

The Minimum Viable Product enables a GM to:

1. Log in (Google Auth) and manage their account
2. Create and fully own their campaigns
3. Create entities with rich text descriptions and AI assistance
4. Plan sessions with scene structure and AI-powered entity detection
5. Never lose work due to accidental navigation or browser issues
6. Search and find content with semantic search across their campaign

**Core Principle:** AI assistance is embedded in every feature from day one.
Imagineer's differentiator is intelligent campaign management, not just
storage. Every MVP feature includes its AI component.

**MVP Tag Legend:**

- `[MVP-1]` - Must have: Blocks basic usage if missing
- `[MVP-2]` - Should have: Core experience features
- `[MVP-3]` - Nice to have: Polish and convenience

---

## In Progress

### Account Settings

- [ ] `[MVP-1]` Account settings page accessible from user menu
- [ ] `[MVP-1]` LLM API key configuration with service selector:
  - Content generation: Anthropic, OpenAI, or Gemini
  - Embedding generation: Ollama (local), Voyage, OpenAI, or
    Gemini
  - Image generation: OpenAI DALL-E, Stability AI, or other
- [ ] `[MVP-1]` Secure storage of API keys (encrypted in database)
- [ ] `[MVP-1]` API key validation on save

### Campaign Dashboard & Settings

- [ ] `[MVP-1]` Campaign dashboard with left navigation menu:
  - Manage Entities
  - Manage Sessions
  - Import Campaign Notes
  - Import Knowledge
- [ ] `[MVP-1]` Campaign settings on main screen:
  - Campaign name and description (rich text)
  - RPG system selector (from game_systems table)
  - Genre selector (dropdown with predefined genres)
  - Default image generation style prompt
- [ ] `[MVP-1]` Player Characters management:
  - List view of all player characters
  - Create/edit PC: character name, player name, description, background
  - Link PCs to entity records

### Evernote Import (Priority)

- [ ] `[MVP-1]` Fix and verify Evernote ENEX file import
- [ ] `[MVP-1]` Import preview before committing
- [ ] `[MVP-1]` Progress indicator during import
- [ ] `[MVP-2]` AI entity extraction from imported notes (post-import)

### Security (from CodeRabbit review)

- [ ] `[MVP-3]` Import validation with duplicate detection
  - Pre-create: Levenshtein name similarity check for NPCs
  - Post-create: Run consistency checks on imported batch

---

## Sessions Section (Major Feature)

### Research & Knowledge Base

- [ ] Research GM campaign content structures (web search)
  - Common campaign organization patterns
  - Session prep best practices
  - Session wrap-up structures
  - "Prep" tasks and patterns used by experienced GMs
- [ ] Update TTRPG Expert knowledge base with research findings

### Text Entry & Entity Linking System

- [x] `[MVP-1]` Wiki-link entity linking with autocomplete
  - Syntax: `[[Entity Name]]` or `[[Entity Name|display text]]`
    with typeahead autocomplete (triggered by typing `[[`)
  - Fuzzy entity name matching via `pg_trgm`
  - Toolbar insert button with entity search popover
  - Visual indicator for linked entities in text
  - Click wiki link to navigate to entity search
  - Entity rename propagation updates all wiki links across
    campaign content in a single transaction
- [ ] `[MVP-1]` Text storage architecture for vectorization
  - Separate text content from JSONB for pgedge vectorizer chunking
  - Ensure all text entries can be chunked and embedded
  - Entity link metadata stored alongside text
- [x] `[MVP-1]` Convert rich text editor output to Markdown
      before saving
  - The MarkdownEditor uses tiptap-markdown for native
    Markdown round-trip serialization.
  - Markdown storage produces better embeddings and is more
    portable than HTML.

### AI Memory System

- [ ] `[MVP-1]` Campaign Memory - persistent context across all sessions
  - Stores key campaign facts, themes, active plot threads
  - GM can view, edit, and delete memory entries
- [ ] `[MVP-1]` Chapter Memory - context scoped to chapter
  - Chapter-specific plot points, NPCs, locations
  - Inherits from Campaign Memory
- [ ] `[MVP-1]` Session Memory - context for individual session
  - Session-specific events, decisions, outcomes
  - Inherits from Chapter and Campaign Memory
- [ ] `[MVP-1]` Memory management UI
  - View memories at each level (Campaign/Chapter/Session)
  - Edit and delete memory entries (full GM control)
  - Memory hierarchy visualization
- [ ] `[MVP-1]` Token compression for chat context
  - Preserve full chat logs for review
  - Compress for AI context window
  - Reference Anthropic memory system guidance
  - Reference pgedge-postgres-mcp chat agent patterns

### Chapters

- [ ] `[MVP-1]` Chapter data model
  - Title, overview (rich text), campaign_id
  - Sort order within campaign
  - Created/updated timestamps
- [ ] `[MVP-1]` Chapter CRUD API endpoints
- [ ] `[MVP-1]` Chapter list view in campaign dashboard
- [ ] `[MVP-1]` Chapter editor UI
  - Title field
  - Rich text overview editor
  - AI scans overview for entity mentions
  - AI suggests new entities to create
  - AI creates/suggests relationship links
- [ ] `[MVP-1]` Entity sidebar for chapters
  - Shows entities mentioned in chapter overview
  - Click entity shows preview popup
  - Option to insert entity link at cursor

### Sessions

- [ ] `[MVP-1]` Session data model
  - Title, date, chapter_id (required)
  - Stage: prep | play | wrap-up
  - Sort order (chronological within chapter)
  - Created/updated timestamps
- [ ] `[MVP-1]` Session CRUD API endpoints
- [ ] `[MVP-1]` Session list view under each chapter
  - Chronological order (creation order)
  - Visual indicator of current stage
- [ ] `[MVP-1]` Session creation UI
  - Title and date fields
  - Auto-assigns to current chapter
  - Starts in "prep" stage

### Session Stages

#### Session Prep

- [ ] `[MVP-1]` Session Prep screen layout
  - Top half: AI-generated summary display (previous session context)
  - Bottom half: Rich text entry for GM notes
  - Left sidebar: Related entity lists by type
  - Right sidebar: AI assistant panel
- [ ] `[MVP-1]` Previous session summary display
  - AI-generated summary from previous session wrap-up
  - Collapsible for more screen space
- [ ] `[MVP-1]` Prep notes editor
  - Rich text with @mention entity linking
  - GM prepares 2-4 hours of content
  - Structure into events/encounters (optional)
- [ ] `[MVP-1]` Entity detection in prep notes
  - AI scans for existing entities
  - AI suggests new entities to create
  - Running list of detected/created entities
- [ ] `[MVP-2]` Prep breakdown into events/encounters
  - Optional linear structure for play traversal
  - Reorderable event list

#### Session Play

- [ ] `[MVP-1]` Session Play screen layout
  - Main content: Prep notes with event/encounter traversal
  - Right panel: Always-visible AI chat assistant
  - Dynamic entity lists: New entities discovered during play
- [ ] `[MVP-1]` Play mode navigation
  - Traverse prep events/encounters linearly
  - Free navigation back to prep notes
- [ ] `[MVP-1]` AI Play Assistant chat
  - On-the-fly help during session
  - Create reactions and side content
  - Answer rules questions
  - Full chat log preserved
- [ ] `[MVP-1]` Real-time entity extraction
  - AI creates entities from chat conversations
  - Running list of new entities with relationships
  - GM can accept immediately (add to campaign knowledge)
  - GM can defer to draft mode (review in wrap-up)
- [ ] `[MVP-2]` Draft canon mode for deferred entities
  - Entities created but not yet canonical
  - Flagged for review in wrap-up stage

#### Session Wrap-up

- [ ] `[MVP-1]` Session Wrap-up screen layout
  - Top half: AI-generated session summary (editable)
  - Bottom half: Additional notes entry
  - Sidebars: Entities created/modified during session
- [ ] `[MVP-1]` AI session summary generation
  - Synthesize from prep notes, play chat, and wrap-up notes
  - Editable by GM before finalizing
  - Contributes to campaign story arc
- [ ] `[MVP-1]` Knowledge extraction
  - Review entities created during play
  - Review draft canon items for approval/edit
  - Extract additional entities from wrap-up notes
- [ ] `[MVP-1]` Relationship creation
  - AI suggests relationships between session entities
  - GM approves/rejects/edits relationships
- [ ] `[MVP-2]` Campaign story integration
  - Add session summary to campaign timeline
  - Update chapter context with session outcomes

### Session UI Components

- [ ] `[MVP-1]` Entity sidebar component (reusable)
  - Grouped by entity type (NPCs, Locations, Items, etc.)
  - Only shows types with relevant entities
  - Scoped to campaign/chapter/session context
  - Click shows entity preview popup
  - Click can insert @mention at cursor
- [ ] `[MVP-1]` Entity preview popup
  - Quick view of entity details
  - Options: Insert link, Navigate to entity, Close
- [ ] `[MVP-1]` AI summary panel component (reusable)
  - Displays AI-generated summaries
  - Collapsible/expandable
  - Shows relevant context for current view
- [ ] `[MVP-2]` Stage navigation component
  - Visual indicator of current stage
  - Free navigation between Prep/Play/Wrap-up
  - Confirmation when leaving with unsaved changes

## MVP Backlog

Priority features for initial release, ordered by dependency.

### Multi-User Foundation & Authentication (Post-MVP Backlog)

#### Future (Post-MVP)

- [ ] Role-based access control (User, Admin roles)
- [ ] Admin user management
- [ ] Additional OAuth providers (Apple, GitHub) if needed
- [ ] Full token refresh endpoint (basic implementation complete)

### Campaign Permissions & Sharing

- [x] `[MVP-1]` Campaigns auto-assigned to creator as GM/owner
- [x] `[MVP-1]` Only owner can modify their campaigns
- [ ] `[MVP-3]` As a GM, invite other users to be a player in your campaign
- [ ] `[MVP-3]` Email-based invite system with notifications
- [ ] Invite resend functionality
- [ ] Only see (read-only) campaigns you have accepted an invitation to as a
      player
- [ ] As a user, see invites to other GMs' campaigns in your dashboard, and
      accept or decline them
- [ ] Permission levels: Read, ReadWrite, ReadWriteDelete

### UI Layout & Architecture (Foundation)

The layout is the UX framework into which all features fit.

- [ ] `[MVP-1]` Three-panel layout for planning/session views
  - Left: Campaign navigator, session list (chronological)
  - Center: Content authoring canvas with rich text
  - Right: Contextual entities panel (auto-populated by AI)
- [ ] `[MVP-1]` Full-screen editing views (not popups) for:
  - Entity editor with tabbed sections
  - Session planning view
  - Scene/vignette editor
  - Timeline event editor
  - (Map canvas deferred to post-MVP)
- [ ] `[MVP-2]` Entity mention system (@mentions in rich text)
  - Autocomplete for existing entities
  - Visual indicator for mentioned entities
  - Click mention to view/edit entity

### Data Loss Prevention

- [ ] `[MVP-1]` Auto-draft to local storage every 30 seconds
- [ ] `[MVP-1]` Dirty state tracking with unsaved indicator
- [ ] `[MVP-1]` Navigation confirmation when unsaved changes exist
- [ ] `[MVP-2]` Draft recovery on return (offer to restore)
- [ ] `[MVP-2]` Explicit Save and Save & Close actions
- [ ] `[MVP-3]` Undo/redo stack for text content

### Scenes / Vignettes / Encounters

- [ ] `[MVP-2]` Scene data model (title, summary, involved entities, GM notes)
- [ ] `[MVP-2]` Scene editor UI with entity sidebar
- [ ] `[MVP-2]` Scenes reorderable within session
- [ ] `[MVP-3]` Memorable moments per scene (Funny, Dramatic, Epic, Intriguing)
- [ ] Scene outline with collapsible beats
- [ ] Link scenes to timeline events

### Session Planning & Management

- [ ] `[MVP-1]` Session planning mode for preparing your next session
- [ ] `[MVP-1]` Keep each session chronologically for your campaign
- [ ] `[MVP-1]` Three-panel planning view:
  - Central canvas for free-form notes
  - Right panel shows entities (auto-populated by AI)
  - AI action buttons: Scan, Check Canon, Suggest, Find Holes
- [ ] `[MVP-1]` AI-assisted session planning:
  - Free-form typing and pasting notes
  - AI automatically finds and links existing entities
  - AI suggests new entities to create from your notes
  - Trigger AI scan at any point for suggestions
  - Scan for plot holes and canon inconsistencies (uses consistency-checker)
- [ ] `[MVP-2]` Session notes with entity linking
- [ ] `[MVP-2]` Session summary generation (Full/Short/Stylized tabs)
- [ ] `[MVP-3]` Memorable moments capture during or after session

### Entity Editor Improvements

- [ ] `[MVP-1]` Replace JSON attributes box with rich text description editor
  - Headers, bolding, bullet points, etc.
  - More user-friendly than raw JSON
- [ ] `[MVP-1]` Add relationship association in entity create/edit form
  - Select existing entities to link as relationships
- [ ] `[MVP-1]` AI assistance buttons for entity editing:
  - [ ] "Flesh out details" - expand description based on entity type and
        relationships
  - [ ] "Check canon" - detect violations against established campaign facts
  - [ ] "Suggest relationships" - recommend likely entity connections

### Search & Discovery

- [x] `[MVP-1]` Semantic search across all campaign elements (RAG-powered)
- [ ] `[MVP-2]` Filter by element type, date range, relationships
- [ ] `[MVP-3]` Keyword and boolean search operators
- [ ] Graph-based relationship search

### RAG Infrastructure (Required for AI Features)

#### pgedge_vectorizer Setup

- [x] `[MVP-1]` Configure pgedge_vectorizer in docker-compose.yml
  - Add to shared_preload_libraries (already done)
  - Set pgedge_vectorizer.databases = 'imagineer'
  - Configure provider (Ollama for local, OpenAI as alternative)
  - Set num_workers and batch_size
- [x] `[MVP-1]` Create API key management
  - Secure storage for embedding API keys
  - Environment variable injection into container
- [x] `[MVP-1]` Enable vectorization on campaign content tables:
  - [x] entities.description (hybrid strategy)
  - [x] entities.gm_notes (hybrid strategy)
  - [x] sessions.notes (hybrid strategy)
  - [x] timeline_events.description

#### Hybrid Search

- [x] `[MVP-1]` Hybrid search API endpoint
  - Vector similarity via chunk tables
  - BM25 lexical via vchord_bm25
  - Combined scoring (configurable weights)
  - Filter by campaign, entity type, date range

#### Retrieval Context Builder

- [ ] `[MVP-1]` Context assembly service
  - Campaign content retrieval (scoped to campaign)
  - Token budget management (fit within LLM context)
  - Source attribution for retrieved chunks

---

## Post-MVP Backlog

Features planned for after initial release.

### Entity History

- [x] Entity event log tracking mentions in sessions
  - Build viewable history over time
  - Automatically establish relationship links to other entities mentioned
    in the same content/event

### World Elements (New Entity Types)

- [ ] Custom entity type creation UI (allow users to define new types)
- [ ] Professions/careers as separate entity type
- [ ] Races and sexes as configurable entities
- [ ] Languages with mastery levels (None, Semi, Proficient, Native)
- [ ] Equipment tracking linked to characters
- [ ] Life date ranges (birth/death tracking) for NPCs
- [ ] Player character attachment (link player user to controlled character)
- [ ] Sibling/family relationships (family trees)
- [ ] Organizations with organization types
- [ ] Political units (kingdoms, city-states, provinces) with types
- [ ] Cultures with descriptions
- [ ] Currencies with exchange rates
- [ ] Land features (mountains, forests) with subtypes
- [ ] Writings (books, stories) with chapter support
- [ ] Glossary/terms with definitions and term groups

### Evernote Import Enhancements

- [ ] Evernote import via API call (not just file upload)
  - List your spaces, notebooks, and notes
  - Import entire space, entire notebook, or individual notes
- [ ] AI processing of imported Evernotes to extract entities and sessions
- [ ] Management UI to verify AI-extracted entities before ingestion
  - AI to dedupe and merge in new details
  - Merged version requires approval before updating

### Visual & Interactive Features

#### Maps

- [ ] Map builder interface for entity locations
- [ ] Multiple maps per campaign
- [ ] Layer-based map system (organize drawing layers)
- [ ] Shape drawing (points, lines, polygons, rectangles, ellipses)
- [ ] Color-coded features with custom color system
- [ ] Icon placement with icon categories and groups
- [ ] GeoJSON feature support

#### Graph Visualization

- [ ] Relationship graph visualization on entity detail page
- [ ] Interactive graph exploration (Cytoscape)
- [ ] Graph queries filtered by element type
- [ ] Full campaign relationship mapping visualization

#### Calendar & Timeline

- [ ] Custom calendar support for non-Earth worlds (fantasy calendars)
- [ ] Configurable months, weekdays, holidays
- [ ] Event scheduling on custom calendars
- [ ] Timeline visualizations with date range queries

### Dashboards

- [ ] Custom dashboard creation with widgets
- [ ] Dashboard widgets showing element lists
- [ ] GM dashboard vs player dashboard variants
- [ ] Real-time updates via WebSocket

### RAG Enhancements (Post-MVP)

- [ ] Search result ranking tuning
  - A/B test vector vs BM25 weights
  - Boost recent content option
- [ ] Caching layer for frequent queries
  - Cache embeddings for common search terms
  - Invalidate on content updates
- [ ] Scenes vectorization (scenes.summary and scenes.gm_notes)

### Rulebook Knowledgebase

- [ ] Database schema for rulebook storage
  - rulebook_sources table (title, game_system, type)
  - rulebook_sections table (hierarchical, with heading_path)
  - Enable vectorization with hybrid chunking
- [ ] PDF import pipeline
  - [ ] PDF text extraction (pdfplumber or PyMuPDF)
  - [ ] Heading detection and hierarchy extraction
  - [ ] Page number tracking
  - [ ] Table and list preservation
  - [ ] Image extraction (for later use)
- [ ] Admin UI for rulebook management
  - Upload PDF files
  - Review extracted sections
  - Edit/correct extraction errors
  - Delete/re-import sources
- [ ] Rulebook content for supported systems:
  - [ ] Call of Cthulhu 7e Keeper's Guide
  - [ ] Call of Cthulhu 7e Investigator's Handbook
  - [ ] GURPS 4e Basic Set (Characters + Campaigns)
  - [ ] Forged in the Dark SRD
  - [ ] Blades in the Dark core rules
- [ ] Rulebook retrieval in context builder (scoped to game system)

### AI/Content Generation

- [ ] AI text generation for world building
- [ ] AI image generation support
- [ ] Prompt-based content creation
- [ ] NPC stat block generation (using rulebook RAG)
- [ ] Character creation assistance (using rulebook RAG)
- [ ] Encounter balancing suggestions

### File Management

- [ ] File attachments to entities
- [ ] Cloud storage integration
- [ ] File upload/download support
- [ ] File metadata tracking

### Document Import & Processing

- [ ] PDF import service (Go or Python microservice)
  - Text extraction with layout preservation
  - Heading hierarchy detection (font size, bold, numbering)
  - Table extraction to structured format
  - Page number tracking for citations
  - Handle multi-column layouts
- [ ] EPUB/MOBI import for digital rulebooks
- [ ] Markdown import with frontmatter parsing
- [ ] Word document (.docx) import
- [ ] Import job queue with progress tracking
- [ ] Import preview before committing to database

### Canon Management

- [ ] Canon conflict detection and resolution UI
- [ ] Source confidence tracking workflow
- [ ] Campaign notes hierarchy (GM-only content)
- [ ] Noteable attachments (attach notes to any element)

### Testing

- [ ] Add comprehensive UI tests (React components)
- [ ] Add API integration tests (Go handlers)
- [ ] Add database integration tests (entity CRUD with real DB)

### Infrastructure

- [ ] MCP custom tool definitions for Imagineer operations

---

## Completed

- [x] Multi-User Foundation & Authentication
  - [x] Create users table (id, google_id, email, name, avatar_url,
    created_at, updated_at)
  - [x] Add owner_id to campaigns table (FK to users)
  - [x] User support added to schema migration
  - [x] Google OAuth integration (sign in / sign up)
  - [x] JWT token generation after successful OAuth
  - [x] Token refresh mechanism (basic implementation; full refresh endpoint
    deferred)
  - [x] Auth middleware extracts user from JWT and attaches to request
    context
  - [x] Campaign queries scoped to owner (users only see their own campaigns)
  - [x] All child data (entities, sessions, etc.) inherits campaign scope
  - [x] React AuthContext for client-side auth state
  - [x] Login page with Google sign-in
  - [x] OAuth callback handling
  - [x] Protected routes in React app
  - [x] User display in navigation (avatar, name, logout)
- [x] First agent: consistency-checker
  - Agent interface with Result, Suggestion, Source structs
  - Five checks: orphaned entities, duplicate names, timeline conflicts,
    invalid references, sessions without discoveries
  - Database queries in internal/database/consistency.go
  - API endpoint POST /api/campaigns/{id}/agents/consistency-check
  - Comprehensive unit tests
- [x] Initial project setup
- [x] Core data model implementation (SCHEMAS.md)
- [x] Database migrations squashed to 001_schema + 002_seed_data
- [x] Game system schemas (CoC 7e, GURPS 4e, FitD)
- [x] Docker Compose configuration (PostgreSQL + MCP Server)
- [x] Migration runner script
- [x] Backup/restore scripts
- [x] Claude sub-agent setup (CLAUDE.md, .claude/agents/)
- [x] GitHub Actions CI workflows
- [x] React/Vite client scaffolding
- [x] Client pages (Dashboard, Campaigns, Entities, Timeline, Import)
- [x] Evernote importer (internal/importers/evernote)
- [x] Google Docs importer (internal/importers/googledocs)
- [x] Makefile with test-all, lint, coverage commands
- [x] Implement API endpoints (15 missing backend endpoints)
- [x] Fix import route mismatch (client vs backend URL pattern)
- [x] Implement Entities.tsx page with full CRUD operations
- [x] Implement Timeline.tsx page with vertical timeline view
- [x] Connect Import.tsx handlers to API hooks
- [x] Fix port configuration (server on 3001, avoid MCP conflict)
- [x] Fix pgx/pq array compatibility issue (entity creation was failing)
- [x] Fix Timeline.tsx entityIds undefined crash (CodeRabbit)
- [x] Fix Timeline.tsx forEach lint error (CodeRabbit)
- [x] Add campaign scoping to relationship/timeline handlers (CodeRabbit)
- [x] Update PostgreSQL from 17 to 18 (pgedge-postgres:18-spock5-standard)
- [x] Configure PostgreSQL extensions (vector, vectorizer, pg_cron, etc.)
- [x] GM notes filtering (API filters gm_notes for non-owners, client hides UI)
- [x] Rename keeper_notes to gm_notes (CoC-specific to generic terminology)
- [x] Add RPG terminology knowledge base for TTRPG expert agent
- [x] Fix entity API URL routes (get/update/delete use /entities/{id})
- [x] Fix Tags input UX (autoSelect and onBlur for better capture)
- [x] Campaigns auto-assigned to creator as GM/owner (was already implemented)
- [x] Only owner can modify their campaigns (was already implemented)
- [x] Campaign description vectorization (migration 003)
  - campaigns.description vectorized with chunk_size=200
  - search_campaign_content() updated with campaigns branch
  - Fixed c.chunk -> c.content bug in search function
- [x] Integration tests for embedding pipeline
  - search_integration_test.go with build tag `integration`
  - Tests: vectorization available, chunks created, search
    results, campaign description search
  - Makefile test-integration target added
- [x] Wiki Links (`[[Entity Name]]` syntax)
  - `[[Entity Name]]` and `[[Entity Name|display text]]`
    wiki-link syntax in the Markdown editor and renderer.
  - Inline autocomplete triggered by `[[` with `pg_trgm` fuzzy
    matching.
  - Toolbar insert button with entity search popover.
  - Clickable wiki links in MarkdownRenderer with entity
    navigation.
  - Entity rename propagation across all campaign content.
  - Entity resolve API endpoint for fuzzy name matching.
- [x] Convert rich text editor output to Markdown before saving
  - The MarkdownEditor uses tiptap-markdown for native Markdown
    round-trip serialization.
- [x] pgedge_vectorizer UUID primary key workaround
  - Hotfix SQL patches enable_vectorization and
    vectorization_trigger for UUID support
  - All chunk tables rebuilt with UUID source_id columns
- [x] LLM Enrichment (Phase 4)
  - LLM client abstraction (Anthropic, OpenAI, Ollama)
  - Enrichment engine with prompt construction and JSON parsing
  - Auto-trigger enrichment on Phase 1 triage completion
  - Entity log table and CRUD API/UI
  - Triage UI enrichment section (diff view, log entries,
    relationships)
  - Custom Ollama Docker image with pre-baked embedding model
  - SSE streaming and polling for progressive delivery
