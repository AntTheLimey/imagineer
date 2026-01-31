# Project Structure

This document describes the directory organization of Imagineer.

## Top-Level Layout

```
imagineer/
├── .claude/              # Claude Code agent definitions and knowledge bases
├── .github/              # GitHub Actions workflows
├── client/               # React web application
├── cmd/                  # Go application entry points
│   ├── server/           # API server
│   └── cli/              # Command-line interface
├── config/               # Configuration files
├── docs/                 # Project documentation
├── internal/             # Internal Go packages
├── migrations/           # Database migrations
├── schemas/              # Game system YAML definitions
├── scripts/              # Automation scripts
├── CLAUDE.md             # Claude Code instructions
├── design.md             # Architecture and design document
├── Makefile              # Build commands
└── README.md             # Project overview
```

## Command Applications (`/cmd`)

### Server (`/cmd/server`)

The API server for the React web client.

```
cmd/server/
├── main.go               # Entry point
└── README.md             # Server documentation
```

### CLI (`/cmd/cli`)

Command-line interface for Imagineer operations.

```
cmd/cli/
├── main.go               # Entry point
└── README.md             # CLI documentation
```

## Internal Packages (`/internal`)

Shared Go packages not exposed as public API.

```
internal/
├── api/                  # HTTP API handlers
│   ├── campaigns.go      # Campaign endpoints
│   ├── entities.go       # Entity endpoints
│   ├── relationships.go  # Relationship endpoints
│   ├── sessions.go       # Session endpoints
│   ├── timeline.go       # Timeline endpoints
│   └── import.go         # Import endpoints
├── database/             # Database operations
│   ├── connection.go     # Connection management
│   ├── queries.go        # SQL queries
│   └── migrations.go     # Migration runner
├── models/               # Data structures
│   ├── campaign.go       # Campaign model
│   ├── entity.go         # Entity model
│   ├── relationship.go   # Relationship model
│   ├── session.go        # Session model
│   ├── timeline.go       # Timeline event model
│   └── conflict.go       # Canon conflict model
├── agents/               # Analysis agents
│   ├── consistency.go    # Consistency checker
│   └── interface.go      # Agent interface
└── importers/            # Content importers
    ├── common/           # Shared importer types
    ├── evernote/         # Evernote .enex parser
    └── googledocs/       # Google Docs importer
```

## Client Structure (`/client`)

The React web application for user interaction.

```
client/
├── src/
│   ├── main.tsx          # Entry point
│   ├── App.tsx           # Root component with routing
│   ├── components/       # Reusable UI components
│   │   └── Layout.tsx    # Main layout with navigation
│   ├── pages/            # Page components (routes)
│   │   ├── Dashboard.tsx # Home dashboard
│   │   ├── Campaigns.tsx # Campaign management
│   │   ├── Entities.tsx  # Entity browser
│   │   ├── Timeline.tsx  # Timeline view
│   │   └── Import.tsx    # Content import
│   ├── types/            # TypeScript type definitions
│   │   └── index.ts      # Shared types
│   └── test/             # Test configuration
├── public/               # Static assets
├── package.json          # Dependencies
├── vite.config.ts        # Vite configuration
└── tsconfig.json         # TypeScript configuration
```

## Game System Schemas (`/schemas`)

YAML definitions for supported TTRPG systems.

```
schemas/
├── coc-7e.yaml           # Call of Cthulhu 7th Edition
├── gurps-4e.yaml         # GURPS 4th Edition
└── fitd.yaml             # Forged in the Dark
```

## Database Migrations (`/migrations`)

SQL migration files applied in order.

```
migrations/
├── 001_initial_schema.sql    # Core tables
└── 002_seed_game_systems.sql # Game system data
```

## Scripts (`/scripts`)

Automation and utility scripts.

```
scripts/
├── migrate.sh            # Database migration runner
├── backup.sh             # PostgreSQL backup
└── restore.sh            # Backup restoration
```

## Configuration Files

Key configuration files and their locations:

| File | Location | Purpose |
|------|----------|---------|
| `Makefile` | Root | Build, test, lint commands |
| `go.mod` | Root | Go module dependencies |
| `package.json` | `/client` | Node.js dependencies |
| `tsconfig.json` | `/client` | TypeScript configuration |
| `vite.config.ts` | `/client` | Vite build configuration |
| `docker-compose.yml` | Root | Docker services |
| `.env` | Root | Environment variables |

## Source Code Conventions

All components follow these conventions:

- Go standard project layout (cmd/, internal/)
- Four-space indentation
- Tests co-located with source (Go) or in test/ (React)
- Documentation in `/docs/`
