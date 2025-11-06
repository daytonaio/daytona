// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package daytona

import (
	"context"
	"fmt"
	"strings"

	"github.com/daytonaio/toolbox_apiclient"

	"github.com/daytonaio/mcp/internal/apiclient"
	"github.com/daytonaio/mcp/internal/common"
	"github.com/daytonaio/mcp/internal/constants"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	log "github.com/sirupsen/logrus"
)

type ShellInput struct {
	SandboxId *string `json:"sandboxId,omitempty" jsonschema:"ID of the sandbox to execute the command in. Don't provide this if not explicitly instructed from user. If not provided, a new sandbox will be created."`
	Command   string  `json:"command" jsonschema:"Command to execute."`
}

type ShellOutput struct {
	Stdout    *string `json:"stdout,omitempty" jsonschema:"Standard output of the command."`
	Stderr    *string `json:"stderr,omitempty" jsonschema:"Standard error output of the command."`
	ExitCode  *int    `json:"exitCode,omitempty" jsonschema:"Exit code of the command."`
	ErrorType *string `json:"errorType,omitempty" jsonschema:"Error type of the command."`
}

func (s *DaytonaMCPServer) getShellTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "shell",
		Title:       "Shell",
		Description: "Execute shell commands in the Daytona sandbox.",
	}
}

func (s *DaytonaMCPServer) handleShell(ctx context.Context, request *mcp.CallToolRequest, input *ShellInput) (*mcp.CallToolResult, *ShellOutput, error) {
	if input.Command == "" {
		return returnCommandError("Command must be a non-empty string", "ValueError")
	}

	sandbox, stop, err := common.GetSandbox(ctx, s.apiClient, input.SandboxId)
	if err != nil {
		return returnCommandError(fmt.Sprintf("Error getting sandbox: %v", err), "SandboxError")
	}
	defer stop()

	// Process the command
	command := strings.TrimSpace(input.Command)
	if strings.Contains(command, "&&") || strings.HasPrefix(command, "cd ") {
		// Wrap complex commands in /bin/sh -c
		command = fmt.Sprintf("/bin/sh -c %s", shellQuote(command))
	}

	proxyUrl, err := apiclient.ExtractProxyUrl(ctx, s.apiClient)
	if err != nil {
		return returnCommandError(fmt.Sprintf("Error extracting proxy URL: %v", err), "ProxyUrlError")
	}

	toolboxApiClient := apiclient.NewToolboxApiClient(constants.DaytonaMcpSource, sandbox.Id, proxyUrl, request.Extra.Header)

	log.Infof("Executing command: %s", command)

	// Execute the command
	result, _, err := toolboxApiClient.ProcessAPI.ExecuteCommand(ctx).Request(*toolbox_apiclient.NewExecuteRequest(command)).Execute()
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

	stdout := strings.TrimSpace(result.Result)
	exitCode := int(*result.ExitCode)

	// Process command output
	cmdResult := ShellOutput{
		Stdout:   &stdout,
		ExitCode: &exitCode,
	}

	// Log truncated output
	outputLen := len(*cmdResult.Stdout)
	logOutput := *cmdResult.Stdout
	if outputLen > 500 {
		logOutput = (*cmdResult.Stdout)[:500] + "..."
	}

	log.Infof("Command completed - exit code: %d, output length: %d", cmdResult.ExitCode, outputLen)

	log.Debugf("Command output (truncated): %s", logOutput)

	// Check for non-zero exit code
	if cmdResult.ExitCode != nil && *cmdResult.ExitCode > 0 {
		log.Infof("Command exited with non-zero status - exit code: %d", cmdResult.ExitCode)
	}

	return &mcp.CallToolResult{
		IsError: false,
	}, &cmdResult, nil
}

// Helper function to return command errors in a consistent format
func returnCommandError(message, errorType string) (*mcp.CallToolResult, *ShellOutput, error) {
	exitCode := -1
	return &mcp.CallToolResult{
			IsError: true,
		},
		&ShellOutput{
			Stdout:    nil,
			Stderr:    &message,
			ExitCode:  &exitCode,
			ErrorType: &errorType,
		}, fmt.Errorf("error: %s", message)
}

// Helper function to quote shell commands
func shellQuote(s string) string {
	// Simple shell quoting - wrap in single quotes and escape existing single quotes
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}
