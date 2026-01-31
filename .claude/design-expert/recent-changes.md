# Recent Changes

This document tracks significant architectural changes and their implications.

## 2026-01-30: Knowledge Base Initialization

### Changes

- Created comprehensive knowledge bases for all sub-agents
- Added TTRPG-specific guidance across agents
- Established pattern library for common operations

### Implications

- Sub-agents now have context for Imagineer-specific decisions
- Consistent guidance across all development activities
- Easier onboarding for new contributors

## 2026-01-30: Claude Instruction Enhancements

### Changes

- Added mandatory delegation policy
- Added automatic CHANGELOG.md and Todo.md updates
- Added design-compliance-validator and mcp-server-expert agents
- Added copyright notice requirement

### Implications

- All work flows through specialized sub-agents
- Documentation stays current automatically
- Design compliance is actively validated

## Initial Setup Phase

### Migration 001: Initial Schema

Created core database tables:

- game_systems
- campaigns
- sessions
- entities
- relationships
- timeline_events
- canon_conflicts
- schema_migrations

**Key Decisions:**

- UUID primary keys throughout
- JSONB for flexible attributes
- GIN indexes for JSON queries
- Trigram extension for fuzzy matching

### Migration 002: Game System Seeds

Populated game system data from YAML schemas:

- Call of Cthulhu 7th Edition
- GURPS 4th Edition
- Forged in the Dark

**Key Decisions:**

- Full schema stored in JSONB
- Dice conventions included
- Skill categories defined

### React Client Setup

Established client architecture:

- Vite for build tooling
- Material-UI for components
- React Router for navigation
- React Query for server state

**Key Decisions:**

- Dark theme as default
- Sidebar navigation
- TypeScript throughout

### Importer Implementation

Created content import system:

- Common interface for all importers
- Evernote .enex parser
- Google Docs fetcher
- Auto entity type detection

**Key Decisions:**

- Importers return structured results
- Storage is separate from parsing
- Warnings captured for partial failures

## Lessons Learned

### 1. Schema Flexibility

JSONB works well for game system variations. Initial concern about query
performance addressed by GIN indexes.

### 2. Type Detection

Auto-detecting entity types from content is valuable but imperfect.
Keywords help but user confirmation important.

### 3. Canon Management

The DRAFT/AUTHORITATIVE/SUPERSEDED model is simple enough to be useful.
Conflict detection needs to be prominent in UI.

## Anticipated Future Changes

### Multi-User Support

When implementing:

- Add user table and authentication
- Add campaign membership
- Implement keeper vs player views
- Consider real-time updates

### API Completion

Pending:

- All CRUD endpoints
- Import endpoints
- Session management
- Conflict resolution

### Agent Development

First agent: consistency-checker

- Detect conflicting entity descriptions
- Identify relationship inconsistencies
- Flag timeline anomalies

### MCP Integration

For Claude integration:

- Define Imagineer-specific tools
- Expose entity queries
- Support relationship exploration
- Enable timeline navigation
