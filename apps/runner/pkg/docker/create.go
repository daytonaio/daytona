// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/containerd/errdefs"
	"github.com/daytonaio/common-go/pkg/timer"
	"github.com/daytonaio/runner/internal/constants"
	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/models/enums"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	log "github.com/sirupsen/logrus"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
)

func (d *DockerClient) Create(ctx context.Context, sandboxDto dto.CreateSandboxDTO) (string, error) {
	defer timer.Timer()()

	startTime := time.Now()
	defer func() {
		obs, err := common.ContainerOperationDuration.GetMetricWithLabelValues("create")
		if err == nil {
			obs.Observe(time.Since(startTime).Seconds())
		}
	}()

	state, err := d.DeduceSandboxState(ctx, sandboxDto.Id)
	if err != nil && state == enums.SandboxStateError {
		return "", err
	}

	if state == enums.SandboxStateStarted || state == enums.SandboxStatePullingSnapshot || state == enums.SandboxStateStarting {
		return sandboxDto.Id, nil
	}

	if state == enums.SandboxStateStopped || state == enums.SandboxStateCreating {
		err = d.Start(ctx, sandboxDto.Id, sandboxDto.AuthToken, sandboxDto.Metadata)
		if err != nil {
			return "", err
		}

		return sandboxDto.Id, nil
	}

	d.statesCache.SetSandboxState(ctx, sandboxDto.Id, enums.SandboxStateCreating)

	ctx = context.WithValue(ctx, constants.ID_KEY, sandboxDto.Id)
	err = d.PullImage(ctx, sandboxDto.Snapshot, sandboxDto.Registry)
	if err != nil {
		return "", err
	}

	d.statesCache.SetSandboxState(ctx, sandboxDto.Id, enums.SandboxStateCreating)

	err = d.validateImageArchitecture(ctx, sandboxDto.Snapshot)
	if err != nil {
		log.Errorf("ERROR: %s.\n", err.Error())
		return "", err
	}

	volumeMountPathBinds := make([]string, 0)
	if sandboxDto.Volumes != nil {
		volumeMountPathBinds, err = d.getVolumesMountPathBinds(ctx, sandboxDto.Volumes)
		if err != nil {
			return "", err
		}
	}

	containerConfig, hostConfig, networkingConfig, err := d.getContainerConfigs(ctx, sandboxDto, volumeMountPathBinds)
	if err != nil {
		return "", err
	}

	c, err := d.apiClient.ContainerCreate(ctx, containerConfig, hostConfig, networkingConfig, &v1.Platform{
		Architecture: "amd64",
		OS:           "linux",
	}, sandboxDto.Id)
	if err != nil {
		// Container already exists and is being created by another process
		if errdefs.IsConflict(err) {
			return sandboxDto.Id, nil
		}
		return "", err
	}

	err = d.Start(ctx, sandboxDto.Id, sandboxDto.AuthToken, sandboxDto.Metadata)
	if err != nil {
		return "", err
	}

	containerShortId := c.ID[:12]
	info, err := d.apiClient.ContainerInspect(context.Background(), sandboxDto.Id)
	if err != nil {
		log.Errorf("Failed to inspect container: %v", err)
	}
	ip := common.GetContainerIpAddress(ctx, info)

	if sandboxDto.NetworkBlockAll != nil && *sandboxDto.NetworkBlockAll {
		go func() {
			err = d.netRulesManager.SetNetworkRules(containerShortId, ip, "")
			if err != nil {
				log.Errorf("Failed to update sandbox network settings: %v", err)
			}
		}()
	} else if sandboxDto.NetworkAllowList != nil && *sandboxDto.NetworkAllowList != "" {
		go func() {
			err = d.netRulesManager.SetNetworkRules(containerShortId, ip, *sandboxDto.NetworkAllowList)
			if err != nil {
				log.Errorf("Failed to update sandbox network settings: %v", err)
			}
		}()
	}

	if sandboxDto.Metadata != nil && sandboxDto.Metadata["limitNetworkEgress"] == "true" {
		go func() {
			err = d.netRulesManager.SetNetworkLimiter(containerShortId, ip)
			if err != nil {
				log.Errorf("Failed to update sandbox network settings: %v", err)
			}
		}()
	}

	return c.ID, nil
}

func (p *DockerClient) validateImageArchitecture(ctx context.Context, image string) error {
	defer timer.Timer()()

	inspect, err := p.apiClient.ImageInspect(ctx, image)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return err
		}
		return fmt.Errorf("failed to inspect image: %w", err)
	}

	arch := strings.ToLower(inspect.Architecture)
	validArchs := []string{"amd64", "x86_64"}

	for _, validArch := range validArchs {
		if arch == validArch {
			return nil
		}
	}

	return common_errors.NewConflictError(fmt.Errorf("image %s architecture (%s) is not x64 compatible", image, inspect.Architecture))
}
