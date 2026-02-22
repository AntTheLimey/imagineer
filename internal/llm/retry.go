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
	"math"
	"strings"
	"time"
)

const maxRetries = 3

// doWithRetry calls fn up to maxRetries times with exponential backoff
// when the HTTP status code is 429 (rate limited) or 503 (service
// unavailable). Quota errors (402 or 429-with-quota-body) fail
// immediately without retrying. The fn must return
// (response, httpStatusCode, error).
func doWithRetry(
	ctx context.Context,
	fn func(ctx context.Context) (CompletionResponse, int, error),
) (CompletionResponse, error) {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, statusCode, err := fn(ctx)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Quota errors fail immediately.
		if isQuotaError(statusCode, err) {
			return CompletionResponse{},
				&QuotaExceededError{
					Provider: "llm",
					Message:  err.Error(),
				}
		}

		// Only retry on 429 (rate limited) or 503 (service unavailable)
		if statusCode != 429 && statusCode != 503 {
			return CompletionResponse{}, err
		}

		if attempt < maxRetries {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			select {
			case <-ctx.Done():
				return CompletionResponse{}, ctx.Err()
			case <-time.After(backoff):
			}
		}
	}
	return CompletionResponse{}, lastErr
}

// isQuotaError checks whether the HTTP status code
// and error message indicate an API quota exhaustion
// rather than a temporary rate limit.
func isQuotaError(statusCode int, err error) bool {
	// Anthropic returns 402 for quota exceeded.
	if statusCode == 402 {
		return true
	}
	// OpenAI returns 429 with "exceeded your current
	// quota" in the message body.
	if statusCode == 429 && err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "quota") ||
			strings.Contains(msg, "billing") {
			return true
		}
	}
	return false
}
