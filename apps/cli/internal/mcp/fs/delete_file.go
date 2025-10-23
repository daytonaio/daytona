// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"context"
	"fmt"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	mcp_headers "github.com/daytonaio/daytona/cli/internal/mcp"
	mcp "github.com/modelcontextprotocol/go-sdk/mcp"

	log "github.com/sirupsen/logrus"
)

type DeleteFileInput struct {
	Id       *string `json:"id,omitempty"`
	FilePath *string `json:"filePath,omitempty"`
}

type DeleteFileOutput struct {
	Message string `json:"message" jsonchema:"Message indicating the successful deletion of the file."`
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

	if input.Id == nil || *input.Id == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("sandbox ID is required")
	}

	if input.FilePath == nil || *input.FilePath == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("filePath parameter is required")
	}

	// Execute delete command
	_, err = apiClient.ToolboxAPI.DeleteFile(ctx, *input.Id).Path(*input.FilePath).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error deleting file: %v", err)
	}

	log.Infof("Deleted file: %s", *input.FilePath)

	return &mcp.CallToolResult{
			IsError: false,
		}, &DeleteFileOutput{
			Message: fmt.Sprintf("Deleted file: %s", *input.FilePath),
		}, nil
}
