// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package tools

import (
	"context"
	"fmt"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	"github.com/mark3labs/mcp-go/mcp"

	log "github.com/sirupsen/logrus"
)

type DeleteFileArgs struct {
	Id       *string `json:"id,omitempty"`
	FilePath *string `json:"filePath,omitempty"`
}

func GetDeleteFileTool() mcp.Tool {
	return mcp.NewTool("delete_file",
		mcp.WithDescription("Delete a file or directory in the Daytona sandbox."),
		mcp.WithString("filePath", mcp.Required(), mcp.Description("Path to the file or directory to delete.")),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID of the sandbox to delete the file in.")),
	)
}

func DeleteFile(ctx context.Context, request mcp.CallToolRequest, args DeleteFileArgs) (*mcp.CallToolResult, error) {
	apiClient, err := apiclient_cli.GetApiClient(nil, daytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, err
	}

	if args.Id == nil || *args.Id == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("sandbox ID is required")
	}

	if args.FilePath == nil || *args.FilePath == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("filePath parameter is required")
	}

	// Use the toolbox file API instead of shell command interpolation.
	_, err = apiClient.ToolboxAPI.DeleteFileDeprecated(ctx, *args.Id).Path(*args.FilePath).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error deleting file: %v", err)
	}

	log.Infof("Deleted file: %s", *args.FilePath)

	return mcp.NewToolResultText(fmt.Sprintf("Deleted file: %s", *args.FilePath)), nil
}
