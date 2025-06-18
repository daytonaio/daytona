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

type DownloadFileArgs struct {
	Id       *string `json:"id,omitempty"`
	FilePath *string `json:"file_path,omitempty"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	Data string `json:"data,omitempty"`
}

func GetDownloadFileTool() mcp.Tool {
	return mcp.NewTool("download_file",
		mcp.WithDescription("Download a file from the Daytona sandbox. Returns the file content either as text or as a base64 encoded image. Handles special cases like matplotlib plots stored as JSON with embedded base64 images."),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the file to download.")),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID of the sandbox to download the file from.")),
	)
}

func DownloadCommand(ctx context.Context, request mcp.CallToolRequest, args DownloadFileArgs) (*mcp.CallToolResult, error) {
	apiClient, err := apiclient.GetApiClient(nil, daytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, err
	}

	if args.Id == nil || *args.Id == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("sandbox ID is required")
	}

	if args.FilePath == nil || *args.FilePath == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("file_path parameter is required")
	}

	// Check if file exists using execute command
	execResponse, _, err := apiClient.ToolboxAPI.ExecuteCommand(ctx, *args.Id).
		ExecuteRequest(*daytonaapiclient.NewExecuteRequest(fmt.Sprintf("test -f %s && echo 'exists' || echo 'not exists'", *args.FilePath))).
		Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error checking file existence: %v", err)
	}

	if execResponse.Result != "exists" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("file not found: %s", *args.FilePath)
	}

	// Download the file
	file, _, err := apiClient.ToolboxAPI.DownloadFile(ctx, *args.Id).Path(*args.FilePath).Execute()
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
	ext := filepath.Ext(*args.FilePath)
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
