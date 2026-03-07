# Conversational Platform — Analysis & Plan

> Analysis and implementation plan for evolving Imagineer from an
> editor-centric CRUD application to a conversational world-building
> platform, as described in VISION.md.

## Context

This plan was produced through a guided product direction session
on March 7, 2026. Every decision recorded here was made by the
product owner through structured questioning. This is not
speculative — these are binding design decisions.

---

## Gap Analysis

### What We Have

**Backend (Go + PostgreSQL):**

- Entity CRUD with JSONB attributes, UUID primary keys, soft
  deletes.
- Relationship model with ontology-driven type constraints,
  cardinality validation, and inverse relationship handling.
- Campaign, chapter, session, and scene data model with full REST
  API.
- RAG infrastructure: pgedge_vectorizer, hybrid vector + BM25
  search, campaign-scoped queries.
- ContextBuilder with content-derived queries, deduplication, and
  4000-token budget tracking.
- LLM client abstraction supporting Anthropic, OpenAI, and Ollama.
- Multi-agent enrichment pipeline (TTRPG Expert, Canon Expert,
  Graph Expert) with phase registry.
- Auth system with Google OAuth, JWT tokens, campaign ownership
  scoping.
- SSE streaming for analysis progress.
- Wiki link insertion and propagation.
- Per-finding surgical revision agent with context window
  extraction.

**Frontend (React + TypeScript + MUI):**

- Tiptap markdown editor with wiki link autocomplete and hover
  popovers.
- Entity autocomplete with debounced search.
- MUI component library and dark theme.
- API client layer and React hooks pattern.
- Analysis wizard with phase-specific screens.
- Session editor with stage navigation, scene strip, play mode
  components.

### What We Need

**Backend (new work):**

- Conversational AI layer: streaming responses, tool use loop,
  message persistence, context management with compaction.
- Tool definitions for AI: search entities, create/update entity,
  create relationship, read document, edit document, query
  timeline, generate stat block.
- Per-document/entity chat thread storage with compaction.
- Intent routing: classify whether procedural code or AI handles
  a request.
- Token metering infrastructure (tracking from day one, billing
  later).
- Local auth mode (username/password) alongside Google OAuth.

**Frontend (mostly rebuild):**

- Split-screen workspace layout: editor (top) + chat (bottom).
- Per-document/entity chat component with persistent threads.
- Campaign dashboard home page with graph widget, map widget,
  recent changes, search, overview text.
- Contextual side panels (entity sidebar, timeline, graph) that
  pull in/out on demand.
- Editor integration for AI-driven programmatic changes
  (insertions, replacements, selections awareness).
- Navigation model: document/entity-centric, not page-per-feature.
- Graph visualization component.

### What Survives

**Backend — mostly reusable (~70%):**

- Data model and migrations (entities, relationships, ontology,
  campaigns, chapters, sessions, scenes).
- Database layer and queries.
- RAG/vector search infrastructure.
- ContextBuilder with token budgets.
- LLM client abstraction (extend for streaming tool use).
- Auth system (extend with local auth).
- REST API for entities, relationships, campaigns.
- Wiki link system.
- Game system YAML schemas.
- Ontology loader and seeder.

**Frontend — selectively reusable (~20%):**

- Tiptap editor and markdown integration.
- Entity autocomplete component.
- Wiki link components (autocomplete, hover popover, insertion).
- API client and hooks pattern.
- MUI theme and component styling.

**Frontend — deprecated:**

- Analysis wizard (identify/revise/enrich phases). Replaced by
  conversational interaction where the AI handles analysis
  inline.
- Editor-centric pages (ChapterEditorPage, EntityEditor,
  SessionEditorPage as standalone pages). Absorbed into unified
  document + chat experience.
- SaveSplitButton with "Save & Analyze" modes. No longer a
  separate action.
- Analysis triage pages. No longer a separate workflow.

---

## Architectural Decisions

### 1. Conversation First

The chat is the primary interface. The editor and world model
serve the conversation. The GM directs the AI through chat; the
AI manifests its work in the editor and world model
simultaneously.

### 2. Procedural First, AI for Reasoning Only

Do not waste tokens on work that procedural code can handle.
Database lookups, graph traversals, text search, wiki link
insertion — all procedural. The AI handles: natural language
intent understanding, creative generation, narrative reasoning,
consistency checking.

The architecture follows this flow:

```
GM input (chat message + editor selection state)
    |
    v
Intent classification (AI — lightweight)
    |
    v
Can procedural code handle it?
   YES --> execute directly, update editor/world model
    NO --> assemble minimal targeted context
           --> AI generates/reasons
           --> execute results via procedural tools
```

### 3. Split-Screen Editor + Chat

Every document and entity page presents the same interface:
editor on top, chat on bottom. The editor accepts both human
keystrokes and programmatic changes from the AI. The AI is
aware of the current editor state including text selections.

