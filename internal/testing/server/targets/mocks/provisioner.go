//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/target/workspace"
	"github.com/stretchr/testify/mock"
)

type mockProvisioner struct {
	mock.Mock
}

func NewMockProvisioner() *mockProvisioner {
	return &mockProvisioner{}
}

func (p *mockProvisioner) CreateWorkspace(params provisioner.WorkspaceParams) error {
	args := p.Called(params)
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

func (p *mockProvisioner) StartWorkspace(params provisioner.WorkspaceParams) error {
	args := p.Called(params)
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
