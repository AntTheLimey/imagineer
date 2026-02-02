# AI Memory System Design

This document describes the hierarchical AI memory system for Imagineer. The
memory system enables contextual AI assistance across campaign management by
maintaining three tiers of memory.

## Overview

The memory system provides persistent context for AI interactions during
campaign management. Rather than treating each conversation in isolation, the
system maintains memories that accumulate over time and surface relevant
context automatically.

The system uses three memory tiers that form a hierarchy:

- Campaign Memory stores long-term facts that persist across the entire
  campaign.
- Chapter Memory stores medium-term context for the current story arc.
- Session Memory stores short-term details for individual sessions.

Each tier inherits from the tier above, creating a layered context that the
AI uses to provide relevant, consistent assistance.

## Design Principles

The memory system follows five core principles.

### File-Based Transparency

GMs maintain full control over all memories. The system stores memories as
viewable, editable records that the GM can modify or delete at any time.
The GM is the ultimate authority on campaign canon.

### Memory Formation Over Summarization

The system extracts specific facts rather than compressing everything into
summaries. A memory might be "Duke Wilhelm is secretly a ghoul" rather than
a paragraph summarizing an entire session. This approach preserves important
details that generic summaries lose.

### Hierarchical Inheritance

Lower memory tiers inherit context from higher tiers. Session Memory
inherits from Chapter Memory, which inherits from Campaign Memory. The AI
receives the complete context stack when responding to queries.

### Full Log Preservation

The system stores complete chat histories permanently. Compression happens
only when assembling context for AI calls. The original conversation
remains intact for reference and re-processing.

### Entity Integration

Memory extraction feeds the entity and relationship graph. When the AI
detects a new NPC, location, or item in conversation, the system queues
that entity for GM review before adding to the campaign database.

## Memory Tiers

### Campaign Memory

Campaign Memory stores persistent facts that remain relevant across the
entire campaign. This tier captures foundational elements that rarely
change.

The system stores the following types of campaign memories:

- The campaign premise establishes the initial situation and goals.
- Setting details describe the world and its rules.
- Tone preferences guide the style of AI-generated content.
- Major factions represent the primary power groups.
- Themes track recurring narrative elements.
- Canon authority decisions record which sources take precedence.
- Long-term plot threads track overarching storylines.
- Entity relationships summarize the major connections in the campaign.

Campaign Memory uses a token budget of 2,000 to 5,000 tokens. The AI
context always includes Campaign Memory regardless of the query type.

### Chapter Memory

Chapter Memory stores context for the current story arc. This tier captures
information relevant to the active portion of the campaign.

The system stores the following types of chapter memories:

- Chapter goals define what the current arc aims to accomplish.
- Active objectives track specific tasks the characters pursue.
- Plot threads document storylines in progress.
- Session summaries provide condensed recaps of recent sessions.
- Decision consequences record major choices and their effects.
- Active NPCs list characters currently relevant to the arc.
- Active locations list places currently relevant to the arc.

Chapter Memory uses a token budget of 3,000 to 5,000 tokens. The AI context
always includes Chapter Memory alongside Campaign Memory.

### Session Memory

Session Memory stores details for individual sessions. This tier captures
granular information about specific play sessions.

The system stores the following types of session memories:

- The full chat log preserves the complete conversation history.
- Session summaries provide brief recaps of the session.
- Scene breakdowns describe individual scenes within the session.
- Decision points record choices the players made.
- Discoveries note new information the characters learned.
- Memorable moments capture highlights from play.
- Extracted entities list potential new entities pending review.

Session Memory uses a token budget of 5,000 to 10,000 tokens for recent
sessions. The AI context includes the current session and the previous two
to three sessions.

## Token Compression Strategy

The system uses Adaptive Focus Memory to manage token budgets while
preserving important context.

### Fidelity Levels

Each message receives one of three fidelity treatments based on recency
and importance.

| Level | Treatment | Application |
|-------|-----------|-------------|
| FULL | The system includes the message verbatim. | Recent messages and critical decisions receive full treatment. |
| COMPRESSED | The system summarizes the message via LLM. | Older context and routine exchanges receive compression. |
| PLACEHOLDER | The system replaces the message with a short stub. | Very old context becomes a placeholder that references extracted facts. |

### Compression Pipeline

The following diagram shows the compression pipeline that processes session
chat logs.

```
Full Chat Log (stored permanently)
         |
         v
+--------------------------------------+
| EXTRACTION PASS                      |
|                                      |
| The system extracts:                 |
| - Entity mentions                    |
| - Plot decisions                     |
| - Canon-relevant facts               |
| - Memorable moments                  |
| - Unresolved threads                 |
+--------------------------------------+
         |
         v
+--------------------------------------+
| SUMMARIZATION PASS                   |
|                                      |
| The system generates:                |
| - Scene-by-scene summaries           |
| - Session summary (2-3 paragraphs)   |
| - Timeline one-liner                 |
+--------------------------------------+
         |
         v
Session Memory (~500-1000 tokens)
```

