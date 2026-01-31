# PostgreSQL Expert Knowledge Base

This directory contains PostgreSQL guidance for Imagineer.

## Documents

- `schema-overview.md` - Database schema documentation
- `indexing-strategy.md` - Index design patterns
- `jsonb-patterns.md` - JSONB column usage

## Quick Reference

### Core Tables

| Table | Purpose |
|-------|---------|
| `game_systems` | TTRPG system definitions |
| `campaigns` | Individual campaigns |
| `entities` | NPCs, locations, items, etc. |
| `relationships` | Entity connections |
| `sessions` | Game sessions |
| `timeline_events` | Chronological events |
| `canon_conflicts` | Source contradictions |

### Key Conventions

- UUID primary keys
- JSONB for flexible attributes
- GIN indexes for JSONB/arrays
- Trigram indexes for fuzzy search
- Timestamps on all tables

### Connection String

```
postgresql://user:pass@host:5432/imagineer?sslmode=require
```

Last Updated: 2026-01-30
