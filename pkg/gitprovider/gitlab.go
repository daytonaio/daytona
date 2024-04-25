// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/xanzy/go-gitlab"
)

type GitLabGitProvider struct {
	token      string
	baseApiUrl *string
}

func NewGitLabGitProvider(token string, baseApiUrl *string) *GitLabGitProvider {
	return &GitLabGitProvider{
		token:      token,
		baseApiUrl: baseApiUrl,
	}
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
		response = append(response, &GitRepository{
			Id:   strconv.Itoa(repo.ID),
			Name: repo.Name,
			Url:  repo.WebURL,
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
			responseBranch.SHA = branch.Commit.ID
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
		response = append(response, &GitPullRequest{
			Name:   mergeRequest.Title,
			Branch: mergeRequest.SourceBranch,
		})
	}

	return response, nil
}

func (g *GitLabGitProvider) ParseGitUrl(gitURL string) (*GitRepository, error) {
	client := g.getApiClient()
	repo , err := parseGitComponents(gitURL)
	if err != nil {
		return nil, err
	}

	repo, err = g.parseSpecificPath(repo, client)

	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (g *GitLabGitProvider) parseSpecificPath(repo *GitRepository, client *gitlab.Client) (*GitRepository, error) {
	parts := strings.Split(*repo.Path, "/")
	repo.Path = nil

	switch {
	case len(parts) >= 2 && parts[0] == "-":
		switch parts[1] {
		case "tree":
			repo.Branch = &parts[2]
		case "merge_requests":
			prNumber, _ := strconv.Atoi(parts[2])
			pull, _, err := client.MergeRequests.GetMergeRequest(fmt.Sprintf("%s/%s",repo.Owner,repo.Name), prNumber, nil)
			if err != nil {
				return nil, err
			}
			repo.Branch = &pull.SourceBranch
			repo.Url = getCloneURL(repo.Source, pull.Author.Username, repo.Name)
			repo.Owner = pull.Author.Username
		case "commits":
			repo.Sha = parts[2]
			repo.Branch = &repo.Sha
		}

	case len(parts) >= 2 && parts[0] == "merge_requests":
		prNumber, _ := strconv.Atoi(parts[1])
		pull, _, err := client.MergeRequests.GetMergeRequest(fmt.Sprintf("%s/%s",repo.Owner,repo.Name), prNumber, nil)
		if err != nil {
			return nil, err
		}

		repo.Branch = &pull.SourceBranch
		repo.Url = getCloneURL(repo.Source, pull.Author.Username, repo.Name)
		repo.Owner = pull.Author.Username
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
