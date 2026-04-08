// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build e2e

package e2e_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSandboxLifecycle validates the full sandbox lifecycle:
// create → poll STARTED → validate fields → delete → verify cleanup.
func TestSandboxLifecycle(t *testing.T) {
	cfg := LoadConfig(t)
	client := NewAPIClient(cfg)
	runID := testRunID()

	var sandboxID string
	var toolboxProxyURL string
	overallStart := time.Now()

	t.Run("Create", func(t *testing.T) {
		createStart := time.Now()

		createReq := map[string]interface{}{
			"name":   fmt.Sprintf("e2e-lifecycle-%s", runID[4:]),
			"labels": sandboxLabels(runID),
		}
		if cfg.Snapshot != "" {
			createReq["snapshot"] = cfg.Snapshot
		}

		resp, body := client.DoRequest(t, http.MethodPost, "/sandbox", createReq)
		require.Equal(t, http.StatusOK, resp.StatusCode, "create sandbox failed: %s", string(body))

		var sandbox map[string]interface{}
		require.NoError(t, json.Unmarshal(body, &sandbox), "failed to parse sandbox response")

		id, ok := sandbox["id"].(string)
		require.True(t, ok && id != "", "sandbox response must have non-empty id")
		sandboxID = id

		state, _ := sandbox["state"].(string)
		assert.NotEmpty(t, state, "sandbox must have state field")

		desiredState, _ := sandbox["desiredState"].(string)
		assert.Equal(t, "started", desiredState, "desiredState must be 'started' after creation")

		t.Logf("sandbox %s created (API response) in %s", sandboxID, time.Since(createStart))
	})

	require.NotEmpty(t, sandboxID, "sandboxID must be set by Create subtest")

	// Register cleanup on the PARENT test so it runs after ALL subtests complete
	t.Cleanup(func() {
		client.DeleteSandbox(t, sandboxID)
	})

	t.Run("PollUntilStarted", func(t *testing.T) {
		pollStart := time.Now()

		sandbox := client.PollSandboxState(t, sandboxID, "started", cfg.PollTimeout, cfg.PollInterval)

		state, _ := sandbox["state"].(string)
		assert.Equal(t, "started", state, "sandbox must be in started state")

		proxyURL, _ := sandbox["toolboxProxyUrl"].(string)
		assert.NotEmpty(t, proxyURL, "toolboxProxyUrl must be non-empty when sandbox is started")
		toolboxProxyURL = proxyURL

		t.Logf("sandbox %s reached STARTED in %s (total from create: %s)",
			sandboxID, time.Since(pollStart), time.Since(overallStart))
		t.Logf("toolboxProxyUrl: %s", toolboxProxyURL)
	})

	t.Run("ValidateFields", func(t *testing.T) {
		sandbox, statusCode := client.GetSandbox(t, sandboxID)
		require.Equal(t, http.StatusOK, statusCode, "GET /sandbox/{id} must return 200")
		require.NotNil(t, sandbox, "sandbox response must not be nil")

		name, _ := sandbox["name"].(string)
		assert.Contains(t, name, "e2e-lifecycle", "sandbox name must contain e2e-lifecycle prefix")

		labels, _ := sandbox["labels"].(map[string]interface{})
		assert.Equal(t, "true", labels["e2e"], "sandbox must have e2e=true label")
		assert.Equal(t, runID, labels["test-run"], "sandbox must have test-run label matching runID")

		cpu, _ := sandbox["cpu"].(float64)
		assert.Greater(t, cpu, float64(0), "sandbox cpu must be positive")

		memory, _ := sandbox["memory"].(float64)
		assert.Greater(t, memory, float64(0), "sandbox memory must be positive")

		disk, _ := sandbox["disk"].(float64)
		assert.Greater(t, disk, float64(0), "sandbox disk must be positive")

		createdAt, _ := sandbox["createdAt"].(string)
		assert.NotEmpty(t, createdAt, "sandbox createdAt must be present")

		t.Logf("validated sandbox %s: cpu=%.0f memory=%.0fGB disk=%.0fGB createdAt=%s",
			sandboxID, cpu, memory, disk, createdAt)
	})

	var deleteStart time.Time

	t.Run("Delete", func(t *testing.T) {
		deleteStart = time.Now()
		resp, body := client.DoRequest(t, http.MethodDelete, "/sandbox/"+sandboxID, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode, "DELETE /sandbox/{id} must return 200: %s", string(body))
		t.Logf("delete request for sandbox %s returned 200", sandboxID)
	})

	t.Run("VerifyCleanup", func(t *testing.T) {
		cleanupTimeout := 60 * time.Second
		deadline := time.Now().Add(cleanupTimeout)
		interval := 2 * time.Second

		for time.Now().Before(deadline) {
			sandbox, statusCode := client.GetSandbox(t, sandboxID)

			if statusCode == http.StatusNotFound {
				t.Logf("sandbox %s fully removed (404) in %s", sandboxID, time.Since(deleteStart))
				return
			}

			if sandbox != nil {
				state, _ := sandbox["state"].(string)
				t.Logf("waiting for cleanup: sandbox %s state=%s", sandboxID, state)
			}

			time.Sleep(interval)
		}

		t.Fatalf("sandbox %s was not fully removed within 60s", sandboxID)
	})
}
