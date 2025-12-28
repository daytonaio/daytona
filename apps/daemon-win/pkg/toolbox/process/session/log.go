// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"errors"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// GetSessionCommandLogs godoc
//
//	@Summary		Get session command logs
//	@Description	Get logs for a specific command within a session.
//	@Tags			process
//	@Produce		text/plain
//	@Param			sessionId	path		string	true	"Session ID"
//	@Param			commandId	path		string	true	"Command ID"
//	@Success		200			{string}	string	"Log content"
//	@Router			/process/session/{sessionId}/command/{commandId}/logs [get]
//
//	@id				GetSessionCommandLogs
func (s *SessionController) GetSessionCommandLogs(c *gin.Context) {
	sessionId := c.Param("sessionId")
	cmdId := c.Param("commandId")

	session, ok := sessions[sessionId]
	if !ok {
		c.AbortWithError(http.StatusNotFound, errors.New("session not found"))
		return
	}

	command, ok := sessions[sessionId].commands[cmdId]
	if !ok {
		c.AbortWithError(http.StatusNotFound, errors.New("command not found"))
		return
	}

	logFilePath, _ := command.LogFilePath(session.Dir(s.configDir))

	logBytes, err := os.ReadFile(logFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		if os.IsPermission(err) {
			c.AbortWithError(http.StatusForbidden, err)
			return
		}
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.String(http.StatusOK, string(logBytes))
}
