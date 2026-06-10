// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"context"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/errors"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/options"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
	toolbox "github.com/daytonaio/daytona/libs/toolbox-api-client-go"
)

// GitService provides Git operations for a sandbox.
//
// GitService enables common Git workflows including cloning repositories, staging
// and committing changes, managing branches, and syncing with remote repositories.
// It is accessed through the [Sandbox.Git] field.
//
// Example:
//
//	// Clone a repository
//	err := sandbox.Git.Clone(ctx, "https://github.com/user/repo.git", "/home/user/repo")
//
//	// Make changes and commit
//	err = sandbox.Git.Add(ctx, "/home/user/repo", []string{"."})
//	resp, err := sandbox.Git.Commit(ctx, "/home/user/repo", "Initial commit", "John Doe", "john@example.com")
//
//	// Push to remote
//	err = sandbox.Git.Push(ctx, "/home/user/repo",
//	    options.WithPushUsername("username"),
//	    options.WithPushPassword("token"),
//	)
type GitService struct {
	toolboxClient *toolbox.APIClient
	otel          *otelState
}

// NewGitService creates a new GitService with the provided toolbox client.
//
// This is typically called internally by the SDK when creating a [Sandbox].
// Users should access GitService through [Sandbox.Git] rather than creating
// it directly.
func NewGitService(toolboxClient *toolbox.APIClient, otel *otelState) *GitService {
	return &GitService{
		toolboxClient: toolboxClient,
		otel:          otel,
	}
}

// Clone clones a Git repository into the specified path.
//
// The url parameter specifies the repository URL (HTTPS or SSH format).
// The path parameter specifies the destination directory for the cloned repository.
//
// Optional parameters can be configured using functional options:
//   - [options.WithBranch]: Clone a specific branch instead of the default
//   - [options.WithCommitId]: Checkout a specific commit after cloning
//   - [options.WithUsername]: Username for authentication (HTTPS)
//   - [options.WithPassword]: Password or token for authentication (HTTPS)
//
// Example:
//
//	// Clone the default branch
//	err := sandbox.Git.Clone(ctx, "https://github.com/user/repo.git", "/home/user/repo")
//
//	// Clone a specific branch with authentication
//	err := sandbox.Git.Clone(ctx, "https://github.com/user/private-repo.git", "/home/user/repo",
//	    options.WithBranch("develop"),
//	    options.WithUsername("username"),
//	    options.WithPassword("github_token"),
//	)
//
//	// Clone and checkout a specific commit
//	err := sandbox.Git.Clone(ctx, "https://github.com/user/repo.git", "/home/user/repo",
//	    options.WithCommitId("abc123"),
//	)
//
// Returns an error if the clone operation fails.
func (g *GitService) Clone(ctx context.Context, url, path string, opts ...func(*options.GitClone)) error {
	return withInstrumentationVoid(ctx, g.otel, "Git", "Clone", func(ctx context.Context) error {
		cloneOpts := options.Apply(opts...)

		req := toolbox.NewGitCloneRequest(path, url)
		if cloneOpts.Branch != nil {
			req.SetBranch(*cloneOpts.Branch)
		}
		if cloneOpts.CommitId != nil {
			req.SetCommitId(*cloneOpts.CommitId)
		}
		if cloneOpts.Username != nil {
			req.SetUsername(*cloneOpts.Username)
		}
		if cloneOpts.Password != nil {
			req.SetPassword(*cloneOpts.Password)
		}
		if cloneOpts.InsecureSkipTLS != nil {
			req.SetInsecureSkipTls(*cloneOpts.InsecureSkipTLS)
		}
		if cloneOpts.Depth != nil {
			req.SetDepth(*cloneOpts.Depth)
		}

		httpResp, err := g.toolboxClient.GitAPI.CloneRepository(ctx).Request(*req).Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}

		return nil
	})
}

