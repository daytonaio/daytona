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
	"syscall"
	"time"

	"github.com/daytonaio/daemon/internal/util"
	"github.com/daytonaio/daemon/pkg/common"
	"github.com/gin-gonic/gin"

	log "github.com/sirupsen/logrus"
)

const TERMINATION_GRACE_PERIOD = 5 * time.Second
const TERMINATION_CHECK_INTERVAL = 100 * time.Millisecond

var sessions = map[string]*session{}

func (s *SessionController) CreateSession(c *gin.Context) {
	ctx, cancel := context.WithCancel(context.Background())

	cmd := exec.CommandContext(ctx, common.GetShell())
	cmd.Env = os.Environ()

	// Set up a new process group so we can kill all child processes when the session is deleted
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// for backward compatibility (only sdk clients before 0.103.X), we use the home directory as the default directory
	sdkVersion := util.ExtractSdkVersionFromHeader(c.Request.Header)
	versionComparison, err := util.CompareVersions(sdkVersion, "0.103.0-0")
	if err != nil {
		log.Error(err)
		versionComparison = util.Pointer(1)
	}
	isLegacy := versionComparison != nil && *versionComparison < 0 && sdkVersion != "0.0.0-dev"
	if isLegacy {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			cancel()
			return
		}
		cmd.Dir = homeDir
	}

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
	if !ok {
		_ = c.AbortWithError(http.StatusNotFound, errors.New("session not found"))
		return
	}

	// Cancel context first - this signals CommandContext to stop
	session.cancel()

	// Terminate process group if still running
	if err := s.terminateSession(c.Request.Context(), session); err != nil {
		log.Errorf("Failed to terminate session %s: %v", session.id, err)
		// Continue with cleanup even if termination fails
	}

	// Clean up session directory
	err := os.RemoveAll(session.Dir(s.configDir))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	delete(sessions, session.id)
	c.Status(http.StatusNoContent)
}

func (s *SessionController) ListSessions(c *gin.Context) {
	sessionDTOs := []Session{}

	for sessionId := range sessions {
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

	_, ok := sessions[sessionId]
	if !ok {
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
	if !ok {
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
	if !ok {
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

func (s *SessionController) terminateSession(ctx context.Context, session *session) error {
	if session.cmd == nil || session.cmd.Process == nil {
		return nil
	}

	pid := session.cmd.Process.Pid

	// Send SIGTERM to entire process group (negative PID)
	err := syscall.Kill(-pid, syscall.SIGTERM)
	if err != nil {
		// If SIGTERM fails, try SIGKILL immediately
		log.Warnf("SIGTERM failed for session %s, trying SIGKILL: %v", session.id, err)
		return syscall.Kill(-pid, syscall.SIGKILL)
	}

	// Wait for graceful termination
	if s.waitForTermination(ctx, pid, TERMINATION_GRACE_PERIOD, TERMINATION_CHECK_INTERVAL) {
		log.Debugf("Session %s terminated gracefully", session.id)
		return nil
	}

	// Force kill entire process group if still alive
	log.Debugf("Session %s timeout, sending SIGKILL to process group", session.id)
	return syscall.Kill(-pid, syscall.SIGKILL)
}

func (s *SessionController) waitForTermination(ctx context.Context, pid int, timeout, interval time.Duration) bool {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return false
		case <-ticker.C:
			err := syscall.Kill(-pid, 0)
			if err != nil {
				return true
			}
		}
	}
}
