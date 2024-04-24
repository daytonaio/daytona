// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker_test

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func (s *DockerClientTestSuite) TestDestroyWorkspace() {
	networks := []types.NetworkResource{
		{
			ID:   workspace1.Id,
			Name: workspace1.Id,
		},
	}

	s.mockClient.On("NetworkList", mock.Anything, types.NetworkListOptions{}).Return(networks, nil)
	s.mockClient.On("NetworkRemove", mock.Anything, workspace1.Id).Return(nil)

	err := s.dockerClient.DestroyWorkspace(workspace1)
	require.Nil(s.T(), err)
}

func (s *DockerClientTestSuite) TestDestroyProject() {
	containerName := s.dockerClient.GetProjectContainerName(project1)

	s.mockClient.On("ContainerRemove", mock.Anything, containerName,
		container.RemoveOptions{
			Force:         true,
			RemoveVolumes: true,
		},
	).Return(nil)

	err := s.dockerClient.DestroyProject(project1)
	require.Nil(s.T(), err)
}
