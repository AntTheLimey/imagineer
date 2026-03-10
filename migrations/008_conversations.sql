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

CREATE INDEX idx_messages_conversation_counts
    ON messages(conversation_id, compacted);
COMMENT ON INDEX idx_messages_conversation_counts
    IS 'Supports COUNT aggregations in '
       'conversation_list view.';

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
    -- Validate scope_type at function level for
    -- clear error messages before hitting the
    -- CHECK constraint.
    IF p_scope_type NOT IN (
        'entity', 'chapter', 'session',
        'scene', 'campaign'
    ) THEN
        RAISE EXCEPTION
            'invalid scope_type: %', p_scope_type
            USING ERRCODE = 'invalid_parameter_value';
    END IF;

    INSERT INTO conversations
        (campaign_id, scope_type, scope_id)
    VALUES
        (p_campaign_id, p_scope_type, p_scope_id)
    ON CONFLICT ON CONSTRAINT uq_conversation_scope
    DO UPDATE SET updated_at = conversations.updated_at
    RETURNING conversations.id INTO v_id;

    -- xmax = 0 means the row was freshly inserted
    -- (no prior version existed).
    SELECT (xmax = 0) INTO v_created
    FROM conversations
    WHERE conversations.id = v_id;

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
BEGIN
    RETURN QUERY
    SELECT c.id, c.campaign_id,
        c.scope_type::TEXT, c.scope_id,
        c.summary, c.summary_tokens,
        (c.summary_tokens
         + COALESCE(SUM(m.tokens)
             OVER (), 0))::INT,
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
    -- Verify conversation exists.
    IF NOT EXISTS (
        SELECT 1 FROM conversations
        WHERE id = p_conversation_id
    ) THEN
        RAISE EXCEPTION
            'conversation % not found',
            p_conversation_id
            USING ERRCODE = 'no_data_found';
    END IF;

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
