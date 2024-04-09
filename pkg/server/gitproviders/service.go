// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitproviders

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/gitprovider"
)

type GitProviderServiceConfig struct {
	ConfigStore gitprovider.ConfigStore
}

type GitProviderService struct {
	configStore gitprovider.ConfigStore
}

func NewGitProviderService(config GitProviderServiceConfig) *GitProviderService {
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
