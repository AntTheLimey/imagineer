<!--
  Imagineer - TTRPG Campaign Intelligence Platform

  Copyright (c) 2025 - 2026
  This software is released under The MIT License
-->

# Phase 1 Conversational Backend Design

This design document describes the server-side conversation
layer for Imagineer. Phase 1 transforms Imagineer from an
editor-centric CRUD application into a conversational
world-building platform. Phase 0 (Seed Canticle) is complete.
This phase builds the backend infrastructure that enables the
conversational interface described in VISION.md.

Every design decision recorded here was made by the product
owner through a structured questioning process.

## Design Principles

Four principles guide all implementation decisions in this
phase.

### Conversation First

The chat is the primary interface. The editor and world model
serve the conversation. The GM directs the AI through chat,
and the AI manifests its work in the editor and world model
simultaneously.

### Procedural First, AI for Reasoning Only

Procedural code handles all work that does not require
intelligence. Database lookups, graph traversals, text search,
and wiki link insertion are all procedural operations. The AI
handles natural language intent understanding, creative
generation, narrative reasoning, and consistency checking.
This principle prevents wasting tokens on mechanical tasks.

### Always LLM (Intent Routing)

All user messages pass through the LLM. The LLM decides
which tools to call. The system uses no separate intent
classification layer. Tools are cheap; the LLM is the router.
This approach avoids the complexity of a separate classifier
and allows the LLM to handle nuanced requests naturally.

### World-Grounded Generation

The TTRPG expert never generates content in isolation. Every
generative request triggers a world model scan that includes
scope resolution, orphan entity discovery, underconnected NPC
identification, and relationship gap analysis. The expert
actively seeks opportunities to connect new content to
existing fabric, turning loose threads into narrative
connective tissue. The goal is not "plausible content" but
"content that makes the world more coherent."

This principle is the core differentiator. A generic LLM
produces a plausible inn with a plausible innkeeper. The
TTRPG expert, grounded in the world model, produces an inn
in the right district, staffed by the barmaid the PCs met
three sessions ago but never defined a workplace for, run by
an innkeeper who has a reason to know the merchant mentioned
in passing last chapter. The world coheres through use.

## Data Model

Three new tables support the conversation layer.

### Conversations Table

The `conversations` table tracks each conversation thread,
scoped to a specific campaign document or entity.

```sql
CREATE TABLE conversations (
    id              BIGSERIAL PRIMARY KEY,
    campaign_id     BIGINT NOT NULL
                        REFERENCES campaigns(id),
    scope_type      TEXT NOT NULL,
    scope_id        BIGINT NOT NULL,
    title           TEXT,
    summary         TEXT,
    summary_tokens  INT DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON COLUMN conversations.scope_type IS
    'entity, chapter, session, scene, or campaign';
COMMENT ON COLUMN conversations.summary IS
    'Compacted conversation summary';
COMMENT ON COLUMN conversations.summary_tokens IS
    'Token count of the current summary';
```

### Messages Table

The `messages` table stores every message in a conversation,
including user turns, assistant responses, tool calls, and
tool results.

```sql
CREATE TABLE messages (
    id              BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL
                        REFERENCES conversations(id),
    role            TEXT NOT NULL,
    content         TEXT NOT NULL,
    tool_name       TEXT,
    tool_input      JSONB,
    tool_result     JSONB,
    tokens          INT,
    compacted       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON COLUMN messages.role IS
    'user, assistant, system, tool_call, or tool_result';
COMMENT ON COLUMN messages.tool_name IS
    'Populated for tool_call and tool_result messages';
COMMENT ON COLUMN messages.tool_input IS
    'Tool input parameters for tool_call messages';
COMMENT ON COLUMN messages.tool_result IS
    'Tool output for tool_result messages';
COMMENT ON COLUMN messages.compacted IS
    'TRUE when message has been summarised and excluded '
    'from the context window';
```

### Token Usage Log Table

The `token_usage_log` table records every LLM call for usage
tracking and future billing.

```sql
CREATE TABLE token_usage_log (
    id              BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL
                        REFERENCES conversations(id),
    campaign_id     BIGINT NOT NULL
                        REFERENCES campaigns(id),
    user_id         BIGINT NOT NULL
                        REFERENCES users(id),
    model           TEXT NOT NULL,
    input_tokens    INT NOT NULL,
    output_tokens   INT NOT NULL,
    total_tokens    INT NOT NULL,
    llm_call_type   TEXT NOT NULL,
    agent_name      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON COLUMN token_usage_log.llm_call_type IS
    'conversation, compaction, or agent_tool';
COMMENT ON COLUMN token_usage_log.agent_name IS
    'For agent_tool calls: ttrpg_expert, canon_expert, '
    'or graph_expert';
```

