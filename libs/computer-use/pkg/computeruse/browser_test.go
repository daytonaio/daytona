// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestBrowserWaitForReadyTimesOut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not ready", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	manager := testBrowserManager(t, server.URL)
	done := make(chan error)

	_, err := manager.waitForReady(done)
	if err == nil || !strings.Contains(err.Error(), "did not become ready") {
		t.Fatalf("waitForReady error = %v, want readiness timeout", err)
	}
}

func TestBrowserWaitForReadyFailsWhenProcessExits(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	manager := testBrowserManager(t, server.URL)
	done := make(chan error, 1)
	done <- errors.New("exit status 1")

	_, err := manager.waitForReady(done)
	if err == nil || !strings.Contains(err.Error(), "exited before CDP became ready") {
		t.Fatalf("waitForReady error = %v, want process exit", err)
	}
}

func testBrowserManager(t *testing.T, serverURL string) *browserManager {
	t.Helper()
	port, err := strconv.Atoi(serverURL[strings.LastIndex(serverURL, ":")+1:])
	if err != nil {
		t.Fatalf("failed to parse server port from %q: %v", serverURL, err)
	}
	return &browserManager{
		port:         port,
		httpClient:   http.DefaultClient,
		readyTimeout: 20 * time.Millisecond,
		pollInterval: 1 * time.Millisecond,
	}
}
