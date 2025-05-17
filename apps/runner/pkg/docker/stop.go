// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/docker/docker/api/types/container"
)

func (d *DockerClient) Stop(ctx context.Context, containerName string) error {
	d.cache.SetSandboxState(ctx, containerName, enums.SandboxStateStopping)

	err := d.apiClient.ContainerStop(ctx, containerName, container.StopOptions{
		Signal: "SIGKILL",
	})
	if err != nil {
		return err
	}

	err = d.waitForContainerStopped(ctx, containerName, 10*time.Second)
	if err != nil {
		return err
	}

	d.cache.SetSandboxState(ctx, containerName, enums.SandboxStateStopped)

	return nil
}

func (d *DockerClient) waitForContainerStopped(ctx context.Context, containerName string, timeout time.Duration) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("timeout waiting for container %s to stop", containerName)
		case <-ticker.C:
			c, err := d.ContainerInspect(ctx, containerName)
			if err != nil {
				return err
			}

			if !c.State.Running {
				return nil
			}
		}
	}
}
