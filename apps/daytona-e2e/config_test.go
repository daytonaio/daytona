// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build e2e

package e2e_test

import (
	"os"
	"testing"
	"time"
)

// Config holds all configuration for E2E tests loaded from environment variables.
type Config struct {
	BaseURL      string
	APIKey       string
	Snapshot     string
	PollTimeout  time.Duration
	PollInterval time.Duration
}

// LoadConfig loads test configuration from environment variables.
// Calls t.Skip() if required variables are not set.
func LoadConfig(t *testing.T) Config {
	t.Helper()

	baseURL := os.Getenv("DAYTONA_API_URL")
	if baseURL == "" {
		t.Skip("DAYTONA_API_URL not set — skipping E2E test")
	}

	apiKey := os.Getenv("DAYTONA_API_KEY")
	if apiKey == "" {
		t.Skip("DAYTONA_API_KEY not set — skipping E2E test")
	}

	pollTimeout := 5 * time.Minute
	if v := os.Getenv("DAYTONA_POLL_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			t.Fatalf("invalid DAYTONA_POLL_TIMEOUT %q: %v", v, err)
		}
		pollTimeout = d
	}

	pollInterval := 2 * time.Second
	if v := os.Getenv("DAYTONA_POLL_INTERVAL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			t.Fatalf("invalid DAYTONA_POLL_INTERVAL %q: %v", v, err)
		}
		pollInterval = d
	}

	return Config{
		BaseURL:      baseURL,
		APIKey:       apiKey,
		Snapshot:     os.Getenv("DAYTONA_SNAPSHOT"),
		PollTimeout:  pollTimeout,
		PollInterval: pollInterval,
	}
}
