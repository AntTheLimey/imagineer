# Development Guidelines

This document provides guidelines for maintaining design consistency.

## Design Compliance Guidelines

### Before Writing Code

1. Review relevant sections of design.md
2. Check component-responsibilities.md for ownership
3. Find similar patterns in existing code
4. Consider security implications

### During Implementation

1. Follow existing patterns
2. Add tests alongside code
3. Document public APIs
4. Validate inputs at boundaries

### After Implementation

1. Run full test suite
2. Check for linter warnings
3. Update documentation if needed
4. Update CHANGELOG.md

## Pattern Library

### Adding a New Entity Type

1. Add to EntityType enum in `/internal/importers/common/types.go`
2. Add to TypeScript union in `/client/src/types/index.ts`
3. Add detection logic to importers if auto-detectable
4. Update entity creation UI if user-selectable

```go
// In types.go
const (
    EntityTypeNPC      EntityType = "npc"
    // ...
    EntityTypeNewType  EntityType = "new_type"  // Add here
)
```

### Adding a New API Endpoint

1. Define handler in `/internal/api/`
2. Register route in server setup
3. Add TypeScript types in client
4. Add service function in client
5. Write tests for handler

```go
// In api/entities.go
func (h *Handler) HandleGetEntity(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // 1. Parse request
    id := chi.URLParam(r, "id")

    // 2. Validate authorization
    if !h.canAccess(ctx, id) {
        http.Error(w, "forbidden", http.StatusForbidden)
        return
    }

    // 3. Execute operation
    entity, err := h.db.GetEntity(ctx, id)
    if err != nil {
        http.Error(w, "not found", http.StatusNotFound)
        return
    }

    // 4. Return response
    json.NewEncoder(w).Encode(entity)
}
```

### Adding a New Importer

1. Create package in `/internal/importers/newformat/`
2. Implement Importer interface
3. Register in importer factory
4. Add UI tab in Import page
5. Write tests with sample files

```go
// In importers/newformat/parser.go
type NewFormatImporter struct{}

func (i *NewFormatImporter) Import(
    ctx context.Context,
    source io.Reader,
    options common.ImportOptions,
) (*common.ImportResult, error) {
    result := &common.ImportResult{
        Entities:      []common.ExtractedEntity{},
        Relationships: []common.ExtractedRelationship{},
        Events:        []common.ExtractedEvent{},
    }

    // Parse source
    // Extract entities
    // Return result

    return result, nil
}

func (i *NewFormatImporter) Name() string {
    return "New Format"
}

func (i *NewFormatImporter) SupportedFormats() []string {
    return []string{".newformat"}
}
```

### Adding a Database Migration

1. Create numbered file in `/migrations/`
2. Write up migration (and down if needed)
3. Test on copy of data
4. Run via `make migrate`

```sql
-- migrations/003_add_entity_aliases.sql

-- Add aliases column for alternative entity names
ALTER TABLE entities
ADD COLUMN aliases TEXT[] DEFAULT '{}';

-- Index for searching aliases
CREATE INDEX idx_entities_aliases ON entities USING GIN(aliases);

-- Comment for documentation
COMMENT ON COLUMN entities.aliases IS
    'Alternative names for the entity (nicknames, titles, etc.)';
```

### Adding a React Component

1. Create component in appropriate directory
2. Add TypeScript types
3. Follow MUI theming
4. Add tests

```typescript
// In components/EntityCard.tsx
import { Card, CardContent, Typography } from '@mui/material';
import { Entity } from '../types';

interface EntityCardProps {
    entity: Entity;
    onClick?: (id: string) => void;
}

export function EntityCard({ entity, onClick }: EntityCardProps) {
    return (
        <Card onClick={() => onClick?.(entity.id)}>
            <CardContent>
                <Typography variant="h6">{entity.name}</Typography>
                <Typography color="text.secondary">
                    {entity.entityType}
                </Typography>
            </CardContent>
        </Card>
    );
}
```

## Code Review Checklist

### General

- [ ] Follows existing patterns
- [ ] Tests included
- [ ] Documentation updated
- [ ] No linter warnings
- [ ] Changes are minimal and focused

### Security

- [ ] Input validated
- [ ] Authorization checked
- [ ] Keeper notes protected
- [ ] No injection vectors

### Imagineer-Specific

- [ ] Entity types validated
- [ ] Canon confidence handled
- [ ] Duplicate names checked
- [ ] Game system respected

## Common Anti-Patterns to Avoid

### Auto-Resolving Canon Conflicts

Never automatically choose between conflicting data.

```go
// BAD
if conflict {
    useNewerVersion()  // Silent data loss!
}

// GOOD
if conflict {
    return createConflict(old, new)  // Human decides
}
```

### Exposing Keeper Notes

Always filter keeper-only content for player views.

```go
// BAD
return entity  // Includes keeper notes!

// GOOD
return filterForPlayer(entity)  // Safe
```

### Skipping Name Similarity Check

Always check for similar names before creating entities.

```go
// BAD
db.Insert(entity)

// GOOD
similar := findSimilar(entity.Name)
if len(similar) > 0 {
    return handleDuplicate(similar)
}
db.Insert(entity)
```

### Hardcoding Game System Logic

Use schemas, not code, for game system specifics.

```go
// BAD
if system == "coc-7e" {
    validateCoC7e(entity)
}

// GOOD
schema := loadSchema(system)
validate(entity, schema)
```

## Error Handling Patterns

### API Errors

```go
// Log detailed error internally
log.Error("database error",
    "operation", "get_entity",
    "entity_id", id,
    "error", err,
)

// Return generic error to client
return &APIError{
    Code:    "ENTITY_NOT_FOUND",
    Message: "Entity not found",
}
```

### Validation Errors

```go
// Collect all validation errors
var errors []string
if entity.Name == "" {
    errors = append(errors, "name is required")
}
if !validEntityType(entity.EntityType) {
    errors = append(errors, "invalid entity type")
}
if len(errors) > 0 {
    return &ValidationError{Errors: errors}
}
```

### Import Errors

```go
// Partial success is valid for imports
result := &ImportResult{}
for _, item := range items {
    entity, err := parseItem(item)
    if err != nil {
        result.Warnings = append(result.Warnings,
            fmt.Sprintf("skipped item: %v", err))
        continue
    }
    result.Entities = append(result.Entities, entity)
}
return result, nil
```
