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
	Id                  *string                           `json:"id,omitempty"`
	Target              *string                           `json:"target,omitempty"`
	Snapshot            *string                           `json:"snapshot,omitempty"`
	User                *string                           `json:"user,omitempty"`
	Env                 *map[string]string                `json:"env,omitempty"`
	Labels              *map[string]string                `json:"labels,omitempty"`
	Public              *bool                             `json:"public,omitempty"`
	Class               *string                           `json:"class,omitempty"`
	Cpu                 *int32                            `json:"cpu,omitempty"`
	Memory              *int32                            `json:"memory,omitempty"`
	Disk                *int32                            `json:"disk,omitempty"`
	AutoStopInterval    *int32                            `json:"autoStopInterval,omitempty"`
	AutoArchiveInterval *int32                            `json:"autoArchiveInterval,omitempty"`
	Volumes             *[]daytonaapiclient.SandboxVolume `json:"volumes,omitempty"`
	BuildInfo           *daytonaapiclient.CreateBuildInfo `json:"buildInfo,omitempty"`
}

func GetCreateSandboxTool() mcp.Tool {
	return mcp.NewTool("create_sandbox",
		mcp.WithDescription("Create a new sandbox with Daytona"),
		mcp.WithString("id", mcp.Description("If a sandbox ID is provided it is first checked if it exists and is running, if so, the existing sandbox will be used. However, a model is not able to provide custom sandbox ID but only the ones Daytona commands return and should always leave ID field empty if the intention is to create a new sandbox.")),
		mcp.WithString("target", mcp.DefaultString("us"), mcp.Description("Target region of the sandbox.")),
		mcp.WithString("snapshot", mcp.Description("Snapshot of the sandbox (don't specify any if not explicitly instructed from user).")),
		mcp.WithString("user", mcp.Description("User associated with the sandbox.")),
		mcp.WithObject("env", mcp.Description("Environment variables for the sandbox. Format: {\"key\": \"value\", \"key2\": \"value2\"}"), mcp.AdditionalProperties(map[string]any{"type": "string"})),
		mcp.WithObject("labels", mcp.Description("Labels for the sandbox. Format: {\"key\": \"value\", \"key2\": \"value2\"}"), mcp.AdditionalProperties(map[string]any{"type": "string"})),
		mcp.WithBoolean("public", mcp.Description("Whether the sandbox http preview is publicly accessible.")),
		mcp.WithString("class", mcp.Description("Class type of the sandbox.")),
		mcp.WithNumber("cpu", mcp.Description("CPU cores allocated to the sandbox."), mcp.Min(0), mcp.Max(4)),
		mcp.WithNumber("memory", mcp.Description("Memory allocated to the sandbox in GB."), mcp.Min(0), mcp.Max(8)),
		mcp.WithNumber("disk", mcp.Description("Disk space allocated to the sandbox in GB."), mcp.Min(0), mcp.Max(10)),
		mcp.WithNumber("autoStopInterval", mcp.DefaultNumber(15), mcp.Min(0), mcp.Description("Auto-stop interval in minutes (0 means disabled) for the sandbox.")),
		mcp.WithNumber("autoArchiveInterval", mcp.DefaultNumber(10080), mcp.Min(0), mcp.Description("Auto-archive interval in minutes (0 means the maximum interval will be used) for the sandbox.")),
		mcp.WithArray("volumes", mcp.Description("Volumes to attach to the sandbox."), mcp.Items(map[string]any{"type": "object", "properties": map[string]any{"volumeId": map[string]any{"type": "string"}, "mountPath": map[string]any{"type": "string"}}})),
		mcp.WithObject("buildInfo", mcp.Description("Build information for the sandbox."), mcp.Properties(map[string]any{"dockerfile_content": map[string]any{"type": "string"}, "context_hashes": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}})),
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

	createSandboxReq := createSandboxRequest(args)

	// Create new sandbox with retries
	maxRetries := 3
	retryDelay := time.Second * 2

	for retry := range maxRetries {
		sandbox, _, err := apiClient.SandboxAPI.CreateSandbox(ctx).CreateSandbox(*createSandboxReq).Execute()
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

func createSandboxRequest(args CreateSandboxArgs) *daytonaapiclient.CreateSandbox {
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

	if args.User != nil && *args.User != "" {
		createSandbox.SetUser(*args.User)
	}

	if args.Env != nil {
		createSandbox.SetEnv(*args.Env)
	}

	if args.Labels != nil {
		createSandbox.SetLabels(*args.Labels)
	}

	if args.Public != nil {
		createSandbox.SetPublic(*args.Public)
	}

	if args.Class != nil && *args.Class != "" {
		createSandbox.SetClass(*args.Class)
	}

	if args.Cpu != nil {
		createSandbox.SetCpu(*args.Cpu)
	}

	if args.Memory != nil {
		createSandbox.SetMemory(*args.Memory)
	}

	if args.Disk != nil {
		createSandbox.SetDisk(*args.Disk)
	}

	if args.Volumes != nil {
		createSandbox.SetVolumes(*args.Volumes)
	}

	if args.BuildInfo != nil {
		createSandbox.SetBuildInfo(*args.BuildInfo)
	}

	return createSandbox
}
