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
	"time"

	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/antonypegg/imagineer/internal/models"
)

// maxOrchestratorIterations limits the number of
// tool-use round trips the orchestrator can perform
// before it must stop. This prevents runaway loops.
const maxOrchestratorIterations = 10

// defaultSystemPrompt is used when no prompt template
// path is configured.
const defaultSystemPrompt = "You are a helpful assistant."

// defaultCacheTTL is used when the config specifies
// a zero TTL.
const defaultCacheTTL = 30 * time.Minute

// Orchestrator manages the streaming conversation loop
// with tool execution. It coordinates between the LLM
// provider, the tool registry, the session cache, and
// the database layer.
type Orchestrator struct {
	db         *database.DB
	provider   llm.StreamingProvider
	tools      *ToolRegistry
	cache      *SessionCache
	promptPath string
	schemasDir string
}

// OrchestratorConfig holds the configuration for
// creating a new Orchestrator.
type OrchestratorConfig struct {
	DB         *database.DB
	Provider   llm.StreamingProvider
	Tools      *ToolRegistry
	PromptPath string
	SchemasDir string
	CacheTTL   time.Duration
}

// NewOrchestrator creates an orchestrator with the
// given configuration. If CacheTTL is zero, it
// defaults to 30 minutes.
func NewOrchestrator(cfg OrchestratorConfig) *Orchestrator {
	ttl := cfg.CacheTTL
	if ttl == 0 {
		ttl = defaultCacheTTL
	}

	return &Orchestrator{
		db:         cfg.DB,
		provider:   cfg.Provider,
		tools:      cfg.Tools,
		cache:      NewSessionCache(ttl),
		promptPath: cfg.PromptPath,
		schemasDir: cfg.SchemasDir,
	}
}

// Stop shuts down background resources such as the
// session cache sweeper. Call this when the orchestrator
// is no longer needed.
func (o *Orchestrator) Stop() {
	if o.cache != nil {
		o.cache.Stop()
	}
}

// HandleMessage starts processing a user message in a
// background goroutine and returns a channel of stream
// events. The channel is closed when processing
// completes or an error occurs.
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
		o.runLoop(
			ctx, outCh,
			conversationID, campaignID,
			userID, content,
		)
	}()

	return outCh, nil
}

// sendEvent sends an event to the output channel,
// respecting context cancellation.
func sendEvent(
	ctx context.Context,
	ch chan<- llm.StreamEvent,
	event llm.StreamEvent,
) bool {
	select {
	case <-ctx.Done():
		return false
	case ch <- event:
		return true
	}
}

// sendError sends an error event to the output channel.
func sendError(
	ctx context.Context,
	ch chan<- llm.StreamEvent,
	err error,
) {
	sendEvent(ctx, ch, llm.StreamEvent{
		Type:  llm.EventError,
		Error: err,
	})
}

