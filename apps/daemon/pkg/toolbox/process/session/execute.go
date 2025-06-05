// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/daytonaio/daemon/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Add a standard error response struct
type ErrorResponse struct {
	Error string `json:"error"`
}

func (s *SessionController) SessionExecuteCommand(c *gin.Context) {
	sessionId := c.Param("sessionId")

	var request SessionExecuteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	if request.Async {
		request.RunAsync = true
	}

	// Validate command is not empty (if not already handled by binding)
	if strings.TrimSpace(request.Command) == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("command cannot be empty"))
		return
	}

	session, ok := sessions[sessionId]
	if !ok {
		c.AbortWithError(http.StatusNotFound, errors.New("session not found"))
		return
	}

	var cmdId *string
	var logFile *os.File

	cmdId = util.Pointer(uuid.NewString())

	command := &Command{
		Id:      *cmdId,
		Command: request.Command,
	}
	session.commands[*cmdId] = command

	logFilePath, exitCodeFilePath := command.LogFilePath(session.Dir(s.configDir))

	if err := os.MkdirAll(filepath.Dir(logFilePath), 0755); err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to create log directory: %w", err))
		return
	}

	logFile, err := os.Create(logFilePath)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to create log file: %w", err))
		return
	}

	defer logFile.Close()

	cmdToExec := fmt.Sprintf("{ %s; } > %s 2>&1 ; echo \"$?\" > %s\n", request.Command, logFile.Name(), exitCodeFilePath)

	_, err = session.stdinWriter.Write([]byte(cmdToExec))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to write command: %w", err))
		return
	}

	if request.RunAsync {
		c.JSON(http.StatusAccepted, SessionExecuteResponse{
			CommandId: cmdId,
		})
		return
	}

	for {
		select {
		case <-session.ctx.Done():
			session.commands[*cmdId].ExitCode = util.Pointer(1)

			c.AbortWithError(http.StatusBadRequest, errors.New("session cancelled"))
			return
		default:
			exitCode, err := os.ReadFile(exitCodeFilePath)
			if err != nil {
				if os.IsNotExist(err) {
					time.Sleep(50 * time.Millisecond)
					continue
				}
				c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to read exit code file: %w", err))
				return
			}

			exitCodeInt, err := strconv.Atoi(strings.TrimRight(string(exitCode), "\n"))
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to convert exit code to int: %w", err))
				return
			}

			sessions[sessionId].commands[*cmdId].ExitCode = &exitCodeInt

			logBytes, err := os.ReadFile(logFilePath)
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to read log file: %w", err))
				return
			}

			logContent := string(logBytes)

			c.JSON(http.StatusOK, SessionExecuteResponse{
				CommandId: cmdId,
				Output:    &logContent,
				ExitCode:  &exitCodeInt,
			})
			return
		}
	}
}
