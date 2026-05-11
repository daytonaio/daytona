// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build e2e

package e2e_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

// SessionClient is a thin wrapper over the API client with session-specific helpers.
// Mirrors the toolbox_test.go pattern of black-box HTTP/WS testing.
type SessionClient struct {
	api *APIClient
}

func NewSessionClient(api *APIClient) *SessionClient {
	return &SessionClient{api: api}
}

// CodeRun calls POST /api/sessions/code-run with the given body.
// Returns the parsed response body and HTTP status.
func (ic *SessionClient) CodeRun(t *testing.T, body map[string]interface{}) (map[string]interface{}, int) {
	t.Helper()
	resp, raw := ic.api.DoRequest(t, http.MethodPost, "/sessions/code-run", body)
	var parsed map[string]interface{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &parsed)
	}
	return parsed, resp.StatusCode
}

// Connect calls POST /api/sessions/connect and returns the parsed response.
func (ic *SessionClient) Connect(t *testing.T, body map[string]interface{}) (map[string]interface{}, int) {
	t.Helper()
	resp, raw := ic.api.DoRequest(t, http.MethodPost, "/sessions/connect", body)
	var parsed map[string]interface{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &parsed)
	}
	return parsed, resp.StatusCode
}

// CreateSession calls POST /api/sessions.
func (ic *SessionClient) CreateSession(t *testing.T, body map[string]interface{}) (map[string]interface{}, int) {
	t.Helper()
	resp, raw := ic.api.DoRequest(t, http.MethodPost, "/sessions", body)
	var parsed map[string]interface{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &parsed)
	}
	return parsed, resp.StatusCode
}

// ListSessions calls GET /api/sessions.
func (ic *SessionClient) ListSessions(t *testing.T, template string) ([]interface{}, int) {
	t.Helper()
	path := "/sessions"
	if template != "" {
		path += "?template=" + url.QueryEscape(template)
	}
	resp, raw := ic.api.DoRequest(t, http.MethodGet, path, nil)
	var parsed []interface{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &parsed)
	}
	return parsed, resp.StatusCode
}

// DeleteSession calls DELETE /api/sessions/:id.
func (ic *SessionClient) DeleteSession(t *testing.T, id string) int {
	t.Helper()
	resp, _ := ic.api.DoRequest(t, http.MethodDelete, "/sessions/"+id, nil)
	return resp.StatusCode
}

// CreateTransient calls POST /api/sessions/transients.
// Returns the parsed SessionDto (including the embedded `access` bundle) and HTTP status.
func (ic *SessionClient) CreateTransient(t *testing.T, body map[string]interface{}) (map[string]interface{}, int) {
	t.Helper()
	resp, raw := ic.api.DoRequest(t, http.MethodPost, "/sessions/transients", body)
	var parsed map[string]interface{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &parsed)
	}
	return parsed, resp.StatusCode
}

// GetSessionAccess calls GET /api/sessions/:id/access. Refreshes the SDK's
// direct-to-sandbox access bundle and acts as a keep-alive (bumps lastUsedAt).
func (ic *SessionClient) GetSessionAccess(t *testing.T, id string) (map[string]interface{}, int) {
	t.Helper()
	resp, raw := ic.api.DoRequest(t, http.MethodGet, "/sessions/"+id+"/access", nil)
	var parsed map[string]interface{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &parsed)
	}
	return parsed, resp.StatusCode
}

// ListTemplates calls GET /api/sessions/templates.
func (ic *SessionClient) ListTemplates(t *testing.T) ([]interface{}, int) {
	t.Helper()
	resp, raw := ic.api.DoRequest(t, http.MethodGet, "/sessions/templates", nil)
	var parsed []interface{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &parsed)
	}
	return parsed, resp.StatusCode
}

