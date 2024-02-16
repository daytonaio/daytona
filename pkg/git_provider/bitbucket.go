package git_provider

import (
	"github.com/ktrysmt/go-bitbucket"
)

type BitbucketGitProvider struct {
	username string
	token    string
}

func (g *BitbucketGitProvider) GetNamespaces() ([]GitNamespace, error) {
	client := g.getApiClient()
	user, err := g.GetUserData()
	if err != nil {
		panic(err)
	}

	wsList, err := client.Workspaces.List()
	if err != nil {
		panic(err)
	}

	namespaces := make([]GitNamespace, wsList.Size)
	for i, org := range wsList.Workspaces {
		namespaces[i].Id = org.Slug
		namespaces[i].Name = org.Name
	}

	namespaces = append(namespaces, GitNamespace{Id: user.Username, Name: user.Username})

	return namespaces, nil
}

func (g *BitbucketGitProvider) GetRepositories(namespace string) ([]GitRepository, error) {
	client := g.getApiClient()
	var response []GitRepository

	repoList, err := client.Repositories.ListForAccount(&bitbucket.RepositoriesOptions{
		Owner:   namespace,
		Page:    &[]int{1}[0],
		Keyword: nil,
	})
	if err != nil {
		panic(err)
	}

	for _, repo := range repoList.Items {
		htmlLink, ok := repo.Links["html"].(map[string]interface{})
		if !ok {
			panic("Invalid HTML link")
		}

		response = append(response, GitRepository{
			FullName: repo.Full_name,
			Name:     repo.Name,
			Url:      htmlLink["href"].(string),
		})
	}

	return response, err
}

func (g *BitbucketGitProvider) GetUserData() (GitUser, error) {
	client := g.getApiClient()

	user, err := client.User.Profile()
	if err != nil {
		return GitUser{}, err
	}

	return GitUser{Username: user.Username}, nil
}

func (g *BitbucketGitProvider) getApiClient() *bitbucket.Client {
	client := bitbucket.NewBasicAuth(g.username, g.token)

	return client
}
