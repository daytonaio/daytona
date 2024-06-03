// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	bitbucketv1 "github.com/gfleury/go-bitbucket-v1"
	"github.com/mitchellh/mapstructure"
)

type BitbucketServerGitProvider struct {
	*AbstractGitProvider

	username   string
	token      string
	baseApiUrl *string
}

const bitbucketServerResponseLimit = 100

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

	bitbucketDCClient, ok := client.(*bitbucketv1.APIClient)
	if !ok {
		return "", fmt.Errorf("Invalid Bitbucket Data Center/Server client")
	}

	until := ""
	if staticContext.Sha != nil {
		until = *staticContext.Sha
	}

	commits, err := bitbucketDCClient.DefaultApi.GetCommits(staticContext.Id, staticContext.Name, map[string]interface{}{
		"until": until,
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

	pr, err := bitbucketDCClient.DefaultApi.GetPullRequest(staticContext.Id, staticContext.Name, int(*staticContext.PrNumber))
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
	var staticContext StaticGitContext

	re := regexp.MustCompile(`(https?://[^/]+)/projects/([^/]+)/repos/([^/]+)(?:/([^/]+))?(?:/([^/]+))?`)
	matches := re.FindStringSubmatch(repoUrl)

	if len(matches) < 4 {
		return nil, fmt.Errorf("Could not extract project key and repo name from URL: %s", repoUrl)
	}

	baseUrl := matches[1]
	projectKey := matches[2]
	repoName := matches[3]
	// action is either 'pull-requests', 'browse', 'commits', 'branches'
	action := matches[4]
	// identifier is either pull request number or path or commit SHA
	identifier := matches[5]

	staticContext.Id = projectKey
	staticContext.Name = repoName
	staticContext.Owner = projectKey
	// For '.git' or repo clone over https format, refer https://community.atlassian.com/t5/Bitbucket-questions/Project-key-in-repositories-URL/qaq-p/578207
	// and https://community.atlassian.com/t5/Bitbucket-questions/remote-url-in-Bitbucket-server-what-does-scm-represent-is-it/qaq-p/2060987
	staticContext.Url = fmt.Sprintf("%s/scm/%s/%s.git", baseUrl, projectKey, repoName)
	staticContext.Source = strings.TrimPrefix(baseUrl, "https://")

	switch action {
	case "pull-requests":
		if prNumber, err := strconv.Atoi(identifier); err == nil {
			prUint := uint32(prNumber)
			staticContext.PrNumber = &prUint
		}
	case "browse":
		if identifier != "" {
			staticContext.Path = &identifier
		}
	case "commits":
		if identifier != "" {
			staticContext.Sha = &identifier
		}
	case "branches":
		if identifier != "" {
			staticContext.Branch = &identifier
		}
	default:
		if strings.Contains(repoUrl, "commits?until=") {
			parts := strings.Split(repoUrl, "commits?until=")
			if len(parts) == 2 {
				sha := parts[1]
				staticContext.Sha = &sha
			}
		}
	}

	return &staticContext, nil
}
