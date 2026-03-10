<!--
  Imagineer - TTRPG Campaign Intelligence Platform

  Copyright (c) 2025 - 2026
  This software is released under The MIT License
-->

# Phase 1 Conversational Backend Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use
> superpowers:executing-plans to implement this plan
> task-by-task.

**Goal:** Build the server-side conversation layer that
transforms Imagineer from an editor-centric CRUD app into a
conversational world-building platform.

**Architecture:** An orchestrator manages a streaming tool use
loop — assembling context, calling the Anthropic streaming API,
executing tools against the world model, and feeding results
back until the LLM produces a final response. All exchanges
stream to the client via SSE. Three existing expert agents
(TTRPG, Canon, Graph) are wrapped as agent tools with their
own sub-tool access for world-grounded generation.

**Tech Stack:** Go 1.22+, PostgreSQL (pgx v5), Anthropic
Messages API (streaming), chi router, SSE, testify.

**Design Document:**
`docs/plans/2026-03-09-conversational-backend-design.md`

---

## Task 1: Database Migration

Create the migration adding conversations, messages, and
token_usage_log tables.

**Files:**

- Create: `migrations/008_conversations.sql`

**Step 1: Write the migration**

```sql
/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

-- ============================================
-- Migration 008: Conversation Layer
--
-- Adds conversations, messages, and
-- token_usage_log tables with full database
-- support: triggers, indexes, views, and
-- stored functions for the Phase 1
-- conversational backend.
-- ============================================

BEGIN;

-- ============================================
-- Section 1: Tables
-- ============================================

CREATE TABLE conversations (
    id             BIGSERIAL PRIMARY KEY,
    campaign_id    BIGINT NOT NULL
                   REFERENCES campaigns(id)
                   ON DELETE CASCADE,
    scope_type     TEXT NOT NULL
                   CHECK (scope_type IN (
                       'entity', 'chapter',
                       'session', 'scene',
                       'campaign'
                   )),
    scope_id       BIGINT NOT NULL,
    title          TEXT,
    summary        TEXT,
    summary_tokens INT NOT NULL DEFAULT 0
                   CHECK (summary_tokens >= 0),
    created_at     TIMESTAMPTZ NOT NULL
                   DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL
                   DEFAULT NOW(),

    CONSTRAINT uq_conversation_scope
        UNIQUE (campaign_id, scope_type, scope_id)
);

COMMENT ON TABLE conversations IS
    'Chat threads scoped to a campaign document '
    'or entity. One thread per scope. Access via '
    'get_or_create_conversation() only.';
COMMENT ON COLUMN conversations.scope_type IS
    'entity, chapter, session, scene, or campaign';
COMMENT ON COLUMN conversations.scope_id IS
    'PK of the scoped record. For campaign scope '
    'this equals campaign_id. Not a FK because '
    'the target table varies by scope_type.';
COMMENT ON COLUMN conversations.summary IS
    'Rolling summary of compacted messages. '
    'NULL until first compaction.';
COMMENT ON COLUMN conversations.summary_tokens IS
    'Token count of current summary for context '
    'window budget calculation.';
COMMENT ON CONSTRAINT uq_conversation_scope
    ON conversations IS
    'One thread per scope. Used by '
    'get_or_create_conversation() ON CONFLICT.';

CREATE TABLE messages (
    id              BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL
                    REFERENCES conversations(id)
                    ON DELETE CASCADE,
    role            TEXT NOT NULL
                    CHECK (role IN (
                        'user', 'assistant',
                        'system', 'tool_call',
                        'tool_result'
                    )),
    content         TEXT NOT NULL DEFAULT '',
    tool_name       TEXT,
    tool_use_id     TEXT,
    tool_input      JSONB,
    tool_result     JSONB,
    tokens          INT CHECK (tokens >= 0),
    compacted       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL
                    DEFAULT NOW(),

    CONSTRAINT tool_fields_consistent CHECK (
        (role IN ('tool_call', 'tool_result')
            AND tool_name IS NOT NULL)
        OR role NOT IN ('tool_call', 'tool_result')
    )
);

COMMENT ON TABLE messages IS
    'Individual messages within a conversation. '
    'Never deleted. Compacted messages excluded '
    'from context assembly but retained for '
    'full transcript retrieval.';
COMMENT ON COLUMN messages.role IS
    'user, assistant, system, tool_call, or '
    'tool_result';
COMMENT ON COLUMN messages.content IS
    'Message text. Empty string for pure '
    'tool_call messages.';
COMMENT ON COLUMN messages.tool_use_id IS
    'Anthropic tool_use_id correlating a '
    'tool_call to its tool_result. Required '
    'for context reconstruction.';
COMMENT ON COLUMN messages.compacted IS
    'TRUE = incorporated into summary, excluded '
    'from live context. Set by compact_messages().';
COMMENT ON CONSTRAINT tool_fields_consistent
    ON messages IS
    'tool_name required for tool_call/tool_result.';

CREATE TABLE token_usage_log (
    id              BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL
                    REFERENCES conversations(id)
                    ON DELETE RESTRICT,
    campaign_id     BIGINT NOT NULL
                    REFERENCES campaigns(id)
                    ON DELETE RESTRICT,
    user_id         BIGINT NOT NULL
                    REFERENCES users(id)
                    ON DELETE RESTRICT,
    model           TEXT NOT NULL,
    input_tokens    INT NOT NULL
                    CHECK (input_tokens >= 0),
    output_tokens   INT NOT NULL
                    CHECK (output_tokens >= 0),
    total_tokens    INT NOT NULL
                    CHECK (total_tokens >= 0),
    llm_call_type   TEXT NOT NULL
                    CHECK (llm_call_type IN (
                        'conversation',
                        'compaction',
                        'agent_tool'
                    )),
    agent_name      TEXT,
    created_at      TIMESTAMPTZ NOT NULL
                    DEFAULT NOW(),

    CONSTRAINT total_tokens_consistent CHECK (
        total_tokens = input_tokens + output_tokens
    ),
    CONSTRAINT agent_name_when_agent_tool CHECK (
        llm_call_type != 'agent_tool'
        OR agent_name IS NOT NULL
    )
);

COMMENT ON TABLE token_usage_log IS
    'Append-only ledger of every LLM call. '
    'ON DELETE RESTRICT — billing records must '
    'not be silently removed.';
COMMENT ON COLUMN token_usage_log.llm_call_type IS
    'conversation, compaction, or agent_tool';
COMMENT ON COLUMN token_usage_log.agent_name IS
    'For agent_tool: ttrpg_expert, canon_expert, '
    'or graph_expert. NULL otherwise.';
COMMENT ON CONSTRAINT total_tokens_consistent
    ON token_usage_log IS
    'total = input + output at write time.';
COMMENT ON CONSTRAINT agent_name_when_agent_tool
    ON token_usage_log IS
    'agent_name required when type is agent_tool.';

-- ============================================
-- Section 2: Triggers
-- ============================================

CREATE TRIGGER update_conversations_updated_at
    BEFORE UPDATE ON conversations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TRIGGER update_conversations_updated_at
    ON conversations IS
    'Keeps updated_at current on every UPDATE.';

CREATE OR REPLACE FUNCTION
touch_conversation_on_message()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE conversations
    SET updated_at = NOW()
    WHERE id = NEW.conversation_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION touch_conversation_on_message()
    IS 'Bumps conversations.updated_at on new '
       'message insert.';

CREATE TRIGGER touch_conversation_updated_at
    AFTER INSERT ON messages
    FOR EACH ROW
    EXECUTE FUNCTION
        touch_conversation_on_message();

COMMENT ON TRIGGER touch_conversation_updated_at
    ON messages IS
    'INSERT only. Keeps conversation list sort '
    'order accurate.';

-- ============================================
-- Section 3: Indexes
-- ============================================

CREATE INDEX idx_conversations_campaign
    ON conversations(campaign_id,
                     updated_at DESC);
COMMENT ON INDEX idx_conversations_campaign IS
    'Conversation list ordered by last activity.';

CREATE INDEX idx_messages_context_window
    ON messages(conversation_id, created_at)
    WHERE compacted = FALSE;
COMMENT ON INDEX idx_messages_context_window IS
    'Hot-path partial index for context assembly. '
    'Covers only uncompacted messages.';

CREATE INDEX idx_messages_conversation_full
    ON messages(conversation_id, created_at);
COMMENT ON INDEX idx_messages_conversation_full IS
    'Full ordered scan for transcript view.';

CREATE INDEX idx_messages_compaction_candidates
    ON messages(conversation_id, id)
    WHERE compacted = FALSE;
COMMENT ON INDEX idx_messages_compaction_candidates
    IS 'Oldest uncompacted messages for '
       'compaction.';

CREATE INDEX idx_token_usage_campaign_time
    ON token_usage_log(campaign_id, created_at)
    INCLUDE (input_tokens, output_tokens,
             total_tokens, llm_call_type);
COMMENT ON INDEX idx_token_usage_campaign_time IS
    'Covering index for campaign token '
    'aggregation. Index-only scan for SUM '
    'queries.';

CREATE INDEX idx_token_usage_user_time
    ON token_usage_log(user_id, created_at)
    INCLUDE (input_tokens, output_tokens,
             total_tokens, llm_call_type);
COMMENT ON INDEX idx_token_usage_user_time IS
    'Covering index for per-user token '
    'aggregation.';

CREATE INDEX idx_token_usage_conversation
    ON token_usage_log(conversation_id)
    INCLUDE (total_tokens, llm_call_type,
             created_at);
COMMENT ON INDEX idx_token_usage_conversation IS
    'Per-conversation token total queries.';

-- ============================================
-- Section 4: Views
-- ============================================

CREATE VIEW conversation_list AS
SELECT
    c.id,
    c.campaign_id,
    c.scope_type,
    c.scope_id,
    c.title,
    c.summary_tokens,
    c.created_at,
    c.updated_at,
    m.role                 AS last_message_role,
    LEFT(m.content, 200)   AS last_message_preview,
    m.created_at           AS last_message_at,
    COALESCE(counts.total_messages, 0)
                           AS total_messages,
    COALESCE(counts.uncompacted_messages, 0)
                           AS uncompacted_messages
FROM conversations c
LEFT JOIN LATERAL (
    SELECT role, content, created_at
    FROM messages
    WHERE conversation_id = c.id
      AND compacted = FALSE
      AND role IN ('user', 'assistant')
    ORDER BY created_at DESC
    LIMIT 1
) m ON TRUE
LEFT JOIN LATERAL (
    SELECT
        COUNT(*)              AS total_messages,
        COUNT(*) FILTER (
            WHERE NOT compacted
        )                     AS uncompacted_messages
    FROM messages
    WHERE conversation_id = c.id
) counts ON TRUE;

COMMENT ON VIEW conversation_list IS
    'Conversation list with message preview '
    'and counts. One row per conversation.';

CREATE VIEW conversation_context AS
SELECT
    c.id                  AS conversation_id,
    c.campaign_id,
    c.scope_type,
    c.scope_id,
    c.title,
    c.summary,
    c.summary_tokens,
    m.id                  AS message_id,
    m.role,
    m.content,
    m.tool_name,
    m.tool_use_id,
    m.tool_input,
    m.tool_result,
    m.tokens              AS message_tokens,
    m.created_at          AS message_created_at
FROM conversations c
JOIN messages m
    ON m.conversation_id = c.id
   AND m.compacted = FALSE
ORDER BY c.id, m.created_at;

COMMENT ON VIEW conversation_context IS
    'Uncompacted messages with conversation '
    'metadata. Filter by conversation_id.';

CREATE VIEW token_usage_by_campaign AS
SELECT
    campaign_id,
    DATE_TRUNC('day', created_at)
                          AS usage_date,
    llm_call_type,
    model,
    COUNT(*)              AS call_count,
    SUM(input_tokens)     AS total_input,
    SUM(output_tokens)    AS total_output,
    SUM(total_tokens)     AS total_tokens
FROM token_usage_log
GROUP BY campaign_id,
    DATE_TRUNC('day', created_at),
    llm_call_type, model;

COMMENT ON VIEW token_usage_by_campaign IS
    'Daily token aggregation by campaign.';

CREATE VIEW token_usage_by_user AS
SELECT
    user_id,
    DATE_TRUNC('day', created_at)
                          AS usage_date,
    llm_call_type,
    model,
    COUNT(*)              AS call_count,
    SUM(input_tokens)     AS total_input,
    SUM(output_tokens)    AS total_output,
    SUM(total_tokens)     AS total_tokens
FROM token_usage_log
GROUP BY user_id,
    DATE_TRUNC('day', created_at),
    llm_call_type, model;

COMMENT ON VIEW token_usage_by_user IS
    'Daily token aggregation by user.';

-- ============================================
-- Section 5: Stored Functions
-- ============================================

CREATE OR REPLACE FUNCTION
get_or_create_conversation(
    p_campaign_id BIGINT,
    p_scope_type  TEXT,
    p_scope_id    BIGINT
)
RETURNS TABLE (
    id             BIGINT,
    campaign_id    BIGINT,
    scope_type     TEXT,
    scope_id       BIGINT,
    title          TEXT,
    summary        TEXT,
    summary_tokens INT,
    created_at     TIMESTAMPTZ,
    updated_at     TIMESTAMPTZ,
    was_created    BOOLEAN
) AS $$
DECLARE
    v_id      BIGINT;
    v_created BOOLEAN;
BEGIN
    INSERT INTO conversations
        (campaign_id, scope_type, scope_id)
    VALUES
        (p_campaign_id, p_scope_type, p_scope_id)
    ON CONFLICT ON CONSTRAINT uq_conversation_scope
    DO NOTHING
    RETURNING conversations.id INTO v_id;

    IF v_id IS NOT NULL THEN
        v_created := TRUE;
    ELSE
        v_created := FALSE;
        SELECT conversations.id INTO v_id
        FROM conversations
        WHERE conversations.campaign_id
                  = p_campaign_id
          AND conversations.scope_type
                  = p_scope_type
          AND conversations.scope_id
                  = p_scope_id;
    END IF;

    RETURN QUERY
    SELECT c.id, c.campaign_id,
        c.scope_type::TEXT, c.scope_id,
        c.title, c.summary, c.summary_tokens,
        c.created_at, c.updated_at, v_created
    FROM conversations c
    WHERE c.id = v_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION
    get_or_create_conversation(
        BIGINT, TEXT, BIGINT)
    IS 'Atomic upsert: returns existing or '
       'creates new conversation for a scope. '
       'Returns was_created for logging.';

CREATE OR REPLACE FUNCTION
assemble_conversation_context(
    p_conversation_id BIGINT
)
RETURNS TABLE (
    conversation_id          BIGINT,
    campaign_id              BIGINT,
    scope_type               TEXT,
    scope_id                 BIGINT,
    summary                  TEXT,
    summary_tokens           INT,
    uncompacted_token_estimate INT,
    message_id               BIGINT,
    role                     TEXT,
    content                  TEXT,
    tool_name                TEXT,
    tool_use_id              TEXT,
    tool_input               JSONB,
    tool_result              JSONB,
    message_tokens           INT,
    message_created_at       TIMESTAMPTZ
) AS $$
DECLARE
    v_summary_tokens INT;
    v_token_estimate INT;
BEGIN
    SELECT c.summary_tokens
    INTO v_summary_tokens
    FROM conversations c
    WHERE c.id = p_conversation_id;

    IF NOT FOUND THEN
        RETURN;
    END IF;

    SELECT v_summary_tokens
           + COALESCE(SUM(m.tokens), 0)
    INTO v_token_estimate
    FROM messages m
    WHERE m.conversation_id
              = p_conversation_id
      AND m.compacted = FALSE;

    RETURN QUERY
    SELECT c.id, c.campaign_id,
        c.scope_type::TEXT, c.scope_id,
        c.summary, c.summary_tokens,
        v_token_estimate,
        m.id, m.role::TEXT, m.content,
        m.tool_name, m.tool_use_id,
        m.tool_input, m.tool_result,
        m.tokens, m.created_at
    FROM conversations c
    JOIN messages m
        ON m.conversation_id = c.id
       AND m.compacted = FALSE
    WHERE c.id = p_conversation_id
    ORDER BY m.created_at;
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION
    assemble_conversation_context(BIGINT)
    IS 'Hot path: conversation metadata + '
       'uncompacted messages + token estimate '
       'in one round-trip. STABLE.';

CREATE OR REPLACE FUNCTION compact_messages(
    p_conversation_id BIGINT,
    p_message_ids     BIGINT[],
    p_new_summary     TEXT,
    p_summary_tokens  INT
)
RETURNS INT AS $$
DECLARE
    v_compacted_count INT;
BEGIN
    UPDATE messages
    SET compacted = TRUE
    WHERE conversation_id = p_conversation_id
      AND id = ANY(p_message_ids)
      AND compacted = FALSE;
    GET DIAGNOSTICS v_compacted_count
        = ROW_COUNT;

    UPDATE conversations
    SET summary        = p_new_summary,
        summary_tokens = p_summary_tokens
    WHERE id = p_conversation_id;

    RETURN v_compacted_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION
    compact_messages(BIGINT, BIGINT[], TEXT, INT)
    IS 'Atomic: marks messages compacted + '
       'updates summary. Both succeed or both '
       'roll back.';

CREATE OR REPLACE FUNCTION
get_token_usage_summary(
    p_campaign_id BIGINT,
    p_user_id     BIGINT,
    p_since       TIMESTAMPTZ,
    p_until       TIMESTAMPTZ DEFAULT NULL
)
RETURNS TABLE (
    llm_call_type       TEXT,
    model               TEXT,
    call_count          BIGINT,
    total_input_tokens  BIGINT,
    total_output_tokens BIGINT,
    total_tokens        BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT t.llm_call_type, t.model,
        COUNT(*)::BIGINT,
        SUM(t.input_tokens)::BIGINT,
        SUM(t.output_tokens)::BIGINT,
        SUM(t.total_tokens)::BIGINT
    FROM token_usage_log t
    WHERE (p_campaign_id IS NULL
            OR t.campaign_id = p_campaign_id)
      AND (p_user_id IS NULL
            OR t.user_id = p_user_id)
      AND t.created_at >= p_since
      AND t.created_at
              < COALESCE(p_until, NOW())
    GROUP BY t.llm_call_type, t.model
    ORDER BY SUM(t.total_tokens) DESC;
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION
    get_token_usage_summary(
        BIGINT, BIGINT, TIMESTAMPTZ,
        TIMESTAMPTZ)
    IS 'Flexible token aggregation. Filter by '
       'campaign, user, or both. p_until '
       'defaults to NOW().';

-- ============================================
-- Section 6: Record Migration
-- ============================================
INSERT INTO schema_migrations (version)
VALUES ('008_conversations');

COMMIT;
```

