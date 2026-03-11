<!--
  Imagineer - TTRPG Campaign Intelligence Platform

  Copyright (c) 2025 - 2026
  This software is released under The MIT License
-->

# Frontend Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use
> superpowers:executing-plans to implement this plan
> task-by-task.

**Goal:** Build the conversation-first frontend described
in the frontend design document, replacing the existing
editor-centric UI with a writer's desk workspace featuring
an editor, floating entity pinboard, and contextual chat.

**Architecture:** React 18 + TypeScript + MUI v5 + Tiptap
editor + React Query for server state + SSE for streaming.
The workspace has three layers (editor, pinboard, chat)
connected through a structural change pipeline. Navigation
is wiki-style (entity-to-entity) plus search, replacing
sidebar tree navigation.

**Tech Stack:** React 18, TypeScript 5, MUI v5,
@tiptap/react, @tanstack/react-query, EventSource API,
React Router v6, Vitest + React Testing Library

**Design Document:**
`docs/plans/2026-03-10-frontend-design.md`

---

## Phase 1: Foundation (Tasks 1-5)

Phase 1 adds the TypeScript types, API clients, React
Query hooks, and context providers that every subsequent
task depends on. No visible UI changes occur during this
phase.

---

### Task 1: Conversation Types and API Client

Add TypeScript types for conversations and messages that
match the backend models created in the conversational
backend plan. Create the conversation API client following
the existing pattern in `client/src/api/entities.ts`.

**Depends on:** Nothing (first task).

**Files:**

- Modify: `client/src/types/index.ts`
- Create: `client/src/api/conversations.ts`
- Create: `client/src/api/__tests__/conversations.test.ts`
- Modify: `client/src/api/index.ts`

**Step 1: Add conversation types to the type file**

Append the following types to `client/src/types/index.ts`
after the existing type definitions.

```typescript
// Conversation scope types
type ScopeType =
    | 'campaign'
    | 'chapter'
    | 'session'
    | 'scene'
    | 'entity';

type MessageRole = 'user' | 'assistant' | 'system';

interface Conversation {
    id: number;
    campaignId: number;
    scopeType: ScopeType;
    scopeId: number;
    title: string;
    summary?: string;
    messageCount: number;
    totalInputTokens: number;
    totalOutputTokens: number;
    createdAt: string;
    updatedAt: string;
}

interface Message {
    id: number;
    conversationId: number;
    role: MessageRole;
    content: string;
    toolName?: string;
    toolInput?: Record<string, unknown>;
    toolResult?: Record<string, unknown>;
    tokenCount: number;
    createdAt: string;
}

// SSE event types for streaming
type StreamEventType =
    | 'text_delta'
    | 'tool_use'
    | 'tool_result'
    | 'done'
    | 'error';

interface StreamEvent {
    type: StreamEventType;
    text?: string;
    toolName?: string;
    toolInput?: Record<string, unknown>;
    toolResult?: Record<string, unknown>;
    error?: string;
    usage?: {
        inputTokens: number;
        outputTokens: number;
    };
}

interface TokenUsage {
    inputTokens: number;
    outputTokens: number;
}
```

Export all new types from the file.

**Step 2: Write failing tests for the API client**

Create the test file at
`client/src/api/__tests__/conversations.test.ts`. The
tests verify that each API function calls the correct
endpoint with the expected HTTP method and parameters.
Follow the pattern in `client/src/api/client.test.ts`.

Write tests for the following functions:

- `list` calls `GET /campaigns/:id/conversations`
  with optional query parameters.
- `get` calls
  `GET /campaigns/:id/conversations/:conversationId`.
- `create` calls `POST /campaigns/:id/conversations`
  with the request body containing `scopeType` and
  `scopeId`.

Mock the `apiClient` module to intercept calls.

Run the tests to verify they fail:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run src/api/__tests__/conversations.test.ts
```

**Step 3: Implement the conversation API client**

Create `client/src/api/conversations.ts` following the
pattern in `client/src/api/entities.ts`. The file imports
`apiClient` from `./client` and exports a
`conversationsApi` object.

Define input types within the file:

```typescript
export interface ListConversationsParams {
    campaignId: number;
    scopeType?: ScopeType;
    scopeId?: number;
    page?: number;
    pageSize?: number;
}

export interface CreateConversationInput {
    scopeType: ScopeType;
    scopeId: number;
    title?: string;
}
```

Implement the following functions on the
`conversationsApi` object:

- `list(params)` sends a GET request to
  `/campaigns/${campaignId}/conversations` with
  query parameters for optional filters.
- `get(campaignId, conversationId)` sends a GET
  request to the single-conversation endpoint.
- `create(campaignId, input)` sends a POST request
  with the input body.
- `getMessages(campaignId, conversationId)` sends
  a GET request to
  `/campaigns/${campaignId}/conversations/${conversationId}/messages`.

The `sendMessage` function is handled separately in
Task 3 because SSE streaming requires a different
approach than the standard `apiClient` fetch pattern.

**Step 4: Export from the API index**

Add the conversation API exports to
`client/src/api/index.ts`:

```typescript
export { conversationsApi } from './conversations';
export type {
    ListConversationsParams,
    CreateConversationInput,
} from './conversations';
```

**Step 5: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run src/api/__tests__/conversations.test.ts
```

**Commit message:**
`feat(client): add conversation types and API client`

---

### Task 2: Conversation React Query Hooks

Create React Query hooks for conversations following
the pattern in `client/src/hooks/useEntities.ts`. The
hooks provide cached server state for conversation lists,
single conversations, and mutations.

**Depends on:** Task 1 (types and API client).

**Files:**

- Create: `client/src/hooks/useConversations.ts`
- Create:
  `client/src/hooks/__tests__/useConversations.test.ts`
- Modify: `client/src/hooks/index.ts`

**Step 1: Write failing tests**

Create
`client/src/hooks/__tests__/useConversations.test.ts`.
Wrap each test in a `QueryClientProvider` with a fresh
`QueryClient`. Mock `conversationsApi` to return test
data.

Write tests for the following hooks:

- `useConversations(campaignId)` fetches and returns
  a list of conversations for the campaign.
- `useConversation(campaignId, conversationId)` fetches
  and returns a single conversation.
- `useCreateConversation()` calls the create API and
  invalidates the conversation list cache on success.

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/hooks/__tests__/useConversations.test.ts
```

**Step 2: Implement the hooks**

Create `client/src/hooks/useConversations.ts` with the
following query key factory:

```typescript
export const conversationKeys = {
    all: ['conversations'] as const,
    lists: () =>
        [...conversationKeys.all, 'list'] as const,
    list: (
        campaignId: number,
        params?: ListConversationsParams,
    ) =>
        [...conversationKeys.lists(), campaignId, params]
        as const,
    details: () =>
        [...conversationKeys.all, 'detail'] as const,
    detail: (id: number) =>
        [...conversationKeys.details(), id] as const,
};
```

Implement the following hooks:

- `useConversations(campaignId, params?)` wraps
  `useQuery` with the list query key.
- `useConversation(campaignId, conversationId)` wraps
  `useQuery` with the detail query key.
- `useCreateConversation()` wraps `useMutation` and
  invalidates conversation list queries on success.

The `useSendMessage` hook is deferred to Task 3 because
the hook depends on the SSE streaming infrastructure.

**Step 3: Export from the hooks index**

Add exports to `client/src/hooks/index.ts`:

```typescript
// Conversation hooks
export {
    useConversations,
    useConversation,
    useCreateConversation,
    conversationKeys,
} from './useConversations';
```

**Step 4: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/hooks/__tests__/useConversations.test.ts
```

**Commit message:**
`feat(client): add conversation React Query hooks`

---

### Task 3: SSE Streaming Hook

Create a reusable hook for consuming SSE streams from
the conversation backend. The backend returns SSE events
from a POST endpoint, so the hook uses `fetch` with a
streaming `ReadableStream` reader rather than the
browser `EventSource` API (which only supports GET).

**Depends on:** Task 1 (types).

**Files:**

- Create: `client/src/hooks/useSSEStream.ts`
- Create:
  `client/src/hooks/__tests__/useSSEStream.test.ts`
- Modify: `client/src/hooks/index.ts`

**Step 1: Write failing tests**

Create
`client/src/hooks/__tests__/useSSEStream.test.ts`.
Mock `fetch` to return a `Response` with a
`ReadableStream` body that emits SSE-formatted events.

The SSE format follows this pattern:

```
event: text_delta
data: {"text":"Hello"}

event: done
data: {"usage":{"inputTokens":10,"outputTokens":5}}

```

Write tests for the following scenarios:

- The hook accumulates text from multiple
  `text_delta` events into `accumulatedText`.
- The hook calls `onTextDelta` for each text delta.
- The hook calls `onToolUse` when a `tool_use` event
  arrives.
- The hook calls `onToolResult` when a `tool_result`
  event arrives.
- The hook calls `onDone` with usage data when the
  stream completes.
- The hook calls `onError` when an `error` event
  arrives.
- The hook sets `isStreaming` to `true` while the
  stream is active and `false` when the stream ends.
- Calling `cancel` aborts the stream and sets
  `isStreaming` to `false`.
- The hook cleans up when the component unmounts
  (the `AbortController` fires).

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/hooks/__tests__/useSSEStream.test.ts
```

**Step 2: Implement the SSE streaming hook**

Create `client/src/hooks/useSSEStream.ts` with the
following interface:

```typescript
interface UseSSEStreamOptions {
    onTextDelta?: (text: string) => void;
    onToolUse?: (
        name: string,
        input: Record<string, unknown>,
    ) => void;
    onToolResult?: (
        result: Record<string, unknown>,
    ) => void;
    onDone?: (usage?: TokenUsage) => void;
    onError?: (error: string) => void;
}

interface UseSSEStreamReturn {
    sendMessage: (
        campaignId: number,
        conversationId: number,
        content: string,
    ) => void;
    cancel: () => void;
    isStreaming: boolean;
    accumulatedText: string;
    error: string | null;
    usage: TokenUsage | null;
}

