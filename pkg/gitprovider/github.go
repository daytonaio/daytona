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
	user, err := g.GetUserData()
	if err != nil {
		return nil, err
	}
	var response []GitBranch

	repoBranches, _, err := client.Repositories.ListBranches(context.Background(), user.Username, repo.Name, &github.ListOptions{})
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

// func (g *GitHubGitProvider) ParseGitUrl(url string) (*types.Repository, error) {
// 	repo := parseGitUrl(url)

// 	if strings.Contains(url, "pull/") {
// 		parts := strings.Split(repo.Path, "pull/")
// 		prNumber, err := strconv.Atoi(strings.Split(parts[1], "/")[0])
// 		if err != nil {
// 			return nil, err
// 		}
// 		repo.Path = ""
// 		repo.PRNumber = prNumber
// 	}
// 	return nil, nil
// }

// public override parseGitUrl(gitUrl: string): StaticGitContext {
// 	const staticContext = super.parseGitUrl(gitUrl)

// 	if (staticContext.path?.includes('pull/')) {
// 		const parts = staticContext.path.split('pull/')
// 		const prNumber = Number(parts[1].split('/')[0])

// 		return {
// 			...staticContext,
// 			path: undefined,
// 			prNumber,
// 		}
// 	}

// 	if (staticContext.path?.includes('tree/')) {
// 		const parts = staticContext.path.split('tree/')
// 		const branch = parts[1].split('/')[0]

// 		return {
// 			...staticContext,
// 			path: undefined,
// 			branch,
// 		}
// 	}

// 	if (staticContext.path?.includes('blob/')) {
// 		const parts = staticContext.path.split('blob/')
// 		const branch = parts[1].split('/')[0]
// 		const path = parts[1].split('/').slice(1).join('/')

// 		return {
// 			...staticContext,
// 			path,
// 			branch,
// 		}
// 	}

// 	if (staticContext.path?.includes('commit/')) {
// 		const parts = staticContext.path.split('commit/')
// 		const sha = parts[1].split('/')[0]

// 		return {
// 			...staticContext,
// 			path: undefined,
// 			sha,
// 			branch: sha,
// 		}
// 	}

// 	if (staticContext.path?.includes('commits/')) {
// 		const parts = staticContext.path.split('commits/')
// 		const sha = parts[1].split('/')[0]

// 		return {
// 			...staticContext,
// 			path: undefined,
// 			sha,
// 			branch: sha,
// 		}
// 	}

// 	return staticContext
// }