**Step 2: Run migration**

Run: `make migrate`

Expected: Migration 008 applies without errors.

**Step 3: Verify tables, views, and functions exist**

Run: `psql -c "\dt conversations; \dt messages; \dt token_usage_log;" imagineer`

Run: `psql -c "\dv conversation_list; \dv conversation_context; \dv token_usage_by_campaign; \dv token_usage_by_user;" imagineer`

Run: `psql -c "\df get_or_create_conversation; \df assemble_conversation_context; \df compact_messages; \df get_token_usage_summary;" imagineer`

Expected: All three tables, four views, and four
stored functions listed.

**Step 4: Commit**

```bash
git add migrations/008_conversations.sql
git commit -m "feat: add conversation layer migration with triggers, views, and stored functions (008)"
```

---

## Task 2: Conversation Models

Add Go types for conversations, messages, and token usage to
the models package.

**Files:**

- Modify: `internal/models/models.go`
- Test: `internal/models/models_test.go`

**Step 1: Write the model types**

Add to the end of `internal/models/models.go` (before the
closing of the file):

```go
// --- Conversation types ---

// ScopeType identifies what a conversation is scoped to.
type ScopeType string

const (
    ScopeTypeEntity   ScopeType = "entity"
    ScopeTypeChapter  ScopeType = "chapter"
    ScopeTypeSession  ScopeType = "session"
    ScopeTypeScene    ScopeType = "scene"
    ScopeTypeCampaign ScopeType = "campaign"
)

// MessageRole identifies the sender of a message.
type MessageRole string

const (
    RoleUser       MessageRole = "user"
    RoleAssistant  MessageRole = "assistant"
    RoleSystem     MessageRole = "system"
    RoleToolCall   MessageRole = "tool_call"
    RoleToolResult MessageRole = "tool_result"
)

// LLMCallType categorises an LLM invocation for metering.
type LLMCallType string

const (
    CallTypeConversation LLMCallType = "conversation"
    CallTypeCompaction   LLMCallType = "compaction"
    CallTypeAgentTool    LLMCallType = "agent_tool"
)

// Conversation represents a chat thread scoped to a
// campaign document or entity.
type Conversation struct {
    ID            int64     `json:"id"`
    CampaignID    int64     `json:"campaignId"`
    ScopeType     ScopeType `json:"scopeType"`
    ScopeID       int64     `json:"scopeId"`
    Title         *string   `json:"title,omitempty"`
    Summary       *string   `json:"summary,omitempty"`
    SummaryTokens int       `json:"summaryTokens"`
    CreatedAt     time.Time `json:"createdAt"`
    UpdatedAt     time.Time `json:"updatedAt"`
    Messages      []Message `json:"messages,omitempty"`
}

// CreateConversationRequest contains fields for creating
// a new conversation.
type CreateConversationRequest struct {
    ScopeType ScopeType `json:"scopeType"`
    ScopeID   int64     `json:"scopeId"`
}

// Message represents a single message in a conversation.
type Message struct {
    ID             int64           `json:"id"`
    ConversationID int64           `json:"conversationId"`
    Role           MessageRole     `json:"role"`
    Content        string          `json:"content"`
    ToolName       *string         `json:"toolName,omitempty"`
    ToolUseID      *string         `json:"toolUseId,omitempty"`
    ToolInput      json.RawMessage `json:"toolInput,omitempty"`
    ToolResult     json.RawMessage `json:"toolResult,omitempty"`
    Tokens         *int            `json:"tokens,omitempty"`
    Compacted      bool            `json:"compacted"`
    CreatedAt      time.Time       `json:"createdAt"`
}

// SendMessageRequest contains the user message and
// optional editor state.
type SendMessageRequest struct {
    Content        string  `json:"content"`
    EditorSelection *string `json:"editorSelection,omitempty"`
}

// TokenUsageLog records token consumption for a single
// LLM call.
type TokenUsageLog struct {
    ID             int64       `json:"id"`
    ConversationID int64       `json:"conversationId"`
    CampaignID     int64       `json:"campaignId"`
    UserID         int64       `json:"userId"`
    Model          string      `json:"model"`
    InputTokens    int         `json:"inputTokens"`
    OutputTokens   int         `json:"outputTokens"`
    TotalTokens    int         `json:"totalTokens"`
    LLMCallType    LLMCallType `json:"llmCallType"`
    AgentName      *string     `json:"agentName,omitempty"`
    CreatedAt      time.Time   `json:"createdAt"`
}

// ConversationContext is the return type from
// assemble_conversation_context(). Contains
// conversation metadata and all uncompacted
// messages in one struct.
type ConversationContext struct {
    Conversation  Conversation
    Messages      []Message
    TokenEstimate int
}

// ConversationListItem is a row from the
// conversation_list view, including derived
// fields for display.
type ConversationListItem struct {
    ID                  int64     `json:"id"`
    CampaignID          int64     `json:"campaignId"`
    ScopeType           ScopeType `json:"scopeType"`
    ScopeID             int64     `json:"scopeId"`
    Title               *string   `json:"title,omitempty"`
    SummaryTokens       int       `json:"summaryTokens"`
    CreatedAt           time.Time `json:"createdAt"`
    UpdatedAt           time.Time `json:"updatedAt"`
    LastMessageRole     *string   `json:"lastMessageRole,omitempty"`
    LastMessagePreview  *string   `json:"lastMessagePreview,omitempty"`
    LastMessageAt       *time.Time `json:"lastMessageAt,omitempty"`
    TotalMessages       int       `json:"totalMessages"`
    UncompactedMessages int       `json:"uncompactedMessages"`
}

// TokenUsageSummary is the return type from
// get_token_usage_summary(). One row per
// (llm_call_type, model) combination.
type TokenUsageSummary struct {
    LLMCallType       string `json:"llmCallType"`
    Model             string `json:"model"`
    CallCount         int64  `json:"callCount"`
    TotalInputTokens  int64  `json:"totalInputTokens"`
    TotalOutputTokens int64  `json:"totalOutputTokens"`
    TotalTokens       int64  `json:"totalTokens"`
}
```

