// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"context"
	"fmt"
	"io"

	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/provider/util"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
)

func (d *DockerClient) CreateWorkspace(workspace *workspace.Workspace, logWriter io.Writer) error {
	if logWriter != nil {
		logWriter.Write([]byte("Initializing network\n"))
	}
	ctx := context.Background()

	networks, err := d.apiClient.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return err
	}

	for _, network := range networks {
		if network.Name == workspace.Id {
			if logWriter != nil {
				logWriter.Write([]byte("Network already exists\n"))
			}
			return nil
		}
	}

	_, err = d.apiClient.NetworkCreate(ctx, workspace.Id, types.NetworkCreate{
		Attachable: true,
		Driver:     "bridge",
	})
	if err != nil {
		return err
	}

	if logWriter != nil {
		logWriter.Write([]byte("Network initialized\n"))
	}
	return nil
}

func (d *DockerClient) CreateProject(project *workspace.Project, daytonaDownloadUrl string, cr *containerregistry.ContainerRegistry, logWriter io.Writer) error {
	err := d.PullImage(project.Image, cr, logWriter)
	if err != nil {
		return err
	}

	return d.initProjectContainer(project, daytonaDownloadUrl)
}

func (d *DockerClient) initProjectContainer(project *workspace.Project, daytonaDownloadUrl string) error {
	ctx := context.Background()

	_, err := d.apiClient.ContainerCreate(ctx, GetContainerCreateConfig(project, daytonaDownloadUrl), &container.HostConfig{
		Privileged:  true,
		NetworkMode: container.NetworkMode(project.WorkspaceId),
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeVolume,
				Source: d.GetProjectVolumeName(project),
				Target: fmt.Sprintf("/home/%s/%s", project.User, project.Name),
			},
		},
		ExtraHosts: []string{
			"host.docker.internal:host-gateway",
		},
	}, nil, nil, d.GetProjectContainerName(project))
	if err != nil {
		return err
	}

	return nil
}

func GetContainerCreateConfig(project *workspace.Project, daytonaDownloadUrl string) *container.Config {
	envVars := []string{}

	for key, value := range project.EnvVars {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
	}

	return &container.Config{
		Hostname: project.Name,
		Image:    project.Image,
		Labels: map[string]string{
			"daytona.workspace.id":                     project.WorkspaceId,
			"daytona.workspace.project.name":           project.Name,
			"daytona.workspace.project.repository.url": project.Repository.Url,
		},
		User:         project.User,
		Env:          envVars,
		Entrypoint:   []string{"bash", "-c", util.GetProjectStartScript(daytonaDownloadUrl, project.ApiKey)},
		AttachStdout: true,
		AttachStderr: true,
	}
}
