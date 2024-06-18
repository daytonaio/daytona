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
	err := s.dockerClient.DestroyWorkspace(workspace1)
	require.Nil(s.T(), err)
}

func (s *DockerClientTestSuite) TestDestroyProject() {
	s.mockClient.On("ContainerList", mock.Anything, mock.Anything).Return([]types.Container{}, nil)

	containerName := s.dockerClient.GetProjectContainerName(project1)

	s.mockClient.On("ContainerInspect", mock.Anything, containerName).Return(types.ContainerJSON{
		Config: &container.Config{},
	}, nil)

	s.mockClient.On("ContainerRemove", mock.Anything, containerName,
		container.RemoveOptions{
			Force:         true,
			RemoveVolumes: true,
		},
	).Return(nil)

	s.mockClient.On("VolumeRemove", mock.Anything, s.dockerClient.GetProjectVolumeName(project1), true).Return(nil)

	err := s.dockerClient.DestroyProject(project1)
	require.Nil(s.T(), err)
}
