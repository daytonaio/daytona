// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

type GitProviderConfig struct {
	Id         string  `json:"id" validate:"required"`
	Username   string  `json:"username" validate:"required"`
	BaseApiUrl *string `json:"baseApiUrl,omitempty" validate:"optional"`
	Token      string  `json:"token" validate:"required"`
} // @name GitProvider

type GitUser struct {
	Id       string `json:"id" validate:"required"`
	Username string `json:"username" validate:"required"`
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required"`
} // @name GitUser

type CloneTarget string // @name CloneTarget

const (
	CloneTargetBranch CloneTarget = "branch"
	CloneTargetCommit CloneTarget = "commit"
)

type GitRepository struct {
	Id       string      `json:"id" validate:"required"`
	Url      string      `json:"url" validate:"required"`
	Name     string      `json:"name" validate:"required"`
	Branch   *string     `json:"branch,omitempty" validate:"optional"`
	Sha      string      `json:"sha" validate:"required"`
	Owner    string      `json:"owner" validate:"required"`
	PrNumber *uint32     `json:"prNumber,omitempty" validate:"optional"`
	Source   string      `json:"source" validate:"required"`
	Path     *string     `json:"path,omitempty" validate:"optional"`
	Target   CloneTarget `json:"clonetarget,omitempty" validate:"optional"`
} // @name GitRepository

type GitNamespace struct {
	Id   string `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
} // @name GitNamespace

type GitBranch struct {
	Name string `json:"name" validate:"required"`
	Sha  string `json:"sha" validate:"required"`
} // @name GitBranch

type GitPullRequest struct {
	Name            string `json:"name" validate:"required"`
	Branch          string `json:"branch" validate:"required"`
	Sha             string `json:"sha" validate:"required"`
	SourceRepoId    string `json:"sourceRepoId" validate:"required"`
	SourceRepoUrl   string `json:"sourceRepoUrl" validate:"required"`
	SourceRepoOwner string `json:"sourceRepoOwner" validate:"required"`
	SourceRepoName  string `json:"sourceRepoName" validate:"required"`
} // @name GitPullRequest
