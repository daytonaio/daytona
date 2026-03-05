// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package execute

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"os/exec"
	"syscall"
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
func (ec *ExecuteController) ExecuteCommand(c *gin.Context) {
	var request ExecuteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithError(http.StatusBadRequest, errors.New("command is required"))
		return
	}

	cmdParts := parseCommand(request.Command)
	if len(cmdParts) == 0 {
		c.AbortWithError(http.StatusBadRequest, errors.New("empty command"))
		return
	}

	// set maximum execution time
	timeout := 360 * time.Second
	if request.Timeout != nil && *request.Timeout > 0 {
		timeout = time.Duration(*request.Timeout) * time.Second
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()

	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	if request.Cwd != nil {
		cmd.Dir = *request.Cwd
	}

	// Set up process group so we can kill all child processes on timeout
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	var outputBuf bytes.Buffer
	cmd.Stdout = &outputBuf
	cmd.Stderr = &outputBuf

	if err := cmd.Start(); err != nil {
		c.JSON(http.StatusOK, ExecuteResponse{
			ExitCode: -1,
			Result:   err.Error(),
		})
	}

	// Wait for command
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// Wait for either completion or context timeout
	select {
	case <-ctx.Done():
		// Context timed out - gracefully terminate the process tree
		err := common.TerminateProcessTreeGracefully(
			context.Background(),
			ec.logger,
			cmd.Process,
			ec.terminationGracePeriod,
			ec.terminationCheckInterval,
		)
		if err != nil {
			ec.logger.ErrorContext(ctx, "Failed to terminate process group", "Error", err)
		}
		<-done // Wait for process cleanup
		c.AbortWithError(http.StatusRequestTimeout, errors.New("Command execution timed out after "+timeout.String()+". The timeout can be increased by adjusting the timeout parameter in the request."))
		return
	case <-done:
		// Command completed normally
		output := outputBuf.Bytes()
		exitCode := -1
		if cmd.ProcessState != nil {
			exitCode = cmd.ProcessState.ExitCode()
		}
		c.JSON(http.StatusOK, ExecuteResponse{
			ExitCode: exitCode,
			Result:   string(output),
		})
	}
}

// parseCommand splits a command string properly handling quotes
func parseCommand(command string) []string {
	var args []string
	var current bytes.Buffer
	var inQuotes bool
	var quoteChar rune

	for _, r := range command {
		switch {
		case r == '"' || r == '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = r
			} else if quoteChar == r {
				inQuotes = false
				quoteChar = 0
			} else {
				current.WriteRune(r)
			}
		case r == ' ' && !inQuotes:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}
