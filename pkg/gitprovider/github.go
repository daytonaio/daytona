// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type GitHubGitProvider struct {
	token string
}

func NewGitHubGitProvider(token string) *GitHubGitProvider {
	return &GitHubGitProvider{
		token: token,
	}
}

func (g *GitHubGitProvider) GetNamespaces() ([]*GitNamespace, error) {
	client := g.getApiClient()
	user, err := g.GetUser()
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

	namespaces := []*GitNamespace{}

	for _, org := range orgList {
		namespace := &GitNamespace{}
		if org.Login != nil {
			namespace.Id = *org.Login
			namespace.Name = *org.Login
		} else if org.Name != nil {
			namespace.Name = *org.Name
		}
		namespaces = append(namespaces, namespace)
	}

	namespaces = append([]*GitNamespace{{Id: personalNamespaceId, Name: user.Username}}, namespaces...)

	return namespaces, nil
}

func (g *GitHubGitProvider) GetRepositories(namespace string) ([]*GitRepository, error) {
	client := g.getApiClient()
	var response []*GitRepository
	query := "fork:true "

	if namespace == personalNamespaceId {
		user, err := g.GetUser()
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
		response = append(response, &GitRepository{
			Id:   *repo.Name,
			Name: *repo.Name,
			Url:  *repo.HTMLURL,
		})
	}

	return response, err
}

func (g *GitHubGitProvider) GetRepoBranches(repositoryId string, namespaceId string) ([]*GitBranch, error) {
	client := g.getApiClient()

	if namespaceId == personalNamespaceId {
		user, err := g.GetUser()
		if err != nil {
			return nil, err
		}
		namespaceId = user.Username
	}

	var response []*GitBranch

	repoBranches, _, err := client.Repositories.ListBranches(context.Background(), namespaceId, repositoryId, &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, branch := range repoBranches {
		responseBranch := &GitBranch{
			Name: *branch.Name,
		}
		if branch.Commit != nil && branch.Commit.SHA != nil {
			responseBranch.SHA = *branch.Commit.SHA
		}
		response = append(response, responseBranch)
	}

	return response, nil
}

func (g *GitHubGitProvider) GetRepoPRs(repositoryId string, namespaceId string) ([]*GitPullRequest, error) {
	client := g.getApiClient()

	if namespaceId == personalNamespaceId {
		user, err := g.GetUser()
		if err != nil {
			return nil, err
		}
		namespaceId = user.Username
	}

	var response []*GitPullRequest

	prList, _, err := client.PullRequests.List(context.Background(), namespaceId, repositoryId, &github.PullRequestListOptions{
		State: "open",
	})
	if err != nil {
		return nil, err
	}

	for _, pr := range prList {
		response = append(response, &GitPullRequest{
			Name:   *pr.Title,
			Branch: *pr.Head.Ref,
		})
	}

	return response, nil
}

func (g *GitHubGitProvider) ParseGitUrl(gitURL string) (*GitRepository, error) {
	client := g.getApiClient()
	repo, err := parseGitComponents(gitURL)

	if err != nil {
		return nil, err
	}

	repo, err = g.parseSpecificPath(repo, client)

	if err != nil {
		return nil, err
	}

	return repo, nil
}


func (g *GitHubGitProvider) parseSpecificPath(repo *GitRepository, client *github.Client) (*GitRepository, error) {
	parts := strings.Split(*repo.Path, "/")
	repo.Path = nil

	switch {
	case len(parts) >= 2 && parts[0] == "pull":
		prNumber, _ := strconv.Atoi(parts[1])
		pull, _, err := client.PullRequests.Get(context.Background(), repo.Owner, repo.Name, prNumber)
		if err != nil {
			return nil, err
		}
		repo.Branch = pull.Head.Ref
		repo.Url = *pull.Head.Repo.CloneURL
		repo.Owner = *pull.Head.Repo.Owner.Login

	case len(parts) >= 1 && parts[0] == "tree":
		repo.Branch = &parts[1]
	case len(parts) >= 2 && parts[0] == "blob":
		repo.Branch = &parts[1]
		branchPath := strings.Join(parts[2:], "/")
		repo.Path = &branchPath
	case len(parts) >= 2 && (parts[0] == "commit" || parts[0] == "commits"):
		repo.Sha = parts[1]
		repo.Branch = &repo.Sha
	}

	return repo, nil
}

func (g *GitHubGitProvider) GetUser() (*GitUser, error) {
	client := g.getApiClient()

	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		return nil, err
	}

	response := &GitUser{}

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
