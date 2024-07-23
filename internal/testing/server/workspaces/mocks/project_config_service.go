//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
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

func (m *mockProjectConfigService) Delete(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *mockProjectConfigService) Find(name string) (*config.ProjectConfig, error) {
	args := m.Called(name)
	return args.Get(0).(*config.ProjectConfig), args.Error(1)
}

func (m *mockProjectConfigService) FindDefault(url string) (*config.ProjectConfig, error) {
	args := m.Called(url)
	return args.Get(0).(*config.ProjectConfig), args.Error(1)
}

func (m *mockProjectConfigService) List() ([]*config.ProjectConfig, error) {
	args := m.Called()
	return args.Get(0).([]*config.ProjectConfig), args.Error(1)
}

func (m *mockProjectConfigService) FilterByGitUrl(url string) ([]*config.ProjectConfig, error) {
	args := m.Called(url)
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

func (m *mockProjectConfigService) ToProjectConfig(createProjectConfigDto dto.CreateProjectConfigDTO) *config.ProjectConfig {
	args := m.Called(createProjectConfigDto)
	return args.Get(0).(*config.ProjectConfig)
}
