<!--
  Imagineer - TTRPG Campaign Intelligence Platform

  Copyright (c) 2025 - 2026
  This software is released under The MIT License
-->

# Tagging System Design

This document records the validated design for inline tags
in Imagineer. The product owner produced these decisions
through a brainstorming session on March 11, 2026. All
decisions recorded here are binding design commitments.

## Overview

Inline tags are the primary mechanism for GM-initiated
asynchronous AI requests embedded in document prose. The GM
drops a tag mid-writing, keeps editing, and the AI resolves
the tag in the background. Results appear on the pinboard
for review, iteration, and acceptance.

Tags are natural language requests, not commands. A tag can
ask for anything: entity creation, questions, prose
generation, relationship changes, tone edits, or any other
world-building task.

## Tag Syntax and Detection

The `@...@` syntax uses bookended `@` symbols to delimit a
tag. The text between the symbols is a natural language
request to the AI.

### Detection Mechanism

A Tiptap extension (`InlineTagExtension`) watches for the
`@...@` pattern. Detection triggers when the GM types the
closing `@` symbol. The extension performs the following
steps:

- The extension scans backward from the cursor to find an
  unmatched opening `@`.
- The extension validates the content length against the
  user-configured maximum (default 250 characters).
- If the content exceeds the maximum, the extension ignores
  the tag and treats the text as literal content.
- The extension ignores `@` symbols inside code blocks and
  code spans.
- The extension emits a `tag-detected` event with the tag
  content and position.

### Tag Identity

Each detected tag receives a unique ID (UUID) generated
client-side at detection time. This ID ties together the
editor decoration, the conversation, and the pinboard card.

### Persistence in the Document

The tag text stays in the document as plain text with a
Tiptap decoration overlay rather than a custom node. The
markdown serialization is `@request text@` with no special
markup. The decoration tracks the tag's visual state.

### Live Typing Feedback

While the GM types inside an open tag (after the first `@`,
before the closing `@`), the editor provides visual
feedback:

- The text from the opening `@` to the cursor receives a
  subtle growing highlight (light amber, the same colour
  family as the Processing state but more transparent).
- As the GM approaches the maximum length limit, the
  highlight shifts to a warning colour (soft red) in the
  last 20 characters.
- If the maximum length is exceeded without a closing `@`,
  the highlight disappears. The tag is silently abandoned
  and the text is treated as literal content.

## Context Assembly

When a tag is detected, the client assembles a context
payload and sends the payload to the AI.

### Context Window

The context window uses a character-based radius around the
tag rather than a paragraph-based approach. Campaign notes
are messy (bullets, tables, scattered fragments) and do not
have clean paragraph boundaries.

- The window extends up to 500 characters backward from the
  opening `@`.
- The window extends up to 500 characters forward from the
  closing `@`.
- The tag content is always included in full regardless of
  length.
- Boundaries snap outward to the nearest word boundary or
  complete wiki link (if a `[[...]]` link would be sliced,
  the boundary extends to include the full link).
- The context radius is user-configurable (default 500
  characters per side).

### Additional Structured Context

The system includes structured data beyond the raw text
window:

- For every `[[Entity Name]]` found within the context
  window, the system fetches the entity's full data from
  the React Query cache and includes the data in the
  payload.
- All entities currently pinned on the pinboard are
  included in the payload.

## AI Resolution

Each tag generates a single LLM call with structured
output.

### Self-Classification

The AI receives a system prompt instructing the AI to
analyse the request in context, self-classify the intent
(entity creation, question, prose generation, relationship
changes, tone edit, or other), and return structured output.

### Structured Response

The AI returns a structured response containing the
following fields:

- `responseType` identifies what kind of result the
  response contains.
- `prose` contains any text or prose response.
- `proposedChanges` contains an array of structural changes
  (entities to create, relationships to add, or document
  edits to make).
- `documentEdits` contains an array of edit descriptors,
  each with `from`, `to`, and `newContent` fields
  describing edits to the document (edits can span wider
  than the tag itself).

### Streaming

Results stream back via SSE using the existing
`useSSEStream` hook.

## Tags Are Conversations

Each tag creates a micro-conversation using the existing
conversation infrastructure. This is the core architectural
decision: no new backend infrastructure is needed.

