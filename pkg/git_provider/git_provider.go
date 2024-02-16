package git_provider

import "github.com/daytonaio/daytona/common/grpc/proto/types"

type GitProvider interface {
	GetNamespaces() ([]GitNamespace, error)
	GetRepositories(namespace string) ([]GitRepository, error)
	GetUserData() (GitUser, error)
}

type GitUser struct {
	Username string
}

type GitNamespace struct {
	Id   string
	Name string
}

type GitRepository struct {
	FullName string
	Name     string
	Url      string
}

func CreateGitProvider(providerId string, gitProviders []*types.GitProvider) GitProvider {
	var chosenProvider *types.GitProvider
	for _, gitProvider := range gitProviders {
		if gitProvider.Id == providerId {
			chosenProvider = gitProvider
		}
	}

	switch providerId {
	case "github":
		return &GitHubGitProvider{
			token: chosenProvider.Token,
		}
	case "gitlab":
		return &GitLabGitProvider{
			token: chosenProvider.Token,
		}
	case "bitbucket":
		return &BitbucketGitProvider{
			username: chosenProvider.Username,
			token:    chosenProvider.Token,
		}
	default:
		return nil
	}
}