// Status returns the current Git status of a repository.
//
// The path parameter specifies the repository directory to check.
//
// The returned [types.GitStatus] contains:
//   - CurrentBranch: The name of the currently checked out branch
//   - Ahead: Number of commits ahead of the remote tracking branch
//   - Behind: Number of commits behind the remote tracking branch
//   - BranchPublished: Whether the branch has been pushed to remote
//   - FileStatus: List of files with their staging and working tree status
//
// Example:
//
//	status, err := sandbox.Git.Status(ctx, "/home/user/repo")
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("On branch %s\n", status.CurrentBranch)
//	fmt.Printf("Ahead: %d, Behind: %d\n", status.Ahead, status.Behind)
//	for _, file := range status.FileStatus {
//	    fmt.Printf("%s %s\n", file.Status, file.Path)
//	}
//
// Returns an error if the status operation fails or the path is not a Git repository.
func (g *GitService) Status(ctx context.Context, path string) (*types.GitStatus, error) {
	return withInstrumentation(ctx, g.otel, "Git", "Status", func(ctx context.Context) (*types.GitStatus, error) {
		status, httpResp, err := g.toolboxClient.GitAPI.GetStatus(ctx).Path(path).Execute()
		if err != nil {
			return nil, errors.ConvertToolboxError(err, httpResp)
		}

		// Convert pointer values to direct values
		ahead := 0
		if status.Ahead != nil {
			ahead = int(*status.Ahead)
		}
		behind := 0
		if status.Behind != nil {
			behind = int(*status.Behind)
		}
		branchPublished := false
		if status.BranchPublished != nil {
			branchPublished = *status.BranchPublished
		}

		return &types.GitStatus{
			CurrentBranch:   status.GetCurrentBranch(),
			Ahead:           ahead,
			Behind:          behind,
			BranchPublished: branchPublished,
			FileStatus:      convertFileStatus(status.GetFileStatus()),
			Detached:        status.GetDetached(),
			Upstream:        status.GetUpstream(),
		}, nil
	})
}

// Add stages files for the next commit.
//
// The path parameter specifies the repository directory.
// The files parameter is a list of file paths (relative to the repository root)
// to stage. Use "." to stage all changes.
//
// Example:
//
//	// Stage specific files
//	err := sandbox.Git.Add(ctx, "/home/user/repo", []string{"file1.txt", "src/main.go"})
//
//	// Stage all changes
//	err := sandbox.Git.Add(ctx, "/home/user/repo", []string{"."})
//
// Returns an error if the add operation fails.
func (g *GitService) Add(ctx context.Context, path string, files []string) error {
	return withInstrumentationVoid(ctx, g.otel, "Git", "Add", func(ctx context.Context) error {
		req := toolbox.NewGitAddRequest(files, path)
		httpResp, err := g.toolboxClient.GitAPI.AddFiles(ctx).Request(*req).Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}

		return nil
	})
}

// Commit creates a new Git commit with the staged changes.
//
// Parameters:
//   - path: The repository directory
//   - message: The commit message
//   - author: The author name for the commit
//   - email: The author email for the commit
//
// Optional parameters can be configured using functional options:
//   - [options.WithAllowEmpty]: Allow creating commits with no changes
//
// Example:
//
//	// Create a commit
//	resp, err := sandbox.Git.Commit(ctx, "/home/user/repo",
//	    "Add new feature",
//	    "John Doe",
//	    "john@example.com",
//	)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Created commit: %s\n", resp.SHA)
//
//	// Create an empty commit
//	resp, err := sandbox.Git.Commit(ctx, "/home/user/repo",
//	    "Empty commit for CI trigger",
//	    "John Doe",
//	    "john@example.com",
//	    options.WithAllowEmpty(true),
//	)
//
// Returns the [types.GitCommitResponse] containing the commit SHA, or an error if
// the commit fails.
func (g *GitService) Commit(ctx context.Context, path, message, author, email string, opts ...func(*options.GitCommit)) (*types.GitCommitResponse, error) {
	return withInstrumentation(ctx, g.otel, "Git", "Commit", func(ctx context.Context) (*types.GitCommitResponse, error) {
		commitOpts := options.Apply(opts...)

		req := toolbox.NewGitCommitRequest(author, email, message, path)
		if commitOpts.AllowEmpty != nil {
			req.SetAllowEmpty(*commitOpts.AllowEmpty)
		}

		result, httpResp, err := g.toolboxClient.GitAPI.CommitChanges(ctx).Request(*req).Execute()
		if err != nil {
			return nil, errors.ConvertToolboxError(err, httpResp)
		}

		return &types.GitCommitResponse{
			SHA: result.GetHash(),
		}, nil
	})
}

