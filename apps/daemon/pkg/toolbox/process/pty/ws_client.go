// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package pty

// writeMessage serializes all writes to the underlying websocket connection.
// gorilla/websocket does not support concurrent writers.
func (cl *wsClient) writeMessage(messageType int, data []byte) error {
	cl.writeMu.Lock()
	defer cl.writeMu.Unlock()
	return cl.conn.WriteMessage(messageType, data)
}

func (cl *wsClient) close() {
	cl.closeOnce.Do(func() {
		close(cl.done)
		_ = cl.conn.Close()
	})
}
