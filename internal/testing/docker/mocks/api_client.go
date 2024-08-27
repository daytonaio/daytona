//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/mock"
)

type MockApiClient struct {
	mock.Mock
	client.APIClient
}

func NewMockApiClient() *MockApiClient {
	return &MockApiClient{}
}

func (m *MockApiClient) NetworkList(ctx context.Context, options network.ListOptions) ([]network.Inspect, error) {
	args := m.Called(ctx, options)
	return args.Get(0).([]network.Inspect), args.Error(1)
}

func (m *MockApiClient) NetworkCreate(ctx context.Context, name string, options network.CreateOptions) (network.CreateResponse, error) {
	args := m.Called(ctx, name, options)
	return args.Get(0).(network.CreateResponse), args.Error(1)
}

func (m *MockApiClient) NetworkRemove(ctx context.Context, networkID string) error {
	args := m.Called(ctx, networkID)
	return args.Error(0)
}

func (m *MockApiClient) ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error) {
	args := m.Called(ctx, options)
	return args.Get(0).([]image.Summary), args.Error(1)
}

func (m *MockApiClient) ImagePull(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
	args := m.Called(ctx, ref, options)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockApiClient) ContainerLogs(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error) {
	args := m.Called(ctx, container, options)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockApiClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
	args := m.Called(ctx, config, hostConfig, networkingConfig, platform, containerName)
	return args.Get(0).(container.CreateResponse), args.Error(1)
}

func (m *MockApiClient) ContainerRemove(ctx context.Context, container string, options container.RemoveOptions) error {
	args := m.Called(ctx, container, options)
	return args.Error(0)
}

func (m *MockApiClient) ContainerStart(ctx context.Context, container string, startOptions container.StartOptions) error {
	args := m.Called(ctx, container, startOptions)
	return args.Error(0)
}

func (m *MockApiClient) ContainerStop(ctx context.Context, container string, stopOptions container.StopOptions) error {
	args := m.Called(ctx, container, stopOptions)
	return args.Error(0)
}

func (m *MockApiClient) ContainerExecCreate(ctx context.Context, container string, config container.ExecOptions) (types.IDResponse, error) {
	args := m.Called(ctx, container, config)
	return args.Get(0).(types.IDResponse), args.Error(1)
}

func (m *MockApiClient) ContainerExecStart(ctx context.Context, execID string, config container.ExecStartOptions) error {
	args := m.Called(ctx, execID, config)
	return args.Error(0)
}

func (m *MockApiClient) ContainerExecAttach(ctx context.Context, execID string, config container.ExecStartOptions) (types.HijackedResponse, error) {
	args := m.Called(ctx, execID, config)
	return args.Get(0).(types.HijackedResponse), args.Error(1)
}

func (m *MockApiClient) ContainerExecInspect(ctx context.Context, execID string) (container.ExecInspect, error) {
	args := m.Called(ctx, execID)
	return args.Get(0).(container.ExecInspect), args.Error(1)
}

func (m *MockApiClient) ContainerInspect(ctx context.Context, container string) (types.ContainerJSON, error) {
	args := m.Called(ctx, container)
	return args.Get(0).(types.ContainerJSON), args.Error(1)
}

func (m *MockApiClient) ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
	args := m.Called(ctx, options)
	return args.Get(0).([]types.Container), args.Error(1)
}

func (m *MockApiClient) VolumeRemove(ctx context.Context, volume string, force bool) error {
	args := m.Called(ctx, volume, force)
	return args.Error(0)
}
