// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/ktrysmt/go-bitbucket"
)

type BitbucketGitProvider struct {
	*AbstractGitProvider

	username string
	token    string
}

func NewBitbucketGitProvider(username string, token string) *BitbucketGitProvider {
	provider := &BitbucketGitProvider{
		username:            username,
		token:               token,
		AbstractGitProvider: &AbstractGitProvider{},
	}
	provider.AbstractGitProvider.GitProvider = provider

	return provider
}

func (g *BitbucketGitProvider) GetNamespaces() ([]*GitNamespace, error) {
	client := g.getApiClient()
	wsList, err := client.Workspaces.List()
	if err != nil {
		return nil, err
	}

	namespaces := []*GitNamespace{}

	for _, org := range wsList.Workspaces {
		namespace := &GitNamespace{}
		namespace.Id = org.Slug
		namespace.Name = org.Name
		namespaces = append(namespaces, namespace)
	}

	return namespaces, nil
}

func (g *BitbucketGitProvider) GetRepositories(namespace string) ([]*GitRepository, error) {
	client := g.getApiClient()
	var response []*GitRepository

	if namespace == personalNamespaceId {
		user, err := g.GetUser()
		if err != nil {
			return nil, err
		}
		namespace = user.Username
	}

	repoList, err := client.Repositories.ListForAccount(&bitbucket.RepositoriesOptions{
		Owner:   namespace,
		Page:    &[]int{1}[0],
		Keyword: nil,
	})
	if err != nil {
		return nil, err
	}

	for _, repo := range repoList.Items {
		htmlLink, ok := repo.Links["html"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("Invalid repo links")
		}

		repoUrl, ok := htmlLink["href"].(string)
		if !ok {
			return nil, fmt.Errorf("Invalid repo html link")
		}

		u, err := url.Parse(repoUrl)
		if err != nil {
			return nil, err
		}

		owner, name, err := g.getOwnerAndRepoFromFullName(repo.Full_name)
		if err != nil {
			return nil, err
		}

		response = append(response, &GitRepository{
			Id:     repo.Full_name,
			Name:   name,
			Url:    repoUrl,
			Source: u.Host,
			Owner:  owner,
		})
	}

	return response, err
}

func (g *BitbucketGitProvider) GetRepoBranches(repositoryId string, namespaceId string) ([]*GitBranch, error) {
	client := g.getApiClient()
	var response []*GitBranch

	owner, repo, err := g.getOwnerAndRepoFromFullName(repositoryId)
	if err != nil {
		return nil, err
	}

	branches, err := client.Repositories.Repository.ListBranches(&bitbucket.RepositoryBranchOptions{
		RepoSlug: repo,
		Owner:    owner,
	})
	if err != nil {
		return nil, err
	}

	for _, branch := range branches.Branches {
		hash, ok := branch.Target["hash"].(string)
		if !ok {
			return nil, fmt.Errorf("Invalid branch hash")
		}

		response = append(response, &GitBranch{
			Name: branch.Name,
			Sha:  hash,
		})
	}

	return response, nil
}

func (g *BitbucketGitProvider) GetRepoPRs(repositoryId string, namespaceId string) ([]*GitPullRequest, error) {
	client := g.getApiClient()
	var response []*GitPullRequest

	owner, repo, err := g.getOwnerAndRepoFromFullName(repositoryId)
	if err != nil {
		return nil, err
	}

	prList, err := client.Repositories.PullRequests.Get(&bitbucket.PullRequestsOptions{
		Owner:    owner,
		RepoSlug: repo,
	})
	if err != nil {
		return nil, err
	}

	marshalled, err := json.Marshal(prList)
	if err != nil {
		return nil, err
	}

	var prResponse prResponseData
	err = json.Unmarshal(marshalled, &prResponse)
	if err != nil {
		return nil, err
	}

	for _, pr := range prResponse.Values {
		htmlLink, ok := pr.Source.Repository.Links["html"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("Invalid repo links")
		}

		repoUrl, ok := htmlLink["href"].(string)
		if !ok {
			return nil, fmt.Errorf("Invalid repo html link")
		}

		response = append(response, &GitPullRequest{
			Name:            pr.Title,
			Branch:          pr.Source.Branch.Name,
			Sha:             pr.Source.Commit.Hash,
			SourceRepoId:    pr.Source.Repository.Full_name,
			SourceRepoUrl:   repoUrl,
			SourceRepoOwner: owner,
			SourceRepoName:  repo,
		})
	}

	return response, nil
}

