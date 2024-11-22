//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/provisioner"
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

func (p *mockProvisioner) CreateProject(params provisioner.ProjectParams) error {
	args := p.Called(params)
	return args.Error(0)
}

func (p *mockProvisioner) CreateWorkspace(workspace *workspace.Workspace, target *provider.ProviderTarget) error {
	args := p.Called(workspace, target)
	return args.Error(0)
}

func (p *mockProvisioner) DestroyProject(proj *project.Project, target *provider.ProviderTarget) error {
	args := p.Called(proj, target)
	return args.Error(0)
}

func (p *mockProvisioner) DestroyWorkspace(workspace *workspace.Workspace, target *provider.ProviderTarget) error {
	args := p.Called(workspace, target)
	return args.Error(0)
}

func (p *mockProvisioner) GetWorkspaceInfo(ctx context.Context, w *workspace.Workspace, target *provider.ProviderTarget) (*workspace.WorkspaceInfo, error) {
	args := p.Called(ctx, w, target)
	return args.Get(0).(*workspace.WorkspaceInfo), args.Error(1)
}

func (p *mockProvisioner) StartProject(params provisioner.ProjectParams) error {
	args := p.Called(params)
	return args.Error(0)
}

func (p *mockProvisioner) StartWorkspace(workspace *workspace.Workspace, target *provider.ProviderTarget) error {
	args := p.Called(workspace, target)
	return args.Error(0)
}

func (p *mockProvisioner) StopProject(proj *project.Project, target *provider.ProviderTarget) error {
	args := p.Called(proj, target)
	return args.Error(0)
}

func (p *mockProvisioner) StopWorkspace(workspace *workspace.Workspace, target *provider.ProviderTarget) error {
	args := p.Called(workspace, target)
	return args.Error(0)
}
