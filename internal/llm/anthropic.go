/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	anthropicAPIURL     = "https://api.anthropic.com/v1/messages"
	anthropicAPIVersion = "2023-06-01"
	anthropicModel      = "claude-sonnet-4-20250514"
)

// AnthropicProvider implements the Provider interface for Anthropic's API.
type AnthropicProvider struct {
	apiKey string
	client *http.Client
}

// NewAnthropicProvider creates a new Anthropic LLM provider.
func NewAnthropicProvider(apiKey string) (*AnthropicProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("anthropic API key is required")
	}
	return &AnthropicProvider{
		apiKey: apiKey,
		client: &http.Client{Timeout: 120 * time.Second},
	}, nil
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type anthropicError struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Complete sends a completion request to the Anthropic API.
func (p *AnthropicProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}

	body := anthropicRequest{
		Model:     anthropicModel,
		MaxTokens: maxTokens,
		System:    req.SystemPrompt,
		Messages: []anthropicMessage{
			{Role: "user", Content: req.UserPrompt},
		},
	}

	return doWithRetry(ctx, func(ctx context.Context) (CompletionResponse, int, error) {
		payload, err := json.Marshal(body)
		if err != nil {
			return CompletionResponse{}, 0, fmt.Errorf("failed to marshal request: %w", err)
		}

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicAPIURL, bytes.NewReader(payload))
		if err != nil {
			return CompletionResponse{}, 0, fmt.Errorf("failed to create request: %w", err)
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("x-api-key", p.apiKey)
		httpReq.Header.Set("anthropic-version", anthropicAPIVersion)

		resp, err := p.client.Do(httpReq)
		if err != nil {
			return CompletionResponse{}, 0, fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return CompletionResponse{}, resp.StatusCode, fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			var apiErr anthropicError
			_ = json.Unmarshal(respBody, &apiErr)
			return CompletionResponse{}, resp.StatusCode, fmt.Errorf(
				"anthropic API error (status %d): %s", resp.StatusCode, apiErr.Error.Message)
		}

		var result anthropicResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return CompletionResponse{}, resp.StatusCode, fmt.Errorf("failed to parse response: %w", err)
		}

		if len(result.Content) == 0 {
			return CompletionResponse{}, resp.StatusCode, fmt.Errorf("empty response from Anthropic API")
		}

		return CompletionResponse{
			Content:    result.Content[0].Text,
			TokensUsed: result.Usage.InputTokens + result.Usage.OutputTokens,
		}, resp.StatusCode, nil
	})
}
