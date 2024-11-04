// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker_test

import (
	"testing"

	"github.com/daytonaio/daytona/internal/testing/docker/mocks"
	"github.com/daytonaio/daytona/pkg/docker"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/daytonaio/daytona/pkg/workspace/buildconfig"
	"github.com/stretchr/testify/suite"
)

var workspace1 = &workspace.Workspace{
	Name: "test",
	Repository: &gitprovider.GitRepository{
		Id:   "123",
		Url:  "https://github.com/daytonaio/daytona",
		Name: "daytona",
	},
	Image:       "test-image:tag",
	User:        "test-user",
	TargetId:    "123",
	BuildConfig: &buildconfig.BuildConfig{},
}

var target1 = &target.Target{
	Id:           "123",
	Name:         "test",
	TargetConfig: "local",
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
