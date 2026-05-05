// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package coderun

import (
	"bytes"
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
// that inherits the stdio pipes would keep Wait blocked forever, hanging the
// request.
const stdioDrainGracePeriod = 100 * time.Millisecond

// CodeRun godoc
//
//	@Summary		Execute code
//	@Description	Execute Python, JavaScript, or TypeScript code and return output, exit code, and artifacts
//	@Tags			process
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CodeRunRequest	true	"Code execution request"
//	@Success		200		{object}	CodeRunResponse
//	@Router			/process/code-run [post]
//
//	@id				CodeRun
func CodeRun(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request CodeRunRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.Error(common_errors.NewInvalidBodyRequestError(err))
			return
		}

		toolbox, err := GetToolbox(request.Language)
		if err != nil {
			c.Error(common_errors.NewBadRequestError(err))
			return
		}

		if err := common.ValidateEnvKeys(request.Envs); err != nil {
			c.Error(common_errors.NewBadRequestError(err))
			return
		}

		ctx, cancel := context.WithCancel(c.Request.Context())
		defer cancel()

		var timeoutReached atomic.Bool
		if request.Timeout != nil && *request.Timeout > 0 {
			timeout := time.Duration(*request.Timeout) * time.Second
			timer := time.AfterFunc(timeout, func() {
				timeoutReached.Store(true)
				cancel()
			})
			defer timer.Stop()
		}

		runCommand := toolbox.GetRunCommand(request.Code, request.Argv)
		cmd := exec.CommandContext(ctx, common.GetShell())
		cmd.Stdin = strings.NewReader(runCommand)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		// Kill the entire process group on cancel so any descendants spawned
		// by the user's code (subshells, daemons, runaway children) go too.
		cmd.Cancel = func() error {
			if cmd.Process == nil {
				return os.ErrProcessDone
			}
			return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		// Bound the post-exit I/O drain to keep a backgrounded descendant
		// holding the stdio pipes from hanging this request indefinitely.
		cmd.WaitDelay = stdioDrainGracePeriod
		common.ApplyEnvs(cmd, request.Envs)

		var outputBuf bytes.Buffer
		cmd.Stdout = &outputBuf
		cmd.Stderr = &outputBuf

		if err := cmd.Start(); err != nil {
			c.JSON(http.StatusOK, CodeRunResponse{
				ExitCode: -1,
				Result:   err.Error(),
			})
			return
		}

		err = cmd.Wait()
		output := outputBuf.Bytes()
		if err != nil {
			if timeoutReached.Load() {
				c.Error(common_errors.NewRequestTimeoutError(errors.New("command execution timeout")))
				return
			}

			var exitError *exec.ExitError
			if errors.As(err, &exitError) {
				result, artifacts := ParseArtifacts(string(output))
				c.JSON(http.StatusOK, CodeRunResponse{
					ExitCode:  exitError.ExitCode(),
					Result:    result,
					Artifacts: artifacts,
				})
				return
			}

			// Process exited successfully but stdio drain was cut short by
			// WaitDelay (e.g. backgrounded descendant kept the pipes open).
			// Surface the partial output as a successful run.
			if errors.Is(err, exec.ErrWaitDelay) {
				result, artifacts := ParseArtifacts(string(output))
				c.JSON(http.StatusOK, CodeRunResponse{
					ExitCode:  0,
					Result:    result,
					Artifacts: artifacts,
				})
				return
			}

			result, artifacts := ParseArtifacts(string(output))
			c.JSON(http.StatusOK, CodeRunResponse{
				ExitCode:  -1,
				Result:    result,
				Artifacts: artifacts,
			})
			return
		}

		if cmd.ProcessState == nil {
			result, artifacts := ParseArtifacts(string(output))
			c.JSON(http.StatusOK, CodeRunResponse{
				ExitCode:  -1,
				Result:    result,
				Artifacts: artifacts,
			})
			return
		}

		result, artifacts := ParseArtifacts(string(output))
		c.JSON(http.StatusOK, CodeRunResponse{
			ExitCode:  cmd.ProcessState.ExitCode(),
			Result:    result,
			Artifacts: artifacts,
		})
	}
}
