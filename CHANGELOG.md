# Changelog

All notable changes to the Imagineer project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

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
  - Tags, keeper notes, and source confidence tracking
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

### Changed

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
- Initial migration (001_initial_schema.sql) with all core tables
- Game system seed data (002_seed_game_systems.sql) for CoC 7e, GURPS 4e, FitD
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

- Security review identified areas for hardening before production: authentication, keeper note filtering, SSL for database, environment-based credentials

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
