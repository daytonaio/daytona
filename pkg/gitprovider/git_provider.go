// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/types"
)

const personalNamespaceId = "<PERSONAL>"

type GitProvider interface {
	GetNamespaces() ([]types.GitNamespace, error)
	GetRepositories(namespace string) ([]types.GitRepository, error)
	GetUser() (types.GitUser, error)
	GetRepoBranches(repositoryId string, namespaceId string) ([]types.GitBranch, error)
	GetRepoPRs(repositoryId string, namespaceId string) ([]types.GitPullRequest, error)
	// ParseGitUrl(string) (*types.GitRepository, error)
}

func GetGitProvider(providerId string, gitProviders []types.GitProvider) GitProvider {
	var chosenProvider *types.GitProvider
	for _, gitProvider := range gitProviders {
		if gitProvider.Id == providerId {
			chosenProvider = &gitProvider
			break
		}
	}

	if chosenProvider == nil {
		return nil
	}

	switch providerId {
	case "github":
		return &GitHubGitProvider{
			token: chosenProvider.Token,
		}
	case "gitlab":
		return &GitLabGitProvider{
			token: chosenProvider.Token,
		}
	case "bitbucket":
		return &BitbucketGitProvider{
			username: chosenProvider.Username,
			token:    chosenProvider.Token,
		}
	case "gitlab-self-managed":
		return &GitLabGitProvider{
			token:      chosenProvider.Token,
			baseApiUrl: chosenProvider.BaseApiUrl,
		}
	case "codeberg":
		return &GiteaGitProvider{
			token:      chosenProvider.Token,
			baseApiUrl: "https://codeberg.org",
		}
	case "gitea":
		return &GiteaGitProvider{
			token:      chosenProvider.Token,
			baseApiUrl: chosenProvider.BaseApiUrl,
		}
	default:
		return nil
	}
}

func GetUsernameFromToken(providerId string, token string, baseApiUrl string) (string, error) {
	var gitProvider GitProvider

	switch providerId {
	case "github":
		gitProvider = &GitHubGitProvider{
			token: token,
		}
	case "gitlab":
		fallthrough
	case "gitlab-self-managed":
		gitProvider = &GitLabGitProvider{
			token:      token,
			baseApiUrl: baseApiUrl,
		}
	case "bitbucket":
		gitProvider = &BitbucketGitProvider{
			token: token,
		}
	case "codeberg":
		gitProvider = &GiteaGitProvider{
			token:      token,
			baseApiUrl: "https://codeberg.org",
		}
	case "gitea":
		gitProvider = &GiteaGitProvider{
			token:      token,
			baseApiUrl: baseApiUrl,
		}
	default:
		return "", errors.New("provider not found")
	}

	gitUser, err := gitProvider.GetUser()
	if err != nil {
		return "", errors.New("user not found")
	}

	return gitUser.Username, nil
}
