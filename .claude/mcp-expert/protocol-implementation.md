# MCP Protocol Implementation

Details on implementing MCP for Imagineer.

## Protocol Overview

MCP (Model Context Protocol) enables LLM clients to interact with
external tools and resources in a standardized way.

### Message Types

| Type | Direction | Purpose |
|------|-----------|---------|
| Initialize | Client → Server | Start session |
| Call Tool | Client → Server | Execute tool |
| List Tools | Client → Server | Get available tools |
| List Resources | Client → Server | Get available resources |
| Response | Server → Client | Return results |

## Tool Definition

### Structure

```json
{
    "name": "entity_search",
    "description": "Search for entities by name or type",
    "inputSchema": {
        "type": "object",
        "properties": {
            "campaign_id": {
                "type": "string",
                "description": "Campaign UUID"
            },
            "query": {
                "type": "string",
                "description": "Search query"
            },
            "entity_type": {
                "type": "string",
                "enum": ["npc", "location", "item", ...],
                "description": "Filter by entity type"
            }
        },
        "required": ["campaign_id"]
    }
}
```

### Implementation

```go
func (s *Server) handleEntitySearch(ctx context.Context, params map[string]any) (any, error) {
    // 1. Extract parameters
    campaignID, ok := params["campaign_id"].(string)
    if !ok {
        return nil, errors.New("campaign_id required")
    }

    // 2. Authorize
    if err := s.authorize(ctx, campaignID); err != nil {
        return nil, err
    }

    // 3. Execute search
    entities, err := s.db.SearchEntities(ctx, campaignID, query)
    if err != nil {
        return nil, err
    }

    // 4. Filter for role
    filtered := filterForRole(entities, s.getRole(ctx))

    // 5. Return results
    return filtered, nil
}
```

## Resource Definition

### Structure

```json
{
    "uri": "imagineer://campaigns/{id}/entities",
    "name": "Campaign Entities",
    "description": "List of entities in the campaign",
    "mimeType": "application/json"
}
```

### Implementation

```go
func (s *Server) handleGetResource(ctx context.Context, uri string) (any, error) {
    // Parse URI
    parts := parseResourceURI(uri)

    switch parts.Resource {
    case "entities":
        return s.getEntities(ctx, parts.CampaignID)
    case "timeline":
        return s.getTimeline(ctx, parts.CampaignID)
    default:
        return nil, errors.New("unknown resource")
    }
}
```

## Error Handling

### Error Response

```json
{
    "error": {
        "code": "NOT_FOUND",
        "message": "Entity not found"
    }
}
```

### Error Codes

| Code | Meaning |
|------|---------|
| `INVALID_PARAMS` | Bad input parameters |
| `NOT_FOUND` | Resource not found |
| `UNAUTHORIZED` | Authentication required |
| `FORBIDDEN` | Access denied |
| `INTERNAL_ERROR` | Server error |

## Session Management

### Initialization

```go
func (s *Server) handleInitialize(ctx context.Context, params InitParams) (*InitResult, error) {
    return &InitResult{
        ProtocolVersion: "2024-11-05",
        Capabilities: Capabilities{
            Tools:     true,
            Resources: true,
        },
        ServerInfo: ServerInfo{
            Name:    "Imagineer",
            Version: "0.1.0",
        },
    }, nil
}
```

### Authentication

For Imagineer MCP:

1. Client provides session token
2. Server validates token
3. Server determines user and role
4. All subsequent calls use this context

## Rate Limiting

| Operation | Limit |
|-----------|-------|
| Tool calls | 60/minute |
| Resource reads | 100/minute |
| List operations | 10/minute |

## Logging

### Safe Logging

```go
log.Info("tool_call",
    "tool", toolName,
    "campaign_id", campaignID,
    "user_id", userID,
    // Never log: query content, entity details
)
```

### Audit Trail

Record for audit:

- Tool name
- Timestamp
- User ID
- Campaign ID
- Success/failure
