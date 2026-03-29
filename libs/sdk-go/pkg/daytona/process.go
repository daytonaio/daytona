// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/common"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/errors"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/options"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
	"github.com/daytonaio/daytona/libs/toolbox-api-client-go"
	"github.com/gorilla/websocket"
)

// 3-byte multiplexing markers inserted by the shell labelers
var (
	stdoutPrefixBytes = []byte{0x01, 0x01, 0x01}
	stderrPrefixBytes = []byte{0x02, 0x02, 0x02}
)

const maxPrefixLen = 3

// ProcessService provides process execution operations for a sandbox.
//
// ProcessService enables command execution, session management, and PTY (pseudo-terminal)
// operations. It supports both synchronous command execution and interactive terminal
// sessions. Access through [Sandbox.Process].
//
// Example:
//
//	// Execute a command
//	result, err := sandbox.Process.ExecuteCommand(ctx, "echo 'Hello, World!'")
//	fmt.Println(result.Result)
//
//	// Execute with options
//	result, err := sandbox.Process.ExecuteCommand(ctx, "ls -la",
//	    options.WithCwd("/home/user/project"),
//	    options.WithExecuteTimeout(30*time.Second),
//	)
//
//	// Create an interactive PTY session
//	handle, err := sandbox.Process.CreatePty(ctx, "my-terminal")
//	defer handle.Disconnect()
type ProcessService struct {
	toolboxClient *toolbox.APIClient
	otel          *otelState
}

// NewProcessService creates a new ProcessService with the provided toolbox client.
//
// This is typically called internally by the SDK when creating a [Sandbox].
// Users should access ProcessService through [Sandbox.Process] rather than
// creating it directly.
func NewProcessService(toolboxClient *toolbox.APIClient, otel *otelState) *ProcessService {
	return &ProcessService{
		toolboxClient: toolboxClient,
		otel:          otel,
	}
}

// ExecuteCommand executes a shell command and returns the result.
//
// The command is executed in a shell context. For complex commands, consider
// using proper shell escaping or wrapping in a script.
//
// Optional parameters can be configured using functional options:
//   - [options.WithCwd]: Set the working directory for command execution
//   - [options.WithCommandEnv]: Set environment variables
//   - [options.WithExecuteTimeout]: Set execution timeout
//
// Example:
//
//	// Simple command
//	result, err := sandbox.Process.ExecuteCommand(ctx, "echo 'Hello'")
//	if err != nil {
//	    return err
//	}
//	fmt.Println(result.Result)
//
//	// Command with options
//	result, err := sandbox.Process.ExecuteCommand(ctx, "npm install",
//	    options.WithCwd("/home/user/project"),
//	    options.WithExecuteTimeout(5*time.Minute),
//	)
//
//	// Check exit code
//	if result.ExitCode != 0 {
//	    fmt.Printf("Command failed with exit code %d\n", result.ExitCode)
//	}
//
// Returns [types.ExecuteResponse] containing the output and exit code, or an error.
func (p *ProcessService) ExecuteCommand(ctx context.Context, command string, opts ...func(*options.ExecuteCommand)) (*types.ExecuteResponse, error) {
	return withInstrumentation(ctx, p.otel, "Process", "ExecuteCommand", func(ctx context.Context) (*types.ExecuteResponse, error) {
		execOpts := options.Apply(opts...)

		req := toolbox.NewExecuteRequest(command)
		if execOpts.Cwd != nil {
			req.SetCwd(*execOpts.Cwd)
		}
		if execOpts.Timeout != nil {
			req.SetTimeout(int32(execOpts.Timeout.Seconds()))
		}
		// Note: env parameter not supported in current toolbox API

		resp, httpResp, err := p.toolboxClient.ProcessAPI.ExecuteCommand(ctx).Request(*req).Execute()
		if err != nil {
			return nil, errors.ConvertToolboxError(err, httpResp)
		}

		exitCode := 0
		if resp.ExitCode != nil {
			exitCode = int(*resp.ExitCode)
		}

		return &types.ExecuteResponse{
			ExitCode: exitCode,
			Result:   resp.Result,
		}, nil
	})
}

// CodeRun executes code in a language-specific way.
//
// NOTE: This method is currently unavailable as the toolbox-api-client-go does not expose
// a CodeRun endpoint. For code execution, use [ProcessService.ExecuteCommand] or
// [CodeInterpreterService].
//
// Optional parameters can be configured using functional options:
//   - [options.WithCodeRunParams]: Set code execution parameters
//   - [options.WithCodeRunTimeout]: Set execution timeout
func (p *ProcessService) CodeRun(ctx context.Context, code string, opts ...func(*options.CodeRun)) (*types.ExecuteResponse, error) {
	return withInstrumentation(ctx, p.otel, "Process", "CodeRun", func(ctx context.Context) (*types.ExecuteResponse, error) {
		return nil, errors.NewDaytonaError("CodeRun is not supported by the current toolbox API. Use ExecuteCommand() or CodeInterpreter service instead.", 0, nil)
	})
}