// Branches lists all branches in a Git repository.
//
// The path parameter specifies the repository directory.
//
// Example:
//
//	branches, err := sandbox.Git.Branches(ctx, "/home/user/repo")
//	if err != nil {
//	    return err
//	}
//	for _, branch := range branches {
//	    fmt.Println(branch)
//	}
//
// Returns a slice of branch names or an error if the operation fails.
func (g *GitService) Branches(ctx context.Context, path string) ([]string, error) {
	return withInstrumentation(ctx, g.otel, "Git", "Branches", func(ctx context.Context) ([]string, error) {
		result, httpResp, err := g.toolboxClient.GitAPI.ListBranches(ctx).Path(path).Execute()
		if err != nil {
			return nil, errors.ConvertToolboxError(err, httpResp)
		}

		return result.GetBranches(), nil
	})
}

// Checkout switches to a different branch or commit.
//
// The path parameter specifies the repository directory.
// The name parameter specifies the branch name or commit SHA to checkout.
//
// Example:
//
//	// Switch to a branch
//	err := sandbox.Git.Checkout(ctx, "/home/user/repo", "develop")
//
//	// Checkout a specific commit
//	err := sandbox.Git.Checkout(ctx, "/home/user/repo", "abc123def")
//
// Returns an error if the checkout fails (e.g., branch doesn't exist, uncommitted changes).
func (g *GitService) Checkout(ctx context.Context, path, name string) error {
	return withInstrumentationVoid(ctx, g.otel, "Git", "Checkout", func(ctx context.Context) error {
		req := toolbox.NewGitCheckoutRequest(name, path)
		httpResp, err := g.toolboxClient.GitAPI.CheckoutBranch(ctx).Request(*req).Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}

		return nil
	})
}

// CreateBranch creates a new branch at the current HEAD.
//
// The path parameter specifies the repository directory.
// The name parameter specifies the name for the new branch.
//
// Note: This creates the branch but does not switch to it. Use [GitService.Checkout]
// to switch to the new branch after creation.
//
// Example:
//
//	// Create a new branch
//	err := sandbox.Git.CreateBranch(ctx, "/home/user/repo", "feature/new-feature")
//	if err != nil {
//	    return err
//	}
//
//	// Switch to the new branch
//	err = sandbox.Git.Checkout(ctx, "/home/user/repo", "feature/new-feature")
//
// Returns an error if the branch creation fails (e.g., branch already exists).
func (g *GitService) CreateBranch(ctx context.Context, path, name string) error {
	return withInstrumentationVoid(ctx, g.otel, "Git", "CreateBranch", func(ctx context.Context) error {
		req := toolbox.NewGitBranchRequest(name, path)
		httpResp, err := g.toolboxClient.GitAPI.CreateBranch(ctx).Request(*req).Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}

		return nil
	})
}

// DeleteBranch deletes a branch from the repository.
//
// The path parameter specifies the repository directory.
// The name parameter specifies the branch to delete.
//
// Optional parameters can be configured using functional options:
//   - [options.WithForce]: Force delete the branch even if not fully merged
//
// Note: You cannot delete the currently checked out branch.
//
// Example:
//
//	// Delete a merged branch
//	err := sandbox.Git.DeleteBranch(ctx, "/home/user/repo", "feature/old-feature")
//
//	// Force delete an unmerged branch
//	err := sandbox.Git.DeleteBranch(ctx, "/home/user/repo", "feature/abandoned",
//	    options.WithForce(true),
//	)
//
// Returns an error if the deletion fails.
func (g *GitService) DeleteBranch(ctx context.Context, path, name string, opts ...func(*options.GitDeleteBranch)) error {
	return withInstrumentationVoid(ctx, g.otel, "Git", "DeleteBranch", func(ctx context.Context) error {
		// Apply options (force parameter not yet supported in toolbox API)
		_ = options.Apply(opts...)
		req := toolbox.NewGitDeleteBranchRequest(name, path)
		httpResp, err := g.toolboxClient.GitAPI.DeleteBranch(ctx).Request(*req).Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}

		return nil
	})
}

