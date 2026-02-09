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
	openaiAPIURL = "https://api.openai.com/v1/chat/completions"
	openaiModel  = "gpt-4o"
)

// OpenAIProvider implements the Provider interface for OpenAI's API.
type OpenAIProvider struct {
	apiKey string
	client *http.Client
}

// NewOpenAIProvider creates a new OpenAI LLM provider.
func NewOpenAIProvider(apiKey string) (*OpenAIProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("openai API key is required")
	}
	return &OpenAIProvider{
		apiKey: apiKey,
		client: &http.Client{Timeout: 120 * time.Second},
	}, nil
}

type openaiRequest struct {
	Model       string          `json:"model"`
	Messages    []openaiMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

type openaiError struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Complete sends a completion request to the OpenAI API.
func (p *OpenAIProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}

	messages := []openaiMessage{}
	if req.SystemPrompt != "" {
		messages = append(messages, openaiMessage{Role: "system", Content: req.SystemPrompt})
	}
	messages = append(messages, openaiMessage{Role: "user", Content: req.UserPrompt})

	body := openaiRequest{
		Model:       openaiModel,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: req.Temperature,
	}

	return doWithRetry(ctx, func(ctx context.Context) (CompletionResponse, int, error) {
		payload, err := json.Marshal(body)
		if err != nil {
			return CompletionResponse{}, 0, fmt.Errorf("failed to marshal request: %w", err)
		}

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, openaiAPIURL, bytes.NewReader(payload))
		if err != nil {
			return CompletionResponse{}, 0, fmt.Errorf("failed to create request: %w", err)
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

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
			var apiErr openaiError
			_ = json.Unmarshal(respBody, &apiErr)
			return CompletionResponse{}, resp.StatusCode, fmt.Errorf(
				"openai API error (status %d): %s", resp.StatusCode, apiErr.Error.Message)
		}

		var result openaiResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return CompletionResponse{}, resp.StatusCode, fmt.Errorf("failed to parse response: %w", err)
		}

		if len(result.Choices) == 0 {
			return CompletionResponse{}, resp.StatusCode, fmt.Errorf("empty response from OpenAI API")
		}

		return CompletionResponse{
			Content:    result.Choices[0].Message.Content,
			TokensUsed: result.Usage.TotalTokens,
		}, resp.StatusCode, nil
	})
}