### Data Model Design Decisions

The data model reflects several deliberate choices.

The system maintains one conversation thread per document
or entity, identified by the combination of `scope_type`
and `scope_id`.

Auto-compaction triggers at the 8,000-token threshold.
The compaction process marks older messages as
`compacted=TRUE`, which excludes them from the context
window while retaining them in the database for full
transcript retrieval.

The system logs token usage per LLM call rather than per
user message. A single user message may trigger multiple
LLM calls through the tool use loop and agent tools.

The `llm_call_type` field distinguishes main conversation
turns from compaction calls and agent tool sub-calls.

## Tool Use Loop (Orchestrator)

The orchestrator manages the streaming conversation loop on
the server side. The orchestrator contains all business logic
for context assembly, tool execution, and message persistence.

### Orchestrator Flow

The orchestrator executes this sequence for each user message.

1. The user sends a message via
   `POST /conversations/{id}/messages`.
2. The server loads the conversation context:
   - The system prompt from a disk template (cached).
   - Campaign and scope context (cached per conversation).
   - Conversation history from the session cache (or the
     database on a cache miss).
   - RAG context from a vector search on the user message
     (fresh per turn).
   - Tool definitions (singleton, loaded once).
3. The server calls the Anthropic streaming API.
4. The server forwards stream events to the client via SSE:
   - `text_delta` events render incrementally on the client.
   - `tool_use` events display a tool activity indicator.
5. If the response contains `tool_use` blocks:
   - The orchestrator executes each tool against the world
     model (database queries, entity CRUD).
   - For agent tools, the orchestrator makes a sub-LLM call
     with the specialist prompt and tool results.
   - The orchestrator feeds tool results back to the
     Anthropic API.
   - Streaming continues from step 4.
6. When the response is final text with no more `tool_use`
   blocks:
   - The orchestrator persists all messages (user, assistant,
     tool_call, tool_result).
   - The orchestrator logs token usage.
   - The orchestrator checks the compaction threshold.
   - The orchestrator sends a `done` SSE event.

### Server-Side Session Cache

An in-memory conversation state cache avoids hitting the
database on every turn within an active chat session.

The cache uses the conversation ID as the key. Each entry
contains the recent messages, system prompt, campaign
context, and tool definitions. The cache evicts entries
after an idle timeout (configurable, default 30 minutes).

The database is always the source of truth. The cache
rebuilds from the database on a miss. On compaction, the
cache is invalidated and rebuilt from the database.

### Context Assembly Caching

Different context components have different freshness
requirements.

The system prompt loads from disk once at startup and
remains cached indefinitely. In development mode, the
system invalidates the cache on file change.

Campaign and scope context caches per conversation. The
system invalidates this cache when the scope changes or
when tool use modifies entities within the conversation.

Tool definitions are a singleton loaded once at startup.

RAG context is fresh per turn. The system performs a vector
search on each user message to pull in relevant world model
context.

Conversation history comes from the session cache for recent
messages or from the database on a cache miss.

## Streaming Provider Interface

A new interface alongside the existing `Provider` supports
streaming and tool use.

### Interface Design

The following types define the streaming provider contract.

```go
type StreamingProvider interface {
    CompleteStream(
        ctx context.Context,
        req StreamingRequest,
    ) (<-chan StreamEvent, error)
}

type StreamingRequest struct {
    SystemPrompt string
    Messages     []Message
    Tools        []ToolDefinition
    MaxTokens    int
    Temperature  float64
}

type StreamEvent struct {
    Type      StreamEventType
    Text      string
    ToolName  string
    ToolInput json.RawMessage
    ToolID    string
    Usage     *TokenUsage
    Error     error
}
```

The `StreamEventType` enum includes `TextDelta`, `ToolUse`,
`ToolResult`, `Usage`, `Done`, and `Error`.

### Separation of Concerns

The `StreamingProvider` is a thin wrapper around the
Anthropic SDK. The provider sends requests and returns a
channel of `StreamEvent` values. The provider contains no
business logic, no tool execution, and no context assembly.

The orchestrator is the thick layer. The orchestrator
assembles context, calls the `StreamingProvider`, executes
tools, feeds results back, and manages the loop. All
business logic lives in the orchestrator.

### Phase 1 Scope

Phase 1 implements `StreamingProvider` for Anthropic only.
The interface exists so that OpenAI and Ollama
implementations can be added later without changing the
orchestrator.

