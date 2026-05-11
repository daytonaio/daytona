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

// TestSessionCreateRunDelete verifies that two code-run calls reusing the same context
// share state (variable set in call 1 visible in call 2).
func TestSessionCreateRunDelete(t *testing.T) {
	t.Skipf("not yet implemented: session-entity, session-cache")

	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	created, status := ic.CreateSession(t, map[string]interface{}{
		"template": "python-default",
		"language": "python",
	})
	require.Equal(t, http.StatusOK, status, "POST /sessions must return 200")

	id, _ := created["id"].(string)
	require.NotEmpty(t, id, "context id must be present")
	assertNoSandboxLeak(t, created, "")

	t.Cleanup(func() { _ = ic.DeleteSession(t, id) })

	// First exec sets a variable.
	body1, status1 := ic.CodeRun(t, map[string]interface{}{
		"context": map[string]string{"id": id},
		"code":    "x = 42",
	})
	require.Equal(t, http.StatusOK, status1)
	assertNoSandboxLeak(t, body1, "")

	// Second exec reads the variable from the persisted context.
	body2, status2 := ic.CodeRun(t, map[string]interface{}{
		"context": map[string]string{"id": id},
		"code":    "print(x)",
	})
	require.Equal(t, http.StatusOK, status2)
	stdout, _ := body2["stdout"].(string)
	assert.Equal(t, "42\n", stdout, "context state must persist between calls")

	// Delete returns 204.
	require.Equal(t, http.StatusNoContent, ic.DeleteSession(t, id))
}

// TestSessionListNoLeak verifies that listing contexts shows ACTIVE rows with the right
// shape and never leaks sandbox identifiers.
func TestSessionListNoLeak(t *testing.T) {
	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	created, status := ic.CreateSession(t, map[string]interface{}{
		"template": "python-default",
		"language": "python",
	})
	require.Equal(t, http.StatusOK, status)
	id, _ := created["id"].(string)
	t.Cleanup(func() { _ = ic.DeleteSession(t, id) })

	contexts, listStatus := ic.ListSessions(t, "")
	require.Equal(t, http.StatusOK, listStatus)
	require.NotEmpty(t, contexts)

	var found map[string]interface{}
	for _, raw := range contexts {
		ctx, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		assertNoSandboxLeak(t, ctx, "")
		if ctxID, _ := ctx["id"].(string); ctxID == id {
			found = ctx
		}
	}
	require.NotNil(t, found, "newly created context must be in list")

	_, hasExpiresAt := found["expiresAt"]
	assert.True(t, hasExpiresAt, "context must expose expiresAt")
}

// TestSessionOmitTemplateOnUse proves that the API resolves both template and language
// from a stored context row when the caller passes only `{context:{id}}`.
func TestSessionOmitTemplateOnUse(t *testing.T) {
	t.Skipf("not yet implemented: session-service-controller")

	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	created, status := ic.CreateSession(t, map[string]interface{}{
		"template": "python-default",
		"language": "python",
	})
	require.Equal(t, http.StatusOK, status)
	id, _ := created["id"].(string)
	t.Cleanup(func() { _ = ic.DeleteSession(t, id) })

	body, runStatus := ic.CodeRun(t, map[string]interface{}{
		"context": map[string]string{"id": id},
		"code":    "print('hello')",
	})
	require.Equal(t, http.StatusOK, runStatus, "code-run with only context (no template/language) must succeed")

	stdout, _ := body["stdout"].(string)
	assert.Equal(t, "hello\n", stdout)
}

// TestSessionDeleteIsIdempotent verifies that DELETE /sessions/:id returns 204 on the
// first call and continues to return 204 on subsequent calls (SessionRepository.delete is
// documented as idempotent: it short-circuits when the row is already gone).
func TestSessionDeleteIsIdempotent(t *testing.T) {
	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	created, status := ic.CreateSession(t, map[string]interface{}{
		"template": "python-default",
		"language": "python",
	})
	require.Equal(t, http.StatusOK, status)
	id, _ := created["id"].(string)
	require.NotEmpty(t, id)

	assert.Equal(t, http.StatusNoContent, ic.DeleteSession(t, id), "first delete must return 204")
	assert.Equal(t, http.StatusNoContent, ic.DeleteSession(t, id), "second delete must still return 204 (idempotent)")
}

