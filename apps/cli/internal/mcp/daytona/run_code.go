// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package daytona

import (
	"context"
	"fmt"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	mcp_headers "github.com/daytonaio/daytona/cli/internal/mcp"
	"github.com/daytonaio/daytona/cli/internal/mcp/util"
	"github.com/invopop/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	log "github.com/sirupsen/logrus"
)

type RunCodeInput struct {
	SandboxId *string        `json:"sandboxId" jsonschema:"type=string,description=ID of the sandbox to run the code in. Don't provide this if not explicitly instructed from user. If not provided, a new sandbox will be created."`
	Code      string         `json:"code" jsonschema:"required,type=string,description=Code to run."`
	Params    *CodeRunParams `json:"params,omitempty" jsonschema:"type=object,description=Parameters for the code run."`
	Timeout   int            `json:"timeout" jsonschema:"required,default=0,type=integer,description=Maximum time in seconds to wait for the code to complete. If not provided, the default timeout 0 (meaning indefinitely) will be used."`
}

type CodeRunParams struct {
	Argv []string          `json:"argv,omitempty" jsonschema:"type=array,items=string,description=Command line arguments."`
	Env  map[string]string `json:"env,omitempty" jsonschema:"type=object,additionalProperties=string,description=Environment variables."`
}

type ExecuteResponse struct {
	ExitCode             int                    `json:"exitCode" jsonschema:"type=integer,description=Exit code of the code run."`
	Result               string                 `json:"result" jsonschema:"type=string,description=Result of the code run."`
	Artifacts            *ExecutionArtifacts    `json:"artifacts" jsonschema:"type=object,description=Artifacts of the code run."`
	AdditionalProperties map[string]interface{} `json:"additionalProperties,omitempty" jsonschema:"type=object,additionalProperties=any,description=Additional properties."`
}

type ExecutionArtifacts struct {
	Stdout string  `json:"stdout" jsonschema:"type=string,description=Standard output of the code run."`
	Charts []Chart `json:"charts,omitempty" jsonschema:"type=array,items=object,description=Charts of the code run."`
}

type Chart struct {
	Type     string        `json:"type" jsonschema:"type=string,description=Type of the chart."`
	Title    string        `json:"title" jsonschema:"type=string,description=Title of the chart."`
	Elements []interface{} `json:"elements,omitempty" jsonschema:"type=array,items=object,description=Elements of the chart."`
	Png      string        `json:"png,omitempty" jsonschema:"type=string,description=PNG of the chart."`
}

func getRunCodeTool() *mcp.Tool {
	return &mcp.Tool{
		Name:         "run_code",
		Title:        "Run Code",
		Description:  "Run code in the Daytona sandbox.",
		InputSchema:  jsonschema.Reflect(RunCodeInput{}),
		OutputSchema: jsonschema.Reflect(ExecuteResponse{}),
	}
}

func handleRunCode(ctx context.Context, request *mcp.CallToolRequest, input *RunCodeInput) (*mcp.CallToolResult, *ExecuteResponse, error) {
	return &mcp.CallToolResult{IsError: false}, nil, fmt.Errorf("not implemented")
	apiClient, err := apiclient_cli.GetApiClient(nil, mcp_headers.DaytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, err
	}

	if input.Code == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("code is required")
	}

	if input.Timeout <= 0 {
		log.Warnf("Timeout is less than 0, setting to 0")
		input.Timeout = 0
	}

	_, err = util.GetSandbox(ctx, apiClient, input.SandboxId)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("failed to get sandbox: %v", err)
	}

	// TODO: Implement code execution
	return &mcp.CallToolResult{IsError: false}, nil, nil

	// timeout := float32(input.Timeout)

	// executeResponse, resp, err := apiClient.ToolboxAPI.ExecuteCommand(ctx, input.SandboxId).ExecuteRequest(apiclient.ExecuteRequest{
	// 	Command: input.Code,
	// 	AdditionalProperties: map[string]interface{}{
	// 		"env": input.Params.Env,
	// 		"argv": input.Params.Argv,
	// 	},
	// 	Timeout: &timeout,
	// }).Execute()
	// if err != nil {
	// 	return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("failed to execute code: %v", err)
	// }

	// return &mcp.CallToolResult{
	// 	IsError: false,
	// }, &ExecuteResponse{
	// 	ExitCode: int(result.ExitCode),
	// 	Result:   result.Result,
	// 	Artifacts: &ExecutionArtifacts{
	// 		Stdout: result.Stdout,
	// 	},
	// }, nil
}
