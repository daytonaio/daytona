// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker_test

import (
	"bufio"
	"net"

	t_docker "github.com/daytonaio/daytona/internal/testing/docker"
	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func (s *DockerClientTestSuite) TestExecSync() {
	containerName := s.dockerClient.GetProjectContainerName(project1)
	_, client := net.Pipe()

	s.mockClient.On("ContainerExecCreate", mock.Anything, containerName, types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"test-cmd"},
		User:         project1.User,
		Env:          []string{"DEBIAN_FRONTEND=noninteractive"},
	}).Return(types.IDResponse{
		ID: "123",
	}, nil)
	s.mockClient.On("ContainerExecAttach", mock.Anything, "123", types.ExecStartCheck{}).Return(types.HijackedResponse{
		Conn:   client,
		Reader: bufio.NewReader(t_docker.NewPipeReader("")),
	}, nil)
	s.mockClient.On("ContainerExecInspect", mock.Anything, "123").Return(types.ContainerExecInspect{
		ExitCode: 0,
	}, nil)

	result, err := s.dockerClient.ExecSync(containerName, types.ExecConfig{
		Cmd:  []string{"test-cmd"},
		User: project1.User,
	}, nil)
	require.Nil(s.T(), err)
	require.Equal(s.T(), 0, result.ExitCode)
	require.Equal(s.T(), "", result.StdOut)
}
