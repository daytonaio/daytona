// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build e2e

package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// linkedExecResult is the parsed shape of a toolbox /process/execute response
// (only the fields these tests rely on).
type linkedExecResult struct {
	exitCode int
	result   string
}

// execInSandbox runs `command` inside the sandbox via the toolbox proxy and
// returns the exit code + combined stdout/stderr. `commandTimeout` is the
// per-process timeout the toolbox enforces server-side.
func execInSandbox(t *testing.T, cfg Config, toolboxBaseURL, command string, commandTimeout int) linkedExecResult {
	t.Helper()

	body, err := json.Marshal(map[string]interface{}{
		"command": command,
		"timeout": commandTimeout,
	})
	require.NoError(t, err, "marshal exec body")

	req, err := http.NewRequest(http.MethodPost, toolboxBaseURL+"/process/execute", bytes.NewReader(body))
	require.NoError(t, err, "new exec request")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	// Allow the HTTP round-trip a bit more time than the in-sandbox command
	// timeout so we observe a clean timeout result instead of a transport error.
	httpCli := &http.Client{Timeout: time.Duration(commandTimeout+15) * time.Second}
	resp, err := httpCli.Do(req)
	require.NoError(t, err, "exec request failed")
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "read exec response")
	require.Equal(t, http.StatusOK, resp.StatusCode, "process/execute must return 200: %s", string(respBody))

	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(respBody, &raw), "parse exec response: %s", string(respBody))

	exitCode := 0
	if v, ok := raw["exitCode"].(float64); ok {
		exitCode = int(v)
	}
	result, _ := raw["result"].(string)
	return linkedExecResult{exitCode: exitCode, result: result}
}

// createLinkedSandboxPair creates an owner sandbox, waits for it to reach
// "started", then creates a follower linked to it via `linkedSandbox` and
// waits for *that* to start. Returns both started sandbox JSON bodies.
// Each sandbox auto-registers t.Cleanup via CreateSandbox.
func createLinkedSandboxPair(t *testing.T, client *APIClient, cfg Config, runID, ownerName, followerName string) (owner, follower map[string]interface{}) {
	t.Helper()

	ownerReq := map[string]interface{}{
		"name":   ownerName,
		"labels": sandboxLabels(runID),
	}
	if cfg.Snapshot != "" {
		ownerReq["snapshot"] = cfg.Snapshot
	}
	ownerCreate := client.CreateSandbox(t, ownerReq)
	ownerID, _ := ownerCreate["id"].(string)
	require.NotEmpty(t, ownerID, "owner sandbox must have id")

	owner = client.PollSandboxState(t, ownerID, "started", cfg.PollTimeout, cfg.PollInterval)

	followerReq := map[string]interface{}{
		"name":               followerName,
		"labels":             sandboxLabels(runID),
		"linkedSandbox":      ownerID,
		"autoDeleteInterval": 0,
	}
	if cfg.Snapshot != "" {
		followerReq["snapshot"] = cfg.Snapshot
	}
	followerCreate := client.CreateSandbox(t, followerReq)
	followerID, _ := followerCreate["id"].(string)
	require.NotEmpty(t, followerID, "follower sandbox must have id")

	follower = client.PollSandboxState(t, followerID, "started", cfg.PollTimeout, cfg.PollInterval)
	return owner, follower
}

