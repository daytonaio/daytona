// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"context"
	"time"

	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/docker/docker/api/types/container"
)

func (d *DockerClient) StartProject(project *workspace.Project) error {
	return d.startProjectContainer(project)
}

func (d *DockerClient) startProjectContainer(project *workspace.Project) error {
	containerName := d.GetProjectContainerName(project)
	ctx := context.Background()

	inspect, err := d.apiClient.ContainerInspect(ctx, containerName)

	if err == nil && inspect.State.Running {
		return nil
	}

	err = d.apiClient.ContainerStart(ctx, containerName, container.StartOptions{})
	if err != nil {
		return err
	}

	// make sure container is running
	//	TODO: timeout
	for {
		inspect, err := d.apiClient.ContainerInspect(ctx, containerName)
		if err != nil {
			return err
		}

		if inspect.State.Running {
			break
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}
