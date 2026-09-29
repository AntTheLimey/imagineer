<!--
  Imagineer - TTRPG Campaign Intelligence Platform

  Copyright (c) 2025 - 2026
  This software is released under The MIT License
-->

# Frontend Design

This document records the validated frontend design for
Imagineer. The product owner produced these decisions through
a guided brainstorming session on March 10, 2026. All
decisions recorded here are binding design commitments.

## The Metaphor

The interface embodies a writer's desk with a living world
pinboard. The GM writes at the centre of the workspace,
reference material and entities are pinned around the writing
surface, and an AI collaborator lives in a conversation space
below. The world model grows silently as a byproduct of the
creative process.

Imagineer is not an editor with AI bolted on. Imagineer is
not a database with a pretty face. Imagineer is the place
where a Game Master thinks, writes, and explores with an AI
partner that knows the GM's world as deeply as the GM does.

## How GMs Actually Prep

The design is grounded in how Game Masters actually work, not
how software designers think they should work.

GMs do not prep sessions. GMs prep story. The session
boundary is irrelevant to the creative process because the
session is just when you stop playing.

GMs manage several concerns simultaneously:

- Players and their trajectories are the starting points.
- A destination (an epic conclusion the GM steers toward)
  anchors the arc.
- A web of scenes fills the space between those two points,
  with NPCs placed at specific moments.
- Multiple paths through the web exist because no outcome is
  guaranteed.
- Constant replanning happens when players detour.

Writing is the thinking. GMs write rapid-fire notes in
shorthand, fragments, and imperatives rather than polished
prose. Rough notes and polished prose coexist in the same
document. Refinement happens selectively, not uniformly.

Context switching is a design smell, not a feature. In a
well-designed system, the GM should not need to switch
contexts because:

- Villain activity is written inline.
- The system surfaces connections to past content.
- Future content is just more writing.
- Tangents get captured without breaking flow.

## The Workspace

The workspace comprises three layers that together form the
GM's creative environment.

### The Editor (Centre Stage)

The editor is a freeform writing surface where rough notes
and polished prose coexist. The GM lives in the editor during
prep. The editor supports the following capabilities:

- Freeform text with progressive refinement allows the GM to
  select text and ask the AI to "polish this."
- Inline @ tags dispatch async requests to the AI without
  breaking the GM's flow.
- Entity name detection runs on typing pause and suggests
  wiki links for recognized names using procedural matching
  at no LLM cost.
- AI-generated draft content appears directly in the editor,
  clearly marked as draft through a subtle highlight or
  border.
- Selection awareness means that when the GM highlights text,
  the chat knows what the GM is referring to.

### The Pinboard (Floating Entity Cards)

Floating, draggable, collapsible entity cards surround the
editor. The pinboard supports these interactions:

- The GM pins cards manually by dragging from search, pulling
  from a mention, or accepting an AI-created entity.
- Cards are draggable and collapsible, transitioning between
  full card, title strip, and dismissed states.
- Pull-on-demand means the GM highlights an entity name and
  the entity card appears.
- AI-proposed entities appear as ghost cards that are visible
  but not yet part of the world model.
- The GM approves ghost cards to make them real entities, or
  dismisses them so they vanish without trace.

When the AI generates a new entity (for example, from an
`@[NEED NPC]` tag), the system follows this flow:

1. A new NPC suggestion appears as the most likely result,
   with a sidebar showing potentially matching existing NPCs.
2. Accepting the suggestion creates the entity, updates the
   text with a linked name, and gives the name a hover-over
   detail popover.
3. Choosing an existing NPC instead uses that NPC's name,
   links the text to that NPC, and adds a contextual reminder
   (for example, "last met in X town, on bad terms").

### The Chat (Below the Editor)

The chat panel is not always visible. The GM pulls the chat
up when needed. The chat serves two purposes:

- Deliberate review allows the GM to say "look at what I've
  written, find the gaps."
- Card resolution allows the GM to pull a finding card into
  the chat, discuss the finding, and direct the AI to act.

The chat is context-aware. The chat knows the current
document, the editor selection, and which card the GM has
pulled in for discussion.

## AI Interaction Model

The AI interaction model consists of three distinct patterns
that serve different creative needs.

### Pattern A: Inline Tags (Async, Mid-Flow)

