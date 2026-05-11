// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package interpreter

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// wsClient owns the writer goroutine for one attached WebSocket connection.
// The reader goroutine exists only to keep gorilla's PingHandler alive (incoming
// frames are discarded — execute requests come through the request-line handler
// at attach time, not over the persistent stream).
type wsClient struct {
	id        string
	conn      *websocket.Conn
	send      chan wsFrame
	done      chan struct{}
	closeOnce sync.Once
	logger    logTarget
}

func (cl *wsClient) writer() {
	defer close(cl.done)
	for {
		select {
		case frame, ok := <-cl.send:
			if !ok {
				return
			}
			if err := cl.writeFrame(frame); err != nil {
				cl.logger.Debug("ws write failed", "error", err)
				return
			}
			if frame.close != nil {
				return
			}
		}
	}
}

func (cl *wsClient) reader() {
	for {
		if _, _, err := cl.conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (cl *wsClient) StartReader() {
	go cl.reader()
}

func (cl *wsClient) writeFrame(frame wsFrame) error {
	if err := cl.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		return err
	}
	if frame.close != nil {
		payload := websocket.FormatCloseMessage(frame.close.code, frame.close.message)
		return cl.conn.WriteMessage(websocket.CloseMessage, payload)
	}
	if frame.output == nil {
		return nil
	}
	data, err := json.Marshal(frame.output)
	if err != nil {
		return err
	}
	return cl.conn.WriteMessage(websocket.TextMessage, data)
}

func (cl *wsClient) requestClose(code int, message string) {
	cl.closeOnce.Do(func() {
		select {
		case cl.send <- wsFrame{close: &closeRequest{code: code, message: message}}:
		default:
			cl.logger.Warn("ws send channel full, force-closing", "id", cl.id)
		}
		// Best effort: ensure the underlying conn is closed soon even if the
		// writer goroutine has already exited.
		go func() {
			t := time.NewTimer(2 * time.Second)
			defer t.Stop()
			select {
			case <-cl.done:
			case <-t.C:
			}
			_ = cl.conn.Close()
		}()
	})
}

// AwaitDone blocks until the writer goroutine exits.
func (cl *wsClient) AwaitDone() {
	<-cl.done
}

// RequestClose is the exported version of requestClose for use from the server pkg.
func (cl *wsClient) RequestClose(code int, message string) {
	cl.requestClose(code, message)
}