**Step 2: Write tests for valid scope types**

Add to `internal/models/models_test.go`:

```go
func TestScopeTypeValues(t *testing.T) {
    valid := []models.ScopeType{
        models.ScopeTypeEntity,
        models.ScopeTypeChapter,
        models.ScopeTypeSession,
        models.ScopeTypeScene,
        models.ScopeTypeCampaign,
    }
    for _, st := range valid {
        assert.NotEmpty(t, string(st))
    }
}

func TestMessageRoleValues(t *testing.T) {
    valid := []models.MessageRole{
        models.RoleUser,
        models.RoleAssistant,
        models.RoleSystem,
        models.RoleToolCall,
        models.RoleToolResult,
    }
    for _, r := range valid {
        assert.NotEmpty(t, string(r))
    }
}
```

**Step 3: Run tests**

Run: `go test -v ./internal/models/...`

Expected: All tests pass.

**Step 4: Commit**

```bash
git add internal/models/models.go \
        internal/models/models_test.go
git commit -m "feat: add conversation, message, and token usage model types"
```

---

## Task 3: Database Layer — Conversation CRUD

Add database functions for creating/upserting, getting, and
listing conversations. Uses the `get_or_create_conversation()`
stored function and the `conversation_list` view.

**Files:**

- Create: `internal/database/conversations.go`
- Create: `internal/database/conversations_test.go`

**Step 1: Write failing tests**

Create `internal/database/conversations_test.go` with table-
driven tests. These are unit tests using a mock/stub or
integration tests using the real DB. Follow the existing
pattern in `internal/database/db_test.go`.

Since the existing test patterns use testify and the project
has integration test infrastructure, write tests that can run
against a real database:

```go
// +build integration

package database_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "imagineer/internal/database"
    "imagineer/internal/models"
)

func TestGetOrCreateConversation(t *testing.T) {
    db := setupTestDB(t)
    ctx := context.Background()
    campaignID := createTestCampaign(t, db)

    // First call: creates the conversation
    conv, wasCreated, err :=
        db.GetOrCreateConversation(ctx,
            campaignID,
            models.ScopeTypeChapter,
            int64(1),
        )

    require.NoError(t, err)
    assert.True(t, wasCreated)
    assert.Equal(t, campaignID, conv.CampaignID)
    assert.Equal(t, models.ScopeTypeChapter,
        conv.ScopeType)
    assert.Equal(t, int64(1), conv.ScopeID)
    assert.NotZero(t, conv.ID)
    assert.NotZero(t, conv.CreatedAt)

    // Second call: returns existing
    conv2, wasCreated2, err :=
        db.GetOrCreateConversation(ctx,
            campaignID,
            models.ScopeTypeChapter,
            int64(1),
        )
    require.NoError(t, err)
    assert.False(t, wasCreated2)
    assert.Equal(t, conv.ID, conv2.ID)
}

func TestGetConversation(t *testing.T) {
    db := setupTestDB(t)
    ctx := context.Background()
    campaignID := createTestCampaign(t, db)

    created, _, err :=
        db.GetOrCreateConversation(ctx,
            campaignID,
            models.ScopeTypeEntity,
            int64(42),
        )
    require.NoError(t, err)

    got, err := db.GetConversation(ctx, created.ID)
    require.NoError(t, err)
    assert.Equal(t, created.ID, got.ID)
    assert.Equal(t, models.ScopeTypeEntity,
        got.ScopeType)
}

func TestListConversations(t *testing.T) {
    db := setupTestDB(t)
    ctx := context.Background()
    campaignID := createTestCampaign(t, db)

    // Create two conversations
    _, _, err := db.GetOrCreateConversation(ctx,
        campaignID,
        models.ScopeTypeChapter,
        int64(1),
    )
    require.NoError(t, err)
    _, _, err = db.GetOrCreateConversation(ctx,
        campaignID,
        models.ScopeTypeEntity,
        int64(2),
    )
    require.NoError(t, err)

    convs, err := db.ListConversations(ctx,
        campaignID, "", 0, 20)
    require.NoError(t, err)
    assert.GreaterOrEqual(t, len(convs), 2)
    // Verify view-derived fields are populated
    for _, c := range convs {
        assert.NotZero(t, c.CampaignID)
    }
}

func TestListConversationsFiltered(t *testing.T) {
    db := setupTestDB(t)
    ctx := context.Background()
    campaignID := createTestCampaign(t, db)

    _, _, err := db.GetOrCreateConversation(ctx,
        campaignID,
        models.ScopeTypeChapter,
        int64(1),
    )
    require.NoError(t, err)
    _, _, err = db.GetOrCreateConversation(ctx,
        campaignID,
        models.ScopeTypeEntity,
        int64(2),
    )
    require.NoError(t, err)

    convs, err := db.ListConversations(ctx,
        campaignID, "chapter", 0, 20)
    require.NoError(t, err)
    for _, c := range convs {
        assert.Equal(t, models.ScopeTypeChapter,
            c.ScopeType)
    }
}
```

**Step 2: Run tests to verify they fail**

Run: `go test -v -tags integration ./internal/database/ -run TestGetOrCreateConversation`

Expected: FAIL — `GetOrCreateConversation` not defined.

**Step 3: Implement conversation CRUD**

Create `internal/database/conversations.go`:

```go
package database

import (
    "context"
    "fmt"

    "imagineer/internal/models"
)

// GetOrCreateConversation calls the
// get_or_create_conversation() stored function.
// Returns the conversation and whether it was
// newly created.
func (db *DB) GetOrCreateConversation(
    ctx context.Context,
    campaignID int64,
    scopeType models.ScopeType,
    scopeID int64,
) (*models.Conversation, bool, error) {
    var conv models.Conversation
    var wasCreated bool
    err := db.QueryRow(ctx,
        `SELECT id, campaign_id, scope_type,
            scope_id, title, summary,
            summary_tokens, created_at,
            updated_at, was_created
         FROM get_or_create_conversation(
             $1, $2, $3)`,
        campaignID,
        string(scopeType),
        scopeID,
    ).Scan(
        &conv.ID, &conv.CampaignID,
        &conv.ScopeType, &conv.ScopeID,
        &conv.Title, &conv.Summary,
        &conv.SummaryTokens,
        &conv.CreatedAt, &conv.UpdatedAt,
        &wasCreated,
    )
    if err != nil {
        return nil, false, fmt.Errorf(
            "get_or_create_conversation: %w", err)
    }
    return &conv, wasCreated, nil
}

// GetConversation retrieves a conversation by ID.
func (db *DB) GetConversation(
    ctx context.Context,
    id int64,
) (*models.Conversation, error) {
    var conv models.Conversation
    err := db.QueryRow(ctx,
        `SELECT id, campaign_id, scope_type,
            scope_id, title, summary,
            summary_tokens,
            created_at, updated_at
         FROM conversations
         WHERE id = $1`,
        id,
    ).Scan(
        &conv.ID, &conv.CampaignID,
        &conv.ScopeType, &conv.ScopeID,
        &conv.Title, &conv.Summary,
        &conv.SummaryTokens,
        &conv.CreatedAt, &conv.UpdatedAt,
    )
    if err != nil {
        return nil, fmt.Errorf(
            "failed to get conversation %d: %w",
            id, err)
    }
    return &conv, nil
}

// ListConversations queries the conversation_list
// view, returning conversations with message
// preview and counts.
func (db *DB) ListConversations(
    ctx context.Context,
    campaignID int64,
    scopeType string,
    scopeID int64,
    limit int,
) ([]models.ConversationListItem, error) {
    if limit <= 0 {
        limit = 20
    }

    query := `SELECT id, campaign_id,
            scope_type, scope_id, title,
            summary_tokens, created_at,
            updated_at, last_message_role,
            last_message_preview,
            last_message_at, total_messages,
            uncompacted_messages
         FROM conversation_list
         WHERE campaign_id = $1`
    args := []any{campaignID}
    argN := 2

    if scopeType != "" {
        query += fmt.Sprintf(
            " AND scope_type = $%d", argN)
        args = append(args, scopeType)
        argN++
    }
    if scopeID > 0 {
        query += fmt.Sprintf(
            " AND scope_id = $%d", argN)
        args = append(args, scopeID)
        argN++
    }

    query += fmt.Sprintf(
        " ORDER BY updated_at DESC LIMIT $%d",
        argN)
    args = append(args, limit)

    rows, err := db.Query(ctx, query, args...)
    if err != nil {
        return nil, fmt.Errorf(
            "failed to list conversations: %w",
            err)
    }
    defer rows.Close()

    var items []models.ConversationListItem
    for rows.Next() {
        var c models.ConversationListItem
        err := rows.Scan(
            &c.ID, &c.CampaignID,
            &c.ScopeType, &c.ScopeID,
            &c.Title, &c.SummaryTokens,
            &c.CreatedAt, &c.UpdatedAt,
            &c.LastMessageRole,
            &c.LastMessagePreview,
            &c.LastMessageAt,
            &c.TotalMessages,
            &c.UncompactedMessages,
        )
        if err != nil {
            return nil, fmt.Errorf(
                "failed to scan conversation: %w",
                err)
        }
        items = append(items, c)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf(
            "conversation rows error: %w", err)
    }
    return items, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test -v -tags integration ./internal/database/ -run TestGetOrCreateConversation`

Run: `go test -v -tags integration ./internal/database/ -run TestGetConversation`

Run: `go test -v -tags integration ./internal/database/ -run TestListConversation`

Expected: All pass.

**Step 5: Commit**

```bash
git add internal/database/conversations.go \
        internal/database/conversations_test.go
git commit -m "feat: add conversation CRUD using stored functions and views"
```

---

## Task 4: Database Layer — Messages

Add database functions for creating and listing messages,
assembling conversation context via the stored function,
and compacting messages via the stored function.

**Files:**

- Modify: `internal/database/conversations.go`
- Modify: `internal/database/conversations_test.go`

**Step 1: Write failing tests for message operations**

Add to `internal/database/conversations_test.go`:

