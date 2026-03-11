<!--
  Imagineer - TTRPG Campaign Intelligence Platform

  Copyright (c) 2025 - 2026
  This software is released under The MIT License
-->

# Tagging System Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use
> superpowers:executing-plans to implement this plan
> task-by-task.

**Goal:** Implement the `@...@` inline tagging system
that lets GMs embed natural language AI requests in
document prose, with results appearing on the pinboard
for review, iteration, and acceptance.

**Architecture:** Tags are conversations. Each `@...@`
tag creates a micro-conversation via the existing
conversation infrastructure. The Tiptap editor detects
tags, assembles context (character window + entity
data), and streams results via SSE. Results appear as
pinboard cards with Accept, Try Again, Pull to Chat,
and Dismiss actions.

**Tech Stack:** React 18, TypeScript 5, MUI v5, Tiptap
3, React Query, existing SSE streaming hook, existing
conversation API.

**Design Document:**
`docs/plans/2026-03-11-tagging-system-design.md`

---

## Existing Code

Tasks 1 through 9 of the frontend implementation plan
have already been completed. The following files exist
and are available for use.

### Types and API Layer

The types file at `client/src/types/index.ts` defines
all shared types including `Conversation`, `Message`,
`StreamEvent`, `ScopeType`, `TokenUsage`, and all
entity-related types.

The conversation API client at
`client/src/api/conversations.ts` exports
`conversationsApi` with `list`, `get`, `create`, and
`getMessages` methods. The `ListConversationsParams`
and `CreateConversationInput` interfaces live in the
same file.

### React Query Hooks

The conversation hooks at
`client/src/hooks/useConversations.ts` export
`useConversations`, `useConversation`,
`useCreateConversation`, and `conversationKeys`.

The SSE streaming hook at
`client/src/hooks/useSSEStream.ts` exports
`useSSEStream`, which POSTs to the conversation
messages endpoint and parses SSE events
(`text_delta`, `tool_use`, `tool_result`, `done`,
`error`).

The entity hooks at
`client/src/hooks/useEntities.ts` export
`useEntities` and `useEntity`.

The inline entity detection hook at
`client/src/hooks/useInlineEntityDetection.ts`
performs client-side entity name detection with regex
and a 500ms debounce.

### Workspace Infrastructure

The workspace context at
`client/src/contexts/WorkspaceContext.tsx` exports
`WorkspaceProvider` and `useWorkspace` with selection,
pinboard, pending changes, and chat state management.

The workspace editor at
`client/src/components/Editor/WorkspaceEditor.tsx`
integrates Tiptap with `SelectionTracker`,
`DraftMark`, debounced content changes, and auto-save.

The workspace page at
`client/src/components/Workspace/WorkspacePage.tsx`
renders the three-layer workspace with pinboard,
editor, and chat areas.

### Editor Extensions

The selection tracker at
`client/src/components/Editor/SelectionTracker.ts` is
a Tiptap extension that reports selection state to the
workspace context.

The draft mark at
`client/src/components/Editor/DraftMark.ts` is a
Tiptap mark that visually distinguishes AI-generated
content.

The wiki link node at
`client/src/components/MarkdownEditor/WikiLinkNode.ts`
is a Tiptap node that renders `[[wiki links]]`.

### Patterns and Conventions

The copyright header uses the `// ---` line-comment
style, not `/*...*/` block comments.

All source files use four-space indentation.

React Query hooks follow the query key factory pattern
with cache invalidation via predicate.

SSE streaming uses POST-based fetch with
`ReadableStream`.

Tests use Vitest and `@testing-library/react`. Run
tests from the `client/` directory with the following
command:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/ \
    .worktrees/frontend-rebuild/client \
    && npx vitest run <path>
```

---

## Task 1: Conversation Origin Type Extension

This task extends the conversation types and API to
support the `origin` field that the tagging system
requires.

### Files

The following files require changes:

- Modify: `client/src/types/index.ts` to add
  `ConversationOrigin` type and new fields on the
  `Conversation` interface.
- Modify: `client/src/api/conversations.ts` to add
  `origin` and `tagId` to `CreateConversationInput`,
  add `shelved` filter to `ListConversationsParams`,
  and add a `shelveConversation` method.
- Modify: `client/src/hooks/useConversations.ts` to
  add a `useShelveConversation` mutation hook.
- Create:
  `client/src/hooks/__tests__/useConversations.shelve.test.tsx`
  for the new shelve functionality.

### Step 1: Write failing tests for shelve mutation

Create the test file at
`client/src/hooks/__tests__/useConversations.shelve.test.tsx`.
The tests verify the following behaviours:

- The shelve mutation calls `PATCH` on the endpoint
  `/campaigns/:campaignId/conversations/:conversationId/shelve`.
- Shelving a conversation invalidates the
  conversation list cache.
- `ListConversationsParams` accepts a `shelved`
  boolean filter.

Run the tests to verify they fail:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/ \
    .worktrees/frontend-rebuild/client \
    && npx vitest run \
    src/hooks/__tests__/useConversations.shelve.test.tsx
```

The expected output shows all three tests failing
because the `shelveConversation` method and
`useShelveConversation` hook do not exist yet.

