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
	"errors"
	"fmt"
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

func TestIsQuotaError_402(t *testing.T) {
	if !isQuotaError(402, fmt.Errorf("payment required")) {
		t.Error("expected isQuotaError to return true for status 402")
	}
}

func TestIsQuotaError_429WithQuotaMessage(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"quota keyword", fmt.Errorf("you have exceeded your current quota")},
		{"billing keyword", fmt.Errorf("rate limit exceeded for this billing period")},
		{"mixed case quota", fmt.Errorf("Quota limit reached")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !isQuotaError(429, tt.err) {
				t.Errorf("expected isQuotaError to return true for 429 with message: %s", tt.err)
			}
		})
	}
}

func TestIsQuotaError_429RateLimit(t *testing.T) {
	err := fmt.Errorf("rate limited, please slow down")
	if isQuotaError(429, err) {
		t.Error("expected isQuotaError to return false for normal 429 rate limit")
	}
}

func TestIsQuotaError_429RateLimitExceeded(t *testing.T) {
	err := fmt.Errorf("Rate limit exceeded")
	if isQuotaError(429, err) {
		t.Error("expected isQuotaError to return false for 'Rate limit exceeded' (temporary rate limit, not quota)")
	}
}

func TestIsQuotaError_429NilError(t *testing.T) {
	if isQuotaError(429, nil) {
		t.Error("expected isQuotaError to return false for 429 with nil error")
	}
}

func TestIsQuotaError_500(t *testing.T) {
	err := fmt.Errorf("internal server error")
	if isQuotaError(500, err) {
		t.Error("expected isQuotaError to return false for status 500")
	}
}

func TestDoWithRetry_QuotaError_NoRetry(t *testing.T) {
	attempts := 0
	_, err := doWithRetry(context.Background(), func(ctx context.Context) (CompletionResponse, int, error) {
		attempts++
		return CompletionResponse{}, 402, fmt.Errorf("payment required: quota exhausted")
	})

	if attempts != 1 {
		t.Errorf("expected exactly 1 attempt, got %d", attempts)
	}

	var quotaErr *QuotaExceededError
	if !errors.As(err, &quotaErr) {
		t.Fatalf("expected QuotaExceededError, got %T: %v", err, err)
	}
	if quotaErr.Provider != "llm" {
		t.Errorf("expected provider 'llm', got %q", quotaErr.Provider)
	}
}
