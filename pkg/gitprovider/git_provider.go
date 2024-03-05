// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"errors"

	"github.com/daytonaio/daytona/cmd/daytona/config"
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
	default:
		return nil
	}
}

func GetUsernameFromToken(providerId string, gitProviders []config.GitProvider, token string) (string, error) {
	var gitProvider GitProvider

	switch providerId {
	case "github":
		gitProvider = &GitHubGitProvider{
			token: token,
		}
	case "gitlab":
		gitProvider = &GitLabGitProvider{
			token: token,
		}
	case "bitbucket":
		gitProvider = &BitbucketGitProvider{
			token: token,
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
