// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package toolbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/daytonaio/daytona/cli/config"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/gorilla/websocket"
	"golang.org/x/term"
)

type ExecuteRequest struct {
	Command string   `json:"command"`
	Cwd     *string  `json:"cwd,omitempty"`
	Timeout *float32 `json:"timeout,omitempty"`
	Tty     bool     `json:"tty,omitempty"`
}

type TTYExecuteRequest struct {
	Command string  `json:"command"`
	Cwd     *string `json:"cwd,omitempty"`
	Cols    int     `json:"cols"`
	Rows    int     `json:"rows"`
}

type ExecuteResponse struct {
	ExitCode float32 `json:"exitCode"`
	Result   string  `json:"result"`
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

	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	activeProfile, err := cfg.GetActiveProfile()
	if err != nil {
		return nil, err
	}

	if activeProfile.Api.Key != nil {
		req.Header.Set("Authorization", "Bearer "+*activeProfile.Api.Key)
	} else if activeProfile.Api.Token != nil {
		req.Header.Set("Authorization", "Bearer "+activeProfile.Api.Token.AccessToken)
	}

	if activeProfile.ActiveOrganizationId != nil {
		req.Header.Set("X-Daytona-Organization-ID", *activeProfile.ActiveOrganizationId)
	}

	client := &http.Client{}
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

func (c *Client) ExecuteCommandWithTTY(ctx context.Context, sandbox *apiclient.Sandbox, request TTYExecuteRequest) error {
	proxyURL, err := c.getProxyURL(ctx, sandbox.Id, sandbox.Target)
	if err != nil {
		return err
	}

	return c.executeCommandViaTTY(ctx, proxyURL, sandbox.Id, request)
}

func (c *Client) executeCommandViaTTY(ctx context.Context, proxyURL, sandboxId string, request TTYExecuteRequest) error {
	// Convert HTTP URL to WebSocket URL
	wsURL, err := c.buildWebSocketURL(proxyURL, sandboxId)
	if err != nil {
		return fmt.Errorf("failed to build WebSocket URL: %w", err)
	}

	// Set up WebSocket connection with headers
	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")

	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}

	activeProfile, err := cfg.GetActiveProfile()
	if err != nil {
		return err
	}

	if activeProfile.Api.Key != nil {
		headers.Set("Authorization", "Bearer "+*activeProfile.Api.Key)
	} else if activeProfile.Api.Token != nil {
		headers.Set("Authorization", "Bearer "+activeProfile.Api.Token.AccessToken)
	}

	if activeProfile.ActiveOrganizationId != nil {
		headers.Set("X-Daytona-Organization-ID", *activeProfile.ActiveOrganizationId)
	}

	// Connect to WebSocket with timeout
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 30 * time.Second
	ws, _, err := dialer.Dial(wsURL, headers)
	if err != nil {
		return fmt.Errorf("failed to connect to TTY WebSocket: %w", err)
	}
	defer ws.Close()

	// Send initial TTY request
	if err := ws.WriteJSON(request); err != nil {
		return fmt.Errorf("failed to send TTY request: %w", err)
	}

	// Wait for connection confirmation
	if err := c.waitForTTYConnection(ws); err != nil {
		return err
	}

	// Handle TTY interaction
	return c.handleTTYSession(ctx, ws)
}

func (c *Client) buildWebSocketURL(proxyURL, sandboxId string) (string, error) {
	baseURL := strings.TrimSuffix(proxyURL, "/")

	// Parse the URL to convert HTTP to WebSocket
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	// Convert scheme
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	default:
		return "", fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}

	// Build the WebSocket path
	u.Path = fmt.Sprintf("/%s/process/exec/tty", sandboxId)

	return u.String(), nil
}

func (c *Client) waitForTTYConnection(ws *websocket.Conn) error {
	// Set a timeout for connection establishment
	ws.SetReadDeadline(time.Now().Add(10 * time.Second))
	defer ws.SetReadDeadline(time.Time{})

	for {
		messageType, data, err := ws.ReadMessage()
		if err != nil {
			return fmt.Errorf("failed to read connection message: %w", err)
		}

		if messageType == websocket.TextMessage {
			// Try to parse as control message
			var ctrl map[string]interface{}
			if err := json.Unmarshal(data, &ctrl); err == nil {
				if msgType, ok := ctrl["type"].(string); ok && msgType == "control" {
					if status, ok := ctrl["status"].(string); ok {
						if status == "connected" {
							return nil // Connection established
						}
						if status == "error" {
							if errMsg, ok := ctrl["error"].(string); ok {
								return fmt.Errorf("TTY connection error: %s", errMsg)
							}
							return fmt.Errorf("TTY connection error")
						}
					}
				}
			}
		}
	}
}

