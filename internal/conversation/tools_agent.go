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
	"encoding/json"
	"fmt"

	"github.com/antonypegg/imagineer/internal/llm"
)

// maxAgentIterations limits the number of tool-use
// round trips an agent tool can perform before it
// must return. This prevents runaway loops.
const maxAgentIterations = 10

// agentToolInput is the input schema for every agent
// tool. The LLM sends a question to the expert.
type agentToolInput struct {
	Question string `json:"question"`
}

// agentToolOutput is the JSON structure returned by
// agent tools. It wraps the expert's final text
// response.
type agentToolOutput struct {
	Response string `json:"response"`
}

// System prompts for each agent tool. These are
// conversational prompts designed for interactive
// use, not the JSON-output enrichment prompts.

const ttrpgExpertPrompt = `You are a TTRPG expert assistant for a campaign management platform. You help Game Masters with:

- Scene design and NPC voice generation
- Encounter building with game system validation
- Session structuring with pacing guidance
- Rules clarification and mechanics advice
- World-building consistency checks

You have access to tools to search and read campaign content, entities, and game system schemas. Use these tools to ground your responses in the campaign's established world.

When creating content:
1. First search for relevant existing entities and content
2. Check for potential naming conflicts before suggesting new characters
3. Validate mechanics against the game system schema
4. Ensure new content is consistent with established facts

Respond conversationally. Be helpful, creative, and accurate.`

const canonExpertPrompt = `You are a canon consistency expert for a TTRPG campaign. You help Game Masters maintain continuity by:

- Checking new content against established facts
- Identifying potential contradictions before they become problems
- Tracking timeline consistency
- Validating character behaviour against established personalities

You have access to tools to search campaign content and entities. Use these to verify facts before making claims about campaign canon.

Be conservative: only flag genuine contradictions, not expansions of existing lore. When you find a conflict, clearly state what contradicts what and suggest how to resolve it.

Respond conversationally. Be precise and cite your sources.`

const graphExpertPrompt = `You are a knowledge graph expert for a TTRPG campaign. You help Game Masters understand and manage entity relationships by:

- Analysing relationship structures between entities
- Identifying missing or redundant relationships
- Suggesting new connections based on narrative content
- Detecting potential duplicate entities

You have access to tools to search entities and retrieve their relationships. Use these to provide grounded analysis.

Respond conversationally. Explain graph concepts in terms GMs understand.`

// filterTools creates a new ToolRegistry containing
// only the tools whose names appear in the allowed
// list. Tools not found in the source registry are
// silently skipped.
func filterTools(
	source *ToolRegistry,
	names ...string,
) *ToolRegistry {
	filtered := NewToolRegistry()
	for _, name := range names {
		if tool, ok := source.Get(name); ok {
			filtered.Register(tool)
		}
	}
	return filtered
}

