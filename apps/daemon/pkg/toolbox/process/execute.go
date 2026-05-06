// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package process

import (
	"errors"
	"fmt"
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

// shellSingleQuote wraps s in single quotes, escaping any embedded single
// quotes via the standard '\” idiom. The result is safe to drop into a
// POSIX shell command line.
func shellSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// detachedWrapperCommand wraps userCommand so that the spawned shell:
//
//   - verifies setsid is available and surfaces a clear error on stderr
//     (which the daemon captures) before redirecting stdio, so callers
//     using minimal images without util-linux get an actionable message
//     instead of a bare exit code 127;
//
//   - exec's setsid -f, replacing itself; setsid then forks a session-leader
//     child and exits immediately, so cmd.CombinedOutput returns within
//     milliseconds instead of waiting for the user's daemon to terminate;
//
//   - has stdin/stdout/stderr pre-redirected to /dev/null before the exec,
//     so the daemonized descendant inherits /dev/null on those fds rather
//     than the daemon's stdout/stderr pipes (which would otherwise pin the
//     request open until the daemon exited).
//
// The user command is still passed through bash's stdin (avoiding ARG_MAX
// limits at the cmd.Args level) and is re-quoted into the inner `sh -c`
// argument here.
func detachedWrapperCommand(shell, userCommand string) string {
	const setsidProbe = "command -v setsid >/dev/null 2>&1 || " +
		`{ echo "runDetached requires setsid (install util-linux)" >&2; exit 127; }` + "\n"
	return setsidProbe + fmt.Sprintf(
		"exec setsid -f %s -c %s </dev/null >/dev/null 2>&1\n",
		shellSingleQuote(shell),
		shellSingleQuote(userCommand),
	)
}

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

		shell := common.GetShell()
		stdinScript := request.Command
		if request.RunDetached {
			// Replace the user's command with a setsid-fork wrapper that
			// returns control to the daemon as soon as the launcher forks,
			// while detaching the user's process from this request's stdio.
			stdinScript = detachedWrapperCommand(shell, request.Command)
		}

		// Pipe command via stdin to avoid OS ARG_MAX limits on large commands
		cmd := exec.Command(shell)
		cmd.Stdin = strings.NewReader(stdinScript)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		if request.Cwd != nil {
			cmd.Dir = *request.Cwd
		}
		common.ApplyEnvs(cmd, request.Envs)

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
