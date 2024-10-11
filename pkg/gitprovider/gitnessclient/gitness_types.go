// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitnessclient

import "time"

type MembershipResponse struct {
	Created int64  `json:"created"`
	Updated int64  `json:"updated"`
	Role    string `json:"role"`
	Space   Space  `json:"space"`
	AddedBy User   `json:"added_by"`
}

type Principal struct {
	ID          int64  `json:"id"`
	UID         string `json:"uid"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Type        string `json:"type"`
	Created     int64  `json:"created"`
	Updated     int64  `json:"updated"`
}

type AddedBy struct {
	ID          int64  `json:"id"`
	UID         string `json:"uid"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Type        string `json:"type"`
	Created     int64  `json:"created"`
	Updated     int64  `json:"updated"`
}

type SpaceMemberResponse struct {
	Created   int64     `json:"created"`
	Updated   int64     `json:"updated"`
	Role      string    `json:"role"`
	Principal Principal `json:"principal"`
	AddedBy   AddedBy   `json:"added_by"`
}

type Space struct {
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

type User struct {
	ID          int    `json:"id"`
	UID         string `json:"uid"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Type        string `json:"type"`
	Created     int64  `json:"created"`
	Updated     int64  `json:"updated"`
}
type RepoBranch struct {
	Name string `json:"name"`
	Sha  string `json:"sha"`
}

type UserResponse struct {
	UID         string `json:"uid"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Admin       bool   `json:"admin"`
	Blocked     bool   `json:"blocked"`
	Created     int64  `json:"created"`
	Updated     int64  `json:"updated"`
}

type PR struct {
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
	GitUrl           string
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

type Repository struct {
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

type Webhook struct {
	Id          int64    `json:"id"`
	Version     int64    `json:"version"`
	ParentID    int64    `json:"parent_id"`
	ParentType  string   `json:"parent_type"`
	CreatedBy   int64    `json:"created_by"`
	Created     int64    `json:"created"`
	Updated     int64    `json:"updated"`
	Identifier  string   `json:"identifier"`
	DisplayName string   `json:"display_name"`
	Description string   `json:"description"`
	Url         string   `json:"url"`
	Enabled     bool     `json:"enabled"`
	Insecure    bool     `json:"insecure"`
	Triggers    []string `json:"triggers"`
	HasSecret   bool     `json:"has_secret"`
	Uid         string   `json:"uid"`
}

type WebhookEventData struct {
	Commit struct {
		Added  []string `json:"added"`
		Author struct {
			Identity struct {
				Email string `json:"email"`
				Name  string `json:"name"`
			} `json:"identity"`
			When time.Time `json:"when"`
		} `json:"author"`
		Committer struct {
			Identity struct {
				Email string `json:"email"`
				Name  string `json:"name"`
			} `json:"identity"`
			When time.Time `json:"when"`
		} `json:"committer"`
		Message  string   `json:"message"`
		Modified []string `json:"modified"`
		Removed  []string `json:"removed"`
		Sha      string   `json:"sha"`
	} `json:"commit"`
	Commits []struct {
		Added  []string `json:"added"`
		Author struct {
			Identity struct {
				Email string `json:"email"`
				Name  string `json:"name"`
			} `json:"identity"`
			When time.Time `json:"when"`
		} `json:"author"`
		Committer struct {
			Identity struct {
				Email string `json:"email"`
				Name  string `json:"name"`
			} `json:"identity"`
			When time.Time `json:"when"`
		} `json:"committer"`
		Message  string   `json:"message"`
		Modified []string `json:"modified"`
		Removed  []string `json:"removed"`
		Sha      string   `json:"sha"`
	} `json:"commits"`
	Forced     bool `json:"forced"`
	HeadCommit struct {
		Added  []string `json:"added"`
		Author struct {
			Identity struct {
				Email string `json:"email"`
				Name  string `json:"name"`
			} `json:"identity"`
			When time.Time `json:"when"`
		} `json:"author"`
		Committer struct {
			Identity struct {
				Email string `json:"email"`
				Name  string `json:"name"`
			} `json:"identity"`
			When time.Time `json:"when"`
		} `json:"committer"`
		Message  string   `json:"message"`
		Modified []string `json:"modified"`
		Removed  []string `json:"removed"`
		Sha      string   `json:"sha"`
	} `json:"head_commit"`
	OldSha    string `json:"old_sha"`
	Principal struct {
		Created     int64  `json:"created"`
		DisplayName string `json:"display_name"`
		Email       string `json:"email"`
		ID          int    `json:"id"`
		Type        string `json:"type"`
		UID         string `json:"uid"`
		Updated     int64  `json:"updated"`
	} `json:"principal"`
	Ref struct {
		Name string `json:"name"`
		Repo struct {
			DefaultBranch string `json:"default_branch"`
			Description   string `json:"description"`
			GitSSHURL     string `json:"git_ssh_url"`
			GitURL        string `json:"git_url"`
			ID            int    `json:"id"`
			Identifier    string `json:"identifier"`
			Path          string `json:"path"`
			UID           string `json:"uid"`
			URL           string `json:"url"`
		} `json:"repo"`
	} `json:"ref"`
	Repo struct {
		DefaultBranch string `json:"default_branch"`
		Description   string `json:"description"`
		GitSSHURL     string `json:"git_ssh_url"`
		GitURL        string `json:"git_url"`
		ID            int    `json:"id"`
		Identifier    string `json:"identifier"`
		Path          string `json:"path"`
		UID           string `json:"uid"`
		URL           string `json:"url"`
	} `json:"repo"`
	Sha               string `json:"sha"`
	TotalCommitsCount int    `json:"total_commits_count"`
	Trigger           string `json:"trigger"`
}