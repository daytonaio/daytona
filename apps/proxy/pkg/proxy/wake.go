// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"context"
	"fmt"
	"net/http"
	"time"

	apiclient "github.com/daytonaio/apiclient"
	log "github.com/sirupsen/logrus"
)

const (
	WakeTimeoutSeconds     = 60
	WakePollIntervalMillis = 1000
	SandboxStateCacheTTL   = 5 * time.Second  // Short TTL for state cache
	SandboxStateStartedTTL = 30 * time.Second // Longer TTL when sandbox is started
	PostStartDelayMillis   = 2000             // Delay after sandbox starts to let services initialize
)

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

// EnsureSandboxStarted proactively checks if a sandbox needs to be woken before proxying.
// It uses a cache to avoid API spam. Returns true if sandbox is ready for proxying.
// If the sandbox needs to be started, it will start it and wait for it to be ready.
func (p *Proxy) EnsureSandboxStarted(ctx context.Context, sandboxId string) bool {
	// Try to get state from cache first
	cachedState, err := p.sandboxStateCache.Get(ctx, sandboxId)
	if err == nil {
		// Cache hit - check if sandbox is already started
		if cachedState.State == string(apiclient.SANDBOXSTATE_STARTED) {
			log.Debugf("Sandbox %s is started (cached), proceeding with proxy", sandboxId)
			return true
		}

		// Sandbox is not started - check if wake-on-http is enabled
		if cachedState.State == string(apiclient.SANDBOXSTATE_STOPPED) {
			if cachedState.WakeOnRequest == string(apiclient.WAKEONREQUEST_HTTP) ||
				cachedState.WakeOnRequest == string(apiclient.WAKEONREQUEST_HTTP_AND_SSH) {
				log.Infof("Sandbox %s is stopped (cached), starting it proactively", sandboxId)
				return p.startAndWaitForSandbox(ctx, sandboxId)
			}
		}

		// Other states (starting, stopping, etc.) - let the request proceed
		// and the error handlers will deal with any issues
		return true
	}

	// Cache miss - fetch from API
	info, _, err := p.apiclient.PreviewAPI.GetSandboxInfo(ctx, sandboxId).Execute()
	if err != nil {
		log.Warnf("Failed to get sandbox info for %s: %v, proceeding with proxy", sandboxId, err)
		return true // Let the request proceed, error handlers will catch issues
	}

	state := info.State
	wakeOnRequest := info.WakeOnRequest

	// Cache the state
	stateInfo := SandboxStateInfo{
		State:         string(state),
		WakeOnRequest: string(wakeOnRequest),
	}

	// Use longer TTL for started sandboxes
	cacheTTL := SandboxStateCacheTTL
	if state == apiclient.SANDBOXSTATE_STARTED {
		cacheTTL = SandboxStateStartedTTL
	}

	if err := p.sandboxStateCache.Set(ctx, sandboxId, stateInfo, cacheTTL); err != nil {
		log.Warnf("Failed to cache sandbox state for %s: %v", sandboxId, err)
	}

	// Check if sandbox needs to be woken
	if state == apiclient.SANDBOXSTATE_STOPPED {
		if wakeOnRequest == apiclient.WAKEONREQUEST_HTTP || wakeOnRequest == apiclient.WAKEONREQUEST_HTTP_AND_SSH {
			log.Infof("Sandbox %s is stopped, starting it proactively (wakeOnRequest: %s)", sandboxId, wakeOnRequest)
			return p.startAndWaitForSandbox(ctx, sandboxId)
		}
	}

	// Sandbox is already started or in another state
	return true
}

// startAndWaitForSandbox starts the sandbox and waits for it to be ready
func (p *Proxy) startAndWaitForSandbox(ctx context.Context, sandboxId string) bool {
	// Start the sandbox
	_, _, err := p.apiclient.SandboxAPI.StartSandbox(ctx, sandboxId).Execute()
	if err != nil {
		log.Errorf("Failed to start sandbox %s: %v", sandboxId, err)
		return false
	}

	// Clear caches since the sandbox is starting
	if err := p.runnerCache.Delete(ctx, sandboxId); err != nil {
		log.Warnf("Failed to clear runner cache for sandbox %s: %v", sandboxId, err)
	}
	if err := p.sandboxStateCache.Delete(ctx, sandboxId); err != nil {
		log.Warnf("Failed to clear state cache for sandbox %s: %v", sandboxId, err)
	}

	// Wait for sandbox to be started
	started := p.waitForSandboxStarted(ctx, sandboxId)

	// Update cache with new state
	if started {
		stateInfo := SandboxStateInfo{
			State:         string(apiclient.SANDBOXSTATE_STARTED),
			WakeOnRequest: "", // Will be refreshed on next check
		}
		if err := p.sandboxStateCache.Set(ctx, sandboxId, stateInfo, SandboxStateStartedTTL); err != nil {
			log.Warnf("Failed to update state cache for %s: %v", sandboxId, err)
		}

		// Add a small delay to let services inside the container initialize
		log.Debugf("Sandbox %s started, waiting %dms for services to initialize", sandboxId, PostStartDelayMillis)
		time.Sleep(PostStartDelayMillis * time.Millisecond)
	}

	return started
}

