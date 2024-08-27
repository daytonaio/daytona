// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker_test

import (
	"testing"

	"github.com/daytonaio/daytona/internal/testing/docker/mocks"
	"github.com/daytonaio/daytona/pkg/docker"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/daytonaio/daytona/pkg/workspace/project"
	"github.com/daytonaio/daytona/pkg/workspace/project/config"
	"github.com/stretchr/testify/suite"
)

var project1 = &project.Project{
	ProjectConfig: config.ProjectConfig{
		Name: "test",
		Repository: &gitprovider.GitRepository{
			Id:   "123",
			Url:  "https://github.com/daytonaio/daytona",
			Name: "daytona",
		},
		Image: "test-image:tag",
		User:  "test-user",
	},
	WorkspaceId: "123",
	Target:      "local",
}

var workspace1 = &workspace.Workspace{
	Id:     "123",
	Name:   "test",
	Target: "local",
	Projects: []*project.Project{
		project1,
	},
}

type DockerClientTestSuiteConfig struct {
	dockerClient docker.IDockerClient
	mockClient   *mocks.MockApiClient
}

func NewDockerClientTestSuite(config DockerClientTestSuiteConfig) *DockerClientTestSuite {
	return &DockerClientTestSuite{
		dockerClient: config.dockerClient,
		mockClient:   config.mockClient,
	}
}

type DockerClientTestSuite struct {
	suite.Suite
	dockerClient docker.IDockerClient
	mockClient   *mocks.MockApiClient
}

func (s *DockerClientTestSuite) AfterTest(_, _ string) {
	s.mockClient.AssertExpectations(s.T())
	s.mockClient.ExpectedCalls = nil
}

func TestDockerClient(t *testing.T) {
	mockClient := mocks.NewMockApiClient()

	dockerClient := docker.NewDockerClient(docker.DockerClientConfig{
		ApiClient: mockClient,
	})

	suite.Run(t, NewDockerClientTestSuite(DockerClientTestSuiteConfig{
		dockerClient: dockerClient,
		mockClient:   mockClient,
	}))
}
