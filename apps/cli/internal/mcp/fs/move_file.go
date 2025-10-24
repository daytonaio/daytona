// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"context"
	"fmt"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	mcp_headers "github.com/daytonaio/daytona/cli/internal/mcp"
	"github.com/daytonaio/daytona/cli/internal/mcp/util"
	mcp "github.com/modelcontextprotocol/go-sdk/mcp"

	log "github.com/sirupsen/logrus"
)

type MoveFileInput struct {
	SandboxId  string `json:"sandboxId" jsonschema:"ID of the sandbox to move the file in."`
	SourcePath string `json:"sourcePath" jsonschema:"Source path of the file to move."`
	DestPath   string `json:"destPath" jsonschema:"Destination path where to move the file."`
}

type MoveFileOutput struct {
	Message string `json:"message,omitempty" jsonschema:"Message indicating the successful movement of the file."`
}

func getMoveFileTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "move_file",
		Title:       "Move File",
		Description: "Move or rename a file in the Daytona sandbox.",
	}
}

func handleMoveFile(ctx context.Context, request *mcp.CallToolRequest, input *MoveFileInput) (*mcp.CallToolResult, *MoveFileOutput, error) {
	apiClient, err := apiclient_cli.GetApiClient(nil, mcp_headers.DaytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, err
	}

	if input.SandboxId == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("sandbox ID is required")
	}

	_, err = util.GetSandbox(ctx, apiClient, &input.SandboxId)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("failed to get sandbox: %v", err)
	}

	// Get source and destination paths from request arguments
	if input.SourcePath == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("sourcePath parameter is required")
	}

	if input.DestPath == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("destPath parameter is required")
	}

	_, err = apiClient.ToolboxAPI.MoveFile(ctx, input.SandboxId).Source(input.SourcePath).Destination(input.DestPath).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error moving file: %v", err)
	}

	log.Infof("Moved file from %s to %s", input.SourcePath, input.DestPath)

	return &mcp.CallToolResult{
			IsError: false,
		}, &MoveFileOutput{
			Message: fmt.Sprintf("Moved file from %s to %s", input.SourcePath, input.DestPath),
		}, nil
}
