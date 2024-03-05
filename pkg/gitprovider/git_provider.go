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

// func GetRepository(url string) {
// 	// GetProvider
// 	provider := GitProvider{}

// 	repo := provider.ParseGitUrl(url)

// 	if repo.Prnumber {
// 		repo = provider.GetPullRequestContext(repo)
// 	}

// 	sha := provider.GetLastCommitSha(repo)
// 	repo.Sha = sha

// 	return repo
// }

// func parseGitUrl(url string) (*types.Repository, error) {
// 	if strings.HasPrefix(url, "git@") {
// 		return parseGitUrlWithSsh(url)
// 	}

// 	if !strings.HasPrefix(url, "http") {
// 		return nil, errors.New("Can not parse git url")
// 	}

// 	removedProtocol := strings.Replace(url, "(^\w+:|^)://", "", -1)
// 	splitted := strings.Split(removedProtocol, "/")
// 	source := splitted[0]
// 	owner := splitted[1]
// 	repo := strings.Replace(splitted[2], ".git", "", -1)

// 	path := strings.Join(splitted[3:], "/")

// 	...

// }

// public parseGitUrl(gitUrl: string): StaticGitContext {
// 	if (gitUrl.startsWith('git@')) {
// 		return this.parseGitUrlWithSsh(gitUrl)
// 	}

// 	if (!gitUrl.startsWith('http')) {
// 		throw new Error(`Can not parse git url: ${gitUrl}`)
// 	}

// 	const removedProtocol = gitUrl.replace(/(^\w+:|^)\/\//, '')
// 	const splitted = removedProtocol.split('/')
// 	const source = splitted[0]
// 	const owner = splitted[1]
// 	const repo = splitted[2].replace(/\.git$/, '')
// 	const path = splitted.slice(3).join('/')

// 	const cloneUrl = this.getCloneUrl(source, owner, repo)
// 	const webUrl = this.getWebUrl(source, owner, repo)

// 	return {
// 		owner,
// 		repo,
// 		source,
// 		path,
// 		cloneUrl,
// 		webUrl,
// 		providerId: this.providerName,
// 	}
// }