func (g *BitbucketGitProvider) GetUser() (*GitUser, error) {
	client := g.getApiClient()

	user, err := client.User.Profile()
	if err != nil {
		return nil, err
	}

	response := &GitUser{}
	response.Id = user.AccountId
	response.Username = user.Username
	response.Name = user.DisplayName

	emails, err := client.User.Emails()
	if err != nil {
		return response, err
	}

	if emails != nil {
		userEmail, ok := emails.(map[string]interface{})
		if ok {
			response.Email = userEmail["values"].([]interface{})[0].(map[string]interface{})["email"].(string)
		}
	}

	return response, nil
}

func (g *BitbucketGitProvider) GetLastCommitSha(staticContext *StaticGitContext) (string, error) {
	client := g.getApiClient()

	branch := ""
	if staticContext.Branch != nil {
		branch = *staticContext.Branch
	}

	include := ""
	if staticContext.Sha != nil {
		include = *staticContext.Sha
	}

	commits, err := client.Repositories.Commits.GetCommits(&bitbucket.CommitsOptions{
		Owner:       staticContext.Owner,
		RepoSlug:    staticContext.Id,
		Branchortag: branch,
		Include:     include,
	})

	if err != nil {
		return "", err
	}

	commitsResponse, ok := commits.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Invalid commits response")
	}

	valuesResponse, ok := commitsResponse["values"].([]interface{})
	if !ok {
		return "", fmt.Errorf("Invalid commits values")
	}

	commit := valuesResponse[0]
	commitResponse, ok := commit.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Invalid commit response")
	}

	commitHash, ok := commitResponse["hash"].(string)
	if !ok {
		return "", fmt.Errorf("Invalid commit hash")
	}

	return commitHash, nil
}

