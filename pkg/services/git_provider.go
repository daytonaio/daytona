// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"net/http"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/models"
)

type IGitProviderService interface {
	GetConfig(id string) (*models.GitProviderConfig, error)
	ListConfigsForUrl(url string) ([]*models.GitProviderConfig, error)
	GetGitProvider(id string) (gitprovider.GitProvider, error)
	GetGitProviderForUrl(url string) (gitprovider.GitProvider, string, error)
	GetGitProviderForHttpRequest(req *http.Request) (gitprovider.GitProvider, error)
	GetGitUser(gitProviderId string) (*gitprovider.GitUser, error)
	GetNamespaces(gitProviderId string, options gitprovider.ListOptions) ([]*gitprovider.GitNamespace, error)
	GetRepoBranches(gitProviderId string, namespaceId string, repositoryId string, options gitprovider.ListOptions) ([]*gitprovider.GitBranch, error)
	GetRepoPRs(gitProviderId string, namespaceId string, repositoryId string, options gitprovider.ListOptions) ([]*gitprovider.GitPullRequest, error)
	GetRepositories(gitProviderId string, namespaceId string, options gitprovider.ListOptions) ([]*gitprovider.GitRepository, error)
	ListConfigs() ([]*models.GitProviderConfig, error)
	RemoveGitProvider(gitProviderId string) error
	SetGitProviderConfig(providerConfig *models.GitProviderConfig) error
	GetLastCommitSha(repo *gitprovider.GitRepository) (string, error)
	RegisterPrebuildWebhook(gitProviderId string, repo *gitprovider.GitRepository, endpointUrl string) (string, error)
	GetPrebuildWebhook(gitProviderId string, repo *gitprovider.GitRepository, endpointUrl string) (*string, error)
	UnregisterPrebuildWebhook(gitProviderId string, repo *gitprovider.GitRepository, id string) error
}
