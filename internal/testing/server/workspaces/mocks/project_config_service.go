//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/daytonaio/daytona/pkg/server/projectconfig/prebuild/dto"
	"github.com/daytonaio/daytona/pkg/workspace/project/config"
	"github.com/daytonaio/daytona/pkg/workspace/project/config/prebuild"
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

func (m *mockProjectConfigService) Find(filter *config.Filter) (*config.ProjectConfig, error) {
	args := m.Called(filter)
	return args.Get(0).(*config.ProjectConfig), args.Error(1)
}

func (m *mockProjectConfigService) List(filter *config.Filter) ([]*config.ProjectConfig, error) {
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

func (m *mockProjectConfigService) SetPrebuild(createProjectDto dto.CreatePrebuildDTO) error {
	args := m.Called(createProjectDto)
	return args.Error(0)
}

func (m *mockProjectConfigService) FindPrebuild(projectConfigName, id string) (*prebuild.PrebuildConfig, error) {
	args := m.Called(projectConfigName, id)
	return args.Get(0).(*prebuild.PrebuildConfig), args.Error(1)
}

func (m *mockProjectConfigService) ListPrebuilds(*config.PrebuildFilter) ([]*dto.PrebuildDTO, error) {
	args := m.Called()
	return args.Get(0).([]*dto.PrebuildDTO), args.Error(1)
}

func (m *mockProjectConfigService) DeletePrebuild(projectConfigName string, prebuild *prebuild.PrebuildConfig) error {
	args := m.Called(projectConfigName, prebuild)
	return args.Error(0)
}
