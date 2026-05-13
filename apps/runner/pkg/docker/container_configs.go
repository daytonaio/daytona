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
	"github.com/daytonaio/runner/pkg/volume"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"

	"github.com/docker/docker/api/types/container"
)

func (d *DockerClient) getContainerConfigs(ctx context.Context, sandboxDto dto.CreateSandboxDTO, image *image.InspectResponse, volumeMountPathBinds []string, mounter volume.Mounter) (*container.Config, *container.HostConfig, *network.NetworkingConfig, error) {
	containerConfig, err := d.getContainerCreateConfig(ctx, sandboxDto, image, mounter)
	if err != nil {
		return nil, nil, nil, err
	}

	hostConfig, err := d.getContainerHostConfig(sandboxDto, volumeMountPathBinds, mounter)
	if err != nil {
		return nil, nil, nil, err
	}

	networkingConfig := d.getContainerNetworkingConfig()
	return containerConfig, hostConfig, networkingConfig, nil
}

// volumesToMounterSpec converts the incoming sandbox volume DTOs into the
// package-neutral volume.Volume shape consumed by InContainerMounter.
func volumesToMounterSpec(in []dto.VolumeDTO) []volume.Volume {
	out := make([]volume.Volume, 0, len(in))
	for _, v := range in {
		subpath := ""
		if v.Subpath != nil {
			subpath = *v.Subpath
		}
		out = append(out, volume.Volume{
			VolumeID:          v.VolumeId,
			MountPath:         v.MountPath,
			Subpath:           subpath,
			ReadOnly:          v.ReadOnly,
			LayeredDisk:       v.LayeredDisk,
			LayeredRegion:     v.LayeredRegion,
			LayeredMountToken: v.LayeredMountToken,
		})
	}
	return out
}

func (d *DockerClient) getContainerCreateConfig(ctx context.Context, sandboxDto dto.CreateSandboxDTO, image *image.InspectResponse, mounter volume.Mounter) (*container.Config, error) {
	if image == nil {
		return nil, fmt.Errorf("image not found for sandbox: %s", sandboxDto.Id)
	}

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

	if sandboxDto.OrganizationId != nil && *sandboxDto.OrganizationId != "" {
		envVars = append(envVars, "DAYTONA_ORGANIZATION_ID="+*sandboxDto.OrganizationId)
	}

	if sandboxDto.RegionId != nil && *sandboxDto.RegionId != "" {
		envVars = append(envVars, "DAYTONA_REGION_ID="+*sandboxDto.RegionId)
	}

	// In-container mounters contribute the volume spec + scoped credentials
	// via env so the daemon can mount-s3 from within the sandbox. This may
	// hit the network (e.g. STS AssumeRole) and can fail; surface that error.
	if icm, ok := mounter.(volume.InContainerMounter); ok && len(sandboxDto.Volumes) > 0 {
		extra, err := icm.ContainerEnv(ctx, volumesToMounterSpec(sandboxDto.Volumes))
		if err != nil {
			return nil, fmt.Errorf("in-container volume env: %w", err)
		}
		envVars = append(envVars, extra...)
	}

	labels := make(map[string]string)
	if len(sandboxDto.Volumes) > 0 {
		volumeMountPaths := make([]string, len(sandboxDto.Volumes))
		for i, v := range sandboxDto.Volumes {
			volumeMountPaths[i] = v.MountPath
		}
		labels["daytona.volume_mount_paths"] = strings.Join(volumeMountPaths, ",")
	}
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

func (d *DockerClient) getContainerHostConfig(sandboxDto dto.CreateSandboxDTO, volumeMountPathBinds []string, mounter volume.Mounter) (*container.HostConfig, error) {
	var binds []string

	binds = append(binds, fmt.Sprintf("%s:%s:ro", d.daemonPath, common.DAEMON_PATH))

	// Mount the plugin if available
	if d.computerUsePluginPath != "" {
		binds = append(binds, fmt.Sprintf("%s:/usr/local/lib/daytona-computer-use:ro", d.computerUsePluginPath))
	}

	if len(volumeMountPathBinds) > 0 {
		binds = append(binds, volumeMountPathBinds...)
	}

	// In-container mounters may need extra RO binds (e.g. the mount-s3 binary).
	if icm, ok := mounter.(volume.InContainerMounter); ok {
		binds = append(binds, icm.ContainerBinds()...)
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

	if !d.resourceLimitsDisabled && d.filesystem == "xfs" {
		hostConfig.StorageOpt = map[string]string{
			"size": fmt.Sprintf("%dG", sandboxDto.StorageQuota),
		}
	}

	if d.gpuEnabled {
		nvidiaDevices := []string{
			"/dev/nvidia0",
			"/dev/nvidiactl",
			"/dev/nvidia-uvm",
			"/dev/nvidia-uvm-tools",
			"/dev/nvidia-modeset",
		}
		for _, dev := range nvidiaDevices {
			hostConfig.Devices = append(hostConfig.Devices, container.DeviceMapping{
				PathOnHost:        dev,
				PathInContainer:   dev,
				CgroupPermissions: "rwm",
			})
		}
	}

	return hostConfig, nil
}

func (d *DockerClient) getContainerNetworkingConfig() *network.NetworkingConfig {
	containerNetwork := config.GetContainerNetwork()
	var networkingConfig *network.NetworkingConfig
	if containerNetwork != "" {
		networkingConfig = &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				containerNetwork: {},
			},
		}
	}

	if !d.interSandboxNetworkEnabled {
		if networkingConfig == nil {
			networkingConfig = &network.NetworkingConfig{
				EndpointsConfig: map[string]*network.EndpointSettings{},
			}
		}
		networkingConfig.EndpointsConfig[RUNNER_BRIDGE_NETWORK_NAME] = &network.EndpointSettings{}
	}

	return networkingConfig
}