func (c *Client) handleTTYSession(ctx context.Context, ws *websocket.Conn) error {
	// Set terminal to raw mode
	oldState, err := c.setRawMode()
	if err != nil {
		return fmt.Errorf("failed to set terminal to raw mode: %w", err)
	}
	defer c.restoreTerminal(oldState)

	// Set up signal handling for resize and interruption
	sigChan := make(chan os.Signal, 1)
	c.setupSignalHandling(sigChan, ws)
	defer signal.Stop(sigChan)

	// Create channels for coordination
	done := make(chan struct{})
	var wg sync.WaitGroup

	// Start input forwarding (stdin -> WebSocket)
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.forwardInput(ws, done)
	}()

	// Start output handling (WebSocket -> stdout)
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.handleOutput(ws, done)
	}()

	// Wait for context cancellation or session end
	select {
	case <-ctx.Done():
		close(done)
		// Give goroutines some time to clean up
		timeout := time.After(2 * time.Second)
		finished := make(chan struct{})
		go func() {
			wg.Wait()
			close(finished)
		}()
		select {
		case <-finished:
		case <-timeout:
			// Force close WebSocket if cleanup is taking too long
			ws.Close()
		}
	case <-done:
		// Session ended naturally
		wg.Wait()
	}

	return nil
}

func (c *Client) setRawMode() (*term.State, error) {
	// Check if stdin is a terminal
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return nil, fmt.Errorf("stdin is not a terminal - TTY mode requires an interactive terminal")
	}

	// Set terminal to raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, fmt.Errorf("failed to set terminal to raw mode: %w", err)
	}

	return oldState, nil
}

func (c *Client) restoreTerminal(oldState *term.State) {
	if oldState != nil {
		if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
			// Best effort to restore, log error but don't fail
			fmt.Fprintf(os.Stderr, "Warning: failed to restore terminal state: %v\n", err)
		}
	}
}

func (c *Client) forwardInput(ws *websocket.Conn, done chan struct{}) {
	buffer := make([]byte, 1024)

	for {
		select {
		case <-done:
			return
		default:
			// Read from stdin with timeout to prevent blocking indefinitely
			n, err := os.Stdin.Read(buffer)
			if err != nil {
				if err == io.EOF {
					// EOF on stdin, close connection gracefully
					ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				}
				return
			}

			// Forward to WebSocket
			if err := ws.WriteMessage(websocket.BinaryMessage, buffer[:n]); err != nil {
				// WebSocket write failed, connection likely closed
				return
			}
		}
	}
}

func (c *Client) handleOutput(ws *websocket.Conn, done chan struct{}) {
	for {
		select {
		case <-done:
			return
		default:
			messageType, data, err := ws.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					// Parse close message for exit code
					if closeErr, ok := err.(*websocket.CloseError); ok {
						c.handleCloseMessage(closeErr.Text)
					}
				}
				close(done) // Signal session end
				return
			}

			switch messageType {
			case websocket.BinaryMessage:
				// Terminal output - write to stdout
				os.Stdout.Write(data)
			case websocket.TextMessage:
				// Could be control message or text output
				var ctrl map[string]interface{}
				if err := json.Unmarshal(data, &ctrl); err == nil {
					if msgType, ok := ctrl["type"].(string); ok && msgType == "control" {
						// Handle control messages if needed
						continue
					}
				}
				// Not a control message, treat as text output
				os.Stdout.Write(data)
			case websocket.CloseMessage:
				// Extract exit information from close message
				if len(data) >= 2 {
					c.handleCloseMessage(string(data[2:]))
				}
				close(done) // Signal session end
				return
			}
		}
	}
}

func (c *Client) handleCloseMessage(reason string) {
	if reason == "" {
		// Normal exit
		return
	}

	// Try to parse as JSON exit data
	type exitData struct {
		ExitCode   *int    `json:"exitCode,omitempty"`
		ExitReason *string `json:"exitReason,omitempty"`
	}

	var exit exitData
	if err := json.Unmarshal([]byte(reason), &exit); err == nil {
		if exit.ExitCode != nil && *exit.ExitCode != 0 {
			os.Exit(*exit.ExitCode)
		}
	}
	// If parsing fails, assume normal exit
}
