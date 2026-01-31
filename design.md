# Imagineer Design Document

## Overview
Imagineer is a system-agnostic TTRPG campaign intelligence platform.

## Architecture Decisions

### ADR-001: PostgreSQL with JSONB for Flexibility
- **Decision:** Use PostgreSQL with JSONB columns for system-specific attributes
- **Rationale:** Provides both relational integrity and schema flexibility
- **Consequences:** Need careful indexing strategy for JSONB queries

### ADR-002: pgEdge Postgres MCP Server for AI Integration
- **Decision:** Use pgEdge MCP server for Claude Desktop/Code integration
- **Rationale:** Go-based, extensible custom definitions, production-ready Docker support
- **Consequences:** Currently read-only; may need separate write API or wait for write mode

### ADR-003: Agent-Based Analysis Tools
- **Decision:** Implement analysis as composable agents rather than monolithic features
- **Rationale:** Flexibility, testability, can evolve independently
- **Consequences:** Need clear agent interface contracts

## Data Model
See SCHEMAS.md for detailed entity schemas.

## Technology Stack
- **Language:** Go (standard project layout)
- **Database:** pgEdge Postgres 18 (Docker)
- **MCP Server:** pgEdge Postgres MCP (cloned to ~/PROJECTS/pgedge-postgres-mcp)
- **Configuration:** YAML files
- **Containerization:** Docker Compose
