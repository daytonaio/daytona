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

// TestSessionE2ESmoke is the canonical "did anything break globally" canary.
// One end-to-end happy path: list templates → create python context → stream a multi-line
// program with stdout/error/display/completed control frames → list contexts shows it →
// delete → list shows it gone.
func TestSessionE2ESmoke(t *testing.T) {
	t.Skipf("not yet implemented: full feature")

	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	templates, status := ic.ListTemplates(t)
	require.Equal(t, http.StatusOK, status)
	require.NotEmpty(t, templates)

	ctx, status := ic.CreateSession(t, map[string]interface{}{
		"template": "python-default",
		"language": "python",
	})
	require.Equal(t, http.StatusOK, status)
	contextID, _ := ctx["id"].(string)
	require.NotEmpty(t, contextID)
	t.Cleanup(func() { ic.DeleteSession(t, contextID) })

	connectResp, status := ic.Connect(t, map[string]interface{}{
		"context": map[string]interface{}{"id": contextID},
	})
	require.Equal(t, http.StatusOK, status)
	wsURL, _ := connectResp["wsUrl"].(string)
	require.NotEmpty(t, wsURL)

	ws, closeWS := dialSessionWebSocket(t, wsURL)
	defer closeWS()

	sendExec(t, ws, `import pandas as pd
print("hi")
pd.DataFrame({'a':[1,2]})
`, nil)

	frames := collectFrames(t, ws, 30*time.Second)
	stdout := joinFramesByType(frames, "stdout")
	assert.Contains(t, stdout, "hi", "smoke: must see hi in stdout")

	disp := findFrame(frames, "display")
	require.NotNil(t, disp, "smoke: must see at least one display frame")
	assert.NotNil(t, disp["formats"], "display must carry formats[]")

	control := findFrame(frames, "control")
	require.NotNil(t, control, "smoke: must see a control frame")
	assert.Equal(t, "completed", control["text"])

	contexts, _ := ic.ListSessions(t, "")
	found := false
	for _, raw := range contexts {
		c, _ := raw.(map[string]interface{})
		if id, _ := c["id"].(string); id == contextID {
			found = true
			break
		}
	}
	assert.True(t, found, "list-contexts must include the created context before delete")

	require.Equal(t, http.StatusNoContent, ic.DeleteSession(t, contextID))

	contexts2, _ := ic.ListSessions(t, "")
	for _, raw := range contexts2 {
		c, _ := raw.(map[string]interface{})
		if id, _ := c["id"].(string); id == contextID {
			t.Fatalf("context %s still listed after delete", contextID)
		}
	}
}