func (g *BitbucketGitProvider) GetBranchByCommit(staticContext *StaticGitContext) (string, error) {
	client := g.getApiClient()

	branches, err := client.Repositories.Repository.ListBranches(&bitbucket.RepositoryBranchOptions{
		RepoSlug: staticContext.Name,
		Owner:    staticContext.Owner,
	})
	if err != nil {
		return "", err
	}

	var branchName string
	for _, branch := range branches.Branches {
		hash, ok := branch.Target["hash"].(string)
		if !ok {
			continue
		}

		if hash == *staticContext.Sha {
			branchName = branch.Name
			break
		}

		commits, err := client.Repositories.Commits.GetCommits(&bitbucket.CommitsOptions{
			RepoSlug:    staticContext.Name,
			Owner:       staticContext.Owner,
			Branchortag: branch.Name,
		})
		if err != nil {
			return "", err
		}
		commitsResponse, ok := commits.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("Invalid commits response")
		}

		valuesResponse, ok := commitsResponse["values"].([]interface{})
		if !ok {
			return "", fmt.Errorf("Invalid commits values")
		}

		if len(valuesResponse) == 0 {
			continue
		}

		for _, commit := range valuesResponse {
			commitResponse, ok := commit.(map[string]interface{})
			if !ok {
				return "", fmt.Errorf("Invalid commit response")
			}

			commitHash, ok := commitResponse["hash"].(string)
			if !ok {
				return "", fmt.Errorf("Invalid commit hash")
			}
			if commitHash == *staticContext.Sha {
				branchName = branch.Name
				break
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

func (g *BitbucketGitProvider) GetUrlFromRepository(repository *GitRepository) string {
	url := strings.TrimSuffix(repository.Url, ".git")

	if repository.Branch != nil && *repository.Branch != "" {
		if repository.Path != nil {
			url += "/src/" + *repository.Branch + "/" + *repository.Path
		} else if repository.Sha == *repository.Branch {
			url += "/commit/" + *repository.Branch
		} else {
			url += "/branch/" + *repository.Branch
		}
	} else if repository.Path != nil {
		url += "/src/main/" + *repository.Path
	}

	return url
}

func (g *BitbucketGitProvider) getApiClient() *bitbucket.Client {
	client := bitbucket.NewBasicAuth(g.username, g.token)
	return client
}

func (g *BitbucketGitProvider) GetPrContext(staticContext *StaticGitContext) (*StaticGitContext, error) {
	if staticContext.PrNumber == nil {
		return staticContext, nil
	}

	repo := *staticContext

	client := g.getApiClient()

	pr, err := client.Repositories.PullRequests.Get(&bitbucket.PullRequestsOptions{
		Owner:    staticContext.Owner,
		RepoSlug: staticContext.Id,
		ID:       fmt.Sprint(*staticContext.PrNumber),
	})
	if err != nil {
		return nil, err
	}

	prMap := pr.(map[string]interface{})
	source, ok := prMap["source"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Invalid PR source")
	}

	repository, ok := source["repository"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Invalid PR repository")
	}

	fullName, ok := repository["full_name"].(string)
	if !ok {
		return nil, fmt.Errorf("Invalid PR repository full name")
	}

	parts := strings.Split(fullName, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid PR repository full name")
	}

	repo.Owner = parts[0]
	repo.Name = parts[1]
	repo.Id = fullName

	branch, ok := source["branch"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Invalid PR branch")
	}

	branchName, ok := branch["name"].(string)
	if !ok {
		return nil, fmt.Errorf("Invalid PR branch name")
	}

	repo.Branch = &branchName

	return &repo, nil
}

func (g *BitbucketGitProvider) ParseStaticGitContext(repoUrl string) (*StaticGitContext, error) {
	staticContext, err := g.AbstractGitProvider.ParseStaticGitContext(repoUrl)
	if err != nil {
		return nil, err
	}

	if staticContext.Path == nil {
		return staticContext, nil
	}

	parts := strings.Split(*staticContext.Path, "/")

	switch {
	case len(parts) >= 2 && parts[0] == "pull-requests":
		prNumber, _ := strconv.Atoi(parts[1])
		prUint := uint32(prNumber)
		staticContext.PrNumber = &prUint
		staticContext.Path = nil
	case len(parts) >= 1 && (parts[0] == "src" || parts[0] == "branch"):
		staticContext.Branch = &parts[1]
		if len(parts) > 2 {
			path := strings.Join(parts[2:], "/")
			staticContext.Path = &path
		} else {
			staticContext.Path = nil
		}
	case len(parts) >= 3 && parts[0] == "commits" && parts[1] == "branch":
		staticContext.Branch = &parts[2]
		staticContext.Path = nil
	case len(parts) >= 2 && parts[0] == "commits":
		staticContext.Sha = &parts[1]
		staticContext.Branch = staticContext.Sha
		staticContext.Path = nil
	}

	return staticContext, nil
}

func (g *BitbucketGitProvider) GetDefaultBranch(staticContext *StaticGitContext) (*string, error) {
	client := g.getApiClient()
	branches, err := client.Repositories.Repository.ListBranches(&bitbucket.RepositoryBranchOptions{
		Owner:    staticContext.Owner,
		RepoSlug: staticContext.Id,
	})
	if err != nil {
		return nil, err
	}

	for _, branch := range branches.Branches {
		if branch.Type == "main" {
			return &branch.Name, nil
		}
	}

	return nil, fmt.Errorf("Default branch not found")
}

func (b *BitbucketGitProvider) getOwnerAndRepoFromFullName(fullName string) (string, string, error) {
	parts := strings.Split(fullName, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("Invalid full name")
	}

	name := parts[len(parts)-1]

	owner := strings.Join(parts[:len(parts)-1], "/")

	return owner, name, nil
}

type prResponseData struct {
	Values []struct {
		Title  string `json:"title"`
		Source struct {
			Branch struct {
				Name string `json:"name"`
			} `json:"branch"`
			Commit struct {
				Hash string `json:"hash"`
			} `json:"commit"`
			Repository struct {
				UUID      string                 `json:"uuid"`
				Links     map[string]interface{} `json:"links"`
				Full_name string                 `json:"full_name"`
			} `json:"repository"`
		} `json:"source"`
	} `json:"values"`
}
