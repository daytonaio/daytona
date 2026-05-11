// Copyright Daytona Platforms Inc.
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
//
// Also cleans up stranded session-instance sandboxes labeled `daytona.io/session=true`
// that were created during session e2e runs (test-run=* label).
func TestCleanupStaleSandboxes(t *testing.T) {
	cfg := LoadConfig(t)
	client := NewAPIClient(cfg)

	cleanedCount := 0
	cleanedCount += cleanupStaleByLabel(t, client, `{"e2e":"true"}`)
	cleanedCount += cleanupStaleByLabel(t, client, `{"daytona.io/session":"true","e2e":"true"}`)

	t.Logf("cleaned up %d stale E2E sandboxes total", cleanedCount)
}

func cleanupStaleByLabel(t *testing.T, client *APIClient, labels string) int {
	t.Helper()
	path := "/sandbox/paginated?labels=" + labels

	resp, body := client.DoRequest(t, http.MethodGet, path, nil)
	if resp.StatusCode != http.StatusOK {
		t.Logf("failed to list E2E sandboxes (status %d): %s", resp.StatusCode, string(body))
		return 0
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Logf("failed to parse sandbox list: %v", err)
		return 0
	}

	items, _ := result["items"].([]interface{})
	staleCutoff := time.Now().Add(-30 * time.Minute)
	cleaned := 0

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

		t.Logf("cleaning up stale sandbox %s (created %s, state %s, labels=%s)", sandboxID, createdAtStr, state, labels)
		client.DeleteSandbox(t, sandboxID)
		cleaned++
	}

	return cleaned
}
