// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build e2e

package e2e_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

// TestCleanupStaleSandboxes finds and deletes E2E sandboxes older than 30 minutes.
// This is a safety net for orphaned sandboxes from failed test runs.
// Passes even when no stale sandboxes are found.
func TestCleanupStaleSandboxes(t *testing.T) {
	cfg := LoadConfig(t)
	client := NewAPIClient(cfg)

	labels := `{"e2e":"true"}`
	path := "/sandbox/paginated?labels=" + labels

	resp, body := client.DoRequest(t, http.MethodGet, path, nil)
	if resp.StatusCode != http.StatusOK {
		t.Logf("failed to list E2E sandboxes (status %d): %s", resp.StatusCode, string(body))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Logf("failed to parse sandbox list: %v", err)
		return
	}

	items, _ := result["items"].([]interface{})
	staleCutoff := time.Now().Add(-30 * time.Minute)
	cleanedCount := 0

	for _, item := range items {
		sandbox, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		sandboxID, _ := sandbox["id"].(string)
		if sandboxID == "" {
			continue
		}

		createdAtStr, _ := sandbox["createdAt"].(string)
		state, _ := sandbox["state"].(string)

		createdAt, err := time.Parse(time.RFC3339Nano, createdAtStr)
		if err != nil {
			t.Logf("skipping sandbox %s: cannot parse createdAt %q: %v", sandboxID, createdAtStr, err)
			continue
		}

		if createdAt.After(staleCutoff) {
			continue
		}

		t.Logf("cleaning up stale sandbox %s (created %s, state %s)", sandboxID, createdAtStr, state)
		client.DeleteSandbox(t, sandboxID)
		cleanedCount++
	}

	t.Logf("cleaned up %d stale E2E sandboxes", cleanedCount)
}
