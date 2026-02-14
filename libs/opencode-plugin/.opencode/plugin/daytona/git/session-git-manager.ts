/**
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { Sandbox } from '@daytonaio/sdk'
import { logger } from '../core/logger'
import { toast } from '../core/toast'
import { DaytonaSandboxGitManager } from './sandbox-git-manager'
import { HostGitManager } from './host-git-manager'
import type { PluginInput } from '@opencode-ai/plugin'

/**
 * SessionGitManager: Combines DaytonaSandboxGitManager and HostGitManager for session lifecycle git operations.
 */
export class SessionGitManager {
  private readonly sandboxGit: DaytonaSandboxGitManager
  private readonly hostGit: HostGitManager
  private readonly sandbox: Sandbox
  private readonly repoPath: string
  private readonly worktree: string
  private readonly branch: string
  private readonly localBranch: string
  /** Numbered remote (sandbox-2) matches localBranch (opencode/2) */
  private readonly remoteName: string

  constructor(sandbox: Sandbox, repoPath: string, worktree: string, branchNumber: number) {
    this.sandbox = sandbox
    this.repoPath = repoPath
    this.worktree = worktree
    this.branch = 'opencode'
    this.localBranch = `opencode/${branchNumber}`
    this.remoteName = `sandbox-${branchNumber}`
    this.sandboxGit = new DaytonaSandboxGitManager(sandbox, repoPath)
    this.hostGit = new HostGitManager()
  }

  /**
   * Allocate and reserve the next opencode/N number in the local repo at `worktree`.
   * This keeps all host-git concerns inside the git manager layer.
   */
  static allocateAndReserveBranchNumber(worktree: string, prefix = 'opencode'): number {
    return new HostGitManager().allocateAndReserveBranchNumber(worktree, prefix)
  }

  private async getSshUrl(): Promise<string> {
    const sshAccess = await this.sandbox.createSshAccess(10)
    return `ssh://${sshAccess.token}@ssh.app.daytona.io${this.repoPath}`
  }

  /**
   * Check if local git repository exists
   * @returns true if repo exists, false otherwise
   */
  hasLocalRepo(): boolean {
    return this.hostGit.hasRepo(this.worktree)
  }

  /**
   * Initialize git in the sandbox and sync with host
   * Used when a new sandbox is created for a session
   */
  async initializeAndSync(pluginCtx?: PluginInput) {
    if (pluginCtx?.client?.tui) {
      toast.initialize(pluginCtx.client.tui)
    }
    try {
      // Check if local git repo exists before initializing sandbox repo
      if (!this.hostGit.hasRepo(this.worktree)) {
        // Always ensure the directory exists, even if git syncing is disabled
        await this.sandboxGit.ensureDirectory()
        logger.warn('No local git repository found. Git syncing is disabled.')
        toast.show({
          title: 'Git syncing disabled',
          message: 'No local git repository found. Git syncing is disabled for this session.',
          variant: 'warning',
        })
        return
      }

      await this.sandboxGit.ensureRepo()
      const sshUrl = await this.getSshUrl()
      const pushed = await this.hostGit.pushLocalToSandboxRemote(this.remoteName, sshUrl, this.branch, this.worktree)
      if (pushed) {
        await this.sandboxGit.resetToRemote(this.branch)
      }
    } catch (err: any) {
      toast.show({
        title: 'Git sync error',
        message: err?.message || 'Failed to sync git repo.',
        variant: 'error',
      })
      throw err
    }
  }

  /**
   * Auto-commit in the sandbox and pull latest from host
   * Used on session idle
   * Returns true if changes were synced, false if no changes or no local repo
   */
  async autoCommitAndPull(pluginCtx?: PluginInput): Promise<boolean> {
    if (pluginCtx?.client?.tui) {
      toast.initialize(pluginCtx.client.tui)
    }
    try {
      // Check if local git repo exists before attempting any git operations
      if (!this.hostGit.hasRepo(this.worktree)) {
        logger.warn('No local git repository found. Git syncing is disabled.')
        return false
      }

      await this.sandboxGit.ensureRepo()
      const hasChanges = await this.sandboxGit.autoCommit()

      // Only sync and notify if there were actual changes
      if (!hasChanges) {
        return false
      }

      const sshUrl = await this.getSshUrl()

      await this.hostGit.pull(this.remoteName, sshUrl, this.branch, this.worktree, this.localBranch)
      toast.show({
        title: 'Changes synced',
        message: `Changes have been synced to ${this.localBranch} in your local repository`,
        variant: 'success',
      })
      return true
    } catch (err: any) {
      toast.show({
        title: 'Sync failed',
        message: err?.message || 'Failed to auto-commit and pull.',
        variant: 'error',
      })
      logger.error(`[idle/git] error sandboxId=${this.sandbox.id}: ${err}`)
      throw err
    }
  }
}
