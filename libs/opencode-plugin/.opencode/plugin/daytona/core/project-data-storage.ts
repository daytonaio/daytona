/**
 * Handles file storage operations for project session data
 * Stores data per-project in ~/.local/share/opencode/storage/daytona/{projectId}.json
 */

import { existsSync, readFileSync, writeFileSync, mkdirSync, readdirSync } from 'fs'
import { join } from 'path'
import { logger } from './logger'
import type { ProjectSessionData, SessionInfo } from './types'

export class ProjectDataStorage {
  private readonly storageDir: string

  constructor(storageDir: string) {
    this.storageDir = storageDir

    // Ensure storage directory exists
    if (!existsSync(this.storageDir)) {
      mkdirSync(this.storageDir, { recursive: true })
    }
  }

  /**
   * Get the file path for a project's session data
   */
  private getProjectFilePath(projectId: string): string {
    return join(this.storageDir, `${projectId}.json`)
  }

  /**
   * List known project IDs from storage.
   */
  private listProjectIds(): string[] {
    try {
      return readdirSync(this.storageDir)
        .filter(name => name.endsWith('.json'))
        .map(name => name.slice(0, -'.json'.length))
    } catch (err) {
      logger.error(`Failed to list project data files: ${err}`)
      return []
    }
  }

  /**
   * Load project session data from disk
   */
  load(projectId: string): ProjectSessionData | null {
    const filePath = this.getProjectFilePath(projectId)
    try {
      if (existsSync(filePath)) {
        return JSON.parse(readFileSync(filePath, 'utf-8')) as ProjectSessionData
      }
    } catch (err) {
      logger.error(`Failed to load project data for ${projectId}: ${err}`)
    }
    return null
  }

  /**
   * Get a session for a project. If not found in the requested project, search all other
   * projects on disk and, if found, migrate it into the requested project.
   */
  getSession(projectId: string, worktree: string, sessionId: string): SessionInfo | undefined {
    const current = this.load(projectId)
    const currentSession = current?.sessions?.[sessionId]
    if (currentSession) {
      return currentSession
    }

    // Look in other projects and migrate if found.
    for (const otherProjectId of this.listProjectIds()) {
      if (otherProjectId === projectId) continue

      const otherData = this.load(otherProjectId)
      const found = otherData?.sessions?.[sessionId]
      if (!found) continue

      const destination: ProjectSessionData =
        current ?? {
          projectId,
          worktree,
          lastBranchNumber: 0,
          sessions: {},
        }

      // Remove from source first (best effort).
      try {
        delete otherData!.sessions[sessionId]
        this.save(otherProjectId, otherData!.worktree, otherData!.sessions, otherData!.lastBranchNumber)
      } catch (err) {
        logger.warn(`Failed to remove session ${sessionId} from project ${otherProjectId}: ${err}`)
      }

      // Add to destination and persist.
      destination.sessions[sessionId] = found
      if (found.branchNumber !== undefined) {
        destination.lastBranchNumber = Math.max(destination.lastBranchNumber ?? 0, found.branchNumber)
      }
      // Prefer the worktree for the project we're actually operating on.
      destination.worktree = worktree
      this.save(projectId, destination.worktree, destination.sessions, destination.lastBranchNumber)

      logger.info(`Migrated session ${sessionId} from project ${otherProjectId} to project ${projectId}`)
      return found
    }

    return undefined
  }

  /**
   * Save project session data to disk
   */
  save(
    projectId: string,
    worktree: string,
    sessions: Record<string, SessionInfo>,
    lastBranchNumber?: number,
  ): void {
    const filePath = this.getProjectFilePath(projectId)
    const projectData: ProjectSessionData = {
      projectId,
      worktree,
      lastBranchNumber,
      sessions,
    }

    try {
      writeFileSync(filePath, JSON.stringify(projectData, null, 2))
      logger.info(`Saved project data for ${projectId}`)
    } catch (err) {
      logger.error(`Failed to save project data for ${projectId}: ${err}`)
    }
  }

  /**
   * Get the next available branch number for a project
   */
  getNextBranchNumber(projectId: string): number {
    const projectData = this.load(projectId)
    if (!projectData) {
      return 1
    }

    const branchNumbers = Object.values(projectData.sessions)
      .map(s => s.branchNumber)
      .filter((n): n is number => n !== undefined)

    // Use a persisted monotonic pointer so we never reuse deleted numbers.
    const pointer = projectData.lastBranchNumber ?? 0
    const maxInSessions = branchNumbers.length > 0 ? Math.max(...branchNumbers) : 0
    return Math.max(pointer, maxInSessions) + 1
  }

  /**
   * Get branch number for a sandbox
   */
  getBranchNumberForSandbox(projectId: string, sandboxId: string): number | undefined {
    const projectData = this.load(projectId)
    if (!projectData) {
      return undefined
    }
    const session = Object.values(projectData.sessions).find(s => s.sandboxId === sandboxId)
    return session?.branchNumber
  }

  /**
   * Update a single session in the project file
   */
  updateSession(projectId: string, worktree: string, sessionId: string, sandboxId: string, branchNumber?: number): void {
    const projectData = this.load(projectId) || {
      projectId,
      worktree,
      lastBranchNumber: 0,
      sessions: {},
    }

    const now = Date.now()
    if (!projectData.sessions[sessionId]) {
      // Assign branch number if not provided
      const assignedBranchNumber = branchNumber ?? this.getNextBranchNumber(projectId)
      projectData.sessions[sessionId] = {
        sandboxId,
        branchNumber: assignedBranchNumber,
        created: now,
        lastAccessed: now,
      }
      projectData.lastBranchNumber = Math.max(projectData.lastBranchNumber ?? 0, assignedBranchNumber)
    } else {
      projectData.sessions[sessionId].sandboxId = sandboxId
      projectData.sessions[sessionId].lastAccessed = now
      // Only update branch number if it wasn't set before
      if (projectData.sessions[sessionId].branchNumber === undefined) {
        const assignedBranchNumber = branchNumber ?? this.getNextBranchNumber(projectId)
        projectData.sessions[sessionId].branchNumber = assignedBranchNumber
        projectData.lastBranchNumber = Math.max(projectData.lastBranchNumber ?? 0, assignedBranchNumber)
      }
    }

    this.save(projectId, worktree, projectData.sessions, projectData.lastBranchNumber)
  }

  /**
   * Remove a session from the project file
   */
  removeSession(projectId: string, worktree: string, sessionId: string): void {
    const projectData = this.load(projectId)
    if (projectData && projectData.sessions[sessionId]) {
      delete projectData.sessions[sessionId]
      // Intentionally keep lastBranchNumber so branch numbering remains monotonic
      this.save(projectId, worktree, projectData.sessions, projectData.lastBranchNumber)
    }
  }
}