```go
func TestCreateMessage(t *testing.T) {
    db := setupTestDB(t)
    ctx := context.Background()
    campaignID := createTestCampaign(t, db)

    conv, _, err :=
        db.GetOrCreateConversation(ctx,
            campaignID,
            models.ScopeTypeChapter,
            int64(1),
        )
    require.NoError(t, err)

    msg, err := db.CreateMessage(ctx, conv.ID,
        models.RoleUser,
        "Hello, help me with Session 5",
        nil, nil, nil, nil, nil)
    require.NoError(t, err)
    assert.Equal(t, conv.ID, msg.ConversationID)
    assert.Equal(t, models.RoleUser, msg.Role)
    assert.Equal(t,
        "Hello, help me with Session 5",
        msg.Content)
    assert.False(t, msg.Compacted)

    // Verify trigger bumped updated_at
    got, err := db.GetConversation(ctx, conv.ID)
    require.NoError(t, err)
    assert.True(t,
        got.UpdatedAt.After(conv.UpdatedAt) ||
            got.UpdatedAt.Equal(conv.UpdatedAt))
}

func TestListMessages(t *testing.T) {
    db := setupTestDB(t)
    ctx := context.Background()
    campaignID := createTestCampaign(t, db)

    conv, _, err :=
        db.GetOrCreateConversation(ctx,
            campaignID,
            models.ScopeTypeChapter,
            int64(1),
        )
    require.NoError(t, err)

    _, err = db.CreateMessage(ctx, conv.ID,
        models.RoleUser, "First message",
        nil, nil, nil, nil, nil)
    require.NoError(t, err)
    _, err = db.CreateMessage(ctx, conv.ID,
        models.RoleAssistant, "Response",
        nil, nil, nil, nil, nil)
    require.NoError(t, err)

    msgs, err := db.ListMessages(ctx,
        conv.ID, 50, 0)
    require.NoError(t, err)
    assert.Len(t, msgs, 2)
    assert.Equal(t, "First message",
        msgs[0].Content)
}

func TestAssembleConversationContext(t *testing.T) {
    db := setupTestDB(t)
    ctx := context.Background()
    campaignID := createTestCampaign(t, db)

    conv, _, err :=
        db.GetOrCreateConversation(ctx,
            campaignID,
            models.ScopeTypeChapter,
            int64(1),
        )
    require.NoError(t, err)

    tokens := 100
    _, err = db.CreateMessage(ctx, conv.ID,
        models.RoleUser, "Hello",
        nil, nil, nil, nil, &tokens)
    require.NoError(t, err)

    cc, err := db.AssembleConversationContext(
        ctx, conv.ID)
    require.NoError(t, err)
    assert.Equal(t, conv.ID,
        cc.Conversation.ID)
    assert.Len(t, cc.Messages, 1)
    assert.Equal(t, 100, cc.TokenEstimate)
}

func TestCompactMessages(t *testing.T) {
    db := setupTestDB(t)
    ctx := context.Background()
    campaignID := createTestCampaign(t, db)

    conv, _, err :=
        db.GetOrCreateConversation(ctx,
            campaignID,
            models.ScopeTypeChapter,
            int64(1),
        )
    require.NoError(t, err)

    msg1, err := db.CreateMessage(ctx, conv.ID,
        models.RoleUser, "Old message",
        nil, nil, nil, nil, nil)
    require.NoError(t, err)
    _, err = db.CreateMessage(ctx, conv.ID,
        models.RoleAssistant, "Recent message",
        nil, nil, nil, nil, nil)
    require.NoError(t, err)

    // Compact the first message
    count, err := db.CompactMessages(ctx,
        conv.ID, []int64{msg1.ID},
        "Summary of old exchange", 50)
    require.NoError(t, err)
    assert.Equal(t, 1, count)

    // Verify summary was updated
    got, err := db.GetConversation(ctx, conv.ID)
    require.NoError(t, err)
    assert.NotNil(t, got.Summary)
    assert.Equal(t, "Summary of old exchange",
        *got.Summary)
    assert.Equal(t, 50, got.SummaryTokens)

    // Verify context excludes compacted
    cc, err := db.AssembleConversationContext(
        ctx, conv.ID)
    require.NoError(t, err)
    assert.Len(t, cc.Messages, 1)
    assert.Equal(t, "Recent message",
        cc.Messages[0].Content)
}
```

**Step 2: Run tests to verify they fail**

Run: `go test -v -tags integration ./internal/database/ -run TestCreateMessage`

Expected: FAIL — `CreateMessage` not defined.

**Step 3: Implement message functions**

Add to `internal/database/conversations.go`:

```go
// CreateMessage inserts a message into a conversation.
// The touch_conversation_on_message trigger
// automatically bumps conversations.updated_at on
// insert -- no manual UPDATE needed.
func (db *DB) CreateMessage(
    ctx context.Context,
    conversationID int64,
    role models.MessageRole,
    content string,
    toolName *string,
    toolUseID *string,
    toolInput json.RawMessage,
    toolResult json.RawMessage,
    tokens *int,
) (*models.Message, error) {
    var msg models.Message
    err := db.QueryRow(ctx,
        `INSERT INTO messages
            (conversation_id, role, content,
             tool_name, tool_use_id,
             tool_input, tool_result, tokens)
         VALUES ($1, $2, $3, $4, $5, $6,
                 $7, $8)
         RETURNING id, conversation_id, role,
            content, tool_name, tool_use_id,
            tool_input, tool_result, tokens,
            compacted, created_at`,
        conversationID, string(role), content,
        toolName, toolUseID,
        toolInput, toolResult, tokens,
    ).Scan(
        &msg.ID, &msg.ConversationID,
        &msg.Role, &msg.Content,
        &msg.ToolName, &msg.ToolUseID,
        &msg.ToolInput, &msg.ToolResult,
        &msg.Tokens,
        &msg.Compacted, &msg.CreatedAt,
    )
    if err != nil {
        return nil, fmt.Errorf(
            "failed to create message: %w", err)
    }
    return &msg, nil
}

// ListMessages returns messages for a conversation
// with pagination. Ordered by created_at ascending.
func (db *DB) ListMessages(
    ctx context.Context,
    conversationID int64,
    limit int,
    beforeID int64,
) ([]models.Message, error) {
    if limit <= 0 {
        limit = 50
    }

    query := `SELECT id, conversation_id, role,
            content, tool_name, tool_use_id,
            tool_input, tool_result, tokens,
            compacted, created_at
         FROM messages
         WHERE conversation_id = $1`
    args := []any{conversationID}

    if beforeID > 0 {
        query += " AND id < $3"
        args = append(args, beforeID)
    }

    query += fmt.Sprintf(
        " ORDER BY created_at ASC LIMIT $%d",
        len(args)+1)
    args = append(args, limit)

    return db.scanMessages(ctx, query, args...)
}

// AssembleConversationContext calls the
// assemble_conversation_context() stored function.
// Returns conversation metadata, uncompacted
// messages, and token budget estimate in one
// round-trip. This is the hot path -- called
// every conversation turn on cache miss.
func (db *DB) AssembleConversationContext(
    ctx context.Context,
    conversationID int64,
) (*models.ConversationContext, error) {
    rows, err := db.Query(ctx,
        `SELECT conversation_id, campaign_id,
            scope_type, scope_id, summary,
            summary_tokens,
            uncompacted_token_estimate,
            message_id, role, content,
            tool_name, tool_use_id,
            tool_input, tool_result,
            message_tokens, message_created_at
         FROM assemble_conversation_context($1)`,
        conversationID,
    )
    if err != nil {
        return nil, fmt.Errorf(
            "assemble_conversation_context: %w",
            err)
    }
    defer rows.Close()

    var cc models.ConversationContext
    first := true
    for rows.Next() {
        var m models.Message
        var convID, campaignID, scopeID int64
        var scopeType string
        var summary *string
        var summaryTokens, tokenEstimate int

        err := rows.Scan(
            &convID, &campaignID,
            &scopeType, &scopeID,
            &summary, &summaryTokens,
            &tokenEstimate,
            &m.ID, &m.Role, &m.Content,
            &m.ToolName, &m.ToolUseID,
            &m.ToolInput, &m.ToolResult,
            &m.Tokens, &m.CreatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf(
                "scan context row: %w", err)
        }

        if first {
            cc.Conversation = models.Conversation{
                ID:            convID,
                CampaignID:    campaignID,
                ScopeType: models.ScopeType(
                    scopeType),
                ScopeID:       scopeID,
                Summary:       summary,
                SummaryTokens: summaryTokens,
            }
            cc.TokenEstimate = tokenEstimate
            first = false
        }

        m.ConversationID = convID
        cc.Messages = append(cc.Messages, m)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf(
            "context rows error: %w", err)
    }
    if first {
        return nil, fmt.Errorf(
            "conversation %d not found",
            conversationID)
    }
    return &cc, nil
}

// CompactMessages calls the compact_messages()
// stored function. Atomically marks messages as
// compacted and updates the conversation summary.
func (db *DB) CompactMessages(
    ctx context.Context,
    conversationID int64,
    messageIDs []int64,
    newSummary string,
    summaryTokens int,
) (int, error) {
    var compactedCount int
    err := db.QueryRow(ctx,
        `SELECT compact_messages($1, $2, $3, $4)`,
        conversationID,
        messageIDs,
        newSummary,
        summaryTokens,
    ).Scan(&compactedCount)
    if err != nil {
        return 0, fmt.Errorf(
            "compact_messages: %w", err)
    }
    return compactedCount, nil
}

// scanMessages is a shared helper for scanning
// message rows. Includes tool_use_id in the scan.
func (db *DB) scanMessages(
    ctx context.Context,
    query string,
    args ...any,
) ([]models.Message, error) {
    rows, err := db.Query(ctx, query, args...)
    if err != nil {
        return nil, fmt.Errorf(
            "failed to query messages: %w", err)
    }
    defer rows.Close()

    var msgs []models.Message
    for rows.Next() {
        var m models.Message
        err := rows.Scan(
            &m.ID, &m.ConversationID,
            &m.Role, &m.Content,
            &m.ToolName, &m.ToolUseID,
            &m.ToolInput, &m.ToolResult,
            &m.Tokens,
            &m.Compacted, &m.CreatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf(
                "failed to scan message: %w", err)
        }
        msgs = append(msgs, m)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf(
            "message rows error: %w", err)
    }
    return msgs, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test -v -tags integration ./internal/database/ -run TestCreateMessage`

Run: `go test -v -tags integration ./internal/database/ -run TestListMessages`

Run: `go test -v -tags integration ./internal/database/ -run TestAssembleConversationContext`

Run: `go test -v -tags integration ./internal/database/ -run TestCompactMessages`

Expected: All pass.

**Step 5: Commit**

```bash
git add internal/database/conversations.go \
        internal/database/conversations_test.go
git commit -m "feat: add message CRUD, context assembly, and compaction via stored functions"
```

---

## Task 5: Database Layer — Token Usage Logging

Add the function for logging token usage.

**Files:**

- Modify: `internal/database/conversations.go`
- Modify: `internal/database/conversations_test.go`

**Step 1: Write failing test**

Add to `internal/database/conversations_test.go`:

```go
func TestLogTokenUsage(t *testing.T) {
    db := setupTestDB(t)
    ctx := context.Background()
    campaignID := createTestCampaign(t, db)
    userID := createTestUser(t, db)

    conv, _, err :=
        db.GetOrCreateConversation(ctx,
            campaignID,
            models.ScopeTypeChapter,
            int64(1),
        )
    require.NoError(t, err)

    err = db.LogTokenUsage(ctx,
        models.TokenUsageLog{
            ConversationID: conv.ID,
            CampaignID:     campaignID,
            UserID:         userID,
            Model:          "claude-sonnet-4-20250514",
            InputTokens:    1200,
            OutputTokens:   380,
            TotalTokens:    1580,
            LLMCallType:    models.CallTypeConversation,
        },
    )
    require.NoError(t, err)
}

func TestGetTokenUsageSummary(t *testing.T) {
    db := setupTestDB(t)
    ctx := context.Background()
    campaignID := createTestCampaign(t, db)
    userID := createTestUser(t, db)

    conv, _, err :=
        db.GetOrCreateConversation(ctx,
            campaignID,
            models.ScopeTypeChapter,
            int64(1),
        )
    require.NoError(t, err)

    // Log two entries
    for i := 0; i < 2; i++ {
        err = db.LogTokenUsage(ctx,
            models.TokenUsageLog{
                ConversationID: conv.ID,
                CampaignID:     campaignID,
                UserID:         userID,
                Model:          "claude-sonnet-4-20250514",
                InputTokens:    1000,
                OutputTokens:   500,
                TotalTokens:    1500,
                LLMCallType:    models.CallTypeConversation,
            },
        )
        require.NoError(t, err)
    }

    since := time.Now().Add(-1 * time.Hour)
    summaries, err := db.GetTokenUsageSummary(
        ctx, &campaignID, &userID, since, nil)
    require.NoError(t, err)
    assert.NotEmpty(t, summaries)
    assert.Equal(t, int64(2),
        summaries[0].CallCount)
    assert.Equal(t, int64(3000),
        summaries[0].TotalTokens)
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v -tags integration ./internal/database/ -run TestLogTokenUsage`

Expected: FAIL — `LogTokenUsage` not defined.

**Step 3: Implement**

Add to `internal/database/conversations.go`:

