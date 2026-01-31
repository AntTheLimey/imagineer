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

### Security (from CodeRabbit review)

- [ ] `[MVP-3]` Import validation with duplicate detection
  - Pre-create: Levenshtein name similarity check for NPCs
  - Post-create: Run consistency checks on imported batch

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

- [ ] `[MVP-1]` Semantic search across all campaign elements (RAG-powered)
- [ ] `[MVP-2]` Filter by element type, date range, relationships
- [ ] `[MVP-3]` Keyword and boolean search operators
- [ ] Graph-based relationship search

### RAG Infrastructure (Required for AI Features)

#### pgedge_vectorizer Setup

- [ ] `[MVP-1]` Configure pgedge_vectorizer in docker-compose.yml
  - Add to shared_preload_libraries (already done)
  - Set pgedge_vectorizer.databases = 'imagineer'
  - Configure provider (OpenAI initially)
  - Set num_workers and batch_size
- [ ] `[MVP-1]` Create API key management
  - Secure storage for embedding API keys
  - Environment variable injection into container
- [ ] `[MVP-1]` Enable vectorization on campaign content tables:
  - [ ] entities.description (hybrid strategy)
  - [ ] entities.gm_notes (hybrid strategy)
  - [ ] sessions.notes (hybrid strategy)
  - [ ] timeline_events.description

#### Hybrid Search

- [ ] `[MVP-1]` Hybrid search API endpoint
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

- [ ] Entity event log tracking mentions in sessions
  - Build viewable history over time
  - Automatically establish relationship links to other entities mentioned
    in the same content/event

### World Elements (New Entity Types)

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
  - [x] Migration `003_add_users.sql` to add user support
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
- [x] Database migrations (001_initial_schema, 002_seed_game_systems)
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
