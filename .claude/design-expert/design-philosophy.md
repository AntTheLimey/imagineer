# Design Philosophy

This document captures the core design philosophy and guiding principles
for Imagineer.

## Mission Statement

Imagineer is a TTRPG Campaign Intelligence Platform that helps Game Masters
manage complex campaign worlds across multiple game systems while preserving
narrative consistency and canon integrity.

## Primary Design Goals

### 1. Canon Integrity

The most critical goal is maintaining consistent, trustworthy campaign data.

- **Source Tracking**: Every piece of information tracks its origin
- **Confidence Levels**: Data is marked as DRAFT, AUTHORITATIVE, or SUPERSEDED
- **Conflict Detection**: System identifies contradictions between sources
- **Human Resolution**: Canon conflicts require human decision, never auto-resolved

### 2. Multi-System Support

Support multiple TTRPG systems without system-specific coupling.

- **Schema-Driven**: Game systems defined in YAML, not code
- **Extensible Attributes**: JSONB columns adapt to any system
- **Consistent Core**: Entity model works across all systems
- **System-Specific Validation**: Schemas validate system-specific data

### 3. Separation of Concerns

Clear boundaries between GM and player information.

- **GM Notes**: GM-only content never exposed to players
- **Discovery Tracking**: Track when players learn information
- **Session-Based Visibility**: Content revealed through play
- **Access Control**: API enforces visibility rules

### 4. Relationship-First Design

Entities are defined by their connections as much as their attributes.

- **Explicit Relationships**: All connections are first-class objects
- **Relationship Metadata**: Tone, strength, bidirectionality captured
- **Graph Navigation**: Easy traversal of entity networks
- **Conflict Implications**: Relationship changes may create conflicts

## Architectural Principles

### Simplicity Over Flexibility

Choose simpler designs that solve known problems over flexible designs
that might solve future problems.

- Avoid premature abstraction
- Solve current requirements well
- Refactor when new requirements emerge

### Explicit Over Implicit

Make behavior explicit rather than relying on conventions or magic.

- Configuration in code, not convention
- Clear error messages
- Explicit type definitions
- Documented behavior

### Safety Over Convenience

Protect data integrity even at cost of convenience.

- Duplicate name checks before creation
- Canon conflicts require resolution
- No silent data overwrites
- Validation at boundaries

### Testability

Design for testability from the start.

- Interfaces for dependencies
- Minimal global state
- Pure functions where possible
- Clear test boundaries

## Technology Choices

### Backend: Go

- Strong typing for reliability
- Excellent concurrency model
- Fast compilation and execution
- Standard library sufficiency

### Frontend: React/TypeScript

- Type safety with TypeScript
- Component-based architecture
- Material-UI for consistent design
- React Query for server state

### Database: PostgreSQL

- JSONB for flexible schemas
- UUID primary keys for distribution
- GIN indexes for JSON queries
- Trigram indexes for fuzzy matching

### Game System Schemas: YAML

- Human-readable format
- Easy to version control
- Simple to extend
- Language-agnostic

## Quality Standards

### Code Quality

- 80%+ test coverage
- No linter warnings
- Documented public APIs
- Four-space indentation

### Data Quality

- All entities have names
- Relationships have both endpoints
- Dates have precision indicators
- Sources are tracked

### User Experience

- Responsive interface
- Clear error messages
- Consistent navigation
- Accessible design

## Success Criteria

Imagineer is successful when:

1. GMs can manage campaigns without losing track of canon
2. Canon conflicts are detected and surfaced for resolution
3. Multiple game systems work with the same core interface
4. GM notes remain private from players
5. Entity relationships are easy to explore and update
6. Importing content preserves source attribution
