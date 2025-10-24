// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"context"
	"fmt"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	mcp_headers "github.com/daytonaio/daytona/cli/internal/mcp"
	mcp "github.com/modelcontextprotocol/go-sdk/mcp"

	log "github.com/sirupsen/logrus"
)

type CreateFolderInput struct {
	SandboxId  string `json:"sandboxId" jsonschema:"ID of the sandbox to create the folder in."`
	FolderPath string `json:"folderPath" jsonschema:"Path to the folder to create."`
	Mode       string `json:"mode,omitempty" jsonschema:"Mode of the folder to create (defaults to 0755)."`
}

type CreateFolderOutput struct {
	Message string `json:"message,omitempty" jsonschema:"Message indicating the successful creation of the folder."`
}

func getCreateFolderTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "create_folder",
		Title:       "Create Folder",
		Description: "Create a new folder in the Daytona sandbox.",
	}
}

func handleCreateFolder(ctx context.Context, request *mcp.CallToolRequest, input *CreateFolderInput) (*mcp.CallToolResult, *CreateFolderOutput, error) {
	apiClient, err := apiclient_cli.GetApiClient(nil, mcp_headers.DaytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, err
	}

	if input.SandboxId == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("sandbox ID is required")
	}

	if input.FolderPath == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("folderPath parameter is required")
	}

	if input.Mode == "" {
		input.Mode = "0755"
	}

	// Create the folder
	_, err = apiClient.ToolboxAPI.CreateFolder(ctx, input.SandboxId).Path(input.FolderPath).Mode(input.Mode).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error creating folder: %v", err)
	}

	log.Infof("Created folder: %s", input.FolderPath)

	return &mcp.CallToolResult{
			IsError: false,
		}, &CreateFolderOutput{
			Message: fmt.Sprintf("Created folder: %s", input.FolderPath),
		}, nil
}
