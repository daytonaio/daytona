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
	Branch          *string // Branch to clone (defaults to repository's default branch)
	CommitId        *string // Specific commit SHA to checkout after cloning
	Username        *string // Username for HTTPS authentication
	Password        *string // Password or token for HTTPS authentication
	InsecureSkipTLS *bool   // Skip TLS certificate verification (insecure). Use only for trusted internal Git servers with self-signed or private-CA certs.
	Depth           *int32  // Create a shallow clone truncated to the given number of commits
}

// WithDepth creates a shallow clone truncated to the given number of commits.
//
// Example:
//
//	err := sandbox.Git.Clone(ctx, url, path, options.WithDepth(1))
func WithDepth(depth int32) func(*GitClone) {
	return func(opts *GitClone) {
		opts.Depth = &depth
	}
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

// WithInsecureSkipTLS opts into skipping TLS certificate verification for the
// clone. Use ONLY when cloning from a trusted internal Git server with a
// self-signed or private-CA certificate. Credentials, if supplied, will be
// transmitted over an unverified TLS connection and are exposed to any MITM
// on the route.
//
// Example:
//
//	err := sandbox.Git.Clone(ctx, url, path,
//	    options.WithInsecureSkipTLS(true),
//	)
func WithInsecureSkipTLS(insecureSkipTLS bool) func(*GitClone) {
	return func(opts *GitClone) {
		opts.InsecureSkipTLS = &insecureSkipTLS
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
	Username    *string // Username for HTTPS authentication
	Password    *string // Password or token for HTTPS authentication
	Branch      *string // Branch to push (defaults to the current branch)
	Remote      *string // Remote to push to (defaults to "origin")
	SetUpstream *bool   // Record the pushed branch as the upstream tracking branch
}

// WithPushBranch sets the branch to push instead of the current branch.
func WithPushBranch(branch string) func(*GitPush) {
	return func(opts *GitPush) {
		opts.Branch = &branch
	}
}

// WithPushRemote sets the remote to push to instead of "origin".
func WithPushRemote(remote string) func(*GitPush) {
	return func(opts *GitPush) {
		opts.Remote = &remote
	}
}

// WithSetUpstream records the pushed branch as the upstream tracking branch
// (git push --set-upstream).
func WithSetUpstream(setUpstream bool) func(*GitPush) {
	return func(opts *GitPush) {
		opts.SetUpstream = &setUpstream
	}
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
	Branch   *string // Branch to pull (defaults to the current branch's upstream)
	Remote   *string // Remote to pull from (defaults to "origin")
}

// WithPullBranch sets the branch to pull instead of the current branch's upstream.
func WithPullBranch(branch string) func(*GitPull) {
	return func(opts *GitPull) {
		opts.Branch = &branch
	}
}

// WithPullRemote sets the remote to pull from instead of "origin".
func WithPullRemote(remote string) func(*GitPull) {
	return func(opts *GitPull) {
		opts.Remote = &remote
	}
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

// GitInit holds optional parameters for [daytona.GitService.Init].
type GitInit struct {
	Bare          *bool   // Create a bare repository without a working tree
	InitialBranch *string // Name of the initial branch (defaults to the Git default)
}

// WithBare creates a bare repository without a working tree.
func WithBare(bare bool) func(*GitInit) {
	return func(opts *GitInit) {
		opts.Bare = &bare
	}
}

// WithInitialBranch sets the name of the initial branch.
func WithInitialBranch(initialBranch string) func(*GitInit) {
	return func(opts *GitInit) {
		opts.InitialBranch = &initialBranch
	}
}

// GitReset holds optional parameters for [daytona.GitService.Reset].
type GitReset struct {
	Mode   *string  // Reset mode: "soft", "mixed" (default), "hard", "merge" or "keep"
	Target *string  // Revision to reset to (defaults to HEAD)
	Files  []string // Constrain the reset to the given paths
}

// WithResetMode sets the reset mode ("soft", "mixed", "hard", "merge" or "keep").
func WithResetMode(mode string) func(*GitReset) {
	return func(opts *GitReset) {
		opts.Mode = &mode
	}
}

// WithResetTarget sets the revision to reset to.
func WithResetTarget(target string) func(*GitReset) {
	return func(opts *GitReset) {
		opts.Target = &target
	}
}

// WithResetFiles constrains the reset to the given paths.
func WithResetFiles(files []string) func(*GitReset) {
	return func(opts *GitReset) {
		opts.Files = files
	}
}

// GitRestore holds optional parameters for [daytona.GitService.Restore].
type GitRestore struct {
	Staged   *bool   // Restore the staging index for the given files
	Worktree *bool   // Restore the working tree for the given files (default when neither is set)
	Source   *string // Restore file contents from the given revision instead of the index
}

// WithRestoreStaged restores the staging index for the given files.
func WithRestoreStaged(staged bool) func(*GitRestore) {
	return func(opts *GitRestore) {
		opts.Staged = &staged
	}
}

// WithRestoreWorktree restores the working tree for the given files.
func WithRestoreWorktree(worktree bool) func(*GitRestore) {
	return func(opts *GitRestore) {
		opts.Worktree = &worktree
	}
}

// WithRestoreSource restores file contents from the given revision instead of the index.
func WithRestoreSource(source string) func(*GitRestore) {
	return func(opts *GitRestore) {
		opts.Source = &source
	}
}

// GitRemoteAdd holds optional parameters for [daytona.GitService.RemoteAdd].
type GitRemoteAdd struct {
	Fetch     *bool // Fetch from the remote immediately after adding it
	Overwrite *bool // Replace an existing remote with the same name
}

// WithRemoteFetch fetches from the remote immediately after adding it.
func WithRemoteFetch(fetch bool) func(*GitRemoteAdd) {
	return func(opts *GitRemoteAdd) {
		opts.Fetch = &fetch
	}
}

// WithRemoteOverwrite replaces an existing remote with the same name.
func WithRemoteOverwrite(overwrite bool) func(*GitRemoteAdd) {
	return func(opts *GitRemoteAdd) {
		opts.Overwrite = &overwrite
	}
}

// GitConfig holds optional parameters for the git config operations
// ([daytona.GitService.SetConfig], [daytona.GitService.GetConfig] and
// [daytona.GitService.ConfigureUser]).
type GitConfig struct {
	Scope *string // Config scope: "global" (default), "local" or "system"
	Path  *string // Repository path, required when scope is "local"
}

// WithConfigScope sets the config scope ("global", "local" or "system").
func WithConfigScope(scope string) func(*GitConfig) {
	return func(opts *GitConfig) {
		opts.Scope = &scope
	}
}

// WithConfigPath sets the repository path (required for the "local" scope).
func WithConfigPath(path string) func(*GitConfig) {
	return func(opts *GitConfig) {
		opts.Path = &path
	}
}

// GitAuthenticate holds optional parameters for [daytona.GitService.DangerouslyAuthenticate].
type GitAuthenticate struct {
	Host     *string // Host to authenticate against (defaults to "github.com")
	Protocol *string // Protocol to authenticate against (defaults to "https")
}

// WithAuthHost sets the host to authenticate against.
func WithAuthHost(host string) func(*GitAuthenticate) {
	return func(opts *GitAuthenticate) {
		opts.Host = &host
	}
}

// WithAuthProtocol sets the protocol to authenticate against.
func WithAuthProtocol(protocol string) func(*GitAuthenticate) {
	return func(opts *GitAuthenticate) {
		opts.Protocol = &protocol
	}
}
