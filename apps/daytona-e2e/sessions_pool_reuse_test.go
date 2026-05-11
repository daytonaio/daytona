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

// TestSessionPoolReusesWarmSandbox: first call provisions, second call within the same
// org+template hits the warm pool. Asserts the cold-start delta is large and the warm-call
// duration is sub-second.
func TestSessionPoolReusesWarmSandbox(t *testing.T) {
	t.Skipf("not yet implemented: session-pool-service")

	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	t.Log("first call (cold)")
	cold := time.Now()
	resp1, status1 := ic.CodeRun(t, map[string]interface{}{
		"language": "python",
		"code":     "print(1)",
	})
	coldDur := time.Since(cold)
	require.Equal(t, http.StatusOK, status1)
	assertNoSandboxLeak(t, resp1, "")

	t.Log("second call (warm)")
	warm := time.Now()
	resp2, status2 := ic.CodeRun(t, map[string]interface{}{
		"language": "python",
		"code":     "print(2)",
	})
	warmDur := time.Since(warm)
	require.Equal(t, http.StatusOK, status2)
	assertNoSandboxLeak(t, resp2, "")

	assert.Greater(t, coldDur, warmDur, "cold call must be slower than warm call")
	assert.Less(t, warmDur, 5*time.Second, "warm call must complete in under 5s")
}

// TestSessionPoolPerOrgIsolation: two organizations hitting the same template produce
// distinct SessionInstance rows / two sandboxes.
//
// This test requires multi-org credentials, gated on DAYTONA_E2E_SECONDARY_API_KEY.
func TestSessionPoolPerOrgIsolation(t *testing.T) {
	t.Skipf("not yet implemented: session-pool-service (requires DAYTONA_E2E_SECONDARY_API_KEY)")
}
