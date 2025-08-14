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

	cmdId := util.Pointer(uuid.NewString())

	command := &Command{
		Id:      *cmdId,
		Command: request.Command,
	}
	session.commands[*cmdId] = command

	stdoutFilePath, stderrFilePath, exitCodeFilePath := command.LogFilePath(session.Dir(s.configDir))

	if err := os.MkdirAll(filepath.Dir(stdoutFilePath), 0755); err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to create log directory: %w", err))
		return
	}

	stdoutFile, err := os.Create(stdoutFilePath)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to create stdout log file: %w", err))
		return
	}

	defer stdoutFile.Close()

	stderrFile, err := os.Create(stderrFilePath)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to create stderr log file: %w", err))
		return
	}

	defer stderrFile.Close()

	cmdToExec := fmt.Sprintf(
		`{ 
			 %s; 
			 exit_code=$?; 
			 echo %s >&1; 
			 echo %s >&2; 
			 echo "$exit_code" > %s; 
		 } > %s 2> %s`+"\n",
		request.Command,
		COMMAND_EXIT_MARKER, // goes into stdout log
		COMMAND_EXIT_MARKER, // goes into stderr log
		exitCodeFilePath,
		stdoutFilePath,
		stderrFilePath,
	)

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

			stdoutBytes, err := os.ReadFile(stdoutFilePath)
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to read log file: %w", err))
				return
			}
			stdoutContent := strings.TrimSuffix(strings.TrimRight(string(stdoutBytes), " \n\r\t"), COMMAND_EXIT_MARKER)

			stderrBytes, err := os.ReadFile(stderrFilePath)
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to read stderr log file: %w", err))
				return
			}
			stderrContent := strings.TrimSuffix(strings.TrimRight(string(stderrBytes), " \n\r\t"), COMMAND_EXIT_MARKER)

			c.JSON(http.StatusOK, SessionExecuteResponse{
				CommandId: cmdId,
				Output:    util.Pointer(stdoutContent + "\n" + stderrContent),
				Stdout:    &stdoutContent,
				Stderr:    &stderrContent,
				ExitCode:  &exitCodeInt,
			})
			return
		}
	}
}
