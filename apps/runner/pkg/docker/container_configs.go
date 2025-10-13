// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/daytonaio/runner/cmd/runner/config"
	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/system"
)

func (d *DockerClient) getContainerConfigs(ctx context.Context, sandboxDto dto.CreateSandboxDTO, volumeMountPathBinds []string) (*container.Config, *container.HostConfig, *network.NetworkingConfig, error) {
	containerConfig, err := d.getContainerCreateConfig(ctx, sandboxDto)
	if err != nil {
		return nil, nil, nil, err
	}

	hostConfig, err := d.getContainerHostConfig(ctx, sandboxDto, volumeMountPathBinds)
	if err != nil {
		return nil, nil, nil, err
	}

	networkingConfig := d.getContainerNetworkingConfig(ctx)
	return containerConfig, hostConfig, networkingConfig, nil
}

func (d *DockerClient) getContainerCreateConfig(ctx context.Context, sandboxDto dto.CreateSandboxDTO) (*container.Config, error) {
	envVars := []string{
		"DAYTONA_SANDBOX_ID=" + sandboxDto.Id,
		"DAYTONA_SANDBOX_SNAPSHOT=" + sandboxDto.Snapshot,
		"DAYTONA_SANDBOX_USER=" + sandboxDto.OsUser,
	}

	for key, value := range sandboxDto.Env {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
	}

	if sandboxDto.OtelEndpoint != nil && *sandboxDto.OtelEndpoint != "" {
		envVars = append(envVars, "DAYTONA_OTEL_ENDPOINT="+*sandboxDto.OtelEndpoint)
	}

	labels := make(map[string]string)
	if sandboxDto.Metadata != nil {
		if orgID, ok := sandboxDto.Metadata["organizationId"]; ok && orgID != "" {
			labels["daytona.organization_id"] = orgID
		}
		if orgName, ok := sandboxDto.Metadata["organizationName"]; ok && orgName != "" {
			labels["daytona.organization_name"] = orgName
		}
	}

	workingDir := ""
	cmd := []string{}
	entrypoint := sandboxDto.Entrypoint
	if !d.useSnapshotEntrypoint {
		// Inspect image
		image, err := d.apiClient.ImageInspect(ctx, sandboxDto.Snapshot)
		if err != nil {
			return nil, err
		}

		if image.Config.WorkingDir != "" {
			workingDir = image.Config.WorkingDir
		}

		// If workingDir is empty, append flag env var to envVars
		if workingDir == "" {
			envVars = append(envVars, "DAYTONA_USER_HOME_AS_WORKDIR=true")
		}

		entrypoint = []string{common.DAEMON_PATH}

		if len(sandboxDto.Entrypoint) != 0 {
			cmd = append(cmd, sandboxDto.Entrypoint...)
		} else {
			if slices.Equal(image.Config.Entrypoint, strslice.StrSlice{common.DAEMON_PATH}) {
				cmd = append(cmd, image.Config.Cmd...)
			} else {
				cmd = append(cmd, image.Config.Entrypoint...)
			}
		}
	}

	return &container.Config{
		Hostname:     sandboxDto.Id,
		Image:        sandboxDto.Snapshot,
		WorkingDir:   workingDir,
		Env:          envVars,
		Entrypoint:   entrypoint,
		Cmd:          cmd,
		Labels:       labels,
		AttachStdout: true,
		AttachStderr: true,
	}, nil
}

func (d *DockerClient) getContainerHostConfig(ctx context.Context, sandboxDto dto.CreateSandboxDTO, volumeMountPathBinds []string) (*container.HostConfig, error) {
	var binds []string

	binds = append(binds, fmt.Sprintf("%s:%s:ro", d.daemonPath, common.DAEMON_PATH))

	// Mount the plugin if available
	if d.computerUsePluginPath != "" {
		binds = append(binds, fmt.Sprintf("%s:/usr/local/lib/daytona-computer-use:ro", d.computerUsePluginPath))
	}

	if len(volumeMountPathBinds) > 0 {
		binds = append(binds, volumeMountPathBinds...)
	}

	hostConfig := &container.HostConfig{
		Privileged: true,
		Binds:      binds,
	}

	if sandboxDto.OtelEndpoint != nil && strings.Contains(*sandboxDto.OtelEndpoint, "host.docker.internal") {
		hostConfig.ExtraHosts = []string{
			"host.docker.internal:host-gateway",
		}
	}

	if !d.resourceLimitsDisabled {
		hostConfig.Resources = container.Resources{
			CPUPeriod:  100000,
			CPUQuota:   sandboxDto.CpuQuota * 100000,
			Memory:     common.GBToBytes(float64(sandboxDto.MemoryQuota)),
			MemorySwap: common.GBToBytes(float64(sandboxDto.MemoryQuota)),
		}
	}

	containerRuntime := config.GetContainerRuntime()
	if containerRuntime != "" {
		hostConfig.Runtime = containerRuntime
	}

	info, err := d.apiClient.Info(ctx)
	if err != nil {
		return nil, err
	}

	filesystem := d.getFilesystem(info)
	if filesystem == "xfs" {
		hostConfig.StorageOpt = map[string]string{
			"size": fmt.Sprintf("%dG", sandboxDto.StorageQuota),
		}
	}

	return hostConfig, nil
}

func (d *DockerClient) getContainerNetworkingConfig(_ context.Context) *network.NetworkingConfig {
	containerNetwork := config.GetContainerNetwork()
	if containerNetwork != "" {
		return &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				containerNetwork: {},
			},
		}
	}
	return nil
}

func (d *DockerClient) getFilesystem(info system.Info) string {
	for _, driver := range info.DriverStatus {
		if driver[0] == "Backing Filesystem" {
			return driver[1]
		}
	}

	return ""
}
