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

// TestSignedFileDownloadURL validates signed file download URL creation,
// anonymous download, expiry, file-not-found, and stopped-sandbox cases.
func TestSignedFileDownloadURL(t *testing.T) {
	cfg := LoadConfig(t)
	client := NewAPIClient(cfg)
	runID := testRunID()

	createReq := map[string]interface{}{
		"name":   fmt.Sprintf("e2e-signed-dl-%s", runID[4:]),
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
		t.Skip("toolboxProxyUrl not available — skipping signed file download tests")
	}

	baseURL := strings.TrimRight(toolboxProxyURL, "/") + "/" + sandboxID
	httpCli := &http.Client{Timeout: 30 * time.Second}
	// anonCli does not follow redirects so we see raw 401/403 on expired tokens
	anonCli := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	fileContent := fmt.Sprintf("e2e-signed-dl-content-%d", time.Now().UnixNano())
	filePath := "/tmp/e2e-signed-dl.txt"

	t.Run("Setup_UploadFile", func(t *testing.T) {
		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		fw, err := mw.CreateFormFile("file", "e2e-signed-dl.txt")
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
		respBody, _ := io.ReadAll(resp.Body)
		require.Equal(t, http.StatusOK, resp.StatusCode, "file upload must return 200: %s", string(respBody))
	})

	// Case 3 — file-exists precheck fires on getSignedFileDownloadUrl(), not on curl
	t.Run("FileNotFound_Returns404", func(t *testing.T) {
		resp, body := client.DoRequest(t, http.MethodGet,
			fmt.Sprintf("/sandbox/%s/files/signed-download-url?path=/tmp/nonexistent-file-e2e-xyz", sandboxID),
			nil,
		)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode,
			"nonexistent path must return 404 immediately (not on curl): %s", string(body))
	})

	var signedURL, token string

	// Case 1 — happy path: get URL, download anonymously, verify body + Content-Disposition
	t.Run("HappyPath_GetSignedURL", func(t *testing.T) {
		resp, body := client.DoRequest(t, http.MethodGet,
			fmt.Sprintf("/sandbox/%s/files/signed-download-url?path=%s", sandboxID, filePath),
			nil,
		)
		require.Equal(t, http.StatusOK, resp.StatusCode, "get signed URL must return 200: %s", string(body))

		var result map[string]interface{}
		require.NoError(t, json.Unmarshal(body, &result))

		signedURL, _ = result["url"].(string)
		token, _ = result["token"].(string)
		require.NotEmpty(t, signedURL, "signed URL must not be empty")
		require.NotEmpty(t, token, "token must not be empty")
		assert.True(t, strings.HasPrefix(token, "fdl"), "token must start with fdl, got: %s", token)
		t.Logf("signed URL: %s", signedURL)
	})

	if signedURL == "" {
		t.Skip("signed URL not available — skipping proxy-dependent subtests")
	}

	t.Run("HappyPath_AnonymousDownload", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, signedURL, nil)
		require.NoError(t, err)

		resp, err := anonCli.Do(req)
		if err != nil {
			t.Skipf("proxy subdomain unreachable (%v) — skip anonymous download", err)
		}
		defer resp.Body.Close()
		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode,
			"anonymous download must return 200: %s", string(respBody))
		assert.Equal(t, fileContent, strings.TrimSpace(string(respBody)),
			"downloaded content must match uploaded content")
		assert.Contains(t, resp.Header.Get("Content-Disposition"), "attachment",
			"response must have Content-Disposition: attachment")
	})

	// Case 2 — expiry: create a 3s URL, sleep 5s, verify 401
	t.Run("TokenExpiry_Returns401", func(t *testing.T) {
		resp, body := client.DoRequest(t, http.MethodGet,
			fmt.Sprintf("/sandbox/%s/files/signed-download-url?path=%s&expiresInSeconds=3", sandboxID, filePath),
			nil,
		)
		require.Equal(t, http.StatusOK, resp.StatusCode, "short-lived signed URL must return 200: %s", string(body))

		var result map[string]interface{}
		require.NoError(t, json.Unmarshal(body, &result))
		shortURL, _ := result["url"].(string)
		require.NotEmpty(t, shortURL)

		t.Log("waiting 5s for token TTL to expire...")
		time.Sleep(5 * time.Second)

		req, err := http.NewRequest(http.MethodGet, shortURL, nil)
		require.NoError(t, err)
		expResp, err := anonCli.Do(req)
		if err != nil {
			t.Skipf("proxy subdomain unreachable (%v) — skip expiry check", err)
		}
		defer expResp.Body.Close()

		assert.True(t,
			expResp.StatusCode == http.StatusUnauthorized || expResp.StatusCode == http.StatusForbidden,
			"expired token must return 401 or 403, got %d", expResp.StatusCode)
		t.Logf("expiry verified: %d after TTL", expResp.StatusCode)
	})

	_ = token // token available for explicit-expire test if added later

	// Case 4 — stopped sandbox: stop sandbox, verify 409
	t.Run("StoppedSandbox_Returns409", func(t *testing.T) {
		stopResp, stopBody := client.DoRequest(t, http.MethodPost,
			fmt.Sprintf("/sandbox/%s/stop", sandboxID), nil)
		require.Equal(t, http.StatusOK, stopResp.StatusCode,
			"stop sandbox must return 200: %s", string(stopBody))
		client.PollSandboxState(t, sandboxID, "stopped", cfg.PollTimeout, cfg.PollInterval)

		resp, body := client.DoRequest(t, http.MethodGet,
			fmt.Sprintf("/sandbox/%s/files/signed-download-url?path=%s", sandboxID, filePath),
			nil,
		)
		assert.Equal(t, http.StatusConflict, resp.StatusCode,
			"stopped sandbox must return 409: %s", string(body))
	})
}
