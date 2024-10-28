//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/server/workspaceconfig/dto"
	"github.com/daytonaio/daytona/pkg/target/workspace/config"
	"github.com/stretchr/testify/mock"
)

type mockWorkspaceConfigService struct {
	mock.Mock
}

func NewMockWorkspaceConfigService() *mockWorkspaceConfigService {
	return &mockWorkspaceConfigService{}
}

func (m *mockWorkspaceConfigService) Delete(name string, force bool) []error {
	args := m.Called(name, force)
	return args.Get(0).([]error)
}

func (m *mockWorkspaceConfigService) Find(filter *config.WorkspaceConfigFilter) (*config.WorkspaceConfig, error) {
	args := m.Called(filter)
	return args.Get(0).(*config.WorkspaceConfig), args.Error(1)
}

func (m *mockWorkspaceConfigService) List(filter *config.WorkspaceConfigFilter) ([]*config.WorkspaceConfig, error) {
	args := m.Called(filter)
	return args.Get(0).([]*config.WorkspaceConfig), args.Error(1)
}

func (m *mockWorkspaceConfigService) SetDefault(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *mockWorkspaceConfigService) Save(wc *config.WorkspaceConfig) error {
	args := m.Called(wc)
	return args.Error(0)
}

func (m *mockWorkspaceConfigService) SetPrebuild(workspaceConfigName string, createPrebuildDto dto.CreatePrebuildDTO) (*dto.PrebuildDTO, error) {
	args := m.Called(workspaceConfigName, createPrebuildDto)
	return args.Get(0).(*dto.PrebuildDTO), args.Error(1)
}

func (m *mockWorkspaceConfigService) FindPrebuild(workspaceConfigFilter *config.WorkspaceConfigFilter, prebuildFilter *config.PrebuildFilter) (*dto.PrebuildDTO, error) {
	args := m.Called(workspaceConfigFilter, prebuildFilter)
	return args.Get(0).(*dto.PrebuildDTO), args.Error(1)
}

func (m *mockWorkspaceConfigService) ListPrebuilds(workspaceConfigFilter *config.WorkspaceConfigFilter, prebuildFilter *config.PrebuildFilter) ([]*dto.PrebuildDTO, error) {
	args := m.Called(workspaceConfigFilter, prebuildFilter)
	return args.Get(0).([]*dto.PrebuildDTO), args.Error(1)
}

func (m *mockWorkspaceConfigService) DeletePrebuild(workspaceConfigName string, id string, force bool) []error {
	args := m.Called(workspaceConfigName, id, force)
	return args.Get(0).([]error)
}

func (m *mockWorkspaceConfigService) StartRetentionPoller() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockWorkspaceConfigService) EnforceRetentionPolicy() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockWorkspaceConfigService) ProcessGitEvent(data gitprovider.GitEventData) error {
	args := m.Called(data)
	return args.Error(0)
}
