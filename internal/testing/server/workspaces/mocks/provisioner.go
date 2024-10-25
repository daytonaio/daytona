//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/daytonaio/daytona/pkg/workspace/project"
	"github.com/stretchr/testify/mock"
)

type mockProvisioner struct {
	mock.Mock
}

func NewMockProvisioner() *mockProvisioner {
	return &mockProvisioner{}
}

func (p *mockProvisioner) CreateProject(proj *project.Project, targetConfig *provider.TargetConfig, cr *containerregistry.ContainerRegistry, gc *gitprovider.GitProviderConfig) error {
	args := p.Called(proj, targetConfig, cr, gc)
	return args.Error(0)
}

func (p *mockProvisioner) CreateWorkspace(workspace *workspace.Workspace, targetConfig *provider.TargetConfig) error {
	args := p.Called(workspace, targetConfig)
	return args.Error(0)
}

func (p *mockProvisioner) DestroyProject(proj *project.Project, targetConfig *provider.TargetConfig) error {
	args := p.Called(proj, targetConfig)
	return args.Error(0)
}

func (p *mockProvisioner) DestroyWorkspace(workspace *workspace.Workspace, targetConfig *provider.TargetConfig) error {
	args := p.Called(workspace, targetConfig)
	return args.Error(0)
}

func (p *mockProvisioner) GetWorkspaceInfo(ctx context.Context, w *workspace.Workspace, targetConfig *provider.TargetConfig) (*workspace.WorkspaceInfo, error) {
	args := p.Called(ctx, w, targetConfig)
	return args.Get(0).(*workspace.WorkspaceInfo), args.Error(1)
}

func (p *mockProvisioner) StartProject(proj *project.Project, targetConfig *provider.TargetConfig) error {
	args := p.Called(proj, targetConfig)
	return args.Error(0)
}

func (p *mockProvisioner) StartWorkspace(workspace *workspace.Workspace, targetConfig *provider.TargetConfig) error {
	args := p.Called(workspace, targetConfig)
	return args.Error(0)
}

func (p *mockProvisioner) StopProject(proj *project.Project, targetConfig *provider.TargetConfig) error {
	args := p.Called(proj, targetConfig)
	return args.Error(0)
}

func (p *mockProvisioner) StopWorkspace(workspace *workspace.Workspace, targetConfig *provider.TargetConfig) error {
	args := p.Called(workspace, targetConfig)
	return args.Error(0)
}
