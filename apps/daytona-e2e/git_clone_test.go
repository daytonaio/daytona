// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build e2e

package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// gitCloneRepoURL is a small public repo used to exercise both clone paths.
// Kept intentionally small: the goal is to prove the endpoint works end-to-end,
// not to stress memory (covered by the daemon unit tests and the clone design).
const gitCloneRepoURL = "https://github.com/daytonaio/daytona.git"

// cloneHTTPTimeout is the overall budget for the POST /git/clone request.
// Generous because daytonaio/daytona is a few hundred MB and the sandbox
// runner only guarantees modest bandwidth.
const cloneHTTPTimeout = 10 * time.Minute

// TestGitClone covers both clone codepaths in the daemon:
//
//   - GoGitDefault: default clone uses the in-process go-git library.
//   - ExperimentalCLIPath: when DAYTONA_EXPERIMENTAL_USE_GIT_CLONE_CLI=true
//     is set on the sandbox, the daemon shells out to `git`.
//
// Both subtests clone the same small public repo and verify the working tree
// was materialized with a valid .git directory and at least one commit.
func TestGitClone(t *testing.T) {
	cfg := LoadConfig(t)

	t.Run("GoGitDefault", func(t *testing.T) {
		runGitCloneCase(t, cfg, nil)
	})

	t.Run("ExperimentalCLIPath", func(t *testing.T) {
		runGitCloneCase(t, cfg, map[string]string{
			"DAYTONA_EXPERIMENTAL_USE_GIT_CLONE_CLI": "true",
		})
	})
}

// runGitCloneCase creates a sandbox (optionally with extra env vars),
// clones gitCloneRepoURL via the toolbox /git/clone endpoint, and verifies
// the resulting working tree.
func runGitCloneCase(t *testing.T, cfg Config, envVars map[string]string) {
	t.Helper()

	client := NewAPIClient(cfg)
	runID := testRunID()
	clonePath := fmt.Sprintf("/tmp/e2e-clone-%s", runID[4:])

	createReq := map[string]interface{}{
		"name":   fmt.Sprintf("e2e-git-clone-%s", runID[4:]),
		"labels": sandboxLabels(runID),
	}
	if cfg.Snapshot != "" {
		createReq["snapshot"] = cfg.Snapshot
	}
	if len(envVars) > 0 {
		createReq["envVars"] = envVars
	}

	sandbox := client.CreateSandbox(t, createReq)
	sandboxID, _ := sandbox["id"].(string)
	require.NotEmpty(t, sandboxID, "sandbox must have id")

	started := client.PollSandboxState(t, sandboxID, "started", cfg.PollTimeout, cfg.PollInterval)

	toolboxProxyURL, _ := started["toolboxProxyUrl"].(string)
	if toolboxProxyURL == "" {
		t.Skip("toolboxProxyUrl not available — skipping git clone test")
	}
	baseURL := strings.TrimRight(toolboxProxyURL, "/") + "/" + sandboxID

	t.Logf("sandbox %s started; cloning %s to %s (envVars=%v)",
		sandboxID, gitCloneRepoURL, clonePath, envVars)

	// POST /git/clone
	cloneReq, err := json.Marshal(map[string]interface{}{
		"url":  gitCloneRepoURL,
		"path": clonePath,
	})
	require.NoError(t, err)

	cloneHTTP := &http.Client{Timeout: cloneHTTPTimeout}

	req, err := http.NewRequest(http.MethodPost, baseURL+"/git/clone", bytes.NewReader(cloneReq))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	cloneStart := time.Now()
	resp, err := cloneHTTP.Do(req)
	require.NoError(t, err, "git clone request failed")
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode,
		"git clone must return 200: status=%d body=%s (elapsed %s)",
		resp.StatusCode, string(respBody), time.Since(cloneStart))
	t.Logf("clone succeeded in %s", time.Since(cloneStart))

	// Verify the working tree was materialized: .git present + at least one commit.
	execHTTP := &http.Client{Timeout: 30 * time.Second}

	t.Run("DotGitPresent", func(t *testing.T) {
		out := execSandboxCommand(t, execHTTP, cfg, baseURL,
			fmt.Sprintf("ls -la %s/.git/HEAD", clonePath))
		assert.Contains(t, out.result, "HEAD", "%s/.git/HEAD must exist: %s", clonePath, out.result)
		assert.Equal(t, 0, out.exitCode, "ls must exit 0; got %d: %s", out.exitCode, out.result)
	})

	t.Run("HasAtLeastOneCommit", func(t *testing.T) {
		out := execSandboxCommand(t, execHTTP, cfg, baseURL,
			fmt.Sprintf("git -C %s log --oneline -1", clonePath))
		assert.Equal(t, 0, out.exitCode,
			"git log must exit 0; got %d: %s", out.exitCode, out.result)
		assert.NotEmpty(t, strings.TrimSpace(out.result), "git log output must not be empty")
	})
}

// sandboxExecResult is the shape of a /process/execute response we care about.
type sandboxExecResult struct {
	result   string
	exitCode int
}

// execSandboxCommand runs a shell command inside the sandbox via the toolbox
// /process/execute endpoint and returns stdout + exit code.
func execSandboxCommand(t *testing.T, httpCli *http.Client, cfg Config, baseURL, command string) sandboxExecResult {
	t.Helper()

	body, err := json.Marshal(map[string]interface{}{
		"command": command,
		"timeout": 60,
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, baseURL+"/process/execute", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpCli.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"/process/execute must return 200 for %q: %s", command, string(respBody))

	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal(respBody, &parsed))

	// exitCode must be present and numeric; a missing/malformed field would
	// otherwise silently default to 0 and make a broken response look like a
	// successful command.
	exitCodeRaw, ok := parsed["exitCode"]
	require.True(t, ok, "/process/execute response missing exitCode: %s", string(respBody))
	exitCodeF, ok := exitCodeRaw.(float64)
	require.True(t, ok, "/process/execute exitCode is not numeric: %v (%s)", exitCodeRaw, string(respBody))

	// result may legitimately be empty (e.g. command produced no stdout), so
	// tolerate the zero value but still fail on wrong type.
	result := ""
	if raw, present := parsed["result"]; present && raw != nil {
		s, ok := raw.(string)
		require.True(t, ok, "/process/execute result is not a string: %v (%s)", raw, string(respBody))
		result = s
	}

	return sandboxExecResult{result: result, exitCode: int(exitCodeF)}
}
