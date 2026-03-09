# Imagineer — Product Vision

> The North Star for what Imagineer is, who it serves, and why it
> exists. All architectural and implementation decisions should trace
> back to this document.

## What Imagineer Is

Imagineer is a conversational world-building and campaign management
platform where the AI is your co-author, and the structured world
model is a living byproduct of your creative process.

It is not an editor with AI bolted on. It is not a database with a
pretty face. It is the place where a Game Master thinks, writes,
plays, and explores — with an AI partner that knows their world as
deeply as they do.

## The Core Insight

Game Masters already have a creative workflow. They brainstorm
ideas, write session plans, run sessions, take notes, and iterate.
The structured knowledge — the entities, relationships, timelines,
and canon — is not the primary artifact. It is a **byproduct** of
the creative process, extracted and maintained by the system while
the GM works.

The conversation is the interface. The world model is the memory.

## Who It Serves

A Game Master who:

- Runs narrative-heavy campaigns across multiple sessions and
  chapters.
- Cares about consistency, continuity, and canon across a large
  and growing world.
- Wants AI assistance that understands *their* world, not just
  generic TTRPG advice.
- Wants to see their world come alive — browseable, connected,
  explorable.

## The Four Modes

Imagineer supports four modes of interaction. A GM moves between
them fluidly, sometimes within a single sitting.

### 1. Ingest

*"I have an existing campaign. Here's everything."*

The GM throws their existing content at Imagineer — notes, plans,
play logs, images, handouts, PDFs, recordings. The system processes
it all, builds a rough world model (entities, relationships,
timeline, locations), and presents it for review. The GM then
refines through conversation: correcting, confirming, and filling
gaps. The world model solidifies through dialogue, not through
form-filling.

### 2. Brainstorm

*"I have an idea for a new campaign. Help me build it."*

The GM starts with nothing — or a seed of an idea — and talks it
into existence with the AI. The conversation explores themes,
settings, factions, conflicts, and characters. As the idea takes
shape, the world model begins to form underneath. By the end of a
brainstorming session, the GM has a campaign skeleton and the system
has its first entities, relationships, and narrative arcs.

### 3. Prep / Write

*"I need to plan what happens next."*

This is the primary working mode. The GM opens a conversation with
the AI, with contextual panels flanking the chat:

- **Default left panel**: Last session's actual play notes — what
  happened, what the players did, where they left off.
- **Default right panel**: Existing prep notes and planned content
  that hasn't been played yet.
- **Centre**: The AI conversation, where the GM directs and the AI
  co-authors.

The AI's first job in prep mode is **reconciliation** — diffing
what was planned against what actually happened and surfacing what
needs to change. If the GM planned a confrontation with a warlord
but the players assassinated him last session, the AI flags it and
proposes how the scene transforms.

The GM works at whatever scope feels right:

- **Campaign scope** — overall arc, themes, endgame.
- **Chapter scope** — a narrative unit with its own tensions and
  geography.
- **Scene scope** — the atomic unit of story, with NPCs, beats,
  and encounters.

Sessions are not a structural unit of the world. A session is a
finite window of linear progression through the story — the
aperture through which planned content becomes reality, twisting
future potential as it goes. Scenes are the real building blocks;
sessions are just how they get played.

The AI co-edits documents in real time alongside the chat. The GM
says "make the second-in-command panicking, not the captain" and
the document updates. New entities, relationships, and plot threads
are extracted and added to the world model as the GM works —
silently, in the background, surfaced only when review is needed.

### 4. Explore

*"Let me wander through my world."*

The GM browses the wiki, follows relationship threads, explores
the entity graph, reads the timeline. This mode is not about
production — it is about the joy of seeing a living, connected
world that you built. It is what makes the GM visit Imagineer
between sessions, not just during prep.

The explore mode is also where the GM discovers things the system
has inferred — connections they hadn't noticed, timeline
inconsistencies, orphaned plot threads, characters who haven't
appeared in three chapters.

## The Lifecycle

