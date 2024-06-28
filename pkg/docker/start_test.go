// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker_test

import (
	"github.com/daytonaio/daytona/pkg/docker"
	"github.com/daytonaio/daytona/pkg/provider/util"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func (s *DockerClientTestSuite) TestStartProject() {
	s.mockClient.On("ContainerList", mock.Anything, mock.Anything).Return([]types.Container{}, nil)

	containerName := s.dockerClient.GetProjectContainerName(project1)

	s.mockClient.On("ContainerStart", mock.Anything, containerName, container.StartOptions{}).Return(nil)
	s.mockClient.On("ContainerInspect", mock.Anything, containerName).Return(types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			State: &types.ContainerState{
				Running: false,
			},
		},
		Config: &container.Config{
			Labels: map[string]string{},
		},
	}, nil).Once()
	s.mockClient.On("ContainerInspect", mock.Anything, containerName).Return(types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			State: &types.ContainerState{
				Running: true,
			},
		},
		Config: &container.Config{
			Labels: map[string]string{},
		},
	}, nil)

	s.setupExecTest([]string{"bash", "-c", util.GetProjectStartScript("", project1.ApiKey)}, containerName, project1.User, []string{})

	err := s.dockerClient.StartProject(&docker.CreateProjectOptions{
		Project: project1,
	}, "")
	require.Nil(s.T(), err)
}
