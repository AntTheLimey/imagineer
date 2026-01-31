# Data Flow

This document describes how data flows through the Imagineer system.

## Overview

```
[External Sources] → [Importers] → [Database] → [API] → [Client]
       ↑                               ↑
   Evernote                         Agents
   Google Docs                 (consistency checker)
```

## Import Flow

### Evernote Import

```
1. User uploads .enex file via Import page
2. Client sends file to POST /api/import/evernote
3. API handler passes to EvernoteImporter.Import()
4. Parser extracts:
   - Note content → Entity descriptions
   - Note tags → Entity tags
   - Note title → Entity name
   - Auto-detects entity type from content
5. Returns ImportResult with entities, relationships, events
6. API stores entities in database
7. Client displays import summary
```

### Google Docs Import

```
1. User provides document URL via Import page
2. Client sends URL to POST /api/import/googledocs
3. API handler calls GoogleDocsImporter.Import()
4. Importer fetches document via export URL
5. Parser extracts:
   - Headers → Section names
   - "Key: Value" patterns → Attributes
   - Relationship patterns → Relationships
   - Date patterns → Timeline events
6. Returns ImportResult
7. API stores extracted data
```

## Entity Data Flow

### Creating an Entity

```
1. User fills entity form in Entities page
2. Client validates input
3. POST /api/campaigns/{id}/entities
4. API handler:
   a. Validates against campaign
   b. Checks for duplicate names (Levenshtein distance)
   c. Sets source_confidence to DRAFT
   d. Inserts into database
5. Returns created entity
6. Client updates UI
```

### Entity Retrieval

```
1. Client requests GET /api/campaigns/{id}/entities
2. API handler:
   a. Validates campaign access
   b. Queries entities with filters
   c. Joins with relationships if requested
3. Returns entity list with relationships
4. Client displays in Entities page
```

## Relationship Data Flow

### Creating a Relationship

```
1. User links two entities
2. POST /api/campaigns/{id}/relationships
3. API handler:
   a. Validates both entities exist
   b. Validates both in same campaign
   c. Inserts relationship
4. Returns created relationship
5. Client updates relationship visualization
```

## Timeline Data Flow

### Timeline Event Creation

```
1. User adds event via Timeline page
2. POST /api/campaigns/{id}/timeline
3. API handler:
   a. Validates date precision
   b. Links to related entities
   c. Sets is_player_known flag
4. Returns created event
5. Client adds to timeline view
```

### Timeline Retrieval

```
1. Client requests GET /api/campaigns/{id}/timeline
2. API handler:
   a. Queries events ordered by date
   b. Joins with related entities
   c. Filters by player visibility if requested
3. Returns event list
4. Client renders timeline
```

## Session Data Flow

### Session Planning

```
1. GM creates session via Sessions page
2. POST /api/campaigns/{id}/sessions
3. Session created with status=PLANNED
4. GM adds prep notes, planned scenes
5. Session stored with preparation data
```

### Session Completion

```
1. GM marks session as completed
2. PATCH /api/sessions/{id}
3. API handler:
   a. Updates status to COMPLETED
   b. Records actual date
   c. Links discoveries to entities
4. Triggers consistency check agent
5. Agent analyzes for canon conflicts
```

## Canon Conflict Detection

### Conflict Discovery

```
1. Consistency checker agent runs
2. Agent queries entities for conflicts:
   - Same entity, different descriptions
   - Contradicting relationships
   - Timeline inconsistencies
3. Creates canon_conflict records
4. Sets status to DETECTED
```

### Conflict Resolution

```
1. User reviews conflicts in UI
2. User selects authoritative version
3. PATCH /api/conflicts/{id}
4. API handler:
   a. Updates winning entity to AUTHORITATIVE
   b. Updates losing entity to SUPERSEDED
   c. Sets conflict status to RESOLVED
5. Stores resolution notes
```

## Database Query Patterns

### Entity Queries

```sql
-- Find entities by name (fuzzy match)
SELECT * FROM entities
WHERE campaign_id = $1
  AND name % $2  -- Trigram similarity
ORDER BY similarity(name, $2) DESC;

-- Find entities by tag
SELECT * FROM entities
WHERE campaign_id = $1
  AND tags @> ARRAY[$2];

-- Find entities by attribute
SELECT * FROM entities
WHERE campaign_id = $1
  AND attributes->>'occupation' = $2;
```

### Relationship Queries

```sql
-- Find all relationships for an entity
SELECT r.*, e.name as target_name
FROM relationships r
JOIN entities e ON r.target_entity_id = e.id
WHERE r.source_entity_id = $1;

-- Find bidirectional relationships
SELECT * FROM relationships
WHERE source_entity_id = $1
   OR (target_entity_id = $1 AND bidirectional = true);
```

## Client State Management

### React Query Patterns

```typescript
// Entity list query
const { data: entities } = useQuery({
  queryKey: ['entities', campaignId],
  queryFn: () => fetchEntities(campaignId)
});

// Entity mutation
const mutation = useMutation({
  mutationFn: createEntity,
  onSuccess: () => {
    queryClient.invalidateQueries(['entities', campaignId]);
  }
});
```

## API Response Formats

### Success Response

```json
{
  "data": { ... },
  "meta": {
    "total": 100,
    "page": 1,
    "per_page": 20
  }
}
```

### Error Response

```json
{
  "error": {
    "code": "ENTITY_NOT_FOUND",
    "message": "Entity with ID xyz not found"
  }
}
```
