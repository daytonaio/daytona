// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitnessclient

import "time"

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
type apiRepoBranch struct {
	Name string `json:"name"`
	Sha  string `json:"sha"`
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

type ApiRepository struct {
	Id             int    `json:"id"`
	ParentId       int    `json:"parent_id"`
	Identifier     string `json:"identifier"`
	Path           string `json:"path"`
	Description    string `json:"description"`
	IsPublic       bool   `json:"is_public"`
	CreatedBy      int    `json:"created_by"`
	Created        int64  `json:"created"`
	Updated        int64  `json:"updated"`
	Size           int    `json:"size"`
	SizeUpdated    int    `json:"size_updated"`
	DefaultBranch  string `json:"default_branch"`
	ForkId         int    `json:"fork_id"`
	NumForks       int    `json:"num_forks"`
	NumPulls       int    `json:"num_pulls"`
	NumClosedPulls int    `json:"num_closed_pulls"`
	NumOpenPulls   int    `json:"num_open_pulls"`
	NumMergedPulls int    `json:"num_merged_pulls"`
	Importing      bool   `json:"importing"`
	GitUrl         string `json:"git_url"`
	Uid            string `json:"uid"`
}

type StaticContext struct {
	Id       string  `json:"id"`
	Url      string  `json:"url"`
	Name     string  `json:"name"`
	Branch   *string `json:"branch,omitempty"`
	Sha      *string `json:"sha,omitempty"`
	Owner    string  `json:"owner"`
	PrNumber *uint32 `json:"prNumber,omitempty"`
	Source   string  `json:"source"`
	Path     *string `json:"path,omitempty"`
}
