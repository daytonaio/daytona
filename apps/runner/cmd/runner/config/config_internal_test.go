// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package config

import "testing"

func TestGetDindRegistryMirror(t *testing.T) {
	orig := config
	t.Cleanup(func() { config = orig })

	// Nil config must not panic and returns an empty string (feature disabled).
	config = nil
	if got := GetDindRegistryMirror(); got != "" {
		t.Fatalf("expected empty mirror for nil config, got %q", got)
	}

	config = &Config{DindRegistryMirror: "http://registry-mirror:5001"}
	if got := GetDindRegistryMirror(); got != "http://registry-mirror:5001" {
		t.Fatalf("unexpected mirror: %q", got)
	}
}
