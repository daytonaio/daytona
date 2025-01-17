// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

var sessions = map[string]*Session{}

func CreateSession(configDir string) func(c *gin.Context) {
	return func(c *gin.Context) {
		cmd := exec.Command("/bin/sh")

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

		outReader, outWriter := io.Pipe()

		cmd.Stdout = outWriter
		cmd.Stderr = outWriter

		err = cmd.Start()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		err = os.MkdirAll(filepath.Join(configDir, "sessions", request.SessionId), 0755)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		sessions[request.SessionId] = &Session{
			Cmd:         cmd,
			Alias:       request.Alias,
			OutReader:   bufio.NewReader(outReader),
			StdinWriter: stdinWriter,
		}

		c.Status(http.StatusCreated)
	}
}

func DeleteSession(configDir string) func(c *gin.Context) {
	return func(c *gin.Context) {
		sessionId := c.Param("sessionId")

		session, ok := sessions[sessionId]
		if !ok {
			c.AbortWithError(404, errors.New("session not found"))
			return
		}

		err := session.Cmd.Process.Kill()
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		delete(sessions, sessionId)

		err = os.RemoveAll(filepath.Join(configDir, "sessions", sessionId))
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func ListSessions(c *gin.Context) {
	sessionDTOs := []SessionDTO{}

	for sessionId, session := range sessions {
		sessionDTOs = append(sessionDTOs, SessionDTO{
			SessionId: sessionId,
			Alias:     session.Alias,
		})
	}

	c.JSON(http.StatusOK, sessionDTOs)
}
