/**
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { Sandbox } from '@daytona/sdk'
import { logger } from '../core/logger'

export class DaytonaSandboxGitManager {
  constructor(
    private readonly sandbox: Sandbox,
    private readonly repoPath: string,
  ) {}

  async ensureDirectory(): Promise<void> {
    await this.sandbox.fs.createFolder(this.repoPath, '755')
  }

  private async runGitCommand(command: string): Promise<string> {
    const result = await this.sandbox.process.executeCommand(command, this.repoPath)
    const output = result.result.trim()

    if (result.exitCode !== undefined && result.exitCode !== 0) {
      const installHint = output.includes('git: command not found')
        ? '\nHint: git is not installed in this sandbox. Rebuild the sandbox image with git installed.'
        : ''
      throw new Error(`Git command failed: ${command}\nOutput: ${output || '(no output)'}${installHint}`)
    }

    return result.result
  }

  async ensureRepo(): Promise<void> {
    await this.ensureDirectory()
    const isGit = await this.runGitCommand(
      'if [ -e .git ]; then git rev-parse --is-inside-work-tree; else echo false; fi',
    )
    if (isGit.trim() !== 'true') {
      await this.runGitCommand('git init')
      await this.runGitCommand('git config user.email "sandbox@example.com"')
      await this.runGitCommand('git config user.name "Daytona Sandbox"')
      logger.info(`Initialized git repo in sandbox at ${this.repoPath}`)
    }
  }

  async autoCommit(): Promise<boolean> {
    // Check if there are any changes to commit
    const status = await this.runGitCommand('git status --porcelain')
    if (!status.trim()) {
      logger.info(`No changes to commit in sandbox at ${this.repoPath}`)
      return false
    }
    await this.runGitCommand('git add .')
    await this.runGitCommand('git commit -am "Auto-commit from Daytona plugin"')
    logger.info(`Auto-committed changes in sandbox at ${this.repoPath}`)
    return true
  }

  async resetToRemote(branch: string): Promise<void> {
    const checkout = await this.runGitCommand(`git checkout -B ${branch}`)
    logger.info(`Checked out branch '${branch}': ${checkout}`)
    await this.runGitCommand('git reset --hard')
    await this.runGitCommand('git clean -fd')
    logger.info('Reset sandbox worktree to pushed state.')
    const status = await this.runGitCommand('git status --porcelain')
    if (status.trim()) {
      logger.warn(`Sandbox has uncommitted changes after reset:\n${status}`)
    } else {
      logger.info('No uncommitted changes in sandbox after reset.')
    }
  }
}
