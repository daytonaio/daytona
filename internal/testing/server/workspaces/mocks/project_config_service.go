//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/server/projectconfig/dto"
	"github.com/daytonaio/daytona/pkg/workspace/project/config"
	"github.com/stretchr/testify/mock"
)

type mockProjectConfigService struct {
	mock.Mock
}

func NewMockProjectConfigService() *mockProjectConfigService {
	return &mockProjectConfigService{}
}

func (m *mockProjectConfigService) Delete(name string, force bool) []error {
	args := m.Called(name, force)
	return args.Get(0).([]error)
}

func (m *mockProjectConfigService) Find(filter *config.ProjectConfigFilter) (*config.ProjectConfig, error) {
	args := m.Called(filter)
	return args.Get(0).(*config.ProjectConfig), args.Error(1)
}

func (m *mockProjectConfigService) List(filter *config.ProjectConfigFilter) ([]*config.ProjectConfig, error) {
	args := m.Called(filter)
	return args.Get(0).([]*config.ProjectConfig), args.Error(1)
}

func (m *mockProjectConfigService) SetDefault(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *mockProjectConfigService) Save(pc *config.ProjectConfig) error {
	args := m.Called(pc)
	return args.Error(0)
}

func (m *mockProjectConfigService) SetPrebuild(projectConfigName string, createProjectDto dto.CreatePrebuildDTO) (*dto.PrebuildDTO, error) {
	args := m.Called(projectConfigName, createProjectDto)
	return args.Get(0).(*dto.PrebuildDTO), args.Error(1)
}

func (m *mockProjectConfigService) FindPrebuild(projectConfigFilter *config.ProjectConfigFilter, prebuildFilter *config.PrebuildFilter) (*dto.PrebuildDTO, error) {
	args := m.Called(projectConfigFilter, prebuildFilter)
	return args.Get(0).(*dto.PrebuildDTO), args.Error(1)
}

func (m *mockProjectConfigService) ListPrebuilds(projectConfigFilter *config.ProjectConfigFilter, prebuildFilter *config.PrebuildFilter) ([]*dto.PrebuildDTO, error) {
	args := m.Called(projectConfigFilter, prebuildFilter)
	return args.Get(0).([]*dto.PrebuildDTO), args.Error(1)
}

func (m *mockProjectConfigService) DeletePrebuild(projectConfigName string, id string, force bool) []error {
	args := m.Called(projectConfigName, id, force)
	return args.Get(0).([]error)
}

func (m *mockProjectConfigService) StartRetentionPoller() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockProjectConfigService) EnforceRetentionPolicy() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockProjectConfigService) ProcessGitEvent(data gitprovider.GitEventData) error {
	args := m.Called(data)
	return args.Error(0)
}
