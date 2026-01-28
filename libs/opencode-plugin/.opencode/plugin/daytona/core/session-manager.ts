/**
 * Manages Daytona sandbox sessions and persists session-sandbox mappings
 * Stores data per-project in ~/.local/share/opencode/storage/daytona/{projectId}.json
 */

import { Daytona, type Sandbox } from '@daytonaio/sdk'
import { logger } from './logger'
import type { SessionSandboxMap, SandboxInfo } from './types'
import { SessionGitManager } from '../git/session-git-manager'
import { ProjectDataStorage } from './project-data-storage'
import type { PluginInput } from '@opencode-ai/plugin'
import { toast } from './toast'

export class DaytonaSessionManager {
  private readonly apiKey: string
  private readonly dataStorage: ProjectDataStorage
  private sessionSandboxes: SessionSandboxMap
  private currentProjectId?: string
  public readonly repoPath: string

  constructor(apiKey: string, storageDir: string, repoPath: string) {
    this.apiKey = apiKey
    this.dataStorage = new ProjectDataStorage(storageDir)
    this.repoPath = repoPath
    this.sessionSandboxes = new Map()
  }

  /**
   * Check if a sandbox is fully initialized (has process property)
   */
  private isFullyInitialized(sandbox: Sandbox | SandboxInfo | undefined): sandbox is Sandbox {
    return sandbox !== undefined && 'process' in sandbox
  }

  /**
   * Check if a sandbox is partially initialized (has id but not process)
   */
  private isPartiallyInitialized(sandbox: Sandbox | SandboxInfo | undefined): sandbox is SandboxInfo {
    return sandbox !== undefined && 'id' in sandbox && !('process' in sandbox)
  }

  /**
   * Load sessions for a specific project into memory
   */
  private loadProjectSessions(projectId: string): void {
    const projectData = this.dataStorage.load(projectId)
    if (projectData) {
      for (const [sessionId, sessionInfo] of Object.entries(projectData.sessions)) {
        this.sessionSandboxes.set(sessionId, { id: sessionInfo.sandboxId })
      }
      logger.info(`Loaded ${Object.keys(projectData.sessions).length} sessions for project ${projectId}`)
    }
  }

  /**
   * Set the current project context
   */
  setProjectContext(projectId: string): void {
    if (this.currentProjectId !== projectId) {
      this.currentProjectId = projectId
      this.loadProjectSessions(projectId)
    }
  }

  /**
   * Get branch number for a sandbox
   */
  getBranchNumberForSandbox(projectId: string, sandboxId: string): number | undefined {
    return this.dataStorage.getBranchNumberForSandbox(projectId, sandboxId)
  }