The extraction pass identifies discrete facts and entities. The
summarization pass creates condensed versions of the narrative flow. The
combination produces a compact memory that preserves both facts and story
context.

## Database Schema

The following tables store the memory system data.

### Campaign Memories Table

The `campaign_memories` table stores long-term facts for a campaign.

```sql
CREATE TABLE campaign_memories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID REFERENCES campaigns(id) ON DELETE CASCADE,
    memory_type TEXT NOT NULL,
    title TEXT,
    content TEXT NOT NULL,
    embedding VECTOR(1536),
    source_session_id UUID REFERENCES sessions(id),
    importance INT DEFAULT 5,
    is_spoiler BOOLEAN DEFAULT false,
    gm_created BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON COLUMN campaign_memories.memory_type IS
    'Type of memory: premise, theme, faction, plot_thread';
COMMENT ON COLUMN campaign_memories.importance IS
    'Priority ranking from 1 (low) to 10 (high)';
COMMENT ON COLUMN campaign_memories.is_spoiler IS
    'True if memory contains information hidden from players';
COMMENT ON COLUMN campaign_memories.gm_created IS
    'True if GM manually created this memory';
```

### Chapter Memories Table

The `chapter_memories` table stores arc-level context for a campaign
chapter.

```sql
CREATE TABLE chapter_memories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID REFERENCES campaigns(id) ON DELETE CASCADE,
    chapter_id UUID REFERENCES chapters(id) ON DELETE CASCADE,
    goals TEXT[],
    active_threads TEXT[],
    summary TEXT,
    embedding VECTOR(1536),
    is_current BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON COLUMN chapter_memories.goals IS
    'Array of goal descriptions for the chapter';
COMMENT ON COLUMN chapter_memories.active_threads IS
    'Array of plot thread descriptions';
COMMENT ON COLUMN chapter_memories.is_current IS
    'True if this is the active chapter';
```

### Session Memories Table

The `session_memories` table stores session-level details.

```sql
CREATE TABLE session_memories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID REFERENCES sessions(id) ON DELETE CASCADE,
    memory_type TEXT NOT NULL,
    title TEXT,
    content TEXT NOT NULL,
    embedding VECTOR(1536),
    scene_index INT,
    entities_mentioned UUID[],
    importance INT DEFAULT 5,
    is_player_visible BOOLEAN DEFAULT false,
    gm_edited BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON COLUMN session_memories.memory_type IS
    'Type of memory: summary, scene, decision, discovery';
COMMENT ON COLUMN session_memories.scene_index IS
    'Position within session for scene-type memories';
COMMENT ON COLUMN session_memories.entities_mentioned IS
    'Array of entity UUIDs referenced in this memory';
COMMENT ON COLUMN session_memories.is_player_visible IS
    'True if players can see this memory';
COMMENT ON COLUMN session_memories.gm_edited IS
    'True if GM has edited the AI-generated content';
```

### Session Chat Logs Table

The `session_chat_logs` table stores complete conversation histories.

```sql
CREATE TABLE session_chat_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID REFERENCES sessions(id) ON DELETE CASCADE,
    message_index INT NOT NULL,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    tool_calls JSONB,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(session_id, message_index)
);

COMMENT ON COLUMN session_chat_logs.role IS
    'Message author: user, assistant, or system';
COMMENT ON COLUMN session_chat_logs.tool_calls IS
    'JSON array of tool calls made by the assistant';
```

### Entity Extraction Queue Table

The `memory_entity_extractions` table queues detected entities for GM
review.

```sql
CREATE TABLE memory_entity_extractions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_memory_id UUID REFERENCES session_memories(id),
    extracted_name TEXT NOT NULL,
    extracted_type TEXT,
    evidence_text TEXT,
    matched_entity_id UUID REFERENCES entities(id),
    match_confidence FLOAT,
    status TEXT DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON COLUMN memory_entity_extractions.extracted_type IS
    'Detected entity type: npc, location, item, faction';
COMMENT ON COLUMN memory_entity_extractions.evidence_text IS
    'Text excerpt where the entity was mentioned';
COMMENT ON COLUMN memory_entity_extractions.matched_entity_id IS
    'Existing entity if a match was found';
COMMENT ON COLUMN memory_entity_extractions.match_confidence IS
    'Confidence score for the entity match (0.0 to 1.0)';
COMMENT ON COLUMN memory_entity_extractions.status IS
    'Review status: pending, approved, or rejected';
```

## Context Assembly

