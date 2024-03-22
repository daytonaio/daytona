// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"errors"
	"fmt"
	"strings"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/types"
)

const personalNamespaceId = "<PERSONAL>"

type GitProvider interface {
	GetNamespaces() ([]GitNamespace, error)
	GetRepositories(namespace string) ([]types.Repository, error)
	GetUserData() (GitUser, error)
	GetRepoBranches(types.Repository, string) ([]GitBranch, error)
	GetRepoPRs(types.Repository, string) ([]GitPullRequest, error)
	// ParseGitUrl(string) (*types.Repository, error)
}

type GitUser struct {
	Id       string
	Username string
	Name     string
	Email    string
}

type GitNamespace struct {
	Id   string
	Name string
}

type GitBranch struct {
	Name string
	SHA  string
}

type GitPullRequest struct {
	Name   string
	Branch string
}

type CheckoutOption struct {
	Title string
	Id    string
}

var (
	CheckoutDefault = CheckoutOption{Title: "Clone the default branch", Id: "default"}
	CheckoutBranch  = CheckoutOption{Title: "Branches", Id: "branch"}
	CheckoutPR      = CheckoutOption{Title: "Pull/Merge requests", Id: "pullrequest"}
)

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

func GetUsernameFromToken(providerId string, gitProviders []config.GitProvider, token string, baseApiUrl string) (string, error) {
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

	gitUser, err := gitProvider.GetUserData()
	if err != nil {
		return "", errors.New("user not found")
	}

	return gitUser.Username, nil
}

func GetGitProviderFromHost(url string, gitProviders []serverapiclient.GitProvider) *serverapiclient.GitProvider {
	for _, gitProvider := range gitProviders {
		if strings.Contains(url, fmt.Sprintf("%s.", *gitProvider.Id)) {
			return &gitProvider
		}

		if *gitProvider.BaseApiUrl != "" && strings.Contains(url, getHostnameFromUrl(*gitProvider.BaseApiUrl)) {
			return &gitProvider
		}
	}
	return nil
}

func getHostnameFromUrl(url string) string {
	input := url
	input = strings.TrimPrefix(input, "https://")
	input = strings.TrimPrefix(input, "http://")
	input = strings.TrimPrefix(input, "www.")

	// Remove everything after the first '/'
	if slashIndex := strings.Index(input, "/"); slashIndex != -1 {
		input = input[:slashIndex]
	}

	return input
}
