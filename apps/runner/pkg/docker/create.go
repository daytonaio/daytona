// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"time"

	"github.com/daytonaio/runner/internal/constants"
	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/models/enums"

	log "github.com/sirupsen/logrus"
)

func (d *DockerClient) Create(ctx context.Context, sandboxDto dto.CreateSandboxDTO) (string, error) {
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

	if state == enums.SandboxStateStarted || state == enums.SandboxStatePullingImage || state == enums.SandboxStateStarting {
		return sandboxDto.Id, nil
	}

	if state == enums.SandboxStateStopped || state == enums.SandboxStateCreating {
		err = d.Start(ctx, sandboxDto.Id)
		if err != nil {
			return "", err
		}

		return sandboxDto.Id, nil
	}

	d.cache.SetSandboxState(ctx, sandboxDto.Id, enums.SandboxStateCreating)

	ctx = context.WithValue(ctx, constants.ID_KEY, sandboxDto.Id)
	err = d.PullImage(ctx, sandboxDto.Image, sandboxDto.Registry)
	if err != nil {
		return "", err
	}

	d.cache.SetSandboxState(ctx, sandboxDto.Id, enums.SandboxStateCreating)

	err = d.validateImageArchitecture(ctx, sandboxDto.Image)
	if err != nil {
		log.Errorf("ERROR: %s.\n", err.Error())
		return "", err
	}

	err = d.ensureBinary(d.daytonaBinaryURL, d.daytonaBinaryPath, "Daytona")
	if err != nil {
		log.Errorf("ERROR: %s.\n", err.Error())
		return "", err
	}

	err = d.ensureBinary(d.terminalBinaryURL, d.terminalBinaryPath, "Terminal")
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

	containerConfig, hostConfig, err := d.getContainerConfigs(ctx, sandboxDto, volumeMountPathBinds)
	if err != nil {
		return "", err
	}

	c, err := d.apiClient.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, sandboxDto.Id)
	if err != nil {
		return "", err
	}

	return c.ID, d.Start(ctx, sandboxDto.Id)
}
