package gitprovider

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

// GitnessClient is a client for interacting with the Gitness API.
type GitnessClient struct {
	Token   string
	BaseURL *url.URL
}

// Intermediate structs for the API response

type apiMembershipResponse struct {
	Created int64    `json:"created"`
	Updated int64    `json:"updated"`
	Role    string   `json:"role"`
	Space   apiSpace `json:"space"`
	AddedBy apiUser  `json:"added_by"`
}

type apiSpace struct {
	ID          int    `json:"id"`
	ParentID    int    `json:"parent_id"`
	Path        string `json:"path"`
	Identifier  string `json:"identifier"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
	CreatedBy   int    `json:"created_by"`
	Created     int64  `json:"created"`
	Updated     int64  `json:"updated"`
	UID         string `json:"uid"`
}

type apiUser struct {
	ID          int    `json:"id"`
	UID         string `json:"uid"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Type        string `json:"type"`
	Created     int64  `json:"created"`
	Updated     int64  `json:"updated"`
}

type apiUserResponse struct {
	UID         string `json:"uid"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Admin       bool   `json:"admin"`
	Blocked     bool   `json:"blocked"`
	Created     int64  `json:"created"`
	Updated     int64  `json:"updated"`
}

type apiPR struct {
	Title        string `json:"title"`
	SourceBranch string `json:"source_branch"`
	SourceSha    string `json:"source_sha"`
	SourceRepoId int    `json:"source_repo_id"`
	Author       struct {
		DisplayName string `json:"display_name"`
		Email       string `json:"email"`
	} `json:"author"`
}

type Commit struct {
	Sha        string   `json:"sha"`
	ParentShas []string `json:"parent_shas"`
	Title      string   `json:"title"`
	Message    string   `json:"message"`
	Author     struct {
		Identity struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"identity"`
		When time.Time `json:"when"`
	} `json:"author"`
	Committer struct {
		Identity struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"identity"`
		When time.Time `json:"when"`
	} `json:"committer"`
	Stats struct {
		Total struct {
			Insertions int `json:"insertions"`
			Deletions  int `json:"deletions"`
			Changes    int `json:"changes"`
		} `json:"total"`
	} `json:"stats"`
}
type CommitsResponse struct {
	Commits []Commit `json:"commits"`
}

type PullRequest struct {
	Number           int     `json:"number"`
	Created          int64   `json:"created"`
	Edited           int64   `json:"edited"`
	State            string  `json:"state"`
	IsDraft          bool    `json:"is_draft"`
	Title            string  `json:"title"`
	Description      string  `json:"description"`
	SourceRepoID     int     `json:"source_repo_id"`
	SourceBranch     string  `json:"source_branch"`
	SourceSha        string  `json:"source_sha"`
	TargetRepoID     int     `json:"target_repo_id"`
	TargetBranch     string  `json:"target_branch"`
	Merged           *bool   `json:"merged"`
	MergeMethod      *string `json:"merge_method"`
	MergeCheckStatus string  `json:"merge_check_status"`
	MergeTargetSha   string  `json:"merge_target_sha"`
	MergeBaseSha     string  `json:"merge_base_sha"`
	Author           struct {
		ID          int    `json:"id"`
		UID         string `json:"uid"`
		DisplayName string `json:"display_name"`
		Email       string `json:"email"`
		Type        string `json:"type"`
		Created     int64  `json:"created"`
		Updated     int64  `json:"updated"`
	} `json:"author"`
	Merger *struct {
		ID          int    `json:"id"`
		UID         string `json:"uid"`
		DisplayName string `json:"display_name"`
		Email       string `json:"email"`
		Type        string `json:"type"`
		Created     int64  `json:"created"`
		Updated     int64  `json:"updated"`
	} `json:"merger"`
	Stats struct {
		Commits      int `json:"commits"`
		FilesChanged int `json:"files_changed"`
	} `json:"stats"`
}

func NewGitnessClient(token string, baseUrl *url.URL) *GitnessClient {
	return &GitnessClient{
		Token:   token,
		BaseURL: baseUrl,
	}

}
func (g *GitnessClient) GetUser() (*GitUser, error) {
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

	user := &GitUser{
		Id:       apiUser.UID,
		Username: apiUser.UID,
		Name:     apiUser.DisplayName,
		Email:    apiUser.Email,
	}

	return user, nil
}

func (g *GitnessClient) GetSpaces() ([]*GitNamespace, error) {
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

	var namespaces []*GitNamespace
	for _, membership := range apiMemberships {
		namespace := &GitNamespace{
			Id:   membership.Space.UID,
			Name: membership.Space.Identifier,
		}
		namespaces = append(namespaces, namespace)
	}

	return namespaces, nil
}

//TODO: func (g *GitnessClient)GetRepositories(namespace string) ([] *GitRepository, error){}

func (g *GitnessClient) GetRepoBranches(repositoryId string, namespaceId string) ([]*GitBranch, error) {
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

	var branches []*GitBranch
	if err := json.Unmarshal(body, &branches); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return branches, nil
}

func (g *GitnessClient) GetRepoPRs(repositoryId string, namespaceId string) ([]*GitPullRequest, error) {
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

	var apiPRs []apiPR
	if err := json.Unmarshal(body, &apiPRs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	var pullRequests []*GitPullRequest
	for _, pr := range apiPRs {
		pullRequest := &GitPullRequest{
			Name:            pr.Title,
			Branch:          pr.SourceBranch,
			Sha:             pr.SourceSha,
			SourceRepoId:    fmt.Sprintf("%d", pr.SourceRepoId),
			SourceRepoUrl:   fmt.Sprintf("%s/%s/%s", g.BaseURL.String(), namespaceId, repositoryId),
			SourceRepoOwner: pr.Author.DisplayName,
			SourceRepoName:  repositoryId,
		}
		pullRequests = append(pullRequests, pullRequest)
	}

	return pullRequests, nil
}

func (g *GitnessClient) GetLastCommitSha(staticContext *StaticGitContext) (string, error) {

	path := getRepoRef(staticContext.Url)

	apiURL := ""
	if staticContext.Branch != nil {
		apiURL = fmt.Sprintf("%s/api/v1/repos/%s/commits?git_ref=%s&page=1&include_stats=false", g.BaseURL.String(), path, *staticContext.Branch)
	} else {
		// In this case gitness will use default branch of the repository
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

	// Use the getLastCommit function to get the last commit
	lastCommit, err := getLastCommit(body)
	if err != nil {
		return "", err
	}

	// Return the SHA of the last commit
	return lastCommit.Sha, nil

}

func getRepoRef(url string) string {

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

	// Sort the commits by the "when" field
	sort.Slice(commitsResponse.Commits, func(i, j int) bool {
		return commitsResponse.Commits[i].Committer.When.Before(commitsResponse.Commits[j].Committer.When)
	})

	// Return the last commit
	if len(commitsResponse.Commits) == 0 {
		return Commit{}, fmt.Errorf("no commits found")
	}

	return commitsResponse.Commits[len(commitsResponse.Commits)-1], nil
}

func (g *GitnessClient) getPrContext(staticContext *StaticGitContext) (*StaticGitContext, error) {
	if staticContext.PrNumber == nil {
		return staticContext, nil
	}
	repoRef := getRepoRef(staticContext.Url)
	apiURL := fmt.Sprintf("%s/api/v1/repos/%s/pullreq/%d", g.BaseURL.String(), repoRef, *staticContext.PrNumber)

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
	repo := *staticContext
	repo.Branch = &pr.SourceBranch
	repo.Url = fmt.Sprintf("%s/%s.git", g.BaseURL.String(), repoRef)
	repo.Id = staticContext.Name
	repo.Name = staticContext.Name
	repo.Owner = pr.Author.UID

	return &repo, nil
}
