/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  ToolboxApi,
  SandboxState,
  SandboxApi,
  Sandbox as SandboxDto,
  PortPreviewUrl,
  SandboxVolume,
  BuildInfo,
  SandboxBackupStateEnum,
  Configuration,
} from '@daytonaio/api-client'
import { FileSystem } from './FileSystem'
import { Git } from './Git'
import { CodeRunParams, Process } from './Process'
import { LspLanguageId, LspServer } from './LspServer'
import { DaytonaError } from './errors/DaytonaError'
import { prefixRelativePath } from './utils/Path'
import { ComputerUse } from './ComputerUse'

/**
 * Interface defining methods that a code toolbox must implement
 * @interface
 */
export interface SandboxCodeToolbox {
  /** Generates a command to run the provided code */
  getRunCommand(code: string, params?: CodeRunParams): string
}

/**
 * Represents a Daytona Sandbox.
 *
 * @property {FileSystem} fs - File system operations interface
 * @property {Git} git - Git operations interface
 * @property {Process} process - Process execution interface
 * @property {ComputerUse} computerUse - Computer use operations interface for desktop automation
 * @property {string} id - Unique identifier for the Sandbox
 * @property {string} organizationId - Organization ID of the Sandbox
 * @property {string} [snapshot] - Daytona snapshot used to create the Sandbox
 * @property {string} user - OS user running in the Sandbox
 * @property {Record<string, string>} env - Environment variables set in the Sandbox
 * @property {Record<string, string>} labels - Custom labels attached to the Sandbox
 * @property {boolean} public - Whether the Sandbox is publicly accessible
 * @property {string} target - Target location of the runner where the Sandbox runs
 * @property {number} cpu - Number of CPUs allocated to the Sandbox
 * @property {number} gpu - Number of GPUs allocated to the Sandbox
 * @property {number} memory - Amount of memory allocated to the Sandbox in GiB
 * @property {number} disk - Amount of disk space allocated to the Sandbox in GiB
 * @property {SandboxState} state - Current state of the Sandbox (e.g., "started", "stopped")
 * @property {string} [errorReason] - Error message if Sandbox is in error state
 * @property {SandboxBackupStateEnum} [backupState] - Current state of Sandbox backup
 * @property {string} [backupCreatedAt] - When the backup was created
 * @property {number} [autoStopInterval] - Auto-stop interval in minutes
 * @property {number} [autoArchiveInterval] - Auto-archive interval in minutes
 * @property {number} [autoDeleteInterval] - Auto-delete interval in minutes
 * @property {string} [runnerDomain] - Domain name of the Sandbox runner
 * @property {Array<SandboxVolume>} [volumes] - Volumes attached to the Sandbox
 * @property {BuildInfo} [buildInfo] - Build information for the Sandbox if it was created from dynamic build
 * @property {string} [createdAt] - When the Sandbox was created
 * @property {string} [updatedAt] - When the Sandbox was last updated
 *
 * @class
 */
export class Sandbox implements SandboxDto {
  public readonly fs: FileSystem
  public readonly git: Git
  public readonly process: Process
  public readonly computerUse: ComputerUse

  public id!: string
  public organizationId!: string
  public snapshot?: string
  public user!: string
  public env!: Record<string, string>
  public labels!: Record<string, string>
  public public!: boolean
  public target!: string
  public cpu!: number
  public gpu!: number
  public memory!: number
  public disk!: number
  public state?: SandboxState
  public errorReason?: string
  public backupState?: SandboxBackupStateEnum
  public backupCreatedAt?: string
  public autoStopInterval?: number
  public autoArchiveInterval?: number
  public autoDeleteInterval?: number
  public runnerDomain?: string
  public volumes?: Array<SandboxVolume>
  public buildInfo?: BuildInfo
  public createdAt?: string
  public updatedAt?: string

  private rootDir: string

  /**
   * Creates a new Sandbox instance
   *
   * @param {SandboxDto} sandboxDto - The API Sandbox instance
   * @param {SandboxApi} sandboxApi - API client for Sandbox operations
   * @param {ToolboxApi} toolboxApi - API client for toolbox operations
   * @param {SandboxCodeToolbox} codeToolbox - Language-specific toolbox implementation
   */
  constructor(
    sandboxDto: SandboxDto,
    private readonly clientConfig: Configuration,
    private readonly sandboxApi: SandboxApi,
    private readonly toolboxApi: ToolboxApi,
    private readonly codeToolbox: SandboxCodeToolbox,
  ) {
    this.processSandboxDto(sandboxDto)
    this.rootDir = ''
    this.fs = new FileSystem(this.id, this.toolboxApi, async () => await this.getRootDir())
    this.git = new Git(this.id, this.toolboxApi, async () => await this.getRootDir())
    this.process = new Process(
      this.id,
      this.clientConfig,
      this.codeToolbox,
      this.toolboxApi,
      async () => await this.getRootDir(),
    )
    this.computerUse = new ComputerUse(this.id, this.toolboxApi)
  }

