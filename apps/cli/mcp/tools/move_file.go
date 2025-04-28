// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package tools

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona-ai-saas/cli/apiclient"
	"github.com/mark3labs/mcp-go/mcp"

	log "github.com/sirupsen/logrus"
)

func MoveFile(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiClient, err := apiclient.GetApiClient(nil, daytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, err
	}

	// Get source and destination paths from request arguments
	sourcePath, ok := request.Params.Arguments["source_path"].(string)
	if !ok {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("source_path parameter is required")
	}

	destPath, ok := request.Params.Arguments["dest_path"].(string)
	if !ok {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("dest_path parameter is required")
	}

	// Get sandbox from tracking file
	sandboxID, err := getActiveSandbox()
	if err != nil || sandboxID == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("no sandbox ID found in tracking file: %v", err)
	}

	_, err = apiClient.ToolboxAPI.MoveFile(ctx, sandboxID).Source(sourcePath).Destination(destPath).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error moving file: %v", err)
	}

	log.Infof("Moved file from %s to %s", sourcePath, destPath)

	return mcp.NewToolResultText(fmt.Sprintf("Moved file from %s to %s", sourcePath, destPath)), nil
}
