// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package interpreter

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// WebSocket upgrader with permissive origin policy
var interpreterUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// attachWebSocket connects a new WebSocket client to the interpreter session
func (s *InterpreterSession) attachWebSocket(ws *websocket.Conn) {
    cl := &wsClient{ id: uuid.NewString(), conn: ws, send: make(chan *OutputMessage, 256) }
    s.clientMu.Lock()
    // Replace existing client if present
    if s.client != nil {
        s.client.close()
    }
    s.client = cl
    s.clientMu.Unlock()
    log.Infof("Client %s attached to interpreter session %s", cl.id, s.info.ID)

	// Start writer goroutine
	go s.clientWriter(cl)

	// Send connection success message
    successMsg := &OutputMessage{Type: "control", Text: "connected"}
	if successJSON, err := json.Marshal(successMsg); err == nil {
		_ = ws.WriteMessage(websocket.TextMessage, successJSON)
	}

    // Reader loop (blocks until disconnect)
    s.clientReader(cl)

    // Cleanup on disconnect
    s.clientMu.Lock()
    if s.client != nil && s.client.id == cl.id {
        s.client = nil
    }
    s.clientMu.Unlock()
    cl.close()
    log.Infof("Client %s detached from interpreter session %s", cl.id, s.info.ID)
}

// clientWriter sends output messages to a specific WebSocket client
func (s *InterpreterSession) clientWriter(cl *wsClient) {
	for {
		select {
		case <-s.ctx.Done():
			return
		case msg, ok := <-cl.send:
			if !ok {
				return
			}
			_ = cl.conn.SetWriteDeadline(time.Now().Add(writeWait))
			
			// Send as JSON
			data, err := json.Marshal(msg)
			if err != nil {
				log.Errorf("Failed to marshal output message: %v", err)
				return
			}
			
			if err := cl.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		}
	}
}

// clientReader reads messages from a WebSocket client (currently unused, but keeps connection alive)
func (s *InterpreterSession) clientReader(cl *wsClient) {
	conn := cl.conn
	conn.SetReadLimit(readLimit)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Debug("ws read error:", err)
			}
			return
		}
		// Currently, we don't process messages from clients
		// This could be extended to support client commands in the future
	}
}

// broadcastOutput sends an output message to all connected WebSocket clients
func (s *InterpreterSession) broadcastOutput(msg *OutputMessage) {
    s.clientMu.RLock()
    cl := s.client
    s.clientMu.RUnlock()
    if cl == nil { return }
    select {
    case cl.send <- msg:
    default:
        // drop slow consumer
        _ = cl.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(
            websocket.ClosePolicyViolation, "slow consumer",
        ))
        cl.close()
        s.clientMu.Lock()
        if s.client != nil && s.client.id == cl.id {
            s.client = nil
        }
        s.clientMu.Unlock()
    }
}

// closeAllClients closes all WebSocket connections
func (s *InterpreterSession) closeAllClients() {
    s.clientMu.Lock()
    if s.client != nil {
        _ = s.client.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(
            websocket.CloseNormalClosure, "session closed",
        ))
        s.client.close()
        s.client = nil
    }
    s.clientMu.Unlock()
}

// close closes a WebSocket client connection
func (cl *wsClient) close() {
	cl.closeOnce.Do(func() {
		close(cl.send)
		_ = cl.conn.Close()
	})
}

