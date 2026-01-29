/**
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { Sandbox } from '@daytonaio/sdk'
import { logger } from '../core/logger'

export class DaytonaSandboxGitManager {
  constructor(
    private readonly sandbox: Sandbox,
    private readonly repoPath: string,
  ) {}

  async ensureDirectory(): Promise<void> {
    await this.sandbox.fs.createFolder(this.repoPath, '755')
  }

  async ensureRepo(): Promise<void> {
    await this.ensureDirectory()
    const isGit = await this.sandbox.process.executeCommand('git rev-parse --is-inside-work-tree', this.repoPath)
    if (!isGit || isGit.result.trim() !== 'true') {
      await this.sandbox.process.executeCommand('git init', this.repoPath)
      await this.sandbox.process.executeCommand('git config user.email "sandbox@example.com"', this.repoPath)
      await this.sandbox.process.executeCommand('git config user.name "Daytona Sandbox"', this.repoPath)
      logger.info(`Initialized git repo in sandbox at ${this.repoPath}`)
    }
  }

  async autoCommit(): Promise<boolean> {
    try {
      // Check if there are any changes to commit
      const statusResult = await this.sandbox.process.executeCommand('git status --porcelain', this.repoPath)
      if (!statusResult.result.trim()) {
        logger.info(`No changes to commit in sandbox at ${this.repoPath}`)
        return false
      }
      await this.sandbox.process.executeCommand('git add .', this.repoPath)
      await this.sandbox.process.executeCommand(
        'git commit -am "Auto-commit from Daytona plugin"',
        this.repoPath,
      )
      logger.info(`Auto-committed changes in sandbox at ${this.repoPath}`)
      return true
    } catch (err) {
      logger.error(`Failed to auto-commit in sandbox at ${this.repoPath}: ${err}`)
      return false
    }
  }

  async resetToRemote(branch: string): Promise<void> {
    try {
      const result = await this.sandbox.process.executeCommand(`git checkout -B ${branch}`, this.repoPath)
      logger.info(`Checked out branch '${branch}': ${result.result}`)
      await this.sandbox.process.executeCommand('git reset --hard', this.repoPath)
      await this.sandbox.process.executeCommand('git clean -fd', this.repoPath)
      logger.info('Reset sandbox worktree to pushed state.')
      const statusResult = await this.sandbox.process.executeCommand('git status --porcelain', this.repoPath)
      if (statusResult.result.trim()) {
        logger.warn(`Sandbox has uncommitted changes after reset:\n${statusResult.result}`)
      } else {
        logger.info('No uncommitted changes in sandbox after reset.')
      }
    } catch (err) {
      logger.error(`Failed to reset sandbox worktree: ${err}`)
    }
  }
}
