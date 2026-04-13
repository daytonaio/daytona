// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build e2e

package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// APIClient provides HTTP access to the Daytona API for E2E tests.
type APIClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewAPIClient creates a new API client from the given configuration.
func NewAPIClient(cfg Config) *APIClient {
	return &APIClient{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// DoRequest performs an authenticated HTTP request to the Daytona API.
// Retries on HTTP 429 (rate limit) with exponential backoff.
// Returns the response and body bytes.
func (c *APIClient) DoRequest(t *testing.T, method, path string, body interface{}) (*http.Response, []byte) {
	t.Helper()

	var bodyData []byte
	if body != nil {
		var err error
		bodyData, err = json.Marshal(body)
		require.NoError(t, err, "failed to marshal request body")
	}

	url := c.baseURL + path

	var resp *http.Response
	var respBody []byte

	backoff := time.Second
	for attempt := 1; attempt <= 3; attempt++ {
		var bodyReader io.Reader
		if bodyData != nil {
			bodyReader = bytes.NewReader(bodyData)
		}

		req, err := http.NewRequest(method, url, bodyReader)
		require.NoError(t, err, "failed to create HTTP request")
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err = c.httpClient.Do(req)
		require.NoError(t, err, "HTTP request failed: %s %s", method, url)

		respBody, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		require.NoError(t, err, "failed to read response body")

		if resp.StatusCode != http.StatusTooManyRequests {
			break
		}

		if attempt < 3 {
			t.Logf("rate limited (429), retrying in %s (attempt %d/3)", backoff, attempt)
			time.Sleep(backoff)
			backoff *= 2
		}
	}

	return resp, respBody
}

// GetSandbox fetches a sandbox by ID and returns the parsed JSON body and HTTP status code.
func (c *APIClient) GetSandbox(t *testing.T, sandboxID string) (map[string]interface{}, int) {
	t.Helper()
	resp, body := c.DoRequest(t, http.MethodGet, "/sandbox/"+sandboxID, nil)
	if resp.StatusCode == http.StatusNotFound {
		return nil, http.StatusNotFound
	}
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, resp.StatusCode
	}
	return result, resp.StatusCode
}

// CreateSandbox creates a sandbox via POST /sandbox and immediately registers t.Cleanup to delete it.
func (c *APIClient) CreateSandbox(t *testing.T, req map[string]interface{}) map[string]interface{} {
	t.Helper()
	resp, body := c.DoRequest(t, http.MethodPost, "/sandbox", req)
	require.Equal(t, http.StatusOK, resp.StatusCode, "create sandbox failed: %s", string(body))

	var sandbox map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &sandbox), "failed to parse sandbox response")

	sandboxID, ok := sandbox["id"].(string)
	require.True(t, ok && sandboxID != "", "sandbox response missing id field")

	t.Cleanup(func() {
		c.DeleteSandbox(t, sandboxID)
	})

	return sandbox
}

// DeleteSandbox deletes a sandbox by ID via DELETE /sandbox/{id}.
func (c *APIClient) DeleteSandbox(t *testing.T, sandboxID string) {
	t.Helper()
	resp, body := c.DoRequest(t, http.MethodDelete, "/sandbox/"+sandboxID, nil)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		t.Logf("warning: delete sandbox %s returned %d: %s", sandboxID, resp.StatusCode, string(body))
	} else {
		t.Logf("deleted sandbox %s (status %d)", sandboxID, resp.StatusCode)
	}
}

// PollSandboxState polls GET /sandbox/{id} until the sandbox reaches targetState or timeout.
func (c *APIClient) PollSandboxState(t *testing.T, sandboxID string, targetState string, timeout time.Duration, pollInterval time.Duration) map[string]interface{} {
	t.Helper()

	deadline := time.Now().Add(timeout)
	interval := pollInterval
	maxInterval := 10 * time.Second

	for time.Now().Before(deadline) {
		sandbox, statusCode := c.GetSandbox(t, sandboxID)
		if statusCode == http.StatusNotFound {
			t.Logf("poll %s: sandbox not found (404)", sandboxID)
		} else if sandbox != nil {
			currentState, _ := sandbox["state"].(string)
			t.Logf("poll %s: state=%s (want %s)", sandboxID, currentState, targetState)

			if currentState == "error" || currentState == "build_failed" {
				errReason, _ := sandbox["errorReason"].(string)
				t.Fatalf("sandbox %s entered error state %q: %s", sandboxID, currentState, errReason)
			}

			if currentState == targetState {
				return sandbox
			}
		}

		time.Sleep(interval)
		interval *= 2
		if interval > maxInterval {
			interval = maxInterval
		}
	}

	t.Fatalf("timeout waiting for sandbox %s to reach state %q after %s", sandboxID, targetState, timeout)
	return nil
}

func testRunID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	suffix := make([]byte, 4)
	for i := range suffix {
		suffix[i] = chars[rand.Intn(len(chars))]
	}
	return fmt.Sprintf("e2e-%d-%s", time.Now().Unix(), string(suffix))
}

func sandboxLabels(runID string) map[string]string {
	return map[string]string{
		"e2e":      "true",
		"test-run": runID,
	}
}