function useSSEStream(
    options: UseSSEStreamOptions,
): UseSSEStreamReturn
```

Implementation details:

- Store an `AbortController` ref using `useRef`.
- The `sendMessage` function sends a POST request to
  `/api/campaigns/${campaignId}/conversations/${conversationId}/messages`
  with `{ content }` as the JSON body and
  `Accept: text/event-stream` header.
- Read the response body as a stream using
  `response.body.getReader()` and a `TextDecoder`.
- Parse SSE events by splitting on double newlines.
  Extract the `event:` and `data:` fields from each
  event block.
- Dispatch parsed events to the appropriate callback.
- Accumulate text deltas in a `useState` string.
- Track streaming state (`isStreaming`) in `useState`.
- The `cancel` function calls `abort()` on the
  `AbortController`.
- The `useEffect` cleanup function calls `cancel` to
  prevent leaks on unmount.
- Include the JWT auth token in the request headers
  by importing `getStoredToken` from
  `../contexts/AuthContext`.

**Step 3: Export from the hooks index**

Add to `client/src/hooks/index.ts`:

```typescript
// SSE streaming hook
export { useSSEStream } from './useSSEStream';
export type {
    UseSSEStreamOptions,
    UseSSEStreamReturn,
} from './useSSEStream';
```

**Step 4: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/hooks/__tests__/useSSEStream.test.ts
```

**Commit message:**
`feat(client): add SSE streaming hook for conversations`

---

### Task 4: Workspace Context

Create a React context provider that holds workspace-wide
state shared between the editor, pinboard, and chat. The
context tracks editor selection, pinned entities, pending
structural changes, and chat visibility.

**Depends on:** Task 1 (types).

**Files:**

- Create: `client/src/contexts/WorkspaceContext.tsx`
- Create:
  `client/src/contexts/__tests__/WorkspaceContext.test.tsx`

**Step 1: Write failing tests**

Create
`client/src/contexts/__tests__/WorkspaceContext.test.tsx`.
Use `renderHook` with a `WorkspaceProvider` wrapper to
test the context hook.

Write tests for the following state transitions:

- `setSelection` updates `selectedText` and
  `selectionRange`.
- `pinEntity` adds an entity ID to `pinnedEntityIds`.
- `unpinEntity` removes an entity ID from
  `pinnedEntityIds`.
- Pinning the same entity twice does not create
  duplicates.
- `addPendingChange` adds a change to
  `pendingChanges`.
- `approvePendingChange` sets the change status to
  `'approved'`.
- `dismissPendingChange` sets the change status to
  `'dismissed'`.
- `approveAllPending` sets all pending changes to
  `'approved'`.
- `setChatVisible` toggles the `chatVisible` flag.
- `setActiveCardId` updates the active card.

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/contexts/__tests__/WorkspaceContext.test.tsx
```

**Step 2: Implement the workspace context**

Create `client/src/contexts/WorkspaceContext.tsx` with the
following types and provider:

```typescript
interface PendingEntity {
    tempId: string;
    name: string;
    entityType: EntityType;
    description?: string;
    source: 'ai_suggestion' | 'inline_tag';
    status: 'proposed' | 'approved' | 'dismissed';
}

interface PendingChange {
    id: string;
    type:
        | 'create_entity'
        | 'create_relationship'
        | 'link_entity';
    entity?: PendingEntity;
    description: string;
    status: 'pending' | 'approved' | 'dismissed';
}

interface WorkspaceContextValue {
    // Document state
    documentId: number | null;
    scopeType: ScopeType | null;
    scopeId: number | null;

    // Editor selection
    selectedText: string;
    selectionRange: {
        from: number;
        to: number;
    } | null;
    setSelection: (
        text: string,
        range: { from: number; to: number } | null,
    ) => void;

    // Pinboard
    pinnedEntityIds: number[];
    pinEntity: (id: number) => void;
    unpinEntity: (id: number) => void;

    // Pending changes (structural pipeline)
    pendingChanges: PendingChange[];
    addPendingChange: (
        change: PendingChange,
    ) => void;
    approvePendingChange: (id: string) => void;
    dismissPendingChange: (id: string) => void;
    approveAllPending: () => void;

    // Chat state
    chatVisible: boolean;
    setChatVisible: (visible: boolean) => void;
    activeCardId: string | null;
    setActiveCardId: (
        id: string | null,
    ) => void;
}
```

Use `useReducer` rather than multiple `useState` calls
to manage the interconnected state. Export the context,
provider component, and a `useWorkspace` convenience
hook that calls `useContext` with an error guard.

**Step 3: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/contexts/__tests__/WorkspaceContext.test.tsx
```

**Commit message:**
`feat(client): add workspace context provider`

---

### Task 5: Client-Side Entity Name Detection Hook

Create a hook that performs procedural entity name
detection against the local entity cache. The hook runs
on a typing pause and detects mentions of known entity
names in the editor text at zero LLM cost.

This hook differs from the existing
`useEntityDetection` hook in `client/src/hooks/`, which
calls the backend entity detection API. This new hook
runs entirely client-side using the React Query entity
cache and matches entity names against the text content
using substring matching.

**Depends on:** Task 1 (types).

**Files:**

- Create:
  `client/src/hooks/useInlineEntityDetection.ts`
- Create:
  `client/src/hooks/__tests__/useInlineEntityDetection.test.ts`
- Modify: `client/src/hooks/index.ts`

**Step 1: Write failing tests**

Create
`client/src/hooks/__tests__/useInlineEntityDetection.test.ts`.
Mock the entity list API to return a known set of
entities.

Write tests for the following scenarios:

- The hook detects a known entity name in the text and
  returns a `DetectedEntity` with the correct entity
  ID, name, type, and text position.
- The hook debounces detection (does not run on every
  keystroke).
- The hook marks already-linked entities (those
  inside `[[...]]` wiki link syntax) with
  `isLinked: true`.
- The hook returns an empty array when no entity names
  appear in the text.
- The hook handles case-insensitive matching.
- The hook ignores entity names shorter than three
  characters to prevent false positives.

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/hooks/__tests__/useInlineEntityDetection.test.ts
```

**Step 2: Implement the detection hook**

Create `client/src/hooks/useInlineEntityDetection.ts`
with the following interface:

```typescript
interface DetectedEntity {
    entityId: number;
    entityName: string;
    entityType: EntityType;
    position: { from: number; to: number };
    isLinked: boolean;
}

function useInlineEntityDetection(
    content: string,
    campaignId: number,
): DetectedEntity[]
```

Implementation details:

- Use `useEntities` with the campaign ID to access the
  local entity cache.
- Debounce the detection logic using a `useEffect`
  with a 500ms `setTimeout`.
- For each entity in the cache, search for the entity
  name as a case-insensitive substring in the content.
- Check whether each match position falls within a
  `[[...]]` wiki link node to set `isLinked`.
- Return all matches sorted by position.

**Step 3: Export from the hooks index**

Add to `client/src/hooks/index.ts`:

```typescript
// Inline entity detection hook
export {
    useInlineEntityDetection,
} from './useInlineEntityDetection';
export type {
    DetectedEntity,
} from './useInlineEntityDetection';
```

**Step 4: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/hooks/__tests__/useInlineEntityDetection.test.ts
```

**Commit message:**
`feat(client): add client-side entity name detection hook`

---

## Phase 2: Workspace Layout (Tasks 6-8)

Phase 2 builds the visible workspace shell and routing
structure. After this phase, the new layout renders
and routes resolve to the correct components.

---

### Task 6: Workspace Shell

Build the new top-level layout component that replaces
`AppShell`. The workspace shell provides a minimal
header with global search and a full-height content
area. No sidebar navigation tree exists in the new
layout.

**Depends on:** Task 4 (WorkspaceContext).

**Files:**

- Create: `client/src/layouts/WorkspaceShell.tsx`
- Create: `client/src/components/GlobalSearch.tsx`
- Create:
  `client/src/layouts/__tests__/WorkspaceShell.test.tsx`
- Create:
  `client/src/components/__tests__/GlobalSearch.test.tsx`
- Modify: `client/src/layouts/index.ts`

**Step 1: Write failing tests for WorkspaceShell**

Create
`client/src/layouts/__tests__/WorkspaceShell.test.tsx`.
Wrap the component in `MemoryRouter` and mock auth
context.

Write tests for the following:

- The shell renders a header with the logo, search
  box, campaign selector, and user menu.
- The shell renders child content in the main area.
- The header logo links to `/`.
- The campaign selector displays the current campaign
  name.

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/layouts/__tests__/WorkspaceShell.test.tsx
```

**Step 2: Write failing tests for GlobalSearch**

Create
`client/src/components/__tests__/GlobalSearch.test.tsx`.

Write tests for the following:

- The component renders an MUI `Autocomplete` input.
- Typing a query calls the entity search API after a
  debounce.
- Selecting a result navigates to the entity page.
- The component displays entity type badges next to
  each result.

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/__tests__/GlobalSearch.test.tsx
```

**Step 3: Implement GlobalSearch**

Create `client/src/components/GlobalSearch.tsx`.

The component uses MUI `Autocomplete` with
`freeSolo` mode. The search input debounces by 300ms
and calls `entitiesApi.list()` with the search term.
Results display entity name, type badge, and a
truncated description. Selecting a result calls
`navigate()` from React Router to go to
`/campaigns/${campaignId}/entities/${entityId}`.

**Step 4: Implement WorkspaceShell**

Create `client/src/layouts/WorkspaceShell.tsx`.

The shell layout follows this structure:

```
Box (flex column, height: 100vh)
  AppBar (position: sticky)
    Toolbar
      Logo link (/)
      GlobalSearch (flex: 1, max-width: 600px)
      Campaign selector (Select or Button)
      User avatar + menu
  Box (flex: 1, overflow: hidden)
    <Outlet /> (child routes)
```

Use MUI `AppBar`, `Toolbar`, `IconButton`, `Menu`,
and `MenuItem` components. The campaign selector reads
from `CampaignContext`. The user menu provides
settings and logout links.

**Step 5: Export from layouts index**

Add to `client/src/layouts/index.ts`:

```typescript
export { WorkspaceShell } from './WorkspaceShell';
```

**Step 6: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/layouts/__tests__/WorkspaceShell.test.tsx \
    src/components/__tests__/GlobalSearch.test.tsx
