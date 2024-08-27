//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/workspace/project"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/stretchr/testify/mock"
)

type MockGitService struct {
	mock.Mock
}

func (m *MockGitService) CloneRepository(p *project.Project, auth *http.BasicAuth) error {
	args := m.Called(p, auth)
	return args.Error(0)
}

func (m *MockGitService) CloneRepositoryCmd(p *project.Project, auth *http.BasicAuth) []string {
	args := m.Called(p, auth)
	return args.Get(0).([]string)
}

func (m *MockGitService) RepositoryExists(p *project.Project) (bool, error) {
	args := m.Called(p)
	return args.Bool(0), args.Error(1)
}

func (m *MockGitService) SetGitConfig(userData *gitprovider.GitUser) error {
	args := m.Called(userData)
	return args.Error(0)
}

func (m *MockGitService) GetGitStatus() (*project.GitStatus, error) {
	args := m.Called()
	return args.Get(0).(*project.GitStatus), args.Error(1)
}

func NewMockGitService() *MockGitService {
	gitService := new(MockGitService)
	return gitService
}
