// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/daytonaio/mcp/internal/common"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	log "github.com/sirupsen/logrus"
)

type FileInfoInput struct {
	SandboxId string `json:"sandboxId" jsonschema:"ID of the sandbox to get the file information from."`
	FilePath  string `json:"filePath" jsonschema:"Path to the file to get information about."`
}

type FileInfoOutput struct {
	FileInfo string `json:"fileInfo,omitempty" jsonschema:"Information about the file."`
}

func (s *DaytonaFileSystemMCPServer) getFileInfoTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "get_file_info",
		Title:       "Get File Info",
		Description: "Get information about a file in the Daytona sandbox.",
	}
}

func (s *DaytonaFileSystemMCPServer) handleFileInfo(ctx context.Context, request *mcp.CallToolRequest, input *FileInfoInput) (*mcp.CallToolResult, *FileInfoOutput, error) {
	if input.SandboxId == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("sandbox ID is required")
	}

	_, err := common.GetSandbox(ctx, s.apiClient, &input.SandboxId)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("failed to get sandbox: %v", err)
	}

	if input.FilePath == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("filePath parameter is required")
	}

	// Get file info
	fileInfo, _, err := s.apiClient.ToolboxAPI.GetFileInfo(ctx, input.SandboxId).Path(input.FilePath).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error getting file info: %v", err)
	}

	// Convert file info to JSON
	fileInfoJSON, err := json.MarshalIndent(fileInfo, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error marshaling file info: %v", err)
	}

	log.Infof("Retrieved file info for: %s", input.FilePath)

	return &mcp.CallToolResult{
			IsError: false,
		}, &FileInfoOutput{
			FileInfo: string(fileInfoJSON),
		}, nil
}
