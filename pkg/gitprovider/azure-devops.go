// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/daytonaio/daytona/internal/util"
	azureWebhook "github.com/go-playground/webhooks/v6/azuredevops"
	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/microsoft/azure-devops-go-api/azuredevops/location"
	"github.com/microsoft/azure-devops-go-api/azuredevops/servicehooks"
)

type AzureDevOpsGitProvider struct {
	*AbstractGitProvider

	token      string
	baseApiUrl string
}

func NewAzureDevOpsGitProvider(token string, baseApiUrl string) *AzureDevOpsGitProvider {
	provider := &AzureDevOpsGitProvider{
		token:               token,
		baseApiUrl:          baseApiUrl,
		AbstractGitProvider: &AbstractGitProvider{},
	}
	provider.AbstractGitProvider.GitProvider = provider

	return provider
}

func (g *AzureDevOpsGitProvider) GetNamespaces() ([]*GitNamespace, error) {
	client, _, err := g.getApiClient()
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	pageArgs := core.GetProjectsArgs{}

	namespaces := []*GitNamespace{}

	pages, err := client.GetProjects(ctx, pageArgs)
	for ; err == nil; pages, err = client.GetProjects(ctx, pageArgs) {
		projectsResponse := *pages
		for _, project := range projectsResponse.Value {
			namespaces = append(namespaces, &GitNamespace{Id: project.Id.String(), Name: *project.Name})
		}
		if pages.ContinuationToken == "" {
			return namespaces, nil
		}
		pageArgs = core.GetProjectsArgs{
			ContinuationToken: &pages.ContinuationToken,
		}
	}
	return nil, g.FormatError(err)
}

func (g *AzureDevOpsGitProvider) GetRepositories(namespace string) ([]*GitRepository, error) {
	client, err := g.getGitClient()
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	repos, err := client.GetRepositories(ctx, git.GetRepositoriesArgs{
		Project: &namespace,
	})
	if err != nil {
		return nil, g.FormatError(err)
	}

	repositories := []*GitRepository{}

	for _, repo := range *repos {
		u, err := url.Parse(*repo.WebUrl)
		if err != nil {
			return nil, err
		}
		defaultBranch := *repo.DefaultBranch
		defaultBranch = strings.TrimPrefix(defaultBranch, "refs/heads/")
		owner := g.getOwnerName()

		gitRepo := &GitRepository{
			Id:     repo.Id.String(),
			Name:   *repo.Name,
			Branch: defaultBranch,
			Url:    *repo.WebUrl,
			Source: u.Host,
		}

		if owner != "" {
			gitRepo.Owner = owner
		}

		repositories = append(repositories, gitRepo)
	}

	return repositories, nil
}

func (g *AzureDevOpsGitProvider) GetUser() (*GitUser, error) {
	client := g.getLocationClient()
	ctx := context.Background()
	connectionData, err := client.GetConnectionData(ctx, location.GetConnectionDataArgs{})
	if err != nil {
		return nil, g.FormatError(err)
	}

	user := &GitUser{}
	user.Id = connectionData.AuthenticatedUser.Id.String()
	user.Username = *connectionData.AuthorizedUser.ProviderDisplayName
	user.Name = *connectionData.AuthenticatedUser.ProviderDisplayName

	if props, ok := connectionData.AuthenticatedUser.Properties.(map[string]interface{}); ok {
		if accounts, ok := props["Accounts"].(map[string]interface{}); ok {
			if value, ok := accounts["$value"].(string); ok {
				user.Email = value
			}
		}
	}

	return user, nil
}

func (g *AzureDevOpsGitProvider) GetRepoBranches(repositoryId string, namespaceId string) ([]*GitBranch, error) {
	client, err := g.getGitClient()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	branches, err := client.GetBranches(ctx, git.GetBranchesArgs{
		RepositoryId: &repositoryId,
	})
	if err != nil {
		return nil, g.FormatError(err)
	}

	var response []*GitBranch

	for _, branch := range *branches {
		responseBranch := &GitBranch{
			Name: *branch.Name,
		}
		if branch.Commit != nil {
			responseBranch.Sha = *branch.Commit.CommitId
		}
		response = append(response, responseBranch)
	}

	return response, nil
}

