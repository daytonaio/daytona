// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	bitbucketv1 "github.com/gfleury/go-bitbucket-v1"
	bitbucket "github.com/ktrysmt/go-bitbucket"
	"github.com/mitchellh/mapstructure"
)

type BitbucketGitProvider struct {
	*AbstractGitProvider

	username string
	token    string
}

type BitbucketServerGitProvider struct {
	*AbstractGitProvider

	username   string
	token      string
	baseApiUrl *string
}

const bitbucketServerResponseLimit = 100

func NewBitbucketGitProvider(username string, token string) *BitbucketGitProvider {
	provider := &BitbucketGitProvider{
		username:            username,
		token:               token,
		AbstractGitProvider: &AbstractGitProvider{},
	}
	provider.AbstractGitProvider.GitProvider = provider

	return provider
}

func NewBitbucketServerGitProvider(username string, token string, baseApiUrl *string) *BitbucketServerGitProvider {
	provider := &BitbucketServerGitProvider{
		username:            username,
		token:               token,
		AbstractGitProvider: &AbstractGitProvider{},
		baseApiUrl:          baseApiUrl,
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

func (g *BitbucketGitProvider) getApiClient() *bitbucket.Client {
	client := bitbucket.NewBasicAuth(g.username, g.token)
	return client
}

func (g *BitbucketGitProvider) getPrContext(staticContext *StaticGitContext) (*StaticGitContext, error) {
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

func (g *BitbucketGitProvider) parseStaticGitContext(repoUrl string) (*StaticGitContext, error) {
	staticContext, err := g.AbstractGitProvider.parseStaticGitContext(repoUrl)
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

func (g *BitbucketServerGitProvider) getApiClient() interface{} {
	conf := bitbucketv1.NewConfiguration(*g.baseApiUrl)
	ctx := context.WithValue(context.Background(), bitbucketv1.ContextBasicAuth, bitbucketv1.BasicAuth{
		UserName: g.username,
		Password: g.token,
	})
	client := bitbucketv1.NewAPIClient(ctx, conf)
	return client
}

func (g *BitbucketServerGitProvider) GetNamespaces() ([]*GitNamespace, error) {
	client := g.getApiClient()
	var namespaces []*GitNamespace

	// Bitbucket Data Center/Server
	bitbucketDCClient, ok := client.(*bitbucketv1.APIClient)
	if !ok {
		return nil, fmt.Errorf("Invalid Bitbucket Data Center/Server client")
	}

	projectsRaw, err := bitbucketDCClient.DefaultApi.GetProjects(map[string]any{
		"limit": bitbucketServerResponseLimit,
	})
	if err != nil {
		return nil, err
	}

	projectsRaw.Body.Close()

	projects, err := bitbucketv1.GetProjectsResponse(projectsRaw)
	if err != nil {
		return nil, err
	}

	for _, project := range projects {
		namespace := &GitNamespace{}
		namespace.Id = project.Key
		namespace.Name = project.Name

		namespaces = append(namespaces, namespace)
	}

	return namespaces, nil
}

func (g *BitbucketServerGitProvider) GetRepositories(namespace string) ([]*GitRepository, error) {
	client := g.getApiClient()
	var response []*GitRepository

	bitbucketDCClient, ok := client.(*bitbucketv1.APIClient)
	if !ok {
		return nil, fmt.Errorf("Invalid Bitbucket Data Center/Server client")
	}

	start := 0
	for {
		var repoList *bitbucketv1.APIResponse
		var err error
		if namespace == personalNamespaceId {
			repoList, err = bitbucketDCClient.DefaultApi.GetRepositories_19(nil)
		} else {
			repoList, err = bitbucketDCClient.DefaultApi.GetRepositoriesWithOptions(namespace, map[string]interface{}{
				"start": start,
			})
		}

		if err != nil {
			return nil, err
		}

		pageRepos, err := bitbucketv1.GetRepositoriesResponse(repoList)
		if err != nil {
			return nil, err
		}

		for _, repo := range pageRepos {
			var repoUrl string
			for _, link := range repo.Links.Clone {
				if link.Name == "https" {
					repoUrl = link.Href
					break
				}
			}

			response = append(response, &GitRepository{
				Id:     repo.Slug,
				Name:   repo.Name,
				Url:    repoUrl,
				Source: *g.baseApiUrl,
				Owner:  repo.Owner.Name,
			})
		}

		hasNextPage, nextPageStart := bitbucketv1.HasNextPage(repoList)
		if !hasNextPage {
			break
		}
		start = nextPageStart
	}

	return response, nil
}

func (g *BitbucketServerGitProvider) GetRepoBranches(repositoryId string, namespaceId string) ([]*GitBranch, error) {
	client := g.getApiClient()
	var response []*GitBranch

	bitbucketDCClient, ok := client.(*bitbucketv1.APIClient)
	if !ok {
		return nil, fmt.Errorf("Invalid Bitbucket Data Center/Server client")
	}
	branches, err := bitbucketDCClient.DefaultApi.GetBranches(namespaceId, repositoryId, nil)
	if err != nil {
		return nil, err
	}

	branchList, err := bitbucketv1.GetBranchesResponse(branches)
	if err != nil {
		return nil, err
	}

	for _, branch := range branchList {
		response = append(response, &GitBranch{
			Name: branch.DisplayID,
			Sha:  branch.LatestCommit,
		})
	}

	return response, nil
}

func (g *BitbucketServerGitProvider) GetRepoPRs(repositoryId string, namespaceId string) ([]*GitPullRequest, error) {
	client := g.getApiClient()
	var response []*GitPullRequest

	bitbucketDCClient, ok := client.(*bitbucketv1.APIClient)
	if !ok {
		return nil, fmt.Errorf("Invalid Bitbucket Data Center/Server client")
	}

	prList, err := bitbucketDCClient.DefaultApi.GetPullRequests(nil)
	if err != nil {
		return nil, err
	}

	pullRequest, err := bitbucketv1.GetPullRequestsResponse(prList)
	if err != nil {
		return nil, err
	}

	for _, pr := range pullRequest {
		var repoUrl string
		for _, link := range pr.FromRef.Repository.Links.Clone {
			if link.Name == "https" {
				repoUrl = link.Href
				break
			}
		}

		response = append(response, &GitPullRequest{
			Name:            pr.Title,
			Branch:          pr.FromRef.DisplayID,
			Sha:             pr.FromRef.LatestCommit,
			SourceRepoId:    pr.FromRef.Repository.Slug,
			SourceRepoUrl:   repoUrl,
			SourceRepoOwner: pr.FromRef.Repository.Owner.Name,
			SourceRepoName:  pr.FromRef.Repository.Name,
		})
	}

	return response, nil
}

func (g *BitbucketServerGitProvider) GetUser() (*GitUser, error) {
	client := g.getApiClient()

	bitbucketDCClient, ok := client.(*bitbucketv1.APIClient)
	if !ok {
		return nil, fmt.Errorf("Invalid Bitbucket Data Center/Server client")
	}

	// Since BitbucketServer or gfleury/go-bitbucket-v1 doesn't offer an endpoint to query the
	// currently authenticated user, We instead query the '/rest/api/1.0/application-properties' endpoint
	// which does not put load on the server and then extract the username from the response header.
	// Refer this developer community comment: https://community.developer.atlassian.com/t/obtain-authorised-users-username-from-api/24422/2
	res, err := bitbucketDCClient.DefaultApi.GetApplicationProperties()
	if err != nil {
		return nil, err
	}

	username := res.Header.Get("X-Ausername")
	if username == "" {
		return nil, fmt.Errorf("X-Ausername header is missing")
	}

	user, err := bitbucketDCClient.DefaultApi.GetUser(username)
	if err != nil {
		return nil, err
	}

	var userInfo bitbucketv1.User
	err = mapstructure.Decode(user.Values, &userInfo)

	if err != nil {
		return nil, err
	}

	response := &GitUser{}
	response.Id = fmt.Sprintf("%d", userInfo.ID)
	response.Username = username
	response.Name = userInfo.DisplayName

	return response, nil
}

func (g *BitbucketServerGitProvider) GetLastCommitSha(staticContext *StaticGitContext) (string, error) {
	client := g.getApiClient()

	branch := ""
	if staticContext.Branch != nil {
		branch = *staticContext.Branch
	}

	bitbucketDCClient, ok := client.(*bitbucketv1.APIClient)
	if !ok {
		return "", fmt.Errorf("Invalid Bitbucket Data Center/Server client")
	}

	commits, err := bitbucketDCClient.DefaultApi.GetCommits(staticContext.ProjectKey, staticContext.Id, map[string]interface{}{
		"until": branch,
	})

	if err != nil {
		return "", err
	}

	if len(commits.Values) == 0 {
		return "", fmt.Errorf("No commits found")
	}

	commitList, err := bitbucketv1.GetCommitsResponse(commits)
	if err != nil {
		return "", err
	}

	return commitList[0].ID, nil
}

func (g *BitbucketServerGitProvider) getPrContext(staticContext *StaticGitContext) (*StaticGitContext, error) {
	if staticContext.PrNumber == nil {
		return staticContext, nil
	}

	repo := *staticContext

	client := g.getApiClient()

	bitbucketDCClient, ok := client.(*bitbucketv1.APIClient)
	if !ok {
		return nil, fmt.Errorf("Invalid Bitbucket Data Center/Server client")
	}

	pr, err := bitbucketDCClient.DefaultApi.GetPullRequest(staticContext.ProjectKey, staticContext.Id, int(*staticContext.PrNumber))
	if err != nil {
		return nil, err
	}

	prInfo, err := bitbucketv1.GetPullRequestResponse(pr)
	if err != nil {
		return nil, err
	}

	repo.Owner = prInfo.FromRef.Repository.Owner.DisplayName
	repo.Name = prInfo.FromRef.Repository.Slug
	repo.Id = prInfo.FromRef.Repository.Slug
	repo.Branch = &prInfo.FromRef.DisplayID

	return &repo, nil
}

func (g *BitbucketServerGitProvider) parseStaticGitContext(repoUrl string) (*StaticGitContext, error) {
	staticContext, err := g.AbstractGitProvider.parseStaticGitContext(repoUrl)
	if err != nil {
		return nil, err
	}

	urlParts := strings.Split(repoUrl, "/")
	if len(urlParts) < 5 {
		return nil, fmt.Errorf("Invalid repository URL")
	}

	// Example URL: https://bitbucket.example.com/projects/<PROJECT_KEY>/repos/<REPO_NAME>
	var projectKey, repoName string
	for i, part := range urlParts {
		if strings.ToLower(part) == "projects" && i+2 < len(urlParts) {
			projectKey = urlParts[i+1]
			repoName = urlParts[i+3]
			break
		}
	}

	if projectKey == "" || repoName == "" {
		return nil, fmt.Errorf("Could not extract project key and repository name from URL")
	}

	staticContext.ProjectKey = projectKey
	staticContext.Id = repoName

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
