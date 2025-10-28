// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package interpreter

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// NewInterpreterController creates a new interpreter controller
func NewInterpreterController(workDir string) *InterpreterController {
	// Initialize the context manager with default working directory
	InitManager(workDir)
	return &InterpreterController{workDir: workDir}
}

// CreateContext creates a new interpreter context
// @Summary Create a new interpreter context
// @Description Creates a new isolated interpreter context with optional working directory and language
// @Tags interpreter
// @Accept json
// @Produce json
// @Param request body CreateContextRequest true "Context configuration"
// @Success 200 {object} CreateContextResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /process/interpreter/context [post]
func (c *InterpreterController) CreateContext(ctx *gin.Context) {
	var req CreateContextRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Validate language
	if req.Language != "" && req.Language != "python" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported language. Only 'python' is supported."})
		return
	}

	// Use default cwd if not provided
	cwd := req.Cwd
	if cwd == "" {
		cwd = c.workDir
	}

	// Create the context
	session, err := CreateContext(cwd, req.Language)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return context information
	info := session.Info()
	ctx.JSON(http.StatusOK, CreateContextResponse{
		ID:        info.ID,
		Cwd:       info.Cwd,
		Language:  info.Language,
		CreatedAt: info.CreatedAt,
		Active:    info.Active,
	})
}

// Execute executes code in a Python interpreter context via WebSocket
// @Summary Execute Python code in an interpreter context
// @Description Executes Python code in a specified context (or default context if not specified) via WebSocket streaming
// @Tags interpreter
// @Accept json
// @Produce json
// @Router /process/interpreter/execute [get]
func (c *InterpreterController) Execute(ctx *gin.Context) {
	// Upgrade to websocket first
	ws, err := interpreterUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Read first message as JSON payload with code/timeout/envs/contextId
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

	// Determine timeout (nil or 0 -> no timeout)
	var timeout time.Duration
	if req.Timeout != nil {
		timeout = time.Duration(*req.Timeout) * time.Second
	} else {
		timeout = 0
	}

	// Get or create context
	var session *InterpreterSession
	if req.ContextID == "" {
		// Use default context
		session, err = GetOrCreateDefaultContext()
		if err != nil {
			_ = ws.WriteMessage(websocket.TextMessage, []byte(`{"type":"`+ChunkTypeError+`","name":"ContextError","value":"failed to get default context: `+err.Error()+`","traceback":""}`))
			_ = ws.Close()
			return
		}
	} else {
		// Use specified context
		session, err = GetContext(req.ContextID)
		if err != nil {
			_ = ws.WriteMessage(websocket.TextMessage, []byte(`{"type":"`+ChunkTypeError+`","name":"ContextError","value":"context not found: `+req.ContextID+`","traceback":""}`))
			_ = ws.Close()
			return
		}
	}

	// Verify session is ready
	sessionInfo := session.Info()
	if !sessionInfo.Active {
		_ = ws.WriteMessage(websocket.TextMessage, []byte(`{"type":"`+ChunkTypeError+`","name":"ContextError","value":"context is not active","traceback":""}`))
		_ = ws.Close()
		return
	}

	// Enqueue execution; queue will attach ws for the duration of this job
	go session.enqueueAndExecute(req.Code, req.Envs, timeout, ws)
}

// DeleteContext deletes an interpreter context
// @Summary Delete an interpreter context
// @Description Deletes an interpreter context and shuts down its worker process
// @Tags interpreter
// @Produce json
// @Param id path string true "Context ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /process/interpreter/context/{id} [delete]
func (c *InterpreterController) DeleteContext(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Context ID is required"})
		return
	}

	// Prevent deletion of default context
	if id == "default" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete default context"})
		return
	}

	err := DeleteContext(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Context deleted successfully"})
}

// ListContexts lists all user-created interpreter contexts (excludes default)
// @Summary List all user-created interpreter contexts
// @Description Returns information about all user-created interpreter contexts (excludes default context)
// @Tags interpreter
// @Produce json
// @Success 200 {array} InterpreterSessionInfo
// @Router /process/interpreter/context [get]
func (c *InterpreterController) ListContexts(ctx *gin.Context) {
	allContexts := ListContexts()
	
	// Filter out default context
	userContexts := make([]InterpreterSessionInfo, 0)
	for _, context := range allContexts {
		if context.ID != "default" {
			userContexts = append(userContexts, context)
		}
	}
	
	ctx.JSON(http.StatusOK, gin.H{"contexts": userContexts})
}
