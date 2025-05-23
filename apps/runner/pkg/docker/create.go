// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/daytonaio/runner/internal/constants"
	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/docker/docker/errdefs"

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

	if state == enums.SandboxStateStarted || state == enums.SandboxStatePullingSnapshot || state == enums.SandboxStateStarting {
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
	err = d.PullImage(ctx, sandboxDto.Snapshot, sandboxDto.Registry)
	if err != nil {
		return "", err
	}

	d.cache.SetSandboxState(ctx, sandboxDto.Id, enums.SandboxStateCreating)

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

	containerConfig, hostConfig, err := d.getContainerConfigs(ctx, sandboxDto, volumeMountPathBinds)
	if err != nil {
		return "", err
	}

	c, err := d.apiClient.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, sandboxDto.Id)
	if err != nil {
		return "", err
	}

	err = d.Start(ctx, sandboxDto.Id)
	if err != nil {
		return "", err
	}

	// wait for the daemon to start listening on port 2280
	container, err := d.ContainerInspect(ctx, c.ID)
	if err != nil {
		return "", common.NewNotFoundError(fmt.Errorf("sandbox container not found: %w", err))
	}

	var containerIP string
	for _, network := range container.NetworkSettings.Networks {
		containerIP = network.IPAddress
		break
	}

	if containerIP == "" {
		return "", errors.New("container has no IP address, it might not be running")
	}

	// Build the target URL
	targetURL := fmt.Sprintf("http://%s:2280", containerIP)
	target, err := url.Parse(targetURL)
	if err != nil {
		return "", common.NewBadRequestError(fmt.Errorf("failed to parse target URL: %w", err))
	}

	for i := 0; i < 10; i++ {
		conn, err := net.DialTimeout("tcp", target.Host, 1*time.Second)
		if err != nil {
			time.Sleep(50 * time.Millisecond)
			continue
		}
		conn.Close()
		break
	}

	return c.ID, nil
}

func (p *DockerClient) validateImageArchitecture(ctx context.Context, image string) error {
	inspect, _, err := p.apiClient.ImageInspectWithRaw(ctx, image)
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

	return common.NewConflictError(fmt.Errorf("image %s architecture (%s) is not x64 compatible", image, inspect.Architecture))
}