### Step 2: Add ConversationOrigin type

Add the following type to `client/src/types/index.ts`
after the existing `ScopeType` definition:

```typescript
export type ConversationOrigin =
    | 'main'
    | 'tag'
    | 'user';
```

Add three new fields to the `Conversation` interface:

```typescript
export interface Conversation {
    id: number;
    campaignId: number;
    scopeType: ScopeType;
    scopeId: number;
    title: string;
    summary?: string;
    messageCount: number;
    totalInputTokens: number;
    totalOutputTokens: number;
    origin?: ConversationOrigin;
    tagId?: string;
    shelved?: boolean;
    createdAt: string;
    updatedAt: string;
}
```

### Step 3: Update API client

Add `origin`, `tagId`, and `shelved` to the input and
filter types in `client/src/api/conversations.ts`:

```typescript
export interface ListConversationsParams {
    campaignId: number;
    scopeType?: ScopeType;
    scopeId?: number;
    shelved?: boolean;
    page?: number;
    pageSize?: number;
}

export interface CreateConversationInput {
    scopeType: ScopeType;
    scopeId: number;
    title?: string;
    origin?: ConversationOrigin;
    tagId?: string;
}
```

Add the `shelveConversation` method to the
`conversationsApi` object:

```typescript
shelve(
    campaignId: number,
    conversationId: number,
): Promise<Conversation> {
    return apiClient.patch<Conversation>(
        `/campaigns/${campaignId}`
            + `/conversations/${conversationId}`
            + `/shelve`,
    );
},
```

### Step 4: Add useShelveConversation hook

Add the following hook to
`client/src/hooks/useConversations.ts`:

```typescript
export function useShelveConversation() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({
            campaignId,
            conversationId,
        }: {
            campaignId: number;
            conversationId: number;
        }) => conversationsApi.shelve(
            campaignId,
            conversationId,
        ),
        onSuccess: (_data, variables) => {
            queryClient.invalidateQueries({
                queryKey: conversationKeys.lists(),
                predicate: (query) => {
                    const key =
                        query.queryKey as unknown[];
                    return (
                        key.length >= 3
                        && key[2]
                            === variables.campaignId
                    );
                },
            });
        },
    });
}
```

### Step 5: Run tests and verify they pass

```bash
cd /Users/antonypegg/PROJECTS/imagineer/ \
    .worktrees/frontend-rebuild/client \
    && npx vitest run \
    src/hooks/__tests__/useConversations.shelve.test.tsx
```

The expected output shows all three tests passing.

### Commit

```
feat: add conversation origin type and shelve support
```

---

## Task 2: User Settings Hook

This task creates a hook for user-configurable tag
settings stored in localStorage.

### Files

The following files require changes:

- Create: `client/src/hooks/useTagSettings.ts` with
  the hook that reads and writes tag settings.
- Create:
  `client/src/hooks/__tests__/useTagSettings.test.ts`
  with tests for the hook.

### Step 1: Write failing tests

Create the test file at
`client/src/hooks/__tests__/useTagSettings.test.ts`.
The tests verify the following behaviours:

- The hook returns default values when localStorage
  is empty: `tagMaxLength` is 250 and
  `tagContextRadius` is 500.
- The `updateSettings` function persists updated
  values to localStorage under the key
  `imagineer_tag_settings`.
- The hook reads previously saved values from
  localStorage on mount.

Run the tests to verify they fail:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/ \
    .worktrees/frontend-rebuild/client \
    && npx vitest run \
    src/hooks/__tests__/useTagSettings.test.ts
```

The expected output shows all three tests failing
because the hook does not exist.

### Step 2: Implement useTagSettings

Create `client/src/hooks/useTagSettings.ts` with the
following implementation:

```typescript
// -------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------

import { useState, useCallback } from 'react';

const STORAGE_KEY = 'imagineer_tag_settings';

export interface TagSettings {
    tagMaxLength: number;
    tagContextRadius: number;
}

const DEFAULT_SETTINGS: TagSettings = {
    tagMaxLength: 250,
    tagContextRadius: 500,
};

function loadSettings(): TagSettings {
    try {
        const raw = localStorage.getItem(STORAGE_KEY);
        if (!raw) {
            return { ...DEFAULT_SETTINGS };
        }
        const parsed = JSON.parse(raw) as Partial<
            TagSettings
        >;
        return {
            tagMaxLength:
                parsed.tagMaxLength
                    ?? DEFAULT_SETTINGS.tagMaxLength,
            tagContextRadius:
                parsed.tagContextRadius
                    ?? DEFAULT_SETTINGS
                        .tagContextRadius,
        };
    } catch {
        return { ...DEFAULT_SETTINGS };
    }
}

export function useTagSettings(): {
    settings: TagSettings;
    updateSettings: (
        updates: Partial<TagSettings>,
    ) => void;
} {
    const [settings, setSettings] = useState<
        TagSettings
    >(loadSettings);

    const updateSettings = useCallback(
        (updates: Partial<TagSettings>) => {
            setSettings((prev) => {
                const next = { ...prev, ...updates };
                localStorage.setItem(
                    STORAGE_KEY,
                    JSON.stringify(next),
                );
                return next;
            });
        },
        [],
    );

    return { settings, updateSettings };
}
```

### Step 3: Run tests and verify they pass

```bash
cd /Users/antonypegg/PROJECTS/imagineer/ \
    .worktrees/frontend-rebuild/client \
    && npx vitest run \
    src/hooks/__tests__/useTagSettings.test.ts