  /**
   * Get or create a sandbox for the given session ID
   */
  async getSandbox(sessionId: string, projectId: string, worktree: string, pluginCtx?: PluginInput): Promise<Sandbox> {
    if (pluginCtx?.client?.tui) {
      toast.initialize(pluginCtx.client.tui)
    }
    if (!this.apiKey) {
      logger.error('DAYTONA_API_KEY is not set. Cannot create or retrieve sandbox.')
      toast.show({
        title: 'Sandbox error',
        message: 'DAYTONA_API_KEY is not set. Please set the environment variable to use Daytona sandboxes.',
        variant: 'error',
      })
      throw new Error('DAYTONA_API_KEY is not set. Please set the environment variable to use Daytona sandboxes.')
    }

    // Load project sessions if needed
    this.setProjectContext(projectId)

    const existing = this.sessionSandboxes.get(sessionId)

    // If we have a fully initialized sandbox, reuse it
    if (this.isFullyInitialized(existing)) {
      // Refresh sandbox state and ensure it's running
      await existing.refreshData()
      if (existing.state !== 'started') {
        logger.info(`Starting sandbox ${existing.id} (current state: ${existing.state})`)
        await existing.start()
      }
      this.dataStorage.updateSession(projectId, worktree, sessionId, existing.id)
      return existing
    }

    // If we have a sandboxId but not a full sandbox object, reconnect to it
    if (this.isPartiallyInitialized(existing)) {
      logger.info(`Reconnecting to existing sandbox: ${existing.id}`)
      const daytona = new Daytona({ apiKey: this.apiKey })
      const sandbox = await daytona.get(existing.id)
      await sandbox.start()
      this.sessionSandboxes.set(sessionId, sandbox)
      // Preserve branch number if it exists for this sandbox
      let branchNumber = this.dataStorage.getBranchNumberForSandbox(projectId, sandbox.id)
      if (!branchNumber) {
        branchNumber = this.dataStorage.getNextBranchNumber(projectId)
      }
      this.dataStorage.updateSession(projectId, worktree, sessionId, sandbox.id, branchNumber)
      toast.show({
        title: 'Sandbox connected',
        message: `Connected to existing sandbox.`,
        variant: 'info',
      })
      return sandbox
    }

    // If not in cache/storage for this project, try to recover from other projects and migrate.
    if (!existing) {
      const migrated = this.dataStorage.getSession(projectId, worktree, sessionId)
      if (migrated?.sandboxId) {
        logger.info(`Recovered session ${sessionId} for project ${projectId} (migrated from another project)`)
        this.sessionSandboxes.set(sessionId, { id: migrated.sandboxId })
        // Re-run getSandbox to go through the normal reconnect path.
        return this.getSandbox(sessionId, projectId, worktree, pluginCtx)
      }
    }

    // Otherwise, create a new sandbox
    logger.info(`Creating new sandbox for session: ${sessionId} in project: ${projectId}`)
    const daytona = new Daytona({ apiKey: this.apiKey })
    const sandbox = await daytona.create()
    this.sessionSandboxes.set(sessionId, sandbox)
    // Get or assign branch number for this sandbox
    let branchNumber = this.dataStorage.getBranchNumberForSandbox(projectId, sandbox.id)
    if (!branchNumber) {
      branchNumber = this.dataStorage.getNextBranchNumber(projectId)
    }
    this.dataStorage.updateSession(projectId, worktree, sessionId, sandbox.id, branchNumber)
    logger.info(`Sandbox created successfully: ${sandbox.id} with branch number ${branchNumber}`)

    // Initialize git repo in the sandbox and sync with host
    try {
      const sessionGit = new SessionGitManager(sandbox, this.repoPath, branchNumber)
      await sessionGit.initializeAndSync(pluginCtx)
    } catch (err: any) {
      logger.error(`Failed to initialize git repo or push local changes in sandbox: ${err}`)
      toast.show({
        title: 'Git error',
        message: err?.message || 'Failed to initialize git repo in sandbox.',
        variant: 'error',
      })
    }
    toast.show({
      title: 'Sandbox created',
      message: `Created new sandbox for session.`,
      variant: 'success',
    })
    return sandbox
  }

  /**
   * Delete the sandbox associated with the given session ID
   */
  async deleteSandbox(sessionId: string, projectId: string): Promise<void> {
    let sandbox = this.sessionSandboxes.get(sessionId)

    // If not in cache, try to load from storage and reconnect
    if (!sandbox || this.isPartiallyInitialized(sandbox)) {
      const storedWorktree = this.dataStorage.load(projectId)?.worktree ?? ''
      const sessionInfo = this.dataStorage.getSession(projectId, storedWorktree, sessionId)
      if (sessionInfo?.sandboxId) {
        const daytona = new Daytona({ apiKey: this.apiKey })
        try {
          sandbox = await daytona.get(sessionInfo.sandboxId)
          this.sessionSandboxes.set(sessionId, sandbox)
        } catch (err) {
          logger.error(`Failed to reconnect to sandbox ${sessionInfo.sandboxId}: ${err}`)
        }
      }
    }

    // Delete the sandbox if we have a fully initialized one
    if (this.isFullyInitialized(sandbox)) {
      logger.info(`Removing sandbox for session: ${sessionId}`)
      await sandbox.delete()
      this.sessionSandboxes.delete(sessionId)

      // Remove from storage
      const projectData = this.dataStorage.load(projectId)
      if (projectData) {
        this.dataStorage.removeSession(projectId, projectData.worktree, sessionId)
      }

      logger.info(`Sandbox deleted successfully.`)
    } else {
      logger.warn(`No sandbox found for session: ${sessionId}`)
    }
  }
}
