# MCP Tools Catalog

Catalog of MCP tools for Imagineer.

## Entity Tools

### entity_search

Search for entities in a campaign.

**Input:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| campaign_id | string | Yes | Campaign UUID |
| query | string | No | Search query |
| entity_type | string | No | Filter by type |
| limit | number | No | Max results (default 20) |

**Output:**

```json
{
    "entities": [
        {
            "id": "uuid",
            "name": "Entity Name",
            "type": "npc",
            "description": "Brief description"
        }
    ],
    "total": 100
}
```

### entity_get

Get full entity details.

**Input:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| entity_id | string | Yes | Entity UUID |

**Output:**

```json
{
    "id": "uuid",
    "name": "Entity Name",
    "type": "npc",
    "description": "Full description",
    "attributes": {},
    "tags": [],
    "relationships": []
}
```

**Note:** GM notes filtered based on role.

### entity_create

Create a new entity.

**Input:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| campaign_id | string | Yes | Campaign UUID |
| name | string | Yes | Entity name |
| type | string | Yes | Entity type |
| description | string | No | Description |
| attributes | object | No | Type-specific attributes |

## Relationship Tools

### relationship_map

Get relationships for an entity.

**Input:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| entity_id | string | Yes | Entity UUID |
| depth | number | No | Traversal depth (default 1) |

**Output:**

```json
{
    "center": { "id": "uuid", "name": "Entity" },
    "relationships": [
        {
            "target": { "id": "uuid", "name": "Other" },
            "type": "knows",
            "tone": "friendly",
            "bidirectional": true
        }
    ]
}
```

### relationship_create

Create a relationship between entities.

**Input:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| source_id | string | Yes | Source entity UUID |
| target_id | string | Yes | Target entity UUID |
| type | string | Yes | Relationship type |
| tone | string | No | Relationship tone |
| bidirectional | boolean | No | Two-way relationship |

## Timeline Tools

### timeline_query

Query timeline events.

**Input:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| campaign_id | string | Yes | Campaign UUID |
| start_date | string | No | Start of range |
| end_date | string | No | End of range |
| entity_id | string | No | Filter by entity |

**Output:**

```json
{
    "events": [
        {
            "id": "uuid",
            "date": "1925-03-15",
            "precision": "exact",
            "description": "Event description",
            "entities": ["uuid1", "uuid2"]
        }
    ]
}
```

### timeline_add

Add a timeline event.

**Input:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| campaign_id | string | Yes | Campaign UUID |
| date | string | Yes | Event date |
| precision | string | No | Date precision |
| description | string | Yes | Event description |
| entity_ids | array | No | Related entities |

## Canon Tools

### canon_check

Check for canon conflicts.

**Input:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| campaign_id | string | Yes | Campaign UUID |
| entity_id | string | No | Check specific entity |

**Output:**

```json
{
    "conflicts": [
        {
            "id": "uuid",
            "entity_id": "uuid",
            "field": "description",
            "values": [
                { "value": "Version A", "source": "Source 1" },
                { "value": "Version B", "source": "Source 2" }
            ],
            "status": "DETECTED"
        }
    ]
}
```

### canon_resolve

Resolve a canon conflict.

**Input:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| conflict_id | string | Yes | Conflict UUID |
| resolution | string | Yes | Resolution notes |
| chosen_value | string | Yes | Authoritative value |

## Campaign Tools

### campaign_list

List user's campaigns.

**Output:**

```json
{
    "campaigns": [
        {
            "id": "uuid",
            "name": "Campaign Name",
            "system": "coc-7e"
        }
    ]
}
```

### campaign_get

Get campaign details.

**Input:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| campaign_id | string | Yes | Campaign UUID |

**Output:**

```json
{
    "id": "uuid",
    "name": "Campaign Name",
    "system": "coc-7e",
    "description": "Campaign description",
    "stats": {
        "entities": 150,
        "relationships": 89,
        "sessions": 12
    }
}
```
