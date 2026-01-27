// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

// Package options provides functional option types for configuring SDK operations.
//
// This package uses the functional options pattern to provide a clean, extensible
// API for configuring optional parameters. Each option function returns a closure
// that modifies the corresponding options struct.
//
// # Usage
//
// Options are passed as variadic arguments to SDK methods:
//
//	err := sandbox.Git.Clone(ctx, url, path,
//	    options.WithBranch("develop"),
//	    options.WithUsername("user"),
//	    options.WithPassword("token"),
//	)
//
// # Generic Apply Function
//
// The [Apply] function creates a new options struct and applies all provided
// option functions to it:
//
//	opts := options.Apply(
//	    options.WithBranch("main"),
//	    options.WithUsername("user"),
//	)
//	// opts.Branch == "main", opts.Username == "user"
package options

// Apply creates a new instance of type T and applies all provided option functions.
//
// This generic function enables a consistent pattern for applying functional options
// across different option types. It allocates a zero-value instance of T, then
// applies each option function in order.
//
// Example:
//
//	opts := options.Apply(
//	    options.WithBranch("main"),
//	    options.WithUsername("user"),
//	)
func Apply[T any](opts ...func(*T)) *T {
	result := new(T)
	for _, opt := range opts {
		opt(result)
	}
	return result
}

// GitClone holds optional parameters for [daytona.GitService.Clone].
//
// Fields are pointers to distinguish between unset values and zero values.
// Use the corresponding With* functions to set these options.
type GitClone struct {
	Branch   *string // Branch to clone (defaults to repository's default branch)
	CommitId *string // Specific commit SHA to checkout after cloning
	Username *string // Username for HTTPS authentication
	Password *string // Password or token for HTTPS authentication
}

// WithBranch sets the branch to clone instead of the repository's default branch.
//
// Example:
//
//	err := sandbox.Git.Clone(ctx, url, path, options.WithBranch("develop"))
func WithBranch(branch string) func(*GitClone) {
	return func(opts *GitClone) {
		opts.Branch = &branch
	}
}

// WithCommitId sets a specific commit SHA to checkout after cloning.
//
// The repository is first cloned, then the specified commit is checked out,
// resulting in a detached HEAD state.
//
// Example:
//
//	err := sandbox.Git.Clone(ctx, url, path, options.WithCommitId("abc123def"))
func WithCommitId(commitID string) func(*GitClone) {
	return func(opts *GitClone) {
		opts.CommitId = &commitID
	}
}

// WithUsername sets the username for HTTPS authentication when cloning.
//
// For GitHub, GitLab, and similar services, the username is typically your
// account username or a placeholder like "git" when using tokens.
//
// Example:
//
//	err := sandbox.Git.Clone(ctx, url, path,
//	    options.WithUsername("username"),
//	    options.WithPassword("github_token"),
//	)
func WithUsername(username string) func(*GitClone) {
	return func(opts *GitClone) {
		opts.Username = &username
	}
}

// WithPassword sets the password or access token for HTTPS authentication when cloning.
//
// For GitHub, use a Personal Access Token (PAT). For GitLab, use a Project
// Access Token or Personal Access Token. For Bitbucket, use an App Password.
//
// Example:
//
//	err := sandbox.Git.Clone(ctx, url, path,
//	    options.WithUsername("username"),
//	    options.WithPassword("ghp_xxxxxxxxxxxx"),
//	)
func WithPassword(password string) func(*GitClone) {
	return func(opts *GitClone) {
		opts.Password = &password
	}
}

// GitCommit holds optional parameters for [daytona.GitService.Commit].
type GitCommit struct {
	AllowEmpty *bool // Allow creating commits with no staged changes
}

// WithAllowEmpty allows creating a commit even when there are no staged changes.
//
// This is useful for triggering CI/CD pipelines or marking points in history
// without actual code changes.
//
// Example:
//
//	resp, err := sandbox.Git.Commit(ctx, path, "Trigger rebuild", author, email,
//	    options.WithAllowEmpty(true),
//	)
func WithAllowEmpty(allowEmpty bool) func(*GitCommit) {
	return func(opts *GitCommit) {
		opts.AllowEmpty = &allowEmpty
	}
}

// GitDeleteBranch holds optional parameters for [daytona.GitService.DeleteBranch].
type GitDeleteBranch struct {
	Force *bool // Force delete even if branch is not fully merged
}

// WithForce enables force deletion of a branch even if it's not fully merged.
//
// Use with caution as this can result in lost commits if the branch contains
// work that hasn't been merged elsewhere.
//
// Example:
//
//	err := sandbox.Git.DeleteBranch(ctx, path, "feature/abandoned",
//	    options.WithForce(true),
//	)
func WithForce(force bool) func(*GitDeleteBranch) {
	return func(opts *GitDeleteBranch) {
		opts.Force = &force
	}
}

// GitPush holds optional parameters for [daytona.GitService.Push].
type GitPush struct {
	Username *string // Username for HTTPS authentication
	Password *string // Password or token for HTTPS authentication
}

// WithPushUsername sets the username for HTTPS authentication when pushing.
//
// Example:
//
//	err := sandbox.Git.Push(ctx, path,
//	    options.WithPushUsername("username"),
//	    options.WithPushPassword("github_token"),
//	)
func WithPushUsername(username string) func(*GitPush) {
	return func(opts *GitPush) {
		opts.Username = &username
	}
}

// WithPushPassword sets the password or access token for HTTPS authentication when pushing.
//
// Example:
//
//	err := sandbox.Git.Push(ctx, path,
//	    options.WithPushUsername("username"),
//	    options.WithPushPassword("ghp_xxxxxxxxxxxx"),
//	)
func WithPushPassword(password string) func(*GitPush) {
	return func(opts *GitPush) {
		opts.Password = &password
	}
}

// GitPull holds optional parameters for [daytona.GitService.Pull].
type GitPull struct {
	Username *string // Username for HTTPS authentication
	Password *string // Password or token for HTTPS authentication
}

// WithPullUsername sets the username for HTTPS authentication when pulling.
//
// Example:
//
//	err := sandbox.Git.Pull(ctx, path,
//	    options.WithPullUsername("username"),
//	    options.WithPullPassword("github_token"),
//	)
func WithPullUsername(username string) func(*GitPull) {
	return func(opts *GitPull) {
		opts.Username = &username
	}
}

// WithPullPassword sets the password or access token for HTTPS authentication when pulling.
//
// Example:
//
//	err := sandbox.Git.Pull(ctx, path,
//	    options.WithPullUsername("username"),
//	    options.WithPullPassword("ghp_xxxxxxxxxxxx"),
//	)
func WithPullPassword(password string) func(*GitPull) {
	return func(opts *GitPull) {
		opts.Password = &password
	}
}
