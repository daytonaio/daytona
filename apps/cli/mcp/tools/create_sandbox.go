// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/daytonaio/daytona/cli/apiclient"
	daytonaapiclient "github.com/daytonaio/daytona/daytonaapiclient"
	"github.com/mark3labs/mcp-go/mcp"

	log "github.com/sirupsen/logrus"
)

type CreateSandboxArgs struct {
	Id                  *string `json:"id,omitempty"`
	Target              *string `json:"target,omitempty"`
	Snapshot            *string `json:"snapshot,omitempty"`
	AutoStopInterval    *int32  `json:"autoStopInterval,omitempty"`
	AutoArchiveInterval *int32  `json:"autoArchiveInterval,omitempty"`
}

func GetCreateSandboxTool() mcp.Tool {
	return mcp.NewTool("create_sandbox",
		mcp.WithDescription("Create a new sandbox with Daytona"),
		mcp.WithString("id", mcp.Description("If a sandbox ID is provided it is first checked if it exists and is running, if so, the existing sandbox will be used. However, a model is not able to provide custom sandbox ID but only the ones Daytona commands return and should always leave ID field empty if the intention is to create a new sandbox.")),
		mcp.WithString("target", mcp.DefaultString("us"), mcp.Description("Target region of the sandbox.")),
		mcp.WithString("snapshot", mcp.Description("Snapshot of the sandbox (don't specify any if not explicitly instructed from user).")),
		mcp.WithNumber("auto_stop_interval", mcp.DefaultNumber(15), mcp.Min(0), mcp.Description("Auto-stop interval in minutes (0 means disabled) for the sandbox.")),
		mcp.WithNumber("auto_archive_interval", mcp.DefaultNumber(10080), mcp.Min(0), mcp.Description("Auto-archive interval in minutes (0 means the maximum interval will be used) for the sandbox.")),
	)
}

func CreateSandbox(ctx context.Context, request mcp.CallToolRequest, args CreateSandboxArgs) (*mcp.CallToolResult, error) {
	apiClient, err := apiclient.GetApiClient(nil, daytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, err
	}

	sandboxId := ""
	if args.Id != nil && *args.Id != "" {
		sandboxId = *args.Id
	}

	if sandboxId != "" {
		sandbox, _, err := apiClient.SandboxAPI.GetSandbox(ctx, sandboxId).Execute()
		if err == nil && sandbox.State != nil && *sandbox.State == daytonaapiclient.SANDBOXSTATE_STARTED {
			return mcp.NewToolResultText(fmt.Sprintf("Reusing existing sandbox %s", sandboxId)), nil
		}

		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("sandbox %s not found or not running", sandboxId)
	}

	createSandbox := daytonaapiclient.NewCreateSandbox()

	if args.Snapshot != nil && *args.Snapshot != "" {
		createSandbox.SetSnapshot(*args.Snapshot)
	}

	if args.Target != nil && *args.Target != "" {
		createSandbox.SetTarget(*args.Target)
	}

	if args.AutoStopInterval != nil {
		createSandbox.SetAutoStopInterval(*args.AutoStopInterval)
	}

	if args.AutoArchiveInterval != nil {
		createSandbox.SetAutoArchiveInterval(*args.AutoArchiveInterval)
	}

	// Create new sandbox with retries
	maxRetries := 3
	retryDelay := time.Second * 2

	for retry := range maxRetries {
		sandbox, _, err := apiClient.SandboxAPI.CreateSandbox(ctx).CreateSandbox(*createSandbox).Execute()
		if err != nil {
			if strings.Contains(err.Error(), "Total CPU quota exceeded") {
				return &mcp.CallToolResult{IsError: true}, fmt.Errorf("CPU quota exceeded. Please delete unused sandboxes or upgrade your plan")
			}

			if retry == maxRetries-1 {
				return &mcp.CallToolResult{IsError: true}, fmt.Errorf("failed to create sandbox after %d retries: %v", maxRetries, err)
			}

			log.Infof("Sandbox creation failed, retrying: %v", err)

			time.Sleep(retryDelay)
			retryDelay = retryDelay * 3 / 2 // Exponential backoff
			continue
		}

		log.Infof("Created new sandbox: %s", sandbox.Id)

		return mcp.NewToolResultText(fmt.Sprintf("Created new sandbox %s", sandbox.Id)), nil
	}

	return &mcp.CallToolResult{IsError: true}, fmt.Errorf("failed to create sandbox after %d retries", maxRetries)
}