// buildAgentTool creates a single agent tool that
// wraps a specialist expert. When invoked, it starts
// a mini conversation with the LLM using the given
// system prompt and sub-tools, running a tool-use loop
// until the expert produces a text response.
func buildAgentTool(
	name string,
	description string,
	systemPrompt string,
	subTools *ToolRegistry,
	provider llm.StreamingProvider,
) Tool {
	return Tool{
		Definition: llm.ToolDefinition{
			Name:        name,
			Description: description,
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"question": {
						"type": "string",
						"description": "The question or request for the expert"
					}
				},
				"required": ["question"]
			}`),
		},
		Execute: func(
			ctx context.Context,
			input json.RawMessage,
		) (json.RawMessage, error) {
			var params agentToolInput
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf(
					"invalid input: %w", err)
			}
			if params.Question == "" {
				return nil, fmt.Errorf(
					"question is required")
			}

			return runAgentLoop(
				ctx, provider, systemPrompt,
				subTools, params.Question)
		},
	}
}

// toolCall captures a single tool invocation from the
// LLM response stream.
type toolCall struct {
	Name  string
	ID    string
	Input json.RawMessage
}

// runAgentLoop executes the agent's tool-use loop.
// It sends the question to the LLM with the expert's
// system prompt and sub-tools, processes any tool
// calls, and repeats until the LLM produces a text
// response or the iteration limit is reached.
func runAgentLoop(
	ctx context.Context,
	provider llm.StreamingProvider,
	systemPrompt string,
	subTools *ToolRegistry,
	question string,
) (json.RawMessage, error) {
	messages := []llm.StreamingMessage{
		{Role: "user", Content: question},
	}

	for iteration := 0; iteration < maxAgentIterations; iteration++ {
		req := llm.StreamingRequest{
			SystemPrompt: systemPrompt,
			Messages:     messages,
			Tools:        subTools.Definitions(),
			MaxTokens:    4096,
			Temperature:  0.7,
		}

		ch, err := provider.CompleteStream(ctx, req)
		if err != nil {
			return nil, fmt.Errorf(
				"agent stream failed: %w", err)
		}

		text, calls, err := collectStreamEvents(ch)
		if err != nil {
			return nil, err
		}

		// If no tool calls were made, return the
		// accumulated text response.
		if len(calls) == 0 {
			return json.Marshal(agentToolOutput{
				Response: text,
			})
		}

		// Append the assistant's text (if any)
		// before the tool calls.
		if text != "" {
			messages = append(messages,
				llm.StreamingMessage{
					Role:    "assistant",
					Content: text,
				})
		}

		// Process each tool call: append the
		// assistant's tool-use message, execute
		// the tool, then append the result.
		for _, call := range calls {
			messages = append(messages,
				llm.StreamingMessage{
					Role:      "assistant",
					ToolName:  call.Name,
					ToolUseID: call.ID,
					ToolInput: call.Input,
				})

			result, execErr := subTools.Execute(
				ctx, call.Name, call.Input)
			if execErr != nil {
				// Return the error as a tool
				// result so the expert can
				// recover gracefully.
				errResult, _ := json.Marshal(
					map[string]string{
						"error": execErr.Error(),
					})
				result = errResult
			}

			messages = append(messages,
				llm.StreamingMessage{
					Role:       "user",
					ToolUseID:  call.ID,
					ToolResult: result,
				})
		}
	}

	return nil, fmt.Errorf(
		"agent exceeded maximum iterations (%d)",
		maxAgentIterations)
}

// collectStreamEvents drains a stream event channel
// and returns accumulated text, any tool calls, and
// the first error encountered.
func collectStreamEvents(
	ch <-chan llm.StreamEvent,
) (string, []toolCall, error) {
	var text string
	var calls []toolCall
	var current *toolCall

	for event := range ch {
		switch event.Type {
		case llm.EventTextDelta:
			text += event.Text

		case llm.EventToolUse:
			// Each EventToolUse carries the
			// complete tool call information.
			current = &toolCall{
				Name:  event.ToolName,
				ID:    event.ToolID,
				Input: event.ToolInput,
			}
			calls = append(calls, *current)

		case llm.EventError:
			return "", nil, fmt.Errorf(
				"agent stream error: %w",
				event.Error)

		case llm.EventDone, llm.EventUsage:
			// Nothing to accumulate.
		}
	}

	return text, calls, nil
}

// BuildAgentTools creates the three agent tools with
// filtered sub-tool registries appropriate for each
// expert's domain.
func BuildAgentTools(
	allTools *ToolRegistry,
	provider llm.StreamingProvider,
) []Tool {
	ttrpgTools := filterTools(allTools,
		"search_entities", "get_entity",
		"create_entity", "update_entity",
		"search_content", "read_game_schema")

	canonTools := filterTools(allTools,
		"search_entities", "get_entity",
		"search_content", "get_related_entities")

	graphTools := filterTools(allTools,
		"search_entities", "get_entity",
		"get_related_entities")

	return []Tool{
		buildAgentTool(
			"ask_ttrpg_expert",
			"Ask a TTRPG expert for help with scene design, "+
				"encounter building, session structuring, "+
				"rules clarification, and world-building. "+
				"The expert can search campaign content and "+
				"validate against game system schemas.",
			ttrpgExpertPrompt,
			ttrpgTools,
			provider,
		),
		buildAgentTool(
			"ask_canon_expert",
			"Ask a canon consistency expert to check new "+
				"content against established facts, identify "+
				"contradictions, and validate timeline and "+
				"character consistency.",
			canonExpertPrompt,
			canonTools,
			provider,
		),
		buildAgentTool(
			"ask_graph_expert",
			"Ask a knowledge graph expert to analyse entity "+
				"relationships, identify missing connections, "+
				"detect duplicates, and suggest relationship "+
				"improvements.",
			graphExpertPrompt,
			graphTools,
			provider,
		),
	}
}
