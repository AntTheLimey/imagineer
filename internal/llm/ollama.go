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
	"os"
	"time"
)

const (
	defaultOllamaHost  = "http://ollama:11434"
	defaultOllamaModel = "llama3.2"
)

// OllamaProvider implements the Provider interface for Ollama's local API.
type OllamaProvider struct {
	host   string
	model  string
	client *http.Client
}

// NewOllamaProvider creates a new Ollama LLM provider. The host defaults
// to http://ollama:11434 (Docker network) but can be overridden via the
// OLLAMA_HOST environment variable. The model defaults to llama3.2 but
// can be overridden via the OLLAMA_MODEL environment variable.
func NewOllamaProvider() (*OllamaProvider, error) {
	host := os.Getenv("OLLAMA_HOST")
	if host == "" {
		host = defaultOllamaHost
	}
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = defaultOllamaModel
	}
	return &OllamaProvider{
		host:   host,
		model:  model,
		client: &http.Client{Timeout: 300 * time.Second},
	}, nil
}

type ollamaRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaResponse struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	EvalCount   int `json:"eval_count"`
	PromptCount int `json:"prompt_eval_count"`
}

// Complete sends a completion request to the Ollama API.
func (p *OllamaProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	messages := []ollamaMessage{}
	if req.SystemPrompt != "" {
		messages = append(messages, ollamaMessage{Role: "system", Content: req.SystemPrompt})
	}
	messages = append(messages, ollamaMessage{Role: "user", Content: req.UserPrompt})

	body := ollamaRequest{
		Model:    p.model,
		Messages: messages,
		Stream:   false,
	}

	apiURL := p.host + "/api/chat"

	return doWithRetry(ctx, func(ctx context.Context) (CompletionResponse, int, error) {
		payload, err := json.Marshal(body)
		if err != nil {
			return CompletionResponse{}, 0, fmt.Errorf("failed to marshal request: %w", err)
		}

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(payload))
		if err != nil {
			return CompletionResponse{}, 0, fmt.Errorf("failed to create request: %w", err)
		}
		httpReq.Header.Set("Content-Type", "application/json")

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
			return CompletionResponse{}, resp.StatusCode, fmt.Errorf(
				"ollama API error (status %d): %s", resp.StatusCode, string(respBody))
		}

		var result ollamaResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return CompletionResponse{}, resp.StatusCode, fmt.Errorf("failed to parse response: %w", err)
		}

		return CompletionResponse{
			Content:    result.Message.Content,
			TokensUsed: result.EvalCount + result.PromptCount,
		}, resp.StatusCode, nil
	})
}
