// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"errors"
	"fmt"

	"github.com/daytonaio/runner/cmd/runner/config"
	"github.com/daytonaio/runner/pkg/api/dto"

	"github.com/docker/docker/api/types/container"
)

func (d *DockerClient) getContainerConfigs(ctx context.Context, sandboxDto dto.CreateSandboxDTO, volumeMountPathBinds []string) (*container.Config, *container.HostConfig, error) {
	containerConfig := d.getContainerCreateConfig(sandboxDto)

	hostConfig, err := d.getContainerHostConfig(ctx, sandboxDto, volumeMountPathBinds)
	if err != nil {
		return nil, nil, err
	}

	return containerConfig, hostConfig, nil
}

func (d *DockerClient) getContainerCreateConfig(sandboxDto dto.CreateSandboxDTO) *container.Config {
	envVars := []string{
		"DAYTONA_WS_ID=" + sandboxDto.Id,
		"DAYTONA_WS_IMAGE=" + sandboxDto.Image,
		"DAYTONA_WS_USER=" + sandboxDto.OsUser,
	}

	for key, value := range sandboxDto.Env {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
	}

	return &container.Config{
		Hostname: sandboxDto.Id,
		Image:    sandboxDto.Image,
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

	if len(volumeMountPathBinds) > 0 {
		binds = append(binds, volumeMountPathBinds...)
	}

	hostConfig := &container.HostConfig{
		Privileged: true,
		ExtraHosts: []string{"host.docker.internal:host-gateway"},
		Resources: container.Resources{
			CPUPeriod:  100000,
			CPUQuota:   sandboxDto.CpuQuota * 100000,
			Memory:     sandboxDto.MemoryQuota * 1024 * 1024 * 1024,
			MemorySwap: sandboxDto.MemoryQuota * 1024 * 1024 * 1024,
		},
		Binds: binds,
	}

	containerRuntime := config.GetContainerRuntime()
	if containerRuntime != "" {
		hostConfig.Runtime = containerRuntime
	}

	filesystem, err := d.getFilesystem(ctx)
	if err != nil {
		return nil, err
	}

	if filesystem == "xfs" {
		hostConfig.StorageOpt = map[string]string{
			"size": fmt.Sprintf("%dG", sandboxDto.StorageQuota),
		}
	}

	return hostConfig, nil
}

func (d *DockerClient) getFilesystem(ctx context.Context) (string, error) {
	info, err := d.apiClient.Info(ctx)
	if err != nil {
		return "", err
	}

	for _, driver := range info.DriverStatus {
		if driver[0] == "Backing Filesystem" {
			return driver[1], nil
		}
	}

	return "", errors.New("filesystem not found")
}
