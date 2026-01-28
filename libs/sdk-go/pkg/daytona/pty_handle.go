// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/errors"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
	"github.com/gorilla/websocket"
)

// PtyHandle manages a WebSocket connection to a PTY (pseudo-terminal) session.
//
// PtyHandle provides methods for sending input, receiving output via channels,
// resizing the terminal, and managing the connection lifecycle. It implements
// [io.Reader] and [io.Writer] interfaces for integration with standard Go I/O.
//
// Create a PtyHandle using [ProcessService.CreatePty].
//
// Example:
//
//	// Create a PTY session
//	handle, err := sandbox.Process.CreatePty(ctx, "my-pty", nil)
//	if err != nil {
//	    return err
//	}
//	defer handle.Disconnect()
//
//	// Wait for connection to be established
//	if err := handle.WaitForConnection(ctx); err != nil {
//	    return err
//	}
//
//	// Send input
//	handle.SendInput([]byte("ls -la\n"))
//
//	// Read output from channel
//	for data := range handle.DataChan() {
//	    fmt.Print(string(data))
//	}
//
//	// Or use as io.Reader
//	io.Copy(os.Stdout, handle)
type PtyHandle struct {
	ws                    *websocket.Conn
	sessionID             string
	dataChan              chan []byte
	handleResize          func(context.Context, int, int) (*types.PtySessionInfo, error)
	handleKill            func(context.Context) error
	exitCode              *int
	err                   *string
	connected             bool
	connectionEstablished bool
	mu                    sync.RWMutex
	done                  chan struct{}
}

