package git_provider

import (
	"github.com/xanzy/go-gitlab"
)

type GitLabGitProvider struct {
	token string
}

func (g *GitLabGitProvider) GetNamespaces() ([]GitNamespace, error) {
	client := g.getApiClient()
	user, err := g.GetUserData()
	if err != nil {
		return nil, err
	}

	groupList, _, err := client.Groups.ListGroups(&gitlab.ListGroupsOptions{})
	if err != nil {
		return nil, err
	}

	namespaces := make([]GitNamespace, len(groupList))
	for i, group := range groupList {
		namespaces[i].Id = group.Name
		namespaces[i].Name = group.Name
	}

	namespaces = append(namespaces, GitNamespace{Id: user.Username, Name: user.Username})

	return namespaces, nil
}

func (g *GitLabGitProvider) GetRepositories(namespace string) ([]GitRepository, error) {
	client := g.getApiClient()
	var response []GitRepository

	repoList, _, err := client.Projects.ListProjects(&gitlab.ListProjectsOptions{})

	if err != nil {
		panic("error getting repositories from GitLab")
	}

	for _, repo := range repoList {
		response = append(response, GitRepository{
			FullName: repo.PathWithNamespace,
			Name:     repo.Name,
			Url:      repo.WebURL,
		})
	}

	return response, err
}

func (g *GitLabGitProvider) GetUserData() (GitUser, error) {
	client := g.getApiClient()

	user, _, err := client.Users.CurrentUser()
	if err != nil {
		return GitUser{}, err
	}

	return GitUser{Username: user.Username}, nil
}

func (g *GitLabGitProvider) getApiClient() *gitlab.Client {
	client, err := gitlab.NewClient(g.token)
	if err != nil {
		panic(err)
	}

	return client
}
