// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package process

import (
	"context"
	"errors"
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

// stdioDrainGracePeriod bounds how long Wait blocks for stdout/stderr to drain
// after the direct child has exited. Without this, a daemonized descendant
// (e.g. a tmux server backgrounded by the user's command) that inherits the
// stdio pipes would keep CombinedOutput blocked forever, hanging the request.
const stdioDrainGracePeriod = 100 * time.Millisecond

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

		if err := common.ValidateEnvKeys(request.Envs); err != nil {
			c.Error(common_errors.NewBadRequestError(err))
			return
		}

		// Derive a cancellable context from the request so client disconnects
		// and explicit timeouts both reach the spawned process.
		ctx, cancel := context.WithCancel(c.Request.Context())
		defer cancel()

		// set maximum execution time
		var timeoutReached atomic.Bool
		if request.Timeout != nil && *request.Timeout > 0 {
			timeout := time.Duration(*request.Timeout) * time.Second
			timer := time.AfterFunc(timeout, func() {
				timeoutReached.Store(true)
				cancel()
			})
			defer timer.Stop()
		}

		// Pipe command via stdin to avoid OS ARG_MAX limits on large commands
		cmd := exec.CommandContext(ctx, common.GetShell())
		cmd.Stdin = strings.NewReader(request.Command)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		// Override the default Cancel (which only kills the direct child) so
		// the entire process group is terminated, including any shells, pipes
		// or daemons forked by the user's command.
		cmd.Cancel = func() error {
			if cmd.Process == nil {
				return os.ErrProcessDone
			}
			return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		// Bound the post-exit I/O drain so a backgrounded process that holds
		// the stdio pipes open cannot hang this request indefinitely.
		cmd.WaitDelay = stdioDrainGracePeriod
		if request.Cwd != nil {
			cmd.Dir = *request.Cwd
		}
		common.ApplyEnvs(cmd, request.Envs)

		output, err := cmd.CombinedOutput()
		if err != nil {
			if timeoutReached.Load() {
				c.Error(common_errors.NewRequestTimeoutError(errors.New("command execution timeout")))
				return
			}
			var exitError *exec.ExitError
			if errors.As(err, &exitError) {
				c.JSON(http.StatusOK, ExecuteResponse{
					ExitCode: exitError.ExitCode(),
					Result:   string(output),
				})
				return
			}
			// The command itself succeeded but a backgrounded descendant kept
			// the stdio pipes open past WaitDelay. Return what we captured as
			// a successful execution rather than a -1 error to the caller.
			if errors.Is(err, exec.ErrWaitDelay) {
				c.JSON(http.StatusOK, ExecuteResponse{
					ExitCode: 0,
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
