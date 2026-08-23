// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package process

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unicode/utf8"

	common_errors "github.com/daytonaio/common-go/pkg/errors"

	"github.com/daytonaio/daemon/pkg/childreap"
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
			c.Error(common_errors.NewBadRequestError(fmt.Errorf("invalid request body: %w", err)))
			return
		}

		if strings.TrimSpace(request.Command) == "" {
			c.Error(common_errors.NewBadRequestError(errors.New("command cannot be empty or whitespace-only")))
			return
		}

		if err := common.ValidateEnvKeys(request.Envs); err != nil {
			c.Error(common_errors.NewBadRequestError(err))
			return
		}
		if request.MaxOutputBytes != nil && (*request.MaxOutputBytes == 0 || *request.MaxOutputBytes > maxOutputBytes) {
			c.Error(common_errors.NewBadRequestError(fmt.Errorf("maxOutputBytes must be between 1 and %d", maxOutputBytes)))
			return
		}

		// Pipe command via stdin to avoid OS ARG_MAX limits on large commands
		cmd := exec.Command(common.GetShell())
		cmd.Stdin = strings.NewReader(request.Command)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		if request.Cwd != nil {
			cmd.Dir = *request.Cwd
		}
		common.ApplyEnvs(cmd, request.Envs)

		// Capture output ourselves so we can route through childreap.Wait
		// (cmd.CombinedOutput would call cmd.Wait directly and race the PID-1
		// reaper). The shared lock preserves safe combined writes while retaining
		// stdout and stderr independently.
		output := newCommandOutput(request.MaxOutputBytes)
		cmd.Stdout = output.stdoutWriter()
		cmd.Stderr = output.stderrWriter()

		if err := cmd.Start(); err != nil {
			c.JSON(http.StatusOK, ExecuteResponse{
				ExitCode: -1,
				Result:   err.Error(),
			})
			return
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

		exitCode, waitErr := childreap.Wait(cmd)
		if timeoutReached.Load() {
			if request.MaxOutputBytes == nil {
				c.Error(common_errors.NewRequestTimeoutError(errors.New("command execution timeout")))
				return
			}
			c.JSON(http.StatusOK, output.response(-1, true))
			return
		}
		if waitErr != nil {
			c.JSON(http.StatusOK, output.response(-1, false))
			return
		}

		c.JSON(http.StatusOK, output.response(exitCode, false))
	}
}

type commandOutput struct {
	mu     sync.Mutex
	result *tailBuffer
	stdout *tailBuffer
	stderr *tailBuffer
}

func newCommandOutput(maxBytes *uint32) *commandOutput {
	max := 0
	if maxBytes != nil {
		max = int(*maxBytes)
	}
	return &commandOutput{
		result: newTailBuffer(max),
		stdout: newTailBuffer(max),
		stderr: newTailBuffer(max),
	}
}

func (o *commandOutput) stdoutWriter() io.Writer { return streamWriter{output: o, stream: o.stdout} }

func (o *commandOutput) stderrWriter() io.Writer { return streamWriter{output: o, stream: o.stderr} }

func (o *commandOutput) response(exitCode int, timedOut bool) ExecuteResponse {
	result, _ := o.result.value()
	stdout, stdoutTruncated := o.stdout.value()
	stderr, stderrTruncated := o.stderr.value()
	return ExecuteResponse{
		ExitCode:        exitCode,
		Result:          result,
		Stdout:          stdout,
		Stderr:          stderr,
		TimedOut:        timedOut,
		StdoutTruncated: stdoutTruncated,
		StderrTruncated: stderrTruncated,
	}
}

type streamWriter struct {
	output *commandOutput
	stream *tailBuffer
}

func (w streamWriter) Write(p []byte) (int, error) {
	w.output.mu.Lock()
	defer w.output.mu.Unlock()

	_, _ = w.stream.Write(p)
	_, _ = w.output.result.Write(p)
	return len(p), nil
}

type tailBuffer struct {
	max       int
	buf       bytes.Buffer
	truncated bool
}

func newTailBuffer(max int) *tailBuffer { return &tailBuffer{max: max} }

func (b *tailBuffer) Write(p []byte) (int, error) {
	n, err := b.buf.Write(p)
	if b.max > 0 {
		if over := b.buf.Len() - b.max; over > 0 {
			b.buf.Next(over)
			b.truncated = true
		}
	}
	return n, err
}

func (b *tailBuffer) value() (string, bool) {
	if b.max == 0 {
		return b.buf.String(), false
	}
	value := bytes.ToValidUTF8(b.buf.Bytes(), []byte("\uFFFD"))
	truncated := b.truncated
	if b.max > 0 && len(value) > b.max {
		value = value[len(value)-b.max:]
		for len(value) > 0 && !utf8.RuneStart(value[0]) {
			value = value[1:]
		}
		truncated = true
	}
	return string(value), truncated
}