// tryWakeSandbox attempts to wake a stopped sandbox if wake-on-http is enabled.
// Returns true if the sandbox was woken and is now started, false otherwise.
// Deprecated: Use EnsureSandboxStarted for proactive wake instead.
func (p *Proxy) tryWakeSandbox(ctx context.Context, sandboxId string) bool {
	// Get sandbox info to check state and wake settings
	info, _, err := p.apiclient.PreviewAPI.GetSandboxInfo(ctx, sandboxId).Execute()
	if err != nil {
		log.Errorf("Failed to get sandbox info for %s: %v", sandboxId, err)
		return false
	}

	state := info.State
	wakeOnRequest := info.WakeOnRequest

	// Only wake if sandbox is stopped
	if state != apiclient.SANDBOXSTATE_STOPPED {
		log.Debugf("Sandbox %s is not stopped (state: %s), not waking", sandboxId, state)
		return false
	}

	// Check if wake-on-http is enabled
	if wakeOnRequest != apiclient.WAKEONREQUEST_HTTP && wakeOnRequest != apiclient.WAKEONREQUEST_HTTP_AND_SSH {
		log.Debugf("Sandbox %s does not have wake-on-http enabled (wakeOnRequest: %s)", sandboxId, wakeOnRequest)
		return false
	}

	log.Infof("Waking sandbox %s (state: %s, wakeOnRequest: %s)", sandboxId, state, wakeOnRequest)

	// Start the sandbox
	_, _, err = p.apiclient.SandboxAPI.StartSandbox(ctx, sandboxId).Execute()
	if err != nil {
		log.Errorf("Failed to start sandbox %s: %v", sandboxId, err)
		return false
	}

	// Clear runner cache since the sandbox is starting on a potentially different runner
	if err := p.runnerCache.Delete(ctx, sandboxId); err != nil {
		log.Warnf("Failed to clear runner cache for sandbox %s: %v", sandboxId, err)
	}

	// Poll for sandbox to be started
	return p.waitForSandboxStarted(ctx, sandboxId)
}

// waitForSandboxStarted polls the sandbox state until it's started or timeout
func (p *Proxy) waitForSandboxStarted(ctx context.Context, sandboxId string) bool {
	timeout := time.After(WakeTimeoutSeconds * time.Second)
	ticker := time.NewTicker(WakePollIntervalMillis * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			log.Warnf("Timeout waiting for sandbox %s to start", sandboxId)
			return false
		case <-ctx.Done():
			log.Warnf("Context cancelled while waiting for sandbox %s to start", sandboxId)
			return false
		case <-ticker.C:
			info, _, err := p.apiclient.PreviewAPI.GetSandboxInfo(ctx, sandboxId).Execute()
			if err != nil {
				log.Errorf("Failed to get sandbox info while polling %s: %v", sandboxId, err)
				continue
			}

			state := info.State
			if state == apiclient.SANDBOXSTATE_STARTED {
				log.Infof("Sandbox %s is now started", sandboxId)
				return true
			}

			if state == apiclient.SANDBOXSTATE_ERROR || state == apiclient.SANDBOXSTATE_BUILD_FAILED {
				log.Errorf("Sandbox %s failed to start (state: %s)", sandboxId, state)
				return false
			}

			log.Debugf("Sandbox %s state: %s, waiting...", sandboxId, state)
		}
	}
}

// isWakeableError checks if the error is a connection error that could be resolved by waking the sandbox
func isWakeableError(err error) bool {
	if err == nil {
		return false
	}
	// Connection refused, timeout, or other network errors indicate the sandbox might be stopped
	errStr := err.Error()
	return contains(errStr, "connection refused") ||
		contains(errStr, "no such host") ||
		contains(errStr, "no route to host") ||
		contains(errStr, "i/o timeout") ||
		contains(errStr, "connection reset")
}

// isWakeableStatusCode checks if the HTTP status code indicates the sandbox might be stopped
func isWakeableStatusCode(statusCode int) bool {
	return statusCode == 502 || statusCode == 503 || statusCode == 504
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// WakeOnRequestErrorHandler returns an error handler for the reverse proxy that attempts to wake stopped sandboxes
func (p *Proxy) WakeOnRequestErrorHandler(sandboxId string, originalHandler func(http.ResponseWriter, *http.Request, error)) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		if isWakeableError(err) {
			log.Infof("Detected wakeable error for sandbox %s: %v", sandboxId, err)
			if p.tryWakeSandbox(r.Context(), sandboxId) {
				// Sandbox woke up - return HTML page with auto-refresh
				w.Header().Set("Retry-After", "5")
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(503)
				fmt.Fprint(w, sandboxStartingHTML)
				return
			}
		}
		// Call original handler or default behavior
		if originalHandler != nil {
			originalHandler(w, r, err)
		} else {
			w.WriteHeader(502)
			fmt.Fprintf(w, "Bad Gateway: %v", err)
		}
	}
}

// WakeOnRequestModifyResponse returns a response modifier that attempts to wake stopped sandboxes
// when the upstream returns a 502/503/504 status code
func (p *Proxy) WakeOnRequestModifyResponse(sandboxId string, originalModifyResponse func(*http.Response) error) func(*http.Response) error {
	return func(res *http.Response) error {
		// First call the original modifier if provided
		if originalModifyResponse != nil {
			if err := originalModifyResponse(res); err != nil {
				return err
			}
		}

		// Check if this is a wakeable status code
		if isWakeableStatusCode(res.StatusCode) {
			log.Infof("Detected wakeable status code %d for sandbox %s", res.StatusCode, sandboxId)
			if p.tryWakeSandbox(res.Request.Context(), sandboxId) {
				// Sandbox woke up - modify response to tell client to retry
				res.StatusCode = 503
				res.Status = "503 Service Unavailable"
				res.Header.Set("Retry-After", "5")
				res.Body = http.NoBody
				res.ContentLength = 0
				log.Infof("Sandbox %s woke up, asking client to retry", sandboxId)
			}
		}

		return nil
	}
}
