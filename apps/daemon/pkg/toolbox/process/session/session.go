// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/daytonaio/daemon/pkg/common"
	"github.com/gin-gonic/gin"
)

var sessions = map[string]*session{}

func (s *SessionController) CreateSession(c *gin.Context) {
	ctx, cancel := context.WithCancel(context.Background())

	cmd := exec.CommandContext(ctx, common.GetShell())
	cmd.Env = os.Environ()
	cmd.Dir = s.projectDir

	var request CreateSessionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		cancel()
		return
	}

	if _, ok := sessions[request.SessionId]; ok {
		c.AbortWithError(http.StatusConflict, errors.New("session already exists"))
		cancel()
		return
	}

	stdinWriter, err := cmd.StdinPipe()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		cancel()
		return
	}

	err = cmd.Start()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		cancel()
		return
	}

	session := &session{
		id:          request.SessionId,
		cmd:         cmd,
		stdinWriter: stdinWriter,
		commands:    map[string]*Command{},
		ctx:         ctx,
		cancel:      cancel,
	}
	sessions[request.SessionId] = session

	err = os.MkdirAll(session.Dir(s.configDir), 0755)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusCreated)
}

func (s *SessionController) DeleteSession(c *gin.Context) {
	sessionId := c.Param("sessionId")

	session, ok := sessions[sessionId]
	if !ok || session.deleted {
		c.AbortWithError(http.StatusNotFound, errors.New("session not found"))
		return
	}

	session.cancel()
	session.deleted = true

	err := os.RemoveAll(session.Dir(s.configDir))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (s *SessionController) ListSessions(c *gin.Context) {
	sessionDTOs := []Session{}

	for sessionId, session := range sessions {
		if session.deleted {
			continue
		}

		commands, err := s.getSessionCommands(sessionId)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		sessionDTOs = append(sessionDTOs, Session{
			SessionId: sessionId,
			Commands:  commands,
		})
	}

	c.JSON(http.StatusOK, sessionDTOs)
}

func (s *SessionController) GetSession(c *gin.Context) {
	sessionId := c.Param("sessionId")

	session, ok := sessions[sessionId]
	if !ok || session.deleted {
		c.AbortWithError(http.StatusNotFound, errors.New("session not found"))
		return
	}

	commands, err := s.getSessionCommands(sessionId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, Session{
		SessionId: sessionId,
		Commands:  commands,
	})
}

func (s *SessionController) GetSessionCommand(c *gin.Context) {
	sessionId := c.Param("sessionId")
	cmdId := c.Param("commandId")

	command, err := s.getSessionCommand(sessionId, cmdId)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}

	c.JSON(http.StatusOK, command)
}

func (s *SessionController) getSessionCommands(sessionId string) ([]*Command, error) {
	session, ok := sessions[sessionId]
	if !ok || session.deleted {
		return nil, errors.New("session not found")
	}

	commands := []*Command{}
	for _, command := range session.commands {
		cmd, err := s.getSessionCommand(sessionId, command.Id)
		if err != nil {
			return nil, err
		}
		commands = append(commands, cmd)
	}

	return commands, nil
}

func (s *SessionController) getSessionCommand(sessionId, cmdId string) (*Command, error) {
	session, ok := sessions[sessionId]
	if !ok || session.deleted {
		return nil, errors.New("session not found")
	}

	command, ok := session.commands[cmdId]
	if !ok {
		return nil, errors.New("command not found")
	}

	if command.ExitCode != nil {
		return command, nil
	}

	_, exitCodeFilePath := command.LogFilePath(session.Dir(s.configDir))
	exitCode, err := os.ReadFile(exitCodeFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return command, nil
		}
		return nil, errors.New("failed to read exit code file")
	}

	exitCodeInt, err := strconv.Atoi(strings.TrimRight(string(exitCode), "\n"))
	if err != nil {
		return nil, errors.New("failed to convert exit code to int")
	}

	command.ExitCode = &exitCodeInt

	return command, nil
}
