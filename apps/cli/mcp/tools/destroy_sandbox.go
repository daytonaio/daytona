// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/daytonaio/daytona/cli/apiclient"
	"github.com/mark3labs/mcp-go/mcp"

	log "github.com/sirupsen/logrus"
)

func DestroySandbox(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiClient, err := apiclient.GetApiClient(nil, daytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, err
	}

	// Check for existing sandbox from tracking file
	sandboxId := ""
	if id, ok := request.Params.Arguments["id"]; ok && id != nil {
		if idStr, ok := id.(string); ok && idStr != "" {
			sandboxId = idStr
		}
	}

	if sandboxId == "" {
		return mcp.NewToolResultText("No latest sandbox was found to destroy."), nil
	}

	// Destroy sandbox with retries
	maxRetries := 3
	retryDelay := time.Second * 2

	for retry := range maxRetries {
		_, err := apiClient.WorkspaceAPI.DeleteWorkspace(ctx, sandboxId).Force(true).Execute()
		if err != nil {
			if retry == maxRetries-1 {
				return &mcp.CallToolResult{IsError: true}, fmt.Errorf("failed to destroy sandbox after %d retries: %v", maxRetries, err)
			}

			log.Infof("Sandbox creation failed, retrying: %v", err)

			time.Sleep(retryDelay)
			retryDelay = retryDelay * 3 / 2 // Exponential backoff
			continue
		}

		log.Infof("Destroyed sandbox with ID: %s", sandboxId)

		return mcp.NewToolResultText(fmt.Sprintf("Destroyed sandbox with ID %s", sandboxId)), nil
	}

	return &mcp.CallToolResult{IsError: true}, fmt.Errorf("failed to destroy sandbox after %d retries", maxRetries)
}
