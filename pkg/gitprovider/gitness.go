package gitprovider

import "strconv"

type GitnessGitProvider struct {
	*AbstractGitProvider

	token      string
	baseApiUrl string
}

func NewGitnessGitProvider(token string, baseApiUrl string) *GitnessGitProvider {
	provider := &GitnessGitProvider{
		token:               token,
		baseApiUrl:          baseApiUrl,
		AbstractGitProvider: &AbstractGitProvider{},
	}
	provider.AbstractGitProvider.GitProvider = provider

	return provider
}

func (g *GitnessGitProvider) GetNamespaces() ([]*GitNamespace, error) {
	return nil, nil
}

func (g *GitnessGitProvider) GetRepositories(namespace string) ([]*GitRepository, error) {
	apiClient := GetAPIClient(g.baseApiUrl, g.token)
	repos, err := apiClient.GetRepos()
	if err != nil {
		return nil, err
	}
	var reposList []*GitRepository
	for _, repo := range repos {
		reposList = append(reposList, &GitRepository{Id: strconv.FormatInt(repo.ID, 10), Name: repo.Identifier})
	}
	return reposList, nil
}

func (g *GitnessGitProvider) GetUser() (*GitUser, error) {
	return nil, nil
}

func (g *GitnessGitProvider) GetRepoBranches(repositoryId string, namespaceId string) ([]*GitBranch, error) {
	return nil, nil
}

func (g *GitnessGitProvider) GetRepoPRs(repositoryId string, namespaceId string) ([]*GitPullRequest, error) {
	return nil, nil
}

func (g *GitnessGitProvider) GetRepositoryFromUrl(repositoryUrl string) (*GitRepository, error) {
	return nil, nil
}

func (g *GitnessGitProvider) GetLastCommitSha(staticContext *StaticGitContext) (string, error) {
	return "", nil
}

func (g *GitnessGitProvider) getPrContext(staticContext *StaticGitContext) (*StaticGitContext, error) {
	return nil, nil
}

func (g *GitnessGitProvider) parseStaticGitContext(repoUrl string) (*StaticGitContext, error) {
	return nil, nil
}
