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

func GetFileInfo(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiClient, err := apiclient.GetApiClient(nil, daytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, err
	}

	sandboxId := ""
	if id, ok := request.Params.Arguments["id"]; ok && id != nil {
		if idStr, ok := id.(string); ok && idStr != "" {
			sandboxId = idStr
		}
	}

	if sandboxId == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("sandbox ID is required")
	}

	// Get file path from request arguments
	filePath, ok := request.Params.Arguments["file_path"].(string)
	if !ok {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("file_path parameter is required")
	}

	// Get file info
	fileInfo, _, err := apiClient.ToolboxAPI.GetFileInfo(ctx, sandboxId).Path(filePath).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error getting file info: %v", err)
	}

	// Convert file info to JSON
	fileInfoJSON, err := json.MarshalIndent(fileInfo, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error marshaling file info: %v", err)
	}

	log.Infof("Retrieved file info for: %s", filePath)

	return mcp.NewToolResultText(string(fileInfoJSON)), nil
}
