# Documentation Style Guide

Complete style requirements for Imagineer documentation.

## Writing Style

### Voice and Tone

Use active voice throughout.

```markdown
<!-- BAD - Passive voice -->
The entity is created by the server.

<!-- GOOD - Active voice -->
The server creates the entity.
```

### Sentence Structure

Use full, grammatically correct sentences.

```markdown
<!-- BAD - Fragment -->
Creates entity.

<!-- GOOD - Complete sentence -->
The system creates a new entity with the specified attributes.
```

### Articles

Use articles (a, an, the) when appropriate.

```markdown
<!-- BAD -->
Server creates connection.

<!-- GOOD -->
The server creates a connection.
```

### Pronoun Clarity

Do not refer to an object as "it" unless the object is in the same sentence.

```markdown
<!-- BAD -->
The server validates the entity. It returns an error if invalid.

<!-- GOOD -->
The server validates the entity. The server returns an error if the
entity is invalid.
```

### Emojis

Do not use emojis unless explicitly requested.

## Document Structure

### Headings

- One first-level heading (`#`) per file
- Multiple second-level headings (`##`)
- Third/fourth level sparingly

### Introductions

Each heading should have an introductory sentence or paragraph.

### Line Wrapping

Wrap all markdown files at 79 characters or less.

## Lists

### Blank Lines

Always leave a blank line before the first item in any list.

```markdown
<!-- BAD -->
The system supports:
- Feature one
- Feature two

<!-- GOOD -->
The system supports:

- Feature one
- Feature two
```

### List Items

Each entry should be a complete sentence with articles.

```markdown
<!-- BAD -->
- Authentication
- Database connections

<!-- GOOD -->
- The system provides token-based authentication.
- Users can manage multiple database connections.
```

### Bold in Lists

Do not use bold font for bullet items.

### Numbered Lists

Only use numbered lists when steps must be performed in order.

## Code Snippets

### Explanatory Text

Include an explanatory sentence before code.

```markdown
In the following example, the `CreateEntity` function validates input:

` ` `go
func CreateEntity(name string) error {
    if name == "" {
        return errors.New("name required")
    }
    return nil
}
` ` `
```

### Inline Code

Use backticks around single commands or code.

```markdown
Run `make test` to execute the test suite.
```

### Code Blocks

Use fenced code blocks with language tags.

```markdown
` ` `sql
SELECT * FROM entities WHERE campaign_id = $1;
` ` `
```

### SQL Formatting

Capitalize SQL keywords; use lowercase for identifiers.

```sql
SELECT id, name FROM entities WHERE status = 'active';
```

## CHANGELOG Format

```markdown
## [Unreleased]

### Changed

- Description of change

### Added

- Description of addition

### Fixed

- Description of fix
```

## Todo.md Format

```markdown
## In Progress

- [ ] Current task

## Backlog

- [ ] Future task

## Completed

- [x] Finished task
```
