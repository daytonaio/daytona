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
	Name                string                     `json:"name,omitempty" jsonschema:"Name of the sandbox. Don't provide this if not explicitly instructed from user. If not provided, the sandbox ID will be used as the name."`
	Target              string                     `json:"target,omitempty" jsonschema:"Target region of the sandbox (defaults to us)."`
	Snapshot            string                     `json:"snapshot,omitempty" jsonschema:"Snapshot of the sandbox (don't specify any if not explicitly instructed from user). Cannot be specified when using a build info entry."`
	User                string                     `json:"user,omitempty" jsonschema:"User associated with the sandbox."`
	Env                 map[string]string          `json:"env,omitempty" jsonschema:"Environment variables for the sandbox. Format: {\"key\": \"value\", \"key2\": \"value2\"}"`
	Labels              map[string]string          `json:"labels,omitempty" jsonschema:"Labels for the sandbox. Format: {\"key\": \"value\", \"key2\": \"value2\"}"`
	Public              *bool                      `json:"public,omitempty" jsonschema:"Whether the sandbox http preview is publicly accessible (defaults to true)."`
	Cpu                 *int32                     `json:"cpu,omitempty" jsonschema:"CPU cores allocated to the sandbox. Cannot specify sandbox resources when using a snapshot."`
	Gpu                 *int32                     `json:"gpu,omitempty" jsonschema:"GPU units allocated to the sandbox. Cannot specify sandbox resources when using a snapshot."`
	Memory              *int32                     `json:"memory,omitempty" jsonschema:"Memory allocated to the sandbox in GB. Cannot specify sandbox resources when using a snapshot."`
	Disk                *int32                     `json:"disk,omitempty" jsonschema:"Disk space allocated to the sandbox in GB. Cannot specify sandbox resources when using a snapshot."`
	AutoStopInterval    *int32                     `json:"autoStopInterval,omitempty" jsonschema:"Auto-stop interval in minutes (0 means disabled) for the sandbox (defaults to 15)."`
	AutoArchiveInterval *int32                     `json:"autoArchiveInterval,omitempty" jsonschema:"Auto-archive interval in minutes (0 means the maximum interval will be used) for the sandbox (defaults to 10080)."`
	AutoDeleteInterval  *int32                     `json:"autoDeleteInterval,omitempty" jsonschema:"Auto-delete interval in minutes (negative value means disabled, 0 means delete immediately upon stopping) for the sandbox (defaults to -1)."`
	Volumes             []apiclient.SandboxVolume  `json:"volumes,omitempty" jsonschema:"Volumes to attach to the sandbox."`
	BuildInfo           *apiclient.CreateBuildInfo `json:"buildInfo,omitempty" jsonschema:"Build information for the sandbox."`
	NetworkBlockAll     *bool                      `json:"networkBlockAll,omitempty" jsonschema:"Whether to block all network access to the sandbox (defaults to false)."`
	NetworkAllowList    *string                    `json:"networkAllowList,omitempty" jsonschema:"Comma-separated list of domains to allow network access to the sandbox."`
}

type CreateSandboxOutput struct {
	Message string `json:"message,omitempty" jsonschema:"description=Message indicating the successful creation of the sandbox."`
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

	if input.Name != "" {
		createSandbox.SetName(input.Name)
	}

	if input.BuildInfo != nil {
		if input.Snapshot != "" {
			return nil, fmt.Errorf("cannot specify a snapshot when using a build info entry")
		}
	} else {
		if input.Cpu != nil || input.Gpu != nil || input.Memory != nil || input.Disk != nil {
			return nil, fmt.Errorf("cannot specify sandbox resources when using a snapshot")
		}
	}

	if input.Snapshot != "" {
		createSandbox.SetSnapshot(input.Snapshot)
	}

	if input.Target != "" {
		createSandbox.SetTarget(input.Target)
	} else {
		createSandbox.SetTarget("us")
	}

	if input.AutoStopInterval != nil {
		createSandbox.SetAutoStopInterval(*input.AutoStopInterval)
	} else {
		createSandbox.SetAutoStopInterval(15)
	}

	if input.AutoArchiveInterval != nil {
		createSandbox.SetAutoArchiveInterval(*input.AutoArchiveInterval)
	} else {
		createSandbox.SetAutoArchiveInterval(10080)
	}

	if input.AutoDeleteInterval != nil {
		createSandbox.SetAutoDeleteInterval(*input.AutoDeleteInterval)
	} else {
		createSandbox.SetAutoDeleteInterval(-1)
	}

	if input.User != "" {
		createSandbox.SetUser(input.User)
	}

	if input.Env != nil {
		createSandbox.SetEnv(input.Env)
	}

	if input.Labels != nil {
		createSandbox.SetLabels(input.Labels)
	}

	if input.Public != nil {
		createSandbox.SetPublic(*input.Public)
	} else {
		createSandbox.SetPublic(true)
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
		createSandbox.SetVolumes(input.Volumes)
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
