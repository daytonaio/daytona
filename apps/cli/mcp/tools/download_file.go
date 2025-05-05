// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/daytonaio/daytona/cli/apiclient"
	daytonaapiclient "github.com/daytonaio/daytona/daytonaapiclient"
	"github.com/mark3labs/mcp-go/mcp"
)

type Content struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	Data string `json:"data,omitempty"`
}

func DownloadCommand(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiClient, err := apiclient.GetApiClient(nil, daytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, err
	}

	// Get file path from request arguments
	filePath, ok := request.Params.Arguments["file_path"].(string)
	if !ok {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("file_path parameter is required")
	}

	// Get sandbox from tracking file
	sandboxID, err := getActiveSandbox()
	if err != nil || sandboxID == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("no sandbox ID found in tracking file: %v", err)
	}

	// Check if file exists using execute command
	execResponse, _, err := apiClient.ToolboxAPI.ExecuteCommand(ctx, sandboxID).
		ExecuteRequest(*daytonaapiclient.NewExecuteRequest(fmt.Sprintf("test -f %s && echo 'exists' || echo 'not exists'", filePath))).
		Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error checking file existence: %v", err)
	}

	if execResponse.Result != "exists" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("file not found: %s", filePath)
	}

	// Download the file
	file, _, err := apiClient.ToolboxAPI.DownloadFile(ctx, sandboxID).Path(filePath).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error downloading file: %v", err)
	}
	defer file.Close()

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error reading file content: %v", err)
	}

	// Process file content based on file type
	ext := filepath.Ext(filePath)
	var result []Content

	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif":
		// For image files, return as base64 encoded data
		result = []Content{{
			Type: "image",
			Data: string(content),
		}}
	case ".json":
		// For JSON files, try to parse and handle special cases like matplotlib plots
		var jsonData map[string]interface{}
		if err := json.Unmarshal(content, &jsonData); err != nil {
			// If not valid JSON, return as text
			result = []Content{{
				Type: "text",
				Text: string(content),
			}}
		} else {
			// Check if it's a matplotlib plot
			if _, ok := jsonData["data"]; ok {
				result = []Content{{
					Type: "image",
					Data: jsonData["data"].(string),
				}}
			} else {
				result = []Content{{
					Type: "text",
					Text: string(content),
				}}
			}
		}
	default:
		// For all other files, return as text
		result = []Content{{
			Type: "text",
			Text: string(content),
		}}
	}

	// Convert result to JSON
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error marshaling result: %v", err)
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}
