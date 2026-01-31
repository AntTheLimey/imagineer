# Quality Standards

This document defines code quality standards for the Imagineer project.

## General Standards

### Formatting

| Rule | Standard |
|------|----------|
| Indentation | Four spaces (all languages) |
| Line length | 79 characters (docs), reasonable for code |
| Trailing whitespace | None |
| Final newline | Required |

### Naming Conventions

**Go:**

| Element | Convention | Example |
|---------|------------|---------|
| Package | lowercase, short | `models`, `api` |
| Exported function | PascalCase | `CreateEntity` |
| Unexported function | camelCase | `validateInput` |
| Constant | PascalCase or ALL_CAPS | `MaxRetries` |
| Interface | PascalCase, often -er | `Importer`, `EntityValidator` |

**TypeScript/React:**

| Element | Convention | Example |
|---------|------------|---------|
| Component | PascalCase | `EntityList` |
| Hook | camelCase, use prefix | `useEntities`, `useCampaign` |
| Function | camelCase | `fetchData` |
| Constant | SCREAMING_SNAKE_CASE | `MAX_RETRIES` |
| Type/Interface | PascalCase | `Entity`, `Campaign` |

### Documentation

**Required documentation:**

- All exported functions/methods
- All exported types
- Complex algorithms
- Non-obvious business logic
- Configuration options

**Documentation style:**

```go
// CreateEntity creates a new entity in the specified campaign.
// It validates the entity type against the campaign's game system
// and returns an error if the type is not supported.
func CreateEntity(ctx context.Context, campaignID string, entity Entity) (*Entity, error)
```

## Code Organization

### Function Length

| Length | Assessment |
|--------|------------|
| < 20 lines | Ideal |
| 20-50 lines | Acceptable |
| 50-100 lines | Consider splitting |
| > 100 lines | Must refactor |

### Cyclomatic Complexity

| Complexity | Assessment |
|------------|------------|
| 1-5 | Simple, easy to test |
| 6-10 | Moderate, acceptable |
| 11-20 | Complex, consider simplifying |
| > 20 | Too complex, must refactor |

### Nesting Depth

| Depth | Assessment |
|-------|------------|
| 1-2 | Ideal |
| 3 | Acceptable |
| 4 | Maximum recommended |
| > 4 | Refactor with early returns or extraction |

**Reducing nesting:**

```go
// BAD - Deep nesting
func process(entity *Entity) error {
    if entity != nil {
        if entity.Valid {
            if entity.CampaignID != "" {
                // actual logic here
            }
        }
    }
    return nil
}

// GOOD - Early returns
func process(entity *Entity) error {
    if entity == nil {
        return nil
    }
    if !entity.Valid {
        return nil
    }
    if entity.CampaignID == "" {
        return nil
    }
    // actual logic here
    return nil
}
```

## Error Handling

### Go Error Handling

**Always check errors:**

```go
// BAD
result, _ := someFunction()

// GOOD
result, err := someFunction()
if err != nil {
    return fmt.Errorf("someFunction failed: %w", err)
}
```

**Wrap errors with context:**

```go
// BAD - No context
if err != nil {
    return err
}

// GOOD - Context added
if err != nil {
    return fmt.Errorf("failed to create entity %s: %w", entity.Name, err)
}
```

### TypeScript Error Handling

**Handle promise rejections:**

```typescript
// BAD
const data = await fetchEntities();

// GOOD
try {
    const data = await fetchEntities();
} catch (error) {
    // Handle error appropriately
}
```

## Testing Standards

### Coverage Requirements

| Component Type | Minimum Coverage |
|----------------|------------------|
| Business logic | 90% |
| API handlers | 85% |
| Utilities | 80% |
| UI components | 70% |

### Test Naming

```go
// Go - Descriptive function names
func TestCreateEntity_WithValidInput_ReturnsEntity(t *testing.T)
func TestCreateEntity_WithDuplicateName_ReturnsError(t *testing.T)
```

```typescript
// TypeScript - describe/it blocks
describe('CreateEntity', () => {
    it('returns entity when input is valid', () => {})
    it('returns error when name is duplicate', () => {})
})
```

## Imagineer-Specific Standards

### Entity Type Validation

Always validate entity types against the supported list:

```go
var validEntityTypes = map[string]bool{
    "npc": true, "location": true, "item": true,
    "faction": true, "clue": true, "creature": true,
    "organization": true, "event": true, "document": true,
    "other": true,
}
```

### Canon Confidence Handling

When creating or updating entities, respect the source confidence rules:

- New entities default to `DRAFT`
- Only explicitly promote to `AUTHORITATIVE`
- Mark replaced data as `SUPERSEDED`

### Game System Schema Validation

Validate entity attributes against the game system schema:

```go
// Validate attributes against game system
func ValidateEntityAttributes(entity Entity, system GameSystem) error {
    // Check required attributes exist
    // Validate attribute types
    // Check skill ranges
}
```
