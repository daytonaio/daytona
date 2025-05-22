/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ToolboxApi, ListBranchResponse, GitStatus } from '@daytonaio/api-client'
import { Sandbox, SandboxInstance } from './Sandbox'
import { prefixRelativePath } from './utils/Path'

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
  constructor(
    private readonly sandbox: Sandbox,
    private readonly toolboxApi: ToolboxApi,
    private readonly instance: SandboxInstance,
    private readonly getRootDir: () => Promise<string>,
  ) {}

  /**
   * Stages the specified files for the next commit, similar to
   * running 'git add' on the command line.
   *
   * @param {string} path - Path to the Git repository root. Relative paths are resolved based on the user's
   * root directory.
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
  public async add(path: string, files: string[]): Promise<void> {
    await this.toolboxApi.gitAddFiles(this.instance.id, {
      path: prefixRelativePath(await this.getRootDir(), path),
      files,
    })
  }

  /**
   * List branches in the repository.
   *
   * @param {string} path - Path to the Git repository root. Relative paths are resolved based on the user's
   * root directory.
   * @returns {Promise<ListBranchResponse>} List of branches in the repository
   *
   * @example
   * const response = await git.branches('workspace/repo');
   * console.log(`Branches: ${response.branches}`);
   */
  public async branches(path: string): Promise<ListBranchResponse> {
    const response = await this.toolboxApi.gitListBranches(
      this.instance.id,
      prefixRelativePath(await this.getRootDir(), path),
    )
    return response.data
  }

  /**
   * Clones a Git repository into the specified path. It supports
   * cloning specific branches or commits, and can authenticate with the remote
   * repository if credentials are provided.
   *
   * @param {string} url - Repository URL to clone from
   * @param {string} path - Path where the repository should be cloned. Relative paths are resolved based on the user's
   * root directory.
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
  public async clone(
    url: string,
    path: string,
    branch?: string,
    commitId?: string,
    username?: string,
    password?: string,
  ): Promise<void> {
    await this.toolboxApi.gitCloneRepository(this.instance.id, {
      url: url,
      branch: branch,
      path: prefixRelativePath(await this.getRootDir(), path),
      username,
      password,
      commit_id: commitId,
    })
  }

  /**
   * Commits staged changes.
   *
   * @param {string} path - Path to the Git repository root. Relative paths are resolved based on the user's
   * root directory.
   * @param {string} message - Commit message describing the changes
   * @param {string} author - Name of the commit author
   * @param {string} email - Email address of the commit author
   * @returns {Promise<void>}
   *
   * @example
   * // Stage and commit changes
   * await git.add('workspace/repo', ['README.md']);
   * await git.commit(
   *   'workspace/repo',
   *   'Update documentation',
   *   'John Doe',
   *   'john@example.com'
   * );
   */
  public async commit(path: string, message: string, author: string, email: string): Promise<GitCommitResponse> {
    const response = await this.toolboxApi.gitCommitChanges(this.instance.id, {
      path: prefixRelativePath(await this.getRootDir(), path),
      message,
      author,
      email,
    })
    return {
      sha: response.data.hash,
    }
  }

  /**
   * Push local changes to the remote repository.
   *
   * @param {string} path - Path to the Git repository root. Relative paths are resolved based on the user's
   * root directory.
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
  public async push(path: string, username?: string, password?: string): Promise<void> {
    await this.toolboxApi.gitPushChanges(this.instance.id, {
      path: prefixRelativePath(await this.getRootDir(), path),
      username,
      password,
    })
  }

  /**
   * Pulls changes from the remote repository.
   *
   * @param {string} path - Path to the Git repository root. Relative paths are resolved based on the user's
   * root directory.
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
  public async pull(path: string, username?: string, password?: string): Promise<void> {
    await this.toolboxApi.gitPullChanges(this.instance.id, {
      path: prefixRelativePath(await this.getRootDir(), path),
      username,
      password,
    })
  }

  /**
   * Gets the current status of the Git repository.
   *
   * @param {string} path - Path to the Git repository root. Relative paths are resolved based on the user's
   * root directory.
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
  public async status(path: string): Promise<GitStatus> {
    const response = await this.toolboxApi.gitGetStatus(
      this.instance.id,
      prefixRelativePath(await this.getRootDir(), path),
    )
    return response.data
  }
}
