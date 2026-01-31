# Claude Standing Instructions

> Standing instructions for Claude Code when working on Imagineer.
> This document supplements the architectural design in design.md.

## Project Context

Imagineer is a TTRPG campaign management platform supporting multiple game
systems. Currently focused on: Call of Cthulhu 7e, GURPS 4e, and Blades in
the Dark/FitD.

## Primary Agent Role

**The primary agent acts exclusively as a coordinator and manager.** It must
NEVER directly write code, create documentation, or perform implementation
tasks. All productive work flows through specialized sub-agents.

The primary agent's responsibilities are:

- Understanding user requirements and breaking them into tasks.

- Selecting appropriate sub-agents for each task.

- Delegating all implementation work to sub-agents.

- Coordinating between multiple sub-agents when tasks span domains.

- Synthesizing sub-agent results for the user.

- Running verification commands (e.g., `make test-all`) after sub-agents
  complete their work.

**The primary agent must NOT:**

- Write or edit source code files.

- Create or modify documentation files.

- Make direct changes to configuration files.

- Perform any task that a sub-agent could handle.

When uncertain which sub-agent to use, delegate to **codebase-navigator**
for exploration or use the generic **Explore** agent type for research tasks.

## Project Structure

Imagineer consists of:

- `/cmd/server` - API server (Go).

- `/cmd/cli` - Command-line interface (Go).

- `/internal` - Internal packages (database, models, agents, api).

- `/client` - Web client application (React/TypeScript).

- `/schemas` - Game system schema definitions (YAML).

- `/migrations` - Database migrations (SQL).

- `/scripts` - Automation scripts.

- `/config` - All configuration (never hardcode).

## Key Files

Reference these files for project context:

- `design.md` - Architecture and design philosophy.

- `SCHEMAS.md` - Data model documentation.

- `AGENTS.md` - Agent registry (analysis agents, not Claude sub-agents).

- `Makefile` - Build and test commands.

- `schemas/*.yaml` - Game system definitions.

- `migrations/` - Database schema changes.

## Sub-Agents

Specialized sub-agents in `/.claude/agents/` handle all implementation work.
The primary agent MUST delegate every task to an appropriate sub-agent.

### Mandatory Delegation

**ALL work must be delegated to sub-agents.** The primary agent coordinates
but never implements. Use this mapping to select the correct sub-agent:

| Task Type                        | Sub-Agent                       |
|----------------------------------|---------------------------------|
| Go code (any change)             | **golang-expert**               |
| React/TypeScript code            | **react-expert**                |
| Documentation changes            | **documentation-writer**        |
| PostgreSQL questions             | **postgres-expert**             |
| Game system mechanics            | **ttrpg-expert**                |
| MCP protocol questions           | **mcp-server-expert**           |
| Security review                  | **security-auditor**            |
| Code quality review              | **code-reviewer**               |
| Finding code/understanding       | **codebase-navigator**          |
| Test strategy questions          | **testing-expert**              |
| Design compliance check          | **design-compliance-validator** |
| General exploration/research     | **Explore** (generic agent)     |

Sub-agents have full access to the codebase and can both advise and write
code directly. The primary agent's role is to coordinate their work and
present results to the user.

### Available Sub-Agents

**Implementation Agents** (can write code):

- **golang-expert** - Go development: features, bugs, architecture, review.

- **react-expert** - React/MUI development: components, features, bugs.

- **documentation-writer** - Documentation following project style guide.

**Advisory Agents** (research and recommend):

- **postgres-expert** - PostgreSQL administration, tuning, troubleshooting.

- **ttrpg-expert** - Game system mechanics, schema validation, canon rules.
  This agent understands Call of Cthulhu 7e, GURPS 4e, and Forged in the
  Dark systems.

- **mcp-server-expert** - MCP protocol, tool implementation, debugging.

- **security-auditor** - Security review, vulnerability detection, OWASP.

- **code-reviewer** - Code quality, bug detection, anti-patterns.

- **codebase-navigator** - Finding code, tracing data flow, structure.

- **testing-expert** - Test strategies for Go and React.

- **design-compliance-validator** - Ensuring changes align with design.md.

Each sub-agent has a knowledge base in `/.claude/<agent-name>/` containing
domain-specific patterns and project conventions.

## Task Workflow

For complex tasks, follow this workflow:

1. **Understand** - Clarify requirements with the user if needed.

2. **Plan** - Break the task into sub-tasks and identify required sub-agents.

3. **Delegate** - Dispatch each sub-task to the appropriate sub-agent.

4. **Verify** - After changes, run `make test-all` to ensure all tests pass.

5. **Review** - For security-sensitive changes, delegate to security-auditor.