// Push pushes local commits to the remote repository.
//
// The path parameter specifies the repository directory.
//
// Optional parameters can be configured using functional options:
//   - [options.WithPushUsername]: Username for authentication
//   - [options.WithPushPassword]: Password or token for authentication
//
// Example:
//
//	// Push to a public repository (no auth required)
//	err := sandbox.Git.Push(ctx, "/home/user/repo")
//
//	// Push with authentication
//	err := sandbox.Git.Push(ctx, "/home/user/repo",
//	    options.WithPushUsername("username"),
//	    options.WithPushPassword("github_token"),
//	)
//
// Returns an error if the push fails (e.g., authentication failure, remote rejection).
func (g *GitService) Push(ctx context.Context, path string, opts ...func(*options.GitPush)) error {
	return withInstrumentationVoid(ctx, g.otel, "Git", "Push", func(ctx context.Context) error {
		pushOpts := options.Apply(opts...)

		req := toolbox.NewGitPushRequest(path)
		if pushOpts.Username != nil {
			req.SetUsername(*pushOpts.Username)
		}
		if pushOpts.Password != nil {
			req.SetPassword(*pushOpts.Password)
		}
		if pushOpts.Branch != nil {
			req.SetBranch(*pushOpts.Branch)
		}
		if pushOpts.Remote != nil {
			req.SetRemote(*pushOpts.Remote)
		}
		if pushOpts.SetUpstream != nil {
			req.SetSetUpstream(*pushOpts.SetUpstream)
		}

		httpResp, err := g.toolboxClient.GitAPI.PushChanges(ctx).Request(*req).Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}

		return nil
	})
}

// Pull fetches and merges changes from the remote repository.
//
// The path parameter specifies the repository directory.
//
// Optional parameters can be configured using functional options:
//   - [options.WithPullUsername]: Username for authentication
//   - [options.WithPullPassword]: Password or token for authentication
//
// Example:
//
//	// Pull from a public repository
//	err := sandbox.Git.Pull(ctx, "/home/user/repo")
//
//	// Pull with authentication
//	err := sandbox.Git.Pull(ctx, "/home/user/repo",
//	    options.WithPullUsername("username"),
//	    options.WithPullPassword("github_token"),
//	)
//
// Returns an error if the pull fails (e.g., merge conflicts, authentication failure).
func (g *GitService) Pull(ctx context.Context, path string, opts ...func(*options.GitPull)) error {
	return withInstrumentationVoid(ctx, g.otel, "Git", "Pull", func(ctx context.Context) error {
		pullOpts := options.Apply(opts...)

		req := toolbox.NewGitPullRequest(path)
		if pullOpts.Username != nil {
			req.SetUsername(*pullOpts.Username)
		}
		if pullOpts.Password != nil {
			req.SetPassword(*pullOpts.Password)
		}
		if pullOpts.Branch != nil {
			req.SetBranch(*pullOpts.Branch)
		}
		if pullOpts.Remote != nil {
			req.SetRemote(*pullOpts.Remote)
		}

		httpResp, err := g.toolboxClient.GitAPI.PullChanges(ctx).Request(*req).Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}

		return nil
	})
}

