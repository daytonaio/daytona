// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"errors"
)

const personalNamespaceId = "<PERSONAL>"

type gitProvider interface {
	GetNamespaces() ([]GitNamespace, error)
	GetRepositories(namespace string) ([]GitRepository, error)
	GetUser() (GitUser, error)
	GetRepoBranches(repositoryId string, namespaceId string) ([]GitBranch, error)
	GetRepoPRs(repositoryId string, namespaceId string) ([]GitPullRequest, error)
	// ParseGitUrl(string) (*GitRepository, error)
}

func GetGitProvider(providerId string, gitProviders []GitProvider) gitProvider {
	var chosenProvider *GitProvider
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
	var gitProvider gitProvider

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
