//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package process

import (
	"errors"
	"log/slog"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/daytonaio/daemon/pkg/common"

	"github.com/gin-gonic/gin"
)

// ExecuteCommand godoc
//
//	@Summary		Execute a command
//	@Description	Execute a shell command and return the output and exit code
//	@Tags			process
//	@Accept			json
//	@Produce		json
//	@Param			request	body		ExecuteRequest	true	"Command execution request"
//	@Success		200		{object}	ExecuteResponse
//	@Router			/process/execute [post]
//
//	@id				ExecuteCommand
func ExecuteCommand(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		var request ExecuteRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.AbortWithError(http.StatusBadRequest, errors.New("command is required"))
			return
		}

		if strings.TrimSpace(request.Command) == "" {
			c.AbortWithError(http.StatusBadRequest, errors.New("empty command"))
			return
		}

		parsedCommand, envVars := common.ParseShellWrapper(request.Command)
		if parsedCommand != request.Command {
			logger.Debug("Parsed shell wrapper", "raw", request.Command, "parsed", parsedCommand, "env", envVars)
		}

		shell := common.GetShell()
		shellArgs := common.GetShellArgs(shell)

		isPowerShell := common.IsPowerShell(shell)
		finalCommand := common.BuildWindowsCommandForShell(parsedCommand, envVars, isPowerShell)

		logger.Debug("ExecuteCommand: prepared",
			"shell", shell,
			"is_powershell", isPowerShell,
			"command", finalCommand,
			"setup_duration", time.Since(startTime),
		)

		execStartTime := time.Now()

		args := append(shellArgs, finalCommand)
		cmd := exec.Command(shell, args...)

		if request.Cwd != nil {
			cmd.Dir = *request.Cwd
		}

		timeout := 360 * time.Second
		if request.Timeout != nil && *request.Timeout > 0 {
			timeout = time.Duration(*request.Timeout) * time.Second
		}

		timeoutReached := false
		timer := time.AfterFunc(timeout, func() {
			timeoutReached = true
			if cmd.Process != nil {
				if err := cmd.Process.Kill(); err != nil {
					logger.Error("Failed to kill process on timeout", "error", err)
					return
				}
			}
		})
		defer timer.Stop()

		output, err := cmd.CombinedOutput()
		execDuration := time.Since(execStartTime)
		logger.Debug("ExecuteCommand: completed",
			"execution_duration", execDuration,
			"total_duration", time.Since(startTime),
		)

		if err != nil {
			if timeoutReached {
				c.AbortWithError(http.StatusRequestTimeout, errors.New("command execution timeout"))
				return
			}
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode := exitError.ExitCode()
				c.JSON(http.StatusOK, ExecuteResponse{
					ExitCode: exitCode,
					Result:   string(output),
				})
				return
			}
			c.JSON(http.StatusOK, ExecuteResponse{
				ExitCode: -1,
				Result:   string(output),
			})
			return
		}

		if cmd.ProcessState == nil {
			c.JSON(http.StatusOK, ExecuteResponse{
				ExitCode: -1,
				Result:   string(output),
			})
			return
		}

		exitCode := cmd.ProcessState.ExitCode()
		c.JSON(http.StatusOK, ExecuteResponse{
			ExitCode: exitCode,
			Result:   string(output),
		})
	}
}
