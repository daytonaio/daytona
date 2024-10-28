// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"errors"
	"fmt"
	"maps"
	"net/http"
	"os"
	"os/exec"
	"slices"

	"github.com/daytonaio/daytona/pkg/common"
	"github.com/gin-gonic/gin"
)

var sessions = map[string]*session{}

func CreateSession(workspaceDir, configDir string) func(c *gin.Context) {
	return func(c *gin.Context) {
		cmd := exec.Command(common.GetShell())
		cmd.Env = os.Environ()
		cmd.Dir = workspaceDir

		var request CreateSessionRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
			return
		}

		if _, ok := sessions[request.SessionId]; ok {
			c.AbortWithError(http.StatusConflict, errors.New("session already exists"))
			return
		}

		stdinWriter, err := cmd.StdinPipe()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		err = cmd.Start()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		session := &session{
			id:          request.SessionId,
			cmd:         cmd,
			stdinWriter: stdinWriter,
			commands:    map[string]*Command{},
		}
		sessions[request.SessionId] = session

		err = os.MkdirAll(session.Dir(configDir), 0755)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Status(http.StatusCreated)
	}
}

func DeleteSession(configDir string) func(c *gin.Context) {
	return func(c *gin.Context) {
		sessionId := c.Param("sessionId")

		session, ok := sessions[sessionId]
		if !ok {
			c.AbortWithError(http.StatusNotFound, errors.New("session not found"))
			return
		}

		err := session.cmd.Process.Kill()
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		delete(sessions, sessionId)

		err = os.RemoveAll(session.Dir(configDir))
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func ListSessions(c *gin.Context) {
	sessionDTOs := []Session{}

	for sessionId, session := range sessions {
		sessionDTOs = append(sessionDTOs, Session{
			SessionId: sessionId,
			Commands:  slices.Collect(maps.Values(session.commands)),
		})
	}

	c.JSON(http.StatusOK, sessionDTOs)
}
