// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPtyHandleAccessorsAndDelegates(t *testing.T) {
	exitCode := 7
	errMsg := "failed"
	handle := &PtyHandle{
		sessionID: "pty-1",
		exitCode:  &exitCode,
		err:       &errMsg,
		dataChan:  make(chan []byte, 1),
		done:      make(chan struct{}),
		handleResize: func(context.Context, int, int) (*types.PtySessionInfo, error) {
			return &types.PtySessionInfo{ID: "pty-1", Cols: 120, Rows: 40}, nil
		},
		handleKill: func(context.Context) error { return nil },
	}

	assert.Equal(t, "pty-1", handle.SessionID())
	assert.Equal(t, &exitCode, handle.ExitCode())
	assert.Equal(t, &errMsg, handle.Error())
	info, err := handle.Resize(context.Background(), 120, 40)
	require.NoError(t, err)
	assert.Equal(t, 120, info.Cols)
	require.NoError(t, handle.Kill(context.Background()))
}

func TestPtyHandleWaitForConnectionBehaviors(t *testing.T) {
	t.Run("already established returns immediately", func(t *testing.T) {
		handle := &PtyHandle{connectionEstablished: true}
		require.NoError(t, handle.WaitForConnection(context.Background()))
	})

	t.Run("returns stored connection error", func(t *testing.T) {
		msg := "connection refused"
		handle := &PtyHandle{err: &msg}
		err := handle.WaitForConnection(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), msg)
	})

	t.Run("times out when never connected", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		handle := &PtyHandle{}
		err := handle.WaitForConnection(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "PTY connection timeout")
	})
}

func TestPtyHandleSendInputWriteReadAndDisconnect(t *testing.T) {
	received := make(chan []byte, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer conn.Close()
		require.NoError(t, conn.WriteJSON(controlMessage{Type: "control", Status: "connected"}))
		_, payload, err := conn.ReadMessage()
		require.NoError(t, err)
		received <- payload
		require.NoError(t, conn.WriteMessage(websocket.BinaryMessage, []byte("terminal output")))
		require.NoError(t, conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, `{"exitCode":0}`), time.Now().Add(time.Second)))
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[len("http"):]
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	handle := newPtyHandle(conn, "pty-1", func(context.Context, int, int) (*types.PtySessionInfo, error) {
		return &types.PtySessionInfo{}, nil
	}, func(context.Context) error { return nil })

	require.NoError(t, handle.WaitForConnection(context.Background()))
	_, err = handle.Write([]byte("echo hello\n"))
	require.NoError(t, err)
	assert.Equal(t, []byte("echo hello\n"), <-received)
	buf := make([]byte, 32)
	n, err := handle.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, "terminal output", string(buf[:n]))
	require.NoError(t, handle.Disconnect())
	assert.False(t, handle.IsConnected())
}

func TestPtyHandleReadEOFAndWriteWithoutConnection(t *testing.T) {
	handle := &PtyHandle{dataChan: make(chan []byte)}
	close(handle.dataChan)
	buf := make([]byte, 8)
	_, err := handle.Read(buf)
	assert.ErrorIs(t, err, io.EOF)

	_, err = (&PtyHandle{}).Write([]byte("test"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "PTY is not connected")
}

func TestPtyHandleControlAndCloseParsing(t *testing.T) {
	t.Run("control message updates connection state", func(t *testing.T) {
		handle := &PtyHandle{}
		handle.handleControlMessage(&controlMessage{Status: "connected"})
		assert.True(t, handle.connectionEstablished)
		assert.True(t, handle.connected)
		handle.handleControlMessage(&controlMessage{Status: "error", Error: "bad"})
		require.NotNil(t, handle.Error())
		assert.Equal(t, "bad", *handle.Error())
	})

	t.Run("close message parses json and defaults", func(t *testing.T) {
		tests := []struct {
			name         string
			reason       string
			exitCode     *int
			errorMessage *string
		}{
			{name: "empty reason", reason: "", exitCode: intPtr(0)},
			{name: "json exit code", reason: `{"exitCode":3}`, exitCode: intPtr(3)},
			{name: "json exit reason", reason: `{"exitCode":2,"exitReason":"terminated"}`, exitCode: intPtr(2), errorMessage: strPtr("terminated")},
			{name: "invalid json defaults zero", reason: "not-json", exitCode: intPtr(0)},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				handle := &PtyHandle{}
				handle.handleCloseMessage(tt.reason)
				require.NotNil(t, handle.ExitCode())
				assert.Equal(t, *tt.exitCode, *handle.ExitCode())
				if tt.errorMessage != nil {
					require.NotNil(t, handle.Error())
					assert.Equal(t, *tt.errorMessage, *handle.Error())
				}
			})
		}
	})
}

func TestPtyHandleWaitLifecycle(t *testing.T) {
	t.Run("returns result when exit code already set", func(t *testing.T) {
		exitCode := 0
		handle := &PtyHandle{exitCode: &exitCode}
		result, err := handle.Wait(context.Background())
		require.NoError(t, err)
		require.NotNil(t, result.ExitCode)
		assert.Equal(t, 0, *result.ExitCode)
	})

	t.Run("returns stored error before exit code", func(t *testing.T) {
		msg := "broken pipe"
		handle := &PtyHandle{err: &msg, done: make(chan struct{})}
		_, err := handle.Wait(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), msg)
	})

	t.Run("returns context error when canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		handle := &PtyHandle{done: make(chan struct{})}
		_, err := handle.Wait(ctx)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func TestPtyHandleHandleMessagesEndToEnd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer conn.Close()
		require.NoError(t, conn.WriteJSON(controlMessage{Type: "control", Status: "connected"}))
		require.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte("plain-text")))
		require.NoError(t, conn.WriteMessage(websocket.BinaryMessage, []byte("plain-binary")))
		require.NoError(t, conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, `{"exitCode":4,"error":"boom"}`), time.Now().Add(time.Second)))
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[len("http"):]
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	handle := newPtyHandle(conn, "pty-2", func(context.Context, int, int) (*types.PtySessionInfo, error) {
		return &types.PtySessionInfo{}, nil
	}, func(context.Context) error { return nil })

	require.NoError(t, handle.WaitForConnection(context.Background()))
	assert.Equal(t, "plain-text", string(<-handle.DataChan()))
	assert.Equal(t, "plain-binary", string(<-handle.DataChan()))
	result, err := handle.Wait(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.ExitCode)
	assert.Equal(t, 4, *result.ExitCode)
	require.NotNil(t, result.Error)
	assert.Equal(t, "boom", *result.Error)
}
