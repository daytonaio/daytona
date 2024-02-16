package git_provider

import (
	"context"
	"fmt"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type GitHubGitProvider struct {
	token string
}

func (g *GitHubGitProvider) GetNamespaces() ([]GitNamespace, error) {
	client := g.getApiClient()
	user, err := g.GetUserData()
	if err != nil {
		return nil, err
	}

	orgList, _, err := client.Organizations.List(context.Background(), user.Username, nil)
	if err != nil {
		return nil, err
	}

	namespaces := make([]GitNamespace, len(orgList))
	for i, org := range orgList {
		if org.Login != nil {
			namespaces[i].Id = *org.Login
		}
		if org.Name != nil {
			namespaces[i].Name = *org.Name
		}
	}

	fmt.Println("test")

	namespaces = append(namespaces, GitNamespace{Id: user.Username, Name: user.Username})
	return namespaces, nil
}

func (g *GitHubGitProvider) GetRepositories(namespace string) ([]GitRepository, error) {
	client := g.getApiClient()
	var response []GitRepository

	repoList, _, err := client.Search.Repositories(context.Background(), "user:idagelic", &github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
			Page:    1,
		},
	})

	if err != nil {
		panic("error getting repositories from GitHub")
	}

	for _, repo := range repoList.Repositories {
		response = append(response, GitRepository{
			FullName: *repo.FullName,
			Name:     *repo.Name,
			Url:      *repo.HTMLURL,
		})
	}

	return response, err
}

func (g *GitHubGitProvider) GetUserData() (GitUser, error) {
	client := g.getApiClient()

	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		return GitUser{}, err
	}

	return GitUser{Username: *user.Login}, nil
}

func (g *GitHubGitProvider) getApiClient() *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: g.token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return client
}
