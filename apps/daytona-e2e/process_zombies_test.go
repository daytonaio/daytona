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
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProcessZombieCleanup validates the daemon's zombie/orphan handling:
//   - the PID-1 reaper collects children reparented to the daemon
//   - deleting a session reaps the session shell (no zombie zsh left behind)
//   - deleting a session also kills the session shell's process group, so
//     long-running children spawned inside the session don't survive
//   - deleting a PTY session walks the shell's process tree BEFORE the
//     shell exits, catching disowned descendants that would otherwise be
//     reparented to PID 1 and survive teardown
//
// Most subtests are pure HTTP. The PTY subtest additionally opens a
// WebSocket through the toolbox proxy to drive the shell.
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

	// --- PTY helpers (only used by the PTY subtest below) ---------------

	createPTY := func(t *testing.T, ptyID string) {
		t.Helper()
		body, err := json.Marshal(map[string]interface{}{
			"id":   ptyID,
			"cols": 120,
			"rows": 30,
		})
		require.NoError(t, err)
		req, err := http.NewRequest(http.MethodPost, baseURL+"/process/pty", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
		req.Header.Set("Content-Type", "application/json")
		resp, err := httpCli.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		require.Equal(t, http.StatusCreated, resp.StatusCode, "create pty failed: %s", string(respBody))
	}

	deletePTY := func(t *testing.T, ptyID string) {
		t.Helper()
		req, err := http.NewRequest(http.MethodDelete, baseURL+"/process/pty/"+ptyID, nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
		resp, err := httpCli.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		require.Equal(t, http.StatusOK, resp.StatusCode, "delete pty failed: %s", string(respBody))
	}

	// dialPTY opens a WebSocket to the PTY's input/output channel. The
	// returned connection sends client input to the shell when written to
	// (as text or binary messages) and receives the shell's output as
	// binary messages.
	dialPTY := func(t *testing.T, ptyID string) *websocket.Conn {
		t.Helper()
		parsed, err := url.Parse(baseURL + "/process/pty/" + ptyID + "/connect")
		require.NoError(t, err)
		switch parsed.Scheme {
		case "http":
			parsed.Scheme = "ws"
		case "https":
			parsed.Scheme = "wss"
		}
		hdr := http.Header{}
		hdr.Set("Authorization", "Bearer "+cfg.APIKey)
		dialer := *websocket.DefaultDialer
		dialer.HandshakeTimeout = 15 * time.Second
		conn, resp, err := dialer.Dial(parsed.String(), hdr)
		if resp != nil {
			defer resp.Body.Close()
		}
		require.NoError(t, err, "ws dial to %s failed", parsed.String())
		return conn
	}

	t.Run("PTYKillTakesDownDisownedDescendants", func(t *testing.T) {
		// Regression test for the PTY teardown order bug:
		//
		// Inside an interactive shell, `cmd & disown` puts cmd into its own
		// process group (escaping kill(-pgid)). Before the fix, kill()
		// cancelled the cmd.Context and closed the PTY master FIRST, which
		// caused the shell to exit fast — at which point the kernel
		// reparented the disowned child to PID 1, and the subsequent
		// gopsutil.Children(shell_pid) tree walk returned nothing (the
		// child's PPID is no longer shell_pid). The disowned child then
		// survived the PTY delete.
		//
		// The fix walks the process tree BEFORE cancelling/closing, while
		// the shell is still alive and the child's PPID still points at it.
		//
		// To trigger the regression we need a child whose teardown depends
		// on the PPID walk specifically — i.e. `disown` so it escapes
		// pgid-based teardown.
		ptyID := fmt.Sprintf("e2e-zombie-pty-%s", runID[4:])
		const sentinel = "8765431"
		needle := "sleep " + sentinel

		createPTY(t, ptyID)
		conn := dialPTY(t, ptyID)

		ptyClosed := false
		closeConn := func() {
			if ptyClosed {
				return
			}
			ptyClosed = true
			_ = conn.Close()
		}
		defer closeConn()

		// Best-effort guarantees that the PTY is gone even if the
		// assertion below panics or fails mid-flight, so we don't leak
		// state across test runs of the same sandbox.
		defer func() {
			if !ptyClosed {
				closeConn()
			}
			req, err := http.NewRequest(http.MethodDelete, baseURL+"/process/pty/"+ptyID, nil)
			if err == nil {
				req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
				if resp, derr := httpCli.Do(req); derr == nil {
					resp.Body.Close()
				}
			}
			// And nuke any straggler sleep with our sentinel.
			_, _ = execCommand(t, fmt.Sprintf("pkill -KILL -f '%s' 2>/dev/null || true", needle))
		}()

		// Give the shell a moment to print its first prompt before we
		// pipe input in (interactive shells can race against early input).
		time.Sleep(500 * time.Millisecond)

		err := conn.WriteMessage(
			websocket.TextMessage,
			[]byte(fmt.Sprintf("sleep %s & disown\n", sentinel)),
		)
		require.NoError(t, err, "ws write (sleep & disown) failed")

		// Wait until the sentinel sleep is actually running before we
		// trigger the teardown — otherwise we'd be racing the shell's
		// fork+exec.
		var before string
		spawnedDeadline := time.Now().Add(8 * time.Second)
		spawned := false
		for time.Now().Before(spawnedDeadline) {
			_, before = execCommand(t, procScanForCmdline(needle))
			if strings.Contains(before, needle) {
				spawned = true
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		require.True(t, spawned, "sentinel sleep never showed up in /proc; pty likely didn't accept input; /proc scan: %s", before)

		// Close the WS first so the DELETE doesn't race the read goroutine
		// inside the daemon.
		closeConn()

		deletePTY(t, ptyID)

		// After delete: killProcessTree must have walked the shell's
		// descendants and SIGKILLed the disowned sleep BEFORE the shell
		// itself was reaped. If the bug regresses, the sleep gets
		// reparented to PID 1 first and stays alive.
		var after string
		killDeadline := time.Now().Add(8 * time.Second)
		gone := false
		for time.Now().Before(killDeadline) {
			_, after = execCommand(t, procScanForCmdline(needle))
			if !strings.Contains(after, needle) {
				gone = true
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		assert.True(t, gone,
			"disowned sentinel sleep must be killed when its PTY is deleted; "+
				"if this fails, the PTY kill() ordering has regressed and "+
				"killProcessTree is being called AFTER the shell has already exited. "+
				"final /proc scan: %s", after)

		// And no zombies from the teardown.
		assertNoZombies(t, "PTY teardown must not leave zombies")
	})
}
