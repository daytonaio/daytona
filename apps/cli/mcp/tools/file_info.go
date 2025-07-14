// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/daytonaio/daytona/cli/apiclient"
	"github.com/mark3labs/mcp-go/mcp"

	log "github.com/sirupsen/logrus"
)

type FileInfoArgs struct {
	Id       *string `json:"id,omitempty"`
	FilePath *string `json:"filePath,omitempty"`
}

func GetFileInfoTool() mcp.Tool {
	return mcp.NewTool("get_file_info",
		mcp.WithDescription("Get information about a file in the Daytona sandbox."),
		mcp.WithString("filePath", mcp.Required(), mcp.Description("Path to the file to get information about.")),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID of the sandbox to get the file information from.")),
	)
}

func FileInfo(ctx context.Context, request mcp.CallToolRequest, args FileInfoArgs) (*mcp.CallToolResult, error) {
	apiClient, err := apiclient.GetApiClient(nil, daytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, err
	}

	if args.Id == nil || *args.Id == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("sandbox ID is required")
	}

	if args.FilePath == nil || *args.FilePath == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("filePath parameter is required")
	}

	// Get file info
	fileInfo, _, err := apiClient.ToolboxAPI.GetFileInfo(ctx, *args.Id).Path(*args.FilePath).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error getting file info: %v", err)
	}

	// Convert file info to JSON
	fileInfoJSON, err := json.MarshalIndent(fileInfo, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error marshaling file info: %v", err)
	}

	log.Infof("Retrieved file info for: %s", *args.FilePath)

	return mcp.NewToolResultText(string(fileInfoJSON)), nil
}
