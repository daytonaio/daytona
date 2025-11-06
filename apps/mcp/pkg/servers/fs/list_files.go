// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/daytonaio/mcp/internal/apiclient"
	"github.com/daytonaio/mcp/internal/common"
	"github.com/daytonaio/mcp/internal/constants"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	log "github.com/sirupsen/logrus"
)

type ListFilesInput struct {
	SandboxId string `json:"sandboxId" jsonschema:"ID of the sandbox to list the files from."`
	Path      string `json:"path" jsonschema:"Path to the directory to list files from (defaults to current directory)."`
}

type ListFilesOutput struct {
	Files string `json:"files,omitempty" jsonschema:"List of files in the directory."`
}

func (s *DaytonaFileSystemMCPServer) getListFilesTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "list_files",
		Title:       "List Files",
		Description: "List files in a directory in the Daytona sandbox.",
	}
}

func (s *DaytonaFileSystemMCPServer) handleListFiles(ctx context.Context, request *mcp.CallToolRequest, input *ListFilesInput) (*mcp.CallToolResult, *ListFilesOutput, error) {
	if input.SandboxId == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("sandbox ID is required")
	}

	// Get directory path from request arguments (optional)
	dirPath := "."
	if input.Path != "" {
		dirPath = input.Path
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

	// List files
	files, _, err := toolboxApiClient.FileSystemAPI.ListFiles(ctx).Path(dirPath).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error listing files: %v", err)
	}

	// Convert files to JSON
	filesJSON, err := json.MarshalIndent(files, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error marshaling files: %v", err)
	}

	log.Infof("Listed files in directory: %s", dirPath)

	return &mcp.CallToolResult{
			IsError: false,
		}, &ListFilesOutput{
			Files: string(filesJSON),
		}, nil
}
