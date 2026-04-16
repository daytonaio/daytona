// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	WakeTimeoutSeconds     = 60
	WakePollIntervalMillis = 1000
	PostStartDelayMillis   = 2000
)

const SandboxStateHeader = "X-Daytona-Sandbox-State"

type SandboxStateInfo struct {
	State         *apiclient.SandboxState `json:"state"`
	WakeOnRequest string                  `json:"wakeOnRequest"`
}

const sandboxStartingHTML = `<!DOCTYPE html>
<html>
<head>
    <meta http-equiv="refresh" content="3">
    <title>Sandbox Starting...</title>
    <style>
        body { font-family: system-ui, sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; background: #1a1a2e; color: #eee; }
        .container { text-align: center; }
        .spinner { width: 50px; height: 50px; border: 4px solid #333; border-top: 4px solid #00d4ff; border-radius: 50%; animation: spin 1s linear infinite; margin: 20px auto; }
        @keyframes spin { 0% { transform: rotate(0deg); } 100% { transform: rotate(360deg); } }
    </style>
</head>
<body>
    <div class="container">
        <div class="spinner"></div>
        <h2>Sandbox is starting...</h2>
        <p>This page will automatically refresh in 3 seconds.</p>
    </div>
</body>
</html>`

func (p *Proxy) startAndWaitForSandbox(ctx context.Context, sandboxID string) bool {
	_, _, err := p.apiclient.SandboxAPI.StartSandbox(ctx, sandboxID).Execute()
	if err != nil {
		log.Errorf("Failed to start sandbox %s. Err: %v", sandboxID, err)
		return false
	}

	if err := p.sandboxRunnerCache.Delete(ctx, sandboxID); err != nil {
		log.Warnf("Failed to clear sandbox runner cache for %s: %v", sandboxID, err)
	}

	started := p.waitForSandboxStarted(ctx, sandboxID)
	if started {
		time.Sleep(PostStartDelayMillis * time.Millisecond)
	}

	return started
}

func (p *Proxy) tryWakeSandbox(ctx context.Context, sandboxID string) bool {
	info, err := p.getSandboxStateInfo(ctx, sandboxID)
	if err != nil {
		log.Errorf("Failed to get sandbox info for %s: %v", sandboxID, err)
		return false
	}

	if *info.State != apiclient.SANDBOXSTATE_STOPPED {
		return false
	}

	if info.WakeOnRequest != "http" && info.WakeOnRequest != "http_and_ssh" {
		return false
	}

	return p.startAndWaitForSandbox(ctx, sandboxID)
}

func (p *Proxy) waitForSandboxStarted(ctx context.Context, sandboxID string) bool {
	timeout := time.After(WakeTimeoutSeconds * time.Second)
	ticker := time.NewTicker(WakePollIntervalMillis * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			log.Warnf("Timeout waiting for sandbox %s to start", sandboxID)
			return false
		case <-ctx.Done():
			log.Warnf("Context cancelled while waiting for sandbox %s to start", sandboxID)
			return false
		case <-ticker.C:
			info, err := p.getSandboxStateInfo(ctx, sandboxID)
			if err != nil {
				log.Errorf("Failed to poll sandbox info for %s: %v", sandboxID, err)
				continue
			}

			switch *info.State {
			case apiclient.SANDBOXSTATE_STARTED:
				log.Debugf("Sandbox %s is now started", sandboxID)
				return true
			case apiclient.SANDBOXSTATE_ERROR, apiclient.SANDBOXSTATE_BUILD_FAILED:
				log.Errorf("Sandbox %s failed to start (state: %s)", sandboxID, *info.State)
				return false
			}
		}
	}
}

// resolvedSandboxID returns the resolved sandbox ID from the gin context if available,
// falling back to the provided fallback value. This handles the case where the URL contains
// a signed preview token instead of a real sandbox ID — the real ID is stored in the gin
// context by GetProxyTarget after authentication.
func resolvedSandboxID(ginCtx *gin.Context, fallback string) string {
	if resolved := ginCtx.GetString(RESOLVED_SANDBOX_ID_KEY); resolved != "" {
		return resolved
	}
	return fallback
}

func isWakeableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "no route to host") ||
		strings.Contains(errStr, "i/o timeout") ||
		strings.Contains(errStr, "connection reset")
}

func (p *Proxy) WakeOnRequestErrorHandler(ginCtx *gin.Context, originalHandler func(http.ResponseWriter, *http.Request, error)) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		sandboxID := resolvedSandboxID(ginCtx, "")
		if sandboxID != "" && isWakeableError(err) {
			if p.tryWakeSandbox(r.Context(), sandboxID) {
				w.Header().Set("Retry-After", "5")
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = fmt.Fprint(w, sandboxStartingHTML)
				return
			}
		}

		if originalHandler != nil {
			originalHandler(w, r, err)
			return
		}

		w.WriteHeader(http.StatusBadGateway)
		_, _ = fmt.Fprintf(w, "Bad Gateway: %v", err)
	}
}

func (p *Proxy) WakeOnRequestModifyResponse(ginCtx *gin.Context, originalModifyResponse func(*http.Response) error) func(*http.Response) error {
	return func(res *http.Response) error {
		if originalModifyResponse != nil {
			if err := originalModifyResponse(res); err != nil {
				return err
			}
		}

		sandboxID := resolvedSandboxID(ginCtx, "")
		if sandboxID == "" || res.Header.Get(SandboxStateHeader) != "stopped" {
			return nil
		}

		if p.tryWakeSandbox(res.Request.Context(), sandboxID) {
			res.StatusCode = http.StatusServiceUnavailable
			res.Status = "503 Service Unavailable"
			res.Header.Set("Retry-After", "5")
			res.Header.Set("Content-Type", "text/html; charset=utf-8")
			res.Body = io.NopCloser(strings.NewReader(sandboxStartingHTML))
			res.ContentLength = int64(len(sandboxStartingHTML))
			res.Header.Set("Content-Length", fmt.Sprintf("%d", len(sandboxStartingHTML)))
		}

		return nil
	}
}

func (p *Proxy) getSandboxStateInfo(ctx context.Context, sandboxID string) (*SandboxStateInfo, error) {
	sandbox, _, err := p.apiclient.SandboxAPI.GetSandbox(ctx, sandboxID).Execute()
	if err != nil {
		return nil, err
	}

	info := SandboxStateInfo{
		State:         sandbox.State,
		WakeOnRequest: "none",
	}

	if sandbox.RunnerId == nil {
		return &info, nil
	}

	// Get runner info to determine wake on request setting
	runner, err := p.getRunnerInfo(ctx, *sandbox.RunnerId)
	if err != nil {
		return nil, err
	}

	if runner.RunnerClass == string(apiclient.RUNNERCLASS_VM) {
		info.WakeOnRequest = "http_and_ssh"
	}

	return &info, nil
}
