<!--
  Imagineer - TTRPG Campaign Intelligence Platform

  Copyright (c) 2025 - 2026
  This software is released under The MIT License
-->

# Session Workflows Design

This design document describes structured stage workflows for
sessions (Prep, Play, Wrap-up), optional scene management, an AI
chat agent, chapter editor upgrades, and deep integration with the
existing analysis and enrichment pipeline.

## Context

The Imagineer session model already provides CRUD operations,
`prep_notes`, `actual_notes`, and several JSONB fields. This
design adds structured stage workflows, optional scene management,
a conversational AI chat agent, chapter editor upgrades, and
integration with the existing analysis and enrichment pipeline.

## Design Philosophy

This section describes the core principles that guide the session
workflow design.

### Free-Form First, AI Overlay Optional

GMs do not prep in scenes. GMs write notes. Forcing structured
scene lists during Prep fights how GMs actually work. Instead,
the GM writes freely in prep notes, and Imagineer optionally
suggests a scene breakdown by analysing what the GM wrote.

During Play, the system stays out of the way. The table is
chaotic. The GM is improvising, tracking initiative, and
roleplaying NPCs. The system captures free-form actual notes
rather than structured scene logging in real time.

Wrap-up is where AI shines. After the session, the GM has written
notes about what happened. Imagineer analyses those notes,
suggests scene breakdowns, extracts entity mentions, flags new
NPCs, identifies consequences, and updates the campaign knowledge
graph. The GM reviews and confirms at their own pace.

Scenes are a derived artifact, not a primary input. Scenes emerge
from analysis of prep notes (suggested encounters) and actual
notes (what really happened). The GM can manually create, edit,
and reorder scenes at any time.

#### Cross-System Fit

Each supported game system benefits differently from this
approach:

- D&D 5e benefits the most because combat encounters, dungeon
  rooms, and social scenes map cleanly to the scene model.
- CoC 7e fits naturally because node-based investigation scenes
  with clue connections are a core part of the system.
- GURPS 4e varies highly by genre, but scene breakdowns remain
  useful across genres.
- FitD resists pre-planning by design because scores emerge from
  play, but post-hoc scene recording during Wrap-up works well.

### Clues as Entity Type

Clues are an entity type, not a separate data model. The entity
system provides name, description, GM notes, relationships (to
NPCs, locations, items), wiki links, search, vectorisation, and
the full entity CRUD. The scene `connections` JSONB provides the
graph linking scenes via clues. The clue entity describes what
the clue is, and the scene connection describes how the clue
links scenes together.

## Scene Data Model

Each session can have zero or more scenes. Scenes are optional
because a session works fine with just free-form prep and actual
notes and no scenes at all.

The following table describes the new `scenes` table:

