/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { GitApi } from '@daytona/toolbox-api-client'
import type { ListBranchResponse, ListRemotesResponse, GitStatus } from '@daytona/toolbox-api-client'
import { WithInstrumentation } from './utils/otel.decorator'

/**
 * Response from the git commit.
 *
 * @interface
 * @property {string} sha - The SHA of the commit
 */
export interface GitCommitResponse {
  sha: string
}

/**
 * Provides Git operations within a Sandbox.
 *
 * @class
 */
export class Git {
  constructor(private readonly apiClient: GitApi) {}

  /**
   * Stages the specified files for the next commit, similar to
   * running 'git add' on the command line.
   *
   * @param {string} path - Path to the Git repository root. Relative paths are resolved based on the sandbox working directory.
   * @param {string[]} files - List of file paths or directories to stage, relative to the repository root
   * @returns {Promise<void>}
   *
   * @example
   * // Stage a single file
   * await git.add('workspace/repo', ['file.txt']);
   *
   * @example
   * // Stage whole repository
   * await git.add('workspace/repo', ['.']);
   */
  @WithInstrumentation()
  public async add(path: string, files: string[]): Promise<void> {
    await this.apiClient.addFiles({
      path,
      files,
    })
  }

  /**
   * List branches in the repository.
   *
   * @param {string} path - Path to the Git repository root. Relative paths are resolved based on the sandbox working directory.
   * @returns {Promise<ListBranchResponse>} List of branches in the repository
   *
   * @example
   * const response = await git.branches('workspace/repo');
   * console.log(`Branches: ${response.branches}`);
   */
  @WithInstrumentation()
  public async branches(path: string): Promise<ListBranchResponse> {
    const response = await this.apiClient.listBranches(path)
    return response.data
  }

  /**
   * Create branch in the repository.
   *
   * @param {string} path - Path to the Git repository root. Relative paths are resolved based on the sandbox working directory.
   * @param {string} name - Name of the new branch to create
   * @returns {Promise<void>}
   *
   * @example
   * await git.createBranch('workspace/repo', 'new-feature');
   */
  @WithInstrumentation()
  public async createBranch(path: string, name: string): Promise<void> {
    await this.apiClient.createBranch({
      path,
      name,
    })
    return
  }

  /**
   * Delete branche in the repository.
   *
   * @param {string} path - Path to the Git repository root. Relative paths are resolved based on the sandbox working directory.
   * @param {string} name - Name of the branch to delete
   * @returns {Promise<void>}
   *
   * @example
   * await git.deleteBranch('workspace/repo', 'new-feature');
   */
  @WithInstrumentation()
  public async deleteBranch(path: string, name: string): Promise<void> {
    await this.apiClient.deleteBranch({
      path,
      name,
    })
    return
  }

  /**
   * Checkout branche in the repository.
   *
   * @param {string} path - Path to the Git repository root. Relative paths are resolved based on the sandbox working directory.
   * @param {string} branch - Name of the branch to checkout
   * @returns {Promise<void>}
   *
   * @example
   * await git.checkoutBranch('workspace/repo', 'new-feature');
   */
  @WithInstrumentation()
  public async checkoutBranch(path: string, branch: string): Promise<void> {
    await this.apiClient.checkoutBranch({
      path,
      branch,
    })
    return
  }

  /**
   * Clones a Git repository into the specified path. It supports
   * cloning specific branches or commits, and can authenticate with the remote
   * repository if credentials are provided.
   *
   * @param {string} url - Repository URL to clone from
   * @param {string} path - Path where the repository should be cloned. Relative paths are resolved based on the sandbox working directory.
   * @param {string} [branch] - Specific branch to clone. If not specified, clones the default branch
   * @param {string} [commitId] - Specific commit to clone. If specified, the repository will be left in a detached HEAD state at this commit
   * @param {string} [username] - Git username for authentication
   * @param {string} [password] - Git password or token for authentication
   * @param {boolean} [insecureSkipTls] - Skip TLS certificate verification (insecure). Use only for trusted internal Git servers with self-signed or private-CA certs.
   * @param {number} [depth] - Create a shallow clone truncated to the given number of commits.
   * @returns {Promise<void>}
   *
   * @example
   * // Clone the default branch
   * await git.clone(
   *   'https://github.com/user/repo.git',
   *   'workspace/repo'
   * );
   *
   * @example
   * // Clone a specific branch with authentication
   * await git.clone(
   *   'https://github.com/user/private-repo.git',
   *   'workspace/private',
   *   branch='develop',
   *   username='user',
   *   password='token'
   * );
   *
   * @example
   * // Clone a specific commit
   * await git.clone(
   *   'https://github.com/user/repo.git',
   *   'workspace/repo-old',
   *   commitId='abc123'
   * );
   */
  @WithInstrumentation()
  public async clone(
    url: string,
    path: string,
    branch?: string,
    commitId?: string,
    username?: string,
    password?: string,
    insecureSkipTls?: boolean,
    depth?: number,
  ): Promise<void> {
    await this.apiClient.cloneRepository({
      url: url,
      branch: branch,
      path,
      username,
      password,
      commit_id: commitId,
      insecure_skip_tls: insecureSkipTls,
      depth,
    })
  }

