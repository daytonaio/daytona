// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitnessclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("status code: %d err: %s", res.StatusCode, string(body))
	}

	return body, nil
}

func (g *GitnessClient) GetCommits(owner string, repositoryName string, branch *string, fromSha *string) (*[]Commit, error) {
	api, err := g.BaseURL.Parse(fmt.Sprintf("/api/v1/repos/%s/commits", url.PathEscape(fmt.Sprintf("%s/%s", owner, repositoryName))))
	if err != nil {
		return nil, fmt.Errorf("failed to parse url : %w", err)
	}

	apiURL := ""
	v := url.Values{}
	if branch != nil {
		v.Add("git_ref", *branch)
		apiURL = api.String() + "?" + v.Encode()
	} else if fromSha != nil {
		v.Add("sha", *fromSha)
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

func (g *GitnessClient) GetRepository(repoUrl string) (*Repository, error) {
	repoRef, err := g.GetRepoRef(repoUrl)
	if err != nil {
		return nil, err
	}

	repoURL, err := g.BaseURL.Parse(fmt.Sprintf("/api/v1/repos/%s", url.PathEscape(*repoRef)))
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
	branchesURL, err := g.BaseURL.Parse(fmt.Sprintf("/api/v1/repos/%s/branches", url.PathEscape(fmt.Sprintf("%s/%s", namespaceId, repositoryId))))
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

func (g *GitnessClient) GetLastCommitSha(owner string, repositoryName string, branch *string) (string, error) {
	api, err := g.BaseURL.Parse(fmt.Sprintf("/api/v1/repos/%s/commits", url.PathEscape(fmt.Sprintf("%s/%s", owner, repositoryName))))
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

func (g *GitnessClient) GetRepoRef(gitUrl string) (*string, error) {
	repoUrl := strings.TrimSuffix(gitUrl, ".git")
	parsedRepoUrl, err := url.Parse(repoUrl)
	if err != nil {
		return nil, err
	}
	parsedRepoUrl.Path = strings.TrimPrefix(parsedRepoUrl.Path, "/git/")
	parts := strings.Split(parsedRepoUrl.Path, "/")
	path := fmt.Sprintf("%s/%s", parts[0], parts[1])
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
		return Commit{}, errors.New("no commits found")
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

func (g *GitnessClient) CreateWebhook(repoId string, namespaceId string, webhook Webhook) (*Webhook, error) {
	webhookEndpoint, parseErr := g.BaseURL.Parse(fmt.Sprintf("/api/v1/repos/%s/webhooks", url.PathEscape(namespaceId+"/"+repoId)))
	if parseErr != nil {
		return nil, parseErr
	}

	jsonData, err := json.Marshal(webhook)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", webhookEndpoint.String(), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("status code: %d err: %s", resp.StatusCode, string(responseData))
	}

	var newWebhook Webhook
	if err := json.Unmarshal(responseData, &newWebhook); err != nil {
		return nil, err
	}

	return &newWebhook, nil
}

func (g *GitnessClient) DeleteWebhook(repoID string, namespaceId string, webhookID string) error {
	webhookEndpoint, parseErr := g.BaseURL.Parse(fmt.Sprintf("/api/v1/repos/%s/webhooks/%s", url.PathEscape(namespaceId+"/"+repoID), webhookID))
	if parseErr != nil {
		return parseErr
	}
	_, reqErr := g.performRequest("DELETE", webhookEndpoint.String())
	if reqErr != nil {
		return reqErr
	}

	return nil
}

func (g *GitnessClient) GetAllWebhooks(repoId string, namespaceId string) ([]*Webhook, error) {
	webhookEndpoint, parseErr := g.BaseURL.Parse(fmt.Sprintf("/api/v1/repos/%s/webhooks", url.PathEscape(namespaceId+"/"+repoId)))
	if parseErr != nil {
		return nil, parseErr
	}

	responseData, reqErr := g.performRequest("GET", webhookEndpoint.String())
	if reqErr != nil {
		return nil, reqErr
	}

	var webhookList []*Webhook
	if unmarshalErr := json.Unmarshal(responseData, &webhookList); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return webhookList, nil
}

func GetCloneUrl(protocol, host, owner, repo string) string {
	return fmt.Sprintf("%s://%s/git/%s/%s.git", protocol, host, owner, repo)
}
