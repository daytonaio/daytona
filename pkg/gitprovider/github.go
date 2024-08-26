// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type GitHubGitProvider struct {
	*AbstractGitProvider

	token      string
	baseApiUrl *string
}

func NewGitHubGitProvider(token string, baseApiUrl *string) *GitHubGitProvider {
	gitProvider := &GitHubGitProvider{
		token:               token,
		baseApiUrl:          baseApiUrl,
		AbstractGitProvider: &AbstractGitProvider{},
	}
	gitProvider.AbstractGitProvider.GitProvider = gitProvider

	return gitProvider
}

func (g *GitHubGitProvider) GetNamespaces() ([]*GitNamespace, error) {
	client := g.getApiClient()
	user, err := g.GetUser()
	if err != nil {
		return nil, err
	}

	orgList, _, err := client.Organizations.List(context.Background(), "", &github.ListOptions{
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
		u, err := url.Parse(*repo.HTMLURL)
		if err != nil {
			return nil, err
		}
		response = append(response, &GitRepository{
			Id:     *repo.Name,
			Name:   *repo.Name,
			Url:    *repo.HTMLURL,
			Branch: repo.DefaultBranch,
			Owner:  *repo.Owner.Login,
			Source: u.Host,
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
			responseBranch.Sha = *branch.Commit.SHA
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
			Name:            *pr.Title,
			Branch:          *pr.Head.Ref,
			Sha:             *pr.Head.SHA,
			SourceRepoId:    *pr.Head.Repo.Name,
			SourceRepoName:  *pr.Head.Repo.Name,
			SourceRepoUrl:   *pr.Head.Repo.HTMLURL,
			SourceRepoOwner: *pr.Head.Repo.Owner.Login,
		})
	}

	return response, nil
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

func (g *GitHubGitProvider) GetLastCommitSha(staticContext *StaticGitContext) (string, error) {
	client := g.getApiClient()

	sha := ""

	if staticContext.Branch != nil {
		sha = *staticContext.Branch
	}

	if staticContext.Sha != nil {
		sha = *staticContext.Sha
	}

	commits, _, err := client.Repositories.ListCommits(context.Background(), staticContext.Owner, staticContext.Name, &github.CommitsListOptions{
		SHA: sha,
		ListOptions: github.ListOptions{
			PerPage: 1,
		},
	})
	if err != nil {
		return "", err
	}
	if len(commits) == 0 {
		return "", nil
	}

	return *commits[0].SHA, nil
}

func (g *GitHubGitProvider) GetUrlFromRepository(repository *GitRepository) string {
	url := strings.TrimSuffix(repository.Url, ".git")

	if repository.Branch != nil && *repository.Branch != "" {
		if repository.Sha == *repository.Branch {
			url += "/commit/" + *repository.Branch
		} else {
			url += "/tree/" + *repository.Branch
		}

		if repository.Path != nil {
			url += "/" + *repository.Path
		}
	} else if repository.Path != nil {
		url += "/blob/main/" + *repository.Path
	}

	return url
}

func (g *GitHubGitProvider) getApiClient() *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: g.token},
	)
	tc := oauth2.NewClient(ctx, ts)

	if g.token == "" {
		tc = nil
	}

	client := github.NewClient(tc)

	if g.baseApiUrl != nil {
		trimmedUrl := strings.TrimPrefix(*g.baseApiUrl, "https://")
		trimmedUrl = strings.TrimSuffix(trimmedUrl, "api/v3/")
		trimmedUrl = strings.TrimSuffix(trimmedUrl, "/")

		client.BaseURL = &url.URL{
			Scheme: "https",
			Host:   trimmedUrl,
			Path:   "api/v3/",
		}
	}

	return client
}

func (g *GitHubGitProvider) GetBranchByCommit(staticContext *StaticGitContext) (string, error) {
	if staticContext.Sha == nil || *staticContext.Sha == "" {
		return *staticContext.Sha, nil
	}

	client := g.getApiClient()

	branches, _, err := client.Repositories.ListBranches(context.Background(), staticContext.Owner, staticContext.Name, nil)
	if err != nil {
		return "", fmt.Errorf("failed to list branches: %v", err)
	}

	var branchName string
	for _, branch := range branches {
		branchCommitSHA := branch.GetCommit().GetSHA()

		if branchCommitSHA == *staticContext.Sha {
			branchName = branch.GetName()
			break
		}

		commitId := branchCommitSHA
		for commitId != "" {
			commit, _, err := client.Repositories.GetCommit(context.Background(), staticContext.Owner, staticContext.Name, commitId)
			if err != nil {
				return "", err
			}

			if *commit.SHA == *staticContext.Sha {
				branchName = branch.GetName()
				break
			}

			if len(commit.Parents) > 0 {
				commitId := commit.Parents[0].SHA
				if *staticContext.Sha == *commitId {
					branchName = branch.GetName()
					break
				}
			} else {
				commitId = ""
			}
		}

		if branchName != "" {
			break
		}
	}

	if branchName == "" {
		return "", fmt.Errorf("branch not found for SHA: %s", *staticContext.Sha)
	}

	return branchName, nil
}

func (g *GitHubGitProvider) getPrContext(staticContext *StaticGitContext) (*StaticGitContext, error) {
	if staticContext.PrNumber == nil {
		return staticContext, nil
	}

	client := g.getApiClient()

	pr, _, err := client.PullRequests.Get(context.Background(), staticContext.Owner, staticContext.Name, int(*staticContext.PrNumber))
	if err != nil {
		return nil, err
	}

	repo := *staticContext
	repo.Branch = pr.Head.Ref
	repo.Url = *pr.Head.Repo.CloneURL
	repo.Id = *pr.Head.Repo.Name
	repo.Name = *pr.Head.Repo.Name
	repo.Owner = *pr.Head.Repo.Owner.Login

	return &repo, nil
}

func (g *GitHubGitProvider) parseStaticGitContext(repoUrl string) (*StaticGitContext, error) {
	staticContext, err := g.AbstractGitProvider.parseStaticGitContext(repoUrl)
	if err != nil {
		return nil, err
	}

	if staticContext.Path == nil {
		return staticContext, nil
	}

	parts := strings.Split(*staticContext.Path, "/")

	switch {
	case len(parts) >= 2 && parts[0] == "pull":
		prNumber, _ := strconv.Atoi(parts[1])
		prUint := uint32(prNumber)
		staticContext.PrNumber = &prUint
		staticContext.Path = nil
	case len(parts) >= 1 && parts[0] == "tree":
		branchPath := strings.Join(parts[1:], "/")
		staticContext.Branch = &branchPath
		staticContext.Path = nil
	case len(parts) >= 2 && parts[0] == "blob":
		staticContext.Branch = &parts[1]
		branchPath := strings.Join(parts[2:], "/")
		staticContext.Path = &branchPath
	case len(parts) >= 2 && parts[0] == "commits":
		staticContext.Branch = &parts[1]
		staticContext.Path = nil
	case len(parts) >= 2 && parts[0] == "commit":
		staticContext.Sha = &parts[1]
		staticContext.Branch = staticContext.Sha
		staticContext.Path = nil
	}

	return staticContext, nil
}
