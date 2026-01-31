# Imagineer REST API Documentation

## 1. Overview

- **Base URL:** `/api`
- **Content-Type:** `application/json`
- **ID Format:** All IDs are UUIDs (e.g., `550e8400-e29b-41d4-a716-446655440000`)

## 2. Authentication

Currently no authentication is required (development mode).

> **Note:** Authentication will be added in a future release.

## 3. Endpoints

### Health Check

Check if the API is running.

```
GET /health
```

**Response:**
```json
{
  "status": "ok"
}
```

---

### Game Systems

#### List Game Systems

Returns all available game systems (Call of Cthulhu 7e, GURPS 4e, Blades in the Dark, etc.).

```
GET /api/game-systems
```

**Response:**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Call of Cthulhu 7e",
    "slug": "coc7e",
    "description": "Lovecraftian horror investigation",
    "createdAt": "2024-01-15T10:00:00Z"
  }
]
```

---

### Campaigns

#### List All Campaigns

```
GET /api/campaigns
```

**Response:**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "name": "Masks of Nyarlathotep",
    "systemId": "550e8400-e29b-41d4-a716-446655440000",
    "description": "A globe-trotting campaign of cosmic horror",
    "settings": {},
    "createdAt": "2024-01-15T10:00:00Z",
    "updatedAt": "2024-01-15T10:00:00Z"
  }
]
```

#### Create Campaign

```
POST /api/campaigns
```

**Request Body:**
```json
{
  "name": "Masks of Nyarlathotep",
  "systemId": "550e8400-e29b-41d4-a716-446655440000",
  "description": "A globe-trotting campaign of cosmic horror",
  "settings": {
    "startYear": 1925,
    "startLocation": "New York"
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | Yes | Campaign name |
| systemId | UUID | Yes | Game system ID |
| description | string | No | Campaign description |
| settings | object | No | System-specific settings |

**Response:** `201 Created` with the created campaign object.

#### Get Campaign

```
GET /api/campaigns/:id
```

**Response:** Campaign object.

#### Update Campaign

```
PUT /api/campaigns/:id
```

**Request Body:** Same fields as create (all optional for update).

**Response:** Updated campaign object.

#### Delete Campaign

```
DELETE /api/campaigns/:id
```

**Response:** `204 No Content`

---

### Entities

Entities include NPCs, locations, items, factions, events, and clues.

#### List Campaign Entities

```
GET /api/campaigns/:id/entities
```

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| type | string | Filter by entity type (e.g., `npc`, `location`) |

**Example:**
```
GET /api/campaigns/550e8400.../entities?type=npc
```

**Response:**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440002",
    "campaignId": "550e8400-e29b-41d4-a716-446655440001",
    "entityType": "npc",
    "name": "Jackson Elias",
    "description": "Author and adventurer investigating dark cults",
    "attributes": {
      "occupation": "Author",
      "age": 38
    },
    "tags": ["ally", "author", "investigator"],
    "keeperNotes": "Will be murdered in Chapter 1",
    "createdAt": "2024-01-15T10:00:00Z",
    "updatedAt": "2024-01-15T10:00:00Z"
  }
]
```

#### Create Entity

```
POST /api/campaigns/:id/entities
```

**Request Body:**
```json
{
  "entityType": "npc",
  "name": "Jackson Elias",
  "description": "Author and adventurer investigating dark cults",
  "attributes": {
    "occupation": "Author",
    "age": 38
  },
  "tags": ["ally", "author"],
  "keeperNotes": "Will be murdered in Chapter 1"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| entityType | EntityType | Yes | Type of entity |
| name | string | Yes | Entity name |
| description | string | No | Public description |
| attributes | object | No | System-specific attributes |
| tags | string[] | No | Searchable tags |
| keeperNotes | string | No | GM-only notes |

**Response:** `201 Created` with the created entity object.

#### Get Entity

```
GET /api/entities/:id
```

**Response:** Entity object.

#### Update Entity

```
PUT /api/entities/:id
```

**Request Body:** Same fields as create (all optional for update).

**Response:** Updated entity object.

#### Delete Entity

```
DELETE /api/entities/:id
```

**Response:** `204 No Content`

---

### Relationships

Define connections between entities (e.g., NPC knows NPC, item located at location).

#### List Campaign Relationships

```
GET /api/campaigns/:id/relationships
```

**Response:**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440003",
    "campaignId": "550e8400-e29b-41d4-a716-446655440001",
    "sourceEntityId": "550e8400-e29b-41d4-a716-446655440002",
    "targetEntityId": "550e8400-e29b-41d4-a716-446655440004",
    "relationshipType": "knows",
    "tone": "friendly",
    "description": "Old friends from university",
    "bidirectional": true,
    "strength": 0.8,
    "createdAt": "2024-01-15T10:00:00Z"
  }
]
```

#### Create Relationship

```
POST /api/campaigns/:id/relationships
```

