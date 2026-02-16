/**
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { logger } from '../core/logger'
import { execSync, spawnSync } from 'child_process'

type ExecResult = {
  ok: boolean
  stdout: string
  stderr: string
  status: number | null
}

function execCommand(cmd: string, options: any = {}): ExecResult {
  try {
    const stdout = execSync(cmd, {
      stdio: ['ignore', 'pipe', 'pipe'],
      encoding: 'utf8',
      ...options,
    }) as unknown as string
    return { ok: true, stdout: stdout ?? '', stderr: '', status: 0 }
  } catch (err: any) {
    const stdout = err?.stdout?.toString?.() ?? ''
    const stderr = err?.stderr?.toString?.() ?? err?.message ?? String(err)
    const status = typeof err?.status === 'number' ? err.status : null
    return { ok: false, stdout, stderr, status }
  }
}

export class HostGitManager {
  // No constructor needed; use global logger
  private operationQueue: Promise<void> = Promise.resolve()
  /** Cached OID of an empty commit used to reserve branch refs (branches must point at commits, not blobs). */
  private emptyCommitOidCache = new Map<string, string>()

  /**
   * Checks if a git repository exists in the current directory
   * @returns true if a git repo exists, false otherwise
   */
  hasRepo(cwd?: string): boolean {
    return execCommand('git rev-parse --is-inside-work-tree', cwd ? { cwd } : {}).ok
  }

  /**
   * Allocates the next available opencode/N branch number by scanning local refs and
   * reserving the chosen number by creating the ref immediately.
   *
   * This avoids relying on OpenCode's project ID and works even in repos with no commits.
   */
  allocateAndReserveBranchNumber(cwd: string, prefix = 'opencode'): number {
    const start = Date.now()
    if (!this.hasRepo(cwd)) {
      throw new Error('No local git repository found.')
    }

    const base = `refs/heads/${prefix}/`
    const listRes = execCommand(`git for-each-ref --format='%(refname:strip=3)' ${base}`, { cwd })
    if (!listRes.ok) throw new Error(listRes.stderr)
    const list = listRes.stdout.trim()
    const nums =
      list.length === 0
        ? []
        : list
            .split('\n')
            .map((s) => s.trim())
            .filter(Boolean)
            .map((s) => Number.parseInt(s, 10))
            .filter((n) => Number.isFinite(n) && n > 0)

    let n = (nums.length ? Math.max(...nums) : 0) + 1
    const maxAttempts = 50 // Circuit-breaker
    let attempts = 0
    while (n < 1_000_000 && attempts < maxAttempts) {
      attempts++
      const ref = `${base}${n}`
      if (this.refExists(cwd, ref)) {
        n++
        continue
      }
      const oid = this.getOrCreateEmptyCommitOid(cwd)
      const result = execCommand(`git update-ref "${ref}" "${oid}"`, { cwd })
      if (result.ok) {
        logger.info(`[branch-alloc] reserved ${prefix}/${n} in ${Date.now() - start}ms`)
        return n
      } else {
        // If we raced or hit an edge case, try the next number.
        n++
      }
    }
    const oid = this.getOrCreateEmptyCommitOid(cwd)
    const last = execCommand(`git update-ref "${base}${n}" "${oid}"`, { cwd })
    throw new Error(`Failed to allocate branch number after ${attempts} attempts. Last error: ${last.stderr}`)
  }

  private refExists(cwd: string, ref: string): boolean {
    return execCommand(`git show-ref --verify --quiet "${ref}"`, { cwd }).ok
  }

  /**
   * Returns a commit OID that branch refs can point at. Uses HEAD if the repo has commits,
   * otherwise creates and caches an empty commit (empty tree + commit). Branch refs must
   * point at commits, not blobs.
   */
  private getOrCreateEmptyCommitOid(cwd: string): string {
    const cached = this.emptyCommitOidCache.get(cwd)
    if (cached) return cached
    const headRes = execCommand('git rev-parse HEAD', { cwd })
    const head = headRes.ok ? headRes.stdout.trim() : ''
    if (head) {
      this.emptyCommitOidCache.set(cwd, head)
      return head
    }

    // Create an empty tree (idempotent) then a commit pointing at it.
    // Branch refs must point at commits, so we can't reserve with a blob.
    const treeResult = spawnSync('git', ['hash-object', '-t', 'tree', '-w', '--stdin'], {
      // Empty stdin => empty tree
      input: '',
      cwd,
      encoding: 'utf8',
    })
    const treeOid = treeResult.stdout?.trim()
    if (treeResult.status !== 0 || !treeOid) {
      const errorMsg =
        treeResult.stderr?.toString() || treeResult.error?.message || String(treeResult.error || 'unknown')
      throw new Error(`Failed to create empty tree: ${errorMsg}`)
    }

    // Provide a default identity for reservation commits when repo has no user.name/user.email (e.g. CI).
    const reservationCommitName = 'OpenCode Plugin'
    const reservationCommitEmail = 'opencode@daytona.io'
    const reservationCommitMessage = 'OpenCode reservation'
    const commitEnv = {
      ...process.env,
      GIT_AUTHOR_NAME: reservationCommitName,
      GIT_AUTHOR_EMAIL: reservationCommitEmail,
      GIT_COMMITTER_NAME: reservationCommitName,
      GIT_COMMITTER_EMAIL: reservationCommitEmail,
    }

    // Create the commit
    const commitRes = execCommand(`git commit-tree ${treeOid} -m "${reservationCommitMessage}"`, {
      cwd,
      env: commitEnv,
    })
    if (!commitRes.ok) throw new Error(`Failed to create empty commit: ${commitRes.stderr}`)
    const commitOid = commitRes.stdout.trim()
    this.emptyCommitOidCache.set(cwd, commitOid)
    return commitOid
  }

  /**
   * Pushes local changes to the sandbox remote.
   * @param remoteName Numbered remote (e.g. sandbox-2) matching opencode/N.
   * @param sshUrl The SSH URL of the sandbox remote.
   * @param branch The branch to push to.
   * @param cwd Worktree path to run git in.
   * @returns true if push was successful, false if no repo exists
   */
  async pushLocalToSandboxRemote(remoteName: string, sshUrl: string, branch: string, cwd: string): Promise<boolean> {
    if (!this.hasRepo(cwd)) {
      logger.warn('No local git repository found. Skipping push to sandbox.')
      return false
    }
    try {
      logger.info(`Pushing to ${remoteName} (${sshUrl}) on branch ${branch}`)
      const operation = this.operationQueue.then(async () => {
        const statusRes = execCommand('git status --porcelain', { cwd })
        if (!statusRes.ok) {
          throw new Error(statusRes.stderr)
        }
        if (statusRes.stdout.trim().length > 0) {
          logger.warn('Local repository has uncommitted changes; pushing HEAD only (no auto-commit).')
        }

        this.setRemote(remoteName, sshUrl, cwd)
        let attempts = 0
        while (attempts < 3) {
          try {
            const pushRes = execCommand(`git push ${remoteName} HEAD:${branch}`, { cwd })
            if (!pushRes.ok) throw new Error(pushRes.stderr)
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

  private setRemote(remoteName: string, sshUrl: string, cwd: string): void {
    try {
      // remove existing remote if it exists
      execCommand(`git remote remove ${remoteName}`, { cwd })
      execCommand(`git remote add ${remoteName} ${sshUrl}`, { cwd })
    } catch (e) {
      logger.warn(`Could not set sandbox remote: ${e}`)
    }
  }

  async pull(remoteName: string, sshUrl: string, branch: string, cwd: string, localBranch?: string): Promise<void> {
    const operation = this.operationQueue.then(async () => {
      this.setRemote(remoteName, sshUrl, cwd)
      let attempts = 0
      let lastError: unknown = undefined
      // The first pull attempt sometimes fails. I'm not sure what the cause is.
      while (attempts < 3) {
        try {
          if (localBranch) {
            // Fetch into FETCH_HEAD only (never into refs/heads) so we don't hit
            // "refusing to fetch into branch checked out" when this branch is checked out.
            const fetchRes = execCommand(`git fetch ${remoteName} ${branch}`, { cwd })
            if (!fetchRes.ok) throw new Error(fetchRes.stderr)

            const updateRefRes = execCommand(`git update-ref refs/heads/${localBranch} FETCH_HEAD`, { cwd })
            if (!updateRefRes.ok) throw new Error(updateRefRes.stderr)

            // Only reset working directory if we're currently on this branch
            const currentBranchRes = execCommand(`git rev-parse --abbrev-ref HEAD`, { cwd })
            const currentBranch = currentBranchRes.ok ? currentBranchRes.stdout.trim() : ''
            if (currentBranch === localBranch) {
              const resetRes = execCommand(`git reset --hard refs/heads/${localBranch}`, { cwd })
              if (!resetRes.ok) throw new Error(resetRes.stderr)
            }

            logger.info(`✓ Force pulled latest changes from sandbox into ${localBranch}`)
          } else {
            const pullRes = execCommand(`git pull ${remoteName} ${branch}`, { cwd })
            if (!pullRes.ok) throw new Error(pullRes.stderr)
            logger.info('✓ Pulled latest changes from sandbox')
          }
          return
        } catch (e) {
          lastError = e
          attempts++
          if (attempts >= 3) {
            logger.error(`Error pulling from sandbox after 3 attempts: ${e}`)
          } else {
            logger.warn(`Pull attempt ${attempts} failed, retrying...`)
          }
        }
      }

      // If we got here, all attempts failed.
      throw lastError ?? new Error('Pull failed after 3 attempts')
    })
    this.operationQueue = operation
    await operation
  }

  async push(remoteName: string, sshUrl: string, branch: string, cwd: string): Promise<void> {
    const operation = this.operationQueue.then(async () => {
      this.setRemote(remoteName, sshUrl, cwd)
      let attempts = 0
      while (attempts < 3) {
        try {
          const pushRes = execCommand(`git push ${remoteName} HEAD:${branch}`, { cwd })
          if (!pushRes.ok) throw new Error(pushRes.stderr)
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
