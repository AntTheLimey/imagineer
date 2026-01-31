# Codebase Navigator Knowledge Base

This directory contains structural information about the Imagineer codebase
to help navigate and understand the project.

## Purpose

This knowledge base provides:

- Project directory structure
- Feature implementation locations
- Data flow documentation
- Key files and their purposes

## Documents

### [project-structure.md](project-structure.md)

Complete directory layout and organization.

### [feature-locations.md](feature-locations.md)

Where specific features are implemented.

### [data-flow.md](data-flow.md)

How data moves between components.

### [key-files.md](key-files.md)

Critical files and their purposes.

## Quick Reference

### Component Overview

| Component | Location | Purpose |
|-----------|----------|---------|
| API Server | `/cmd/server` | HTTP API for web client |
| CLI | `/cmd/cli` | Command-line interface |
| Database | `/internal/database` | PostgreSQL operations |
| Models | `/internal/models` | Data structures |
| API Handlers | `/internal/api` | HTTP handlers |
| Agents | `/internal/agents` | Analysis agents |
| Importers | `/internal/importers` | Content importers |
| Web Client | `/client` | React UI |

### Related Knowledge Bases

- Go patterns: `/.claude/golang-expert/`
- React patterns: `/.claude/react-expert/`
- Database: `/.claude/postgres-expert/`
- TTRPG concepts: `/.claude/ttrpg-expert/`

Last Updated: 2026-01-30
