//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"io"

	"github.com/daytonaio/daytona/pkg/docker"
	"github.com/daytonaio/daytona/pkg/models"
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

func (c *MockClient) CreateWorkspace(w *models.Workspace, serverDownloadUrl string, cr *models.ContainerRegistry, logWriter io.Writer) error {
	args := c.Called(w, serverDownloadUrl, cr, logWriter)
	return args.Error(0)
}

func (c *MockClient) CreateTarget(target *models.Target, logWriter io.Writer) error {
	args := c.Called(target, logWriter)
	return args.Error(0)
}

func (c *MockClient) DestroyWorkspace(w *models.Workspace) error {
	args := c.Called(w)
	return args.Error(0)
}

func (c *MockClient) DestroyTarget(target *models.Target) error {
	args := c.Called(target)
	return args.Error(0)
}

func (c *MockClient) StartWorkspace(w *models.Workspace) error {
	args := c.Called(w)
	return args.Error(0)
}

func (c *MockClient) StopWorkspace(w *models.Workspace) error {
	args := c.Called(w)
	return args.Error(0)
}

func (c *MockClient) GetWorkspaceContainerName(w *models.Workspace) string {
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
