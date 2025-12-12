// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package process

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"os/exec"
	"syscall"
	"time"

	"github.com/daytonaio/daemon/internal/util"
	"github.com/daytonaio/daemon/pkg/common"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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
func ExecuteCommand(c *gin.Context) {
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

	cmd := exec.CommandContext(ctx, cmdParts[0], cmdParts[1:]...)
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
		return
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
		err := common.TerminateProcessTreeGracefully(context.Background(), cmd.Process, util.Pointer(2*time.Second))
		if err != nil {
			log.Errorf("Failed to terminate process group: %v", err)
		}
		<-done // Wait for process cleanup
		c.AbortWithError(http.StatusRequestTimeout, errors.New("command execution timeout"))
		return
	case err := <-done:
		// Command completed normally
		output := outputBuf.Bytes()
		if err != nil {
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
