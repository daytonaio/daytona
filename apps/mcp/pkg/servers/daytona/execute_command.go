// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package daytona

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/daytonaio/toolbox_apiclient"

	"github.com/daytonaio/mcp/internal/apiclient"
	"github.com/daytonaio/mcp/internal/common"
	"github.com/daytonaio/mcp/internal/constants"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	log "github.com/sirupsen/logrus"
)

type ExecuteCommandInput struct {
	SandboxId *string           `json:"sandboxId,omitempty" jsonschema:"ID of the sandbox to execute the command in. Don't provide this if not explicitly instructed from user. If not provided, a new sandbox will be created."`
	Command   string            `json:"command" jsonschema:"Shell command to execute."`
	Cwd       *string           `json:"cwd,omitempty" jsonschema:"Working directory for command execution. If not specified, uses the sandbox working directory."`
	Env       map[string]string `json:"env,omitempty" jsonschema:"Environment variables to set for the command."`
	Timeout   *int              `json:"timeout,omitempty" jsonschema:"Maximum time in seconds to wait for the command to complete. 0 means wait indefinitely."`
}

type ExecuteCommandOutput struct {
	ExitCode  *int                `json:"exitCode,omitempty" jsonschema:"Exit code of the command."`
	Result    *string             `json:"result,omitempty" jsonschema:"Standard output from the command."`
	Artifacts *ExecutionArtifacts `json:"artifacts,omitempty" jsonschema:"Artifacts from the command execution containing stdout and charts."`
	Stderr    *string             `json:"stderr,omitempty" jsonschema:"Standard error output of the command."`
	ErrorType *string             `json:"errorType,omitempty" jsonschema:"Error type of the command."`
}

type ExecutionArtifacts struct {
	Stdout *string `json:"stdout,omitempty" jsonschema:"Standard output of the code run."`
	Charts []Chart `json:"charts,omitempty" jsonschema:"Charts of the code run."`
}

type Chart struct {
	Type     string        `json:"type,omitempty" jsonschema:"Type of the chart."`
	Title    string        `json:"title,omitempty" jsonschema:"Title of the chart."`
	Elements []interface{} `json:"elements,omitempty" jsonschema:"Elements of the chart."`
	Png      string        `json:"png,omitempty" jsonschema:"PNG of the chart."`
}

func (s *DaytonaMCPServer) getExecuteCommandTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "executeCommand",
		Title:       "Execute Command",
		Description: "Executes a shell command in the Sandbox with support for working directory, environment variables, and timeout.",
	}
}

func (s *DaytonaMCPServer) handleExecuteCommand(ctx context.Context, request *mcp.CallToolRequest, input *ExecuteCommandInput) (*mcp.CallToolResult, *ExecuteCommandOutput, error) {
	if input.Command == "" {
		return returnExecuteCommandError("Command must be a non-empty string", "ValueError")
	}

	sandbox, stop, err := common.GetSandbox(ctx, s.apiClient, input.SandboxId)
	if err != nil {
		return returnExecuteCommandError(fmt.Sprintf("Error getting sandbox: %v", err), "SandboxError")
	}
	defer stop()

	// Base64 encode the command
	base64UserCmd := base64.StdEncoding.EncodeToString([]byte(input.Command))
	command := fmt.Sprintf("echo '%s' | base64 -d | sh", base64UserCmd)

	// Add environment variables if provided
	if len(input.Env) > 0 {
		var envExports []string
		for key, value := range input.Env {
			encodedValue := base64.StdEncoding.EncodeToString([]byte(value))
			envExports = append(envExports, fmt.Sprintf("export %s=$(echo '%s' | base64 -d)", key, encodedValue))
		}
		command = strings.Join(envExports, "; ") + "; " + command
	}

	// Wrap in sh -c
	command = fmt.Sprintf(`sh -c "%s"`, command)

	proxyUrl, err := apiclient.ExtractProxyUrl(ctx, s.apiClient)
	if err != nil {
		return returnExecuteCommandError(fmt.Sprintf("Error extracting proxy URL: %v", err), "ProxyUrlError")
	}

	toolboxApiClient := apiclient.NewToolboxApiClient(constants.DAYTONA_MCP_SOURCE, sandbox.Id, proxyUrl, request.Extra.Header)

	log.Infof("Executing command: %s", input.Command)

	// Build the execute request
	executeRequest := toolbox_apiclient.NewExecuteRequest(command)
	if input.Cwd != nil {
		executeRequest.SetCwd(*input.Cwd)
	}
	if input.Timeout != nil {
		executeRequest.SetTimeout(int32(*input.Timeout))
	}

	// Execute the command
	result, _, err := toolboxApiClient.ProcessAPI.ExecuteCommand(ctx).Request(*executeRequest).Execute()
	if err != nil {
		// Classify error types
		errStr := err.Error()
		switch {
		case strings.Contains(errStr, "Connection") || strings.Contains(errStr, "Timeout"):
			return returnExecuteCommandError(fmt.Sprintf("Network error during command execution: %s", errStr), "NetworkError")
		case strings.Contains(errStr, "Unauthorized") || strings.Contains(errStr, "401"):
			return returnExecuteCommandError("Authentication failed during command execution. Please check your API key", "NetworkError")
		default:
			return returnExecuteCommandError(fmt.Sprintf("Command execution failed: %s", errStr), "CommandExecutionError")
		}
	}

	// Parse artifacts from the output
	artifacts := parseArtifacts(result.Result)
	var stdout string
	if artifacts.Stdout != nil {
		stdout = strings.TrimSpace(*artifacts.Stdout)
	}
	exitCode := int(*result.ExitCode)

	// Process command output
	cmdResult := ExecuteCommandOutput{
		ExitCode:  &exitCode,
		Result:    &stdout,
		Artifacts: artifacts,
	}

	// Log truncated output
	outputLen := len(stdout)
	logOutput := stdout
	if outputLen > 500 {
		logOutput = stdout[:500] + "..."
	}

	log.Infof("Command completed - exit code: %d, output length: %d", exitCode, outputLen)

	log.Debugf("Command output (truncated): %s", logOutput)

	// Check for non-zero exit code
	if exitCode > 0 {
		log.Infof("Command exited with non-zero status - exit code: %d", exitCode)
	}

	return &mcp.CallToolResult{
		IsError: false,
	}, &cmdResult, nil
}