```

The expected output shows all three tests passing.

### Commit

```
feat: add useTagSettings hook for tag configuration
```

---

## Task 3: InlineTagExtension (Detection and Decoration)

This task creates the core Tiptap extension that
detects `@...@` tags and manages visual decorations.
The extension is the most complex piece of the tagging
system.

### Files

The following files require changes:

- Create:
  `client/src/components/Editor/InlineTagExtension.ts`
  with the Tiptap extension.
- Create:
  `client/src/components/Editor/__tests__/InlineTagExtension.test.ts`
  with tests for detection and decoration.

### Step 1: Write failing tests

Create the test file at
`client/src/components/Editor/__tests__/InlineTagExtension.test.ts`.
The tests verify the following behaviours:

- Typing `@need an NPC@` triggers `onTagDetected`
  with the correct content (`need an NPC`) and
  position.
- The tag content does not include the `@`
  delimiters.
- Tags exceeding `maxLength` are not detected.
- Multiple tags in the same document each trigger
  separate callbacks.
- Each detected tag has a unique UUID `id`.
- The extension ignores `@` symbols inside code
  blocks (between backticks).
- An opening `@` without a closing `@` shows a
  growing decoration.
- The decoration disappears when `maxLength` is
  exceeded without a closing `@`.
- The `updateTagState` command changes the tag's
  decoration CSS class.
- The `removeTag` command removes tag tracking and
  the decoration.

Run the tests to verify they fail:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/ \
    .worktrees/frontend-rebuild/client \
    && npx vitest run \
    src/components/Editor/__tests__/InlineTagExtension.test.ts
```

The expected output shows all ten tests failing.

### Step 2: Define types

Add the following types at the top of the extension
file:

```typescript
export type TagState =
    | 'detected'
    | 'processing'
    | 'responded'
    | 'failed';

export interface TrackedTag {
    id: string;
    content: string;
    position: { from: number; to: number };
    state: TagState;
}

export interface InlineTagOptions {
    maxLength: number;
    onTagDetected: (tag: {
        id: string;
        content: string;
        position: { from: number; to: number };
    }) => void;
}
```

### Step 3: Implement the extension

Create
`client/src/components/Editor/InlineTagExtension.ts`
as a Tiptap `Extension` using `Extension.create<InlineTagOptions>`.
The extension must implement the following:

1. A ProseMirror Plugin with a `StateField` that
   tracks two pieces of state: a `Map<string,
   TrackedTag>` for all tracked tags and an optional
   `{ from: number }` for any in-progress tag (an
   opening `@` without a closing `@`).

2. A `DecorationSet` derived from the plugin state
   that applies CSS classes based on tag state:

   - `tag-processing` with an amber background for
     the Processing state.
   - `tag-responded` with a blue-green background
     for the Responded state.
   - `tag-failed` with a red tint for the Failed
     state.
   - `tag-typing` with a light amber background for
     in-progress tags.
   - `tag-typing-warning` with a soft red background
     for in-progress tags within 20 characters of
     `maxLength`.

3. An `appendTransaction` handler that scans the
   document text after each transaction, looking for
   `@...@` patterns. When the handler finds a new
   completed tag (one that does not already have a
   tracking entry), the handler generates a UUID,
   creates a `TrackedTag` entry, and calls
   `onTagDetected`. The handler must ignore `@`
   symbols inside code marks.

4. Two editor commands:

   - `updateTagState(tagId: string, state: TagState)`
     updates the tracked tag's state and the
     corresponding decoration.
   - `removeTag(tagId: string)` removes the tag from
     the tracking map and removes the decoration.

### Step 4: Run tests and verify they pass

```bash
cd /Users/antonypegg/PROJECTS/imagineer/ \
    .worktrees/frontend-rebuild/client \
    && npx vitest run \
    src/components/Editor/__tests__/InlineTagExtension.test.ts
```

The expected output shows all ten tests passing.

### Commit

```
feat: add InlineTagExtension for @...@ tag detection
```

---

## Task 4: Context Assembly Utility

This task creates a pure function that assembles the
context payload for a detected tag.

### Files

The following files require changes:

- Create: `client/src/utils/tagContext.ts` with the
  assembly function.
- Create:
  `client/src/utils/__tests__/tagContext.test.ts`
  with tests for the function.

### Step 1: Write failing tests

Create the test file at
`client/src/utils/__tests__/tagContext.test.ts`. The
tests verify the following behaviours:

- The function extracts the correct radius of text
  around the tag position.
- Boundaries snap outward to word boundaries (not
  mid-word).
- Boundaries snap outward to include complete
  `[[wiki links]]` that would be sliced at the
  boundary.
- The function returns the full tag content without
  `@` delimiters.
- The function handles a tag near the start of a
  document (boundary at position 0).
- The function handles a tag near the end of a
  document (boundary at the document length).