## Tool Registry

The tool registry defines the tools that the LLM can call
during a conversation. Tools fall into three categories.

### Procedural Tools

Eight procedural tools perform direct database operations
with no token cost.

- `search_entities` searches entities by name, type, or
  attributes within the campaign scope.
- `get_entity` retrieves full entity details by ID.
- `create_entity` creates a new entity in the campaign.
- `update_entity` updates an existing entity's attributes
  or description.
- `create_relationship` creates a relationship between two
  entities.
- `get_related_entities` retrieves all entities related to
  a given entity.
- `search_content` performs a RAG search across all campaign
  content.
- `read_game_schema` loads a game system YAML schema for
  mechanics reference.

### Document Tools

Two document tools handle content operations.

- `read_document` reads a chapter, session, scene, or
  entity description.
- `edit_document` modifies document content through
  insertions, replacements, and deletions.

### Agent Tools

Three agent tools make sub-LLM calls with specialist
expertise. Each agent tool has access to a subset of
procedural tools as sub-tools, allowing the agent to
actively query the world model during its work. This
mirrors how Claude Code's Agent tool gives sub-agents
access to Read, Grep, and Glob.

`ask_ttrpg_expert` handles scene design, encounter
building, session planning, NPC voice, mechanics
validation, scenario writing, and pacing analysis. The
TTRPG expert operates in evaluate mode (critique) and
create mode (generate). The expert's sub-tools include
`search_entities`, `get_entity`, `get_related_entities`,
`search_content`, `read_document`, and `read_game_schema`.

`ask_canon_expert` handles contradiction detection,
consistency verification, fact-checking against established
canon, and hypothetical change impact assessment. The canon
expert's sub-tools include `search_entities`, `get_entity`,
`get_related_entities`, and `search_content`.

`ask_graph_expert` handles relationship suggestions, orphan
detection, graph hygiene, deduplication checking, and path
and connection discovery. The graph expert's sub-tools
include `search_entities`, `get_entity`, and
`get_related_entities`. The graph expert also has direct
database access for structural checks.

### World-Grounded Generation Flow

When the TTRPG expert receives a generative request (for
example, "I need an inn with an innkeeper"), the expert
follows this internal flow.

1. Scope resolution determines the document the user is
   editing, including the location, time period, and
   chapter.
2. A location query calls
   `search_entities(type=location, scope=current_chapter)`
   to find existing locations, including any existing inn.
3. An orphan scan calls
   `search_entities(type=npc, relationship_count=low)` to
   find NPCs that exist but lack connections, such as NPCs
   mentioned but never placed at a location.
4. A relationship gap query calls
   `get_related_entities(entity=current_location)` to find
   nearby entities that lack links.
5. The expert generates content with connections: creating
   the inn in the right district, assigning underconnected
   NPCs as staff, and giving the innkeeper reasons to know
   existing characters.

The GM did not ask for connective tissue. The GM asked for
an inn. The expert found loose threads and wove them in.
The world coheres through use.

### TTRPG Expert Capabilities

The existing TTRPG expert evaluates content across eight
dimensions: pacing, investigation, spotlight, NPC
development, mechanics, player agency, continuity, and
setting. For conversation mode, the expert adds generative
capabilities.