### Conversation Creation

When a tag is detected, the system creates a conversation
with the following attributes:

- `scopeType` matches the document's scope type (chapter,
  session, or similar).
- `scopeId` matches the document's scope ID.
- `origin` is set to `'tag'`, a new field distinguishing
  tag-initiated conversations from `'user'`-initiated chats
  and `'main'` document chats.
- `tagId` stores the client-generated UUID that links back
  to the editor decoration.
- `title` is auto-generated from the first 50 characters of
  the tag content.

### Conversation Types by Origin

The system supports three conversation origins:

- `main` conversations are created when the GM opens
  document chat or sends the first message. A main
  conversation is scoped to the document and persists for
  the document's lifetime.
- `tag` conversations are created when a tag is detected and
  sent for processing. A tag conversation shares the
  document's scope, is active while the tag is unresolved,
  and is shelved on resolution.
- `user` conversations are created when the GM initiates
  from a selection, an entity, or a pinboard card. A user
  conversation has variable scope and remains active until
  the GM shelves the conversation.

### Benefits

Reusing the conversation infrastructure provides several
advantages:

- Token tracking comes free because conversations already
  track `totalInputTokens` and `totalOutputTokens`.
- "Pull to Chat" is trivial because the conversation
  already exists; the chat panel opens and focuses the
  conversation.
- Persistence is handled because conversations are already
  stored in the database.
- History and shelving require only a status flag.

## Tag Lifecycle and Visual States

Tags progress through a defined state machine.

### State Table

| State | Trigger | Visual | Editable? |
|-----------|--------------|------------|-----------|
| Detected | Closing `@` | Brief flash | Briefly |
| Processing | AI streaming | Amber pulse | No |
| Responded | AI complete | Blue-green | No |
| Resolved | GM accepts | Replaced | N/A |
| Failed | Error/timeout | Red tint | Yes |

### State Descriptions

The Detected state occurs when the GM types the closing `@`
symbol. The tag displays a brief flash and then transitions
to the Processing state. The GM can briefly edit the tag
content before the system sends the request.

The Processing state begins when the system creates the
conversation and the AI begins streaming. The tag displays
a soft amber background with a gentle pulsing dot at the
leading edge. If the GM edits the tag text during this
state, the system cancels the current request (aborting the
SSE stream), reverts the tag to Detected, and fires a new
request with the updated content.

The Responded state indicates that the AI has returned a
result. The tag displays a soft blue-green background with
a subtle left border, and a result card appears on the
pinboard. A tag can cycle between Responded and Processing
multiple times through "Try Again" or chat iteration.

The Resolved state occurs when the GM accepts the result
bundle. The system replaces the tag and surrounding text
with the AI resolution, applies DraftMark highlights to the
new content, and inserts wiki links. The tag ceases to
exist in the document.

The Failed state occurs on error or timeout. The tag
displays a red-ish tint with a small error icon. The GM can
edit the tag text to retry.

### Important Terminology

"Responded" means the AI has returned a result. "Resolved"
means the GM has accepted and the work is done.

### Parallel Processing

Multiple tags process in parallel. Each tag is independent.
Multiple result cards can appear on the pinboard
simultaneously.

### Navigation

Tag processing follows a fire-and-forget model. Processing
continues server-side when the GM navigates away. When the
GM returns, results are waiting on the pinboard.

## Pinboard Result Card Actions

When a tag reaches the Responded state, a result card
appears on the pinboard. The GM has four available actions.

### Accept

The Accept action commits all proposed changes (document
edits and world model changes) as a single bundle. The
system resolves the tag and shelves the conversation.

### Try Again

The Try Again action sends the tag back for reprocessing.
The system appends a message to the tag's conversation
telling the AI to generate a different response.
Conversation history prevents the AI from repeating
previous responses. The tag returns to the Processing
state.

### Pull to Chat

The Pull to Chat action opens the tag's conversation in the
chat panel for directed iteration. The AI already has the
full conversation history. The GM can give specific
feedback (for example, "make the name more Cornish").

### Dismiss

The Dismiss action discards the result entirely. The system
removes the tag decoration from the document and strips the
`@` delimiters, leaving the inner text as plain prose. The
system shelves the conversation.