// CreateSession creates a named session for executing multiple commands.
//
// Sessions allow you to execute multiple commands while maintaining state (like
// environment variables and working directory) between commands.
//
// Example:
//
//	// Create a session
//	err := sandbox.Process.CreateSession(ctx, "my-session")
//	if err != nil {
//	    return err
//	}
//	defer sandbox.Process.DeleteSession(ctx, "my-session")
//
//	// Execute commands in the session
//	result, err := sandbox.Process.ExecuteSessionCommand(ctx, "my-session", "cd /home/user", false)
//	result, err = sandbox.Process.ExecuteSessionCommand(ctx, "my-session", "pwd", false)
//
// Returns an error if session creation fails.
func (p *ProcessService) CreateSession(ctx context.Context, sessionID string) error {
	return withInstrumentationVoid(ctx, p.otel, "Process", "CreateSession", func(ctx context.Context) error {
		req := toolbox.NewCreateSessionRequest(sessionID)
		httpResp, err := p.toolboxClient.ProcessAPI.CreateSession(ctx).Request(*req).Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}

		return nil
	})
}

// GetSession retrieves information about a session.
//
// The sessionID parameter identifies the session to query.
//
// Returns a map containing:
//   - sessionId: The session identifier
//   - commands: List of commands executed in the session
//
// Example:
//
//	info, err := sandbox.Process.GetSession(ctx, "my-session")
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Session: %s\n", info["sessionId"])
//
// Returns an error if the session doesn't exist.
func (p *ProcessService) GetSession(ctx context.Context, sessionID string) (map[string]any, error) {
	return withInstrumentation(ctx, p.otel, "Process", "GetSession", func(ctx context.Context) (map[string]any, error) {
		resp, httpResp, err := p.toolboxClient.ProcessAPI.GetSession(ctx, sessionID).Execute()
		if err != nil {
			return nil, errors.ConvertToolboxError(err, httpResp)
		}

		// Convert to map for backward compatibility
		return map[string]any{
			"sessionId": resp.GetSessionId(),
			"commands":  resp.GetCommands(),
		}, nil
	})
}

// GetEntrypointSession retrieves information about the entrypoint session.
//
// Returns an entrypoint session information containing:
//   - SessionId: The entrypoint session identifier
//   - Commands: List of commands executed in the entrypoint session
//
// Example:
//
//	info, err := sandbox.Process.GetEntrypointSession(ctx)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Session: %s\n", info.SessionId)
//
// Returns an error if the session doesn't exist.
func (p *ProcessService) GetEntrypointSession(ctx context.Context) (*toolbox.Session, error) {
	return withInstrumentation(ctx, p.otel, "Process", "GetEntrypointSession", func(ctx context.Context) (*toolbox.Session, error) {
		resp, httpResp, err := p.toolboxClient.ProcessAPI.GetEntrypointSession(ctx).Execute()
		if err != nil {
			return nil, errors.ConvertToolboxError(err, httpResp)
		}

		return resp, nil
	})

}

// DeleteSession removes a session and releases its resources.
//
// The sessionID parameter identifies the session to delete.
//
// Example:
//
//	err := sandbox.Process.DeleteSession(ctx, "my-session")
//
// Returns an error if the session doesn't exist or deletion fails.
func (p *ProcessService) DeleteSession(ctx context.Context, sessionID string) error {
	return withInstrumentationVoid(ctx, p.otel, "Process", "DeleteSession", func(ctx context.Context) error {
		httpResp, err := p.toolboxClient.ProcessAPI.DeleteSession(ctx, sessionID).Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}

		return nil
	})
}

// ListSessions returns all active sessions.
//
// Example:
//
//	sessions, err := sandbox.Process.ListSessions(ctx)
//	if err != nil {
//	    return err
//	}
//	for _, session := range sessions {
//	    fmt.Printf("Session: %s\n", session["sessionId"])
//	}
//
// Returns a slice of session information maps, or an error.
func (p *ProcessService) ListSessions(ctx context.Context) ([]map[string]any, error) {
	return withInstrumentation(ctx, p.otel, "Process", "ListSessions", func(ctx context.Context) ([]map[string]any, error) {
		resp, httpResp, err := p.toolboxClient.ProcessAPI.ListSessions(ctx).Execute()
		if err != nil {
			return nil, errors.ConvertToolboxError(err, httpResp)
		}

		// Convert to map array for backward compatibility
		result := make([]map[string]any, len(resp))
		for i, session := range resp {
			result[i] = map[string]any{
				"sessionId": session.GetSessionId(),
				"commands":  session.GetCommands(),
			}
		}

		return result, nil
	})
}

