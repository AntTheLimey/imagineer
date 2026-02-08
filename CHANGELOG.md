# Changelog

All notable changes to the Imagineer project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

- Wiki Links (`[[Entity Name]]` syntax)
  - `[[Entity Name]]` and `[[Entity Name|display text]]` wiki-link
    syntax in the Markdown editor and renderer.
  - Inline autocomplete triggered by `[[` with fuzzy entity name
    matching via `pg_trgm`.
  - Toolbar insert button for wiki-link insertion via entity
    search.
  - Clickable wiki links in read mode navigate to entity search.
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