- The function extracts wiki link entity names from
  the context window.
- The function passes through `pinnedEntityIds`
  unchanged.

Run the tests to verify they fail:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/ \
    .worktrees/frontend-rebuild/client \
    && npx vitest run \
    src/utils/__tests__/tagContext.test.ts
```

The expected output shows all eight tests failing.

### Step 2: Implement assembleTagContext

Create `client/src/utils/tagContext.ts` with the
following function signature:

```typescript
export interface TagContext {
    tagContent: string;
    surroundingText: string;
    wikiLinkedEntityIds: number[];
    pinnedEntityIds: number[];
}

export function assembleTagContext(
    documentContent: string,
    tagPosition: { from: number; to: number },
    contextRadius: number,
    pinnedEntityIds: number[],
): TagContext
```

The function performs the following steps:

1. Calculate raw boundaries as
   `max(0, tagPosition.from - contextRadius)` and
   `min(documentContent.length, tagPosition.to + contextRadius)`.
2. Snap the start boundary backward to the nearest
   whitespace character (or position 0).
3. Snap the end boundary forward to the nearest
   whitespace character (or the document length).
4. If a `[[` or `]]` sequence would be sliced at
   either boundary, extend the boundary outward to
   include the complete `[[...]]` link.
5. Extract the text between the snapped boundaries
   as `surroundingText`.
6. Extract the tag content from the document between
   `tagPosition.from` and `tagPosition.to`.
7. Scan `surroundingText` for `[[...]]` patterns
   and collect the entity names into an array.
8. Return the assembled `TagContext` with
   `pinnedEntityIds` passed through.

Note: The `wikiLinkedEntityIds` field contains the
names as strings for this function. The calling hook
resolves names to IDs using the entity cache.
Rename the field to `wikiLinkedEntityNames` and use
`string[]` instead of `number[]`:

```typescript
export interface TagContext {
    tagContent: string;
    surroundingText: string;
    wikiLinkedEntityNames: string[];
    pinnedEntityIds: number[];
}
```

### Step 3: Run tests and verify they pass

```bash
cd /Users/antonypegg/PROJECTS/imagineer/ \
    .worktrees/frontend-rebuild/client \
    && npx vitest run \
    src/utils/__tests__/tagContext.test.ts
```

The expected output shows all eight tests passing.

### Commit

```
feat: add assembleTagContext utility for tag payloads
```

---

## Task 5: Tag Processing Hook (useTagProcessor)

This task creates the orchestration hook that connects
tag detection to conversation creation and SSE
streaming.

### Files

The following files require changes:

- Create: `client/src/hooks/useTagProcessor.ts` with
  the orchestration hook.
- Create:
  `client/src/hooks/__tests__/useTagProcessor.test.ts`
  with tests for the hook.

### Step 1: Write failing tests

Create the test file at
`client/src/hooks/__tests__/useTagProcessor.test.ts`.
Mock `conversationsApi` and `useSSEStream`. The tests
verify the following behaviours:

- `processTag` creates a conversation with
  `origin: 'tag'` and the tag's UUID as `tagId`.
- `processTag` calls `assembleTagContext` with the
  correct parameters (document content, tag
  position, context radius, pinned entity IDs).
- `processTag` sends the assembled context as the
  first message via `useSSEStream.sendMessage`.
- The tag state progresses from `detected` to
  `processing` to `responded`.
- On an SSE error event, the tag state becomes
  `failed` with the error message.
- `retryTag` sends a new message to the same
  conversation requesting a different response.
- `dismissTag` shelves the conversation and removes
  the tag from tracking.
- `acceptTag` shelves the conversation and returns
  the proposed changes and document edits.
- Multiple tags can process in parallel without
  interfering with each other.

Run the tests to verify they fail:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/ \
    .worktrees/frontend-rebuild/client \
    && npx vitest run \
    src/hooks/__tests__/useTagProcessor.test.ts
```

The expected output shows all nine tests failing.

### Step 2: Define types

Add the following types at the top of the hook file:

```typescript
export interface TagProcessorOptions {
    campaignId: number;
    scopeType: ScopeType;
    scopeId: number;
    pinnedEntityIds: number[];
    entities: Entity[];
    documentContent: string;
    contextRadius: number;
}

export interface ProcessedTag {
    tagId: string;
    conversationId: number | null;
    state: TagState;
    result: TagResult | null;
    error: string | null;
}

export interface TagResult {
    responseType: string;
    prose?: string;
    proposedChanges?: ProposedChange[];
    documentEdits?: DocumentEdit[];
}

export interface ProposedChange {
    type:
        | 'create_entity'
        | 'create_relationship'
        | 'edit_entity';
    data: Record<string, unknown>;
    description: string;
}

export interface DocumentEdit {
    from: number;
    to: number;
    newContent: string;
}
```

### Step 3: Implement useTagProcessor

Create `client/src/hooks/useTagProcessor.ts` with the
following return type:

```typescript
export function useTagProcessor(
    options: TagProcessorOptions,
): {
    processTag: (tag: {
        id: string;
        content: string;
        position: { from: number; to: number };
    }) => void;
    retryTag: (tagId: string) => void;
    dismissTag: (tagId: string) => void;
    acceptTag: (tagId: string) => void;
    tags: Map<string, ProcessedTag>;
}
```

