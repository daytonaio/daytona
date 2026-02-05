// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"strings"

	"github.com/daytonaio/runner-android/pkg/runner"
	"github.com/gin-gonic/gin"
)

// ProxyCommandLogsStream handles streaming command logs via WebSocket
// Note: Command logs streaming is not supported for Cuttlefish
func ProxyCommandLogsStream(ctx *gin.Context) {
	r := runner.GetInstance()
	if r == nil {
		ctx.JSON(500, gin.H{"error": "Runner not initialized"})
		return
	}

	// For Cuttlefish, command logs streaming is not supported
	// Android devices use logcat for logging
	ctx.JSON(501, gin.H{
		"error": "Command logs streaming not supported for Cuttlefish",
		"hint":  "Use 'adb logcat' to view Android device logs",
	})
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
