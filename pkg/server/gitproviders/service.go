// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitproviders

import (
	"errors"
	"fmt"
	"strings"

	"github.com/daytonaio/daytona/pkg/gitprovider"
)

type IGitProviderService interface {
	GetConfig(id string) (*gitprovider.GitProviderConfig, error)
	GetConfigForUrl(url string) (*gitprovider.GitProviderConfig, error)
	GetGitProvider(id string) (gitprovider.GitProvider, error)
	GetGitProviderForUrl(url string) (gitprovider.GitProvider, string, error)
	GetGitUser(gitProviderId string) (*gitprovider.GitUser, error)
	GetNamespaces(gitProviderId string) ([]*gitprovider.GitNamespace, error)
	GetRepoBranches(gitProviderId string, namespaceId string, repositoryId string) ([]*gitprovider.GitBranch, error)
	GetRepoPRs(gitProviderId string, namespaceId string, repositoryId string) ([]*gitprovider.GitPullRequest, error)
	GetRepositories(gitProviderId string, namespaceId string) ([]*gitprovider.GitRepository, error)
	ListConfigs() ([]*gitprovider.GitProviderConfig, error)
	RemoveGitProvider(gitProviderId string) error
	SetGitProviderConfig(providerConfig *gitprovider.GitProviderConfig) error
	GetLastCommitSha(repo *gitprovider.GitRepository) (string, error)
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
		// If config is not defined, use the default (public) client without token
		if gitprovider.IsGitProviderNotFound(err) {
			providerConfig = &gitprovider.GitProviderConfig{
				Id:         id,
				Username:   "",
				Token:      "",
				BaseApiUrl: nil,
			}
		} else {
			return nil, err
		}
	}

	return s.newGitProvider(providerConfig)
}

func (s *GitProviderService) ListConfigs() ([]*gitprovider.GitProviderConfig, error) {
	return s.configStore.List()
}

func (s *GitProviderService) GetConfig(id string) (*gitprovider.GitProviderConfig, error) {
	return s.configStore.Find(id)
}

func (s *GitProviderService) GetLastCommitSha(repo *gitprovider.GitRepository) (string, error) {
	var err error
	var provider gitprovider.GitProvider
	providerFound := false

	gitProviders, err := s.configStore.List()
	if err != nil {
		return "", err
	}

	for _, p := range gitProviders {
		if strings.Contains(repo.Url, fmt.Sprintf("%s.", p.Id)) {
			provider, err = s.GetGitProvider(p.Id)
			if err == nil {
				return "", err
			}
			providerFound = true
			break
		}

		hostname, err := getHostnameFromUrl(*p.BaseApiUrl)
		if err != nil {
			return "", err
		}

		if p.BaseApiUrl != nil && strings.Contains(repo.Url, hostname) {
			provider, err = s.GetGitProvider(p.Id)
			if err == nil {
				return "", err
			}
			providerFound = true
			break
		}
	}

	if !providerFound {
		hostname := strings.TrimPrefix(repo.Source, "www.")
		providerId := strings.Split(hostname, ".")[0]

		provider, err = s.newGitProvider(&gitprovider.GitProviderConfig{
			Id:         providerId,
			Username:   "",
			Token:      "",
			BaseApiUrl: nil,
		})
		if err != nil {
			return "", err
		}
	}

	return provider.GetLastCommitSha(&gitprovider.StaticGitContext{
		Id:       repo.Id,
		Url:      repo.Url,
		Name:     repo.Name,
		Branch:   repo.Branch,
		Sha:      &repo.Sha,
		Owner:    repo.Owner,
		PrNumber: repo.PrNumber,
		Source:   repo.Source,
		Path:     repo.Path,
	})
}

func (s *GitProviderService) newGitProvider(config *gitprovider.GitProviderConfig) (gitprovider.GitProvider, error) {
	switch config.Id {
	case "github":
		return gitprovider.NewGitHubGitProvider(config.Token, nil), nil
	case "github-enterprise-server":
		return gitprovider.NewGitHubGitProvider(config.Token, config.BaseApiUrl), nil
	case "gitlab":
		return gitprovider.NewGitLabGitProvider(config.Token, nil), nil
	case "bitbucket":
		return gitprovider.NewBitbucketGitProvider(config.Username, config.Token), nil
	case "bitbucket-server":
		return gitprovider.NewBitbucketServerGitProvider(config.Username, config.Token, config.BaseApiUrl), nil
	case "gitlab-self-managed":
		return gitprovider.NewGitLabGitProvider(config.Token, config.BaseApiUrl), nil
	case "codeberg":
		return gitprovider.NewGiteaGitProvider(config.Token, codebergUrl), nil
	case "gitea":
		return gitprovider.NewGiteaGitProvider(config.Token, *config.BaseApiUrl), nil
	case "gitness":
		return gitprovider.NewGitnessGitProvider(config.Token, config.BaseApiUrl), nil
	case "azure-devops":
		return gitprovider.NewAzureDevOpsGitProvider(config.Token, *config.BaseApiUrl), nil
	default:
		return nil, errors.New("git provider not found")
	}
}
