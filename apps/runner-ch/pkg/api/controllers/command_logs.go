// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/daytonaio/runner-ch/pkg/cloudhypervisor"
	"github.com/daytonaio/runner-ch/pkg/runner"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// ProxyCommandLogsStream handles streaming command logs via WebSocket
func ProxyCommandLogsStream(ctx *gin.Context) {
	r := runner.GetInstance()
	if r == nil {
		ctx.JSON(500, gin.H{"error": "Runner not initialized"})
		return
	}

	targetURL, err := getProxyTarget(ctx, r.CHClient)
	if err != nil {
		return // Error already sent
	}

	if ctx.Query("follow") != "true" {
		// Non-streaming request, use regular proxy
		proxyWithTransport(ctx, r.CHClient, targetURL)
		return
	}

	// For streaming, we need to handle WebSocket
	// In remote mode, this goes through SSH tunnel; in local mode, it connects directly
	fullTargetURL := strings.Replace(targetURL.String(), "http://", "ws://", 1)

	// Create WebSocket dialer using appropriate transport
	var dialer websocket.Dialer
	if r.CHClient.IsRemote() {
		// Remote mode: use SSH tunnel
		transport := cloudhypervisor.GetSSHTunnelTransport(r.CHClient.SSHHost, r.CHClient.SSHKeyPath)
		dialer = websocket.Dialer{
			NetDialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return transport.DialContext(ctx, network, addr)
			},
		}
	} else {
		// Local mode: use default dialer (direct connection)
		dialer = *websocket.DefaultDialer
	}

	ws, _, err := dialer.DialContext(context.Background(), fullTargetURL+"?follow=true", nil)
	if err != nil {
		ctx.JSON(400, gin.H{"error": fmt.Sprintf("failed to connect to WebSocket: %v", err)})
		return
	}

	ctx.Header("Content-Type", "application/octet-stream")

	ws.SetCloseHandler(func(code int, text string) error {
		if code == websocket.CloseNormalClosure {
			return nil
		}
		ctx.AbortWithStatus(code)
		return nil
	})

	defer ws.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, msg, err := ws.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					return
				}
				log.Errorf("Error reading WebSocket message: %v", err)
				ws.Close()
				return
			}

			_, err = ctx.Writer.Write(msg)
			if err != nil {
				log.Errorf("Error writing message: %v", err)
				ws.Close()
				return
			}
			ctx.Writer.Flush()
		}
	}
}

// ShouldProxyCommandLogs checks if the request is for command logs streaming
func ShouldProxyCommandLogs(path string) bool {
	// Match /process/session/{sessionId}/command/{commandId}/logs
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) >= 5 {
		return parts[0] == "process" && parts[1] == "session" && parts[3] == "command" && parts[len(parts)-1] == "logs"
	}
	return false
}