// Init initializes a new Git repository at the specified path.
//
// Optional parameters can be configured using functional options:
//   - [options.WithBare]: Create a bare repository without a working tree
//   - [options.WithInitialBranch]: Name of the initial branch
//
// Example:
//
//	err := sandbox.Git.Init(ctx, "/home/user/repo", options.WithInitialBranch("main"))
func (g *GitService) Init(ctx context.Context, path string, opts ...func(*options.GitInit)) error {
	return withInstrumentationVoid(ctx, g.otel, "Git", "Init", func(ctx context.Context) error {
		initOpts := options.Apply(opts...)

		req := toolbox.NewGitInitRequest(path)
		if initOpts.Bare != nil {
			req.SetBare(*initOpts.Bare)
		}
		if initOpts.InitialBranch != nil {
			req.SetInitialBranch(*initOpts.InitialBranch)
		}

		httpResp, err := g.toolboxClient.GitAPI.InitRepository(ctx).Request(*req).Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}
		return nil
	})
}

// Reset resets the current HEAD to the specified state.
//
// Optional parameters can be configured using functional options:
//   - [options.WithResetMode]: Reset mode ("soft", "mixed", "hard", "merge" or "keep")
//   - [options.WithResetTarget]: Revision to reset to (defaults to HEAD)
//   - [options.WithResetFiles]: Constrain the reset to the given paths
//
// Example:
//
//	// Unstage all changes (mixed reset to HEAD)
//	err := sandbox.Git.Reset(ctx, "/home/user/repo")
//
//	// Hard reset to a previous commit
//	err := sandbox.Git.Reset(ctx, "/home/user/repo",
//	    options.WithResetMode("hard"),
//	    options.WithResetTarget("HEAD~1"),
//	)
func (g *GitService) Reset(ctx context.Context, path string, opts ...func(*options.GitReset)) error {
	return withInstrumentationVoid(ctx, g.otel, "Git", "Reset", func(ctx context.Context) error {
		resetOpts := options.Apply(opts...)

		req := toolbox.NewGitResetRequest(path)
		if resetOpts.Mode != nil {
			req.SetMode(*resetOpts.Mode)
		}
		if resetOpts.Target != nil {
			req.SetTarget(*resetOpts.Target)
		}
		if len(resetOpts.Files) > 0 {
			req.SetFiles(resetOpts.Files)
		}

		httpResp, err := g.toolboxClient.GitAPI.ResetChanges(ctx).Request(*req).Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}
		return nil
	})
}

// Restore restores working tree files or unstages changes for the given paths.
//
// Optional parameters can be configured using functional options:
//   - [options.WithRestoreStaged]: Restore the staging index for the given files
//   - [options.WithRestoreWorktree]: Restore the working tree for the given files
//   - [options.WithRestoreSource]: Restore from the given revision instead of the index
//
// Example:
//
//	// Discard working tree changes
//	err := sandbox.Git.Restore(ctx, "/home/user/repo", []string{"file.txt"})
//
//	// Unstage changes
//	err := sandbox.Git.Restore(ctx, "/home/user/repo", []string{"file.txt"},
//	    options.WithRestoreStaged(true),
//	)
func (g *GitService) Restore(ctx context.Context, path string, files []string, opts ...func(*options.GitRestore)) error {
	return withInstrumentationVoid(ctx, g.otel, "Git", "Restore", func(ctx context.Context) error {
		restoreOpts := options.Apply(opts...)

		req := toolbox.NewGitRestoreRequest(files, path)
		if restoreOpts.Staged != nil {
			req.SetStaged(*restoreOpts.Staged)
		}
		if restoreOpts.Worktree != nil {
			req.SetWorktree(*restoreOpts.Worktree)
		}
		if restoreOpts.Source != nil {
			req.SetSource(*restoreOpts.Source)
		}

		httpResp, err := g.toolboxClient.GitAPI.RestoreFiles(ctx).Request(*req).Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}
		return nil
	})
}

