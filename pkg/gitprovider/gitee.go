// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const giteeApiUrl = "https://gitee.com/api/v5"

type GiteeGitProvider struct {
	*AbstractGitProvider

	token      string
	baseApiUrl string
}

func NewGiteeGitProvider(token string, baseApiUrl string) *GiteeGitProvider {
	if baseApiUrl == "" {
		baseApiUrl = giteeApiUrl
	}

	provider := &GiteeGitProvider{
		token:               token,
		baseApiUrl:          baseApiUrl,
		AbstractGitProvider: &AbstractGitProvider{},
	}
	provider.AbstractGitProvider.GitProvider = provider

	return provider
}

func (g *GiteeGitProvider) CanHandle(repoUrl string) (bool, error) {
	// Handle SSH URLs first
	if strings.HasPrefix(repoUrl, "git@") {
		parts := strings.Split(strings.TrimPrefix(repoUrl, "git@"), ":")
		return len(parts) > 0 && parts[0] == "gitee.com", nil
	}

	// Handle HTTPS URLs
	u, err := url.Parse(repoUrl)
	if err != nil {
		return false, nil // Return false without error for invalid URLs
	}

	return strings.Contains(u.Host, "gitee.com"), nil
}

func (g *GiteeGitProvider) GetUser() (*GitUser, error) {
	url := fmt.Sprintf("%s/user", g.baseApiUrl)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", g.token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: %s", resp.Status)
	}

	var user struct {
		ID    int    `json:"id"`
		Login string `json:"login"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &GitUser{
		Id:       strconv.Itoa(user.ID),
		Username: user.Login,
		Name:     user.Name,
		Email:    user.Email,
	}, nil
}

func (g *GiteeGitProvider) GetNamespaces(options ListOptions) ([]*GitNamespace, error) {
	url := fmt.Sprintf("%s/user/orgs", g.baseApiUrl)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", g.token))

	// Add pagination parameters
	q := req.URL.Query()
	q.Add("page", strconv.Itoa(options.Page))
	q.Add("per_page", strconv.Itoa(options.PerPage))
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get namespaces: %s", resp.Status)
	}

	var orgs []struct {
		ID   int    `json:"id"`
		Path string `json:"path"`
		Name string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&orgs); err != nil {
		return nil, err
	}

	namespaces := make([]*GitNamespace, 0, len(orgs))
	for _, org := range orgs {
		namespaces = append(namespaces, &GitNamespace{
			Id:   strconv.Itoa(org.ID),
			Name: org.Name,
		})
	}

	// Add personal namespace on first page
	if options.Page == 1 {
		user, err := g.GetUser()
		if err != nil {
			return nil, err
		}
		namespaces = append([]*GitNamespace{{
			Id:   personalNamespaceId,
			Name: user.Username,
		}}, namespaces...)
	}

	return namespaces, nil
}

func (g *GiteeGitProvider) GetRepositories(namespace string, options ListOptions) ([]*GitRepository, error) {
	var url string
	if namespace == personalNamespaceId {
		url = fmt.Sprintf("%s/user/repos", g.baseApiUrl)
	} else {
		url = fmt.Sprintf("%s/orgs/%s/repos", g.baseApiUrl, namespace)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", g.token))

	// Add pagination parameters
	q := req.URL.Query()
	q.Add("page", strconv.Itoa(options.Page))
	q.Add("per_page", strconv.Itoa(options.PerPage))
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get repositories: %s", resp.Status)
	}

	var repos []struct {
		ID            int    `json:"id"`
		Name          string `json:"name"`
		Path          string `json:"path"`
		DefaultBranch string `json:"default_branch"`
		HTMLURL       string `json:"html_url"`
		SSHURL        string `json:"ssh_url"`
		Private       bool   `json:"private"`
		Fork          bool   `json:"fork"`
		Owner         struct {
			Login string `json:"login"`
		} `json:"owner"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}

	repositories := make([]*GitRepository, 0, len(repos))
	for _, repo := range repos {
		repositories = append(repositories, &GitRepository{
			Id:     strconv.Itoa(repo.ID),
			Name:   repo.Name,
			Branch: repo.DefaultBranch,
			Url:    repo.HTMLURL,
			Owner:  repo.Owner.Login,
			Source: "gitee.com",
		})
	}

	return repositories, nil
}

func (g *GiteeGitProvider) ParseStaticGitContext(repoUrl string) (*StaticGitContext, error) {
	// Handle SSH URLs (git@gitee.com:owner/repo.git)
	if strings.HasPrefix(repoUrl, "git@") {
		sshParts := strings.Split(strings.TrimPrefix(repoUrl, "git@gitee.com:"), "/")
		if len(sshParts) < 2 {
			return nil, fmt.Errorf("invalid SSH URL format")
		}
		owner := sshParts[0]
		name := strings.TrimSuffix(sshParts[1], ".git")
		return &StaticGitContext{
			Source: "gitee.com",
			Owner:  owner,
			Name:   name,
			Url:    repoUrl,
		}, nil
	}

	// Handle HTTPS URLs
	u, err := url.Parse(repoUrl)
	if err != nil {
		return nil, err
	}

	if !strings.Contains(u.Host, "gitee.com") {
		return nil, fmt.Errorf("not a Gitee URL")
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid URL format")
	}

	context := &StaticGitContext{
		Source: u.Host,
		Owner:  parts[0],
		Name:   parts[1],
		Url:    repoUrl,
	}

	if len(parts) > 2 {
		switch parts[2] {
		case "tree":
			if len(parts) > 3 {
				branch := parts[3]
				context.Branch = &branch
			}
		case "commit":
			if len(parts) > 3 {
				sha := parts[3]
				context.Sha = &sha
			}
		case "blob":
			if len(parts) > 4 {
				branch := parts[3]
				context.Branch = &branch
				path := strings.Join(parts[4:], "/")
				context.Path = &path
			}
		case "pulls":
			if len(parts) > 3 {
				prNum, err := strconv.ParseUint(parts[3], 10, 32)
				if err != nil {
					return nil, fmt.Errorf("invalid pull request number: %v", err)
				}
				prNumber := uint32(prNum)
				context.PrNumber = &prNumber
			}
		}
	}

	return context, nil
}

func (g *GiteeGitProvider) GetUrlFromRepo(repo *GitRepository, branch string) string {
	if branch == "" {
		return fmt.Sprintf("https://gitee.com/%s/%s", repo.Owner, repo.Name)
	}
	return fmt.Sprintf("https://gitee.com/%s/%s/tree/%s", repo.Owner, repo.Name, branch)
}
