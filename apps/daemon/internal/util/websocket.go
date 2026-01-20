// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// UpgradeToWebSocket is a toolbox utility function that upgrades an HTTP connection to a WebSocket connection.
// It automatically extracts and accepts SDK version subprotocols (if present) from the request headers and accepts them during handshake.
// It uses a permissive CORS (CheckOrigin always returns true) to allow connections from any origin.
func UpgradeToWebSocket(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	// Extract SDK version subprotocol from request headers
	subprotocol := ExtractSdkVersionSubprotocol(r.Header)
	var protocols []string
	if subprotocol != "" {
		protocols = []string{subprotocol}
	}

	// Create a new upgrader for this request to prevent concurrency issues
	upgrader := websocket.Upgrader{
		CheckOrigin:  func(r *http.Request) bool { return true },
		Subprotocols: protocols,
	}

	// Upgrade the connection to a WebSocket protocol
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return ws, nil
}
