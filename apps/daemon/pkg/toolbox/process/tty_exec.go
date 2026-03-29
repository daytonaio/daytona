// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package process

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/creack/pty"
	"github.com/daytonaio/daemon/internal/util"
	"github.com/daytonaio/daemon/pkg/common"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	cmap "github.com/orcaman/concurrent-map/v2"
)

// Constants for TTY exec sessions
const (
	ttyExecWriteWait = 10 * time.Second
	ttyExecReadLimit = 64 * 1024
)

// ttyExecSession represents a TTY execution session
type ttyExecSession struct {
	logger *slog.Logger

	// Session metadata
	id        string
	command   string
	cwd       string
	timeout   *uint32
	createdAt time.Time
	active    bool

	// PTY process handling
	cmd    *exec.Cmd
	ptmx   *os.File
	ctx    context.Context
	cancel context.CancelFunc

	// WebSocket client management
	clients   cmap.ConcurrentMap[string, *ttyExecClient]
	clientsMu sync.RWMutex

	// Input channel from clients to PTY
	inCh chan []byte

	// Guards session fields
	mu sync.Mutex

	// Timeout handling
	timeoutReached atomic.Bool
}

// ttyExecClient represents a WebSocket client connection
type ttyExecClient struct {
	id        string
	conn      *websocket.Conn
	send      chan []byte   // outbound queue for this client (PTY -> WS)
	done      chan struct{} // closed when the client is shutting down
	closeOnce sync.Once
}

// TTYExecManager manages TTY execution sessions
type TTYExecManager struct {
	sessions cmap.ConcurrentMap[string, *ttyExecSession]
}

// Global manager instance
var ttyExecManager = &TTYExecManager{
	sessions: cmap.New[*ttyExecSession](),
}

// NewTTYExecManager creates a new TTY exec manager instance
func NewTTYExecManager() *TTYExecManager {
	return &TTYExecManager{
		sessions: cmap.New[*ttyExecSession](),
	}
}

// Get retrieves a TTY exec session by ID
func (m *TTYExecManager) Get(id string) (*ttyExecSession, bool) {
	s, ok := m.sessions.Get(id)
	return s, ok
}

// Set adds a TTY exec session to the manager
func (m *TTYExecManager) Set(id string, session *ttyExecSession) {
	m.sessions.Set(id, session)
}

// Remove removes a TTY exec session from the manager
func (m *TTYExecManager) Remove(id string) {
	m.sessions.Remove(id)
}

// createTTYExecSession creates a new TTY execution session
func createTTYExecSession(logger *slog.Logger, req ExecuteRequest) (*ttyExecSession, error) {
	sessionID := uuid.NewString()

	// Set defaults
	cwd := ""
	if req.Cwd != nil {
		cwd = *req.Cwd
	}

	session := &ttyExecSession{
		logger:    logger.With(slog.String("ttyExecSessionId", sessionID)),
		id:        sessionID,
		command:   req.Command,
		cwd:       cwd,
		timeout:   req.Timeout,
		createdAt: time.Now(),
		active:    false,
		clients:   cmap.New[*ttyExecClient](),
		inCh:      make(chan []byte, 1024),
	}

	// Add to manager
	ttyExecManager.Set(sessionID, session)

	return session, nil
}

// start initializes and starts the TTY execution session
func (s *ttyExecSession) start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Already running?
	if s.active && s.cmd != nil && s.ptmx != nil {
		return nil
	}

	// Prevent restarting
	if s.cmd != nil {
		return errors.New("TTY exec session has already been used and cannot be restarted")
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.ctx = ctx
	s.cancel = cancel

	// Create command
	shell := common.GetShell()
	if shell == "" {
		return errors.New("no shell resolved")
	}

	s.cmd = exec.CommandContext(ctx, shell)

	// Set working directory if specified
	if s.cwd != "" {
		s.cmd.Dir = s.cwd
	}

	// Set environment variables
	s.cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	// Start PTY with default size (will be resized when client connects)
	ptmx, err := pty.StartWithSize(s.cmd, &pty.Winsize{
		Rows: 24,
		Cols: 80,
	})
	if err != nil {
		s.cancel()
		return fmt.Errorf("failed to start PTY: %w", err)
	}

	s.ptmx = ptmx
	s.active = true

	s.logger.Debug("TTY exec session started", "sessionId", s.id, "command", s.command)

	// Set up timeout if specified
	if s.timeout != nil && *s.timeout > 0 {
		timeout := time.Duration(*s.timeout) * time.Second
		go func() {
			timer := time.NewTimer(timeout)
			defer timer.Stop()
			select {
			case <-timer.C:
				s.timeoutReached.Store(true)
				s.logger.Debug("TTY exec session timeout reached", "sessionId", s.id)
				s.kill()
			case <-ctx.Done():
				// Session ended before timeout
				return
			}
		}()
	}

	// Start goroutines for PTY I/O
	go s.handlePTYInput()
	go s.handlePTYOutput()
	go s.handleProcessExit()

	return nil
}

