// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"context"
	"encoding/json"
	"fmt"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	mcp_headers "github.com/daytonaio/daytona/cli/internal/mcp"
	"github.com/invopop/jsonschema"
	mcp "github.com/modelcontextprotocol/go-sdk/mcp"

	log "github.com/sirupsen/logrus"
)

type ListFilesInput struct {
	SandboxId *string `json:"sandboxId,omitempty" jsonschema:"required,description=ID of the sandbox to list the files from."`
	Path      *string `json:"path,omitempty" jsonschema:"required,description=Path to the directory to list files from (defaults to current directory)."`
}

type ListFilesOutput struct {
	Files string `json:"files" jsonschema:"description=List of files in the directory."`
}

func getListFilesTool() *mcp.Tool {
	return &mcp.Tool{
		Name:         "list_files",
		Title:        "List Files",
		Description:  "List files in a directory in the Daytona sandbox.",
		InputSchema:  jsonschema.Reflect(ListFilesInput{}),
		OutputSchema: jsonschema.Reflect(ListFilesOutput{}),
	}
}

func handleListFiles(ctx context.Context, request *mcp.CallToolRequest, input *ListFilesInput) (*mcp.CallToolResult, *ListFilesOutput, error) {
	apiClient, err := apiclient_cli.GetApiClient(nil, mcp_headers.DaytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, err
	}

	if input.SandboxId == nil || *input.SandboxId == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("sandbox ID is required")
	}

	// Get directory path from request arguments (optional)
	dirPath := "."
	if input.Path != nil && *input.Path != "" {
		dirPath = *input.Path
	}

	// List files
	files, _, err := apiClient.ToolboxAPI.ListFiles(ctx, *input.SandboxId).Path(dirPath).Execute()
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
