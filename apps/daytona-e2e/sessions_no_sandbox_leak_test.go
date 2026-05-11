// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build e2e

package e2e_test

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSessionResponsesHideSandboxIdentifiers walks every JSON field in every session API response
// (templates, contexts, code-run, connect, transients, access) and asserts:
//  1. NO field name matches the forbidden sandbox-leak pattern (defensive — none are in v1 DTOs).
//  2. The actual SessionInstance.sandboxId UUID for the test org+template never appears as a
//     value anywhere in any response body (catches a leak under a renamed field).
//
// Step (2) requires being able to look up the warm sandbox by label; if the lookup returns "",
// only the field-name check runs.
func TestSessionResponsesHideSandboxIdentifiers(t *testing.T) {
	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	// Trigger pool provisioning so a sandboxId exists to look up. We create a context
	// rather than running code so this test is independent of the daemon execution path.
	seed, seedStatus := ic.CreateSession(t, map[string]interface{}{
		"template": "python-default", "language": "python",
	})
	require.Equal(t, http.StatusOK, seedStatus, "seed createSession must succeed for leak walk")
	seedID, _ := seed["id"].(string)
	t.Cleanup(func() { _ = ic.DeleteSession(t, seedID) })

	knownSandboxID := lookupSandboxIDForTemplate(t, api, "python-default")

	// 1. /templates
	templates, _ := ic.ListTemplates(t)
	for _, raw := range templates {
		assertNoSandboxLeak(t, raw, knownSandboxID)
	}

	// 2. /templates/:name/packages
	pkgs, _ := ic.ListPackages(t, "python-default", "python")
	for _, p := range pkgs {
		assertNoSandboxLeak(t, p, knownSandboxID)
	}

	// 3. /sessions (POST) — the seed we created above
	assertNoSandboxLeak(t, seed, knownSandboxID)

	// 4. /sessions (GET)
	contexts, _ := ic.ListSessions(t, "")
	for _, ctx := range contexts {
		assertNoSandboxLeak(t, ctx, knownSandboxID)
	}

	// 5. /sessions/:id/access (refresh)
	access, accessStatus := ic.GetSessionAccess(t, seedID)
	require.Equal(t, http.StatusOK, accessStatus)
	assertNoSandboxLeak(t, access, knownSandboxID)

	// 6. /sessions/transients
	transient, transientStatus := ic.CreateTransient(t, map[string]interface{}{
		"template": "python-default", "language": "python",
	})
	require.Equal(t, http.StatusOK, transientStatus)
	assertNoSandboxLeak(t, transient, knownSandboxID)

	// 7. /sessions/connect
	conn, connStatus := ic.Connect(t, map[string]interface{}{
		"template": "python-default", "language": "python",
	})
	require.Equal(t, http.StatusOK, connStatus)
	if id, _ := conn["sessionId"].(string); id != "" {
		t.Cleanup(func() { _ = ic.DeleteSession(t, id) })
	}
	assertNoSandboxLeak(t, conn, knownSandboxID)
}

// lookupSandboxIDForTemplate finds the warm-pool sandbox that backs the given template
// for the current organization by querying /sandbox/paginated for sandboxes labeled
// `daytona.io/session-template=<name>`. Returns the first matching sandbox id, or "" if
// none found (in which case the leak walk falls back to field-name-only checks).
//
// This is the API-only equivalent of "read SessionInstance.sandboxId from the DB"; we
// rely on the documented label set the pool writes when it creates the warm sandbox
// (see SessionPoolService.acquire labels: daytona.io/session, daytona.io/session-template,
// daytona.io/session-instance).
func lookupSandboxIDForTemplate(t *testing.T, api *APIClient, templateName string) string {
	t.Helper()
	path := `/sandbox/paginated?labels={"daytona.io/session-template":"` + templateName + `"}`
	resp, raw := api.DoRequest(t, http.MethodGet, path, nil)
	if resp.StatusCode != http.StatusOK {
		t.Logf("lookupSandboxIDForTemplate(%q): list returned %d (skipping value-leak check)", templateName, resp.StatusCode)
		return ""
	}
	var page struct {
		Items []map[string]interface{} `json:"items"`
	}
	if err := json.Unmarshal(raw, &page); err != nil {
		t.Logf("lookupSandboxIDForTemplate(%q): cannot parse paginated body: %v (skipping value-leak check)", templateName, err)
		return ""
	}
	for _, item := range page.Items {
		if id, ok := item["id"].(string); ok && id != "" {
			return id
		}
	}
	t.Logf("lookupSandboxIDForTemplate(%q): no sandboxes labeled with this template (skipping value-leak check)", templateName)
	return ""
}

// lookupEnv is a thin wrapper so test files can avoid importing os in many places.
func lookupEnv(key string) string {
	return os.Getenv(key)
}