```go
// LogTokenUsage records token consumption for a
// single LLM call.
func (db *DB) LogTokenUsage(
    ctx context.Context,
    usage models.TokenUsageLog,
) error {
    _, err := db.Exec(ctx,
        `INSERT INTO token_usage_log
            (conversation_id, campaign_id,
             user_id, model, input_tokens,
             output_tokens, total_tokens,
             llm_call_type, agent_name)
         VALUES ($1, $2, $3, $4, $5, $6,
                 $7, $8, $9)`,
        usage.ConversationID,
        usage.CampaignID,
        usage.UserID,
        usage.Model,
        usage.InputTokens,
        usage.OutputTokens,
        usage.TotalTokens,
        string(usage.LLMCallType),
        usage.AgentName,
    )
    if err != nil {
        return fmt.Errorf(
            "failed to log token usage: %w", err)
    }
    return nil
}

// GetTokenUsageSummary calls
// get_token_usage_summary(). Pass nil for
// campaignID or userID for broader aggregation.
func (db *DB) GetTokenUsageSummary(
    ctx context.Context,
    campaignID *int64,
    userID *int64,
    since time.Time,
    until *time.Time,
) ([]models.TokenUsageSummary, error) {
    rows, err := db.Query(ctx,
        `SELECT llm_call_type, model,
            call_count, total_input_tokens,
            total_output_tokens, total_tokens
         FROM get_token_usage_summary(
             $1, $2, $3, $4)`,
        campaignID, userID, since, until,
    )
    if err != nil {
        return nil, fmt.Errorf(
            "get_token_usage_summary: %w", err)
    }
    defer rows.Close()

    var summaries []models.TokenUsageSummary
    for rows.Next() {
        var s models.TokenUsageSummary
        err := rows.Scan(
            &s.LLMCallType, &s.Model,
            &s.CallCount,
            &s.TotalInputTokens,
            &s.TotalOutputTokens,
            &s.TotalTokens,
        )
        if err != nil {
            return nil, fmt.Errorf(
                "scan token usage summary: %w",
                err)
        }
        summaries = append(summaries, s)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf(
            "token usage rows error: %w", err)
    }
    return summaries, nil
}
```

**Step 4: Run tests**

Run: `go test -v -tags integration ./internal/database/ -run TestLogTokenUsage`

Run: `go test -v -tags integration ./internal/database/ -run TestGetTokenUsageSummary`

Expected: All pass.

**Step 5: Commit**

```bash
git add internal/database/conversations.go \
        internal/database/conversations_test.go
git commit -m "feat: add token usage logging and summary via stored function"
```

---

## Task 6: Streaming Provider Interface

Define the `StreamingProvider` interface and supporting types.

**Files:**

- Create: `internal/llm/streaming.go`
- Create: `internal/llm/streaming_test.go`

**Step 1: Write the interface and types**

Create `internal/llm/streaming.go`:

```go
package llm

import (
    "context"
    "encoding/json"
)

// StreamEventType identifies the kind of streaming event.
type StreamEventType int

const (
    EventTextDelta StreamEventType = iota
    EventToolUse
    EventUsage
    EventDone
    EventError
)

// StreamEvent represents a single event from the
// streaming LLM response.
type StreamEvent struct {
    Type      StreamEventType
    Text      string
    ToolName  string
    ToolInput json.RawMessage
    ToolID    string
    Usage     *TokenUsage
    Error     error
}

// TokenUsage records token counts from an LLM call.
type TokenUsage struct {
    InputTokens  int
    OutputTokens int
}

// ToolDefinition describes a tool the LLM can call.
type ToolDefinition struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    InputSchema json.RawMessage `json:"input_schema"`
}

// StreamingMessage represents a message in a multi-turn
// conversation for the streaming API.
type StreamingMessage struct {
    Role       string          `json:"role"`
    Content    string          `json:"content,omitempty"`
    ToolUseID  string          `json:"tool_use_id,omitempty"`
    ToolName   string          `json:"tool_name,omitempty"`
    ToolInput  json.RawMessage `json:"tool_input,omitempty"`
    ToolResult json.RawMessage `json:"tool_result,omitempty"`
}

// StreamingRequest contains all parameters for a
// streaming LLM call with tool use.
type StreamingRequest struct {
    SystemPrompt string
    Messages     []StreamingMessage
    Tools        []ToolDefinition
    MaxTokens    int
    Temperature  float64
}

// StreamingProvider streams LLM responses with tool use
// support. Implementors return a channel that emits
// StreamEvent values until the response completes.
type StreamingProvider interface {
    CompleteStream(
        ctx context.Context,
        req StreamingRequest,
    ) (<-chan StreamEvent, error)
}
```

**Step 2: Write test for event type constants**

Create `internal/llm/streaming_test.go`:

```go
package llm

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestStreamEventTypes(t *testing.T) {
    // Verify distinct values
    types := []StreamEventType{
        EventTextDelta,
        EventToolUse,
        EventUsage,
        EventDone,
        EventError,
    }
    seen := make(map[StreamEventType]bool)
    for _, et := range types {
        assert.False(t, seen[et],
            "duplicate event type: %d", et)
        seen[et] = true
    }
}

func TestStreamingRequestDefaults(t *testing.T) {
    req := StreamingRequest{
        SystemPrompt: "You are a helpful assistant.",
        MaxTokens:    4096,
        Temperature:  0.7,
    }
    assert.Equal(t, 4096, req.MaxTokens)
    assert.Equal(t, 0.7, req.Temperature)
    assert.Empty(t, req.Messages)
    assert.Empty(t, req.Tools)
}
```

**Step 3: Run tests**

Run: `go test -v ./internal/llm/ -run TestStreamEvent`

Run: `go test -v ./internal/llm/ -run TestStreamingRequest`

Expected: All pass.

**Step 4: Commit**

```bash
git add internal/llm/streaming.go \
        internal/llm/streaming_test.go
git commit -m "feat: add StreamingProvider interface and types"
```

---

## Task 7: Anthropic Streaming Implementation

Implement `StreamingProvider` for the Anthropic Messages API
with SSE streaming and tool use support.

**Files:**

- Create: `internal/llm/anthropic_stream.go`
- Create: `internal/llm/anthropic_stream_test.go`

**Step 1: Write a test using a mock HTTP server**

Create `internal/llm/anthropic_stream_test.go`:

```go
package llm

import (
    "context"
    "fmt"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestAnthropicStreamTextResponse(t *testing.T) {
    // Mock Anthropic SSE response
    server := httptest.NewServer(
        http.HandlerFunc(
            func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Content-Type",
                    "text/event-stream")
                flusher := w.(http.Flusher)

                // content_block_start
                fmt.Fprintf(w,
                    "event: content_block_start\n"+
                    "data: {\"index\":0,"+
                    "\"content_block\":{\"type\":"+
                    "\"text\",\"text\":\"\"}}\n\n")
                flusher.Flush()

                // content_block_delta with text
                fmt.Fprintf(w,
                    "event: content_block_delta\n"+
                    "data: {\"index\":0,\"delta\":"+
                    "{\"type\":\"text_delta\","+
                    "\"text\":\"Hello\"}}\n\n")
                flusher.Flush()

                // message_delta with usage
                fmt.Fprintf(w,
                    "event: message_delta\n"+
                    "data: {\"delta\":{\"stop_reason"+
                    "\":\"end_turn\"},\"usage\":"+
                    "{\"output_tokens\":5}}\n\n")
                flusher.Flush()

                // message_stop
                fmt.Fprintf(w,
                    "event: message_stop\n"+
                    "data: {}\n\n")
                flusher.Flush()
            },
        ),
    )
    defer server.Close()

    provider := &AnthropicStreamProvider{
        apiKey:  "test-key",
        baseURL: server.URL,
        client:  server.Client(),
    }

    ch, err := provider.CompleteStream(
        context.Background(),
        StreamingRequest{
            SystemPrompt: "Test",
            Messages: []StreamingMessage{
                {Role: "user", Content: "Hi"},
            },
            MaxTokens:  100,
            Temperature: 0.7,
        },
    )
    require.NoError(t, err)

    var events []StreamEvent
    for ev := range ch {
        events = append(events, ev)
    }

    // Should have at least a text delta and done
    hasText := false
    hasDone := false
    for _, ev := range events {
        if ev.Type == EventTextDelta {
            hasText = true
            assert.Equal(t, "Hello", ev.Text)
        }
        if ev.Type == EventDone {
            hasDone = true
        }
    }
    assert.True(t, hasText, "expected text delta event")
    assert.True(t, hasDone, "expected done event")
}

func TestAnthropicStreamToolUseResponse(t *testing.T) {
    server := httptest.NewServer(
        http.HandlerFunc(
            func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Content-Type",
                    "text/event-stream")
                flusher := w.(http.Flusher)

                // content_block_start with tool_use
                fmt.Fprintf(w,
                    "event: content_block_start\n"+
                    "data: {\"index\":0,"+
                    "\"content_block\":{\"type\":"+
                    "\"tool_use\",\"id\":"+
                    "\"toolu_123\",\"name\":"+
                    "\"search_entities\","+
                    "\"input\":{}}}\n\n")
                flusher.Flush()

                // content_block_delta with input JSON
                fmt.Fprintf(w,
                    "event: content_block_delta\n"+
                    "data: {\"index\":0,\"delta\":"+
                    "{\"type\":\"input_json_delta\","+
                    "\"partial_json\":"+
                    "\"{\\\"query\\\":\\\"inn\\\"}\""+
                    "}}\n\n")
                flusher.Flush()

                // content_block_stop
                fmt.Fprintf(w,
                    "event: content_block_stop\n"+
                    "data: {\"index\":0}\n\n")
                flusher.Flush()

                // message_delta
                fmt.Fprintf(w,
                    "event: message_delta\n"+
                    "data: {\"delta\":{\"stop_reason"+
                    "\":\"tool_use\"},\"usage\":"+
                    "{\"output_tokens\":20}}\n\n")
                flusher.Flush()

                // message_stop
                fmt.Fprintf(w,
                    "event: message_stop\n"+
                    "data: {}\n\n")
                flusher.Flush()
            },
        ),
    )
    defer server.Close()

    provider := &AnthropicStreamProvider{
        apiKey:  "test-key",
        baseURL: server.URL,
        client:  server.Client(),
    }

    ch, err := provider.CompleteStream(
        context.Background(),
        StreamingRequest{
            SystemPrompt: "Test",
            Messages: []StreamingMessage{
                {Role: "user", Content: "Find inns"},
            },
            Tools: []ToolDefinition{
                {
                    Name:        "search_entities",
                    Description: "Search entities",
                },
            },
            MaxTokens:  100,
            Temperature: 0.7,
        },
    )
    require.NoError(t, err)

    var events []StreamEvent
    for ev := range ch {
        events = append(events, ev)
    }

    hasToolUse := false
    for _, ev := range events {
        if ev.Type == EventToolUse {
            hasToolUse = true
            assert.Equal(t, "search_entities",
                ev.ToolName)
            assert.Equal(t, "toolu_123", ev.ToolID)
        }
    }
    assert.True(t, hasToolUse,
        "expected tool use event")
}
```

**Step 2: Run tests to verify they fail**

Run: `go test -v ./internal/llm/ -run TestAnthropicStream`

Expected: FAIL — `AnthropicStreamProvider` not defined.

**Step 3: Implement the Anthropic streaming provider**

Create `internal/llm/anthropic_stream.go`. This is the
largest single implementation file. It wraps the Anthropic
Messages API streaming endpoint, parsing SSE events and
emitting `StreamEvent` values on a channel.

Key implementation details:

- Use `net/http` to POST to `{baseURL}/v1/messages` with
  `"stream": true` in the request body.
- Parse the SSE response line by line.
- Map Anthropic events (`content_block_start`,
  `content_block_delta`, `content_block_stop`,
  `message_delta`, `message_stop`) to `StreamEvent` types.
- Accumulate partial JSON for tool_use input deltas.
- Emit `EventToolUse` with complete input when a tool_use
  content block stops.
- Emit `EventDone` on `message_stop`.
- Emit `EventError` on HTTP errors or parse failures.
- Close the channel when done.
- Respect context cancellation.

The struct:

```go
type AnthropicStreamProvider struct {
    apiKey  string
    baseURL string
    client  *http.Client
}

func NewAnthropicStreamProvider(
    apiKey string,
) (*AnthropicStreamProvider, error) {
    if apiKey == "" {
        return nil, fmt.Errorf(
            "anthropic API key is required")
    }
    return &AnthropicStreamProvider{
        apiKey:  apiKey,
        baseURL: anthropicAPIURL,
        client:  &http.Client{Timeout: 300 * time.Second},
    }, nil
}
```

The `CompleteStream` method builds the request body with
messages, tools, system prompt, and `"stream": true`, then
launches a goroutine that reads the SSE response and sends
events on the channel.

Refer to the Anthropic streaming API documentation for the
exact SSE event format. The key events are:

- `message_start` — contains `message.usage.input_tokens`
- `content_block_start` — `content_block.type` is `"text"`
  or `"tool_use"`; for tool_use also has `id` and `name`
