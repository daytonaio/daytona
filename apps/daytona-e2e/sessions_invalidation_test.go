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

// TestSessionInvalidatedOnSandboxStop verifies that stopping the underlying
// sandbox marks all of its contexts INVALID and surfaces a clean SessionInvalidatedError
// to the next caller.
//
// The test looks up the underlying SessionInstance.sandboxId via the e2e suite's existing
// DB connection (same pattern cleanup_test.go uses) — there is intentionally no API route
// for this; v1 has no admin CRUD for SessionInstance.
func TestSessionInvalidatedOnSandboxStop(t *testing.T) {
	t.Skipf("not yet implemented: session-pool-service, session-cache")

	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	ctx, status := ic.CreateSession(t, map[string]interface{}{"language": "python"})
	require.Equal(t, http.StatusOK, status)
	contextID, _ := ctx["id"].(string)
	require.NotEmpty(t, contextID)
	t.Cleanup(func() { ic.DeleteSession(t, contextID) })

	// In a real run, the test would now look up the SessionInstance.sandboxId via DB and
	// stop the sandbox. The implementation hook is the snapshot-drift and sandbox-stop
	// observability path, captured here as a placeholder.
	t.Log("placeholder: would stop the sandbox, then call code-run and assert HTTP 410 ContextInvalidated")

	resp, _ := ic.CodeRun(t, map[string]interface{}{
		"context": map[string]interface{}{"id": contextID},
		"code":    "print('after-stop')",
	})
	if resp == nil {
		t.Fatal("expected an error response from code-run after sandbox stop")
	}
	if errObj, ok := resp["error"].(map[string]interface{}); ok {
		assert.Equal(t, "SessionInvalidated", errObj["name"])
		assert.NotEmpty(t, errObj["invalidatedAt"])
	}
}

// TestSessionInvalidatedOnSnapshotDrift verifies that repointing the template
// to a different snapshot triggers a pool reconcile that marks all dependent contexts
// INVALID. v1 has no admin CRUD for SessionTemplate, so the test mutates session_template
// directly via the e2e suite's DB connection — this is a single documented expedient,
// scoped to e2e, and will move to the admin endpoint when v1.1 ships it.
func TestSessionInvalidatedOnSnapshotDrift(t *testing.T) {
	t.Skipf("not yet implemented: session-pool-service")

	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	ctx, status := ic.CreateSession(t, map[string]interface{}{"language": "python"})
	require.Equal(t, http.StatusOK, status)
	contextID, _ := ctx["id"].(string)
	require.NotEmpty(t, contextID)
	t.Cleanup(func() { ic.DeleteSession(t, contextID) })

	// Placeholder for the direct DB UPDATE on session_template that flips the snapshot id.
	t.Log("placeholder: would repoint session_template to a different snapshot via direct DB UPDATE")

	// Wait one reconciler tick (default 30s) plus slack.
	time.Sleep(45 * time.Second)

	resp, _ := ic.CodeRun(t, map[string]interface{}{
		"context": map[string]interface{}{"id": contextID},
		"code":    "print('after-drift')",
	})
	if errObj, ok := resp["error"].(map[string]interface{}); ok {
		assert.Equal(t, "SessionInvalidated", errObj["name"])
	}
}
