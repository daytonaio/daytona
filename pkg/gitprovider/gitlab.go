// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"log"
	"strconv"

	"github.com/daytonaio/daytona/pkg/types"
	"github.com/xanzy/go-gitlab"
)

type GitLabGitProvider struct {
	token      string
	baseApiUrl string
}

func (g *GitLabGitProvider) GetNamespaces() ([]types.GitNamespace, error) {
	client := g.getApiClient()
	user, err := g.GetUser()
	if err != nil {
		return nil, err
	}

	groupList, _, err := client.Groups.ListGroups(&gitlab.ListGroupsOptions{})
	if err != nil {
		return nil, err
	}

	namespaces := make([]types.GitNamespace, len(groupList)+1) // +1 for the personal namespace
	namespaces[0] = types.GitNamespace{Id: personalNamespaceId, Name: user.Username}

	for i, group := range groupList {
		namespaces[i+1].Id = strconv.Itoa(group.ID)
		namespaces[i+1].Name = group.Name
	}

	return namespaces, nil
}

func (g *GitLabGitProvider) GetRepositories(namespace string) ([]types.GitRepository, error) {
	client := g.getApiClient()
	var response []types.GitRepository
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
		response = append(response, types.GitRepository{
			Id:   strconv.Itoa(repo.ID),
			Name: repo.Name,
			Url:  repo.WebURL,
		})
	}

	return response, err
}

func (g *GitLabGitProvider) GetRepoBranches(repositoryId string, namespaceId string) ([]types.GitBranch, error) {
	client := g.getApiClient()
	var response []types.GitBranch

	branches, _, err := client.Branches.ListBranches(repositoryId, &gitlab.ListBranchesOptions{})
	if err != nil {
		return nil, err
	}

	for _, branch := range branches {
		responseBranch := types.GitBranch{
			Name: branch.Name,
		}
		if branch.Commit != nil {
			responseBranch.SHA = branch.Commit.ID
		}
		response = append(response, responseBranch)
	}

	return response, nil
}

func (g *GitLabGitProvider) GetRepoPRs(repositoryId string, namespaceId string) ([]types.GitPullRequest, error) {
	client := g.getApiClient()
	var response []types.GitPullRequest

	mergeRequests, _, err := client.MergeRequests.ListProjectMergeRequests(repositoryId, &gitlab.ListProjectMergeRequestsOptions{})
	if err != nil {
		return nil, err
	}

	for _, mergeRequest := range mergeRequests {
		response = append(response, types.GitPullRequest{
			Name:   mergeRequest.Title,
			Branch: mergeRequest.SourceBranch,
		})
	}

	return response, nil
}

func (g *GitLabGitProvider) GetUser() (types.GitUser, error) {
	client := g.getApiClient()

	user, _, err := client.Users.CurrentUser()
	if err != nil {
		return types.GitUser{}, err
	}

	userId := strconv.Itoa(user.ID)

	response := types.GitUser{
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

	if g.baseApiUrl == "" {
		client, err = gitlab.NewClient(g.token)
	} else {
		client, err = gitlab.NewClient(g.token, gitlab.WithBaseURL(g.baseApiUrl))
	}
	if err != nil {
		log.Fatal(err)
	}

	return client
}
