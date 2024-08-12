// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker_test

import (
	"bufio"
	"net"

	t_docker "github.com/daytonaio/daytona/internal/testing/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func (s *DockerClientTestSuite) TestExecSync() {
	s.mockClient.On("ContainerList", mock.Anything, mock.Anything).Return([]types.Container{}, nil)

	containerName := s.dockerClient.GetProjectContainerName(project1)

	s.setupExecTest([]string{"test-cmd"}, containerName, project1.User, []string{})

	result, err := s.dockerClient.ExecSync(containerName, container.ExecOptions{
		Cmd:  []string{"test-cmd"},
		User: project1.User,
	}, nil)
	require.Nil(s.T(), err)
	require.Equal(s.T(), 0, result.ExitCode)
	require.Equal(s.T(), "", result.StdOut)
}

func (s *DockerClientTestSuite) setupExecTest(cmd []string, containerName, user string, env []string) {
	_, client := net.Pipe()

	s.mockClient.On("ContainerExecCreate", mock.Anything, containerName, container.ExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
		User:         user,
		Env: append([]string{
			"DEBIAN_FRONTEND=noninteractive",
		}, env...),
	}).Return(types.IDResponse{
		ID: "123",
	}, nil)
	s.mockClient.On("ContainerExecAttach", mock.Anything, "123", container.ExecStartOptions{}).Return(types.HijackedResponse{
		Conn:   client,
		Reader: bufio.NewReader(t_docker.NewPipeReader("")),
	}, nil)
	s.mockClient.On("ContainerExecInspect", mock.Anything, "123").Return(container.ExecInspect{
		ExitCode: 0,
	}, nil)
}
