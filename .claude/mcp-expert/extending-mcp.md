# Extending MCP

Guide to adding new tools and resources.

## Adding a New Tool

### 1. Define Tool Schema

```go
// In tools/definitions.go
var EntityCompareToolDef = ToolDefinition{
    Name:        "entity_compare",
    Description: "Compare two entities to find differences",
    InputSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "entity_id_1": map[string]any{
                "type":        "string",
                "description": "First entity UUID",
            },
            "entity_id_2": map[string]any{
                "type":        "string",
                "description": "Second entity UUID",
            },
        },
        "required": []string{"entity_id_1", "entity_id_2"},
    },
}
```

### 2. Implement Handler

```go
// In tools/entity_compare.go
func (h *Handler) HandleEntityCompare(ctx context.Context, params map[string]any) (any, error) {
    // Extract parameters
    id1, _ := params["entity_id_1"].(string)
    id2, _ := params["entity_id_2"].(string)

    // Validate
    if id1 == "" || id2 == "" {
        return nil, &ToolError{
            Code:    "INVALID_PARAMS",
            Message: "Both entity IDs required",
        }
    }

    // Authorize - verify both entities accessible
    for _, id := range []string{id1, id2} {
        if err := h.authorize(ctx, id); err != nil {
            return nil, err
        }
    }

    // Execute
    entity1, err := h.db.GetEntity(ctx, id1)
    if err != nil {
        return nil, err
    }
    entity2, err := h.db.GetEntity(ctx, id2)
    if err != nil {
        return nil, err
    }

    // Compare
    diff := compareEntities(entity1, entity2)

    // Filter for role
    filtered := filterDiff(diff, h.getRole(ctx))

    return filtered, nil
}
```

### 3. Register Tool

```go
// In server/setup.go
func (s *Server) registerTools() {
    s.tools["entity_compare"] = &Tool{
        Definition: EntityCompareToolDef,
        Handler:    s.handlers.HandleEntityCompare,
    }
}
```

### 4. Add Tests

```go
// In tools/entity_compare_test.go
func TestEntityCompare_Success(t *testing.T) {
    // Setup
    ctx := testContext()
    params := map[string]any{
        "entity_id_1": "uuid-1",
        "entity_id_2": "uuid-2",
    }

    // Execute
    result, err := handler.HandleEntityCompare(ctx, params)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
}

func TestEntityCompare_MissingParam(t *testing.T) {
    params := map[string]any{
        "entity_id_1": "uuid-1",
        // entity_id_2 missing
    }

    result, err := handler.HandleEntityCompare(ctx, params)

    assert.Error(t, err)
    assert.Nil(t, result)
}
```

## Adding a New Resource

### 1. Define Resource

```go
// In resources/definitions.go
var GameSystemResource = ResourceDefinition{
    URI:         "imagineer://systems/{system_code}",
    Name:        "Game System",
    Description: "Game system schema and mechanics",
    MimeType:    "application/json",
}
```

### 2. Implement Handler

```go
// In resources/game_system.go
func (h *Handler) HandleGetGameSystem(ctx context.Context, uri string) (any, error) {
    // Parse URI
    systemCode := extractSystemCode(uri)

    // Get system
    system, err := h.db.GetGameSystem(ctx, systemCode)
    if err != nil {
        return nil, err
    }

    return system, nil
}
```

### 3. Register Resource

```go
// In server/setup.go
func (s *Server) registerResources() {
    s.resources["systems"] = &Resource{
        Definition: GameSystemResource,
        Handler:    s.handlers.HandleGetGameSystem,
    }
}
```

## Security Checklist for New Tools

- [ ] All parameters validated
- [ ] Authorization verified
- [ ] Campaign access checked
- [ ] Keeper notes filtered for players
- [ ] Rate limiting applied
- [ ] Errors don't leak info
- [ ] Audit logging added

## Testing Checklist

- [ ] Success case tested
- [ ] Missing params tested
- [ ] Invalid params tested
- [ ] Authorization tested
- [ ] Player role filtering tested
- [ ] Error responses verified

## Documentation Checklist

- [ ] Tool added to tools-catalog.md
- [ ] Parameters documented
- [ ] Response format documented
- [ ] Examples provided
- [ ] Security notes added