  /**
   * Commits staged changes.
   *
   * @param {string} path - Path to the Git repository root. Relative paths are resolved based on the sandbox working directory.
   * @param {string} message - Commit message describing the changes
   * @param {string} author - Name of the commit author
   * @param {string} email - Email address of the commit author
   * @param {boolean} [allowEmpty] - Allow creating an empty commit when no changes are staged
   * @returns {Promise<void>}
   *
   * @example
   * // Stage and commit changes
   * await git.add('workspace/repo', ['README.md']);
   * await git.commit(
   *   'workspace/repo',
   *   'Update documentation',
   *   'John Doe',
   *   'john@example.com',
   *   true
   * );
   *
   */
  @WithInstrumentation()
  public async commit(
    path: string,
    message: string,
    author: string,
    email: string,
    allowEmpty?: boolean,
  ): Promise<GitCommitResponse> {
    const response = await this.apiClient.commitChanges({
      path,
      message,
      author,
      email,
      allow_empty: allowEmpty,
    })
    return {
      sha: response.data.hash,
    }
  }

  /**
   * Push local changes to the remote repository.
   *
   * @param {string} path - Path to the Git repository root. Relative paths are resolved based on the sandbox working directory.
   * @param {string} [username] - Git username for authentication
   * @param {string} [password] - Git password or token for authentication
   * @param {string} [branch] - Branch to push. Defaults to the current branch
   * @param {string} [remote] - Remote to push to. Defaults to "origin"
   * @param {boolean} [setUpstream] - Record the pushed branch as the upstream tracking branch
   * @returns {Promise<void>}
   *
   * @example
   * // Push to a public repository
   * await git.push('workspace/repo');
   *
   * @example
   * // Push to a private repository
   * await git.push(
   *   'workspace/repo',
   *   'user',
   *   'token'
   * );
   *
   * @example
   * // Push a new branch and set its upstream
   * await git.push('workspace/repo', undefined, undefined, 'feature', undefined, true);
   */
  @WithInstrumentation()
  public async push(
    path: string,
    username?: string,
    password?: string,
    branch?: string,
    remote?: string,
    setUpstream?: boolean,
  ): Promise<void> {
    await this.apiClient.pushChanges({
      path,
      username,
      password,
      branch,
      remote,
      set_upstream: setUpstream,
    })
  }

  /**
   * Pulls changes from the remote repository.
   *
   * @param {string} path - Path to the Git repository root. Relative paths are resolved based on the sandbox working directory.
   * @param {string} [username] - Git username for authentication
   * @param {string} [password] - Git password or token for authentication
   * @param {string} [branch] - Branch to pull. Defaults to the current branch's upstream
   * @param {string} [remote] - Remote to pull from. Defaults to "origin"
   * @returns {Promise<void>}
   *
   * @example
   * // Pull from a public repository
   * await git.pull('workspace/repo');
   *
   * @example
   * // Pull from a private repository
   * await git.pull(
   *   'workspace/repo',
   *   'user',
   *   'token'
   * );
   *
   * @example
   * // Pull a specific branch from a specific remote
   * await git.pull('workspace/repo', undefined, undefined, 'main', 'upstream');
   */
  @WithInstrumentation()
  public async pull(
    path: string,
    username?: string,
    password?: string,
    branch?: string,
    remote?: string,
  ): Promise<void> {
    await this.apiClient.pullChanges({
      path,
      username,
      password,
      branch,
      remote,
    })
  }

