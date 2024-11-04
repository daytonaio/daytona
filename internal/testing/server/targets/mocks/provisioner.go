//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/gitprovider"
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

func (p *mockProvisioner) CreateWorkspace(ws *workspace.Workspace, t *target.Target, cr *containerregistry.ContainerRegistry, gc *gitprovider.GitProviderConfig) error {
	args := p.Called(ws, t, cr, gc)
	return args.Error(0)
}

func (p *mockProvisioner) CreateTarget(t *target.Target) error {
	args := p.Called(t)
	return args.Error(0)
}

func (p *mockProvisioner) DestroyWorkspace(ws *workspace.Workspace, t *target.Target) error {
	args := p.Called(ws, t)
	return args.Error(0)
}

func (p *mockProvisioner) DestroyTarget(t *target.Target) error {
	args := p.Called(t)
	return args.Error(0)
}

func (p *mockProvisioner) GetTargetInfo(ctx context.Context, t *target.Target) (*target.TargetInfo, error) {
	args := p.Called(ctx, t)
	return args.Get(0).(*target.TargetInfo), args.Error(1)
}

func (p *mockProvisioner) StartWorkspace(ws *workspace.Workspace, t *target.Target) error {
	args := p.Called(ws, t)
	return args.Error(0)
}

func (p *mockProvisioner) StartTarget(t *target.Target) error {
	args := p.Called(t)
	return args.Error(0)
}

func (p *mockProvisioner) StopWorkspace(ws *workspace.Workspace, t *target.Target) error {
	args := p.Called(ws, t)
	return args.Error(0)
}

func (p *mockProvisioner) StopTarget(t *target.Target) error {
	args := p.Called(t)
	return args.Error(0)
}

func (p *mockProvisioner) GetWorkspaceInfo(ctx context.Context, w *workspace.Workspace, t *target.Target) (*workspace.WorkspaceInfo, error) {
	args := p.Called(ctx, w, t)
	return args.Get(0).(*workspace.WorkspaceInfo), args.Error(1)
}
