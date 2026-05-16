// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package gitprovider

type SigningMethod string // @name SigningMethod

const (
	SigningMethodSSH SigningMethod = "ssh"
	SigningMethodGPG SigningMethod = "gpg"
)

type GitProviderConfig struct {
	Id            string         `json:"id" validate:"required"`
	ProviderId    string         `json:"providerId" validate:"required"`
	Username      string         `json:"username" validate:"required"`
	BaseApiUrl    *string        `json:"baseApiUrl,omitempty" validate:"optional"`
	Token         string         `json:"token" validate:"required"`
	Alias         string         `json:"alias" validate:"required"`
	SigningKey    *string        `json:"signingKey,omitempty" validate:"optional"`
	SigningMethod *SigningMethod `json:"signingMethod,omitempty" validate:"optional"`
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
	Id                string      `json:"id" validate:"required"`
	Url               string      `json:"url" validate:"required"`
	Name              string      `json:"name" validate:"required"`
	Branch            string      `json:"branch" validate:"required"`
	Sha               string      `json:"sha" validate:"required"`
	Owner             string      `json:"owner" validate:"required"`
	PrNumber          *uint32     `json:"prNumber,omitempty" validate:"optional"`
	Source            string      `json:"source" validate:"required"`
	Path              *string     `json:"path,omitempty" validate:"optional"`
	Target            CloneTarget `json:"cloneTarget,omitempty" validate:"optional"`
	Depth             *int        `json:"depth,omitempty" validate:"optional"`
	SingleBranch      *bool       `json:"single_branch,omitempty" validate:"optional"`
	ShallowSince      string      `json:"shallow_since,omitempty" validate:"optional"`
	NoTags            *bool       `json:"no_tags,omitempty" validate:"optional"`
	Filter            string      `json:"filter,omitempty" validate:"optional"`
	Sparse            *bool       `json:"sparse,omitempty" validate:"optional"`
	SparsePaths       []string    `json:"sparse_paths,omitempty" validate:"optional"`
	ReferencePath     string      `json:"reference_path,omitempty" validate:"optional"`
	Dissociate        *bool       `json:"dissociate,omitempty" validate:"optional"`
	RecurseSubmodules *bool       `json:"recurse_submodules,omitempty" validate:"optional"`
	ShallowSubmodules *bool       `json:"shallow_submodules,omitempty" validate:"optional"`
	FilterSubmodules  *bool       `json:"filter_submodules,omitempty" validate:"optional"`
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

type GitEventData struct {
	Url           string   `json:"url" validate:"required"`
	Branch        string   `json:"branch" validate:"required"`
	Sha           string   `json:"sha" validate:"required"`
	Owner         string   `json:"user" validate:"required"`
	AffectedFiles []string `json:"affectedFiles" validate:"required"`
} //	@name	GitEventData
