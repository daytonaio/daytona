// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var fileWatchUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WatchFiles handles WebSocket connections for file watching
func WatchFiles(c *gin.Context) {
	// Get parameters from query
	path := c.Query("path")
	if path == "" {
		c.AbortWithError(http.StatusBadRequest,
			fmt.Errorf("path parameter is required"))
		return
	}

	recursive := c.Query("recursive") == "true"

	// Upgrade HTTP connection to WebSocket
	ws, err := fileWatchUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorf("Failed to upgrade to WebSocket: %v", err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	defer ws.Close()

	// Create and start file watcher
	watcher := NewFileWatcher(path, recursive)
	if err := watcher.Start(); err != nil {
		log.Errorf("Failed to start file watcher: %v", err)
		if writeErr := ws.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseInternalServerErr,
				"Failed to start file watcher")); writeErr != nil {
			log.Errorf("Failed to write close message: %v", writeErr)
		}
		return
	}

	defer watcher.Stop()

	// Set up ping/pong to detect disconnected clients
	if err := ws.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
		log.Errorf("Failed to set read deadline: %v", err)
		return
	}
	ws.SetPongHandler(func(string) error {
		if err := ws.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			log.Errorf("Failed to set read deadline in pong handler: %v", err)
		}
		return nil
	})

	// Goroutine to handle ping messages
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	// Goroutine to read from WebSocket (to detect disconnections)
	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err,
					websocket.CloseGoingAway,
					websocket.CloseAbnormalClosure) {
					log.Errorf("WebSocket read error: %v", err)
				}
				return
			}
		}
	}()

	// Main event loop
	for {
		select {
		case event, ok := <-watcher.Events():
			if !ok {
				// Events channel closed, watcher stopped
				return
			}

			// Marshal event to JSON
			jsonData, err := json.Marshal(event)
			if err != nil {
				log.Errorf("Failed to marshal file event: %v", err)
				continue
			}

			// Send event to client
			if err := ws.WriteMessage(websocket.TextMessage, jsonData); err != nil {
				log.Errorf("Failed to write message to WebSocket: %v", err)
				return
			}

		case err, ok := <-watcher.Errors():
			if !ok {
				// Errors channel closed
				return
			}

			log.Errorf("File watcher error: %v", err)

			// Send error as close message and terminate connection
			if writeErr := ws.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseInternalServerErr,
					err.Error())); writeErr != nil {
				log.Errorf("Failed to write close message: %v", writeErr)
			}
			return

		case <-pingTicker.C:
			// Send ping to keep connection alive
			if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Errorf("Failed to send ping: %v", err)
				return
			}

		case <-readDone:
			// Client disconnected
			log.Debug("WebSocket client disconnected")
			return

		case <-c.Request.Context().Done():
			// Request context cancelled
			log.Debug("Request context cancelled")
			return
		}
	}
}
