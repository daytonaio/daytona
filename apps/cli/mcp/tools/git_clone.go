// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package tools

import (
	"context"
	"fmt"

	"github.com/daytonaio/apiclient"
	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	"github.com/mark3labs/mcp-go/mcp"

	log "github.com/sirupsen/logrus"
)

type GitCloneArgs struct {
	Id       *string `json:"id,omitempty"`
	Url      *string `json:"url,omitempty"`
	Path     *string `json:"path,omitempty"`
	Branch   *string `json:"branch,omitempty"`
	CommitId *string `json:"commitId,omitempty"`
	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
}

func GetGitCloneTool() mcp.Tool {
	return mcp.NewTool("git_clone",
		mcp.WithDescription("Clone a Git repository into the Daytona sandbox."),
		mcp.WithString("url", mcp.Required(), mcp.Description("URL of the Git repository to clone.")),
		mcp.WithString("path", mcp.Description("Directory to clone the repository into (defaults to current directory).")),
		mcp.WithString("branch", mcp.Description("Branch to clone.")),
		mcp.WithString("commitId", mcp.Description("Commit ID to clone.")),
		mcp.WithString("username", mcp.Description("Username to clone the repository with.")),
		mcp.WithString("password", mcp.Description("Password to clone the repository with.")),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID of the sandbox to clone the repository in.")),
	)
}

func GitClone(ctx context.Context, request mcp.CallToolRequest, args GitCloneArgs) (*mcp.CallToolResult, error) {
	apiClient, err := apiclient_cli.GetApiClient(nil, daytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, err
	}

	if args.Id == nil || *args.Id == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("sandbox ID is required")
	}

	gitCloneRequest, err := getGitCloneRequest(args)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, err
	}

	_, err = apiClient.ToolboxAPI.GitCloneRepository(ctx, *args.Id).GitCloneRequest(*gitCloneRequest).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error cloning repository: %v", err)
	}

	log.Infof("Cloned repository: %s to %s", gitCloneRequest.Url, gitCloneRequest.Path)

	return mcp.NewToolResultText(fmt.Sprintf("Cloned repository: %s to %s", gitCloneRequest.Url, gitCloneRequest.Path)), nil
}

func getGitCloneRequest(args GitCloneArgs) (*apiclient.GitCloneRequest, error) {
	gitCloneRequest := apiclient.GitCloneRequest{}

	if args.Url == nil || *args.Url == "" {
		return nil, fmt.Errorf("url parameter is required")
	}

	gitCloneRequest.Url = *args.Url

	gitCloneRequest.Path = "."
	if args.Path != nil && *args.Path != "" {
		gitCloneRequest.Path = *args.Path
	}

	if args.Branch != nil && *args.Branch != "" {
		gitCloneRequest.Branch = args.Branch
	}

	if args.CommitId != nil && *args.CommitId != "" {
		gitCloneRequest.CommitId = args.CommitId
	}

	if args.Username != nil && *args.Username != "" {
		gitCloneRequest.Username = args.Username
	}

	if args.Password != nil && *args.Password != "" {
		gitCloneRequest.Password = args.Password
	}

	return &gitCloneRequest, nil
}
