# Feature Locations

This document maps Imagineer features to their implementation locations.

## Entity Management

### Entity CRUD

| Feature | Location |
|---------|----------|
| Entity model | `/internal/models/entity.go` |
| Entity API handlers | `/internal/api/entities.go` |
| Entity database queries | `/internal/database/queries.go` |
| Entity UI page | `/client/src/pages/Entities.tsx` |
| Entity types (TS) | `/client/src/types/index.ts` |

### Entity Types

The system supports these entity types (defined in `/internal/importers/common/types.go`):

- `npc` - Non-player characters
- `location` - Places and settings
- `item` - Objects and artifacts
- `faction` - Groups and organizations
- `clue` - Investigation clues
- `creature` - Monsters and beings
- `organization` - Formal organizations
- `event` - Historical events
- `document` - In-game documents
- `other` - Miscellaneous

## Campaign Management

| Feature | Location |
|---------|----------|
| Campaign model | `/internal/models/campaign.go` |
| Campaign API handlers | `/internal/api/campaigns.go` |
| Campaign UI page | `/client/src/pages/Campaigns.tsx` |
| Campaign types (TS) | `/client/src/types/index.ts` |

## Relationship Mapping

| Feature | Location |
|---------|----------|
| Relationship model | `/internal/models/relationship.go` |
| Relationship API handlers | `/internal/api/relationships.go` |
| Relationship tones | `/client/src/types/index.ts` |

### Relationship Tones

Defined in TypeScript types:

- `friendly`, `hostile`, `neutral`
- `romantic`, `professional`
- `fearful`, `respectful`, `unknown`

## Timeline Events

| Feature | Location |
|---------|----------|
| Timeline event model | `/internal/models/timeline.go` |
| Timeline API handlers | `/internal/api/timeline.go` |
| Timeline UI page | `/client/src/pages/Timeline.tsx` |
| Date precision types | `/client/src/types/index.ts` |

## Session Management

| Feature | Location |
|---------|----------|
| Session model | `/internal/models/session.go` |
| Session API handlers | `/internal/api/sessions.go` |
| Session status types | `/client/src/types/index.ts` |

### Session Statuses

- `PLANNED` - Upcoming session
- `COMPLETED` - Session finished
- `SKIPPED` - Session cancelled

## Canon Conflict Detection

| Feature | Location |
|---------|----------|
| Conflict model | `/internal/models/conflict.go` |
| Conflict detection | `/internal/agents/consistency.go` |
| Conflict resolution UI | (planned) |

### Canon Confidence Levels

Defined in database and types:

- `DRAFT` - Initial entry, not confirmed
- `AUTHORITATIVE` - Confirmed as canon
- `SUPERSEDED` - Replaced by newer information

## Content Import

### Evernote Import

| Feature | Location |
|---------|----------|
| ENEX parser | `/internal/importers/evernote/parser.go` |
| Import UI tab | `/client/src/pages/Import.tsx` |

### Google Docs Import

| Feature | Location |
|---------|----------|
| Google Docs fetcher | `/internal/importers/googledocs/parser.go` |
| Import UI tab | `/client/src/pages/Import.tsx` |

### Common Importer Interface

| Feature | Location |
|---------|----------|
| Importer interface | `/internal/importers/common/types.go` |
| ExtractedEntity type | `/internal/importers/common/types.go` |
| ImportResult type | `/internal/importers/common/types.go` |

## Game System Schemas

| System | Location |
|--------|----------|
| Call of Cthulhu 7e | `/schemas/coc-7e.yaml` |
| GURPS 4e | `/schemas/gurps-4e.yaml` |
| Forged in the Dark | `/schemas/fitd.yaml` |

### Schema Contents

Each schema defines:

- Character attributes
- Derived statistics
- Skill categories
- Dice conventions
- System-specific mechanics

## Database Schema

| Feature | Location |
|---------|----------|
| Initial schema | `/migrations/001_initial_schema.sql` |
| Game system seeds | `/migrations/002_seed_game_systems.sql` |
| Migration runner | `/scripts/migrate.sh` |

### Core Tables

- `game_systems` - TTRPG system definitions
- `campaigns` - Individual campaigns
- `sessions` - Game sessions
- `entities` - NPCs, locations, items, etc.
- `relationships` - Entity connections
- `timeline_events` - Chronological events
- `canon_conflicts` - Source contradictions
