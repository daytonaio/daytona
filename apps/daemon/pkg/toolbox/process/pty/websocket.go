// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package pty

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// WebSocket upgrader with permissive origin policy
var ptyUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// attachWebSocket connects a new WebSocket client to the PTY session
func (s *PTYSession) attachWebSocket(ws *websocket.Conn) {
	cl := &wsClient{
		id:   uuid.NewString(),
		conn: ws,
		send: make(chan []byte, 256), // if full, drop slow client
	}

	// Register client FIRST so it can receive PTY output via broadcast
	s.clientsMu.Lock()
	s.clients[cl.id] = cl
	count := len(s.clients)
	s.clientsMu.Unlock()
	log.Infof("Client %s attached to PTY session %s (clients=%d)", cl.id, s.info.ID, count)

	// Start PTY data flow - writer (PTY -> this client)
	go s.clientWriter(cl)

	// Send success control message after client is registered and ready
	successMsg := map[string]interface{}{
		"type":   "control",
		"status": "connected",
	}
	if successJSON, err := json.Marshal(successMsg); err == nil {
		_ = ws.WriteMessage(websocket.TextMessage, successJSON)
	}

	// reader (this client -> PTY); blocks until disconnect
	s.clientReader(cl)

	// on exit, unregister
	s.clientsMu.Lock()
	delete(s.clients, cl.id)
	s.clientsMu.Unlock()

	close(cl.send)
	_ = cl.conn.Close()

	s.clientsMu.RLock()
	remaining := len(s.clients)
	s.clientsMu.RUnlock()
	log.Infof("Client %s detached from PTY session %s (clients=%d)", cl.id, s.info.ID, remaining)
}

// clientWriter sends PTY output to a specific WebSocket client
func (s *PTYSession) clientWriter(cl *wsClient) {
	for {
		select {
		case <-s.ctx.Done():
			return
		case b, ok := <-cl.send:
			if !ok {
				return
			}
			_ = cl.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := cl.conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
				return
			}
		}
	}
}

// clientReader reads input from a WebSocket client and sends to PTY
func (s *PTYSession) clientReader(cl *wsClient) {
	conn := cl.conn
	conn.SetReadLimit(readLimit)

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Debug("ws read error:", err)
			}
			return
		}
		// Send all message data to PTY (text or binary)
		if err := s.sendToPTY(data); err != nil {
			// Send error to client and close connection
			_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(
				websocket.CloseInternalServerErr, "PTY session unavailable",
			))
			return
		}
	}
}

// broadcast sends data to all connected WebSocket clients
func (s *PTYSession) broadcast(b []byte) {
	// send to each client; drop slow clients to avoid stalling the PTY
	s.clientsMu.RLock()
	for id, cl := range s.clients {
		select {
		case cl.send <- b:
		default:
			// client's outbound queue is full -> drop the client
			go func(id string, cl *wsClient) {
				log.Warnf("Dropping slow client %s on session %s", id, s.info.ID)
				_ = cl.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(
					websocket.ClosePolicyViolation, "slow consumer",
				))
				_ = cl.conn.Close()
			}(id, cl)
		}
	}
	s.clientsMu.RUnlock()
}

// closeClientsWithExitCode closes all WebSocket connections with structured exit data
func (s *PTYSession) closeClientsWithExitCode(exitCode int, exitReason string) {
	var wsCloseCode int
	var exitReasonStr *string

	// Map PTY exit codes to WebSocket close codes
	if exitCode == 0 {
		wsCloseCode = websocket.CloseNormalClosure
		exitReasonStr = nil // undefined for clean exit
	} else {
		wsCloseCode = websocket.CloseInternalServerErr
		// Set human-readable reason for non-zero exits
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
		default:
			reason := "non-zero exit"
			exitReasonStr = &reason
		}
	}

	// Create structured close data as JSON
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
	for id, cl := range s.clients {
		_ = cl.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(
			wsCloseCode, string(closeJSON),
		))
		_ = cl.conn.Close()
		close(cl.send)
		delete(s.clients, id)
	}
	s.clientsMu.Unlock()
}