- `content_block_delta` — `delta.type` is `"text_delta"`
  (has `text`) or `"input_json_delta"` (has `partial_json`)
- `content_block_stop` — signals end of a content block
- `message_delta` — contains `usage.output_tokens` and
  `delta.stop_reason`
- `message_stop` — signals end of the message

**Step 4: Run tests**

Run: `go test -v ./internal/llm/ -run TestAnthropicStream`

Expected: All pass.

**Step 5: Commit**

```bash
git add internal/llm/anthropic_stream.go \
        internal/llm/anthropic_stream_test.go
git commit -m "feat: add Anthropic streaming provider with tool use support"
```

---

## Task 8: Tool Registry and Definitions

Create the tool registry that holds tool definitions and
dispatches execution.

**Files:**

- Create: `internal/conversation/tools.go`
- Create: `internal/conversation/tools_test.go`

**Step 1: Write failing tests for the registry**

```go
package conversation

import (
    "context"
    "encoding/json"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "imagineer/internal/llm"
)

func TestToolRegistryRegisterAndGet(t *testing.T) {
    registry := NewToolRegistry()

    registry.Register(Tool{
        Definition: llm.ToolDefinition{
            Name:        "search_entities",
            Description: "Search campaign entities",
        },
        Execute: func(ctx context.Context,
            input json.RawMessage,
        ) (json.RawMessage, error) {
            return json.Marshal(map[string]string{
                "result": "found",
            })
        },
    })

    tool, ok := registry.Get("search_entities")
    assert.True(t, ok)
    assert.Equal(t, "search_entities",
        tool.Definition.Name)
}

func TestToolRegistryDefinitions(t *testing.T) {
    registry := NewToolRegistry()
    registry.Register(Tool{
        Definition: llm.ToolDefinition{
            Name:        "tool_a",
            Description: "Tool A",
        },
    })
    registry.Register(Tool{
        Definition: llm.ToolDefinition{
            Name:        "tool_b",
            Description: "Tool B",
        },
    })

    defs := registry.Definitions()
    assert.Len(t, defs, 2)
}

func TestToolRegistryExecute(t *testing.T) {
    registry := NewToolRegistry()
    registry.Register(Tool{
        Definition: llm.ToolDefinition{
            Name: "echo",
        },
        Execute: func(ctx context.Context,
            input json.RawMessage,
        ) (json.RawMessage, error) {
            return input, nil
        },
    })

    input := json.RawMessage(`{"msg":"hello"}`)
    result, err := registry.Execute(
        context.Background(), "echo", input)
    require.NoError(t, err)
    assert.JSONEq(t, `{"msg":"hello"}`,
        string(result))
}

func TestToolRegistryExecuteNotFound(t *testing.T) {
    registry := NewToolRegistry()
    _, err := registry.Execute(
        context.Background(), "nonexistent", nil)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "unknown tool")
}
```

**Step 2: Run tests to verify they fail**

Run: `go test -v ./internal/conversation/ -run TestToolRegistry`

Expected: FAIL — package does not exist.

**Step 3: Implement the tool registry**

Create `internal/conversation/tools.go`:

```go
package conversation

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"

    "imagineer/internal/llm"
)

// ToolExecuteFunc is the function signature for tool
// execution. It receives the raw JSON input from the
// LLM and returns the raw JSON result.
type ToolExecuteFunc func(
    ctx context.Context,
    input json.RawMessage,
) (json.RawMessage, error)

// Tool pairs a tool definition (sent to the LLM) with
// its execution function (called when the LLM invokes
// the tool).
type Tool struct {
    Definition llm.ToolDefinition
    Execute    ToolExecuteFunc
}

// ToolRegistry holds registered tools and dispatches
// execution by name.
type ToolRegistry struct {
    mu    sync.RWMutex
    tools map[string]Tool
    order []string // preserve registration order
}

// NewToolRegistry creates an empty tool registry.
func NewToolRegistry() *ToolRegistry {
    return &ToolRegistry{
        tools: make(map[string]Tool),
    }
}

// Register adds a tool to the registry.
func (r *ToolRegistry) Register(tool Tool) {
    r.mu.Lock()
    defer r.mu.Unlock()
    name := tool.Definition.Name
    if _, exists := r.tools[name]; !exists {
        r.order = append(r.order, name)
    }
    r.tools[name] = tool
}

// Get retrieves a tool by name.
func (r *ToolRegistry) Get(name string) (Tool, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    t, ok := r.tools[name]
    return t, ok
}

// Definitions returns all tool definitions in
// registration order, suitable for passing to the LLM.
func (r *ToolRegistry) Definitions() []llm.ToolDefinition {
    r.mu.RLock()
    defer r.mu.RUnlock()
    defs := make([]llm.ToolDefinition, 0, len(r.order))
    for _, name := range r.order {
        defs = append(defs, r.tools[name].Definition)
    }
    return defs
}

// Execute runs a tool by name with the given input.
func (r *ToolRegistry) Execute(
    ctx context.Context,
    name string,
    input json.RawMessage,
) (json.RawMessage, error) {
    r.mu.RLock()
    tool, ok := r.tools[name]
    r.mu.RUnlock()
    if !ok {
        return nil, fmt.Errorf("unknown tool: %s", name)
    }
    return tool.Execute(ctx, input)
}
```

**Step 4: Run tests**

Run: `go test -v ./internal/conversation/ -run TestToolRegistry`

Expected: All pass.

**Step 5: Commit**

```bash
git add internal/conversation/tools.go \
        internal/conversation/tools_test.go
git commit -m "feat: add tool registry with definition and execution dispatch"
```

---

## Task 9: Procedural Tool Implementations

Implement the 8 procedural tools that perform direct
database operations.

**Files:**

- Create: `internal/conversation/tools_procedural.go`
- Create: `internal/conversation/tools_procedural_test.go`

**Step 1: Write failing test for search_entities tool**

```go
func TestSearchEntitiesTool(t *testing.T) {
    // Unit test using a mock DB interface
    input := json.RawMessage(
        `{"query":"innkeeper","entity_type":"npc"}`)
    tool := buildSearchEntitiesTool(mockDB)
    result, err := tool.Execute(
        context.Background(), input)
    require.NoError(t, err)
    assert.NotEmpty(t, result)
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v ./internal/conversation/ -run TestSearchEntitiesTool`

Expected: FAIL — `buildSearchEntitiesTool` not defined.

**Step 3: Implement procedural tools**

Create `internal/conversation/tools_procedural.go`. Each
tool is a factory function that takes a `*database.DB`
and the current campaign ID, and returns a `Tool`.

Key implementation pattern — each tool:

1. Defines a JSON schema for the LLM to follow.
2. Parses the `json.RawMessage` input into a typed struct.
3. Calls the appropriate `db.*` method.
4. Marshals the result back to `json.RawMessage`.

Tools to implement:

- `buildSearchEntitiesTool(db, campaignID)` — calls
  `db.SearchEntities` or `db.ListEntities` with filters.
- `buildGetEntityTool(db)` — calls `db.GetEntity`.
- `buildCreateEntityTool(db, campaignID)` — calls
  `db.CreateEntity`.
- `buildUpdateEntityTool(db)` — calls `db.UpdateEntity`.
- `buildCreateRelationshipTool(db, campaignID)` — calls
  `db.CreateRelationship`.
- `buildGetRelatedEntitiesTool(db)` — calls
  `db.GetRelatedEntities` or the relationships view query.
- `buildSearchContentTool(db, campaignID)` — calls
  `db.SearchCampaignContent`.
- `buildReadGameSchemaTool(schemasDir)` — reads YAML
  from the `schemas/` directory.

Each tool's `InputSchema` is a JSON Schema object. Example
for `search_entities`:

```json
{
  "type": "object",
  "properties": {
    "query": {
      "type": "string",
      "description": "Search term for entity name"
    },
    "entity_type": {
      "type": "string",
      "description": "Filter by entity type (npc, location, etc.)"
    }
  },
  "required": ["query"]
}
```

**Step 4: Run tests**

Run: `go test -v ./internal/conversation/ -run TestSearch`

Expected: All pass.

**Step 5: Write tests for remaining procedural tools**

Add tests for `get_entity`, `create_entity`,
`create_relationship`, `get_related_entities`,
`search_content`, and `read_game_schema`.

**Step 6: Run all procedural tool tests**

Run: `go test -v ./internal/conversation/ -run TestProcedural`

Expected: All pass.

**Step 7: Commit**

```bash
git add internal/conversation/tools_procedural.go \
        internal/conversation/tools_procedural_test.go
git commit -m "feat: add 8 procedural tool implementations for conversation"
```

---

## Task 10: Document Tool Implementations

Implement `read_document` and `edit_document` tools.

**Files:**

- Create: `internal/conversation/tools_document.go`
- Create: `internal/conversation/tools_document_test.go`

**Step 1: Write failing tests**

Test `read_document` with a chapter scope, entity scope.
Test `edit_document` with a replacement operation.

**Step 2: Run tests to verify they fail**

Expected: FAIL — functions not defined.

**Step 3: Implement document tools**

`read_document` accepts `{scope_type, scope_id}` and
returns the text content of the chapter, session, scene,
or entity description. It calls the appropriate `db.Get*`
method based on scope_type.

`edit_document` accepts `{scope_type, scope_id, operation,
old_text, new_text}` where operation is "replace", "insert",
or "append". It loads the document, applies the edit, and
saves it back. Uses the existing `db.Update*` methods.

**Step 4: Run tests**

Expected: All pass.

**Step 5: Commit**

```bash
git add internal/conversation/tools_document.go \
        internal/conversation/tools_document_test.go
git commit -m "feat: add read_document and edit_document tools"
```

---

## Task 11: Agent Tool Implementations

Wrap the existing TTRPG, Canon, and Graph experts as
agent tools with sub-tool access.

**Files:**

- Create: `internal/conversation/tools_agent.go`
- Create: `internal/conversation/tools_agent_test.go`

**Step 1: Write failing tests**

Test that `ask_ttrpg_expert` assembles the right context
and calls the LLM with the specialist system prompt and
sub-tool definitions. Use a mock streaming provider.

**Step 2: Run tests to verify they fail**

Expected: FAIL — functions not defined.

**Step 3: Implement agent tools**

Each agent tool:

1. Builds a sub-tool registry from the allowed procedural
   tools (TTRPG gets 6 sub-tools, Canon gets 4, Graph
   gets 3).
2. Loads the expert's system prompt from the existing
   `internal/agents/*/prompts.go` files.
3. Calls `streamingProvider.CompleteStream` with the
   specialist prompt, the user's question, and the
   sub-tool definitions.
4. Runs its own mini tool-use loop for sub-tool calls.
5. Returns the expert's final text response.

The agent tool factory:

```go
func buildAgentTool(
    name string,
    systemPromptFn func() string,
    subTools *ToolRegistry,
    provider llm.StreamingProvider,
) Tool
```

Use the existing system prompts:

- TTRPG: `ttrpg.BuildSystemPrompt(...)` from
  `internal/agents/ttrpg/prompts.go`
- Canon: `canon.BuildSystemPrompt(...)` from
  `internal/agents/canon/prompts.go`
- Graph: `graph.BuildSystemPrompt(...)` from
  `internal/agents/graph/prompts.go`

The TTRPG expert's prompt must be extended with generative
capabilities. Add a section to the system prompt for
conversation mode that enables:

- Scene design and NPC voice generation
- Encounter building with schema validation
- Session structuring with time estimates
- World-grounded generation flow (the scope resolution,
  orphan scan, relationship gap query sequence)

**Step 4: Run tests**

Expected: All pass.

**Step 5: Commit**

```bash
git add internal/conversation/tools_agent.go \
        internal/conversation/tools_agent_test.go
git commit -m "feat: wrap existing experts as agent tools with sub-tool access"
```

---

## Task 12: Session Cache

Implement the in-memory conversation state cache with idle
eviction.

**Files:**

- Create: `internal/conversation/session_cache.go`
- Create: `internal/conversation/session_cache_test.go`

**Step 1: Write failing tests**

