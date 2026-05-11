// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build e2e

package e2e_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSessionResolveWorksWithoutRedis: with REDIS_URL pointed at a nonexistent host, every
// API call still succeeds (degraded latency, no errors). Postgres is the durable source of
// truth; Redis is a strict cache that gracefully degrades.
//
// Gated behind DAYTONA_E2E_REDIS_FAILURE_TEST=1 so it doesn't break the default CI run.
func TestSessionResolveWorksWithoutRedis(t *testing.T) {
	if os.Getenv("DAYTONA_E2E_REDIS_FAILURE_TEST") != "1" {
		t.Skip("DAYTONA_E2E_REDIS_FAILURE_TEST not set")
	}
	t.Skipf("not yet implemented: session-cache")

	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	ctx, status := ic.CreateSession(t, map[string]interface{}{"language": "python"})
	require.Equal(t, http.StatusOK, status)
	contextID, _ := ctx["id"].(string)
	require.NotEmpty(t, contextID)
	t.Cleanup(func() { ic.DeleteSession(t, contextID) })

	resp, status := ic.CodeRun(t, map[string]interface{}{
		"context": map[string]interface{}{"id": contextID},
		"code":    "print('hello-no-redis')",
	})
	require.Equal(t, http.StatusOK, status)
	stdout, _ := resp["stdout"].(string)
	require.Contains(t, stdout, "hello-no-redis", "must succeed even with Redis unreachable")
}
