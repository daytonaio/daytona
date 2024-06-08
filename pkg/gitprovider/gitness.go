package gitprovider

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/daytonaio/daytona/pkg/gitnessclient"
)

type GitNessGitProvider struct {
	*AbstractGitProvider

	token      string
	baseApiUrl *string // http://localhost:3000/api/v1/
}

func NewGitNessGitProvider(token string, baseApiUrl *string) *GitNessGitProvider {
	gitProvider := &GitNessGitProvider{
		token:               token,
		baseApiUrl:          baseApiUrl,
		AbstractGitProvider: &AbstractGitProvider{},
	}
	gitProvider.AbstractGitProvider.GitProvider = gitProvider

	return gitProvider
}

func (g *GitNessGitProvider) GetNamespaces() ([]*GitNamespace, error) {
	client := g.getApiClient()
	response, err := client.GetSpaces()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Namespace : %w", err)
	}
	var namespaces []*GitNamespace
	for _, membership := range response {
		namespace := &GitNamespace{
			Id:   membership.Space.UID,
			Name: membership.Space.Identifier,
		}
		namespaces = append(namespaces, namespace)
	}
	return namespaces, nil
}

func (g *GitNessGitProvider) getApiClient() *gitnessclient.GitnessClient {
	url, _ := url.Parse(*g.baseApiUrl)
	return gitnessclient.NewGitnessClient(g.token, url)
}

func (g *GitNessGitProvider) GetRepositories(namespace string) ([]*GitRepository, error) {
	client := g.getApiClient()
	response, err := client.GetRepositories(namespace)
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
			Branch: &repo.DefaultBranch,
			Source: u.Host,
		}

		repos = append(repos, repo)
	}

	return repos, nil

}

func (g *GitNessGitProvider) GetRepoBranches(repositoryId string, namespaceId string) ([]*GitBranch, error) {
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

func (g *GitNessGitProvider) GetRepoPRs(repositoryId string, namespaceId string) ([]*GitPullRequest, error) {
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
			SourceRepoUrl:   fmt.Sprintf("%s/%s/%s", *g.baseApiUrl, namespaceId, repositoryId),
			SourceRepoOwner: pr.Author.DisplayName,
			SourceRepoName:  repositoryId,
		}
		pullRequests = append(pullRequests, pullRequest)
	}
	return pullRequests, nil
}

func (g *GitNessGitProvider) GetUser() (*GitUser, error) {
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

func (g *GitNessGitProvider) GetLastCommitSha(staticContext *StaticGitContext) (string, error) {
	client := g.getApiClient()

	return client.GetLastCommitSha(staticContext)
}

func (g *GitNessGitProvider) getPrContext(staticContext *StaticGitContext) (*StaticGitContext, error) {
	client := g.getApiClient()
	return client.getPrContext(staticContext)
}

func (g *GitNessGitProvider) parseStaticGitContext(repoUrl string) (*StaticGitContext, error) {
	staticContext, err := g.AbstractGitProvider.parseStaticGitContext(repoUrl)
	if err != nil {
		return nil, err
	}

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
	case len(parts) >= 1 && parts[0] == "files" && parts[2] != "~":
		staticContext.Branch = &parts[1]
		staticContext.Path = nil
	case len(parts) >= 2 && parts[0] == "files" && parts[2] == "~":
		staticContext.Branch = &parts[1]
		branchPath := strings.Join(parts[1:], "/")
		staticContext.Path = &branchPath
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