  /**
   * Gets the root directory path for the logged in user inside the Sandbox.
   *
   * @returns {Promise<string | undefined>} The absolute path to the Sandbox root directory for the logged in user
   *
   * @example
   * const rootDir = await sandbox.getUserRootDir();
   * console.log(`Sandbox root: ${rootDir}`);
   */
  public async getUserRootDir(): Promise<string | undefined> {
    const response = await this.toolboxApi.getProjectDir(this.id)
    return response.data.dir
  }

  /**
   * Creates a new Language Server Protocol (LSP) server instance.
   *
   * The LSP server provides language-specific features like code completion,
   * diagnostics, and more.
   *
   * @param {LspLanguageId} languageId - The language server type (e.g., "typescript")
   * @param {string} pathToProject - Path to the project root directory. Relative paths are resolved based on the user's
   * root directory.
   * @returns {LspServer} A new LSP server instance configured for the specified language
   *
   * @example
   * const lsp = await sandbox.createLspServer('typescript', 'workspace/project');
   */
  public async createLspServer(languageId: LspLanguageId | string, pathToProject: string): Promise<LspServer> {
    return new LspServer(
      languageId as LspLanguageId,
      prefixRelativePath(await this.getRootDir(), pathToProject),
      this.toolboxApi,
      this.id,
    )
  }

  /**
   * Sets labels for the Sandbox.
   *
   * Labels are key-value pairs that can be used to organize and identify Sandboxes.
   *
   * @param {Record<string, string>} labels - Dictionary of key-value pairs representing Sandbox labels
   * @returns {Promise<void>}
   *
   * @example
   * // Set sandbox labels
   * await sandbox.setLabels({
   *   project: 'my-project',
   *   environment: 'development',
   *   team: 'backend'
   * });
   */
  public async setLabels(labels: Record<string, string>): Promise<Record<string, string>> {
    this.labels = (await this.sandboxApi.replaceLabels(this.id, { labels })).data.labels
    return this.labels
  }

  /**
   * Start the Sandbox.
   *
   * This method starts the Sandbox and waits for it to be ready.
   *
   * @param {number} [timeout] - Maximum time to wait in seconds. 0 means no timeout.
   *                            Defaults to 60-second timeout.
   * @returns {Promise<void>}
   * @throws {DaytonaError} - `DaytonaError` - If Sandbox fails to start or times out
   *
   * @example
   * const sandbox = await daytona.getCurrentSandbox('my-sandbox');
   * await sandbox.start(40);  // Wait up to 40 seconds
   * console.log('Sandbox started successfully');
   */
  public async start(timeout = 60): Promise<void> {
    if (timeout < 0) {
      throw new DaytonaError('Timeout must be a non-negative number')
    }
    const startTime = Date.now()
    const response = await this.sandboxApi.startSandbox(this.id, undefined, { timeout: timeout * 1000 })
    this.processSandboxDto(response.data)
    const timeElapsed = Date.now() - startTime
    await this.waitUntilStarted(timeout ? timeout - timeElapsed / 1000 : 0)
  }

  /**
   * Stops the Sandbox.
   *
   * This method stops the Sandbox and waits for it to be fully stopped.
   *
   * @param {number} [timeout] - Maximum time to wait in seconds. 0 means no timeout.
   *                            Defaults to 60-second timeout.
   * @returns {Promise<void>}
   *
   * @example
   * const sandbox = await daytona.getCurrentSandbox('my-sandbox');
   * await sandbox.stop();
   * console.log('Sandbox stopped successfully');
   */
  public async stop(timeout = 60): Promise<void> {
    if (timeout < 0) {
      throw new DaytonaError('Timeout must be a non-negative number')
    }
    const startTime = Date.now()
    await this.sandboxApi.stopSandbox(this.id, undefined, { timeout: timeout * 1000 })
    await this.refreshData()
    const timeElapsed = Date.now() - startTime
    await this.waitUntilStopped(timeout ? timeout - timeElapsed / 1000 : 0)
  }

  /**
   * Deletes the Sandbox.
   * @returns {Promise<void>}
   */
  public async delete(timeout = 60): Promise<void> {
    await this.sandboxApi.deleteSandbox(this.id, true, undefined, { timeout: timeout * 1000 })
    await this.refreshData()
  }