The context assembly process builds the AI prompt from memory tiers and
retrieved content.

### Token Budget Allocation

The following table shows token allocation for a 200K context window.

| Component | Tokens | Purpose |
|-----------|--------|---------|
| System prompt | 2,000-4,000 | The system prompt defines core AI behavior. |
| Campaign Memory | 2,000-5,000 | Campaign Memory provides long-term context. |
| Chapter Memory | 3,000-5,000 | Chapter Memory provides current arc context. |
| Session Memory | 5,000-10,000 | Session Memory provides recent session details. |
| RAG context | 5,000-15,000 | RAG retrieval provides entities and rules. |
| Current conversation | 20,000-50,000 | The active chat requires substantial space. |
| Response reserve | 4,000-8,000 | The model needs space for output. |

### Retrieval Priority

The context assembly service retrieves content in the following priority
order.

1. Campaign premise and themes always appear first.
2. The current chapter summary always appears second.
3. Recent session summaries from the last three to five sessions appear
   third.
4. Query-relevant memories appear based on semantic search.
5. Mentioned entities appear based on name and semantic matching.
6. Relationship context appears based on graph traversal.

## GM Control Interface

The GM maintains full control over all memories through a dedicated
management interface.

### Memory Manager UI

The following mockup shows the memory manager interface.

```
+-----------------------------------------------------+
| CAMPAIGN MEMORY MANAGER                             |
+---------+-------------------------------------------+
| [Search...] [Filter by type v] [Date range v]       |
+---------+-------------------------------------------+
| CAMPAIGN FACTS                                      |
| +-- "The cult operates beneath the library" [Edit]  |
| +-- "Duke Wilhelm is secretly a ghoul" [Edit][Del]  |
| +-- [+ Add Campaign Fact]                           |
+---------+-------------------------------------------+
| CHAPTER: The Arkham Conspiracy                      |
| +-- Session 12 summary [View][Edit]                 |
| +-- [+ Add Chapter Memory]                          |
+---------+-------------------------------------------+
| SESSION 14 (Current)                                |
| +-- Full chat log [View][Download]                  |
| +-- AI Summary [Regenerate][Edit]                   |
| +-- Extracted Entities (5) [Review]                 |
| +-- Memorable Moments (2) [View][Add]               |
+---------+-------------------------------------------+
| [Export All] [Clear Session] [Settings]             |
+-----------------------------------------------------+
```

### Available Operations

The following table describes the operations available to GMs.

| Operation | Description |
|-----------|-------------|
| View | The GM reads the memory content. |
| Edit | The GM modifies the memory text. |
| Delete | The GM removes the memory from the system. |
| Add | The GM creates a new memory manually. |
| Regenerate | The system re-runs AI summarization on the source content. |
| Export | The GM downloads memories as JSON or markdown. |

## Entity Integration

The memory system integrates with the entity graph through automatic
extraction and mention linking.

### Automatic Entity Extraction

The system detects entity mentions during chat and queues entities for
review.

1. The detection phase identifies names, places, and items in the text.
2. The matching phase checks detected entities against existing entities
   using Levenshtein distance and semantic similarity.
3. The queue phase adds potential new entities to the extraction queue.
4. The review phase presents queued entities to the GM for approval,
   rejection, or editing.

### Mention Integration

Text containing @mentions creates bidirectional links between memories and
entities. The memory references the entity. The entity shows the memory in
a "mentioned in" list. Clicking an @mention displays an entity preview.

## Implementation Phases

The following phases outline the implementation roadmap.

### Phase 1: Foundation

Phase 1 establishes the core infrastructure.

- The team implements the database schema.
- The system stores chat logs during sessions.
- The system generates basic session summaries.
- The GM interface provides view, edit, and delete operations.

### Phase 2: Intelligent Memory

Phase 2 adds AI-powered memory processing.

- The entity extraction pipeline detects entities in conversation.
- The embedding system generates vectors for memories.
- Semantic search enables retrieval of relevant memories.
- The context assembly service builds AI prompts from memories.

### Phase 3: Advanced Features

Phase 3 implements sophisticated analysis.

- The system detects chapter boundaries automatically.
- The relationship extractor identifies connections between entities.
- Canon conflict detection surfaces contradictions.
- Importance scoring prioritizes memories automatically.

### Phase 4: Polish

Phase 4 optimizes performance and completes tooling.

- Compression optimization reduces token usage.
- Performance benchmarking validates response times.
- Token budget monitoring alerts to context overflow.
- Export and import tools support backup and migration.

## Related Documents

The following documents provide additional context.

- The main design document at `design.md` describes the overall architecture.
- The schema documentation at `SCHEMAS.md` describes the data model.
- The RAG architecture section of `design.md` describes the embedding and
  retrieval system.
