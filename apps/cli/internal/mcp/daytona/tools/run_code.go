package tools

import (
	"context"
	"fmt"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	mcp_headers "github.com/daytonaio/daytona/cli/internal/mcp"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	log "github.com/sirupsen/logrus"
)

type RunCodeInput struct {
	SandboxId string         `json:"sandboxId" jsonchema:"ID of the sandbox to run the code in. If not provided, a new sandbox will be created."`
	Code      string         `json:"code" jsonchema:"Code to run."`
	Params    *CodeRunParams `json:"params,omitempty" jsonchema:"Parameters for the code run."`
	Timeout   int            `json:"timeout" jsonchema:"Maximum time in seconds to wait for the code to complete. If not provided, the default timeout 0 (meaning indefinitely) will be used."`
}

type CodeRunParams struct {
	Argv []string          `json:"argv,omitempty" jsonchema:"Command line arguments."`
	Env  map[string]string `json:"env,omitempty" jsonchema:"Environment variables."`
}

type ExecuteResponse struct {
	ExitCode             int                    `json:"exitCode" jsonchema:"Exit code of the code run."`
	Result               string                 `json:"result" jsonchema:"Result of the code run."`
	Artifacts            *ExecutionArtifacts    `json:"artifacts" jsonchema:"Artifacts of the code run."`
	AdditionalProperties map[string]interface{} `json:"additionalProperties,omitempty" jsonchema:"Additional properties."`
}

type ExecutionArtifacts struct {
	Stdout string  `json:"stdout" jsonchema:"Standard output of the code run."`
	Charts []Chart `json:"charts,omitempty" jsonchema:"Charts of the code run."`
}

type Chart struct {
	Type     string        `json:"type" jsonchema:"Type of the chart."`
	Title    string        `json:"title" jsonchema:"Title of the chart."`
	Elements []interface{} `json:"elements,omitempty" jsonchema:"Elements of the chart."`
	Png      string        `json:"png,omitempty" jsonchema:"PNG of the chart."`
}

func GetRunCodeTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "run_code",
		Title:       "Run Code",
		Description: "Run code in the Daytona sandbox.",
	}
}

func HandleRunCode(ctx context.Context, request *mcp.CallToolRequest, input *RunCodeInput) (*mcp.CallToolResult, *ExecuteResponse, error) {
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

	return &mcp.CallToolResult{}, nil, nil
}
