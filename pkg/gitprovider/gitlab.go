// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/xanzy/go-gitlab"
)

type GitLabGitProvider struct {
	*AbstractGitProvider

	token      string
	baseApiUrl *string
}

func NewGitLabGitProvider(token string, baseApiUrl *string) *GitLabGitProvider {
	gitProvider := &GitLabGitProvider{
		token:               token,
		baseApiUrl:          baseApiUrl,
		AbstractGitProvider: &AbstractGitProvider{},
	}
	gitProvider.AbstractGitProvider.GitProvider = gitProvider

	return gitProvider
}

func (g *GitLabGitProvider) GetNamespaces() ([]*GitNamespace, error) {
	client := g.getApiClient()
	user, err := g.GetUser()
	if err != nil {
		return nil, err
	}

	groupList, _, err := client.Groups.ListGroups(&gitlab.ListGroupsOptions{})
	if err != nil {
		return nil, err
	}

	namespaces := []*GitNamespace{}

	for _, group := range groupList {
		namespaces = append(namespaces, &GitNamespace{
			Id:   strconv.Itoa(group.ID),
			Name: group.Name,
		})
	}

	namespaces = append([]*GitNamespace{{Id: personalNamespaceId, Name: user.Username}}, namespaces...)

	return namespaces, nil
}

