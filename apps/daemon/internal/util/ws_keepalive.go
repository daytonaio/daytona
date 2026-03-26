// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

// SetupWSKeepAlive configures WebSocket keepalive ping/pong handling on the
// given connection. It installs a custom PingHandler that queues pong payloads
// onto a buffered channel instead of writing to the connection directly. This
// avoids write-mutex contention between the PingHandler and the caller's
// single writer goroutine.
//
// The caller's writer goroutine must drain the returned channel via
// WritePendingPongs before each data write so that keepalive pongs are never
// delayed.
func SetupWSKeepAlive(conn *websocket.Conn, logger *slog.Logger) <-chan []byte {
	pongCh := make(chan []byte, 10)
	conn.SetPingHandler(func(message string) error {
		select {
		case pongCh <- []byte(message):
		default:
			logger.Warn("pong channel full, dropping pong response")
		}
		return nil
	})
	return pongCh
}

// WritePendingPongs drains all queued pong responses and writes them to the
// connection. This MUST be called from the single writer goroutine before each
// data write so that keepalive pongs are never delayed by data writes. Because
// only one goroutine writes to the conn, WriteControl acquires the
// gorilla/websocket write mutex instantly — no contention, no silent drops.
func WritePendingPongs(conn *websocket.Conn, pongCh <-chan []byte, deadline time.Duration, logger *slog.Logger) {
	for {
		select {
		case pongData := <-pongCh:
			if err := conn.WriteControl(websocket.PongMessage, pongData, time.Now().Add(deadline)); err != nil {
				logger.Debug("failed to write pong", "error", err)
				return
			}
		default:
			return
		}
	}
}
