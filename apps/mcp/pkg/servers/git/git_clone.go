// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"context"
	"fmt"

	"github.com/daytonaio/mcp/internal/apiclient"
	"github.com/daytonaio/mcp/internal/common"
	"github.com/daytonaio/mcp/internal/constants"
	"github.com/daytonaio/toolbox_apiclient"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	log "github.com/sirupsen/logrus"
)

type GitCloneInput struct {
	SandboxId string `json:"sandboxId" jsonschema:"ID of the sandbox to clone the repository in."`
	Url       string `json:"url" jsonschema:"URL of the Git repository to clone."`
	Path      string `json:"path,omitempty" jsonschema:"Directory to clone the repository into (defaults to current directory)."`
	Branch    string `json:"branch,omitempty" jsonschema:"Branch to clone."`
	CommitId  string `json:"commitId,omitempty" jsonschema:"Commit ID to clone."`
	Username  string `json:"username,omitempty" jsonschema:"Username to clone the repository with."`
	Password  string `json:"password,omitempty" jsonschema:"Password to clone the repository with."`
}

type GitCloneOutput struct {
	Message string `json:"message,omitempty" jsonschema:"Message indicating the successful cloning of the repository."`
}

func (s *DaytonaGitMCPServer) getGitCloneTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "git_clone",
		Title:       "Git Clone",
		Description: "Clone a Git repository into the Daytona sandbox.",
	}
}

func (s *DaytonaGitMCPServer) handleGitClone(ctx context.Context, request *mcp.CallToolRequest, input *GitCloneInput) (*mcp.CallToolResult, *GitCloneOutput, error) {
	if input.SandboxId == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("sandbox ID is required")
	}

	if input.Url == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("url parameter is required")
	}

	if input.Path == "" {
		input.Path = "."
	}

	if input.Branch == "" {
		input.Branch = "main"
	}

	sandbox, stop, err := common.GetSandbox(ctx, s.apiClient, &input.SandboxId)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("failed to get sandbox: %v", err)
	}
	defer stop()

	gitCloneRequest := toolbox_apiclient.GitCloneRequest{
		Url:      input.Url,
		Path:     input.Path,
		Branch:   &input.Branch,
		CommitId: &input.CommitId,
		Username: &input.Username,
		Password: &input.Password,
	}

	proxyUrl, err := apiclient.ExtractProxyUrl(ctx, s.apiClient)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error extracting proxy URL: %v", err)
	}

	toolboxApiClient := apiclient.NewToolboxApiClient(constants.DaytonaGitMcpSource, sandbox.Id, proxyUrl, request.Extra.Header)

	_, err = toolboxApiClient.GitAPI.CloneRepository(ctx).Request(gitCloneRequest).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error cloning repository: %v", err)
	}

	log.Infof("Cloned repository: %s to %s", gitCloneRequest.Url, gitCloneRequest.Path)

	return &mcp.CallToolResult{IsError: false}, &GitCloneOutput{
		Message: fmt.Sprintf("Cloned repository: %s to %s", gitCloneRequest.Url, gitCloneRequest.Path),
	}, nil
}
