// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package pty

import "time"

// writeMessage serializes all writes to the underlying websocket connection.
// gorilla/websocket does not support concurrent writers.
func (cl *wsClient) writeMessage(messageType int, data []byte) error {
	cl.writeMu.Lock()
	defer cl.writeMu.Unlock()
	// Bound the write so a slow/stuck client can't block direct writers such as
	// the exit control message and the close frame on the session exit path.
	_ = cl.conn.SetWriteDeadline(time.Now().Add(writeWait))
	return cl.conn.WriteMessage(messageType, data)
}

func (cl *wsClient) close() {
	cl.closeOnce.Do(func() {
		close(cl.done)
		_ = cl.conn.Close()
	})
}
