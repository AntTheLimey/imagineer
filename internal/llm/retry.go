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
	"time"
)

const maxRetries = 3

// doWithRetry calls fn up to maxRetries times with exponential backoff
// when the HTTP status code is 429 (rate limited) or 503 (service
// unavailable). The fn must return (response, httpStatusCode, error).
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
