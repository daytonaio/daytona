// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"
	"strconv"

	"github.com/daytonaio/daytona/pkg/types"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type GitHubGitProvider struct {
	token string
}

func (g *GitHubGitProvider) GetNamespaces() ([]GitNamespace, error) {
	client := g.getApiClient()
	user, err := g.GetUserData()
	if err != nil {
		return nil, err
	}

	orgList, _, err := client.Organizations.List(context.Background(), user.Username, &github.ListOptions{
		PerPage: 100,
		Page:    1,
	})
	if err != nil {
		return nil, err
	}

	namespaces := make([]GitNamespace, len(orgList)+1) // +1 for the user namespace
	namespaces[0] = GitNamespace{Id: personalNamespaceId, Name: user.Username}

	for i, org := range orgList {
		if org.Login != nil {
			namespaces[i+1].Id = *org.Login
			namespaces[i+1].Name = *org.Login
		} else if org.Name != nil {
			namespaces[i+1].Name = *org.Name
		}
	}

	return namespaces, nil
}

func (g *GitHubGitProvider) GetRepositories(namespace string) ([]types.Repository, error) {
	client := g.getApiClient()
	var response []types.Repository
	query := "fork:true "

	if namespace == personalNamespaceId {
		user, err := g.GetUserData()
		if err != nil {
			return nil, err
		}
		query += "user:" + user.Username
	} else {
		query += "org:" + namespace
	}

	repoList, _, err := client.Search.Repositories(context.Background(), query, &github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
			Page:    1,
		},
	})

	if err != nil {
		return nil, err
	}

	for _, repo := range repoList.Repositories {
		response = append(response, types.Repository{
			Name: *repo.Name,
			Url:  *repo.HTMLURL,
		})
	}

	return response, err
}

func (g *GitHubGitProvider) GetRepoBranches(repo types.Repository, namespaceId string) ([]GitBranch, error) {
	client := g.getApiClient()

	if namespaceId == personalNamespaceId {
		user, err := g.GetUserData()
		if err != nil {
			return nil, err
		}
		namespaceId = user.Username
	}

	var response []GitBranch

	repoBranches, _, err := client.Repositories.ListBranches(context.Background(), namespaceId, repo.Name, &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, branch := range repoBranches {
		responseBranch := GitBranch{
			Name: *branch.Name,
		}
		if branch.Commit != nil && branch.Commit.SHA != nil {
			responseBranch.SHA = *branch.Commit.SHA
		}
		response = append(response, responseBranch)
	}

	return response, nil
}

func (g *GitHubGitProvider) GetRepoPRs(repo types.Repository, namespaceId string) ([]GitPullRequest, error) {
	client := g.getApiClient()

	if namespaceId == personalNamespaceId {
		user, err := g.GetUserData()
		if err != nil {
			return nil, err
		}
		namespaceId = user.Username
	}

	var response []GitPullRequest

	prList, _, err := client.PullRequests.List(context.Background(), namespaceId, repo.Name, &github.PullRequestListOptions{
		State: "open",
	})
	if err != nil {
		return nil, err
	}

	for _, pr := range prList {
		response = append(response, GitPullRequest{
			Name:   *pr.Title,
			Branch: *pr.Head.Ref,
		})
	}

	return response, nil
}

func (g *GitHubGitProvider) GetUserData() (GitUser, error) {
	client := g.getApiClient()

	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		return GitUser{}, err
	}

	response := GitUser{}

	if user.ID != nil {
		response.Id = strconv.FormatInt(*user.ID, 10)
	}

	if user.Name != nil {
		response.Name = *user.Name
	}

	if user.Login != nil {
		response.Username = *user.Login
	}

	if user.Email != nil {
		response.Email = *user.Email
	}

	return response, nil
}

func (g *GitHubGitProvider) getApiClient() *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: g.token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return client
}
