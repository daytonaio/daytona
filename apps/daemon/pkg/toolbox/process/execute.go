// Copyright 2025 Daytona Platforms Inc.
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
		var request ExecuteRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.AbortWithError(http.StatusBadRequest, errors.New("command is required"))
			return
		}

		if len(request.Command) == 0 {
			c.AbortWithError(http.StatusBadRequest, errors.New("empty command"))
			return
		}

		// Pipe command via stdin to avoid OS ARG_MAX limits on large commands
		cmd := exec.Command(common.GetShell())
		cmd.Stdin = strings.NewReader(request.Command)
		if request.Cwd != nil {
			cmd.Dir = *request.Cwd
		}

		// set maximum execution time
		timeout := 360 * time.Second
		if request.Timeout != nil && *request.Timeout > 0 {
			timeout = time.Duration(*request.Timeout) * time.Second
		}

		timeoutReached := false
		timer := time.AfterFunc(timeout, func() {
			timeoutReached = true
			if cmd.Process != nil {
				// kill the process group
				err := cmd.Process.Kill()
				if err != nil {
					logger.Error("failed to kill process", "error", err)
					return
				}
			}
		})
		defer timer.Stop()

		output, err := cmd.CombinedOutput()
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
