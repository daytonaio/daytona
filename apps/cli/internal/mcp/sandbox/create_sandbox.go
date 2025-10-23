// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"fmt"
	"strings"
	"time"

	apiclient "github.com/daytonaio/apiclient"
	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	mcp_headers "github.com/daytonaio/daytona/cli/internal/mcp"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	log "github.com/sirupsen/logrus"
)

type CreateSandboxInput struct {
	Id                  *string                    `json:"id,omitempty" jsonchema:"ID of the sandbox to create."`
	Name                *string                    `json:"name,omitempty" jsonchema:"Name of the sandbox. If not provided, the sandbox ID will be used as the name."`
	Target              *string                    `json:"target,omitempty" jsonchema:"Target region of the sandbox."`
	Snapshot            *string                    `json:"snapshot,omitempty" jsonchema:"Snapshot of the sandbox (don't specify any if not explicitly instructed from user). Cannot be specified when using a build info entry."`
	User                *string                    `json:"user,omitempty" jsonchema:"User associated with the sandbox."`
	Env                 *map[string]string         `json:"env,omitempty" jsonchema:"Environment variables for the sandbox. Format: {\"key\": \"value\", \"key2\": \"value2\"}"`
	Labels              *map[string]string         `json:"labels,omitempty" jsonchema:"Labels for the sandbox. Format: {\"key\": \"value\", \"key2\": \"value2\"}"`
	Public              *bool                      `json:"public,omitempty" jsonchema:"Whether the sandbox http preview is publicly accessible."`
	Cpu                 *int32                     `json:"cpu,omitempty" jsonchema:"CPU cores allocated to the sandbox. Cannot specify sandbox resources when using a snapshot."`
	Gpu                 *int32                     `json:"gpu,omitempty" jsonchema:"GPU units allocated to the sandbox. Cannot specify sandbox resources when using a snapshot."`
	Memory              *int32                     `json:"memory,omitempty" jsonchema:"Memory allocated to the sandbox in GB. Cannot specify sandbox resources when using a snapshot."`
	Disk                *int32                     `json:"disk,omitempty" jsonchema:"Disk space allocated to the sandbox in GB. Cannot specify sandbox resources when using a snapshot."`
	AutoStopInterval    *int32                     `json:"autoStopInterval,omitempty" jsonchema:"Auto-stop interval in minutes (0 means disabled) for the sandbox."`
	AutoArchiveInterval *int32                     `json:"autoArchiveInterval,omitempty" jsonchema:"Auto-archive interval in minutes (0 means the maximum interval will be used) for the sandbox."`
	AutoDeleteInterval  *int32                     `json:"autoDeleteInterval,omitempty" jsonchema:"Auto-delete interval in minutes (negative value means disabled, 0 means delete immediately upon stopping) for the sandbox."`
	Volumes             *[]apiclient.SandboxVolume `json:"volumes,omitempty" jsonchema:"Volumes to attach to the sandbox."`
	BuildInfo           *apiclient.CreateBuildInfo `json:"buildInfo,omitempty" jsonchema:"Build information for the sandbox."`
	NetworkBlockAll     *bool                      `json:"networkBlockAll,omitempty" jsonchema:"Whether to block all network access to the sandbox."`
	NetworkAllowList    *string                    `json:"networkAllowList,omitempty" jsonchema:"Comma-separated list of domains to allow network access to the sandbox."`
}

type CreateSandboxOutput struct {
	Message string `json:"message" jsonchema:"Message indicating the successful creation of the sandbox."`
}

func getCreateSandboxTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "create_sandbox",
		Title:       "Create Sandbox",
		Description: "Create a new sandbox with Daytona.",
	}
}

