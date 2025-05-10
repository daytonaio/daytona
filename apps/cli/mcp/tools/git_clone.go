// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package tools

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cli/apiclient"
	daytonaapiclient "github.com/daytonaio/daytona/daytonaapiclient"
	"github.com/mark3labs/mcp-go/mcp"

	log "github.com/sirupsen/logrus"
)

func GitClone(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiClient, err := apiclient.GetApiClient(nil, daytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, err
	}

	sandboxId := ""
	if id, ok := request.Params.Arguments["id"]; ok && id != nil {
		if idStr, ok := id.(string); ok && idStr != "" {
			sandboxId = idStr
		}
	}

	if sandboxId == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("sandbox ID is required")
	}

	gitCloneRequest, err := getGitCloneRequest(request)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, err
	}

	_, err = apiClient.ToolboxAPI.GitCloneRepository(ctx, sandboxId).GitCloneRequest(*gitCloneRequest).Execute()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("error cloning repository: %v", err)
	}

	log.Infof("Cloned repository: %s to %s", gitCloneRequest.Url, gitCloneRequest.Path)

	return mcp.NewToolResultText(fmt.Sprintf("Cloned repository: %s to %s", gitCloneRequest.Url, gitCloneRequest.Path)), nil
}

func getGitCloneRequest(request mcp.CallToolRequest) (*daytonaapiclient.GitCloneRequest, error) {
	gitCloneRequest := daytonaapiclient.GitCloneRequest{}

	url, ok := request.Params.Arguments["url"].(string)
	if !ok {
		return nil, fmt.Errorf("url parameter is required")
	}

	gitCloneRequest.Url = url

	gitCloneRequest.Path = "."
	path, ok := request.Params.Arguments["path"].(string)
	if ok && path != "" {
		gitCloneRequest.Path = path
	}

	branch, ok := request.Params.Arguments["branch"].(string)
	if ok && branch != "" {
		gitCloneRequest.Branch = &branch
	}

	commitId, ok := request.Params.Arguments["commit_id"].(string)
	if ok && commitId != "" {
		gitCloneRequest.CommitId = &commitId
	}

	username, ok := request.Params.Arguments["username"].(string)
	if ok && username != "" {
		gitCloneRequest.Username = &username
	}

	password, ok := request.Params.Arguments["password"].(string)
	if ok && password != "" {
		gitCloneRequest.Password = &password
	}

	return &gitCloneRequest, nil
}
