// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"context"

	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func (d *DockerClient) DestroyWorkspace(workspace *workspace.Workspace) error {
	ctx := context.Background()

	networks, err := d.apiClient.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return err
	}

	for _, network := range networks {
		if network.Name == workspace.Id {
			err := d.apiClient.NetworkRemove(ctx, network.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *DockerClient) DestroyProject(project *workspace.Project) error {
	return d.removeProjectContainer(project)
}

func (d *DockerClient) removeProjectContainer(project *workspace.Project) error {
	ctx := context.Background()

	err := d.apiClient.ContainerRemove(ctx, d.GetProjectContainerName(project), container.RemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	})
	if err != nil && !client.IsErrNotFound(err) {
		return err
	}

	err = d.apiClient.VolumeRemove(ctx, d.GetProjectVolumeName(project), true)
	if err != nil && !client.IsErrNotFound(err) {
		return err
	}

	return nil
}
