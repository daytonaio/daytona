// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package tools

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/daytonaio/daytona/cli/apiclient"
	daytonaapiclient "github.com/daytonaio/daytona/daytonaapiclient"
	"github.com/mark3labs/mcp-go/mcp"

	log "github.com/sirupsen/logrus"
)

func CreateSandbox(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiClient, err := apiclient.GetApiClient(nil, daytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, err
	}

	// Check for existing sandbox from tracking file
	sandboxID, err := getActiveSandbox()
	if err == nil && sandboxID != "" {
		// Try to get existing sandbox
		sandbox, _, err := apiClient.WorkspaceAPI.GetWorkspace(ctx, sandboxID).Execute()
		if err == nil && sandbox != nil {
			if sandbox.State != nil && *sandbox.State == daytonaapiclient.WORKSPACESTATE_STARTED {
				log.Infof("Reusing existing sandbox: %s", sandbox.Id)

				return mcp.NewToolResultText(fmt.Sprintf("Reusing existing sandbox %s", sandboxID)), nil
			}
		}

		// Sandbox not found or error, clear tracking
		_ = clearActiveSandbox()
	}

	createSandbox := daytonaapiclient.NewCreateWorkspace()

	if image, ok := request.Params.Arguments["image"]; ok && image != nil {
		if imageStr, ok := image.(string); ok && imageStr != "" {
			createSandbox.SetImage(imageStr)
		}
	}

	if target, ok := request.Params.Arguments["target"]; ok && target != nil {
		if targetStr, ok := target.(string); ok && targetStr != "" {
			createSandbox.SetTarget(targetStr)
		}
	}

	if autoStopInterval, ok := request.Params.Arguments["auto_stop_interval"]; ok && autoStopInterval != nil {
		if autoStopIntervalStr, ok := autoStopInterval.(string); ok && autoStopIntervalStr != "" {
			autoStopIntervalValue, err := strconv.Atoi(autoStopIntervalStr)
			if err != nil {
				log.Error(fmt.Errorf("invalid auto stop interval value, fallack to default (15m)"))
				autoStopIntervalValue = 15
			}

			createSandbox.SetAutoStopInterval(int32(autoStopIntervalValue))
		}
	}

	// Create new sandbox with retries
	maxRetries := 3
	retryDelay := time.Second * 2

	for retry := range maxRetries {
		sandbox, _, err := apiClient.WorkspaceAPI.CreateWorkspace(ctx).CreateWorkspace(*createSandbox).Execute()
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

		// Save sandbox ID to tracking file
		err = setActiveSandbox(sandbox.Id)
		if err != nil {
			log.Infof("Failed to save sandbox ID: %s; %v", sandbox.Id, err)
		}

		log.Infof("Created new sandbox: %s", sandbox.Id)

		return mcp.NewToolResultText(fmt.Sprintf("Created new sandbox %s", sandbox.Id)), nil
	}

	return &mcp.CallToolResult{IsError: true}, fmt.Errorf("failed to create sandbox after %d retries", maxRetries)
}