The hook performs the following orchestration:

1. `processTag` receives a detected tag from the
   `InlineTagExtension`. The function calls
   `assembleTagContext` to build the context payload.
   The function resolves wiki link entity names to
   IDs using the `entities` array from options. The
   function creates a conversation via
   `useCreateConversation` with `origin: 'tag'` and
   `tagId` set to the tag's UUID. The function then
   calls `useSSEStream.sendMessage` to send the
   assembled context as the first message. The
   function updates the `ProcessedTag` entry's state
   to `processing`.

2. As SSE events arrive, the hook updates the
   `ProcessedTag` state. On a `done` event, the hook
   parses the accumulated text as JSON to extract
   `TagResult` fields and sets the state to
   `responded`. On an `error` event, the hook sets
   the state to `failed` with the error message.

3. `retryTag` sends a "try again, generate a
   different response" message to the existing
   conversation. The hook resets the tag state to
   `processing`.

4. `dismissTag` calls `useShelveConversation` on the
   tag's conversation and removes the `ProcessedTag`
   entry from the map.

5. `acceptTag` calls `useShelveConversation` on the
   tag's conversation and returns the result's
   proposed changes and document edits for the
   caller to apply.

### Step 4: Run tests and verify they pass

```bash
cd /Users/antonypegg/PROJECTS/imagineer/ \
    .worktrees/frontend-rebuild/client \
    && npx vitest run \
    src/hooks/__tests__/useTagProcessor.test.ts
```

The expected output shows all nine tests passing.

### Commit

```
feat: add useTagProcessor hook for tag orchestration
```

---

## Task 6: Chat Panel with Multi-Conversation Support

This task builds the chat panel with tabs for multiple
simultaneous conversations (main, tag, and user
conversations). The task is large and breaks into four
sub-commits.

### Files

The following files require changes:

- Create:
  `client/src/components/Chat/ChatMessage.tsx`
  for rendering individual messages.
- Create:
  `client/src/components/Chat/StreamingMessage.tsx`
  for showing in-progress streaming text.
- Create:
  `client/src/components/Chat/ToolUseIndicator.tsx`
  for showing tool use chips.
- Create:
  `client/src/components/Chat/SelectionPreview.tsx`
  for showing selected editor text as context.
- Create:
  `client/src/components/Chat/ChatInput.tsx`
  for the message input field.
- Create:
  `client/src/components/Chat/ConversationTabs.tsx`
  for the conversation tab list.
- Create:
  `client/src/components/Chat/ChatPanel.tsx`
  for the top-level chat panel orchestration.
- Create:
  `client/src/components/Chat/__tests__/ChatMessage.test.tsx`
  with tests for message rendering.
- Create:
  `client/src/components/Chat/__tests__/ChatInput.test.tsx`
  with tests for input behaviour.
- Create:
  `client/src/components/Chat/__tests__/ConversationTabs.test.tsx`
  with tests for tab management.
- Create:
  `client/src/components/Chat/__tests__/SelectionPreview.test.tsx`
  with tests for selection display.
- Create:
  `client/src/components/Chat/__tests__/ChatPanel.test.tsx`
  with tests for panel orchestration.

### Sub-task 6a: ChatMessage, StreamingMessage, and ToolUseIndicator

Write failing tests, then implement.

`ChatMessage` renders user and assistant messages with
markdown formatting. User messages appear right-aligned
with a primary colour background. Assistant messages
appear left-aligned with a neutral background.

`StreamingMessage` shows accumulated text with a
blinking cursor indicator at the end. The component
receives `text` and `isStreaming` props.

`ToolUseIndicator` shows a small MUI `Chip` with the
tool name. The component receives a `toolName` string
prop.

Tests for sub-task 6a verify the following:

- `ChatMessage` renders user messages with the
  correct alignment.
- `ChatMessage` renders assistant messages with
  markdown formatting.
- `StreamingMessage` shows the blinking cursor when
  `isStreaming` is true.
- `StreamingMessage` hides the cursor when
  `isStreaming` is false.
- `ToolUseIndicator` renders a chip with the tool
  name.

Run tests:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/ \
    .worktrees/frontend-rebuild/client \
    && npx vitest run \
    src/components/Chat/__tests__/ChatMessage.test.tsx
```

Commit:

```
feat: add ChatMessage, StreamingMessage, ToolUseIndicator
```

### Sub-task 6b: SelectionPreview and ChatInput

Write failing tests, then implement.

`SelectionPreview` shows the currently selected editor
text as a compact chip above the input field. The
component reads `selectedText` from the workspace
context. When no text is selected, the component
renders nothing.

`ChatInput` provides a text field with a send button.
Pressing Enter sends the message (Shift+Enter inserts
a newline). The input disables the send button while
streaming. The component accepts `onSend` and
`isStreaming` props.

Tests for sub-task 6b verify the following:

- `SelectionPreview` renders selected text as a
  chip.
- `SelectionPreview` renders nothing when no text is
  selected.
- `ChatInput` calls `onSend` when the user presses
  Enter.
- `ChatInput` does not call `onSend` on
  Shift+Enter.
- `ChatInput` disables the send button when
  `isStreaming` is true.

Run tests:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/ \
    .worktrees/frontend-rebuild/client \
    && npx vitest run \
    src/components/Chat/__tests__/ChatInput.test.tsx \
    src/components/Chat/__tests__/SelectionPreview.test.tsx
```