// ListPackages calls GET /api/sessions/templates/:name/packages.
func (ic *SessionClient) ListPackages(t *testing.T, template, language string) ([]interface{}, int) {
	t.Helper()
	path := fmt.Sprintf("/sessions/templates/%s/packages?language=%s", url.PathEscape(template), url.QueryEscape(language))
	resp, raw := ic.api.DoRequest(t, http.MethodGet, path, nil)
	var parsed []interface{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &parsed)
	}
	return parsed, resp.StatusCode
}

// dialSessionWebSocket opens a WebSocket against the signed URL returned by /sessions/connect.
// The signed URL embeds the auth token in the subdomain (no extra Authorization header needed).
// Returns a connected websocket and a cleanup function.
func dialSessionWebSocket(t *testing.T, wsURL string) (*websocket.Conn, func()) {
	t.Helper()
	dialer := websocket.Dialer{HandshakeTimeout: 30 * time.Second}
	conn, _, err := dialer.Dial(wsURL, nil)
	require.NoError(t, err, "failed to dial session websocket: %s", wsURL)
	return conn, func() {
		_ = conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "test done"),
			time.Now().Add(time.Second))
		_ = conn.Close()
	}
}

// collectFrames drains the WebSocket until a {type:"control",text:"completed"} frame is observed
// or the timeout elapses. Returns the frames in arrival order.
func collectFrames(t *testing.T, ws *websocket.Conn, timeout time.Duration) []map[string]interface{} {
	t.Helper()
	deadline := time.Now().Add(timeout)
	frames := make([]map[string]interface{}, 0, 16)
	for time.Now().Before(deadline) {
		_ = ws.SetReadDeadline(deadline)
		var msg map[string]interface{}
		if err := ws.ReadJSON(&msg); err != nil {
			t.Logf("collectFrames: read returned %v after %d frames", err, len(frames))
			return frames
		}
		frames = append(frames, msg)
		if frameType, _ := msg["type"].(string); frameType == "control" {
			if text, _ := msg["text"].(string); text == "completed" {
				return frames
			}
		}
	}
	return frames
}

// sendExec writes an exec request frame to the WebSocket.
func sendExec(t *testing.T, ws *websocket.Conn, code string, opts map[string]interface{}) {
	t.Helper()
	payload := map[string]interface{}{"code": code}
	for k, v := range opts {
		payload[k] = v
	}
	_ = ws.SetWriteDeadline(time.Now().Add(5 * time.Second))
	require.NoError(t, ws.WriteJSON(payload), "failed to write exec frame")
}

// joinFramesByType concatenates the `text` field of every frame whose type matches.
func joinFramesByType(frames []map[string]interface{}, frameType string) string {
	var b strings.Builder
	for _, f := range frames {
		if t, _ := f["type"].(string); t == frameType {
			if text, ok := f["text"].(string); ok {
				b.WriteString(text)
			}
		}
	}
	return b.String()
}

// findFrame returns the first frame matching the given type, or nil.
func findFrame(frames []map[string]interface{}, frameType string) map[string]interface{} {
	for _, f := range frames {
		if t, _ := f["type"].(string); t == frameType {
			return f
		}
	}
	return nil
}

// findFrames returns every frame matching the given type.
func findFrames(frames []map[string]interface{}, frameType string) []map[string]interface{} {
	out := make([]map[string]interface{}, 0)
	for _, f := range frames {
		if t, _ := f["type"].(string); t == frameType {
			out = append(out, f)
		}
	}
	return out
}

// sandboxLeakFieldRegex matches field names that should never appear in any session API response
// (defensive — none of these are in the v1 DTOs as designed).
var sandboxLeakFieldRegex = regexp.MustCompile(`(?i)^(sandbox|sandboxId|instanceId|snapshot|snapshotId|templateId)$`)

// assertNoSandboxLeak walks an arbitrary JSON tree and fails if either:
//  1. Any field name matches sandboxLeakFieldRegex (defensive — locks down "future drift cannot
//     reintroduce a leak by adding a field"), or
//  2. Any string value equals knownSandboxId (catches a leak under a renamed field).
//
// Pass an empty knownSandboxId to skip check (2). UUID-shape regex alone is useless because
// context-ids and template-ids share that shape.
func assertNoSandboxLeak(t *testing.T, body interface{}, knownSandboxID string) {
	t.Helper()
	walk(t, "", body, knownSandboxID)
}

