//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/daytonaio/daytona/internal/util/apiclient/server/conversion"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/stretchr/testify/mock"
)

type mockGitService struct {
	mock.Mock
}

func (m *mockGitService) CloneRepository(project *serverapiclient.Project, auth *http.BasicAuth) error {
	args := m.Called(project, auth)
	return args.Error(0)
}

func (m *mockGitService) RepositoryExists(project *serverapiclient.Project) (bool, error) {
	args := m.Called(project)
	return args.Bool(0), args.Error(1)
}

func (m *mockGitService) SetGitConfig(userData *serverapiclient.GitUser) error {
	args := m.Called(userData)
	return args.Error(0)
}

func NewMockGitService(repositoryShouldExist bool, project *workspace.Project) *mockGitService {
	gitService := new(mockGitService)
	gitService.On("RepositoryExists", conversion.ToProjectDTO(project)).Return(repositoryShouldExist, nil)
	if !repositoryShouldExist {
		gitService.On("CloneRepository", project, mock.Anything).Return(nil)
	}
	gitService.On("SetGitConfig", mock.Anything).Return(nil)

	return gitService
}

func NewHostModeMockGitService() *mockGitService {
	gitService := new(mockGitService)
	return gitService
}
