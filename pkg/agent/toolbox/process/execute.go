// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package process

import (
	"bytes"
	"os/exec"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type ExecuteRequest struct {
	Command string `json:"command" binding:"required"`
} // @name ExecuteRequest

type ExecuteResponse struct {
	Code   int    `json:"code"`
	Result string `json:"result"`
} // @name ExecuteResponse

func ExecuteCommand(c *gin.Context) {
	var request ExecuteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, ExecuteResponse{
			Code:   400,
			Result: "Invalid request: command is required",
		})
		return
	}

	cmdParts := parseCommand(request.Command)
	if len(cmdParts) == 0 {
		c.JSON(400, ExecuteResponse{
			Code:   400,
			Result: "Empty command",
		})
		return
	}

	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)

	//	set up process group for proper cleanup
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// set maximum execution time
	timer := time.AfterFunc(10*time.Second, func() {
		if cmd.Process != nil {
			// kill the process group
			syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
	})
	defer timer.Stop()

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		}
	}

	result := stdout.String()
	if stderr.Len() > 0 {
		if result != "" {
			result += "\n"
		}
		result += stderr.String()
	}

	// ensure process is killed if still running
	if cmd.Process != nil {
		cmd.Process.Kill()
		cmd.Process.Release()
	}

	c.JSON(200, ExecuteResponse{
		Code:   exitCode,
		Result: result,
	})
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
