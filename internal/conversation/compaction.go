/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package conversation

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/antonypegg/imagineer/internal/models"
)

// CompactionThreshold is the number of estimated
// uncompacted message tokens above which compaction
// should be triggered. The orchestrator's runLoop
// checks this after each completed response.
const CompactionThreshold = 8000

// minCompactableMessages is the minimum number of
// uncompacted messages required before compaction is
// worthwhile. With fewer messages, there is not enough
// material to justify a summarisation call.
const minCompactableMessages = 4

// compactionMaxTokens is the maximum number of tokens
// the LLM may use when generating a compaction summary.
const compactionMaxTokens = 2048

// compactionTemperature controls the creativity of the
// compaction summary. A low value keeps the output
// factual and deterministic.
const compactionTemperature = 0.3

// compactionSystemPrompt instructs the LLM to produce a
// structured conversation summary suitable for replacing
// the original messages.
const compactionSystemPrompt = `You are a conversation summariser for a TTRPG campaign management platform. Your task is to create a concise summary of a conversation between a Game Master and an AI assistant.

## Instructions

1. Summarise the key points of the conversation:
   - Decisions made by the GM
   - Entities created, modified, or discussed
   - Actions taken (edits, searches, etc.)
   - Unresolved questions or threads
   - Important context for continuing the conversation

2. If an existing summary is provided, merge the new information with it.

3. Be concise but comprehensive. The summary replaces the original messages, so nothing important should be lost.

4. Write in third person past tense ("The GM discussed...", "An NPC named X was created...").

5. Format as a structured list with clear sections.`

// ShouldCompact returns true when the total token count
// across the given uncompacted messages exceeds the
// CompactionThreshold. Messages with a nil Tokens field
// are treated as zero tokens.
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

// Compact summarises old messages and marks them as
// compacted in the database. It preserves the most
// recent messages to maintain conversational continuity.
func (o *Orchestrator) Compact(
	ctx context.Context,
	conversationID int64,
	campaignID int64,
	userID int64,
) error {
	if o.db == nil {
		return fmt.Errorf("database is required for compaction")
	}
	if o.provider == nil {
		return fmt.Errorf("LLM provider is required for compaction")
	}

	// Step 1: Load uncompacted messages and existing
	// summary from the database.
	cc, err := o.db.AssembleConversationContext(
		ctx, conversationID)
	if err != nil {
		return fmt.Errorf(
			"loading conversation context: %w", err)
	}

	// Step 2: Check that there are enough messages to
	// justify a summarisation call.
	if len(cc.Messages) < minCompactableMessages {
		return nil
	}

	// Step 3: Split messages into old (to compact) and
	// recent (to keep). The last 2 messages are always
	// preserved to maintain conversational flow.
	splitIdx := len(cc.Messages) - 2
	oldMessages := cc.Messages[:splitIdx]
	existingSummary := ""
	if cc.Conversation.Summary != nil {
		existingSummary = *cc.Conversation.Summary
	}

	// Step 4: Build the compaction prompt and call the
	// LLM to generate a summary.
	userPrompt := BuildCompactionPrompt(
		existingSummary, oldMessages)

	req := llm.StreamingRequest{
		SystemPrompt: compactionSystemPrompt,
		Messages: []llm.StreamingMessage{
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		MaxTokens:   compactionMaxTokens,
		Temperature: compactionTemperature,
	}

	llmCh, err := o.provider.CompleteStream(ctx, req)
	if err != nil {
		return fmt.Errorf(
			"starting compaction stream: %w", err)
	}

	// Step 5: Collect the full summary text from the
	// streaming response.
	var summary strings.Builder
	var usage llm.TokenUsage

	for event := range llmCh {
		switch event.Type {
		case llm.EventTextDelta:
			summary.WriteString(event.Text)
		case llm.EventUsage:
			if event.Usage != nil {
				usage.InputTokens += event.Usage.InputTokens
				usage.OutputTokens += event.Usage.OutputTokens
			}
		case llm.EventError:
			if event.Error != nil {
				return fmt.Errorf(
					"compaction LLM error: %w",
					event.Error)
			}
		}
	}

	newSummary := summary.String()
	if newSummary == "" {
		return fmt.Errorf(
			"compaction produced empty summary")
	}

	// Step 6: Gather the IDs of the old messages to
	// mark as compacted.
	oldIDs := make([]int64, len(oldMessages))
	for i, m := range oldMessages {
		oldIDs[i] = m.ID
	}

	// Step 7: Estimate the token count for the summary
	// using a simple character-based heuristic.
	estimatedTokens := len(newSummary) / 4

	// Step 8: Persist the compaction in the database.
	compactedCount, err := o.db.CompactMessages(
		ctx, conversationID, oldIDs,
		newSummary, estimatedTokens)
	if err != nil {
		return fmt.Errorf(
			"persisting compaction: %w", err)
	}

	slog.Info("conversation compacted",
		"conversation_id", conversationID,
		"compacted_messages", compactedCount,
		"summary_tokens", estimatedTokens,
	)

	// Step 9: Invalidate the cache so the next request
	// rebuilds context from the database.
	if o.cache != nil {
		o.cache.Invalidate(conversationID)
	}

	// Step 10: Log token usage for the compaction call.
	if usage.InputTokens > 0 || usage.OutputTokens > 0 {
		totalTokens := usage.InputTokens + usage.OutputTokens
		if logErr := o.db.LogTokenUsage(
			ctx, models.TokenUsageLog{
				ConversationID: conversationID,
				CampaignID:     campaignID,
				UserID:         userID,
				Model:          o.model,
				InputTokens:    usage.InputTokens,
				OutputTokens:   usage.OutputTokens,
				TotalTokens:    totalTokens,
				LLMCallType:    models.CallTypeCompaction,
			}); logErr != nil {
			slog.Warn(
				"failed to log compaction token usage",
				"conversation_id", conversationID,
				"error", logErr,
			)
		}
	}

	return nil
}

// BuildCompactionPrompt constructs the user-facing
// prompt for the compaction LLM call. It includes
// the existing summary (if any) and formats the
// messages to be summarised as a conversation
// transcript.
func BuildCompactionPrompt(
	existingSummary string,
	messages []models.Message,
) string {
	var b strings.Builder

	if existingSummary != "" {
		b.WriteString("## Existing Summary\n\n")
		b.WriteString(existingSummary)
		b.WriteString("\n\n")
		b.WriteString(
			"Merge the following conversation " +
				"into the existing summary above.\n\n")
	}

	b.WriteString("## Conversation to Summarise\n\n")

	for _, m := range messages {
		role := string(m.Role)
		switch m.Role {
		case models.MessageRoleUser:
			role = "GM"
		case models.MessageRoleAssistant:
			role = "Assistant"
		case models.MessageRoleToolCall:
			role = "Tool Call"
		case models.MessageRoleToolResult:
			role = "Tool Result"
		}

		if m.Content != "" {
			fmt.Fprintf(&b, "**%s**: %s\n\n", role,
				m.Content)
		} else if m.ToolName != nil {
			fmt.Fprintf(&b, "**%s** (%s)\n\n",
				role, *m.ToolName)
		}
	}

	return b.String()
}
