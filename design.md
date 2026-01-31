# Imagineer Design Document

## Overview
Imagineer is a system-agnostic TTRPG campaign intelligence platform.

## Architecture Decisions

### ADR-001: PostgreSQL with JSONB for Flexibility
- **Decision:** Use PostgreSQL with JSONB columns for system-specific attributes
- **Rationale:** Provides both relational integrity and schema flexibility
- **Consequences:** Need careful indexing strategy for JSONB queries

### ADR-002: pgEdge Postgres MCP Server for AI Integration
- **Decision:** Use pgEdge MCP server for Claude Desktop/Code integration
- **Rationale:** Go-based, extensible custom definitions, production-ready Docker support
- **Consequences:** Currently read-only; may need separate write API or wait for write mode

### ADR-003: Agent-Based Analysis Tools
- **Decision:** Implement analysis as composable agents rather than monolithic features
- **Rationale:** Flexibility, testability, can evolve independently
- **Consequences:** Need clear agent interface contracts

### ADR-004: RAG Architecture for AI Assistance
- **Decision:** Use Retrieval Augmented Generation with hybrid search (vector +
  BM25) for all AI features
- **Rationale:** AI assistance requires campaign context and rulebook knowledge
  to be useful. Pure LLM calls without retrieval produce generic/hallucinated
  content that contradicts established canon.
- **Consequences:** Need embedding pipeline, chunking strategy, and separate
  knowledge stores for campaign vs rulebook content

#### Two Knowledge Domains

**1. Campaign Content (Per-Campaign, Private)**

Content that gets vectorized per campaign:
- Entity descriptions, attributes, GM notes
- Session notes and scene content
- Timeline events and their descriptions
- Relationship descriptions
- Imported document content
- Canon conflicts and resolutions

Use cases:
- "Check Canon" - semantic search for contradictions
- "Scan for Entities" - find existing entities mentioned in text
- "Suggest Relationships" - find semantically related entities
- "Find Plot Holes" - search for unresolved threads
- Session recap generation with full context

**2. Rulebook Knowledgebase (Per-Game-System, Shared)**

Reference material shared across campaigns using the same game system:
- Core rulebooks (character creation, combat, skills, magic)
- Bestiaries and creature stats
- Equipment lists with stats and costs
- Published adventures and scenarios (as inspiration)
- Setting guides and lore

Use cases:
- NPC stat block generation following system rules
- Player character creation assistance
- Encounter balancing suggestions
- Rule lookups during session prep
- "How does [mechanic] work?" queries

#### Embedding Pipeline with pgedge_vectorizer

The pgedge_vectorizer extension provides automatic, trigger-based embedding:

```
┌─────────────────────────────────────────────────────────────┐
│ INSERT/UPDATE on vectorized table                           │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│ TRIGGER (AFTER INSERT/UPDATE)                               │
│ ├─ Detects content changes (skips if unchanged)             │
│ ├─ Deletes old chunks if content changed                    │
│ ├─ Chunks text (token_based, markdown, or hybrid)           │
│ └─ Queues chunks for async embedding                        │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│ BACKGROUND WORKERS                                          │
│ ├─ Poll queue for pending items                             │
│ ├─ Batch API calls (configurable batch_size)                │
│ ├─ Retry with exponential backoff on failure                │
│ └─ Update chunk tables with embeddings                      │
└─────────────────────────────────────────────────────────────┘
```

**Setup Example:**
```sql
-- Enable vectorization on entities table
SELECT pgedge_vectorizer.enable_vectorization(
    source_table := 'entities'::regclass,
    source_column := 'description'::name,
    chunk_strategy := 'hybrid',      -- Best for mixed content
    chunk_size := 400,               -- Tokens per chunk
    chunk_overlap := 50              -- ~12% overlap
);
-- Creates: entities_description_chunks table with HNSW index
-- Creates: auto-trigger on INSERT/UPDATE
-- Processes all existing rows
```

**Chunking Strategies:**

| Strategy | Best For | Behaviour |
|----------|----------|-----------|
| `token_based` | Plain text, fast | Fixed token windows with overlap |
| `markdown` | Structured docs | Preserves headings, code blocks |
| `hybrid` | RAG (recommended) | Two-pass: structure-aware + merge small chunks |

Hybrid strategy prepends heading context to chunks:
`[Context: # Chapter > ## Section] Chunk content here...`

**Embedding Models (configured via postgresql.conf):**
```ini
pgedge_vectorizer.provider = 'openai'          # openai|voyage|ollama
pgedge_vectorizer.model = 'text-embedding-3-small'
pgedge_vectorizer.api_key_file = '/path/to/.api-key'
pgedge_vectorizer.batch_size = 20              # Items per API call
pgedge_vectorizer.num_workers = 2              # Background workers
```