// Helper function to return command errors in a consistent format
func returnExecuteCommandError(message, errorType string) (*mcp.CallToolResult, *ExecuteCommandOutput, error) {
	exitCode := -1
	return &mcp.CallToolResult{
			IsError: true,
		},
		&ExecuteCommandOutput{
			Result:    nil,
			Stderr:    &message,
			ExitCode:  &exitCode,
			ErrorType: &errorType,
		}, fmt.Errorf("error: %s", message)
}

// parseArtifacts parses artifacts from command output
// Looks for lines starting with "dtn_artifact_k39fd2:" and extracts chart metadata
func parseArtifacts(output string) *ExecutionArtifacts {
	charts := []Chart{}
	stdout := output

	if output == "" {
		artifacts := &ExecutionArtifacts{
			Stdout: &stdout,
			Charts: charts,
		}
		return artifacts
	}

	// Split output by lines to find artifact markers
	lines := strings.Split(output, "\n")
	artifactLines := []string{}

	for _, line := range lines {
		// Look for the artifact marker pattern
		if strings.HasPrefix(line, "dtn_artifact_k39fd2:") {
			artifactLines = append(artifactLines, line)

			// Try to parse the artifact
			artifactJson := strings.TrimSpace(strings.TrimPrefix(line, "dtn_artifact_k39fd2:"))
			var artifactData map[string]interface{}
			if err := json.Unmarshal([]byte(artifactJson), &artifactData); err == nil {
				if artifactType, ok := artifactData["type"].(string); ok && artifactType == "chart" {
					if value, ok := artifactData["value"].(map[string]interface{}); ok {
						chart := parseChart(value)
						if chart != nil {
							charts = append(charts, *chart)
						}
					}
				}
			}
		}
	}

	// Remove artifact lines from stdout along with their following newlines
	for _, line := range artifactLines {
		stdout = strings.ReplaceAll(stdout, line+"\n", "")
		stdout = strings.ReplaceAll(stdout, line, "")
	}

	artifacts := &ExecutionArtifacts{
		Stdout: &stdout,
		Charts: charts,
	}

	return artifacts
}

// parseChart parses chart data from artifact value
func parseChart(chartData map[string]interface{}) *Chart {
	chart := &Chart{}

	if chartType, ok := chartData["type"].(string); ok {
		chart.Type = chartType
	}
	if title, ok := chartData["title"].(string); ok {
		chart.Title = title
	}
	if elements, ok := chartData["elements"].([]interface{}); ok {
		chart.Elements = elements
	}
	if png, ok := chartData["png"].(string); ok {
		chart.Png = png
	}

	return chart
}
