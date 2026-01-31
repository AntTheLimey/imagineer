# Go Expert Knowledge Base

This directory contains Go development guidance for Imagineer.

## Documents

- `architecture-overview.md` - Project structure and patterns
- `code-conventions.md` - Coding standards
- `database-patterns.md` - Database access patterns
- `testing-strategy.md` - Go testing approach

## Quick Reference

### Project Layout

```
cmd/server/     - API server entry point
cmd/cli/        - CLI entry point
internal/       - Private packages
  api/          - HTTP handlers
  database/     - Database operations
  models/       - Data structures
  agents/       - Analysis agents
  importers/    - Content importers
```

### Key Packages

| Package | Purpose |
|---------|---------|
| `internal/api` | HTTP handlers |
| `internal/database` | PostgreSQL access |
| `internal/models` | Data structures |
| `internal/importers/common` | Importer interface |

Last Updated: 2026-01-30
