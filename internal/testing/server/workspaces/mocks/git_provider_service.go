//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/stretchr/testify/mock"
)

type MockGitProviderService struct {
	mock.Mock
}

func NewMockGitProviderService() *MockGitProviderService {
	return &MockGitProviderService{}
}

func (m *MockGitProviderService) GetConfig(id string) (*gitprovider.GitProviderConfig, error) {
	args := m.Called(id)
	return args.Get(0).(*gitprovider.GitProviderConfig), args.Error(1)
}

func (m *MockGitProviderService) GetConfigForUrl(url string) (*gitprovider.GitProviderConfig, error) {
	args := m.Called(url)
	return args.Get(0).(*gitprovider.GitProviderConfig), args.Error(1)
}

func (m *MockGitProviderService) GetGitProvider(id string) (gitprovider.GitProvider, error) {
	args := m.Called(id)
	return args.Get(0).(gitprovider.GitProvider), args.Error(1)
}

func (m *MockGitProviderService) GetGitProviderForUrl(url string) (gitprovider.GitProvider, string, error) {
	args := m.Called(url)
	return args.Get(0).(gitprovider.GitProvider), args.String(0), args.Error(1)
}

func (m *MockGitProviderService) GetGitUser(gitProviderId string) (*gitprovider.GitUser, error) {
	args := m.Called(gitProviderId)
	return args.Get(0).(*gitprovider.GitUser), args.Error(1)
}

func (m *MockGitProviderService) GetNamespaces(gitProviderId string) ([]*gitprovider.GitNamespace, error) {
	args := m.Called(gitProviderId)
	return args.Get(0).([]*gitprovider.GitNamespace), args.Error(1)
}

func (m *MockGitProviderService) GetRepoBranches(gitProviderId string, namespaceId string, repositoryId string) ([]*gitprovider.GitBranch, error) {
	args := m.Called(gitProviderId, namespaceId, repositoryId)
	return args.Get(0).([]*gitprovider.GitBranch), args.Error(1)
}

func (m *MockGitProviderService) GetRepoPRs(gitProviderId string, namespaceId string, repositoryId string) ([]*gitprovider.GitPullRequest, error) {
	args := m.Called(gitProviderId, namespaceId, repositoryId)
	return args.Get(0).([]*gitprovider.GitPullRequest), args.Error(1)
}

func (m *MockGitProviderService) GetRepositories(gitProviderId string, namespaceId string) ([]*gitprovider.GitRepository, error) {
	args := m.Called(gitProviderId, namespaceId)
	return args.Get(0).([]*gitprovider.GitRepository), args.Error(1)
}

func (m *MockGitProviderService) ListConfigs() ([]*gitprovider.GitProviderConfig, error) {
	args := m.Called()
	return args.Get(0).([]*gitprovider.GitProviderConfig), args.Error(1)
}

func (m *MockGitProviderService) RemoveGitProvider(gitProviderId string) error {
	args := m.Called(gitProviderId)
	return args.Error(0)
}

func (m *MockGitProviderService) SetGitProviderConfig(providerConfig *gitprovider.GitProviderConfig) error {
	args := m.Called(providerConfig)
	return args.Error(0)
}

func (m *MockGitProviderService) GetLastCommitSha(repo *gitprovider.GitRepository) (string, error) {
	args := m.Called(repo)
	return args.String(0), args.Error(1)
}
