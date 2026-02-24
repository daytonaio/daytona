// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package interpreter

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func NewInterpreterController(logger *slog.Logger, workDir string) *Controller {
	InitManager(workDir)
	// Pre-warm the default interpreter context to reduce latency on first request
	go func() {
		_, err := GetOrCreateDefaultContext(logger)
		if err != nil {
			logger.Debug("Failed to pre-create default interpreter context", "error", err)
		}
	}()
	return &Controller{logger: logger.With(slog.String("component", "interpreter_controller")), workDir: workDir}
}

// CreateContext creates a new interpreter context
//
//	@Summary		Create a new interpreter context
//	@Description	Creates a new isolated interpreter context with optional working directory and language
//	@Tags			interpreter
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CreateContextRequest	true	"Context configuration"
//	@Success		200		{object}	InterpreterContext
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/process/interpreter/context [post]
//
//	@id				CreateInterpreterContext
func (c *Controller) CreateContext(ctx *gin.Context) {
	var req CreateContextRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request payload: %w", err))
		return
	}

	language := LanguagePython
	if req.Language != nil {
		language = *req.Language
	}
	if language != LanguagePython {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("unsupported language: %s (only '%s' is supported currently)", language, LanguagePython))
		return
	}

	cwd := c.workDir
	if req.Cwd != nil {
		cwd = *req.Cwd
	}

	iCtx, err := CreateContext(c.logger, cwd, language)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	info := iCtx.Info()
	ctx.JSON(http.StatusOK, ContextInfo{
		ID:        info.ID,
		Cwd:       info.Cwd,
		CreatedAt: info.CreatedAt,
		Active:    info.Active,
		Language:  info.Language,
	})
}

// Execute executes code in an interpreter context via WebSocket
//
//	@Summary		Execute code in an interpreter context
//	@Description	Executes code in a specified context (or default context if not specified) via WebSocket streaming
//	@Tags			interpreter
//	@Accept			json
//	@Produce		json
//	@Router			/process/interpreter/execute [get]
//	@Success		101	{string}	string		"Switching Protocols"
//	@Header			101	{string}	Upgrade		"websocket"
//	@Header			101	{string}	Connection	"Upgrade"
//
//	@id				ExecuteInterpreterCode
func (c *Controller) Execute(ctx *gin.Context) {
	// Upgrade to WebSocket
	ws, err := util.UpgradeToWebSocket(ctx.Writer, ctx.Request)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	_, payload, err := ws.ReadMessage()
	if err != nil {
		writeWSError(ws, "failed to read first message", websocket.CloseProtocolError)
		return
	}

	var req ExecuteRequest
	err = json.Unmarshal(payload, &req)
	if err != nil {
		writeWSError(ws, "invalid JSON payload", websocket.CloseProtocolError)
		return
	}

	if req.Code == "" {
		writeWSError(ws, "code is required", websocket.ClosePolicyViolation)
		return
	}

	timeout := 10 * time.Minute
	if req.Timeout != nil {
		if *req.Timeout < 0 {
			writeWSError(ws, "timeout must be greater than or equal to 0", websocket.ClosePolicyViolation)
			return
		}
		if *req.Timeout == 0 {
			timeout = 0
		} else {
			timeout = time.Duration(*req.Timeout) * time.Second
		}
	}

	var iCtx *Context

	if req.ContextID == nil {
		iCtx, err = GetOrCreateDefaultContext(c.logger)
		if err != nil {
			writeWSError(ws, "failed to get default context: "+err.Error(), websocket.CloseInternalServerErr)
			return
		}
	} else {
		iCtx, err = GetContext(*req.ContextID)
		if err != nil {
			writeWSError(ws, "context not found: "+*req.ContextID, websocket.ClosePolicyViolation)
			return
		}
	}

	contextInfo := iCtx.Info()
	if !contextInfo.Active {
		writeWSError(ws, "context is not active", websocket.ClosePolicyViolation)
		return
	}

	var envs map[string]string
	if req.Envs != nil {
		envs = *req.Envs
	}

	go iCtx.enqueueAndExecute(req.Code, envs, timeout, ws)
}

// writeWSError sends an error message to the WebSocket and closes the connection
func writeWSError(ws *websocket.Conn, value string, closeCode int) {
	closeMessage := websocket.FormatCloseMessage(closeCode, value)
	_ = ws.WriteControl(websocket.CloseMessage, closeMessage, time.Now().Add(writeWait))
	_ = ws.Close()
}

// DeleteContext deletes an interpreter context
//
//	@Summary		Delete an interpreter context
//	@Description	Deletes an interpreter context and shuts down its worker process
//	@Tags			interpreter
//	@Produce		json
//	@Param			id	path		string	true	"Context ID"
//	@Success		200	{object}	map[string]string
//	@Failure		400	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Router			/process/interpreter/context/{id} [delete]
//
//	@id				DeleteInterpreterContext
func (c *Controller) DeleteContext(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("context ID is required"))
		return
	}

	if id == "default" {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("cannot delete default context"))
		return
	}

	err := DeleteContext(id)
	if err != nil {
		if common_errors.IsNotFoundError(err) {
			ctx.AbortWithError(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Context deleted successfully"})
}

// ListContexts lists all user-created interpreter contexts (excludes default)
//
//	@Summary		List all user-created interpreter contexts
//	@Description	Returns information about all user-created interpreter contexts (excludes default context)
//	@Tags			interpreter
//	@Produce		json
//	@Success		200	{object}	ListContextsResponse
//	@Router			/process/interpreter/context [get]
//
//	@id				ListInterpreterContexts
func (c *Controller) ListContexts(ctx *gin.Context) {
	allContexts := ListContexts()

	userContexts := make([]ContextInfo, 0, len(allContexts))
	for _, context := range allContexts {
		if context.ID != "default" {
			userContexts = append(userContexts, context)
		}
	}

	ctx.JSON(http.StatusOK, ListContextsResponse{Contexts: userContexts})
}
