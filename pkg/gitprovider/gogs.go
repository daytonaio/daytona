// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	gogs "github.com/gogs/go-gogs-client"
)

type GogsGitProvider struct {
	*AbstractGitProvider

	token      string
	baseApiUrl string
}

func NewGogsGitProvider(token string, baseApiUrl string) *GogsGitProvider {
	provider := &GogsGitProvider{
		token:               token,
		baseApiUrl:          baseApiUrl,
		AbstractGitProvider: &AbstractGitProvider{},
	}
	provider.AbstractGitProvider.GitProvider = provider

	return provider
}

func (g *GogsGitProvider) getApiClient() *gogs.Client {
	return gogs.NewClient(g.baseApiUrl, g.token)
}

func (g *GogsGitProvider) CanHandle(repoUrl string) (bool, error) {
	staticContext, err := g.ParseStaticGitContext(repoUrl)
	if err != nil {
		return false, err
	}

	return strings.Contains(g.baseApiUrl, staticContext.Source), nil
}

func (g *GogsGitProvider) GetNamespaces(options ListOptions) ([]*GitNamespace, error) {
	client := g.getApiClient()
	var namespaces []*GitNamespace
	user, err := g.GetUser()
	if err != nil {
		return nil, err
	}
	orgs, err := client.ListMyOrgs()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Namespace : %w", err)
	}

	namespaces = append([]*GitNamespace{{Id: personalNamespaceId, Name: user.Username}}, namespaces...)
	for _, org := range orgs {
		namespaces = append(namespaces, &GitNamespace{Id: org.UserName, Name: org.UserName})
	}
	return namespaces, nil
}

func (g *GogsGitProvider) GetUser() (*GitUser, error) {
	client := g.getApiClient()
	user, err := client.GetSelfInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch User : %w", err)
	}
	fullName := user.FullName
	userName := user.UserName
	if fullName == "" {
		fullName = userName
	}
	return &GitUser{
		Id:       strconv.FormatInt(user.ID, 10),
		Username: user.UserName,
		Name:     fullName,
		Email:    user.Email,
	}, nil
}

func (g *GogsGitProvider) GetRepositories(namespace string, options ListOptions) ([]*GitRepository, error) {
	client := g.getApiClient()
	var repoList []*gogs.Repository
	if namespace == personalNamespaceId {
		repos, err := client.ListMyRepos()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch Repositories : %w", err)
		}
		repoList = repos
	} else {
		repos, err := client.ListOrgRepos(namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch Repositories : %w", err)
		}
		repoList = repos
	}
	repos := []*GitRepository{}
	for _, repo := range repoList {
		u, err := url.Parse(repo.HTMLURL)
		if err != nil {
			return nil, err
		}

		repos = append(repos, &GitRepository{
			Id:     repo.Name,
			Name:   repo.Name,
			Url:    repo.HTMLURL,
			Branch: repo.DefaultBranch,
			Owner:  repo.Owner.UserName,
			Source: u.Host,
		})
	}

	return repos, nil
}

func (g *GogsGitProvider) GetRepoBranches(repositoryId string, namespaceId string, options ListOptions) ([]*GitBranch, error) {
	client := g.getApiClient()
	if namespaceId == personalNamespaceId {
		user, err := g.GetUser()
		if err != nil {
			return nil, err
		}
		namespaceId = user.Username
	}
	var branches []*GitBranch

	repoBranches, err := client.ListRepoBranches(namespaceId, repositoryId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Branches: %w", err)
	}
	for _, branch := range repoBranches {
		responseBranch := &GitBranch{
			Name: branch.Name,
		}
		if branch.Commit != nil {
			responseBranch.Sha = branch.Commit.ID
		}
		branches = append(branches, responseBranch)
	}

	return branches, nil
}

func (g *GogsGitProvider) GetDefaultBranch(staticContext *StaticGitContext) (*string, error) {
	client := g.getApiClient()
	repo, err := client.GetRepo(staticContext.Owner, staticContext.Name)
	if err != nil {
		return nil, err
	}

	return &repo.DefaultBranch, nil
}

func (g *GogsGitProvider) GetRepoPRs(repositoryId string, namespaceId string, options ListOptions) ([]*GitPullRequest, error) {
	// Gogs does not have any API endpoint to fetch PRs
	return []*GitPullRequest{}, nil
}

func (g *GogsGitProvider) GetBranchByCommit(staticContext *StaticGitContext) (string, error) {
	client := g.getApiClient()
	repoBranches, err := client.ListRepoBranches(staticContext.Owner, staticContext.Name)
	if err != nil {
		return "", fmt.Errorf("failed to get branch by commit: %w", err)
	}

	var branchName string
	for _, branch := range repoBranches {
		commitId := branch.Commit.ID
		if *staticContext.Sha == commitId {
			branchName = branch.Name
			break
		}

		for commitId != "" {
			commit, err := client.GetSingleCommit(staticContext.Owner, staticContext.Name, commitId)
			if err != nil {
				continue
			}

			if *staticContext.Sha == commit.SHA {
				branchName = branch.Name
				break
			}
			if len(commit.Parents) > 0 {
				commitId = commit.Parents[0].SHA
				if *staticContext.Sha == commitId {
					branchName = branch.Name
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
		return "", fmt.Errorf("status code: %d branch not found for SHA: %s", http.StatusNotFound, *staticContext.Sha)
	}
	return branchName, nil
}

func (g *GogsGitProvider) GetLastCommitSha(staticContext *StaticGitContext) (string, error) {
	client := g.getApiClient()
	branchName := ""
	if staticContext.Branch != nil {
		branchName = *staticContext.Branch
	}
	branch, err := client.GetRepoBranch(staticContext.Owner, staticContext.Name, branchName)
	if err != nil {
		return "", fmt.Errorf("failed to get last commit SHA: %w", err)
	}
	return branch.Commit.ID, nil
}

func (g *GogsGitProvider) ParseStaticGitContext(repoUrl string) (*StaticGitContext, error) {
	staticContext, err := g.AbstractGitProvider.ParseStaticGitContext(repoUrl)
	if err != nil {
		return nil, err
	}

	if staticContext.Path == nil {
		return staticContext, nil
	}

	parts := strings.Split(*staticContext.Path, "/")
	switch {
	case len(parts) >= 2 && parts[0] == "pulls":
		prNumber, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}
		prUint := uint32(prNumber)
		staticContext.PrNumber = &prUint
		staticContext.Path = nil
	case len(parts) >= 2 && parts[0] == "src":
		staticContext.Branch = &parts[1]
		if len(parts) > 2 {
			branchPath := strings.Join(parts[2:], "/")
			staticContext.Path = &branchPath
		} else {
			staticContext.Path = nil
		}
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

func (g *GogsGitProvider) GetPrContext(staticContext *StaticGitContext) (*StaticGitContext, error) {
	// Gogs does not have any API endpoint to fetch PRs
	return nil, fmt.Errorf("creating workspaces from Pull Requests is not supported for Gogs")
}

func (g *GogsGitProvider) GetUrlFromContext(repoContext *GetRepositoryContext) string {
	url := strings.TrimSuffix(repoContext.Url, ".git")

	if repoContext.Branch != nil && *repoContext.Branch != "" {
		if repoContext.Sha != nil && *repoContext.Sha == *repoContext.Branch {
			url += "/commit/" + *repoContext.Branch
		} else {
			url += "/src/" + *repoContext.Branch
		}

		if repoContext.Path != nil && *repoContext.Path != "" {
			url += "/" + *repoContext.Path
		}
	} else if repoContext.Path != nil {
		url += "/src/main/" + *repoContext.Path
	}

	return url
}
