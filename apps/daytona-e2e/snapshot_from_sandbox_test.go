// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build e2e

package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// snapshotPollTimeout bounds how long we wait for a sandbox to leave the
// SNAPSHOTTING state. Commit + push to the internal registry can take a
// while for non-trivial container filesystems, so this is intentionally
// generous.
const snapshotPollTimeout = 10 * time.Minute

// TestDockerFilesystemSnapshot exercises the full snapshot-from-sandbox flow
// for a Docker (CONTAINER-class) sandbox:
//
//  1. Create a sandbox and wait until it is STARTED.
//  2. Write a marker file into its filesystem via the toolbox.
//  3. POST /sandbox/{id}/snapshot to commit the container, push the
//     resulting image to the internal registry, and persist a Snapshot row.
//  4. Poll the sandbox until it leaves SNAPSHOTTING and returns to STARTED.
//  5. Validate the persisted Snapshot row (state=active, size>0, ref set).
//  6. Create a second sandbox from that snapshot and verify the marker
//     file is still present, proving the filesystem state was captured.
//
// The snapshot endpoint is gated behind the SANDBOX_LINUX_VM feature flag.
// `@openfeature/nestjs-sdk`'s @RequireFlagsEnabled decorator throws a
// NotFoundException with the message `Cannot ${method} ${url}` when the
// flag is off, so a 404 with that body shape from this endpoint is
// interpreted as a flag-disabled skip rather than a route-missing failure.
func TestDockerFilesystemSnapshot(t *testing.T) {
	cfg := LoadConfig(t)
	client := NewAPIClient(cfg)
	runID := testRunID()

	snapshotName := fmt.Sprintf("e2e-fs-snap-%s", runID[4:])
	markerPath := "/tmp/e2e-fs-snapshot-marker.txt"
	markerContent := fmt.Sprintf("e2e-fs-snapshot-content-%d", time.Now().UnixNano())

	// ------------------------------------------------------------------
	// Source sandbox: create, wait, write marker file
	// ------------------------------------------------------------------

	createReq := map[string]interface{}{
		"name":   fmt.Sprintf("e2e-fs-snap-src-%s", runID[4:]),
		"labels": sandboxLabels(runID),
	}
	if cfg.Snapshot != "" {
		createReq["snapshot"] = cfg.Snapshot
	}

	srcSandbox := client.CreateSandbox(t, createReq)
	srcSandboxID, _ := srcSandbox["id"].(string)
	require.NotEmpty(t, srcSandboxID, "source sandbox must have id")

	srcStarted := client.PollSandboxState(t, srcSandboxID, "started", cfg.PollTimeout, cfg.PollInterval)
	srcToolboxURL, _ := srcStarted["toolboxProxyUrl"].(string)
	require.NotEmpty(t, srcToolboxURL, "source sandbox must expose toolboxProxyUrl")

	srcBaseURL := strings.TrimRight(srcToolboxURL, "/") + "/" + srcSandboxID
	httpCli := &http.Client{Timeout: 60 * time.Second}

	t.Logf("source sandbox %s ready; writing marker file %s", srcSandboxID, markerPath)
	uploadFile(t, httpCli, cfg, srcBaseURL, markerPath, markerContent)

	// ------------------------------------------------------------------
	// Trigger snapshot
	// ------------------------------------------------------------------

	snapResp, snapBody := client.DoRequest(t, http.MethodPost,
		"/sandbox/"+srcSandboxID+"/snapshot",
		map[string]interface{}{"name": snapshotName},
	)

	switch snapResp.StatusCode {
	case http.StatusOK:
		// continue
	case http.StatusNotFound:
		// @openfeature/nestjs-sdk's @RequireFlagsEnabled throws a
		// NotFoundException("Cannot ${method} ${url}") when a required
		// flag is disabled, so the SANDBOX_LINUX_VM-gated endpoint
		// returns this 404 in environments where OpenFeature is not
		// configured to enable the flag (e.g. the default e2e setup).
		// Genuine "no such route" 404s share the same body shape, so
		// treat both as a skip; misregistered routes would still be
		// caught by API unit tests.
		t.Skipf("snapshot endpoint not available (likely SANDBOX_LINUX_VM disabled): %s",
			string(snapBody))
	case http.StatusForbidden:
		t.Skipf("snapshot endpoint forbidden: %s", string(snapBody))
	case http.StatusUnprocessableEntity:
		// The runner doesn't support snapshotting (non-VM/non-CONTAINER
		// class). Treated as an environment skip rather than a failure.
		t.Skipf("snapshot endpoint reported unsupported runner class: %s", string(snapBody))
	default:
		t.Fatalf("POST /sandbox/%s/snapshot returned %d: %s",
			srcSandboxID, snapResp.StatusCode, string(snapBody))
	}

	// Always best-effort delete the snapshot, even if later assertions fail.
	t.Cleanup(func() {
		deleteSnapshotByName(t, client, snapshotName)
	})

	// Sandbox enters SNAPSHOTTING and returns to its previous state on
	// completion. Wait for STARTED with the longer snapshot timeout.
	snapshotStart := time.Now()
	srcAfter := client.PollSandboxState(t, srcSandboxID, "started",
		snapshotPollTimeout, cfg.PollInterval)
	require.NotNil(t, srcAfter, "source sandbox must return to started after snapshot")
	t.Logf("source sandbox %s back to STARTED in %s", srcSandboxID, time.Since(snapshotStart))

	// ------------------------------------------------------------------
	// Validate persisted Snapshot row
	// ------------------------------------------------------------------

	snapshot := waitForSnapshotActive(t, client, snapshotName, 2*time.Minute, 2*time.Second)

	state, _ := snapshot["state"].(string)
	require.Equal(t, "active", state, "snapshot must be ACTIVE; got %q", state)

	snapshotID, _ := snapshot["id"].(string)
	require.NotEmpty(t, snapshotID, "snapshot row must have an id")

	if size, ok := snapshot["size"].(float64); ok {
		assert.Greater(t, size, float64(0),
			"snapshot size must be > 0 GB (got %v)", size)
	} else {
		t.Errorf("snapshot.size must be a non-null number, got %T %v",
			snapshot["size"], snapshot["size"])
	}

	if ref, ok := snapshot["ref"].(string); ok {
		assert.NotEmpty(t, ref, "snapshot.ref must be set after persist")
		// Sandbox-derived snapshots are pushed to the internal registry
		// under a `daytona-{hash}:daytona` tag.
		assert.Contains(t, ref, ":daytona",
			"snapshot.ref should reference the canonical daytona tag, got %q", ref)
	} else {
		t.Errorf("snapshot.ref must be a string, got %T %v",
			snapshot["ref"], snapshot["ref"])
	}

	t.Logf("snapshot %q persisted: id=%s state=%s size=%v ref=%v",
		snapshotName, snapshotID, state, snapshot["size"], snapshot["ref"])

	// ------------------------------------------------------------------
	// Derived sandbox: create from snapshot, verify marker file is intact
	// ------------------------------------------------------------------

	derivedReq := map[string]interface{}{
		"name":     fmt.Sprintf("e2e-fs-snap-dst-%s", runID[4:]),
		"snapshot": snapshotName,
		"labels":   sandboxLabels(runID),
	}

	derivedSandbox := client.CreateSandbox(t, derivedReq)
	derivedID, _ := derivedSandbox["id"].(string)
	require.NotEmpty(t, derivedID, "derived sandbox must have id")

	derivedStarted := client.PollSandboxState(t, derivedID, "started",
		cfg.PollTimeout, cfg.PollInterval)
	derivedToolboxURL, _ := derivedStarted["toolboxProxyUrl"].(string)
	require.NotEmpty(t, derivedToolboxURL, "derived sandbox must expose toolboxProxyUrl")

	derivedBaseURL := strings.TrimRight(derivedToolboxURL, "/") + "/" + derivedID

	t.Run("MarkerFileSurvivesSnapshot", func(t *testing.T) {
		downloaded := downloadFile(t, httpCli, cfg, derivedBaseURL, markerPath)
		assert.Contains(t, downloaded, markerContent,
			"marker file written before snapshot must be present in the derived sandbox; got %q",
			downloaded)
	})
}

