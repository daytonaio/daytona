// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build e2e

package e2e_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSessionRejectsMissingApiKey verifies that POST /api/sessions/code-run without an
// Authorization header returns 401.
func TestSessionRejectsMissingApiKey(t *testing.T) {
	cfg := LoadConfig(t)
	url := strings.TrimRight(cfg.BaseURL, "/") + "/sessions/code-run"

	body, err := json.Marshal(map[string]interface{}{"language": "python", "code": "print(1)"})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	httpCli := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpCli.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestSessionRejectsCrossOrgContext verifies that org A creating a context and org B trying to
// use its id returns 404 (NOT 403, to avoid existence leak).
func TestSessionRejectsCrossOrgContext(t *testing.T) {
	t.Skipf("not yet implemented: session-entity")

	cfg := LoadConfig(t)
	apiA := NewAPIClient(cfg)
	icA := NewSessionClient(apiA)

	keyB := lookupEnv("DAYTONA_E2E_API_KEY_ORG_B")
	if keyB == "" {
		t.Skip("DAYTONA_E2E_API_KEY_ORG_B not set — cannot test cross-org isolation")
	}
	cfgB := cfg
	cfgB.APIKey = keyB
	apiB := NewAPIClient(cfgB)
	icB := NewSessionClient(apiB)

	created, status := icA.CreateSession(t, map[string]interface{}{
		"template": "python-default", "language": "python",
	})
	require.Equal(t, http.StatusOK, status)
	id, _ := created["id"].(string)
	t.Cleanup(func() { _ = icA.DeleteSession(t, id) })

	_, runStatus := icB.CodeRun(t, map[string]interface{}{
		"context": map[string]string{"id": id},
		"code":    "print('cross org')",
	})
	assert.Equal(t, http.StatusNotFound, runStatus,
		"cross-org context use must return 404 (not 403, to avoid existence leak)")
}
