// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/xanzy/go-gitlab"
)

type GitLabGitProvider struct {
	*AbstractGitProvider

	token      string
	baseApiUrl *string
}

func NewGitLabGitProvider(token string, baseApiUrl *string) *GitLabGitProvider {
	gitProvider := &GitLabGitProvider{
		token:               token,
		baseApiUrl:          baseApiUrl,
		AbstractGitProvider: &AbstractGitProvider{},
	}
	gitProvider.AbstractGitProvider.GitProvider = gitProvider

	return gitProvider
}

func (g *GitLabGitProvider) CanHandle(repoUrl string) (bool, error) {
	staticContext, err := g.ParseStaticGitContext(repoUrl)
	if err != nil {
		return false, err
	}

	if g.baseApiUrl == nil {
		return staticContext.Source == "gitlab.com", nil
	}

	return strings.Contains(*g.baseApiUrl, staticContext.Source), nil
}

func (g *GitLabGitProvider) GetNamespaces() ([]*GitNamespace, error) {
	client := g.getApiClient()
	user, err := g.GetUser()
	if err != nil {
		return nil, err
	}

	groupList, _, err := client.Groups.ListGroups(&gitlab.ListGroupsOptions{})
	if err != nil {
		return nil, g.FormatError(err)
	}

	namespaces := []*GitNamespace{}

	for _, group := range groupList {
		namespaces = append(namespaces, &GitNamespace{
			Id:   strconv.Itoa(group.ID),
			Name: group.Name,
		})
	}

	namespaces = append([]*GitNamespace{{Id: personalNamespaceId, Name: user.Username}}, namespaces...)

	return namespaces, nil
}

func (g *GitLabGitProvider) GetRepositories(namespace string) ([]*GitRepository, error) {
	client := g.getApiClient()
	var response []*GitRepository
	var repoList []*gitlab.Project

	if namespace == personalNamespaceId {
		user, err := g.GetUser()
		if err != nil {
			return nil, err
		}

		repos, _, err := client.Projects.ListUserProjects(user.Id, &gitlab.ListProjectsOptions{
			ListOptions: gitlab.ListOptions{
				PerPage: 100,
				Page:    1,
			},
		})
		if err != nil {
			return nil, g.FormatError(err)
		}
		repoList = repos
	} else {
		repos, _, err := client.Groups.ListGroupProjects(namespace, &gitlab.ListGroupProjectsOptions{
			ListOptions: gitlab.ListOptions{
				PerPage: 100,
				Page:    1,
			},
		})
		if err != nil {
			return nil, g.FormatError(err)
		}
		repoList = repos
	}

	for _, repo := range repoList {
		u, err := url.Parse(repo.WebURL)
		if err != nil {
			return nil, err
		}

		response = append(response, &GitRepository{
			Id:     strconv.Itoa(repo.ID),
			Name:   repo.Path,
			Url:    repo.WebURL,
			Branch: repo.DefaultBranch,
			Owner:  repo.Namespace.Path,
			Source: u.Host,
		})
	}

	return response, nil
}

func (g *GitLabGitProvider) GetRepoBranches(repositoryId string, namespaceId string) ([]*GitBranch, error) {
	client := g.getApiClient()
	var response []*GitBranch

	branches, _, err := client.Branches.ListBranches(repositoryId, &gitlab.ListBranchesOptions{})
	if err != nil {
		return nil, g.FormatError(err)
	}

	for _, branch := range branches {
		responseBranch := &GitBranch{
			Name: branch.Name,
		}
		if branch.Commit != nil {
			responseBranch.Sha = branch.Commit.ID
		}
		response = append(response, responseBranch)
	}

	return response, nil
}

func (g *GitLabGitProvider) GetRepoPRs(repositoryId string, namespaceId string) ([]*GitPullRequest, error) {
	client := g.getApiClient()
	var response []*GitPullRequest

	mergeRequests, _, err := client.MergeRequests.ListProjectMergeRequests(repositoryId, &gitlab.ListProjectMergeRequestsOptions{})
	if err != nil {
		return nil, g.FormatError(err)
	}

	for _, mergeRequest := range mergeRequests {
		sourceRepo, _, err := client.Projects.GetProject(mergeRequest.SourceProjectID, nil)
		if err != nil {
			log.Warn(g.FormatError(err))
			continue
		}

		response = append(response, &GitPullRequest{
			Name:            mergeRequest.Title,
			Branch:          mergeRequest.SourceBranch,
			Sha:             mergeRequest.SHA,
			SourceRepoId:    fmt.Sprint(mergeRequest.SourceProjectID),
			SourceRepoUrl:   sourceRepo.WebURL,
			SourceRepoOwner: sourceRepo.Namespace.Path,
			SourceRepoName:  sourceRepo.Path,
		})
	}

	return response, nil
}

