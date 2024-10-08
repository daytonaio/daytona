//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"io"

	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/docker"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/daytonaio/daytona/pkg/workspace/project"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/mock"
)

// This is meant to mock DockerClient calls in provider tests
type MockClient struct {
	mock.Mock
}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (c *MockClient) CreateProject(p *project.Project, serverDownloadUrl string, cr *containerregistry.ContainerRegistry, logWriter io.Writer) error {
	args := c.Called(p, serverDownloadUrl, cr, logWriter)
	return args.Error(0)
}

func (c *MockClient) CreateWorkspace(workspace *workspace.Workspace, logWriter io.Writer) error {
	args := c.Called(workspace, logWriter)
	return args.Error(0)
}

func (c *MockClient) DestroyProject(p *project.Project) error {
	args := c.Called(p)
	return args.Error(0)
}

func (c *MockClient) DestroyWorkspace(workspace *workspace.Workspace) error {
	args := c.Called(workspace)
	return args.Error(0)
}

func (c *MockClient) StartProject(p *project.Project) error {
	args := c.Called(p)
	return args.Error(0)
}

func (c *MockClient) StopProject(p *project.Project) error {
	args := c.Called(p)
	return args.Error(0)
}

func (c *MockClient) GetProjectInfo(p *project.Project) (*project.ProjectInfo, error) {
	args := c.Called(p)
	return args.Get(0).(*project.ProjectInfo), args.Error(1)
}

func (c *MockClient) GetWorkspaceInfo(ws *workspace.Workspace) (*workspace.WorkspaceInfo, error) {
	args := c.Called(ws)
	return args.Get(0).(*workspace.WorkspaceInfo), args.Error(1)
}

func (c *MockClient) GetProjectContainerName(p *project.Project) string {
	args := c.Called(p)
	return args.String(0)
}

func (c *MockClient) ExecSync(containerID string, config container.ExecOptions, outputWriter io.Writer) (*docker.ExecResult, error) {
	args := c.Called(containerID, config, outputWriter)
	return args.Get(0).(*docker.ExecResult), args.Error(1)
}

func (c *MockClient) GetContainerLogs(containerName string, logWriter io.Writer) error {
	args := c.Called(containerName, logWriter)
	return args.Error(0)
}
