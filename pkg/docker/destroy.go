// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"context"
	"fmt"
	"os"

	"github.com/daytonaio/daytona/pkg/ssh"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/target/workspace"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func (d *DockerClient) DestroyTarget(target *target.Target, targetDir string, sshClient *ssh.Client) error {
	if sshClient == nil {
		return os.RemoveAll(targetDir)
	} else {
		return sshClient.Exec(fmt.Sprintf("rm -rf %s", targetDir), nil)
	}
}

func (d *DockerClient) DestroyWorkspace(workspace *workspace.Workspace, workspaceDir string, sshClient *ssh.Client) error {
	err := d.removeWorkspaceContainer(workspace)
	if err != nil {
		return err
	}

	if sshClient == nil {
		return os.RemoveAll(workspaceDir)
	} else {
		return sshClient.Exec(fmt.Sprintf("rm -rf %s", workspaceDir), nil)
	}
}

func (d *DockerClient) removeWorkspaceContainer(w *workspace.Workspace) error {
	ctx := context.Background()

	containerName := d.GetWorkspaceContainerName(w)

	c, err := d.apiClient.ContainerInspect(ctx, containerName)
	if err != nil {
		if client.IsErrNotFound(err) {
			return nil
		}
		return err
	}

	err = d.RemoveContainer(containerName)
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

func (d *DockerClient) RemoveContainer(containerName string) error {
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