func (g *GitLabGitProvider) GetUser() (*GitUser, error) {
	client := g.getApiClient()

	user, _, err := client.Users.CurrentUser()
	if err != nil {
		return nil, g.FormatError(err)
	}

	userId := strconv.Itoa(user.ID)

	response := &GitUser{
		Id:       userId,
		Username: user.Username,
		Name:     user.Name,
		Email:    user.Email,
	}

	return response, nil
}

func (g *GitLabGitProvider) GetBranchByCommit(staticContext *StaticGitContext) (string, error) {
	client := g.getApiClient()

	branches, _, err := client.Branches.ListBranches(staticContext.Id, &gitlab.ListBranchesOptions{})
	if err != nil {
		return "", g.FormatError(err)
	}

	var branchName string
	for _, branch := range branches {
		if *staticContext.Sha == branch.Commit.ID {
			branchName = branch.Name
			break
		}

		commitId := branch.Commit.ID
		for commitId != "" {
			commit, _, err := client.Commits.GetCommit(staticContext.Id, commitId)
			if err != nil {
				return "", g.FormatError(err)
			}

			if *staticContext.Sha == commit.ID {
				branchName = branch.Name
				break
			}

			if len(commit.ParentIDs) > 0 {
				commitId = commit.ParentIDs[0]
				if *staticContext.Sha == commitId {
					branchName = branch.Name
					break
				}
			} else {
				commitId = ""
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

func (g *GitLabGitProvider) GetLastCommitSha(staticContext *StaticGitContext) (string, error) {
	client := g.getApiClient()

	var sha *string

	if staticContext.Branch != nil {
		sha = staticContext.Branch
	}

	if staticContext.Sha != nil {
		sha = staticContext.Sha
	}

	commits, _, err := client.Commits.ListCommits(staticContext.Id, &gitlab.ListCommitsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 1,
		},
		RefName: sha,
	})
	if err != nil {
		return "", g.FormatError(err)
	}
	if len(commits) == 0 {
		return "", nil
	}

	return commits[0].ID, nil
}

func (g *GitLabGitProvider) GetUrlFromContext(repoContext *GetRepositoryContext) string {
	url := strings.TrimSuffix(repoContext.Url, ".git")

	if repoContext.Branch != nil && *repoContext.Branch != "" {
		if repoContext.Sha != nil && *repoContext.Sha == *repoContext.Branch {
			url += "/-/commit/" + *repoContext.Branch
		} else {
			url += "/-/tree/" + *repoContext.Branch
		}

		if repoContext.Path != nil {
			url += "/" + *repoContext.Path
		}
	} else if repoContext.Path != nil {
		url += "/-/blob/main/" + *repoContext.Path
	}

	return url
}

func (g *GitLabGitProvider) getApiClient() *gitlab.Client {
	var client *gitlab.Client
	var err error

	if g.baseApiUrl == nil {
		client, err = gitlab.NewClient(g.token)
	} else {
		client, err = gitlab.NewClient(g.token, gitlab.WithBaseURL(*g.baseApiUrl))
	}
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func (g *GitLabGitProvider) ParseStaticGitContext(repoUrl string) (*StaticGitContext, error) {
	if strings.HasPrefix(repoUrl, "git@") {
		return g.parseSshGitUrl(repoUrl)
	}

	// Determine protocol based on baseApiUrl or repoUrl
	isHttps := true
	if g.baseApiUrl != nil && strings.HasPrefix(*g.baseApiUrl, "http://") {
		isHttps = false
	} else if strings.HasPrefix(repoUrl, "http://") {
		isHttps = false
	}

	if !strings.HasPrefix(repoUrl, "http") {
		return nil, fmt.Errorf("can not parse git URL: %s", repoUrl)
	}

	repoUrl = strings.TrimSuffix(repoUrl, ".git")

	u, err := url.Parse(repoUrl)
	if err != nil {
		return nil, err
	}

	staticContext := &StaticGitContext{
		Source: u.Host,
	}

	parts := strings.Split(u.Path, "/-/")

	ownerRepo := strings.TrimPrefix(parts[0], "/")

	if len(parts) == 2 {
		staticContext.Path = &parts[1]
	}

	if len(parts) > 2 {
		return nil, fmt.Errorf("can not parse git URL: %s", repoUrl)
	}

	ownerRepoParts := strings.Split(ownerRepo, "/")
	if len(ownerRepoParts) < 2 {
		return nil, fmt.Errorf("can not parse git URL: %s", repoUrl)
	}

	staticContext.Name = ownerRepoParts[len(ownerRepoParts)-1]
	staticContext.Owner = strings.Join(ownerRepoParts[:len(ownerRepoParts)-1], "/")
	staticContext.Url = getCloneUrl(staticContext.Source, staticContext.Owner, staticContext.Name, isHttps)
	staticContext.Id = fmt.Sprintf("%s/%s", staticContext.Owner, staticContext.Name)

	if staticContext.Path == nil {
		return staticContext, nil
	}

	switch {
	case strings.Contains(*staticContext.Path, "merge_requests"):
		parts := strings.Split(*staticContext.Path, "merge_requests/")
		mrParts := strings.Split(parts[1], "/")
		if len(mrParts) < 1 {
			return nil, fmt.Errorf("can not parse git URL: %s", repoUrl)
		}
		mrNumber, err := strconv.Atoi(mrParts[0])
		if err != nil {
			return nil, err
		}
		mrNumberUint := uint32(mrNumber)
		staticContext.PrNumber = &mrNumberUint
		staticContext.Path = nil
	case strings.Contains(*staticContext.Path, "tree/"):
		parts := strings.Split(*staticContext.Path, "tree/")
		if len(parts) < 2 {
			return nil, fmt.Errorf("can not parse git URL: %s", repoUrl)
		}

		branchParts := strings.Split(parts[1], "/")
		if len(branchParts) < 1 {
			return nil, fmt.Errorf("can not parse git URL: %s", repoUrl)
		}

		staticContext.Branch = &branchParts[0]
		staticContext.Path = nil
	case strings.Contains(*staticContext.Path, "blob/"):
		parts := strings.Split(*staticContext.Path, "blob/")
		if len(parts) < 2 {
			return nil, fmt.Errorf("can not parse git URL: %s", repoUrl)
		}

		branchParts := strings.Split(parts[1], "/")
		if len(branchParts) < 1 {
			return nil, fmt.Errorf("can not parse git URL: %s", repoUrl)
		}

		staticContext.Branch = &branchParts[0]
		branchPath := strings.Join(branchParts[1:], "/")
		staticContext.Path = &branchPath
	case strings.Contains(*staticContext.Path, "commit/"):
		parts := strings.Split(*staticContext.Path, "commit/")
		if len(parts) < 2 {
			return nil, fmt.Errorf("can not parse git URL: %s", repoUrl)
		}

		commitParts := strings.Split(parts[1], "/")
		if len(commitParts) < 1 {
			return nil, fmt.Errorf("can not parse git URL: %s", repoUrl)
		}

		staticContext.Sha = &commitParts[0]
		staticContext.Branch = &commitParts[0]
		staticContext.Path = nil
	case strings.Contains(*staticContext.Path, "commits/"):
		parts := strings.Split(*staticContext.Path, "commits/")
		if len(parts) < 2 {
			return nil, fmt.Errorf("can not parse git URL: %s", repoUrl)
		}

		branchParts := strings.Split(parts[1], "/")
		if len(branchParts) < 1 {
			return nil, fmt.Errorf("can not parse git URL: %s", repoUrl)
		}

		sha1Pattern := regexp.MustCompile(`^[a-fA-F0-9]{40}$`)
		if sha1Pattern.MatchString(branchParts[0]) {
			staticContext.Sha = &branchParts[0]
			staticContext.Branch = &branchParts[0]
		} else {
			staticContext.Branch = &branchParts[0]
		}
		staticContext.Path = nil
	}

	return staticContext, nil
}

func (g *GitLabGitProvider) GetPrContext(staticContext *StaticGitContext) (*StaticGitContext, error) {
	if staticContext.PrNumber == nil {
		return staticContext, nil
	}

	client := g.getApiClient()

	pull, _, err := client.MergeRequests.GetMergeRequest(staticContext.Id, int(*staticContext.PrNumber), nil)
	if err != nil {
		return nil, g.FormatError(err)
	}

	project, _, err := client.Projects.GetProject(staticContext.Id, nil)
	if err != nil {
		return nil, g.FormatError(err)
	}

	repo := *staticContext
	repo.Branch = &pull.SourceBranch
	repo.Url = project.HTTPURLToRepo
	repo.Owner = pull.Author.Username

	return &repo, nil
}

func (g *GitLabGitProvider) GetDefaultBranch(staticContext *StaticGitContext) (*string, error) {
	client := g.getApiClient()

	project, _, err := client.Projects.GetProject(staticContext.Id, nil)
	if err != nil {
		return nil, g.FormatError(err)
	}

	return &project.DefaultBranch, nil
}

func (g *GitLabGitProvider) RegisterPrebuildWebhook(repo *GitRepository, endpointUrl string) (string, error) {
	client := g.getApiClient()

	pushEvents := true
	projectID := fmt.Sprintf("%s/%s", repo.Owner, repo.Name)

	hook, _, err := client.Projects.AddProjectHook(projectID, &gitlab.AddProjectHookOptions{
		URL:        &endpointUrl,
		PushEvents: &pushEvents,
	})
	if err != nil {
		return "", g.FormatError(err)
	}

	return strconv.Itoa(hook.ID), nil
}

func (g *GitLabGitProvider) GetPrebuildWebhook(repo *GitRepository, endpointUrl string) (*string, error) {
	client := g.getApiClient()

	projectID := fmt.Sprintf("%s/%s", repo.Owner, repo.Name)
	hooks, _, err := client.Projects.ListProjectHooks(projectID, &gitlab.ListProjectHooksOptions{
		PerPage: 100,
	})
	if err != nil {
		return nil, g.FormatError(err)
	}

	for _, hook := range hooks {
		if hook.URL == endpointUrl {
			return util.Pointer(strconv.Itoa(hook.ID)), nil
		}
	}

	return nil, nil
}

func (g *GitLabGitProvider) UnregisterPrebuildWebhook(repo *GitRepository, hookId string) error {
	client := g.getApiClient()

	hookIdInt, err := strconv.Atoi(hookId)
	if err != nil {
		return err
	}

	projectID := fmt.Sprintf("%s/%s", repo.Owner, repo.Name)
	_, err = client.Projects.DeleteProjectHook(projectID, hookIdInt)
	if err != nil {
		return g.FormatError(err)
	}

	return nil
}

func (g *GitLabGitProvider) GetCommitsRange(repo *GitRepository, initialSha string, currentSha string) (int, error) {
	client := g.getApiClient()

	projectID := fmt.Sprintf("%s/%s", repo.Owner, repo.Name)
	commits, _, err := client.Repositories.Compare(projectID, &gitlab.CompareOptions{
		From: &initialSha,
		To:   &currentSha,
	})
	if err != nil {
		return 0, g.FormatError(err)
	}

	return len(commits.Commits), nil
}

func (g *GitLabGitProvider) ParseEventData(request *http.Request) (*GitEventData, error) {
	payload, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}

	var webhookData gitlab.PushEvent
	err = json.Unmarshal(payload, &webhookData)
	if err != nil {
		return nil, err
	}
	if webhookData.EventName != "push" {
		return nil, nil
	}

	var owner string
	if webhookData.Project.Namespace != "" {
		owner = webhookData.Project.Namespace
	}

	gitEventData := &GitEventData{
		Url:    util.CleanUpRepositoryUrl(webhookData.Project.WebURL) + ".git",
		Branch: strings.TrimPrefix(webhookData.Ref, "refs/heads/"),
		Sha:    webhookData.After,
		Owner:  owner,
	}

	for _, commit := range webhookData.Commits {
		gitEventData.AffectedFiles = append(gitEventData.AffectedFiles, commit.Added...)
		gitEventData.AffectedFiles = append(gitEventData.AffectedFiles, commit.Modified...)
		gitEventData.AffectedFiles = append(gitEventData.AffectedFiles, commit.Removed...)
	}

	return gitEventData, nil
}

func (g *GitLabGitProvider) FormatError(err error) error {
	re := regexp.MustCompile(`([A-Z]+)\s(https:\/\/\S+):\s(\d{3})\s(\{message:\s\d{3}\s.+\})`)
	match := re.FindStringSubmatch(err.Error())
	if len(match) == 5 {
		return fmt.Errorf("status code: %s err: Request to %s failed with %s", match[3], match[2], match[4])
	}

	return fmt.Errorf("status code: %d err: failed to format error message Request failed with %s", http.StatusInternalServerError, err.Error())
}
