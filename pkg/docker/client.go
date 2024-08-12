// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"context"
	"fmt"
	"io"

	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/ssh"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/daytonaio/daytona/pkg/workspace/project"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type CreateProjectOptions struct {
	Project    *project.Project
	ProjectDir string
	Cr         *containerregistry.ContainerRegistry
	LogWriter  io.Writer
	Gpc        *gitprovider.GitProviderConfig
	SshClient  *ssh.Client
}

type IDockerClient interface {
	CreateProject(opts *CreateProjectOptions) error
	CreateWorkspace(workspace *workspace.Workspace, workspaceDir string, logWriter io.Writer, sshClient *ssh.Client) error

	DestroyProject(project *project.Project, projectDir string, sshClient *ssh.Client) error
	DestroyWorkspace(workspace *workspace.Workspace, workspaceDir string, sshClient *ssh.Client) error

	StartProject(opts *CreateProjectOptions, daytonaDownloadUrl string) error
	StopProject(project *project.Project, logWriter io.Writer) error

	GetProjectInfo(project *project.Project) (*project.ProjectInfo, error)
	GetWorkspaceInfo(ws *workspace.Workspace) (*workspace.WorkspaceInfo, error)

	GetProjectContainerName(project *project.Project) string
	GetProjectVolumeName(project *project.Project) string
	ExecSync(containerID string, config container.ExecOptions, outputWriter io.Writer) (*ExecResult, error)
	GetContainerLogs(containerName string, logWriter io.Writer) error
	PullImage(imageName string, cr *containerregistry.ContainerRegistry, logWriter io.Writer) error
	PushImage(imageName string, cr *containerregistry.ContainerRegistry, logWriter io.Writer) error
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

func (d *DockerClient) GetProjectContainerName(project *project.Project) string {
	containers, err := d.apiClient.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(filters.Arg("label", fmt.Sprintf("daytona.workspace.id=%s", project.WorkspaceId)), filters.Arg("label", fmt.Sprintf("daytona.project.name=%s", project.Name))),
		All:     true,
	})
	if err != nil || len(containers) == 0 {
		return project.WorkspaceId + "-" + project.Name
	}

	return containers[0].ID
}

func (d *DockerClient) GetProjectVolumeName(project *project.Project) string {
	return project.WorkspaceId + "-" + project.Name
}

func (d *DockerClient) getComposeContainers(c types.ContainerJSON) (string, []types.Container, error) {
	ctx := context.Background()

	for k, v := range c.Config.Labels {
		if k == "com.docker.compose.project" {
			containers, err := d.apiClient.ContainerList(ctx, container.ListOptions{
				Filters: filters.NewArgs(filters.Arg("label", fmt.Sprintf("com.docker.compose.project=%s", v))),
			})
			return v, containers, err
		}
	}

	return "", nil, nil
}
