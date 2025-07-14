// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/daytonaio/apiclient"
	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	"github.com/mark3labs/mcp-go/mcp"

	log "github.com/sirupsen/logrus"
)

type CreateSandboxArgs struct {
	Id                  *string                    `json:"id,omitempty"`
	Target              *string                    `json:"target,omitempty"`
	Snapshot            *string                    `json:"snapshot,omitempty"`
	User                *string                    `json:"user,omitempty"`
	Env                 *map[string]string         `json:"env,omitempty"`
	Labels              *map[string]string         `json:"labels,omitempty"`
	Public              *bool                      `json:"public,omitempty"`
	Cpu                 *int32                     `json:"cpu,omitempty"`
	Gpu                 *int32                     `json:"gpu,omitempty"`
	Memory              *int32                     `json:"memory,omitempty"`
	Disk                *int32                     `json:"disk,omitempty"`
	AutoStopInterval    *int32                     `json:"autoStopInterval,omitempty"`
	AutoArchiveInterval *int32                     `json:"autoArchiveInterval,omitempty"`
	AutoDeleteInterval  *int32                     `json:"autoDeleteInterval,omitempty"`
	Volumes             *[]apiclient.SandboxVolume `json:"volumes,omitempty"`
	BuildInfo           *apiclient.CreateBuildInfo `json:"buildInfo,omitempty"`
}

func GetCreateSandboxTool() mcp.Tool {
	return mcp.NewTool("create_sandbox",
		mcp.WithDescription("Create a new sandbox with Daytona"),
		mcp.WithString("id", mcp.Description("If a sandbox ID is provided it is first checked if it exists and is running, if so, the existing sandbox will be used. However, a model is not able to provide custom sandbox ID but only the ones Daytona commands return and should always leave ID field empty if the intention is to create a new sandbox.")),
		mcp.WithString("target", mcp.DefaultString("us"), mcp.Description("Target region of the sandbox.")),
		mcp.WithString("snapshot", mcp.Description("Snapshot of the sandbox (don't specify any if not explicitly instructed from user). Cannot be specified when using a build info entry.")),
		mcp.WithString("user", mcp.Description("User associated with the sandbox.")),
		mcp.WithObject("env", mcp.Description("Environment variables for the sandbox. Format: {\"key\": \"value\", \"key2\": \"value2\"}"), mcp.AdditionalProperties(map[string]any{"type": "string"})),
		mcp.WithObject("labels", mcp.Description("Labels for the sandbox. Format: {\"key\": \"value\", \"key2\": \"value2\"}"), mcp.AdditionalProperties(map[string]any{"type": "string"})),
		mcp.WithBoolean("public", mcp.Description("Whether the sandbox http preview is publicly accessible.")),
		mcp.WithNumber("cpu", mcp.Description("CPU cores allocated to the sandbox. Cannot specify sandbox resources when using a snapshot."), mcp.Max(4)),
		mcp.WithNumber("gpu", mcp.Description("GPU units allocated to the sandbox. Cannot specify sandbox resources when using a snapshot."), mcp.Max(1)),
		mcp.WithNumber("memory", mcp.Description("Memory allocated to the sandbox in GB. Cannot specify sandbox resources when using a snapshot."), mcp.Max(8)),
		mcp.WithNumber("disk", mcp.Description("Disk space allocated to the sandbox in GB. Cannot specify sandbox resources when using a snapshot."), mcp.Max(10)),
		mcp.WithNumber("autoStopInterval", mcp.DefaultNumber(15), mcp.Min(0), mcp.Description("Auto-stop interval in minutes (0 means disabled) for the sandbox.")),
		mcp.WithNumber("autoArchiveInterval", mcp.DefaultNumber(10080), mcp.Min(0), mcp.Description("Auto-archive interval in minutes (0 means the maximum interval will be used) for the sandbox.")),
		mcp.WithNumber("autoDeleteInterval", mcp.DefaultNumber(-1), mcp.Description("Auto-delete interval in minutes (negative value means disabled, 0 means delete immediately upon stopping) for the sandbox.")),
		mcp.WithArray("volumes", mcp.Description("Volumes to attach to the sandbox."), mcp.Items(map[string]any{"type": "object", "properties": map[string]any{"volumeId": map[string]any{"type": "string"}, "mountPath": map[string]any{"type": "string"}}})),
		mcp.WithObject("buildInfo", mcp.Description("Build information for the sandbox."), mcp.Properties(map[string]any{"dockerfileContent": map[string]any{"type": "string"}, "contextHashes": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}})),
	)
}

func CreateSandbox(ctx context.Context, request mcp.CallToolRequest, args CreateSandboxArgs) (*mcp.CallToolResult, error) {
	apiClient, err := apiclient_cli.GetApiClient(nil, daytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, err
	}

	sandboxId := ""
	if args.Id != nil && *args.Id != "" {
		sandboxId = *args.Id
	}

	if sandboxId != "" {
		sandbox, _, err := apiClient.SandboxAPI.GetSandbox(ctx, sandboxId).Execute()
		if err == nil && sandbox.State != nil && *sandbox.State == apiclient.SANDBOXSTATE_STARTED {
			return mcp.NewToolResultText(fmt.Sprintf("Reusing existing sandbox %s", sandboxId)), nil
		}

		return &mcp.CallToolResult{IsError: true}, fmt.Errorf("sandbox %s not found or not running", sandboxId)
	}

	createSandboxReq, err := createSandboxRequest(args)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, err
	}

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

func createSandboxRequest(args CreateSandboxArgs) (*apiclient.CreateSandbox, error) {
	createSandbox := apiclient.NewCreateSandbox()

	if args.BuildInfo != nil {
		if args.Snapshot != nil && *args.Snapshot != "" {
			return nil, fmt.Errorf("cannot specify a snapshot when using a build info entry")
		}
	} else {
		if args.Cpu != nil || args.Gpu != nil || args.Memory != nil || args.Disk != nil {
			return nil, fmt.Errorf("cannot specify sandbox resources when using a snapshot")
		}
	}

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

	if args.AutoDeleteInterval != nil {
		createSandbox.SetAutoDeleteInterval(*args.AutoDeleteInterval)
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

	return createSandbox, nil
}
