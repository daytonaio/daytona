package git_provider

import (
	"strconv"

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

	namespaces := make([]GitNamespace, len(groupList)+1) // +1 for the personal namespace
	namespaces[0] = GitNamespace{Id: personalNamespaceId, Name: user.Username}

	for i, group := range groupList {
		namespaces[i+1].Id = strconv.Itoa(group.ID)
		namespaces[i+1].Name = group.Name
	}

	return namespaces, nil
}

func (g *GitLabGitProvider) GetRepositories(namespace string) ([]GitRepository, error) {
	client := g.getApiClient()
	var response []GitRepository
	var repoList []*gitlab.Project
	var err error

	if namespace == personalNamespaceId {
		user, err := g.GetUserData()
		if err != nil {
			return nil, err
		}

		repoList, _, err = client.Projects.ListUserProjects(user.Id, &gitlab.ListProjectsOptions{
			ListOptions: gitlab.ListOptions{
				PerPage: 100,
				Page:    1,
			},
		})
		if err != nil {
			return nil, err
		}
	} else {
		repoList, _, err = client.Groups.ListGroupProjects(namespace, &gitlab.ListGroupProjectsOptions{
			ListOptions: gitlab.ListOptions{
				PerPage: 100,
				Page:    1,
			},
		})
		if err != nil {
			return nil, err
		}
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

	userId := strconv.Itoa(user.ID)

	return GitUser{Id: userId, Username: user.Username}, nil
}

func (g *GitLabGitProvider) getApiClient() *gitlab.Client {
	client, err := gitlab.NewClient(g.token)
	if err != nil {
		panic(err)
	}

	return client
}