func handleCreateSandbox(ctx context.Context, request *mcp.CallToolRequest, input *CreateSandboxInput) (*mcp.CallToolResult, *CreateSandboxOutput, error) {
	apiClient, err := apiclient_cli.GetApiClient(nil, mcp_headers.DaytonaMCPHeaders)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, err
	}

	sandboxId := ""
	if input.Id != nil && *input.Id != "" {
		sandboxId = *input.Id
	}

	if sandboxId != "" {
		sandbox, _, err := apiClient.SandboxAPI.GetSandbox(ctx, sandboxId).Execute()
		if err == nil && sandbox.State != nil && *sandbox.State == apiclient.SANDBOXSTATE_STARTED {
			return &mcp.CallToolResult{IsError: false}, &CreateSandboxOutput{
				Message: fmt.Sprintf("Reusing existing sandbox %s", sandboxId),
			}, nil
		}

		return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("sandbox %s not found or not running", sandboxId)
	}

	createSandboxReq, err := createSandboxRequest(*input)
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, nil, err
	}

	// Create new sandbox with retries
	maxRetries := 3
	retryDelay := time.Second * 2

	for retry := range maxRetries {
		sandbox, _, err := apiClient.SandboxAPI.CreateSandbox(ctx).CreateSandbox(*createSandboxReq).Execute()
		if err != nil {
			if strings.Contains(err.Error(), "Total CPU quota exceeded") {
				return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("CPU quota exceeded. Please delete unused sandboxes or upgrade your plan")
			}

			if retry == maxRetries-1 {
				return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("failed to create sandbox after %d retries: %v", maxRetries, err)
			}

			log.Infof("Sandbox creation failed, retrying: %v", err)

			time.Sleep(retryDelay)
			retryDelay = retryDelay * 3 / 2 // Exponential backoff
			continue
		}

		log.Infof("Created new sandbox: %s", sandbox.Id)

		return &mcp.CallToolResult{IsError: false}, &CreateSandboxOutput{
			Message: fmt.Sprintf("Created new sandbox %s", sandbox.Id),
		}, nil
	}

	return &mcp.CallToolResult{IsError: true}, nil, fmt.Errorf("failed to create sandbox after %d retries", maxRetries)
}

func createSandboxRequest(input CreateSandboxInput) (*apiclient.CreateSandbox, error) {
	createSandbox := apiclient.NewCreateSandbox()

	if input.Name != nil && *input.Name != "" {
		createSandbox.SetName(*input.Name)
	}

	if input.BuildInfo != nil {
		if input.Snapshot != nil && *input.Snapshot != "" {
			return nil, fmt.Errorf("cannot specify a snapshot when using a build info entry")
		}
	} else {
		if input.Cpu != nil || input.Gpu != nil || input.Memory != nil || input.Disk != nil {
			return nil, fmt.Errorf("cannot specify sandbox resources when using a snapshot")
		}
	}

	if input.Snapshot != nil && *input.Snapshot != "" {
		createSandbox.SetSnapshot(*input.Snapshot)
	}

	if input.Target != nil && *input.Target != "" {
		createSandbox.SetTarget(*input.Target)
	}

	if input.AutoStopInterval != nil {
		createSandbox.SetAutoStopInterval(*input.AutoStopInterval)
	}

	if input.AutoArchiveInterval != nil {
		createSandbox.SetAutoArchiveInterval(*input.AutoArchiveInterval)
	}

	if input.AutoDeleteInterval != nil {
		createSandbox.SetAutoDeleteInterval(*input.AutoDeleteInterval)
	}

	if input.User != nil && *input.User != "" {
		createSandbox.SetUser(*input.User)
	}

	if input.Env != nil {
		createSandbox.SetEnv(*input.Env)
	}

	if input.Labels != nil {
		createSandbox.SetLabels(*input.Labels)
	}

	if input.Public != nil {
		createSandbox.SetPublic(*input.Public)
	}

	if input.Cpu != nil {
		createSandbox.SetCpu(*input.Cpu)
	}

	if input.Memory != nil {
		createSandbox.SetMemory(*input.Memory)
	}

	if input.Disk != nil {
		createSandbox.SetDisk(*input.Disk)
	}

	if input.Volumes != nil {
		createSandbox.SetVolumes(*input.Volumes)
	}

	if input.BuildInfo != nil {
		createSandbox.SetBuildInfo(*input.BuildInfo)
	}

	if input.NetworkBlockAll != nil {
		createSandbox.SetNetworkBlockAll(*input.NetworkBlockAll)
	}

	if input.NetworkAllowList != nil {
		createSandbox.SetNetworkAllowList(*input.NetworkAllowList)
	}

	return createSandbox, nil
}
