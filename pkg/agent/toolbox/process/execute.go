// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package process

import (
	"bytes"
	"errors"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

func ExecuteCommand(projectDir string) func(c *gin.Context) {
	return func(c *gin.Context) {
		var request ExecuteRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.AbortWithError(400, errors.New("command is required"))
			return
		}

		cmdParts := parseCommand(request.Command)
		if len(cmdParts) == 0 {
			c.AbortWithError(400, errors.New("empty command"))
			return
		}

		cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
		cmd.Dir = projectDir

		// set maximum execution time
		timeout := 10 * time.Second
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
					log.Error(err)
					return
				}
			}
		})
		defer timer.Stop()

		output, err := cmd.CombinedOutput()
		if err != nil {
			if timeoutReached {
				c.AbortWithError(408, errors.New("command execution timeout"))
				return
			}
			c.AbortWithError(400, err)
			return
		}

		if cmd.ProcessState == nil {
			c.JSON(200, ExecuteResponse{
				Code:   -1,
				Result: string(output),
			})
			return
		}

		c.JSON(200, ExecuteResponse{
			Code:   cmd.ProcessState.ExitCode(),
			Result: string(output),
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
