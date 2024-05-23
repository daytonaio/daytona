// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker_test

import (
	t_docker "github.com/daytonaio/daytona/internal/testing/docker"
	"github.com/daytonaio/daytona/pkg/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func (s *DockerClientTestSuite) TestCreateWorkspace() {
	s.mockClient.On("NetworkList", mock.Anything, types.NetworkListOptions{}).Return([]types.NetworkResource{}, nil)
	s.mockClient.On("NetworkCreate", mock.Anything, workspace1.Id,
		types.NetworkCreate{
			Attachable: true,
		},
	).Return(types.NetworkCreateResponse{}, nil)

	err := s.dockerClient.CreateWorkspace(workspace1, nil)
	require.Nil(s.T(), err)
}

func (s *DockerClientTestSuite) TestCreateProject() {
	var networkingConfig *network.NetworkingConfig
	var platform *v1.Platform

	containerName := s.dockerClient.GetProjectContainerName(project1)

	s.mockClient.On("ImageList", mock.Anything,
		image.ListOptions{
			Filters: filters.NewArgs(filters.Arg("reference", project1.Image)),
		},
	).Return([]image.Summary{}, nil)

	s.mockClient.On("ImagePull", mock.Anything, project1.Image, mock.Anything).Return(t_docker.NewPipeReader(""), nil)

	s.mockClient.On("ContainerCreate", mock.Anything, docker.GetContainerCreateConfig(project1, "download-url"),
		&container.HostConfig{
			Privileged:  true,
			NetworkMode: container.NetworkMode(project1.WorkspaceId),
		},
		networkingConfig,
		platform,
		containerName,
	).Return(container.CreateResponse{ID: "123"}, nil)

	err := s.dockerClient.CreateProject(project1, "download-url", nil, nil)
	require.Nil(s.T(), err)
}
