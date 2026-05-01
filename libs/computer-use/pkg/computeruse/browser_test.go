// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestBrowserWaitForReadyReadsDevToolsActivePort(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"webSocketDebuggerUrl":"ws://127.0.0.1:9222/devtools/browser/abc"}`))
	}))
	defer server.Close()

	manager := testBrowserManager(t, server.URL)
	manager.port = 0
	manager.profileDir = t.TempDir()
	port := server.URL[strings.LastIndex(server.URL, ":")+1:]
	if err := os.WriteFile(filepath.Join(manager.profileDir, "DevToolsActivePort"), []byte(port+"\n/devtools/browser/abc\n"), 0644); err != nil {
		t.Fatalf("failed to write DevToolsActivePort: %v", err)
	}

	got, err := manager.waitForReady(make(chan error))
	if err != nil {
		t.Fatalf("waitForReady failed: %v", err)
	}
	if got != "ws://127.0.0.1:9222/devtools/browser/abc" {
		t.Fatalf("CDP URL = %q", got)
	}
	if strconv.Itoa(manager.port) != port {
		t.Fatalf("manager port = %d, want %s", manager.port, port)
	}
}

func TestEnsurePortAvailableFailsForBoundPort(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to bind test port: %v", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	if err := ensurePortAvailable(port); err == nil {
		t.Fatalf("ensurePortAvailable(%d) succeeded, want failure", port)
	}
}

func TestBrowserPortDefaultsToDynamicPort(t *testing.T) {
	t.Setenv("DAYTONA_BROWSER_CDP_PORT", "")

	port, dynamic := browserPort()

	if port != 0 || !dynamic {
		t.Fatalf("browserPort() = (%d, %v), want (0, true)", port, dynamic)
	}
}

func TestBrowserPortUsesConfiguredPort(t *testing.T) {
	t.Setenv("DAYTONA_BROWSER_CDP_PORT", "9333")

	port, dynamic := browserPort()

	if port != 9333 || dynamic {
		t.Fatalf("browserPort() = (%d, %v), want (9333, false)", port, dynamic)
	}
}

func TestBrowserStatusClearsCDPFieldsWhenStopped(t *testing.T) {
	manager := &browserManager{
		port:        9222,
		localCDPURL: "ws://127.0.0.1:9222/devtools/browser/abc",
	}

	status := manager.status("https://proxy.example.com/toolbox/sandbox-1")

	if status.Running {
		t.Fatal("browser status running = true, want false")
	}
	if status.WebSocketDebuggerURL != "" || status.LocalWebSocketDebuggerURL != "" || status.ProxyPath != "" {
		t.Fatalf("stopped browser status returned CDP fields: %+v", status)
	}
	if manager.localCDPURL != "" {
		t.Fatalf("manager local CDP URL = %q, want cleared", manager.localCDPURL)
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