| Field               | Type          | Purpose                    |
|---------------------|---------------|----------------------------|
| id                  | BIGSERIAL     | Primary key.               |
| session_id          | BIGINT FK     | Parent session.            |
| campaign_id         | BIGINT FK     | For query efficiency.      |
| title               | TEXT NOT NULL  | Scene title (e.g. "The     |
|                     |               | Warehouse Ambush").        |
| description         | TEXT          | What the scene is about.   |
| scene_type          | TEXT          | The type of scene (see     |
|                     |               | values below).             |
| status              | TEXT          | planned, active,           |
|                     |               | completed, or skipped.     |
| sort_order          | INT           | Position in the scene      |
|                     |               | list.                      |
| objective           | TEXT          | What the GM wants to       |
|                     |               | achieve.                   |
| gm_notes            | TEXT          | Private GM-only notes.     |
| entity_ids          | BIGINT[]      | NPCs, locations, and       |
|                     |               | items involved.            |
| system_data         | JSONB         | System-specific data.      |
| source              | TEXT          | manual, ai_suggested,      |
|                     |               | or ai_discovered.          |
| source_confidence   | TEXT          | DRAFT, AUTHORITATIVE,      |
|                     |               | or SUGGESTION.             |
| connections         | JSONB         | Node-based scenario        |
|                     |               | links (see below).         |
| created_at          | TIMESTAMPTZ   | When created.              |
| updated_at          | TIMESTAMPTZ   | When last modified.        |

The `scene_type` field accepts the following values: `combat`,
`social`, `investigation`, `exploration`, `downtime`, `chase`,
`puzzle`, `ritual`, `stealth`, `travel`, `score`, `free_play`,
`montage`, and `other`.

The `connections` field enables CoC-style node mapping where
scenes link via clues. Each connection has the following
structure:

```json
[
    {
        "target_scene_id": 42,
        "clue_entity_id": 17,
        "description": "Finding the ledger leads here"
    }
]
```

This field remains unused for linear D&D sessions. The
`system_data` JSONB lets each game system store relevant
mechanical data (such as D&D encounter XP budgets, CoC sanity
loss estimates, or FitD clock definitions) without schema changes.

The `description` and `gm_notes` TEXT fields are vectorised for
search, following the same pgedge_vectorizer pattern that chapters
and sessions use.

## Database Changes

This section describes all database schema changes required for
session workflows.

### New Tables

The system creates the `scenes` table as described in the Scene
Data Model section above.

The system creates a `session_chat_messages` table to store chat
agent conversations:

| Field        | Type        | Purpose                       |
|--------------|-------------|-------------------------------|
| id           | BIGSERIAL   | Primary key.                  |
| session_id   | BIGINT FK   | Parent session.               |
| campaign_id  | BIGINT FK   | For query efficiency.         |
| role         | TEXT        | Either `user` or `assistant`. |
| content      | TEXT        | The message text.             |
| sort_order   | INT         | Message sequence.             |
| created_at   | TIMESTAMPTZ | When sent.                    |

The system vectorises both `user` and `assistant` messages.
Campaign search can surface chat history. For example, searching
"what did I ask about the warehouse?" finds the relevant
exchange.

### Modified Tables

The system adds a `play_notes` TEXT field to the sessions table.
This field stores scratchpad content separately from
`actual_notes`. The scratchpad captures ephemeral tracking
(HP, initiative) while `actual_notes` stores the narrative record
of what happened.

### Dropped Fields

The system drops the following JSONB fields from the sessions
table:

- `planned_scenes` is replaced by the `scenes` table.
- `discoveries` is replaced by entity and relationship data from
  the enrichment pipeline.
- `player_decisions` is replaced by entity log entries from the
  enrichment pipeline.
- `consequences` is replaced by entity log entries from the
  enrichment pipeline.

The enrichment pipeline handles all of these concerns as entities
and relationships rather than unstructured JSONB blobs.

### Vectorisation Additions

The system adds vectorisation for the following fields:

- `scenes.description` enables search across scene descriptions.
- `scenes.gm_notes` enables search across GM-only scene notes.
- `session_chat_messages.content` enables search across chat
  history.

## AI Scene Discovery and Suggestion

The AI analyses GM notes and suggests scene breakdowns. Scene
discovery is another analysis pass in the existing pipeline, like
entity detection.

### When Scene Discovery Runs

Scene discovery runs at specific stages:

- During Prep, the GM writes prep notes and clicks "Save &
  Analyse". The analyser scans for scene-like structures and
  suggests a breakdown.
- During Wrap-up, the GM has written actual notes about what
  happened. The analyser suggests scenes reflecting what really
  occurred, which may differ from the prep.
- Scene discovery never runs during Play because the GM is busy
  at the table.

### Discovery Signals

The AI looks for the following signals in text to identify
scenes:

| Signal              | Weight     | Example                   |
|---------------------|------------|---------------------------|
| Location changes    | High       | "When they arrive at the  |
|                     |            | docks..."                 |
| NPC introductions   | Medium     | "Madame Lefevre greets    |
|                     |            | them..."                  |
| Combat or conflict  | High       | "The cultists attack..."  |
| setups              |            |                           |
| Time transitions    | Medium     | "The next morning..."     |
| Investigation beats | High (CoC) | "Searching the study      |
|                     |            | reveals..."               |
| Objective shifts    | Medium     | "Now they need to find    |
|                     |            | the key"                  |
| Skill challenges    | Medium     | "To cross the chasm..."   |
|                     | (D&D)     |                           |
| Score or job        | High       | "The plan is to break     |
| framing             | (FitD)    | into..."                  |

### System-Aware Weighting

The enrichment prompt is tuned per game system. CoC analysis
prioritises investigation nodes and clue connections. D&D
analysis prioritises encounters and exploration. FitD analysis
prioritises score phases and clock triggers.

### Output and Review

Each suggested scene comes back as a `SUGGESTION` with
`source: 'ai_discovered'`. The GM sees the suggestions in a
review panel and can accept, edit, or dismiss each one. Accepted
scenes promote to `DRAFT` or `AUTHORITATIVE`.

Scene discovery uses the same `ContentAnalysisItem` pipeline. The
analyser produces items of type `scene_suggestion`, the triage
page lets the GM review the items, and accepting an item creates
the scene record.

## Session Stage Workflow

Each session moves through three stages plus a terminal state.
The UI adapts based on which stage is active.

### Prep (Before the Session)

The Prep stage provides a workspace for session preparation:

- A full-screen editor displays `prep_notes` with free-form rich
  text and wiki links.
- An optional scene list panel appears in the sidebar, showing
  planned scenes.
- The "Save & Analyse" button runs scene discovery, entity
  detection, and relationship extraction on prep notes.
- The GM can manually add, reorder, and edit scenes, or accept
  AI suggestions.
- A scene connections editor supports node-based investigations,
  primarily for CoC.
- The session stage is set to `prep`.

### Play (At the Table)

The Play stage provides a multi-panel workspace designed for use
at the table.

#### Scene Bar (Top Strip)

A horizontal scrollable strip of scene cards displays scene
titles and type icons. The GM clicks a card to switch the active
scene. The active scene is highlighted. The GM can drag cards to
reorder scenes mid-session. A "+" button enables quick scene
creation. The bar supports multiple scenes marked as "active"
for split-party situations. If no scenes exist, the strip is
hidden or shows just a "Notes" tab.

#### Main Area

The main area contains three panels:

- The scene viewer occupies approximately 40% of the width. The
  viewer shows a read-only view of the active scene's
  description, objective, GM notes, and linked entities as
  clickable chips. When the party splits, tabs or a toggle let
  the GM flip between active scenes. If no scenes exist, the
  viewer shows `prep_notes` as a scrollable read-only document
  with clickable wiki links. Mixed mode shows whatever scenes
  exist plus a "Notes" tab for full prep notes.
- The entity drawer occupies approximately 30% of the width. The
  GM clicks any entity chip in the scene viewer or any wiki link
  in the notes, and the entity details load in the drawer. The
  drawer displays description, stats (from system_data and
  attributes), relationships, and GM notes. The drawer stays
  pinned until the GM clicks a different entity. This panel
  reuses the existing EntityPreviewPanel.
- The scratchpad occupies approximately 30% of the width. The
  scratchpad provides a free-form notes area that persists for
  the session. The GM uses the scratchpad for tracking HP,
  initiative order, who is doing what, loot, and temporary
  conditions. The system saves scratchpad content to
  `play_notes`. The scratchpad auto-saves frequently.

#### Entity Sidebar (Left Edge, Collapsible)

The entity sidebar has two states:

- In collapsed state (approximately 40px wide), a vertical icon
  strip shows entity type icons (person, location, item, faction)
  with count badges. The GM clicks an icon to expand the sidebar.
- In expanded state (approximately 220px wide), a grouped list
  shows entities relevant to the current scene or session. The
  list groups entities by type: NPCs, Locations, Items, and
  Factions. Each entry displays a name and type icon on a single
  line. The GM clicks an entry to load the entity in the entity
  drawer. Collapsing the sidebar returns full width to the main
  panels.

#### Chat Agent (Bottom Dock)

A collapsible panel docks at the bottom of the Play screen. The
Chat Agent section below describes this panel in detail.

The session stage is set to `play`.

### Wrap-up (After the Session)

The Wrap-up stage provides a workspace for post-session
processing:

- The actual notes editor occupies approximately 50% of the
  width on the left side. The editor is fully editable.
- The analysis and triage panel occupies approximately 50% of
  the width on the right side. The panel shows scene suggestions,
  entity suggestions, and relationship suggestions.
- A collapsible reference drawer at the bottom shows `prep_notes`
  in read-only mode.
- The "Save, Analyse & Enrich" button runs the full pipeline on
  actual notes.
- Chat history from the current session is included in the
  enrichment input.
- A prep versus actual diff summary shows planned scenes that
  were completed, skipped, or still pending, and any unplanned
  scenes that the AI discovered.
- The promotion flow operates as follows: as the GM accepts
  suggestions, content moves from SUGGESTION to DRAFT to
  AUTHORITATIVE. Bulk accept is available for low-risk items
  such as log entries and entity links. Higher-impact items
  such as new entities and relationships require individual
  confirmation.
- An "Import Session Notes" button opens a modal with a paste
  text area or file upload (.txt, .md, .html). An append versus
  replace toggle defaults to append. Imported content gets a
  visible separator in the notes. After import, the standard
  "Save, Analyse & Enrich" flow applies.
- The session stage is set to `wrap_up`.

### Completed

Once the GM is satisfied, the GM marks the session as
`completed`. Completed sessions are read-only unless the GM
explicitly reopens a session.

Stage progression follows this order: prep, play, wrap_up,
completed. The GM can move backward (for example, back to Prep
if the session is postponed).

Stage transitions are manual via a button. The system does not
auto-detect stage changes.

## Full-Screen Editor Layout

The session editor uses a full-screen layout with dedicated
routes for each stage.

### Route Structure

The system provides the following routes:

- `/campaigns/:id/sessions/:sessionId/prep` displays the Prep
  stage.
- `/campaigns/:id/sessions/:sessionId/play` displays the Play
  stage.
- `/campaigns/:id/sessions/:sessionId/wrapup` displays the
  Wrap-up stage.

A stage tab bar at the top lets the GM switch between stages,
with the current stage highlighted. All three routes share the
same `SessionEditorPage` component, which renders different
content based on the active stage.

### Shared Header

The shared header displays the session title, campaign
breadcrumb, stage tabs, and save controls. The `SaveSplitButton`
from existing C-track work slots into the header, with the
default action varying by stage:

- Prep uses "Save & Analyse" as the default action.
- Play uses "Save" as the default action.
- Wrap-up uses "Save, Analyse & Enrich" as the default action.

### Chapter Editor

Chapters use the same full-screen route pattern with the
`SaveSplitButton`, entity sidebar, and import button. Chapters
are simpler than sessions because chapters have no stages, no
scenes, no chat agent, and no play mode. A chapter is a
narrative arc with an `overview` text field that receives the
same editorial treatment as the session Prep stage.

## Chat Agent

This section describes the conversational AI chat agent that
assists GMs during play.

### Overview

The chat agent appears as a collapsible panel docked at the
bottom of the Play screen. The agent provides a conversational
interface for GM help during the session.

### Capabilities

The chat agent supports the following actions:

| Action          | Example Prompt             | Result              |
|-----------------|----------------------------|---------------------|
| Create scene    | "The party went to the     | The agent adds a    |
|                 | market"                    | new scene to the    |
|                 |                            | strip as DRAFT      |
|                 |                            | with inferred type. |
| Quick NPC       | "I need a bartender at     | The agent creates   |
|                 | The Rusty Anchor"          | an entity with      |
|                 |                            | name, motivation,   |
|                 |                            | quirk, and basic    |
|                 |                            | stats from the game |
|                 |                            | system schema,      |
|                 |                            | tagged DRAFT.       |
| Quick entity    | "Create a shop called      | The agent creates   |
|                 | Meridian Imports"          | a location entity,  |
|                 |                            | tagged DRAFT.       |
| Rules lookup    | "What's the DC for a hard  | The agent answers   |
|                 | check?"                    | from the game       |
|                 |                            | system schema. No   |
|                 |                            | data is created.    |
| Stat block      | "Give me stats for a       | The agent returns   |
|                 | street thug"               | system-appropriate  |
|                 |                            | stats inline. The   |
|                 |                            | agent does not save |
|                 |                            | the stats unless    |
|                 |                            | the GM says "save   |
|                 |                            | that".              |
| Situation help  | "The rogue wants to        | The agent suggests  |
|                 | pickpocket the guard"      | mechanics, DCs, and |
|                 |                            | consequences based  |
|                 |                            | on the game system. |

### Campaign-Wide Read, Session-Scoped Write

The chat agent's memory (conversation history) is per-session.
The agent's knowledge spans the campaign.

The agent can read all entities, all sessions, all chapters,
entity logs, relationships, chat history from other sessions,
and campaign search results (vector).

The agent can create the following in the current session only:
entities (DRAFT), scenes (DRAFT), and entity log entries (DRAFT).

The agent cannot modify or delete existing data, create
AUTHORITATIVE content, or run enrichment or analysis.

The agent has tool definitions that map to existing API
endpoints. The vector search endpoint is the primary discovery
mechanism.

### Persistent Memory

Every prompt and response is saved to `session_chat_messages`.
Each message is a separate row with either a `user` or
`assistant` role. The system vectorises all messages for
campaign-wide search. When the GM opens the Play screen, the
full chat history for that session loads from the table.

During Wrap-up, the enrichment pipeline can optionally analyse
chat history alongside actual notes.

### Rich Entity Linking in Responses

Chat responses are rendered with entity awareness. A
post-processing step matches entity names against the campaign
entity list using case-insensitive matching. Matched names
render as clickable links that load the entity in the Play
screen's entity drawer.

#### Relationship Colourisation

When a response mentions two or more known entities that have a
relationship, visual indicators show the relationship tone:

- Positive relationships (allied, friendly, employs) display a
  green accent.
- Negative relationships (enemy, rival, opposed) display a red
  accent.
- Neutral relationships (knows, located_at, member_of) display a
  blue accent.

The LLM writes natural language. The client layer adds the rich
interactive layer. The system uses longest entity name match for
overlapping names. If multiple relationships exist between a
pair, the negative relationship takes precedence as a warning
to the GM.

## Entity Log Integration

The existing entity log captures narrative events such as
"Baron Harkness was killed by cultists" and "the party found
the silver key". The entity log already has `session_id` and
`chapter_id` FK fields, plus `occurred_at` for in-game dates.

Events like NPC deaths, clue discoveries, and alliance changes
are captured through the enrichment pipeline. When the GM writes
actual notes and runs enrichment, the LLM generates `log_entry`
suggestions for relevant entities. The GM reviews and accepts
the suggestions through the triage page.

The design does not include a structured event type system at
this time. Free-text content in the entity log is sufficient. A
typed event system (died, discovered, destroyed, relocated) can
be added later for filtering and display.

## Implementation Phases

This section describes the five implementation phases. Each
phase is independently shippable.

### Phase 1: Foundation (Database and Basic UI)

Phase 1 delivers the database schema and basic session editor:

- The `scenes` table migration creates the scene data model.
- The `session_chat_messages` table migration creates the chat
  storage.
- The `play_notes` field is added to the sessions table.
- The `planned_scenes`, `discoveries`, `player_decisions`, and
  `consequences` fields are dropped from the sessions table.
- Scene CRUD is implemented across the Go database layer, API
  handlers, and React hooks.
- The session editor launches as a full-screen route with stage
  tabs.
- The Prep stage provides the notes editor and scene list
  sidebar with manual scene management only.
- The chapter editor is upgraded to a full-screen route with the
  `SaveSplitButton`, entity sidebar, and import button.
- Vectorisation is configured for `scenes.description`,
  `scenes.gm_notes`, and `session_chat_messages.content`.

### Phase 2: Play Mode

Phase 2 delivers the Play stage workspace:

- The Play layout includes the scene viewer (or prep notes
  viewer), entity drawer, scratchpad, and collapsible entity
  sidebar.
- The scene strip provides status toggles for active, completed,
  and skipped scenes.
- The adaptive layout handles three modes: scenes exist,
  notes-only, and mixed.
- The import session notes feature supports paste and upload into
  actual notes.

### Phase 3: Wrap-up and Enrichment Integration

Phase 3 delivers post-session processing:

- The Wrap-up layout includes the triage panel for reviewing
  suggestions.
- The `scene_suggestion` analysis item type is added to the
  enrichment pipeline.
- Scene discovery runs on both prep notes and actual notes.
- The prep versus actual diff summary highlights what changed.
- The session completion state marks sessions as read-only.
- Chat history is included in enrichment input during Wrap-up.

### Phase 4: Chat Agent

Phase 4 delivers the conversational AI assistant:

- The chat UI panel loads and saves persistent message history.
- Campaign-wide read access operates via tool definitions.
- Entity and scene creation tools let the agent create DRAFT
  content.
- Entity linking and relationship colourisation enrich chat
  responses.
- Rules lookup queries game system schemas for answers.
- The minimal viable version supports rules questions and text
  responses first. Tool integration for entity and scene
  creation follows as a fast-follow.

### Phase 5: Node Mapping (CoC Focus)

Phase 5 delivers investigation node mapping:

- The scene connections editor supports clue-based node graphs.
- A visual node map view provides connection indicators
  (starting as a simple list with connection lines).
- The clue entity type tracks discovery through entity log
  entries.

## Verification

After each phase, the team runs the following checks.

### Automated Verification

The following commands must pass after each phase:

```bash
go build ./...
go test ./...
cd client && npx tsc --noEmit
cd client && npx vitest run
make test-all
```

### Manual Testing: Session Workflows

The following manual tests verify session workflow behaviour:

- Open a session and verify that stage tabs (Prep, Play,
  Wrap-up) appear.
- In Prep, write notes, add scenes manually, and confirm that
  "Save & Analyse" discovers scenes.
- In Play, browse scenes or prep notes, click entities, and use
  the scratchpad.
- In Wrap-up, paste imported notes, run enrichment, review
  suggestions, and mark the session as completed.

### Manual Testing: Chat Agent

The following manual tests verify chat agent behaviour:

- Open Play mode and expand the chat dock.
- Ask a rules question and verify the answer comes from the game
  system schema.
- Type "Create a bartender NPC" and verify the entity appears in
  the entity sidebar.
- Verify that entity names in responses are clickable and load
  the entity in the entity drawer.
- Verify that related entities show relationship colour
  indicators.
- Close and reopen the session and verify that chat history
  persists.

### Manual Testing: Chapter Editor

The following manual tests verify chapter editor behaviour:

- Verify the chapter opens in a full-screen route with the
  `SaveSplitButton`.
- Verify the entity sidebar displays campaign entities.
- Verify the import button allows pasting content.
