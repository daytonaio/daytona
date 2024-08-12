// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker_test

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"

	t_docker "github.com/daytonaio/daytona/internal/testing/docker"
	"github.com/daytonaio/daytona/pkg/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func (s *DockerClientTestSuite) TestCreateWorkspace() {
	workspaceDir := s.T().TempDir()

	err := s.dockerClient.CreateWorkspace(workspace1, workspaceDir, nil, nil)
	require.Nil(s.T(), err)

	_, err = os.Stat(workspaceDir)
	require.Nil(s.T(), err)
}

func (s *DockerClientTestSuite) TestCreateProject() {
	s.mockClient.On("ContainerList", mock.Anything, mock.Anything).Return([]types.Container{}, nil)

	var networkingConfig *network.NetworkingConfig
	var platform *v1.Platform

	projectDir := os.TempDir()

	containerName := s.dockerClient.GetProjectContainerName(project1)

	s.mockClient.On("ImageList", mock.Anything,
		image.ListOptions{
			Filters: filters.NewArgs(filters.Arg("reference", project1.Image)),
		},
	).Return([]image.Summary{}, nil)

	s.mockClient.On("ImagePull", mock.Anything, project1.Image, mock.Anything).Return(t_docker.NewPipeReader(""), nil)
	s.mockClient.On("ImagePull", mock.Anything, "daytonaio/workspace-project", mock.Anything).Return(t_docker.NewPipeReader(""), nil)

	s.mockClient.On("ContainerRemove", mock.Anything, mock.Anything, container.RemoveOptions{RemoveVolumes: true, Force: true}).Return(nil)
	s.mockClient.On("ContainerStart", mock.Anything, mock.Anything, container.StartOptions{}).Return(nil)
	s.mockClient.On("ContainerExecCreate", mock.Anything, mock.Anything, mock.Anything).Return(types.IDResponse{ID: "exec-id"}, nil)
	s.mockClient.On("ContainerStop", mock.Anything, "123", container.StopOptions{}).Return(nil)

	_, client := net.Pipe()
	s.mockClient.On("ContainerExecAttach", mock.Anything, "exec-id", container.ExecStartOptions{}).
		Return(types.HijackedResponse{
			Conn:   client,
			Reader: bufio.NewReader(t_docker.NewPipeReader("")),
		}, nil)
	s.mockClient.On("ContainerExecInspect", mock.Anything, "exec-id").Return(container.ExecInspect{}, nil)

	s.mockClient.On("ContainerCreate", mock.Anything, &container.Config{
		Image:      "daytonaio/workspace-project",
		Entrypoint: []string{"sleep"},
		Cmd:        []string{"infinity"},
		Env:        []string{"GIT_SSL_NO_VERIFY=true"},
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: filepath.Dir(projectDir),
				Target: "/workdir",
			},
		},
	}, networkingConfig, platform, fmt.Sprintf("git-clone-%s-%s", project1.WorkspaceId, project1.Name),
	).Return(container.CreateResponse{ID: "123"}, nil)
	s.mockClient.On("ContainerCreate", mock.Anything, docker.GetContainerCreateConfig(project1),
		&container.HostConfig{
			Privileged: true,
			ExtraHosts: []string{
				"host.docker.internal:host-gateway",
			},
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: projectDir,
					Target: fmt.Sprintf("/home/%s/%s", project1.User, project1.Name),
				},
			},
		},
		networkingConfig,
		platform,
		containerName,
	).Return(container.CreateResponse{ID: "123"}, nil)

	err := s.dockerClient.CreateProject(&docker.CreateProjectOptions{
		Project:    project1,
		ProjectDir: projectDir,
		Cr:         nil,
		LogWriter:  nil,
		Gpc:        nil,
		SshClient:  nil,
	})
	require.Nil(s.T(), err)
}