func (g *GitLabGitProvider) GetRepositories(namespace string) ([]*GitRepository, error) {
	client := g.getApiClient()
	var response []*GitRepository
	var repoList []*gitlab.Project
	var err error

	if namespace == personalNamespaceId {
		user, err := g.GetUser()
		if err != nil {
			return nil, err
		}

		repoList, _, err = client.Projects.ListUserProjects(user.Id, &gitlab.ListProjectsOptions{
			ListOptions: gitlab.ListOptions{
				PerPage: 100,
				Page:    1,
			},
		})
		if err != nil {
			return nil, err
		}
	} else {
		repoList, _, err = client.Groups.ListGroupProjects(namespace, &gitlab.ListGroupProjectsOptions{
			ListOptions: gitlab.ListOptions{
				PerPage: 100,
				Page:    1,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	for _, repo := range repoList {
		u, err := url.Parse(repo.WebURL)
		if err != nil {
			return nil, err
		}

		response = append(response, &GitRepository{
			Id:     strconv.Itoa(repo.ID),
			Name:   repo.Path,
			Url:    repo.WebURL,
			Branch: &repo.DefaultBranch,
			Owner:  repo.Namespace.Path,
			Source: u.Host,
		})
	}

	return response, err
}

func (g *GitLabGitProvider) GetRepoBranches(repositoryId string, namespaceId string) ([]*GitBranch, error) {
	client := g.getApiClient()
	var response []*GitBranch

	branches, _, err := client.Branches.ListBranches(repositoryId, &gitlab.ListBranchesOptions{})
	if err != nil {
		return nil, err
	}

	for _, branch := range branches {
		responseBranch := &GitBranch{
			Name: branch.Name,
		}
		if branch.Commit != nil {
			responseBranch.Sha = branch.Commit.ID
		}
		response = append(response, responseBranch)
	}

	return response, nil
}

func (g *GitLabGitProvider) GetRepoPRs(repositoryId string, namespaceId string) ([]*GitPullRequest, error) {
	client := g.getApiClient()
	var response []*GitPullRequest

	mergeRequests, _, err := client.MergeRequests.ListProjectMergeRequests(repositoryId, &gitlab.ListProjectMergeRequestsOptions{})
	if err != nil {
		return nil, err
	}

	for _, mergeRequest := range mergeRequests {
		sourceRepo, _, err := client.Projects.GetProject(mergeRequest.SourceProjectID, nil)
		if err != nil {
			return nil, err
		}

		response = append(response, &GitPullRequest{
			Name:            mergeRequest.Title,
			Branch:          mergeRequest.SourceBranch,
			Sha:             mergeRequest.SHA,
			SourceRepoId:    fmt.Sprint(mergeRequest.SourceProjectID),
			SourceRepoUrl:   sourceRepo.WebURL,
			SourceRepoOwner: sourceRepo.Namespace.Path,
			SourceRepoName:  sourceRepo.Path,
		})
	}

	return response, nil
}

func (g *GitLabGitProvider) GetUser() (*GitUser, error) {
	client := g.getApiClient()

	user, _, err := client.Users.CurrentUser()
	if err != nil {
		return nil, err
	}

	userId := strconv.Itoa(user.ID)

	response := &GitUser{
		Id:       userId,
		Username: user.Username,
		Name:     user.Name,
		Email:    user.Email,
	}

	return response, nil
}

func (g *GitLabGitProvider) GetLastCommitSha(staticContext *StaticGitContext) (string, error) {
	client := g.getApiClient()

	var sha *string

	if staticContext.Branch != nil {
		sha = staticContext.Branch
	}

	if staticContext.Sha != nil {
		sha = staticContext.Sha
	}

	commits, _, err := client.Commits.ListCommits(staticContext.Id, &gitlab.ListCommitsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 1,
		},
		RefName: sha,
	})
	if err != nil {
		return "", err
	}
	if len(commits) == 0 {
		return "", nil
	}

	return commits[0].ID, nil
}

func (g *GitLabGitProvider) getApiClient() *gitlab.Client {
	var client *gitlab.Client
	var err error

	if g.baseApiUrl == nil {
		client, err = gitlab.NewClient(g.token)
	} else {
		client, err = gitlab.NewClient(g.token, gitlab.WithBaseURL(*g.baseApiUrl))
	}
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func (g *GitLabGitProvider) parseStaticGitContext(repoUrl string) (*StaticGitContext, error) {
	if strings.HasPrefix(repoUrl, "git@") {
		return g.parseSshGitUrl(repoUrl)
	}

	if !strings.HasPrefix(repoUrl, "http") {
		return nil, fmt.Errorf("can not parse git URL: %s", repoUrl)
	}

	repoUrl = strings.TrimSuffix(repoUrl, ".git")

	u, err := url.Parse(repoUrl)
	if err != nil {
		return nil, err
	}

	staticContext := &StaticGitContext{
		Source: u.Host,
	}

	parts := strings.Split(u.Path, "/-/")

	ownerRepo := strings.TrimPrefix(parts[0], "/")

	if len(parts) == 2 {
		staticContext.Path = &parts[1]
	}

	if len(parts) > 2 {
		return nil, fmt.Errorf("can not parse git URL: %s", repoUrl)
	}

	ownerRepoParts := strings.Split(ownerRepo, "/")
	if len(ownerRepoParts) < 2 {
		return nil, fmt.Errorf("can not parse git URL: %s", repoUrl)
	}

	staticContext.Name = ownerRepoParts[len(ownerRepoParts)-1]
	staticContext.Owner = strings.Join(ownerRepoParts[:len(ownerRepoParts)-1], "/")
	staticContext.Url = getCloneUrl(staticContext.Source, staticContext.Owner, staticContext.Name)
	staticContext.Id = fmt.Sprintf("%s/%s", staticContext.Owner, staticContext.Name)

	if staticContext.Path == nil {
		return staticContext, nil
	}

	switch {
	case strings.Contains(*staticContext.Path, "merge_requests"):
		parts := strings.Split(*staticContext.Path, "merge_requests/")
		mrParts := strings.Split(parts[1], "/")
		if len(mrParts) < 1 {
			return nil, fmt.Errorf("can not parse git URL: %s", repoUrl)
		}
		mrNumber, err := strconv.Atoi(mrParts[0])
		if err != nil {
			return nil, err
		}
		mrNumberUint := uint32(mrNumber)
		staticContext.PrNumber = &mrNumberUint
		staticContext.Path = nil
	case strings.Contains(*staticContext.Path, "tree/"):
		parts := strings.Split(*staticContext.Path, "tree/")
		if len(parts) < 2 {
			return nil, fmt.Errorf("can not parse git URL: %s", repoUrl)
		}

		branchParts := strings.Split(parts[1], "/")
		if len(branchParts) < 1 {
			return nil, fmt.Errorf("can not parse git URL: %s", repoUrl)
		}

		staticContext.Branch = &branchParts[0]
		staticContext.Path = nil
	case strings.Contains(*staticContext.Path, "blob/"):
		parts := strings.Split(*staticContext.Path, "blob/")
		if len(parts) < 2 {
			return nil, fmt.Errorf("can not parse git URL: %s", repoUrl)
		}

		branchParts := strings.Split(parts[1], "/")
		if len(branchParts) < 1 {
			return nil, fmt.Errorf("can not parse git URL: %s", repoUrl)
		}

		staticContext.Branch = &branchParts[0]
		branchPath := strings.Join(branchParts[1:], "/")
		staticContext.Path = &branchPath
	case strings.Contains(*staticContext.Path, "commit/"):
		parts := strings.Split(*staticContext.Path, "commit/")
		if len(parts) < 2 {
			return nil, fmt.Errorf("can not parse git URL: %s", repoUrl)
		}

		commitParts := strings.Split(parts[1], "/")
		if len(commitParts) < 1 {
			return nil, fmt.Errorf("can not parse git URL: %s", repoUrl)
		}

		staticContext.Sha = &commitParts[0]
		staticContext.Branch = &commitParts[0]
		staticContext.Path = nil
	case strings.Contains(*staticContext.Path, "commits/"):
		parts := strings.Split(*staticContext.Path, "commits/")
		if len(parts) < 2 {
			return nil, fmt.Errorf("can not parse git URL: %s", repoUrl)
		}

		branchParts := strings.Split(parts[1], "/")
		if len(branchParts) < 1 {
			return nil, fmt.Errorf("can not parse git URL: %s", repoUrl)
		}

		staticContext.Branch = &branchParts[0]
		staticContext.Path = nil
	}

	return staticContext, nil
}

func (g *GitLabGitProvider) getPrContext(staticContext *StaticGitContext) (*StaticGitContext, error) {
	if staticContext.PrNumber == nil {
		return staticContext, nil
	}

	client := g.getApiClient()

	pull, _, err := client.MergeRequests.GetMergeRequest(staticContext.Id, int(*staticContext.PrNumber), nil)
	if err != nil {
		return nil, err
	}

	project, _, err := client.Projects.GetProject(staticContext.Id, nil)
	if err != nil {
		return nil, err
	}

	repo := *staticContext
	repo.Branch = &pull.SourceBranch
	repo.Url = project.HTTPURLToRepo
	repo.Owner = pull.Author.Username

	return &repo, nil
}