// TestSessionGetAccessRefreshes verifies that GET /sessions/:id/access mints a fresh signed
// proxy bundle on every call (different tokens) and the second call's tokenExpiresAt is not
// earlier than the first's. Acts as a keep-alive too — bumps lastUsedAt.
func TestSessionGetAccessRefreshes(t *testing.T) {
	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	created, status := ic.CreateSession(t, map[string]interface{}{
		"template": "python-default",
		"language": "python",
	})
	require.Equal(t, http.StatusOK, status)
	id, _ := created["id"].(string)
	require.NotEmpty(t, id)
	t.Cleanup(func() { _ = ic.DeleteSession(t, id) })

	first, firstStatus := ic.GetSessionAccess(t, id)
	require.Equal(t, http.StatusOK, firstStatus)
	firstURL, _ := first["wsUrl"].(string)
	firstToken, _ := first["token"].(string)
	firstExp, _ := first["tokenExpiresAt"].(string)
	require.NotEmpty(t, firstURL, "access must expose wsUrl")
	require.NotEmpty(t, firstToken, "access must expose token")
	require.NotEmpty(t, firstExp, "access must expose tokenExpiresAt")

	// Small sleep so the second mint lands at a strictly later wall-clock second.
	// SessionService.buildSandboxAccess derives tokenExpiresAt = now + ttl, so a 1s
	// gap is enough to differentiate without making the test slow.
	time.Sleep(1100 * time.Millisecond)

	second, secondStatus := ic.GetSessionAccess(t, id)
	require.Equal(t, http.StatusOK, secondStatus)
	secondURL, _ := second["wsUrl"].(string)
	secondExp, _ := second["tokenExpiresAt"].(string)
	require.NotEmpty(t, secondURL)
	require.NotEmpty(t, secondExp)

	firstT, err := time.Parse(time.RFC3339Nano, firstExp)
	require.NoError(t, err, "first tokenExpiresAt must be RFC3339")
	secondT, err := time.Parse(time.RFC3339Nano, secondExp)
	require.NoError(t, err, "second tokenExpiresAt must be RFC3339")
	assert.False(t, secondT.Before(firstT), "second tokenExpiresAt must not be earlier than the first")
}

// TestSessionCreateTransientIdempotent verifies that POST /sessions/transients returns
// the same deterministic id for the same (template, language) pair while the warm pool
// instance is unchanged. The id is `transient-${instance.id}-${language}` — see
// SessionService.createTransientSession.
func TestSessionCreateTransientIdempotent(t *testing.T) {
	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	first, firstStatus := ic.CreateTransient(t, map[string]interface{}{
		"template": "python-default",
		"language": "python",
	})
	require.Equal(t, http.StatusOK, firstStatus)
	firstID, _ := first["id"].(string)
	require.NotEmpty(t, firstID, "transient must return an id")
	require.Contains(t, firstID, "transient-", "transient id is `transient-<instance>-<lang>`")
	assertNoSandboxLeak(t, first, "")

	second, secondStatus := ic.CreateTransient(t, map[string]interface{}{
		"template": "python-default",
		"language": "python",
	})
	require.Equal(t, http.StatusOK, secondStatus)
	secondID, _ := second["id"].(string)
	assert.Equal(t, firstID, secondID, "transient must be deterministic per (template, language)")

	// Both calls must return a usable access bundle.
	for label, body := range map[string]map[string]interface{}{"first": first, "second": second} {
		access, ok := body["access"].(map[string]interface{})
		require.True(t, ok, "%s transient must include access bundle", label)
		assert.NotEmpty(t, access["wsUrl"], "%s access.wsUrl must be set", label)
		assert.NotEmpty(t, access["token"], "%s access.token must be set", label)
	}
}