| Provider | Models | Dimensions | Notes |
|----------|--------|------------|-------|
| OpenAI | text-embedding-3-small | 1536 | Cost-effective, batch-capable |
| OpenAI | text-embedding-3-large | 3072 | Higher quality |
| Voyage | voyage-2 | 1024 | Optimised for retrieval |
| Ollama | nomic-embed-text | 768 | Local, no API costs |

**Query-Time Embedding:**
```sql
SELECT pgedge_vectorizer.generate_embedding('search query text');
-- Returns vector for similarity comparison
```

#### Hybrid Search Strategy

Combine vector similarity with BM25 lexical search using the auto-created
chunk tables:

```sql
-- Hybrid search on entity descriptions
WITH query_embedding AS (
    SELECT pgedge_vectorizer.generate_embedding($1) AS vec
),
vector_results AS (
    SELECT
        e.id,
        e.name,
        c.content,
        c.chunk_index,
        1 - (c.embedding <=> (SELECT vec FROM query_embedding)) AS vector_score
    FROM entities e
    JOIN entities_description_chunks c ON e.id = c.source_id
    WHERE e.campaign_id = $2
      AND c.embedding IS NOT NULL
    ORDER BY c.embedding <=> (SELECT vec FROM query_embedding)
    LIMIT 20
),
bm25_results AS (
    SELECT e.id, bm25_score(c.content, $1) AS lexical_score
    FROM entities e
    JOIN entities_description_chunks c ON e.id = c.source_id
    WHERE e.campaign_id = $2
    LIMIT 20
)
SELECT
    v.id,
    v.name,
    v.content,
    COALESCE(v.vector_score, 0) * 0.7 +
    COALESCE(b.lexical_score, 0) * 0.3 AS combined_score
FROM vector_results v
LEFT JOIN bm25_results b ON v.id = b.id
ORDER BY combined_score DESC
LIMIT 10;
```

Benefits of hybrid:
- Vector finds semantic matches ("tavern" matches "inn", "pub", "alehouse")
- BM25 ensures exact keyword matches aren't missed
- Handles proper nouns (character names) that embeddings struggle with

#### Rulebook Knowledgebase Schema

Rulebook content stored separately from campaign data:

```sql
CREATE TABLE rulebook_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_system_id UUID REFERENCES game_systems(id),
    title TEXT NOT NULL,                    -- "Call of Cthulhu 7e Keeper's Guide"
    source_type TEXT NOT NULL,              -- 'core_rules'|'bestiary'|'adventure'|'supplement'
    version TEXT,                           -- Edition/printing
    publisher TEXT,
    import_source TEXT,                     -- 'pdf'|'manual'|'api'
    total_pages INT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE rulebook_sections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id UUID REFERENCES rulebook_sources(id) ON DELETE CASCADE,
    parent_section_id UUID REFERENCES rulebook_sections(id),
    heading_level INT NOT NULL,             -- 1-6 (h1-h6)
    title TEXT NOT NULL,                    -- "Chapter 4: Combat"
    page_start INT,
    page_end INT,
    content TEXT NOT NULL,                  -- Full section text
    heading_path TEXT[],                    -- ['Core Rules', 'Combat', 'Initiative']
    sort_order INT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Enable vectorization with hybrid strategy for structure preservation
SELECT pgedge_vectorizer.enable_vectorization(
    'rulebook_sections'::regclass,
    'content'::name,
    chunk_strategy := 'hybrid',
    chunk_size := 400,
    chunk_overlap := 50
);
-- Creates: rulebook_sections_content_chunks with heading context
```

**Querying Rulebooks by Game System:**
```sql
-- Find relevant rules for NPC creation in Call of Cthulhu
SELECT
    rs.title AS source,
    sec.title AS section,
    c.content,
    c.embedding <=> pgedge_vectorizer.generate_embedding(
        'creating NPC investigators characteristics skills'
    ) AS distance
FROM rulebook_sources rs
JOIN rulebook_sections sec ON rs.id = sec.source_id
JOIN rulebook_sections_content_chunks c ON sec.id = c.source_id
WHERE rs.game_system_id = $1  -- CoC 7e game system
  AND c.embedding IS NOT NULL
ORDER BY distance
LIMIT 10;
```

#### Retrieval Context for AI Calls

When invoking AI assistance, retrieve relevant context:

