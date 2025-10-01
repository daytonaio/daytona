// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package pty

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/daytonaio/daemon/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	cmap "github.com/orcaman/concurrent-map/v2"
	log "github.com/sirupsen/logrus"
)

// NewPTYController creates a new PTY controller
func NewPTYController(workDir string) *PTYController {
	return &PTYController{workDir: workDir}
}

// CreatePTYSession creates a new PTY session
func (p *PTYController) CreatePTYSession(c *gin.Context) {
	var req PTYCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate session ID
	if req.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session ID is required"})
		return
	}

	// Check if session with this ID already exists
	if _, exists := ptyManager.Get(req.ID); exists {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("PTY session with ID '%s' already exists", req.ID)})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid value for cols - must be less than 1000"})
		return
	}
	if *req.Rows > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid value for rows - must be less than 1000"})
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
	}

	// Add to manager first to prevent race conditions
	ptyManager.Add(session)

	// Start PTY immediately if not lazy start (default behavior)
	if !req.LazyStart {
		if err := session.start(); err != nil {
			// If start fails, remove from manager
			ptyManager.Delete(req.ID)
			log.WithError(err).Error("failed to start PTY at create")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start PTY session"})
			return
		}
	}

	c.JSON(http.StatusCreated, PTYCreateResponse{SessionID: req.ID})
}

// ListPTYSessions lists all PTY sessions
func (p *PTYController) ListPTYSessions(c *gin.Context) {
	c.JSON(http.StatusOK, PTYListResponse{Sessions: ptyManager.List()})
}

// GetPTYSession gets information about a specific PTY session
func (p *PTYController) GetPTYSession(c *gin.Context) {
	id := c.Param("sessionId")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session ID is required"})
		return
	}

	if s, ok := ptyManager.Get(id); ok {
		c.JSON(http.StatusOK, s.Info())
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "PTY session not found"})
}

// DeletePTYSession deletes a PTY session
func (p *PTYController) DeletePTYSession(c *gin.Context) {
	id := c.Param("sessionId")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session ID is required"})
		return
	}

	if s, ok := ptyManager.Delete(id); ok {
		s.kill()
		log.Debugf("Deleted PTY session %s", id)
		c.JSON(http.StatusOK, gin.H{"message": "PTY session deleted"})
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "PTY session not found"})
}

// ConnectPTYSession handles WebSocket connections to PTY sessions
func (p *PTYController) ConnectPTYSession(c *gin.Context) {
	id := c.Param("sessionId")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session ID is required"})
		return
	}

	// Always upgrade to WebSocket first
	ws, err := ptyUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.WithError(err).Error("ws upgrade failed")
		return
	}

	session, err := ptyManager.VerifyPTYSessionReady(id)
	if err != nil {
		log.Debugf("failed to connect to PTY session: %v", err)
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

// ResizePTYSession resizes a PTY session
func (p *PTYController) ResizePTYSession(c *gin.Context) {
	id := c.Param("sessionId")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session ID is required"})
		return
	}

	var req PTYResizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Cols > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cols must be less than 1000"})
		return
	}
	if req.Rows > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rows must be less than 1000"})
		return
	}

	session, err := ptyManager.VerifyPTYSessionForResize(id)
	if err != nil {
		c.JSON(http.StatusGone, gin.H{"error": err.Error()})
		return
	}

	if err := session.resize(req.Cols, req.Rows); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Debugf("Resized PTY session %s to %dx%d", id, req.Cols, req.Rows)

	// Return updated session info
	updatedInfo := session.Info()
	c.JSON(http.StatusOK, updatedInfo)
}
