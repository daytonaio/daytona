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

// wsExecStdout runs exactly one execution against an existing session and
// returns the concatenated stdout. The session-daemon closes the WebSocket
// after each `completed` frame (one execution per connection), so consecutive
// executions reconnect to the SAME session via context.id — the daemon-side
// context (interpreter globals, files, bash overlay) persists between connects.
func wsExecStdout(t *testing.T, ic *SessionClient, sessionID, code string) string {
	t.Helper()
	conn, status := ic.Connect(t, map[string]interface{}{"context": map[string]interface{}{"id": sessionID}})
	require.Equal(t, http.StatusOK, status, "reconnect to existing session must return 200")
	require.Equal(t, sessionID, conn["sessionId"], "reconnect must reuse the same session id")

	ws, closeWs := dialSessionWebSocket(t, conn["wsUrl"].(string))
	defer closeWs()
	sendExec(t, ws, code, nil)
	return joinFramesByType(collectFrames(t, ws, 60*time.Second), "stdout")
}

// TestSessionPythonStatefulFilesAndBash drives a single Python session across
// THREE consecutive executions to prove:
//
//  1. State persists between executions — a file written to /workspace in exec 1
//     is still on disk in exec 2, and an interpreter global set in exec 1 is
//     still bound in exec 2 (the per-session subprocess is reused, reset=false).
//  2. bash() works from the same session — exec 3 calls bash("grep ...") and
//     reads back the match. The just-bash OverlayFs reads through to the real
//     /workspace, so grep sees the file Python wrote in exec 1.
func TestSessionPythonStatefulFilesAndBash(t *testing.T) {
	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	conn, status := ic.Connect(t, map[string]interface{}{
		"template": "python-default",
		"language": "python",
	})
	require.Equal(t, http.StatusOK, status, "POST /sessions/connect must return 200")
	sessionID, _ := conn["sessionId"].(string)
	require.NotEmpty(t, sessionID)
	t.Cleanup(func() { _ = ic.DeleteSession(t, sessionID) })

	// Exec 1: create files on disk + bind an interpreter global.
	exec1 := `import os
os.makedirs('/workspace/proj', exist_ok=True)
with open('/workspace/proj/data.txt', 'w') as f:
    f.write('alpha\nbeta\ngamma\n')
STATE = 'persisted-marker'
print('created', os.path.exists('/workspace/proj/data.txt'))`
	out1 := wsExecStdout(t, ic, sessionID, exec1)
	assert.Contains(t, out1, "created True", "exec 1 must create the file")

	// Exec 2: the file AND the interpreter global must survive into the next exec.
	exec2 := `import os
print('exists', os.path.exists('/workspace/proj/data.txt'))
print('state', STATE)
print('contents', repr(open('/workspace/proj/data.txt').read()))`
	out2 := wsExecStdout(t, ic, sessionID, exec2)
	assert.Contains(t, out2, "exists True", "file written in exec 1 must persist into exec 2")
	assert.Contains(t, out2, "state persisted-marker", "interpreter global must persist across executions")
	assert.Contains(t, out2, "beta", "exec 2 must read back the file contents")

	// Exec 3: call bash() from the same session; grep the file created in exec 1.
	exec3 := strings.Join([]string{
		`r = bash("grep beta /workspace/proj/data.txt")`,
		`print("grep:", r.stdout.strip())`,
		`print("exit", r.exit_code)`,
	}, "\n")
	out3 := wsExecStdout(t, ic, sessionID, exec3)
	assert.Contains(t, out3, "grep: beta", "python bash() grep must match the line written in exec 1")
	assert.Contains(t, out3, "exit 0", "grep must exit 0 on a match")
}

// TestSessionTypescriptStatefulFilesAndBash drives a single TypeScript session
// across THREE consecutive executions to prove:
//
//  1. State persists between executions — a globalThis value set in exec 1 is
//     still bound in exec 2 (the V8 context is reused, reset=false).
//  2. A file created via bash() in exec 1 is still visible to bash() in exec 2
//     (each session keeps a single OverlayFs for its lifetime).
//  3. bash() works from the TS isolate — exec 3 runs bash("grep ...") (just-bash
//     runs host-side; results are copied across the isolated-vm boundary).
func TestSessionTypescriptStatefulFilesAndBash(t *testing.T) {
	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	conn, status := ic.Connect(t, map[string]interface{}{
		"template": "python-default",
		"language": "typescript",
	})
	require.Equal(t, http.StatusOK, status, "POST /sessions/connect must return 200")
	sessionID, _ := conn["sessionId"].(string)
	require.NotEmpty(t, sessionID)
	t.Cleanup(func() { _ = ic.DeleteSession(t, sessionID) })

	// Exec 1: create files via bash() + bind a global on the isolate.
	exec1 := strings.Join([]string{
		`await bash('mkdir -p /workspace/tsproj');`,
		`await bash('echo alpha > /workspace/tsproj/data.txt');`,
		`await bash('echo beta >> /workspace/tsproj/data.txt');`,
		`await bash('echo gamma >> /workspace/tsproj/data.txt');`,
		`globalThis.MARKER = 'ts-persisted';`,
		`const shown = await bash('cat /workspace/tsproj/data.txt');`,
		`console.log('created', shown.stdout.trim());`,
	}, "\n")
	out1 := wsExecStdout(t, ic, sessionID, exec1)
	assert.Contains(t, out1, "alpha", "exec 1 must create + read the file via bash()")
	assert.Contains(t, out1, "gamma", "exec 1 file must contain all appended lines")

	// Exec 2: the global AND the bash-overlay file must survive into the next exec.
	exec2 := strings.Join([]string{
		`console.log('marker', globalThis.MARKER);`,
		`const again = await bash('cat /workspace/tsproj/data.txt');`,
		`console.log('persisted', again.stdout.trim().split('\n').length, 'lines');`,
	}, "\n")
	out2 := wsExecStdout(t, ic, sessionID, exec2)
	assert.Contains(t, out2, "marker ts-persisted", "isolate global must persist across executions")
	assert.Contains(t, out2, "persisted 3 lines", "file created in exec 1 must persist into exec 2")

	// Exec 3: grep from the TS isolate over the file created in exec 1.
	exec3 := `const r = await bash('grep beta /workspace/tsproj/data.txt'); console.log('grep', r.stdout.trim(), 'exit', r.exitCode);`
	out3 := wsExecStdout(t, ic, sessionID, exec3)
	assert.Contains(t, out3, "grep beta", "ts bash() grep must match the line created in exec 1")
	assert.Contains(t, out3, "exit 0", "grep must exit 0 on a match")
}
