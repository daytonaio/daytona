// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker_test

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func (s *DockerClientTestSuite) TestStartProject() {
	containerName := s.dockerClient.GetProjectContainerName(project1)

	s.mockClient.On("ContainerStart", mock.Anything, containerName, container.StartOptions{}).Return(nil)
	s.mockClient.On("ContainerInspect", mock.Anything, containerName).Return(types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			State: &types.ContainerState{
				Running: false,
			},
		},
	}, nil).Once()
	s.mockClient.On("ContainerInspect", mock.Anything, containerName).Return(types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			State: &types.ContainerState{
				Running: true,
			},
		},
	}, nil)

	err := s.dockerClient.StartProject(project1)
	require.Nil(s.T(), err)
}
