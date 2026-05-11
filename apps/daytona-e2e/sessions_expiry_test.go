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

// TestSessionExpiresOnIdle: with an overridden idle TTL of 10s, a context that
// goes unused for ~70s (one cron tick + slack) returns HTTP 410 with name=ContextExpired,
// reason=idle.
func TestSessionExpiresOnIdle(t *testing.T) {
	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	withTtlOverride(t, 10, 86400)

	ctx, status := ic.CreateSession(t, map[string]interface{}{"language": "python"})
	require.Equal(t, http.StatusOK, status)
	contextID, _ := ctx["id"].(string)
	require.NotEmpty(t, contextID)
	t.Cleanup(func() { ic.DeleteSession(t, contextID) })

	t.Log("waiting ~70s for idle TTL + sweep cron")
	time.Sleep(70 * time.Second)

	resp, statusCode := ic.CodeRun(t, map[string]interface{}{
		"context": map[string]interface{}{"id": contextID},
		"code":    "print('after-idle')",
	})
	assert.Equal(t, http.StatusGone, statusCode, "expected 410 after idle TTL")
	if errObj, ok := resp["error"].(map[string]interface{}); ok {
		assert.Equal(t, "SessionExpired", errObj["name"])
		assert.Equal(t, "idle", errObj["reason"])
	}
}

// TestSessionExpiresOnAbsolute: absolute TTL=30s; even constant pings can't keep
// the context alive past the absolute deadline.
func TestSessionExpiresOnAbsolute(t *testing.T) {
	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	withTtlOverride(t, 3600, 30)

	ctx, status := ic.CreateSession(t, map[string]interface{}{"language": "python"})
	require.Equal(t, http.StatusOK, status)
	contextID, _ := ctx["id"].(string)
	require.NotEmpty(t, contextID)
	t.Cleanup(func() { ic.DeleteSession(t, contextID) })

	deadline := time.Now().Add(80 * time.Second)
	for time.Now().Before(deadline) {
		_, statusCode := ic.CodeRun(t, map[string]interface{}{
			"context": map[string]interface{}{"id": contextID},
			"code":    "1",
		})
		if statusCode == http.StatusGone {
			return
		}
		time.Sleep(5 * time.Second)
	}
	t.Fatal("context did not expire within the expected window")
}

// TestSessionHardDeleteAfterGrace: after a context reaches EXPIRED, set a 5s grace
// period and verify the row is hard-deleted (next call returns 404, not 410). Relies on
// hard-delete running every minute (see plan: hard-delete is @Cron(EVERY_MINUTE) precisely
// to make this test viable without a "trigger sweep now" admin hook).
func TestSessionHardDeleteAfterGrace(t *testing.T) {
	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	withTtlOverride(t, 10, 86400)
	withGracePeriodOverride(t, 5)

	ctx, status := ic.CreateSession(t, map[string]interface{}{"language": "python"})
	require.Equal(t, http.StatusOK, status)
	contextID, _ := ctx["id"].(string)
	require.NotEmpty(t, contextID)

	t.Log("waiting ~70s for idle expiry, then ~70s more for hard-delete")
	time.Sleep(150 * time.Second)

	_, statusCode := ic.CodeRun(t, map[string]interface{}{
		"context": map[string]interface{}{"id": contextID},
		"code":    "1",
	})
	assert.Equal(t, http.StatusNotFound, statusCode, "expected 404 after grace period (row hard-deleted)")
}
