// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"time"

	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

func (d *DockerClient) Stop(ctx context.Context, containerId string) error {
	d.cache.SetSandboxState(ctx, containerId, enums.SandboxStateStopping)

	err := d.apiClient.ContainerStop(ctx, containerId, container.StopOptions{
		Signal: "SIGKILL",
	})
	if err != nil {
		return err
	}

	var c types.ContainerJSON

	// TODO: timeout
	for {
		c, err = d.apiClient.ContainerInspect(ctx, containerId)
		if err != nil {
			return err
		}

		if !c.State.Running {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	d.cache.SetSandboxState(ctx, containerId, enums.SandboxStateStopped)

	return nil
}
