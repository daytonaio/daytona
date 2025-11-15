// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"encoding/json"
	"fmt"

	common_daemon "github.com/daytonaio/common-go/pkg/daemon"
	"github.com/daytonaio/runner/cmd/runner/config"
	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/docker/docker/api/types/network"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/system"
)

func (d *DockerClient) getContainerConfigs(ctx context.Context, sandboxDto dto.CreateSandboxDTO, volumeMountPathBinds []string) (*container.Config, *container.HostConfig, *network.NetworkingConfig, error) {
	containerConfig := d.getContainerCreateConfig(sandboxDto)

	hostConfig, err := d.getContainerHostConfig(ctx, sandboxDto, volumeMountPathBinds)
	if err != nil {
		return nil, nil, nil, err
	}

	networkingConfig := d.getContainerNetworkingConfig(ctx)
	return containerConfig, hostConfig, networkingConfig, nil
}

func (d *DockerClient) getContainerCreateConfig(sandboxDto dto.CreateSandboxDTO) *container.Config {
	envVars := []string{
		"DAYTONA_SANDBOX_ID=" + sandboxDto.Id,
		"DAYTONA_SANDBOX_SNAPSHOT=" + sandboxDto.Snapshot,
		"DAYTONA_SANDBOX_USER=" + sandboxDto.OsUser,
	}

	for key, value := range sandboxDto.Env {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
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

	entrypoint := sandboxDto.Entrypoint
	if d.daemonEntrypoint {
		// Store the snapshot entrypoint as JSON in environment variable
		// This allows the daemon wrapper to execute it
		if len(sandboxDto.Entrypoint) > 0 {
			entrypointJSON, err := json.Marshal(sandboxDto.Entrypoint)
			if err == nil {
				envVars = append(envVars, "DAYTONA_SNAPSHOT_ENTRYPOINT="+string(entrypointJSON))
			}
		}

		// Get workdir from metadata or use default
		workDir := ""
		if sandboxDto.Metadata != nil {
			if wd, ok := sandboxDto.Metadata["workDir"]; ok {
				workDir = wd
			}
		}
		if workDir == "" {
			workDir = common_daemon.UseUserHomeAsWorkDir
		}

		// Set daemon as entrypoint with wrapper script
		entrypoint = d.getDaemonWrapperEntrypoint(workDir)
	}

	return &container.Config{
		Hostname: sandboxDto.Id,
		Image:    sandboxDto.Snapshot,
		// User:         sandboxDto.OsUser,
		Env:          envVars,
		Entrypoint:   entrypoint,
		Labels:       labels,
		AttachStdout: true,
		AttachStderr: true,
	}
}

func (d *DockerClient) getContainerHostConfig(ctx context.Context, sandboxDto dto.CreateSandboxDTO, volumeMountPathBinds []string) (*container.HostConfig, error) {
	var binds []string

	binds = append(binds, fmt.Sprintf("%s:/usr/local/bin/daytona:ro", d.daemonPath))

	// Mount the plugin if available
	if d.computerUsePluginPath != "" {
		binds = append(binds, fmt.Sprintf("%s:/usr/local/lib/daytona-computer-use:ro", d.computerUsePluginPath))
	}

	if len(volumeMountPathBinds) > 0 {
		binds = append(binds, volumeMountPathBinds...)
	}

	hostConfig := &container.HostConfig{
		Privileged: true,
		ExtraHosts: []string{"host.docker.internal:host-gateway"},
		Binds:      binds,
	}

	if !d.resourceLimitsDisabled {
		hostConfig.Resources = container.Resources{
			CPUPeriod:  100000,
			CPUQuota:   sandboxDto.CpuQuota * 100000,
			Memory:     sandboxDto.MemoryQuota * 1024 * 1024 * 1024,
			MemorySwap: sandboxDto.MemoryQuota * 1024 * 1024 * 1024,
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