// controlMessage represents a WebSocket control message
type controlMessage struct {
	Type   string `json:"type"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// exitData represents the exit information from the PTY
type exitData struct {
	ExitCode   *int    `json:"exitCode,omitempty"`
	ExitReason *string `json:"exitReason,omitempty"`
	Error      *string `json:"error,omitempty"`
}

// newPtyHandle creates a new PTY handle with a WebSocket connection
func newPtyHandle(
	ws *websocket.Conn,
	sessionID string,
	handleResize func(context.Context, int, int) (*types.PtySessionInfo, error),
	handleKill func(context.Context) error,
) *PtyHandle {
	h := &PtyHandle{
		ws:           ws,
		sessionID:    sessionID,
		dataChan:     make(chan []byte, 100),
		handleResize: handleResize,
		handleKill:   handleKill,
		done:         make(chan struct{}),
	}

	// Start message handler
	go h.handleMessages()

	return h
}

// DataChan returns a channel for receiving PTY output.
//
// The channel receives raw bytes from the terminal. It is closed when the
// PTY session ends or the connection is closed.
//
// Example:
//
//	for data := range handle.DataChan() {
//	    fmt.Print(string(data))
//	}
func (h *PtyHandle) DataChan() <-chan []byte {
	return h.dataChan
}

// SessionID returns the unique identifier for this PTY session.
func (h *PtyHandle) SessionID() string {
	return h.sessionID
}

// ExitCode returns the exit code of the PTY process, or nil if still running.
func (h *PtyHandle) ExitCode() *int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.exitCode
}

// Error returns the error message if the PTY session failed, or nil otherwise.
func (h *PtyHandle) Error() *string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.err
}

// IsConnected returns true if the WebSocket connection is active.
func (h *PtyHandle) IsConnected() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.connected && h.ws != nil
}

// WaitForConnection waits for the WebSocket connection to be established.
//
// This method blocks until the PTY session is ready to receive input and send
// output, or until a timeout (10 seconds) expires. Always call this after
// creating a PTY to ensure the connection is ready.
//
// Example:
//
//	handle, _ := sandbox.Process.CreatePty(ctx, "my-pty", nil)
//	if err := handle.WaitForConnection(ctx); err != nil {
//	    return fmt.Errorf("PTY connection failed: %w", err)
//	}
//
// Returns an error if the connection times out or fails.
func (h *PtyHandle) WaitForConnection(ctx context.Context) error {
	h.mu.RLock()
	if h.connectionEstablished {
		h.mu.RUnlock()
		return nil
	}
	h.mu.RUnlock()

	// Create a timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return errors.NewDaytonaTimeoutError("PTY connection timeout")
		case <-ticker.C:
			h.mu.RLock()
			if h.connectionEstablished {
				h.mu.RUnlock()
				return nil
			}
			if h.err != nil {
				err := h.err
				h.mu.RUnlock()
				return errors.NewDaytonaError(*err, 0, nil)
			}
			h.mu.RUnlock()
		}
	}
}

// SendInput sends input data to the PTY session.
//
// The data is sent as raw bytes and will be processed as if typed in the terminal.
// Use this to send commands, keystrokes, or any terminal input.
//
// Example:
//
//	// Send a command
//	handle.SendInput([]byte("ls -la\n"))
//
//	// Send Ctrl+C
//	handle.SendInput([]byte{0x03})
//
// Returns an error if the PTY is not connected or sending fails.
func (h *PtyHandle) SendInput(data []byte) error {
	h.mu.RLock()
	if !h.connected || h.ws == nil {
		h.mu.RUnlock()
		return errors.NewDaytonaError("PTY is not connected", 0, nil)
	}
	ws := h.ws
	h.mu.RUnlock()

	if err := ws.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return errors.NewDaytonaError(fmt.Sprintf("Failed to send input to PTY: %v", err), 0, nil)
	}

	return nil
}

// Resize changes the PTY terminal dimensions.
//
// This notifies terminal applications about the new dimensions via SIGWINCH signal.
// Call this when the terminal display size changes.
//
// Parameters:
//   - cols: Number of columns (width in characters)
//   - rows: Number of rows (height in characters)
//
// Example:
//
//	info, err := handle.Resize(ctx, 120, 40)
//
// Returns updated [types.PtySessionInfo] or an error.
func (h *PtyHandle) Resize(ctx context.Context, cols, rows int) (*types.PtySessionInfo, error) {
	return h.handleResize(ctx, cols, rows)
}

// Disconnect closes the WebSocket connection and releases resources.
//
// Call this when done with the PTY session. This does not terminate the
// underlying process - use [PtyHandle.Kill] for that.
//
// Example:
//
//	defer handle.Disconnect()
//
// Returns an error if the WebSocket close fails.
func (h *PtyHandle) Disconnect() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	var wsErr error
	if h.ws != nil {
		wsErr = h.ws.Close()
		h.ws = nil
		h.connected = false
	}

	return wsErr
}

// Wait blocks until the PTY process exits and returns the result.
//
// Example:
//
//	result, err := handle.Wait(ctx)
//	if err != nil {
//	    return err
//	}
//	if result.ExitCode != nil {
//	    fmt.Printf("Process exited with code: %d\n", *result.ExitCode)
//	}
//
// Returns [types.PtyResult] with exit code and any error, or an error if
// the context is cancelled.
func (h *PtyHandle) Wait(ctx context.Context) (*types.PtyResult, error) {
	// Check if already exited
	h.mu.RLock()
	if h.exitCode != nil {
		exitCode := h.exitCode
		err := h.err
		h.mu.RUnlock()
		return &types.PtyResult{
			ExitCode: exitCode,
			Error:    err,
		}, nil
	}
	h.mu.RUnlock()

	// Wait for exit
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-h.done:
			h.mu.RLock()
			result := &types.PtyResult{
				ExitCode: h.exitCode,
				Error:    h.err,
			}
			h.mu.RUnlock()
			return result, nil
		case <-ticker.C:
			h.mu.RLock()
			if h.exitCode != nil {
				result := &types.PtyResult{
					ExitCode: h.exitCode,
					Error:    h.err,
				}
				h.mu.RUnlock()
				return result, nil
			}
			if h.err != nil {
				err := *h.err
				h.mu.RUnlock()
				return nil, errors.NewDaytonaError(err, 0, nil)
			}
			h.mu.RUnlock()
		}
	}
}

// Kill terminates the PTY session and its associated process.
//
// This operation is irreversible. The process receives a SIGKILL signal
// and terminates immediately.
//
// Example:
//
//	err := handle.Kill(ctx)
//
// Returns an error if the kill operation fails.
func (h *PtyHandle) Kill(ctx context.Context) error {
	return h.handleKill(ctx)
}

// Read implements [io.Reader] for reading PTY output.
//
// This method blocks until data is available or the PTY closes (returns [io.EOF]).
// Use with [io.Copy], [bufio.Scanner], or any standard Go I/O utilities.
//
// Example:
//
//	// Copy all output to stdout
//	io.Copy(os.Stdout, handle)
//
//	// Use with bufio.Scanner
//	scanner := bufio.NewScanner(handle)
//	for scanner.Scan() {
//	    fmt.Println(scanner.Text())
//	}
func (h *PtyHandle) Read(p []byte) (n int, err error) {
	data, ok := <-h.dataChan
	if !ok {
		return 0, io.EOF
	}
	n = copy(p, data)
	return n, nil
}

// Write implements [io.Writer] for sending input to the PTY.
//
// Example:
//
//	// Write directly
//	handle.Write([]byte("echo hello\n"))
//
//	// Use with io.Copy
//	io.Copy(handle, strings.NewReader("echo hello\n"))
func (h *PtyHandle) Write(p []byte) (n int, err error) {
	if err := h.SendInput(p); err != nil {
		return 0, err
	}
	return len(p), nil
}

// handleMessages processes incoming WebSocket messages
func (h *PtyHandle) handleMessages() {
	defer close(h.done)
	defer close(h.dataChan) // Close channel when done to signal EOF to readers

	for {
		h.mu.RLock()
		ws := h.ws
		h.mu.RUnlock()

		if ws == nil {
			return
		}

		messageType, data, err := ws.ReadMessage()
		if err != nil {
			h.mu.Lock()
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				// Normal closure - try to parse close message
				if closeErr, ok := err.(*websocket.CloseError); ok {
					h.handleCloseMessage(closeErr.Text)
				} else {
					// Default to exit code 0 for normal closure
					exitCode := 0
					h.exitCode = &exitCode
				}
			} else {
				errMsg := err.Error()
				h.err = &errMsg
			}
			h.connected = false
			h.mu.Unlock()
			return
		}

		switch messageType {
		case websocket.TextMessage:
			// Try to parse as control message
			var ctrl controlMessage
			if err := json.Unmarshal(data, &ctrl); err == nil && ctrl.Type == "control" {
				h.handleControlMessage(&ctrl)
			} else {
				// Regular text output - send to channel
				// Make a copy of data since WebSocket reuses the buffer
				dataCopy := make([]byte, len(data))
				copy(dataCopy, data)
				h.dataChan <- dataCopy
			}
		case websocket.BinaryMessage:
			// Binary data (terminal output) - send to channel
			// Make a copy of data since WebSocket reuses the buffer
			dataCopy := make([]byte, len(data))
			copy(dataCopy, data)
			h.dataChan <- dataCopy
		case websocket.CloseMessage:
			h.mu.Lock()
			// Extract close message data
			if len(data) >= 2 {
				h.handleCloseMessage(string(data[2:]))
			}
			h.connected = false
			h.mu.Unlock()
			return
		}
	}
}

// handleControlMessage processes control messages from the server
func (h *PtyHandle) handleControlMessage(msg *controlMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()

	switch msg.Status {
	case "connected":
		h.connected = true
		h.connectionEstablished = true
	case "error":
		errMsg := msg.Error
		if errMsg == "" {
			errMsg = "Unknown connection error"
		}
		h.err = &errMsg
		h.connected = false
	}
}

// handleCloseMessage parses the close message and extracts exit information
func (h *PtyHandle) handleCloseMessage(reason string) {
	if reason == "" {
		// Default to exit code 0 for empty close reason
		exitCode := 0
		h.exitCode = &exitCode
		return
	}

	// Try to parse as JSON exit data
	var exit exitData
	if err := json.Unmarshal([]byte(reason), &exit); err == nil {
		if exit.ExitCode != nil {
			h.exitCode = exit.ExitCode
		}
		if exit.ExitReason != nil {
			h.err = exit.ExitReason
		}
		if exit.Error != nil {
			h.err = exit.Error
		}
	} else {
		// Not JSON, default to exit code 0
		exitCode := 0
		h.exitCode = &exitCode
	}
}
