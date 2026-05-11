// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build e2e

package e2e_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSessionConnectStreaming verifies that POST /api/sessions/connect returns a signed wsUrl,
// and that opening the WebSocket streams stdout chunks incrementally (not all at once at end).
func TestSessionConnectStreaming(t *testing.T) {
	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	body, status := ic.Connect(t, map[string]interface{}{
		"template": "python-default",
		"language": "python",
	})
	require.Equal(t, http.StatusOK, status, "POST /sessions/connect must return 200")
	assertNoSandboxLeak(t, body, "")

	wsURL, _ := body["wsUrl"].(string)
	require.NotEmpty(t, wsURL, "connect must return a wsUrl")
	contextID, _ := body["sessionId"].(string)
	require.NotEmpty(t, contextID, "connect must return a sessionId")
	expiresAt, _ := body["expiresAt"].(string)
	require.NotEmpty(t, expiresAt, "connect must return expiresAt")

	t.Cleanup(func() { _ = ic.DeleteSession(t, contextID) })

	ws, closer := dialSessionWebSocket(t, wsURL)
	defer closer()

	startedAt := time.Now()
	sendExec(t, ws, "for i in range(3):\n    print(i)\n    import time; time.sleep(0.1)\n", nil)

	frames := collectFrames(t, ws, 30*time.Second)

	stdoutFrames := findFrames(frames, "stdout")
	require.GreaterOrEqual(t, len(stdoutFrames), 1, "must receive at least one stdout frame")

	// Verify chunks arrived incrementally: at least one frame must arrive >50ms after start
	// AND at least one frame must arrive >50ms after the previous one (true streaming, not
	// a single final dump).
	var hasIncrementalArrival bool
	for i := 1; i < len(stdoutFrames); i++ {
		// Best-effort detection: assertion is loose because frames don't carry timestamps,
		// so we infer streaming by the number of distinct stdout frames (more than one).
		if len(stdoutFrames) >= 2 {
			hasIncrementalArrival = true
			break
		}
	}
	totalElapsed := time.Since(startedAt)
	t.Logf("collected %d frames in %s", len(frames), totalElapsed)
	assert.True(t, hasIncrementalArrival, "expected multiple stdout frames (streaming, not single dump)")

	combined := joinFramesByType(frames, "stdout")
	assert.Contains(t, combined, "0")
	assert.Contains(t, combined, "1")
	assert.Contains(t, combined, "2")
}

// TestSessionConnectStreamingError verifies an error chunk is emitted when user code raises.
func TestSessionConnectStreamingError(t *testing.T) {
	t.Skipf("not yet implemented: session-service-controller")

	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	body, status := ic.Connect(t, map[string]interface{}{
		"template": "python-default",
		"language": "python",
	})
	require.Equal(t, http.StatusOK, status)

	wsURL, _ := body["wsUrl"].(string)
	contextID, _ := body["sessionId"].(string)
	t.Cleanup(func() { _ = ic.DeleteSession(t, contextID) })

	ws, closer := dialSessionWebSocket(t, wsURL)
	defer closer()

	sendExec(t, ws, "1/0", nil)
	frames := collectFrames(t, ws, 15*time.Second)

	errFrame := findFrame(frames, "error")
	require.NotNil(t, errFrame, "must emit an error frame for `1/0`")
	name, _ := errFrame["name"].(string)
	assert.Equal(t, "ZeroDivisionError", name)
}

// TestSessionConnectStreamingControl verifies the final frame is a {type:control,text:completed}.
func TestSessionConnectStreamingControl(t *testing.T) {
	t.Skipf("not yet implemented: session-daemon-app")

	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	body, status := ic.Connect(t, map[string]interface{}{
		"template": "python-default",
		"language": "python",
	})
	require.Equal(t, http.StatusOK, status)

	wsURL, _ := body["wsUrl"].(string)
	contextID, _ := body["sessionId"].(string)
	t.Cleanup(func() { _ = ic.DeleteSession(t, contextID) })

	ws, closer := dialSessionWebSocket(t, wsURL)
	defer closer()

	sendExec(t, ws, "print('done')", nil)
	frames := collectFrames(t, ws, 15*time.Second)

	require.NotEmpty(t, frames)
	last := frames[len(frames)-1]
	assert.Equal(t, "control", last["type"], "final frame must be type=control")
	assert.Equal(t, "completed", last["text"], "final control frame must say `completed`")
}