// handlePTYInput processes input from clients and sends to PTY
func (s *ttyExecSession) handlePTYInput() {
	defer func() {
		if s.ptmx != nil {
			s.ptmx.Close()
		}
	}()

	for {
		select {
		case <-s.ctx.Done():
			return
		case data := <-s.inCh:
			if s.ptmx != nil {
				if _, err := s.ptmx.Write(data); err != nil {
					s.logger.Debug("failed to write to PTY", "error", err)
					return
				}
			}
		}
	}
}

// handlePTYOutput reads from PTY and broadcasts to all clients
func (s *ttyExecSession) handlePTYOutput() {
	if s.ptmx == nil {
		return
	}

	buf := make([]byte, 1024)
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			n, err := s.ptmx.Read(buf)
			if err != nil {
				s.logger.Debug("PTY read error", "error", err)
				return
			}
			if n > 0 {
				s.broadcast(buf[:n])
			}
		}
	}
}

// handleProcessExit waits for the command to exit and cleans up
func (s *ttyExecSession) handleProcessExit() {
	if s.cmd == nil || s.cmd.Process == nil {
		return
	}

	// Wait for the process to exit
	err := s.cmd.Wait()

	s.mu.Lock()
	s.active = false
	s.mu.Unlock()

	var exitCode int
	var exitReason string

	if s.timeoutReached.Load() {
		exitCode = -1
		exitReason = "timeout"
	} else if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
			exitReason = "non-zero exit"
		} else {
			exitCode = -1
			exitReason = "execution error"
		}
	} else {
		exitCode = 0
		exitReason = "success"
	}

	s.logger.Debug("TTY exec session exited", "sessionId", s.id, "exitCode", exitCode, "reason", exitReason)

	// Close all WebSocket connections with exit information
	s.closeClientsWithExitCode(exitCode, exitReason)

	// Clean up
	s.cancel()

	// Remove from manager after a delay to allow clients to receive exit info
	go func() {
		time.Sleep(1 * time.Second)
		ttyExecManager.Remove(s.id)
	}()
}

// kill terminates the PTY session
func (s *ttyExecSession) kill() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cmd != nil && s.cmd.Process != nil {
		s.logger.Debug("Killing TTY exec session", "sessionId", s.id)
		_ = s.cmd.Process.Kill()
	}

	if s.cancel != nil {
		s.cancel()
	}

	s.active = false
}

// sendToPTY sends data to the PTY input channel
func (s *ttyExecSession) sendToPTY(data []byte) error {
	// Handle the command execution for the initial connect
	if s.command != "" && strings.TrimSpace(s.command) != "" {
		// Send the command followed by Enter
		cmdData := append([]byte(s.command), '\r')
		select {
		case s.inCh <- cmdData:
		case <-s.ctx.Done():
			return errors.New("session context cancelled")
		}
		// Clear the command so it's only sent once
		s.command = ""
	}

	// Send the actual input data
	select {
	case s.inCh <- data:
		return nil
	case <-s.ctx.Done():
		return errors.New("session context cancelled")
	}
}

// attachWebSocket connects a WebSocket client to the TTY exec session
func (s *ttyExecSession) attachWebSocket(ws *websocket.Conn) {
	client := &ttyExecClient{
		id:   uuid.NewString(),
		conn: ws,
		send: make(chan []byte, 256),
		done: make(chan struct{}),
	}

	// Register client
	s.clients.Set(client.id, client)
	s.logger.Debug("Client attached to TTY exec session", "clientId", client.id, "sessionId", s.id)

	// Start client writer
	go s.clientWriter(client)

	// Send success control message
	successMsg := map[string]interface{}{
		"type":   "control",
		"status": "connected",
	}
	if successJSON, err := json.Marshal(successMsg); err == nil {
		_ = ws.WriteMessage(websocket.TextMessage, successJSON)
	}

	// Execute the command when first client connects
	if s.command != "" {
		if err := s.sendToPTY([]byte{}); err != nil {
			s.logger.Error("Failed to execute command", "error", err)
		}
	}

	// Start client reader (blocks until disconnect)
	s.clientReader(client)

	// Clean up on disconnect
	s.clients.Remove(client.id)
	client.close()
	s.logger.Debug("Client detached from TTY exec session", "clientId", client.id, "sessionId", s.id)
}