Commit:

```
feat: add SelectionPreview and ChatInput components
```

### Sub-task 6c: ConversationTabs

Write failing tests, then implement.

`ConversationTabs` renders a horizontal tab list of
active conversations. The main chat tab always appears
first. Tag conversations display a tag icon. User
conversations display a chat bubble icon. Shelved
conversations collapse into a "History" section. Each
tab shows an unread indicator when new messages arrive.

Props:

```typescript
interface ConversationTabsProps {
    conversations: Conversation[];
    activeConversationId: number | null;
    onSelectConversation: (id: number) => void;
}
```

Tests for sub-task 6c verify the following:

- The main conversation tab renders first.
- Tag conversations show a tag icon.
- Shelved conversations appear in a collapsible
  history section.
- Clicking a tab calls `onSelectConversation` with
  the conversation ID.
- The active tab receives visual emphasis.

Run tests:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/ \
    .worktrees/frontend-rebuild/client \
    && npx vitest run \
    src/components/Chat/__tests__/ConversationTabs.test.tsx
```

Commit:

```
feat: add ConversationTabs for multi-conversation UI
```

### Sub-task 6d: ChatPanel (orchestration)

Write failing tests, then implement.

`ChatPanel` is the top-level chat component that
integrates `ConversationTabs`, the message list,
`StreamingMessage`, and `ChatInput`. The panel
collapses at the bottom of the workspace. The panel
creates and fetches conversations per scope using the
existing hooks. The panel uses `useSSEStream` for
message sending. A resize handle at the top edge lets
the user adjust the panel height.

Props:

```typescript
interface ChatPanelProps {
    campaignId: number;
    scopeType: ScopeType;
    scopeId: number;
}
```

Tests for sub-task 6d verify the following:

- The panel renders `ConversationTabs` and the
  message area.
- Sending a message calls `useSSEStream.sendMessage`.
- Switching tabs loads messages for the selected
  conversation.
- The resize handle adjusts the panel height.

Run tests:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/ \
    .worktrees/frontend-rebuild/client \
    && npx vitest run \
    src/components/Chat/__tests__/ChatPanel.test.tsx
```

Commit:

```
feat: add ChatPanel with multi-conversation support
```

---

## Task 7: Token Cost Display

This task builds token cost display for individual
messages and a running total across all active
conversations.

### Files

The following files require changes:

- Create:
  `client/src/components/Chat/TokenCostBadge.tsx`
  for per-message token display.
- Create:
  `client/src/components/Chat/TokenRunningTotal.tsx`
  for the running total in the workspace header.
- Create:
  `client/src/components/Chat/__tests__/TokenCostBadge.test.tsx`
  with tests for formatting.
- Create:
  `client/src/components/Chat/__tests__/TokenRunningTotal.test.tsx`
  with tests for sum calculation.

### Step 1: Write failing tests for TokenCostBadge

The tests verify the following behaviours:

- The badge renders "~450 tokens" for a count of
  450.
- The badge renders "~1.2k tokens" for a count of
  1200.
- The badge renders nothing when the count is zero.
- The badge uses caption typography with muted
  colour.

Run the tests:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/ \
    .worktrees/frontend-rebuild/client \
    && npx vitest run \
    src/components/Chat/__tests__/TokenCostBadge.test.tsx
```

### Step 2: Implement TokenCostBadge

```typescript
interface TokenCostBadgeProps {
    tokenCount: number;
}
```

The component formats the count:

- Below 1000: `~{count} tokens`
- 1000 and above: `~{(count / 1000).toFixed(1)}k tokens`
- Zero: return `null`

Use MUI `Typography` with `variant="caption"` and
`color="text.secondary"`.

### Step 3: Write failing tests for TokenRunningTotal

The tests verify the following behaviours:

- The component sums `totalInputTokens` and
  `totalOutputTokens` across all conversations.
- The component renders the formatted total.
- The component renders nothing when the total is
  zero.

### Step 4: Implement TokenRunningTotal

```typescript
interface TokenRunningTotalProps {
    conversations: Conversation[];
}
```

The component computes the sum of
`totalInputTokens + totalOutputTokens` across all
provided conversations and renders the result using
`TokenCostBadge`.

### Step 5: Run all token tests

```bash
cd /Users/antonypegg/PROJECTS/imagineer/ \
    .worktrees/frontend-rebuild/client \
    && npx vitest run \
    src/components/Chat/__tests__/TokenCostBadge.test.tsx \
    src/components/Chat/__tests__/TokenRunningTotal.test.tsx
