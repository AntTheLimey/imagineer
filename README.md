# Imagineer

**TTRPG Campaign Intelligence Platform**

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Node Version](https://img.shields.io/badge/Node-20+-339933?style=flat&logo=node.js)](https://nodejs.org/)
[![License](https://img.shields.io/badge/License-PostgreSQL-336791?style=flat)](LICENSE.md)
[![CI](https://github.com/AntTheLimey/imagineer/actions/workflows/ci.yml/badge.svg)](https://github.com/AntTheLimey/imagineer/actions)

---

## Overview

Imagineer is a system-agnostic campaign management platform for tabletop role-playing games (TTRPGs). It helps Game Masters organize complex campaigns by providing intelligent tools for tracking entities, relationships, timelines, and narrative continuity.

### Key Capabilities

- **Multi-System Support** - Built-in schemas for Call of Cthulhu 7e, GURPS 4e, and Forged in the Dark
- **Entity Management** - Track NPCs, locations, items, factions, clues, and creatures with system-specific attributes
- **Relationship Mapping** - Model connections between entities with tone, strength, and directionality
- **Timeline Management** - Maintain in-game chronology with flexible date precision
- **Canon Management** - Detect and resolve conflicting information with DRAFT/AUTHORITATIVE/SUPERSEDED states
- **Import Tools** - Bring in existing content from Evernote and Google Docs

---

## Features

### Campaign Intelligence
- **Entity Tracking** - Comprehensive management of NPCs, locations, items, factions, clues, and creatures
- **Relationship Graphs** - Map connections between entities with relationship types, tones, and strength values
- **Timeline Events** - Track in-game chronology with exact, approximate, month, or year precision
- **Session Management** - Plan sessions, record outcomes, and track player discoveries

### Canon Management
- **Conflict Detection** - Automatically identify contradictory information across sources
- **Source Tracking** - Maintain provenance for all campaign data
- **Version Control** - DRAFT, AUTHORITATIVE, and SUPERSEDED states for content lifecycle
- **Resolution Workflow** - Surface conflicts for human decision-making

### Import Capabilities
- **Evernote Integration** - Import notes and structure from Evernote exports
- **Google Docs Support** - Pull campaign content from Google Docs

### Multi-System Architecture
- **Flexible Schemas** - JSONB-based attributes adapt to any game system
- **System Templates** - Pre-built schemas for popular TTRPGs
- **Custom Systems** - Define your own attribute and skill schemas

---

## Quick Start

```bash
# Clone the repository
git clone https://github.com/AntTheLimey/imagineer.git
cd imagineer

# Copy configuration examples
cp .env.example .env
cp config/db/db.json.example config/db/db.json
cp .mcp.json.example .mcp.json

# Start services (PostgreSQL)
make up

# Run database migrations
make migrate

# Start the backend server (port 8081)
PORT=8081 go run ./cmd/server

# In a separate terminal, start the frontend (port 5173)
cd client && npm install && npm run dev
```

Access the application at `http://localhost:5173`

---

## Architecture

Imagineer follows a modern full-stack architecture with clear separation of concerns:

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  React Client   │────▶│   Go API Server │────▶│   PostgreSQL    │
│  (TypeScript)   │◀────│   (chi router)  │◀────│   (JSONB)       │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

### Technology Stack

| Layer | Technology |
|-------|------------|
| **Frontend** | React 18, TypeScript, Material-UI, TanStack Query |
| **Backend** | Go 1.24+, chi router, pgx |
| **Database** | PostgreSQL 16+ with JSONB for flexible schemas |
| **Testing** | Go testing, Vitest, React Testing Library |
| **Build** | Vite, Make |

---

## Game Systems

Imagineer ships with support for three distinct game systems:

| System | Code | Dice | Description |
|--------|------|------|-------------|
| **Call of Cthulhu 7e** | `coc-7e` | d100 | Lovecraftian horror investigation |
| **GURPS 4e** | `gurps-4e` | 3d6 | Generic universal role-playing system |
| **Forged in the Dark** | `fitd` | d6 pool | Heist and scoundrel games (Blades in the Dark, Scum & Villainy) |

System definitions are stored in `schemas/` as YAML files. You can add custom systems by creating new schema files.

---

## Development

### Prerequisites

- **Go** 1.24 or later
- **Node.js** 20 or later
- **PostgreSQL** 16 or later
- **Docker** and Docker Compose (for containerized development)

### Common Commands

```bash
# Start all services
make up

# Stop all services
make down

# View service status
make status

# Run all tests with linting and coverage
make test-all

# Run Go tests only
make test-server

# Run React tests only
make test-client

# Run linters
make lint

# Build binaries
make build

# Open psql session
make psql

# View migration status
make migrate-status
```

### Environment Configuration

Copy the example files and configure as needed:

- `.env` - Database credentials and API keys
- `config/db/db.json` - Database connection settings
- `.mcp.json` - MCP server configuration for AI integration

---

## Project Structure

```
imagineer/
├── cmd/
│   ├── server/          # API server entry point
│   └── cli/             # Command-line tools
├── client/              # React frontend application
│   ├── src/             # TypeScript source code
│   └── dist/            # Production build output
├── internal/            # Private Go packages
│   ├── api/             # HTTP handlers and routing
│   ├── database/        # Database access layer
│   ├── models/          # Domain models
│   ├── importers/       # Evernote/Google Docs importers
│   └── agents/          # AI analysis agents
├── migrations/          # SQL migration files
├── schemas/             # Game system YAML definitions
├── config/              # Configuration files
├── scripts/             # Build and deployment scripts
└── docker-compose.yml   # Container orchestration
```

---

## Data Model

Imagineer uses a flexible data model built on PostgreSQL with JSONB columns:

- **Campaigns** - Top-level container for all campaign data
- **Entities** - Polymorphic table for NPCs, locations, items, factions, clues, creatures
- **Relationships** - Typed connections between entities with tone and strength
- **Sessions** - Session planning and outcome tracking
- **Timeline Events** - In-game chronology with flexible date precision
- **Canon Conflicts** - Detected contradictions awaiting resolution

See [SCHEMAS.md](SCHEMAS.md) for detailed schema documentation.

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes with clear messages
4. **Push** to your branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

### Development Guidelines

- Follow Go standard project layout conventions
- Search for duplicates before creating new entities
- Run `make test-all` before submitting PRs
- Never modify migrations that have been applied
- Mark AI suggestions clearly as SUGGESTION

---

## License

This project is licensed under the PostgreSQL License - see the [LICENSE.md](LICENSE.md) file for details.

---

## Acknowledgments

- Built with [Go](https://go.dev/), [React](https://react.dev/), and [PostgreSQL](https://www.postgresql.org/)
- UI components from [Material-UI](https://mui.com/)
- Inspired by the needs of GMs everywhere who struggle to keep their campaigns organized

---

*Imagineer - Because your campaign deserves intelligence.*