// ExecuteSessionCommand executes a command within a session.
//
// Parameters:
//   - sessionID: The session to execute the command in
//   - command: The command to execute
//   - runAsync: If true, return immediately without waiting for completion
//   - suppressInputEcho: If true, suppress input echo
//
// When runAsync is true, use [ProcessService.GetSessionCommand] to check status
// and [ProcessService.GetSessionCommandLogs] to retrieve output.
//
// Example:
//
//	// Synchronous execution
//	result, err := sandbox.Process.ExecuteSessionCommand(ctx, "my-session", "ls -la", false)
//	if err != nil {
//	    return err
//	}
//	fmt.Println(result["stdout"])
//
//	// Asynchronous execution
//	result, err := sandbox.Process.ExecuteSessionCommand(ctx, "my-session", "long-running-cmd", true)
//	cmdID := result["id"].(string)
//	// Later: check status with GetSessionCommand(ctx, "my-session", cmdID)
//
// Returns command result including id, exitCode (if completed), stdout, and stderr.
func (p *ProcessService) ExecuteSessionCommand(ctx context.Context, sessionID, command string, runAsync bool, suppressInputEcho bool) (map[string]any, error) {
	return withInstrumentation(ctx, p.otel, "Process", "ExecuteSessionCommand", func(ctx context.Context) (map[string]any, error) {
		req := toolbox.NewSessionExecuteRequest(command)
		req.SetRunAsync(runAsync)
		req.SetSuppressInputEcho(suppressInputEcho)
		resp, httpResp, err := p.toolboxClient.ProcessAPI.SessionExecuteCommand(ctx, sessionID).Request(*req).Execute()
		if err != nil {
			return nil, errors.ConvertToolboxError(err, httpResp)
		}

		// Convert to map for backward compatibility
		result := map[string]any{
			"id": resp.GetCmdId(),
		}
		if resp.ExitCode != nil {
			result["exitCode"] = resp.GetExitCode()
		}
		if resp.Stdout != nil {
			result["stdout"] = resp.GetStdout()
		}
		if resp.Stderr != nil {
			result["stderr"] = resp.GetStderr()
		}

		return result, nil
	})
}

// GetSessionCommand retrieves the status of a command in a session.
//
// Parameters:
//   - sessionID: The session containing the command
//   - commandID: The command identifier (from ExecuteSessionCommand result)
//
// Example:
//
//	status, err := sandbox.Process.GetSessionCommand(ctx, "my-session", cmdID)
//	if err != nil {
//	    return err
//	}
//	if exitCode, ok := status["exitCode"]; ok {
//	    fmt.Printf("Command completed with exit code: %v\n", exitCode)
//	} else {
//	    fmt.Println("Command still running")
//	}
//
// Returns command status including id, command text, and exitCode (if completed).
func (p *ProcessService) GetSessionCommand(ctx context.Context, sessionID, commandID string) (map[string]any, error) {
	return withInstrumentation(ctx, p.otel, "Process", "GetSessionCommand", func(ctx context.Context) (map[string]any, error) {
		resp, httpResp, err := p.toolboxClient.ProcessAPI.GetSessionCommand(ctx, sessionID, commandID).Execute()
		if err != nil {
			return nil, errors.ConvertToolboxError(err, httpResp)
		}

		// Convert to map for backward compatibility
		result := map[string]any{
			"id":      resp.GetId(),
			"command": resp.GetCommand(),
		}
		if resp.ExitCode != nil {
			result["exitCode"] = resp.GetExitCode()
		}

		return result, nil
	})
}

// GetSessionCommandLogs retrieves the output logs of a command.
//
// Parameters:
//   - sessionID: The session containing the command
//   - commandID: The command identifier
//
// Example:
//
//	logs, err := sandbox.Process.GetSessionCommandLogs(ctx, "my-session", cmdID)
//	if err != nil {
//	    return err
//	}
//	fmt.Println(logs["logs"])
//
// Returns a map containing the "logs" key with command output.
func (p *ProcessService) GetSessionCommandLogs(ctx context.Context, sessionID, commandID string) (map[string]any, error) {
	return withInstrumentation(ctx, p.otel, "Process", "GetSessionCommandLogs", func(ctx context.Context) (map[string]any, error) {
		logs, httpResp, err := p.toolboxClient.ProcessAPI.GetSessionCommandLogs(ctx, sessionID, commandID).Execute()
		if err != nil {
			return nil, errors.ConvertToolboxError(err, httpResp)
		}

		// Convert to map for backward compatibility
		// The API returns logs as a plain string, so we return it as "logs"
		return map[string]any{
			"logs": logs,
		}, nil
	})
}

// GetSessionCommandLogsStream streams command logs as they become available.
//
// This method establishes a WebSocket connection to stream logs in real-time.
// The stdout and stderr channels receive log chunks as strings and are closed
// when the stream ends or an error occurs.
//
// Parameters:
//   - sessionID: The session containing the command
//   - commandID: The command identifier
//   - stdout: Channel to receive stdout output
//   - stderr: Channel to receive stderr output
//
// The caller should provide buffered channels to avoid blocking.
//
// Example:
//
//	stdout := make(chan string, 100)
//	stderr := make(chan string, 100)
//
//	go func() {
//	    err := sandbox.Process.GetSessionCommandLogsStream(ctx, "session", "cmd", stdout, stderr)
//	    if err != nil {
//	        log.Printf("Stream error: %v", err)
//	    }
//	}()
//
//	for {
//	    select {
//	    case chunk, ok := <-stdout:
//	        if !ok {
//	            stdout = nil
//	        } else {
//	            fmt.Print(chunk)
//	        }
//	    case chunk, ok := <-stderr:
//	        if !ok {
//	            stderr = nil
//	        } else {
//	            fmt.Fprint(os.Stderr, chunk)
//	        }
//	    }
//	    if stdout == nil && stderr == nil {
//	        break
//	    }
//	}
//
// Returns an error if the connection fails or stream encounters an error.
func (p *ProcessService) GetSessionCommandLogsStream(ctx context.Context, sessionID, commandID string, stdout, stderr chan<- string) error {
	return withInstrumentationVoid(ctx, p.otel, "Process", "GetSessionCommandLogsStream", func(ctx context.Context) error {
		defer func() {
			close(stdout)
			close(stderr)
		}()

		// Convert HTTP URL to WebSocket URL
		httpURL := p.toolboxClient.GetConfig().Servers[0].URL
		wsURL := common.ConvertToWebSocketURL(httpURL)

		// Get authentication headers from the toolbox client configuration
		headers := make(map[string][]string)
		cfg := p.toolboxClient.GetConfig()
		for key, value := range cfg.DefaultHeader {
			headers[key] = []string{value}
		}

		// Connect to WebSocket with follow=true to stream logs
		wsEndpoint := fmt.Sprintf("%s/process/session/%s/command/%s/logs?follow=true", wsURL, sessionID, commandID)
		conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsEndpoint, headers)
		if err != nil {
			return errors.NewDaytonaError(fmt.Sprintf("Failed to connect to log stream: %v", err), 0, nil)
		}
		defer conn.Close()

		// Process the WebSocket stream and demux stdout/stderr
		return processWebsocketStream(ctx, conn, stdout, stderr)
	})
}