// clientWriter sends PTY output to a WebSocket client
func (s *ttyExecSession) clientWriter(client *ttyExecClient) {
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-client.done:
			return
		case data := <-client.send:
			_ = client.conn.SetWriteDeadline(time.Now().Add(ttyExecWriteWait))
			if err := client.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
				return
			}
		}
	}
}

// clientReader reads input from WebSocket client and sends to PTY
func (s *ttyExecSession) clientReader(client *ttyExecClient) {
	conn := client.conn
	conn.SetReadLimit(ttyExecReadLimit)

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				s.logger.Debug("WebSocket read error", "error", err)
			}
			return
		}

		if err := s.sendToPTY(data); err != nil {
			_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(
				websocket.CloseInternalServerErr, "TTY exec session unavailable",
			))
			return
		}
	}
}

// broadcast sends data to all connected WebSocket clients
func (s *ttyExecSession) broadcast(data []byte) {
	s.clientsMu.RLock()
	for id, client := range s.clients.Items() {
		select {
		case client.send <- data:
		case <-client.done:
			// Client is shutting down, skip
		default:
			// Client's outbound queue is full, drop the client
			go func(id string, client *ttyExecClient) {
				_ = client.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(
					websocket.ClosePolicyViolation, "slow consumer",
				))
				client.close()
			}(id, client)
		}
	}
	s.clientsMu.RUnlock()
}

// closeClientsWithExitCode closes all WebSocket connections with structured exit data
func (s *ttyExecSession) closeClientsWithExitCode(exitCode int, exitReason string) {
	var wsCloseCode int
	var exitReasonStr *string

	// Map exit codes to WebSocket close codes
	if exitCode == 0 {
		wsCloseCode = websocket.CloseNormalClosure
		exitReasonStr = nil
	} else {
		wsCloseCode = websocket.CloseInternalServerErr
		switch {
		case exitCode == 130:
			reason := "Ctrl+C"
			exitReasonStr = &reason
		case exitCode == 137:
			reason := "SIGKILL"
			exitReasonStr = &reason
		case exitCode == 143:
			reason := "SIGTERM"
			exitReasonStr = &reason
		case exitCode > 128:
			sigNum := exitCode - 128
			reason := fmt.Sprintf("signal %d", sigNum)
			exitReasonStr = &reason
		case exitReason == "timeout":
			reason := "timeout"
			exitReasonStr = &reason
		default:
			reason := "non-zero exit"
			exitReasonStr = &reason
		}
	}

	// Create structured close data
	type CloseData struct {
		ExitCode   int     `json:"exitCode"`
		ExitReason *string `json:"exitReason,omitempty"`
	}

	closeData := CloseData{
		ExitCode:   exitCode,
		ExitReason: exitReasonStr,
	}

	closeJSON, _ := json.Marshal(closeData)

	s.clientsMu.Lock()
	for id, client := range s.clients.Items() {
		_ = client.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(
			wsCloseCode, string(closeJSON),
		))
		client.close()
		s.clients.Remove(id)
	}
	s.clientsMu.Unlock()
}

// close closes the client connection
func (c *ttyExecClient) close() {
	c.closeOnce.Do(func() {
		close(c.done)
		close(c.send)
		c.conn.Close()
	})
}

// ConnectTTYExecSession handles WebSocket connections to TTY exec sessions
func ConnectTTYExecSession(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("sessionId")
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "session ID is required"})
			return
		}

		// Upgrade to WebSocket
		ws, err := util.UpgradeToWebSocket(c.Writer, c.Request)
		if err != nil {
			logger.Error("WebSocket upgrade failed", "error", err)
			return
		}

		// Get session from manager
		session, exists := ttyExecManager.Get(sessionID)
		if !exists {
			logger.Debug("TTY exec session not found", "sessionId", sessionID)
			errorMsg := map[string]interface{}{
				"type":   "control",
				"status": "error",
				"error":  "TTY exec session not found",
			}
			if errorJSON, err := json.Marshal(errorMsg); err == nil {
				_ = ws.WriteMessage(websocket.TextMessage, errorJSON)
			}
			_ = ws.Close()
			return
		}

		// Start the session if not already active
		if !session.active {
			if err := session.start(); err != nil {
				logger.Error("Failed to start TTY exec session", "sessionId", sessionID, "error", err)
				errorMsg := map[string]interface{}{
					"type":   "control",
					"status": "error",
					"error":  "Failed to start TTY exec session: " + err.Error(),
				}
				if errorJSON, err := json.Marshal(errorMsg); err == nil {
					_ = ws.WriteMessage(websocket.TextMessage, errorJSON)
				}
				_ = ws.Close()
				return
			}
		}

		// Attach WebSocket to session
		session.attachWebSocket(ws)
	}
}
