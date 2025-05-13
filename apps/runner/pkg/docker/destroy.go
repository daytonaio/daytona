// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/errdefs"

	log "github.com/sirupsen/logrus"
)

func (d *DockerClient) Destroy(ctx context.Context, containerName string) error {
	startTime := time.Now()
	defer func() {
		obs, err := common.ContainerOperationDuration.GetMetricWithLabelValues("destroy")
		if err == nil {
			obs.Observe(time.Since(startTime).Seconds())
		}
	}()

	state, err := d.DeduceSandboxState(ctx, containerName)
	if err != nil && state == enums.SandboxStateError {
		return err
	}

	if state == enums.SandboxStateDestroyed || state == enums.SandboxStateDestroying {
		return nil
	}

	d.cache.SetSandboxState(ctx, containerName, enums.SandboxStateDestroying)

	_, err = d.ContainerInspect(ctx, containerName)
	if err != nil {
		if errdefs.IsNotFound(err) {
			d.cache.SetSandboxState(ctx, containerName, enums.SandboxStateDestroyed)
		}
		return err
	}

	err = d.apiClient.ContainerRemove(ctx, containerName, container.RemoveOptions{
		Force: true,
	})
	if err != nil {
		if errdefs.IsNotFound(err) {
			d.cache.SetSandboxState(ctx, containerName, enums.SandboxStateDestroyed)
		}
		return err
	}

	d.cache.SetSandboxState(ctx, containerName, enums.SandboxStateDestroyed)

	return nil
}

func (d *DockerClient) RemoveDestroyed(ctx context.Context, containerName string) error {

	// Check if container exists and is in destroyed state
	state, err := d.DeduceSandboxState(ctx, containerName)
	if err != nil {
		return err
	}

	if state != enums.SandboxStateDestroyed {
		return common.NewBadRequestError(fmt.Errorf("container %s is not in destroyed state", containerName))
	}

	// Remove the container
	err = d.apiClient.ContainerRemove(ctx, containerName, container.RemoveOptions{
		Force: true,
	})
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil // Container already removed
		}
		return err
	}

	log.Infof("Destroyed container %s removed successfully", containerName)

	return nil
}
