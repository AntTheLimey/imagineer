# Key Files

This document identifies the most important files in the Imagineer codebase.

## Configuration Files

### CLAUDE.md

**Location**: `/CLAUDE.md`

**Purpose**: Standing instructions for Claude Code when working on Imagineer.

**Contains**:

- Sub-agent coordination pattern
- Task workflow (7 steps)
- Canon management rules
- Code standards
- Testing requirements

**When to Reference**: Before making any changes, to understand project conventions.

### design.md

**Location**: `/design.md`

**Purpose**: Architecture and design philosophy.

**Contains**:

- Design decisions and rationale
- Component responsibilities
- Security model
- Future roadmap

**When to Reference**: When making architectural decisions or adding new features.

### Makefile

**Location**: `/Makefile`

**Purpose**: Build, test, and development commands.

**Key Commands**:

- `make up` - Start Docker services
- `make down` - Stop services
- `make migrate` - Run database migrations
- `make test-all` - Run all tests
- `make client-dev` - Start React dev server

## Database Files

### Initial Schema

**Location**: `/migrations/001_initial_schema.sql`

**Purpose**: Creates all core database tables.

**Tables Created**:

- `game_systems`
- `campaigns`
- `sessions`
- `entities`
- `relationships`
- `timeline_events`
- `canon_conflicts`
- `schema_migrations`

### Game System Seeds

**Location**: `/migrations/002_seed_game_systems.sql`

**Purpose**: Populates game system data from YAML schemas.

**Systems Seeded**:

- Call of Cthulhu 7e
- GURPS 4e
- Forged in the Dark

## Type Definitions

### Go Types (Common Importer)

**Location**: `/internal/importers/common/types.go`

**Purpose**: Shared types for all importers.

**Key Types**:

- `EntityType` - Enumeration of entity types
- `ExtractedEntity` - Entity from imported content
- `ExtractedRelationship` - Relationship from import
- `ExtractedEvent` - Timeline event from import
- `ImportResult` - Complete import result
- `Importer` - Interface for importers
- `ImportOptions` - Configuration for imports

### TypeScript Types

**Location**: `/client/src/types/index.ts`

**Purpose**: TypeScript type definitions for the client.

**Key Types**:

- `GameSystem` - Game system configuration
- `Campaign` - Campaign data
- `Entity` - Entity with all fields
- `EntityType` - Entity type union
- `Relationship` - Entity relationship
- `RelationshipTone` - Relationship tone union
- `Session` - Game session
- `SessionStatus` - Session status union
- `TimelineEvent` - Timeline event
- `DatePrecision` - Date precision union
- `CanonConflict` - Canon conflict data
- `SourceConfidence` - Confidence level union

## Game System Schemas

### Call of Cthulhu 7e

**Location**: `/schemas/coc-7e.yaml`

**Purpose**: CoC 7e game mechanics definition.

**Contents**:

- Percentile characteristics (STR, CON, etc.)
- Derived attributes (HP, MP, SAN, Luck)
- Skill categories
- Roll mechanics (success levels)
- Sanity system

### GURPS 4e

**Location**: `/schemas/gurps-4e.yaml`

**Purpose**: GURPS 4e game mechanics definition.

**Contents**:

- Point-buy attributes (ST, DX, IQ, HT)
- Secondary characteristics
- Skill difficulty levels
- Advantages/disadvantages

### Forged in the Dark

**Location**: `/schemas/fitd.yaml`

**Purpose**: FitD game mechanics definition.

**Contents**:

- D6 dice pool mechanics
- Action ratings by attribute
- Position and effect
- Stress and trauma
- Faction system

## React Components

### App Component

**Location**: `/client/src/App.tsx`

**Purpose**: Root component with routing.

**Contains**:

- React Router configuration
- Material-UI theme provider
- React Query provider
- Route definitions

### Layout Component

**Location**: `/client/src/components/Layout.tsx`

**Purpose**: Main layout with navigation.

**Contains**:

- Sidebar with navigation links
- App bar with title
- Content area

### Page Components

| Page | Location | Purpose |
|------|----------|---------|
| Dashboard | `/client/src/pages/Dashboard.tsx` | Home overview |
| Campaigns | `/client/src/pages/Campaigns.tsx` | Campaign list |
| Entities | `/client/src/pages/Entities.tsx` | Entity browser |
| Timeline | `/client/src/pages/Timeline.tsx` | Timeline view |
| Import | `/client/src/pages/Import.tsx` | Content import |

## Importer Implementations

### Evernote Parser

**Location**: `/internal/importers/evernote/parser.go`

**Purpose**: Parse Evernote .enex export files.

**Key Functions**:

- `Import()` - Main import function
- `parseENEX()` - XML parsing
- `detectEntityType()` - Auto-detection from content

### Google Docs Parser

**Location**: `/internal/importers/googledocs/parser.go`

**Purpose**: Import from Google Docs.

**Key Functions**:

- `Import()` - Main import function
- `fetchDocument()` - HTTP fetch
- `parseDocument()` - Content extraction
- `extractRelationships()` - Pattern matching

## Scripts

### Migration Runner

**Location**: `/scripts/migrate.sh`

**Purpose**: Apply database migrations.

**Features**:

- Colored output
- Tracks applied migrations
- Supports up/down/status commands

### Backup Script

**Location**: `/scripts/backup.sh`

**Purpose**: Backup PostgreSQL database.

**Features**:

- Timestamped backups
- Backup rotation
- Compression

### Restore Script

**Location**: `/scripts/restore.sh`

**Purpose**: Restore from backup.

**Features**:

- Interactive selection
- Confirmation prompt
