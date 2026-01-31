# Common Issues

This document catalogs frequently encountered issues during code review.

## Go Anti-Patterns

### Ignoring Errors

```go
// BAD - Silently ignoring error
result, _ := db.QueryRow(ctx, query, id)

// GOOD - Handle the error
result, err := db.QueryRow(ctx, query, id)
if err != nil {
    return fmt.Errorf("failed to fetch entity: %w", err)
}
```

### Empty Error Checks

```go
// BAD - Error checked but not handled
if err != nil {
    // nothing here
}

// GOOD - Proper error handling
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
```

### Not Using Context

```go
// BAD - No context propagation
func GetEntity(id string) (*Entity, error)

// GOOD - Context passed through
func GetEntity(ctx context.Context, id string) (*Entity, error)
```

### Mutex Copying

```go
// BAD - Mutex copied on receiver
type EntityCache struct {
    mu sync.Mutex
    data map[string]*Entity
}
func (c EntityCache) Get(id string) *Entity  // WRONG - copies mutex

// GOOD - Pointer receiver
func (c *EntityCache) Get(id string) *Entity
```

### Defer in Loop

```go
// BAD - Deferred close won't run until function ends
for _, file := range files {
    f, _ := os.Open(file)
    defer f.Close()  // Memory leak
}

// GOOD - Use closure or explicit close
for _, file := range files {
    func() {
        f, _ := os.Open(file)
        defer f.Close()
        // process file
    }()
}
```

## React Anti-Patterns

### Missing Hook Dependencies

```typescript
// BAD - Missing dependency
useEffect(() => {
    fetchEntity(entityId);
}, []);  // entityId missing!

// GOOD - All dependencies included
useEffect(() => {
    fetchEntity(entityId);
}, [entityId]);
```

### State Initialization in Render

```typescript
// BAD - Computed on every render
const Component = () => {
    const [data] = useState(expensiveComputation());  // Runs every render!
};

// GOOD - Lazy initialization
const Component = () => {
    const [data] = useState(() => expensiveComputation());
};
```

### Index as Key

```typescript
// BAD - Index as key
{entities.map((entity, i) => <EntityCard key={i} entity={entity} />)}

// GOOD - Stable unique key
{entities.map((entity) => <EntityCard key={entity.id} entity={entity} />)}
```

### Inline Object Props

```typescript
// BAD - New object on every render
<EntityList style={{ margin: 10 }} />

// GOOD - Stable reference
const listStyle = useMemo(() => ({ margin: 10 }), []);
<EntityList style={listStyle} />
```

### Missing useCallback

```typescript
// BAD - New function on every render, breaks memoization
const handleClick = () => updateEntity(entity.id);
<MemoizedButton onClick={handleClick} />

// GOOD - Stable function reference
const handleClick = useCallback(() => {
    updateEntity(entity.id);
}, [entity.id]);
```

## Database Issues

### N+1 Queries

```go
// BAD - N+1 queries
entities, _ := GetAllEntities(campaignID)
for _, e := range entities {
    relationships, _ := GetRelationships(e.ID)  // N additional queries!
}

// GOOD - Single query with join
entities, _ := GetEntitiesWithRelationships(campaignID)
```

### Missing Indexes

```sql
-- BAD - Frequently queried column without index
SELECT * FROM entities WHERE campaign_id = $1;

-- GOOD - Add index
CREATE INDEX idx_entities_campaign_id ON entities(campaign_id);
```

### JSONB Full Scans

```sql
-- BAD - Scans all JSONB values
SELECT * FROM entities WHERE attributes->>'occupation' = 'Detective';

-- GOOD - Use GIN index
CREATE INDEX idx_entities_attributes ON entities USING GIN(attributes);
```

## Imagineer-Specific Issues

### Not Checking for Duplicate Names

```go
// BAD - Creates duplicate entity
func CreateEntity(ctx context.Context, entity Entity) error {
    return db.Insert(entity)
}

// GOOD - Check for similar names first
func CreateEntity(ctx context.Context, entity Entity) error {
    existing, err := FindSimilarEntities(ctx, entity.CampaignID, entity.Name)
    if len(existing) > 0 {
        return fmt.Errorf("similar entity already exists: %s", existing[0].Name)
    }
    return db.Insert(entity)
}
```

### Exposing GM Notes to Players

```go
// BAD - Returns all fields including GM notes
func GetEntityForPlayer(id string) (*Entity, error) {
    return db.GetEntity(id)
}

// GOOD - Filter sensitive fields
func GetEntityForPlayer(id string) (*Entity, error) {
    entity, err := db.GetEntity(id)
    if err != nil {
        return nil, err
    }
    entity.GMNotes = ""  // Strip GM-only content
    return entity, nil
}
```

### Not Validating Game System

```go
// BAD - Assumes attributes are valid
func UpdateEntity(entity Entity) error {
    return db.Update(entity)
}

// GOOD - Validate against game system schema
func UpdateEntity(entity Entity) error {
    system, _ := GetGameSystem(entity.CampaignID)
    if err := ValidateAttributes(entity.Attributes, system); err != nil {
        return fmt.Errorf("invalid attributes: %w", err)
    }
    return db.Update(entity)
}
```

### Auto-Resolving Canon Conflicts

```go
// BAD - Automatically picking one version
func ImportEntity(entity Entity) error {
    existing, _ := FindByName(entity.Name)
    if existing != nil {
        return db.Update(entity)  // Silently overwrites!
    }
    return db.Insert(entity)
}

// GOOD - Create conflict for human review
func ImportEntity(entity Entity) error {
    existing, _ := FindByName(entity.Name)
    if existing != nil {
        return CreateConflict(existing, entity)  // Human decides
    }
    return db.Insert(entity)
}
```