// GetEntrypointLogs retrieves the output logs of the sandbox entrypoint.
//
// Example:
//
//	logs, err := sandbox.Process.GetEntrypointLogs(ctx)
//	if err != nil {
//	    return err
//	}
//	fmt.Println(logs)
//
// Returns a string containing the entrypoint command output logs.
func (p *ProcessService) GetEntrypointLogs(ctx context.Context) (string, error) {
	return withInstrumentation(ctx, p.otel, "Process", "GetEntrypointLogs", func(ctx context.Context) (string, error) {
		logs, httpResp, err := p.toolboxClient.ProcessAPI.GetEntrypointLogs(ctx).Execute()
		if err != nil {
			return "", errors.ConvertToolboxError(err, httpResp)
		}

		return logs, nil
	})
}

// GetEntrypointLogsStream streams entrypoint logs as they become available.
//
// This method establishes a WebSocket connection to stream sandbox entrypoint logs in real-time.
// The stdout and stderr channels receive log chunks as strings and are closed
// when the stream ends or an error occurs.
//
// Parameters:
//   - stdout: Channel to receive stdout output
//   - stderr: Channel to receive stderr output
//
// The caller should provide buffered channels to avoid blocking.
//
// Example:
//
//	stdout := make(chan string, 100)
//	stderr := make(chan string, 100)
//
//	go func() {
//	    err := sandbox.Process.GetEntrypointLogsStream(ctx, stdout, stderr)
//	    if err != nil {
//	        log.Printf("Stream error: %v", err)
//	    }
//	}()
//
//	for {
//	    select {
//	    case chunk, ok := <-stdout:
//	        if !ok {
//	            stdout = nil
//	        } else {
//	            fmt.Print(chunk)
//	        }
//	    case chunk, ok := <-stderr:
//	        if !ok {
//	            stderr = nil
//	        } else {
//	            fmt.Fprint(os.Stderr, chunk)
//	        }
//	    }
//	    if stdout == nil && stderr == nil {
//	        break
//	    }
//	}
//
// Returns an error if the connection fails or stream encounters an error.
func (p *ProcessService) GetEntrypointLogsStream(ctx context.Context, stdout, stderr chan<- string) error {
	return withInstrumentationVoid(ctx, p.otel, "Process", "GetEntrypointLogsStream", func(ctx context.Context) error {
		defer func() {
			close(stdout)
			close(stderr)
		}()
		// Convert HTTP URL to WebSocket URL
		httpURL := p.toolboxClient.GetConfig().Servers[0].URL
		wsURL := common.ConvertToWebSocketURL(httpURL)
		// Get authentication headers from the toolbox client configuration
		headers := make(map[string][]string)
		cfg := p.toolboxClient.GetConfig()
		for key, value := range cfg.DefaultHeader {
			headers[key] = []string{value}
		}
		// Connect to WebSocket with follow=true to stream logs
		wsEndpoint := fmt.Sprintf("%s/process/session/entrypoint/logs?follow=true", wsURL)
		conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsEndpoint, headers)
		if err != nil {
			return errors.NewDaytonaError(fmt.Sprintf("Failed to connect to log stream: %v", err), 0, nil)
		}
		defer conn.Close()
		// Process the WebSocket stream and demux stdout/stderr
		return processWebsocketStream(ctx, conn, stdout, stderr)
	})
}

