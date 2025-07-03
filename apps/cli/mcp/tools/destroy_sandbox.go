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

type DestroySandboxArgs struct {
	Id *string `json:"id,omitempty"`
}

func GetDestroySandboxTool() mcp.Tool {
	return mcp.NewTool("destroy_sandbox",
		mcp.WithDescription("Destroy a sandbox with Daytona"),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID of the sandbox to destroy.")),
	)
}

func DestroySandbox(ctx context.Context, request mcp.CallToolRequest, args DestroySandboxArgs) (*mcp.CallToolResult, error) {
	apiClient, err := apiclient.GetApiClient(nil, daytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, err
	}

	if args.Id == nil || *args.Id == "" {
		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("sandbox ID is required")
	}

	// Destroy sandbox with retries
	maxRetries := 3
	retryDelay := time.Second * 2

	for retry := range maxRetries {
		_, err := apiClient.SandboxAPI.DeleteSandbox(ctx, *args.Id).Force(true).Execute()
		if err != nil {
			if retry == maxRetries-1 {
				return &mcp.CallToolResult{IsError: true}, fmt.Errorf("failed to destroy sandbox after %d retries: %v", maxRetries, err)
			}

			log.Infof("Sandbox creation failed, retrying: %v", err)

			time.Sleep(retryDelay)
			retryDelay = retryDelay * 3 / 2 // Exponential backoff
			continue
		}

		log.Infof("Destroyed sandbox with ID: %s", *args.Id)

		return mcp.NewToolResultText(fmt.Sprintf("Destroyed sandbox with ID %s", *args.Id)), nil
	}

	return &mcp.CallToolResult{IsError: true}, fmt.Errorf("failed to destroy sandbox after %d retries", maxRetries)
}
