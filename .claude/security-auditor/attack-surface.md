# Attack Surface

This document maps API endpoints and input validation requirements.

## API Endpoints

### Campaign Endpoints

| Endpoint | Method | Auth | Input Validation |
|----------|--------|------|------------------|
| `/api/campaigns` | GET | Required | None |
| `/api/campaigns` | POST | Required | Name length, system ID |
| `/api/campaigns/{id}` | GET | Required + Owner | UUID format |
| `/api/campaigns/{id}` | PUT | Required + Owner | All fields |
| `/api/campaigns/{id}` | DELETE | Required + Owner | UUID format |

### Entity Endpoints

| Endpoint | Method | Auth | Input Validation |
|----------|--------|------|------------------|
| `/api/campaigns/{id}/entities` | GET | Required + Owner | UUID, filters |
| `/api/campaigns/{id}/entities` | POST | Required + Owner | All fields |
| `/api/entities/{id}` | GET | Required + Owner | UUID format |
| `/api/entities/{id}` | PUT | Required + Owner | All fields |
| `/api/entities/{id}` | DELETE | Required + Owner | UUID format |

**Critical**: Entity GET must filter GM notes for player access.

### Relationship Endpoints

| Endpoint | Method | Auth | Input Validation |
|----------|--------|------|------------------|
| `/api/campaigns/{id}/relationships` | GET | Required + Owner | UUID |
| `/api/campaigns/{id}/relationships` | POST | Required + Owner | Entity IDs exist |

### Session Endpoints

| Endpoint | Method | Auth | Input Validation |
|----------|--------|------|------------------|
| `/api/campaigns/{id}/sessions` | GET | Required + Owner | UUID |
| `/api/campaigns/{id}/sessions` | POST | Required + Owner | All fields |

**Critical**: Session prep notes are GM-only.

### Timeline Endpoints

| Endpoint | Method | Auth | Input Validation |
|----------|--------|------|------------------|
| `/api/campaigns/{id}/timeline` | GET | Required + Owner | UUID, date range |
| `/api/campaigns/{id}/timeline` | POST | Required + Owner | Date format |

### Import Endpoints

| Endpoint | Method | Auth | Input Validation |
|----------|--------|------|------------------|
| `/api/import/evernote` | POST | Required | File size, type |
| `/api/import/googledocs` | POST | Required | URL format |

## Input Validation Requirements

### UUID Fields

All ID fields must be valid UUIDs:

```go
func validateUUID(id string) error {
    _, err := uuid.Parse(id)
    return err
}
```

### String Fields

| Field | Max Length | Pattern | Required |
|-------|------------|---------|----------|
| Entity name | 200 | Printable | Yes |
| Description | 10000 | Any | No |
| GM notes | 50000 | Any | No |
| Tags | 50 each | Alphanumeric | No |

### Entity Type

Must be one of the valid entity types:

```go
var validTypes = []string{
    "npc", "location", "item", "faction", "clue",
    "creature", "organization", "event", "document", "other",
}
```

### Date Fields

Timeline dates must match expected formats:

```go
var dateFormats = []string{
    "2006-01-02",           // Exact
    "2006-01",              // Month precision
    "2006",                 // Year precision
}
```

### URL Fields (Google Docs Import)

Must be valid Google Docs URL:

```go
func validateGoogleDocsURL(url string) error {
    if !strings.HasPrefix(url, "https://docs.google.com/") {
        return errors.New("invalid Google Docs URL")
    }
    return nil
}
```

### File Upload (Evernote Import)

| Check | Limit |
|-------|-------|
| File size | 50MB max |
| File type | .enex only |
| XML depth | 100 levels max |

## Authorization Requirements

### Campaign Access

```go
func authorizeForCampaign(ctx context.Context, userID, campaignID string) error {
    campaign, err := db.GetCampaign(ctx, campaignID)
    if err != nil {
        return err
    }
    if campaign.OwnerID != userID {
        return ErrForbidden
    }
    return nil
}
```

### Entity Access

Entities inherit authorization from their campaign.

### Player vs GM

```go
func filterEntityForRole(entity *Entity, role string) *Entity {
    if role == "player" {
        entity.GMNotes = ""
        entity.DiscoveredSession = ""
    }
    return entity
}
```

## Rate Limiting

| Endpoint Category | Limit |
|-------------------|-------|
| Read operations | 100/minute |
| Write operations | 20/minute |
| Import operations | 5/minute |
| Authentication | 10/minute |

## Error Responses

### Safe Error Messages

```go
// DON'T leak internal details
return fmt.Errorf("database error: %v", err)

// DO return generic messages
return errors.New("an error occurred")
```

### Error Codes

| Code | Meaning | Details to Client |
|------|---------|-------------------|
| 400 | Bad Request | Validation errors |
| 401 | Unauthorized | "Authentication required" |
| 403 | Forbidden | "Access denied" |
| 404 | Not Found | "Resource not found" |
| 500 | Server Error | "An error occurred" |
