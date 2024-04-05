// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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
	username string
	token    string
}

func NewBitbucketGitProvider(username string, token string) *BitbucketGitProvider {
	return &BitbucketGitProvider{
		username: username,
		token:    token,
	}
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
			log.Fatal("Invalid HTML link")
		}

		url := htmlLink["href"].(string)
		repoSlug := url[strings.LastIndex(url, "/")+1:]

		response = append(response, &GitRepository{
			Id:   repoSlug,
			Name: repo.Name,
			Url:  url,
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
			SHA:  branch.Target.Hash,
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

func (g *BitbucketGitProvider) getApiClient() *bitbucket.Client {
	client := bitbucket.NewBasicAuth(g.username, g.token)
	return client
}