  /**
   * Waits for the Sandbox to reach the 'started' state.
   *
   * This method polls the Sandbox status until it reaches the 'started' state
   * or encounters an error.
   *
   * @param {number} [timeout] - Maximum time to wait in seconds. 0 means no timeout.
   *                               Defaults to 60 seconds.
   * @returns {Promise<void>}
   * @throws {DaytonaError} - `DaytonaError` - If the sandbox ends up in an error state or fails to start within the timeout period.
   */
  public async waitUntilStarted(timeout = 60) {
    if (timeout < 0) {
      throw new DaytonaError('Timeout must be a non-negative number')
    }

    const checkInterval = 100 // Wait 100 ms between checks
    const startTime = Date.now()

    while (this.state !== 'started') {
      await this.refreshData()

      // @ts-expect-error this.refreshData() can modify this.state so this check is fine
      if (this.state === 'started') {
        return
      }

      if (this.state === 'error') {
        const errMsg = `Sandbox ${this.id} failed to start with status: ${this.state}, error reason: ${this.errorReason}`
        throw new DaytonaError(errMsg)
      }

      if (timeout !== 0 && Date.now() - startTime > timeout * 1000) {
        throw new DaytonaError('Sandbox failed to become ready within the timeout period')
      }

      await new Promise((resolve) => setTimeout(resolve, checkInterval))
    }
  }

  /**
   * Wait for Sandbox to reach 'stopped' state.
   *
   * This method polls the Sandbox status until it reaches the 'stopped' state
   * or encounters an error.
   *
   * @param {number} [timeout] - Maximum time to wait in seconds. 0 means no timeout.
   *                               Defaults to 60 seconds.
   * @returns {Promise<void>}
   * @throws {DaytonaError} - `DaytonaError` - If the sandbox fails to stop within the timeout period.
   */
  public async waitUntilStopped(timeout = 60) {
    if (timeout < 0) {
      throw new DaytonaError('Timeout must be a non-negative number')
    }

    const checkInterval = 100 // Wait 100 ms between checks
    const startTime = Date.now()

    while (this.state !== 'stopped') {
      await this.refreshData()

      // @ts-expect-error this.refreshData() can modify this.state so this check is fine
      if (this.state === 'stopped') {
        return
      }

      if (this.state === 'error') {
        const errMsg = `Sandbox failed to stop with status: ${this.state}, error reason: ${this.errorReason}`
        throw new DaytonaError(errMsg)
      }

      if (timeout !== 0 && Date.now() - startTime > timeout * 1000) {
        throw new DaytonaError('Sandbox failed to become stopped within the timeout period')
      }

      await new Promise((resolve) => setTimeout(resolve, checkInterval))
    }
  }

  /**
   * Refreshes the Sandbox data from the API.
   *
   * @returns {Promise<void>}
   *
   * @example
   * await sandbox.refreshData();
   * console.log(`Sandbox ${sandbox.id}:`);
   * console.log(`State: ${sandbox.state}`);
   * console.log(`Resources: ${sandbox.cpu} CPU, ${sandbox.memory} GiB RAM`);
   */
  public async refreshData(): Promise<void> {
    const response = await this.sandboxApi.getSandbox(this.id)
    this.processSandboxDto(response.data)
  }

  /**
   * Set the auto-stop interval for the Sandbox.
   *
   * The Sandbox will automatically stop after being idle (no new events) for the specified interval.
   * Events include any state changes or interactions with the Sandbox through the sdk.
   * Interactions using Sandbox Previews are not included.
   *
   * @param {number} interval - Number of minutes of inactivity before auto-stopping.
   *                           Set to 0 to disable auto-stop. Default is 15 minutes.
   * @returns {Promise<void>}
   * @throws {DaytonaError} - `DaytonaError` - If interval is not a non-negative integer
   *
   * @example
   * // Auto-stop after 1 hour
   * await sandbox.setAutostopInterval(60);
   * // Or disable auto-stop
   * await sandbox.setAutostopInterval(0);
   */
  public async setAutostopInterval(interval: number): Promise<void> {
    if (!Number.isInteger(interval) || interval < 0) {
      throw new DaytonaError('autoStopInterval must be a non-negative integer')
    }

    await this.sandboxApi.setAutostopInterval(this.id, interval)
    this.autoStopInterval = interval
  }

