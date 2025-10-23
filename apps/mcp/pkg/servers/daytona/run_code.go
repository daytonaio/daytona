// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package daytona

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type RunCodeInput struct {
	SandboxId *string       `json:"sandboxId,omitempty" jsonschema:"ID of the sandbox to run the code in. Don't provide this if not explicitly instructed from user. If not provided, a new sandbox will be created."`
	Code      string        `json:"code" jsonschema:"Code to run."`
	Params    CodeRunParams `json:"params,omitempty" jsonschema:"Parameters for the code run."`
	Timeout   *int          `json:"timeout,omitempty" jsonschema:"Maximum time in seconds to wait for the code to complete. If not provided, the default timeout 0 (meaning indefinitely) will be used."`
}

type CodeRunParams struct {
	Argv []string          `json:"argv,omitempty" jsonschema:"Command line arguments."`
	Env  map[string]string `json:"env,omitempty" jsonschema:"Environment variables."`
}

type ExecuteResponse struct {
	ExitCode             *int                   `json:"exitCode,omitempty" jsonschema:"Exit code of the code run."`
	Result               *string                `json:"result,omitempty" jsonschema:"Result of the code run."`
	Artifacts            *ExecutionArtifacts    `json:"artifacts,omitempty" jsonschema:"Artifacts of the code run."`
	AdditionalProperties map[string]interface{} `json:"additionalProperties,omitempty" jsonschema:"Additional properties."`
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

func (s *DaytonaMCPServer) getRunCodeTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "run_code",
		Title:       "Run Code",
		Description: "Run code in the Daytona sandbox.",
	}
}

func (s *DaytonaMCPServer) handleRunCode(ctx context.Context, request *mcp.CallToolRequest, input *RunCodeInput) (*mcp.CallToolResult, *ExecuteResponse, error) {
	// TODO: implement once code interpreter is finished
	return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("not implemented")
}
