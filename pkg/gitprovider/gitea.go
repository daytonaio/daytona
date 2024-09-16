// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/daytonaio/daytona/internal/util"
	giteaWebhook "github.com/go-playground/webhooks/v6/gitea"
)

type GiteaGitProvider struct {
	*AbstractGitProvider

	token      string
	baseApiUrl string
}

func NewGiteaGitProvider(token string, baseApiUrl string) *GiteaGitProvider {
	provider := &GiteaGitProvider{
		token:               token,
		baseApiUrl:          baseApiUrl,
		AbstractGitProvider: &AbstractGitProvider{},
	}
	provider.AbstractGitProvider.GitProvider = provider

	return provider
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

	orgList, res, err := client.ListMyOrgs(gitea.ListOrgsOptions{
		ListOptions: gitea.ListOptions{
			Page:     1,
			PageSize: 100,
		},
	})
	if err != nil {
		return nil, g.FormatError(res, err)
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

		repos, res, err := client.ListUserRepos(user.Username, gitea.ListReposOptions{
			ListOptions: gitea.ListOptions{
				Page:     1,
				PageSize: 100,
			},
		})
		if err != nil {
			return nil, g.FormatError(res, err)
		}
		repoList = repos
	} else {
		repos, res, err := client.ListOrgRepos(namespace, gitea.ListOrgReposOptions{
			ListOptions: gitea.ListOptions{
				Page:     1,
				PageSize: 100,
			},
		})
		if err != nil {
			return nil, g.FormatError(res, err)
		}
		repoList = repos
	}

	response := []*GitRepository{}

	for _, repo := range repoList {
		u, err := url.Parse(repo.HTMLURL)
		if err != nil {
			return nil, err
		}
		response = append(response, &GitRepository{
			Id:     repo.Name,
			Name:   repo.Name,
			Url:    repo.HTMLURL,
			Branch: repo.DefaultBranch,
			Owner:  repo.Owner.UserName,
			Source: u.Host,
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

	repoBranches, res, err := client.ListRepoBranches(namespaceId, repositoryId, gitea.ListRepoBranchesOptions{
		ListOptions: gitea.ListOptions{
			Page:     1,
			PageSize: 100,
		},
	})
	if err != nil {
		return nil, g.FormatError(res, err)
	}

	response := []*GitBranch{}

	for _, branch := range repoBranches {
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

	prList, res, err := client.ListRepoPullRequests(namespaceId, repositoryId, gitea.ListPullRequestsOptions{
		ListOptions: gitea.ListOptions{
			Page:     1,
			PageSize: 100,
		},
		State: gitea.StateOpen,
		Sort:  "recentupdate",
	})
	if err != nil {
		return nil, g.FormatError(res, err)
	}

	response := []*GitPullRequest{}

	for _, pr := range prList {
		response = append(response, &GitPullRequest{
			Name:            pr.Title,
			Branch:          pr.Head.Ref,
			Sha:             pr.Head.Sha,
			SourceRepoId:    pr.Head.Repository.Name,
			SourceRepoName:  pr.Head.Repository.Name,
			SourceRepoUrl:   pr.Head.Repository.HTMLURL,
			SourceRepoOwner: pr.Head.Repository.Owner.UserName,
		})
	}

	return response, nil
}

func (g *GiteaGitProvider) GetUser() (*GitUser, error) {
	client, err := g.getApiClient()
	if err != nil {
		return nil, err
	}

	user, res, err := client.GetMyUserInfo()
	if user == nil || err != nil {
		return nil, g.FormatError(res, err)
	}

	return &GitUser{
		Id:       strconv.FormatInt(user.ID, 10),
		Username: user.UserName,
		Name:     user.FullName,
		Email:    user.Email,
	}, nil
}

func (g *GiteaGitProvider) GetBranchByCommit(staticContext *StaticGitContext) (string, error) {
	client, err := g.getApiClient()
	if err != nil {
		return "", err
	}

	repoBranches, res, err := client.ListRepoBranches(staticContext.Owner, staticContext.Name, gitea.ListRepoBranchesOptions{
		ListOptions: gitea.ListOptions{
			Page:     1,
			PageSize: 100,
		},
	})
	if err != nil {
		return "", g.FormatError(res, err)
	}
	var branchName string
	for _, branch := range repoBranches {
		if *staticContext.Sha == branch.Commit.ID {
			branchName = branch.Name
			break
		}

		commitId := branch.Commit.ID
		for commitId != "" {
			commit, _, err := client.GetSingleCommit(staticContext.Owner, staticContext.Id, commitId)
			if err != nil {
				continue
			}

			if *staticContext.Sha == commit.SHA {
				branchName = branch.Name
				break
			}
			if len(commit.Parents) > 0 {
				commitId = commit.Parents[0].SHA
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

func (g *GiteaGitProvider) GetLastCommitSha(staticContext *StaticGitContext) (string, error) {
	client, err := g.getApiClient()
	if err != nil {
		return "", err
	}

	branch := ""
	if staticContext.Branch != nil {
		branch = *staticContext.Branch
	}

	commits, res, err := client.ListRepoCommits(staticContext.Owner, staticContext.Id, gitea.ListCommitOptions{
		SHA: branch,
	})
	if err != nil {
		return "", g.FormatError(res, err)
	}

	if len(commits) == 0 {
		return "", nil
	}

	return commits[0].SHA, nil
}

func (g *GiteaGitProvider) getApiClient() (*gitea.Client, error) {
	ctx := context.Background()

	options := []gitea.ClientOption{
		gitea.SetContext(ctx),
	}

	if g.token != "" {
		options = append(options, gitea.SetToken(g.token))
	}

	return gitea.NewClient(g.baseApiUrl, options...)
}

func (g *GiteaGitProvider) GetUrlFromContext(repoContext *GetRepositoryContext) string {
	url := strings.TrimSuffix(repoContext.Url, ".git")

	if repoContext.Branch != nil && *repoContext.Branch != "" {
		if repoContext.Sha != nil && *repoContext.Sha == *repoContext.Branch {
			url += "/src/commit/" + *repoContext.Branch
		} else {
			url += "/src/branch/" + *repoContext.Branch
		}

		if repoContext.Path != nil && *repoContext.Path != "" {
			url += "/" + *repoContext.Path
		}
	} else if repoContext.Path != nil {
		url += "/src/branch/main/" + *repoContext.Path
	}

	return url
}

func (g *GiteaGitProvider) GetPrContext(staticContext *StaticGitContext) (*StaticGitContext, error) {
	if staticContext.PrNumber == nil {
		return staticContext, nil
	}

	client, err := g.getApiClient()
	if err != nil {
		return nil, err
	}

	pr, res, err := client.GetPullRequest(staticContext.Owner, staticContext.Id, int64(*staticContext.PrNumber))
	if err != nil {
		return nil, g.FormatError(res, err)
	}

	repo := *staticContext
	repo.Branch = &pr.Head.Ref
	repo.Url = pr.Head.Repository.CloneURL
	repo.Name = pr.Head.Repository.Name
	repo.Id = pr.Head.Repository.Name
	repo.Owner = pr.Head.Repository.Owner.UserName

	return &repo, nil
}

func (g *GiteaGitProvider) ParseStaticGitContext(repoUrl string) (*StaticGitContext, error) {
	staticContext, err := g.AbstractGitProvider.ParseStaticGitContext(repoUrl)
	if err != nil {
		return nil, err
	}

	if staticContext.Path == nil {
		return staticContext, nil
	}

	parts := strings.Split(*staticContext.Path, "/")

	switch {
	case len(parts) >= 2 && parts[0] == "pulls":
		prNumber, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}
		prUint := uint32(prNumber)
		staticContext.PrNumber = &prUint
		staticContext.Path = nil
	case len(parts) >= 3 && parts[0] == "src" && parts[1] == "branch":
		staticContext.Branch = &parts[2]
		if len(parts) > 3 {
			branchPath := strings.Join(parts[3:], "/")
			staticContext.Path = &branchPath
		} else {
			staticContext.Path = nil
		}
	case len(parts) >= 3 && parts[0] == "src" && parts[1] == "commit":
		staticContext.Sha = &parts[2]
		staticContext.Branch = staticContext.Sha
		if len(parts) > 3 {
			branchPath := strings.Join(parts[3:], "/")
			staticContext.Path = &branchPath
		} else {
			staticContext.Path = nil
		}
	case len(parts) >= 2 && parts[0] == "commit":
		staticContext.Sha = &parts[1]
		staticContext.Branch = staticContext.Sha
		staticContext.Path = nil
	case len(parts) == 3 && parts[0] == "commits" && parts[1] == "branch":
		staticContext.Branch = &parts[2]
		staticContext.Path = nil
	}

	return staticContext, nil
}

func (g *GiteaGitProvider) GetDefaultBranch(staticContext *StaticGitContext) (*string, error) {
	client, err := g.getApiClient()
	if err != nil {
		return nil, err
	}

	repo, res, err := client.GetRepo(staticContext.Owner, staticContext.Id)
	if err != nil {
		return nil, g.FormatError(res, err)
	}

	return &repo.DefaultBranch, nil
}

func (g *GiteaGitProvider) RegisterPrebuildWebhook(repo *GitRepository, endpointUrl string) (string, error) {
	client, err := g.getApiClient()
	if err != nil {
		return "", fmt.Errorf("failed to get api client: %w", err)
	}

	hookOpts := gitea.CreateHookOption{
		Type: "gitea",
		Config: map[string]string{
			"url":          endpointUrl,
			"content_type": "json",
		},
		Events: []string{"push"},
		Active: true,
	}

	hook, res, err := client.CreateRepoHook(repo.Owner, repo.Name, hookOpts)
	if err != nil {
		return "", g.FormatError(res, err)
	}

	return strconv.Itoa(int(hook.ID)), nil
}

func (g *GiteaGitProvider) GetPrebuildWebhook(repo *GitRepository, endpointUrl string) (*string, error) {
	client, err := g.getApiClient()
	if err != nil {
		return nil, err
	}

	hooks, res, err := client.ListRepoHooks(repo.Owner, repo.Name, gitea.ListHooksOptions{
		ListOptions: gitea.ListOptions{
			PageSize: 100,
			Page:     1,
		},
	})
	if err != nil {
		return nil, g.FormatError(res, err)
	}

	for _, hook := range hooks {
		if hook.Config["url"] == endpointUrl {
			return util.Pointer(strconv.Itoa(int(hook.ID))), nil
		}
	}

	return nil, nil
}

func (g *GiteaGitProvider) UnregisterPrebuildWebhook(repo *GitRepository, id string) error {
	client, err := g.getApiClient()
	if err != nil {
		return err
	}

	hookId, err := strconv.Atoi(id)
	if err != nil {
		return err
	}

	res, err := client.DeleteRepoHook(repo.Owner, repo.Name, int64(hookId))

	return g.FormatError(res, err)
}

func (g *GiteaGitProvider) GetCommitsRange(repo *GitRepository, initialSha string, currentSha string) (int, error) {
	client, err := g.getApiClient()
	if err != nil {
		return 0, err
	}

	initialCommits, res, err := client.ListRepoCommits(repo.Owner, repo.Name, gitea.ListCommitOptions{
		SHA: initialSha,
	})
	if err != nil {
		return 0, g.FormatError(res, err)
	}

	currentCommits, res, err := client.ListRepoCommits(repo.Owner, repo.Name, gitea.ListCommitOptions{
		SHA: currentSha,
	})
	if err != nil {
		return 0, g.FormatError(res, err)
	}

	return (len(currentCommits) - len(initialCommits)), nil
}

func (g *GiteaGitProvider) ParseEventData(request *http.Request) (*GitEventData, error) {
	if request.Header.Get("X-Gitea-Event") != "push" {
		return nil, fmt.Errorf("invalid event key")
	}

	hook, err := giteaWebhook.New()
	if err != nil {
		return nil, err
	}

	event, err := hook.Parse(request, giteaWebhook.PushEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse event: %w", err)
	}

	pushEvent, ok := event.(giteaWebhook.PushPayload)
	if !ok {
		return nil, fmt.Errorf("failed to parse push event: %w", err)
	}

	owner := pushEvent.Repo.Owner.FullName

	gitEventData := &GitEventData{
		Owner:  owner,
		Url:    util.CleanUpRepositoryUrl(pushEvent.Repo.HTMLURL) + ".git",
		Branch: strings.TrimPrefix(pushEvent.Ref, "refs/heads/"),
		Sha:    pushEvent.After,
	}

	for _, commit := range pushEvent.Commits {
		gitEventData.AffectedFiles = append(gitEventData.AffectedFiles, commit.Added...)
		gitEventData.AffectedFiles = append(gitEventData.AffectedFiles, commit.Modified...)
		gitEventData.AffectedFiles = append(gitEventData.AffectedFiles, commit.Removed...)
	}

	return gitEventData, nil
}

func (g *GiteaGitProvider) FormatError(response *gitea.Response, err error) error {
	fmt.Println(response.Body)
	return fmt.Errorf("status code: %d err: %s", response.StatusCode, err.Error())
}