  /**
   * Gets the current status of the Git repository.
   *
   * @param {string} path - Path to the Git repository root. Relative paths are resolved based on the sandbox working directory.
   * @returns {Promise<GitStatus>} Current repository status including:
   *                               - currentBranch: Name of the current branch
   *                               - ahead: Number of commits ahead of the remote branch
   *                               - behind: Number of commits behind the remote branch
   *                               - branchPublished: Whether the branch has been published to the remote repository
   *                               - fileStatus: List of file statuses
   *
   * @example
   * const status = await sandbox.git.status('workspace/repo');
   * console.log(`Current branch: ${status.currentBranch}`);
   * console.log(`Commits ahead: ${status.ahead}`);
   * console.log(`Commits behind: ${status.behind}`);
   */
  @WithInstrumentation()
  public async status(path: string): Promise<GitStatus> {
    const response = await this.apiClient.getStatus(path)
    return response.data
  }

  /**
   * Initializes a new Git repository at the specified path.
   *
   * @param {string} path - Path where the repository should be initialized. Relative paths are resolved based on the sandbox working directory.
   * @param {boolean} [bare] - Create a bare repository without a working tree
   * @param {string} [initialBranch] - Name of the initial branch. If not specified, uses the Git default
   * @returns {Promise<void>}
   *
   * @example
   * await git.init('workspace/repo', false, 'main');
   */
  @WithInstrumentation()
  public async init(path: string, bare?: boolean, initialBranch?: string): Promise<void> {
    await this.apiClient.initRepository({
      path,
      bare,
      initial_branch: initialBranch,
    })
  }

  /**
   * Resets the current HEAD to the specified state.
   *
   * @param {string} path - Path to the Git repository root. Relative paths are resolved based on the sandbox working directory.
   * @param {string} [mode] - Reset mode, one of "soft", "mixed" (default), "hard", "merge" or "keep"
   * @param {string} [target] - Revision to reset to. Defaults to HEAD
   * @param {string[]} [files] - Constrain the reset to the given paths
   * @returns {Promise<void>}
   *
   * @example
   * // Unstage all changes (mixed reset to HEAD)
   * await git.reset('workspace/repo');
   *
   * @example
   * // Hard reset to a previous commit
   * await git.reset('workspace/repo', 'hard', 'HEAD~1');
   */
  @WithInstrumentation()
  public async reset(path: string, mode?: string, target?: string, files?: string[]): Promise<void> {
    await this.apiClient.resetChanges({
      path,
      mode,
      target,
      files,
    })
  }

  /**
   * Restores working tree files or unstages changes.
   *
   * @param {string} path - Path to the Git repository root. Relative paths are resolved based on the sandbox working directory.
   * @param {string[]} files - File paths to restore
   * @param {boolean} [staged] - Restore the staging index for the given files
   * @param {boolean} [worktree] - Restore the working tree for the given files. Defaults to true when neither staged nor worktree is provided
   * @param {string} [source] - Restore file contents from the given revision instead of the index
   * @returns {Promise<void>}
   *
   * @example
   * // Discard working tree changes
   * await git.restore('workspace/repo', ['file.txt']);
   *
   * @example
   * // Unstage changes
   * await git.restore('workspace/repo', ['file.txt'], true);
   */
  @WithInstrumentation()
  public async restore(
    path: string,
    files: string[],
    staged?: boolean,
    worktree?: boolean,
    source?: string,
  ): Promise<void> {
    await this.apiClient.restoreFiles({
      path,
      files,
      staged,
      worktree,
      source,
    })
  }

  /**
   * Adds (or overwrites) a remote in the repository.
   *
   * @param {string} path - Path to the Git repository root. Relative paths are resolved based on the sandbox working directory.
   * @param {string} name - Name of the remote
   * @param {string} url - URL of the remote
   * @param {boolean} [fetch] - Fetch from the remote immediately after adding it
   * @param {boolean} [overwrite] - Replace an existing remote with the same name
   * @returns {Promise<void>}
   *
   * @example
   * await git.remoteAdd('workspace/repo', 'origin', 'https://github.com/user/repo.git');
   */
  @WithInstrumentation()
  public async remoteAdd(path: string, name: string, url: string, fetch?: boolean, overwrite?: boolean): Promise<void> {
    await this.apiClient.addRemote({
      path,
      name,
      url,
      fetch,
      overwrite,
    })
  }

