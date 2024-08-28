//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"net/http"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/stretchr/testify/mock"
)

type MockGitProvider struct {
	mock.Mock
}

func (m *MockGitProvider) GetNamespaces() ([]*gitprovider.GitNamespace, error) {
	args := m.Called()
	return args.Get(0).([]*gitprovider.GitNamespace), args.Error(1)
}

func (m *MockGitProvider) GetRepositories(namespace string) ([]*gitprovider.GitRepository, error) {
	args := m.Called(namespace)
	return args.Get(0).([]*gitprovider.GitRepository), args.Error(1)
}

func (m *MockGitProvider) GetUser() (*gitprovider.GitUser, error) {
	args := m.Called()
	return args.Get(0).(*gitprovider.GitUser), args.Error(1)
}

func (m *MockGitProvider) GetBranchByCommit(staticContext *gitprovider.StaticGitContext) (string, error) {
	args := m.Called(staticContext)
	return args.String(0), args.Error(1)
}

func (m *MockGitProvider) GetRepoBranches(repositoryId string, namespaceId string) ([]*gitprovider.GitBranch, error) {
	args := m.Called(repositoryId, namespaceId)
	return args.Get(0).([]*gitprovider.GitBranch), args.Error(1)
}

func (m *MockGitProvider) GetRepoPRs(repositoryId string, namespaceId string) ([]*gitprovider.GitPullRequest, error) {
	args := m.Called(repositoryId, namespaceId)
	return args.Get(0).([]*gitprovider.GitPullRequest), args.Error(1)
}

func (m *MockGitProvider) GetRepositoryContext(repoContext gitprovider.GetRepositoryContext) (*gitprovider.GitRepository, error) {
	args := m.Called(repoContext)
	return args.Get(0).(*gitprovider.GitRepository), args.Error(1)
}

func (m *MockGitProvider) GetUrlFromContext(repoContext *gitprovider.GetRepositoryContext) string {
	args := m.Called(repoContext)
	return args.String(0)
}

func (m *MockGitProvider) GetLastCommitSha(staticContext *gitprovider.StaticGitContext) (string, error) {
	args := m.Called(staticContext)
	return args.String(0), args.Error(1)
}

func (m *MockGitProvider) GetPrContext(staticContext *gitprovider.StaticGitContext) (*gitprovider.StaticGitContext, error) {
	args := m.Called(staticContext)
	return args.Get(0).(*gitprovider.StaticGitContext), args.Error(1)
}

func (m *MockGitProvider) ParseStaticGitContext(repoUrl string) (*gitprovider.StaticGitContext, error) {
	args := m.Called(repoUrl)
	return args.Get(0).(*gitprovider.StaticGitContext), args.Error(1)
}

func (m *MockGitProvider) GetDefaultBranch(staticContext *gitprovider.StaticGitContext) (*string, error) {
	args := m.Called(staticContext)
	return args.Get(0).(*string), args.Error(1)
}

func (m *MockGitProvider) RegisterPrebuildWebhook(repo *gitprovider.GitRepository, endpointUrl string) (string, error) {
	args := m.Called(repo, endpointUrl)
	return args.String(0), args.Error(1)
}

func (m *MockGitProvider) GetPrebuildWebhook(repo *gitprovider.GitRepository, endpointUrl string) (*string, error) {
	args := m.Called(repo, endpointUrl)
	return args.Get(0).(*string), args.Error(1)
}

func (m *MockGitProvider) UnregisterPrebuildWebhook(repo *gitprovider.GitRepository, id string) error {
	args := m.Called(repo, id)
	return args.Error(0)
}

func (m *MockGitProvider) GetCommitsRange(repo *gitprovider.GitRepository, owner string, initialSha string, currentSha string) (int, error) {
	args := m.Called(repo, owner, initialSha, currentSha)
	return args.Int(0), args.Error(1)
}

func (m *MockGitProvider) ParseEventData(request *http.Request) (*gitprovider.GitEventData, error) {
	args := m.Called(request)
	return args.Get(0).(*gitprovider.GitEventData), args.Error(1)
}
