// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package tools

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cli/apiclient"
	"github.com/mark3labs/mcp-go/mcp"

	log "github.com/sirupsen/logrus"
)

type CreateFolderArgs struct {
	Id         *string `json:"id,omitempty"`
	FolderPath *string `json:"folderPath,omitempty"`
	Mode       *string `json:"mode,omitempty"`
}

func GetCreateFolderTool() mcp.Tool {
	return mcp.NewTool("create_folder",
		mcp.WithDescription("Create a new folder in the Daytona sandbox."),
		mcp.WithString("folderPath", mcp.Required(), mcp.Description("Path to the folder to create.")),
		mcp.WithString("mode", mcp.Description("Mode of the folder to create (defaults to 0755).")),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID of the sandbox to create the folder in.")),
	)
}

func CreateFolder(ctx context.Context, request mcp.CallToolRequest, args CreateFolderArgs) (*mcp.CallToolResult, error) {
	apiClient, err := apiclient.GetApiClient(nil, daytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, err
	}

	if args.Id == nil || *args.Id == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("sandbox ID is required")
	}

	if args.FolderPath == nil || *args.FolderPath == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("folderPath parameter is required")
	}

	mode := "0755" // default mode
	if args.Mode == nil || *args.Mode == "" {
		args.Mode = &mode
	}

	// Create the folder
	_, err = apiClient.ToolboxAPI.CreateFolder(ctx, *args.Id).Path(*args.FolderPath).Mode(*args.Mode).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error creating folder: %v", err)
	}

	log.Infof("Created folder: %s", *args.FolderPath)

	return mcp.NewToolResultText(fmt.Sprintf("Created folder: %s", *args.FolderPath)), nil
}
