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

type DeleteFileInput struct {
	SandboxId string `json:"sandboxId" jsonschema:"ID of the sandbox to delete the file from."`
	FilePath  string `json:"filePath" jsonschema:"Path to the file to delete."`
}

type DeleteFileOutput struct {
	Message string `json:"message,omitempty" jsonschema:"Message indicating the successful deletion of the file."`
}

func getDeleteFileTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "delete_file",
		Title:       "Delete File",
		Description: "Delete a file or directory in the Daytona sandbox.",
	}
}

func handleDeleteFile(ctx context.Context, request *mcp.CallToolRequest, input *DeleteFileInput) (*mcp.CallToolResult, *DeleteFileOutput, error) {
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

	if input.FilePath == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("filePath parameter is required")
	}

	// Execute delete command
	_, err = apiClient.ToolboxAPI.DeleteFile(ctx, input.SandboxId).Path(input.FilePath).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error deleting file: %v", err)
	}

	log.Infof("Deleted file: %s", input.FilePath)

	return &mcp.CallToolResult{
			IsError: false,
		}, &DeleteFileOutput{
			Message: fmt.Sprintf("Deleted file: %s", input.FilePath),
		}, nil
}
