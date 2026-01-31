# Imagineer Task Tracker

## In Progress

- [ ] First agent: consistency-checker

### Security (from CodeRabbit review)

- [ ] Keeper notes GM filtering - requires auth system first
  - API: Filter keeper_notes field from responses for non-GM users
  - Client: Wrap keeper notes UI in GM authorization check
- [ ] Import validation with duplicate detection
  - Pre-create: Levenshtein name similarity check for NPCs
  - Post-create: Run consistency checks on imported batch

## Backlog

### Entity Editor Improvements

- [ ] Replace JSON attributes box with rich text description editor
  - Headers, bolding, bullet points, etc.
  - More user-friendly than raw JSON
- [ ] Add relationship association in entity create/edit form
  - Select existing entities to link as relationships
- [ ] AI assistance buttons for entity editing:
  - [ ] "Flesh out details" - expand description based on entity type and
        relationships
  - [ ] "Check canon" - detect violations against established campaign facts
  - [ ] "Suggest relationships" - recommend likely entity connections

### Testing

- [ ] Add comprehensive UI tests (React components)
- [ ] Add API integration tests (Go handlers)
- [ ] Add database integration tests (entity CRUD with real DB)

### Features

- [ ] Custom calendar support for non-Earth worlds (fantasy calendars)
- [ ] Graph relationship visualization on entity detail page
- [ ] Map builder interface for entity locations (PostGIS + Leaflet)
- [ ] MCP custom tool definitions for Imagineer operations
- [ ] Session management features
- [ ] Relationship mapping visualization
- [ ] Canon conflict detection and resolution UI
- [ ] User authentication system

## Completed

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
