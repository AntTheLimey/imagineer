# Design Expert Knowledge Base

This directory contains documentation about the Imagineer design philosophy,
architecture, and implementation guidance.

## Purpose

These documents serve as the knowledge base for the Design Compliance Validator
role - an agent that ensures all implementation work aligns with the
architectural vision and design goals documented in design.md.

## Documents

### Core Design Documentation

#### [design-philosophy.md](design-philosophy.md)

Captures the core design philosophy, goals, and principles.

#### [architecture-decisions.md](architecture-decisions.md)

Records major architectural decisions and rationales.

#### [security-model.md](security-model.md)

Describes the security design and access control model.

#### [component-responsibilities.md](component-responsibilities.md)

Defines boundaries and responsibilities for each component.

### Implementation Guidance

#### [development-guidelines.md](development-guidelines.md)

Provides guidelines for maintaining design consistency.

#### [recent-changes.md](recent-changes.md)

Tracks significant architectural changes and implications.

## How to Use These Documents

### For Design Compliance Validation

1. Identify relevant design goals from design-philosophy.md
2. Check architecture-decisions.md for precedents
3. Verify component boundaries per component-responsibilities.md
4. Validate security per security-model.md
5. Review patterns in development-guidelines.md

### For Feature Implementation

1. Understand design intent from design-philosophy.md
2. Find precedents in architecture-decisions.md
3. Determine owning component from component-responsibilities.md
4. Follow patterns in development-guidelines.md
5. Check security requirements in security-model.md

## Compliance Levels

- **COMPLIANT**: Implementation fully aligns with design goals
- **MINOR DEVIATION**: Small choices that don't affect core goals
- **MODERATE CONCERN**: Partially conflicts but reconcilable
- **DESIGN VIOLATION**: Clear contradiction of design principles

## Relationship to Other Documentation

### design.md (Root)

Primary authority for architectural decisions. These documents elaborate.

### CLAUDE.md (Root)

Standing instructions for development practices. Complements these docs.

### SCHEMAS.md

Data model documentation. Supplements database design here.

Last Updated: 2026-01-30
