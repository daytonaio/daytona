// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package coderun

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"os/exec"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/pkg/childreap"
	"github.com/daytonaio/daemon/pkg/common"
	"github.com/gin-gonic/gin"
)

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

		runCommand := toolbox.GetRunCommand(request.Code, request.Argv)
		cmd := exec.Command(common.GetShell())
		cmd.Stdin = strings.NewReader(runCommand)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
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

		var timeoutReached atomic.Bool
		if request.Timeout != nil && *request.Timeout > 0 {
			timeout := time.Duration(*request.Timeout) * time.Second
			timer := time.AfterFunc(timeout, func() {
				timeoutReached.Store(true)
				if killErr := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); killErr != nil {
					logger.Error("failed to kill process group", "error", killErr)
				}
			})
			defer timer.Stop()
		}

		exitCode, waitErr := childreap.Wait(cmd)
		output := outputBuf.Bytes()
		if timeoutReached.Load() {
			c.Error(common_errors.NewRequestTimeoutError(errors.New("command execution timeout")))
			return
		}
		if waitErr != nil {
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
			ExitCode:  exitCode,
			Result:    result,
			Artifacts: artifacts,
		})
	}
}
