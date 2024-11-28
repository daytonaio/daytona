// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker_test

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/docker"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func (s *DockerClientTestSuite) TestGetWorkspaceInfo() {
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

	workspaceInfo, err := s.dockerClient.GetWorkspaceInfo(workspace1)
	require.Nil(s.T(), err)
	require.Equal(s.T(), workspace1.Name, workspaceInfo.Name)
	require.Equal(s.T(), workspaceInfo.IsRunning, inspectResult.State.Running)
	require.Equal(s.T(), workspaceInfo.Created, inspectResult.Created)
	require.Equal(s.T(), workspaceInfo.ProviderMetadata, metadata)
}

func (s *DockerClientTestSuite) TestGetTargetInfo() {
	var targetConfig = &models.TargetConfig{
		Name: "test",
		ProviderInfo: models.ProviderInfo{
			Name:    "test-provider",
			Version: "test",
		},
		Options: "test-options",
		Deleted: false,
	}

	targetWithoutWorkspaces := &models.Target{
		Id:             "123",
		Name:           "test",
		TargetConfigId: targetConfig.Id,
		TargetConfig:   *targetConfig,
	}

	targetInfo, err := s.dockerClient.GetTargetInfo(targetWithoutWorkspaces)
	require.Nil(s.T(), err)
	require.Equal(s.T(), targetInfo.Name, targetWithoutWorkspaces.Name)
	require.Equal(s.T(), targetInfo.ProviderMetadata, fmt.Sprintf(docker.TargetMetadataFormat, targetWithoutWorkspaces.Id))
}
