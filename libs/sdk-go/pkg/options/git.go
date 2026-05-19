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
	Branch            *string  // Branch to clone (defaults to repository's default branch)
	CommitId          *string  // Specific commit SHA to checkout after cloning
	Username          *string  // Username for HTTPS authentication
	Password          *string  // Password or token for HTTPS authentication
	Depth             *int     // Number of commits to fetch for a shallow clone
	SingleBranch      *bool    // Restrict clone history to one branch
	ShallowSince      *string  // Fetch commits newer than this date
	NoTags            *bool    // Skip fetching tags
	Filter            *string  // Partial clone filter, such as blob:none
	Sparse            *bool    // Initialize sparse checkout
	SparsePaths       []string // Sparse checkout paths to include
	ReferencePath     *string  // Local Git object store to borrow with --reference-if-able
	Dissociate        *bool    // Copy borrowed objects so the clone is self-contained
	RecurseSubmodules *bool    // Clone submodules recursively
	ShallowSubmodules *bool    // Use shallow clones for submodules
	FilterSubmodules  *bool    // Apply the partial clone filter to submodules
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

// WithDepth sets the shallow clone depth.
//
// Example:
//
//	err := sandbox.Git.Clone(ctx, url, path, options.WithDepth(1))
func WithDepth(depth int) func(*GitClone) {
	return func(opts *GitClone) {
		opts.Depth = &depth
	}
}

// WithSingleBranch controls whether clone history is restricted to one branch.
func WithSingleBranch(singleBranch bool) func(*GitClone) {
	return func(opts *GitClone) {
		opts.SingleBranch = &singleBranch
	}
}

// WithShallowSince fetches only history newer than the supplied date.
func WithShallowSince(shallowSince string) func(*GitClone) {
	return func(opts *GitClone) {
		opts.ShallowSince = &shallowSince
	}
}

// WithNoTags skips fetching tags during clone.
func WithNoTags(noTags bool) func(*GitClone) {
	return func(opts *GitClone) {
		opts.NoTags = &noTags
	}
}

// WithFilter sets a partial clone filter, such as "blob:none".
func WithFilter(filter string) func(*GitClone) {
	return func(opts *GitClone) {
		opts.Filter = &filter
	}
}

// WithSparse initializes sparse checkout for the clone.
func WithSparse(sparse bool) func(*GitClone) {
	return func(opts *GitClone) {
		opts.Sparse = &sparse
	}
}

// WithSparsePaths sets the sparse checkout paths to include.
func WithSparsePaths(paths []string) func(*GitClone) {
	return func(opts *GitClone) {
		opts.SparsePaths = paths
	}
}

// WithReferencePath borrows objects from a local Git repository or mirror if it exists.
func WithReferencePath(path string) func(*GitClone) {
	return func(opts *GitClone) {
		opts.ReferencePath = &path
	}
}

// WithDissociate copies borrowed reference objects into the clone before finishing.
func WithDissociate(dissociate bool) func(*GitClone) {
	return func(opts *GitClone) {
		opts.Dissociate = &dissociate
	}
}

// WithRecurseSubmodules clones submodules recursively.
func WithRecurseSubmodules(recurseSubmodules bool) func(*GitClone) {
	return func(opts *GitClone) {
		opts.RecurseSubmodules = &recurseSubmodules
	}
}

// WithShallowSubmodules uses shallow clones for submodules.
func WithShallowSubmodules(shallowSubmodules bool) func(*GitClone) {
	return func(opts *GitClone) {
		opts.ShallowSubmodules = &shallowSubmodules
	}
}

// WithFilterSubmodules applies the partial clone filter to submodules.
func WithFilterSubmodules(filterSubmodules bool) func(*GitClone) {
	return func(opts *GitClone) {
		opts.FilterSubmodules = &filterSubmodules
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
