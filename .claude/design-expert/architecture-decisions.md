# Architecture Decisions

This document records major architectural decisions for Imagineer.

## ADR-001: Go Standard Project Layout

**Status**: Accepted

**Context**: Need a consistent project structure.

**Decision**: Use Go standard project layout with cmd/, internal/, pkg/.

**Rationale**:

- Industry standard, well understood
- Clear separation of public and private code
- Supports multiple binaries (server, cli)

## ADR-002: PostgreSQL with JSONB

**Status**: Accepted

**Context**: Need flexible schema for multiple game systems.

**Decision**: Use PostgreSQL with JSONB columns for system-specific attributes.

**Rationale**:

- JSONB provides schema flexibility
- GIN indexes support fast queries
- PostgreSQL is mature and well-supported
- Avoids NoSQL complexity

## ADR-003: UUID Primary Keys

**Status**: Accepted

**Context**: Need globally unique identifiers.

**Decision**: Use UUIDs for all primary keys.

**Rationale**:

- Globally unique without coordination
- Safe for distributed systems (future)
- No sequential ID leakage

## ADR-004: YAML Game System Schemas

**Status**: Accepted

**Context**: Need to define game system mechanics.

**Decision**: Use YAML files in /schemas directory.

**Rationale**:

- Human-readable and editable
- Version controllable
- Language-agnostic
- Easy to add new systems

## ADR-005: Canon Confidence Levels

**Status**: Accepted

**Context**: Need to track reliability of information.

**Decision**: Three-level confidence: DRAFT, AUTHORITATIVE, SUPERSEDED.

**Rationale**:

- Simple to understand
- Covers main use cases
- Prevents silent data loss

## ADR-006: Entity Polymorphism via Type Column

**Status**: Accepted

**Context**: Need multiple entity types in one table.

**Decision**: Single entities table with entity_type discriminator.

**Rationale**:

- Simpler queries than table-per-type
- Easier relationship mapping
- JSONB handles type-specific attributes
- Supports future entity types

## ADR-007: React with Material-UI

**Status**: Accepted

**Context**: Need web client framework.

**Decision**: React 18 with TypeScript and Material-UI.

**Rationale**:

- Type safety with TypeScript
- Consistent design with MUI
- Large ecosystem
- React Query for server state

## ADR-008: Importer Interface Pattern

**Status**: Accepted

**Context**: Need extensible import system.

**Decision**: Common Importer interface with ImportResult return type.

**Rationale**:

- Easy to add new import sources
- Consistent result handling
- Separation from storage logic
- Testable in isolation

## ADR-009: Conflict Detection vs Resolution

**Status**: Accepted

**Context**: How to handle conflicting information.

**Decision**: Detect and record conflicts; require human resolution.

**Rationale**:

- Preserves human judgment for narrative decisions
- No silent data loss
- Audit trail of conflicts
- GM makes final call

## ADR-010: GM Notes as First-Class Field

**Status**: Accepted

**Context**: Need GM-only content on entities.

**Decision**: Dedicated gm_notes column, never exposed to players.

**Rationale**:

- Clear separation of concerns
- Easy to filter for player views
- No accidental exposure via JSONB

## ADR-011: Session-Based Discovery Tracking

**Status**: Accepted

**Context**: Need to track when players learn information.

**Decision**: Link entity discoveries to sessions.

**Rationale**:

- Natural mapping to play sessions
- Enables "what do players know" queries
- Supports session recaps

## ADR-012: Relationship Tones

**Status**: Accepted

**Context**: Relationships need emotional context.

**Decision**: Predefined tone enum: friendly, hostile, neutral, etc.

**Rationale**:

- Enables relationship visualization
- Consistent across campaigns
- Easy to query and filter
- Extensible via new enum values

## Future Decisions Pending

### Multi-User Access

How to handle multiple users per campaign (players, co-GMs)?

- Options: Shared campaigns, invitation system, role-based
- Deferred until core features complete

### Real-Time Collaboration

How to handle simultaneous editing?

- Options: Locking, operational transforms, conflict-free types
- Deferred until multi-user implemented

### Plugin System

How to allow custom entity types and behaviors?

- Options: Dynamic schemas, plugin packages, webhooks
- Deferred until core is stable
