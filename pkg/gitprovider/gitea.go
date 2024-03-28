// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"
	"strconv"

	"github.com/daytonaio/daytona/pkg/types"

	"code.gitea.io/sdk/gitea"
)

type GiteaGitProvider struct {
	token      string
	baseApiUrl string
}

func (g *GiteaGitProvider) GetNamespaces() ([]types.GitNamespace, error) {
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

	namespaces := make([]types.GitNamespace, len(orgList)+1) // +1 for the user namespace
	namespaces[0] = types.GitNamespace{Id: personalNamespaceId, Name: user.Username}

	for i, org := range orgList {
		namespaces[i+1] = types.GitNamespace{Id: org.UserName, Name: org.UserName}
	}

	return namespaces, nil
}

func (g *GiteaGitProvider) GetRepositories(namespace string) ([]types.GitRepository, error) {
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

	response := make([]types.GitRepository, 0, len(repoList))

	for _, repo := range repoList {
		response = append(response, types.GitRepository{
			Id:   repo.Name,
			Name: repo.Name,
			Url:  repo.HTMLURL,
		})
	}

	return response, err
}

func (g *GiteaGitProvider) GetRepoBranches(repositoryId string, namespaceId string) ([]types.GitBranch, error) {
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

	response := make([]types.GitBranch, 0, len(repoBranches))

	for _, branch := range repoBranches {
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

func (g *GiteaGitProvider) GetRepoPRs(repositoryId string, namespaceId string) ([]types.GitPullRequest, error) {
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

	response := make([]types.GitPullRequest, 0, len(prList))

	for _, pr := range prList {
		response = append(response, types.GitPullRequest{
			Name:   pr.Title,
			Branch: pr.Head.Ref,
		})
	}

	return response, nil
}

func (g *GiteaGitProvider) GetUser() (types.GitUser, error) {
	client, err := g.getApiClient()
	if err != nil {
		return types.GitUser{}, err
	}

	user, _, err := client.GetMyUserInfo()
	if user == nil || err != nil {
		return types.GitUser{}, err
	}

	return types.GitUser{
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
