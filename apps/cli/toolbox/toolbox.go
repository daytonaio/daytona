// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package toolbox

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/daytonaio/daytona/cli/config"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/gorilla/websocket"
	"golang.org/x/term"
)

// ExitCodeError is returned when the remote TTY process exits with a non-zero exit code.
// It carries the exit code so callers can propagate it without printing an error message.
type ExitCodeError struct {
	Code int
}

func (e *ExitCodeError) Error() string {
	return fmt.Sprintf("exit code %d", e.Code)
}

type ExecuteRequest struct {
	Command string   `json:"command"`
	Cwd     *string  `json:"cwd,omitempty"`
	Timeout *float32 `json:"timeout,omitempty"`
	TTY     *bool    `json:"tty,omitempty"`
}

type ExecuteResponse struct {
	ExitCode float32 `json:"exitCode"`
	Result   string  `json:"result"`
}

// PTYCreateRequest mirrors the daemon's PTYCreateRequest for the /process/pty endpoint.
type PTYCreateRequest struct {
	ID      string            `json:"id"`
	Command *string           `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Cwd     string            `json:"cwd,omitempty"`
	Timeout *uint32           `json:"timeout,omitempty"`
	Cols    *uint16           `json:"cols,omitempty"`
	Rows    *uint16           `json:"rows,omitempty"`
	Envs    map[string]string `json:"envs,omitempty"`
}

// PTYCreateResponse is the response from creating a PTY session.
type PTYCreateResponse struct {
	SessionID string `json:"sessionId"`
}

type Client struct {
	apiClient *apiclient.APIClient
}

func NewClient(apiClient *apiclient.APIClient) *Client {
	return &Client{
		apiClient: apiClient,
	}
}

// Gets the toolbox proxy URL for a sandbox, caching by region in config
func (c *Client) getProxyURL(ctx context.Context, sandboxId, region string) (string, error) {
	// Check config cache first
	cachedURL, err := config.GetToolboxProxyUrl(region)
	if err == nil && cachedURL != "" {
		return cachedURL, nil
	}

	// Fetch from API
	toolboxProxyUrl, _, err := c.apiClient.SandboxAPI.GetToolboxProxyUrl(ctx, sandboxId).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get toolbox proxy URL: %w", err)
	}

	// Best-effort caching
	_ = config.SetToolboxProxyUrl(region, toolboxProxyUrl.Url)

	return toolboxProxyUrl.Url, nil
}

// defaultHTTPTimeout is used for all outbound HTTP requests to the daemon.
const defaultHTTPTimeout = 30 * time.Second

// getAuthHeaders reads the active profile from config once and returns the
// corresponding HTTP headers (Authorization + optional Org ID). Callers should
// call this once per user-facing operation and pass the resulting headers down
// to avoid redundant config file reads.
func getAuthHeaders() (http.Header, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	activeProfile, err := cfg.GetActiveProfile()
	if err != nil {
		return nil, fmt.Errorf("failed to get active profile: %w", err)
	}

	h := http.Header{}
	if activeProfile.Api.Key != nil {
		h.Set("Authorization", "Bearer "+*activeProfile.Api.Key)
	} else if activeProfile.Api.Token != nil {
		h.Set("Authorization", "Bearer "+activeProfile.Api.Token.AccessToken)
	}
	if activeProfile.ActiveOrganizationId != nil {
		h.Set("X-Daytona-Organization-ID", *activeProfile.ActiveOrganizationId)
	}
	return h, nil
}

func (c *Client) ExecuteCommand(ctx context.Context, sandbox *apiclient.Sandbox, request ExecuteRequest) (*ExecuteResponse, error) {
	proxyURL, err := c.getProxyURL(ctx, sandbox.Id, sandbox.Target)
	if err != nil {
		return nil, err
	}

	return c.executeCommandViaProxy(ctx, proxyURL, sandbox.Id, request)
}

// TODO: replace this with the toolbox api client at some point
func (c *Client) executeCommandViaProxy(ctx context.Context, proxyURL, sandboxId string, request ExecuteRequest) (*ExecuteResponse, error) {
	// Build the URL: {proxyUrl}/{sandboxId}/process/execute
	url := fmt.Sprintf("%s/%s/process/execute", strings.TrimSuffix(proxyURL, "/"), sandboxId)

	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	auth, err := getAuthHeaders()
	if err != nil {
		return nil, err
	}
	for k, v := range auth {
		req.Header[k] = v
	}

	client := &http.Client{Timeout: defaultHTTPTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var response ExecuteResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// ExecuteCommandTTY creates a PTY session for the given command and streams it interactively.
func (c *Client) ExecuteCommandTTY(ctx context.Context, sandbox *apiclient.Sandbox, request PTYCreateRequest) error {
	proxyURL, err := c.getProxyURL(ctx, sandbox.Id, sandbox.Target)
	if err != nil {
		return err
	}

	// Load auth headers once and reuse across all sub-calls.
	auth, err := getAuthHeaders()
	if err != nil {
		return err
	}

	// Create PTY session
	sessionID, err := c.createPTYSession(ctx, proxyURL, sandbox.Id, request, auth)
	if err != nil {
		return err
	}

	// Best-effort cleanup when the session ends.
	defer c.deletePTYSession(proxyURL, sandbox.Id, sessionID, auth)

	// Connect to the session as an interactive terminal
	return c.connectAndStreamPTY(ctx, proxyURL, sandbox.Id, sessionID, auth)
}

// createPTYSession creates a new PTY session via the daemon's /process/pty endpoint.
func (c *Client) createPTYSession(ctx context.Context, proxyURL, sandboxId string, request PTYCreateRequest, auth http.Header) (string, error) {
	url := fmt.Sprintf("%s/%s/process/pty", strings.TrimSuffix(proxyURL, "/"), sandboxId)

	body, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range auth {
		req.Header[k] = v
	}

	client := &http.Client{Timeout: defaultHTTPTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to create PTY session: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to create PTY session: status %d: %s", resp.StatusCode, string(respBody))
	}

	var response PTYCreateResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", fmt.Errorf("failed to parse PTY session response: %w", err)
	}

	return response.SessionID, nil
}

// connectAndStreamPTY connects to a PTY session via WebSocket and streams stdin/stdout.
func (c *Client) connectAndStreamPTY(ctx context.Context, proxyURL, sandboxId, sessionID string, auth http.Header) error {
	base := strings.TrimSuffix(proxyURL, "/")
	wsBase := strings.Replace(base, "https://", "wss://", 1)
	wsBase = strings.Replace(wsBase, "http://", "ws://", 1)

	wsURL := fmt.Sprintf("%s/%s/process/pty/%s/connect", wsBase, sandboxId, sessionID)

	// Dial WebSocket
	dialer := websocket.Dialer{
		HandshakeTimeout: 30 * time.Second,
	}
	ws, _, err := dialer.DialContext(ctx, wsURL, auth)
	if err != nil {
		return fmt.Errorf("failed to connect to PTY session: %w", err)
	}
	defer ws.Close()

	// Ensure stdin and stdout are attached to a TTY before enabling raw mode.
	stdinFd := int(os.Stdin.Fd())
	stdoutFd := int(os.Stdout.Fd())
	if !term.IsTerminal(stdinFd) || !term.IsTerminal(stdoutFd) {
		return fmt.Errorf("--tty requires an interactive terminal on both stdin and stdout (are you piping input or running in CI?)")
	}

	// Set up raw terminal mode
	oldState, err := term.MakeRaw(stdinFd)
	if err != nil {
		return fmt.Errorf("failed to setup terminal: %w", err)
	}
	defer term.Restore(stdinFd, oldState)

	// Get initial terminal size and send to server
	cols, rows, err := term.GetSize(stdoutFd)
	if err != nil {
		cols = 80
		rows = 24
	}
	if err := c.resizePTYSession(ctx, proxyURL, sandboxId, sessionID, uint16(cols), uint16(rows), auth); err != nil {
		slog.Debug("initial PTY resize failed", "error", err)
	}

	// Handle terminal resizing (platform-specific: SIGWINCH on Unix, no-op on Windows)
	stopResizeHandler := setupResizeHandler(ctx, proxyURL, sandboxId, sessionID, c, auth)
	defer stopResizeHandler()

	done := make(chan error, 2)

	// Handle termination signals.
	intChan := make(chan os.Signal, 1)
	signal.Notify(intChan, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		signal.Stop(intChan)
		close(intChan)
	}()
	go func() {
		if sig, ok := <-intChan; ok {
			switch sig {
			case syscall.SIGINT:
				ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
				_ = ws.WriteMessage(websocket.BinaryMessage, []byte{3})
			case syscall.SIGTERM:
				_ = ws.Close()
			}
		}
	}()

	// Read from stdin and write to WebSocket.
	quit := make(chan struct{})
	go func() {
		buffer := make([]byte, 4096)
		for {
			n, err := os.Stdin.Read(buffer)
			if err != nil {
				if err != io.EOF {
					select {
					case done <- err:
					default:
					}
				}
				return
			}
			if n > 0 {
				select {
				case <-quit:
					return
				default:
				}
				data := make([]byte, n)
				copy(data, buffer[:n])
				ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := ws.WriteMessage(websocket.BinaryMessage, data); err != nil {
					select {
					case done <- err:
					default:
					}
					return
				}
			}
		}
	}()

	// Read from WebSocket and write to stdout.
	// The existing PTY sends exit codes via WebSocket close frames (not text control messages).
	go func() {
		for {
			msgType, data, err := ws.ReadMessage()
			if err != nil {
				// Check for WebSocket close frame carrying exit code info.
				var closeErr *websocket.CloseError
				if errors.As(err, &closeErr) {
					var closeData struct {
						ExitCode int `json:"exitCode"`
					}
					if jsonErr := json.Unmarshal([]byte(closeErr.Text), &closeData); jsonErr == nil && closeData.ExitCode != 0 {
						done <- &ExitCodeError{Code: closeData.ExitCode}
						return
					}
					done <- nil
					return
				}
				// Any other error (network drop, etc.) — propagate so callers can fail appropriately.
				done <- err
				return
			}

			// Filter control messages (e.g. "connected") sent as TextMessage.
			if msgType == websocket.TextMessage {
				var ctrl struct {
					Type   string `json:"type"`
					Status string `json:"status"`
					Error  string `json:"error"`
				}
				if json.Unmarshal(data, &ctrl) == nil && ctrl.Type == "control" {
					if ctrl.Status == "error" {
						done <- fmt.Errorf("PTY session error: %s", ctrl.Error)
						return
					}
					// "connected" or other informational — skip
					continue
				}
			}

			os.Stdout.Write(data)
		}
	}()

	err = <-done
	close(quit)
	return err
}

// resizePTYSession sends a resize request to the PTY session.
func (c *Client) resizePTYSession(ctx context.Context, proxyURL, sandboxId, sessionID string, cols, rows uint16, auth http.Header) error {
	url := fmt.Sprintf("%s/%s/process/pty/%s/resize", strings.TrimSuffix(proxyURL, "/"), sandboxId, sessionID)

	req := map[string]uint16{
		"cols": cols,
		"rows": rows,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	for k, v := range auth {
		httpReq.Header[k] = v
	}

	client := &http.Client{Timeout: defaultHTTPTimeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resize request failed: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}

// deletePTYSession attempts to delete the PTY session (best-effort cleanup).
func (c *Client) deletePTYSession(proxyURL, sandboxId, sessionId string, auth http.Header) {
	url := fmt.Sprintf("%s/%s/process/pty/%s", strings.TrimSuffix(proxyURL, "/"), sandboxId, sessionId)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return
	}
	for k, v := range auth {
		req.Header[k] = v
	}
	client := &http.Client{Timeout: defaultHTTPTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}