The GM writes `@[natural language request]` anywhere in the
text. The following examples illustrate typical requests:

- `@[need a mining boss NPC]` requests a new NPC.
- `@[need 3 interesting tavern patrons]` requests multiple
  NPCs.
- `@[what creature would make sense in a river-bend mining
  town?]` requests a creative suggestion.
- `@[insert combat stats for 4 henchmen, thugs, not trained
  soldiers]` requests stat blocks.
- `@[need a reason why the captain is loyal despite the
  corruption]` requests narrative motivation.
- `@[how should this scene end if the party refuses to
  help?]` requests a contingency plan.

The system picks up each tag and works on the request in the
background. The GM keeps writing without interruption. When
a result is ready, the result appears contextually:

- A name or short answer replaces the tag as an inline
  suggestion.
- An entity appears as a proposed card on the pinboard, and
  the tag is replaced with a linked entity name.
- Multiple results cluster as options near the tag location.
- Longer content appears as an expandable card.

Tags are not commands. Tags are natural language requests to
the AI, embedded in the text. The @ symbol is just the GM
saying "this bit is for you, not for the document."

Each tag generates one LLM call with visible token cost.

### Pattern B: Deliberate Review (Explicit, On Request)

The GM finishes a chunk of writing and explicitly asks the AI
to analyze holistically. The AI examines:

- Plot holes in the narrative.
- Player agency gaps where players lack meaningful choices.
- Spotlight balance (whether every PC has active moments
  across scenes).
- Missing player choice paths.
- Mechanical issues with game system rules.
- Narrative consequences that the GM has not addressed.

Results appear in two forms:

- Small findings appear as annotations anchored to the
  relevant text, following the Google Docs comment style.
- Large findings appear as actionable cards.

Every finding must be actionable. The AI does not produce
observations alone but produces things the GM can do
something about.

The system shows token cost before confirmation and after
completion.

### Pattern C: Card Resolution (Conversational, In Chat)

The GM pulls a card (a finding, an entity proposal, or an
idea) into the chat. The chat thread now has that card as
context, for example "we're talking about Davros's spotlight
balance."

The GM can interact in several ways:

