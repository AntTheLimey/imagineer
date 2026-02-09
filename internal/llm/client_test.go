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
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/antonypegg/imagineer/internal/models"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name    string
		service models.LLMService
		apiKey  string
		wantErr bool
	}{
		{
			name:    "anthropic with key",
			service: models.LLMServiceAnthropic,
			apiKey:  "test-key",
			wantErr: false,
		},
		{
			name:    "anthropic without key",
			service: models.LLMServiceAnthropic,
			apiKey:  "",
			wantErr: true,
		},
		{
			name:    "openai with key",
			service: models.LLMServiceOpenAI,
			apiKey:  "test-key",
			wantErr: false,
		},
		{
			name:    "openai without key",
			service: models.LLMServiceOpenAI,
			apiKey:  "",
			wantErr: true,
		},
		{
			name:    "ollama no key needed",
			service: models.LLMServiceOllama,
			apiKey:  "",
			wantErr: false,
		},
		{
			name:    "unsupported service",
			service: "unsupported",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProvider(tt.service, tt.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && provider == nil {
				t.Error("NewProvider() returned nil provider without error")
			}
		})
	}
}

func TestAnthropicProvider_Complete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("x-api-key") != "test-key" {
			t.Error("missing or incorrect x-api-key header")
		}
		if r.Header.Get("anthropic-version") != anthropicAPIVersion {
			t.Error("missing or incorrect anthropic-version header")
		}

		// Verify request body
		var req anthropicRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Model != anthropicModel {
			t.Errorf("unexpected model: %s", req.Model)
		}
		if len(req.Messages) != 1 || req.Messages[0].Role != "user" {
			t.Error("unexpected messages format")
		}

		resp := anthropicResponse{
			Content: []struct {
				Text string `json:"text"`
			}{{Text: "test response"}},
		}
		resp.Usage.InputTokens = 10
		resp.Usage.OutputTokens = 5

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := &AnthropicProvider{
		apiKey:  "test-key",
		baseURL: server.URL,
		client:  server.Client(),
	}

	t.Run("successful completion", func(t *testing.T) {
		resp, err := provider.Complete(context.Background(), CompletionRequest{
			SystemPrompt: "You are a test assistant",
			UserPrompt:   "Hello",
			MaxTokens:    100,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Content != "test response" {
			t.Errorf("unexpected content: got %q, want %q", resp.Content, "test response")
		}
		if resp.TokensUsed != 15 {
			t.Errorf("unexpected tokens used: got %d, want 15", resp.TokensUsed)
		}
	})

	t.Run("provider creation", func(t *testing.T) {
		p, err := NewAnthropicProvider("test-key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.apiKey != "test-key" {
			t.Error("apiKey not set correctly")
		}
		if p.baseURL != anthropicAPIURL {
			t.Errorf("baseURL not set correctly: got %q, want %q", p.baseURL, anthropicAPIURL)
		}
	})

	t.Run("empty key rejected", func(t *testing.T) {
		_, err := NewAnthropicProvider("")
		if err == nil {
			t.Error("expected error for empty key")
		}
	})
}

func TestOpenAIProvider_Complete(t *testing.T) {
	t.Run("provider creation", func(t *testing.T) {
		p, err := NewOpenAIProvider("test-key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.apiKey != "test-key" {
			t.Error("apiKey not set correctly")
		}
	})

	t.Run("empty key rejected", func(t *testing.T) {
		_, err := NewOpenAIProvider("")
		if err == nil {
			t.Error("expected error for empty key")
		}
	})
}

func TestOllamaProvider_Complete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ollamaRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Stream {
			t.Error("stream should be false")
		}

		resp := ollamaResponse{
			EvalCount:   20,
			PromptCount: 10,
		}
		resp.Message.Content = "ollama response"

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := &OllamaProvider{
		host:   server.URL,
		model:  "llama3.2",
		client: server.Client(),
	}

	resp, err := provider.Complete(context.Background(), CompletionRequest{
		SystemPrompt: "You are a test assistant",
		UserPrompt:   "Hello",
		MaxTokens:    100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "ollama response" {
		t.Errorf("unexpected content: %s", resp.Content)
	}
	if resp.TokensUsed != 30 {
		t.Errorf("unexpected tokens used: %d", resp.TokensUsed)
	}
}

func TestRetryOn429(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"message":"rate limited"}}`))
			return
		}

		resp := ollamaResponse{
			EvalCount:   5,
			PromptCount: 5,
		}
		resp.Message.Content = "success after retry"

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := &OllamaProvider{
		host:   server.URL,
		model:  "test",
		client: server.Client(),
	}

	resp, err := provider.Complete(context.Background(), CompletionRequest{
		UserPrompt: "test",
	})
	if err != nil {
		t.Fatalf("unexpected error after retries: %v", err)
	}
	if resp.Content != "success after retry" {
		t.Errorf("unexpected content: %s", resp.Content)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}
