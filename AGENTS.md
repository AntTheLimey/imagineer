# Imagineer Agent Registry

## Overview

Agents are focused analysis tools that can be composed into workflows. They
leverage RAG (Retrieval Augmented Generation) with hybrid search to provide
context-aware assistance based on campaign content and rulebook knowledge.

## Agent Interface

All agents implement:
```go
type Agent interface {
    Name() string
    Description() string
    Run(ctx context.Context, params map[string]any) (Result, error)
}

type Result struct {
    Success     bool
    Data        any
    Suggestions []Suggestion
    Errors      []string
    Sources     []Source  // Attribution for RAG-retrieved content
}

type Source struct {
    Type       string // "campaign" or "rulebook"
    EntityID   *uuid.UUID
    EntityName string
    ChunkText  string
    Score      float64
}
```

## RAG-Powered Agents

These agents use hybrid search (vector + BM25) to retrieve relevant context
before invoking the LLM.

### entity-scanner
- **Purpose:** Scan text to identify and link existing entities, suggest new ones
- **RAG Source:** Campaign content (entities, descriptions)
- **Inputs:** text, campaign_id
- **Outputs:**
  - Existing entities found (with confidence scores)
  - Suggested new entities to create
  - Entity type classifications
- **UI Integration:** "Scan for Entities" button in session planning

### canon-checker
- **Purpose:** Detect contradictions with established campaign facts
- **RAG Source:** Campaign content (all vectorized tables)
- **Inputs:** text_to_check, campaign_id, optional entity_id
- **Outputs:**
  - List of potential contradictions
  - Source references for conflicting facts
  - Severity rating (minor/major/critical)
  - Suggested resolutions
- **UI Integration:** "Check Canon" button in editors

### plot-hole-finder
- **Purpose:** Identify unresolved threads and logical inconsistencies
- **RAG Source:** Campaign content (sessions, scenes, timeline)
- **Inputs:** campaign_id, optional session_id, optional date_range
- **Outputs:**
  - Unresolved plot threads
  - Logical inconsistencies
  - Missing connections between events
  - Suggested follow-ups
- **UI Integration:** "Find Plot Holes" button in session planning

### relationship-suggester
- **Purpose:** Recommend connections between entities based on context
- **RAG Source:** Campaign content (entities, relationships)
- **Inputs:** entity_id, campaign_id
- **Outputs:**
  - Suggested relationships with reasoning
  - Relationship types (ally, enemy, family, etc.)
  - Confidence scores
- **UI Integration:** "Suggest Relationships" button in entity editor

### stat-generator
- **Purpose:** Generate NPC/creature stat blocks following game system rules
- **RAG Source:** Rulebook knowledgebase (character creation, bestiaries)
- **Inputs:** entity_id, game_system_id, optional constraints (power level, role)
- **Outputs:**
  - Complete stat block in game system format
  - Suggested skills/abilities based on description
  - Equipment recommendations
- **UI Integration:** "Generate Stats" button in NPC editor

### rule-lookup
- **Purpose:** Answer questions about game mechanics
- **RAG Source:** Rulebook knowledgebase
- **Inputs:** query, game_system_id
- **Outputs:**
  - Relevant rule excerpts with page references
  - Summary of mechanic
  - Related rules
- **UI Integration:** "Ask Rules" in session planning, global search

### description-expander
- **Purpose:** Flesh out sparse entity descriptions
- **RAG Source:** Campaign content + Rulebook knowledgebase
- **Inputs:** entity_id, campaign_id, optional focus_areas
- **Outputs:**
  - Expanded description draft
  - Suggested attributes based on entity type
  - Connections to existing entities
- **UI Integration:** "Flesh Out Details" button in entity editor

### session-recap-generator
- **Purpose:** Generate session summaries in various styles
- **RAG Source:** Campaign content (sessions, scenes, memorable moments)
- **Inputs:** session_id, style (full/short/stylized)
- **Outputs:**
  - Generated recap text
  - Key events highlighted
  - Entities mentioned
- **UI Integration:** Full/Short/Stylized tabs in session view

## Non-RAG Agents

These agents operate on structured data without LLM invocation.

### consistency-checker
- **Purpose:** Find plot holes, timeline conflicts, orphaned entities
- **Inputs:** campaign_id, optional entity_type filter
- **Outputs:** List of inconsistencies with severity and suggested fixes
- **Method:** SQL queries checking referential integrity, timeline logic

### timeline-validator
- **Purpose:** Check event sequence logic and travel time violations
- **Inputs:** campaign_id, date_range (optional)
- **Outputs:** Timeline inconsistencies, impossible travel times
- **Method:** Graph traversal on timeline events with location distances

### name-checker
- **Purpose:** Flag duplicate or similar names (Levenshtein distance)
- **Inputs:** campaign_id, optional new_name to check
- **Outputs:** Similar names with distance scores
- **Method:** pg_trgm similarity queries

### relationship-mapper
- **Purpose:** Generate visualisation data for connection networks
- **Inputs:** campaign_id, entity_id (optional for ego network)
- **Outputs:** Graph data (nodes, edges) in Cytoscape or mermaid format
- **Method:** Graph queries on relationships table

### clue-redundancy
- **Purpose:** Ensure critical clues have multiple discovery paths
- **Inputs:** campaign_id, clue_ids (optional)
- **Outputs:** Clues with <3 discovery methods, suggested alternatives
- **Method:** Count relationship paths to clue entities

### import-parser
- **Purpose:** Extract entities from unstructured text (non-RAG version)
- **Inputs:** text content, campaign_id, entity_type hints
- **Outputs:** Proposed entities for confirmation
- **Method:** NLP extraction with entity type classification

## Planned Agents

### foreshadowing-tracker
- **Purpose:** Track planted foreshadowing and whether it's been paid off
- **RAG Source:** Campaign content (sessions, scenes)
- **Status:** Planned

### discovery-state
- **Purpose:** Track what players know vs what's hidden
- **Inputs:** campaign_id, player_id (optional)
- **Status:** Planned

### encounter-balancer
- **Purpose:** Evaluate encounter difficulty based on party composition
- **RAG Source:** Rulebook knowledgebase (combat rules, creature stats)
- **Status:** Planned

### npc-voice-generator
- **Purpose:** Generate speech patterns and mannerisms for NPCs
- **RAG Source:** Campaign content (NPC descriptions, culture)
- **Status:** Planned

## Agent Composition

Agents can be composed into workflows:

```go
// Example: Session Prep Workflow
workflow := NewWorkflow("session-prep")
workflow.Add(entity-scanner)      // Find mentioned entities
workflow.Add(canon-checker)       // Verify no contradictions
workflow.Add(plot-hole-finder)    // Check for issues
workflow.Add(relationship-suggester) // Suggest connections
workflow.Execute(ctx, params)
```

## RAG Context Assembly

All RAG agents use a shared context builder:

```go
type ContextBuilder struct {
    CampaignID   uuid.UUID
    GameSystemID uuid.UUID
    TokenBudget  int  // Max tokens for retrieved context
}

func (cb *ContextBuilder) Build(query string) ([]Chunk, error) {
    // 1. Retrieve from campaign content (hybrid search)
    // 2. Retrieve from rulebook (if game system set)
    // 3. Rank and deduplicate
    // 4. Fit within token budget
    // 5. Return with source attribution
}
```

Token budget allocation (default 4000 tokens):
- Campaign content: 70% (2800 tokens)
- Rulebook content: 30% (1200 tokens)
- Adjustable per agent based on use case
