// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	gitnessclient "github.com/daytonaio/daytona/pkg/gitprovider/gitnessclient"
)

type GitnessGitProvider struct {
	*AbstractGitProvider
	token      string
	baseApiUrl string
}

func NewGitnessGitProvider(token string, baseApiUrl string) *GitnessGitProvider {
	gitProvider := &GitnessGitProvider{
		token:               token,
		baseApiUrl:          baseApiUrl,
		AbstractGitProvider: &AbstractGitProvider{},
	}
	gitProvider.AbstractGitProvider.GitProvider = gitProvider

	return gitProvider
}

func (g *GitnessGitProvider) CanHandle(repoUrl string) (bool, error) {
	staticContext, err := g.ParseStaticGitContext(repoUrl)
	if err != nil {
		return false, err
	}

	return strings.Contains(g.baseApiUrl, staticContext.Source), nil
}

func (g *GitnessGitProvider) GetNamespaces(options ListOptions) ([]*GitNamespace, error) {
	client := g.getApiClient()
	response, err := client.GetSpaces()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Namespace : %w", err)
	}
	var namespaces []*GitNamespace
	for _, membership := range response {
		namespace := &GitNamespace{
			Id:   membership.Space.Identifier,
			Name: membership.Space.Identifier,
		}
		namespaces = append(namespaces, namespace)
	}
	return namespaces, nil
}

func (g *GitnessGitProvider) getApiClient() *gitnessclient.GitnessClient {
	url, _ := url.Parse(g.baseApiUrl)
	return gitnessclient.NewGitnessClient(g.token, url)
}

func (g *GitnessGitProvider) GetRepositories(namespace string, options ListOptions) ([]*GitRepository, error) {
	client := g.getApiClient()
	response, err := client.GetRepositories(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Repositories : %w", err)
	}
	admin, err := client.GetSpaceAdmin(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Repositories : %w", err)
	}
	var repos []*GitRepository
	for _, repo := range response {
		u, err := url.Parse(repo.GitUrl)
		if err != nil {
			return nil, err
		}

		repo := &GitRepository{
			Id:     repo.Identifier,
			Name:   repo.Identifier,
			Url:    repo.GitUrl,
			Branch: repo.DefaultBranch,
			Source: u.Host,
			Owner:  admin.Principal.DisplayName,
		}
		repos = append(repos, repo)
	}
	return repos, nil
}

func (g *GitnessGitProvider) GetRepoBranches(repositoryId string, namespaceId string, options ListOptions) ([]*GitBranch, error) {
	client := g.getApiClient()
	response, err := client.GetRepoBranches(repositoryId, namespaceId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Branches: %w", err)
	}
	var branches []*GitBranch
	for _, branch := range response {
		repobranch := &GitBranch{
			Name: branch.Name,
			Sha:  branch.Sha,
		}
		branches = append(branches, repobranch)
	}
	return branches, nil
}

func (g *GitnessGitProvider) GetRepoPRs(repositoryId string, namespaceId string, options ListOptions) ([]*GitPullRequest, error) {
	client := g.getApiClient()
	response, err := client.GetRepoPRs(repositoryId, namespaceId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Pull Request : %w", err)
	}
	var pullRequests []*GitPullRequest
	for _, pr := range response {
		pullRequest := &GitPullRequest{
			Name:            pr.Title,
			Branch:          pr.SourceBranch,
			Sha:             pr.SourceSha,
			SourceRepoId:    fmt.Sprintf("%d", pr.SourceRepoId),
			SourceRepoUrl:   gitnessclient.GetCloneUrl(client.BaseURL.Scheme, client.BaseURL.Host, namespaceId, repositoryId),
			SourceRepoOwner: pr.Author.DisplayName,
			SourceRepoName:  repositoryId,
		}
		pullRequests = append(pullRequests, pullRequest)
	}
	return pullRequests, nil
}

