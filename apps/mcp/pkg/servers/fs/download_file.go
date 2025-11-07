// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/daytonaio/mcp/internal/apiclient"
	"github.com/daytonaio/mcp/internal/common"
	"github.com/daytonaio/mcp/internal/constants"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type FileDownloadInput struct {
	SandboxId string `json:"sandboxId" jsonschema:"ID of the sandbox to download the file from."`
	FilePath  string `json:"filePath" jsonschema:"Path to the file to download."`
}

type FileDownloadOutput struct {
	Content []Content `json:"content,omitempty" jsonschema:"Contents of the file."`
}

type Content struct {
	Type string `json:"type,omitempty" jsonschema:"Type of the content."`
	Text string `json:"text,omitempty" jsonschema:"Text of the content."`
	Data string `json:"data,omitempty" jsonschema:"Data of the content."`
}

func (s *DaytonaFileSystemMCPServer) getDownloadFileTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "download_file",
		Title:       "Download File",
		Description: "Download a file from the Daytona sandbox. Returns the file content either as text or as a base64 encoded image. Handles special cases like matplotlib plots stored as JSON with embedded base64 images.",
	}
}

func (s *DaytonaFileSystemMCPServer) handleDownloadFile(ctx context.Context, request *mcp.CallToolRequest, input *FileDownloadInput) (*mcp.CallToolResult, *FileDownloadOutput, error) {
	if input.SandboxId == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("sandbox ID is required")
	}

	if input.FilePath == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("filePath parameter is required")
	}

	sandbox, stop, err := common.GetSandbox(ctx, s.apiClient, &input.SandboxId)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("failed to get sandbox: %v", err)
	}
	defer stop()

	proxyUrl, err := apiclient.ExtractProxyUrl(ctx, s.apiClient)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error extracting proxy URL: %v", err)
	}

	toolboxApiClient := apiclient.NewToolboxApiClient(constants.DAYTONA_FS_MCP_SOURCE, sandbox.Id, proxyUrl, request.Extra.Header)

	// Download the file
	file, _, err := toolboxApiClient.FileSystemAPI.DownloadFile(ctx).Path(input.FilePath).Execute()
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
	ext := filepath.Ext(input.FilePath)
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
