# Component Responsibilities

This document defines clear boundaries and responsibilities for each
Imagineer component.

## Component Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         Web Client                               │
│                    (React/TypeScript)                            │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ HTTP/REST
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                        API Server                                │
│                          (Go)                                    │
└─────────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              │               │               │
              ▼               ▼               ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│    Database     │ │    Importers    │ │     Agents      │
│  (PostgreSQL)   │ │      (Go)       │ │      (Go)       │
└─────────────────┘ └─────────────────┘ └─────────────────┘
```

## API Server (`/cmd/server`)

### Core Responsibilities

- Handle HTTP requests from web client
- Authenticate and authorize requests
- Route requests to appropriate handlers
- Coordinate database operations
- Invoke importers and agents
- Return formatted responses

### What It Does NOT Do

- Store business logic in handlers (use internal packages)
- Access database directly (use database package)
- Parse import files (use importer packages)
- Make architectural decisions (follow design.md)

### Key Boundaries

- All database access through `/internal/database`
- All models defined in `/internal/models`
- Request validation at handler level
- Response formatting at handler level

## Web Client (`/client`)

### Core Responsibilities

- Present user interface for all features
- Handle user input and validation
- Communicate with API server
- Manage client-side state
- Display data in appropriate views

### What It Does NOT Do

- Store sensitive data locally
- Make direct database calls
- Implement business logic
- Validate game system schemas

### Key Boundaries

- All API calls through service layer
- State management through React Query
- Type definitions in `/client/src/types`
- Components in `/client/src/components` or `/pages`

## Internal Packages (`/internal`)

### Database Package (`/internal/database`)

**Responsibilities:**

- Connection management
- Query execution
- Transaction handling
- Migration support

**Boundaries:**

- No business logic
- No HTTP concerns
- No model definitions

### Models Package (`/internal/models`)

**Responsibilities:**

- Data structure definitions
- Validation logic
- Type conversions

**Boundaries:**

- No database queries
- No HTTP concerns
- No external dependencies

### API Package (`/internal/api`)

**Responsibilities:**

- HTTP handler implementations
- Request parsing
- Response formatting
- Calling database and service functions

**Boundaries:**

- No direct SQL
- No business logic (use services)
- No model definitions

### Agents Package (`/internal/agents`)

**Responsibilities:**

- Consistency checking
- Canon conflict detection
- Analysis operations

**Boundaries:**

- Read-only database access for analysis
- No HTTP concerns
- No direct user interaction

### Importers Package (`/internal/importers`)

**Responsibilities:**

- Parse external file formats
- Extract entities, relationships, events
- Return structured import results

**Boundaries:**

- No database access
- No HTTP concerns
- No storage operations

## Database (PostgreSQL)

### Core Responsibilities

- Store all campaign data
- Enforce referential integrity
- Support JSONB queries
- Provide fuzzy matching

### What It Does NOT Do

- Business logic (use triggers sparingly)
- Access control (application responsibility)
- File storage

### Key Boundaries

- All access through Go database package
- Migrations in `/migrations`
- No stored procedures

## Game System Schemas (`/schemas`)

### Core Responsibilities

- Define game mechanics
- Specify attribute types
- Document skill systems
- Describe dice conventions

### What They Do NOT Do

- Store campaign data
- Define entity structures
- Contain executable code

## Communication Patterns

### Client to Server

```
Client → HTTP Request → API Handler → Database/Services → Response
```

### Import Flow

```
Client → File Upload → API Handler → Importer → ImportResult → Database
```

### Agent Flow

```
Scheduler/Request → Agent → Database (read) → Analysis Result → Storage
```

## Deployment Patterns

### Development

- Server: `go run cmd/server/main.go`
- Client: `npm run dev` (Vite)
- Database: Docker Compose

### Production

- Server: Compiled binary
- Client: Static build served by server
- Database: Managed PostgreSQL

## Boundary Rules

### Rule 1: No Cross-Package Imports

Packages should not import from sibling packages at the same level.

```
internal/api → internal/database ✓
internal/api → internal/importers ✓
internal/database → internal/api ✗
```

### Rule 2: Models Are Shared

The models package can be imported by any internal package.

### Rule 3: Database Access Is Centralized

Only the database package executes SQL queries.

### Rule 4: Importers Are Stateless

Importers receive input and return results. No side effects.

### Rule 5: Agents Are Read-Heavy

Agents primarily read data for analysis. Writes are for results only.
