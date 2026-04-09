// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package process

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
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
		if len(request.Envs) > 0 {
			envPairs := make([]string, 0, len(request.Envs))
			for key, value := range request.Envs {
				envPairs = append(envPairs, fmt.Sprintf("%s=%s", key, value))
			}
			cmd.Env = append(os.Environ(), envPairs...)
		}

		// set maximum execution time
		var timeoutReached atomic.Bool
		if request.Timeout != nil && *request.Timeout > 0 {
			timeout := time.Duration(*request.Timeout) * time.Second
			timer := time.AfterFunc(timeout, func() {
				timeoutReached.Store(true)
				if cmd.Process != nil {
					// Kill the entire process group so child processes are also terminated
					if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
						logger.Error("failed to kill process group", "error", err)
					}
				}
			})
			defer timer.Stop()
		}

		output, err := cmd.CombinedOutput()
		if err != nil {
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
