// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"context"
	"time"

	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/docker/docker/api/types/container"
)

func (d *DockerClient) StopProject(project *workspace.Project) error {
	return d.stopProjectContainer(project)
}

func (d *DockerClient) stopProjectContainer(project *workspace.Project) error {
	containerName := d.GetProjectContainerName(project)
	ctx := context.Background()

	err := d.apiClient.ContainerStop(ctx, containerName, container.StopOptions{})
	if err != nil {
		return err
	}

	//	TODO: timeout
	for {
		inspect, err := d.apiClient.ContainerInspect(ctx, containerName)
		if err != nil {
			return err
		}

		if !inspect.State.Running {
			break
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}
