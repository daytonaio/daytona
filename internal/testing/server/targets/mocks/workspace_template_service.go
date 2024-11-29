//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/workspacetemplates/dto"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/stretchr/testify/mock"
)

type mockWorkspaceTemplateService struct {
	mock.Mock
}

func NewMockWorkspaceTemplateService() *mockWorkspaceTemplateService {
	return &mockWorkspaceTemplateService{}
}

func (m *mockWorkspaceTemplateService) Delete(name string, force bool) []error {
	args := m.Called(name, force)
	return args.Get(0).([]error)
}

func (m *mockWorkspaceTemplateService) Find(filter *stores.WorkspaceTemplateFilter) (*models.WorkspaceTemplate, error) {
	args := m.Called(filter)
	return args.Get(0).(*models.WorkspaceTemplate), args.Error(1)
}

func (m *mockWorkspaceTemplateService) List(filter *stores.WorkspaceTemplateFilter) ([]*models.WorkspaceTemplate, error) {
	args := m.Called(filter)
	return args.Get(0).([]*models.WorkspaceTemplate), args.Error(1)
}

func (m *mockWorkspaceTemplateService) SetDefault(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *mockWorkspaceTemplateService) Save(wt *models.WorkspaceTemplate) error {
	args := m.Called(wt)
	return args.Error(0)
}

func (m *mockWorkspaceTemplateService) SetPrebuild(workspaceTemplateName string, createPrebuildDto dto.CreatePrebuildDTO) (*dto.PrebuildDTO, error) {
	args := m.Called(workspaceTemplateName, createPrebuildDto)
	return args.Get(0).(*dto.PrebuildDTO), args.Error(1)
}

func (m *mockWorkspaceTemplateService) FindPrebuild(workspaceTemplateFilter *stores.WorkspaceTemplateFilter, prebuildFilter *stores.PrebuildFilter) (*dto.PrebuildDTO, error) {
	args := m.Called(workspaceTemplateFilter, prebuildFilter)
	return args.Get(0).(*dto.PrebuildDTO), args.Error(1)
}

func (m *mockWorkspaceTemplateService) ListPrebuilds(workspaceTemplateFilter *stores.WorkspaceTemplateFilter, prebuildFilter *stores.PrebuildFilter) ([]*dto.PrebuildDTO, error) {
	args := m.Called(workspaceTemplateFilter, prebuildFilter)
	return args.Get(0).([]*dto.PrebuildDTO), args.Error(1)
}

func (m *mockWorkspaceTemplateService) DeletePrebuild(workspaceTemplateName string, id string, force bool) []error {
	args := m.Called(workspaceTemplateName, id, force)
	return args.Get(0).([]error)
}

func (m *mockWorkspaceTemplateService) StartRetentionPoller() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockWorkspaceTemplateService) EnforceRetentionPolicy() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockWorkspaceTemplateService) ProcessGitEvent(data gitprovider.GitEventData) error {
	args := m.Called(data)
	return args.Error(0)
}
