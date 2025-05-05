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

func ListFiles(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiClient, err := apiclient.GetApiClient(nil, daytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, err
	}

	// Get directory path from request arguments (optional)
	dirPath := "."
	if path, ok := request.Params.Arguments["path"].(string); ok && path != "" {
		dirPath = path
	}

	// Get sandbox from tracking file
	sandboxID, err := getActiveSandbox()
	if err != nil || sandboxID == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("no sandbox ID found in tracking file: %v", err)
	}

	// List files
	files, _, err := apiClient.ToolboxAPI.ListFiles(ctx, sandboxID).Path(dirPath).Execute()
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
