// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/daytonaio/daytona/cli/apiclient"
	daytonaapiclient "github.com/daytonaio/daytona/daytonaapiclient"
	"github.com/mark3labs/mcp-go/mcp"

	log "github.com/sirupsen/logrus"
)

type CommandResult struct {
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	ExitCode  int    `json:"exit_code"`
	ErrorType string `json:"error_type,omitempty"`
}

func ExecuteCommand(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiClient, err := apiclient.GetApiClient(nil, daytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, err
	}

	cmd := request.Params.Arguments["command"].(string)

	// Validate command input
	if cmd == "" {
		return returnCommandError("Command must be a non-empty string", "ValueError")
	}

	// Get sandbox from tracking file
	sandboxID, err := getActiveSandbox()
	if err != nil || sandboxID == "" {
		return returnCommandError("No sandbox ID found in tracking file", "SandboxError")
	}

	// Process the command
	command := strings.TrimSpace(cmd)
	if strings.Contains(command, "&&") || strings.HasPrefix(command, "cd ") {
		// Wrap complex commands in /bin/sh -c
		command = fmt.Sprintf("/bin/sh -c %s", shellQuote(command))
	}

	log.Infof("Executing command: %s", command)

	// Execute the command
	result, _, err := apiClient.ToolboxAPI.ExecuteCommand(ctx, sandboxID).
		ExecuteRequest(*daytonaapiclient.NewExecuteRequest(command)).
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
	cmdResult := CommandResult{
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

	// Convert result to JSON
	resultJSON, err := json.MarshalIndent(cmdResult, "", "  ")
	if err != nil {
		return returnCommandError(fmt.Sprintf("Error marshaling result: %v", err), "CommandExecutionError")
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}

// Helper function to return command errors in a consistent format
func returnCommandError(message, errorType string) (*mcp.CallToolResult, error) {
	return &mcp.CallToolResult{
		IsError: true,
		Result: mcp.Result{
			Meta: map[string]interface{}{
				"Stdout":    "",
				"Stderr":    message,
				"ExitCode":  -1,
				"ErrorType": errorType,
			},
		},
	}, nil
}

// Helper function to quote shell commands
func shellQuote(s string) string {
	// Simple shell quoting - wrap in single quotes and escape existing single quotes
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}
