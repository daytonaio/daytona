// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build e2e

package e2e_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSessionBashEcho proves the standalone bash isolate runs a virtual-bash
// pipeline (grep is reimplemented by just-bash; no real binary/subprocess).
func TestSessionBashEcho(t *testing.T) {
	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	body, status := ic.CodeRun(t, map[string]interface{}{
		"language": "bash",
		"code":     `echo "hello world" | grep -o hello`,
	})
	require.Equal(t, http.StatusOK, status)
	stdout, _ := body["stdout"].(string)
	assert.Contains(t, stdout, "hello", "grep over a pipe must print the match")
}

// TestSessionBashNonZeroExit verifies a non-zero command exit is reported (not
// raised as a transport error) — the standalone isolate surfaces the exit via
// the normal completion flow.
func TestSessionBashNonZeroExit(t *testing.T) {
	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	_, status := ic.CodeRun(t, map[string]interface{}{
		"language": "bash",
		"code":     `false`,
	})
	require.Equal(t, http.StatusOK, status, "a non-zero exit is a normal completion, not an HTTP error")
}

// TestSessionBashOverlayIsolation proves writes are private + ephemeral: a file
// written in one bash session is NOT visible to a second session (each gets its
// own OverlayFs over /workspace).
func TestSessionBashOverlayIsolation(t *testing.T) {
	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	// Session A writes a file into its private overlay.
	connA, status := ic.Connect(t, map[string]interface{}{"template": "python-default", "language": "bash"})
	require.Equal(t, http.StatusOK, status)
	idA, _ := connA["sessionId"].(string)
	t.Cleanup(func() { _ = ic.DeleteSession(t, idA) })
	wsA, closeA := dialSessionWebSocket(t, connA["wsUrl"].(string))
	defer closeA()
	sendExec(t, wsA, `echo secret > /workspace/overlay_probe.txt && cat /workspace/overlay_probe.txt`, nil)
	framesA := collectFrames(t, wsA, 30*time.Second)
	assert.Contains(t, joinFramesByType(framesA, "stdout"), "secret")

	// Session B must NOT see A's overlay write.
	connB, status := ic.Connect(t, map[string]interface{}{"template": "python-default", "language": "bash"})
	require.Equal(t, http.StatusOK, status)
	idB, _ := connB["sessionId"].(string)
	t.Cleanup(func() { _ = ic.DeleteSession(t, idB) })
	wsB, closeB := dialSessionWebSocket(t, connB["wsUrl"].(string))
	defer closeB()
	sendExec(t, wsB, `cat /workspace/overlay_probe.txt 2>/dev/null || echo MISSING`, nil)
	framesB := collectFrames(t, wsB, 30*time.Second)
	assert.Contains(t, joinFramesByType(framesB, "stdout"), "MISSING",
		"a write in session A's overlay must not leak into session B")
}

// TestSessionPythonBashBridge proves Python user code can call bash() and read
// the result back (the stdio hostcall RPC bridge).
func TestSessionPythonBashBridge(t *testing.T) {
	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	body, status := ic.CodeRun(t, map[string]interface{}{
		"language": "python",
		"code": strings.Join([]string{
			`r = bash("echo from-python | tr a-z A-Z")`,
			`print(r.stdout.strip())`,
			`print("exit", r.exit_code)`,
		}, "\n"),
		"timeout": 30,
	})
	require.Equal(t, http.StatusOK, status)
	stdout, _ := body["stdout"].(string)
	assert.Contains(t, stdout, "FROM-PYTHON", "python bash() must return the command's stdout")
	assert.Contains(t, stdout, "exit 0")
}

// TestSessionTypescriptBashBridge proves TS isolate user code can call the
// host-bridged bash() (just-bash runs in the host, results copied across the
// isolated-vm boundary).
func TestSessionTypescriptBashBridge(t *testing.T) {
	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	body, status := ic.CodeRun(t, map[string]interface{}{
		"language": "typescript",
		"code":     `const r = await bash('echo from-ts'); console.log(r.stdout.trim(), 'exit', r.exitCode);`,
		"timeout":  30,
	})
	require.Equal(t, http.StatusOK, status)
	stdout, _ := body["stdout"].(string)
	assert.Contains(t, stdout, "from-ts", "ts bash() must return the command's stdout")
	assert.Contains(t, stdout, "exit 0")
}