```
User Request: "Help me flesh out this NPC: Gerald the wine merchant"

Retrieved Context:
1. Campaign entities mentioning "merchant", "wine", "trade" (semantic)
2. Existing NPCs with similar roles (semantic)
3. Locations where merchants might appear (semantic)
4. Game system rules for NPC creation (rulebook)
5. Example merchant stat blocks (rulebook)

AI Prompt: [System instructions] + [Retrieved context] + [User request]
```

#### Privacy Considerations

- Campaign content stays private to campaign members
- Rulebook content can be shared (with proper licensing)
- Embedding API calls send content externally (note in privacy policy)
- Local embedding option (Ollama) for sensitive campaigns
- Never send GM notes to shared/public systems

## Data Model
See SCHEMAS.md for detailed entity schemas.

## Technology Stack
- **Language:** Go (standard project layout)
- **Database:** pgEdge Postgres 18 (Docker)
- **MCP Server:** pgEdge Postgres MCP (cloned to ~/PROJECTS/pgedge-postgres-mcp)
- **Configuration:** YAML files
- **Containerization:** Docker Compose

## User Experience Design

### Design Philosophy

Imagineer's UI is designed for GMs who spend extended periods preparing and
running sessions. The interface must support deep work without interruption,
protect against data loss, and provide AI assistance at every meaningful
decision point.

**Core Principles:**

1. **Full-Screen Workflows** - Editing is primary work, not secondary. Entity
   creation, session planning, and map building are full-screen experiences,
   not modal popups.

2. **Persistent Context** - The system maintains awareness of what you're
   working on and surfaces relevant information automatically.

3. **AI as Co-Pilot** - AI assistance is integrated contextually, not bolted
   on. The AI understands campaign canon and suggests based on established
   facts.

4. **Zero Data Loss** - Auto-save, explicit save states, and confirmation
   before navigation protect hours of creative work.

### Layout Architecture

#### Three-Panel Layout (Planning/Session Views)

Inspired by GM Assistant's dashboard design, the planning view uses a
three-panel layout optimised for session preparation:

```
┌─────────────────────────────────────────────────────────────────────┐
│  Header: Campaign Name | Session Title | AI Status | User Menu     │
├──────────┬──────────────────────────────────────────┬───────────────┤
│          │                                          │               │
│  LEFT    │           CENTRAL CANVAS                 │    RIGHT      │
│  PANEL   │                                          │    PANEL      │
│          │   Free-form content authoring area       │               │
│ Campaign │   - Session notes / planning             │  Contextual   │
│ Navigator│   - Scene/vignette editor                │  Entities     │
│          │   - Rich text with entity mentions       │               │
│ Sessions │                                          │  - NPCs       │
│ (chrono) │   AI scans content and populates         │  - Locations  │
│          │   right panel with detected entities     │  - Items      │
│ Quick    │                                          │  - Factions   │
│ Actions  │   [Scan for Entities] [Check Canon]      │               │
│          │   [Suggest Connections] [Find Holes]     │  Click to     │
│          │                                          │  expand/edit  │
│          │                                          │               │
└──────────┴──────────────────────────────────────────┴───────────────┘
```

**Left Panel - Navigation & Structure:**
- Campaign selector dropdown
- Chronological session list with visual indicators
- "New Session" and "Plan Next Session" actions
- Quick links to campaign entities, maps, timeline

**Central Canvas - Content Authoring:**
- Rich text editor with @ mentions for entity linking
- Outline mode with collapsible sections (scenes/vignettes)
- AI action buttons contextually placed
- "Show More/Less" for lengthy content

**Right Panel - Contextual Entities:**
- Auto-populated based on mentions in central canvas
- Grouped by type: NPCs, Locations, Items, Factions
- Expandable cards with key details
- Click to open full entity editor (in new view, not popup)
- Visual indicator for new vs existing entities
- "Create This Entity" action for AI-suggested new entities

#### Scenes / Vignettes / Encounters

A **Scene** is a discrete unit of narrative content with:

- **Title** - Short descriptive name ("Arrival at the Village")
- **Summary** - What happens in this scene
- **Involved Entities** - NPCs, locations, items present
- **GM Notes** - Private GM information
- **Memorable Moments** - Categorised highlights (Funny, Dramatic, Epic,
  Intriguing) added during or after play
- **Outline Points** - Structured beats within the scene

Scenes are:
- Reorderable within a session
- Linkable to timeline events
- Searchable by involved entities
- Usable as planning units for future sessions

### Full-Screen Editing

**Problem:** Popups and modals are fragile. A misclick, escape key, or
browser refresh can destroy work in progress.

