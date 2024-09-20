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
	token   string
	BaseURL *url.URL
}

func NewGitnessClient(token string, baseUrl *url.URL) *GitnessClient {
	return &GitnessClient{
		token:   token,
		BaseURL: baseUrl,
	}
}

func (g *GitnessClient) performRequest(method, requestURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(context.Background(), method, requestURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.token)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d err: %s", res.StatusCode, string(body))
	}

	return body, nil
}

func (g *GitnessClient) GetCommits(owner string, repositoryName string, branch *string) (*[]Commit, error) {
	api, err := g.BaseURL.Parse(fmt.Sprintf("/api/v1/repos/%s/commits", url.PathEscape(fmt.Sprintf("%s/%s", owner, repositoryName))))
	if err != nil {
		return nil, fmt.Errorf("failed to parse url : %w", err)
	}

	apiURL := ""
	if branch != nil {
		v := url.Values{}
		v.Add("git_ref", *branch)
		apiURL = api.String() + "?" + v.Encode()
	} else {
		apiURL = api.String()
	}

	body, err := g.performRequest("GET", apiURL)
	if err != nil {
		return nil, fmt.Errorf("error while making request: %s", err.Error())
	}

	var commitsResponse CommitsResponse
	err = json.Unmarshal(body, &commitsResponse)
	if err != nil {
		return nil, err
	}

	return &commitsResponse.Commits, nil
}

func (g *GitnessClient) GetSpaceAdmin(spaceName string) (*SpaceMemberResponse, error) {
	spacesURL, err := g.BaseURL.Parse(fmt.Sprintf("/api/v1/spaces/%s/members", spaceName))
	if err != nil {
		return nil, err
	}

	body, err := g.performRequest("GET", spacesURL.String())
	if err != nil {
		return nil, err
	}

	var apiMemberships []SpaceMemberResponse
	if err := json.Unmarshal(body, &apiMemberships); err != nil {
		return nil, err
	}
	var admin *SpaceMemberResponse
	for _, member := range apiMemberships {
		if member.Role == "space_owner" {
			admin = &member
			break
		}
	}
	if admin == nil {
		return nil, err
	}

	return admin, nil
}

func (g *GitnessClient) GetSpaces() ([]MembershipResponse, error) {
	spacesURL, err := g.BaseURL.Parse("/api/v1/user/memberships")
	if err != nil {
		return nil, err
	}
	values := url.Values{}
	values.Add("order", "asc")
	values.Add("sort", "identifier")
	values.Add("page", "1")
	values.Add("limit", "100")
	apiUrl := spacesURL.String() + "?" + values.Encode()

	body, err := g.performRequest("GET", apiUrl)
	if err != nil {
		return nil, err
	}

	var apiMemberships []MembershipResponse
	if err := json.Unmarshal(body, &apiMemberships); err != nil {
		return nil, err
	}

	return apiMemberships, nil
}

func (g *GitnessClient) GetUser() (*UserResponse, error) {
	userURL, err := g.BaseURL.Parse("/api/v1/user")
	if err != nil {
		return nil, err
	}

	body, err := g.performRequest("GET", userURL.String())
	if err != nil {
		return nil, err
	}

	var apiUser UserResponse
	if err := json.Unmarshal(body, &apiUser); err != nil {
		return nil, err
	}

	return &apiUser, nil
}

func (g *GitnessClient) GetRepositories(namespace string) ([]Repository, error) {
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
		return nil, err
	}

	body, err := g.performRequest("GET", reposURL.String())
	if err != nil {
		return nil, err
	}

	var apiRepos []Repository
	if err := json.Unmarshal(body, &apiRepos); err != nil {
		return nil, err
	}

	return apiRepos, nil
}

func (g *GitnessClient) GetRepository(url string) (*Repository, error) {
	repoRef, err := g.GetRepoRef(url)
	if err != nil {
		return nil, err
	}

	repoURL, err := g.BaseURL.Parse(fmt.Sprintf("/api/v1/repos/%s", *repoRef))
	if err != nil {
		return nil, err
	}

	body, err := g.performRequest("GET", repoURL.String())
	if err != nil {
		return nil, err
	}

	var repo Repository
	if err := json.Unmarshal(body, &repo); err != nil {
		return nil, err
	}

	return &repo, nil
}