func walk(t *testing.T, path string, v interface{}, knownSandboxID string) {
	switch val := v.(type) {
	case map[string]interface{}:
		for k, sub := range val {
			if sandboxLeakFieldRegex.MatchString(k) {
				t.Errorf("session API leak: field %q at %q matches forbidden pattern (value=%v)", k, path, sub)
			}
			walk(t, path+"."+k, sub, knownSandboxID)
		}
	case []interface{}:
		for i, sub := range val {
			walk(t, fmt.Sprintf("%s[%d]", path, i), sub, knownSandboxID)
		}
	case string:
		if knownSandboxID != "" && val == knownSandboxID {
			t.Errorf("session API leak: sandboxID %q surfaced as a value at %q", knownSandboxID, path)
		}
	default:
		_ = reflect.TypeOf(v)
	}
}

// withTtlOverride sets SESSION_IDLE_TTL_SECONDS / SESSION_ABSOLUTE_TTL_SECONDS
// on the API process for the duration of the test. The API reads these at every cron tick, so
// changes take effect on the next minute boundary without an API restart.
//
// Note: this only works if the e2e test runs in-process with the API (or the API binary inherits
// the test process's env). For out-of-process API runs, the test must be skipped via
// DAYTONA_E2E_TTL_OVERRIDE_SUPPORTED.
func withTtlOverride(t *testing.T, idleSec, absSec int) {
	t.Helper()
	if os.Getenv("DAYTONA_E2E_TTL_OVERRIDE_SUPPORTED") != "1" {
		t.Skip("DAYTONA_E2E_TTL_OVERRIDE_SUPPORTED not set; cannot mutate API env vars")
	}
	prevIdle := os.Getenv("SESSION_IDLE_TTL_SECONDS")
	prevAbs := os.Getenv("SESSION_ABSOLUTE_TTL_SECONDS")
	require.NoError(t, os.Setenv("SESSION_IDLE_TTL_SECONDS", fmt.Sprintf("%d", idleSec)))
	require.NoError(t, os.Setenv("SESSION_ABSOLUTE_TTL_SECONDS", fmt.Sprintf("%d", absSec)))
	t.Cleanup(func() {
		if prevIdle == "" {
			_ = os.Unsetenv("SESSION_IDLE_TTL_SECONDS")
		} else {
			_ = os.Setenv("SESSION_IDLE_TTL_SECONDS", prevIdle)
		}
		if prevAbs == "" {
			_ = os.Unsetenv("SESSION_ABSOLUTE_TTL_SECONDS")
		} else {
			_ = os.Setenv("SESSION_ABSOLUTE_TTL_SECONDS", prevAbs)
		}
	})
}

// withGracePeriodOverride sets SESSION_EXPIRED_GRACE_SECONDS for the test duration.
func withGracePeriodOverride(t *testing.T, graceSec int) {
	t.Helper()
	if os.Getenv("DAYTONA_E2E_TTL_OVERRIDE_SUPPORTED") != "1" {
		t.Skip("DAYTONA_E2E_TTL_OVERRIDE_SUPPORTED not set; cannot mutate API env vars")
	}
	prev := os.Getenv("SESSION_EXPIRED_GRACE_SECONDS")
	require.NoError(t, os.Setenv("SESSION_EXPIRED_GRACE_SECONDS", fmt.Sprintf("%d", graceSec)))
	t.Cleanup(func() {
		if prev == "" {
			_ = os.Unsetenv("SESSION_EXPIRED_GRACE_SECONDS")
		} else {
			_ = os.Setenv("SESSION_EXPIRED_GRACE_SECONDS", prev)
		}
	})
}