// runLoop is the core conversation loop. It persists
// the user message, loads context, builds the request,
// streams the response, and handles tool calls.
func (o *Orchestrator) runLoop(
	ctx context.Context,
	outCh chan<- llm.StreamEvent,
	conversationID int64,
	campaignID int64,
	userID int64,
	content string,
) {
	// Ensure tools registry is not nil.
	tools := o.tools
	if tools == nil {
		tools = NewToolRegistry()
	}

	// Step 1: Persist user message.
	if o.db != nil {
		_, err := o.db.CreateMessage(
			ctx, conversationID,
			models.MessageRoleUser,
			content,
			nil, nil, nil, nil, nil,
		)
		if err != nil {
			sendError(ctx, outCh, fmt.Errorf(
				"persisting user message: %w", err))
			return
		}
	}

	// Step 2: Load conversation context.
	var messages []llm.StreamingMessage

	if o.cache != nil {
		if entry, ok := o.cache.Get(conversationID); ok {
			messages = entry.Messages
		}
	}

	if messages == nil && o.db != nil {
		cc, err := o.db.AssembleConversationContext(
			ctx, conversationID)
		if err != nil {
			sendError(ctx, outCh, fmt.Errorf(
				"loading conversation context: %w", err))
			return
		}
		messages = messagesToStreaming(cc.Messages)
	}

	// Append the current user message.
	messages = append(messages, llm.StreamingMessage{
		Role:    "user",
		Content: content,
	})

	// Step 3: Build system prompt.
	systemPrompt := defaultSystemPrompt
	if o.promptPath != "" {
		rendered, err := LoadSystemPrompt(
			o.promptPath, PromptContext{})
		if err != nil {
			sendError(ctx, outCh, fmt.Errorf(
				"loading system prompt: %w", err))
			return
		}
		systemPrompt = rendered
	}

	// Step 4-8: Streaming loop with tool use.
	var totalUsage llm.TokenUsage

	for iteration := 0; iteration < maxOrchestratorIterations; iteration++ {
		if ctx.Err() != nil {
			sendError(ctx, outCh, fmt.Errorf(
				"context cancelled: %w", ctx.Err()))
			return
		}

		// Build streaming request.
		req := llm.StreamingRequest{
			SystemPrompt: systemPrompt,
			Messages:     messages,
			Tools:        tools.Definitions(),
			MaxTokens:    4096,
			Temperature:  0.7,
		}

		// Call provider.
		llmCh, err := o.provider.CompleteStream(ctx, req)
		if err != nil {
			sendError(ctx, outCh, fmt.Errorf(
				"LLM stream failed: %w", err))
			return
		}

		// Process events from the LLM channel.
		var accumulated string
		var pendingCalls []toolCall

		for event := range llmCh {
			switch event.Type {
			case llm.EventTextDelta:
				accumulated += event.Text
				if !sendEvent(ctx, outCh, event) {
					return
				}

			case llm.EventToolUse:
				pendingCalls = append(pendingCalls,
					toolCall{
						Name:  event.ToolName,
						ID:    event.ToolID,
						Input: event.ToolInput,
					})

			case llm.EventUsage:
				if event.Usage != nil {
					totalUsage.InputTokens += event.Usage.InputTokens
					totalUsage.OutputTokens += event.Usage.OutputTokens
				}

			case llm.EventError:
				sendEvent(ctx, outCh, event)
				return

			case llm.EventDone:
				// Processing continues below.
			}
		}

		// If tool calls are present, execute them.
		if len(pendingCalls) > 0 {
			for _, call := range pendingCalls {
				if ctx.Err() != nil {
					sendError(ctx, outCh, fmt.Errorf(
						"context cancelled: %w",
						ctx.Err()))
					return
				}

				// Forward tool use event for SSE
				// visibility.
				if !sendEvent(ctx, outCh,
					llm.StreamEvent{
						Type:      llm.EventToolUse,
						ToolName:  call.Name,
						ToolID:    call.ID,
						ToolInput: call.Input,
					}) {
					return
				}

				// Execute tool.
				result, execErr := tools.Execute(
					ctx, call.Name, call.Input)
				if execErr != nil {
					result = []byte(fmt.Sprintf(
						`{"error":%q}`,
						execErr.Error()))
				}

				// Append assistant tool call and
				// result messages.
				messages = append(messages,
					llm.StreamingMessage{
						Role:      "assistant",
						ToolName:  call.Name,
						ToolUseID: call.ID,
						ToolInput: call.Input,
					},
					llm.StreamingMessage{
						Role:       "user",
						ToolUseID:  call.ID,
						ToolResult: result,
					},
				)

				// Persist tool messages if DB is
				// available.
				if o.db != nil {
					toolName := call.Name
					toolID := call.ID
					_, _ = o.db.CreateMessage(
						ctx, conversationID,
						models.MessageRoleToolCall,
						"",
						&toolName, &toolID,
						call.Input, nil, nil,
					)
					_, _ = o.db.CreateMessage(
						ctx, conversationID,
						models.MessageRoleToolResult,
						"",
						&toolName, &toolID,
						nil, result, nil,
					)
				}
			}
			// Loop back for next LLM call.
			continue
		}

		// No tool calls -- this is the final response.

		// Persist assistant message.
		if o.db != nil {
			_, _ = o.db.CreateMessage(
				ctx, conversationID,
				models.MessageRoleAssistant,
				accumulated,
				nil, nil, nil, nil, nil,
			)
		}

		// Log token usage.
		if o.db != nil && (totalUsage.InputTokens > 0 ||
			totalUsage.OutputTokens > 0) {
			_ = o.db.LogTokenUsage(
				ctx, models.TokenUsageLog{
					ConversationID: conversationID,
					CampaignID:     campaignID,
					UserID:         userID,
					Model:          "unknown",
					InputTokens:    totalUsage.InputTokens,
					OutputTokens:   totalUsage.OutputTokens,
					TotalTokens: totalUsage.InputTokens +
						totalUsage.OutputTokens,
					LLMCallType: models.CallTypeConversation,
				})
		}

		// Update cache.
		if o.cache != nil {
			// Append assistant response to messages
			// for caching.
			messages = append(messages,
				llm.StreamingMessage{
					Role:    "assistant",
					Content: accumulated,
				})
			o.cache.Set(conversationID, &CacheEntry{
				Messages:     messages,
				SystemPrompt: systemPrompt,
				ToolDefs:     tools.Definitions(),
			})
		}

		// Send done event.
		sendEvent(ctx, outCh, llm.StreamEvent{
			Type: llm.EventDone,
		})
		return
	}

	// Iteration limit reached.
	sendError(ctx, outCh, fmt.Errorf(
		"orchestrator exceeded maximum iterations (%d)",
		maxOrchestratorIterations))
}

// messagesToStreaming converts database Message records
// to StreamingMessage format for the LLM provider.
func messagesToStreaming(
	msgs []models.Message,
) []llm.StreamingMessage {
	result := make([]llm.StreamingMessage, 0, len(msgs))
	for _, m := range msgs {
		sm := llm.StreamingMessage{
			Role:    string(m.Role),
			Content: m.Content,
		}
		if m.ToolName != nil {
			sm.ToolName = *m.ToolName
		}
		if m.ToolUseID != nil {
			sm.ToolUseID = *m.ToolUseID
		}
		if m.ToolInput != nil {
			sm.ToolInput = m.ToolInput
		}
		if m.ToolResult != nil {
			sm.ToolResult = m.ToolResult
		}
		result = append(result, sm)
	}
	return result
}
