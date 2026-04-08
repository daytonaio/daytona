// Copyright 2025 Daytona Platforms Inc.
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

// TestToolboxSmoke validates the toolbox proxy chain: API → proxy → runner → container.
// Creates a sandbox, executes a command, and performs a file write + read roundtrip.
func TestToolboxSmoke(t *testing.T) {
	cfg := LoadConfig(t)
	client := NewAPIClient(cfg)
	runID := testRunID()

	createReq := map[string]interface{}{
		"name":   fmt.Sprintf("e2e-toolbox-%s", runID[4:]),
		"labels": sandboxLabels(runID),
	}
	if cfg.Snapshot != "" {
		createReq["snapshot"] = cfg.Snapshot
	}

	sandbox := client.CreateSandbox(t, createReq)
	sandboxID, _ := sandbox["id"].(string)
	require.NotEmpty(t, sandboxID, "sandbox must have id")

	started := client.PollSandboxState(t, sandboxID, "started", cfg.PollTimeout, cfg.PollInterval)

	toolboxProxyURL, _ := started["toolboxProxyUrl"].(string)
	if toolboxProxyURL == "" {
		t.Skip("toolboxProxyUrl not available — skipping toolbox tests")
	}

	t.Logf("sandbox %s started, toolboxProxyUrl: %s", sandboxID, toolboxProxyURL)

	baseURL := strings.TrimRight(toolboxProxyURL, "/") + "/" + sandboxID
	httpCli := &http.Client{Timeout: 30 * time.Second}

	t.Run("ExecuteCommand", func(t *testing.T) {
		body, err := json.Marshal(map[string]interface{}{
			"command": "echo e2e-test-output",
			"timeout": 10,
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

		require.Equal(t, http.StatusOK, resp.StatusCode, "process/execute must return 200: %s", string(respBody))
		t.Logf("execute response: %s", string(respBody))
		assert.Contains(t, string(respBody), "e2e-test-output", "response must contain command output")

		var execResult map[string]interface{}
		if err := json.Unmarshal(respBody, &execResult); err == nil {
			exitCode, _ := execResult["exitCode"].(float64)
			assert.Equal(t, float64(0), exitCode, "exitCode must be 0 for successful command")
		}
	})

	fileContent := fmt.Sprintf("e2e-test-content-%d", time.Now().UnixNano())
	filePath := "/tmp/e2e-test-file.txt"

	t.Run("FileWrite", func(t *testing.T) {
		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		fw, err := mw.CreateFormFile("file", "e2e-test-file.txt")
		require.NoError(t, err)
		_, err = io.WriteString(fw, fileContent)
		require.NoError(t, err)
		require.NoError(t, mw.Close())

		req, err := http.NewRequest(http.MethodPost, baseURL+"/files/upload?path="+filePath, &buf)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
		req.Header.Set("Content-Type", mw.FormDataContentType())

		resp, err := httpCli.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode, "files/upload must return 200: %s", string(respBody))
		t.Logf("file upload response: %s", string(respBody))
	})

	t.Run("FileRead", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, baseURL+"/files/download?path="+filePath, nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

		resp, err := httpCli.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode, "files/download must return 200: %s", string(respBody))
		assert.Contains(t, string(respBody), fileContent, "downloaded file must contain written content")
		t.Logf("file read verified: content matches")
	})
}
