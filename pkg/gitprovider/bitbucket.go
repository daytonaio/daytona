// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ktrysmt/go-bitbucket"
)

type BranchesResponse struct {
	Values []BranchResponse `json:"values"`
}

type BranchResponse struct {
	Name   string `json:"name"`
	Target struct {
		Hash string `json:"hash"`
	} `json:"target"`
}

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
	user, err := g.GetUser()
	if err != nil {
		return nil, err
	}

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

	namespaces = append([]*GitNamespace{{Id: personalNamespaceId, Name: user.Username}}, namespaces...)

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

		repoUrl := htmlLink["href"].(string)
		repoSlug := repoUrl[strings.LastIndex(repoUrl, "/")+1:]

		u, err := url.Parse(repoUrl)
		if err != nil {
			return nil, err
		}

		ownerUsername, ok := repo.Owner["username"].(string)
		if !ok {
			return nil, fmt.Errorf("Invalid repo owner")
		}

		response = append(response, &GitRepository{
			Id:     repoSlug,
			Name:   repo.Name,
			Url:    repoUrl,
			Source: u.Host,
			Owner:  ownerUsername,
		})
	}

	return response, err
}

func (g *BitbucketGitProvider) GetRepoBranches(repositoryId string, namespaceId string) ([]*GitBranch, error) {
	client := g.getApiClient()
	var response []*GitBranch

	if namespaceId == personalNamespaceId {
		user, err := g.GetUser()
		if err != nil {
			return nil, err
		}
		namespaceId = user.Username
	}

	// Custom API call implementation

	authString := fmt.Sprintf("%s:%s", g.username, g.token)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(authString))

	url := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/refs/branches", namespaceId, repositoryId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	req.Header.Set("Authorization", "Basic "+encodedAuth)

	resp, err := client.HttpClient.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	var branchesResponse BranchesResponse

	// Unmarshal JSON into the Branches slice
	err = json.Unmarshal(body, &branchesResponse)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	// Now you can work with the branches
	for _, branch := range branchesResponse.Values {
		response = append(response, &GitBranch{
			Name: branch.Name,
			Sha:  branch.Target.Hash,
		})
	}

	return response, nil
}

func (g *BitbucketGitProvider) GetRepoPRs(repositoryId string, namespaceId string) ([]*GitPullRequest, error) {
	client := g.getApiClient()
	var response []*GitPullRequest

	if namespaceId == personalNamespaceId {
		user, err := g.GetUser()
		if err != nil {
			return nil, err
		}
		namespaceId = user.Username
	}

	prList, err := client.Repositories.PullRequests.Get(&bitbucket.PullRequestsOptions{
		Owner:    namespaceId,
		RepoSlug: repositoryId,
	})

	if err != nil {
		return nil, err
	}

	// TODO: Implement this
	fmt.Println(prList)
	return response, err
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

	links, ok := repository["links"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Invalid PR repository links")
	}

	htmlLink, ok := links["html"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Invalid PR repository html link")
	}

	url := htmlLink["href"].(string)
	repoSlug := url[strings.LastIndex(url, "/")+1:]

	repo.Owner = parts[0]
	repo.Name = parts[1]
	repo.Id = repoSlug

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
	case len(parts) >= 1 && parts[0] == "src":
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
