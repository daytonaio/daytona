// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"

	"github.com/daytonaio/daytona/pkg/types"
	"github.com/xanzy/go-gitlab"
)

type GitLabGitProvider struct {
	token      string
	baseApiUrl string
}

func (g *GitLabGitProvider) GetNamespaces() ([]GitNamespace, error) {
	client := g.getApiClient()
	user, err := g.GetUserData()
	if err != nil {
		return nil, err
	}

	groupList, _, err := client.Groups.ListGroups(&gitlab.ListGroupsOptions{})
	if err != nil {
		return nil, err
	}

	namespaces := make([]GitNamespace, len(groupList)+1) // +1 for the personal namespace
	namespaces[0] = GitNamespace{Id: personalNamespaceId, Name: user.Username}

	for i, group := range groupList {
		namespaces[i+1].Id = strconv.Itoa(group.ID)
		namespaces[i+1].Name = group.Name
	}

	return namespaces, nil
}

func (g *GitLabGitProvider) GetRepositories(namespace string) ([]types.Repository, error) {
	client := g.getApiClient()
	var response []types.Repository
	var repoList []*gitlab.Project
	var err error

	if namespace == personalNamespaceId {
		user, err := g.GetUserData()
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
		response = append(response, types.Repository{
			Id:   strconv.Itoa(repo.ID),
			Name: repo.Name,
			Url:  repo.WebURL,
		})
	}

	return response, err
}

func (g *GitLabGitProvider) GetRepoBranches(repo types.Repository, namespaceId string) ([]GitBranch, error) {
	client := g.getApiClient()
	var response []GitBranch

	branches, _, err := client.Branches.ListBranches(repo.Id, &gitlab.ListBranchesOptions{})
	if err != nil {
		return nil, err
	}

	for _, branch := range branches {
		responseBranch := GitBranch{
			Name: branch.Name,
		}
		if branch.Commit != nil {
			responseBranch.SHA = branch.Commit.ID
		}
		response = append(response, responseBranch)
	}

	return response, nil
}

func (g *GitLabGitProvider) GetRepoPRs(repo types.Repository, namespaceId string) ([]GitPullRequest, error) {
	client := g.getApiClient()
	var response []GitPullRequest

	mergeRequests, _, err := client.MergeRequests.ListProjectMergeRequests(repo.Id, &gitlab.ListProjectMergeRequestsOptions{})
	if err != nil {
		return nil, err
	}

	for _, mergeRequest := range mergeRequests {
		response = append(response, GitPullRequest{
			Name:   mergeRequest.Title,
			Branch: mergeRequest.SourceBranch,
		})
	}

	return response, nil
}

func (g *GitLabGitProvider) ParseGitUrl(url string) (*types.Repository, error) {
	pattern := `^https?://gitlab.com/([^/]+)/([^/]+)(?:/-/tree/([^/]+))?/?$`

	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(url)

	if len(match) < 3 {
		return nil, errors.New("invalid GitLab URL: " + url)
	}

	var branch string

	if len(match) > 3 {
		branch = match[3]
	}

	repoURL := fmt.Sprintf("https://gitlab.com/%s/%s", match[1], match[2])

	repo := &types.Repository{
		Url:    repoURL,
		Branch: branch,
	}

	return repo, nil
}

func (g *GitLabGitProvider) GetUserData() (GitUser, error) {
	client := g.getApiClient()

	user, _, err := client.Users.CurrentUser()
	if err != nil {
		return GitUser{}, err
	}

	userId := strconv.Itoa(user.ID)

	response := GitUser{
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
