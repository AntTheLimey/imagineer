# Go Code Conventions

Coding standards for Imagineer Go code.

## Formatting

- Run `gofmt` on all files
- Four-space indentation (tabs converted to spaces)
- No trailing whitespace

## Naming

### Packages

```go
package api       // lowercase, short
package models    // singular noun
package importers // what it does
```

### Functions

```go
func CreateEntity()     // Exported: PascalCase
func validateInput()    // Unexported: camelCase
```

### Variables

```go
var MaxRetries = 3      // Exported constant
var defaultTimeout = 30 // Unexported
entityID := "..."       // Local: camelCase
```

### Interfaces

```go
type Importer interface {}  // Often -er suffix
type EntityStore interface {} // Describes capability
```

## Error Handling

### Always Check Errors

```go
result, err := someFunction()
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
```

### Wrap with Context

```go
if err != nil {
    return fmt.Errorf("failed to create entity %s: %w", name, err)
}
```

### Use Sentinel Errors

```go
var ErrEntityNotFound = errors.New("entity not found")

if errors.Is(err, ErrEntityNotFound) {
    // Handle not found
}
```

## Context Usage

### Accept Context First

```go
func GetEntity(ctx context.Context, id string) (*Entity, error)
```

### Pass Context Through

```go
func (h *Handler) HandleGet(ctx context.Context, id string) {
    entity, err := h.db.GetEntity(ctx, id)
}
```

## Struct Organization

```go
type Entity struct {
    // IDs first
    ID         string
    CampaignID string

    // Core fields
    Name        string
    EntityType  string
    Description string

    // Optional fields
    GMNotes *string
    Tags        []string

    // Metadata
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

## Interface Design

### Small Interfaces

```go
// Good: Small, focused
type EntityReader interface {
    GetEntity(ctx context.Context, id string) (*Entity, error)
}

// Avoid: Large kitchen-sink interfaces
```

### Accept Interfaces, Return Structs

```go
func NewService(db Database) *Service  // Accept interface
func (s *Service) Get() *Entity        // Return concrete
```

## Testing

### Table-Driven Tests

```go
func TestValidateEntity(t *testing.T) {
    tests := []struct {
        name    string
        entity  Entity
        wantErr bool
    }{
        {"valid", Entity{Name: "Test"}, false},
        {"empty name", Entity{}, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEntity(tt.entity)
            if (err != nil) != tt.wantErr {
                t.Errorf("got err=%v, wantErr=%v", err, tt.wantErr)
            }
        })
    }
}
```

## Resource Cleanup

### Use Defer

```go
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()

    // Process file
}
```

### Handle Cleanup Errors

```go
defer func() {
    if err := f.Close(); err != nil {
        log.Warn("failed to close file", "err", err)
    }
}()
```

## Copyright Header

Every source file should have:

```go
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
