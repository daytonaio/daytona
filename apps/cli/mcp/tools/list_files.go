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

type ListFilesArgs struct {
	Id   *string `json:"id,omitempty"`
	Path *string `json:"path,omitempty"`
}

func GetListFilesTool() mcp.Tool {
	return mcp.NewTool("list_files",
		mcp.WithDescription("List files in a directory in the Daytona sandbox."),
		mcp.WithString("path", mcp.Description("Path to the directory to list files from (defaults to current directory).")),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID of the sandbox to list the files from.")),
	)
}

func ListFiles(ctx context.Context, request mcp.CallToolRequest, args ListFilesArgs) (*mcp.CallToolResult, error) {
	apiClient, err := apiclient.GetApiClient(nil, daytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, err
	}

	if args.Id == nil || *args.Id == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("sandbox ID is required")
	}

	// Get directory path from request arguments (optional)
	dirPath := "."
	if args.Path != nil && *args.Path != "" {
		dirPath = *args.Path
	}

	// List files
	files, _, err := apiClient.ToolboxAPI.ListFiles(ctx, *args.Id).Path(dirPath).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error listing files: %v", err)
	}

	// Convert files to JSON
	filesJSON, err := json.MarshalIndent(files, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error marshaling files: %v", err)
	}

	log.Infof("Listed files in directory: %s", dirPath)

	return mcp.NewToolResultText(string(filesJSON)), nil
}
