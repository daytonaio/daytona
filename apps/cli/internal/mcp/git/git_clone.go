// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"context"
	"fmt"

	"github.com/daytonaio/apiclient"
	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	mcp_headers "github.com/daytonaio/daytona/cli/internal/mcp"
	"github.com/invopop/jsonschema"

	mcp "github.com/modelcontextprotocol/go-sdk/mcp"

	log "github.com/sirupsen/logrus"
)

type GitCloneInput struct {
	SandboxId *string `json:"sandboxId,omitempty" jsonschema:"required,description=ID of the sandbox to clone the repository in."`
	Url       *string `json:"url,omitempty" jsonschema:"required,description=URL of the Git repository to clone."`
	Path      *string `json:"path,omitempty" jsonschema:"default=.,description=Directory to clone the repository into (defaults to current directory)."`
	Branch    *string `json:"branch,omitempty" jsonschema:"default=main,description=Branch to clone."`
	CommitId  *string `json:"commitId,omitempty" jsonschema:"description=Commit ID to clone."`
	Username  *string `json:"username,omitempty" jsonschema:"description=Username to clone the repository with."`
	Password  *string `json:"password,omitempty" jsonschema:"description=Password to clone the repository with."`
}

type GitCloneOutput struct {
	Message string `json:"message" jsonschema:"description=Message indicating the successful cloning of the repository."`
}

func getGitCloneTool() *mcp.Tool {
	return &mcp.Tool{
		Name:         "git_clone",
		Title:        "Git Clone",
		Description:  "Clone a Git repository into the Daytona sandbox.",
		InputSchema:  jsonschema.Reflect(GitCloneInput{}),
		OutputSchema: jsonschema.Reflect(GitCloneOutput{}),
	}
}

func handleGitClone(ctx context.Context, request *mcp.CallToolRequest, input *GitCloneInput) (*mcp.CallToolResult, *GitCloneOutput, error) {
	apiClient, err := apiclient_cli.GetApiClient(nil, mcp_headers.DaytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, err
	}

	if input.SandboxId == nil || *input.SandboxId == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("sandbox ID is required")
	}

	gitCloneRequest, err := getGitCloneRequest(*input)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, err
	}

	_, err = apiClient.ToolboxAPI.GitCloneRepository(ctx, *input.SandboxId).GitCloneRequest(*gitCloneRequest).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("error cloning repository: %v", err)
	}

	log.Infof("Cloned repository: %s to %s", gitCloneRequest.Url, gitCloneRequest.Path)

	return &mcp.CallToolResult{IsError: false}, &GitCloneOutput{
		Message: fmt.Sprintf("Cloned repository: %s to %s", gitCloneRequest.Url, gitCloneRequest.Path),
	}, nil
}

func getGitCloneRequest(input GitCloneInput) (*apiclient.GitCloneRequest, error) {
	gitCloneRequest := apiclient.GitCloneRequest{}

	if input.Url == nil || *input.Url == "" {
		return nil, fmt.Errorf("url parameter is required")
	}

	gitCloneRequest.Url = *input.Url

	gitCloneRequest.Path = "."
	if input.Path != nil && *input.Path != "" {
		gitCloneRequest.Path = *input.Path
	}

	if input.Branch != nil && *input.Branch != "" {
		gitCloneRequest.Branch = input.Branch
	}

	if input.CommitId != nil && *input.CommitId != "" {
		gitCloneRequest.CommitId = input.CommitId
	}

	if input.Username != nil && *input.Username != "" {
		gitCloneRequest.Username = input.Username
	}

	if input.Password != nil && *input.Password != "" {
		gitCloneRequest.Password = input.Password
	}

	return &gitCloneRequest, nil
}
