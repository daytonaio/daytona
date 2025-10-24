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
	"github.com/invopop/jsonschema"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	log "github.com/sirupsen/logrus"
)

type CreateSandboxInput struct {
	Name                *string                    `json:"name,omitempty" jsonschema:"type=string,description=Name of the sandbox. Don't provide this if not explicitly instructed from user. If not provided, the sandbox ID will be used as the name."`
	Target              *string                    `json:"target,omitempty" jsonschema:"default=us,type=string,description=Target region of the sandbox (defaults to us)."`
	Snapshot            *string                    `json:"snapshot,omitempty" jsonschema:"type=string,description=Snapshot of the sandbox (don't specify any if not explicitly instructed from user). Cannot be specified when using a build info entry."`
	User                *string                    `json:"user,omitempty" jsonschema:"type=string,description=User associated with the sandbox."`
	Env                 *map[string]string         `json:"env,omitempty" jsonschema:"type=object,additionalProperties=string,description=Environment variables for the sandbox. Format: {\"key\": \"value\", \"key2\": \"value2\"}"`
	Labels              *map[string]string         `json:"labels,omitempty" jsonschema:"type=object,additionalProperties=string,description=Labels for the sandbox. Format: {\"key\": \"value\", \"key2\": \"value2\"}"`
	Public              *bool                      `json:"public,omitempty" jsonschema:"default=true,type=boolean,description=Whether the sandbox http preview is publicly accessible (defaults to true)."`
	Cpu                 *int32                     `json:"cpu,omitempty" jsonschema:"default=1,type=integer,description=CPU cores allocated to the sandbox. Cannot specify sandbox resources when using a snapshot."`
	Gpu                 *int32                     `json:"gpu,omitempty" jsonschema:"default=1,type=integer,description=GPU units allocated to the sandbox. Cannot specify sandbox resources when using a snapshot."`
	Memory              *int32                     `json:"memory,omitempty" jsonschema:"default=2,type=integer,description=Memory allocated to the sandbox in GB. Cannot specify sandbox resources when using a snapshot."`
	Disk                *int32                     `json:"disk,omitempty" jsonschema:"default=4,type=integer,description=Disk space allocated to the sandbox in GB. Cannot specify sandbox resources when using a snapshot."`
	AutoStopInterval    *int32                     `json:"autoStopInterval,omitempty" jsonschema:"default=15,type=integer,description=Auto-stop interval in minutes (0 means disabled) for the sandbox (defaults to 15)."`
	AutoArchiveInterval *int32                     `json:"autoArchiveInterval,omitempty" jsonschema:"default=10080,type=integer,description=Auto-archive interval in minutes (0 means the maximum interval will be used) for the sandbox (defaults to 10080)."`
	AutoDeleteInterval  *int32                     `json:"autoDeleteInterval,omitempty" jsonschema:"default=-1,type=integer,description=Auto-delete interval in minutes (negative value means disabled, 0 means delete immediately upon stopping) for the sandbox (defaults to -1)."`
	Volumes             *[]apiclient.SandboxVolume `json:"volumes,omitempty" jsonschema:"type=array,items=object,description=Volumes to attach to the sandbox."`
	BuildInfo           *apiclient.CreateBuildInfo `json:"buildInfo,omitempty" jsonschema:"type=object,description=Build information for the sandbox."`
	NetworkBlockAll     *bool                      `json:"networkBlockAll,omitempty" jsonschema:"default=false,type=boolean,description=Whether to block all network access to the sandbox (defaults to false)."`
	NetworkAllowList    *string                    `json:"networkAllowList,omitempty" jsonschema:"type=string,description=Comma-separated list of domains to allow network access to the sandbox."`
}

type CreateSandboxOutput struct {
	Message string `json:"message" jsonschema:"type=string,description=Message indicating the successful creation of the sandbox."`
}

func getCreateSandboxTool() *mcp.Tool {
	return &mcp.Tool{
		Name:         "create_sandbox",
		Title:        "Create Sandbox",
		Description:  "Create a new sandbox with Daytona.",
		InputSchema:  jsonschema.Reflect(CreateSandboxInput{}),
		OutputSchema: jsonschema.Reflect(CreateSandboxOutput{}),
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
