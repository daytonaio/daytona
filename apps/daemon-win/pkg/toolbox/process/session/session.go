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
	"time"

	"github.com/daytonaio/daemon-win/internal/util"
	"github.com/daytonaio/daemon-win/pkg/common"
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v4/process"

	log "github.com/sirupsen/logrus"
)

const TERMINATION_GRACE_PERIOD = 5 * time.Second
const TERMINATION_CHECK_INTERVAL = 100 * time.Millisecond

var sessions = map[string]*session{}

// CreateSession godoc
//
//	@Summary		Create a new session
//	@Description	Create a new shell session for command execution
//	@Tags			process
//	@Accept			json
//	@Produce		json
//	@Param			request	body	CreateSessionRequest	true	"Session creation request"
//	@Success		201
//	@Router			/process/session [post]
//
//	@id				CreateSession
func (s *SessionController) CreateSession(c *gin.Context) {
	ctx, cancel := context.WithCancel(context.Background())

	// Use PowerShell for Windows sessions
	shell := common.GetShell()
	cmd := exec.CommandContext(ctx, shell)
	cmd.Env = os.Environ()

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
		id:           request.SessionId,
		cmd:          cmd,
		stdinWriter:  stdinWriter,
		commands:     map[string]*Command{},
		ctx:          ctx,
		cancel:       cancel,
		isPowerShell: common.IsPowerShell(shell),
	}
	sessions[request.SessionId] = session

	err = os.MkdirAll(session.Dir(s.configDir), 0755)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusCreated)
}

// DeleteSession godoc
//
//	@Summary		Delete a session
//	@Description	Delete an existing shell session
//	@Tags			process
//	@Param			sessionId	path	string	true	"Session ID"
//	@Success		204
//	@Router			/process/session/{sessionId} [delete]
//
//	@id				DeleteSession
func (s *SessionController) DeleteSession(c *gin.Context) {
	sessionId := c.Param("sessionId")

	session, ok := sessions[sessionId]
	if !ok {
		_ = c.AbortWithError(http.StatusNotFound, errors.New("session not found"))
		return
	}

	// Terminate process tree on Windows
	err := s.terminateSession(c.Request.Context(), session)
	if err != nil {
		log.Errorf("Failed to terminate session %s: %v", session.id, err)
		// Continue with cleanup even if termination fails
	}

	// Cancel context after termination
	session.cancel()

	// Clean up session directory
	err = os.RemoveAll(session.Dir(s.configDir))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	delete(sessions, session.id)
	c.Status(http.StatusNoContent)
}

// ListSessions godoc
//
//	@Summary		List all sessions
//	@Description	Get a list of all active shell sessions
//	@Tags			process
//	@Produce		json
//	@Success		200	{array}	Session
//	@Router			/process/session [get]
//
//	@id				ListSessions
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

// GetSession godoc
//
//	@Summary		Get session details
//	@Description	Get details of a specific session including its commands
//	@Tags			process
//	@Produce		json
//	@Param			sessionId	path		string	true	"Session ID"
//	@Success		200			{object}	Session
//	@Router			/process/session/{sessionId} [get]
//
//	@id				GetSession
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

// GetSessionCommand godoc
//
//	@Summary		Get session command details
//	@Description	Get details of a specific command within a session
//	@Tags			process
//	@Produce		json
//	@Param			sessionId	path		string	true	"Session ID"
//	@Param			commandId	path		string	true	"Command ID"
//	@Success		200			{object}	Command
//	@Router			/process/session/{sessionId}/command/{commandId} [get]
//
//	@id				GetSessionCommand
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

	// Clean up the exit code string - trim whitespace, BOM, and line endings
	exitCodeStr := strings.TrimSpace(string(exitCode))
	exitCodeStr = strings.TrimRight(exitCodeStr, "\r\n")
	// Remove UTF-8 BOM if present
	exitCodeStr = strings.TrimPrefix(exitCodeStr, "\xef\xbb\xbf")
	exitCodeStr = strings.TrimSpace(exitCodeStr)

	// If empty after cleanup, command is still running
	if exitCodeStr == "" {
		return command, nil
	}

	exitCodeInt, err := strconv.Atoi(exitCodeStr)
	if err != nil {
		log.Warnf("Failed to parse exit code '%s' for command %s: %v", exitCodeStr, cmdId, err)
		// Default to 1 (error) if we can't parse
		exitCodeInt = 1
	}

	command.ExitCode = &exitCodeInt

	return command, nil
}

func (s *SessionController) terminateSession(ctx context.Context, session *session) error {
	if session.cmd == nil || session.cmd.Process == nil {
		return nil
	}

	pid := session.cmd.Process.Pid

	// First, try to terminate child processes
	_ = s.terminateProcessTree(pid)

	// Then kill the main process
	err := session.cmd.Process.Kill()
	if err != nil {
		log.Warnf("Failed to kill session %s process: %v", session.id, err)
		return err
	}

	log.Debugf("Session %s process killed", session.id)
	return nil
}

func (s *SessionController) terminateProcessTree(pid int) error {
	parent, err := process.NewProcess(int32(pid))
	if err != nil {
		return err
	}

	children, err := parent.Children()
	if err != nil {
		return err
	}

	// Terminate children first (recursively)
	for _, child := range children {
		_ = s.terminateProcessTree(int(child.Pid))
	}

	// Then terminate the children directly
	for _, child := range children {
		if childProc, err := os.FindProcess(int(child.Pid)); err == nil {
			_ = childProc.Kill()
		}
	}

	return nil
}