- The GM asks the AI for suggestions ("how do you suggest we
  address this?").
- The GM directs the AI to act ("add a court scene for Davros
  where he defends himself").
- The GM acts on the discussion by editing the document
  directly.

The AI modifies draft content in the editor and updates
proposed structural changes. The GM acts or directs, and
either way, changes flow through the structural change
pipeline.

## The Structural Change Pipeline

When the AI creates or modifies world model content
(entities, relationships, or scene placement), the content
follows a strict pipeline.

### Draft, Revise, Approve, Commit

The pipeline has four stages.

1. Draft: AI-generated prose appears in the editor, clearly
   marked as draft. Proposed entities appear as ghost cards on
   the pinboard, visible but not yet real.
2. Revise: the GM edits the prose directly or directs the AI
   through chat. As the prose changes, proposed structural
   changes update reactively. If the GM replaces Judge
   Harwick with Ser Jon Barnes, the proposed "create Judge
   Harwick" entity disappears and a "link to Ser Jon Barnes"
   action takes its place.
3. Approve: when the prose is right, the GM reviews pending
   structural changes. These changes should be mostly obvious
   by this point. The GM can approve or dismiss each item
   individually, or approve all at once.
4. Commit: approved changes enter the world model. The system
   creates entities, establishes relationships, and finalizes
   wiki links. Draft highlights clear. Ghost cards become real
   pinboard cards.

The key principle is that the world model never gets polluted
with half-baked ideas. Unapproved proposals vanish without
trace.

## Non-Writing Experiences

Several views serve the GM outside the core writing
workspace.

### Campaign Dashboard (Home Page)

The campaign dashboard is the landing page for each campaign.
The dashboard contains the following elements:

- A campaign overview text that is itself a co-editable
  document with chat available.
- A graph visualization widget with interactive, clickable
  nodes.
- A map widget that starts as a relationship and travel map.
- A recent changes feed.
- A search box.
- A scratchpad access point.

### Graph Explorer (Full-Screen)

The graph explorer provides an interactive entity-relationship
visualization that serves four simultaneous purposes.

1. Discovery reveals hidden connections the GM had not noticed
   (for example, "these two NPCs are both connected to the
   same faction through different paths").
2. Completeness shows well-developed clusters, thin spots,
   and orphaned entities. The graph explorer acts as a health
   check on world-building.
3. Beauty renders the world as a living constellation. The GM
   zooms in for detail, zooms out for the tapestry, and
   experiences the moment where "look at what I built" becomes
   real.
4. Utility provides a visual way to browse the wiki. The GM
   clicks nodes to see entities and follows connections to
   navigate.

### Wiki

The wiki provides classic entity-to-entity navigation through
relationships. Each entity page shows everything on one
page: description, relationships, attributes, game stats, and
mentions across documents. The display is dense and complete
with no progressive disclosure because GMs want to glance
and see the full picture.

Entity pages are documents with chat available. The GM can
start editing or talking to the AI at any point during
exploration. There is no mode switch between exploring and
editing.

### Map

The map starts as a relationship and travel map between
locations. Location entities accumulate geographic detail over
time, including relative positions, distances, terrain, and
connections. The data model is designed so that rendering can
grow from a functional travel diagram to illustrated fantasy
cartography.

### Scratchpad

The scratchpad is a persistent per-campaign idea drawer.

- The scratchpad is always one click away, quick to open and
  close.
- Ideas accumulate over time.
- Chat can populate the scratchpad ("add this to the
  scratchpad").
- The GM reviews the scratchpad periodically to develop ideas
  into scenes or entities, or to discard them.

## Cost Model

The cost model distinguishes between free procedural
operations and user-initiated LLM operations.

### Free Operations (Procedural, Automatic)

The system runs these operations automatically at no token
cost:

- Entity name detection on typing pause (fuzzy and trigram
  matching against existing entities).
- Wiki link suggestions.
- Duplicate name detection.
- Structural change proposals updating reactively as the GM
  edits.
- Graph and map rendering.
- Wiki browsing, search, and navigation.

### User-Initiated Operations (LLM, Cost Visible)

These operations incur token cost, and the system shows the
cost visibly:

- @ tag requests each generate one LLM call, with cost shown
  per request.
- Deliberate review and analysis show cost before
  confirmation and after completion.
- Chat conversation charges tokens per exchange.
- Content generation (scene drafts, NPC backstories, stat
  blocks) incurs token cost.

Token usage is always visible to the GM:

- A per-request cost indicator is small and unobtrusive but
  always present.
- A campaign-level usage summary is accessible from the
  dashboard.
- No surprise bills occur because the GM chooses when to
  incur cost.

## What Survives from the Current Frontend

Several components from the current frontend carry forward
into the new design:

- The Tiptap editor and markdown integration form the
  foundation for the editor layer.
- The entity autocomplete component carries forward.
- Wiki link components (autocomplete, hover popover, and
  insertion) carry forward.
- The API client and React Query hooks pattern carries
  forward.
- The MUI theme and dark color scheme (purple, amber, and
  slate TTRPG aesthetic) carries forward.
- React Router handles navigation.

## What Is Deprecated

The new design deprecates several current frontend features:

- The analysis wizard (identify, revise, and enrich phases)
  is replaced by Pattern B deliberate review.
- Editor-centric standalone pages (`ChapterEditorPage`,
  `EntityEditor`, and `SessionEditorPage`) are absorbed into
  the unified workspace.
- `SaveSplitButton` with "Save & Analyze" modes is no longer
  a separate action.
- Analysis triage pages are no longer a separate workflow.
- Sidebar navigation as primary navigation is replaced by
  wiki-style entity-to-entity navigation and search.

## Open Questions

These decisions are deferred to implementation.

1. Co-editing presentation: should AI document changes appear
   as live typing or as suggestion-mode annotations (accept
   and reject)? The team will prototype both approaches.
2. @ tag syntax: the exact delimiter for inline AI requests
   must not conflict with wiki link syntax, which uses square
   brackets.
3. Graph visualization library: the team will evaluate options
   during implementation. Cytoscape.js is a candidate.
4. Map rendering approach: the team will evaluate options when
   the location data model is mature.
5. Pinboard persistence: should card positions persist across
   sessions, or reset on page load?
6. Mobile and tablet support: desktop is the primary target.
   Mobile provides read access and simple edits. Tablet at
   the table provides play mode.
