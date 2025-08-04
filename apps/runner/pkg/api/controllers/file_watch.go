/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/daytonaio/common-go/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// ProxyFileWatchStream handles WebSocket proxying for file watching between client and daemon
func ProxyFileWatchStream(ctx *gin.Context) {
	targetURL, _, err := getProxyTarget(ctx)
	if err != nil {
		return
	}

	// Convert HTTP URL to WebSocket URL
	fullTargetURL := strings.Replace(targetURL.String(), "http://", "ws://", 1)

	// Forward all query parameters for file watching
	if ctx.Request.URL.RawQuery != "" {
		fullTargetURL += "?" + ctx.Request.URL.RawQuery
	}

	// Establish WebSocket connection to daemon
	ws, _, err := websocket.DefaultDialer.DialContext(context.Background(), fullTargetURL, nil)
	if err != nil {
		ctx.Error(errors.NewBadRequestError(fmt.Errorf("failed to create file watch connection: %w", err)))
		return
	}

	defer ws.Close()

	// Upgrade incoming connection to WebSocket
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	clientWS, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Errorf("Failed to upgrade client connection: %v", err)
		return
	}
	defer clientWS.Close()

	// Simple bidirectional message forwarding
	errChan := make(chan error, 2)

	// Forward messages from client to daemon
	go func() {
		defer func() { errChan <- nil }()
		for {
			messageType, message, err := clientWS.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Errorf("Client WebSocket read error: %v", err)
				}
				return
			}

			err = ws.WriteMessage(messageType, message)
			if err != nil {
				log.Errorf("Error writing to daemon WebSocket: %v", err)
				return
			}
		}
	}()

	// Forward messages from daemon to client
	go func() {
		defer func() { errChan <- nil }()
		for {
			messageType, message, err := ws.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					return
				}
				log.Errorf("Daemon WebSocket read error: %v", err)
				return
			}

			err = clientWS.WriteMessage(messageType, message)
			if err != nil {
				log.Errorf("Error writing to client WebSocket: %v", err)
				return
			}
		}
	}()

	// Wait for either connection to close or error
	select {
	case err := <-errChan:
		if err != nil {
			log.Errorf("WebSocket proxy error: %v", err)
		}
	case <-ctx.Done():
		log.Debug("Request context cancelled")
	}
}
