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
		send: make(chan wsFrame, 1024),
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
		case frame, ok := <-cl.send:
			if !ok {
				return
			}

			err := cl.writeFrame(frame)
			if err != nil {
				log.Debugf("Failed to write frame: %v", err)
			}
			if frame.close != nil {
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
	case cl.send <- wsFrame{output: msg}:
	default:
		log.Debug("Client send channel full - closing slow consumer")
		cl.requestClose(websocket.ClosePolicyViolation, "slow consumer")

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

		// Wait for clientWriter to drain remaining messages with a timeout
		// This ensures close frames and other pending messages have time to be sent
		timer := time.NewTimer(5 * time.Second)
		select {
		case <-cl.done:
			// clientWriter has finished processing all messages
			if !timer.Stop() {
				<-timer.C
			}
		case <-timer.C:
			// Timeout reached, proceed with closing
			log.Debug("Timeout waiting for client writer to finish")
		}

		// Wait for client's close frame response (proper WebSocket handshake)
		// Set a read deadline to prevent hanging if client doesn't respond
		_ = cl.conn.SetReadDeadline(time.Now().Add(5 * time.Second))

		// Drain any remaining messages until we get close frame or timeout
		// This ensures proper WebSocket close handshake per RFC 6455
		for {
			_, _, err := cl.conn.NextReader()
			if err != nil {
				break
			}
		}

		_ = cl.conn.Close()
	})
}

func (cl *wsClient) writeFrame(frame wsFrame) error {
	if frame.output == nil && frame.close == nil {
		return nil
	}

	err := cl.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err != nil {
		return err
	}

	if frame.close != nil {
		payload := websocket.FormatCloseMessage(frame.close.code, frame.close.message)
		return cl.conn.WriteMessage(websocket.CloseMessage, payload)
	}

	data, err := json.Marshal(frame.output)
	if err != nil {
		return err
	}

	return cl.conn.WriteMessage(websocket.TextMessage, data)
}

func (cl *wsClient) requestClose(code int, message string) {
	frame := wsFrame{
		close: &closeRequest{
			code:    code,
			message: message,
		},
	}

	select {
	case cl.send <- frame:
	default:
		log.Debug("Couldn't send close frame to client - closing connection")
	}

	cl.close()
}
