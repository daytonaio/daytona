// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"

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

	return &container.Config{
		Hostname: sandboxDto.Id,
		Image:    sandboxDto.Snapshot,
		// User:         sandboxDto.OsUser,
		Env:          envVars,
		Entrypoint:   sandboxDto.Entrypoint,
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