func (g *GitnessGitProvider) GetUser() (*GitUser, error) {
	client := g.getApiClient()
	response, err := client.GetUser()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch User : %w", err)
	}
	user := &GitUser{
		Id:       response.UID,
		Username: response.UID,
		Name:     response.DisplayName,
		Email:    response.Email,
	}
	return user, nil
}

func (g *GitnessGitProvider) GetUrlFromContext(repoContext *GetRepositoryContext) string {
	url := strings.TrimSuffix(repoContext.Url, ".git")

	if repoContext.Branch != nil && *repoContext.Branch != "" {
		url += "/files/" + *repoContext.Branch

		if repoContext.Path != nil && *repoContext.Path != "" {
			url += "/~/" + *repoContext.Path
		}
	} else if repoContext.Path != nil {
		url += "/files/main/~/" + *repoContext.Path
	}

	return url
}

func (g *GitnessGitProvider) GetBranchByCommit(staticContext *StaticGitContext) (string, error) {
	client := g.getApiClient()
	response, err := client.GetRepoBranches(staticContext.Name, staticContext.Owner)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Branches: %w", err)
	}

	var branchName string
	for _, branch := range response {
		if *staticContext.Sha == branch.Sha {
			branchName = branch.Name
			break
		}

		commits, err := client.GetCommits(staticContext.Owner, staticContext.Name, &branch.Name, nil)
		if err != nil {
			return "", err
		}

		if len(*commits) == 0 {
			continue
		}

		for _, commit := range *commits {
			if commit.Sha == *staticContext.Sha {
				branchName = branch.Name
				break
			}
		}
		if branchName != "" {
			break
		}
	}

	if branchName == "" {
		return "", fmt.Errorf("status code: %d branch not found for SHA: %s", http.StatusNotFound, *staticContext.Sha)
	}

	return branchName, nil
}

func (g *GitnessGitProvider) GetLastCommitSha(staticContext *StaticGitContext) (string, error) {
	client := g.getApiClient()
	return client.GetLastCommitSha(staticContext.Owner, staticContext.Name, staticContext.Branch)
}

func (g *GitnessGitProvider) GetPrContext(staticContext *StaticGitContext) (*StaticGitContext, error) {
	if staticContext.PrNumber == nil {
		return staticContext, nil
	}
	client := g.getApiClient()
	pullReq, err := client.GetPr(staticContext.Url, *staticContext.PrNumber)
	if err != nil {
		return nil, err
	}
	repo := *staticContext
	repo.Branch = &pullReq.SourceBranch
	repo.Url = pullReq.GitUrl
	repo.Id = fmt.Sprint(pullReq.SourceRepoID)
	repo.Owner = pullReq.Author.DisplayName
	ref, err := client.GetRepoRef(staticContext.Url)
	if err != nil {
		return nil, err
	}
	prUrl := strings.Split(*ref, "/")
	if len(prUrl) == 2 {
		repo.Name = prUrl[1]
	}
	return &repo, nil
}