// RemoteAdd adds (or overwrites) a remote in the repository.
//
// Optional parameters can be configured using functional options:
//   - [options.WithRemoteFetch]: Fetch from the remote immediately after adding it
//   - [options.WithRemoteOverwrite]: Replace an existing remote with the same name
//
// Example:
//
//	err := sandbox.Git.RemoteAdd(ctx, "/home/user/repo", "origin", "https://github.com/user/repo.git")
func (g *GitService) RemoteAdd(ctx context.Context, path, name, url string, opts ...func(*options.GitRemoteAdd)) error {
	return withInstrumentationVoid(ctx, g.otel, "Git", "RemoteAdd", func(ctx context.Context) error {
		remoteOpts := options.Apply(opts...)

		req := toolbox.NewGitAddRemoteRequest(name, path, url)
		if remoteOpts.Fetch != nil {
			req.SetFetch(*remoteOpts.Fetch)
		}
		if remoteOpts.Overwrite != nil {
			req.SetOverwrite(*remoteOpts.Overwrite)
		}

		httpResp, err := g.toolboxClient.GitAPI.AddRemote(ctx).Request(*req).Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}
		return nil
	})
}

// Remotes lists the remotes configured in the repository.
//
// Example:
//
//	remotes, err := sandbox.Git.Remotes(ctx, "/home/user/repo")
//	for _, r := range remotes {
//	    fmt.Printf("%s: %s\n", r.Name, r.URL)
//	}
func (g *GitService) Remotes(ctx context.Context, path string) ([]types.GitRemote, error) {
	return withInstrumentation(ctx, g.otel, "Git", "Remotes", func(ctx context.Context) ([]types.GitRemote, error) {
		result, httpResp, err := g.toolboxClient.GitAPI.ListRemotes(ctx).Path(path).Execute()
		if err != nil {
			return nil, errors.ConvertToolboxError(err, httpResp)
		}

		remotes := make([]types.GitRemote, 0, len(result.GetRemotes()))
		for _, r := range result.GetRemotes() {
			remotes = append(remotes, types.GitRemote{Name: r.GetName(), URL: r.GetUrl()})
		}
		return remotes, nil
	})
}

// RemoteGet returns the URL of a remote, or an empty string when it does not exist.
//
// Example:
//
//	url, err := sandbox.Git.RemoteGet(ctx, "/home/user/repo", "origin")
func (g *GitService) RemoteGet(ctx context.Context, path, name string) (string, error) {
	return withInstrumentation(ctx, g.otel, "Git", "RemoteGet", func(ctx context.Context) (string, error) {
		result, httpResp, err := g.toolboxClient.GitAPI.ListRemotes(ctx).Path(path).Execute()
		if err != nil {
			return "", errors.ConvertToolboxError(err, httpResp)
		}

		for _, r := range result.GetRemotes() {
			if r.GetName() == name {
				return r.GetUrl(), nil
			}
		}
		return "", nil
	})
}

// SetConfig sets a git config value at the given scope.
//
// Optional parameters can be configured using functional options:
//   - [options.WithConfigScope]: Config scope ("global" (default), "local" or "system")
//   - [options.WithConfigPath]: Repository path, required when scope is "local"
//
// Example:
//
//	err := sandbox.Git.SetConfig(ctx, "user.name", "John Doe")
func (g *GitService) SetConfig(ctx context.Context, key, value string, opts ...func(*options.GitConfig)) error {
	return withInstrumentationVoid(ctx, g.otel, "Git", "SetConfig", func(ctx context.Context) error {
		configOpts := options.Apply(opts...)

		req := toolbox.NewGitSetConfigRequest(key, value)
		if configOpts.Scope != nil {
			req.SetScope(*configOpts.Scope)
		}
		if configOpts.Path != nil {
			req.SetPath(*configOpts.Path)
		}

		httpResp, err := g.toolboxClient.GitAPI.SetGitConfig(ctx).Request(*req).Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}
		return nil
	})
}

// GetConfig returns a git config value at the given scope, or an empty string
// when the key is not set.
//
// Optional parameters can be configured using functional options:
//   - [options.WithConfigScope]: Config scope ("global" (default), "local" or "system")
//   - [options.WithConfigPath]: Repository path, required when scope is "local"
//
// Example:
//
//	name, err := sandbox.Git.GetConfig(ctx, "user.name")
func (g *GitService) GetConfig(ctx context.Context, key string, opts ...func(*options.GitConfig)) (string, error) {
	return withInstrumentation(ctx, g.otel, "Git", "GetConfig", func(ctx context.Context) (string, error) {
		configOpts := options.Apply(opts...)

		req := g.toolboxClient.GitAPI.GetGitConfig(ctx).Key(key)
		if configOpts.Scope != nil {
			req = req.Scope(*configOpts.Scope)
		}
		if configOpts.Path != nil {
			req = req.Path(*configOpts.Path)
		}

		result, httpResp, err := req.Execute()
		if err != nil {
			return "", errors.ConvertToolboxError(err, httpResp)
		}
		return result.GetValue(), nil
	})
}