// processWebsocketStream demultiplexes a WebSocket stream into separate stdout and stderr channels.
func processWebsocketStream(ctx context.Context, conn *websocket.Conn, stdout, stderr chan<- string) error {
	var buf []byte
	var currentType string // "", "stdout", or "stderr"

	flush := func() {
		if len(buf) > 0 && currentType != "" {
			flushToChannel(buf, currentType, stdout, stderr)
			buf = nil
		}
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return ctx.Err()
		default:
		}

		// Read message from WebSocket
		_, message, err := conn.ReadMessage()
		if err != nil {
			flush()
			// Handle normal closure
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return nil
			}
			return errors.NewDaytonaError(fmt.Sprintf("WebSocket read error: %v", err), 0, nil)
		}

		// Nothing to process, continue
		if len(message) == 0 {
			continue
		}

		buf = append(buf, message...)

		// Process buffer looking for markers
		for {
			stdoutIdx := bytes.Index(buf, stdoutPrefixBytes)
			stderrIdx := bytes.Index(buf, stderrPrefixBytes)

			// Find nearest marker
			nextIdx := -1
			var nextType string
			var markerLen int

			if stdoutIdx != -1 && (stderrIdx == -1 || stdoutIdx < stderrIdx) {
				nextIdx, nextType, markerLen = stdoutIdx, "stdout", len(stdoutPrefixBytes)
			} else if stderrIdx != -1 {
				nextIdx, nextType, markerLen = stderrIdx, "stderr", len(stderrPrefixBytes)
			}

			if nextIdx == -1 {
				// No marker found - emit all but last (maxPrefixLen-1) bytes
				// to handle potential partial marker split across messages
				if len(buf) > maxPrefixLen-1 {
					emitLen := len(buf) - (maxPrefixLen - 1)
					if currentType != "" {
						flushToChannel(buf[:emitLen], currentType, stdout, stderr)
					}
					buf = buf[emitLen:]
				}
				break
			}

			// Emit bytes before marker
			if nextIdx > 0 && currentType != "" {
				flushToChannel(buf[:nextIdx], currentType, stdout, stderr)
			}

			// Skip marker and switch type
			buf = buf[nextIdx+markerLen:]
			currentType = nextType
		}
	}
}

// flushToChannel sends data to the appropriate channel based on the data type
func flushToChannel(data []byte, dataType string, stdout, stderr chan<- string) {
	if len(data) == 0 {
		return
	}
	text := string(data)
	switch dataType {
	case "stdout":
		stdout <- text
	case "stderr":
		stderr <- text
	default:
		// Drop unknown data type
	}
}

// makeHTTPRequest performs a raw HTTP request for TTY execution
func makeHTTPRequest(ctx context.Context, method, url string, body interface{}, headers map[string]string) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set content type for JSON
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add provided headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

// parseJSONResponse parses a JSON response into the provided struct
func parseJSONResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return nil
}

// CreatePtySession creates a PTY (pseudo-terminal) session.
//
// A PTY session provides a terminal interface for interactive applications.
// Use [ProcessService.ConnectPty] to connect to the session after creation.
//
// Parameters:
//   - id: Unique identifier for the session
//
// Optional parameters can be configured using functional options:
//   - [options.WithPtySize]: Set terminal dimensions (rows and columns)
//   - [options.WithPtyEnv]: Set environment variables
//
// Example:
//
//	// Create with default settings
//	session, err := sandbox.Process.CreatePtySession(ctx, "my-terminal")
//
//	// Create with custom size
//	session, err := sandbox.Process.CreatePtySession(ctx, "my-terminal",
//	    options.WithPtySize(types.PtySize{Rows: 24, Cols: 80}),
//	)
//
// Returns [types.PtySessionInfo] containing session details, or an error.
func (p *ProcessService) CreatePtySession(ctx context.Context, id string, opts ...func(*options.PtySession)) (*types.PtySessionInfo, error) {
	return withInstrumentation(ctx, p.otel, "Process", "CreatePtySession", func(ctx context.Context) (*types.PtySessionInfo, error) {
		sessionOpts := options.Apply(opts...)

		req := toolbox.NewPtyCreateRequest()
		if id != "" {
			req.SetId(id)
		}
		if sessionOpts.PtySize != nil {
			req.SetRows(int32(sessionOpts.PtySize.Rows))
			req.SetCols(int32(sessionOpts.PtySize.Cols))
		}
		if sessionOpts.Env != nil {
			req.SetEnvs(sessionOpts.Env)
		}

		resp, httpResp, err := p.toolboxClient.ProcessAPI.CreatePtySession(ctx).Request(*req).Execute()
		if err != nil {
			return nil, errors.ConvertToolboxError(err, httpResp)
		}

		// Return the session info with the created session ID and requested size
		result := &types.PtySessionInfo{
			ID: resp.GetSessionId(),
		}
		if sessionOpts.PtySize != nil {
			result.Rows = sessionOpts.PtySize.Rows
			result.Cols = sessionOpts.PtySize.Cols
		}

		return result, nil
	})
}

// ListPtySessions returns all active PTY sessions.
//
// Example:
//
//	sessions, err := sandbox.Process.ListPtySessions(ctx)
//	if err != nil {
//	    return err
//	}
//	for _, session := range sessions {
//	    fmt.Printf("PTY: %s (%dx%d)\n", session.ID, session.Cols, session.Rows)
//	}
//
// Returns a slice of [types.PtySessionInfo], or an error.
func (p *ProcessService) ListPtySessions(ctx context.Context) ([]*types.PtySessionInfo, error) {
	return withInstrumentation(ctx, p.otel, "Process", "ListPtySessions", func(ctx context.Context) ([]*types.PtySessionInfo, error) {
		resp, httpResp, err := p.toolboxClient.ProcessAPI.ListPtySessions(ctx).Execute()
		if err != nil {
			return nil, errors.ConvertToolboxError(err, httpResp)
		}

		sessions := resp.GetSessions()
		result := make([]*types.PtySessionInfo, len(sessions))
		for i, session := range sessions {
			result[i] = &types.PtySessionInfo{
				ID:   session.GetId(),
				Rows: int(session.GetRows()),
				Cols: int(session.GetCols()),
			}
		}

		return result, nil
	})
}

