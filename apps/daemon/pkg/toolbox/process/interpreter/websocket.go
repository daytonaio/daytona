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
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// attachWebSocket connects a WebSocket client to the interpreter context
func (c *Context) attachWebSocket(ws *websocket.Conn) {
	cl := &wsClient{
		id:   uuid.NewString(),
		conn: ws,
		send: make(chan *OutputMessage, 256),
		done: make(chan struct{}),
	}

	c.mu.Lock()
	if c.client != nil {
		c.client.close()
	}
	c.client = cl
	c.mu.Unlock()

	log.Debugf("Client %s attached to interpreter context %s", cl.id, c.info.ID)

	go c.clientWriter(cl)

	// Wait for clientWriter to exit (signals disconnection)
	<-cl.done

	c.mu.Lock()
	if c.client != nil && c.client.id == cl.id {
		c.client = nil
	}
	c.mu.Unlock()

	cl.close()
	log.Debugf("Client %s detached from interpreter context %s", cl.id, c.info.ID)
}

// clientWriter sends output messages to the WebSocket client
func (c *Context) clientWriter(cl *wsClient) {
	defer close(cl.done)

	for {
		select {
		case <-c.ctx.Done():
			return
		case msg, ok := <-cl.send:
			if !ok {
				return
			}
			_ = cl.conn.SetWriteDeadline(time.Now().Add(writeWait))

			data, err := json.Marshal(msg)
			if err != nil {
				log.Errorf("Failed to marshal output message: %v", err)
				return
			}

			err = cl.conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				return
			}
		}
	}
}

// emitOutput sends an output message to the connected WebSocket client
func (c *Context) emitOutput(msg *OutputMessage) {
	c.mu.Lock()
	cl := c.client
	c.mu.Unlock()

	if cl == nil {
		return
	}

	select {
	case cl.send <- msg:
	default:
		log.Debug("Client send channel full - closing slow consumer")
		closeMsg := websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "slow consumer")
		_ = cl.conn.WriteControl(websocket.CloseMessage, closeMsg, time.Now().Add(writeWait))
		cl.close()

		c.mu.Lock()
		if c.client != nil && c.client.id == cl.id {
			c.client = nil
		}
		c.mu.Unlock()
	}
}

// close closes a WebSocket client connection
func (cl *wsClient) close() {
	cl.closeOnce.Do(func() {
		close(cl.send)
		_ = cl.conn.Close()
	})
}
