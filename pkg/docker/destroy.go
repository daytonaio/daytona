// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"context"
	"fmt"
	"os"

	"github.com/daytonaio/daytona/pkg/ssh"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/daytonaio/daytona/pkg/workspace/project"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func (d *DockerClient) DestroyWorkspace(workspace *workspace.Workspace, workspaceDir string, sshClient *ssh.Client) error {
	if sshClient == nil {
		return os.RemoveAll(workspaceDir)
	} else {
		return sshClient.Exec(fmt.Sprintf("rm -rf %s", workspaceDir), nil)
	}
}

func (d *DockerClient) DestroyProject(project *project.Project, projectDir string, sshClient *ssh.Client) error {
	err := d.removeProjectContainer(project)
	if err != nil {
		return err
	}

	if sshClient == nil {
		return os.RemoveAll(projectDir)
	} else {
		return sshClient.Exec(fmt.Sprintf("rm -rf %s", projectDir), nil)
	}
}

func (d *DockerClient) removeProjectContainer(p *project.Project) error {
	ctx := context.Background()

	containerName := d.GetProjectContainerName(p)

	c, err := d.apiClient.ContainerInspect(ctx, containerName)
	if err != nil {
		if client.IsErrNotFound(err) {
			return nil
		}
		return err
	}

	err = d.removeContainer(containerName)
	if err != nil {
		return err
	}

	err = d.apiClient.VolumeRemove(ctx, containerName, true)
	if err != nil && !client.IsErrNotFound(err) {
		return err
	}

	// TODO: Add logging
	_, composeContainers, err := d.getComposeContainers(c)
	if err != nil {
		return err
	}

	if composeContainers == nil {
		return nil
	}

	for _, c := range composeContainers {
		err = d.apiClient.ContainerRemove(ctx, c.ID, container.RemoveOptions{
			Force:         true,
			RemoveVolumes: true,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DockerClient) removeContainer(containerName string) error {
	ctx := context.Background()

	err := d.apiClient.ContainerRemove(ctx, containerName, container.RemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	})
	if err != nil && !client.IsErrNotFound(err) {
		return err
	}

	return nil
}
