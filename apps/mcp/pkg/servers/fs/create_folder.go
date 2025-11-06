// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"context"
	"fmt"

	"github.com/daytonaio/mcp/internal/apiclient"
	"github.com/daytonaio/mcp/internal/common"
	"github.com/daytonaio/mcp/internal/constants"

	"github.com/modelcontextprotocol/go-sdk/mcp"

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

func (s *DaytonaFileSystemMCPServer) getCreateFolderTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "create_folder",
		Title:       "Create Folder",
		Description: "Create a new folder in the Daytona sandbox.",
	}
}

func (s *DaytonaFileSystemMCPServer) handleCreateFolder(ctx context.Context, request *mcp.CallToolRequest, input *CreateFolderInput) (*mcp.CallToolResult, *CreateFolderOutput, error) {
	if input.SandboxId == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("sandbox ID is required")
	}

	if input.FolderPath == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("folderPath parameter is required")
	}

	sandbox, stop, err := common.GetSandbox(ctx, s.apiClient, &input.SandboxId)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("failed to get sandbox: %v", err)
	}
	defer stop()

	if input.Mode == "" {
		input.Mode = "0755"
	}

	proxyUrl, err := apiclient.ExtractProxyUrl(ctx, s.apiClient)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error extracting proxy URL: %v", err)
	}

	toolboxApiClient := apiclient.NewToolboxApiClient(constants.DaytonaFsMcpSource, sandbox.Id, proxyUrl, request.Extra.Header)

	// Create the folder
	_, err = toolboxApiClient.FileSystemAPI.CreateFolder(ctx).Path(input.FolderPath).Mode(input.Mode).Execute()
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
