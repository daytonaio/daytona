//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provisioner"
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

func (p *mockProvisioner) CreateTarget(t *models.Target) error {
	args := p.Called(t)
	return args.Error(0)
}

func (p *mockProvisioner) DestroyWorkspace(ws *models.Workspace) error {
	args := p.Called(ws)
	return args.Error(0)
}

func (p *mockProvisioner) DestroyTarget(t *models.Target) error {
	args := p.Called(t)
	return args.Error(0)
}

func (p *mockProvisioner) GetTargetInfo(ctx context.Context, t *models.Target) (*models.TargetInfo, error) {
	args := p.Called(ctx, t)
	return args.Get(0).(*models.TargetInfo), args.Error(1)
}

func (p *mockProvisioner) StartWorkspace(params provisioner.WorkspaceParams) error {
	args := p.Called(params)
	return args.Error(0)
}

func (p *mockProvisioner) StartTarget(t *models.Target) error {
	args := p.Called(t)
	return args.Error(0)
}

func (p *mockProvisioner) StopWorkspace(ws *models.Workspace) error {
	args := p.Called(ws)
	return args.Error(0)
}

func (p *mockProvisioner) StopTarget(t *models.Target) error {
	args := p.Called(t)
	return args.Error(0)
}

func (p *mockProvisioner) GetWorkspaceInfo(ctx context.Context, w *models.Workspace) (*models.WorkspaceInfo, error) {
	args := p.Called(ctx, w)
	return args.Get(0).(*models.WorkspaceInfo), args.Error(1)
}