```

The expected output shows all tests passing.

### Commit

```
feat: add token cost badge and running total display
```

---

## Task 8: Tag Result Card

This task builds the pinboard card that appears when a
tag reaches the Responded state.

### Files

The following files require changes:

- Create:
  `client/src/components/Pinboard/TagResultCard.tsx`
  with the card component.
- Create:
  `client/src/components/Pinboard/__tests__/TagResultCard.test.tsx`
  with tests for rendering and actions.

### Step 1: Write failing tests

The tests verify the following behaviours:

- The card renders the tag content as a header.
- The card renders the AI prose response.
- The card renders a summary list of proposed
  changes.
- Clicking Accept calls `onAccept`.
- Clicking Try Again calls `onTryAgain`.
- Clicking Pull to Chat calls `onPullToChat`.
- Clicking Dismiss calls `onDismiss`.
- The Processing state shows a loading spinner and
  no action buttons except cancel.
- The Failed state shows the error message and a
  retry button.
- The card shows a `TokenCostBadge` with the
  conversation's token usage.

Run the tests:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/ \
    .worktrees/frontend-rebuild/client \
    && npx vitest run \
    src/components/Pinboard/__tests__/TagResultCard.test.tsx
```

The expected output shows all ten tests failing.

### Step 2: Implement TagResultCard

Define the component props:

```typescript
interface TagResultCardProps {
    tag: ProcessedTag;
    onAccept: () => void;
    onTryAgain: () => void;
    onPullToChat: () => void;
    onDismiss: () => void;
}
```

The component renders an MUI `Card` with the
following structure:

- `CardHeader` with the tag content (first 80
  characters) as the title.
- `CardContent` with the AI prose response (when
  state is `responded`).
- A list of proposed changes showing the
  `description` field of each `ProposedChange`.
- `TokenCostBadge` in the footer area.
- `CardActions` with four buttons: Accept, Try
  Again, Pull to Chat, and Dismiss (when state is
  `responded`).
- When state is `processing`, the content area shows
  a `CircularProgress` spinner.
- When state is `failed`, the content area shows the
  error message and a single "Retry" button.

### Step 3: Run tests and verify they pass

```bash
cd /Users/antonypegg/PROJECTS/imagineer/ \
    .worktrees/frontend-rebuild/client \
    && npx vitest run \
    src/components/Pinboard/__tests__/TagResultCard.test.tsx
```

The expected output shows all ten tests passing.

### Commit

```
feat: add TagResultCard for pinboard tag results
```

---

## Task 9: Deliberate Review with Tag Gate

This task builds the deliberate review interface that
gates on in-flight tags before submitting the full
document for AI review.

### Files

The following files require changes:

- Create:
  `client/src/components/Chat/ReviewRequest.tsx`
  with the review component.
- Create:
  `client/src/components/Chat/__tests__/ReviewRequest.test.tsx`
  with tests for the review flow.

### Step 1: Write failing tests

The tests verify the following behaviours:

- The Review button renders in the chat panel
  header.
- Clicking the Review button opens a confirmation
  dialog showing the document word count and
  estimated token cost.
- The dialog shows a waiting state when any
  `ProcessedTag` has state `processing`.
- The review auto-starts when all processing tags
  complete (transition from waiting to ready).
- Cancelling the dialog does not send a review
  request.
- On confirmation, the component sends the full
  document content as a review conversation message.
- Review findings are dispatched as pending changes
  to `WorkspaceContext`.

Run the tests:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/ \
    .worktrees/frontend-rebuild/client \
    && npx vitest run \
    src/components/Chat/__tests__/ReviewRequest.test.tsx
```

The expected output shows all seven tests failing.

### Step 2: Implement ReviewRequest

Define the component props:

```typescript
interface ReviewRequestProps {
    campaignId: number;
    conversationId: number | null;
    documentContent: string;
    processingTags: ProcessedTag[];
}
```

The component implements the following behaviour:

1. A "Review" button renders in the chat header
   area.
2. Clicking the button opens an MUI `Dialog` with
   the document word count (split on whitespace) and
   an estimated token cost (word count multiplied by
   1.3, rounded up).
3. If any tag in `processingTags` has state
   `processing`, the dialog shows "Waiting for N tag
   requests to complete..." instead of the confirm
   button.
4. The component uses a `useEffect` to watch
   `processingTags`. When the count of processing
   tags drops to zero and the dialog was previously
   in the waiting state, the component automatically
   sends the review request.
5. The confirm button sends the full document
   content as a message to the review conversation
   via `useSSEStream`.
6. On stream completion, the component parses the
   response for structured findings and dispatches
   each finding as a `PendingChange` to the
   workspace context via `addPendingChange`.

### Step 3: Run tests and verify they pass

```bash
cd /Users/antonypegg/PROJECTS/imagineer/ \
    .worktrees/frontend-rebuild/client \
    && npx vitest run \
    src/components/Chat/__tests__/ReviewRequest.test.tsx
