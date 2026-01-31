# Testing MCP

Testing strategies for MCP components.

## Unit Testing Tools

### Basic Handler Test

```go
func TestEntitySearch_Success(t *testing.T) {
    // Setup
    db := NewMockDB()
    db.On("SearchEntities", mock.Anything, "campaign-1", "").
        Return([]Entity{{ID: "1", Name: "Test"}}, nil)

    handler := NewHandler(db)
    ctx := contextWithUser("user-1", "gm")

    // Execute
    params := map[string]any{
        "campaign_id": "campaign-1",
    }
    result, err := handler.HandleEntitySearch(ctx, params)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)

    entities := result.(map[string]any)["entities"].([]Entity)
    assert.Len(t, entities, 1)
}
```

### Authorization Test

```go
func TestEntitySearch_Unauthorized(t *testing.T) {
    db := NewMockDB()
    db.On("GetCampaign", mock.Anything, "campaign-1").
        Return(&Campaign{OwnerID: "other-user"}, nil)

    handler := NewHandler(db)
    ctx := contextWithUser("user-1", "gm")  // Wrong user

    params := map[string]any{
        "campaign_id": "campaign-1",
    }
    _, err := handler.HandleEntitySearch(ctx, params)

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "forbidden")
}
```

### Role Filtering Test

```go
func TestEntityGet_PlayerFilteredNotes(t *testing.T) {
    db := NewMockDB()
    db.On("GetEntity", mock.Anything, "entity-1").
        Return(&Entity{
            ID:          "entity-1",
            Name:        "NPC",
            GMNotes: "Secret info",  // Should be filtered
        }, nil)

    handler := NewHandler(db)
    ctx := contextWithUser("user-1", "player")  // Player role

    params := map[string]any{
        "entity_id": "entity-1",
    }
    result, err := handler.HandleEntityGet(ctx, params)

    assert.NoError(t, err)
    entity := result.(*Entity)
    assert.Empty(t, entity.GMNotes)  // Filtered
}
```

## Integration Testing

### Full Tool Flow

```go
func TestToolFlow_EntityCRUD(t *testing.T) {
    server := setupTestServer(t)
    defer server.Close()

    // Create entity
    createResult, err := server.CallTool("entity_create", map[string]any{
        "campaign_id": testCampaignID,
        "name":        "Test Entity",
        "type":        "npc",
    })
    assert.NoError(t, err)
    entityID := createResult["id"].(string)

    // Get entity
    getResult, err := server.CallTool("entity_get", map[string]any{
        "entity_id": entityID,
    })
    assert.NoError(t, err)
    assert.Equal(t, "Test Entity", getResult["name"])

    // Search entity
    searchResult, err := server.CallTool("entity_search", map[string]any{
        "campaign_id": testCampaignID,
        "query":       "Test",
    })
    assert.NoError(t, err)
    entities := searchResult["entities"].([]any)
    assert.Len(t, entities, 1)
}
```

### Protocol Test

```go
func TestMCPProtocol_Initialize(t *testing.T) {
    client := NewTestClient(t)

    result, err := client.Initialize()

    assert.NoError(t, err)
    assert.Equal(t, "2024-11-05", result.ProtocolVersion)
    assert.True(t, result.Capabilities.Tools)
}

func TestMCPProtocol_ListTools(t *testing.T) {
    client := NewTestClient(t)
    client.Initialize()

    tools, err := client.ListTools()

    assert.NoError(t, err)
    assert.Contains(t, toolNames(tools), "entity_search")
    assert.Contains(t, toolNames(tools), "entity_get")
}
```

## Error Testing

### Invalid Parameters

```go
func TestEntitySearch_InvalidCampaignID(t *testing.T) {
    handler := NewHandler(db)
    ctx := contextWithUser("user-1", "gm")

    params := map[string]any{
        "campaign_id": "not-a-uuid",
    }
    _, err := handler.HandleEntitySearch(ctx, params)

    assert.Error(t, err)
    toolErr := err.(*ToolError)
    assert.Equal(t, "INVALID_PARAMS", toolErr.Code)
}

func TestEntitySearch_MissingRequired(t *testing.T) {
    handler := NewHandler(db)
    ctx := contextWithUser("user-1", "gm")

    params := map[string]any{}  // Missing campaign_id
    _, err := handler.HandleEntitySearch(ctx, params)

    assert.Error(t, err)
    toolErr := err.(*ToolError)
    assert.Equal(t, "INVALID_PARAMS", toolErr.Code)
}
```

### Database Errors

```go
func TestEntityGet_DatabaseError(t *testing.T) {
    db := NewMockDB()
    db.On("GetEntity", mock.Anything, "entity-1").
        Return(nil, errors.New("connection failed"))

    handler := NewHandler(db)
    ctx := contextWithUser("user-1", "gm")

    params := map[string]any{
        "entity_id": "entity-1",
    }
    _, err := handler.HandleEntityGet(ctx, params)

    assert.Error(t, err)
    // Verify error doesn't leak internal details
    assert.NotContains(t, err.Error(), "connection failed")
}
```

## Test Utilities

### Context Helpers

```go
func contextWithUser(userID, role string) context.Context {
    ctx := context.Background()
    ctx = context.WithValue(ctx, userIDKey, userID)
    ctx = context.WithValue(ctx, roleKey, role)
    return ctx
}
```

### Mock Database

```go
type MockDB struct {
    mock.Mock
}

func (m *MockDB) GetEntity(ctx context.Context, id string) (*Entity, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*Entity), args.Error(1)
}
```

### Test Server

```go
func setupTestServer(t *testing.T) *TestServer {
    db := setupTestDB(t)
    server := NewServer(db)
    server.Start()
    t.Cleanup(server.Close)
    return &TestServer{server}
}
```

## Coverage Requirements

| Component | Target |
|-----------|--------|
| Tool handlers | 90% |
| Authorization | 100% |
| Error handling | 85% |
| Protocol handling | 80% |
