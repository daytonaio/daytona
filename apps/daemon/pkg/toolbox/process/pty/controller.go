// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package pty

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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
	if req.Cols <= 0 {
		req.Cols = 80
	}
	if req.Rows <= 0 {
		req.Rows = 24
	}
	// clamp extremes to avoid ioctl errors
	if req.Cols > 1000 {
		req.Cols = 1000
	}
	if req.Rows > 1000 {
		req.Rows = 1000
	}

	session := &PTYSession{
		info: PTYSessionInfo{
			ID:        req.ID,
			Cwd:       req.Cwd,
			Envs:      req.Envs,
			Cols:      req.Cols,
			Rows:      req.Rows,
			CreatedAt: time.Now(),
			Active:    false,
			LazyStart: req.LazyStart,
		},
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
		log.Infof("Deleted PTY session %s", id)
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

	// Validate session existence and send control message
	session, ok := ptyManager.Get(id)
	if !ok {
		log.Warnf("PTY session %s not found", id)
		// Send error control message
		errorMsg := map[string]interface{}{
			"type":   "control",
			"status": "error",
			"error":  "PTY session not found",
		}
		if errorJSON, err := json.Marshal(errorMsg); err == nil {
			_ = ws.WriteMessage(websocket.TextMessage, errorJSON)
		}
		_ = ws.Close()
		return
	}

	sessionInfo := session.Info()

	// Handle inactive sessions based on lazy start flag
	if !sessionInfo.Active {
		if sessionInfo.LazyStart {
			// Lazy start session - start PTY on first client connection
			log.Infof("Starting lazy PTY session %s on first client connection", id)
			if err := session.start(); err != nil {
				log.WithError(err).Errorf("Failed to start lazy PTY session %s", id)
				// Send error control message
				errorMsg := map[string]interface{}{
					"type":   "control",
					"status": "error",
					"error":  "Failed to start PTY session",
				}
				if errorJSON, err := json.Marshal(errorMsg); err == nil {
					_ = ws.WriteMessage(websocket.TextMessage, errorJSON)
				}
				_ = ws.Close()
				return
			}
		} else {
			// Non-lazy session that's inactive means it has terminated
			log.Warnf("PTY session %s has terminated and is no longer available", id)
			// Send error control message
			errorMsg := map[string]interface{}{
				"type":   "control",
				"status": "error",
				"error":  fmt.Sprintf("PTY session '%s' has terminated and is no longer available", id),
			}
			if errorJSON, err := json.Marshal(errorMsg); err == nil {
				_ = ws.WriteMessage(websocket.TextMessage, errorJSON)
			}
			_ = ws.Close()
			return
		}
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

	session, ok := ptyManager.Get(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "PTY session not found"})
		return
	}

	sessionInfo := session.Info()

	// Check if session can be resized
	if !sessionInfo.Active {
		if sessionInfo.LazyStart {
			// Lazy start session not yet started - allow resize (will update session info)
			log.Infof("Resizing lazy PTY session %s before it starts", id)
		} else {
			// Non-lazy session that's inactive means it has terminated
			c.JSON(http.StatusGone, gin.H{"error": fmt.Sprintf("PTY session '%s' has terminated and cannot be resized", id)})
			return
		}
	}

	var req PTYResizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := session.resize(req.Cols, req.Rows); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Infof("Resized PTY session %s to %dx%d", id, req.Cols, req.Rows)

	// Return updated session info
	updatedInfo := session.Info()
	c.JSON(http.StatusOK, updatedInfo)
}