// TestLinkedSandboxConnectivity verifies that a sandbox created with
// linkedSandbox set is co-located on the same runner as its owner and can
// reach the owner over the runner-local link network by sandbox name.
func TestLinkedSandboxConnectivity(t *testing.T) {
	cfg := LoadConfig(t)
	client := NewAPIClient(cfg)
	runID := testRunID()

	ownerName := fmt.Sprintf("e2e-linkown-%s", runID[4:])
	followerName := fmt.Sprintf("e2e-linkflw-%s", runID[4:])

	owner, follower := createLinkedSandboxPair(t, client, cfg, runID, ownerName, followerName)

	ownerID, _ := owner["id"].(string)
	ownerRunnerID, _ := owner["runnerId"].(string)
	ownerToolboxURL, _ := owner["toolboxProxyUrl"].(string)
	require.NotEmpty(t, ownerToolboxURL, "owner toolboxProxyUrl must be set when started")

	followerID, _ := follower["id"].(string)
	followerToolboxURL, _ := follower["toolboxProxyUrl"].(string)
	require.NotEmpty(t, followerToolboxURL, "follower toolboxProxyUrl must be set when started")

	ownerBaseURL := strings.TrimRight(ownerToolboxURL, "/") + "/" + ownerID
	followerBaseURL := strings.TrimRight(followerToolboxURL, "/") + "/" + followerID

	t.Run("ResponseFields", func(t *testing.T) {
		gotLinkedID, _ := follower["linkedSandboxId"].(string)
		assert.Equal(t, ownerID, gotLinkedID, "follower.linkedSandboxId must equal owner.id")

		gotRunnerID, _ := follower["runnerId"].(string)
		assert.NotEmpty(t, ownerRunnerID, "owner must have runnerId")
		assert.Equal(t, ownerRunnerID, gotRunnerID, "follower must be co-located on same runner as owner")

		// autoDeleteInterval is enforced to 0 for linked sandboxes by the API.
		if v, ok := follower["autoDeleteInterval"].(float64); ok {
			assert.Equal(t, 0, int(v), "linked sandbox must have autoDeleteInterval=0 (ephemeral)")
		}
	})

	t.Run("FollowerCanReachOwnerByName", func(t *testing.T) {
		// Preflight: python3 must be present in the owner sandbox snapshot, as
		// the user-facing example uses `python3 -m http.server`.
		preflight := execInSandbox(t, cfg, ownerBaseURL, "command -v python3 >/dev/null && echo OK || echo MISSING", 10)
		require.Equal(t, 0, preflight.exitCode, "preflight exec failed: %s", preflight.result)
		require.Contains(t, preflight.result, "OK", "snapshot %q must include python3 for this test", cfg.Snapshot)

		// Start a python http server in the owner sandbox on port 3000, serving
		// a known marker file. Wait inside the shell until the server actually
		// binds so the follower-side curl is not racing against startup.
		const marker = "hello-from-owner"
		startServerCmd := fmt.Sprintf(`set -e
mkdir -p /tmp/lnk
echo %q > /tmp/lnk/index.html
cd /tmp/lnk
nohup python3 -m http.server 3000 > /tmp/lnk/srv.log 2>&1 &
for i in $(seq 1 20); do
  if curl -sS --max-time 1 http://127.0.0.1:3000/ >/dev/null 2>&1; then
    echo READY
    exit 0
  fi
  sleep 0.5
done
echo "server did not become ready"
cat /tmp/lnk/srv.log || true
exit 1
`, marker)

		startRes := execInSandbox(t, cfg, ownerBaseURL, startServerCmd, 30)
		require.Equal(t, 0, startRes.exitCode, "owner failed to start http server: %s", startRes.result)
		require.Contains(t, startRes.result, "READY", "owner http server did not become ready: %s", startRes.result)

		// From the follower, curl the owner by its sandbox name. The link
		// network attaches the owner with an alias of its sandbox name, so the
		// follower's DNS resolver should pick that up.
		curlCmd := fmt.Sprintf(
			`curl -sS --max-time 5 -o /tmp/body -w '%%{http_code}' http://%s:3000/ && echo "|" && cat /tmp/body`,
			ownerName,
		)

		var lastResult string
		var lastExit int
		// Tolerate any tail of the runner-side network-attach reconciliation:
		// the follower is marked "started" once its container is running, but
		// the link-network NetworkConnect call can race in rare cases.
		deadline := time.Now().Add(20 * time.Second)
		for time.Now().Before(deadline) {
			res := execInSandbox(t, cfg, followerBaseURL, curlCmd, 15)
			lastExit, lastResult = res.exitCode, res.result
			if res.exitCode == 0 && strings.Contains(res.result, "200") && strings.Contains(res.result, marker) {
				t.Logf("follower successfully reached owner %q via link network: %s", ownerName, strings.TrimSpace(res.result))
				return
			}
			t.Logf("follower curl not yet successful (exit=%d): %s", res.exitCode, strings.TrimSpace(res.result))
			time.Sleep(2 * time.Second)
		}

		t.Fatalf("follower could not reach owner %q on port 3000 via link network: exitCode=%d output=%q",
			ownerName, lastExit, lastResult)
	})
}

// TestLinkedSandboxCascadeDestroy verifies that destroying the owner sandbox
// automatically destroys all linked followers (the ephemeral semantic).
func TestLinkedSandboxCascadeDestroy(t *testing.T) {
	cfg := LoadConfig(t)
	client := NewAPIClient(cfg)
	runID := testRunID()

	ownerName := fmt.Sprintf("e2e-lnkcas-own-%s", runID[4:])
	followerName := fmt.Sprintf("e2e-lnkcas-flw-%s", runID[4:])

	owner, follower := createLinkedSandboxPair(t, client, cfg, runID, ownerName, followerName)

	ownerID, _ := owner["id"].(string)
	followerID, _ := follower["id"].(string)
	require.NotEmpty(t, ownerID)
	require.NotEmpty(t, followerID)

	// Sanity: confirm the link is recorded before we trigger cascade.
	gotLinkedID, _ := follower["linkedSandboxId"].(string)
	require.Equal(t, ownerID, gotLinkedID, "precondition: follower.linkedSandboxId must equal owner.id")

	// Trigger the cascade by destroying the owner.
	resp, body := client.DoRequest(t, http.MethodDelete, "/sandbox/"+ownerID, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "DELETE owner failed: %s", string(body))
	t.Logf("requested destroy of owner %s; waiting for cascade to take down follower %s", ownerID, followerID)

	// Poll the follower until either it's gone (404) or its desiredState is
	// "destroyed" (the SandboxEvents.DESTROYED handler updates this before the
	// runner finishes physical teardown).
	deadline := time.Now().Add(90 * time.Second)
	interval := 2 * time.Second

	for time.Now().Before(deadline) {
		sandbox, statusCode := client.GetSandbox(t, followerID)
		if statusCode == http.StatusNotFound {
			t.Logf("follower %s removed (404) after owner destroy", followerID)
			return
		}
		desiredState, _ := sandbox["desiredState"].(string)
		state, _ := sandbox["state"].(string)
		t.Logf("waiting for cascade: follower state=%s desiredState=%s", state, desiredState)
		if desiredState == "destroyed" {
			t.Logf("follower %s marked desiredState=destroyed after owner destroy", followerID)
			return
		}
		time.Sleep(interval)
	}

	t.Fatalf("follower %s was not cascade-destroyed within 90s after owner %s deletion", followerID, ownerID)
}
