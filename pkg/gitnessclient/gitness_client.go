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

func (g *GitnessClient) GetSpaceAdmin(spaceName string) (*apiSpaceMemberResponse, error) {

	spacesURL, err := g.BaseURL.Parse(fmt.Sprintf("/api/v1/spaces/%s/members", spaceName))
	if err != nil {
		return nil, fmt.Errorf("failed to parse url : %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "GET", spacesURL.String(), nil)
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

	var apiMemberships []apiSpaceMemberResponse
	if err := json.Unmarshal(body, &apiMemberships); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}
	var admin *apiSpaceMemberResponse
	for _, member := range apiMemberships {
		if member.Role == "space_owner" {
			admin = &member
			break
		}
	}
	if admin == nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}
	return admin, nil

}

func (g *GitnessClient) GetSpaces() ([]apiMembershipResponse, error) {
	spacesURL, err := g.BaseURL.Parse("/api/v1/user/memberships")
	if err != nil {
		return nil, fmt.Errorf("failed to parse url : %w", err)
	}
	values := url.Values{}
	values.Add("order", "asc")
	values.Add("sort", "identifier")
	values.Add("page", "1")
	values.Add("limit", "100")
	apiUrl := spacesURL.String() + "?" + values.Encode()

	req, err := http.NewRequestWithContext(context.Background(), "GET", apiUrl, nil)
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
	userURL, err := g.BaseURL.Parse("/api/v1/user")
	if err != nil {
		return nil, fmt.Errorf("failed to parse url : %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "GET", userURL.String(), nil)
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

func (g *GitnessClient) GetRepositories(namespace string) ([]ApiRepository, error) {
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

	reposURL, err := g.BaseURL.Parse(fmt.Sprintf("/api/v1/spaces/%s/+/repos", space))
	if err != nil {
		return nil, fmt.Errorf("failed to parse url : %w", err)
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
	branchesURL, err := g.BaseURL.Parse(fmt.Sprintf("/api/v1/repos/%s/branches", url.PathEscape(namespaceId+"/"+repositoryId)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse url : %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "GET", branchesURL.String(), nil)
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
	prsURL, err := g.BaseURL.Parse(fmt.Sprintf("/api/v1/repos/%s/pullreq", url.PathEscape(namespaceId+"/"+repositoryId)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse url : %w", err)
	}
	values := url.Values{}
	values.Add("state", "open")
	apiUrl := prsURL.String() + "?" + values.Encode()

	req, err := http.NewRequestWithContext(context.Background(), "GET", apiUrl, nil)
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
	ref := g.GetRepoRef(repoURL)
	api, err := g.BaseURL.Parse(fmt.Sprintf("/api/v1/repos/%s/commits", url.PathEscape(ref)))
	if err != nil {
		return "", fmt.Errorf("failed to parse url : %w", err)
	}
	apiURL := ""
	if branch != nil {
		v := url.Values{}
		v.Add("git_ref", *branch)
		apiURL = api.String() + "?" + v.Encode()
	} else {
		apiURL = api.String()
	}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("error while making reuest: %s", err.Error())
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.Token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error while making request: %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get commits status code %d: ", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error while reading response: %s", err.Error())
	}

	lastCommit, err := getLastCommit(body)
	if err != nil {
		return "", fmt.Errorf("error while fetching last commit from list: %s", err.Error())
	}

	return lastCommit.Sha, nil

}

func (g *GitnessClient) GetRepoRef(url string) string {
	repoUrl := strings.TrimSuffix(url, ".git")
	parts := strings.Split(repoUrl, "/")
	var path string
	if parts[3] == "git" {
		path = fmt.Sprintf("%s/%s", parts[4], parts[5])
	} else {
		path = fmt.Sprintf("%s/%s", parts[3], parts[4])
	}

	return path
}
func getLastCommit(jsonData []byte) (Commit, error) {
	var commitsResponse CommitsResponse
	err := json.Unmarshal(jsonData, &commitsResponse)
	if err != nil {
		return Commit{}, fmt.Errorf("json Unmarshling failed: %s", err.Error())
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

	apiUrl, err := g.BaseURL.Parse(fmt.Sprintf("/api/v1/repos/%s/pullreq/%d", url.PathEscape(repoRef), prNumber))
	if err != nil {
		return nil, fmt.Errorf("failed to parse url : %w", err)
	}

	req, err := http.NewRequest("GET", apiUrl.String(), nil)
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
	refPart := strings.Split(repoRef, "/")
	pr.GitUrl = GetGitnessCloneUrl(g.BaseURL.Host, refPart[0], refPart[1])
	return &pr, nil
}

func GetGitnessCloneUrl(source, owner, repo string) string {

	return fmt.Sprintf("https://%s/git/%s/%s.git", source, owner, repo)

}