  /**
   * Set the auto-archive interval for the Sandbox.
   *
   * The Sandbox will automatically archive after being continuously stopped for the specified interval.
   *
   * @param {number} interval - Number of minutes after which a continuously stopped Sandbox will be auto-archived.
   *                           Set to 0 for the maximum interval. Default is 7 days.
   * @returns {Promise<void>}
   * @throws {DaytonaError} - `DaytonaError` - If interval is not a non-negative integer
   *
   * @example
   * // Auto-archive after 1 hour
   * await sandbox.setAutoArchiveInterval(60);
   * // Or use the maximum interval
   * await sandbox.setAutoArchiveInterval(0);
   */
  public async setAutoArchiveInterval(interval: number): Promise<void> {
    if (!Number.isInteger(interval) || interval < 0) {
      throw new DaytonaError('autoArchiveInterval must be a non-negative integer')
    }
    await this.sandboxApi.setAutoArchiveInterval(this.id, interval)
    this.autoArchiveInterval = interval
  }

  /**
   * Set the auto-delete interval for the Sandbox.
   *
   * The Sandbox will automatically delete after being continuously stopped for the specified interval.
   *
   * @param {number} interval - Number of minutes after which a continuously stopped Sandbox will be auto-deleted.
   *                           Set to negative value to disable auto-delete. Set to 0 to delete immediately upon stopping.
   *                           By default, auto-delete is disabled.
   * @returns {Promise<void>}
   *
   * @example
   * // Auto-delete after 1 hour
   * await sandbox.setAutoDeleteInterval(60);
   * // Or delete immediately upon stopping
   * await sandbox.setAutoDeleteInterval(0);
   * // Or disable auto-delete
   * await sandbox.setAutoDeleteInterval(-1);
   */
  public async setAutoDeleteInterval(interval: number): Promise<void> {
    await this.sandboxApi.setAutoDeleteInterval(this.id, interval)
    this.autoDeleteInterval = interval
  }

  /**
   * Retrieves the preview link for the sandbox at the specified port. If the port is closed,
   * it will be opened automatically. For private sandboxes, a token is included to grant access
   * to the URL.
   *
   * @param {number} port - The port to open the preview link on.
   * @returns {PortPreviewUrl} The response object for the preview link, which includes the `url`
   * and the `token` (to access private sandboxes).
   *
   * @example
   * const previewLink = await sandbox.getPreviewLink(3000);
   * console.log(`Preview URL: ${previewLink.url}`);
   * console.log(`Token: ${previewLink.token}`);
   */
  public async getPreviewLink(port: number): Promise<PortPreviewUrl> {
    return (await this.sandboxApi.getPortPreviewUrl(this.id, port)).data
  }

  /**
   * Archives the sandbox, making it inactive and preserving its state. When sandboxes are archived, the entire filesystem
   * state is moved to cost-effective object storage, making it possible to keep sandboxes available for an extended period.
   * The tradeoff between archived and stopped states is that starting an archived sandbox takes more time, depending on its size.
   * Sandbox must be stopped before archiving.
   */
  public async archive(): Promise<void> {
    await this.sandboxApi.archiveSandbox(this.id)
    await this.refreshData()
  }

  private async getRootDir(): Promise<string> {
    if (!this.rootDir) {
      this.rootDir = (await this.getUserRootDir()) || ''
    }
    return this.rootDir
  }

  /**
   * Assigns the API sandbox data to the Sandbox object.
   *
   * @param {SandboxDto} sandboxDto - The API sandbox instance to assign data from
   * @returns {void}
   */
  private processSandboxDto(sandboxDto: SandboxDto) {
    this.id = sandboxDto.id
    this.organizationId = sandboxDto.organizationId
    this.snapshot = sandboxDto.snapshot
    this.user = sandboxDto.user
    this.env = sandboxDto.env
    this.labels = sandboxDto.labels
    this.public = sandboxDto.public
    this.target = sandboxDto.target
    this.cpu = sandboxDto.cpu
    this.gpu = sandboxDto.gpu
    this.memory = sandboxDto.memory
    this.disk = sandboxDto.disk
    this.state = sandboxDto.state
    this.errorReason = sandboxDto.errorReason
    this.backupState = sandboxDto.backupState
    this.backupCreatedAt = sandboxDto.backupCreatedAt
    this.autoStopInterval = sandboxDto.autoStopInterval
    this.autoArchiveInterval = sandboxDto.autoArchiveInterval
    this.autoDeleteInterval = sandboxDto.autoDeleteInterval
    this.runnerDomain = sandboxDto.runnerDomain
    this.volumes = sandboxDto.volumes
    this.buildInfo = sandboxDto.buildInfo
    this.createdAt = sandboxDto.createdAt
    this.updatedAt = sandboxDto.updatedAt
  }
}