func (g *AzureDevOpsGitProvider) GetRepoPRs(repositoryId string, namespaceId string) ([]*GitPullRequest, error) {
	client, err := g.getGitClient()
	if err != nil {
		return nil, err
	}

	repoUUID, err := uuid.Parse(repositoryId)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	prs, err := client.GetPullRequests(ctx, git.GetPullRequestsArgs{
		RepositoryId: &repositoryId,
		SearchCriteria: &git.GitPullRequestSearchCriteria{
			RepositoryId: &repoUUID,
		},
	})
	if err != nil {
		return nil, g.FormatError(err)
	}

	response := []*GitPullRequest{}

	for _, pr := range *prs {
		branch := *pr.SourceRefName
		branch = strings.TrimPrefix(branch, "refs/heads/")

		pullrequest := &GitPullRequest{
			Name:           *pr.Title,
			Sha:            *pr.LastMergeSourceCommit.CommitId,
			SourceRepoId:   repositoryId,
			SourceRepoName: *pr.Repository.Name,
			Branch:         branch,
		}

		repo, err := client.GetRepository(ctx, git.GetRepositoryArgs{
			RepositoryId: &repositoryId,
		})
		if err != nil {
			return nil, g.FormatError(err)
		}

		pullrequest.SourceRepoUrl = *repo.WebUrl

		owner := g.getOwnerName()
		if owner != "" {
			pullrequest.SourceRepoOwner = owner
		}

		response = append(response, pullrequest)
	}

	return response, nil
}

func (g *AzureDevOpsGitProvider) GetLastCommitSha(staticContext *StaticGitContext) (string, error) {
	client, err := g.getGitClient()
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	sha := ""
	gitVersionType := &git.GitVersionTypeValues.Branch

	if staticContext.Branch != nil {
		sha = *staticContext.Branch
	}

	if staticContext.Sha != nil {
		sha = *staticContext.Sha
		gitVersionType = &git.GitVersionTypeValues.Commit
	}

	commits, err := client.GetCommits(ctx, git.GetCommitsArgs{
		RepositoryId: &staticContext.Id,
		SearchCriteria: &git.GitQueryCommitsCriteria{
			ItemVersion: &git.GitVersionDescriptor{
				Version:     &sha,
				VersionType: gitVersionType,
			},
		},
		Top: &[]int{1}[0],
	})
	if err != nil {
		return "", g.FormatError(err)
	}

	if len(*commits) == 0 {
		return "", nil
	}
	return *(*commits)[0].CommitId, nil
}

