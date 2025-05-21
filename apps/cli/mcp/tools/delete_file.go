// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package tools

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cli/apiclient"
	daytonaapiclient "github.com/daytonaio/daytona/daytonaapiclient"
	"github.com/mark3labs/mcp-go/mcp"

	log "github.com/sirupsen/logrus"
)

func DeleteFile(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	// Execute delete command
	execResponse, _, err := apiClient.ToolboxAPI.ExecuteCommand(ctx, sandboxId).
		ExecuteRequest(*daytonaapiclient.NewExecuteRequest(fmt.Sprintf("rm -rf %s", filePath))).
		Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error deleting file: %v", err)
	}

	log.Infof("Deleted file: %s", filePath)

	return mcp.NewToolResultText(fmt.Sprintf("Deleted file: %s\nOutput: %s", filePath, execResponse.Result)), nil
}