// GetPtySessionInfo retrieves information about a PTY session.
//
// Parameters:
//   - sessionID: The PTY session identifier
//
// Example:
//
//	info, err := sandbox.Process.GetPtySessionInfo(ctx, "my-terminal")
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Terminal size: %dx%d\n", info.Cols, info.Rows)
//
// Returns [types.PtySessionInfo] with session details, or an error.
func (p *ProcessService) GetPtySessionInfo(ctx context.Context, sessionID string) (*types.PtySessionInfo, error) {
	return withInstrumentation(ctx, p.otel, "Process", "GetPtySessionInfo", func(ctx context.Context) (*types.PtySessionInfo, error) {
		resp, httpResp, err := p.toolboxClient.ProcessAPI.GetPtySession(ctx, sessionID).Execute()
		if err != nil {
			return nil, errors.ConvertToolboxError(err, httpResp)
		}

		return &types.PtySessionInfo{
			ID:   resp.GetId(),
			Rows: int(resp.GetRows()),
			Cols: int(resp.GetCols()),
		}, nil
	})
}

// KillPtySession terminates a PTY session.
//
// This ends the terminal session and any processes running in it.
//
// Parameters:
//   - sessionID: The PTY session to terminate
//
// Example:
//
//	err := sandbox.Process.KillPtySession(ctx, "my-terminal")
//
// Returns an error if the session doesn't exist or termination fails.
func (p *ProcessService) KillPtySession(ctx context.Context, sessionID string) error {
	return withInstrumentationVoid(ctx, p.otel, "Process", "KillPtySession", func(ctx context.Context) error {
		_, httpResp, err := p.toolboxClient.ProcessAPI.DeletePtySession(ctx, sessionID).Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}

		return nil
	})
}

// ResizePtySession changes the terminal dimensions of a PTY session.
//
// This sends a SIGWINCH signal to applications, notifying them of the size change.
//
// Parameters:
//   - sessionID: The PTY session to resize
//   - ptySize: New terminal dimensions
//
// Example:
//
//	newSize := types.PtySize{Rows: 40, Cols: 120}
//	info, err := sandbox.Process.ResizePtySession(ctx, "my-terminal", newSize)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("New size: %dx%d\n", info.Cols, info.Rows)
//
// Returns updated [types.PtySessionInfo], or an error.
func (p *ProcessService) ResizePtySession(ctx context.Context, sessionID string, ptySize types.PtySize) (*types.PtySessionInfo, error) {
	return withInstrumentation(ctx, p.otel, "Process", "ResizePtySession", func(ctx context.Context) (*types.PtySessionInfo, error) {
		req := toolbox.NewPtyResizeRequest(int32(ptySize.Cols), int32(ptySize.Rows))
		resp, httpResp, err := p.toolboxClient.ProcessAPI.ResizePtySession(ctx, sessionID).Request(*req).Execute()
		if err != nil {
			return nil, errors.ConvertToolboxError(err, httpResp)
		}

		return &types.PtySessionInfo{
			ID:   resp.GetId(),
			Rows: int(resp.GetRows()),
			Cols: int(resp.GetCols()),
		}, nil
	})
}

// ConnectPty establishes a WebSocket connection to an existing PTY session.
//
// Returns a [PtyHandle] for interacting with the terminal. The handle provides:
//   - DataChan(): Channel for receiving terminal output
//   - SendInput(): Method for sending keyboard input
//   - Resize(): Method for changing terminal size
//   - Disconnect(): Method for closing the connection
//
// Parameters:
//   - sessionID: The PTY session to connect to
//
// Example:
//
//	handle, err := sandbox.Process.ConnectPty(ctx, "my-terminal")
//	if err != nil {
//	    return err
//	}
//	defer handle.Disconnect()
//
//	// Wait for connection
//	if err := handle.WaitForConnection(ctx); err != nil {
//	    return err
//	}
//
//	// Read output
//	for data := range handle.DataChan() {
//	    fmt.Print(string(data))
//	}
//
// Returns a [PtyHandle] for terminal interaction, or an error.
func (p *ProcessService) ConnectPty(ctx context.Context, sessionID string) (*PtyHandle, error) {
	return withInstrumentation(ctx, p.otel, "Process", "ConnectPty", func(ctx context.Context) (*PtyHandle, error) {
		// Convert HTTP URL to WebSocket URL
		httpURL := p.toolboxClient.GetConfig().Servers[0].URL
		wsURL := common.ConvertToWebSocketURL(httpURL)

		// Get authentication headers from the toolbox client configuration
		headers := make(map[string][]string)
		cfg := p.toolboxClient.GetConfig()
		for key, value := range cfg.DefaultHeader {
			headers[key] = []string{value}
		}

		conn, _, err := websocket.DefaultDialer.DialContext(ctx, fmt.Sprintf("%s/process/pty/%s/connect", wsURL, sessionID), headers)
		if err != nil {
			return nil, errors.NewDaytonaError(fmt.Sprintf("Failed to connect to PTY: %v", err), 0, nil)
		}

		// Create resize handler
		resizeHandler := func(ctx context.Context, cols, rows int) (*types.PtySessionInfo, error) {
			return p.ResizePtySession(ctx, sessionID, types.PtySize{Cols: cols, Rows: rows})
		}

		// Create kill handler
		killHandler := func(ctx context.Context) error {
			return p.KillPtySession(ctx, sessionID)
		}

		// Create and return the handle
		handle := newPtyHandle(conn, sessionID, resizeHandler, killHandler)

		return handle, nil
	})
}

