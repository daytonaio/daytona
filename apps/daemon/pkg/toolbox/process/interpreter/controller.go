// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package interpreter

import (
	"net/http"
	"time"

	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// NewInterpreterController creates a new interpreter controller
func NewInterpreterController(workDir string) *InterpreterController {
	return &InterpreterController{workDir: workDir}
}

// Execute executes code in a Python interpreter with automatic session management
func (c *InterpreterController) Execute(ctx *gin.Context) {
    // Upgrade to websocket first
    ws, err := interpreterUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
    if err != nil {
        ctx.AbortWithStatus(http.StatusBadRequest)
        return
    }

    // Read first message as JSON payload with code/timeout/envs
    _, payload, err := ws.ReadMessage()
    if err != nil {
        _ = ws.WriteMessage(websocket.TextMessage, []byte(`{"type":"`+ChunkTypeError+`","name":"ProtocolError","value":"failed to read first message","traceback":""}`))
        _ = ws.Close()
        return
    }
    var req InterpreterExecuteRequest
    if err := json.Unmarshal(payload, &req); err != nil {
        _ = ws.WriteMessage(websocket.TextMessage, []byte(`{"type":"`+ChunkTypeError+`","name":"ProtocolError","value":"invalid JSON payload","traceback":""}`))
        _ = ws.Close()
        return
    }
    if req.Code == "" {
        _ = ws.WriteMessage(websocket.TextMessage, []byte(`{"type":"`+ChunkTypeError+`","name":"ValidationError","value":"code is required","traceback":""}`))
        _ = ws.Close()
        return
    }

	// Determine timeout (nil or 0 -> no timeout). If defined, must be >= 0
	var timeout time.Duration
	if req.Timeout != nil {
    timeout = time.Duration(*req.Timeout) * time.Second
	} else {
		timeout = 0
	}

	// Ensure a session exists (create only if none running)
	session, err := GetOrCreateSession(c.workDir)
    if err != nil {
        _ = ws.WriteMessage(websocket.TextMessage, []byte(`{"type":"`+ChunkTypeError+`","name":"SessionError","value":"failed to start interpreter session","traceback":""}`))
        _ = ws.Close()
        return
    }

	// Verify session is ready
	sessionInfo := session.Info()
    if !sessionInfo.Active {
        _ = ws.WriteMessage(websocket.TextMessage, []byte(`{"type":"`+ChunkTypeError+`","name":"SessionError","value":"interpreter session is not active","traceback":""}`))
        _ = ws.Close()
        return
    }

    // Enqueue execution; queue will attach ws for the duration of this job
    go session.enqueueAndExecute(req.Code, req.Envs, timeout, ws)
}
