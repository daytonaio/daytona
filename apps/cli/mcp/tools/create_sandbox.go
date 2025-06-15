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

	sandboxId := ""
	if id, ok := request.Params.Arguments["id"]; ok && id != nil {
		if idStr, ok := id.(string); ok && idStr != "" {
			sandboxId = idStr
		}
	}

	if sandboxId != "" {
		sandbox, _, err := apiClient.SandboxAPI.GetSandbox(ctx, sandboxId).Execute()
		if err == nil && sandbox.State != nil && *sandbox.State == daytonaapiclient.SANDBOXSTATE_STARTED {
			return mcp.NewToolResultText(fmt.Sprintf("Reusing existing sandbox %s", sandboxId)), nil
		}

		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("sandbox %s not found or not running", sandboxId)
	}

	createSandbox := daytonaapiclient.NewCreateSandbox()

	if snapshot, ok := request.Params.Arguments["snapshot"]; ok && snapshot != nil {
		if snapshotStr, ok := snapshot.(string); ok && snapshotStr != "" {
			createSandbox.SetSnapshot(snapshotStr)
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
				log.Error(fmt.Errorf("invalid auto stop interval value, fallback to default (15m)"))
				autoStopIntervalValue = 15
			}

			createSandbox.SetAutoStopInterval(int32(autoStopIntervalValue))
		}
	}

	if autoArchiveInterval, ok := request.Params.Arguments["auto_archive_interval"]; ok && autoArchiveInterval != nil {
		if autoArchiveIntervalStr, ok := autoArchiveInterval.(string); ok && autoArchiveIntervalStr != "" {
			autoArchiveIntervalValue, err := strconv.Atoi(autoArchiveIntervalStr)
			if err != nil {
				log.Error(fmt.Errorf("invalid auto archive interval value, fallback to default (7d)"))
				autoArchiveIntervalValue = 7 * 24 * 60
			}

			createSandbox.SetAutoArchiveInterval(int32(autoArchiveIntervalValue))
		}
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