// uploadFile writes content to path inside the sandbox via the toolbox
// /files/upload endpoint. Fails the test on any non-200 response.
func uploadFile(t *testing.T, httpCli *http.Client, cfg Config, baseURL, path, content string) {
	t.Helper()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("file", "marker.txt")
	require.NoError(t, err)
	_, err = io.WriteString(fw, content)
	require.NoError(t, err)
	require.NoError(t, mw.Close())

	req, err := http.NewRequest(http.MethodPost, baseURL+"/files/upload?path="+path, &buf)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := httpCli.Do(req)
	require.NoError(t, err, "files/upload request failed")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"files/upload must return 200: %s", string(body))
}

// downloadFile reads a file from the sandbox via the toolbox /files/download
// endpoint and returns its contents as a string.
func downloadFile(t *testing.T, httpCli *http.Client, cfg Config, baseURL, path string) string {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, baseURL+"/files/download?path="+path, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := httpCli.Do(req)
	require.NoError(t, err, "files/download request failed")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode,
		"files/download must return 200: %s", string(body))
	return string(body)
}

// waitForSnapshotActive polls GET /snapshots/{nameOrId} until the snapshot
// row reports state=active or the timeout elapses. Treats 404 as "not
// persisted yet" and keeps polling.
func waitForSnapshotActive(t *testing.T, client *APIClient, nameOrID string, timeout, interval time.Duration) map[string]interface{} {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, body := client.DoRequest(t, http.MethodGet, "/snapshots/"+nameOrID, nil)

		if resp.StatusCode == http.StatusNotFound {
			t.Logf("snapshot %q not yet persisted (404), retrying", nameOrID)
			time.Sleep(interval)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET /snapshots/%s returned %d: %s",
				nameOrID, resp.StatusCode, string(body))
		}

		var snapshot map[string]interface{}
		require.NoError(t, json.Unmarshal(body, &snapshot),
			"failed to parse snapshot response: %s", string(body))

		state, _ := snapshot["state"].(string)
		if state == "active" {
			return snapshot
		}

		// `error` or `build_failed` are terminal failure states for the
		// snapshot lifecycle - bail out instead of waiting for a timeout.
		if state == "error" || state == "build_failed" {
			errReason, _ := snapshot["errorReason"].(string)
			t.Fatalf("snapshot %q entered terminal failure state %q: %s",
				nameOrID, state, errReason)
		}

		t.Logf("snapshot %q state=%s, waiting for active", nameOrID, state)
		time.Sleep(interval)
	}

	t.Fatalf("snapshot %q did not become active within %s", nameOrID, timeout)
	return nil
}

