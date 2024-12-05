// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"
	"net/http"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/models"
)

type IGitProviderService interface {
	GetConfig(ctx context.Context, id string) (*models.GitProviderConfig, error)
	ListConfigsForUrl(ctx context.Context, url string) ([]*models.GitProviderConfig, error)
	GetGitProvider(ctx context.Context, id string) (gitprovider.GitProvider, error)
	GetGitProviderForUrl(ctx context.Context, url string) (gitprovider.GitProvider, string, error)
	GetGitProviderForHttpRequest(ctx context.Context, req *http.Request) (gitprovider.GitProvider, error)
	GetGitUser(ctx context.Context, gitProviderId string) (*gitprovider.GitUser, error)
	GetNamespaces(ctx context.Context, gitProviderId string, options gitprovider.ListOptions) ([]*gitprovider.GitNamespace, error)
	GetRepoBranches(ctx context.Context, gitProviderId string, namespaceId string, repositoryId string, options gitprovider.ListOptions) ([]*gitprovider.GitBranch, error)
	GetRepoPRs(ctx context.Context, gitProviderId string, namespaceId string, repositoryId string, options gitprovider.ListOptions) ([]*gitprovider.GitPullRequest, error)
	GetRepositories(ctx context.Context, gitProviderId string, namespaceId string, options gitprovider.ListOptions) ([]*gitprovider.GitRepository, error)
	ListConfigs(ctx context.Context) ([]*models.GitProviderConfig, error)
	RemoveGitProvider(ctx context.Context, gitProviderId string) error
	SetGitProviderConfig(ctx context.Context, providerConfig *models.GitProviderConfig) error
	GetLastCommitSha(ctx context.Context, repo *gitprovider.GitRepository) (string, error)
	RegisterPrebuildWebhook(ctx context.Context, gitProviderId string, repo *gitprovider.GitRepository, endpointUrl string) (string, error)
	GetPrebuildWebhook(ctx context.Context, gitProviderId string, repo *gitprovider.GitRepository, endpointUrl string) (*string, error)
	UnregisterPrebuildWebhook(ctx context.Context, gitProviderId string, repo *gitprovider.GitRepository, id string) error
}
