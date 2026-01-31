# Documentation Templates

Standard templates for Imagineer documentation.

## API Endpoint Documentation

```markdown
# Endpoint Name

Brief description of what this endpoint does.

## Request

` ` `
METHOD /api/path/{param}
` ` `

### Path Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| param | string | Description |

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| page | int | 1 | Page number |

### Request Body

` ` `json
{
    "field": "value"
}
` ` `

## Response

### Success (200)

` ` `json
{
    "data": {
        "id": "uuid",
        "name": "string"
    }
}
` ` `

### Error (404)

` ` `json
{
    "error": {
        "code": "NOT_FOUND",
        "message": "Entity not found"
    }
}
` ` `

## Example

` ` `bash
curl -X GET https://api.example.com/api/entities/123
` ` `
```

## Feature Documentation

```markdown
# Feature Name

Brief description of the feature and its purpose.

## Overview

The feature provides the following capabilities:

- First capability description.
- Second capability description.
- Third capability description.

## Usage

### Basic Usage

Explanation of basic usage.

` ` `go
// Example code
` ` `

### Advanced Usage

Explanation of advanced options.

## Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| option | type | default | Description |

## Related Features

- [Related Feature 1](link)
- [Related Feature 2](link)
```

## Component Documentation

```markdown
# ComponentName

Brief description of the component.

## Props

| Prop | Type | Required | Description |
|------|------|----------|-------------|
| name | string | Yes | Description |
| onClick | function | No | Description |

## Example

` ` `typescript
import { ComponentName } from './ComponentName';

function Example() {
    return <ComponentName name="value" />;
}
` ` `

## Styling

The component uses Material-UI theming.

## Accessibility

- Keyboard navigation supported.
- ARIA labels provided.
```

## Migration Documentation

```markdown
# Migration NNN: Description

## Purpose

Brief description of why this migration exists.

## Changes

### Tables Created

- `table_name` - Description of table

### Columns Added

- `table.column` - Description of column

### Indexes Added

- `idx_name` - Description of index

## Rollback

` ` `sql
-- Rollback SQL if applicable
` ` `

## Verification

` ` `sql
-- SQL to verify migration succeeded
SELECT COUNT(*) FROM new_table;
` ` `
```

## Knowledge Base Document

```markdown
# Topic Name

Brief introduction to the topic.

## Overview

Explanation of the topic and its importance.

## Details

### Subtopic One

Detailed information.

### Subtopic Two

More detailed information.

## Examples

### Good Example

` ` `go
// Good code example
` ` `

### Anti-Pattern

` ` `go
// What to avoid
` ` `

## Related Documents

- [Related Document](link)

## References

- [External Reference](url)

Last Updated: YYYY-MM-DD
```
