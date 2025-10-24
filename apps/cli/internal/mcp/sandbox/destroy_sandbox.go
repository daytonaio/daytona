// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"fmt"
	"time"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	mcp_headers "github.com/daytonaio/daytona/cli/internal/mcp"
	"github.com/invopop/jsonschema"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	log "github.com/sirupsen/logrus"
)

type DestroySandboxInput struct {
	SandboxId *string `json:"sandboxId,omitempty" jsonschema:"required,description=ID of the sandbox to destroy."`
}

type DestroySandboxOutput struct {
	Message string `json:"message" jsonschema:"description=Message indicating the successful destruction of the sandbox."`
}

func getDestroySandboxTool() *mcp.Tool {
	return &mcp.Tool{
		Name:         "destroy_sandbox",
		Title:        "Destroy Sandbox",
		Description:  "Destroy a sandbox with Daytona.",
		InputSchema:  jsonschema.Reflect(DestroySandboxInput{}),
		OutputSchema: jsonschema.Reflect(DestroySandboxOutput{}),
	}
}

func handleDestroySandbox(ctx context.Context, request *mcp.CallToolRequest, input *DestroySandboxInput) (*mcp.CallToolResult, *DestroySandboxOutput, error) {
	apiClient, err := apiclient_cli.GetApiClient(nil, mcp_headers.DaytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, err
	}

	if input.SandboxId == nil || *input.SandboxId == "" {
		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("sandbox ID is required")
	}

	// Destroy sandbox with retries
	maxRetries := 3
	retryDelay := time.Second * 2

	for retry := range maxRetries {
		_, _, err := apiClient.SandboxAPI.DeleteSandbox(ctx, *input.SandboxId).Execute()
		if err != nil {
			if retry == maxRetries-1 {
				return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("failed to destroy sandbox after %d retries: %v", maxRetries, err)
			}

			log.Infof("Sandbox destruction failed, retrying: %v", err)

			time.Sleep(retryDelay)
			retryDelay = retryDelay * 3 / 2 // Exponential backoff
			continue
		}

		log.Infof("Destroyed sandbox with ID: %s", *input.SandboxId)

		return &mcp.CallToolResult{IsError: false}, &DestroySandboxOutput{
			Message: fmt.Sprintf("Destroyed sandbox with ID %s", *input.SandboxId),
		}, nil
	}

	return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("failed to destroy sandbox after %d retries", maxRetries)
}
