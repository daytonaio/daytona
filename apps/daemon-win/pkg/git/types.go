// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"time"

	"github.com/go-git/go-git/v5"
)

type GitCommitInfo struct {
	Hash      string    `json:"hash" validate:"required"`
	Author    string    `json:"author" validate:"required"`
	Email     string    `json:"email" validate:"required"`
	Message   string    `json:"message" validate:"required"`
	Timestamp time.Time `json:"timestamp" validate:"required"`
} // @name GitCommitInfo

type GitStatus struct {
	CurrentBranch   string        `json:"currentBranch" validate:"required"`
	Files           []*FileStatus `json:"fileStatus" validate:"required"`
	BranchPublished bool          `json:"branchPublished" validate:"optional"`
	Ahead           int           `json:"ahead" validate:"optional"`
	Behind          int           `json:"behind" validate:"optional"`
} // @name GitStatus

type FileStatus struct {
	Name     string `json:"name" validate:"required"`
	Extra    string `json:"extra" validate:"required"`
	Staging  Status `json:"staging" validate:"required"`
	Worktree Status `json:"worktree" validate:"required"`
} // @name FileStatus

// Status status code of a file in the Worktree
type Status string // @name Status

const (
	Unmodified         Status = "Unmodified"
	Untracked          Status = "Untracked"
	Modified           Status = "Modified"
	Added              Status = "Added"
	Deleted            Status = "Deleted"
	Renamed            Status = "Renamed"
	Copied             Status = "Copied"
	UpdatedButUnmerged Status = "Updated but unmerged"
)

var MapStatus map[git.StatusCode]Status = map[git.StatusCode]Status{
	git.Unmodified:         Unmodified,
	git.Untracked:          Untracked,
	git.Modified:           Modified,
	git.Added:              Added,
	git.Deleted:            Deleted,
	git.Renamed:            Renamed,
	git.Copied:             Copied,
	git.UpdatedButUnmerged: UpdatedButUnmerged,
}