// deleteSnapshotByName resolves a snapshot's UUID via GET /snapshots/{name}
// and then deletes it via DELETE /snapshots/{uuid}. Best-effort: failures
// are logged but do not fail the test (this is invoked from t.Cleanup).
func deleteSnapshotByName(t *testing.T, client *APIClient, name string) {
	t.Helper()

	resp, body := client.DoRequest(t, http.MethodGet, "/snapshots/"+name, nil)
	if resp.StatusCode == http.StatusNotFound {
		t.Logf("snapshot %q already gone (404) at cleanup", name)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Logf("warning: GET /snapshots/%s during cleanup returned %d: %s",
			name, resp.StatusCode, string(body))
		return
	}

	var snapshot map[string]interface{}
	if err := json.Unmarshal(body, &snapshot); err != nil {
		t.Logf("warning: failed to parse snapshot response during cleanup: %v", err)
		return
	}

	id, _ := snapshot["id"].(string)
	if id == "" {
		t.Logf("warning: snapshot %q returned no id during cleanup; body=%s",
			name, string(body))
		return
	}

	delResp, delBody := client.DoRequest(t, http.MethodDelete, "/snapshots/"+id, nil)
	if delResp.StatusCode != http.StatusOK && delResp.StatusCode != http.StatusNotFound {
		t.Logf("warning: DELETE /snapshots/%s returned %d: %s",
			id, delResp.StatusCode, string(delBody))
		return
	}
	t.Logf("deleted snapshot %q (id=%s, status=%d)", name, id, delResp.StatusCode)
}
