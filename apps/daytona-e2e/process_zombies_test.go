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

// TestProcessZombieCleanup validates the daemon's zombie/orphan handling:
//   - the PID-1 reaper collects children reparented to the daemon
//   - deleting a session reaps the session shell (no zombie zsh left behind)
//   - deleting a session also kills the session shell's process group, so
//     long-running children spawned inside the session don't survive
//
// The tests work over plain HTTP (no PTY/WebSocket) and only require the
// toolbox proxy URL exposed by the API after the sandbox reaches `started`.
func TestProcessZombieCleanup(t *testing.T) {
	cfg := LoadConfig(t)
	client := NewAPIClient(cfg)
	runID := testRunID()

	createReq := map[string]interface{}{
		"name":   fmt.Sprintf("e2e-zombie-%s", runID[4:]),
		"labels": sandboxLabels(runID),
	}
	if cfg.Snapshot != "" {
		createReq["snapshot"] = cfg.Snapshot
	}

	sandbox := client.CreateSandbox(t, createReq)
	sandboxID, _ := sandbox["id"].(string)
	require.NotEmpty(t, sandboxID, "sandbox must have id")

	started := client.PollSandboxState(t, sandboxID, "started", cfg.PollTimeout, cfg.PollInterval)

	toolboxProxyURL, _ := started["toolboxProxyUrl"].(string)
	if toolboxProxyURL == "" {
		t.Skip("toolboxProxyUrl not available — skipping zombie tests")
	}

	baseURL := strings.TrimRight(toolboxProxyURL, "/") + "/" + sandboxID
	httpCli := &http.Client{Timeout: 30 * time.Second}

	t.Logf("sandbox %s started, toolboxProxyUrl: %s", sandboxID, toolboxProxyURL)

	// --- helpers --------------------------------------------------------

	// execCommand runs a shell command via /process/execute and returns
	// (exitCode, result-stdout-and-stderr-combined).
	execCommand := func(t *testing.T, command string) (int, string) {
		t.Helper()
		body, err := json.Marshal(map[string]interface{}{
			"command": command,
			"timeout": 15,
		})
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, baseURL+"/process/execute", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpCli.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode, "execute failed: %s", string(respBody))

		var r struct {
			ExitCode int    `json:"exitCode"`
			Result   string `json:"result"`
		}
		require.NoError(t, json.Unmarshal(respBody, &r))
		return r.ExitCode, r.Result
	}

	// listZombies returns the raw ps lines whose STAT contains 'Z'. Empty
	// string means no zombies are present in the sandbox.
	listZombies := func(t *testing.T) string {
		t.Helper()
		// awk filters by stat column containing 'Z'. `|| true` keeps the
		// pipeline exit code clean when there are no matches.
		_, out := execCommand(t, `ps -eo pid,ppid,stat,comm,args | awk 'NR==1 || $3 ~ /Z/' ; true`)
		// Trim trailing whitespace for cleaner assertions.
		return strings.TrimSpace(out)
	}

	// assertNoZombies polls listZombies for up to ~5s before asserting,
	// since the PID-1 reaper runs asynchronously and zombie collection is
	// best-effort-within-a-tick rather than synchronous.
	assertNoZombies := func(t *testing.T, msg string) {
		t.Helper()
		var last string
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			last = listZombies(t)
			if !strings.Contains(last, "<defunct>") {
				t.Logf("ps zombie scan (clean):\n%s", last)
				return
			}
			time.Sleep(250 * time.Millisecond)
		}
		t.Logf("ps zombie scan (FAIL):\n%s", last)
		assert.NotContains(t, last, "<defunct>", msg)
	}

	createSession := func(t *testing.T, sessionID string) {
		t.Helper()
		body, err := json.Marshal(map[string]interface{}{"sessionId": sessionID})
		require.NoError(t, err)
		req, err := http.NewRequest(http.MethodPost, baseURL+"/process/session", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
		req.Header.Set("Content-Type", "application/json")
		resp, err := httpCli.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		require.Equal(t, http.StatusCreated, resp.StatusCode, "create session failed: %s", string(respBody))
	}

	execSessionCmd := func(t *testing.T, sessionID, command string, async bool) {
		t.Helper()
		body, err := json.Marshal(map[string]interface{}{
			"command":  command,
			"runAsync": async,
		})
		require.NoError(t, err)
		req, err := http.NewRequest(http.MethodPost, baseURL+"/process/session/"+sessionID+"/exec", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
		req.Header.Set("Content-Type", "application/json")
		resp, err := httpCli.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		if async {
			require.Equal(t, http.StatusAccepted, resp.StatusCode, "session exec (async) failed: %s", string(respBody))
		} else {
			require.Equal(t, http.StatusOK, resp.StatusCode, "session exec failed: %s", string(respBody))
		}
	}

	deleteSession := func(t *testing.T, sessionID string) {
		t.Helper()
		req, err := http.NewRequest(http.MethodDelete, baseURL+"/process/session/"+sessionID, nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
		resp, err := httpCli.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		require.Equal(t, http.StatusNoContent, resp.StatusCode, "delete session failed: %s", string(respBody))
	}

	// --- subtests -------------------------------------------------------

	t.Run("OrphanedChildReaped", func(t *testing.T) {
		// Spawn a short-lived backgrounded shell via /process/execute. The
		// outer shell that /process/execute spawned exits immediately after
		// the &, so the inner sh is reparented to PID 1 (the daemon). When
		// it finishes its echo and exits, only the daemon's SIGCHLD reaper
		// can collect it. If the reaper is missing or broken, this leaves a
		// <defunct> entry.
		code, _ := execCommand(t, `nohup sh -c 'echo e2e-orphan-child; exit 0' >/dev/null 2>&1 &`)
		require.Zero(t, code, "execute must accept the background launch")

		assertNoZombies(t, "PID-1 reaper must collect orphaned children")
	})

	t.Run("SessionShellReapedOnDelete", func(t *testing.T) {
		sessionID := fmt.Sprintf("e2e-zombie-sess-%s", runID[4:])
		createSession(t, sessionID)
		// Run something trivial so the session has spawned at least one
		// child wrapper before we tear it down.
		execSessionCmd(t, sessionID, "echo hello-session", false)
		deleteSession(t, sessionID)

		// After delete: the session's shell was Process.Kill()ed AND then
		// cmd.Wait()ed by the daemon. No zombie zsh/bash should remain.
		assertNoZombies(t, "session shell must be reaped after Delete")
	})

	t.Run("SessionChildrenKilledOnDelete", func(t *testing.T) {
		sessionID := fmt.Sprintf("e2e-zombie-child-%s", runID[4:])
		createSession(t, sessionID)

		// Pick an unusual sleep duration so we can fingerprint the process
		// in `ps` output without false positives from anything else in the
		// sandbox.
		const sentinel = "7654321"

		// runAsync spawns a long-running child of the session shell. The
		// child is NOT nohup'd or disown'd — it lives in the session
		// shell's process group, which is exactly the case that the
		// Setpgid-on-create + kill(-pgid)-on-delete change is meant to
		// clean up atomically.
		execSessionCmd(t, sessionID, fmt.Sprintf("sleep %s & echo spawned", sentinel), true)

		// Give the runAsync wrapper time to actually start the sleep child
		// before we delete the session.
		time.Sleep(1 * time.Second)

		// Sanity check: confirm the sleep is running before delete.
		_, before := execCommand(t, fmt.Sprintf("ps -eo pid,ppid,stat,comm,args | grep 'sleep %s' | grep -v grep ; true", sentinel))
		require.Contains(t, before, "sleep "+sentinel, "sentinel sleep must be running before session delete; ps output: %s", before)

		deleteSession(t, sessionID)

		// After delete: kill(-pgid, SIGKILL) plus the descendant walk
		// should have taken the sleep down. Poll briefly to allow signal
		// delivery + reaping to flush through.
		var after string
		deadline := time.Now().Add(5 * time.Second)
		gone := false
		for time.Now().Before(deadline) {
			_, after = execCommand(t, fmt.Sprintf("ps -eo pid,ppid,stat,comm,args | grep 'sleep %s' | grep -v grep ; true", sentinel))
			if !strings.Contains(after, "sleep "+sentinel) {
				gone = true
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		assert.True(t, gone, "sentinel sleep must be killed when its session is deleted; final ps output: %s", after)

		// And no zombies should be left behind by the teardown either.
		assertNoZombies(t, "session teardown must not leave zombies")
	})
}
