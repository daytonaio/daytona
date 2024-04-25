//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/stretchr/testify/mock"
)

type mockGitProviderService struct {
	mock.Mock
}

func NewGitProviderService() *mockGitProviderService {
	return &mockGitProviderService{}
}

func (s *mockGitProviderService) GetConfig(id string) (*gitprovider.GitProviderConfig, error) {
	args := s.Called(id)
	return args.Get(0).(*gitprovider.GitProviderConfig), args.Error(1)
}

func (s *mockGitProviderService) GetConfigForUrl(url string) (*gitprovider.GitProviderConfig, error) {
	args := s.Called(url)
	return args.Get(0).(*gitprovider.GitProviderConfig), args.Error(1)
}

func (s *mockGitProviderService) GetGitProvider(id string) (gitprovider.GitProvider, error) {
	args := s.Called(id)
	return args.Get(0).(gitprovider.GitProvider), args.Error(1)
}

func (s *mockGitProviderService) GetGitProviderForUrl(url string) (gitprovider.GitProvider, error) {
	args := s.Called(url)
	return args.Get(0).(gitprovider.GitProvider), args.Error(1)
}

func (s *mockGitProviderService) GetGitUser(gitProviderId string) (*gitprovider.GitUser, error) {
	args := s.Called(gitProviderId)
	return args.Get(0).(*gitprovider.GitUser), args.Error(1)
}

func (s *mockGitProviderService) GetNamespaces(gitProviderId string) ([]*gitprovider.GitNamespace, error) {
	args := s.Called(gitProviderId)
	return args.Get(0).([]*gitprovider.GitNamespace), args.Error(1)
}

func (s *mockGitProviderService) GetRepoBranches(gitProviderId string, namespaceId string, repositoryId string) ([]*gitprovider.GitBranch, error) {
	args := s.Called(gitProviderId, namespaceId, repositoryId)
	return args.Get(0).([]*gitprovider.GitBranch), args.Error(1)
}

func (s *mockGitProviderService) GetRepoPRs(gitProviderId string, namespaceId string, repositoryId string) ([]*gitprovider.GitPullRequest, error) {
	args := s.Called(gitProviderId, namespaceId, repositoryId)
	return args.Get(0).([]*gitprovider.GitPullRequest), args.Error(1)
}

func (s *mockGitProviderService) GetRepositories(gitProviderId string, namespaceId string) ([]*gitprovider.GitRepository, error) {
	args := s.Called(gitProviderId, namespaceId)
	return args.Get(0).([]*gitprovider.GitRepository), args.Error(1)
}

func (s *mockGitProviderService) ListConfigs() ([]*gitprovider.GitProviderConfig, error) {
	args := s.Called()
	return args.Get(0).([]*gitprovider.GitProviderConfig), args.Error(1)
}

func (s *mockGitProviderService) RemoveGitProvider(gitProviderId string) error {
	args := s.Called(gitProviderId)
	return args.Error(0)
}

func (s *mockGitProviderService) SetGitProviderConfig(providerConfig *gitprovider.GitProviderConfig) error {
	args := s.Called(providerConfig)
	return args.Error(0)
}
