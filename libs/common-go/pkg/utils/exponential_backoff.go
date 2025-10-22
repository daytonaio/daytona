// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

const (
	// Default retry configuration for Docker operations
	DEFAULT_MAX_RETRIES int           = 5
	DEFAULT_BASE_DELAY  time.Duration = 100 * time.Millisecond
	DEFAULT_MAX_DELAY   time.Duration = 5 * time.Second
)

// RetryWithExponentialBackoff executes a function with exponential backoff retry logic
func RetryWithExponentialBackoff(ctx context.Context, operationName string, maxRetries int, baseDelay, maxDelay time.Duration, operationFunc func() error) error {
	if maxRetries <= 1 {
		slog.DebugContext(ctx, "Invalid max retries value, using default", "value", maxRetries, "default", DEFAULT_MAX_RETRIES)
		maxRetries = DEFAULT_MAX_RETRIES
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		logAttempt := attempt + 1
		slog.DebugContext(ctx, "Executing operation",
			"operation", operationName,
			"attempt", logAttempt,
			"max_retries", maxRetries)

		err := operationFunc()
		if err == nil {
			return nil
		}

		if attempt < maxRetries-1 {
			// Calculate exponential backoff delay
			delay := min(baseDelay*time.Duration(1<<attempt), maxDelay)

			slog.WarnContext(ctx, "Operation failed, retrying",
				"operation", operationName,
				"attempt", logAttempt,
				"max_retries", maxRetries,
				"error", err,
				"retry_delay", delay)

			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		return fmt.Errorf("failed to %s after %d attempts: %w", operationName, maxRetries, err)
	}

	return nil
}
