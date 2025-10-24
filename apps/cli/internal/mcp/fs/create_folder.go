// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"context"
	"fmt"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	mcp_headers "github.com/daytonaio/daytona/cli/internal/mcp"
	"github.com/invopop/jsonschema"
	mcp "github.com/modelcontextprotocol/go-sdk/mcp"

	log "github.com/sirupsen/logrus"
)

type CreateFolderInput struct {
	SandboxId  *string `json:"sandboxId,omitempty" jsonschema:"required,type=string,description=ID of the sandbox to create the folder in."`
	FolderPath *string `json:"folderPath,omitempty" jsonschema:"required,type=string,description=Path to the folder to create."`
	Mode       *string `json:"mode,omitempty" jsonschema:"default=0755,type=string,description=Mode of the folder to create (defaults to 0755)."`
}

type CreateFolderOutput struct {
	Message string `json:"message" jsonschema:"type=string,description=Message indicating the successful creation of the folder."`
}

func getCreateFolderTool() *mcp.Tool {
	return &mcp.Tool{
		Name:         "create_folder",
		Title:        "Create Folder",
		Description:  "Create a new folder in the Daytona sandbox.",
		InputSchema:  jsonschema.Reflect(CreateFolderInput{}),
		OutputSchema: jsonschema.Reflect(CreateFolderOutput{}),
	}
}

func handleCreateFolder(ctx context.Context, request *mcp.CallToolRequest, input *CreateFolderInput) (*mcp.CallToolResult, *CreateFolderOutput, error) {
	apiClient, err := apiclient_cli.GetApiClient(nil, mcp_headers.DaytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, err
	}

	if input.SandboxId == nil || *input.SandboxId == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("sandbox ID is required")
	}

	if input.FolderPath == nil || *input.FolderPath == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("folderPath parameter is required")
	}

	mode := "0755" // default mode
	if input.Mode == nil || *input.Mode == "" {
		input.Mode = &mode
	}

	// Create the folder
	_, err = apiClient.ToolboxAPI.CreateFolder(ctx, *input.SandboxId).Path(*input.FolderPath).Mode(*input.Mode).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error creating folder: %v", err)
	}

	log.Infof("Created folder: %s", *input.FolderPath)

	return &mcp.CallToolResult{
			IsError: false,
		}, &CreateFolderOutput{
			Message: fmt.Sprintf("Created folder: %s", *input.FolderPath),
		}, nil
}
