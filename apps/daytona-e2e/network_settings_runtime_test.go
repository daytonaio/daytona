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

	"github.com/stretchr/testify/require"
)

func TestRuntimeNetworkSettings(t *testing.T) {
	cfg := LoadConfig(t)
	client := NewAPIClient(cfg)
	runID := testRunID()

	createReq := map[string]interface{}{
		"name":   fmt.Sprintf("e2e-network-settings-%s", runID[4:]),
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
		t.Skip("toolboxProxyUrl not available — skipping runtime network settings test")
	}

	baseURL := strings.TrimRight(toolboxProxyURL, "/") + "/" + sandboxID
	httpCli := &http.Client{Timeout: 30 * time.Second}

	googleDNS := "https://8.8.8.8"
	quad9 := "https://9.9.9.9"
	openDNS := "https://208.67.222.222"

	type expectations map[string]bool

	probe := func(t *testing.T, url string) (int, string) {
		t.Helper()
		body, err := json.Marshal(map[string]interface{}{
			// -k: ignore TLS cert hostname mismatch when probing raw IP over HTTPS
			"command": fmt.Sprintf("curl -k -sS -o /dev/null -w '%%{http_code}' --max-time 5 %s", url),
			"timeout": 15,
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

		var execResult map[string]interface{}
		require.NoError(t, json.Unmarshal(respBody, &execResult))

		exitCode := int(execResult["exitCode"].(float64))
		result, _ := execResult["result"].(string)
		return exitCode, strings.TrimSpace(result)
	}

	checkEventually := func(t *testing.T, label string, want expectations) {
		t.Helper()
		deadline := time.Now().Add(45 * time.Second)
		lastState := ""

		for time.Now().Before(deadline) {
			allMatch := true
			var lines []string

			for url, shouldReach := range want {
				exitCode, result := probe(t, url)
				reached := exitCode == 0
				ok := reached == shouldReach
				if !ok {
					allMatch = false
				}

				status := "BLOCKED"
				if reached {
					status = fmt.Sprintf("REACHABLE(http=%s)", result)
				}
				lines = append(lines, fmt.Sprintf("%s => %s expectedReachable=%t", url, status, shouldReach))
			}

			lastState = strings.Join(lines, " | ")
			if allMatch {
				t.Logf("[%s] %s", label, lastState)
				return
			}

			time.Sleep(2 * time.Second)
		}

		t.Fatalf("[%s] expectations not met before timeout: %s", label, lastState)
	}

	collectReachability := func(t *testing.T, urls []string) (map[string]bool, string) {
		t.Helper()
		result := make(map[string]bool, len(urls))
		var lines []string
		for _, url := range urls {
			exitCode, probeResult := probe(t, url)
			reached := exitCode == 0
			result[url] = reached
			status := "BLOCKED"
			if reached {
				status = fmt.Sprintf("REACHABLE(http=%s)", probeResult)
			}
			lines = append(lines, fmt.Sprintf("%s => %s", url, status))
		}
		return result, strings.Join(lines, " | ")
	}

	updateNetworkSettings := func(t *testing.T, payload map[string]interface{}) {
		t.Helper()
		resp, body := client.DoRequest(t, http.MethodPost, "/sandbox/"+sandboxID+"/network-settings", payload)
		if resp.StatusCode == http.StatusNotFound &&
			strings.Contains(string(body), "Cannot POST /api/sandbox/") &&
			strings.Contains(string(body), "/network-settings") {
			t.Skip("runtime network settings endpoint is not available in this environment")
		}
		if resp.StatusCode == http.StatusBadRequest &&
			strings.Contains(string(body), "Network access is restricted and cannot be overridden at the sandbox level") {
			t.Skip("organization has limited network egress enabled; runtime override endpoint is intentionally blocked")
		}
		require.Equal(t, http.StatusOK, resp.StatusCode, "update network settings failed: %s", string(body))
	}

	// Preflight: in some CI environments raw-IP HTTPS probes are blocked by platform policy.
	// Runtime network assertions are meaningful only if baseline outbound probes are reachable.
	baselineURLs := []string{googleDNS, quad9, openDNS}
	baseline, baselineLog := collectReachability(t, baselineURLs)
	if !baseline[googleDNS] || !baseline[quad9] || !baseline[openDNS] {
		t.Skipf("skipping runtime network settings assertions: baseline outbound probes are not fully reachable in this environment: %s", baselineLog)
	}

	t.Run("DefaultAllReachable", func(t *testing.T) {
		checkEventually(t, "default", expectations{
			googleDNS: true,
			quad9:     true,
			openDNS:   true,
		})
	})

	t.Run("AllowListSingleIP", func(t *testing.T) {
		updateNetworkSettings(t, map[string]interface{}{
			"networkAllowList": "8.8.8.8/32",
		})
		checkEventually(t, "allowlist-single", expectations{
			googleDNS: true,
			quad9:     false,
			openDNS:   false,
		})
	})

	t.Run("AllowListTwoIPs", func(t *testing.T) {
		updateNetworkSettings(t, map[string]interface{}{
			"networkAllowList": "8.8.8.8/32,9.9.9.9/32",
		})
		checkEventually(t, "allowlist-two", expectations{
			googleDNS: true,
			quad9:     true,
			openDNS:   false,
		})
	})

	t.Run("BlockAll", func(t *testing.T) {
		updateNetworkSettings(t, map[string]interface{}{
			"networkBlockAll": true,
		})
		checkEventually(t, "block-all", expectations{
			googleDNS: false,
			quad9:     false,
			openDNS:   false,
		})
	})

	t.Run("RestoreAll", func(t *testing.T) {
		updateNetworkSettings(t, map[string]interface{}{
			"networkBlockAll": false,
		})
		checkEventually(t, "restore-all", expectations{
			googleDNS: true,
			quad9:     true,
			openDNS:   true,
		})
	})
}