```

**Commit message:**
`feat(client): add workspace shell layout and
global search`

---

### Task 7: Workspace Page Component

Build the three-layer workspace page that combines
the editor, pinboard, and chat into a single view.
The component acts as the container that arranges the
three workspace layers.

**Depends on:** Task 4 (WorkspaceContext), Task 6
(WorkspaceShell).

**Files:**

- Create:
  `client/src/components/Workspace/WorkspacePage.tsx`
- Create:
  `client/src/components/Workspace/__tests__/WorkspacePage.test.tsx`

**Step 1: Write failing tests**

Create
`client/src/components/Workspace/__tests__/WorkspacePage.test.tsx`.
Wrap in `WorkspaceProvider`, `QueryClientProvider`,
and `MemoryRouter`.

Write tests for the following:

- The component renders an editor area.
- The component renders a chat toggle button.
- Clicking the chat toggle shows the chat panel area.
- The component renders a pinboard container area.
- The component wraps children in a
  `WorkspaceProvider`.

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Workspace/__tests__/WorkspacePage.test.tsx
```

**Step 2: Implement WorkspacePage**

Create
`client/src/components/Workspace/WorkspacePage.tsx`.

The component accepts the following props:

```typescript
interface WorkspacePageProps {
    campaignId: number;
    scopeType: ScopeType;
    scopeId: number;
    initialContent: string;
    onContentChange: (content: string) => void;
    onSave: (content: string) => Promise<void>;
}
```

Layout structure:

```
WorkspaceProvider (sets documentId, scopeType, scopeId)
  Box (position: relative, flex: 1, display: flex,
       flex-direction: column)
    Box (flex: 1, display: flex, position: relative,
         overflow: hidden)
      // Pinboard container (position: absolute,
      //                     full area, pointer-events
      //                     set so cards receive
      //                     clicks but editor shows
      //                     through)
      PinboardContainer
      // Editor (centred, z-index above pinboard
      //         background)
      Box (max-width: 800px, mx: auto, flex: 1)
        // Editor placeholder for now
    // Chat panel (collapsible, at bottom)
    ChatPanel (visible based on chatVisible
               from context)
```

Use placeholder `Box` components with labels for the
editor, pinboard, and chat. Subsequent tasks replace
the placeholders with real components.

**Step 3: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Workspace/__tests__/WorkspacePage.test.tsx
```

**Commit message:**
`feat(client): add workspace page layout component`

---

### Task 8: Route Restructuring

Replace the existing route structure with workspace-based
navigation. Keep old routes functional during the
transition by marking deprecated pages.

**Depends on:** Task 6 (WorkspaceShell), Task 7
(WorkspacePage).

**Files:**

- Modify: `client/src/App.tsx`
- Create:
  `client/src/pages/WorkspaceRoutePage.tsx`
- Modify existing page files (add deprecation
  comments only)

**Step 1: Write failing tests**

Add route rendering tests to a new file
`client/src/App.test.tsx` (or modify if one exists).
Wrap `App` in test providers and verify route
resolution.

Write tests for the following routes:

- `/campaigns/:id` renders the campaign dashboard.
- `/campaigns/:id/workspace` renders the workspace
  page.
- `/campaigns/:id/workspace/:scopeType/:scopeId`
  renders the scoped workspace page.
- `/campaigns/:id/entities` renders the entity list
  page.
- `/campaigns/:id/entities/:entityId` renders the
  entity page.
- `/campaigns/:id/graph` renders the graph explorer
  placeholder.
- `/campaigns/:id/map` renders the map view
  placeholder.

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run src/App.test.tsx
```

**Step 2: Create WorkspaceRoutePage**

Create `client/src/pages/WorkspaceRoutePage.tsx`.

The page extracts `campaignId`, `scopeType`, and
`scopeId` from the route parameters using
`useParams`. The page fetches the document content
for the given scope and passes the data to
`WorkspacePage` as props.

For the initial implementation, use a placeholder
content loader that returns empty content. Wire the
real content fetching (chapters, sessions, and
entities) in a follow-up commit.

**Step 3: Update App.tsx routes**

Modify `client/src/App.tsx` to add a new
`WorkspaceShellWrapper` component alongside the
existing `AppShellWrapper`. The new wrapper uses
`WorkspaceShell` instead of `AppShell`.

Add the following new routes under the workspace
shell wrapper:

```typescript
<Route element={<WorkspaceShellWrapper />}>
    <Route
        path="/campaigns/:id"
        element={<CampaignDashboard />}
    />
    <Route
        path="/campaigns/:id/workspace"
        element={<WorkspaceRoutePage />}
    />
    <Route
        path="/campaigns/:id/workspace/:scopeType/:scopeId"
        element={<WorkspaceRoutePage />}
    />
    <Route
        path="/campaigns/:id/entities"
        element={<Entities />}
    />
    <Route
        path="/campaigns/:id/entities/:entityId"
        element={<EntityView />}
    />
    <Route
        path="/campaigns/:id/graph"
        element={<GraphPlaceholder />}
    />
    <Route
        path="/campaigns/:id/map"
        element={<MapPlaceholder />}
    />
    <Route
        path="/campaigns/:id/scratchpad"
        element={<ScratchpadPlaceholder />}
    />
</Route>
```

Keep the existing `AppShellWrapper` routes intact.
Add a comment marking the old routes as deprecated.

**Step 4: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run src/App.test.tsx
```

**Commit message:**
`feat(client): add workspace routes alongside
existing routes`

---

## Phase 3: The Editor (Tasks 9-12)

Phase 3 builds the enhanced editor that sits at the
centre of the workspace. The editor extends the
existing Tiptap foundation with selection tracking,
inline tag detection, and entity name highlighting.

---

### Task 9: Enhanced Tiptap Editor

Build the workspace editor component on top of the
existing Tiptap MarkdownEditor. The new component adds
selection tracking, content change debouncing, draft
content marking, and auto-save with dirty state
tracking.

**Depends on:** Task 4 (WorkspaceContext), Task 7
(WorkspacePage).

**Files:**

- Create:
  `client/src/components/Editor/WorkspaceEditor.tsx`
- Create:
  `client/src/components/Editor/SelectionTracker.ts`
- Create:
  `client/src/components/Editor/DraftMark.ts`
- Create:
  `client/src/components/Editor/__tests__/WorkspaceEditor.test.tsx`

**Step 1: Write failing tests**

Create
`client/src/components/Editor/__tests__/WorkspaceEditor.test.tsx`.
Wrap the component in `WorkspaceProvider`.

Write tests for the following:

- The editor renders with initial content.
- Selecting text in the editor updates
  `selectedText` and `selectionRange` in the
  workspace context.
- The editor calls `onContentChange` with debounced
  content.
- The editor tracks dirty state (has unsaved changes).
- Draft-marked content renders with a distinct CSS
  class.

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Editor/__tests__/WorkspaceEditor.test.tsx
```

**Step 2: Create the SelectionTracker extension**

Create
`client/src/components/Editor/SelectionTracker.ts`.

The extension is a Tiptap `Extension` that listens to
the `selectionUpdate` transaction event. On each
update, the extension extracts the selected text and
the `from`/`to` positions, then calls a callback
function passed via extension options.

```typescript
import { Extension } from '@tiptap/core';

interface SelectionTrackerOptions {
    onSelectionChange: (
        text: string,
        range: {
            from: number;
            to: number;
        } | null,
    ) => void;
}

const SelectionTracker =
    Extension.create<SelectionTrackerOptions>({
    name: 'selectionTracker',
    // ... implementation
});
```

**Step 3: Create the DraftMark mark type**

Create `client/src/components/Editor/DraftMark.ts`.

The mark is a Tiptap `Mark` that renders a `<span>`
with a CSS class `draft-content`. The mark has no
keyboard shortcuts or input rules because the
application applies the mark programmatically when
inserting AI-generated content.

```typescript
import { Mark } from '@tiptap/core';

const DraftMark = Mark.create({
    name: 'draftContent',
    // ... implementation
});
```

**Step 4: Implement WorkspaceEditor**

Create
`client/src/components/Editor/WorkspaceEditor.tsx`.

The component uses `useEditor` from `@tiptap/react`
with the StarterKit, Markdown, WikiLink, Placeholder,
SelectionTracker, and DraftMark extensions. The
component renders the `EditorContent` inside an MUI
`Paper` component with appropriate styling.

Props:

```typescript
interface WorkspaceEditorProps {
    initialContent: string;
    onContentChange: (content: string) => void;
    onSave: (content: string) => Promise<void>;
    readOnly?: boolean;
}
```

The `onContentChange` callback fires on a 300ms
debounce after the editor content changes. Auto-save
triggers on a 5-second debounce when dirty state is
true. The component uses `useWorkspace` to connect
the selection tracker to the context.

