//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"io"

	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/docker"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/workspace"
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

func (c *MockClient) CreateWorkspace(w *workspace.Workspace, serverDownloadUrl string, cr *containerregistry.ContainerRegistry, logWriter io.Writer) error {
	args := c.Called(w, serverDownloadUrl, cr, logWriter)
	return args.Error(0)
}

func (c *MockClient) CreateTarget(target *target.Target, logWriter io.Writer) error {
	args := c.Called(target, logWriter)
	return args.Error(0)
}

func (c *MockClient) DestroyWorkspace(w *workspace.Workspace) error {
	args := c.Called(w)
	return args.Error(0)
}

func (c *MockClient) DestroyTarget(target *target.Target) error {
	args := c.Called(target)
	return args.Error(0)
}

func (c *MockClient) StartWorkspace(w *workspace.Workspace) error {
	args := c.Called(w)
	return args.Error(0)
}

func (c *MockClient) StopWorkspace(w *workspace.Workspace) error {
	args := c.Called(w)
	return args.Error(0)
}

func (c *MockClient) GetWorkspaceInfo(w *workspace.Workspace) (*workspace.WorkspaceInfo, error) {
	args := c.Called(w)
	return args.Get(0).(*workspace.WorkspaceInfo), args.Error(1)
}

func (c *MockClient) GetTargetInfo(t *target.Target) (*target.TargetInfo, error) {
	args := c.Called(t)
	return args.Get(0).(*target.TargetInfo), args.Error(1)
}

func (c *MockClient) GetWorkspaceContainerName(w *workspace.Workspace) string {
	args := c.Called(w)
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
