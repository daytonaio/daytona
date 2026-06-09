// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package pty

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/internal/util"
	"github.com/daytonaio/daemon/pkg/common"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	cmap "github.com/orcaman/concurrent-map/v2"
)

// NewPTYController creates a new PTY controller
func NewPTYController(logger *slog.Logger, workDir string) *PTYController {
	return &PTYController{logger: logger.With(slog.String("component", "PTY_controller")), workDir: workDir}
}

// CreatePTYSession godoc
//
//	@Summary		Create a new PTY session
//	@Description	Create a new pseudo-terminal session with specified configuration
//	@Tags			process
//	@Accept			json
//	@Produce		json
//	@Param			request	body		PTYCreateRequest	true	"PTY session creation request"
//	@Success		201		{object}	PTYCreateResponse
//	@Failure		400		{object}	common.ErrorResponse
//	@Failure		409		{object}	common.ErrorResponse
//	@Failure		500		{object}	common.ErrorResponse
//	@Router			/process/pty [post]
//
//	@id				CreatePtySession
func (p *PTYController) CreatePTYSession(c *gin.Context) {
	var req PTYCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(common_errors.NewInvalidBodyRequestError(err))
		return
	}

	// Validate session ID
	if req.ID == "" {
		c.Error(common_errors.NewBadRequestError(errors.New("session ID is required")))
		return
	}

	// Check if session with this ID already exists
	if _, exists := ptyManager.Get(req.ID); exists {
		c.Error(common_errors.NewConflictError(fmt.Errorf("PTY session with ID '%s' already exists", req.ID)))
		return
	}

	// Defaults
	if req.Cwd == "" {
		req.Cwd = p.workDir
	}
	if req.Envs == nil {
		req.Envs = make(map[string]string, 1)
	}
	if req.Envs["TERM"] == "" {
		req.Envs["TERM"] = "xterm-256color"
	}
	if req.Cols == nil {
		req.Cols = util.Pointer(uint16(80))
	}
	if req.Rows == nil {
		req.Rows = util.Pointer(uint16(24))
	}
	// Set upper limits to avoid ioctl errors
	if *req.Cols > 1000 {
		c.Error(common_errors.NewBadRequestError(errors.New("invalid value for cols - must be less than 1000")))
		return
	}
	if *req.Rows > 1000 {
		c.Error(common_errors.NewBadRequestError(errors.New("invalid value for rows - must be less than 1000")))
		return
	}

	session := &PTYSession{
		info: PTYSessionInfo{
			ID:        req.ID,
			Cwd:       req.Cwd,
			Envs:      req.Envs,
			Cols:      *req.Cols,
			Rows:      *req.Rows,
			CreatedAt: time.Now(),
			Active:    false,
			LazyStart: req.LazyStart,
		},
		clients: cmap.New[*wsClient](),
		logger:  p.logger.With(slog.String("sessionId", req.ID)),
	}

	// Add to manager first to prevent race conditions
	ptyManager.Add(session)

	// Start PTY immediately if not lazy start (default behavior)
	if !req.LazyStart {
		if err := session.start(); err != nil {
			// If start fails, remove from manager
			ptyManager.Delete(req.ID)
			p.logger.Error("failed to start PTY at create", "error", err)
			c.Error(common_errors.NewInternalServerError(errors.New("failed to start PTY session")))
			return
		}
	}

	c.JSON(http.StatusCreated, PTYCreateResponse{SessionID: req.ID})
}

// ListPTYSessions godoc
//
//	@Summary		List all PTY sessions
//	@Description	Get a list of all active pseudo-terminal sessions
//	@Tags			process
//	@Produce		json
//	@Success		200	{object}	PTYListResponse
//	@Failure		500	{object}	common.ErrorResponse
//	@Router			/process/pty [get]
//
//	@id				ListPtySessions
func (p *PTYController) ListPTYSessions(c *gin.Context) {
	c.JSON(http.StatusOK, PTYListResponse{Sessions: ptyManager.List()})
}

// GetPTYSession godoc
//
//	@Summary		Get PTY session information
//	@Description	Get detailed information about a specific pseudo-terminal session
//	@Tags			process
//	@Produce		json
//	@Param			sessionId	path		string	true	"PTY session ID"
//	@Success		200			{object}	PTYSessionInfo
//	@Failure		400			{object}	common.ErrorResponse
//	@Failure		404			{object}	common.ErrorResponse
//	@Router			/process/pty/{sessionId} [get]
//
//	@id				GetPtySession
func (p *PTYController) GetPTYSession(c *gin.Context) {
	id := c.Param("sessionId")
	if id == "" {
		c.Error(common_errors.NewBadRequestError(errors.New("session ID is required")))
		return
	}

	if s, ok := ptyManager.Get(id); ok {
		c.JSON(http.StatusOK, s.Info())
		return
	}
	c.Error(common.NewProcessNotFoundError("PTY session not found"))
}

