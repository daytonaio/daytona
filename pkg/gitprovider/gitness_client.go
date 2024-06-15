package gitprovider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type HTTPClient struct {
	client *http.Client
	base   string
	token  string
}

// Space represents the space object within each item of the array.
type Space struct {
	ID          int64     `json:"id"`
	ParentID    int64     `json:"parent_id"`
	Path        string    `json:"path"`
	Identifier  string    `json:"identifier"`
	Description string    `json:"description"`
	IsPublic    bool      `json:"is_public"`
	CreatedBy   int64     `json:"created_by"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	UID         string    `json:"uid"`
}

// AddedBy represents the added_by object within each item of the array.
type AddedBy struct {
	ID          int64     `json:"id"`
	UID         string    `json:"uid"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email"`
	Type        string    `json:"type"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
}

// Item represents each item in the array.
type NameSpace struct {
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
	Role    string    `json:"role"`
	Space   Space     `json:"space"`
	AddedBy AddedBy   `json:"added_by"`
}

// Items represents the entire array of items.
type NameSpaces []NameSpace

type Repo struct {
	ID             int64  `json:"id"`
	ParentID       int64  `json:"parent_id"`
	Identifier     string `json:"identifier"`
	Path           string `json:"path"`
	Description    string `json:"description"`
	IsPublic       bool   `json:"is_public"`
	CreatedBy      int64  `json:"created_by"`
	Created        int64  `json:"created"`
	Updated        int64  `json:"updated"`
	Size           int    `json:"size"`
	SizeUpdated    int    `json:"size_updated"`
	DefaultBranch  string `json:"default_branch"`
	ForkID         int    `json:"fork_id"`
	NumForks       int    `json:"num_forks"`
	NumPulls       int    `json:"num_pulls"`
	NumClosedPulls int    `json:"num_closed_pulls"`
	NumOpenPulls   int    `json:"num_open_pulls"`
	NumMergedPulls int    `json:"num_merged_pulls"`
	Importing      bool   `json:"importing"`
	GitURL         string `json:"git_url"`
	UID            string `json:"uid"`
}

type Repositories []Repo

type Branch struct {
	Name string `json:"name"`
	SHA  string `json:"sha"`
}

type Branches []Branch

// PR represents the Pull Request object in the API response.
type PR struct {
	Number           int64       `json:"number"`
	Created          int64       `json:"created"`
	Edited           int64       `json:"edited"`
	State            string      `json:"state"`
	IsDraft          bool        `json:"is_draft"`
	Title            string      `json:"title"`
	Description      string      `json:"description"`
	SourceRepoID     int         `json:"source_repo_id"`
	SourceBranch     string      `json:"source_branch"`
	SourceSHA        string      `json:"source_sha"`
	TargetRepoID     int         `json:"target_repo_id"`
	TargetBranch     string      `json:"target_branch"`
	Merged           interface{} `json:"merged"`       // Use interface{} since it can be null
	MergeMethod      interface{} `json:"merge_method"` // Use interface{} since it can be null
	MergeCheckStatus string      `json:"merge_check_status"`
	MergeTargetSHA   string      `json:"merge_target_sha"`
	MergeBaseSHA     string      `json:"merge_base_sha"`
	Author           Author      `json:"author"`
	Merger           interface{} `json:"merger"` // Use interface{} since it can be null
	Stats            Stats       `json:"stats"`
}

// Author represents the author of the Pull Request.
type Author struct {
	ID          int64  `json:"id"`
	UID         string `json:"uid"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Type        string `json:"type"`
	Created     int64  `json:"created"`
	Updated     int64  `json:"updated"`
}

// Stats represents the statistics of the Pull Request.
type Stats struct {
	Commits      int `json:"commits"`
	FilesChanged int `json:"files_changed"`
}

type PullRequests []PR

func GetAPIClient(base, token string) *HTTPClient {
	return &HTTPClient{http.DefaultClient, base, token}
}

func (c *HTTPClient) GetRepos() (Repositories, error) {
	path := "user/repos"

	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", c.base, path), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	resp, err := c.client.Get(fmt.Sprintf("%s/%s", c.base, path))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var repos Repositories
	err = json.NewDecoder(resp.Body).Decode(&repos)
	if err != nil {
		return nil, err
	}
	return repos, nil
}
