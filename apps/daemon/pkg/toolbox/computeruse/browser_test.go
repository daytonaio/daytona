// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	commonproxy "github.com/daytonaio/common-go/pkg/proxy"
	"github.com/gin-gonic/gin"
)

func TestRewriteBrowserCDPURL(t *testing.T) {
	got := RewriteBrowserCDPURL(
		"ws://127.0.0.1:9222/devtools/browser/abc",
		"https://proxy.example.com/toolbox/sandbox-1",
	)
	want := "wss://proxy.example.com/toolbox/sandbox-1/browser/cdp/devtools/browser/abc"
	if got != want {
		t.Fatalf("rewritten URL = %q, want %q", got, want)
	}
}

func TestWrapBrowserCDPHandlerReturnsRewrittenURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/computeruse/browser/cdp", WrapBrowserCDPHandler(func(req *BrowserCDPRequest) (*BrowserCDPResponse, error) {
		return &BrowserCDPResponse{
			WebSocketDebuggerURL:      "ws://127.0.0.1:9222/devtools/browser/abc",
			LocalWebSocketDebuggerURL: "ws://127.0.0.1:9222/devtools/browser/abc",
			Port:                      9222,
		}, nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/computeruse/browser/cdp", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "proxy.example.com")
	req.Header.Set("X-Forwarded-Prefix", "/toolbox/sandbox-1")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body BrowserCDPResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	want := "wss://proxy.example.com/toolbox/sandbox-1/browser/cdp/devtools/browser/abc"
	if body.WebSocketDebuggerURL != want {
		t.Fatalf("webSocketDebuggerUrl = %q, want %q", body.WebSocketDebuggerURL, want)
	}
	if body.ProxyPath != "/browser/cdp/devtools/browser/abc" {
		t.Fatalf("proxyPath = %q", body.ProxyPath)
	}
}

func TestWrapBrowserCDPHandlerUsesToolboxBaseURLHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/computeruse/browser/cdp", WrapBrowserCDPHandler(func(req *BrowserCDPRequest) (*BrowserCDPResponse, error) {
		return &BrowserCDPResponse{
			WebSocketDebuggerURL:      "ws://127.0.0.1:9222/devtools/browser/abc",
			LocalWebSocketDebuggerURL: "ws://127.0.0.1:9222/devtools/browser/abc",
			Port:                      9222,
		}, nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/computeruse/browser/cdp", nil)
	req.Header.Set(commonproxy.DaytonaToolboxBaseURLHeader, "https://proxy.example.com/toolbox/sandbox-1")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body BrowserCDPResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	want := "wss://proxy.example.com/toolbox/sandbox-1/browser/cdp/devtools/browser/abc"
	if body.WebSocketDebuggerURL != want {
		t.Fatalf("webSocketDebuggerUrl = %q, want %q", body.WebSocketDebuggerURL, want)
	}
}

func TestWrapBrowserStatusHandlerRewritesLocalhostURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/computeruse/browser/status", WrapBrowserStatusHandler(func() (*BrowserStatusResponse, error) {
		return &BrowserStatusResponse{
			Status:                    "running",
			Running:                   true,
			WebSocketDebuggerURL:      "ws://localhost:9222/devtools/browser/abc",
			LocalWebSocketDebuggerURL: "ws://localhost:9222/devtools/browser/abc",
			Port:                      9222,
		}, nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/computeruse/browser/status", nil)
	req.Header.Set(commonproxy.DaytonaToolboxBaseURLHeader, "https://proxy.example.com/toolbox/sandbox-1")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body BrowserStatusResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	want := "wss://proxy.example.com/toolbox/sandbox-1/browser/cdp/devtools/browser/abc"
	if body.WebSocketDebuggerURL != want {
		t.Fatalf("webSocketDebuggerUrl = %q, want %q", body.WebSocketDebuggerURL, want)
	}
}
