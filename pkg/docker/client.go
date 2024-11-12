// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"context"
	"fmt"
	"io"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/ssh"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type CreateWorkspaceOptions struct {
	Workspace                *models.Workspace
	WorkspaceDir             string
	ContainerRegistry        *models.ContainerRegistry
	LogWriter                io.Writer
	Gpc                      *models.GitProviderConfig
	SshClient                *ssh.Client
	BuilderImage             string
	BuilderContainerRegistry *models.ContainerRegistry
}

type IDockerClient interface {
	CreateWorkspace(opts *CreateWorkspaceOptions) error
	CreateTarget(target *models.Target, targetDir string, logWriter io.Writer, sshClient *ssh.Client) error

	DestroyWorkspace(workspace *models.Workspace, workspaceDir string, sshClient *ssh.Client) error
	DestroyTarget(target *models.Target, targetDir string, sshClient *ssh.Client) error

	StartWorkspace(opts *CreateWorkspaceOptions, daytonaDownloadUrl string) error
	StopWorkspace(workspace *models.Workspace, logWriter io.Writer) error

	GetWorkspaceInfo(workspace *models.Workspace) (*models.WorkspaceInfo, error)
	GetTargetInfo(t *models.Target) (*models.TargetInfo, error)

	GetWorkspaceContainerName(workspace *models.Workspace) string
	GetWorkspaceVolumeName(workspace *models.Workspace) string
	ExecSync(containerID string, config container.ExecOptions, outputWriter io.Writer) (*ExecResult, error)
	GetContainerLogs(containerName string, logWriter io.Writer) error
	PullImage(imageName string, cr *models.ContainerRegistry, logWriter io.Writer) error
	PushImage(imageName string, cr *models.ContainerRegistry, logWriter io.Writer) error
	DeleteImage(imageName string, force bool, logWriter io.Writer) error

	CreateFromDevcontainer(opts CreateDevcontainerOptions) (string, RemoteUser, error)
	RemoveContainer(containerName string) error
}

type DockerClientConfig struct {
	ApiClient client.APIClient
}

func NewDockerClient(config DockerClientConfig) IDockerClient {
	return &DockerClient{
		apiClient: config.ApiClient,
	}
}

type DockerClient struct {
	apiClient client.APIClient
}

func (d *DockerClient) GetWorkspaceContainerName(workspace *models.Workspace) string {
	containers, err := d.apiClient.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(filters.Arg("label", fmt.Sprintf("daytona.target.id=%s", workspace.TargetId)), filters.Arg("label", fmt.Sprintf("daytona.workspace.name=%s", workspace.Name))),
		All:     true,
	})
	if err != nil || len(containers) == 0 {
		return workspace.TargetId + "-" + workspace.Name
	}

	return containers[0].ID
}

func (d *DockerClient) GetWorkspaceVolumeName(workspace *models.Workspace) string {
	return workspace.TargetId + "-" + workspace.Name
}

func (d *DockerClient) getComposeContainers(c types.ContainerJSON) (string, []types.Container, error) {
	ctx := context.Background()

	for k, v := range c.Config.Labels {
		if k == "com.docker.compose.workspace" {
			containers, err := d.apiClient.ContainerList(ctx, container.ListOptions{
				Filters: filters.NewArgs(filters.Arg("label", fmt.Sprintf("com.docker.compose.workspace=%s", v))),
			})
			return v, containers, err
		}
	}

	return "", nil, nil
}