The four modes form a natural cycle:

```
Ingest / Brainstorm
        |
        v
      Prep / Write  <---.
        |                |
        v                |
      Play (external)    |
        |                |
        v                |
      Post-Play --------'
        |
        v
      Explore
```

1. **Ingest or Brainstorm** — the world model is born.
2. **Prep** — the GM and AI co-create content for the next
   session. The world model grows.
3. **Play** — the session happens (outside Imagineer, at the
   table). The GM captures what occurred — notes, recordings.
4. **Post-play** — the system reconciles plan vs reality, updates
   the world model, and surfaces what changed.
5. **Explore** — the GM browses their world, gets inspired, and
   the cycle feeds back into prep.

Each loop makes the world richer and more solid. The world does
not exist to explore without prep mode, and does not become solid
without play.

## What the AI Does That a Generic LLM Cannot

- **Maintains a structured world model** across sessions —
  entities, relationships, timeline, canon, and narrative arcs.
- **Diffs plans against actual play** and surfaces what needs to
  change before the GM even asks.
- **Enforces narrative consistency** across chapters and sessions
  — catching contradictions, timeline conflicts, and forgotten
  plot threads.
- **Generates content woven into the existing world** — not
  generic TTRPG advice, but content actively connected to
  established characters, locations, factions, and history.
  When asked to create an inn, the AI places it in the right
  district, staffs it with NPCs the players met three sessions
  ago, and gives the innkeeper a reason to know the merchant
  from last chapter. Every generative act strengthens the
  world's connective tissue.
- **Extracts structure from unstructured content** — turning
  session notes, brainstorming conversations, and imported
  documents into entities, relationships, and timeline events.

## What Falls Out Later

These are valuable but not the core product. They become possible
*because* the structured world model exists:

- **Publishable scenario extraction** — generate structured
  campaign materials from played sessions.
- **Fiction and book generation** — narrative output from play
  history, assembled and polished with AI assistance.
- **Shareable world bibles** — export the wiki and world model
  for other GMs or collaborators.

## The Two Reasons a GM Uses Imagineer

1. **They USE it because** the AI assistance is better here than
   anywhere else. It knows their world. It remembers what happened
   three chapters ago. It catches their mistakes. It makes prep
   faster and richer.

2. **They VISIT it because** the wiki and world graph are
   magnetic. Seeing your own world — all its characters, places,
   factions, and connections — rendered as a living, explorable
   thing is inherently compelling. It is the reward for the work
   of creation.

Usage drives the workflow. Engagement drives the habit.

## Design Principles

### Conversation First

The chat is the primary interface. Everything else — documents,
panels, wiki — serves the conversation or is produced by it.

### Structure as Byproduct

The GM should never feel like they are "filling in a database."
Entities, relationships, and timeline events are extracted from
natural creative work — conversations, documents, play notes —
and maintained by the system.

### Work at Any Scope

The GM moves fluidly between campaign, chapter, and scene scope.
The system tracks where they are and adjusts context accordingly.

### Reconciliation Over Analysis

The most valuable thing the AI does is not "analyze your text for
entities." It is "you planned X, the players did Y, here is what
that means for Z." Plan-vs-reality reconciliation is the killer
feature of prep mode.

### World-Grounded Generation

The AI never generates content in isolation. Every generative
request triggers a world model scan: scope resolution, orphan
entity discovery, underconnected NPC identification, and
relationship gap analysis. The AI actively seeks opportunities
to connect new content to existing fabric, turning loose threads
into narrative connective tissue. A generic LLM gives you a
plausible inn; Imagineer gives you an inn that makes your world
more coherent.

### The World Grows Through Use

Every conversation, every prep session, every ingested document
makes the world model richer. The system should never feel empty
or cold. It should feel like it is learning the GM's world
alongside them.

### Explore Is a Reward

The wiki and graph are not utilitarian tools. They are the
tangible manifestation of the GM's creative work. They should
feel alive, beautiful, and worth visiting.
