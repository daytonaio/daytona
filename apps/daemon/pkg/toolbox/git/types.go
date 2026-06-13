// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

type GitAddRequest struct {
	Path string `json:"path" validate:"required"`
	// files to add (use . for all files)
	Files []string `json:"files" validate:"required"`
} //	@name	GitAddRequest

type GitCloneRequest struct {
	URL      string  `json:"url" validate:"required"`
	Path     string  `json:"path" validate:"required"`
	Username *string `json:"username,omitempty" validate:"optional"`
	Password *string `json:"password,omitempty" validate:"optional"`
	Branch   *string `json:"branch,omitempty" validate:"optional"`
	CommitID *string `json:"commit_id,omitempty" validate:"optional"`
	// Skip TLS certificate verification for this clone. Defaults to false (verify).
	// Set to true ONLY for trusted internal Git servers with self-signed or
	// private-CA certs; credentials, if supplied, will be transmitted over an
	// unverified TLS connection and are exposed to any MITM on the route.
	InsecureSkipTLS *bool `json:"insecure_skip_tls,omitempty" validate:"optional"`
	// Depth creates a shallow clone truncated to the given number of commits.
	Depth *int32 `json:"depth,omitempty" validate:"optional"`
} //	@name	GitCloneRequest

type GitCommitRequest struct {
	Path       string `json:"path" validate:"required"`
	Message    string `json:"message" validate:"required"`
	Author     string `json:"author" validate:"required"`
	Email      string `json:"email" validate:"required"`
	AllowEmpty bool   `json:"allow_empty,omitempty"`
} //	@name	GitCommitRequest

type GitCommitResponse struct {
	Hash string `json:"hash" validate:"required"`
} //	@name	GitCommitResponse

type GitBranchRequest struct {
	Path string `json:"path" validate:"required"`
	Name string `json:"name" validate:"required"`
} //	@name	GitBranchRequest

type GitDeleteBranchRequest struct {
	Path string `json:"path" validate:"required"`
	Name string `json:"name" validate:"required"`
} //	@name	GitDeleteBranchRequest

type ListBranchResponse struct {
	Branches []string `json:"branches" validate:"required"`
	// Current is the name of the checked out branch (empty when HEAD is detached).
	Current string `json:"current,omitempty" validate:"optional"`
} //	@name	ListBranchResponse

type GitPullRequest struct {
	Path     string  `json:"path" validate:"required"`
	Username *string `json:"username,omitempty" validate:"optional"`
	Password *string `json:"password,omitempty" validate:"optional"`
	// Branch to pull (defaults to the current branch's upstream).
	Branch *string `json:"branch,omitempty" validate:"optional"`
	// Remote to pull from (defaults to "origin").
	Remote *string `json:"remote,omitempty" validate:"optional"`
} //	@name	GitPullRequest

type GitPushRequest struct {
	Path     string  `json:"path" validate:"required"`
	Username *string `json:"username,omitempty" validate:"optional"`
	Password *string `json:"password,omitempty" validate:"optional"`
	// Branch to push (defaults to the current branch).
	Branch *string `json:"branch,omitempty" validate:"optional"`
	// Remote to push to (defaults to "origin").
	Remote *string `json:"remote,omitempty" validate:"optional"`
	// SetUpstream records the pushed branch as the upstream tracking branch.
	SetUpstream *bool `json:"set_upstream,omitempty" validate:"optional"`
} //	@name	GitPushRequest

type GitCheckoutRequest struct {
	Path   string `json:"path" validate:"required"`
	Branch string `json:"branch" validate:"required"`
} //	@name	GitCheckoutRequest

type GitInitRequest struct {
	Path string `json:"path" validate:"required"`
	// Bare creates a repository without a working tree.
	Bare bool `json:"bare,omitempty"`
	// InitialBranch sets the name of the initial branch.
	InitialBranch *string `json:"initial_branch,omitempty" validate:"optional"`
} //	@name	GitInitRequest

type GitResetRequest struct {
	Path string `json:"path" validate:"required"`
	// Mode is one of soft, mixed (default), hard, merge or keep.
	Mode *string `json:"mode,omitempty" validate:"optional"`
	// Target is the revision to reset to (defaults to HEAD).
	Target *string `json:"target,omitempty" validate:"optional"`
	// Files constrains the reset to the given paths.
	Files []string `json:"files,omitempty" validate:"optional"`
} //	@name	GitResetRequest

type GitRestoreRequest struct {
	Path  string   `json:"path" validate:"required"`
	Files []string `json:"files" validate:"required"`
	// Staged restores the staging index for the given files.
	Staged *bool `json:"staged,omitempty" validate:"optional"`
	// Worktree restores the working tree for the given files.
	Worktree *bool `json:"worktree,omitempty" validate:"optional"`
	// Source restores file contents from the given revision instead of the index.
	Source *string `json:"source,omitempty" validate:"optional"`
} //	@name	GitRestoreRequest

type GitAddRemoteRequest struct {
	Path string `json:"path" validate:"required"`
	Name string `json:"name" validate:"required"`
	URL  string `json:"url" validate:"required"`
	// Fetch fetches from the remote immediately after adding it.
	Fetch bool `json:"fetch,omitempty"`
	// Overwrite replaces an existing remote with the same name.
	Overwrite bool `json:"overwrite,omitempty"`
} //	@name	GitAddRemoteRequest

type GitRemote struct {
	Name string `json:"name" validate:"required"`
	URL  string `json:"url" validate:"required"`
} //	@name	GitRemote

type ListRemotesResponse struct {
	Remotes []GitRemote `json:"remotes" validate:"required"`
} //	@name	ListRemotesResponse

type GitSetConfigRequest struct {
	// Path is the repository path, required when scope is "local".
	Path  *string `json:"path,omitempty" validate:"optional"`
	Key   string  `json:"key" validate:"required"`
	Value string  `json:"value" validate:"required"`
	// Scope is one of global (default), local or system.
	Scope *string `json:"scope,omitempty" validate:"optional"`
} //	@name	GitSetConfigRequest

type GitConfigResponse struct {
	// Value is the config value, null when the key is not set.
	Value *string `json:"value,omitempty" validate:"optional"`
} //	@name	GitConfigResponse

type GitConfigureUserRequest struct {
	// Path is the repository path, required when scope is "local".
	Path  *string `json:"path,omitempty" validate:"optional"`
	Name  string  `json:"name" validate:"required"`
	Email string  `json:"email" validate:"required"`
	// Scope is one of global (default), local or system.
	Scope *string `json:"scope,omitempty" validate:"optional"`
} //	@name	GitConfigureUserRequest

type GitAuthenticateRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	// Host defaults to github.com.
	Host *string `json:"host,omitempty" validate:"optional"`
	// Protocol defaults to https.
	Protocol *string `json:"protocol,omitempty" validate:"optional"`
} //	@name	GitAuthenticateRequest