// ConfigureUser configures the git user name and email at the given scope.
//
// Optional parameters can be configured using functional options:
//   - [options.WithConfigScope]: Config scope ("global" (default), "local" or "system")
//   - [options.WithConfigPath]: Repository path, required when scope is "local"
//
// Example:
//
//	err := sandbox.Git.ConfigureUser(ctx, "John Doe", "john@example.com")
func (g *GitService) ConfigureUser(ctx context.Context, name, email string, opts ...func(*options.GitConfig)) error {
	return withInstrumentationVoid(ctx, g.otel, "Git", "ConfigureUser", func(ctx context.Context) error {
		configOpts := options.Apply(opts...)

		req := toolbox.NewGitConfigureUserRequest(email, name)
		if configOpts.Scope != nil {
			req.SetScope(*configOpts.Scope)
		}
		if configOpts.Path != nil {
			req.SetPath(*configOpts.Path)
		}

		httpResp, err := g.toolboxClient.GitAPI.ConfigureUser(ctx).Request(*req).Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}
		return nil
	})
}

// DangerouslyAuthenticate persists git credentials globally so that subsequent
// operations against the given host authenticate automatically.
//
// This stores the password in plaintext on disk via the git credential store.
//
// Optional parameters can be configured using functional options:
//   - [options.WithAuthHost]: Host to authenticate against (defaults to "github.com")
//   - [options.WithAuthProtocol]: Protocol to authenticate against (defaults to "https")
//
// Example:
//
//	err := sandbox.Git.DangerouslyAuthenticate(ctx, "user", "github_token")
func (g *GitService) DangerouslyAuthenticate(ctx context.Context, username, password string, opts ...func(*options.GitAuthenticate)) error {
	return withInstrumentationVoid(ctx, g.otel, "Git", "DangerouslyAuthenticate", func(ctx context.Context) error {
		authOpts := options.Apply(opts...)

		req := toolbox.NewGitAuthenticateRequest(password, username)
		if authOpts.Host != nil {
			req.SetHost(*authOpts.Host)
		}
		if authOpts.Protocol != nil {
			req.SetProtocol(*authOpts.Protocol)
		}

		httpResp, err := g.toolboxClient.GitAPI.Authenticate(ctx).Request(*req).Execute()
		if err != nil {
			return errors.ConvertToolboxError(err, httpResp)
		}
		return nil
	})
}

// convertFileStatus converts toolbox FileStatus to types FileStatus
func convertFileStatus(files []toolbox.FileStatus) []types.FileStatus {
	result := make([]types.FileStatus, len(files))
	for i, file := range files {
		// Convert Status enum to single-character code
		staging := statusToCode(file.GetStaging())
		worktree := statusToCode(file.GetWorktree())

		// Combine into traditional git status format (staging + worktree)
		statusStr := string(staging) + string(worktree)

		result[i] = types.FileStatus{
			Path:   file.GetName(),
			Status: statusStr,
		}
	}
	return result
}

// statusToCode converts toolbox Status enum to git status character
func statusToCode(status toolbox.Status) rune {
	switch status {
	case toolbox.STATUS_Unmodified:
		return ' '
	case toolbox.STATUS_Modified:
		return 'M'
	case toolbox.STATUS_Added:
		return 'A'
	case toolbox.STATUS_Deleted:
		return 'D'
	case toolbox.STATUS_Renamed:
		return 'R'
	case toolbox.STATUS_Copied:
		return 'C'
	case toolbox.STATUS_Untracked:
		return '?'
	case toolbox.STATUS_UpdatedButUnmerged:
		return 'U'
	default:
		return '?'
	}
}
