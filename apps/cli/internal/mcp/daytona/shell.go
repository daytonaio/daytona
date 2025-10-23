// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package daytona

import (
	"context"
	"fmt"
	"strings"

	apiclient "github.com/daytonaio/apiclient"
	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	mcp_headers "github.com/daytonaio/daytona/cli/internal/mcp"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	log "github.com/sirupsen/logrus"
)

type ShellInput struct {
	Id      *string `json:"id,omitempty" jsonchema:"ID of the sandbox to execute the command in."`
	Command *string `json:"command,omitempty" jsonchema:"Command to execute."`
}

type ShellOutput struct {
	Stdout    string `json:"stdout" jsonchema:"Standard output of the command."`
	Stderr    string `json:"stderr" jsonchema:"Standard error output of the command."`
	ExitCode  int    `json:"exitCode" jsonchema:"Exit code of the command."`
	ErrorType string `json:"errorType,omitempty" jsonchema:"Error type of the command."`
}

func getShellTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "shell",
		Title:       "Shell",
		Description: "Execute shell commands in the Daytona sandbox.",
	}
}

func handleShell(ctx context.Context, request *mcp.CallToolRequest, input *ShellInput) (*mcp.CallToolResult, *ShellOutput, error) {
	apiClient, err := apiclient_cli.GetApiClient(nil, mcp_headers.DaytonaMCPHeaders)
	if err != nil {
		return returnCommandError(fmt.Sprintf("Error getting API client: %v", err), "APIError")
	}

	if input.Id == nil || *input.Id == "" {
		return returnCommandError("Sandbox ID is required", "SandboxError")
	}

	if input.Command == nil || *input.Command == "" {
		return returnCommandError("Command must be a non-empty string", "ValueError")
	}

	// Process the command
	command := strings.TrimSpace(*input.Command)
	if strings.Contains(command, "&&") || strings.HasPrefix(command, "cd ") {
		// Wrap complex commands in /bin/sh -c
		command = fmt.Sprintf("/bin/sh -c %s", shellQuote(command))
	}

	log.Infof("Executing command: %s", command)

	// Execute the command
	result, _, err := apiClient.ToolboxAPI.ExecuteCommand(ctx, *input.Id).
		ExecuteRequest(*apiclient.NewExecuteRequest(command)).
		Execute()

	if err != nil {
		// Classify error types
		errStr := err.Error()
		switch {
		case strings.Contains(errStr, "Connection") || strings.Contains(errStr, "Timeout"):
			return returnCommandError(fmt.Sprintf("Network error during command execution: %s", errStr), "NetworkError")
		case strings.Contains(errStr, "Unauthorized") || strings.Contains(errStr, "401"):
			return returnCommandError("Authentication failed during command execution. Please check your API key", "NetworkError")
		default:
			return returnCommandError(fmt.Sprintf("Command execution failed: %s", errStr), "CommandExecutionError")
		}
	}

	// Process command output
	cmdResult := ShellOutput{
		Stdout:   strings.TrimSpace(result.Result),
		ExitCode: int(result.ExitCode),
	}

	// Log truncated output
	outputLen := len(cmdResult.Stdout)
	logOutput := cmdResult.Stdout
	if outputLen > 500 {
		logOutput = cmdResult.Stdout[:500] + "..."
	}

	log.Infof("Command completed - exit code: %d, output length: %d", cmdResult.ExitCode, outputLen)

	log.Debugf("Command output (truncated): %s", logOutput)

	// Check for non-zero exit code
	if cmdResult.ExitCode > 0 {
		log.Infof("Command exited with non-zero status - exit code: %d", cmdResult.ExitCode)
	}

	return &mcp.CallToolResult{
		IsError: false,
	}, &cmdResult, nil
}

// Helper function to return command errors in a consistent format
func returnCommandError(message, errorType string) (*mcp.CallToolResult, *ShellOutput, error) {
	return &mcp.CallToolResult{
			IsError: true,
		},
		&ShellOutput{
			Stdout:    "",
			Stderr:    message,
			ExitCode:  -1,
			ErrorType: errorType,
		}, fmt.Errorf("error: %s", message)
}

// Helper function to quote shell commands
func shellQuote(s string) string {
	// Simple shell quoting - wrap in single quotes and escape existing single quotes
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}
