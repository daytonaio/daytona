// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitnessclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const personalNamespaceId = "<PERSONAL>"

type GitnessClient struct {
	Token   string
	BaseURL *url.URL
}

func NewGitnessClient(token string, baseUrl *url.URL) *GitnessClient {
	return &GitnessClient{
		Token:   token,
		BaseURL: baseUrl,
	}

}

func (g *GitnessClient) GetSpaces() ([]apiMembershipResponse, error) {
	spacesURL := g.BaseURL.ResolveReference(&url.URL{Path: "/api/v1/user/memberships"}).String()

	values := url.Values{}
	values.Add("order", "asc")
	values.Add("sort", "identifier")
	values.Add("page", "1")
	values.Add("limit", "100")
	spacesURL += "?" + values.Encode()

	req, err := http.NewRequestWithContext(context.Background(), "GET", spacesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiMemberships []apiMembershipResponse
	if err := json.Unmarshal(body, &apiMemberships); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return apiMemberships, nil
}

func (g *GitnessClient) GetUser() (*apiUserResponse, error) {
	userURL := g.BaseURL.ResolveReference(&url.URL{Path: "/api/v1/user"}).String()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", userURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	var apiUser apiUserResponse
	if err := json.Unmarshal(body, &apiUser); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return &apiUser, nil
}

func (g *GitnessClient) GetRepositories(namespace string, page string, limit string) ([]ApiRepository, error) {
	space := ""
	if namespace == personalNamespaceId {
		user, err := g.GetUser()
		if err != nil {
			return nil, err
		}
		space = user.UID
	} else {
		space = namespace
	}

	reposURL, err := g.BaseURL.Parse(fmt.Sprintf("/api/v1/spaces/%s/+/repos?page=%s&limit=%s", space, page, limit))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", reposURL.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch repositories, status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiRepos []ApiRepository
	if err := json.Unmarshal(body, &apiRepos); err != nil {
		return nil, err
	}

	return apiRepos, nil
}

func (g *GitnessClient) GetRepoBranches(repositoryId string, namespaceId string) ([]*apiRepoBranch, error) {
	branchesURL := g.BaseURL.ResolveReference(&url.URL{
		Path: fmt.Sprintf("/api/v1/repos/%s%%2F%s/branches", namespaceId, repositoryId),
	}).String()

	req, err := http.NewRequestWithContext(context.Background(), "GET", branchesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var branches []*apiRepoBranch
	if err := json.Unmarshal(body, &branches); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return branches, nil
}

func (g *GitnessClient) GetRepoPRs(repositoryId string, namespaceId string) ([]*apiPR, error) {
	prsURL := g.BaseURL.ResolveReference(&url.URL{
		Path: fmt.Sprintf("/api/v1/repos/%s%%2F%s/pullreq", namespaceId, repositoryId),
	}).String()

	values := url.Values{}
	values.Add("state", "open")
	prsURL += "?" + values.Encode()

	req, err := http.NewRequestWithContext(context.Background(), "GET", prsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiPRs []*apiPR
	if err := json.Unmarshal(body, &apiPRs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return apiPRs, nil
}

func (g *GitnessClient) GetLastCommitSha(repoURL string, branch *string) (string, error) {

	path := g.GetRepoRef(repoURL)

	apiURL := ""
	if branch != nil {
		apiURL = fmt.Sprintf("%s/api/v1/repos/%s/commits?git_ref=%s&page=1&include_stats=false", g.BaseURL.String(), path, *branch)
	} else {
		apiURL = fmt.Sprintf("%s/api/v1/repos/%s/commits?page=1&include_stats=false", g.BaseURL.String(), path)
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.Token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get commits: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	lastCommit, err := getLastCommit(body)
	if err != nil {
		return "", err
	}

	return lastCommit.Sha, nil

}

func (g *GitnessClient) GetRepoRef(url string) string {

	parts := strings.Split(url, "/")
	path := fmt.Sprintf("%s/%s", parts[3], parts[4])
	return path
}

func getLastCommit(jsonData []byte) (Commit, error) {
	var commitsResponse CommitsResponse
	err := json.Unmarshal(jsonData, &commitsResponse)
	if err != nil {
		return Commit{}, err
	}

	sort.Slice(commitsResponse.Commits, func(i, j int) bool {
		return commitsResponse.Commits[i].Committer.When.Before(commitsResponse.Commits[j].Committer.When)
	})

	if len(commitsResponse.Commits) == 0 {
		return Commit{}, fmt.Errorf("no commits found")
	}

	return commitsResponse.Commits[len(commitsResponse.Commits)-1], nil
}

func (g *GitnessClient) GetPr(repoURL string, prNumber uint32) (*PullRequest, error) {

	repoRef := g.GetRepoRef(repoURL)
	apiURL := fmt.Sprintf("%s/api/v1/repos/%s/pullreq/%d", g.BaseURL.String(), repoRef, prNumber)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.Token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var pr PullRequest
	err = json.Unmarshal(body, &pr)
	if err != nil {
		return nil, err
	}
	pr.GitUrl = fmt.Sprintf("%s/%s.git", g.BaseURL.String(), repoRef)
	return &pr, nil
}
