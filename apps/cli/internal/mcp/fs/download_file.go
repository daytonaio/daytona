// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	mcp_headers "github.com/daytonaio/daytona/cli/internal/mcp"
	"github.com/invopop/jsonschema"
	mcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type FileDownloadInput struct {
	SandboxId *string `json:"sandboxId,omitempty" jsonschema:"required,description=ID of the sandbox to download the file from."`
	FilePath  *string `json:"filePath,omitempty" jsonschema:"required,description=Path to the file to download."`
}

type FileDownloadOutput struct {
	Content []Content `json:"content" jsonschema:"description=Contents of the file."`
}

type Content struct {
	Type string `json:"type" jsonschema:"description=Type of the content."`
	Text string `json:"text,omitempty" jsonschema:"description=Text of the content."`
	Data string `json:"data,omitempty" jsonschema:"description=Data of the content."`
}

func getDownloadFileTool() *mcp.Tool {
	return &mcp.Tool{
		Name:         "download_file",
		Title:        "Download File",
		Description:  "Download a file from the Daytona sandbox. Returns the file content either as text or as a base64 encoded image. Handles special cases like matplotlib plots stored as JSON with embedded base64 images.",
		InputSchema:  jsonschema.Reflect(FileDownloadInput{}),
		OutputSchema: jsonschema.Reflect(FileDownloadOutput{}),
	}
}

func handleDownloadFile(ctx context.Context, request *mcp.CallToolRequest, input *FileDownloadInput) (*mcp.CallToolResult, *FileDownloadOutput, error) {
	apiClient, err := apiclient_cli.GetApiClient(nil, mcp_headers.DaytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, err
	}

	if input.SandboxId == nil || *input.SandboxId == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("sandbox ID is required")
	}

	if input.FilePath == nil || *input.FilePath == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("filePath parameter is required")
	}

	// Download the file
	file, _, err := apiClient.ToolboxAPI.DownloadFile(ctx, *input.SandboxId).Path(*input.FilePath).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error downloading file: %v", err)
	}
	defer file.Close()

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error reading file content: %v", err)
	}

	// Process file content based on file type
	ext := filepath.Ext(*input.FilePath)
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
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error marshaling result: %v", err)
	}

	return &mcp.CallToolResult{
			IsError: false,
		}, &FileDownloadOutput{
			Content: []Content{
				{
					Type: "text",
					Text: string(resultJSON),
				},
			},
		}, nil
}
