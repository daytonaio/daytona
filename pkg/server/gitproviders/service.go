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

	switch id {
	case "github":
		return gitprovider.NewGitHubGitProvider(providerConfig.Token), nil
	case "gitlab":
		return gitprovider.NewGitLabGitProvider(providerConfig.Token, nil), nil
	case "bitbucket":
		return gitprovider.NewBitbucketGitProvider(providerConfig.Username, providerConfig.Token), nil
	case "gitlab-self-managed":
		return gitprovider.NewGitLabGitProvider(providerConfig.Token, providerConfig.BaseApiUrl), nil
	case "codeberg":
		return gitprovider.NewGiteaGitProvider(providerConfig.Token, codebergUrl), nil
	case "gitea":
		return gitprovider.NewGiteaGitProvider(providerConfig.Token, *providerConfig.BaseApiUrl), nil
	default:
		return nil, errors.New("git provider not found")
	}
}

// TODO: ON save
// func GetUsernameFromToken(providerId string, token string, baseApiUrl *string) (string, error) {
// 	var gitProvider gitprovider.GitProvider

// 	switch providerId {
// 	case "github":
// 		gitProvider = gitprovider.NewGitHubGitProvider(token)
// 	case "gitlab":
// 		fallthrough
// 	case "gitlab-self-managed":
// 		gitProvider = gitprovider.NewGitLabGitProvider(token, baseApiUrl)
// 	case "bitbucket":
// 		gitProvider = gitprovider.NewBitbucketGitProvider("", token)
// 	case "codeberg":
// 		gitProvider = gitprovider.NewGiteaGitProvider(token, codebergUrl)
// 	case "gitea":
// 		gitProvider = &GiteaGitProvider{
// 			token:      token,
// 			baseApiUrl: baseApiUrl,
// 		}
// 	default:
// 		return "", errors.New("provider not found")
// 	}

// 	gitUser, err := gitProvider.GetUser()
// 	if err != nil {
// 		return "", errors.New("user not found")
// 	}

// 	return gitUser.Username, nil
// }
