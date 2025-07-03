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

type MoveFileArgs struct {
	Id         *string `json:"id,omitempty"`
	SourcePath *string `json:"source_path,omitempty"`
	DestPath   *string `json:"dest_path,omitempty"`
}

func GetMoveFileTool() mcp.Tool {
	return mcp.NewTool("move_file",
		mcp.WithDescription("Move or rename a file in the Daytona sandbox."),
		mcp.WithString("source_path", mcp.Required(), mcp.Description("Source path of the file to move.")),
		mcp.WithString("dest_path", mcp.Required(), mcp.Description("Destination path where to move the file.")),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID of the sandbox to move the file in.")),
	)
}

func MoveFile(ctx context.Context, request mcp.CallToolRequest, args MoveFileArgs) (*mcp.CallToolResult, error) {
	apiClient, err := apiclient.GetApiClient(nil, daytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, err
	}

	if args.Id == nil || *args.Id == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("sandbox ID is required")
	}

	// Get source and destination paths from request arguments
	if args.SourcePath == nil || *args.SourcePath == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("source_path parameter is required")
	}

	if args.DestPath == nil || *args.DestPath == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("dest_path parameter is required")
	}

	_, err = apiClient.ToolboxAPI.MoveFile(ctx, *args.Id).Source(*args.SourcePath).Destination(*args.DestPath).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error moving file: %v", err)
	}

	log.Infof("Moved file from %s to %s", *args.SourcePath, *args.DestPath)

	return mcp.NewToolResultText(fmt.Sprintf("Moved file from %s to %s", *args.SourcePath, *args.DestPath)), nil
}