- Scene design and pacing helps the GM structure individual
  scenes (e.g., "Help me design the confrontation at the
  anatomical theatre").
- Encounter building creates balanced encounters and
  validates them against the game schema (e.g., "Build a
  combat encounter for 4 investigators").
- Stat block generation produces character statistics using
  `read_game_schema` (e.g., "Give me stats for this NPC").
- NPC voice and dialogue captures character speech patterns
  (e.g., "How would Herzfeld speak to a subordinate who
  failed him?").
- Session planning structures game sessions with time
  estimates (e.g., "I have 3 hours -- structure Session 6
  for me").
- Play time estimation predicts scene duration (e.g., "How
  long will this scene take to run?").
- Narrative structure analysis checks tension beats and arc
  completeness (e.g., "Does this chapter arc have enough
  tension beats?").
- Player engagement suggestions generate hooks for passive
  players (e.g., "My players are getting passive -- suggest
  hooks").
- Rules adjudication answers mechanics questions using the
  game schema (e.g., "How does Sanity loss work for
  witnessing the Engine?").
- Scenario writing drafts read-aloud text and descriptions
  (e.g., "Draft the read-aloud text for entering the
  catacombs").
- Reconciliation identifies differences between prep and
  play (e.g., "What changed between prep and play? What
  needs updating?").

All generative capabilities use the world-grounded
generation flow described above.

### Canon Expert Capabilities

Beyond passive contradiction detection, the canon expert
gains active capabilities for conversation mode.

- Fact verification checks whether a statement is safe to
  make (e.g., "Is it safe to say Kaunitz was in Venice in
  July?").
- Knowledge summaries compile what the campaign establishes
  about a topic (e.g., "What's established about the Bauer
  brothers?").
- Hypothetical checking assesses whether a proposed change
  contradicts existing canon (e.g., "I'm thinking of making
  Herzfeld a former priest -- does that contradict
  anything?").
- Conflict resolution guidance helps the GM choose between
  conflicting accounts (e.g., "We have two conflicting
  accounts of this event -- which should be canon?").

### Graph Expert Capabilities

Beyond structural validation, the graph expert gains active
capabilities for conversation mode.

- Connection suggestions recommend relationships for new
  entities (e.g., "I just created a new faction -- what
  should it connect to?").
- Path discovery traces connections between entities (e.g.,
  "How is Kaunitz connected to the Calcutta cell?").
- Impact analysis predicts the consequences of structural
  changes (e.g., "If I remove this faction, what breaks?").
- Deduplication detection identifies potential duplicate
  entities (e.g., "Is this entity a duplicate of
  something?").

## Conversation Compaction

Compaction keeps the context window manageable while
preserving conversation intelligence.

### Compaction Trigger

Token-count threshold triggers compaction. When total
conversation tokens exceed 8,000, the system runs the
compaction process.

### Compaction Method

The system uses hybrid summarisation.

1. The summariser condenses older messages into a structured
   summary covering key decisions, entities discussed,
   actions taken, and unresolved threads.
2. The summariser merges the new summary with the existing
   conversation summary incrementally (merge, not replace).
3. The system marks compacted messages as `compacted=TRUE`
   in the database.
4. The system retains recent messages (post-compaction) in
   full.

### Ground Truth Validation

Compaction summaries can contain stale information. An
entity's description may have changed through a tool call
after the summary was written. The system prevents drift
through validation.

During compaction, the summariser has access to the same
procedural tools (`search_entities`, `get_entity`,
`get_related_entities`). The summariser validates key facts
in the summary against the current world model. If the
summary references "Kaunitz is the cult leader" but the
Kaunitz entity has since been updated, the summariser
corrects the summary.

This approach uses the same tool infrastructure as the main
conversation, requiring no special code path. The approach
mirrors how the Canon Expert already checks new content
against established facts. The compaction summariser applies
the same principle to its own output.

### Token Budget

The context window budget allocates tokens across five
components.

- The conversation summary targets a maximum of 1,000
  tokens.
- Recent messages use up to 4,000 tokens.
- RAG context uses up to 4,000 tokens (matching the
  existing `ContextBuilder` budget).
- The system prompt and tool definitions use approximately
  2,000 tokens.
- The total context budget of approximately 11,000 tokens
  leaves headroom within the Anthropic context window for
  the LLM's response.

## API Endpoints

Four routes under the campaign namespace handle conversation
operations.

### Create Conversation

```
POST /campaigns/{cid}/conversations
```

This endpoint creates a new conversation. The request body
contains `scope_type` and `scope_id`. The endpoint returns
the conversation object with its ID.

### Get Conversation

```
GET /campaigns/{cid}/conversations/{id}
```

This endpoint retrieves conversation metadata and recent
messages. The `limit` query parameter (default 50) controls
message count. The `before` query parameter accepts a
message ID for pagination. The endpoint returns the
conversation object with paginated messages.

### Send Message

```
POST /campaigns/{cid}/conversations/{id}/messages
```

This endpoint sends a user message and returns an SSE
stream. The request body contains `content` and an optional
`editor_selection` field. The response is an SSE stream
containing `text_delta`, `tool_use`, `tool_result`, `done`,
and `error` events.

### List Conversations

```
GET /campaigns/{cid}/conversations
```

This endpoint lists conversations for a campaign. The
`scope_type` and `scope_id` query parameters filter results.
The `limit` query parameter (default 20) controls the page
size. The endpoint returns a paginated conversation list.

### SSE Event Format

The SSE stream uses the following event types and data
formats.

The `text_delta` event delivers incremental text from the
assistant.

```
event: text_delta
data: {"text": "The inn sits on..."}
```

The `tool_use` event signals that the LLM is calling a tool.

```
event: tool_use
data: {"tool": "search_entities", "input": {...}}
```

The `tool_result` event delivers the result of a tool call.

```
event: tool_result
data: {"tool": "search_entities", "result": {...}}
```

The `done` event signals the end of the response.

```
event: done
data: {"message_id": 42, "tokens": {"input": 1200,
  "output": 380}}
```

The `error` event delivers error information.

```
event: error
data: {"message": "..."}
```

## System Prompt

The system prompt lives on disk at
`config/prompts/conversation.tmpl`, not compiled into Go
code. This design makes iteration fast: the developer edits
the template, restarts the server (or hot-reloads in
development mode), and tests immediately.

The template receives campaign context, scope information,
game system details, and the current date. The prompt
establishes the AI's role as a co-author who knows the GM's
world and instructs the AI to use tools to ground its
responses in the world model.

## File and Package Structure

Phase 1 introduces new packages and files alongside the
existing codebase.

### Conversation Package

The `internal/conversation/` package contains the core
conversation logic.

- `orchestrator.go` contains the main loop: context
  assembly, streaming, tool execution, and message
  persistence.
- `compaction.go` contains the conversation summarisation
  and ground truth validation logic.
- `session_cache.go` contains the in-memory conversation
  state with idle eviction.
- `tools.go` contains the tool registry: definitions,
  execution dispatch, and agent tool sub-tool mapping.
- `tools_procedural.go` contains the procedural tool
  implementations.
- `tools_agent.go` contains the agent tool implementations
  that wrap the existing experts.
- `tools_document.go` contains the document read and edit
  tool implementations.
- `sse.go` contains the SSE response writer.

### LLM Package

The `internal/llm/` package contains the streaming provider
abstraction.

- `streaming.go` defines the `StreamingProvider` interface.
- `anthropic_stream.go` contains the Anthropic streaming
  implementation.

### API and Migration Files

The API layer and database migration round out the new
files.

- `internal/api/conversation_handlers.go` contains the HTTP
  handlers for the conversation endpoints.
- `migrations/NNN_conversations.sql` creates the
  `conversations`, `messages`, and `token_usage_log` tables.
- `config/prompts/conversation.tmpl` contains the system
  prompt template.

### Integration with Existing Code

Tool implementations call existing `internal/database/`
functions for entity CRUD, relationship management, and
search.

Agent tools wrap existing agents from
`internal/agents/ttrpg/`, `internal/agents/canon/`, and
`internal/agents/graph/`. The agent tools reuse the existing
system prompts and analysis logic while extending them with
sub-tool access and generative capabilities.

Context assembly reuses the existing `ContextBuilder`
pattern from `internal/enrichment/context.go` for RAG
context.

SSE streaming follows the existing pattern used for
enrichment progress in `internal/api/`.

## Token Metering

Every LLM call is logged to `token_usage_log` with the
following fields.

- The conversation that triggered the call.
- The campaign and user that the call belongs to.
- The model used for the call.
- Input, output, and total token counts.
- The call type (conversation turn, compaction, or agent
  tool).
- For agent tool calls, the agent name (`ttrpg_expert`,
  `canon_expert`, or `graph_expert`).

This logging enables future billing tiers (bring-your-own
LLM, token credits) without code changes. Billing logic
queries the log table directly.

## Scalability Considerations

The architecture supports a future SaaS offering with many
concurrent users.

The session cache uses per-conversation keys with idle
eviction. Memory scales with the number of active
conversations, not total users.

The orchestrator is stateless between turns because state
lives in the cache or the database. The orchestrator scales
horizontally behind a load balancer.

SSE connections are long-lived but lightweight. The standard
Go HTTP server handles thousands of concurrent connections.

The database is the bottleneck for concurrent tool
execution. Connection pooling through pgx and read replicas
handle this concern at scale.

LLM API calls are the real constraint. Rate limiting and
queue management are needed at scale but not for Phase 1,
which targets a single user.

## Out of Scope for Phase 1

Several features are explicitly deferred to later phases.

- OpenAI and Ollama streaming providers are deferred. The
  `StreamingProvider` interface exists, but only the
  Anthropic implementation ships in Phase 1.
- The frontend is deferred to Phase 2. Phase 1 is backend
  only, testable via curl or Postman.
- Editor integration is deferred to Phase 2 or Phase 3.
  The `edit_document` tool exists, but the frontend wiring
  comes later.
- Local authentication is deferred to Phase 8.
- Billing is deferred. Token metering is in place, but
  billing logic is a future concern.

## References

The following documents provide additional context.

- `VISION.md` describes the product vision and design
  principles.
- `ROADMAP.md` defines the build sequence and phase
  definitions.
- `design.md` describes the architecture and design
  philosophy.
