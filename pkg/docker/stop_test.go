// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker_test

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func (s *DockerClientTestSuite) TestStopProject() {
	s.mockClient.On("ContainerList", mock.Anything, mock.Anything).Return([]types.Container{}, nil)

	containerName := s.dockerClient.GetProjectContainerName(project1)

	s.mockClient.On("ContainerStop", mock.Anything, containerName, container.StopOptions{}).Return(nil)
	s.mockClient.On("ContainerInspect", mock.Anything, containerName).Return(types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			State: &types.ContainerState{
				Running: false,
			},
		},
		Config: &container.Config{
			Labels: map[string]string{},
		},
	}, nil)

	err := s.dockerClient.StopProject(project1, nil)
	require.Nil(s.T(), err)
}
