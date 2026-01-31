# MCP Expert Knowledge Base

This directory contains MCP protocol guidance for Imagineer.

## Purpose

This knowledge base provides:

- MCP protocol implementation details
- Tool and resource patterns
- Authentication guidance
- Testing strategies

## Documents

### [protocol-implementation.md](protocol-implementation.md)

MCP protocol details and wire format.

### [tools-catalog.md](tools-catalog.md)

Available MCP tools for Imagineer.

### [extending-mcp.md](extending-mcp.md)

How to add new tools and resources.

### [testing-mcp.md](testing-mcp.md)

Testing strategies for MCP components.

## Imagineer MCP Context

Imagineer uses MCP to expose campaign management tools to LLM clients
(like Claude). This enables AI-assisted campaign management.

### Planned Tools

| Tool | Purpose |
|------|---------|
| `entity_search` | Find entities by name or type |
| `entity_get` | Get entity details |
| `relationship_map` | Explore entity connections |
| `timeline_query` | Query timeline events |
| `canon_check` | Check for conflicts |

### Security Considerations

- All tools must respect GM/player access
- GM notes never exposed via MCP
- Campaign ownership verified
- Rate limiting applied

Last Updated: 2026-01-30
