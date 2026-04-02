// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package pty

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
)

var testUpgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

// newTestWSPair creates a connected server/client websocket pair and returns the server-side conn.
func newTestWSPair(t *testing.T) *websocket.Conn {
	t.Helper()
	var serverConn *websocket.Conn
	ready := make(chan struct{})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		serverConn, err = testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade: %v", err)
		}
		close(ready)
		select {} // keep alive until server closes
	}))
	t.Cleanup(srv.Close)

	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	clientConn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { clientConn.Close() })

	// Drain messages on the client side so the server doesn't block
	go func() {
		for {
			if _, _, err := clientConn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	<-ready
	return serverConn
}

// TestWriteMessageConcurrent verifies that concurrent calls to writeMessage
// do not panic or race. Without the writeMu this reliably triggers
// "concurrent write to websocket connection".
func TestWriteMessageConcurrent(t *testing.T) {
	conn := newTestWSPair(t)
	cl := &wsClient{
		id:   "test",
		conn: conn,
		send: make(chan []byte, 256),
		done: make(chan struct{}),
	}

	const goroutines = 20
	const writes = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < writes; j++ {
				_ = cl.writeMessage(websocket.TextMessage, []byte("hello"))
			}
		}()
	}

	wg.Wait()
}

// TestWriteMessageAndCloseConcurrent simulates the race between clientWriter
// sending data and closeClientsWithExitCode sending a close frame — the exact
// scenario from the production panic.
func TestWriteMessageAndCloseConcurrent(t *testing.T) {
	conn := newTestWSPair(t)
	cl := &wsClient{
		id:   "test",
		conn: conn,
		send: make(chan []byte, 256),
		done: make(chan struct{}),
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Simulate clientWriter sending PTY data
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			if err := cl.writeMessage(websocket.BinaryMessage, []byte("pty output")); err != nil {
				return
			}
		}
	}()

	// Simulate closeClientsWithExitCode sending a close frame mid-stream
	go func() {
		defer wg.Done()
		_ = cl.writeMessage(websocket.CloseMessage, websocket.FormatCloseMessage(
			websocket.CloseNormalClosure, `{"exitCode":0}`),
		)
		cl.close()
	}()

	wg.Wait()
}