  /**
   * Lists the remotes configured in the repository.
   *
   * @param {string} path - Path to the Git repository root. Relative paths are resolved based on the sandbox working directory.
   * @returns {Promise<ListRemotesResponse>} The configured remotes (name + URL)
   *
   * @example
   * const response = await git.remotes('workspace/repo');
   * response.remotes.forEach((r) => console.log(`${r.name}: ${r.url}`));
   */
  @WithInstrumentation()
  public async remotes(path: string): Promise<ListRemotesResponse> {
    const response = await this.apiClient.listRemotes(path)
    return response.data
  }

  /**
   * Gets the URL of a remote, or undefined when it does not exist.
   *
   * @param {string} path - Path to the Git repository root. Relative paths are resolved based on the sandbox working directory.
   * @param {string} name - Name of the remote
   * @returns {Promise<string | undefined>} The remote URL, or undefined when the remote does not exist
   *
   * @example
   * const url = await git.remoteGet('workspace/repo', 'origin');
   */
  @WithInstrumentation()
  public async remoteGet(path: string, name: string): Promise<string | undefined> {
    const response = await this.apiClient.listRemotes(path)
    return response.data.remotes.find((r) => r.name === name)?.url
  }

  /**
   * Sets a Git config value at the given scope.
   *
   * @param {string} key - Config key in dotted form (e.g. "user.name")
   * @param {string} value - Config value
   * @param {string} [scope] - Config scope, one of "global" (default), "local" or "system"
   * @param {string} [path] - Repository path, required when scope is "local"
   * @returns {Promise<void>}
   *
   * @example
   * await git.setConfig('user.name', 'John Doe');
   */
  @WithInstrumentation()
  public async setConfig(key: string, value: string, scope = 'global', path?: string): Promise<void> {
    await this.apiClient.setGitConfig({
      key,
      value,
      scope,
      path,
    })
  }

  /**
   * Gets a Git config value at the given scope, or undefined when unset.
   *
   * @param {string} key - Config key in dotted form (e.g. "user.name")
   * @param {string} [scope] - Config scope, one of "global" (default), "local" or "system"
   * @param {string} [path] - Repository path, required when scope is "local"
   * @returns {Promise<string | undefined>} The config value, or undefined when the key is not set
   *
   * @example
   * const name = await git.getConfig('user.name');
   */
  @WithInstrumentation()
  public async getConfig(key: string, scope = 'global', path?: string): Promise<string | undefined> {
    const response = await this.apiClient.getGitConfig(key, path, scope)
    return response.data.value
  }

  /**
   * Configures the Git user name and email at the given scope.
   *
   * @param {string} name - User name (user.name)
   * @param {string} email - User email (user.email)
   * @param {string} [scope] - Config scope, one of "global" (default), "local" or "system"
   * @param {string} [path] - Repository path, required when scope is "local"
   * @returns {Promise<void>}
   *
   * @example
   * await git.configureUser('John Doe', 'john@example.com');
   */
  @WithInstrumentation()
  public async configureUser(name: string, email: string, scope = 'global', path?: string): Promise<void> {
    await this.apiClient.configureUser({
      name,
      email,
      scope,
      path,
    })
  }

  /**
   * Persists Git credentials globally so that subsequent operations against the
   * given host authenticate automatically.
   *
   * @remarks This stores the password in plaintext on disk via the Git credential store.
   *
   * @param {string} username - Git username
   * @param {string} password - Git password or token
   * @param {string} [host] - Host to authenticate against. Defaults to "github.com"
   * @param {string} [protocol] - Protocol to authenticate against. Defaults to "https"
   * @returns {Promise<void>}
   *
   * @example
   * await git.dangerouslyAuthenticate('user', 'github_token');
   */
  @WithInstrumentation()
  public async dangerouslyAuthenticate(
    username: string,
    password: string,
    host?: string,
    protocol?: string,
  ): Promise<void> {
    await this.apiClient.authenticate({
      username,
      password,
      host,
      protocol,
    })
  }
}
