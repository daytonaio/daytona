/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { GitApi, ListBranchResponse, GitStatus } from '@daytonaio/toolbox-api-client'
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
  ): Promise<void> {
    await this.apiClient.cloneRepository({
      url: url,
      branch: branch,
      path,
      username,
      password,
      commit_id: commitId,
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
   */
  @WithInstrumentation()
  public async push(path: string, username?: string, password?: string): Promise<void> {
    await this.apiClient.pushChanges({
      path,
      username,
      password,
    })
  }

  /**
   * Pulls changes from the remote repository.
   *
   * @param {string} path - Path to the Git repository root. Relative paths are resolved based on the sandbox working directory.
   * @param {string} [username] - Git username for authentication
   * @param {string} [password] - Git password or token for authentication
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
   */
  @WithInstrumentation()
  public async pull(path: string, username?: string, password?: string): Promise<void> {
    await this.apiClient.pullChanges({
      path,
      username,
      password,
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
}
