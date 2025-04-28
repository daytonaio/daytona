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

func CreateFolder(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiClient, err := apiclient.GetApiClient(nil, daytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, err
	}

	// Get folder path from request arguments
	folderPath, ok := request.Params.Arguments["folder_path"].(string)
	if !ok {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("folder_path parameter is required")
	}

	mode := "0755"
	modeArg, ok := request.Params.Arguments["mode"].(string)
	if ok && modeArg != "" {
		mode = modeArg
	}

	// Get sandbox from tracking file
	sandboxID, err := getActiveSandbox()
	if err != nil || sandboxID == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("no sandbox ID found in tracking file: %v", err)
	}

	// Create the folder
	_, err = apiClient.ToolboxAPI.CreateFolder(ctx, sandboxID).Path(folderPath).Mode(mode).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error creating folder: %v", err)
	}

	log.Infof("Created folder: %s", folderPath)

	return mcp.NewToolResultText(fmt.Sprintf("Created folder: %s", folderPath)), nil
}
