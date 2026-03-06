// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package process

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"os/exec"
	"time"

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

		// TTY mode is not supported on this endpoint; clients must use /process/pty.
		// Reject early before doing any work.
		if request.TTY != nil && *request.TTY {
			c.JSON(http.StatusBadRequest, gin.H{"error": "TTY=true is not supported on this endpoint; use /process/pty instead"})
			return
		}

		cmdParts := parseCommand(request.Command)
		if len(cmdParts) == 0 {
			c.AbortWithError(http.StatusBadRequest, errors.New("empty command"))
			return
		}

		cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
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

// parseCommand splits a command string properly handling quotes and backslashes.
// Outside of quotes, only \' is treated as an escape (to support the POSIX '\''
// idiom produced by buildCommand). All other backslashes are literal.
// Within double-quoted strings, `\"` is treated as a literal double-quote character.
// Other backslashes are written through as-is. This allows arguments produced by
// buildCommand (which escapes internal `"` as `\"`) to round-trip correctly.
func parseCommand(command string) []string {
	var args []string
	var current bytes.Buffer
	var inQuotes bool
	var quoteChar rune

	runes := []rune(command)
	i := 0
	for i < len(runes) {
		r := runes[i]
		switch {
		case r == '\\' && !inQuotes && i+1 < len(runes) && runes[i+1] == '\'':
			// Outside quotes, \' produces a literal single-quote. This is required
			// to decode the '\'' idiom produced by buildCommand, where a single-quote
			// inside a -c script is encoded as '\'' (close-quote, \', open-quote).
			// All other backslashes outside quotes pass through literally so that
			// paths (C:\tmp), regexes, etc. round-trip unchanged.
			i++
			current.WriteRune(runes[i])
		case r == '\\' && inQuotes && quoteChar == '"' && i+1 < len(runes) && runes[i+1] == '"':
			// \" inside a double-quoted string: emit a literal double-quote and skip the next char.
			i++
			current.WriteRune('"')
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
		i++
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}
