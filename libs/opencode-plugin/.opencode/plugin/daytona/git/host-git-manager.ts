import { logger } from '../core/logger'
import { execSync } from 'child_process'

function execSyncSilent(cmd: string, options: any = {}) {
  return execSync(cmd, { stdio: 'ignore', ...options })
}

export class HostGitManager {
  // No constructor needed; use global logger
  private operationQueue: Promise<void> = Promise.resolve()

  /**
   * Checks if a git repository exists in the current directory
   * @returns true if a git repo exists, false otherwise
   */
  hasRepo(): boolean {
    try {
      execSyncSilent('git rev-parse --is-inside-work-tree')
      return true
    } catch {
      return false
    }
  }

  /**
   * Pushes local changes to the sandbox remote.
   * @param remoteName Numbered remote (e.g. sandbox-2) matching opencode/N.
   * @param sshUrl The SSH URL of the sandbox remote.
   * @param branch The branch to push to.
   * @returns true if push was successful, false if no repo exists
   */
  async pushLocalToSandboxRemote(remoteName: string, sshUrl: string, branch: string): Promise<boolean> {
    if (!this.hasRepo()) {
      logger.warn('No local git repository found. Skipping push to sandbox.')
      return false
    }
    try {
      logger.info(`Pushing to ${remoteName} (${sshUrl}) on branch ${branch}`)
      const operation = this.operationQueue.then(async () => {
        execSyncSilent('git add .')
        execSyncSilent('git commit -m "Sync local changes before agent start" || echo "No changes to commit"', {
          shell: '/bin/bash',
        })
        this.setRemote(remoteName, sshUrl)
        let attempts = 0
        while (attempts < 3) {
          try {
            execSyncSilent(`git push ${remoteName} HEAD:${branch}`)
            logger.info(`✓ Pushed local changes to ${remoteName}`)
            return
          } catch (e) {
            attempts++
            if (attempts >= 3) {
              logger.error(`Error pushing to ${remoteName} after 3 attempts: ${e}`)
            } else {
              logger.warn(`Push attempt ${attempts} failed, retrying...`)
            }
          }
        }
      })
      this.operationQueue = operation
      await operation
      return true
    } catch (e) {
      logger.error(`Error pushing to sandbox: ${e}`)
      return false
    }
  }

  private setRemote(remoteName: string, sshUrl: string): void {
    try {
      // remove existing remote if it exists
      execSyncSilent(`git remote remove ${remoteName} || true`)
      execSyncSilent(`git remote add ${remoteName} ${sshUrl}`)
    } catch (e) {
      logger.warn(`Could not set sandbox remote: ${e}`)
    }
  }

  async pull(remoteName: string, sshUrl: string, branch: string, localBranch?: string): Promise<void> {
    const operation = this.operationQueue.then(async () => {
      this.setRemote(remoteName, sshUrl)
      let attempts = 0
      // The first pull attempt sometimes fails. I'm not sure what the cause is.
      while (attempts < 3) {
        try {
          if (localBranch) {
            // Fetch the remote branch into the specified local branch
            execSyncSilent(`git fetch ${remoteName} ${branch}:${localBranch}`)
            logger.info(`✓ Fetched latest changes from sandbox into ${localBranch}`)
          } else {
            execSyncSilent(`git pull ${remoteName} ${branch}`)
            logger.info('✓ Pulled latest changes from sandbox')
          }
          return
        } catch (e) {
          attempts++
          if (attempts >= 3) {
            logger.error(`Error pulling from sandbox after 3 attempts: ${e}`)
          } else {
            logger.warn(`Pull attempt ${attempts} failed, retrying...`)
          }
        }
      }
    })
    this.operationQueue = operation
    await operation
  }

  async push(remoteName: string, sshUrl: string, branch: string): Promise<void> {
    const operation = this.operationQueue.then(async () => {
      this.setRemote(remoteName, sshUrl)
      let attempts = 0
      while (attempts < 3) {
        try {
          execSyncSilent(`git push ${remoteName} HEAD:${branch}`)
          logger.info('✓ Pushed changes to sandbox')
          return
        } catch (e) {
          attempts++
          if (attempts >= 3) {
            logger.error(`Error pushing to sandbox after 3 attempts: ${e}`)
          } else {
            logger.warn(`Push attempt ${attempts} failed, retrying...`)
          }
        }
      }
    })
    this.operationQueue = operation
    await operation
  }
}