The editor and chat share a live context. When the GM
highlights text and says "revise this," the AI sees both the
selection and the instruction. When the AI creates an entity
as a side effect of editing, the entity appears in the world
model and a wiki link appears in the editor.

### 4. Per-Document Chat Threads

Each document and entity has its own persistent chat thread.
Threads use compaction: recent messages in full, older messages
summarised. Full transcripts are stored in the database; the
context window sent to the AI is managed within token budgets.

### 5. Explicit Scope, Flowing Side Effects

The GM navigates to a specific document (Session 5 prep,
Kaunitz's entity page). That sets the working scope. If the
conversation drifts — "oh, Kaunitz should have a sister" — the
AI creates the entity and relationships in the world model
without switching the working scope. The Session 5 document
stays in the editor. Side effects flow to the right places.

### 6. Story Structure

Campaign → Chapters → Scenes. Sessions are play events that
move through scenes, turning planned content into canon. The
GM decides what "chapter" means to them — geographic phases,
narrative arcs, acts, episodes. The system does not enforce a
particular story shape.

A one-shot is a campaign with one chapter.

### 7. Documents in the Database

All text content lives in the database, not on the filesystem.
This enables chunking, vectorization, and RAG across all
content. Every piece of text content — entity descriptions,
scene summaries, chapter overviews, session plans, GM notes —
is a co-editable document with the same editing experience.

### 8. Multi-Provider LLM Abstraction

The LLM client abstraction supports multiple providers. This
is both a technical choice and a business model requirement
(BYOLLM subscription tier). The abstraction must extend to
cover streaming responses and tool use, not just one-shot
request/response.

Reference implementations: `pgedge-postgres-mcp` and
`pgedge-rag-server` in `~/PROJECTS/` for patterns on tool use
loops and streaming.

### 9. Token Metering From Day One

Track token usage per conversation, per campaign, per user from
the start. Do not charge for it initially, but ensure the
infrastructure exists for the future billing tiers:

- Local/free: no metering needed.
- BYOLLM subscription: flat fee, user provides API keys.
- Token credits: usage-based billing.

### 10. Auth Modes

- **Google OAuth**: current implementation, for hosted/SaaS
  deployment.
- **Local auth**: username/password, for anyone downloading and
  running locally. Setup wizard creates an initial admin user.

---

## The Four Modes

### Prep / Write (Primary — Build First)

The core working mode. The GM opens a document, the AI is ready
to co-author.

**Default panels on session prep:**

- Left/top: the document being worked on.
- Bottom: the chat.
- On demand: last session's play notes, existing prep notes,
  entity sidebar, timeline.

**Key AI capabilities in prep mode:**

- Plan-vs-reality reconciliation: diff what was planned against
  what actually happened, surface what needs to change.
- Co-editing: the AI updates the document in response to chat
  instructions.
- Entity creation as side effect: new characters, locations,
  items created in the world model as they emerge in the text.
- Wiki link insertion: automatic linking of entity names.
- Mechanics on demand: "give me stats for this NPC" generates a
  stat block appropriate to the campaign's game system and
  attaches it to the entity.

### Explore (Secondary — Build Second)

The engagement magnet. The GM browses their world.

**Campaign dashboard (home page):**

- Campaign overview text (wiki-style, co-editable).
- Graph visualization widget.
- Map widget.
- Recent changes feed.
- Players widget.
- Latest session widget.
- Search box.

**Entity/document pages:**

- Same editor + chat interface as prep mode.
- No modal distinction between "exploring" and "editing" — the
  tools are always available, the GM uses them or doesn't.

### Play (Tertiary — Build Third)

Reference + quick assist + capture. Not curation.

**What Play mode is:**

- Prep document open for reference.
- Quick AI assist: "what's the innkeeper's name?", "describe the
  alley at night", "give me a quick encounter."
- Scratchpad for notes during session.
- Capture only — no entity creation, no relationship editing, no
  canon decisions during play.

**What Play mode is not:**

- A world model curation tool. There is no time during play to
  maintain the knowledge graph.

**Post-play (part of the Play lifecycle):**

- GM imports play notes (typed, pasted, or transcribed from
  recording).
- AI reconciles play notes against the prep plan.
- Surfaces: new entities mentioned, scenes skipped, scenes
  modified, plot threads introduced or resolved.
- GM reviews and approves world model updates.
- Play solidifies planned content into canon.

### Ingest (Deferred — Manual First)

For onboarding existing campaigns.

**Initial approach (Phase 0):**

- Manual import using Claude Code + MCP server.
- Read existing campaign files, extract entities and
  relationships, populate database directly.
- No UI needed for the product owner's own use.

**Future UI:**

- Upload markdown files (primary), with PDF, ENEX, Google Docs,
  and paste-in as additional formats.
- System processes files, builds rough world model.
- GM refines through conversation — correcting, confirming,
  filling gaps.

### Brainstorm (Emergent — No Separate Build)

Brainstorm mode is prep mode with an empty world. The GM starts
a conversation, talks their campaign idea into existence, and
the world model grows from nothing. No separate implementation
needed — it falls out of prep mode naturally.

---

## Build Sequence

### Phase 0 — Seed Canticle

Populate the Imagineer database with the Canticle campaign using
Claude Code and the MCP server. Read existing markdown files,
extract entities, relationships, chapter/scene structure, and
insert directly. This provides real test data for all subsequent
development.

**No code changes required.** This is a data task.

### Phase 1 — Conversational Backend

Build the server-side conversation layer.

- Chat message storage: per-document/entity threads, message
  persistence, compaction.
- Streaming responses: extend LLM client abstraction for
  streaming + tool use across providers.
- Tool use loop: define tools, execute against world model,
  feed results back to LLM.
- Tool definitions: search entities, create entity, update
  entity, create relationship, read document, edit document,
  query timeline.
- Context assembly: scope-aware context loading with token
  budgets, conversation history management.
- Token metering: track usage per conversation/campaign/user.

**Reference implementations:** Study `pgedge-postgres-mcp` and
`pgedge-rag-server` for streaming and tool use patterns.

### Phase 2 — Core UI

Build the new frontend layout.

- Workspace shell: top-level navigation, campaign selector.
- Split-screen component: editor (top) + chat (bottom),
  resizable divider.
- Chat component: message display, streaming response rendering,
  input box, thread persistence.
- Editor integration: Tiptap editor with programmatic change
  API, selection state exposure to chat context.
- Document/entity navigation: route structure, per-page chat
  thread loading.
- Salvage: Tiptap editor, wiki link components, entity
  autocomplete, MUI theme, API client.

### Phase 3 — Prep Mode

The core working experience.

- Plan-vs-reality reconciliation: AI diffs planned content
  against play notes and surfaces changes needed.
- Co-editing: AI updates editor content via tool use in response
  to chat instructions.
- Side-effect entity creation: AI creates entities and
  relationships in the world model during document editing.
- Wiki link auto-insertion during co-editing.
- Contextual panels: pull-out panels for play notes, prep notes,
  entity sidebar.
- Mechanics on demand: stat block generation tool.

### Phase 4 — Campaign Dashboard + Explore

The engagement magnet.

- Campaign dashboard layout with widget grid.
- Campaign overview (co-editable document + chat).
- Entity list with type filtering and search.
- Graph visualization (interactive, clickable nodes).
- Recent changes feed.
- Search integration (existing RAG infrastructure).
- Entity/document pages using the same split-screen layout.

### Phase 5 — Play Mode

Reference + assist + capture.

- Simplified read-first view of prep document.
- Quick-assist chat (lightweight, fast responses).
- Scratchpad for capture during play.
- Session recording attachment (file upload, no transcription
  yet).

### Phase 6 — Post-Play

Synthesis and reconciliation.

- Play notes import (paste, file upload, markdown).
- AI reconciliation: diff play notes against prep, surface
  changes.
- Review interface: approve/reject world model updates.
- Canon solidification: mark scenes as played, skipped, or
  modified.
- Session completion: record what happened, update timeline.

### Phase 7 — Ingest UI

For other GMs onboarding existing campaigns.

- File upload interface (markdown first, then PDF, ENEX, Google
  Docs).
- Background processing: entity and relationship extraction.
- Review and refinement through conversation.
- Paste-in as universal fallback.

### Phase 8 — Polish

Production readiness.

- Local auth mode with setup wizard.
- BYOLLM configuration UI (API key management per provider).
- Token usage dashboard.
- Performance optimisation.
- Mobile-responsive layout for Play mode (tablet at the table).

---

## Open Questions

These decisions were deferred during the planning session and
should be resolved during implementation:

1. **Co-editing presentation:** Should AI document changes appear
   as live typing (operational transforms / CRDTs) or as
   suggestion-mode annotations (accept/reject)? Both build on
   the same foundation. Prototype both and test.

2. **Graph visualization library:** Evaluate options during
   Phase 4. Cytoscape.js is already referenced in the backlog.

3. **Audio transcription:** Session recordings (m4a) are captured
   but transcription is not yet planned. This could unlock
   powerful post-play synthesis but is a significant
   infrastructure addition.

4. **Map generation:** Referenced as a desire but not scoped.
   Evaluate during or after Phase 4.

5. **Image generation for entities:** Referenced as a desire.
   The LLM abstraction already has an image generation service
   selector. Wire up during Phase 4.

---

## Working Model

Development follows a collaborative model: design together
through conversation, Claude implements via sub-agents, the
product owner tests and provides feedback. No time pressure —
development runs in parallel with continued use of Claude
Desktop for actual Canticle play.

The Canticle campaign data (Phase 0) serves as the primary test
fixture throughout development. Every feature should be
validated against real campaign data, not synthetic test cases.
