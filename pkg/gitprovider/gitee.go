// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"
	"fmt"
	"io"
	"strings"

	"net/http"
	"net/url"
	"strconv"

	"encoding/json"

	"gitee.com/openeuler/go-gitee/gitee"
	"github.com/antihax/optional"
	"golang.org/x/oauth2"
)

type Project struct {
	Id            int32            `json:"id,omitempty"`
	FullName      string           `json:"full_name,omitempty"`
	HumanName     string           `json:"human_name,omitempty"`
	Url           string           `json:"url,omitempty"`
	Namespace     *gitee.Namespace `json:"namespace,omitempty"`
	Path          string           `json:"path,omitempty"`
	Name          string           `json:"name,omitempty"`
	Owner         *gitee.UserBasic `json:"owner,omitempty"`
	Description   string           `json:"description,omitempty"`
	Private       bool             `json:"private,omitempty"`
	Public        bool             `json:"public,omitempty"`
	Internal      bool             `json:"internal,omitempty"`
	HtmlUrl       string           `json:"html_url,omitempty"`
	Homepage      string           `json:"homepage,omitempty"`
	DefaultBranch string           `json:"default_branch,omitempty"`
	Parent        *Project         `json:"parent,omitempty"`
}

type GiteeGitProvider struct {
	*AbstractGitProvider

	token string
}

func NewGiteeGitProvider(token string) *GiteeGitProvider {
	gitProvider := &GiteeGitProvider{
		token:               token,
		AbstractGitProvider: &AbstractGitProvider{},
	}
	gitProvider.AbstractGitProvider.GitProvider = gitProvider

	return gitProvider
}

func (g *GiteeGitProvider) CanHandle(repoUrl string) (bool, error) {
	staticContext, err := g.ParseStaticGitContext(repoUrl)
	if err != nil {
		return false, err
	}

	return strings.Contains("https://gitee.com", staticContext.Source), nil
}

func (g *GiteeGitProvider) getApiClient() *gitee.APIClient {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: g.token})
	conf := gitee.NewConfiguration()
	conf.HTTPClient = oauth2.NewClient(context.Background(), tokenSource)
	return gitee.NewAPIClient(conf)
}

func (g *GiteeGitProvider) GetUser() (*GitUser, error) {
	client := g.getApiClient()
	ctx := context.Background()
	user, _, err := client.UsersApi.GetV5User(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch User : %w", err)
	}

	gitUser := &GitUser{
		Id:       strconv.FormatInt(int64(user.Id), 10),
		Username: user.Login,
		Name:     user.Name,
	}

	// gitee email will be empty if not set to public
	if user.Email != "" {
		gitUser.Email = user.Email
	}

	return gitUser, nil
}

func (g *GiteeGitProvider) GetNamespaces(options ListOptions) ([]*GitNamespace, error) {
	client := g.getApiClient()
	ctx := context.Background()
	userNamespaces, _, err := client.UsersApi.GetV5UserNamespaces(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Namespaces : %w", err)
	}

	var namespaces []*GitNamespace
	for _, namespace := range userNamespaces {
		namespaces = append(namespaces, &GitNamespace{
			Id:   namespace.Name,
			Name: namespace.Name,
		})
	}

	return namespaces, nil
}

func (g *GiteeGitProvider) GetRepositories(namespace string, options ListOptions) ([]*GitRepository, error) {
	reposApiURL := fmt.Sprintf("https://gitee.com/api/v5/user/repos?access_token=%s", g.token)
	body, err := g.performRequest("GET", reposApiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Repositories : %w", err)
	}

	var repos []Project
	err = json.Unmarshal(body, &repos)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal GetRepositories response : %w", err)
	}

	response := []*GitRepository{}
	for _, repo := range repos {
		u, err := url.Parse(repo.HtmlUrl)
		if err != nil {
			return nil, err
		}
		response = append(response, &GitRepository{
			Id:     repo.Name,
			Name:   repo.Name,
			Url:    repo.HtmlUrl,
			Branch: repo.DefaultBranch,
			Owner:  repo.Owner.Login,
			Source: u.Host,
		})
	}
	return response, nil
}

func (g *GiteeGitProvider) GetRepoBranches(repositoryId string, namespaceId string, options ListOptions) ([]*GitBranch, error) {
	client := g.getApiClient()
	ctx := context.Background()
	branches, _, err := client.RepositoriesApi.GetV5ReposOwnerRepoBranches(ctx, namespaceId, repositoryId, nil)
	if err != nil {
		return nil, err
	}

	var response []*GitBranch
	for _, branch := range branches {
		response = append(response, &GitBranch{
			Name: branch.Name,
			Sha:  branch.Commit.Sha,
		})
	}
	return response, nil
}

func (g *GiteeGitProvider) GetRepoPRs(repositoryId string, namespaceId string, options ListOptions) ([]*GitPullRequest, error) {
	client := g.getApiClient()
	ctx := context.Background()
	prs, _, err := client.PullRequestsApi.GetV5ReposOwnerRepoPulls(ctx, namespaceId, repositoryId, nil)
	if err != nil {
		return nil, err
	}

	var response []*GitPullRequest
	for _, pr := range prs {
		response = append(response, &GitPullRequest{
			Name:            pr.Title,
			Branch:          pr.Head.Label,
			Sha:             pr.Head.Sha,
			SourceRepoId:    pr.Head.Repo.Name,
			SourceRepoName:  pr.Head.Repo.Name,
			SourceRepoUrl:   pr.Head.Repo.HtmlUrl,
			SourceRepoOwner: pr.Head.Repo.Owner.Name,
		})
	}
	return response, nil
}

