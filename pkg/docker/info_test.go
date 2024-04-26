// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker_test

import (
	"errors"
	"fmt"

	"github.com/daytonaio/daytona/pkg/docker"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/errdefs"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func (s *DockerClientTestSuite) TestGetProjectInfo() {
	containerName := s.dockerClient.GetProjectContainerName(project1)

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

	projectInfo, err := s.dockerClient.GetProjectInfo(project1)
	require.Nil(s.T(), err)
	require.Equal(s.T(), project1.Name, projectInfo.Name)
	require.Equal(s.T(), projectInfo.IsRunning, inspectResult.State.Running)
	require.Equal(s.T(), projectInfo.Created, inspectResult.Created)
	require.Equal(s.T(), projectInfo.ProviderMetadata, metadata)
}

func (s *DockerClientTestSuite) TestGetProjectInfoNotFound() {
	containerName := s.dockerClient.GetProjectContainerName(project1)

	s.mockClient.On("ContainerInspect", mock.Anything, containerName).Return(types.ContainerJSON{}, errdefs.NotFound(errors.New("not found")))

	projectInfo, err := s.dockerClient.GetProjectInfo(project1)
	require.Nil(s.T(), err)
	require.Equal(s.T(), project1.Name, projectInfo.Name)
	require.Equal(s.T(), projectInfo.IsRunning, false)
	require.Equal(s.T(), projectInfo.Created, "")
	require.Equal(s.T(), projectInfo.ProviderMetadata, docker.ContainerNotFoundMetadata)
}

func (s *DockerClientTestSuite) TestGetWorkspaceInfo() {
	workspaceWithoutProjects := &workspace.Workspace{
		Id:     "123",
		Name:   "test",
		Target: "local",
	}

	wsInfo, err := s.dockerClient.GetWorkspaceInfo(workspaceWithoutProjects)
	require.Nil(s.T(), err)
	require.Equal(s.T(), wsInfo.Name, workspaceWithoutProjects.Name)
	require.Equal(s.T(), wsInfo.ProviderMetadata, fmt.Sprintf(docker.WorkspaceMetadataFormat, workspaceWithoutProjects.Id))
}
