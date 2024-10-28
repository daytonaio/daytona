//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/stretchr/testify/mock"
)

type mockProvisioner struct {
	mock.Mock
}

func NewMockProvisioner() *mockProvisioner {
	return &mockProvisioner{}
}

func (p *mockProvisioner) CreateWorkspace(ws *workspace.Workspace, targetConfig *provider.TargetConfig, cr *containerregistry.ContainerRegistry, gc *gitprovider.GitProviderConfig) error {
	args := p.Called(ws, targetConfig, cr, gc)
	return args.Error(0)
}

func (p *mockProvisioner) CreateTarget(target *target.Target, targetConfig *provider.TargetConfig) error {
	args := p.Called(target, targetConfig)
	return args.Error(0)
}

func (p *mockProvisioner) DestroyWorkspace(ws *workspace.Workspace, targetConfig *provider.TargetConfig) error {
	args := p.Called(ws, targetConfig)
	return args.Error(0)
}

func (p *mockProvisioner) DestroyTarget(target *target.Target, targetConfig *provider.TargetConfig) error {
	args := p.Called(target, targetConfig)
	return args.Error(0)
}

func (p *mockProvisioner) GetTargetInfo(ctx context.Context, w *target.Target, targetConfig *provider.TargetConfig) (*target.TargetInfo, error) {
	args := p.Called(ctx, w, targetConfig)
	return args.Get(0).(*target.TargetInfo), args.Error(1)
}

func (p *mockProvisioner) StartWorkspace(ws *workspace.Workspace, targetConfig *provider.TargetConfig) error {
	args := p.Called(ws, targetConfig)
	return args.Error(0)
}

func (p *mockProvisioner) StartTarget(target *target.Target, targetConfig *provider.TargetConfig) error {
	args := p.Called(target, targetConfig)
	return args.Error(0)
}

func (p *mockProvisioner) StopWorkspace(ws *workspace.Workspace, targetConfig *provider.TargetConfig) error {
	args := p.Called(ws, targetConfig)
	return args.Error(0)
}

func (p *mockProvisioner) StopTarget(target *target.Target, targetConfig *provider.TargetConfig) error {
	args := p.Called(target, targetConfig)
	return args.Error(0)
}

func (p *mockProvisioner) GetWorkspaceInfo(ctx context.Context, w *workspace.Workspace, targetConfig *provider.TargetConfig) (*workspace.WorkspaceInfo, error) {
	args := p.Called(ctx, w, targetConfig)
	return args.Get(0).(*workspace.WorkspaceInfo), args.Error(1)
}