**Step 5: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Editor/__tests__/WorkspaceEditor.test.tsx
```

**Commit message:**
`feat(client): add workspace editor with selection
tracking and draft marks`

---

### Task 10: Inline Tag Detection Extension

Create a Tiptap extension that detects `@[...]` tags in
the editor text and emits events for processing. The
extension watches for the closing `]` character after an
`@[` sequence, extracts the request text, and dispatches
an event. The extension does not call any API directly.

**Depends on:** Task 9 (WorkspaceEditor).

**Files:**

- Create:
  `client/src/components/Editor/InlineTagExtension.ts`
- Create:
  `client/src/components/Editor/__tests__/InlineTagExtension.test.ts`

**Step 1: Write failing tests**

Create
`client/src/components/Editor/__tests__/InlineTagExtension.test.ts`.
Create a Tiptap editor instance in each test with the
extension enabled.

Write tests for the following:

- Typing `@[need an NPC]` triggers the `onTag`
  callback with `content: "need an NPC"` and the
  correct position range.
- The extension correctly identifies the `@[` opening
  and `]` closing delimiters.
- Nested brackets like `@[need [urgent] NPC]` do not
  cause false triggers.
- The extension assigns a unique `id` to each
  detected tag.
- Multiple tags in the same document each trigger
  separate callbacks.

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Editor/__tests__/InlineTagExtension.test.ts
```

**Step 2: Implement InlineTagExtension**

Create
`client/src/components/Editor/InlineTagExtension.ts`.

The extension defines the following types:

```typescript
interface InlineTag {
    id: string;
    content: string;
    position: { from: number; to: number };
    status: 'pending' | 'processing' | 'resolved';
    result?: InlineTagResult;
}

type InlineTagResult =
    | { type: 'text'; text: string }
    | {
        type: 'entity';
        entity: PendingEntity;
        existingMatches?: Entity[];
      }
    | {
        type: 'options';
        options: Array<{
            label: string;
            value: string;
        }>;
      }
    | {
        type: 'card';
        title: string;
        content: string;
      };
```

The extension uses a Tiptap `Plugin` with an
`appendTransaction` hook. On each transaction, the
plugin scans the document text for complete `@[...]`
patterns. When the plugin finds a new complete tag
(one not already tracked), the plugin calls the
`onTag` callback from the extension options.

The extension also provides a `DecorationSet` that
renders tags with a coloured background while the
system processes them.

Extension options:

```typescript
interface InlineTagOptions {
    onTag: (tag: InlineTag) => void;
}
```

**Step 3: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Editor/__tests__/InlineTagExtension.test.ts
```

**Commit message:**
`feat(client): add inline @[...] tag detection
Tiptap extension`

---

### Task 11: Entity Name Detection Integration

Wire the client-side entity detection hook (Task 5)
into the editor to show inline wiki link suggestions.
When the hook detects unlinked entity names, the
editor highlights them with a subtle decoration, and
a popover offers to convert the name to a wiki link.

**Depends on:** Task 5 (useInlineEntityDetection),
Task 9 (WorkspaceEditor).

**Files:**

- Create:
  `client/src/components/Editor/EntityDetectionPlugin.ts`
- Create:
  `client/src/components/Editor/EntitySuggestionPopover.tsx`
- Create:
  `client/src/components/Editor/__tests__/EntityDetectionPlugin.test.ts`

**Step 1: Write failing tests**

Create
`client/src/components/Editor/__tests__/EntityDetectionPlugin.test.ts`.
Provide a mock entity list and editor content
containing entity names.

Write tests for the following:

- Detected entity names receive a
  `entity-suggestion` CSS class decoration.
- Clicking a highlighted entity name opens a popover
  with the entity name and a "Link" button.
- Clicking "Link" replaces the plain text with a wiki
  link node.
- Already-linked entities do not receive the
  decoration.

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Editor/__tests__/EntityDetectionPlugin.test.ts
```

**Step 2: Implement EntityDetectionPlugin**

Create
`client/src/components/Editor/EntityDetectionPlugin.ts`.

The plugin is a Tiptap `Plugin` that receives a list
of `DetectedEntity` objects (from the detection hook)
through its plugin state. The plugin creates a
`DecorationSet` with inline decorations at each
detected entity position, applying the
`entity-suggestion` CSS class.

**Step 3: Implement EntitySuggestionPopover**

Create
`client/src/components/Editor/EntitySuggestionPopover.tsx`.

The popover uses MUI `Popover` and appears when the
user clicks on a decorated entity name. The popover
displays the entity name, type badge, and a "Link"
button. Clicking "Link" dispatches a Tiptap command
that replaces the text range with a wiki link node.

**Step 4: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Editor/__tests__/EntityDetectionPlugin.test.ts
```

**Commit message:**
`feat(client): add entity name detection and
wiki link suggestions in editor`

---

### Task 12: Editor-Chat Selection Bridge

Wire the editor selection state to the chat panel so
that highlighted text in the editor appears as context
in the chat input. When the GM selects text and opens
the chat, the chat displays a context preview above the
input field.

**Depends on:** Task 4 (WorkspaceContext), Task 9
(WorkspaceEditor).

**Files:**

- Modify:
  `client/src/components/Editor/WorkspaceEditor.tsx`
  (minor, verify selection tracking works)
- Create:
  `client/src/components/Chat/SelectionPreview.tsx`
- Create:
  `client/src/components/Chat/__tests__/SelectionPreview.test.tsx`

**Step 1: Write failing tests**

Create
`client/src/components/Chat/__tests__/SelectionPreview.test.tsx`.
Render the component inside a `WorkspaceProvider` with
pre-set selection state.

Write tests for the following:

- When `selectedText` is empty, the component renders
  nothing.
- When `selectedText` has content, the component
  renders a preview chip with truncated text.
- The preview chip has a "Clear" button that calls
  `setSelection('', null)`.
- Long selected text truncates to 100 characters with
  an ellipsis.

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Chat/__tests__/SelectionPreview.test.tsx
```

**Step 2: Implement SelectionPreview**

Create
`client/src/components/Chat/SelectionPreview.tsx`.

The component reads `selectedText` from the workspace
context. When the text is non-empty, the component
renders an MUI `Chip` with a `FormatQuote` icon and
truncated text. A close icon on the chip clears the
selection.

**Step 3: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Chat/__tests__/SelectionPreview.test.tsx
```

**Commit message:**
`feat(client): add editor-to-chat selection bridge`

---

## Phase 4: The Chat (Tasks 13-16)

Phase 4 builds the chat panel, token cost display,
card pull-in mechanism, and deliberate review
interface.

---

### Task 13: Chat Panel Component

Build the collapsible chat panel that appears below the
editor. The panel displays a message history, streams
incoming assistant responses, shows tool use indicators,
and accepts user input.

**Depends on:** Task 3 (useSSEStream), Task 4
(WorkspaceContext), Task 12 (SelectionPreview).

**Files:**

- Create: `client/src/components/Chat/ChatPanel.tsx`
- Create: `client/src/components/Chat/ChatMessage.tsx`
- Create: `client/src/components/Chat/ChatInput.tsx`
- Create:
  `client/src/components/Chat/StreamingMessage.tsx`
- Create:
  `client/src/components/Chat/ToolUseIndicator.tsx`
- Create:
  `client/src/components/Chat/__tests__/ChatPanel.test.tsx`
- Create:
  `client/src/components/Chat/__tests__/ChatMessage.test.tsx`
- Create:
  `client/src/components/Chat/__tests__/ChatInput.test.tsx`

**Step 1: Write failing tests for ChatMessage**

Create
`client/src/components/Chat/__tests__/ChatMessage.test.tsx`.

Write tests for the following:

- A user message renders with a "You" label and the
  message text.
- An assistant message renders with markdown
  formatting.
- An assistant message with tool calls shows the tool
  use indicator.
- The message displays a token count badge.

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Chat/__tests__/ChatMessage.test.tsx
```

**Step 2: Write failing tests for ChatInput**

Create
`client/src/components/Chat/__tests__/ChatInput.test.tsx`.

Write tests for the following:

- The input renders a text field and a send button.
- The send button is disabled when the input is
  empty.
- The send button is disabled when a message is
  streaming.
- Pressing Enter sends the message.
- The selection preview appears above the input when
  text is selected in the editor.

**Step 3: Write failing tests for ChatPanel**

Create
`client/src/components/Chat/__tests__/ChatPanel.test.tsx`.

Write tests for the following:

- The panel is hidden when `chatVisible` is `false`
  in the workspace context.
- The panel shows the message list when `chatVisible`
  is `true`.
- The panel has a resize handle at the top edge.
- Sending a message calls the SSE stream hook.
- A streaming response displays in real time via
  `StreamingMessage`.

Run all chat tests to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run src/components/Chat/__tests__/
```

**Step 4: Implement ChatMessage**

Create `client/src/components/Chat/ChatMessage.tsx`.

The component renders user messages in a right-aligned
bubble and assistant messages in a left-aligned bubble.
Assistant message content renders through
`react-markdown` for formatting. The component shows
a small `TokenCostBadge` (Task 14) below the message
when token count data is available.

**Step 5: Implement ToolUseIndicator**

Create
`client/src/components/Chat/ToolUseIndicator.tsx`.

The component renders a small MUI `Chip` with a
`CircularProgress` spinner and a label like
"Searching entities..." based on the tool name. The
component maps known tool names to human-readable
descriptions.

**Step 6: Implement StreamingMessage**

Create
`client/src/components/Chat/StreamingMessage.tsx`.

The component accepts `accumulatedText` and
`isStreaming` props. The component renders the
accumulated text with markdown formatting and shows
a blinking cursor when streaming is active.

**Step 7: Implement ChatInput**

Create `client/src/components/Chat/ChatInput.tsx`.

The component renders an MUI `TextField` with a send
`IconButton`. The component integrates
`SelectionPreview` above the input. The component
accepts an `onSend` callback and a `disabled` prop.
Pressing Enter (without Shift) triggers the send
action.

**Step 8: Implement ChatPanel**

Create `client/src/components/Chat/ChatPanel.tsx`.

The panel reads `chatVisible` from the workspace
context. When visible, the panel renders as a `Box`
at the bottom of the workspace with a configurable
height (default 300px). The panel includes a drag
handle at the top edge for resizing, a scrollable
message list, and the `ChatInput` at the bottom.

The panel manages a local conversation state:

- On mount, fetch or create a conversation for the
  current scope using `useCreateConversation`.
- Display existing messages using `ChatMessage`.
- Use `useSSEStream` to handle new message sending
  and streaming.
- On stream completion, add the assistant message
  to the local message list.

**Step 9: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run src/components/Chat/__tests__/
```

**Commit message:**
`feat(client): add collapsible chat panel with
streaming support`

---

### Task 14: Token Cost Display

Build a small badge component that shows token usage for
each interaction and a running conversation total.

**Depends on:** Task 1 (TokenUsage type).

**Files:**

- Create:
  `client/src/components/Chat/TokenCostBadge.tsx`
- Create:
  `client/src/components/Chat/__tests__/TokenCostBadge.test.tsx`

**Step 1: Write failing tests**

Create
`client/src/components/Chat/__tests__/TokenCostBadge.test.tsx`.

Write tests for the following:

- The badge renders token counts in the format
  "~450 tokens".
- Counts above 1000 render with "k" suffix
  (for example, "~1.2k tokens").
- The badge renders nothing when the token count
  is zero.
- The badge displays in a small, muted font style.

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Chat/__tests__/TokenCostBadge.test.tsx
```

**Step 2: Implement TokenCostBadge**

Create
`client/src/components/Chat/TokenCostBadge.tsx`.

The component renders an MUI `Typography` with variant
`caption` and a muted colour. The component accepts
`inputTokens` and `outputTokens` props and displays
the sum. The component formats large numbers with a
"k" suffix.

**Step 3: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Chat/__tests__/TokenCostBadge.test.tsx
```

**Commit message:**
`feat(client): add token cost badge component`

---

### Task 15: Card Pull-In Mechanism

Enable pulling finding cards and entity proposal cards
from the pinboard into the chat for discussion. When a
card enters the chat context, the chat prepends the card
data to messages for context-aware conversation.

**Depends on:** Task 4 (WorkspaceContext), Task 13
(ChatPanel).

**Files:**

- Create:
  `client/src/components/Chat/CardContext.tsx`
- Create:
  `client/src/components/Chat/__tests__/CardContext.test.tsx`

**Step 1: Write failing tests**

Create
`client/src/components/Chat/__tests__/CardContext.test.tsx`.

Write tests for the following:

- When a card is pulled in, the component renders a
  context chip above the chat messages showing the
  card title.
- The context chip has a dismiss button that removes
  the card context.
- Multiple cards can be pulled in simultaneously.
- The `activeCardId` in WorkspaceContext matches the
  pulled-in card.

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Chat/__tests__/CardContext.test.tsx
```

**Step 2: Implement CardContext**

Create `client/src/components/Chat/CardContext.tsx`.

The component renders a horizontal row of MUI `Chip`
components above the chat message list. Each chip
shows the card title and has a delete icon that
removes the card from context. The component reads
the `activeCardId` from the workspace context and
renders the corresponding card title.

Provide a `pullToChat` function via props or context
that other components (finding cards, entity cards)
call to add their data to the chat context.

**Step 3: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Chat/__tests__/CardContext.test.tsx
```

**Commit message:**
`feat(client): add card pull-in mechanism for chat`

---

### Task 16: Deliberate Review Request

Build the interface for Pattern B from the design
document, where the GM requests a holistic analysis of
the current document content. The review sends the full
document to the backend and displays results as editor
annotations and pinboard finding cards.

**Depends on:** Task 3 (useSSEStream), Task 9
(WorkspaceEditor), Task 13 (ChatPanel).

**Files:**

- Create:
  `client/src/components/Chat/ReviewRequest.tsx`
- Create:
  `client/src/components/Chat/__tests__/ReviewRequest.test.tsx`

**Step 1: Write failing tests**

Create
`client/src/components/Chat/__tests__/ReviewRequest.test.tsx`.

Write tests for the following:

- Clicking the "Review" button opens a confirmation
  dialog.
- The confirmation dialog shows an estimated token
  cost based on document length.
- Confirming the dialog sends the full document
  content to the conversation API.
- Cancelling the dialog does not send any request.
- The component parses finding results from the SSE
  stream response.

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Chat/__tests__/ReviewRequest.test.tsx
```

**Step 2: Implement ReviewRequest**

Create
`client/src/components/Chat/ReviewRequest.tsx`.

The component renders a "Review" button (MUI
`IconButton` with a `RateReview` icon) in the chat
input area. Clicking the button opens an MUI `Dialog`
that shows the document word count, estimated token
count (approximately 0.75 tokens per word), and a
"Confirm" button.

On confirmation, the component sends a message to the
conversation with a system-formatted review request
that includes the full document content. The SSE
stream response is parsed for finding data.

Finding data structure:

```typescript
interface Finding {
    id: string;
    category:
        | 'plot_hole'
        | 'spotlight'
        | 'agency'
        | 'mechanics'
        | 'continuity'
        | 'pacing';
    severity: 'info' | 'warning' | 'error';
    title: string;
    description: string;
    suggestion: string;
    lineReference?: string;
    textAnchor?: { from: number; to: number };
}
```

When findings arrive through the stream, the component
dispatches them to the workspace context as pending
changes and emits them for rendering as pinboard
finding cards.

**Step 3: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Chat/__tests__/ReviewRequest.test.tsx
```

**Commit message:**
`feat(client): add deliberate review request dialog`

---

## Phase 5: The Pinboard (Tasks 17-20)

Phase 5 builds the floating entity card pinboard that
surrounds the editor. Cards are draggable, collapsible,
and support ghost state for AI-proposed entities.

---

### Task 17: Pinboard Container

Build the container that manages floating entity cards
around the editor. The container handles card
positioning, z-index management, and drag interactions.

**Depends on:** Task 4 (WorkspaceContext), Task 7
(WorkspacePage).

**New dependency:** Install `@dnd-kit/core` and
`@dnd-kit/utilities` for drag-and-drop support.

**Files:**

- Create:
  `client/src/components/Pinboard/PinboardContainer.tsx`
- Create:
  `client/src/components/Pinboard/__tests__/PinboardContainer.test.tsx`

**Step 1: Install dnd-kit**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npm install @dnd-kit/core @dnd-kit/utilities
```

**Step 2: Write failing tests**

Create
`client/src/components/Pinboard/__tests__/PinboardContainer.test.tsx`.

Write tests for the following:

- The container renders one card per pinned entity in
  the workspace context.
- The container renders ghost cards for pending entity
  changes.
- Each card is positioned absolutely within the
  container.
- Dismissing a card calls `unpinEntity` on the
  workspace context.
- New cards receive an auto-calculated position that
  avoids overlapping existing cards.

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Pinboard/__tests__/PinboardContainer.test.tsx
```

**Step 3: Implement PinboardContainer**

Create
`client/src/components/Pinboard/PinboardContainer.tsx`.

The container is a `Box` with `position: absolute`,
`inset: 0`, and `pointer-events: none`. Child cards
set `pointer-events: auto` so that clicks pass
through the container to the editor beneath but land
on the cards.

The container reads `pinnedEntityIds` and
`pendingChanges` from the workspace context. For each
pinned entity, the container renders an `EntityCard`
(Task 18). For each pending entity change, the
container renders a ghost `EntityCard`.

Card positions are stored in local state as a map
from card ID to `{ x: number; y: number }`.
Auto-positioning logic places new cards in available
space around the editor (left and right gutters).
The `DndContext` from `@dnd-kit/core` wraps the
cards to enable drag repositioning.

**Step 4: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Pinboard/__tests__/PinboardContainer.test.tsx
```

**Commit message:**
`feat(client): add pinboard container with
drag-and-drop positioning`

---

### Task 18: Entity Pinboard Card

Build the entity card component for the pinboard. The
card has three visual states: full, title strip, and
ghost (AI-proposed).

**Depends on:** Task 17 (PinboardContainer).

**Files:**

- Create:
  `client/src/components/Pinboard/EntityCard.tsx`
- Create:
  `client/src/components/Pinboard/__tests__/EntityCard.test.tsx`

**Step 1: Write failing tests**

Create
`client/src/components/Pinboard/__tests__/EntityCard.test.tsx`.

Write tests for the following:

- In full state, the card displays name, type badge,
  truncated description, and key relationships.
- In title-strip state, the card shows only the name
  and type icon at minimal height.
- Clicking the collapse button toggles between full
  and title-strip states.
- In ghost state, the card shows a dashed border and
  "Approve" and "Dismiss" buttons.
- Clicking "Approve" on a ghost card calls
  `approvePendingChange` on the workspace context.
- Clicking "Dismiss" on a ghost card calls
  `dismissPendingChange` on the workspace context.
- Real cards show "Unpin" and "Open" action buttons.
- Clicking "Open" navigates to the entity page.

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Pinboard/__tests__/EntityCard.test.tsx
```

**Step 2: Implement EntityCard**

Create
`client/src/components/Pinboard/EntityCard.tsx`.

The component uses MUI `Card` with conditional
styling based on the card state. Props:

```typescript
interface EntityCardProps {
    entityId?: number;
    pendingEntity?: PendingEntity;
    pendingChangeId?: string;
    isGhost?: boolean;
    onUnpin?: () => void;
    onApprove?: () => void;
    onDismiss?: () => void;
    onPullToChat?: () => void;
}
```

For real entities (non-ghost), the component uses
`useEntity` to fetch entity data and
`useEntityRelationships` to fetch the first four
relationships. For ghost entities, the component
renders the pending entity data directly.

The card uses MUI `CardHeader` (name + type badge),
`CardContent` (description, relationships), and
`CardActions` (action buttons). The ghost state
applies a `border: 2px dashed` style and shows
approve/dismiss buttons.

**Step 3: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Pinboard/__tests__/EntityCard.test.tsx
```

**Commit message:**
`feat(client): add entity pinboard card with ghost
state support`

---

### Task 19: Finding Card

Build the finding card component that displays results
from deliberate review analysis. Each finding has a
severity, category, suggestion, and an optional text
anchor in the editor.

**Depends on:** Task 16 (Finding type), Task 17
(PinboardContainer).

**Files:**

- Create:
  `client/src/components/Pinboard/FindingCard.tsx`
- Create:
  `client/src/components/Pinboard/__tests__/FindingCard.test.tsx`

**Step 1: Write failing tests**

Create
`client/src/components/Pinboard/__tests__/FindingCard.test.tsx`.

Write tests for the following:

- The card displays the severity icon (info uses
  `InfoOutlined`, warning uses `WarningAmber`, error
  uses `ErrorOutline`).
- The card displays the category as a coloured badge.
- The card displays the title and description text.
- The card displays the suggestion text in a distinct
  section.
- Clicking "Discuss" calls the `onPullToChat`
  callback.
- Clicking "Dismiss" calls the `onDismiss` callback.
- A finding with a `textAnchor` shows a "Show in
  editor" link.

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Pinboard/__tests__/FindingCard.test.tsx
```

**Step 2: Implement FindingCard**

Create
`client/src/components/Pinboard/FindingCard.tsx`.

The component uses MUI `Card` with a coloured left
border based on severity (info = blue, warning =
amber, error = red). Props:

```typescript
interface FindingCardProps {
    finding: Finding;
    onDismiss: () => void;
    onPullToChat: () => void;
    onShowInEditor?: () => void;
}
```

The card layout uses `CardHeader` with severity icon
and category chip, `CardContent` with description
and suggestion, and `CardActions` with "Discuss" and
"Dismiss" buttons.

**Step 3: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Pinboard/__tests__/FindingCard.test.tsx
```

**Commit message:**
`feat(client): add finding card for review results`

---

### Task 20: Inline Tag Result Cards

Build the UI for displaying results from `@[...]` tag
processing. When the backend resolves a tag, the result
appears as a card on the pinboard near the tag position
with options to accept or dismiss.

**Depends on:** Task 10 (InlineTagExtension), Task 18
(EntityCard).

**Files:**

- Create:
  `client/src/components/Pinboard/TagResultCard.tsx`
- Create:
  `client/src/components/Pinboard/__tests__/TagResultCard.test.tsx`

**Step 1: Write failing tests**

Create
`client/src/components/Pinboard/__tests__/TagResultCard.test.tsx`.

Write tests for the following:

- The card displays the AI-proposed entity name and
  type.
- When existing entity matches exist, the card shows
  a "Did you mean:" section with clickable entity
  names.
- Clicking "Accept new" creates the entity, replaces
  the tag with a wiki link, and removes the card.
- Clicking an existing entity replaces the tag with
  a wiki link to that entity and removes the card.
- Clicking "Dismiss" removes the card and leaves the
  raw tag text in the editor.

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Pinboard/__tests__/TagResultCard.test.tsx
```

**Step 2: Implement TagResultCard**

Create
`client/src/components/Pinboard/TagResultCard.tsx`.

The component extends the ghost entity card style with
additional sections for existing matches. Props:

```typescript
interface TagResultCardProps {
    tag: InlineTag;
    existingMatches: Entity[];
    onAcceptNew: () => void;
    onAcceptExisting: (entityId: number) => void;
    onDismiss: () => void;
}
```

The card displays the proposed entity in the main
section. When `existingMatches` has entries, a
divider and "Did you mean:" section lists the
matching entities as clickable MUI `ListItem`
components.

**Step 3: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Pinboard/__tests__/TagResultCard.test.tsx
```

**Commit message:**
`feat(client): add tag result card for inline
AI suggestions`

---

## Phase 6: Wiki and Entity Pages (Tasks 21-23)

Phase 6 builds the wiki-style entity pages, entity list
page, and campaign dashboard. These pages provide
non-writing views for exploring and managing the world
model.

---

### Task 21: Entity Page (Wiki View)

Build the full entity page for wiki-style exploration.
The page displays everything about an entity on a
single dense view with inline editing and chat
availability.

**Depends on:** Task 6 (WorkspaceShell), Task 13
(ChatPanel).

**Files:**

- Create: `client/src/pages/EntityPage.tsx`
- Create:
  `client/src/pages/__tests__/EntityPage.test.tsx`

**Step 1: Write failing tests**

Create
`client/src/pages/__tests__/EntityPage.test.tsx`.
Mock the entity API, relationships API, and router
params.

Write tests for the following:

- The page renders the entity name as a large heading
  with a type badge.
- The page renders the full description as markdown
  with wiki links.
- The page renders attributes (game stats).
- The page renders tags as MUI chips.
- The page renders relationships grouped by type,
  with each showing the linked entity name as a
  clickable link.
- Clicking a relationship entity name navigates to
  that entity's page (wiki-style browsing).
- The page renders GM notes (when present).
- The page renders metadata (created, updated, source
  confidence, version).
- The chat panel is available (toggle button present).
- Clicking "Edit" toggles the description to an
  editable state.

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/pages/__tests__/EntityPage.test.tsx
```

**Step 2: Implement EntityPage**

Create `client/src/pages/EntityPage.tsx`.

The page reads `campaignId` and `entityId` from the
route params. The page uses `useEntity` and
`useEntityRelationships` to fetch data.

Layout (all in a single scrollable column, max-width
900px, centred):

- `Typography` variant `h3` for the entity name with
  a `Chip` for entity type.
- `Divider`.
- Description section: renders through
  `MarkdownRenderer` (or `WorkspaceEditor` in edit
  mode).
- Attributes section: renders key-value pairs from
  the JSONB attributes object.
- Tags section: horizontal row of MUI `Chip`
  components.
- Relationships section: `Accordion` groups by
  relationship type, each listing linked entities
  as `Link` components that navigate to the target
  entity page.
- GM Notes section: renders through
  `MarkdownRenderer`.
- Metadata section: `Typography` variant `caption`
  lines for timestamps, source confidence, and
  version.

The chat panel mounts at the bottom of the page with
the scope set to `entity` and the entity ID.

**Step 3: Update routes**

Modify `client/src/App.tsx` to use `EntityPage` for
the workspace shell route
`/campaigns/:id/entities/:entityId`.

**Step 4: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/pages/__tests__/EntityPage.test.tsx
```

**Commit message:**
`feat(client): add wiki-style entity page`

---

### Task 22: Entity List Page

Build a searchable, filterable list of all entities in
a campaign with both list and card grid views.

**Depends on:** Task 6 (WorkspaceShell).

**Files:**

- Create: `client/src/pages/EntityListPage.tsx`
- Create:
  `client/src/pages/__tests__/EntityListPage.test.tsx`

**Step 1: Write failing tests**

Create
`client/src/pages/__tests__/EntityListPage.test.tsx`.
Mock the entity list API.

Write tests for the following:

- The page renders a search input.
- Typing in the search input filters entities by name.
- The page renders entity type filter chips.
- Clicking a type filter chip filters the list to
  that entity type.
- Each entity result shows name, type badge,
  description snippet, and relationship count.
- Clicking an entity result navigates to the entity
  page.
- A toggle switches between list and card grid views.

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/pages/__tests__/EntityListPage.test.tsx
```

**Step 2: Implement EntityListPage**

Create `client/src/pages/EntityListPage.tsx`.

The page uses `useEntities` with search and type
filter parameters. The page maintains local state for
the search term, selected type filter, and view mode
(list or grid).

In list mode, results render as MUI `List` with
`ListItem` components. In grid mode, results render
as a responsive grid of MUI `Card` components. Both
views use React Router `Link` for navigation.

**Step 3: Update routes**

Modify `client/src/App.tsx` to use `EntityListPage`
for the workspace shell route
`/campaigns/:id/entities`.

**Step 4: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/pages/__tests__/EntityListPage.test.tsx
```

**Commit message:**
`feat(client): add searchable entity list page`

---

### Task 23: Campaign Dashboard

Build the campaign home page with widget cards for the
campaign overview, graph visualization, recent changes
feed, and search.

**Depends on:** Task 6 (WorkspaceShell).

**Files:**

- Create: `client/src/pages/NewCampaignDashboard.tsx`
- Create:
  `client/src/pages/__tests__/NewCampaignDashboard.test.tsx`

**Step 1: Write failing tests**

Create
`client/src/pages/__tests__/NewCampaignDashboard.test.tsx`.
Mock the campaign, entity, and stats APIs.

Write tests for the following:

- The page renders the campaign name as a heading.
- The page renders a campaign overview text widget.
- The page renders a graph visualization placeholder
  widget.
- The page renders a recent changes feed widget.
- The page renders a search box that links to the
  entity list with a query parameter.
- The page renders a scratchpad access button.
- The page renders a token usage summary widget.
- Each widget renders as a distinct MUI `Card`.

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/pages/__tests__/NewCampaignDashboard.test.tsx
```

**Step 2: Implement NewCampaignDashboard**

Create `client/src/pages/NewCampaignDashboard.tsx`.

The page uses a responsive MUI `Grid` container with
widget cards. Each widget is a self-contained
component that fetches its own data.

Widget components (defined inline or as separate
components within the file):

- `OverviewWidget` renders the campaign description
  as markdown.
- `GraphWidget` renders a placeholder card with
  "Entity Graph" title and a link to the full graph
  explorer page.
- `RecentChangesWidget` uses the entity list sorted
  by `updatedAt` to show the five most recently
  modified entities.
- `SearchWidget` renders a search input that
  navigates to `/campaigns/:id/entities?q=<query>`.
- `ScratchpadWidget` renders a button that navigates
  to `/campaigns/:id/scratchpad`.
- `TokenUsageWidget` renders a summary of token usage
  from `useCampaignStats`.
- `MapWidget` renders a placeholder card with a link
  to the map page.

**Step 3: Update routes**

Modify `client/src/App.tsx` to use
`NewCampaignDashboard` for the workspace shell route
`/campaigns/:id`.

**Step 4: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/pages/__tests__/NewCampaignDashboard.test.tsx
```

**Commit message:**
`feat(client): add campaign dashboard with widgets`

---

## Phase 7: Scratchpad (Task 24)

Phase 7 builds the persistent per-campaign idea drawer
that the GM can access from any page.

---

### Task 24: Scratchpad Drawer

Build a persistent per-campaign scratchpad that slides
in from the right side of the screen. The initial
version uses `localStorage` for persistence, with a
documented TODO for server-side storage.

**Depends on:** Task 6 (WorkspaceShell).

**Files:**

- Create:
  `client/src/components/Scratchpad/ScratchpadDrawer.tsx`
- Create:
  `client/src/components/Scratchpad/ScratchpadNote.tsx`
- Create:
  `client/src/components/Scratchpad/__tests__/ScratchpadDrawer.test.tsx`

**Step 1: Write failing tests**

Create
`client/src/components/Scratchpad/__tests__/ScratchpadDrawer.test.tsx`.

Write tests for the following:

- The drawer is closed by default.
- Opening the drawer renders the note list.
- Clicking "Add note" creates a new empty note.
- Typing in a note updates the note content.
- Clicking "Delete" removes the note from the list.
- Notes persist to localStorage keyed by campaign ID.
- Closing and reopening the drawer shows the
  previously saved notes.

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Scratchpad/__tests__/ScratchpadDrawer.test.tsx
```

**Step 2: Implement ScratchpadNote**

Create
`client/src/components/Scratchpad/ScratchpadNote.tsx`.

The component renders an MUI `TextField` for the note
body with a delete `IconButton`. The component
accepts `value`, `onChange`, and `onDelete` props.

**Step 3: Implement ScratchpadDrawer**

Create
`client/src/components/Scratchpad/ScratchpadDrawer.tsx`.

The drawer uses MUI `Drawer` with `anchor="right"`.
Props:

```typescript
interface ScratchpadDrawerProps {
    campaignId: number;
    open: boolean;
    onClose: () => void;
}
```

Note state management:

```typescript
interface ScratchpadNote {
    id: string;
    content: string;
    createdAt: string;
}
```

The component loads notes from localStorage on mount
using the key `imagineer_scratchpad_${campaignId}`.
The component saves notes to localStorage on every
change using a debounced effect.

The drawer header shows "Scratchpad" and an "Add
note" button. The body renders a scrollable list of
`ScratchpadNote` components.

Include a `// TODO: Replace localStorage with
server-side persistence` comment at the top of the
persistence logic.

**Step 4: Add scratchpad button to WorkspaceShell**

Modify `client/src/layouts/WorkspaceShell.tsx` to add
a scratchpad `IconButton` (using the `NoteAlt` icon)
in the header. Clicking the button opens the
`ScratchpadDrawer`.

**Step 5: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Scratchpad/__tests__/ScratchpadDrawer.test.tsx
```

**Commit message:**
`feat(client): add scratchpad drawer with
localStorage persistence`

---

## Phase 8: Graph Explorer (Task 25)

Phase 8 builds the full-screen interactive entity
relationship graph.

---

### Task 25: Graph Explorer

Build the full-screen graph visualization that displays
entities as nodes and relationships as edges. The graph
uses Cytoscape.js for rendering and interaction.

**Depends on:** Task 6 (WorkspaceShell), Task 8
(routes).

**New dependency:** Install `cytoscape` and
`react-cytoscapejs` for graph rendering.

**Files:**

- Create: `client/src/pages/GraphExplorer.tsx`
- Create:
  `client/src/pages/__tests__/GraphExplorer.test.tsx`

**Step 1: Install Cytoscape dependencies**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npm install cytoscape react-cytoscapejs \
    && npm install -D @types/cytoscape \
    @types/react-cytoscapejs
```

**Step 2: Write failing tests**

Create
`client/src/pages/__tests__/GraphExplorer.test.tsx`.
Mock the entity and relationship APIs to return test
data.

Write tests for the following:

- The component renders a Cytoscape graph container.
- Each entity appears as a node.
- Each relationship appears as an edge.
- Nodes are coloured by entity type.
- Clicking a node shows entity details in a sidebar
  panel.
- Double-clicking a node navigates to the entity
  page.
- The page renders zoom controls.
- A type filter allows hiding nodes of a specific
  entity type.

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/pages/__tests__/GraphExplorer.test.tsx
```

**Step 3: Implement GraphExplorer**

Create `client/src/pages/GraphExplorer.tsx`.

The page reads `campaignId` from the route params.
The page uses `useEntities` and `useRelationships`
to fetch all entities and relationships for the
campaign.

Transform entities and relationships into Cytoscape
elements:

```typescript
const elements = [
    // Nodes
    ...entities.map((e) => ({
        data: {
            id: String(e.id),
            label: e.name,
            type: e.entityType,
        },
    })),
    // Edges
    ...relationships.map((r) => ({
        data: {
            id: `e${r.id}`,
            source: String(r.sourceEntityId),
            target: String(r.targetEntityId),
            label: r.relationshipType,
        },
    })),
];
```

Configure the graph with a `cose` (force-directed)
layout. Style nodes with colours based on entity
type (NPCs = blue, locations = green, items = amber,
factions = purple). Style edges with labels.

The sidebar detail panel shows entity name, type,
description, and relationship count when a node is
selected.

**Step 4: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/pages/__tests__/GraphExplorer.test.tsx
```

**Commit message:**
`feat(client): add interactive graph explorer with
Cytoscape.js`

---

## Phase 9: Map (Task 26)

Phase 9 builds the initial location relationship map.

---

### Task 26: Location Map (Initial Version)

Build the initial relationship and travel map that
visualises location entities and their connections.
The map reuses the Cytoscape foundation from Task 25
with a filtered data set and a geographic-style layout.

**Depends on:** Task 25 (GraphExplorer, Cytoscape
setup).

**Files:**

- Create: `client/src/pages/MapView.tsx`
- Create:
  `client/src/pages/__tests__/MapView.test.tsx`

**Step 1: Write failing tests**

Create
`client/src/pages/__tests__/MapView.test.tsx`.
Mock the entity API to return location-type entities
and the relationship API to return relationships
between locations.

Write tests for the following:

- The component renders a Cytoscape graph container.
- Only location entities appear as nodes (other types
  are excluded).
- Relationships between locations appear as edges.
- Clicking a node shows location details.
- The page title indicates "Map" context.

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/pages/__tests__/MapView.test.tsx
```

**Step 2: Implement MapView**

Create `client/src/pages/MapView.tsx`.

The page reuses the Cytoscape rendering pattern from
`GraphExplorer.tsx` but filters entities to the
`location` type only. Relationships between
non-location entities are excluded from the edge set.

Use the `grid` layout initially (because locations
lack coordinate data). Add a `// TODO: Switch to
coordinate-based layout when location entities have
position data` comment.

The detail sidebar shows location name, description,
and connected locations.

**Step 3: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/pages/__tests__/MapView.test.tsx
```

**Commit message:**
`feat(client): add initial location map view`

---

## Phase 10: Integration and Polish (Tasks 27-29)

Phase 10 wires the full structural change pipeline,
cleans up deprecated code, and adds end-to-end smoke
tests.

---

### Task 27: Structural Change Pipeline Integration

Wire the Draft, Revise, Approve, Commit pipeline across
the editor, pinboard, and chat layers. The
`StructuralChangeManager` component watches for pending
changes and provides a review panel for approving or
dismissing changes in bulk.

**Depends on:** Task 4 (WorkspaceContext), Task 18
(EntityCard), Task 13 (ChatPanel).

**Files:**

- Create:
  `client/src/components/Workspace/StructuralChangeManager.tsx`
- Create:
  `client/src/components/Workspace/__tests__/StructuralChangeManager.test.tsx`

**Step 1: Write failing tests**

Create
`client/src/components/Workspace/__tests__/StructuralChangeManager.test.tsx`.
Provide a workspace context with pending changes.

Write tests for the following:

- The component renders a badge showing "N pending
  changes" when changes exist.
- Clicking the badge opens a review panel listing
  all pending changes.
- Each change in the review panel has "Approve" and
  "Dismiss" buttons.
- Clicking "Approve" on a `create_entity` change
  calls the entity creation API.
- Clicking "Dismiss" removes the change from the
  pending list.
- Clicking "Approve All" approves every pending
  change and calls the respective APIs.
- On successful entity creation, the entity cache
  refreshes.
- The component renders nothing when no pending
  changes exist.

Run to verify failure:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Workspace/__tests__/StructuralChangeManager.test.tsx
```

**Step 2: Implement StructuralChangeManager**

Create
`client/src/components/Workspace/StructuralChangeManager.tsx`.

The component reads `pendingChanges` from the
workspace context. When changes exist, the component
renders an MUI `Badge` on a `Fab` button positioned
at the bottom-right of the workspace.

Clicking the button opens an MUI `Drawer` from the
right with a list of changes. Each change displays
a description, type icon, and action buttons.

The "Approve" action dispatches to the appropriate
API based on the change type:

- `create_entity` calls `useCreateEntity`.
- `create_relationship` calls
  `useCreateRelationship`.
- `link_entity` calls `useUpdateEntity` to add
  the wiki link.

On success, the component calls
`approvePendingChange` on the context and
invalidates the entity query cache.

The "Approve All" button iterates all pending
changes and approves each sequentially.

**Step 3: Integrate into WorkspacePage**

Modify
`client/src/components/Workspace/WorkspacePage.tsx`
to include `StructuralChangeManager` as a child
component.

**Step 4: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/components/Workspace/__tests__/StructuralChangeManager.test.tsx
```

**Commit message:**
`feat(client): add structural change pipeline
manager`

---

### Task 28: Deprecated Page Cleanup

Move old pages that the workspace replaces to a
`deprecated` directory. Update route imports and remove
references to deprecated components.

**Depends on:** All prior tasks.

**Files:**

- Create: `client/src/deprecated/` directory
- Move the following files to `client/src/deprecated/`:
  - `pages/AnalysisWizard.tsx` and its test
  - `pages/IdentifyPhasePage.tsx` and its test
  - `pages/RevisePhasePage.tsx` and its test
  - `pages/EnrichPhasePage.tsx` and its test
  - `pages/EntityEditor.tsx`
  - `pages/ChapterEditorPage.tsx`
  - `pages/SessionEditorPage.tsx` and its test
  - `components/SaveSplitButton/` directory
  - `contexts/AnalysisWizardContext.tsx`
  - `layouts/AppShell.tsx`
- Modify: `client/src/App.tsx` (remove old routes)
- Modify: `client/src/hooks/index.ts`
  (remove analysis wizard exports)
- Modify: `client/src/layouts/index.ts`
  (remove AppShell export)

**Step 1: Create the deprecated directory**

```bash
mkdir -p \
    /Users/antonypegg/PROJECTS/imagineer/client/src/deprecated
```

**Step 2: Move deprecated files**

Move each file to the deprecated directory. Use `git
mv` to preserve history:

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client/src \
    && git mv pages/AnalysisWizard.tsx \
    deprecated/ \
    && git mv pages/AnalysisWizard.test.tsx \
    deprecated/ \
    && git mv pages/IdentifyPhasePage.tsx \
    deprecated/ \
    && git mv pages/IdentifyPhasePage.test.tsx \
    deprecated/ \
    && git mv pages/RevisePhasePage.tsx \
    deprecated/ \
    && git mv pages/RevisePhasePage.test.tsx \
    deprecated/ \
    && git mv pages/EnrichPhasePage.tsx \
    deprecated/ \
    && git mv pages/EnrichPhasePage.test.tsx \
    deprecated/ \
    && git mv pages/EntityEditor.tsx \
    deprecated/ \
    && git mv pages/ChapterEditorPage.tsx \
    deprecated/ \
    && git mv pages/SessionEditorPage.tsx \
    deprecated/ \
    && git mv pages/SessionEditorPage.test.tsx \
    deprecated/ \
    && git mv components/SaveSplitButton \
    deprecated/ \
    && git mv contexts/AnalysisWizardContext.tsx \
    deprecated/ \
    && git mv layouts/AppShell.tsx \
    deprecated/
```

**Step 3: Remove old routes from App.tsx**

Remove the `AppShellWrapper` routes and the
`FullScreenWrapper` routes for deprecated pages.
Remove the `AnalysisWizardRedirect` component.
Remove imports for all moved files.

**Step 4: Update index exports**

Remove deprecated exports from:

- `client/src/hooks/index.ts` (remove
  `useAnalysisWizard` and analysis-related hooks).
- `client/src/layouts/index.ts` (remove `AppShell`).
- `client/src/api/index.ts` (remove
  `contentAnalysisApi` if no longer used).

**Step 5: Verify no broken imports**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx tsc --noEmit
```

**Step 6: Verify all tests pass**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run
```

**Commit message:**
`refactor(client): move deprecated pages to
deprecated directory`

---

### Task 29: End-to-End Smoke Tests

Write integration tests that verify the core workspace
workflows with mocked API responses. The tests exercise
the full component tree from workspace page down to
individual cards and chat messages.

**Depends on:** All prior tasks.

**Files:**

- Create:
  `client/src/tests/integration/workspace.test.tsx`

**Step 1: Write the integration test file**

Create
`client/src/tests/integration/workspace.test.tsx`.

Set up a shared test harness that wraps the
`WorkspacePage` in all required providers:
`QueryClientProvider`, `WorkspaceProvider`,
`MemoryRouter`, and `CampaignProvider`. Mock all
API modules (`conversationsApi`, `entitiesApi`,
`relationshipsApi`).

**Step 2: Write and tag flow test**

Write a test that verifies the following sequence:

1. Render the workspace with initial content.
2. Type text containing an `@[need a tavern NPC]`
   tag.
3. Verify the tag detection fires.
4. Mock the SSE stream to return an entity suggestion.
5. Verify a ghost entity card appears on the
   pinboard.
6. Click "Approve" on the ghost card.
7. Verify the entity creation API is called.

**Step 3: Chat flow test**

Write a test that verifies the following sequence:

1. Render the workspace.
2. Toggle the chat panel open.
3. Type a message and click send.
4. Mock the SSE stream to return text deltas.
5. Verify the streaming message appears character
   by character.
6. Verify the token cost badge appears after the
   stream completes.

**Step 4: Wiki navigation test**

Write a test that verifies the following sequence:

1. Render the entity page for a test entity with
   relationships.
2. Click a relationship's linked entity name.
3. Verify the URL changes to the target entity's
   page.

**Step 5: Review flow test**

Write a test that verifies the following sequence:

1. Render the workspace with document content.
2. Click the "Review" button.
3. Confirm the review dialog.
4. Mock the SSE stream to return findings.
5. Verify finding cards appear on the pinboard.

**Step 6: Structural pipeline test**

Write a test that verifies the following sequence:

1. Render the workspace with a pending ghost entity
   in the workspace context.
2. Click the pending changes badge.
3. Click "Approve" on the entity.
4. Verify the entity creation API is called.
5. Verify the ghost card transitions to a real card.

**Step 7: Run all integration tests**

```bash
cd /Users/antonypegg/PROJECTS/imagineer/client \
    && npx vitest run \
    src/tests/integration/workspace.test.tsx
```

**Commit message:**
`test(client): add end-to-end workspace integration
tests`

---

## Task Dependency Summary

The following table summarises the dependency
relationships between tasks. A task may begin only after
all tasks listed in the "Depends on" column have been
completed.

| Task | Name                            | Depends on   |
|------|---------------------------------|--------------|
| 1    | Conversation Types and API      | None         |
| 2    | Conversation React Query Hooks  | 1            |
| 3    | SSE Streaming Hook              | 1            |
| 4    | Workspace Context               | 1            |
| 5    | Entity Name Detection Hook      | 1            |
| 6    | Workspace Shell                 | 4            |
| 7    | Workspace Page Component        | 4, 6         |
| 8    | Route Restructuring             | 6, 7         |
| 9    | Enhanced Tiptap Editor          | 4, 7         |
| 10   | Inline Tag Detection Extension  | 9            |
| 11   | Entity Detection Integration    | 5, 9         |
| 12   | Editor-Chat Selection Bridge    | 4, 9         |
| 13   | Chat Panel Component            | 3, 4, 12     |
| 14   | Token Cost Display              | 1            |
| 15   | Card Pull-In Mechanism          | 4, 13        |
| 16   | Deliberate Review Request       | 3, 9, 13     |
| 17   | Pinboard Container              | 4, 7         |
| 18   | Entity Pinboard Card            | 17           |
| 19   | Finding Card                    | 16, 17       |
| 20   | Inline Tag Result Cards         | 10, 18       |
| 21   | Entity Page (Wiki View)         | 6, 13        |
| 22   | Entity List Page                | 6            |
| 23   | Campaign Dashboard              | 6            |
| 24   | Scratchpad Drawer               | 6            |
| 25   | Graph Explorer                  | 6, 8         |
| 26   | Location Map                    | 25           |
| 27   | Structural Change Pipeline      | 4, 13, 18    |
| 28   | Deprecated Page Cleanup         | All prior     |
| 29   | End-to-End Smoke Tests          | All prior     |

## New Package Dependencies

The following npm packages must be installed during
implementation. Each task that requires a new package
includes the install command.

- `@dnd-kit/core` and `@dnd-kit/utilities` for pinboard
  drag-and-drop (Task 17).
- `cytoscape`, `react-cytoscapejs`,
  `@types/cytoscape`, and `@types/react-cytoscapejs`
  for graph visualization (Task 25).

## File Summary

The following list enumerates every file that this plan
creates or modifies, grouped by directory.

Types:

- Modify: `client/src/types/index.ts`

API layer:

- Create: `client/src/api/conversations.ts`
- Create:
  `client/src/api/__tests__/conversations.test.ts`
- Modify: `client/src/api/index.ts`

Hooks:

- Create: `client/src/hooks/useConversations.ts`
- Create:
  `client/src/hooks/__tests__/useConversations.test.ts`
- Create: `client/src/hooks/useSSEStream.ts`
- Create:
  `client/src/hooks/__tests__/useSSEStream.test.ts`
- Create:
  `client/src/hooks/useInlineEntityDetection.ts`
- Create:
  `client/src/hooks/__tests__/useInlineEntityDetection.test.ts`
- Modify: `client/src/hooks/index.ts`

Contexts:

- Create:
  `client/src/contexts/WorkspaceContext.tsx`
- Create:
  `client/src/contexts/__tests__/WorkspaceContext.test.tsx`

Layouts:

- Create: `client/src/layouts/WorkspaceShell.tsx`
- Create:
  `client/src/layouts/__tests__/WorkspaceShell.test.tsx`
- Modify: `client/src/layouts/index.ts`

Editor components:

- Create:
  `client/src/components/Editor/WorkspaceEditor.tsx`
- Create:
  `client/src/components/Editor/SelectionTracker.ts`
- Create: `client/src/components/Editor/DraftMark.ts`
- Create:
  `client/src/components/Editor/InlineTagExtension.ts`
- Create:
  `client/src/components/Editor/EntityDetectionPlugin.ts`
- Create:
  `client/src/components/Editor/EntitySuggestionPopover.tsx`
- Create:
  `client/src/components/Editor/__tests__/WorkspaceEditor.test.tsx`
- Create:
  `client/src/components/Editor/__tests__/InlineTagExtension.test.ts`
- Create:
  `client/src/components/Editor/__tests__/EntityDetectionPlugin.test.ts`

Chat components:

- Create: `client/src/components/Chat/ChatPanel.tsx`
- Create: `client/src/components/Chat/ChatMessage.tsx`
- Create: `client/src/components/Chat/ChatInput.tsx`
- Create:
  `client/src/components/Chat/StreamingMessage.tsx`
- Create:
  `client/src/components/Chat/ToolUseIndicator.tsx`
- Create:
  `client/src/components/Chat/SelectionPreview.tsx`
- Create:
  `client/src/components/Chat/TokenCostBadge.tsx`
- Create: `client/src/components/Chat/CardContext.tsx`
- Create:
  `client/src/components/Chat/ReviewRequest.tsx`
- Create:
  `client/src/components/Chat/__tests__/ChatPanel.test.tsx`
- Create:
  `client/src/components/Chat/__tests__/ChatMessage.test.tsx`
- Create:
  `client/src/components/Chat/__tests__/ChatInput.test.tsx`
- Create:
  `client/src/components/Chat/__tests__/SelectionPreview.test.tsx`
- Create:
  `client/src/components/Chat/__tests__/TokenCostBadge.test.tsx`
- Create:
  `client/src/components/Chat/__tests__/CardContext.test.tsx`
- Create:
  `client/src/components/Chat/__tests__/ReviewRequest.test.tsx`

Pinboard components:

- Create:
  `client/src/components/Pinboard/PinboardContainer.tsx`
- Create:
  `client/src/components/Pinboard/EntityCard.tsx`
- Create:
  `client/src/components/Pinboard/FindingCard.tsx`
- Create:
  `client/src/components/Pinboard/TagResultCard.tsx`
- Create:
  `client/src/components/Pinboard/__tests__/PinboardContainer.test.tsx`
- Create:
  `client/src/components/Pinboard/__tests__/EntityCard.test.tsx`
- Create:
  `client/src/components/Pinboard/__tests__/FindingCard.test.tsx`
- Create:
  `client/src/components/Pinboard/__tests__/TagResultCard.test.tsx`

Workspace components:

- Create:
  `client/src/components/Workspace/WorkspacePage.tsx`
- Create:
  `client/src/components/Workspace/StructuralChangeManager.tsx`
- Create:
  `client/src/components/Workspace/__tests__/WorkspacePage.test.tsx`
- Create:
  `client/src/components/Workspace/__tests__/StructuralChangeManager.test.tsx`

Scratchpad components:

- Create:
  `client/src/components/Scratchpad/ScratchpadDrawer.tsx`
- Create:
  `client/src/components/Scratchpad/ScratchpadNote.tsx`
- Create:
  `client/src/components/Scratchpad/__tests__/ScratchpadDrawer.test.tsx`

Global components:

- Create: `client/src/components/GlobalSearch.tsx`
- Create:
  `client/src/components/__tests__/GlobalSearch.test.tsx`

Pages:

- Create: `client/src/pages/WorkspaceRoutePage.tsx`
- Create: `client/src/pages/EntityPage.tsx`
- Create: `client/src/pages/EntityListPage.tsx`
- Create:
  `client/src/pages/NewCampaignDashboard.tsx`
- Create: `client/src/pages/GraphExplorer.tsx`
- Create: `client/src/pages/MapView.tsx`
- Create:
  `client/src/pages/__tests__/EntityPage.test.tsx`
- Create:
  `client/src/pages/__tests__/EntityListPage.test.tsx`
- Create:
  `client/src/pages/__tests__/NewCampaignDashboard.test.tsx`
- Create:
  `client/src/pages/__tests__/GraphExplorer.test.tsx`
- Create:
  `client/src/pages/__tests__/MapView.test.tsx`

Integration tests:

- Create:
  `client/src/tests/integration/workspace.test.tsx`

App and routing:

- Modify: `client/src/App.tsx`
- Create: `client/src/App.test.tsx`

Deprecated:

- Create: `client/src/deprecated/` directory
- Move multiple files (see Task 28)