**Request Body:**
```json
{
  "sourceEntityId": "550e8400-e29b-41d4-a716-446655440002",
  "targetEntityId": "550e8400-e29b-41d4-a716-446655440004",
  "relationshipType": "knows",
  "tone": "friendly",
  "description": "Old friends from university",
  "bidirectional": true,
  "strength": 0.8
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| sourceEntityId | UUID | Yes | Source entity ID |
| targetEntityId | UUID | Yes | Target entity ID |
| relationshipType | string | Yes | Type of relationship |
| tone | RelationshipTone | No | Emotional tone |
| description | string | No | Relationship description |
| bidirectional | boolean | No | If true, relationship goes both ways |
| strength | number | No | Relationship strength (0.0 - 1.0) |

**Response:** `201 Created` with the created relationship object.

---

### Timeline

Track events in your campaign's chronology.

#### List Timeline Events

```
GET /api/campaigns/:id/timeline
```

**Response:**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440005",
    "campaignId": "550e8400-e29b-41d4-a716-446655440001",
    "description": "Jackson Elias arrives in New York",
    "eventDate": "1925-01-15",
    "datePrecision": "day",
    "entityIds": ["550e8400-e29b-41d4-a716-446655440002"],
    "isPlayerKnown": true,
    "createdAt": "2024-01-15T10:00:00Z"
  }
]
```

#### Create Timeline Event

```
POST /api/campaigns/:id/timeline
```

**Request Body:**
```json
{
  "description": "Jackson Elias arrives in New York",
  "eventDate": "1925-01-15",
  "datePrecision": "day",
  "entityIds": ["550e8400-e29b-41d4-a716-446655440002"],
  "isPlayerKnown": true
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| description | string | Yes | Event description |
| eventDate | string | No | Date of event (ISO 8601) |
| datePrecision | DatePrecision | Yes | How precise the date is |
| entityIds | UUID[] | No | Related entity IDs |
| isPlayerKnown | boolean | No | Visible to players |

**Response:** `201 Created` with the created timeline event object.

---

### Statistics

#### Dashboard Statistics

Get aggregate statistics for the dashboard.

```
GET /api/stats/dashboard
```

**Response:**
```json
{
  "campaignCount": 3,
  "npcCount": 47,
  "locationCount": 23,
  "timelineEventCount": 156,
  "itemCount": 31,
  "factionCount": 8,
  "totalEntityCount": 109
}
```

---

### Import

Import content from external sources.

#### Import from Evernote

```
POST /api/import/evernote
```

**Content-Type:** `multipart/form-data`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| file | file | Yes | Evernote export file (.enex) |

**Example (curl):**
```bash
curl -X POST http://localhost:8080/api/import/evernote \
  -F "file=@campaign-notes.enex"
```

**Response:**
```json
{
  "imported": 15,
  "skipped": 2,
  "errors": []
}
```

#### Import from Google Docs

```
POST /api/import/google-docs
```

**Request Body (URL):**
```json
{
  "url": "https://docs.google.com/document/d/..."
}
```

**Request Body (Content):**
```json
{
  "content": "# Session Notes\n\nThe investigators arrived at..."
}
```

**Response:**
```json
{
  "imported": 5,
  "entities": ["550e8400-e29b-41d4-a716-446655440006"]
}
```

---

## 4. Error Responses

All errors return a JSON object with an `error` field:

```json
{
  "error": "Error message here"
}
```

### HTTP Status Codes

| Code | Meaning | Description |
|------|---------|-------------|
| 400 | Bad Request | Invalid input, missing required fields, or malformed JSON |
| 404 | Not Found | Resource does not exist |
| 500 | Internal Server Error | Server-side error |

### Example Error Responses

**400 Bad Request:**
```json
{
  "error": "name is required"
}
```

**404 Not Found:**
```json
{
  "error": "campaign not found"
}
```

---

## 5. Data Types

### EntityType

Type of campaign entity.

| Value | Description |
|-------|-------------|
| `npc` | Non-player character |
| `location` | Place or area |
| `item` | Object, artifact, or document |
| `faction` | Organization or group |
| `event` | Historical or planned event |
| `clue` | Investigative clue or evidence |

### SourceConfidence

Confidence level for imported or inferred data.

| Value | Description |
|-------|-------------|
| `authoritative` | Canon, confirmed by GM |
| `high` | Very likely accurate |
| `medium` | Probably accurate |
| `low` | Uncertain, needs verification |
| `speculative` | Guess or suggestion |

### RelationshipTone

Emotional quality of a relationship.

| Value | Description |
|-------|-------------|
| `friendly` | Positive relationship |
| `neutral` | No strong feelings |
| `hostile` | Negative relationship |
| `fearful` | One party fears the other |
| `romantic` | Romantic involvement |
| `professional` | Business or work relationship |

### DatePrecision

How precisely a date is known.

| Value | Description |
|-------|-------------|
| `exact` | Exact date and time known |
| `day` | Known to the day |
| `week` | Known to the week |
| `month` | Known to the month |
| `year` | Known to the year |
| `decade` | Known to the decade |
| `approximate` | Rough estimate |
| `unknown` | Date not known |
