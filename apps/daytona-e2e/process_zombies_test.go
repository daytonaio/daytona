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
	// 60s timeout: session delete chains terminateSession (≤5s grace) +
	// gopsutil tree walk + filesystem cleanup. Plenty of headroom for a
	// slow slim-image container while still surfacing a real hang.
	httpCli := &http.Client{Timeout: 60 * time.Second}

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

	// procScanZombies reads /proc/[pid]/stat directly and prints "PID
	// COMM" for every zombie. We avoid `ps` because the slim sandbox image
	// doesn't ship one. /proc/PID/stat format is:
	//   pid (comm) state ppid ...
	// `comm` may contain spaces or parens, so we strip everything up to
	// and including ") " before reading the state field.
	const procScanZombies = `awk 'FNR==1{` +
		`comm=$0; sub(/\).*/,"",comm); sub(/^[^(]*\(/,"",comm);` +
		`rest=$0; sub(/.*\) /,"",rest); split(rest,f," ");` +
		`split(FILENAME,p,"/");` +
		`if (f[1]=="Z") print "Z " p[3] " " comm` +
		`}' /proc/[0-9]*/stat 2>/dev/null`

	// listZombies returns one line per zombie process ("Z <pid> <comm>"),
	// or an empty string if there are none.
	listZombies := func(t *testing.T) string {
		t.Helper()
		_, out := execCommand(t, procScanZombies)
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
			if last == "" {
				t.Logf("/proc zombie scan: clean")
				return
			}
			time.Sleep(250 * time.Millisecond)
		}
		t.Logf("/proc zombie scan (FAIL):\n%s", last)
		assert.Empty(t, last, msg)
	}

	// procScanForCmdline returns a shell command that prints every
	// /proc/[pid]/cmdline whose contents include `needle`, formatted as
	// "<pid> <full cmdline>". Used in place of `ps … | grep …`.
	procScanForCmdline := func(needle string) string {
		// Single-quoted `needle` is embedded into a shell `case` glob, so
		// we reject quotes to keep the test honest about what it's
		// matching. Tests pass simple ASCII tokens.
		if strings.ContainsAny(needle, "'\"\\") {
			t.Fatalf("procScanForCmdline: unsupported needle %q", needle)
		}
		return `for c in /proc/[0-9]*/cmdline; do ` +
			`[ -r "$c" ] || continue; ` +
			`a=$(tr '\0' ' ' < "$c" 2>/dev/null); ` +
			`case "$a" in *'` + needle + `'*) ` +
			`echo "$(basename $(dirname $c)) $a";; ` +
			`esac; ` +
			`done`
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
		needle := "sleep " + sentinel
		_, before := execCommand(t, procScanForCmdline(needle))
		require.Contains(t, before, needle, "sentinel sleep must be running before session delete; /proc scan: %s", before)

		deleteSession(t, sessionID)

		// After delete: the daemon's signalProcessTree walk should have
		// taken the sleep down before the session shell was reaped. Poll
		// briefly to allow signal delivery + reaping to flush through.
		var after string
		deadline := time.Now().Add(5 * time.Second)
		gone := false
		for time.Now().Before(deadline) {
			_, after = execCommand(t, procScanForCmdline(needle))
			if !strings.Contains(after, needle) {
				gone = true
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		assert.True(t, gone, "sentinel sleep must be killed when its session is deleted; final /proc scan: %s", after)

		// And no zombies should be left behind by the teardown either.
		assertNoZombies(t, "session teardown must not leave zombies")
	})
}
