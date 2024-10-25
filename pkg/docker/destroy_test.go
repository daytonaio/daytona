// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker_test

import (
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func (s *DockerClientTestSuite) TestDestroyTarget() {
	targetDir := s.T().TempDir()

	err := s.dockerClient.DestroyTarget(target1, targetDir, nil)
	require.Nil(s.T(), err)

	_, err = os.Stat(targetDir)
	require.True(s.T(), os.IsNotExist(err))
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

	projectDir := s.T().TempDir()

	err := s.dockerClient.DestroyProject(project1, projectDir, nil)
	require.Nil(s.T(), err)

	_, err = os.Stat(projectDir)
	require.True(s.T(), os.IsNotExist(err))
}
