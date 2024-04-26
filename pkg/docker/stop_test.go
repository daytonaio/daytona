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
	containerName := s.dockerClient.GetProjectContainerName(project1)

	s.mockClient.On("ContainerStop", mock.Anything, containerName, container.StopOptions{}).Return(nil)
	s.mockClient.On("ContainerInspect", mock.Anything, containerName).Return(types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			State: &types.ContainerState{
				Running: false,
			},
		},
	}, nil)

	err := s.dockerClient.StopProject(project1)
	require.Nil(s.T(), err)
}