```

The expected output shows all seven tests passing.

### Commit

```
feat: add deliberate review with tag gate
```

---

## Task 10: Integration Wiring

This task wires the tagging system into the
`WorkspacePage` and `WorkspaceEditor`, connecting all
the pieces built in Tasks 1 through 9.

### Files

The following files require changes:

- Modify:
  `client/src/components/Editor/WorkspaceEditor.tsx`
  to add `InlineTagExtension` to the extensions list
  and wire the `onTagDetected` callback.
- Modify:
  `client/src/components/Workspace/WorkspacePage.tsx`
  to add `useTagProcessor`, wire tag processing to
  the pinboard and chat panel.
- Modify: `client/src/components/Editor/index.ts` to
  export `InlineTagExtension`.
- Create:
  `client/src/components/Workspace/__tests__/TagIntegration.test.tsx`
  with integration tests.

### Step 1: Write failing integration tests

Create the test file at
`client/src/components/Workspace/__tests__/TagIntegration.test.tsx`.
The tests verify the following end-to-end behaviours:

- Tag detection in the editor triggers tag
  processing via `useTagProcessor`.
- A processed tag result appears as a
  `TagResultCard` on the pinboard.
- The Accept action applies document edits to the
  editor.
- The Dismiss action removes the tag decoration from
  the editor.
- Pull to Chat opens the chat panel focused on the
  tag's conversation.

Run the tests:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/ \
    .worktrees/frontend-rebuild/client \
    && npx vitest run \
    src/components/Workspace/__tests__/TagIntegration.test.tsx
```

The expected output shows all five tests failing.

### Step 2: Add InlineTagExtension to WorkspaceEditor

Modify `client/src/components/Editor/WorkspaceEditor.tsx`
to accept a new optional prop:

```typescript
export interface WorkspaceEditorProps {
    initialContent: string;
    onContentChange: (content: string) => void;
    onSave: (content: string) => Promise<void>;
    readOnly?: boolean;
    onTagDetected?: (tag: {
        id: string;
        content: string;
        position: { from: number; to: number };
    }) => void;
    tagMaxLength?: number;
}
```

Add `InlineTagExtension` to the extensions array,
configured with `maxLength` from `tagMaxLength` (or
the default of 250) and `onTagDetected` from the prop.
Only add the extension when `onTagDetected` is
provided.

### Step 3: Wire useTagProcessor in WorkspacePage

Modify
`client/src/components/Workspace/WorkspacePage.tsx` to
perform the following:

1. Import and call `useTagProcessor` with the
   current campaign ID, scope type, scope ID, pinned
   entity IDs, entities from `useEntities`, document
   content, and context radius from `useTagSettings`.
2. Pass `processTag` from the hook as the
   `onTagDetected` prop to `WorkspaceEditor`.
3. Pass `tagMaxLength` from `useTagSettings` as the
   `tagMaxLength` prop to `WorkspaceEditor`.
4. Replace the pinboard placeholder with a mapping
   over the `tags` map from `useTagProcessor`.
   Render a `TagResultCard` for each tag in the
   `responded` or `processing` state.
5. Wire the Accept action to call `acceptTag`, which
   applies document edits via the editor's
   `commands` and adds proposed changes to the
   workspace context via `addPendingChange`.
6. Wire the Dismiss action to call `dismissTag` and
   then call `editor.commands.removeTag(tagId)`.
7. Wire Pull to Chat to call `setChatVisible(true)`
   and switch to the tag's conversation tab.
8. Wire Try Again to call `retryTag`.
9. Replace the chat placeholder with the `ChatPanel`
   component.

### Step 4: Export InlineTagExtension

Add the export to `client/src/components/Editor/index.ts`:

```typescript
export { default as InlineTagExtension } from
    './InlineTagExtension';
```

### Step 5: Run integration tests

```bash
cd /Users/antonypegg/PROJECTS/imagineer/ \
    .worktrees/frontend-rebuild/client \
    && npx vitest run \
    src/components/Workspace/__tests__/TagIntegration.test.tsx
```

The expected output shows all five tests passing.

### Step 6: Run the full test suite

```bash
cd /Users/antonypegg/PROJECTS/imagineer/ \
    .worktrees/frontend-rebuild/client \
    && npx vitest run
```

All tests across the tagging system should pass.

### Commit

```
feat: wire tagging system into workspace editor
```

---

## Relationship to Existing Plan

This plan replaces the following tasks from the
original frontend implementation plan
(`docs/plans/2026-03-10-frontend-implementation-plan.md`):

- Old Task 10 (Inline Tag Detection) is replaced by
  Tasks 3 and 10 in this plan.
- Old Task 13 (Chat Panel) is replaced by Task 6 in
  this plan, which now includes multi-conversation
  support.
- Old Task 14 (Token Cost Display) is replaced by
  Task 7 in this plan, which now includes a running
  total.
- Old Task 15 (Card Pull-In) is incorporated into
  Task 8 as the Pull to Chat action on
  `TagResultCard`.
- Old Task 16 (Deliberate Review) is replaced by
  Task 9 in this plan, which now includes the tag
  gate.
- Old Task 20 (Tag Result Cards) is replaced by
  Task 8 in this plan with updated actions and
  lifecycle.

The following tasks from the original plan remain
unchanged and should execute alongside this plan:

- Task 11: Entity Name Detection Integration.
- Task 12: Editor-Chat Selection Bridge.
- Tasks 17 through 19: Pinboard Container, Entity
  Card, and Finding Card.
- Tasks 21 through 29: Entity pages, dashboard,
  scratchpad, graph, map, pipeline, cleanup, and
  smoke tests.