**Solution:** All substantial editing happens in dedicated full-screen views:

| Content Type    | Edit View                                        |
|-----------------|--------------------------------------------------|
| Entity          | Full-page form with tabbed sections              |
| Session         | Three-panel planning view                        |
| Scene/Vignette  | Focused editor with entity sidebar               |
| Map             | Full-screen map canvas with toolbars             |
| Timeline Event  | Dedicated view with entity linking               |
| Campaign        | Settings page with sections                      |

**Navigation Patterns:**

- Editing views have a clear "back" path (breadcrumb or explicit button)
- Navigation away from unsaved changes triggers confirmation
- Auto-save drafts every 30 seconds to local storage
- Explicit "Save" and "Save & Close" actions
- "Discard Changes" requires confirmation

### Data Loss Prevention

Multiple layers protect user work:

1. **Auto-Draft** - Every 30 seconds, current state saves to browser local
   storage. On return, offer to restore draft.

2. **Dirty State Tracking** - Track unsaved changes. Show indicator in UI.
   Warn before navigation.

3. **Explicit Saves** - Server-side saves are explicit actions. Success/error
   feedback is clear and immediate.

4. **Undo/Redo** - In-editor undo stack for text content.

5. **Version History** - For entities with `version` field, allow viewing
   previous versions.

6. **Session Storage Backup** - On browser close with unsaved changes, attempt
   to persist to session storage with recovery on next visit.

### AI Integration Points

AI assistance appears contextually throughout the application:

#### Session Planning View

- **Scan for Entities** - Analyse text, identify mentioned entities, populate
  right panel. Distinguish existing entities from suggested new ones.
- **Check Canon** - Review content for contradictions with established campaign
  facts. Surface conflicts with source references.
- **Suggest Connections** - Based on mentioned entities, suggest relationships
  that could exist between them.
- **Find Plot Holes** - Identify logical inconsistencies, unresolved threads,
  or continuity issues.
- **Generate Summary** - Create Full/Short/Stylized recaps of session content.

#### Entity Editor

- **Flesh Out Details** - Expand sparse description based on entity type and
  existing relationships.
- **Suggest Relationships** - Recommend likely connections to other entities.
- **Check Canon** - Verify this entity doesn't conflict with established facts.
- **Generate Portrait** - AI image generation for NPC portraits (future).

#### Import Processing

- **Extract Entities** - Parse imported content to identify entities.
- **Deduplicate** - Find potential duplicates using name similarity.
- **Merge Suggestions** - When duplicate found, suggest how to merge new
  information with existing.

#### Search & Discovery

- **Semantic Search** - Find entities by meaning, not just keywords.
- **Relationship Exploration** - "Show me everyone connected to this NPC."
- **Timeline Analysis** - "What happened before this event?"

### AI Interaction Patterns

AI actions follow consistent patterns:

1. **Request** - User clicks AI action button
2. **Processing** - Loading indicator shows AI is working
3. **Preview** - AI suggestions shown in review state (not auto-applied)
4. **Approval** - User accepts, modifies, or rejects suggestions
5. **Application** - Accepted changes applied to content

AI never auto-modifies content. All AI suggestions require explicit user
approval before becoming part of the campaign data.

### Memorable Moments

Based on GM Assistant's design, sessions can have categorised highlights:

| Category    | Icon  | Description                              |
|-------------|-------|------------------------------------------|
| Funny       | 😄    | Humorous moments, jokes, mishaps         |
| Dramatic    | 🎭    | High tension, emotional beats            |
| Intriguing  | 🗿    | Mystery, discovery, plot reveals         |
| Epic        | 🐉    | Heroic actions, major achievements       |

Memorable Moments are:
- Added during session or in post-session recap
- Attributed to specific characters/players
- Linked to scenes where they occurred
- Searchable and browsable across campaign history

### Mobile Considerations

While desktop is primary, the UI should gracefully adapt:

- **Mobile:** Single-panel view with panel switching
- **Tablet:** Two-panel view (nav + content, or content + entities)
- **Desktop:** Full three-panel layout

Planning and editing workflows are desktop-focused. Mobile provides read
access and simple edits.

### Visual Design Notes

- **Dark Theme Primary** - GMs often prep at night. Dark theme reduces eye
  strain. Light theme available.
- **Information Density** - Show meaningful data, not empty space. GMs want
  information at a glance.
- **Consistent Iconography** - Entity types, scene categories, and actions
  have consistent visual language.
- **Color Coding** - Entity types have associated colors for quick scanning.
- **Typography** - Readable fonts, clear hierarchy, generous line height for
  extended reading.