```go
func TestSessionCacheGetSet(t *testing.T) {
    cache := NewSessionCache(30 * time.Minute)
    entry := &CacheEntry{
        Messages: []llm.StreamingMessage{
            {Role: "user", Content: "Hello"},
        },
    }
    cache.Set(42, entry)
    got, ok := cache.Get(42)
    assert.True(t, ok)
    assert.Len(t, got.Messages, 1)
}

func TestSessionCacheEviction(t *testing.T) {
    cache := NewSessionCache(10 * time.Millisecond)
    cache.Set(1, &CacheEntry{})
    time.Sleep(50 * time.Millisecond)
    _, ok := cache.Get(1)
    assert.False(t, ok)
}

func TestSessionCacheInvalidate(t *testing.T) {
    cache := NewSessionCache(30 * time.Minute)
    cache.Set(1, &CacheEntry{})
    cache.Invalidate(1)
    _, ok := cache.Get(1)
    assert.False(t, ok)
}
```

**Step 2: Run tests to verify they fail**

Expected: FAIL — types not defined.

**Step 3: Implement**

Create `internal/conversation/session_cache.go`:

```go
// CacheEntry holds in-memory conversation state.
type CacheEntry struct {
    Messages      []llm.StreamingMessage
    SystemPrompt  string
    ToolDefs      []llm.ToolDefinition
    LastAccessed  time.Time
}

// SessionCache provides idle-evicting conversation state.
type SessionCache struct {
    mu       sync.RWMutex
    entries  map[int64]*CacheEntry
    ttl      time.Duration
    stopCh   chan struct{}
}

func NewSessionCache(ttl time.Duration) *SessionCache
func (c *SessionCache) Get(convID int64)
    (*CacheEntry, bool)
func (c *SessionCache) Set(convID int64,
    entry *CacheEntry)
func (c *SessionCache) Invalidate(convID int64)
func (c *SessionCache) Stop()
```

The cache starts a background goroutine that sweeps
expired entries every `ttl/2`. `Get` updates
`LastAccessed`. `Stop` shuts down the sweeper.

**Step 4: Run tests**

Run: `go test -v ./internal/conversation/ -run TestSessionCache`

Expected: All pass.

**Step 5: Commit**

```bash
git add internal/conversation/session_cache.go \
        internal/conversation/session_cache_test.go
git commit -m "feat: add in-memory session cache with idle eviction"
```

---

## Task 13: SSE Response Writer

Create the SSE writer that formats and sends events to
the HTTP response.

**Files:**

- Create: `internal/conversation/sse.go`
- Create: `internal/conversation/sse_test.go`

**Step 1: Write failing tests**

Test that `SSEWriter.WriteTextDelta` writes the correct
SSE format. Use `httptest.NewRecorder`.

```go
func TestSSEWriterTextDelta(t *testing.T) {
    w := httptest.NewRecorder()
    sse := NewSSEWriter(w)
    err := sse.WriteTextDelta("Hello world")
    require.NoError(t, err)
    assert.Contains(t, w.Body.String(),
        "event: text_delta")
    assert.Contains(t, w.Body.String(),
        `"text":"Hello world"`)
}

func TestSSEWriterToolUse(t *testing.T) {
    w := httptest.NewRecorder()
    sse := NewSSEWriter(w)
    err := sse.WriteToolUse("search_entities",
        json.RawMessage(`{"query":"inn"}`))
    require.NoError(t, err)
    assert.Contains(t, w.Body.String(),
        "event: tool_use")
    assert.Contains(t, w.Body.String(),
        "search_entities")
}

func TestSSEWriterDone(t *testing.T) {
    w := httptest.NewRecorder()
    sse := NewSSEWriter(w)
    err := sse.WriteDone(42, 1200, 380)
    require.NoError(t, err)
    assert.Contains(t, w.Body.String(),
        "event: done")
    assert.Contains(t, w.Body.String(),
        `"message_id":42`)
}
```

**Step 2: Run tests to verify they fail**

Expected: FAIL — `NewSSEWriter` not defined.

**Step 3: Implement**

Create `internal/conversation/sse.go`:

```go
type SSEWriter struct {
    w       http.ResponseWriter
    flusher http.Flusher
}

func NewSSEWriter(w http.ResponseWriter) *SSEWriter
func (s *SSEWriter) SetHeaders()
func (s *SSEWriter) WriteTextDelta(text string) error
func (s *SSEWriter) WriteToolUse(
    name string, input json.RawMessage) error
func (s *SSEWriter) WriteToolResult(
    name string, result json.RawMessage) error
func (s *SSEWriter) WriteDone(
    messageID int64, inputTokens, outputTokens int,
) error
func (s *SSEWriter) WriteError(msg string) error
```

Each method formats the event as:
```
event: <type>\ndata: <json>\n\n
```
and calls `flusher.Flush()`.

**Step 4: Run tests**

Run: `go test -v ./internal/conversation/ -run TestSSEWriter`

Expected: All pass.

**Step 5: Commit**

```bash
git add internal/conversation/sse.go \
        internal/conversation/sse_test.go
git commit -m "feat: add SSE response writer for conversation streaming"
```

---

## Task 14: System Prompt Template

Create the system prompt template file and the loader.

**Files:**

- Create: `config/prompts/conversation.tmpl`
- Create: `internal/conversation/prompt.go`
- Create: `internal/conversation/prompt_test.go`

**Step 1: Create the template**

Create `config/prompts/conversation.tmpl` using Go's
`text/template` syntax. The template receives a struct
with campaign name, scope type, scope ID, game system
name, current date, and entity context.

The prompt should establish the AI as a collaborative
world-building partner who:

- Knows the campaign context
- Uses tools to ground responses in the world model
- Follows the world-grounded generation principle
- Creates connective tissue between new and existing
  content
- Never generates in isolation
- Asks for confirmation before creating major entities

**Step 2: Write the template loader**

Create `internal/conversation/prompt.go`:

```go
type PromptContext struct {
    CampaignName  string
    ScopeType     string
    ScopeID       int64
    GameSystem    string
    CurrentDate   string
    EntityContext string
}

func LoadSystemPrompt(
    templatePath string,
    ctx PromptContext,
) (string, error)
```

Uses `text/template` to parse and execute the template.

**Step 3: Write tests**

Test that the loader renders the template with campaign
name and scope type substituted. Test missing template
file returns error.

**Step 4: Run tests**

Expected: All pass.

**Step 5: Commit**

```bash
git add config/prompts/conversation.tmpl \
        internal/conversation/prompt.go \
        internal/conversation/prompt_test.go
git commit -m "feat: add system prompt template and loader"
```

---

## Task 15: Orchestrator — Core Loop

Implement the orchestrator that manages the streaming
conversation loop with tool execution.

**Files:**

- Create: `internal/conversation/orchestrator.go`
- Create: `internal/conversation/orchestrator_test.go`

**Step 1: Define the orchestrator struct**

```go
type Orchestrator struct {
    db            *database.DB
    provider      llm.StreamingProvider
    tools         *ToolRegistry
    cache         *SessionCache
    promptPath    string
    schemasDir    string
}

type OrchestratorConfig struct {
    DB            *database.DB
    Provider      llm.StreamingProvider
    PromptPath    string
    SchemasDir    string
    CacheTTL      time.Duration
}

func NewOrchestrator(cfg OrchestratorConfig)
    *Orchestrator
```

**Step 2: Write failing test for HandleMessage**

The main method is `HandleMessage` which:

1. Loads conversation context.
2. Appends the user message.
3. Calls the streaming provider.
4. Processes stream events.
5. Executes tools on tool_use events.
6. Feeds tool results back for continuation.
7. Persists messages.
8. Returns stream events via a channel.

Test with a mock streaming provider that returns a simple
text response (no tool use).

```go
func TestOrchestratorSimpleResponse(t *testing.T) {
    mockProvider := &mockStreamingProvider{
        events: []llm.StreamEvent{
            {Type: llm.EventTextDelta,
             Text: "Hello!"},
            {Type: llm.EventDone},
        },
    }
    orch := NewOrchestrator(OrchestratorConfig{
        DB:       mockDB,
        Provider: mockProvider,
    })

    events, err := orch.HandleMessage(
        context.Background(),
        conversationID,
        campaignID,
        userID,
        "Hi there",
    )
    require.NoError(t, err)

    var collected []llm.StreamEvent
    for ev := range events {
        collected = append(collected, ev)
    }
    assert.True(t, len(collected) >= 2)
}
```

**Step 3: Run test to verify it fails**

Expected: FAIL — `HandleMessage` not defined.

**Step 4: Implement HandleMessage**

The core loop:

```go
func (o *Orchestrator) HandleMessage(
    ctx context.Context,
    conversationID int64,
    campaignID int64,
    userID int64,
    content string,
) (<-chan llm.StreamEvent, error) {
    outCh := make(chan llm.StreamEvent, 64)

    go func() {
        defer close(outCh)
        o.runLoop(ctx, outCh,
            conversationID, campaignID,
            userID, content)
    }()

    return outCh, nil
}

func (o *Orchestrator) runLoop(
    ctx context.Context,
    outCh chan<- llm.StreamEvent,
    conversationID int64,
    campaignID int64,
    userID int64,
    userContent string,
) {
    // 1. Persist user message
    // 2. Build messages list from cache/DB
    // 3. Assemble system prompt
    // 4. Build streaming request
    // 5. Call provider.CompleteStream
    // 6. Forward text deltas to outCh
    // 7. On tool_use: execute tool, send result
    //    to outCh, append to messages, loop back
    //    to step 4
    // 8. On done: persist assistant message,
    //    log token usage, check compaction
    //    threshold, send done event
}
```

The tool execution loop continues until the LLM's
`stop_reason` is `"end_turn"` (not `"tool_use"`).

**Step 5: Write test for tool use loop**

Test with a mock provider that first returns a tool_use
event, then on the second call (with tool result) returns
a text response.

**Step 6: Run tests**

Expected: All pass.

**Step 7: Commit**

```bash
git add internal/conversation/orchestrator.go \
        internal/conversation/orchestrator_test.go
git commit -m "feat: add conversation orchestrator with tool use loop"
```

---

## Task 16: Compaction

Implement conversation compaction with hybrid summarisation
and ground truth validation.

**Files:**

- Create: `internal/conversation/compaction.go`
- Create: `internal/conversation/compaction_test.go`

**Step 1: Write failing tests**

Test that `ShouldCompact` returns true when total tokens
exceed 8000. Test that `Compact` produces a summary and
marks old messages as compacted.

**Step 2: Run tests to verify they fail**

Expected: FAIL.

**Step 3: Implement compaction**

```go
const CompactionThreshold = 8000

func (o *Orchestrator) ShouldCompact(
    messages []models.Message,
) bool {
    total := 0
    for _, m := range messages {
        if m.Tokens != nil {
            total += *m.Tokens
        }
    }
    return total > CompactionThreshold
}

func (o *Orchestrator) Compact(
    ctx context.Context,
    conversationID int64,
    campaignID int64,
    userID int64,
) error {
    // 1. Call db.AssembleConversationContext()
    //    to get uncompacted messages + summary
    // 2. Split into "old" (to summarise) and
    //    "recent" (to keep)
    // 3. Build compaction prompt with old
    //    messages and existing summary
    // 4. Call streaming provider with
    //    summarisation prompt + sub-tools for
    //    ground truth validation
    // 5. Call db.CompactMessages() which
    //    atomically marks messages compacted
    //    and updates the summary in one
    //    stored function call
    // 6. Log token usage as "compaction" type
}
```

The compaction prompt instructs the LLM to:

- Summarise the conversation into key decisions, entities
  discussed, actions taken, and unresolved threads
- Merge with existing summary (if any)
- Use `search_entities` and `get_entity` tools to validate
  facts against current world model state
- Correct any stale information in the summary

**Step 4: Run tests**

Expected: All pass.

**Step 5: Commit**

```bash
git add internal/conversation/compaction.go \
        internal/conversation/compaction_test.go
git commit -m "feat: add conversation compaction with ground truth validation"
```

---

## Task 17: API Handlers

Create the HTTP handlers for the four conversation
endpoints.

**Files:**

- Create: `internal/api/conversation_handlers.go`
- Create: `internal/api/conversation_handlers_test.go`

**Step 1: Write failing tests**

Test the create, get, list, and send message endpoints.
Use `httptest` for HTTP-level testing.