func (g *GiteeGitProvider) GetBranchByCommit(staticContext *StaticGitContext) (string, error) {
	client := g.getApiClient()
	ctx := context.Background()
	branches, _, err := client.RepositoriesApi.GetV5ReposOwnerRepoBranches(ctx, staticContext.Owner, staticContext.Name, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get branch by commit: %w", err)
	}

	var branchName string
	for _, branch := range branches {
		commitId := branch.Commit.Sha
		if *staticContext.Sha == commitId {
			branchName = branch.Name
			break
		}

		commits, _, err := client.RepositoriesApi.GetV5ReposOwnerRepoCommits(ctx, staticContext.Owner, staticContext.Name, &gitee.GetV5ReposOwnerRepoCommitsOpts{
			Sha: optional.NewString(branchName),
		})
		if err != nil {
			continue
		}

		if len(commits) == 0 {
			continue
		}

		for _, commit := range commits {
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

func (g *GiteeGitProvider) GetLastCommitSha(staticContext *StaticGitContext) (string, error) {
	client := g.getApiClient()
	ctx := context.Background()

	var sha *string

	if staticContext.Branch != nil {
		sha = staticContext.Branch
	}

	if staticContext.Sha != nil {
		sha = staticContext.Sha
	}
	commits, _, err := client.RepositoriesApi.GetV5ReposOwnerRepoCommits(ctx, staticContext.Owner, staticContext.Name, &gitee.GetV5ReposOwnerRepoCommitsOpts{
		Sha: optional.NewString(*sha),
	})
	if err != nil {
		return "", err
	}
	if len(commits) == 0 {
		return "", nil
	}
	return commits[0].Sha, nil
}

func (g *GiteeGitProvider) GetDefaultBranch(staticContext *StaticGitContext) (*string, error) {
	client := g.getApiClient()
	ctx := context.Background()
	repo, _, err := client.RepositoriesApi.GetV5ReposOwnerRepo(ctx, staticContext.Owner, staticContext.Name, nil)
	if err != nil {
		return nil, err
	}

	return &repo.DefaultBranch, nil
}

func (g *GiteeGitProvider) GetUrlFromContext(repoContext *GetRepositoryContext) string {
	url := strings.TrimSuffix(repoContext.Url, ".git")

	if repoContext.Branch != nil && *repoContext.Branch != "" {
		if repoContext.Sha != nil && *repoContext.Sha == *repoContext.Branch {
			//https://gitee.com/<owner>/<repo>/commit/<sha>
			if repoContext.Path != nil {
				url += "/blob/" + *repoContext.Sha + "/" + *repoContext.Path
			} else {
				url += "/commit/" + *repoContext.Branch
			}
		} else {
			if repoContext.Path != nil {
				url += "/blob/" + *repoContext.Branch + "/" + *repoContext.Path
			} else {
				url += "/tree/" + *repoContext.Branch
			}
		}
	} else if repoContext.Path != nil {
		url += "/blob/master/" + *repoContext.Path
	}

	return url
}

func (g *GiteeGitProvider) ParseStaticGitContext(repoUrl string) (*StaticGitContext, error) {
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

	case len(parts) >= 2 && parts[0] == "commit":
		staticContext.Sha = &parts[1]
		staticContext.Branch = staticContext.Sha
		staticContext.Path = nil

	case len(parts) >= 2 && (parts[0] == "commits" || parts[0] == "tree"):
		staticContext.Branch = &parts[1]
		staticContext.Path = nil

	case len(parts) >= 2 && parts[0] == "blob":
		staticContext.Sha = &parts[1]
		staticContext.Branch = staticContext.Sha
		path := strings.Join(parts[2:], "/")
		staticContext.Path = &path

	}

	return staticContext, nil
}

func (g *GiteeGitProvider) GetPrContext(staticContext *StaticGitContext) (*StaticGitContext, error) {
	if staticContext.PrNumber == nil {
		return staticContext, nil
	}
	client := g.getApiClient()
	ctx := context.Background()
	pr, _, err := client.PullRequestsApi.GetV5ReposOwnerRepoPullsNumber(ctx, staticContext.Owner, staticContext.Name, int32(*staticContext.PrNumber), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR context: %w", err)
	}

	repo := *staticContext
	repo.Branch = &pr.Head.Ref
	repo.Url = pr.Head.Repo.HtmlUrl
	repo.Name = pr.Head.Repo.Name
	repo.Id = pr.Head.Repo.Name
	repo.Owner = pr.Head.Repo.Owner.Login
	return &repo, nil
}

func (g *GiteeGitProvider) performRequest(method, requestURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(context.Background(), method, requestURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.token))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("status code: %d err: %s", res.StatusCode, string(body))
	}

	return body, nil
}
