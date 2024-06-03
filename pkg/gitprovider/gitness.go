package gitprovider

import (
	"net/url"
	"strconv"
	"strings"
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
	return client.GetSpaces()
}

func (g *GitNessGitProvider) getApiClient() *GitnessClient {
	url, _ := url.Parse(*g.baseApiUrl)
	return NewGitnessClient(g.token, url)
}

// func (g *GitNessGitProvider) GetRepositories(namespace string) ([]*GitRepository, error) {}

func (g *GitNessGitProvider) GetRepoBranches(repositoryId string, namespaceId string) ([]*GitBranch, error) {
	client := g.getApiClient()
	return client.GetRepoBranches(repositoryId, namespaceId)
}

func (g *GitNessGitProvider) GetRepoPRs(repositoryId string, namespaceId string) ([]*GitPullRequest, error) {
	client := g.getApiClient()
	return client.GetRepoPRs(repositoryId, namespaceId)
}

func (g *GitNessGitProvider) GetUser() (*GitUser, error) {
	client := g.getApiClient()
	return client.GetUser()
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
