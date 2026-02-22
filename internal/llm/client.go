/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

// Package llm provides a unified abstraction for LLM service providers
// (Anthropic, OpenAI, Ollama) used by the enrichment engine.
package llm

import (
	"context"
	"fmt"

	"github.com/antonypegg/imagineer/internal/models"
)

// Provider defines the interface for LLM service providers.
type Provider interface {
	Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
}

// CompletionRequest holds the parameters for an LLM completion call.
type CompletionRequest struct {
	SystemPrompt string
	UserPrompt   string
	MaxTokens    int
	Temperature  float64
}

// CompletionResponse holds the result of an LLM completion call.
type CompletionResponse struct {
	Content    string
	TokensUsed int
}

// QuotaExceededError indicates the LLM provider has
// rejected the request due to an exhausted API quota
// (as opposed to a temporary rate limit).
type QuotaExceededError struct {
	Provider string
	Message  string
}

func (e *QuotaExceededError) Error() string {
	return fmt.Sprintf(
		"%s quota exceeded: %s",
		e.Provider, e.Message,
	)
}

// NewProvider creates an LLM provider based on the service type and API key.
func NewProvider(service models.LLMService, apiKey string) (Provider, error) {
	switch service {
	case models.LLMServiceAnthropic:
		return NewAnthropicProvider(apiKey)
	case models.LLMServiceOpenAI:
		return NewOpenAIProvider(apiKey)
	case models.LLMServiceOllama:
		return NewOllamaProvider()
	default:
		return nil, fmt.Errorf("unsupported LLM service: %s", service)
	}
}
