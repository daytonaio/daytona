// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker_test

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func (s *DockerClientTestSuite) TestGetWorkspaceProviderMetadata() {
	s.mockClient.On("ContainerList", mock.Anything, mock.Anything).Return([]types.Container{}, nil)

	containerName := s.dockerClient.GetWorkspaceContainerName(workspace1)

	inspectResult := types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			State: &types.ContainerState{
				Running: true,
			},
			Created: "test-created",
		},
		Config: &container.Config{
			Labels: map[string]string{
				"test": "label",
			},
		},
	}
	metadata := `{"test":"label"}`

	s.mockClient.On("ContainerInspect", mock.Anything, containerName).Return(inspectResult, nil)

	workspaceMetadata, err := s.dockerClient.GetWorkspaceProviderMetadata(workspace1)
	require.Nil(s.T(), err)
	require.Equal(s.T(), workspaceMetadata, metadata)
}
