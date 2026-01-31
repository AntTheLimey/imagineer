# Imagineer Development Guide

This guide covers setting up your local development environment for Imagineer, a TTRPG campaign management platform.

## 1. Prerequisites

Before you begin, ensure you have the following installed:

| Tool | Minimum Version | Notes |
|------|-----------------|-------|
| Go | 1.23+ | [Download](https://go.dev/dl/) |
| Node.js | 20+ | [Download](https://nodejs.org/) or use nvm |
| PostgreSQL | 16+ | Local install or Docker |
| Docker | Latest | [Download](https://www.docker.com/products/docker-desktop/) |
| Docker Compose | Latest | Included with Docker Desktop |
| Make | Any | Pre-installed on macOS/Linux |

Verify your installations:

```bash
go version          # go1.23.x or later
node --version      # v20.x.x or later
docker --version    # Docker version 24.x or later
make --version      # GNU Make 3.x or later
```

## 2. Initial Setup

### Clone and Configure

```bash
git clone https://github.com/AntTheLimey/imagineer.git
cd imagineer

# Copy configuration files
cp .env.example .env
cp config/db/db.json.example config/db/db.json
cp .mcp.json.example .mcp.json

# Edit .env and config files with your database credentials
```

Review and update the copied configuration files:

- `.env` - Environment variables for the application
- `config/db/db.json` - Database connection settings
- `.mcp.json` - MCP server configuration (for AI assistant integration)

### Database Setup (Docker - Recommended)

Using Docker is the recommended approach as it provides a consistent, isolated environment.

```bash
# Start PostgreSQL container
make up

# Run database migrations
make migrate

# Check migration status
make migrate-status

# Access PostgreSQL shell
make psql
```

### Database Setup (Local PostgreSQL)

If you prefer running PostgreSQL locally:

```bash
# Create database and user
createdb imagineer
createuser imagineer -P
# Enter a password when prompted

# Grant privileges
psql -c "GRANT ALL PRIVILEGES ON DATABASE imagineer TO imagineer;"

# Update config/db/db.json with your credentials
# Example:
# {
#   "host": "localhost",
#   "port": 5432,
#   "user": "imagineer",
#   "password": "your_password",
#   "database": "imagineer",
#   "sslmode": "disable"
# }

# Run migrations
for f in migrations/*.sql; do psql -d imagineer -f "$f"; done
```

## 3. Running the Application

### Backend (Go API Server)

The Go server handles API requests and database operations.

```bash
# Run with default settings (port 8080)
go run ./cmd/server

# Run on a different port (recommended if MCP server uses 8080)
PORT=8081 go run ./cmd/server

# Or build and run the binary
make build
./bin/imagineer

# With environment variables
PORT=8081 DB_HOST=localhost ./bin/imagineer
```

The server will be available at `http://localhost:8080` (or your configured port).

### Frontend (React)

The React frontend provides the user interface.

```bash
cd client
npm install
npm run dev
```

The development server starts at `http://localhost:5173`.

The frontend proxies `/api` requests to the backend automatically (configured in `vite.config.ts`). Ensure your backend is running before using the frontend.

### Running Both Services

For full development, run both in separate terminals:

**Terminal 1 - Backend:**
```bash
PORT=8081 go run ./cmd/server
```

**Terminal 2 - Frontend:**
```bash
cd client && npm run dev
```

## 4. Running Tests

### All Tests

```bash
make test-all
```

### Go Tests Only

```bash
# Using make
make test-server

# Using go directly
go test -v ./...

# Test a specific package
go test -v ./internal/api/...

# Run with race detection
go test -race ./...
```

### React Tests Only

```bash
# Using make
make test-client

# Using npm directly
cd client && npm test

# Watch mode
cd client && npm test -- --watch

# Run a specific test file
cd client && npm test -- src/hooks/useCharacters.test.ts
```

### Test Coverage

```bash
# Generate coverage report
make coverage

# Go coverage with HTML report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
open coverage.html

# React coverage
cd client && npm test -- --coverage
```

## 5. Code Quality

### Linting

```bash
# Run all linters
make lint

# Go linting (requires golangci-lint)
golangci-lint run

# React/TypeScript linting
cd client && npm run lint
```

### Formatting

```bash
# Format Go code
gofmt -w .

# Or use goimports for import organization
goimports -w .

# Format React/TypeScript code
cd client && npm run format
# Or with prettier directly
cd client && npx prettier --write src/
```

### Type Checking

```bash
# TypeScript type checking (no emit)
cd client && npx tsc --noEmit
```

## 6. Database Operations

### Container Management

```bash
# View container status
make status

# View container logs
make logs

# Follow logs in real-time
make logs-f

# Stop services (preserves data)
make down

# Stop and remove volumes (DESTRUCTIVE - loses all data)
make reset
```

### Backup and Restore

```bash
# Create a backup
make backup
# Creates: backups/imagineer_YYYYMMDD_HHMMSS.sql

# List available backups
ls -la backups/

# Restore from a backup
make restore BACKUP=backups/imagineer_20240115_143022.sql
```

### Migration Management

```bash
# Check current migration status
make migrate-status

# Run pending migrations
make migrate

# Create a new migration
# Name format: YYYYMMDDHHMMSS_description.sql
touch migrations/$(date +%Y%m%d%H%M%S)_add_new_feature.sql
```

**Important:** Never modify migrations that have already been applied to shared environments.

## 7. Project Structure

```
imagineer/
├── cmd/
│   └── server/              # Main application entry point
│       └── main.go          # Server initialization and startup
│
├── internal/                # Private application code
│   ├── api/                 # HTTP handlers and routing
│   │   ├── handlers/        # Request handlers by domain
│   │   ├── middleware/      # Auth, logging, CORS, etc.
│   │   └── routes.go        # Route definitions
│   │
│   ├── database/            # PostgreSQL queries and connections
│   │   ├── queries/         # SQL query functions
│   │   └── connection.go    # Database connection pool
│   │
│   ├── models/              # Data structures and domain types
│   │   ├── campaign.go      # Campaign entity
│   │   ├── character.go     # Character entity
│   │   └── ...
│   │
│   └── importers/           # External data importers
│       ├── evernote/        # Evernote ENEX parser
│       └── googledocs/      # Google Docs parser
│
├── client/                  # React frontend
│   └── src/
│       ├── api/             # API client and request functions
│       │   └── client.ts    # Axios/fetch configuration
│       │
│       ├── hooks/           # React Query hooks
│       │   ├── useCharacters.ts
│       │   ├── useCampaigns.ts
│       │   └── ...
│       │
│       ├── pages/           # Page components (routes)
│       │   ├── Dashboard/
│       │   ├── Characters/
│       │   └── ...
│       │
│       ├── components/      # Reusable UI components
│       └── types/           # TypeScript type definitions
│
├── config/                  # Configuration files
│   └── db/                  # Database configuration
│
├── migrations/              # SQL migration files
│   ├── 001_initial.sql
│   └── ...
│
├── schemas/                 # Game system definitions (YAML)
│   ├── coc7e.yaml          # Call of Cthulhu 7e
│   ├── gurps4e.yaml        # GURPS 4th Edition
│   └── bitd.yaml           # Blades in the Dark
│
├── docs/                    # Documentation
└── Makefile                 # Build and development commands
```

### Key Directories Explained

| Directory | Purpose |
|-----------|---------|
| `cmd/server/` | Application entry point. Contains `main.go` which initializes the server, database connection, and routes. |
| `internal/api/` | HTTP layer. Handlers process requests, middleware handles cross-cutting concerns like authentication and logging. |
| `internal/database/` | Data access layer. Contains SQL queries wrapped in Go functions, connection pooling, and transaction management. |
| `internal/models/` | Domain models. Go structs representing entities like campaigns, characters, and sessions. Includes validation logic. |
| `internal/importers/` | Import utilities. Parsers for bringing in content from Evernote exports and Google Docs. |
| `client/src/api/` | Frontend API client. Configured HTTP client and typed request functions for each endpoint. |
| `client/src/hooks/` | React Query hooks. Data fetching and caching logic using TanStack Query. Provides loading states and mutations. |
| `client/src/pages/` | Route components. Each page corresponds to a route and composes components and hooks. |
| `schemas/` | Game system schemas. YAML files defining attributes, skills, and mechanics for each supported TTRPG system. |
| `migrations/` | Database migrations. Sequential SQL files that build up the database schema. Applied in order by filename. |

## 8. Environment Variables

Key environment variables (see `.env.example` for complete list):

### Application

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server listen port | `8080` |
| `ENV` | Environment (`development`, `production`) | `development` |
| `LOG_LEVEL` | Logging verbosity (`debug`, `info`, `warn`, `error`) | `info` |

### Database

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | Database user | `imagineer` |
| `DB_PASSWORD` | Database password | - |
| `DB_NAME` | Database name | `imagineer` |
| `DB_SSLMODE` | SSL mode (`disable`, `require`) | `disable` |

### Frontend

| Variable | Description | Default |
|----------|-------------|---------|
| `VITE_API_URL` | Backend API URL | `/api` |

## 9. Troubleshooting

### Port Conflicts

**Problem:** "address already in use" error when starting the server.

**Solution:**
```bash
# Find what's using the port
lsof -i :8080

# Kill the process or use a different port
PORT=8081 go run ./cmd/server
```

### PostgreSQL Connection Issues

**Problem:** "connection refused" or "no pg_hba.conf entry" errors.

**Solutions:**

1. **Check PostgreSQL is running:**
   ```bash
   # Docker
   docker ps | grep postgres
   make status

   # Local
   pg_isready
   ```

2. **Check listen_addresses (local PostgreSQL):**
   Edit `postgresql.conf`:
   ```
   listen_addresses = 'localhost'  # or '*' for all interfaces
   ```

3. **Check pg_hba.conf:**
   Add or modify the entry:
   ```
   host    imagineer    imagineer    127.0.0.1/32    md5
   ```

4. **Restart PostgreSQL after config changes:**
   ```bash
   # macOS with Homebrew
   brew services restart postgresql

   # Linux
   sudo systemctl restart postgresql
   ```

### Node Modules Issues

**Problem:** Missing dependencies, version conflicts, or corrupted packages.

**Solution:**
```bash
cd client
rm -rf node_modules package-lock.json
npm install
```

### Go Module Issues

**Problem:** Missing or outdated Go dependencies.

**Solution:**
```bash
go mod tidy
go mod download
```

### Docker Issues

**Problem:** Container won't start or behaves unexpectedly.

**Solutions:**

1. **Check container logs:**
   ```bash
   make logs
   docker logs imagineer-postgres
   ```

2. **Reset containers completely:**
   ```bash
   make down
   docker volume rm imagineer_postgres_data  # if exists
   make up
   make migrate
   ```

3. **Rebuild containers:**
   ```bash
   docker-compose down
   docker-compose build --no-cache
   docker-compose up -d
   ```

### Migration Errors

**Problem:** Migration fails or database is in inconsistent state.

**Solutions:**

1. **Check which migrations have been applied:**
   ```bash
   make migrate-status
   # Or query directly
   make psql
   SELECT * FROM schema_migrations;
   ```

2. **For development, reset and rerun:**
   ```bash
   make reset
   make migrate
   ```

3. **For specific migration issues, check the SQL:**
   ```bash
   # Review the failing migration
   cat migrations/NNNN_name.sql
   ```

### Frontend Proxy Issues

**Problem:** API requests fail with 404 or CORS errors.

**Solutions:**

1. **Ensure backend is running on expected port:**
   Check `vite.config.ts` for proxy configuration.

2. **Check the proxy target matches your backend:**
   ```typescript
   // vite.config.ts
   server: {
     proxy: {
       '/api': {
         target: 'http://localhost:8081',  // Match your backend port
         changeOrigin: true,
       },
     },
   },
   ```

3. **Restart Vite dev server after config changes:**
   ```bash
   cd client && npm run dev
   ```

## 10. Getting Help

- Check existing issues: https://github.com/AntTheLimey/imagineer/issues
- Review project documentation in the `docs/` directory
- Consult `CLAUDE.md` for AI assistant guidelines
