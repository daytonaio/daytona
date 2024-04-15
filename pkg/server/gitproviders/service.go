// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitproviders

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/gitprovider"
)

type IGitProviderService interface {
	GetConfig(id string) (*gitprovider.GitProviderConfig, error)
	GetConfigForUrl(url string) (*gitprovider.GitProviderConfig, error)
	GetGitProvider(id string) (gitprovider.GitProvider, error)
	GetGitProviderForUrl(url string) (gitprovider.GitProvider, error)
	GetGitUser(gitProviderId string) (*gitprovider.GitUser, error)
	GetNamespaces(gitProviderId string) ([]*gitprovider.GitNamespace, error)
	GetRepoBranches(gitProviderId string, namespaceId string, repositoryId string) ([]*gitprovider.GitBranch, error)
	GetRepoPRs(gitProviderId string, namespaceId string, repositoryId string) ([]*gitprovider.GitPullRequest, error)
	GetRepositories(gitProviderId string, namespaceId string) ([]*gitprovider.GitRepository, error)
	ListConfigs() ([]*gitprovider.GitProviderConfig, error)
	RemoveGitProvider(gitProviderId string) error
	SetGitProviderConfig(providerConfig *gitprovider.GitProviderConfig) error
}

type GitProviderServiceConfig struct {
	ConfigStore gitprovider.ConfigStore
}

type GitProviderService struct {
	configStore gitprovider.ConfigStore
}

func NewGitProviderService(config GitProviderServiceConfig) IGitProviderService {
	return &GitProviderService{
		configStore: config.ConfigStore,
	}
}

var codebergUrl = "https://codeberg.org"

func (s *GitProviderService) GetGitProvider(id string) (gitprovider.GitProvider, error) {
	providerConfig, err := s.configStore.Find(id)
	if err != nil {
		return nil, err
	}

	return s.newGitProvider(providerConfig)
}

func (s *GitProviderService) ListConfigs() ([]*gitprovider.GitProviderConfig, error) {
	return s.configStore.List()
}

func (s *GitProviderService) GetConfig(id string) (*gitprovider.GitProviderConfig, error) {
	return s.configStore.Find(id)
}

func (s *GitProviderService) newGitProvider(config *gitprovider.GitProviderConfig) (gitprovider.GitProvider, error) {
	switch config.Id {
	case "github":
		return gitprovider.NewGitHubGitProvider(config.Token), nil
	case "gitlab":
		return gitprovider.NewGitLabGitProvider(config.Token, nil), nil
	case "bitbucket":
		return gitprovider.NewBitbucketGitProvider(config.Username, config.Token), nil
	case "gitlab-self-managed":
		return gitprovider.NewGitLabGitProvider(config.Token, config.BaseApiUrl), nil
	case "codeberg":
		return gitprovider.NewGiteaGitProvider(config.Token, codebergUrl), nil
	case "gitea":
		return gitprovider.NewGiteaGitProvider(config.Token, *config.BaseApiUrl), nil
	default:
		return nil, errors.New("git provider not found")
	}
}
