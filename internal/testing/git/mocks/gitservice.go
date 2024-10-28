//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/stretchr/testify/mock"
)

type MockGitService struct {
	mock.Mock
}

func (m *MockGitService) CloneRepository(repo *gitprovider.GitRepository, auth *http.BasicAuth) error {
	args := m.Called(repo, auth)
	return args.Error(0)
}

func (m *MockGitService) CloneRepositoryCmd(repo *gitprovider.GitRepository, auth *http.BasicAuth) []string {
	args := m.Called(repo, auth)
	return args.Get(0).([]string)
}

func (m *MockGitService) RepositoryExists() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockGitService) SetGitConfig(userData *gitprovider.GitUser, providerConfig *gitprovider.GitProviderConfig) error {
	args := m.Called(userData, providerConfig)
	return args.Error(0)
}

func (m *MockGitService) GetGitStatus() (*workspace.GitStatus, error) {
	args := m.Called()
	return args.Get(0).(*workspace.GitStatus), args.Error(1)
}

func NewMockGitService() *MockGitService {
	gitService := new(MockGitService)
	return gitService
}
