// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build e2e

package e2e_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHealthAPI verifies the API server is running and healthy.
func TestHealthAPI(t *testing.T) {
	cfg := LoadConfig(t)
	client := NewAPIClient(cfg)

	resp, body := client.DoRequest(t, http.MethodGet, "/health", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "GET /health returned non-200: %s", string(body))

	t.Logf("GET /health response: %s", string(body))
}

// TestSandboxListReachable verifies the sandbox API is reachable and auth works.
// Uses a label filter that won't match anything to avoid listing real sandboxes.
func TestSandboxListReachable(t *testing.T) {
	cfg := LoadConfig(t)
	client := NewAPIClient(cfg)

	// Use a nonexistent run ID so we get an empty list (not real sandboxes)
	labels := `{"e2e":"true","test-run":"nonexistent-0"}`
	path := "/sandbox/paginated?labels=" + labels

	resp, body := client.DoRequest(t, http.MethodGet, path, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode, "GET /sandbox/paginated returned non-200: %s", string(body))

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result), "failed to parse paginated response: %s", string(body))

	// Response must contain "items" array (may be empty)
	items, ok := result["items"]
	assert.True(t, ok, "paginated response missing 'items' field")
	if ok {
		_, isSlice := items.([]interface{})
		assert.True(t, isSlice, "paginated response 'items' is not an array")
	}

	t.Logf("GET /sandbox/paginated: total=%v items", result["total"])
}
