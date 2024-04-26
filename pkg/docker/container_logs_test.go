// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker_test

import (
	"io"

	t_docker "github.com/daytonaio/daytona/internal/testing/docker"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func (s *DockerClientTestSuite) TestGetContainerLogs() {
	containerName := s.dockerClient.GetProjectContainerName(project1)
	logWriter := io.MultiWriter(&util.DebugLogWriter{})

	s.mockClient.On("ContainerLogs", mock.Anything, containerName,
		container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
		},
	).Return(t_docker.NewPipeReader(""), nil)

	err := s.dockerClient.GetContainerLogs(containerName, logWriter)
	require.Nil(s.T(), err)

}