func (g *GitnessGitProvider) ParseStaticGitContext(repoUrl string) (*StaticGitContext, error) {
	staticContext, err := g.AbstractGitProvider.ParseStaticGitContext(repoUrl)
	if err != nil {
		return nil, err
	}
	if strings.Contains(repoUrl, "/git/") {
		ref, err := g.getApiClient().GetRepoRef(repoUrl)
		if err != nil {
			return nil, err
		}
		refParts := strings.Split(*ref, "/")
		staticContext.Owner = refParts[0]
		staticContext.Name = refParts[1]
		staticContext.Id = refParts[1]
		staticContext.Path = nil
	}

	parsedUrl, err := url.Parse(repoUrl)
	if err != nil {
		return nil, err
	}
	staticContext.Url = gitnessclient.GetCloneUrl(parsedUrl.Scheme, parsedUrl.Host, staticContext.Owner, staticContext.Name)
	if staticContext.Path == nil {
		return staticContext, nil
	}

	parts := strings.Split(*staticContext.Path, "/")

	switch {

	case len(parts) >= 2 && parts[0] == "pulls":
		prNumber, _ := strconv.Atoi(parts[1])
		prUint := uint32(prNumber)
		staticContext.PrNumber = &prUint
		staticContext.Path = nil

	case len(parts) == 4 && parts[0] == "files" && parts[2] == "~":
		staticContext.Branch = &parts[1]
		branchPath := strings.Join(parts[2:], "/")
		staticContext.Path = &branchPath

	case len(parts) >= 1 && parts[0] == "files" && parts[1] != "~":
		staticContext.Branch = &parts[1]
		staticContext.Path = nil

	case len(parts) >= 2 && parts[0] == "commits":
		staticContext.Branch = &parts[1]
		staticContext.Path = nil

	case len(parts) >= 2 && parts[0] == "commit":
		staticContext.Sha = &parts[1]
		staticContext.Branch = staticContext.Sha
		staticContext.Path = nil
	}

	return staticContext, nil
}

func (g *GitnessGitProvider) GetDefaultBranch(staticContext *StaticGitContext) (*string, error) {
	client := g.getApiClient()
	return client.GetDefaultBranch(staticContext.Url)
}

func (g *GitnessGitProvider) RegisterPrebuildWebhook(repo *GitRepository, endpointUrl string) (string, error) {
	client := g.getApiClient()
	webhook, err := client.CreateWebhook(repo.Id, repo.Owner, gitnessclient.Webhook{
		Triggers:    []string{"branch_updated"},
		Url:         endpointUrl,
		Identifier:  "daytona-webhook_" + repo.Id,
		DisplayName: "Daytona Webhook",
		Enabled:     true,
	})
	if err != nil {
		return "", err
	}
	return webhook.Uid, nil
}

func (g *GitnessGitProvider) GetPrebuildWebhook(repo *GitRepository, endpointUrl string) (*string, error) {
	client := g.getApiClient()
	webhooks, err := client.GetAllWebhooks(repo.Id, repo.Owner)
	if err != nil {
		return nil, err
	}
	for _, webhook := range webhooks {
		if webhook.Url == endpointUrl {
			return &webhook.Uid, nil
		}
	}
	return nil, nil
}

func (g *GitnessGitProvider) UnregisterPrebuildWebhook(repo *GitRepository, id string) error {
	client := g.getApiClient()
	return client.DeleteWebhook(repo.Id, repo.Owner, id)
}

func (g *GitnessGitProvider) GetCommitsRange(repo *GitRepository, initialSha string, currentSha string) (int, error) {
	client := g.getApiClient()
	commits, err := client.GetCommits(repo.Owner, repo.Name, &repo.Branch, &initialSha)
	if err != nil {
		return 0, err
	}
	return len(*commits), nil
}

func (g *GitnessGitProvider) ParseEventData(request *http.Request) (*GitEventData, error) {
	payload, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}
	var webhookEvent gitnessclient.WebhookEventData
	err = json.Unmarshal(payload, &webhookEvent)
	if err != nil {
		return nil, err
	}

	if webhookEvent.Trigger != "branch_updated" {
		return nil, nil
	}

	gitEventData := &GitEventData{
		Url:    webhookEvent.Repo.GitURL,
		Branch: strings.TrimPrefix(webhookEvent.Ref.Name, "refs/heads/"),
		Sha:    webhookEvent.Sha,
		Owner:  webhookEvent.Principal.DisplayName,
	}

	for _, commit := range webhookEvent.Commits {
		gitEventData.AffectedFiles = append(gitEventData.AffectedFiles, commit.Modified...)
		gitEventData.AffectedFiles = append(gitEventData.AffectedFiles, commit.Added...)
		gitEventData.AffectedFiles = append(gitEventData.AffectedFiles, commit.Removed...)
	}
	return gitEventData, nil
}