func (g *GitnessClient) GetRepoBranches(repositoryId string, namespaceId string) ([]*RepoBranch, error) {
	branchesURL, err := g.BaseURL.Parse(fmt.Sprintf("/api/v1/repos/%s/branches", url.PathEscape(namespaceId+"/"+repositoryId)))
	if err != nil {
		return nil, err
	}

	body, err := g.performRequest("GET", branchesURL.String())
	if err != nil {
		return nil, err
	}

	var branches []*RepoBranch
	if err := json.Unmarshal(body, &branches); err != nil {
		return nil, err
	}

	return branches, nil
}

func (g *GitnessClient) GetRepoPRs(repositoryId string, namespaceId string) ([]*PR, error) {
	prsURL, err := g.BaseURL.Parse(fmt.Sprintf("/api/v1/repos/%s/pullreq", url.PathEscape(namespaceId+"/"+repositoryId)))
	if err != nil {
		return nil, err
	}
	values := url.Values{}
	values.Add("state", "open")
	apiUrl := prsURL.String() + "?" + values.Encode()

	body, err := g.performRequest("GET", apiUrl)
	if err != nil {
		return nil, err
	}

	var apiPRs []*PR
	if err := json.Unmarshal(body, &apiPRs); err != nil {
		return nil, err
	}

	return apiPRs, nil
}

func (g *GitnessClient) GetLastCommitSha(repoURL string, branch *string) (string, error) {
	ref, err := g.GetRepoRef(repoURL)
	if err != nil {
		return "", err
	}
	api, err := g.BaseURL.Parse(fmt.Sprintf("/api/v1/repos/%s/commits", url.PathEscape(*ref)))
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

	body, err := g.performRequest("GET", apiURL)
	if err != nil {
		return "", fmt.Errorf("error while making request: %w", err)
	}

	lastCommit, err := getLastCommit(body)
	if err != nil {
		return "", fmt.Errorf("error while fetching last commit from list: %w", err)
	}

	return lastCommit.Sha, nil
}

func (g *GitnessClient) GetRepoRef(url string) (*string, error) {
	repoUrl := strings.TrimSuffix(url, ".git")
	parts := strings.Split(repoUrl, "/")
	if len(parts) < 5 {
		return nil, fmt.Errorf("failed to parse repository reference: invalid url passed")
	}
	var path string
	if parts[3] == "git" && len(parts) >= 6 {
		path = fmt.Sprintf("%s/%s", parts[4], parts[5])
	} else {
		path = fmt.Sprintf("%s/%s", parts[3], parts[4])
	}
	return &path, nil
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
	repoRef, err := g.GetRepoRef(repoURL)
	if err != nil {
		return nil, err
	}
	apiUrl, err := g.BaseURL.Parse(fmt.Sprintf("/api/v1/repos/%s/pullreq/%d", url.PathEscape(*repoRef), prNumber))
	if err != nil {
		return nil, err
	}

	body, err := g.performRequest("GET", apiUrl.String())
	if err != nil {
		return nil, err
	}

	var pr PullRequest
	err = json.Unmarshal(body, &pr)
	if err != nil {
		return nil, err
	}
	refPart := strings.Split(*repoRef, "/")
	pr.GitUrl = GetCloneUrl(g.BaseURL.Scheme, g.BaseURL.Host, refPart[0], refPart[1])
	return &pr, nil
}

func (g *GitnessClient) GetDefaultBranch(url string) (*string, error) {
	repo, err := g.GetRepository(url)
	if err != nil {
		return nil, err
	}

	return &repo.DefaultBranch, nil
}

func GetCloneUrl(protocol, host, owner, repo string) string {
	return fmt.Sprintf("%s://%s/git/%s/%s.git", protocol, host, owner, repo)
}