// ExecuteTTY executes a command in TTY mode by creating a PTY session.
//
// TTY (pseudo-terminal) mode provides an interactive terminal interface for commands
// that require terminal interaction. This method creates a PTY session and returns
// a session ID that can be used to connect via WebSocket.
//
// Optional parameters can be configured using functional options:
//   - [options.WithTTYCwd]: Set the working directory for command execution
//   - [options.WithTTYTimeout]: Set command execution timeout
//   - [options.WithTTYSize]: Set terminal dimensions
//
// Example:
//
//	// Execute an interactive command
//	response, err := sandbox.Process.ExecuteTTY(ctx, "vim /home/user/file.txt",
//	    options.WithTTYCwd("/home/user"),
//	    options.WithTTYSize(types.PtySize{Rows: 30, Cols: 120}),
//	)
//	if err != nil {
//	    return err
//	}
//
//	// Connect to the session using the session ID
//	handle, err := sandbox.Process.ConnectPty(ctx, response.SessionID)
//	if err != nil {
//	    return err
//	}
//	defer handle.Disconnect()
//
// Returns [types.ExecuteTTYResponse] containing the session ID, or an error.
func (p *ProcessService) ExecuteTTY(ctx context.Context, command string, opts ...func(*options.ExecuteTTY)) (*types.ExecuteTTYResponse, error) {
	return withInstrumentation(ctx, p.otel, "Process", "ExecuteTTY", func(ctx context.Context) (*types.ExecuteTTYResponse, error) {
		execOpts := options.Apply(opts...)

		// Build the request for TTY execution
		req := toolbox.NewExecuteRequest(command)
		if execOpts.Cwd != nil {
			req.SetCwd(*execOpts.Cwd)
		}
		if execOpts.Timeout != nil {
			req.SetTimeout(int32(execOpts.Timeout.Seconds()))
		}

		// Create a custom request body with TTY flag
		// Since the generated client doesn't support TTY, we'll create a raw HTTP request
		reqBody := map[string]interface{}{
			"command": command,
			"tty":     true,
		}
		if execOpts.Cwd != nil {
			reqBody["cwd"] = *execOpts.Cwd
		}
		if execOpts.Timeout != nil {
			reqBody["timeout"] = int32(execOpts.Timeout.Seconds())
		}

		// Use raw HTTP client to make the TTY execution request
		endpoint := "/process/execute"

		// Get base URL from toolbox client
		cfg := p.toolboxClient.GetConfig()
		baseURL := cfg.Servers[0].URL

		response, err := makeHTTPRequest(ctx, "POST", baseURL+endpoint, reqBody, cfg.DefaultHeader)
		if err != nil {
			return nil, errors.NewDaytonaError(fmt.Sprintf("Failed to execute TTY command: %v", err), 0, nil)
		}

		// Parse the TTY response
		var ttyResp struct {
			SessionID string `json:"sessionId"`
		}
		if err := parseJSONResponse(response, &ttyResp); err != nil {
			return nil, errors.NewDaytonaError(fmt.Sprintf("Failed to parse TTY response: %v", err), 0, nil)
		}

		return &types.ExecuteTTYResponse{
			SessionID: ttyResp.SessionID,
		}, nil
	})
}

// ExecuteTTYAndConnect executes a command in TTY mode and immediately connects to it.
//
// This is a convenience method that combines [ProcessService.ExecuteTTY] and
// [ProcessService.ConnectTTYExec] into a single operation.
//
// Optional parameters can be configured using functional options:
//   - [options.WithTTYCwd]: Set the working directory for command execution
//   - [options.WithTTYTimeout]: Set command execution timeout
//   - [options.WithTTYSize]: Set terminal dimensions
//
// Example:
//
//	handle, err := sandbox.Process.ExecuteTTYAndConnect(ctx, "vim /home/user/file.txt",
//	    options.WithTTYCwd("/home/user"),
//	    options.WithTTYSize(types.PtySize{Rows: 30, Cols: 120}),
//	)
//	if err != nil {
//	    return err
//	}
//	defer handle.Disconnect()
//
//	// Wait for connection
//	if err := handle.WaitForConnection(ctx); err != nil {
//	    return err
//	}
//
//	// Interact with the terminal
//	for data := range handle.DataChan() {
//	    fmt.Print(string(data))
//	}
//
// Returns a [PtyHandle] for terminal interaction, or an error.
func (p *ProcessService) ExecuteTTYAndConnect(ctx context.Context, command string, opts ...func(*options.ExecuteTTY)) (*PtyHandle, error) {
	return withInstrumentation(ctx, p.otel, "Process", "ExecuteTTYAndConnect", func(ctx context.Context) (*PtyHandle, error) {
		// Execute the command in TTY mode
		response, err := p.ExecuteTTY(ctx, command, opts...)
		if err != nil {
			return nil, err
		}

		// Connect to the TTY execution session
		return p.ConnectTTYExec(ctx, response.SessionID)
	})
}

