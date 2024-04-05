// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"
	"strconv"

	"code.gitea.io/sdk/gitea"
)

type GiteaGitProvider struct {
	token      string
	baseApiUrl string
}

func NewGiteaGitProvider(token string, baseApiUrl string) *GiteaGitProvider {
	return &GiteaGitProvider{
		token:      token,
		baseApiUrl: baseApiUrl,
	}
}

func (g *GiteaGitProvider) GetNamespaces() ([]*GitNamespace, error) {
	client, err := g.getApiClient()
	if err != nil {
		return nil, err
	}

	user, err := g.GetUser()
	if err != nil {
		return nil, err
	}

	orgList, _, err := client.ListMyOrgs(gitea.ListOrgsOptions{
		ListOptions: gitea.ListOptions{
			Page:     1,
			PageSize: 100,
		},
	})
	if err != nil {
		return nil, err
	}

	namespaces := []*GitNamespace{}

	for _, org := range orgList {
		namespaces = append(namespaces, &GitNamespace{Id: org.UserName, Name: org.UserName})
	}
	namespaces = append([]*GitNamespace{{Id: personalNamespaceId, Name: user.Username}}, namespaces...)

	return namespaces, nil
}

func (g *GiteaGitProvider) GetRepositories(namespace string) ([]*GitRepository, error) {
	client, err := g.getApiClient()
	if err != nil {
		return nil, err
	}

	var repoList []*gitea.Repository

	if namespace == personalNamespaceId {
		user, err := g.GetUser()
		if err != nil {
			return nil, err
		}

		repoList, _, err = client.ListUserRepos(user.Username, gitea.ListReposOptions{
			ListOptions: gitea.ListOptions{
				Page:     1,
				PageSize: 100,
			},
		})
		if err != nil {
			return nil, err
		}
	} else {
		repoList, _, err = client.ListOrgRepos(namespace, gitea.ListOrgReposOptions{
			ListOptions: gitea.ListOptions{
				Page:     1,
				PageSize: 100,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	response := []*GitRepository{}

	for _, repo := range repoList {
		response = append(response, &GitRepository{
			Id:   repo.Name,
			Name: repo.Name,
			Url:  repo.HTMLURL,
		})
	}

	return response, err
}

func (g *GiteaGitProvider) GetRepoBranches(repositoryId string, namespaceId string) ([]*GitBranch, error) {
	client, err := g.getApiClient()
	if err != nil {
		return nil, err
	}

	if namespaceId == personalNamespaceId {
		user, err := g.GetUser()
		if err != nil {
			return nil, err
		}
		namespaceId = user.Username
	}

	repoBranches, _, err := client.ListRepoBranches(namespaceId, repositoryId, gitea.ListRepoBranchesOptions{
		ListOptions: gitea.ListOptions{
			Page:     1,
			PageSize: 100,
		},
	})
	if err != nil {
		return nil, err
	}

	response := []*GitBranch{}

	for _, branch := range repoBranches {
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

func (g *GiteaGitProvider) GetRepoPRs(repositoryId string, namespaceId string) ([]*GitPullRequest, error) {
	client, err := g.getApiClient()
	if err != nil {
		return nil, err
	}

	if namespaceId == personalNamespaceId {
		user, err := g.GetUser()
		if err != nil {
			return nil, err
		}
		namespaceId = user.Username
	}

	prList, _, err := client.ListRepoPullRequests(namespaceId, repositoryId, gitea.ListPullRequestsOptions{
		ListOptions: gitea.ListOptions{
			Page:     1,
			PageSize: 100,
		},
		State: gitea.StateOpen,
		Sort:  "recentupdate",
	})
	if err != nil {
		return nil, err
	}

	response := []*GitPullRequest{}

	for _, pr := range prList {
		response = append(response, &GitPullRequest{
			Name:   pr.Title,
			Branch: pr.Head.Ref,
		})
	}

	return response, nil
}

func (g *GiteaGitProvider) GetUser() (*GitUser, error) {
	client, err := g.getApiClient()
	if err != nil {
		return nil, err
	}

	user, _, err := client.GetMyUserInfo()
	if user == nil || err != nil {
		return nil, err
	}

	return &GitUser{
		Id:       strconv.FormatInt(user.ID, 10),
		Username: user.UserName,
		Name:     user.FullName,
		Email:    user.Email,
	}, nil
}

func (g *GiteaGitProvider) getApiClient() (*gitea.Client, error) {
	ctx := context.Background()

	return gitea.NewClient(g.baseApiUrl, gitea.SetContext(ctx), gitea.SetToken(g.token))
}