func (g *AzureDevOpsGitProvider) GetBranchByCommit(staticContext *StaticGitContext) (string, error) {
	client, err := g.getGitClient()
	if err != nil {
		return "", err
	}

	branches, err := client.GetBranches(context.Background(), git.GetBranchesArgs{
		RepositoryId: &staticContext.Id,
	})
	if err != nil {
		return "", g.FormatError(err)
	}

	var branchName string
	for _, branch := range *branches {
		if *branch.Commit.CommitId == *staticContext.Sha {
			branchName = *branch.Name
			break
		}

		searchCriteria := &git.GitQueryCommitsCriteria{
			ItemVersion: &git.GitVersionDescriptor{
				Version:     &branchName,
				VersionType: &git.GitVersionTypeValues.Branch,
			},
			FromCommitId: staticContext.Sha,
			ToCommitId:   staticContext.Sha,
		}

		commits, err := client.GetCommitsBatch(context.Background(), git.GetCommitsBatchArgs{
			SearchCriteria: searchCriteria,
			RepositoryId:   &staticContext.Id,
		})
		if err != nil {
			continue
		}

		if len(*commits) == 0 {
			continue
		}

		for _, commit := range *commits {
			if *commit.CommitId == *staticContext.Sha {
				branchName = *branch.Name
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

func (g *AzureDevOpsGitProvider) GetUrlFromContext(repoContext *GetRepositoryContext) string {
	url := strings.TrimSuffix(repoContext.Url, ".git")
	if repoContext.Name != nil {
		url = strings.TrimSuffix(url, *repoContext.Name)
		url += "_git/" + *repoContext.Name
	}
	query := ""

	if repoContext.Branch != nil && *repoContext.Branch != "" {
		if repoContext.Sha != nil && *repoContext.Sha == *repoContext.Branch {
			query += "version=GC" + *repoContext.Branch
		} else {
			query += "version=GB" + *repoContext.Branch
		}

		if repoContext.Path != nil && *repoContext.Path != "" {
			if query != "" {
				query += "&"
			}

			query += "path=" + *repoContext.Path
		}
	} else if repoContext.Path != nil {
		query = "version=GBmain&path=" + *repoContext.Path
	} else {
		url = strings.Replace(url, "/_git", "", 1)
	}

	if query != "" {
		url += "?" + query
	}

	return url
}

func (g *AzureDevOpsGitProvider) GetPrContext(staticContext *StaticGitContext) (*StaticGitContext, error) {
	var pullRequestId int
	if staticContext.PrNumber == nil {
		return staticContext, nil
	} else {
		pullRequestId = int(*staticContext.PrNumber)
	}

	client, err := g.getGitClient()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	pr, err := client.GetPullRequest(ctx, git.GetPullRequestArgs{
		RepositoryId:  &staticContext.Id,
		PullRequestId: &pullRequestId,
	})
	if err != nil {
		return nil, g.FormatError(err)
	}

	repo := *staticContext
	sourceRefName := *pr.SourceRefName
	sourceRefName = strings.TrimPrefix(sourceRefName, "refs/heads/")

	repo.Branch = &sourceRefName
	repo.Id = staticContext.Id
	repo.Name = staticContext.Name

	owner := g.getOwnerName()
	if owner != "" {
		repo.Owner = owner
	}

	return &repo, nil
}

func (g *AzureDevOpsGitProvider) ParseStaticGitContext(repoUrl string) (*StaticGitContext, error) {
	repoUrl = strings.TrimSpace(repoUrl)
	if strings.HasPrefix(repoUrl, "git@") {
		return g.parseAzureDevopsSshGitUrl(repoUrl)
	}

	if strings.HasPrefix(repoUrl, "http") {
		return g.parseAzureDevopsHttpGitUrl(repoUrl)
	}

	return nil, errors.New("can not parse git URL: " + repoUrl)
}

func (g *AzureDevOpsGitProvider) GetDefaultBranch(staticContext *StaticGitContext) (*string, error) {
	client, err := g.getGitClient()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	repo, err := client.GetRepository(ctx, git.GetRepositoryArgs{
		RepositoryId: &staticContext.Id,
	})
	if err != nil {
		return nil, err
	}

	defaultBranch := *repo.DefaultBranch
	defaultBranch = strings.TrimPrefix(defaultBranch, "refs/heads/")
	return &defaultBranch, nil
}

func (g *AzureDevOpsGitProvider) parseAzureDevopsSshGitUrl(gitURL string) (*StaticGitContext, error) {
	re := regexp.MustCompile(`git@ssh.([\w\.]+):(.+?)/(.+?)/(.+?)/(.+?)?$`)
	matches := re.FindStringSubmatch(gitURL)
	if len(matches) != 6 {
		return nil, errors.New("cannot parse git URL: " + gitURL)
	}

	repo := &StaticGitContext{}

	repo.Source = matches[1]
	repo.Owner = matches[3]
	repo.Name = matches[5]
	project := matches[4]
	repo.Url = g.getAzureDevopsCloneUrl(repo.Source, repo.Owner, repo.Name, project)
	repoId, err := g.getAzureDevopsRepoId(repo.Name, project)
	if err != nil {
		return nil, err
	}
	repo.Id = repoId

	return repo, nil
}

func (g *AzureDevOpsGitProvider) parseAzureDevopsHttpGitUrl(gitURL string) (*StaticGitContext, error) {
	u, err := url.Parse(gitURL)
	if err != nil {
		return nil, err
	}

	repo := &StaticGitContext{}
	urlPattern := `^(https?://)?(?P<source>[^/]+)/(?P<org>[^/]+)(?:/(?P<project>[^/_]+))?/_git/(?P<repo>[^/?]+)(?:\?.*)?(?:/.*)?$`
	urlPatternRegex := regexp.MustCompile(urlPattern)
	matches := urlPatternRegex.FindStringSubmatch(gitURL)
	if len(matches) < 6 {
		return nil, errors.New("cannot parse git URL: " + gitURL)
	}

	repo.Source = u.Host
	repo.Owner = matches[3]
	repo.Name = matches[5]
	projectName := matches[4]

	urlPath := strings.TrimPrefix(u.Path, "/")

	if projectName == "" {
		projectName = repo.Name
		parts := strings.SplitN(urlPath, fmt.Sprintf("%s/", repo.Owner), 2)
		if len(parts) != 2 {
			return nil, errors.New("cannot parse git URL: " + gitURL)
		}

		urlPath = strings.Join([]string{parts[0], repo.Owner, projectName, parts[1]}, "/")
		urlPath = strings.TrimPrefix(urlPath, "/")
	}

	parts := strings.Split(urlPath, "/")

	repo.Url = g.getAzureDevopsCloneUrl(repo.Source, repo.Owner, repo.Name, projectName)
	repo.Name, _ = url.QueryUnescape(repo.Name)
	projectName, _ = url.QueryUnescape(projectName)

	repoId, err := g.getAzureDevopsRepoId(repo.Name, projectName)
	if err != nil {
		return nil, err
	}
	repo.Id = repoId

	queryParams, err := url.QueryUnescape(u.RawQuery)
	if err != nil {
		return nil, err
	}

	if len(parts) <= 4 && queryParams == "" {
		return repo, nil
	}

	switch {
	case len(parts) >= 6 && parts[4] == "pullrequest":
		prNumber, _ := strconv.Atoi(parts[5])
		prUint := uint32(prNumber)
		repo.PrNumber = &prUint
		repo.Path = nil
	case len(parts) >= 6 && parts[4] == "commit":
		repo.Sha = &parts[5]
		repo.Branch = &parts[5]
		repo.Path = nil
	}

	switch {
	case strings.Contains(queryParams, "itemVersion=GB"):
		fallthrough
	case strings.Contains(queryParams, "version=GB"):
		fallthrough
	case strings.Contains(queryParams, "refName=refs/heads/"):
		re := regexp.MustCompile(`(itemVersion|version|refName)=(GB|GT|refs/heads/)(.+?)(&|$)`)
		matches := re.FindStringSubmatch(queryParams)
		if len(matches) != 5 {
			return nil, errors.New("cannot parse git URL: " + gitURL)
		}
		if repo.Branch == nil {
			repo.Branch = &matches[3]
		}
		repo.Path = nil
	case strings.Contains(queryParams, "path="):
		re := regexp.MustCompile(`path=/(.+?)(&|$)`)
		matches := re.FindStringSubmatch(queryParams)
		if len(matches) != 3 {
			return nil, errors.New("cannot parse git URL: " + gitURL)
		}

		repo.Path = &matches[1]
	}

	return repo, nil
}

func (g *AzureDevOpsGitProvider) getAzureDevopsCloneUrl(source string, owner string, repo string, project string) string {
	return fmt.Sprintf("https://%s/%s/%s/_git/%s", source, owner, project, repo)
}

func (g *AzureDevOpsGitProvider) getAzureDevopsRepoId(repo string, project string) (string, error) {
	client, err := g.getGitClient()
	if err != nil {
		return "", err
	}

	repository, err := client.GetRepository(context.Background(), git.GetRepositoryArgs{
		RepositoryId: &repo,
		Project:      &project,
	})
	if err != nil {
		return "", err
	}

	return repository.Id.String(), nil
}

func (g *AzureDevOpsGitProvider) getOwnerName() string {
	baseUrl := g.baseApiUrl
	re := regexp.MustCompile(`/([^/]+)/?$`)
	matches := re.FindStringSubmatch(baseUrl)
	if len(matches) == 2 {
		return matches[1]
	}

	return ""
}

func (g *AzureDevOpsGitProvider) getGitClient() (git.Client, error) {
	ctx := context.Background()
	connection := azuredevops.NewPatConnection(g.baseApiUrl, g.token)

	client, err := git.NewClient(ctx, connection)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (g *AzureDevOpsGitProvider) getApiClient() (core.Client, *azuredevops.Connection, error) {
	ctx := context.Background()
	connection := azuredevops.NewPatConnection(g.baseApiUrl, g.token)

	client, err := core.NewClient(ctx, connection)
	if err != nil {
		return nil, nil, err
	}
	return client, connection, nil
}

func (g *AzureDevOpsGitProvider) getLocationClient() location.Client {
	ctx := context.Background()
	connection := azuredevops.NewPatConnection(g.baseApiUrl, g.token)

	client := location.NewClient(ctx, connection)
	return client
}

func (g *AzureDevOpsGitProvider) RegisterPrebuildWebhook(repo *GitRepository, endpointUrl string) (string, error) {
	coreClient, conn, err := g.getApiClient()
	if err != nil {
		return "", err
	}

	serviceHooksClient := servicehooks.NewClient(context.Background(), conn)

	project, err := coreClient.GetProject(context.Background(), core.GetProjectArgs{
		ProjectId: &repo.Name,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get project: %w", err)
	}
	projectID := project.Id.String()

	subscription := servicehooks.Subscription{
		PublisherId:      util.Pointer("tfs"),
		EventType:        util.Pointer("git.push"),
		ResourceVersion:  util.Pointer("1.0"),
		ConsumerActionId: util.Pointer("httpRequest"),
		ConsumerId:       util.Pointer("webHooks"),
		ConsumerInputs: &map[string]string{
			"url":         endpointUrl,
			"httpHeaders": "X-AzureDevops-Event:git.push\nX-Owner:" + repo.Owner,
		},
		PublisherInputs: &map[string]string{
			"projectId":  projectID,
			"repository": repo.Id,
		},
		Status: &servicehooks.SubscriptionStatusValues.Enabled,
	}

	var hook *servicehooks.Subscription
	hook, err = serviceHooksClient.CreateSubscription(context.Background(), servicehooks.CreateSubscriptionArgs{
		Subscription: &subscription,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create subscription: %w", err)
	}

	return hook.Id.String(), nil
}

func (g *AzureDevOpsGitProvider) GetPrebuildWebhook(repo *GitRepository, endpointUrl string) (*string, error) {
	_, conn, err := g.getApiClient()
	if err != nil {
		return nil, err
	}

	serviceHooksClient := servicehooks.NewClient(context.Background(), conn)
	hooks, err := serviceHooksClient.ListSubscriptions(context.Background(), servicehooks.ListSubscriptionsArgs{
		PublisherId:      util.Pointer("tfs"),
		ConsumerActionId: util.Pointer("httpRequest"),
		ConsumerId:       util.Pointer("webHooks"),
		EventType:        util.Pointer("git.push"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	for _, hook := range *hooks {
		if (*hook.ConsumerInputs)["url"] == endpointUrl {
			return util.Pointer(hook.Id.String()), nil
		}
	}

	return nil, nil
}

func (g *AzureDevOpsGitProvider) UnregisterPrebuildWebhook(repo *GitRepository, id string) error {
	_, conn, err := g.getApiClient()
	if err != nil {
		return err
	}

	uuid, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	serviceHooksClient := servicehooks.NewClient(context.Background(), conn)

	if err := serviceHooksClient.DeleteSubscription(context.Background(), servicehooks.DeleteSubscriptionArgs{
		SubscriptionId: &uuid,
	}); err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	return nil
}

func (g *AzureDevOpsGitProvider) GetCommitsRange(repo *GitRepository, initialSha string, currentSha string) (int, error) {
	gitClient, err := g.getGitClient()
	if err != nil {
		return 0, err
	}

	commits, err := gitClient.GetCommitDiffs(context.Background(), git.GetCommitDiffsArgs{
		RepositoryId: &repo.Id,
		Project:      &repo.Name,
		BaseVersionDescriptor: &git.GitBaseVersionDescriptor{
			BaseVersion:     &initialSha,
			BaseVersionType: &git.GitVersionTypeValues.Commit,
		},
		TargetVersionDescriptor: &git.GitTargetVersionDescriptor{
			TargetVersion:     &currentSha,
			TargetVersionType: &git.GitVersionTypeValues.Commit,
		},
	})
	if err != nil {
		return 0, err
	}

	return *commits.AheadCount, nil
}

func (g *AzureDevOpsGitProvider) ParseEventData(request *http.Request) (*GitEventData, error) {
	if request.Header.Get("X-AzureDevops-Event") != "git.push" {
		return nil, fmt.Errorf("invalid event key: %s", request.Header.Get("X-AzureDevops-Event"))
	}

	hook, err := azureWebhook.New()
	if err != nil {
		return nil, err
	}
	event, err := hook.Parse(request, azureWebhook.GitPushEventType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse event: %w", err)
	}
	pushEvent, ok := event.(azureWebhook.GitPushEvent)
	if !ok {
		return nil, fmt.Errorf("failed to parse push event: %w", err)
	}

	owner := request.Header.Get("X-Owner")

	gitEventData := &GitEventData{
		Owner:  owner,
		Url:    util.CleanUpRepositoryUrl(pushEvent.Resource.Repository.RemoteURL),
		Branch: strings.TrimPrefix(pushEvent.Resource.Repository.DefaultBranch, "refs/heads/"),
		Sha:    pushEvent.Resource.Commits[0].CommitID,
	}

	for _, commit := range pushEvent.Resource.Commits {
		gitEventData.AffectedFiles = append(gitEventData.AffectedFiles, commit.CommitID)
	}

	return gitEventData, nil
}

func (g *AzureDevOpsGitProvider) FormatError(err error) error {
	data, marshalErr := json.Marshal(err)
	if marshalErr != nil {
		return fmt.Errorf("status code: %d err: failed to format the error message: Request failed with %s", http.StatusInternalServerError, marshalErr.Error())
	}

	jsonData := azuredevops.WrappedError{}
	unmarshalErr := json.Unmarshal(data, &jsonData)
	if unmarshalErr != nil {
		return fmt.Errorf("status code: %d err: failed to format the error message: Request failed with %s", http.StatusInternalServerError, unmarshalErr.Error())
	}

	statusCode := http.StatusInternalServerError
	message := "unknown error"

	if jsonData.StatusCode != nil {
		statusCode = *jsonData.StatusCode
	}

	if jsonData.Message != nil {
		message = *jsonData.Message
	}

	return fmt.Errorf("status code: %d err: Request failed with %s", statusCode, message)
}