```go
func TestGetOrCreateConversationHandler(t *testing.T) {
    // POST /api/campaigns/1/conversations
    // Body: {"scopeType":"chapter","scopeId":1}
    // Expect: 200 or 201 with conversation object
    // Uses get_or_create_conversation() stored fn
}

func TestGetConversationHandler(t *testing.T) {
    // GET /api/campaigns/1/conversations/1
    // Expect: 200 with conversation + messages
}

func TestListConversationsHandler(t *testing.T) {
    // GET /api/campaigns/1/conversations
    // Expect: 200 with array
}

func TestSendMessageHandler(t *testing.T) {
    // POST /api/campaigns/1/conversations/1/messages
    // Body: {"content":"Hello"}
    // Expect: SSE stream with Content-Type:
    //         text/event-stream
}
```

**Step 2: Run tests to verify they fail**

Expected: FAIL — handler not defined.

**Step 3: Implement handlers**

Create `internal/api/conversation_handlers.go`:

```go
type ConversationHandler struct {
    db           *database.DB
    orchestrator *conversation.Orchestrator
}

func NewConversationHandler(
    db *database.DB,
    orch *conversation.Orchestrator,
) *ConversationHandler

func (h *ConversationHandler) Create(
    w http.ResponseWriter, r *http.Request)

func (h *ConversationHandler) Get(
    w http.ResponseWriter, r *http.Request)

func (h *ConversationHandler) List(
    w http.ResponseWriter, r *http.Request)

func (h *ConversationHandler) SendMessage(
    w http.ResponseWriter, r *http.Request)
```

`SendMessage` is the SSE streaming handler. It:

1. Parses the request body (`SendMessageRequest`).
2. Verifies campaign ownership.
3. Sets SSE headers.
4. Calls `orchestrator.HandleMessage`.
5. Reads from the returned channel and writes SSE events
   using the `SSEWriter`.

**Step 4: Run tests**

Expected: All pass.

**Step 5: Commit**

```bash
git add internal/api/conversation_handlers.go \
        internal/api/conversation_handlers_test.go
git commit -m "feat: add conversation API handlers with SSE streaming"
```

---

## Task 18: Router Registration

Register the conversation routes in the router.

**Files:**

- Modify: `internal/api/router.go`
- Modify: `cmd/server/main.go`

**Step 1: Add routes to router.go**

Add conversation routes under the campaigns group:

```go
// Inside the campaigns route group, after existing
// routes:
r.Route("/{campaignId}/conversations", func(r chi.Router) {
    r.Post("/", convHandler.Create)
    r.Get("/", convHandler.List)
    r.Route("/{conversationId}", func(r chi.Router) {
        r.Get("/", convHandler.Get)
        r.Post("/messages", convHandler.SendMessage)
    })
})
```

**Step 2: Wire up in main.go**

The `NewRouter` function needs to accept the orchestrator
or the conversation handler. Add the necessary
constructor arguments and create the handler.

The orchestrator needs:

- `*database.DB` (already available)
- `llm.StreamingProvider` (create from user's API key)
- System prompt template path
- Schemas directory

Create the streaming provider from the user's stored
Anthropic API key. This requires the user to have
configured their API key in account settings.

**Step 3: Verify compilation**

Run: `go build ./cmd/server/...`

Expected: Compiles without errors.

**Step 4: Commit**

```bash
git add internal/api/router.go cmd/server/main.go
git commit -m "feat: register conversation routes and wire orchestrator"
```

---

## Task 19: Tool Registration Factory

Create a factory function that builds and registers all
13 tools for a conversation context.

**Files:**

- Create: `internal/conversation/registry_factory.go`
- Create: `internal/conversation/registry_factory_test.go`

**Step 1: Write failing test**

Test that `BuildToolRegistry` returns a registry with
all 13 tools registered.

**Step 2: Implement**

```go
func BuildToolRegistry(
    db *database.DB,
    campaignID int64,
    provider llm.StreamingProvider,
    schemasDir string,
) *ToolRegistry {
    registry := NewToolRegistry()

    // Procedural tools (8)
    registry.Register(
        buildSearchEntitiesTool(db, campaignID))
    registry.Register(
        buildGetEntityTool(db))
    registry.Register(
        buildCreateEntityTool(db, campaignID))
    registry.Register(
        buildUpdateEntityTool(db))
    registry.Register(
        buildCreateRelationshipTool(db, campaignID))
    registry.Register(
        buildGetRelatedEntitiesTool(db))
    registry.Register(
        buildSearchContentTool(db, campaignID))
    registry.Register(
        buildReadGameSchemaTool(schemasDir))

    // Document tools (2)
    registry.Register(
        buildReadDocumentTool(db))
    registry.Register(
        buildEditDocumentTool(db))

    // Agent tools (3)
    // Build sub-tool registries for each agent
    ttrpgSubTools := buildTTRPGSubToolRegistry(
        db, campaignID, schemasDir)
    canonSubTools := buildCanonSubToolRegistry(
        db, campaignID)
    graphSubTools := buildGraphSubToolRegistry(
        db, campaignID)

    registry.Register(
        buildAgentTool("ask_ttrpg_expert",
            ttrpg.ConversationSystemPrompt,
            ttrpgSubTools, provider))
    registry.Register(
        buildAgentTool("ask_canon_expert",
            canon.ConversationSystemPrompt,
            canonSubTools, provider))
    registry.Register(
        buildAgentTool("ask_graph_expert",
            graph.ConversationSystemPrompt,
            graphSubTools, provider))

    return registry
}
```

**Step 3: Run test**

Expected: PASS — registry contains 13 tools.

**Step 4: Commit**

```bash
git add internal/conversation/registry_factory.go \
        internal/conversation/registry_factory_test.go
git commit -m "feat: add tool registry factory with all 13 tools"
```

---

## Task 20: Expert System Prompts for Conversation Mode

Extend the existing expert system prompts with
conversation-mode capabilities.

**Files:**

- Modify: `internal/agents/ttrpg/prompts.go`
- Modify: `internal/agents/canon/prompts.go`
- Modify: `internal/agents/graph/prompts.go`
- Test: `internal/agents/ttrpg/prompts_test.go`

**Step 1: Add conversation system prompt to TTRPG expert**

Add a `ConversationSystemPrompt` function alongside the
existing `BuildSystemPrompt`. The conversation prompt
extends the existing 8-dimension analysis with generative
capabilities:

- Scene design, encounter building, session planning
- NPC voice and dialogue generation
- Stat block generation (using `read_game_schema` tool)
- World-grounded generation flow instructions
- Reconciliation between prep and play

The prompt instructs the expert to always:

1. Query scope context first
2. Search for orphaned/underconnected entities
3. Weave new content into existing world fabric
4. Reference established NPCs, locations, factions

**Step 2: Add conversation prompt to Canon expert**

Add `ConversationSystemPrompt` with active capabilities:

- Fact verification ("is it safe to say X?")
- Knowledge compilation ("what do we know about X?")
- Hypothetical checking
- Conflict resolution guidance

**Step 3: Add conversation prompt to Graph expert**

Add `ConversationSystemPrompt` with active capabilities:

- Connection suggestions for new entities
- Path discovery between entities
- Impact analysis for structural changes
- Deduplication detection

**Step 4: Write tests for prompt rendering**

Test that each conversation prompt renders without errors
and contains key sections.

**Step 5: Run tests**

Run: `go test -v ./internal/agents/...`

Expected: All pass.

**Step 6: Commit**

```bash
git add internal/agents/ttrpg/prompts.go \
        internal/agents/canon/prompts.go \
        internal/agents/graph/prompts.go \
        internal/agents/ttrpg/prompts_test.go
git commit -m "feat: add conversation-mode system prompts for all three experts"
```

---

## Task 21: Integration Test — End-to-End Conversation

Write an integration test that exercises the full
conversation flow: create conversation, send message,
receive streaming response with tool use.

**Files:**

- Create: `internal/conversation/integration_test.go`

**Step 1: Write integration test**

```go
// +build integration

func TestConversationEndToEnd(t *testing.T) {
    // 1. Set up real DB with test campaign and entities
    // 2. Create a mock streaming provider that simulates
    //    a tool_use response followed by a text response
    // 3. Create orchestrator with real DB + mock provider
    // 4. Call HandleMessage("Find NPCs near the inn")
    // 5. Verify:
    //    - User message persisted
    //    - Tool call events emitted (search_entities)
    //    - Tool result events emitted
    //    - Text delta events emitted
    //    - Done event emitted
    //    - Assistant message persisted
    //    - Token usage logged
}
```

**Step 2: Run test**

Run: `go test -v -tags integration ./internal/conversation/ -run TestConversationEndToEnd`

Expected: PASS.

**Step 3: Commit**

```bash
git add internal/conversation/integration_test.go
git commit -m "test: add end-to-end conversation integration test"
```

---

## Task 22: Final Verification

Run all tests and verify the full build.

**Files:** None (verification only).

**Step 1: Run all Go tests**

Run: `make test-server`

Expected: All tests pass.

**Step 2: Run lint**

Run: `make lint`

Expected: No lint errors.

**Step 3: Run full test suite**

Run: `make test-all`

Expected: All suites pass.

**Step 4: Verify server starts**

Run: `go build ./cmd/server/...`

Expected: Compiles without errors.

**Step 5: Manual smoke test with curl**

```bash
# Start server
make run &

# Create conversation
curl -X POST http://localhost:3001/api/campaigns/1/conversations \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"scopeType":"chapter","scopeId":1}'

# Send message (SSE stream)
curl -N -X POST http://localhost:3001/api/campaigns/1/conversations/1/messages \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"content":"Tell me about the NPCs in Vienna"}'
```

Expected: SSE events stream to the terminal.

**Step 6: Commit any final fixes**

---

## Summary

| Task | Description | Notes |
|------|-------------|-------|
| 1 | Database migration | Triggers, views, 4 stored fns |
| 2 | Conversation models | +ConversationContext, ListItem, TokenUsageSummary |
| 3 | Conversation CRUD | Uses `get_or_create_conversation()`, `conversation_list` view |
| 4 | Message DB layer | Uses `assemble_conversation_context()`, `compact_messages()` |
| 5 | Token usage logging | Uses `get_token_usage_summary()` |
| 6 | Streaming provider interface | 4 steps |
| 7 | Anthropic streaming impl | 5 steps |
| 8 | Tool registry | 5 steps |
| 9 | Procedural tools (8) | 7 steps |
| 10 | Document tools (2) | 5 steps |
| 11 | Agent tools (3) | 5 steps |
| 12 | Session cache | 5 steps |
| 13 | SSE writer | 5 steps |
| 14 | System prompt template | 5 steps |
| 15 | Orchestrator core loop | 7 steps |
| 16 | Compaction | Calls `db.CompactMessages()` (stored fn) |
| 17 | API handlers | 5 steps |
| 18 | Router registration | 4 steps |
| 19 | Tool registry factory | 4 steps |
| 20 | Expert conversation prompts | 6 steps |
| 21 | Integration test | 3 steps |
| 22 | Final verification | 6 steps |

**Key design decisions:**

- **messages.tool_use_id** column correlates tool_call
  to tool_result for context reconstruction.
- **Triggers** handle `updated_at` on conversations
  automatically -- no manual UPDATE in Go code.
- **Stored functions** encapsulate critical operations:
  - `get_or_create_conversation()` -- atomic upsert
  - `assemble_conversation_context()` -- hot-path
    single round-trip context assembly
  - `compact_messages()` -- atomic mark + summary
    update
  - `get_token_usage_summary()` -- flexible
    aggregation
- **Views** provide pre-joined read paths:
  - `conversation_list` -- list with preview/counts
  - `conversation_context` -- uncompacted messages
  - `token_usage_by_campaign` / `token_usage_by_user`

Each step is 2-5 minutes. Tasks are ordered by
dependency: foundation first (migration, models, DB
layer), then infrastructure (streaming, tools, cache),
then orchestration, then API surface, then verification.

**Independent tasks that can run in parallel:**

- Tasks 6-7 (streaming provider) have no dependency on
  Tasks 3-5 (DB layer).
- Task 12 (session cache) is independent of Tasks 9-11
  (tool implementations).
- Task 13 (SSE writer) is independent of everything
  except the package existing.
- Task 14 (system prompt) is independent.

**Critical path:**

Tasks 1-2 → 3-5 → 8-9 → 15 → 17-18 → 22
