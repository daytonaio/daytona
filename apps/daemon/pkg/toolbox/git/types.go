// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

type GitAddRequest struct {
	Path string `json:"path" validate:"required"`
	// files to add (use . for all files)
	Files []string `json:"files" validate:"required"`
} // @name GitAddRequest

type GitCloneRequest struct {
	URL      string  `json:"url" validate:"required"`
	Path     string  `json:"path" validate:"required"`
	Username *string `json:"username,omitempty" validate:"optional"`
	Password *string `json:"password,omitempty" validate:"optional"`
	Branch   *string `json:"branch,omitempty" validate:"optional"`
	CommitID *string `json:"commit_id,omitempty" validate:"optional"`
} // @name GitCloneRequest

type GitCommitRequest struct {
	Path       string `json:"path" validate:"required"`
	Message    string `json:"message" validate:"required"`
	Author     string `json:"author" validate:"required"`
	Email      string `json:"email" validate:"required"`
	AllowEmpty bool   `json:"allow_empty,omitempty"`
} // @name GitCommitRequest

type GitCommitResponse struct {
	Hash string `json:"hash" validate:"required"`
} // @name GitCommitResponse

type GitBranchRequest struct {
	Path string `json:"path" validate:"required"`
	Name string `json:"name" validate:"required"`
} // @name GitBranchRequest

type GitDeleteBranchRequest struct {
	Path string `json:"path" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type ListBranchResponse struct {
	Branches []string `json:"branches" validate:"required"`
} // @name ListBranchResponse

type GitRepoRequest struct {
	Path     string  `json:"path" validate:"required"`
	Username *string `json:"username,omitempty" validate:"optional"`
	Password *string `json:"password,omitempty" validate:"optional"`
} // @name GitRepoRequest

type GitCheckoutRequest struct {
	Path   string `json:"path" validate:"required"`
	Branch string `json:"branch" validate:"required"`
} // @name GitCheckoutRequest