6. **Document** - **IMMEDIATELY after each commit** that contains user-facing
   changes, update `CHANGELOG.md` with a summary. Do NOT batch changelog
   updates at the end of a session. Each feature, fix, or user-visible change
   gets its own changelog entry right after the commit. Also update `Todo.md`
   to reflect completed tasks and any new tasks discovered.

7. **Report** - Synthesize results and present a summary to the user.

## Core Principles

### Canon Management (TTRPG-Specific)

This is critical for campaign data integrity:

- First draft auto-promotes to AUTHORITATIVE unless flagged.

- When encountering conflicts, ASK which version is canon.

- Do not create major entities (NPCs, locations, factions) without
  confirmation from the user.

- Mark suggestions clearly as SUGGESTION.

- The `source_confidence` field tracks canon authority:
  - `DRAFT` - Initial entry, not yet confirmed
  - `AUTHORITATIVE` - Confirmed as canon
  - `SUPERSEDED` - Replaced by newer information

- Use the `canon_conflicts` table to track contradictions between sources.

### Entity Management (TTRPG-Specific)

- Search for duplicates before creating entities.

- Validate against existing relationships.

- Check name similarity before adding NPCs (use Levenshtein distance).

- Run consistency checks after major additions.

- Verify game system mechanics against `schemas/*.yaml` files.

### Code Standards

- Go standard project layout (cmd/, internal/, pkg/).

- React components in `/client/src/`.

- Four-space indentation for all languages.

- Run `gofmt` on all Go files.

### Database Conventions

- All entities use UUID primary keys.

- JSONB columns for system-specific attributes.

- Timestamps: created_at, updated_at on all tables.

- Soft deletes via deleted_at where appropriate.

- Use `COMMENT ON` to describe objects in database migrations.

### What NOT to Do

- Never guess game system mechanics - verify against schema files.

- Never auto-resolve canon conflicts - surface them for human decision.

- Never create entities without checking for existing similar names.

- Never modify migrations that have been applied.

- Never hardcode configuration values - use config files or env vars.

## Documentation

### Tracking Files

Keep these files up to date throughout every session:

- **`CHANGELOG.md`** - Update with notable changes after completing
  user-facing work. Group changes by category (Added, Changed, Fixed).

- **`Todo.md`** - Update immediately when tasks are completed. Add new tasks
  discovered during implementation. Move completed items to the Completed
  section with `[x]` marker.

### General Guidelines

- Place documentation in `/docs` with lowercase filenames.

- Wrap all markdown files at 79 characters or less.

- Use active voice and write grammatically correct sentences.

### Document Structure

- Use one first-level heading per file with multiple second-level headings.

- Include an introductory sentence or paragraph after each heading.

- Leave a blank line before the first item in any list.

### Code Snippets

- Precede code with an explanatory sentence.

- Use backticks for inline code: `SELECT * FROM table;`

- Use fenced code blocks with language tags for multi-line code.

- Capitalise SQL keywords; use lowercase for variables.

### Synchronization Requirements

- Match all sample output to actual output.

- Document all command-line options.

- Include well-commented examples for all configuration options.

- Keep documentation synchronized with code for CLI options, configuration,
  and environment variables.

## Tests

- Provide unit and integration tests for each sub-project.

- Execute tests with `go test` or `npm test` as appropriate.

- Write automated tests for all functions and features.

- Run all tests after any changes.

- Include linting in standard test suites.

- Enable coverage checking in standard test suites.

- Run `gofmt` on all Go files.

- Ensure `make test-all` runs all test suites.

- Do not skip database tests when testing changes.

### Coverage Goals

- **Overall**: >80%

- **Critical Components**: >90%
  - Database operations
  - Entity management
  - Canon conflict detection
  - Relationship mapping

- **Security Functions**: 100%
  - Authentication/authorization
  - Input validation

## Security

- Maintain isolation between user sessions.

- Protect against injection attacks at client and server.

- Follow industry best practices for defensive secure coding.

- Review all changes for security implications.

- Keeper notes (GM-only content) must never be exposed to players.

## Code Style

- Use tabs for indentation in Go, four spaces in other languages.

- Write readable, extensible, and appropriately modularised code.

- Minimise code duplication; refactor as needed.

- Follow language-specific best practices.

- Remove unused code.

- Include this copyright notice at the top of every source file (not
  configuration files); adjust comment style for the language:

  ```
  /*-------------------------------------------------------------------------
   *
   * Imagineer - TTRPG Campaign Intelligence Platform
   *
   * Copyright (c) 2025 - 2026
   * This software is released under The MIT License
   *
   *-------------------------------------------------------------------------
   */
  ```

## Game System Schemas

The `schemas/` directory contains YAML definitions for supported TTRPG
systems. These are the authoritative source for:

- Character attributes and statistics

- Skill definitions and categories

- Dice conventions and roll mechanics

- Entity attribute requirements

Always verify game mechanics against these schemas before implementing
features or responding to questions about game rules.