// DeletePTYSession godoc
//
//	@Summary		Delete a PTY session
//	@Description	Delete a pseudo-terminal session and terminate its process
//	@Tags			process
//	@Produce		json
//	@Param			sessionId	path		string	true	"PTY session ID"
//	@Success		200			{object}	gin.H
//	@Failure		400			{object}	common.ErrorResponse
//	@Failure		404			{object}	common.ErrorResponse
//	@Router			/process/pty/{sessionId} [delete]
//
//	@id				DeletePtySession
func (p *PTYController) DeletePTYSession(c *gin.Context) {
	id := c.Param("sessionId")
	if id == "" {
		c.Error(common_errors.NewBadRequestError(errors.New("session ID is required")))
		return
	}

	if s, ok := ptyManager.Delete(id); ok {
		s.kill()
		p.logger.Debug("Deleted PTY session", "sessionId", id)
		c.JSON(http.StatusOK, gin.H{"message": "PTY session deleted"})
		return
	}
	c.Error(common.NewProcessNotFoundError("PTY session not found"))
}

// ConnectPTYSession godoc
//
//	@Summary		Connect to PTY session via WebSocket
//	@Description	Establish a WebSocket connection to interact with a pseudo-terminal session
//	@Tags			process
//	@Param			sessionId	path	string	true	"PTY session ID"
//	@Success		101			"Switching Protocols - WebSocket connection established"
//	@Failure		400			{object}	common.ErrorResponse
//	@Router			/process/pty/{sessionId}/connect [get]
//
//	@id				ConnectPtySession
func (p *PTYController) ConnectPTYSession(c *gin.Context) {
	id := c.Param("sessionId")
	if id == "" {
		c.Error(common_errors.NewBadRequestError(errors.New("session ID is required")))
		return
	}

	// Upgrade to WebSocket
	ws, err := util.UpgradeToWebSocket(c.Writer, c.Request)
	if err != nil {
		p.logger.Error("ws upgrade failed", "error", err)
		return
	}

	session, err := ptyManager.VerifyPTYSessionReady(id)
	if err != nil {
		p.logger.Debug("failed to connect to PTY session", "sessionId", id, "error", err)
		// Send error control message
		errorMsg := map[string]interface{}{
			"type":   "control",
			"status": "error",
			"error":  "Failed to connect to PTY session: " + err.Error(),
		}
		if errorJSON, err := json.Marshal(errorMsg); err == nil {
			_ = ws.WriteMessage(websocket.TextMessage, errorJSON)
		}
		_ = ws.Close()
		return
	}

	// Attach to session - this will send the control message internally
	session.attachWebSocket(ws)
}

// ResizePTYSession godoc
//
//	@Summary		Resize a PTY session
//	@Description	Resize the terminal dimensions of a pseudo-terminal session
//	@Tags			process
//	@Accept			json
//	@Produce		json
//	@Param			sessionId	path		string				true	"PTY session ID"
//	@Param			request		body		PTYResizeRequest	true	"Resize request with new dimensions"
//	@Success		200			{object}	PTYSessionInfo
//	@Failure		400			{object}	common.ErrorResponse
//	@Failure		410			{object}	common.ErrorResponse
//	@Failure		500			{object}	common.ErrorResponse
//	@Router			/process/pty/{sessionId}/resize [post]
//
//	@id				ResizePtySession
func (p *PTYController) ResizePTYSession(c *gin.Context) {
	id := c.Param("sessionId")
	if id == "" {
		c.Error(common_errors.NewBadRequestError(errors.New("session ID is required")))
		return
	}

	var req PTYResizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(common_errors.NewInvalidBodyRequestError(err))
		return
	}

	if req.Cols > 1000 {
		c.Error(common_errors.NewBadRequestError(errors.New("cols must be less than 1000")))
		return
	}
	if req.Rows > 1000 {
		c.Error(common_errors.NewBadRequestError(errors.New("rows must be less than 1000")))
		return
	}

	session, err := ptyManager.VerifyPTYSessionForResize(id)
	if err != nil {
		c.Error(common.NewSessionEndedError(err.Error()))
		return
	}

	if err := session.resize(req.Cols, req.Rows); err != nil {
		c.Error(common_errors.NewInternalServerError(err))
		return
	}
	p.logger.Debug("Resized PTY session", "sessionId", id, "cols", req.Cols, "rows", req.Rows)

	// Return updated session info
	updatedInfo := session.Info()
	c.JSON(http.StatusOK, updatedInfo)
}
