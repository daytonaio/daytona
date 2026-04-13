// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package process

import (
	"errors"
	"log/slog"
	"net/http"
	"os/exec"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	common_errors "github.com/daytonaio/common-go/pkg/errors"

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

		if strings.TrimSpace(request.Command) == "" {
			c.AbortWithError(http.StatusBadRequest, errors.New("empty command"))
			return
		}

		// Pipe command via stdin to avoid OS ARG_MAX limits on large commands
		cmd := exec.Command(common.GetShell())
		cmd.Stdin = strings.NewReader(request.Command)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		if request.Cwd != nil {
			cmd.Dir = *request.Cwd
		}

		killProcessGroup := func() {
			if cmd.Process != nil {
				if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
					logger.Error("failed to kill process group", "error", err)
				}
			}
		}

		// set maximum execution time
		var timeoutReached atomic.Bool
		if request.Timeout != nil && *request.Timeout > 0 {
			timeout := time.Duration(*request.Timeout) * time.Second
			timer := time.AfterFunc(timeout, func() {
				timeoutReached.Store(true)
				killProcessGroup()
			})
			defer timer.Stop()
		}

		// Monitor HTTP request context for client disconnection.
		// When the client drops the connection (e.g. via AbortSignal), Go's
		// net/http cancels this context, and we kill the process group.
		var clientDisconnected atomic.Bool
		done := make(chan struct{})
		go func() {
			select {
			case <-c.Request.Context().Done():
				clientDisconnected.Store(true)
				killProcessGroup()
			case <-done:
			}
		}()

		output, err := cmd.CombinedOutput()
		close(done)
		if err != nil {
			if clientDisconnected.Load() {
				return
			}
			if timeoutReached.Load() {
				c.Error(common_errors.NewRequestTimeoutError(errors.New("command execution timeout")))
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
