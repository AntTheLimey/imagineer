# Documentation Examples

Good examples and anti-patterns for Imagineer documentation.

## Good Examples

### Clear Introduction

```markdown
# Entity Management

Entities represent the people, places, and things in a campaign world.
The entity system provides tools for creating, organizing, and relating
these elements.
```

### Proper List Format

```markdown
The entity system supports these types:

- NPCs represent non-player characters in the game world.
- Locations define places where action takes place.
- Items include objects, artifacts, and equipment.
- Factions group entities with shared goals or allegiances.
```

### Good Code Documentation

```markdown
In the following example, the `CreateEntity` function validates the
entity name before inserting into the database:

` ` `go
func CreateEntity(ctx context.Context, entity Entity) error {
    if entity.Name == "" {
        return errors.New("entity name is required")
    }
    return db.Insert(ctx, entity)
}
` ` `
```

### Clear Step-by-Step

```markdown
## Creating a New Campaign

Follow these steps to create a new campaign:

1. Navigate to the Campaigns page.
2. Click the "New Campaign" button.
3. Enter the campaign name and description.
4. Select the game system from the dropdown.
5. Click "Create" to save the campaign.
```

## Anti-Patterns

### Vague Introduction

```markdown
<!-- BAD -->
# Entity Management

This is about entities.
```

```markdown
<!-- GOOD -->
# Entity Management

Entities represent the people, places, and things in a campaign world.
```

### Missing Blank Line Before List

```markdown
<!-- BAD -->
The system supports:
- Feature one
- Feature two

<!-- GOOD -->
The system supports:

- Feature one.
- Feature two.
```

### Incomplete Bullet Points

```markdown
<!-- BAD -->
- Authentication
- Relationships
- Timeline

<!-- GOOD -->
- The system provides token-based authentication.
- Users can create relationships between entities.
- The timeline view displays events chronologically.
```

### Bold in Lists

```markdown
<!-- BAD -->
- **Authentication** - handles login
- **Database** - stores data

<!-- GOOD -->
- Authentication handles user login and session management.
- The database layer stores all campaign data.
```

### Passive Voice

```markdown
<!-- BAD -->
The entity is validated by the server. An error is returned if invalid.

<!-- GOOD -->
The server validates the entity. The server returns an error if the
entity is invalid.
```

### Ambiguous Pronouns

```markdown
<!-- BAD -->
The importer processes the file. It extracts entities from it. It then
stores them in the database.

<!-- GOOD -->
The importer processes the file. The parser extracts entities from the
content. The system stores the extracted entities in the database.
```

### Long Lines

```markdown
<!-- BAD -->
The entity management system provides comprehensive tools for creating, organizing, updating, and relating the various people, places, things, and concepts that exist within a campaign world.

<!-- GOOD -->
The entity management system provides comprehensive tools for creating,
organizing, updating, and relating the various people, places, things,
and concepts that exist within a campaign world.
```

### Code Without Explanation

```markdown
<!-- BAD -->
` ` `go
func ValidateEntity(e Entity) error {
    if e.Name == "" {
        return ErrNameRequired
    }
    return nil
}
` ` `

<!-- GOOD -->
In the following example, the `ValidateEntity` function checks that the
entity has a valid name:

` ` `go
func ValidateEntity(e Entity) error {
    if e.Name == "" {
        return ErrNameRequired
    }
    return nil
}
` ` `
```

### Missing Language Tag

```markdown
<!-- BAD -->
` ` `
SELECT * FROM entities;
` ` `

<!-- GOOD -->
` ` `sql
SELECT * FROM entities;
` ` `
```

## Changelog Examples

### Good Entry

```markdown
### Changed

- CLAUDE.md Task Workflow now includes step 6 "Document" for automatic
  CHANGELOG.md and Todo.md updates after completing user-facing changes
```

### Bad Entry

```markdown
<!-- BAD - Too vague -->
### Changed

- Updated docs
```

### Good Feature Entry

```markdown
### Added

#### Content Importers

- Evernote importer (internal/importers/evernote)
  - Parses .enex XML export files
  - Extracts notes as entities with tags
  - Auto-detects entity types from content
```

### Bad Feature Entry

```markdown
<!-- BAD - Missing details -->
### Added

- Evernote import
```