// ConnectTTYExec establishes a WebSocket connection to a TTY execution session.
//
// This method connects to a TTY execution session created by [ProcessService.ExecuteTTY].
// Unlike regular PTY sessions, TTY execution sessions execute a specific command.
//
// Parameters:
//   - sessionID: The TTY execution session ID from ExecuteTTY response
//
// Example:
//
//	// First execute a command in TTY mode
//	response, err := sandbox.Process.ExecuteTTY(ctx, "vim file.txt")
//	if err != nil {
//	    return err
//	}
//
//	// Then connect to the session
//	handle, err := sandbox.Process.ConnectTTYExec(ctx, response.SessionID)
//	if err != nil {
//	    return err
//	}
//	defer handle.Disconnect()
//
//	// Wait for connection and interact
//	if err := handle.WaitForConnection(ctx); err != nil {
//	    return err
//	}
//
//	for data := range handle.DataChan() {
//	    fmt.Print(string(data))
//	}
//
// Returns a [PtyHandle] for terminal interaction, or an error.
func (p *ProcessService) ConnectTTYExec(ctx context.Context, sessionID string) (*PtyHandle, error) {
	return withInstrumentation(ctx, p.otel, "Process", "ConnectTTYExec", func(ctx context.Context) (*PtyHandle, error) {
		// Convert HTTP URL to WebSocket URL
		httpURL := p.toolboxClient.GetConfig().Servers[0].URL
		wsURL := common.ConvertToWebSocketURL(httpURL)

		// Get authentication headers from the toolbox client configuration
		headers := make(map[string][]string)
		cfg := p.toolboxClient.GetConfig()
		for key, value := range cfg.DefaultHeader {
			headers[key] = []string{value}
		}

		// Connect to TTY execution WebSocket endpoint
		conn, _, err := websocket.DefaultDialer.DialContext(ctx, fmt.Sprintf("%s/process/exec/%s/connect", wsURL, sessionID), headers)
		if err != nil {
			return nil, errors.NewDaytonaError(fmt.Sprintf("Failed to connect to TTY execution session: %v", err), 0, nil)
		}

		// Create dummy resize and kill handlers for TTY exec sessions
		resizeHandler := func(ctx context.Context, cols, rows int) (*types.PtySessionInfo, error) {
			// TTY execution sessions don't support resizing in the same way as PTY sessions
			return &types.PtySessionInfo{
				ID:   sessionID,
				Rows: rows,
				Cols: cols,
			}, nil
		}

		killHandler := func(ctx context.Context) error {
			// For TTY execution sessions, killing is handled by the session timeout or completion
			return nil
		}

		// Create and return the handle
		handle := newPtyHandle(conn, sessionID, resizeHandler, killHandler)

		return handle, nil
	})
}

// CreatePty creates a new PTY session and immediately connects to it.
//
// This is a convenience method that combines [ProcessService.CreatePtySession] and
// [ProcessService.ConnectPty] into a single operation.
//
// Parameters:
//   - id: Unique identifier for the PTY session
//
// Optional parameters can be configured using functional options:
//   - [options.WithCreatePtySize]: Set terminal dimensions
//   - [options.WithCreatePtyEnv]: Set environment variables
//
// Example:
//
//	handle, err := sandbox.Process.CreatePty(ctx, "interactive-shell",
//	    options.WithCreatePtySize(types.PtySize{Rows: 24, Cols: 80}),
//	    options.WithCreatePtyEnv(map[string]string{"TERM": "xterm-256color"}),
//	)
//	if err != nil {
//	    return err
//	}
//	defer handle.Disconnect()
//
//	// Wait for connection
//	if err := handle.WaitForConnection(ctx); err != nil {
//	    return err
//	}
//
//	// Send a command
//	handle.SendInput([]byte("ls -la\n"))
//
//	// Read output
//	for data := range handle.DataChan() {
//	    fmt.Print(string(data))
//	}
//
// Returns a [PtyHandle] for terminal interaction, or an error.
func (p *ProcessService) CreatePty(ctx context.Context, id string, opts ...func(*options.CreatePty)) (*PtyHandle, error) {
	return withInstrumentation(ctx, p.otel, "Process", "CreatePty", func(ctx context.Context) (*PtyHandle, error) {
		createOpts := options.Apply(opts...)

		// Convert to CreatePtySession options
		sessionOpts := []func(*options.PtySession){}
		if createOpts.PtySize != nil {
			sessionOpts = append(sessionOpts, options.WithPtySize(*createOpts.PtySize))
		}
		if createOpts.Env != nil {
			sessionOpts = append(sessionOpts, options.WithPtyEnv(createOpts.Env))
		}

		// Create the PTY session
		_, err := p.CreatePtySession(ctx, id, sessionOpts...)
		if err != nil {
			return nil, err
		}

		// Connect to the session
		return p.ConnectPty(ctx, id)
	})
}