### Future Actions

The team may add more actions later based on usage
feedback.

## Resolution and Document Changes

When the GM accepts a tag result (the Accept action), a
single commit triggers multiple coordinated changes.

### Document Edits

The system applies the AI's proposed `documentEdits` to the
editor:

- The edits can replace the tag and surrounding text with
  rewritten prose.
- The edits can insert wiki links elsewhere in the document
  where a new entity is mentioned.
- The edits can modify other parts of the document for
  consistency.
- All changed text receives the DraftMark highlight.

### World Model Changes

Proposed structural changes (entity creation, relationship
wiring, and entity edits) flow through the existing
structural change pipeline. The system bundles all changes
as a single batch.

### Design for A, Architect for C

The default UI presents all changes as a single
"Accept All" / "Reject All" bundle. Under the hood, each
change is tracked individually in the `pendingChanges`
array (`WorkspaceContext`), so pivoting to individual
accept/dismiss per change is a UI-only switch if needed
later.

### Dismissal

When the GM dismisses a tag result, the system removes the
tag decoration and strips the `@` delimiters from the
document. The inner text remains as plain prose. The system
shelves the conversation.

## Chat Panel and Multi-Conversation UI

The workspace supports multiple simultaneous conversations.

### Chat Panel Layout

The chat panel displays conversations with the following
structure:

- The panel shows a list of active conversations as tabs or
  a compact list.
- The main document chat (if one exists) always appears
  first.
- Tag and user conversations are ordered by most recent
  activity.
- Shelved conversations collapse into a "History" section.
- Each conversation shows an origin icon (document, tag, or
  chat bubble), a title, and an unread indicator.

### Shelving

When a tag is resolved, the system automatically shelves
the tag's conversation. The GM can manually shelve any
conversation. Shelved conversations remain in the database
but are hidden from the active list. The GM can reopen
shelved conversations.

## Deliberate Review Integration

The deliberate review (whole-document analysis) interacts
with the tag system through a gate rule.

### Gate Rule

When the GM triggers a deliberate review, the system checks
for tags in the Processing state. If any are found, the
review displays "Waiting for N tag requests to complete..."
and starts automatically once all in-flight tags reach the
Responded or Failed state.

### Rationale

The review should see the document in a stable state with
no in-flight mutations. The gate rule eliminates conflict
between review results and tag results. The wait is
typically brief because tag resolutions are single focused
LLM calls.

## Token Cost Visibility

Cost display is calibrated to the type of interaction.

### Tags

Tags are quick and flow-embedded. The system shows the
token cost after resolution on the result card. The system
does not interrupt the writing flow with cost estimates.

### Deliberate Review

Deliberate reviews are intentional and expensive. The
system shows an estimated cost before kickoff and the
actual cost after completion.

### Running Total

A persistent counter in the workspace UI shows accumulated
token usage for the current document session. The counter
sums usage across all active conversations.

## User Settings

Two user-configurable settings control tag behaviour.

| Setting | Default | Description |
|---------------------|---------|---------------------------|
| `tagMaxLength` | 250 | Maximum characters between |
| | | `@` delimiters |
| `tagContextRadius` | 500 | Characters of surrounding |
| | | text per side |

The client caches settings locally. Changes take effect
immediately for new tags. In-flight tags use the settings
that were active when the system sent the request.

## Relationship to Existing Systems

The tagging system integrates with several existing
subsystems.

### Wiki Links

Wiki links (`[[...]]`) and tags are complementary, not
competing. Wiki links reference existing entities. Tags
request new AI work. Wiki links within a tag's context
window automatically pull entity data into the AI's
context, making wiki links a natural way for the GM to
provide the AI with relevant world state.

### Entity Name Detection

The `useInlineEntityDetection` hook runs independently of
the tag system. The hook detects known entity names in prose
and suggests wiki links. Tags do not interact with entity
detection directly, but entities created through tag
resolution will be detected by future runs of the hook.

### Structural Change Pipeline

Tag resolutions feed into the existing structural change
pipeline. Proposed entities and relationships flow through
the same Draft, Approve, Commit stages as all other
structural changes.

### DraftMark

AI-generated content inserted by tag resolution uses the
existing DraftMark system to highlight new and changed
text.
